package mcp

import "context"

const (
	// MethodSetLogLevel configures the minimum log level for client
	// https://modelcontextprotocol.io/specification/2025-03-26/server/utilities/logging
	MethodSetLogLevel Method = "logging/setLevel"
)

func (c *controller) CallLoggingSetLevel(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	return nil, ErrNotImplemented
}
