package api

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	view "github.com/hasmcp/hasmcp-ce/backend/internal/data/view/api"
	"github.com/mustafaturan/monoflake"
)

func FromHTTPRequestToCreateProviderRequestEntity(c *fiber.Ctx) *entity.CreateProviderRequest {
	var payload view.CreateProviderRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}

	return &entity.CreateProviderRequest{
		Provider: FromProviderViewToProviderEntity(payload.Provider),
	}
}

func FromProviderViewToProviderEntity(p view.Provider) entity.Provider {
	var oauth2Config entity.ProviderOauth2Config
	if p.Oauth2Config != nil {
		oauth2Config = entity.ProviderOauth2Config{
			ClientID:     p.Oauth2Config.ClientID,
			ClientSecret: p.Oauth2Config.ClientSecret,
			AuthURL:      p.Oauth2Config.AuthURL,
			TokenURL:     p.Oauth2Config.TokenURL,
		}
	}
	return entity.Provider{
		ID:             monoflake.IDFromBase62(p.ID).Int64(),
		ApiType:        entity.StringToApiType(p.ApiType),
		VisibilityType: entity.StringToVisibilityType(p.VisibilityType),
		BaseURL:        p.BaseURL,
		DocumentURL:    p.DocumentURL,
		IconURL:        p.IconURL,
		Name:           p.Name,
		Description:    p.Description,
		Oauth2Config:   oauth2Config,
	}
}

func FromCreateProviderResponseEntityToHTTPResponse(rs *entity.CreateProviderResponse) []byte {
	payload, _ := json.Marshal(view.CreateProviderResponse{
		Provider: FromProviderEntityToProviderView(rs.Provider),
	})
	return payload
}

func FromHTTPRequestToGetProviderRequestEntity(c *fiber.Ctx) *entity.GetProviderRequest {
	id := monoflake.IDFromBase62(c.Params("id")).Int64()

	return &entity.GetProviderRequest{
		ID: id,
	}
}

func FromGetProviderResponseEntityToHTTPResponse(rs *entity.GetProviderResponse) []byte {
	response := view.GetProviderResponse{
		Provider: FromProviderEntityToProviderView(rs.Provider),
	}

	payload, _ := json.Marshal(response)

	return payload
}

func FromUpdateProviderResponseEntityToHTTPResponse(rs *entity.UpdateProviderResponse) []byte {
	response := view.UpdateProviderResponse{
		Provider: FromProviderEntityToProviderView(rs.Provider),
	}

	payload, _ := json.Marshal(response)

	return payload
}

func FromHTTPRequestToListProvidersRequestEntity(c *fiber.Ctx) *entity.ListProvidersRequest {
	limit, _ := strconv.ParseInt(c.Query("limit"), 10, 32)
	token, _ := strconv.ParseInt(c.Query("token"), 10, 64)

	return &entity.ListProvidersRequest{
		Filters: entity.ProviderFilters{
			NameContains:    c.Query("name_contains"),
			BaseURLContains: c.Query("base_url_contains"),
			ApiType:         entity.StringToApiType(c.Query("api_type")),
			VisibilityType:  entity.StringToVisibilityType(c.Query("visibility_type")),
		},
		Pagination: entity.Pagination{
			Limit: int(limit),
			Token: token,
		},
	}
}

func FromListProvidersResponseEntityToHTTPResponse(rs *entity.ListProvidersResponse) []byte {
	payload := view.ListProvidersResponse{
		Providers: FromProviderEntitiesToProviderViews(rs.Providers),
	}
	res, _ := json.Marshal(payload)
	return res
}

func FromHTTPRequestToUpdateProviderRequestEntity(c *fiber.Ctx) *entity.UpdateProviderRequest {
	var payload view.UpdateProviderRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}

	data := payload.Provider
	data.ID = c.Params("id")
	provider := FromProviderViewToProviderEntity(data)

	return &entity.UpdateProviderRequest{
		Provider: provider,
	}
}

func FromHTTPRequestToDeleteProviderRequestEntity(c *fiber.Ctx) *entity.DeleteProviderRequest {
	id := monoflake.IDFromBase62(c.Params("id")).Int64()

	return &entity.DeleteProviderRequest{
		ID: id,
	}
}

func FromProviderEntitiesToProviderViews(providers []entity.Provider) []view.Provider {
	v := make([]view.Provider, len(providers))
	for i, p := range providers {
		v[i] = FromProviderEntityToProviderView(p)
	}
	return v
}

func FromProviderEntityToProviderView(p entity.Provider) view.Provider {
	tools := make([]view.ProviderTool, len(p.Tools))
	for i, v := range p.Tools {
		tools[i] = FromProviderToolEntityToProviderToolView(v)
	}

	var oauth2Config *view.ProviderOauth2Config
	if p.Oauth2Config.ClientID != "" && p.Oauth2Config.AuthURL != "" && p.Oauth2Config.TokenURL != "" {
		oauth2Config = &view.ProviderOauth2Config{
			ClientID:     p.Oauth2Config.ClientID,
			ClientSecret: "***",
			AuthURL:      p.Oauth2Config.AuthURL,
			TokenURL:     p.Oauth2Config.TokenURL,
		}
	}

	return view.Provider{
		ID:             monoflake.ID(p.ID).String(),
		CreatedAt:      FromTimeToRFC3339String(p.CreatedAt),
		UpdatedAt:      FromTimeToRFC3339String(p.UpdatedAt),
		Version:        p.Version,
		ApiType:        p.ApiType.String(),
		VisibilityType: p.VisibilityType.String(),
		BaseURL:        p.BaseURL,
		DocumentURL:    p.DocumentURL,
		IconURL:        p.IconURL,
		SecretPrefix:   p.SecretPrefix,
		Name:           p.Name,
		Description:    p.Description,
		Tools:          tools,
		Oauth2Config:   oauth2Config,
	}
}
