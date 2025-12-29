package oauth2mcp

import (
	"encoding/json"

	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	view "github.com/hasmcp/hasmcp-ce/backend/internal/data/view/api"
)

func FromErrorEntityToErrorView(e erre.Error) view.Error {
	return view.Error{
		Code:    int(e.Code),
		Message: e.Message,
		Data:    e.Data,
	}
}

func FromErrorEntityToHTTPResponse(e erre.Error) []byte {
	err := FromErrorEntityToErrorView(e)
	data := map[string]view.Error{
		"error": err,
	}
	payload, _ := json.Marshal(data)
	return payload
}

func FromErrorToHTTPResponse(err error) ([]byte, int) {
	e, ok := err.(erre.Error)
	if !ok {
		e = erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "internal server error",
			Data: map[string]any{
				"reason": err.Error(),
			},
		}
	}
	return FromErrorEntityToHTTPResponse(e), int(e.Code)
}
