package crud

import (
	"context"
	"encoding/hex"
	"errors"
	"regexp"
	"time"

	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	erre "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/err"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/locksmith"
)

type VariableController interface {
	CreateVariable(ctx context.Context, req entity.CreateVariableRequest) (*entity.CreateVariableResponse, error)
	SaveVariable(ctx context.Context, req entity.SaveVariableRequest) error
	ListVariables(ctx context.Context) (*entity.ListVariablesResponse, error)
	DeleteVariable(ctx context.Context, req entity.DeleteVariableRequest) error
	UpdateVariable(ctx context.Context, req entity.UpdateVariableRequest) (*entity.UpdateVariableResponse, error)
}

const (
	_validationAttrVariableNameMaxLength = 128
)

var (
	_regexPatternVariableName = regexp.MustCompile(`^[A-Z0-9_]{1,128}$`)
)

func (c *controller) CreateVariable(ctx context.Context, req entity.CreateVariableRequest) (*entity.CreateVariableResponse, error) {
	if err := c.validateCreateVariableRequest(req); err != nil {
		return nil, erre.Error{
			Code:    400,
			Message: err.Error(),
		}
	}

	var nonce string
	variable := req.Variable
	val := string(variable.Value)
	if variable.Type == entity.VariableTypeSecret {
		res, err := c.locksmith.Encrypt(ctx, &locksmith.EncryptRequest{
			Plaintext: []byte(val),
		})
		if err != nil {
			return nil, erre.Error{
				Code:    erre.ErrorCodeInternalServerError,
				Message: "failed to encrypt secret",
				Data: map[string]any{
					"reason": err.Error(),
				},
			}
		}
		val = hex.EncodeToString(res.Ciphertext)
		nonce = hex.EncodeToString(res.Nonce)
	}
	v := model.Variable{
		ID:    c.idgen.Next(),
		Type:  uint8(variable.Type),
		Name:  variable.Name,
		Value: val,
		Nonce: nonce,
	}
	err := c.storage.CreateVariable(ctx, v)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to create variable",
			Data: map[string]any{
				"reason": err.Error(),
			},
		}
	}

	return &entity.CreateVariableResponse{
		Variable: entity.Variable{
			ID:    v.ID,
			Type:  entity.VariableType(v.Type),
			Value: variable.Value,
			Name:  v.Name,
		},
	}, nil
}

func (c *controller) ListVariables(ctx context.Context) (*entity.ListVariablesResponse, error) {
	vars, err := c.storage.ListVariables(ctx)
	if err != nil {
		return nil, err
	}

	variables := make([]entity.Variable, 0, len(vars))
	for _, v := range vars {
		val := []byte(v.Value)
		if entity.VariableType(v.Type) == entity.VariableTypeSecret {
			val, err = hex.DecodeString(v.Value)
			if err != nil {
				return nil, err
			}
			nonce, err := hex.DecodeString(v.Nonce)
			if err != nil {
				return nil, err
			}
			res, err := c.locksmith.Decrypt(ctx, &locksmith.DecryptRequest{
				Ciphertext: val,
				Nonce:      nonce,
			})
			if err != nil {
				return nil, err
			}
			val = res.Plaintext
		}

		variables = append(variables, entity.Variable{
			ID:        v.ID,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			Type:      entity.VariableType(v.Type),
			Name:      v.Name,
			Value:     val,
		})
	}

	return &entity.ListVariablesResponse{
		Variables: variables,
	}, nil
}

func (c *controller) DeleteVariable(ctx context.Context, req entity.DeleteVariableRequest) error {
	defer c.cache.Evict(ctx, entity.ObjectTypeVariable, req.ID)

	err := c.storage.DeleteVariable(ctx, req.ID)
	if err != nil {
		return err
	}
	return nil
}

func (c *controller) UpdateVariable(ctx context.Context, req entity.UpdateVariableRequest) (*entity.UpdateVariableResponse, error) {
	defer c.cache.Evict(ctx, entity.ObjectTypeVariable, req.Variable.ID)
	var nonce []byte

	variable := req.Variable
	// Fetch existing variable to check type
	v, err := c.storage.GetVariable(ctx, variable.ID)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeNotFound,
			Message: "failed to fetch updated variable",
			Data: map[string]any{
				"reason":     err.Error(),
				"variableId": variable.ID,
			},
		}
	}

	attrs := map[model.VariableAttribute]any{
		model.VariableAttributeValue:     variable.Value,
		model.VariableAttributeUpdatedAt: time.Now().UTC(),
	}

	if entity.VariableType(v.Type) == entity.VariableTypeSecret {
		val := []byte(variable.Value)
		res, err := c.locksmith.Encrypt(ctx, &locksmith.EncryptRequest{
			Plaintext: val,
		})
		if err != nil {
			return nil, err
		}
		val = res.Ciphertext
		nonce = res.Nonce
		attrs[model.VariableAttributeValue] = hex.EncodeToString(val)
		attrs[model.VariableAttributeNonce] = hex.EncodeToString(nonce)
	}

	err = c.storage.UpdateVariable(ctx, variable.ID, attrs)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to update variable",
			Data: map[string]any{
				"reason":     err.Error(),
				"variableId": variable.ID,
				"attrs":      attrs,
			},
		}
	}

	v, err = c.storage.GetVariable(ctx, variable.ID)
	if err != nil {
		return nil, erre.Error{
			Code:    erre.ErrorCodeInternalServerError,
			Message: "failed to fetch updated variable",
			Data: map[string]any{
				"reason":     err.Error(),
				"variableId": variable.ID,
				"attrs":      attrs,
			},
		}
	}
	val := []byte(v.Value)
	if entity.VariableType(v.Type) == entity.VariableTypeSecret {
		val, err = hex.DecodeString(v.Value)
		if err != nil {
			return nil, err
		}
		nonce, err := hex.DecodeString(v.Nonce)
		if err != nil {
			return nil, err
		}
		res, err := c.locksmith.Decrypt(ctx, &locksmith.DecryptRequest{
			Ciphertext: val,
			Nonce:      nonce,
		})
		if err != nil {
			return nil, err
		}
		val = res.Plaintext
	}

	return &entity.UpdateVariableResponse{
		Variable: entity.Variable{
			ID:        v.ID,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			Type:      entity.VariableType(v.Type),
			Name:      v.Name,
			Value:     val,
		},
	}, nil
}

func (c *controller) SaveVariable(ctx context.Context, req entity.SaveVariableRequest) error {
	currentVariable, err := c.storage.GetVariableByName(ctx, req.Variable.Name)
	if err == nil && currentVariable != nil && currentVariable.ID != 0 {
		defer c.cache.Evict(ctx, entity.ObjectTypeVariable, currentVariable.ID)
	}

	if err := c.validateSaveVariableRequest(req); err != nil {
		return erre.Error{
			Code:    400,
			Message: err.Error(),
		}
	}

	var nonce string
	variable := req.Variable
	val := string(variable.Value)
	if variable.Type == entity.VariableTypeSecret {
		res, err := c.locksmith.Encrypt(ctx, &locksmith.EncryptRequest{
			Plaintext: []byte(val),
		})
		if err != nil {
			return erre.Error{
				Code:    500,
				Message: err.Error(),
			}
		}
		val = hex.EncodeToString(res.Ciphertext)
		nonce = hex.EncodeToString(res.Nonce)
	}
	now := time.Now().UTC()
	v := model.Variable{
		ID:        c.idgen.Next(),
		Type:      uint8(variable.Type),
		Name:      variable.Name,
		Value:     val,
		Nonce:     nonce,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = c.storage.SaveVariable(ctx, v)
	if err != nil {
		return erre.Error{
			Code:    500,
			Message: err.Error(),
		}
	}

	return nil
}

func (c *controller) validateCreateVariableRequest(req entity.CreateVariableRequest) error {
	return c.validateVariable(req.Variable)
}

func (c *controller) validateSaveVariableRequest(req entity.SaveVariableRequest) error {
	return c.validateVariable(req.Variable)
}

func (c *controller) validateVariable(v entity.Variable) error {
	if !_regexPatternVariableName.MatchString(v.Name) {
		return errors.New("invalid varible name must be `[A-Z0-9_]{1,128}`")
	}

	if v.Type != entity.VariableTypeEnv && v.Type != entity.VariableTypeSecret {
		return errors.New("invalid variable type")
	}
	if v.Name == "" {
		return errors.New("name is required")
	}
	if len(v.Name) > _validationAttrVariableNameMaxLength {
		return errors.New("name exceeds maximum length")
	}
	if len(v.Value) == 0 {
		return errors.New("value is required")
	}
	return nil
}
