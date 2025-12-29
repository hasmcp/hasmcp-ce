package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/cache"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/crud"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp"
	mcpjwt "github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/jwt"
	oauth2mcp "github.com/hasmcp/hasmcp-ce/backend/internal/controller/oauth2mcp"
	oauth2mcpjwt "github.com/hasmcp/hasmcp-ce/backend/internal/controller/oauth2mcp/jwt"

	apihandler "github.com/hasmcp/hasmcp-ce/backend/internal/handler/api"
	mcphandler "github.com/hasmcp/hasmcp-ce/backend/internal/handler/mcp"
	oauth2mcphandler "github.com/hasmcp/hasmcp-ce/backend/internal/handler/oauth2mcp"
	statichandler "github.com/hasmcp/hasmcp-ce/backend/internal/handler/static"

	"github.com/hasmcp/hasmcp-ce/backend/internal/repository/base"
	"github.com/hasmcp/hasmcp-ce/backend/internal/repository/postgres"
	"github.com/hasmcp/hasmcp-ce/backend/internal/repository/sqlite"
	"github.com/hasmcp/hasmcp-ce/backend/internal/repository/storage"

	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/httpc"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/idgen"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/locksmith"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/log"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/memq"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/pubsub"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/server"
)

type (
	// App hosts the components that runs/stop the whole app with dependencies
	App struct {
		Controllers  controllers
		Handlers     handlers
		Repositories repositories
		Services     services
	}

	services struct {
		Config    config.Service
		Logger    log.Service
		IDGen     idgen.Service
		Locksmith locksmith.Service
		Memq      memq.Service
		PubSub    pubsub.Service
		Server    server.Service
	}

	handlers struct {
		StaticHandler    statichandler.Handler
		ApiHandler       apihandler.Handler
		McpHandler       mcphandler.Handler
		Oauth2McpHandler oauth2mcphandler.Handler
	}

	controllers struct {
		Cache     cache.Controller
		Crud      crud.Controller
		Oauth2Mcp oauth2mcp.Controller
	}

	repositories struct {
		Repository base.Repository
	}
)

// New inits a new app
func New() (*App, error) {
	// Initialize config provider
	config, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "config", err)
	}

	// Initialize logger
	logger, err := log.New(
		log.Params{
			Config: config,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "logger", err)
	}

	// HTTPC
	httpc, err := httpc.New(
		httpc.Params{
			Config: config,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "httpc", err)
	}

	// IDGen
	idgen, err := idgen.New(
		idgen.Params{
			Config: config,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "idgen", err)
	}

	// Locksmith
	locksmith, err := locksmith.New(
		locksmith.Params{
			Config: config,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "locksmith", err)
	}

	// Memq
	memq, err := memq.New(
		memq.Params{},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "memq", err)
	}

	// pubsub
	pubsub, err := pubsub.New(
		pubsub.Params{
			IDGen:  idgen,
			Config: config,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "pubsub", err)
	}

	// DB repository
	var db base.Repository

	// Repository
	postgres, err := postgres.New(postgres.Params{
		Config: config,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "postgres", err)
	}
	db = postgres

	if db == nil {
		sqlite, err := sqlite.New(sqlite.Params{
			Config: config,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", "sqlite", err)
		}
		db = sqlite
	}

	if db == nil {
		return nil, errors.New("either postgres or sqlite must be enabled to run the app")
	}

	// Storage
	storage, err := storage.New(storage.Params{
		DB: db,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "storage", err)
	}

	// Controllers
	cache, err := cache.New(cache.Params{
		Locksmith: locksmith,
		Storage:   storage,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "cache", err)
	}

	mcpJWT, err := mcpjwt.New(mcpjwt.Params{
		Config: config,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "mcpjwt", err)
	}

	mcp, err := mcp.New(mcp.Params{
		Config:    config,
		IDGen:     idgen,
		HTTPC:     httpc,
		Locksmith: locksmith,
		Memq:      memq,
		McpJWT:    mcpJWT,
		Cache:     cache,
		PubSub:    pubsub,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "mcp", err)
	}

	crud, err := crud.New(crud.Params{
		Config:     config,
		IDGen:      idgen,
		Locksmith:  locksmith,
		Cache:      cache,
		Repository: db,
		Storage:    storage,
		Mcp:        mcp,
		McpJWT:     mcpJWT,
	})
	if err != nil {
		return nil, err
	}

	oauth2JWT, err := oauth2mcpjwt.New(oauth2mcpjwt.Params{
		Config: config,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "oauth2jwt", err)
	}

	oauth2mcp, err := oauth2mcp.New(oauth2mcp.Params{
		Config:    config,
		Locksmith: locksmith,
		Crud:      crud,
		JWT:       oauth2JWT,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "oauth2mcp", err)
	}

	server, err := server.New(server.Params{
		Config: config,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "server", err)
	}

	// Handlers
	apihandler, err := apihandler.New(apihandler.Params{
		Config: config,
		Server: server,
		Crud:   crud,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "api handler", err)
	}

	mcphandler, err := mcphandler.New(mcphandler.Params{
		Config:  config,
		Server:  server,
		Mcp:     mcp,
		JWTAuth: mcpJWT,
	})
	if err != nil {
		return nil, err
	}

	oauth2mcphandler, err := oauth2mcphandler.New(oauth2mcphandler.Params{
		Config: config,
		Server: server,
		Oauth2: oauth2mcp,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "oauth2 handler", err)
	}

	statichandler, err := statichandler.New(statichandler.Params{
		Server: server,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "static handler", err)
	}

	return &App{
		// Core services
		Services: services{
			Config:    config,
			Logger:    logger,
			IDGen:     idgen,
			Locksmith: locksmith,
			Memq:      memq,
			PubSub:    pubsub,
			Server:    server,
		},

		// Repositorys
		Repositories: repositories{
			Repository: db,
		},

		// Controllers
		Controllers: controllers{
			Cache:     cache,
			Crud:      crud,
			Oauth2Mcp: oauth2mcp,
		},

		// Handlers
		Handlers: handlers{
			ApiHandler:       apihandler,
			McpHandler:       mcphandler,
			Oauth2McpHandler: oauth2mcphandler,
			StaticHandler:    statichandler,
		},
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	return a.Services.Server.Start(ctx)
}

func (a *App) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	return a.Services.Server.Shutdown(ctx)
}

// TODO: Start/stop all services, handlers, controllers and repositories to gracefull restarts
