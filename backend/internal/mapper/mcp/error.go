package mcp

import (
	"encoding/json"
	"net/http"

	"github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/jsonrpc"
)

func FromErrorToJsonRpcResponse(id any, err error) ([]byte, int) {
	e, ok := err.(jsonrpc.Error)
	if !ok {
		e = jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "internal server error",
			Data: map[string]any{
				"reason": err.Error(),
			},
		}
	}

	data := &jsonrpc.ErrorResponse{
		JSONRpc: jsonrpc.Version,
		ID:      id,
		Error:   e,
	}

	payload, _ := json.Marshal(data)
	return payload, fromJsonRpcErrorCodeToHTTPStatusCode(e.Code)
}

func fromJsonRpcErrorCodeToHTTPStatusCode(code int) int {
	switch code {
	case jsonrpc.ErrCodeInvalidJsonReceived:
		return http.StatusBadRequest
	case jsonrpc.ErrCodeInvalidRequest:
		return http.StatusBadRequest
	case jsonrpc.ErrCodeMethodNotFound:
		return http.StatusNotFound
	case jsonrpc.ErrCodeInvalidParams:
		return http.StatusBadRequest
	case jsonrpc.ErrCodeInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
