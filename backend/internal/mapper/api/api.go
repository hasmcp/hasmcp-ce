package api

import "time"

var (
	_unsetTime time.Time
)

func FromTimeToRFC3339String(t time.Time) string {
	if t.Equal(_unsetTime) {
		return ""
	}

	return t.Format(time.RFC3339)
}
