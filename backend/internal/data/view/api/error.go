package view

type Error struct {
	Code    int            `json:"code,omitempty"`
	Message string         `json:"message,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}
