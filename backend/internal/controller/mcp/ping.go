package mcp

import (
	"context"

	"github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/jsonrpc"
)

const (
	// MethodPing verifies connection liveness between client and server.
	// https://modelcontextprotocol.io/specification/2024-11-05/basic/utilities/ping/
	MethodPing Method = "ping"
)

func (c *controller) CallPing(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	return &CallSessionResponse{
		HTTPStatusCode:     200,
		McpSessionID:       req.McpSessionID,
		McpProtocolVersion: req.McpProtocolVersion,
		Result: &jsonrpc.ResultResponse{
			JSONRpc: jsonrpc.Version,
			ID:      req.Request.ID,
			Result:  []byte("{}"),
		},
	}, nil
}
