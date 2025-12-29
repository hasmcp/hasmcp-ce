package jwt

import (
	"context"
	"fmt"
	"strings"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	protocol "github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp/protocol/p250618"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	"github.com/mustafaturan/monoflake"
)

type (
	Controller interface {
		IssuerController
		AuthenticatorController
		SessionVerifierController
	}

	IssuerController interface {
		Issue(ctx context.Context, p IssueParams) (*IssueResult, error)
	}

	AuthenticatorController interface {
		Authenticate(ctx context.Context, p AuthParams) (*AuthResult, error)
	}

	SessionVerifierController interface {
		VerifySession(ctx context.Context, p AuthParams) (*SessionResult, error)
	}

	controller struct {
		secret []byte
	}

	Params struct {
		Config config.Service
	}

	ServerClaims struct {
		ServerID string `json:"serverID"`
		Scope    string `json:"scope"`
		jwtv5.RegisteredClaims
	}

	SessionClaims struct {
		ServerID         string                           `json:"serverID"`
		InitializeParams protocol.InitializeRequestParams `json:"initializeParams"`
		jwtv5.RegisteredClaims
	}

	AuthParams struct {
		AccessToken []byte
	}

	AuthResult struct {
		ServerID    int64
		Permissions map[string]struct{}
	}

	SessionResult struct {
		ServerID         int64
		SessionID        int64
		InitializeParams protocol.InitializeRequestParams
	}

	jwtConfig struct {
		Secret string `yaml:"secret"`
	}

	IssueParams struct {
		Claims jwtv5.Claims
	}

	IssueResult struct {
		Token string
	}
)

const (
	_cfgKey = "mcpjwt"
)

func New(p Params) (Controller, error) {
	var cfg jwtConfig
	err := p.Config.Populate(_cfgKey, &cfg)
	if err != nil {
		return nil, err
	}

	return &controller{
		secret: []byte(cfg.Secret),
	}, nil
}

func (c *controller) Issue(ctx context.Context, p IssueParams) (*IssueResult, error) {
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, p.Claims)
	tokenString, err := token.SignedString(c.secret)
	if err != nil {
		return nil, err
	}

	return &IssueResult{
		Token: tokenString,
	}, nil
}

func (c *controller) Authenticate(ctx context.Context, p AuthParams) (*AuthResult, error) {
	claims := &ServerClaims{}
	token, err := jwtv5.ParseWithClaims(string(p.AccessToken), claims, func(token *jwtv5.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return c.secret, nil
	}, jwtv5.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, erre.Error{
			Code:    401,
			Message: "Could not authenticate",
			Data: map[string]any{
				"details": err.Error(),
			},
		}
	}

	claims, ok := token.Claims.(*ServerClaims)
	if !ok || !token.Valid {
		return nil, erre.Error{
			Code:    401,
			Message: "Invalid access token",
		}
	}

	scope := strings.Split(claims.Scope, " ")
	permissions := make(map[string]struct{}, len(scope))
	for _, s := range scope {
		permissions[s] = struct{}{}
	}

	return &AuthResult{
		ServerID:    monoflake.IDFromBase62(claims.ServerID).Int64(),
		Permissions: permissions,
	}, nil
}

func (c *controller) VerifySession(ctx context.Context, p AuthParams) (*SessionResult, error) {
	claims := &SessionClaims{}
	token, err := jwtv5.ParseWithClaims(string(p.AccessToken), claims, func(token *jwtv5.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return c.secret, nil
	}, jwtv5.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, erre.Error{
			Code:    401,
			Message: "Could not authenticate",
			Data: map[string]any{
				"details": err.Error(),
			},
		}
	}

	claims, ok := token.Claims.(*SessionClaims)
	if !ok || !token.Valid {
		return nil, erre.Error{
			Code:    401,
			Message: "Invalid access token",
		}
	}

	return &SessionResult{
		ServerID:         monoflake.IDFromBase62(claims.ServerID).Int64(),
		SessionID:        monoflake.IDFromBase62(claims.ID).Int64(),
		InitializeParams: claims.InitializeParams,
	}, nil
}
