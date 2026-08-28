package api

import (
	"encoding/json"
	"testing"

	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	view "github.com/hasmcp/hasmcp-ce/backend/internal/data/view/api"
	"github.com/mustafaturan/monoflake"
)

func TestCreateServerToolResponseIncludesAssociationIDs(t *testing.T) {
	const (
		serverID   int64 = 101
		providerID int64 = 202
		toolID     int64 = 303
	)

	payload := FromCreateServerToolResponseEntityToHTTPResponse(&entity.CreateServerToolResponse{
		Tool: entity.ServerTool{ServerID: serverID, ProviderID: providerID, ToolID: toolID},
	})

	var response view.CreateServerToolResponse
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatal(err)
	}
	want := view.ServerTool{
		ServerID:   monoflake.ID(serverID).String(),
		ProviderID: monoflake.ID(providerID).String(),
		ToolID:     monoflake.ID(toolID).String(),
	}
	if response.Tool != want {
		t.Fatalf("mapped tool = %#v, want %#v", response.Tool, want)
	}
}
