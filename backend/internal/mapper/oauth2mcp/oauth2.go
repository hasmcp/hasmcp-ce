package oauth2mcp

import (
	"github.com/gofiber/fiber/v2"
	oauth2 "github.com/hasmcp/hasmcp-ce/backend/internal/controller/oauth2mcp"
	"github.com/mustafaturan/monoflake"
)

func FromHTTPRequestToOauth2AuthorizeRequestEntity(c *fiber.Ctx) *oauth2.AuthorizeRequest {
	uri := c.Context().URI()
	return &oauth2.AuthorizeRequest{
		ServerID: monoflake.IDFromBase62(c.Query("server_id")).Int64(),
		HostName: string(uri.Host()),
	}
}

func FromHTTPRequestToOauth2CallbackRequestEntity(c *fiber.Ctx) *oauth2.CallbackRequest {
	uri := c.Context().URI()
	return &oauth2.CallbackRequest{
		HostName: string(uri.Host()),
		State:    c.Query("state"),
		Code:     c.Query("code"),
	}
}
