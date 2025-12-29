package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	view "github.com/hasmcp/hasmcp-ce/backend/internal/data/view/api"
	"github.com/mustafaturan/monoflake"
)

// SchemaProperty defines the schema for a single property (e.g., "userId").
// We use json tags to control the output field names.
type SchemaProperty struct {
	Description string `json:"description"`
	Type        string `json:"type"`
}

// JSONSchema defines the structure of the entire JSON Schema document.
type JSONSchema struct {
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	Type        string                    `json:"type"`
	Properties  map[string]SchemaProperty `json:"properties"`
	Required    []string                  `json:"required"`
}

func FromHTTPRequestToCreateProviderToolRequestEntity(c *fiber.Ctx) *entity.CreateProviderToolRequest {
	var payload view.CreateProviderToolRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}

	data := payload.Tool
	data.ProviderID = c.Params("id")
	tool := FromProviderToolViewToProviderToolEntity(data)

	return &entity.CreateProviderToolRequest{
		Tool: tool,
	}
}

func FromProviderToolViewToProviderToolEntity(e view.ProviderTool) entity.ProviderTool {
	headers := make([]entity.ToolHeader, len(e.Headers))
	for i, h := range e.Headers {
		headers[i] = entity.ToolHeader{
			Key:   h.Key,
			Value: h.Value,
		}
	}

	return entity.ProviderTool{
		ID:                  monoflake.IDFromBase62(e.ID).Int64(),
		ProviderID:          monoflake.IDFromBase62(e.ProviderID).Int64(),
		Method:              entity.StringToMethodType(e.Method),
		Path:                e.Path,
		Name:                e.Name,
		Title:               e.Title,
		Description:         e.Description,
		PathArgsJSONSchema:  e.PathArgsJSONSchema,
		QueryArgsJSONSchema: e.QueryArgsJSONSchema,
		ReqBodyJSONSchema:   e.ReqBodyJSONSchema,
		ResBodyJSONSchema:   e.ResBodyJSONSchema,
		Headers:             headers,
		Oauth2Scopes:        e.Oauth2Scopes,
	}
}

func FromCreateProviderToolResponseEntityToHTTPResponse(rs *entity.CreateProviderToolResponse) []byte {
	resp := view.CreateProviderToolResponse{
		Tool: FromProviderToolEntityToProviderToolView(rs.Tool),
	}

	payload, _ := json.Marshal(resp)

	return payload
}

func FromProviderToolEntitiesToProviderToolViews(es []entity.ProviderTool) []view.ProviderTool {
	tools := make([]view.ProviderTool, len(es))
	for i, e := range es {
		tools[i] = FromProviderToolEntityToProviderToolView(e)
	}
	return tools
}

func FromHTTPRequestToListProviderToolsRequestEntity(c *fiber.Ctx) *entity.ListProviderToolsRequest {
	providerIDParam := c.Params("id")
	if providerIDParam == "" {
		return nil
	}

	return &entity.ListProviderToolsRequest{
		ProviderID: monoflake.IDFromBase62(providerIDParam).Int64(),
	}
}

func FromListProviderToolsResponseEntityToHTTPResponse(rs *entity.ListProviderToolsResponse) []byte {
	resp := &view.ListProviderToolsResponse{
		Tools: FromProviderToolEntitiesToProviderToolViews(rs.Tools),
	}

	payload, _ := json.Marshal(resp)

	return payload
}

func FromHTTPRequestToGetProviderToolRequestEntity(c *fiber.Ctx) *entity.GetProviderToolRequest {
	providerIDParam := c.Params("id")
	toolIDParam := c.Params("toolID")
	if providerIDParam == "" || toolIDParam == "" {
		return nil
	}

	return &entity.GetProviderToolRequest{
		ProviderID: monoflake.IDFromBase62(providerIDParam).Int64(),
		ToolID:     monoflake.IDFromBase62(toolIDParam).Int64(),
	}
}

func FromGetProviderToolResponseEntityToHTTPResponse(rs *entity.GetProviderToolResponse) []byte {
	payload, _ := json.Marshal(view.GetProviderToolResponse{
		Tool: FromProviderToolEntityToProviderToolView(rs.Tool),
	})

	return payload
}

func FromProviderToolEntityToProviderToolView(e entity.ProviderTool) view.ProviderTool {
	headers := make([]view.ToolHeader, len(e.Headers))
	for j, h := range e.Headers {
		headers[j] = view.ToolHeader{
			Key:   h.Key,
			Value: h.Value,
		}
	}

	return view.ProviderTool{
		ID:                  monoflake.ID(e.ID).String(),
		ProviderID:          monoflake.ID(e.ProviderID).String(),
		Method:              e.Method.String(),
		Path:                e.Path,
		Name:                e.Name,
		Title:               e.Title,
		Description:         e.Description,
		PathArgsJSONSchema:  e.PathArgsJSONSchema,
		QueryArgsJSONSchema: e.QueryArgsJSONSchema,
		ReqBodyJSONSchema:   e.ReqBodyJSONSchema,
		ResBodyJSONSchema:   e.ResBodyJSONSchema,
		Headers:             headers,
		Oauth2Scopes:        e.Oauth2Scopes,
	}
}

func FromUpdateProviderToolResponseEntityToHTTPResponse(rs *entity.UpdateProviderToolResponse) []byte {
	payload, _ := json.Marshal(view.UpdateProviderToolResponse{
		Tool: FromProviderToolEntityToProviderToolView(rs.Tool),
	})

	return payload
}

func FromHTTPRequestToUpdateProviderToolRequestEntity(c *fiber.Ctx) *entity.UpdateProviderToolRequest {
	providerIDParam := c.Params("id")
	providerToolIDParam := c.Params("toolID")
	var payload view.UpdateProviderToolRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}

	data := payload.Tool
	data.ID = providerToolIDParam
	data.ProviderID = providerIDParam

	return &entity.UpdateProviderToolRequest{
		Tool: FromProviderToolViewToProviderToolEntity(data),
	}
}

func FromHTTPRequestToDeleteProviderToolRequestEntity(c *fiber.Ctx) *entity.DeleteProviderToolRequest {
	providerIDParam := c.Params("id")
	toolIDParam := c.Params("toolID")
	if providerIDParam == "" || toolIDParam == "" {
		return nil
	}

	return &entity.DeleteProviderToolRequest{
		ProviderID: monoflake.IDFromBase62(providerIDParam).Int64(),
		ToolID:     monoflake.IDFromBase62(toolIDParam).Int64(),
	}
}
