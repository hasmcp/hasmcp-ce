package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	crudctrl "github.com/hasmcp/hasmcp-ce/backend/internal/controller/crud"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	view "github.com/hasmcp/hasmcp-ce/backend/internal/data/view/api"
	"github.com/mustafaturan/monoflake"
)

type serverToolResponseCRUD struct {
	crudctrl.Controller
	request  entity.CreateServerToolRequest
	response *entity.CreateServerToolResponse
}

func (c *serverToolResponseCRUD) CreateServerTool(
	_ context.Context, request entity.CreateServerToolRequest) (*entity.CreateServerToolResponse, error) {
	c.request = request
	return c.response, nil
}

func TestCreateServerToolHTTPResponseIncludesAssociationIDs(t *testing.T) {
	const (
		serverID   int64 = 101
		providerID int64 = 202
		toolID     int64 = 303
	)

	crud := &serverToolResponseCRUD{
		response: &entity.CreateServerToolResponse{
			Tool: entity.ServerTool{ServerID: serverID, ProviderID: providerID, ToolID: toolID},
		},
	}
	h := &handler{crud: crud}
	app := fiber.New()
	app.Post("/api/v1/servers/:id/tools", h.createServerTool())

	requestBody, err := json.Marshal(view.CreateServerToolRequest{
		Tool: view.ServerTool{ToolID: monoflake.ID(toolID).String()},
	})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/servers/"+monoflake.ID(serverID).String()+"/tools",
		bytes.NewReader(requestBody),
	)
	request.Header.Set(headerContentType, headerContentTypeValueApplicationJSON)

	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusCreated)
	}

	wantRequest := entity.CreateServerToolRequest{
		Tool: entity.ServerTool{ServerID: serverID, ToolID: toolID},
	}
	if crud.request != wantRequest {
		t.Fatalf("controller request = %#v, want %#v", crud.request, wantRequest)
	}

	var body view.CreateServerToolResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	wantTool := view.ServerTool{
		ServerID:   monoflake.ID(serverID).String(),
		ProviderID: monoflake.ID(providerID).String(),
		ToolID:     monoflake.ID(toolID).String(),
	}
	if body.Tool != wantTool {
		t.Fatalf("response tool = %#v, want %#v", body.Tool, wantTool)
	}
}
