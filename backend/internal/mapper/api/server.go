package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	view "github.com/hasmcp/hasmcp-ce/backend/internal/data/view/api"
	"github.com/mustafaturan/monoflake"
)

func FromHTTPRequestToCreateServerRequestEntity(c *fiber.Ctx) *entity.CreateServerRequest {
	var payload view.CreateServerRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}

	return &entity.CreateServerRequest{
		Server: FromServerViewToServerEntity(payload.Server),
	}
}

func FromServerViewToServerEntity(s view.Server) entity.Server {
	providers := make([]entity.Provider, len(s.Providers))
	for i, p := range s.Providers {
		tools := make([]entity.ProviderTool, len(p.Tools))
		for j, e := range p.Tools {
			tools[j] = entity.ProviderTool{
				ID: monoflake.IDFromBase62(e.ID).Int64(),
			}
		}
		providers[i] = entity.Provider{
			ID:    monoflake.IDFromBase62(p.ID).Int64(),
			Tools: tools,
		}
	}

	resources := make([]entity.Resource, len(s.Resources))
	for i, r := range s.Resources {
		resources[i] = entity.Resource{
			ID: monoflake.IDFromBase62(r.ID).Int64(),
		}
	}

	prompts := make([]entity.Prompt, len(s.Prompts))
	for i, p := range s.Prompts {
		prompts[i] = entity.Prompt{
			ID: monoflake.IDFromBase62(p.ID).Int64(),
		}
	}

	return entity.Server{
		ID:                         monoflake.IDFromBase62(s.ID).Int64(),
		RequestHeadersProxyEnabled: s.RequestHeadersProxyEnabled,
		Name:                       s.Name,
		Instructions:               s.Instructions,
		Version:                    s.Version,
		Providers:                  providers,
		Resources:                  resources,
		Prompts:                    prompts,
	}
}

func FromCreateServerResponseEntityToHTTPResponse(res *entity.CreateServerResponse) []byte {
	payload, _ := json.Marshal(view.CreateServerResponse{
		Server: FromServerEntityToServerView(res.Server),
	})
	return payload
}

func FromServerEntitiesToServerViews(ss []entity.Server) []view.Server {
	servers := make([]view.Server, len(ss))
	for i, s := range ss {
		servers[i] = FromServerEntityToServerView(s)
	}
	return servers
}

func FromServerEntityToServerView(s entity.Server) view.Server {
	return view.Server{
		ID:                         monoflake.ID(s.ID).String(),
		CreatedAt:                  FromTimeToRFC3339String(s.CreatedAt),
		UpdatedAt:                  FromTimeToRFC3339String(s.UpdatedAt),
		RequestHeadersProxyEnabled: s.RequestHeadersProxyEnabled,
		Name:                       s.Name,
		Instructions:               s.Instructions,
		Version:                    s.Version,
		Providers:                  FromProviderEntitiesToProviderViews(s.Providers),
		Resources:                  FromResourceEntitiesToResourceViews(s.Resources),
		Prompts:                    FromPromptEntitiesToPromptViews(s.Prompts),
	}
}

func FromHTTPRequestToListServersRequestEntity(c *fiber.Ctx) *entity.ListServersRequest {
	return &entity.ListServersRequest{}
}

func FromListServersResponseEntityToHTTPResponse(res *entity.ListServersResponse) []byte {
	payload := view.ListServersResponse{
		Servers: FromServerEntitiesToServerViews(res.Servers),
	}

	data, _ := json.Marshal(payload)
	return data
}

func FromHTTPRequestToUpdateServerRequestEntity(c *fiber.Ctx) *entity.UpdateServerRequest {
	var payload view.UpdateServerRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}

	data := payload.Server
	data.ID = c.Params("id")

	return &entity.UpdateServerRequest{
		Server: FromServerViewToServerEntity(data),
	}
}

func FromUpdateServerResponseEntityToHTTPResponse(res *entity.UpdateServerResponse) []byte {
	payload, _ := json.Marshal(view.UpdateServerResponse{
		Server: FromServerEntityToServerView(res.Server),
	})
	return payload
}

func FromHTTPRequestToDeleteServerRequestEntity(c *fiber.Ctx) *entity.DeleteServerRequest {
	serverIDParam := c.Params("id")
	if serverIDParam == "" {
		return nil
	}

	return &entity.DeleteServerRequest{
		ID: monoflake.IDFromBase62(serverIDParam).Int64(),
	}
}

func FromHTTPRequestToGetServerRequestEntity(c *fiber.Ctx) *entity.GetServerRequest {
	serverIDParam := c.Params("id")
	if serverIDParam == "" {
		return nil
	}

	return &entity.GetServerRequest{
		ID: monoflake.IDFromBase62(serverIDParam).Int64(),
	}
}

func FromGetServerResponseEntityToHTTPResponse(res *entity.GetServerResponse) []byte {
	payload, _ := json.Marshal(view.GetServerResponse{
		Server: FromServerEntityToServerView(res.Server),
	})
	return payload
}
