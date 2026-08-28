package crud

import (
	"context"
	"testing"

	cachectrl "github.com/hasmcp/hasmcp-ce/backend/internal/controller/cache"
	mcpctrl "github.com/hasmcp/hasmcp-ce/backend/internal/controller/mcp"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	"github.com/hasmcp/hasmcp-ce/backend/internal/data/model"
	storagepkg "github.com/hasmcp/hasmcp-ce/backend/internal/repository/storage"
)

type serverToolResponseStorage struct {
	storagepkg.Repository
	providerID int64
	added      model.ServerTool
}

func (s *serverToolResponseStorage) GetServer(_ context.Context, id int64) (*model.Server, error) {
	return &model.Server{ID: id}, nil
}

func (s *serverToolResponseStorage) GetProviderTool(_ context.Context, id int64) (*model.ProviderTool, error) {
	return &model.ProviderTool{ID: id, ProviderID: s.providerID}, nil
}

func (s *serverToolResponseStorage) AddToolToServer(_ context.Context, tool model.ServerTool) error {
	s.added = tool
	return nil
}

type serverToolResponseCache struct {
	cachectrl.Controller
	objectType entity.ObjectType
	id         int64
}

func (c *serverToolResponseCache) Evict(_ context.Context, objectType entity.ObjectType, id int64) {
	c.objectType = objectType
	c.id = id
}

type serverToolResponseMCP struct {
	mcpctrl.Controller
	change entity.ResourceChange
}

func (m *serverToolResponseMCP) HandleChanges(_ context.Context, change entity.ResourceChange) error {
	m.change = change
	return nil
}

func TestCreateServerToolReturnsPersistedAssociation(t *testing.T) {
	const (
		serverID   int64 = 101
		providerID int64 = 202
		toolID     int64 = 303
	)

	store := &serverToolResponseStorage{providerID: providerID}
	cache := &serverToolResponseCache{}
	mcp := &serverToolResponseMCP{}
	controller := &controller{storage: store, cache: cache, mcp: mcp}

	response, err := controller.CreateServerTool(context.Background(), entity.CreateServerToolRequest{
		Tool: entity.ServerTool{ServerID: serverID, ToolID: toolID},
	})
	if err != nil {
		t.Fatal(err)
	}

	wantTool := entity.ServerTool{ServerID: serverID, ProviderID: providerID, ToolID: toolID}
	if response == nil || response.Tool != wantTool {
		t.Fatalf("created tool = %#v, want %#v", response, wantTool)
	}
	if want := (model.ServerTool{ServerID: serverID, ProviderID: providerID, ToolID: toolID}); store.added != want {
		t.Fatalf("persisted tool = %#v, want %#v", store.added, want)
	}
	if cache.objectType != entity.ObjectTypeServer || cache.id != serverID {
		t.Fatalf("cache eviction = (%v, %d), want (%v, %d)", cache.objectType, cache.id, entity.ObjectTypeServer, serverID)
	}
	wantChange := entity.ResourceChange{
		ObjectType:      entity.ObjectTypeServerTool,
		EventType:       entity.ObjectEventTypeUpdate,
		ResoureID:       toolID,
		ResourceOwnerID: serverID,
	}
	if mcp.change != wantChange {
		t.Fatalf("MCP change = %#v, want %#v", mcp.change, wantChange)
	}
}
