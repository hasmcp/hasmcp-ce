package oauth2mcp

import (
	"context"
	"regexp"
	"strings"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/crud"
	"github.com/hasmcp/hasmcp-ce/backend/internal/controller/oauth2mcp/jwt"
	crude "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/locksmith"
	"github.com/mustafaturan/monoflake"
	"golang.org/x/oauth2"
)

type (
	Controller interface {
		Authorize(ctx context.Context, req AuthorizeRequest) (*AuthorizeResponse, error)
		Callback(ctx context.Context, req CallbackRequest) (*CallbackResponse, error)
	}

	controller struct {
		locksmith    locksmith.Service
		oauth2Config oauth2Config
		crud         crud.Controller
		jwt          jwt.Controller
	}

	Params struct {
		Config    config.Service
		Locksmith locksmith.Service
		Crud      crud.Controller
		JWT       jwt.Controller
	}

	oauth2Config struct {
		HTTPScheme string `yaml:"httpScheme"`
		Secret     string `yaml:"secret"`
	}

	AuthorizeRequest struct {
		HostName string
		ServerID int64
	}

	AuthorizeResponse struct {
		AuthCodeURL string
	}

	CallbackRequest struct {
		HostName string
		State    string
		Code     string
	}

	CallbackResponse struct {
		InternalRedirectURL string
	}
)

const (
	_cfgKey = "oauth2McpProvider"
)

var (
	_regexVariableName = regexp.MustCompile(`\$\{([A-Z0-9_]+)\}`)
)

func New(p Params) (Controller, error) {
	var cfg oauth2Config
	err := p.Config.Populate(_cfgKey, &cfg)
	if err != nil {
		return nil, err
	}

	return &controller{
		oauth2Config: cfg,
		locksmith:    p.Locksmith,
		crud:         p.Crud,
		jwt:          p.JWT,
	}, nil
}

// Authorize authneticates user with the mcp server's provider oauth2 mechanism
// with the necessary scopes. In the current version there is a catch: if there
// are multiple mcp servers with different scope requirements then the new
// request overrides the other for the current user.
func (c *controller) Authorize(ctx context.Context, req AuthorizeRequest) (*AuthorizeResponse, error) {
	res, err := c.crud.GetServer(ctx, crude.GetServerRequest{
		ID: req.ServerID,
	})
	if err != nil {
		return nil, err
	}

	server := res.Server

	// NOTE: It is possible to have a MCP server without a provider
	// It is intended to auth only the ones have a provider
	if len(server.Providers) != 1 {
		return nil, erre.Error{
			Code:    erre.ErrorCodeUnprocessableEntity,
			Message: "a provider must be assigned to MCP server to authorize",
			Data: map[string]any{
				"serverID":      server.ID,
				"providerCount": len(server.Providers),
			},
		}
	}

	providerRes, err := c.crud.GetProvider(ctx, crude.GetProviderRequest{
		ID: server.Providers[0].ID,
	})
	if err != nil {
		return nil, err
	}

	provider := providerRes.Provider
	oauth2Cfg := provider.Oauth2Config
	if len(oauth2Cfg.ClientSecretEncrypted) > 0 {
		res, err := c.locksmith.Decrypt(ctx, &locksmith.DecryptRequest{
			Nonce:      oauth2Cfg.ClientSecretEncryptionNonce,
			Ciphertext: oauth2Cfg.ClientSecretEncrypted,
		})
		if err != nil {
			return nil, err
		}
		oauth2Cfg.ClientSecret = string(res.Plaintext)
	}

	clientID := oauth2Cfg.ClientID
	clientSecret := oauth2Cfg.ClientSecret
	tokenURL := oauth2Cfg.TokenURL
	authURL := oauth2Cfg.AuthURL

	if clientID == "" || clientSecret == "" || tokenURL == "" || authURL == "" {
		return nil, erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: "provider client credentials are missing",
			Data: map[string]any{
				"clientIDMissing":     clientID == "",
				"clientSecretMissing": clientSecret == "",
				"tokenURLMissing":     tokenURL == "",
				"authURLMissing":      authURL == "",
			},
		}
	}

	toolSet := make(map[int64][]string, len(provider.Tools))
	for _, e := range provider.Tools {
		toolSet[e.ID] = e.Oauth2Scopes
	}

	scopeSet := map[string]struct{}{}
	scopes := make([]string, 0)
	for _, e := range server.Providers[0].Tools {
		var ok bool
		for _, s := range toolSet[e.ID] {
			s = strings.Trim(s, " ")
			if s == "" {
				continue
			}
			if _, ok = scopeSet[s]; ok {
				continue
			}
			scopeSet[s] = struct{}{}
			scopes = append(scopes, s)
		}
	}

	providerID := monoflake.ID(provider.ID).String()
	// NOTE: This URL must be registered to provider for better security
	redirectURL := c.oauth2Config.HTTPScheme + "://" + req.HostName + "/oauth2/callback"

	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   provider.Oauth2Config.AuthURL,
			TokenURL:  provider.Oauth2Config.TokenURL,
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}

	rand, err := c.locksmith.GenerateRandomString64(ctx)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "couldn't generate random string for nonce",
			Data: map[string]any{
				"reason": err.Error(),
			},
		}
	}

	// Give some time user to authenticate on the external service
	expiresAt := time.Now().UTC().Add(time.Second * 180)

	serverID := monoflake.ID(server.ID).String()
	// Generate jwt token for state
	token, err := c.jwt.Issue(ctx, jwt.IssueParams{
		Claims: jwt.ProviderClaims{
			RegisteredClaims: jwtv5.RegisteredClaims{
				ID:        rand[:16],
				ExpiresAt: jwtv5.NewNumericDate(expiresAt),
				Audience:  []string{providerID, serverID},
			},
		},
	})

	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "couldn't generate jwt token for state",
			Data: map[string]any{
				"reason": err.Error(),
			},
		}
	}

	url := oauthConfig.AuthCodeURL(token.Token, oauth2.AccessTypeOffline)
	return &AuthorizeResponse{
		AuthCodeURL: url,
	}, nil
}

func (c *controller) Callback(ctx context.Context, req CallbackRequest) (*CallbackResponse, error) {
	if req.Code == "" {
		return nil, erre.Error{
			Code:    erre.ErrorCodeUnprocessableEntity,
			Message: "invalid authorization code",
		}
	}

	// Validate jwt token for state
	res, err := c.jwt.VerifyState(ctx, jwt.VerifyStateParams{
		AccessToken: []byte(req.State),
	})
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeUnauthorized,
			Message: "failed to verify state",
			Data: map[string]any{
				"reason": err.Error(),
			},
		}
	}

	providerID := monoflake.ID(res.ProviderID).String()
	serverID := monoflake.ID(res.ServerID).String()

	// Get provider
	providerRes, err := c.crud.GetProvider(ctx, crude.GetProviderRequest{
		ID: res.ProviderID,
	})
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeUnauthorized,
			Message: "couldn't find the provider",
			Data: map[string]any{
				"providerID": providerID,
				"serverID":   serverID,
			},
		}
	}

	provider := providerRes.Provider
	oauth2Cfg := provider.Oauth2Config
	if len(oauth2Cfg.ClientSecretEncrypted) > 0 {
		res, err := c.locksmith.Decrypt(ctx, &locksmith.DecryptRequest{
			Nonce:      oauth2Cfg.ClientSecretEncryptionNonce,
			Ciphertext: oauth2Cfg.ClientSecretEncrypted,
		})
		if err != nil {
			return nil, err
		}
		oauth2Cfg.ClientSecret = string(res.Plaintext)
	}

	cfg := &oauth2.Config{
		ClientID:     oauth2Cfg.ClientID,
		ClientSecret: oauth2Cfg.ClientSecret,
		RedirectURL:  c.oauth2Config.HTTPScheme + "://" + req.HostName + "/oauth2/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:   oauth2Cfg.AuthURL,
			TokenURL:  oauth2Cfg.TokenURL,
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}

	token, err := cfg.Exchange(ctx, req.Code)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeUnauthorized,
			Message: "couldn't get access token from provider",
			Data: map[string]any{
				"providerID": providerID,
				"serverID":   serverID,
				"reason":     err.Error(),
				"code":       req.Code,
			},
		}
	}

	accessTokenName := provider.SecretPrefix + "_ACCESS_TOKEN"
	refreshTokenName := provider.SecretPrefix + "_REFRESH_TOKEN"

	if len(provider.Tools) > 0 {
		foundAuthorizationVariable := false
		for _, t := range provider.Tools {
			for _, h := range t.Headers {
				if h.Key == "Authorization" {
					varNames := extractVariables(h.Value)
					if len(varNames) > 0 {
						accessTokenName = varNames[0]
						foundAuthorizationVariable = true
					}
					break
				}
			}
			if foundAuthorizationVariable {
				break
			}
		}
	}

	err = c.crud.SaveVariable(ctx, crude.SaveVariableRequest{
		Variable: crude.Variable{
			Name:  accessTokenName,
			Type:  crude.VariableTypeSecret,
			Value: []byte(token.AccessToken),
		},
	})
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "couldn't save access token to db",
			Data: map[string]any{
				"providerID": providerID,
				"serverID":   serverID,
				"reason":     err.Error(),
			},
		}
	}

	if token.RefreshToken != "" {
		err = c.crud.SaveVariable(ctx, crude.SaveVariableRequest{
			Variable: crude.Variable{
				Name:  refreshTokenName,
				Type:  crude.VariableTypeSecret,
				Value: []byte(token.RefreshToken),
			},
		})

		if err != nil {
			return nil, erre.Error{
				Code:    erre.ErrorCodeInternalServerError,
				Message: "couldn't save refresh token to db",
				Data: map[string]any{
					"providerID": providerID,
					"serverID":   serverID,
					"reason":     err.Error(),
				},
			}
		}
	}

	return &CallbackResponse{
		InternalRedirectURL: "/servers/" + serverID + "?message=Succesfully+added+access+and+refresh+tokens+to+variables",
	}, nil
}

func extractVariables(s string) []string {
	allMatches := _regexVariableName.FindAllStringSubmatch(s, -1)

	var names []string
	for _, match := range allMatches {
		if len(match) > 1 {
			names = append(names, match[1])
		}
	}

	return names
}
