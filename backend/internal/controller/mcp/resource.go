package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	protocol "github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/protocol/p250618"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/jsonrpc"
)

const (
	// MethodResourcesList lists all available server resources.
	// https://modelcontextprotocol.io/specification/2024-11-05/server/resources/
	MethodResourcesList Method = "resources/list"

	// MethodResourcesRead retrieves content of a specific resource by URI.
	// https://modelcontextprotocol.io/specification/2024-11-05/server/resources/
	MethodResourcesRead Method = "resources/read"

	// MethodResourcesRead retrieves content of a specific resource by URI.
	// https://modelcontextprotocol.io/specification/2025-03-26/server/resources#subscriptions
	MethodResourcesSubscribe Method = "resources/subscribe"

	// MethodResourcesTemplatesList provides URI templates for constructing resource URIs.
	// https://modelcontextprotocol.io/specification/2024-11-05/server/resources/
	MethodResourcesTemplatesList Method = "resources/templates/list"
)

func (c *controller) CallResourcesList(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
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
		var params protocol.ListResourcesRequestParams
		if err := json.Unmarshal(req.Request.Params, &params); err != nil {
			return nil, jsonrpc.Error{
				Code:    jsonrpc.ErrCodeInvalidJsonReceived,
				Message: "invalid json for resources/list",
				Data:    map[string]any{"reason": err.Error()},
			}
		}
		cursor, _ = strconv.Atoi(stringPtrToString(params.Cursor))
	}

	// Get all resource IDs for this server
	allResourceIDs := srv.resourceIDs

	// Paginate
	paginatedIDs, nextCursorIndex := paginate(allResourceIDs, cursor, _paginationLimitResourceList)

	// Build result
	resources := make([]protocol.Resource, len(paginatedIDs))
	for i, id := range paginatedIDs {
		resources[i] = srv.protocol.resources[id]
	}

	var nextCursor *string
	if nextCursorIndex != -1 {
		nextCursor = stringPtr(strconv.Itoa(nextCursorIndex))
	}

	response := protocol.ListResourcesResult{
		Resources:  resources,
		NextCursor: nextCursor,
	}
	data, err := json.Marshal(response)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "failed to marshal resources/list response",
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

func (c *controller) CallResourcesRead(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
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

	var params protocol.ReadResourceRequestParams
	if err := json.Unmarshal(req.Request.Params, &params); err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInvalidJsonReceived,
			Message: "invalid json for resources/read",
			Data:    map[string]any{"reason": err.Error()},
		}
	}

	// Find the resource by URI
	var foundResource *protocol.Resource
	for _, res := range srv.protocol.resources {
		if res.Uri == params.Uri {
			r := res
			foundResource = &r
			break
		}
	}

	if foundResource == nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInvalidParams,
			Message: "resource not found at specified URI",
			Data:    map[string]any{"uri": params.Uri},
		}
	}

	// Fetch the resource content via HTTP
	httpReq, err := http.NewRequestWithContext(ctx, "GET", params.Uri, nil)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeServerError,
			Message: "failed to create request for resource",
			Data:    map[string]any{"uri": params.Uri, "reason": err.Error()},
		}
	}

	httpRes, err := c.httpc.Call(ctx, httpReq)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeServerError,
			Message: "failed to fetch resource content",
			Data:    map[string]any{"uri": params.Uri, "reason": err.Error()},
		}
	}
	defer httpRes.Body.Close()

	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "failed to read resource content",
			Data:    map[string]any{"uri": params.Uri, "reason": err.Error()},
		}
	}

	mimeType := httpRes.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = stringPtrToString(foundResource.MimeType)
	}

	resultContent := protocol.ReadResourceResultContentsElem{
		Uri:      params.Uri,
		MimeType: &mimeType,
	}

	if strings.HasPrefix(mimeType, "text/") {
		resultContent.Text = string(body)
	} else {
		resultContent.Blob = base64.StdEncoding.EncodeToString(body)
	}

	response := protocol.ReadResourceResult{
		Contents: []protocol.ReadResourceResultContentsElem{resultContent},
	}
	data, err := json.Marshal(response)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeServerError,
			Message: "failed to marshal resources/read response",
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

func (c *controller) CallResourcesTemplatesList(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	// TODO: This method is not implemented as the storage layer
	// does not currently have a model for "Resource Templates".
	return nil, ErrNotImplemented
}

func (c *controller) CallResourcesSubscribe(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	return nil, ErrNotImplemented
}
