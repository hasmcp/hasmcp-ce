package model

import (
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
)

func FromProviderModelsToProviderEntities(ps []model.Provider) []crud.Provider {
	providers := make([]crud.Provider, len(ps))
	for i, p := range ps {
		providers[i] = FromProviderModelToProviderEntity(p)
	}
	return providers
}

func FromProviderModelToProviderEntity(p model.Provider) crud.Provider {
	clientSecretEncrypted, _ := hex.DecodeString(p.Oauth2Config.ClientSecretEncrypted)
	clientSecretEncryptionNonce, _ := hex.DecodeString(p.Oauth2Config.ClientSecretEncryptionNonce)
	return crud.Provider{
		ID:             p.ID,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
		Version:        p.Version,
		ApiType:        crud.ApiType(p.ApiType),
		VisibilityType: crud.VisibilityType(p.VisibilityType),
		BaseURL:        p.BaseURL,
		DocumentURL:    p.DocumentURL,
		IconURL:        p.IconURL,
		SecretPrefix:   p.SecretPrefix,
		Name:           p.Name,
		Description:    p.Description,
		Tools:          FromProviderToolModelsToProviderToolEntities(p.Tools),
		Oauth2Config: crud.ProviderOauth2Config{
			ClientID:                    p.Oauth2Config.ClientID,
			ClientSecretEncrypted:       clientSecretEncrypted,
			ClientSecretEncryptionNonce: clientSecretEncryptionNonce,
			AuthURL:                     p.Oauth2Config.AuthURL,
			TokenURL:                    p.Oauth2Config.TokenURL,
		},
	}
}

func FromProviderToolModelsToProviderToolEntities(es []model.ProviderTool) []crud.ProviderTool {
	tools := make([]crud.ProviderTool, len(es))
	for i, e := range es {
		tools[i] = FromProviderToolModelToProviderToolEntity(e)
	}
	return tools
}

func FromProviderToolModelToProviderToolEntity(e model.ProviderTool) crud.ProviderTool {
	var headers []crud.ToolHeader
	if e.Headers != nil {
		_ = json.Unmarshal(e.Headers, &headers)
	}
	return crud.ProviderTool{
		ID:                  e.ID,
		ProviderID:          e.ProviderID,
		Method:              crud.MethodType(e.Method),
		Path:                e.Path,
		Name:                e.Name,
		Title:               e.Title,
		Description:         e.Description,
		PathArgsJSONSchema:  e.PathArgsJSONSchema,
		QueryArgsJSONSchema: e.QueryArgsJSONSchema,
		ReqBodyJSONSchema:   e.ReqBodyJSONSchema,
		ResBodyJSONSchema:   e.ResBodyJSONSchema,
		Headers:             headers,
		Oauth2Scopes:        strings.Split(e.Oauth2Scopes, ","),
	}
}

func FromVariableModelToVariableEntity(v model.Variable) crud.Variable {
	val, _ := hex.DecodeString(v.Value)
	nonce, _ := hex.DecodeString(v.Nonce)

	return crud.Variable{
		ID:        v.ID,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
		Type:      crud.VariableType(v.Type),
		Value:     val,
		Nonce:     nonce,
		Name:      v.Name,
	}
}

func FromResourceModelToReourceEntity(r model.Resource) crud.Resource {
	return crud.Resource{
		ID:          r.ID,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		Name:        r.Name,
		Description: r.Description,
		URI:         r.URI,
		MimeType:    r.MimeType,
		Size:        r.Size,
		Annotations: r.Annotations,
	}
}

func FromResourceModelsToResourceEntities(rs []model.Resource) []crud.Resource {
	resources := make([]crud.Resource, len(rs))
	for i, r := range rs {
		resources[i] = FromResourceModelToReourceEntity(r)
	}
	return resources
}

func FromPromptModelToPromptEntity(p model.Prompt) crud.Prompt {
	return crud.Prompt{
		ID:          p.ID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		Name:        p.Name,
		Description: p.Description,
		Arguments:   p.Arguments,
		Messages:    p.Messages,
	}
}

func FromPromptModelsToPromptEntities(ps []model.Prompt) []crud.Prompt {
	prompts := make([]crud.Prompt, len(ps))
	for i, p := range ps {
		prompts[i] = FromPromptModelToPromptEntity(p)
	}
	return prompts
}

func FromServerModelsToServerEntities(servers []model.Server) []crud.Server {
	s := make([]crud.Server, len(servers))
	for i, m := range servers {
		s[i] = FromServerModelToServerEntity(m)
	}
	return s
}

func FromServerModelToServerEntity(s model.Server) crud.Server {
	providers := make([]crud.Provider, 0, len(s.Tools))
	providerIndex := map[int64]int{}
	for _, e := range s.Tools {
		index, ok := providerIndex[e.ProviderID]
		if !ok {
			index = len(providers)
			providerIndex[e.ProviderID] = index
			providers = append(providers, crud.Provider{
				ID:    e.ProviderID,
				Tools: make([]crud.ProviderTool, 0, 1),
			})
		}
		providers[index].Tools = append(providers[index].Tools, crud.ProviderTool{
			ID: e.ToolID,
		})
	}

	resources := make([]crud.Resource, len(s.Resources))
	for i, p := range s.Resources {
		resources[i] = crud.Resource{
			ID: p.ResourceID,
		}
	}

	prompts := make([]crud.Prompt, len(s.Prompts))
	for i, p := range s.Prompts {
		prompts[i] = crud.Prompt{
			ID: p.PromptID,
		}
	}

	return crud.Server{
		ID:                         s.ID,
		CreatedAt:                  s.CreatedAt,
		UpdatedAt:                  s.UpdatedAt,
		RequestHeadersProxyEnabled: s.RequestHeadersProxyEnabled,
		Name:                       s.Name,
		Instructions:               s.Instructions,
		Version:                    s.Version,
		Providers:                  providers,
		Resources:                  resources,
		Prompts:                    prompts,
	}
}
