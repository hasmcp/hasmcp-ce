package mcp

import (
	"context"
	"sync"

	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/cache"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/jwt"
	protocol "github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/protocol/p250618"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/jsonrpc"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/httpc"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/idgen"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/locksmith"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/memq"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/pubsub"
	"github.com/mustafaturan/monoflake"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

type (
	Controller interface {
		// MCP Protocol methods
		CallSession(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error)
		DeleteSession(ctx context.Context, req DeleteSessionRequest) error
		SubscribeSession(ctx context.Context, req SubscribeSessionRequest) (*SubscribeSessionResponse, error)
		UnsubscribeSession(ctx context.Context, req UnsubscribeSessionRequest) error

		// Tail IO Sessions
		StartTailIO(ctx context.Context, req StartTailIORequest) (*StartTailIOResponse, error)
		StopTailIO(ctx context.Context, req StopTailIORequest) error

		// CRUD updates
		HandleChanges(ctx context.Context, change entity.ResourceChange) error
	}

	controller struct {
		idgen     idgen.Service
		httpc     httpc.Service
		locksmith locksmith.Service
		memq      memq.Service
		pubsub    pubsub.Service
		jwt       jwt.Controller
		cache     cache.Controller

		servers sync.Map

		queueIDForResourceUpdates uint32
	}

	serverSession struct {
		pubsubID         int64
		initializeParams protocol.InitializeRequestParams
	}

	server struct {
		toolIDs                    []int64
		resourceIDs                []int64
		promptIDs                  []int64
		requestHeadersProxyEnabled bool
		sessions                   *sync.Map
		protocol                   protocolComponents
	}

	protocolComponents struct {
		implementation protocol.Implementation
		tools          map[int64]protocol.Tool
		prompts        map[int64]protocol.Prompt
		resources      map[int64]protocol.Resource
	}

	Params struct {
		Config    config.Service
		IDGen     idgen.Service
		HTTPC     httpc.Service
		Locksmith locksmith.Service
		Memq      memq.Service
		PubSub    pubsub.Service
		Cache     cache.Controller
		McpJWT    jwt.Controller
	}

	DeleteSessionRequest struct {
		ServerID           int64
		McpSessionID       string
		McpProtocolVersion string
		Permissions        map[string]struct{}
	}

	CallSessionRequest struct {
		Headers            map[string][]string
		ServerID           int64
		McpSessionID       string
		McpProtocolVersion string
		Permissions        map[string]struct{}
		Request            jsonrpc.Request
	}

	CallSessionResponse struct {
		HTTPStatusCode     int
		McpSessionID       string
		McpProtocolVersion string
		Result             *jsonrpc.ResultResponse
	}

	mcpConfig struct {
	}

	Method string

	err string
)

const (
	ScopeSessionCreate = "session:create"
	ScopeSessionCall   = "session:call"
	ScopeSessionStream = "session:stream"
	ScopeSessionDelete = "session:delete"
	ScopeServerTail    = "server:tail"
)

const (
	_cfgKey = "mcp"

	_logPrefix = "[mcp] "

	ErrNotImplemented err = "not implemented"

	_serverProtocolVersion = "2025-06-18"

	// some clients currently does not support pagination, keeping this number high for now
	_paginationLimitToolsList    = 100
	_paginationLimitResourceList = 10
	_paginationLimitPromptList   = 10
)

func (e err) Error() string {
	return string(e)
}

func New(p Params) (Controller, error) {
	var cfg mcpConfig
	err := p.Config.Populate(_cfgKey, &cfg)
	if err != nil {
		return nil, err
	}

	c := &controller{
		idgen:     p.IDGen,
		httpc:     p.HTTPC,
		locksmith: p.Locksmith,
		memq:      p.Memq,
		pubsub:    p.PubSub,

		jwt:   p.McpJWT,
		cache: p.Cache,

		servers: sync.Map{},
	}

	res, err := c.memq.Create(context.Background(), memq.CreateRequest{
		Name: "MCP_RESOURCE_UPDATES",
		Size: 100000,
	})
	if err != nil {
		return nil, err
	}

	err = c.memq.AddWorkers(context.Background(), memq.AddWorkersRequest{
		Count:   1,
		QueueID: res.ID,
		Handle:  c.applyUpdatesOnChanges,
	})
	if err != nil {
		return nil, err
	}

	c.queueIDForResourceUpdates = res.ID
	return c, nil
}

func (c *controller) HandleChanges(ctx context.Context, change entity.ResourceChange) error {
	err := c.memq.AddTask(ctx, memq.AddTaskRequest{
		QueueID: c.queueIDForResourceUpdates,
		Task: memq.Task{
			ID:  c.idgen.Next(),
			Val: change,
		},
	})
	if err != nil {
		zlog.Error().Err(err).Msg("failed to queue the resource change")
		return err
	}
	return nil
}

func (c *controller) applyUpdatesOnChanges(ctx context.Context, t memq.Task) error {
	change := t.Val.(entity.ResourceChange)
	serverID := change.ResourceOwnerID
	if change.ObjectType == entity.ObjectTypeServer && change.EventType == entity.ObjectEventTypeDelete {
		// loop through sessions and close
		c.servers.Delete(change.ResoureID)
		return nil
	}

	currentServer, err := c.getServer(serverID)
	if err != nil {
		return nil
	}

	newServer, err := c.buildServer(context.Background(), serverID)
	if err != nil {
		zlog.Error().Err(err).Int64("id", serverID).Msg(_logPrefix + "failed to build server, to not to risk deleting it!")
		c.servers.Delete(serverID)
		return err
	}

	newServer.sessions = currentServer.sessions
	c.servers.Store(serverID, newServer)

	zlog.Info().Any("tools", newServer.toolIDs).Msg(_logPrefix + "saved new server with tools")

	toolsListChanged := false
	promptsListChanged := false
	resourcesListChanged := false

	switch change.ObjectType {
	case entity.ObjectTypeServer:
		toolsListChanged = true
		if len(currentServer.toolIDs) == len(newServer.toolIDs) {
			toolsListChanged = false
			// added
			for _, t := range newServer.toolIDs {
				if currentServer.protocol.tools[t].Name == "" {
					toolsListChanged = true
					break
				}
			}
			// removed
			for _, t := range currentServer.toolIDs {
				if newServer.protocol.tools[t].Name == "" {
					toolsListChanged = true
					break
				}
			}
		}
		promptsListChanged = true
		if len(currentServer.promptIDs) == len(newServer.promptIDs) {
			promptsListChanged = false
			// added
			for _, t := range newServer.promptIDs {
				if currentServer.protocol.prompts[t].Name == "" {
					promptsListChanged = true
					break
				}
			}
			// removed
			for _, t := range currentServer.promptIDs {
				if newServer.protocol.prompts[t].Name == "" {
					promptsListChanged = true
					break
				}
			}
		}
		resourcesListChanged = true
		if len(currentServer.resourceIDs) == len(newServer.resourceIDs) {
			resourcesListChanged = false
			// added
			for _, t := range newServer.resourceIDs {
				if currentServer.protocol.resources[t].Name == "" {
					resourcesListChanged = true
					break
				}
			}
			// removed
			for _, t := range currentServer.resourceIDs {
				if newServer.protocol.resources[t].Name == "" {
					resourcesListChanged = true
					break
				}
			}
		}
	case entity.ObjectTypeProviderTool, entity.ObjectTypeProvider, entity.ObjectTypeServerTool:
		toolsListChanged = true
	case entity.ObjectTypePrompt, entity.ObjectTypeServerPrompt:
		promptsListChanged = true
	case entity.ObjectTypeResource, entity.ObjectTypeServerResource:
		resourcesListChanged = true
	}

	if toolsListChanged {
		currentServer.sessions.Range(func(key, _ any) bool {
			sessionID, ok := key.(int64)
			if ok {
				c.sendSessionNotification(ctx, CallSessionRequest{
					ServerID:     serverID,
					McpSessionID: monoflake.ID(sessionID).String(),
					Request: jsonrpc.Request{
						Method: MethodNotificationToolsListChanged,
						Params: []byte(""),
					},
				})
			}
			return true
		})
	}

	if promptsListChanged {
		currentServer.sessions.Range(func(key, _ any) bool {
			sessionID, ok := key.(int64)
			if ok {
				c.sendSessionNotification(ctx, CallSessionRequest{
					ServerID:     change.ResourceOwnerID,
					McpSessionID: monoflake.ID(sessionID).String(),
					Request: jsonrpc.Request{
						Method: MethodNotificationPromptsListChanged,
						Params: []byte(""),
					},
				})
			}
			return true
		})
	}

	if resourcesListChanged {
		currentServer.sessions.Range(func(key, _ any) bool {
			sessionID, ok := key.(int64)
			if ok {
				c.sendSessionNotification(ctx, CallSessionRequest{
					ServerID:     change.ResourceOwnerID,
					McpSessionID: monoflake.ID(sessionID).String(),
					Request: jsonrpc.Request{
						Method: MethodNotificationResourcesListChanged,
						Params: []byte(""),
					},
				})
			}
			return true
		})
	}

	currentServer = nil

	return nil
}

func (c *controller) getServer(id int64) (*server, error) {
	serverVal, ok := c.servers.Load(id)
	if !ok {
		server, err := c.buildServer(context.Background(), id)
		if err != nil {
			return nil, erre.Error{
				Code:    404,
				Message: "Not found",
				Data: map[string]any{
					"reason": err.Error(),
				},
			}
		}
		c.servers.Store(id, server)
		return server, nil
	}

	return serverVal.(*server), nil
}

func (c *controller) getSession(serverID, sessionID int64) (*serverSession, error) {
	// server
	serv, err := c.getServer(serverID)
	if err != nil {
		return nil, err
	}

	// load session
	sessionVal, ok := serv.sessions.Load(sessionID)
	if !ok {
		return nil, erre.Error{
			Code:    404,
			Message: "Session not found",
		}
	}

	return sessionVal.(*serverSession), nil
}

func (r CallSessionRequest) MarshalZerologObject(e *zerolog.Event) {
	e.
		Int64("serverID", r.ServerID).
		Str("sessionID", r.McpSessionID).
		Str("protocolVersion", r.McpProtocolVersion).
		Str("method", r.Request.Method).
		Str("params", string(r.Request.Params))
}

func (r CallSessionResponse) MarshalZerologObject(e *zerolog.Event) {
	e.
		Str("sessionID", r.McpSessionID).
		Str("protocolVersion", r.McpProtocolVersion).
		Any("result", r.Result)
}
