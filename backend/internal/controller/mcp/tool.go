package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/cache"
	protocol "github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/protocol/p250618"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/jsonrpc"
	"github.com/mustafaturan/monoflake"
	zlog "github.com/rs/zerolog/log"
)

const (
	// MethodToolsList lists all available executable tools.
	// https://modelcontextprotocol.io/specification/2024-11-05/server/tools/
	MethodToolsList Method = "tools/list"

	// MethodToolsCall invokes a specific tool with provided parameters.
	// https://modelcontextprotocol.io/specification/2024-11-05/server/tools/
	MethodToolsCall Method = "tools/call"
)

type (
	InputSchema struct {
		PathArgs  json.RawMessage `json:"pathArgs,omitempty"`
		QueryArgs json.RawMessage `json:"queryArgs,omitempty"`
		BodyArgs  json.RawMessage `json:"bodyArgs,omitempty"`
	}
)

const (
	_argsPathArgs  = "pathArgs"
	_argsQueryArgs = "queryArgs"
	_argsBodyArgs  = "bodyArgs"
)

var (
	_regexVariableName = regexp.MustCompile(`\$\{([A-Z0-9_]+)\}`)
)

func (c *controller) CallToolsList(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	srv, err := c.getServer(req.ServerID)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInvalidParams,
			Message: "server not found",
			Data: map[string]any{
				"reason": err.Error(),
			},
		}
	}

	var cursor int

	if len(req.Request.Params) > 0 {
		var params protocol.ListToolsRequestParams
		err = json.Unmarshal(req.Request.Params, &params)
		if err != nil {
			return nil, jsonrpc.Error{
				Code:    jsonrpc.ErrCodeInvalidJsonReceived,
				Message: "failed to parse json request",
				Data: map[string]any{
					"reason": err.Error(),
				},
			}
		}

		if params.Cursor != nil {
			cursor, err = strconv.Atoi(*params.Cursor)
			if err != nil {
				return nil, jsonrpc.Error{
					Code:    jsonrpc.ErrCodeInvalidParams,
					Message: "failed to parse cursor",
					Data: map[string]any{
						"reason": err.Error(),
					},
				}
			}
		}
	}

	toolIDs, cursor := paginate(srv.toolIDs, cursor, _paginationLimitToolsList)
	tools := make([]protocol.Tool, len(toolIDs))
	for i, id := range toolIDs {
		tool := srv.protocol.tools[id]
		tools[i] = tool
	}

	var nextCursor *string
	if cursor > 0 {
		nextCursor = stringPtr(strconv.Itoa(cursor))
	}

	response := protocol.ListToolsResult{
		NextCursor: nextCursor,
		Tools:      tools,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "failed to unmarshal response",
			Data: map[string]any{
				"reason": err.Error(),
			},
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

func (c *controller) CallToolsCall(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	var params protocol.CallToolRequestParams
	err := json.Unmarshal(req.Request.Params, &params)
	if err != nil {
		return nil, err
	}

	server, err := c.getServer(req.ServerID)
	if err != nil {
		return nil, err
	}

	toolIDPart := ""
	if len(params.Name) >= 12 {
		toolIDPart = params.Name[1:12]
	}

	toolID := monoflake.IDFromBase62(toolIDPart).Int64()
	_, ok := server.protocol.tools[toolID]
	if !ok {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInvalidParams,
			Message: "Tool not found",
			Data: map[string]any{
				"toolName": params.Name,
			},
		}
	}

	// verify the tool schema
	pathArgs := params.Arguments[_argsPathArgs]
	queryArgs := params.Arguments[_argsQueryArgs]
	bodyArgs := params.Arguments[_argsBodyArgs]

	tool, err := c.cache.GetTool(ctx, toolID)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "Tool tool not found",
			Data: map[string]any{
				"reason":   err.Error(),
				"toolName": params.Name,
			},
		}
	}

	provider, err := c.cache.GetProvider(ctx, tool.ProviderID)
	if err != nil {
		zlog.Error().Err(err).
			Str("toolName", params.Name).
			Int64("epID", tool.ID).
			Int64("pID", tool.ProviderID).
			Msg("tool provider is not found")
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "Tool provider is not found",
			Data: map[string]any{
				"reason":     err.Error(),
				"toolName":   params.Name,
				"providerID": tool.ProviderID,
				"toolID":     tool.ID,
			},
		}
	}

	url, err := buildURL(provider.BaseURL, tool.Path, pathArgs, queryArgs)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "Tool tool url is malformed",
			Data: map[string]any{
				"reason":   err.Error(),
				"toolName": params.Name,
			},
		}
	}

	callerHeaders := map[string][]string{}
	if server.requestHeadersProxyEnabled {
		callerHeaders = req.Headers
	}

	headers := buildHeaders(callerHeaders, tool.Headers, c.cache)
	remoteReq := &http.Request{
		Method: tool.Method.String(),
		URL:    url,
		Header: headers,
		Body:   io.NopCloser(bytes.NewReader(bodyArgs)),
	}

	res, err := c.httpc.Call(ctx, remoteReq)
	if err != nil {
		return nil, err
	}

	body := res.Body
	defer body.Close()
	resBody, _ := io.ReadAll(body)

	resPayload := protocol.CallToolResult{
		Content: []protocol.ContentBlock{
			protocol.TextContent{
				Text: string(resBody),
				Type: "text",
			},
		},
	}

	result, err := json.Marshal(resPayload)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "Tool tool url is malformed",
			Data: map[string]any{
				"reason":   err.Error(),
				"toolName": params.Name,
			},
		}
	}

	return &CallSessionResponse{
		HTTPStatusCode:     200,
		McpSessionID:       req.McpSessionID,
		McpProtocolVersion: req.McpProtocolVersion,
		Result: &jsonrpc.ResultResponse{
			JSONRpc: jsonrpc.Version,
			ID:      req.Request.ID,
			Result:  result,
		},
	}, nil
}

func buildHeaders(callerHeaders map[string][]string, toolHeaders []entity.ToolHeader, cache cache.Controller) http.Header {
	headers := http.Header{}
	// Pass proxy headers
	for k, vals := range callerHeaders {
		for _, v := range vals {
			headers.Add(k, v)
		}
	}
	for _, h := range toolHeaders {
		key, val := h.Key, h.Value
		if len(callerHeaders[key]) > 0 {
			continue
		}
		names := extractVariables(val)
		for _, n := range names {
			v, err := cache.GetVariable(context.Background(), n)
			if err != nil {
				continue
			}
			val = strings.Replace(val, "${"+n+"}", v, 1)
		}
		headers.Add(key, val)
	}
	return headers
}

func extractVariables(s string) []string {
	allMatches := _regexVariableName.FindAllStringSubmatch(s, -1)

	var names []string
	for _, match := range allMatches {
		if len(match) > 1 {
			names = append(names, match[1])
		}
	}

	return names
}

func buildURL(baseURL, path string, pathArgs, queryArgs json.RawMessage) (*url.URL, error) {
	builtPath := buildPath(string(path), pathArgs)
	builtQuery := buildQuery(queryArgs)
	return url.Parse(baseURL + builtPath + builtQuery)
}

func buildPath(pathTemplate string, pathArgs json.RawMessage) string {
	if len(pathArgs) == 0 {
		return pathTemplate
	}
	var args map[string]string
	err := json.Unmarshal(pathArgs, &args)
	if err != nil {
		return ""
	}

	for key, value := range args {
		pathTemplate = strings.ReplaceAll(pathTemplate, "{"+key+"}", value)
	}

	return pathTemplate
}

func buildQuery(queryArgs json.RawMessage) string {
	if len(queryArgs) == 0 {
		return ""
	}
	var args map[string]string
	err := json.Unmarshal(queryArgs, &args)
	if err != nil {
		return ""
	}

	queries := make([]string, 0, len(args))
	for key, value := range args {
		queries = append(queries, key+"="+value)
	}
	if len(queries) > 0 {
		return "?" + strings.Join(queries, "&")
	}
	return ""
}
