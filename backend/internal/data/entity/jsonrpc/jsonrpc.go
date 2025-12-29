package jsonrpc

import "encoding/json"

type (
	Request struct {
		JSONRpc string          `json:"jsonrpc"`
		ID      any             `json:"id,omitempty"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params,omitempty"`
	}

	ResultResponse struct {
		JSONRpc string          `json:"jsonrpc"`
		Result  json.RawMessage `json:"result"`
		ID      any             `json:"id"`
	}

	ErrorResponse struct {
		JSONRpc string `json:"jsonrpc"`
		Error   Error  `json:"error"`
		ID      any    `json:"id,omitempty"`
	}

	Error struct {
		Code    int            `json:"code"`
		Message string         `json:"message"`
		Data    map[string]any `json:"data"`
	}
)

const (
	ErrCodeInvalidJsonReceived = -32700
	ErrCodeInvalidRequest      = -32600
	ErrCodeMethodNotFound      = -32601
	ErrCodeInvalidParams       = -32602
	ErrCodeInternalError       = -32603
	ErrCodeServerError         = -32000
)

const (
	Version = "2.0"
)

// Error implements standard error interface
func (e Error) Error() string {
	return e.Message
}
