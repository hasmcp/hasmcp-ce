package mcp

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/jsonrpc"
	"github.com/hasmcp/hasmcp-ce/backend/internal/handler/mcp/middleware/jwt"
	zlog "github.com/rs/zerolog/log"
)

// Common HTTP header constants used across transports
const (
	_headerKeyLastEventID     = "last-event-id"
	_headerKeyProtocolVersion = "mcp-protocol-version"
	_headerKeySessionID       = "mcp-session-id"
	_hedaerKeyXHasMcpKey      = "x-hasmcp-key"
	_headerKeyAcceptEncoding  = "accept-encoding"
)

var (
	// _excludeHeaders is used to exclude headers from the callsessionrequest
	// the rest of the headers will be passed through to the actual service
	_excludeHeaders = map[string]struct{}{
		_headerKeyLastEventID:     {},
		_headerKeyProtocolVersion: {},
		_headerKeySessionID:       {},
		_hedaerKeyXHasMcpKey:      {},
		_headerKeyAcceptEncoding:  {},
	}
)

func FromHTTPRequestToMcpCallSessionRequest(c *fiber.Ctx) *mcp.CallSessionRequest {
	headers := c.GetReqHeaders()
	for k := range headers {
		sk := strings.ToLower(k)
		if _, ok := _excludeHeaders[sk]; ok {
			delete(headers, k)
		}
	}

	authRes := jwt.AuthResult(c)
	mcpSessionID := c.Get(_headerKeySessionID)
	mcpProtocolVersion := c.Get(_headerKeyProtocolVersion)

	var request jsonrpc.Request
	err := json.Unmarshal(c.BodyRaw(), &request)
	if err != nil {
		return nil
	}

	if request.JSONRpc != "2.0" {
		return nil
	}

	return &mcp.CallSessionRequest{
		Headers:            headers,
		ServerID:           authRes.ServerID,
		McpSessionID:       mcpSessionID,
		McpProtocolVersion: mcpProtocolVersion,
		Request:            request,
		Permissions:        authRes.Permissions,
	}
}

func FromMcpCallSessionResponseToHTTPResponse(res mcp.CallSessionResponse) []byte {
	if res.Result == nil {
		return []byte("")
	}
	data, err := json.Marshal(res.Result)
	if err != nil {
		zlog.Error().Err(err).Any("jsonRpcResult", res.Result).Msg("failed to marshal the jsonrpc response")
		return []byte("")
	}
	return data
}

func FromHTTPRequestToMcpSubscribeSessionRequest(c *fiber.Ctx) mcp.SubscribeSessionRequest {
	authRes := jwt.AuthResult(c)
	mcpSessionID := c.Get(_headerKeySessionID)
	mcpProtocolVersion := c.Get(_headerKeyProtocolVersion)
	lastEventID := c.Get(_headerKeyLastEventID)
	return mcp.SubscribeSessionRequest{
		ServerID:           authRes.ServerID,
		McpSessionID:       mcpSessionID,
		McpProtocolVersion: mcpProtocolVersion,
		LastEventID:        lastEventID,
		Permissions:        authRes.Permissions,
	}
}

func FromHTTPRequestToMcpStartTailIORequest(c *fiber.Ctx) mcp.StartTailIORequest {
	authRes := jwt.AuthResult(c)
	return mcp.StartTailIORequest{
		ServerID:    authRes.ServerID,
		Permissions: authRes.Permissions,
	}
}

func FromHTTPRequestToMcpDeleteSessionRequest(c *fiber.Ctx) mcp.DeleteSessionRequest {
	authRes := jwt.AuthResult(c)
	mcpSessionID := c.Get(_headerKeySessionID)
	mcpProtocolVersion := c.Get(_headerKeyProtocolVersion)
	return mcp.DeleteSessionRequest{
		ServerID:           authRes.ServerID,
		McpSessionID:       mcpSessionID,
		McpProtocolVersion: mcpProtocolVersion,
		Permissions:        authRes.Permissions,
	}
}
