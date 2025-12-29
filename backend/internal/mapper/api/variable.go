package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	view "github.com/hasmcp/hasmcp-ce/backend/internal/data/view/api"
	"github.com/mustafaturan/monoflake"
)

func FromHTTPRequestToCreateVariableRequestEntity(c *fiber.Ctx) *entity.CreateVariableRequest {
	var payload view.CreateVariableRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}

	return &entity.CreateVariableRequest{
		Variable: FromVariableViewToVariableEntity(payload.Variable),
	}
}

func FromVariableViewToVariableEntity(v view.Variable) entity.Variable {
	return entity.Variable{
		ID:    monoflake.IDFromBase62(v.ID).Int64(),
		Type:  entity.StringToVariableType(v.Type),
		Name:  v.Name,
		Value: []byte(v.Value),
	}
}

func FromHTTPRequestToUpdateVariableRequestEntity(c *fiber.Ctx) *entity.UpdateVariableRequest {
	id := monoflake.IDFromBase62(c.Params("id")).Int64()
	if id <= 0 {
		return nil
	}
	var payload view.UpdateVariableRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}

	variable := FromVariableViewToVariableEntity(payload.Variable)
	variable.ID = id

	return &entity.UpdateVariableRequest{
		Variable: variable,
	}
}

func FromUpdateVariableResponseToHTTPResponse(rs entity.UpdateVariableResponse) []byte {
	payload := view.UpdateVariableRequest{
		Variable: FromVariableEntityToVariableView(rs.Variable),
	}

	res, _ := json.Marshal(payload)
	return res
}

func FromHTTPRequestToDeleteVariableRequestEntity(c *fiber.Ctx) *entity.DeleteVariableRequest {
	return &entity.DeleteVariableRequest{
		ID: monoflake.IDFromBase62(c.Params("id")).Int64(),
	}
}

func FromCreateVariableResponseEntityToHTTPResponse(rs *entity.CreateVariableResponse) []byte {
	payload := view.CreateVariableResponse{
		Variable: FromVariableEntityToVariableView(rs.Variable),
	}
	res, _ := json.Marshal(payload)
	return res
}

func FromUpdateVariableResponseEntityToHTTPResponse(rs *entity.UpdateVariableResponse) []byte {
	payload := view.UpdateVariableResponse{
		Variable: FromVariableEntityToVariableView(rs.Variable),
	}
	res, _ := json.Marshal(payload)
	return res
}

func FromVariableEntitiesToVariableViews(variables []entity.Variable) []view.Variable {
	vars := make([]view.Variable, len(variables))
	for i, v := range variables {
		vars[i] = FromVariableEntityToVariableView(v)
	}
	return vars
}

func FromVariableEntityToVariableView(v entity.Variable) view.Variable {
	value := string(v.Value)
	if v.Type == entity.VariableTypeSecret {
		value = "***"
	}
	return view.Variable{
		ID:    monoflake.ID(v.ID).String(),
		Type:  v.Type.String(),
		Name:  v.Name,
		Value: value,
	}
}

func FromListVariablesResponseEntityToHTTPResponse(rs *entity.ListVariablesResponse) []byte {
	payload := view.ListVariablesResponse{
		Variables: FromVariableEntitiesToVariableViews(rs.Variables),
	}
	res, _ := json.Marshal(payload)
	return res
}
