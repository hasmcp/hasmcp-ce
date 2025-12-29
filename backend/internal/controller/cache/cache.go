package cache

import (
	"context"
	"encoding/hex"
	"sync"

	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	modelmapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/model"
	"github.com/hasmcp/hasmcp-ce/backend/internal/repository/storage"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/locksmith"
	zlog "github.com/rs/zerolog/log"
)

type (
	Controller interface {
		GetTool(ctx context.Context, id int64) (*entity.ProviderTool, error)
		GetServer(ctx context.Context, id int64) (*entity.Server, error)
		GetPrompt(ctx context.Context, id int64) (*entity.Prompt, error)
		GetProvider(ctx context.Context, id int64) (*entity.Provider, error)
		GetResource(ctx context.Context, id int64) (*entity.Resource, error)
		GetVariable(ctx context.Context, name string) (string, error)
		Evict(ctx context.Context, objectType entity.ObjectType, id int64)
		ReloadTool(ctx context.Context, id int64) (*entity.ProviderTool, error)
		ReloadServer(ctx context.Context, id int64) (*entity.Server, error)
		ReloadPrompt(ctx context.Context, id int64) (*entity.Prompt, error)
		ReloadProvider(ctx context.Context, id int64) (*entity.Provider, error)
		ReloadResource(ctx context.Context, id int64) (*entity.Resource, error)
		ReloadVariable(ctx context.Context, name string) (string, error)
	}

	controller struct {
		locksmith locksmith.Service
		storage   storage.Repository

		variableRefs *sync.Map
		variables    *sync.Map
		tools        *sync.Map
		providers    *sync.Map
		resources    *sync.Map
		prompts      *sync.Map
		servers      *sync.Map
	}

	Params struct {
		Locksmith locksmith.Service
		Storage   storage.Repository
	}

	err string
)

const (
	ErrNotFound err = "not found"
)

func New(p Params) (Controller, error) {
	storage := p.Storage
	ls := p.Locksmith

	providers := sync.Map{}
	tools := sync.Map{}
	prompts := sync.Map{}
	resources := sync.Map{}
	variables := sync.Map{}
	variableRefs := sync.Map{}
	servers := sync.Map{}

	variableModels, err := storage.ListVariables(context.Background())
	if err != nil {
		return nil, err
	}

	for _, v := range variableModels {
		variable := modelmapper.FromVariableModelToVariableEntity(v)
		val := variable.Value
		if variable.Type == entity.VariableTypeSecret {
			res, err := ls.Decrypt(context.Background(), &locksmith.DecryptRequest{
				Ciphertext: variable.Value,
				Nonce:      variable.Nonce,
			})
			if err != nil {
				return nil, err
			}
			val = res.Plaintext
		}
		variables.Store(v.Name, string(val))
		variableRefs.Store(variable.ID, v.Name)
	}

	return &controller{
		locksmith: ls,
		storage:   storage,

		variableRefs: &variableRefs,
		variables:    &variables,
		tools:        &tools,
		servers:      &servers,
		providers:    &providers,
		prompts:      &prompts,
		resources:    &resources,
	}, nil
}

func (c *controller) Evict(ctx context.Context, objectType entity.ObjectType, id int64) {
	switch objectType {
	case entity.ObjectTypeVariable:
		if v, ok := c.variableRefs.Load(id); ok {
			c.variableRefs.Delete(id)
			if name, ok := v.(string); ok {
				c.variables.Delete(name)
			}
		}
	case entity.ObjectTypeProviderTool:
		c.tools.Delete(id)
	case entity.ObjectTypeProvider:
		c.providers.Delete(id)
	case entity.ObjectTypeServer:
		c.servers.Delete(id)
	case entity.ObjectTypePrompt:
		c.prompts.Delete(id)
	case entity.ObjectTypeResource:
		c.resources.Delete(id)
	}
}

func (c *controller) GetServer(ctx context.Context, id int64) (*entity.Server, error) {
	v, ok := c.servers.Load(id)
	if !ok {
		return c.ReloadServer(ctx, id)
	}

	s, _ := v.(*entity.Server)

	return s, nil
}

func (c *controller) ReloadServer(ctx context.Context, id int64) (*entity.Server, error) {
	s, err := c.storage.GetServer(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	var ok bool
	server := modelmapper.FromServerModelToServerEntity(*s)
	for i := 0; i < len(server.Providers); i++ {
		p, err := c.GetProvider(ctx, server.Providers[i].ID)
		if err != nil {
			return nil, err
		}
		toolSet := make(map[int64]struct{}, len(server.Providers[i].Tools))
		for _, e := range server.Providers[i].Tools {
			toolSet[e.ID] = struct{}{}
		}
		tools := make([]entity.ProviderTool, 0, len(toolSet))
		for _, e := range p.Tools {
			if _, ok = toolSet[e.ID]; ok {
				tools = append(tools, e)
			}
		}

		// copy provider with desired tools
		server.Providers[i] = entity.Provider{
			ID:        p.ID,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,

			Version:        p.Version,
			ApiType:        p.ApiType,
			VisibilityType: p.VisibilityType,
			BaseURL:        p.BaseURL,
			DocumentURL:    p.DocumentURL,
			IconURL:        p.IconURL,
			SecretPrefix:   p.SecretPrefix,
			Name:           p.Name,
			Description:    p.Description,
			Oauth2Config:   p.Oauth2Config,

			Tools: tools,
		}
	}

	for i := 0; i < len(server.Prompts); i++ {
		p, err := c.GetPrompt(ctx, server.Prompts[i].ID)
		if err != nil {
			return nil, err
		}
		server.Prompts[i] = *p
	}

	for i := 0; i < len(server.Resources); i++ {
		r, err := c.GetResource(ctx, server.Resources[i].ID)
		if err != nil {
			return nil, err
		}
		server.Resources[i] = *r
	}

	c.servers.Store(server.ID, &server)
	return &server, nil
}

func (c *controller) GetTool(ctx context.Context, id int64) (*entity.ProviderTool, error) {
	v, ok := c.tools.Load(id)
	if !ok {
		return c.ReloadTool(ctx, id)
	}

	e, _ := v.(*entity.ProviderTool)

	return e, nil
}

func (c *controller) ReloadTool(ctx context.Context, id int64) (*entity.ProviderTool, error) {
	e, err := c.storage.GetProviderTool(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	tool := modelmapper.FromProviderToolModelToProviderToolEntity(*e)
	c.tools.Store(e.ID, &tool)
	return &tool, nil
}

func (c *controller) GetPrompt(ctx context.Context, id int64) (*entity.Prompt, error) {
	v, ok := c.prompts.Load(id)
	if !ok {
		return c.ReloadPrompt(ctx, id)
	}

	p := v.(*entity.Prompt)

	return p, nil
}

func (c *controller) ReloadPrompt(ctx context.Context, id int64) (*entity.Prompt, error) {
	p, err := c.storage.GetPrompt(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	prompt := modelmapper.FromPromptModelToPromptEntity(*p)
	c.prompts.Store(p.ID, &prompt)
	return &prompt, nil
}

func (c *controller) GetProvider(ctx context.Context, id int64) (*entity.Provider, error) {
	v, ok := c.providers.Load(id)
	if !ok {
		return c.ReloadProvider(ctx, id)
	}

	p := v.(*entity.Provider)

	return p, nil
}

func (c *controller) ReloadProvider(ctx context.Context, id int64) (*entity.Provider, error) {
	p, err := c.storage.GetProvider(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	provider := modelmapper.FromProviderModelToProviderEntity(*p)
	c.providers.Store(p.ID, &provider)
	return &provider, nil
}

func (c *controller) GetResource(ctx context.Context, id int64) (*entity.Resource, error) {
	v, ok := c.providers.Load(id)
	if !ok {
		r, err := c.storage.GetResource(ctx, id)
		if err != nil {
			return nil, ErrNotFound
		}
		resource := modelmapper.FromResourceModelToReourceEntity(*r)
		c.resources.Store(r.ID, &resource)
		return &resource, nil
	}

	r := v.(*entity.Resource)

	return r, nil
}

func (c *controller) ReloadResource(ctx context.Context, id int64) (*entity.Resource, error) {
	r, err := c.storage.GetResource(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}
	resource := modelmapper.FromResourceModelToReourceEntity(*r)
	c.resources.Store(r.ID, &resource)
	return &resource, nil
}

func (c *controller) GetVariable(ctx context.Context, name string) (string, error) {
	v, ok := c.variables.Load(name)
	if !ok {
		return c.ReloadVariable(ctx, name)
	}

	return v.(string), nil
}

func (c *controller) ReloadVariable(ctx context.Context, name string) (string, error) {
	v, err := c.storage.GetVariableByName(ctx, name)
	if err != nil {
		return "", err
	}
	c.variableRefs.Store(v.ID, v.Name)

	if entity.VariableType(v.Type) == entity.VariableTypeEnv {
		c.variables.Store(v.Name, v.Value)
		return v.Value, nil
	}

	val, err := hex.DecodeString(v.Value)
	if err != nil {
		zlog.Error().Str("name", v.Name).Err(err).Msg("failed to decrypt the secret")
		return "", err
	}
	nonce, err := hex.DecodeString(v.Nonce)
	if err != nil {
		zlog.Error().Str("name", v.Name).Err(err).Msg("failed to decrypt the secret")
		return "", err
	}

	res, err := c.locksmith.Decrypt(ctx, &locksmith.DecryptRequest{
		Ciphertext: val,
		Nonce:      nonce,
	})
	if err != nil {
		zlog.Error().Str("name", v.Name).Err(err).Msg("failed to decrypt the secret")
		return "", err
	}

	c.variables.Store(v.Name, string(res.Plaintext))
	return string(res.Plaintext), nil
}

func (e err) Error() string {
	return string(e)
}
