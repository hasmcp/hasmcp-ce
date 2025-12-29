package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"text/template"

	protocol "github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/protocol/p250618"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/jsonrpc"
	"github.com/mustafaturan/monoflake"
	zlog "github.com/rs/zerolog/log"
)

const (
	// MethodPromptsList lists all available prompt templates.
	// https://modelcontextprotocol.io/specification/2025-06-18/server/prompts
	MethodPromptsList Method = "prompts/list"

	// MethodPromptsGet retrieves a specific prompt template with filled parameters.
	// https://modelcontextprotocol.io/specification/2025-06-18/server/prompts
	MethodPromptsGet Method = "prompts/get"
)

func (c *controller) CallPromptsList(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	srv, err := c.getServer(req.ServerID)
	if err != nil {
		return nil, err
	}
	var cursor int
	if len(req.Request.Params) > 0 {
		var params protocol.ListPromptsRequestParams
		if err := json.Unmarshal(req.Request.Params, &params); err != nil {
			return nil, jsonrpc.Error{
				Code:    jsonrpc.ErrCodeInvalidJsonReceived,
				Message: "invalid params for prompts/list",
				Data:    map[string]any{"reason": err.Error()},
			}
		}
		cursor, _ = strconv.Atoi(stringPtrToString(params.Cursor))
	}

	// Get all prompt IDs for this server
	allPromptIDs := make([]int64, 0, len(srv.protocol.prompts))
	for id := range srv.protocol.prompts {
		allPromptIDs = append(allPromptIDs, id)
	}
	// Sort for consistent pagination
	sort.Slice(allPromptIDs, func(i, j int) bool { return allPromptIDs[i] < allPromptIDs[j] })

	// Paginate
	paginatedIDs, nextCursorIndex := paginate(allPromptIDs, cursor, _paginationLimitPromptList)

	// Build result
	prompts := make([]protocol.Prompt, len(paginatedIDs))
	for i, id := range paginatedIDs {
		prompts[i] = srv.protocol.prompts[id]
	}

	var nextCursor *string
	if nextCursorIndex != -1 {
		nextCursor = stringPtr(strconv.Itoa(nextCursorIndex))
	}

	response := protocol.ListPromptsResult{
		Prompts:    prompts,
		NextCursor: nextCursor,
	}
	data, err := json.Marshal(response)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "failed to marshal prompts/list response",
			Data:    map[string]any{"reason": err.Error()},
		}
	}

	return &CallSessionResponse{
		HTTPStatusCode:     200,
		McpSessionID:       req.McpSessionID,
		McpProtocolVersion: req.McpProtocolVersion,
		Result: &jsonrpc.ResultResponse{
			JSONRpc: jsonrpc.Version,
			Result:  data,
			ID:      req.Request.ID,
		},
	}, nil
}

func (c *controller) CallPromptsGet(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	srv, err := c.getServer(req.ServerID)
	if err != nil {
		return nil, err
	}

	var params protocol.GetPromptRequestParams
	if err := json.Unmarshal(req.Request.Params, &params); err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInvalidJsonReceived,
			Message: "invalid params for prompts/get",
			Data:    map[string]any{"reason": err.Error()},
		}
	}

	// Find the prompt by name (which is the monoflake ID string)
	promptIDPart := ""
	if len(params.Name) >= 12 {
		promptIDPart = params.Name[1:12]
	}

	promptID := monoflake.IDFromBase62(promptIDPart).Int64()
	if promptID == 0 {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInvalidParams,
			Message: "prompt not found",
			Data:    map[string]any{"name": params.Name},
		}
	}

	protocolPrompt, ok := srv.protocol.prompts[promptID]
	if !ok {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInvalidParams,
			Message: "prompt not found for this server",
			Data:    map[string]any{"name": params.Name},
		}
	}

	// Get full prompt data from storage
	dbPrompt, err := c.cache.GetPrompt(ctx, promptID)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeServerError,
			Message: "failed to retrieve prompt data",
			Data:    map[string]any{"id": promptID, "reason": err.Error()},
		}
	}

	var messages []protocol.PromptMessage
	if err := json.Unmarshal(dbPrompt.Messages, &messages); err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "failed to parse prompt messages",
			Data:    map[string]any{"id": promptID, "reason": err.Error()},
		}
	}

	// --- Start Templating Logic ---
	// If arguments are provided, apply them to the messages
	if len(params.Arguments) > 0 {
		for i := range messages {
			// Content is unmarshalled as map[string]any because it's an interface
			if contentMap, ok := messages[i].Content.(map[string]any); ok {
				// Check if it's TextContent
				if typeVal, ok := contentMap["type"]; ok && typeVal == "text" {
					// Check if text field exists
					if textVal, ok := contentMap["text"]; ok {
						// Cast text to string
						if textStr, ok := textVal.(string); ok {
							// Use text/template for replacement
							tmpl, err := template.New("prompt").Parse(textStr)
							if err != nil {
								zlog.Warn().Err(err).Int64("promptID", promptID).Msg("failed to parse prompt template")
								// Don't fail the request, just use the un-templated string
								continue
							}

							var buf bytes.Buffer
							if err := tmpl.Execute(&buf, params.Arguments); err != nil {
								zlog.Warn().Err(err).Int64("promptID", promptID).Msg("failed to execute prompt template")
								// Don't fail the request, just use the un-templated string
								continue
							}

							// Update the map with the new templated text
							contentMap["text"] = buf.String()
						}
					}
				}
			}
		}
	}
	// --- End Templating Logic ---

	response := protocol.GetPromptResult{
		Description: protocolPrompt.Description,
		Messages:    messages,
	}
	data, err := json.Marshal(response)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "failed to marshal prompts/get response",
			Data:    map[string]any{"reason": err.Error()},
		}
	}

	return &CallSessionResponse{
		HTTPStatusCode: 200,
		McpSessionID:   req.McpSessionID,
		Result: &jsonrpc.ResultResponse{
			JSONRpc: jsonrpc.Version,
			Result:  data,
			ID:      req.Request.ID,
		},
	}, nil
}
