package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	view "github.com/hasmcp/hasmcp-ce/backend/internal/data/view/api"
	"github.com/mustafaturan/monoflake"
)

func FromHTTPRequestToCreateServerResourceAssociationRequestEntity(c *fiber.Ctx) *entity.CreateServerResourceRequest {
	var payload view.CreateServerResourceRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}
	data := payload.Resource
	data.ServerID = c.Params("id")

	return &entity.CreateServerResourceRequest{
		Resource: FromServerResourceViewToServerResourceEntity(data),
	}
}

func FromServerResourceViewToServerResourceEntity(r view.ServerResource) entity.ServerResource {
	return entity.ServerResource{
		ServerID:   monoflake.IDFromBase62(r.ServerID).Int64(),
		ResourceID: monoflake.IDFromBase62(r.ResourceID).Int64(),
	}
}

func FromServerResourceEntityToServerResourceView(r entity.ServerResource) view.ServerResource {
	return view.ServerResource{
		ServerID:   monoflake.ID(r.ServerID).String(),
		ResourceID: monoflake.ID(r.ResourceID).String(),
	}
}

func FromServerResourceEntitiesToServerResourceViews(rs []entity.ServerResource) []view.ServerResource {
	resources := make([]view.ServerResource, len(rs))
	for i, r := range rs {
		resources[i] = FromServerResourceEntityToServerResourceView(r)
	}
	return resources
}

func FromCreateServerResourceAssociationResponseEntityToHTTPResponse(res *entity.CreateServerResourceResponse) []byte {
	resp := view.CreateServerResourceResponse{
		Resource: FromServerResourceEntityToServerResourceView(res.Resoure),
	}

	payload, _ := json.Marshal(resp)
	return payload
}

func FromHTTPRequestToListServerResourcesRequestEntity(c *fiber.Ctx) *entity.ListServerResourcesRequest {
	serverIDParam := c.Params("id")
	if serverIDParam == "" {
		return nil
	}
	return &entity.ListServerResourcesRequest{
		ServerID: monoflake.IDFromBase62(serverIDParam).Int64(),
	}
}

func FromListServerResourcesResponseEntityToHTTPResponse(res *entity.ListServerResourcesResponse) []byte {
	payload, _ := json.Marshal(view.ListServerResourcesResponse{
		Resources: FromServerResourceEntitiesToServerResourceViews(res.Resources),
	})
	return payload
}

func FromHTTPRequestToDeleteServerResourceAssociationRequestEntity(c *fiber.Ctx) *entity.DeleteServerResourceRequest {
	serverIDParam := c.Params("id")
	resourceIDParam := c.Params("resourceID")
	if serverIDParam == "" || resourceIDParam == "" {
		return nil
	}

	return &entity.DeleteServerResourceRequest{
		ServerID:   monoflake.IDFromBase62(serverIDParam).Int64(),
		ResourceID: monoflake.IDFromBase62(resourceIDParam).Int64(),
	}
}
