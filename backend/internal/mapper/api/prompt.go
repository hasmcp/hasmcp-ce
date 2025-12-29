package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	view "github.com/hasmcp/hasmcp-ce/backend/internal/data/view/api"
	"github.com/mustafaturan/monoflake"
)

func FromHTTPRequestToCreatePromptRequestEntity(c *fiber.Ctx) *entity.CreatePromptRequest {
	var payload view.CreatePromptRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}

	return &entity.CreatePromptRequest{
		Prompt: FromPromptViewToPromptEntity(payload.Prompt),
	}
}

func FromPromptViewToPromptEntity(p view.Prompt) entity.Prompt {
	return entity.Prompt{
		ID:          monoflake.IDFromBase62(p.ID).Int64(),
		Name:        p.Name,
		Description: p.Description,
		Arguments:   p.Arguments,
		Messages:    p.Messages,
	}
}

func FromCreatePromptResponseEntityToHTTPResponse(res *entity.CreatePromptResponse) []byte {
	p := res.Prompt
	payload, _ := json.Marshal(view.CreatePromptResponse{
		Prompt: FromPromptEntityToPromptView(p),
	})
	return payload
}

func FromHTTPRequestToListPromptsRequestEntity(c *fiber.Ctx) *entity.ListPromptsRequest {
	return &entity.ListPromptsRequest{}
}

func FromListPromptsResponseEntityToHTTPResponse(res *entity.ListPromptsResponse) []byte {
	payload, _ := json.Marshal(view.ListPromptsResponse{
		Prompts: FromPromptEntitiesToPromptViews(res.Prompts),
	})
	return payload
}

func FromHTTPRequestToGetPromptRequestEntity(c *fiber.Ctx) *entity.GetPromptRequest {
	id := monoflake.IDFromBase62(c.Params("id")).Int64()
	if id == 0 {
		return nil
	}
	return &entity.GetPromptRequest{ID: id}
}

func FromGetPromptResponseEntityToHTTPResponse(res *entity.GetPromptResponse) []byte {
	p := res.Prompt
	payload, _ := json.Marshal(view.GetPromptResponse{
		Prompt: FromPromptEntityToPromptView(p),
	})
	return payload
}

func FromPromptEntityToPromptView(p entity.Prompt) view.Prompt {
	return view.Prompt{
		ID:          monoflake.ID(p.ID).String(),
		CreatedAt:   FromTimeToRFC3339String(p.CreatedAt),
		UpdatedAt:   FromTimeToRFC3339String(p.UpdatedAt),
		Name:        p.Name,
		Description: p.Description,
		Arguments:   p.Arguments,
		Messages:    p.Messages,
	}
}

func FromPromptEntitiesToPromptViews(prompts []entity.Prompt) []view.Prompt {
	v := make([]view.Prompt, len(prompts))
	for i, p := range prompts {
		v[i] = FromPromptEntityToPromptView(p)
	}
	return v
}

func FromHTTPRequestToUpdatePromptRequestEntity(c *fiber.Ctx) *entity.UpdatePromptRequest {
	var payload view.UpdatePromptRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}

	data := payload.Prompt
	data.ID = c.Params("id")

	return &entity.UpdatePromptRequest{
		Prompt: FromPromptViewToPromptEntity(data),
	}
}

func FromHTTPRequestToDeletePromptRequestEntity(c *fiber.Ctx) *entity.DeletePromptRequest {
	id := monoflake.IDFromBase62(c.Params("id")).Int64()
	if id == 0 {
		return nil
	}
	return &entity.DeletePromptRequest{ID: id}
}
