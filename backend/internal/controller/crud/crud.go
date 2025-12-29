package crud

import (
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/cache"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/jwt"
	"github.com/hasmcp/hasmcp-ce/backend/internal/repository/base"
	"github.com/hasmcp/hasmcp-ce/backend/internal/repository/storage"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/idgen"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/locksmith"
)

type (
	Params struct {
		Config    config.Service
		IDGen     idgen.Service
		Locksmith locksmith.Service

		Cache  cache.Controller
		Mcp    mcp.Controller
		McpJWT jwt.IssuerController

		Repository base.Repository
		Storage    storage.Repository
	}

	Controller interface {
		VariableController
		ProviderController
		ProviderToolController
		ServerController
		ServerTokenController
		ServerToolController
		PromptController
		ResourceController
		ServerPromptController
		ServerResourceController
	}

	controller struct {
		idgen     idgen.Service
		locksmith locksmith.Service

		cache  cache.Controller
		mcp    mcp.Controller
		mcpJWT jwt.IssuerController

		repository base.Repository
		storage    storage.Repository
	}

	crudConfig struct {
	}
)

const (
	_cfgKey = "crudctrl"
)

func New(p Params) (Controller, error) {
	var cfg crudConfig
	err := p.Config.Populate(_cfgKey, &cfg)
	if err != nil {
		return nil, err
	}

	c := &controller{
		idgen:     p.IDGen,
		locksmith: p.Locksmith,

		cache:  p.Cache,
		mcp:    p.Mcp,
		mcpJWT: p.McpJWT,

		repository: p.Repository,
		storage:    p.Storage,
	}
	return c, nil
}
