package mcp

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/jwt"
	protocol "github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/protocol/p250618"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/jsonrpc"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/pubsub"
	"github.com/mustafaturan/monoflake"
	zlog "github.com/rs/zerolog/log"
)

const (
	// MethodInitialize initiates connection and negotiates protocol capabilities.
	// https://modelcontextprotocol.io/specification/2024-11-05/basic/lifecycle/#initialization
	MethodInitialize Method = "initialize"
)

var (
	_serverCapabilities = protocol.ServerCapabilities{
		Experimental: nil,
		Logging:      nil,
		Completions:  nil,
		Prompts: &protocol.ServerCapabilitiesPrompts{
			ListChanged: boolPtr(true),
		},
		Resources: &protocol.ServerCapabilitiesResources{
			ListChanged: boolPtr(true),
			Subscribe:   boolPtr(false), // Enabled subscribe capability
		},
		Tools: &protocol.ServerCapabilitiesTools{
			ListChanged: boolPtr(true),
		},
	}

	_regexPatternToolName = regexp.MustCompile("[^A-Za-z0-9]")
)

func (c *controller) CallInitialize(ctx context.Context, req CallSessionRequest) (*CallSessionResponse, error) {
	var params protocol.InitializeRequestParams
	err := json.Unmarshal(req.Request.Params, &params)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInvalidJsonReceived,
			Message: "Invalid params",
			Data: map[string]any{
				"params": string(req.Request.Params),
				"reason": err.Error(),
			},
		}
	}

	var srv *server
	srv, err = c.getServer(req.ServerID)
	if err != nil {
		srv, err = c.buildServer(ctx, req.ServerID)
		if err != nil {
			return nil, err
		}
		c.servers.Store(req.ServerID, srv)
	}

	// create pubsub for this session
	pubsubResp, err := c.pubsub.Create(ctx, pubsub.CreatePubSubRequest{})
	if err != nil {
		zlog.Warn().Err(err).Msg("failed to create PubSub")
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "Failed to initialize PubSub for HTTP streaming",
			Data: map[string]any{
				"reason": err.Error(),
			},
		}
	}

	// create debug pubsub for this server
	// must set the ID to serverID to receive all sessions debug logs
	// if it exists, it won't be creating another one
	_, _ = c.pubsub.Create(ctx, pubsub.CreatePubSubRequest{
		ID: req.ServerID,
	})

	// create session
	sessionID := pubsubResp.ID
	srv.sessions.Store(sessionID, &serverSession{
		initializeParams: params,
		pubsubID:         sessionID,
	})

	result := protocol.InitializeResult{
		ProtocolVersion: _serverProtocolVersion,
		Capabilities:    _serverCapabilities,
		ServerInfo:      srv.protocol.implementation,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "Failed to marshal initialize result",
			Data: map[string]any{
				"reason": err.Error(),
			},
		}
	}

	mcpSessionToken, err := c.jwt.Issue(ctx, jwt.IssueParams{
		Claims: jwt.SessionClaims{
			ServerID:         monoflake.ID(req.ServerID).String(),
			InitializeParams: params,
			RegisteredClaims: jwtv5.RegisteredClaims{
				ID:        monoflake.ID(sessionID).String(),
				ExpiresAt: jwtv5.NewNumericDate(time.Now().UTC().AddDate(1, 0, 0)),
			},
		},
	})
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInternalError,
			Message: "Failed to issue session token",
			Data: map[string]any{
				"reason": err.Error(),
			},
		}
	}

	return &CallSessionResponse{
		HTTPStatusCode:     200,
		McpSessionID:       mcpSessionToken.Token,
		McpProtocolVersion: req.McpProtocolVersion,
		Result: &jsonrpc.ResultResponse{
			JSONRpc: jsonrpc.Version,
			ID:      req.Request.ID,
			Result:  data,
		},
	}, nil
}

func (c *controller) buildServer(ctx context.Context, serverID int64) (*server, error) {
	// build server
	// 1. Get server from db
	// 2. Get server tools from db
	// 3. Get server prompts from db
	// 3. Get server resources from db
	// 5. Build tools
	mcpsrv, err := c.cache.ReloadServer(ctx, serverID)
	if err != nil {
		return nil, jsonrpc.Error{
			Code:    jsonrpc.ErrCodeInvalidParams,
			Message: "Server not found",
			Data: map[string]any{
				"serverID": serverID,
				"reason":   err.Error(),
			},
		}
	}

	providerIDs := make([]int64, len(mcpsrv.Providers))
	toolIDs := make([]int64, 0)
	tools := make(map[int64]protocol.Tool)
	for i, p := range mcpsrv.Providers {
		providerIDs[i] = p.ID
		for _, e := range p.Tools {
			toolIDs = append(toolIDs, e.ID)
			title := e.Title
			if title == "" {
				title = entity.MethodType(e.Method).String() + " " + e.Path
			}
			name := e.Name

			inputSchemaProperties := protocol.ToolInputSchemaProperties{}
			required := make([]string, 0, 3)
			if len(e.PathArgsJSONSchema) > 2 {
				var props map[string]any
				_ = json.Unmarshal(e.PathArgsJSONSchema, &props)
				inputSchemaProperties["pathArgs"] = props
				required = append(required, "pathArgs")
			}

			if len(e.QueryArgsJSONSchema) > 2 {
				var props map[string]any
				_ = json.Unmarshal(e.QueryArgsJSONSchema, &props)
				inputSchemaProperties["queryArgs"] = props
				required = append(required, "queryArgs")
			}

			if len(e.ReqBodyJSONSchema) > 0 {
				var props map[string]any
				_ = json.Unmarshal(e.ReqBodyJSONSchema, &props)
				inputSchemaProperties["bodyArgs"] = props
				required = append(required, "bodyArgs")
			}

			if len(required) == 0 {
				required = nil
			}

			tools[e.ID] = protocol.Tool{
				// NOTE: Some of the clients still show the Name only instead of title.
				// NOTE: Gemini-CLI expects the name starts with letter
				Name:        toMcpName('T', e.ID, name, title, len(mcpsrv.Name)),
				Description: stringPtr(e.Description),
				Title:       stringPtr(title),
				InputSchema: protocol.ToolInputSchema{
					Type:       "object",
					Properties: inputSchemaProperties,
					Required:   required,
				},
			}
		}
	}

	prompts := make(map[int64]protocol.Prompt, len(mcpsrv.Prompts))
	promptIDs := make([]int64, len(mcpsrv.Prompts))
	for i, p := range mcpsrv.Prompts {
		promptIDs[i] = p.ID
		prompts[p.ID] = protocol.Prompt{
			// NOTE: Some of the clients still show the Name only instead of title.
			Name:        toMcpName('P', p.ID, p.Name, "", len(mcpsrv.Name)),
			Title:       stringPtr(p.Name),
			Description: stringPtr(p.Description),
		}
	}

	resources := make(map[int64]protocol.Resource, len(mcpsrv.Resources))
	resourceIDs := make([]int64, len(mcpsrv.Resources))
	for i, r := range mcpsrv.Resources {
		resourceIDs[i] = r.ID
		resources[r.ID] = protocol.Resource{
			// NOTE: Some of the clients still show the Name only instead of title.
			Name:        toMcpName('R', r.ID, r.Name, "", len(mcpsrv.Name)),
			Title:       stringPtr(r.Name),
			Description: stringPtr(r.Description),
			Uri:         r.URI,
			MimeType:    stringPtr(r.MimeType),
			Size:        intPtr(int(r.Size)),
		}
	}

	return &server{
		requestHeadersProxyEnabled: mcpsrv.RequestHeadersProxyEnabled,
		toolIDs:                    toolIDs,
		resourceIDs:                resourceIDs,
		promptIDs:                  promptIDs,
		sessions:                   &sync.Map{},
		protocol: protocolComponents{
			implementation: protocol.Implementation{
				Name:    mcpsrv.Name,
				Version: strconv.Itoa(int(mcpsrv.Version)),
			},
			tools:     tools,
			prompts:   prompts,
			resources: resources,
		},
	}, nil
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtr(i int) *int {
	return &i
}

func toMcpName(mcpFeaturePrefix rune, id int64, name, title string, serverNameLen int) string {
	title = strings.ReplaceAll(title, "[", "")
	title = strings.ReplaceAll(title, "]", "")
	title = strings.ReplaceAll(title, "{", "")
	title = strings.ReplaceAll(title, "}", "")
	if name != "" {
		name = string(mcpFeaturePrefix) + monoflake.ID(id).String() + "_" + name
	} else {
		name = string(mcpFeaturePrefix) + monoflake.ID(id).String() + "_" + title
	}
	name = _regexPatternToolName.ReplaceAllString(name, "_")
	name = strings.ReplaceAll(name, "__", "_")
	// 42 is the cursor limit for name, per spec mcp clients should display the
	// title not the name but it will take time to have this standard practice in
	// all mcp clients
	if len(name) > 41-serverNameLen {
		name = name[0 : 41-serverNameLen]
	}
	return name
}
