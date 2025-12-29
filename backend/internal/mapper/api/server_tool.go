package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	view "github.com/hasmcp/hasmcp-ce/backend/internal/data/view/api"
	"github.com/mustafaturan/monoflake"
)

func FromHTTPRequestToCreateServerToolRequestEntity(c *fiber.Ctx) *entity.CreateServerToolRequest {
	var payload view.CreateServerToolRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}
	data := payload.Tool
	data.ServerID = c.Params("id")

	return &entity.CreateServerToolRequest{
		Tool: FromServerToolViewToServerToolEntity(data),
	}
}

func FromServerToolViewToServerToolEntity(e view.ServerTool) entity.ServerTool {
	return entity.ServerTool{
		ServerID:   monoflake.IDFromBase62(e.ServerID).Int64(),
		ProviderID: monoflake.IDFromBase62(e.ProviderID).Int64(),
		ToolID:     monoflake.IDFromBase62(e.ToolID).Int64(),
	}
}

func FromServerToolEntityToServerToolView(e entity.ServerTool) view.ServerTool {
	return view.ServerTool{
		ServerID:   monoflake.ID(e.ServerID).String(),
		ProviderID: monoflake.ID(e.ProviderID).String(),
		ToolID:     monoflake.ID(e.ToolID).String(),
	}
}

func FromServerToolEntitiesToServerToolViews(es []entity.ServerTool) []view.ServerTool {
	tools := make([]view.ServerTool, len(es))
	for i, e := range es {
		tools[i] = FromServerToolEntityToServerToolView(e)
	}
	return tools
}

func FromCreateServerToolResponseEntityToHTTPResponse(rs *entity.CreateServerToolResponse) []byte {
	resp := view.CreateServerToolResponse{
		Tool: FromServerToolEntityToServerToolView(rs.Tool),
	}
	payload, _ := json.Marshal(resp)
	return payload
}

func FromHTTPRequestToListServerToolsRequestEntity(c *fiber.Ctx) *entity.ListServerToolsRequest {
	serverIDParam := c.Params("id")
	if serverIDParam == "" {
		return nil
	}

	return &entity.ListServerToolsRequest{
		ServerID: monoflake.IDFromBase62(serverIDParam).Int64(),
	}
}

func FromListServerToolsResponseEntityToHTTPResponse(rs *entity.ListServerToolsResponse) []byte {
	data := view.ListServerToolsResponse{
		Tools: FromServerToolEntitiesToServerToolViews(rs.Tools),
	}

	payload, _ := json.Marshal(data)
	return payload
}

func FromHTTPRequestToDeleteServerToolRequestEntity(c *fiber.Ctx) *entity.DeleteServerToolRequest {
	serverIDParam := c.Params("id")
	if serverIDParam == "" {
		return nil
	}
	toolID := c.Params("toolID")

	return &entity.DeleteServerToolRequest{
		Tool: entity.ServerTool{
			ServerID: monoflake.IDFromBase62(serverIDParam).Int64(),
			ToolID:   monoflake.IDFromBase62(toolID).Int64(),
		},
	}
}
