package jwt

import (
	"context"
	"fmt"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	"github.com/mustafaturan/monoflake"
)

type (
	Controller interface {
		IssuerController
		Oauth2StateVerifierController
	}

	IssuerController interface {
		Issue(ctx context.Context, p IssueParams) (*IssueResult, error)
	}

	Oauth2StateVerifierController interface {
		VerifyState(ctx context.Context, p VerifyStateParams) (*VerifyStateResult, error)
	}

	controller struct {
		secret []byte
	}

	Params struct {
		Config config.Service
	}

	VerifyStateParams struct {
		AccessToken []byte
	}

	VerifyStateResult struct {
		ProviderID int64
		ServerID   int64
	}

	ProviderClaims struct {
		jwtv5.RegisteredClaims
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
	_cfgKey = "oauth2McpProviderJwt"
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

func (c *controller) VerifyState(ctx context.Context, p VerifyStateParams) (*VerifyStateResult, error) {
	claims := &ProviderClaims{}
	token, err := jwtv5.ParseWithClaims(string(p.AccessToken), claims, func(token *jwtv5.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return c.secret, nil
	}, jwtv5.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, erre.Error{
			Code:    401,
			Message: "could not authenticate",
			Data: map[string]any{
				"details": err.Error(),
			},
		}
	}

	claims, ok := token.Claims.(*ProviderClaims)
	if !ok || !token.Valid {
		return nil, erre.Error{
			Code:    401,
			Message: "invalid access token",
		}
	}

	if len(claims.Audience) < 2 {
		return nil, erre.Error{
			Code:    401,
			Message: "invalid access token",
		}
	}

	return &VerifyStateResult{
		ProviderID: monoflake.IDFromBase62(claims.Audience[0]).Int64(),
		ServerID:   monoflake.IDFromBase62(claims.Audience[1]).Int64(),
	}, nil
}
