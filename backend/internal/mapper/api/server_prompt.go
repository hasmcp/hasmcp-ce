package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	view "github.com/hasmcp/hasmcp-ce/backend/internal/data/view/api"
	"github.com/mustafaturan/monoflake"
)

func FromHTTPRequestToCreateServerPromptAssociationRequestEntity(c *fiber.Ctx) *entity.CreateServerPromptRequest {
	var payload view.CreateServerPromptRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}
	data := payload.Prompt
	data.ServerID = c.Params("id")

	return &entity.CreateServerPromptRequest{
		Prompt: FromServerPromptViewToServerPromptEntity(data),
	}
}

func FromServerPromptViewToServerPromptEntity(p view.ServerPrompt) entity.ServerPrompt {
	return entity.ServerPrompt{
		ServerID: monoflake.IDFromBase62(p.ServerID).Int64(),
		PromptID: monoflake.IDFromBase62(p.PromptID).Int64(),
	}
}

func FromServerPromptEntityToServerPromptView(p entity.ServerPrompt) view.ServerPrompt {
	return view.ServerPrompt{
		ServerID: monoflake.ID(p.ServerID).String(),
		PromptID: monoflake.ID(p.PromptID).String(),
	}
}

func FromServerPromptEntitiesToServerPromptViews(ps []entity.ServerPrompt) []view.ServerPrompt {
	prompts := make([]view.ServerPrompt, len(ps))
	for i, p := range ps {
		prompts[i] = FromServerPromptEntityToServerPromptView(p)
	}
	return prompts
}

func FromCreateServerPromptAssociationResponseEntityToHTTPResponse(res *entity.CreateServerPromptResponse) []byte {
	resp := view.CreateServerPromptResponse{
		Prompt: FromServerPromptEntityToServerPromptView(res.Prompt),
	}
	payload, _ := json.Marshal(resp)
	return payload
}

func FromHTTPRequestToListServerPromptsRequestEntity(c *fiber.Ctx) *entity.ListServerPromptsRequest {
	serverIDParam := c.Params("id")
	if serverIDParam == "" {
		return nil
	}
	return &entity.ListServerPromptsRequest{
		ServerID: monoflake.IDFromBase62(serverIDParam).Int64(),
	}
}

func FromListServerPromptsResponseEntityToHTTPResponse(res *entity.ListServerPromptsResponse) []byte {

	payload, _ := json.Marshal(view.ListServerPromptsResponse{
		Prompts: FromServerPromptEntitiesToServerPromptViews(res.Prompts),
	})
	return payload
}

func FromHTTPRequestToDeleteServerPromptAssociationRequestEntity(c *fiber.Ctx) *entity.DeleteServerPromptRequest {
	serverIDParam := c.Params("id")
	promptIDParam := c.Params("promptID")
	if serverIDParam == "" || promptIDParam == "" {
		return nil
	}

	return &entity.DeleteServerPromptRequest{
		ServerID: monoflake.IDFromBase62(serverIDParam).Int64(),
		PromptID: monoflake.IDFromBase62(promptIDParam).Int64(),
	}
}
