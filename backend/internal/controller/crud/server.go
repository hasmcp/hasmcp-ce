package crud

import (
	"context"
	"fmt"
	"regexp"
	"time"

	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
	modelmapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/model"
	zlog "github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type ServerController interface {
	CreateServer(ctx context.Context, req entity.CreateServerRequest) (*entity.CreateServerResponse, error)
	UpdateServer(ctx context.Context, req entity.UpdateServerRequest) (*entity.UpdateServerResponse, error)
	DeleteServer(ctx context.Context, req entity.DeleteServerRequest) error
	GetServer(ctx context.Context, req entity.GetServerRequest) (*entity.GetServerResponse, error)
	ListServers(ctx context.Context, req entity.ListServersRequest) (*entity.ListServersResponse, error)
}

const (
	_initialServerVersion = int32(1)

	_validationAttrServerNameMaxLength         = 16
	_validationAttrServerInstructionsMaxLength = 4096
	_validationAttrServerProvidersMax          = 1
)

var (
	_regexPatternServerName = regexp.MustCompile(`^[a-zA-Z0-9]{1,16}$`)
)

func (c *controller) CreateServer(ctx context.Context, req entity.CreateServerRequest) (*entity.CreateServerResponse, error) {
	if err := c.validateCreateServerRequest(req); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	server := req.Server
	server.ID = c.idgen.Next()
	server.CreatedAt = now
	server.UpdatedAt = now
	server.Version = _initialServerVersion

	s := modelmapper.FromServerEntityServerModel(server)

	err := c.storage.CreateServer(ctx, s)
	if err != nil {
		return nil, err
	}

	return &entity.CreateServerResponse{Server: modelmapper.FromServerModelToServerEntity(s)}, nil
}

func (c *controller) UpdateServer(ctx context.Context, req entity.UpdateServerRequest) (*entity.UpdateServerResponse, error) {
	if err := c.validateUpdateServerRequest(req); err != nil {
		return nil, err
	}

	// Check if server exists
	_, err := c.GetServer(ctx, entity.GetServerRequest{
		ID: req.Server.ID,
	})
	if err != nil {
		return nil, err
	}

	s := modelmapper.FromServerEntityServerModel(req.Server)

	ctx = c.storage.ContextWithTx(ctx)
	err = c.storage.UpdateServer(ctx, s.ID, s.Version, map[model.ServerAttribute]any{
		model.ServerAttributeName:                       s.Name,
		model.ServerAttributeInstructions:               s.Instructions,
		model.ServerAttributeVersion:                    s.Version,
		model.ServerAttributeTools:                      s.Tools,
		model.ServerAttributeResources:                  s.Resources,
		model.ServerAttributePrompts:                    s.Prompts,
		model.ServerAttributeRequestHeadersProxyEnabled: s.RequestHeadersProxyEnabled,
	})

	if err != nil {
		rollbackErr := c.storage.TxRollback(ctx)
		zlog.Error().Err(rollbackErr).Str("updateErr", err.Error()).Msg("failed to rollback transaction on updating server")
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to update the server",
			Data: map[string]any{
				"reason": err.Error(),
			},
		}
	}
	err = c.storage.TxCommit(ctx)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to commit changes to db",
			Data: map[string]any{
				"reason": err.Error(),
			},
		}
	}

	defer func() {
		freshCtx := context.Background()
		c.cache.Evict(freshCtx, entity.ObjectTypeServer, req.Server.ID)
		err := c.mcp.HandleChanges(freshCtx, entity.ResourceChange{
			ObjectType:      entity.ObjectTypeServer,
			EventType:       entity.ObjectEventTypeUpdate,
			ResoureID:       s.ID,
			ResourceOwnerID: s.ID,
		})
		if err != nil {
			zlog.Error().Err(err).Msg("failed to deliver changes")
		}
	}()

	serverRes, err := c.GetServer(context.Background(), entity.GetServerRequest{
		ID: s.ID,
	})
	if err != nil {
		return nil, err
	}

	return &entity.UpdateServerResponse{
		Server: serverRes.Server,
	}, nil
}

func (c *controller) GetServer(ctx context.Context, req entity.GetServerRequest) (*entity.GetServerResponse, error) {
	// Check if server exists
	server, err := c.storage.GetServer(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, erre.Error{
				Code:    erre.ErrorCodeNotFound,
				Message: "server not found",
				Data: map[string]any{
					"reason":   err.Error(),
					"serverID": req.ID,
				},
			}
		}
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to get server",
			Data: map[string]any{
				"reason":   err.Error(),
				"serverID": req.ID,
			},
		}
	}
	if server == nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeNotFound,
			Message: "server not found",
			Data: map[string]any{
				"serverID": req.ID,
			},
		}
	}

	return &entity.GetServerResponse{
		Server: modelmapper.FromServerModelToServerEntity(*server),
	}, nil
}

func (c *controller) ListServers(ctx context.Context, req entity.ListServersRequest) (*entity.ListServersResponse, error) {
	servers, err := c.storage.ListServers(ctx)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to list servers",
			Data: map[string]any{
				"reason": err.Error(),
			},
		}
	}

	resp := &entity.ListServersResponse{
		Servers: modelmapper.FromServerModelsToServerEntities(servers),
	}

	return resp, nil
}

func (c *controller) DeleteServer(ctx context.Context, req entity.DeleteServerRequest) error {
	defer c.cache.Evict(ctx, entity.ObjectTypeServer, req.ID)

	err := c.storage.DeleteServer(ctx, req.ID)

	if err != nil {
		return erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to delete server",
			Data: map[string]any{
				"reason":   err.Error(),
				"serverID": req.ID,
			},
		}
	}

	_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
		ObjectType:      entity.ObjectTypeServer,
		EventType:       entity.ObjectEventTypeDelete,
		ResoureID:       req.ID,
		ResourceOwnerID: req.ID,
	})

	return nil
}

func (c *controller) validateCreateServerRequest(req entity.CreateServerRequest) error {
	s := req.Server
	if len(s.Name) == 0 || len(s.Name) > _validationAttrServerNameMaxLength {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: fmt.Sprintf("name must be between 1 and %d characters", _validationAttrServerNameMaxLength),
			Data: map[string]any{
				"name": s.Name,
			},
		}
	}

	if !_regexPatternServerName.MatchString(s.Name) {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: "name must be `[a-zA-Z0-9]{1,16}",
			Data: map[string]any{
				"name": s.Name,
			},
		}
	}

	if len(s.Instructions) > _validationAttrServerInstructionsMaxLength {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: fmt.Sprintf("instructions must be less than %d characters", _validationAttrServerInstructionsMaxLength),
			Data: map[string]any{
				"instructions": s.Instructions,
			},
		}
	}
	if len(s.Providers) > _validationAttrServerProvidersMax {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: fmt.Sprintf("max providers allowed per MCP server is set to %d", _validationAttrServerProvidersMax),
			Data: map[string]any{
				"providersCount": len(s.Providers),
			},
		}
	}
	return nil
}

func (c *controller) validateUpdateServerRequest(req entity.UpdateServerRequest) error {
	s := req.Server
	if s.ID == 0 {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: "invalid server ID",
			Data: map[string]any{
				"serverID": s.ID,
			},
		}
	}
	if s.Version == 0 {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: "invalid server version",
			Data: map[string]any{
				"version": s.Version,
			},
		}
	}
	if len(s.Name) == 0 || len(s.Name) > _validationAttrServerNameMaxLength {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: fmt.Sprintf("name must be between 1 and %d characters", _validationAttrServerNameMaxLength),
			Data: map[string]any{
				"name": s.Name,
			},
		}
	}

	if !_regexPatternServerName.MatchString(s.Name) {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: "name must be `[a-zA-Z0-9]{1,16}",
			Data: map[string]any{
				"name": s.Name,
			},
		}
	}

	if len(s.Instructions) > _validationAttrServerInstructionsMaxLength {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: fmt.Sprintf("instructions must be less than %d characters", _validationAttrServerInstructionsMaxLength),
			Data: map[string]any{
				"instructions": s.Instructions,
			},
		}
	}
	if len(s.Providers) > _validationAttrServerProvidersMax {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: fmt.Sprintf("max providers allowed per MCP server is set to %d", _validationAttrServerProvidersMax),
			Data: map[string]any{
				"providersCount": len(s.Providers),
			},
		}
	}
	return nil
}
