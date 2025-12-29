package api

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	view "github.com/hasmcp/hasmcp-ce/backend/internal/data/view/api"
	"github.com/mustafaturan/monoflake"
)

func FromHTTPRequestToCreateServerTokenRequestEntity(c *fiber.Ctx) *entity.CreateServerTokenRequest {
	var payload view.CreateServerTokenRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}

	data := payload.Token
	data.ServerID = c.Params("id")

	return &entity.CreateServerTokenRequest{
		Token: FromServerTokenViewToServerTokenEntity(data),
	}
}

func FromServerTokenViewToServerTokenEntity(t view.ServerToken) entity.ServerToken {
	expiresAt, _ := time.Parse(time.RFC3339, t.ExpiresAt)
	return entity.ServerToken{
		ServerID:  monoflake.IDFromBase62(t.ServerID).Int64(),
		ExpiresAt: expiresAt,
		Scope:     t.Scope,
	}
}

func FromServerTokenEntitiesToServerTokenViews(ts []entity.ServerToken) []view.ServerToken {
	tokens := make([]view.ServerToken, len(ts))
	for i, t := range ts {
		tokens[i] = FromServerTokenEntityToServerTokenView(t)
	}
	return tokens
}

func FromServerTokenEntityToServerTokenView(t entity.ServerToken) view.ServerToken {
	return view.ServerToken{
		ID:        monoflake.ID(t.ID).String(),
		ServerID:  monoflake.ID(t.ServerID).String(),
		CreatedAt: FromTimeToRFC3339String(t.CreatedAt),
		ExpiresAt: FromTimeToRFC3339String(t.ExpiresAt),
		Scope:     t.Scope,
		Value:     string(t.ActualValue),
	}
}

func FromCreateServerTokenResponseEntityToHTTPResponse(rs *entity.CreateServerTokenResponse) []byte {
	t := rs.ServerToken
	resp := view.CreateServerTokenResponse{
		Token: FromServerTokenEntityToServerTokenView(t),
	}

	payload, _ := json.Marshal(resp)
	return payload
}
