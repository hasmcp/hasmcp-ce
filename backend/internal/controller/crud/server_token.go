package crud

import (
	"context"
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/jwt"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/mustafaturan/monoflake"
)

const (
	_jwtTokenIssuer = "HasMCP"

	_jwtDefaultScope = mcp.ScopeSessionCreate + " " + mcp.ScopeSessionCall + " " + mcp.ScopeSessionDelete + " " + mcp.ScopeSessionStream
)

type ServerTokenController interface {
	CreateServerToken(ctx context.Context, req entity.CreateServerTokenRequest) (*entity.CreateServerTokenResponse, error)
}

func (c *controller) CreateServerToken(
	ctx context.Context, req entity.CreateServerTokenRequest) (*entity.CreateServerTokenResponse, error) {
	if err := c.validateCreateServerTokenRequest(req); err != nil {
		return nil, err
	}

	t := req.Token

	// Check if server exists
	_, err := c.GetServer(ctx, entity.GetServerRequest{
		ID: t.ServerID,
	})
	if err != nil {
		return nil, err
	}

	scope := t.Scope
	if scope == "" {
		scope = _jwtDefaultScope
	}

	id := c.idgen.Next()
	tokenRes, err := c.mcpJWT.Issue(ctx, jwt.IssueParams{
		Claims: jwt.ServerClaims{
			ServerID: monoflake.ID(t.ServerID).String(),
			Scope:    scope,
			RegisteredClaims: jwtv5.RegisteredClaims{
				ID:        c.idgen.NextString(),
				ExpiresAt: jwtv5.NewNumericDate(t.ExpiresAt),
				Issuer:    _jwtTokenIssuer,
				IssuedAt:  jwtv5.NewNumericDate(time.Now()),
			},
		},
	})
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to hash server token",
			Data: map[string]any{
				"serverID": t.ServerID,
				"reason":   err.Error(),
			},
		}
	}
	return &entity.CreateServerTokenResponse{
		ServerToken: entity.ServerToken{
			ID:          id,
			ServerID:    t.ServerID,
			CreatedAt:   t.CreatedAt,
			ExpiresAt:   t.ExpiresAt,
			Scope:       scope,
			ActualValue: []byte(tokenRes.Token),
		},
	}, nil
}

func (c *controller) validateCreateServerTokenRequest(req entity.CreateServerTokenRequest) error {
	t := req.Token
	if t.ServerID <= 0 {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: "server ID must be greater than 0",
			Data: map[string]any{
				"serverID": t.ServerID,
			},
		}
	}

	if t.ExpiresAt.Before(time.Now().UTC()) {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: fmt.Sprintf("expires at (%s) must be in the future", t.ExpiresAt),
			Data: map[string]any{
				"serverID": t.ServerID,
			},
		}
	}
	return nil
}
