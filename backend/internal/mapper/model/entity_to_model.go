package model

import (
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
)

func FromServerEntityServerModel(s crud.Server) model.Server {
	tools := make([]model.ServerTool, 0, len(s.Providers))
	for _, p := range s.Providers {
		for _, e := range p.Tools {
			tools = append(tools, model.ServerTool{
				ToolID:     e.ID,
				ProviderID: p.ID,
				ServerID:   s.ID,
			})
		}
	}

	resources := make([]model.ServerResource, 0, len(s.Resources))
	for _, r := range s.Resources {
		resources = append(resources, model.ServerResource{
			ResourceID: r.ID,
			ServerID:   s.ID,
		})
	}

	prompts := make([]model.ServerPrompt, 0, len(s.Prompts))
	for _, p := range s.Prompts {
		prompts = append(prompts, model.ServerPrompt{
			PromptID: p.ID,
			ServerID: s.ID,
		})
	}

	return model.Server{
		ID:                         s.ID,
		CreatedAt:                  s.CreatedAt,
		UpdatedAt:                  s.UpdatedAt,
		RequestHeadersProxyEnabled: s.RequestHeadersProxyEnabled,
		Name:                       s.Name,
		Instructions:               s.Instructions,
		Version:                    s.Version,
		Tools:                      tools,
		Resources:                  resources,
		Prompts:                    prompts,
	}
}
