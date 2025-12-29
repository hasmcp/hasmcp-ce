package crud

import (
	"context"
	"encoding/hex"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"

	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
	modelmapper "github.com/hasmcp/hasmcp-ce/backend/internal/mapper/model"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/locksmith"

	zlog "github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type ProviderController interface {
	CreateProvider(ctx context.Context, req entity.CreateProviderRequest) (*entity.CreateProviderResponse, error)
	GetProvider(ctx context.Context, req entity.GetProviderRequest) (*entity.GetProviderResponse, error)
	ListProviders(ctx context.Context, req entity.ListProvidersRequest) (*entity.ListProvidersResponse, error)
	UpdateProvider(ctx context.Context, req entity.UpdateProviderRequest) (*entity.UpdateProviderResponse, error)
	DeleteProvider(ctx context.Context, req entity.DeleteProviderRequest) error
}

const (
	_validationAttrProviderNameMaxLength        = 16
	_validationAttrProviderDescriptionMaxLength = 4096
	_validationAttrProviderBaseURLMaxLength     = 255
	_validationAttrProviderDocumentURLMaxLength = 255
	_validationAttrProviderIconURLMaxLength     = 255

	_providerInitialVersion = int32(1)
)

var (
	_regexPatternProviderName = regexp.MustCompile(`^[a-zA-Z0-9]{1,16}$`)
)

func (c *controller) CreateProvider(ctx context.Context, req entity.CreateProviderRequest) (*entity.CreateProviderResponse, error) {
	if err := c.validateCreateProviderRequest(req); err != nil {
		return nil, err
	}

	p := req.Provider

	var clientSecretEncrypted, clientSecretEncryptionNonce string
	if p.Oauth2Config.ClientSecret != "" {
		res, err := c.locksmith.Encrypt(ctx, &locksmith.EncryptRequest{
			Plaintext: []byte(p.Oauth2Config.ClientSecret),
		})
		if err != nil {
			return nil, erre.Error{
				Code:    erre.ErrorCodeUnprocessableEntity,
				Message: "crud: failed to encrypt the client secret",
				Data: map[string]any{
					"reason": err.Error(),
				},
			}
		}

		clientSecretEncrypted = hex.EncodeToString(res.Ciphertext)
		clientSecretEncryptionNonce = hex.EncodeToString(res.Nonce)
	}

	now := time.Now().UTC()
	id := c.idgen.Next()
	provider := model.Provider{
		ID:             id,
		CreatedAt:      now,
		UpdatedAt:      now,
		Version:        _providerInitialVersion,
		ApiType:        uint8(p.ApiType),
		VisibilityType: uint8(p.VisibilityType),
		BaseURL:        p.BaseURL,
		DocumentURL:    p.DocumentURL,
		IconURL:        p.IconURL,
		SecretPrefix:   buildSecretPrefix(p.BaseURL),
		Name:           p.Name,
		Description:    p.Description,
		Oauth2Config: model.ProviderOauth2Config{
			ID:                          id,
			ProviderID:                  id,
			ClientID:                    p.Oauth2Config.ClientID,
			ClientSecretEncrypted:       clientSecretEncrypted,
			ClientSecretEncryptionNonce: clientSecretEncryptionNonce,
			AuthURL:                     p.Oauth2Config.AuthURL,
			TokenURL:                    p.Oauth2Config.TokenURL,
		},
	}

	err := c.storage.CreateProvider(ctx, provider)
	if err != nil {
		return nil, err
	}
	return &entity.CreateProviderResponse{
		Provider: modelmapper.FromProviderModelToProviderEntity(provider),
	}, nil
}

func (c *controller) GetProvider(ctx context.Context, req entity.GetProviderRequest) (*entity.GetProviderResponse, error) {
	provider, err := c.storage.GetProvider(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, erre.Error{
				Code:    erre.ErrorCodeNotFound,
				Message: "provider not found",
				Data: map[string]any{
					"reason":     err.Error(),
					"providerID": req.ID,
				},
			}
		}
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to get provider",
			Data: map[string]any{
				"reason":     err.Error(),
				"providerID": req.ID,
			},
		}
	}
	if provider == nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeNotFound,
			Message: "provider not found",
			Data: map[string]any{
				"providerID": req.ID,
			},
		}
	}

	p := modelmapper.FromProviderModelToProviderEntity(*provider)

	return &entity.GetProviderResponse{
		Provider: p,
	}, nil
}

func (c *controller) ListProviders(ctx context.Context, req entity.ListProvidersRequest) (*entity.ListProvidersResponse, error) {
	providers, err := c.storage.ListProviders(ctx, nil)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to list providers",
			Data: map[string]any{
				"reason":     err.Error(),
				"filters":    req.Filters,
				"pagination": req.Pagination,
			},
		}
	}

	entities := modelmapper.FromProviderModelsToProviderEntities(providers)
	return &entity.ListProvidersResponse{Providers: entities}, nil
}

func (c *controller) DeleteProvider(ctx context.Context, req entity.DeleteProviderRequest) error {
	serverIDs, err := c.storage.ListServerIDsByProviderID(ctx, req.ID)
	if err != nil {
		zlog.Error().Err(err).Msg("failed to list server ids by resource")
	}

	err = c.storage.DeleteProvider(ctx, req.ID)
	if err != nil {
		return erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to delete provider",
			Data: map[string]any{
				"reason":     err.Error(),
				"providerID": req.ID,
			},
		}
	}

	c.cache.Evict(ctx, entity.ObjectTypeProvider, req.ID)
	for _, id := range serverIDs {
		_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
			ObjectType:      entity.ObjectTypeProvider,
			EventType:       entity.ObjectEventTypeUpdate,
			ResoureID:       req.ID,
			ResourceOwnerID: id,
		})
	}

	return nil
}

func (c *controller) UpdateProvider(ctx context.Context, req entity.UpdateProviderRequest) (*entity.UpdateProviderResponse, error) {
	if err := c.validateUpdateProviderRequest(req); err != nil {
		return nil, err
	}

	p := req.Provider

	attrs := make(map[model.ProviderAttribute]any)
	if p.Name != "" {
		attrs[model.ProviderAttributeName] = p.Name
	}
	if p.Description != "" {
		attrs[model.ProviderAttributeDescription] = p.Description
	}
	if p.DocumentURL != "" {
		attrs[model.ProviderAttributeDocumentURL] = p.DocumentURL
	}
	if p.IconURL != "" {
		attrs[model.ProviderAttributeIconURL] = p.IconURL
	}

	if p.Oauth2Config.AuthURL != "" && p.Oauth2Config.TokenURL != "" &&
		p.Oauth2Config.ClientID != "" && p.Oauth2Config.ClientSecret != "" && p.Oauth2Config.ClientSecret != "***" {
		var clientSecretEncrypted, clientSecretEncryptionNonce string
		if p.Oauth2Config.ClientSecret != "" {
			res, err := c.locksmith.Encrypt(ctx, &locksmith.EncryptRequest{
				Plaintext: []byte(p.Oauth2Config.ClientSecret),
			})
			if err != nil {
				return nil, erre.Error{
					Code:    erre.ErrorCodeUnprocessableEntity,
					Message: "crud: failed to encrypt the client secret",
					Data: map[string]any{
						"reason": err.Error(),
					},
				}
			}

			clientSecretEncrypted = hex.EncodeToString(res.Ciphertext)
			clientSecretEncryptionNonce = hex.EncodeToString(res.Nonce)
		}

		attrs[model.ProviderAttributeOauth2Config] = &model.ProviderOauth2Config{
			ID:                          req.Provider.ID,
			ProviderID:                  req.Provider.ID,
			ClientID:                    p.Oauth2Config.ClientID,
			ClientSecretEncrypted:       clientSecretEncrypted,
			ClientSecretEncryptionNonce: clientSecretEncryptionNonce,
			AuthURL:                     p.Oauth2Config.AuthURL,
			TokenURL:                    p.Oauth2Config.TokenURL,
		}
	}

	err := c.storage.UpdateProvider(ctx, p.ID, attrs)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to update provider",
			Data: map[string]any{
				"reason":     err.Error(),
				"providerID": p.ID,
				"attrs":      attrs,
			},
		}
	}

	providerRes, err := c.storage.GetProvider(ctx, p.ID)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to get provider",
			Data: map[string]any{
				"reason":     err.Error(),
				"providerID": p.ID,
				"attrs":      attrs,
			},
		}
	}

	serverIDs, err := c.storage.ListServerIDsByProviderID(ctx, p.ID)
	if err != nil {
		zlog.Error().Err(err).Msg("failed to list server ids by provider")
	}

	c.cache.Evict(ctx, entity.ObjectTypeProvider, req.Provider.ID)
	for _, id := range serverIDs {
		_ = c.mcp.HandleChanges(ctx, entity.ResourceChange{
			ObjectType:      entity.ObjectTypeProvider,
			EventType:       entity.ObjectEventTypeUpdate,
			ResoureID:       req.Provider.ID,
			ResourceOwnerID: id,
		})
	}

	return &entity.UpdateProviderResponse{
		Provider: modelmapper.FromProviderModelToProviderEntity(*providerRes),
	}, nil
}

func (c *controller) validateUpdateProviderRequest(req entity.UpdateProviderRequest) error {
	p := req.Provider
	var anyChanges bool
	if p.ID <= 0 {
		return errors.New("invalid provider ID")
	}

	if p.Name != "" {
		anyChanges = true
		if len(p.Name) > _validationAttrProviderNameMaxLength {
			return errors.New("name exceeds maximum length")
		}

		if !_regexPatternProviderName.MatchString(p.Name) {
			return errors.New("invalid provider name must be `[a-zA-Z0-9]{1,16}`")
		}
	}

	if p.Description != "" {
		anyChanges = true
		if len(p.Description) > _validationAttrProviderDescriptionMaxLength {
			return errors.New("description exceeds maximum length")
		}
	}

	if p.DocumentURL != "" {
		anyChanges = true
		if len(p.DocumentURL) > _validationAttrProviderDocumentURLMaxLength {
			return errors.New("document URL exceeds maximum length")
		}
		if err := validateURL(p.DocumentURL); err != nil {
			return err
		}
	}

	if p.IconURL != "" {
		anyChanges = true
		if len(p.IconURL) > _validationAttrProviderIconURLMaxLength {
			return errors.New("icon URL exceeds maximum length")
		}
		if err := validateURL(p.IconURL); err != nil {
			return err
		}
	}

	if p.Oauth2Config.AuthURL != "" {
		anyChanges = true
		if err := validateURL(p.Oauth2Config.AuthURL); err != nil {
			return err
		}
	}

	if p.Oauth2Config.TokenURL != "" {
		anyChanges = true
		if err := validateURL(p.Oauth2Config.TokenURL); err != nil {
			return err
		}
	}

	if !anyChanges {
		return erre.Error{
			Code:    erre.ErrorCodeBadRequest,
			Message: "no changes provided for provider update",
		}
	}

	return nil
}

func (c *controller) validateCreateProviderRequest(req entity.CreateProviderRequest) error {
	p := req.Provider
	if p.ApiType == entity.ApiTypeInvalid {
		return errors.New("invalid API type")
	}

	if p.VisibilityType == entity.VisibilityTypeInvalid {
		return errors.New("invalid visibility type")
	}

	if p.BaseURL == "" {
		return errors.New("base URL is required")
	}
	if len(p.BaseURL) > _validationAttrProviderBaseURLMaxLength {
		return errors.New("base URL exceeds maximum length")
	}

	if err := validateURL(p.BaseURL); err != nil {
		return err
	}

	if p.DocumentURL != "" {
		if len(p.DocumentURL) > _validationAttrProviderDocumentURLMaxLength {
			return errors.New("document URL exceeds maximum length")
		}
		if err := validateURL(p.DocumentURL); err != nil {
			return err
		}
	}

	if p.IconURL != "" {
		if len(p.IconURL) > _validationAttrProviderIconURLMaxLength {
			return errors.New("icon URL exceeds maximum length")
		}
		if err := validateURL(p.IconURL); err != nil {
			return err
		}
	}

	if p.Name == "" {
		return errors.New("name is required")
	}
	if len(p.Name) > _validationAttrProviderNameMaxLength {
		return errors.New("name exceeds maximum length")
	}

	if !_regexPatternProviderName.MatchString(p.Name) {
		return errors.New("invalid provider name must be `[a-zA-Z0-9]{1,16}`")
	}

	if p.Description == "" {
		return errors.New("description is required")
	}
	if len(p.Description) > _validationAttrProviderDescriptionMaxLength {
		return errors.New("description exceeds maximum length")
	}

	if p.Oauth2Config.AuthURL != "" {
		if err := validateURL(p.Oauth2Config.AuthURL); err != nil {
			return err
		}
	}

	if p.Oauth2Config.TokenURL != "" {
		if err := validateURL(p.Oauth2Config.TokenURL); err != nil {
			return err
		}
	}

	return nil
}

func buildSecretPrefix(u string) string {
	parsed, _ := url.Parse(u)
	host := parsed.Hostname()
	host = strings.TrimPrefix(host, "www.")
	return strings.ToUpper(strings.Replace(host, ".", "_", -1))
}

func validateURL(u string) error {
	if u == "" {
		return errors.New("URL is required")
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return errors.New("invalid URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("URL must be http or https")
	}
	if parsed.Hostname() == "" {
		return errors.New("URL must have a valid hostname")
	}
	return nil
}
