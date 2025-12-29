package api

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	entity "github.com/hasmcp/hasmcp-ce/backend/internal/data/entity/crud"
	view "github.com/hasmcp/hasmcp-ce/backend/internal/data/view/api"
	"github.com/mustafaturan/monoflake"
)

func FromHTTPRequestToCreateResourceRequestEntity(c *fiber.Ctx) *entity.CreateResourceRequest {
	var payload view.CreateResourceRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}

	data := payload.Resource

	return &entity.CreateResourceRequest{
		Resource: FromResoureViewToResourceEntity(data),
	}
}

func FromResoureViewToResourceEntity(r view.Resource) entity.Resource {
	return entity.Resource{
		Name:        r.Name,
		Description: r.Description,
		URI:         r.URI,
		MimeType:    r.MimeType,
		Size:        r.Size,
		Annotations: r.Annotations,
	}
}

func FromCreateResourceResponseEntityToHTTPResponse(res *entity.CreateResourceResponse) []byte {
	r := res.Resource
	payload, _ := json.Marshal(view.CreateResourceResponse{
		Resource: FromResourceEntityToResourceView(r),
	})
	return payload
}

func FromResourceEntitiesToResourceViews(resources []entity.Resource) []view.Resource {
	v := make([]view.Resource, len(resources))
	for i, r := range resources {
		v[i] = FromResourceEntityToResourceView(r)
	}
	return v
}

func FromResourceEntityToResourceView(r entity.Resource) view.Resource {
	return view.Resource{
		ID:          monoflake.ID(r.ID).String(),
		CreatedAt:   FromTimeToRFC3339String(r.CreatedAt),
		UpdatedAt:   FromTimeToRFC3339String(r.UpdatedAt),
		Name:        r.Name,
		Description: r.Description,
		URI:         r.URI,
		MimeType:    r.MimeType,
		Size:        r.Size,
		Annotations: r.Annotations,
	}
}

func FromHTTPRequestToListResourcesRequestEntity(c *fiber.Ctx) *entity.ListResourcesRequest {
	return &entity.ListResourcesRequest{}
}

func FromListResourcesResponseEntityToHTTPResponse(res *entity.ListResourcesResponse) []byte {
	payload, _ := json.Marshal(view.ListResourcesResponse{
		Resources: FromResourceEntitiesToResourceViews(res.Resources),
	})
	return payload
}

func FromHTTPRequestToGetResourceRequestEntity(c *fiber.Ctx) *entity.GetResourceRequest {
	id := monoflake.IDFromBase62(c.Params("id")).Int64()
	if id == 0 {
		return nil
	}
	return &entity.GetResourceRequest{ID: id}
}

func FromGetResourceResponseEntityToHTTPResponse(res *entity.GetResourceResponse) []byte {
	payload, _ := json.Marshal(view.GetResourceResponse{
		Resource: FromResourceEntityToResourceView(res.Resource),
	})
	return payload
}

func FromHTTPRequestToUpdateResourceRequestEntity(c *fiber.Ctx) *entity.UpdateResourceRequest {
	id := monoflake.IDFromBase62(c.Params("id")).Int64()
	if id <= 0 {
		return nil
	}

	var payload view.UpdateResourceRequest
	if err := json.Unmarshal(c.BodyRaw(), &payload); err != nil {
		return nil
	}

	data := payload.Resource
	resource := FromResoureViewToResourceEntity(data)
	resource.ID = id

	return &entity.UpdateResourceRequest{
		Resource: resource,
	}
}

func FromUpdateResourceResponseEntityToHTTPResponse(res *entity.UpdateResourceResponse) []byte {
	payload, _ := json.Marshal(view.UpdateResourceResponse{
		Resource: FromResourceEntityToResourceView(res.Resource),
	})
	return payload
}

func FromHTTPRequestToDeleteResourceRequestEntity(c *fiber.Ctx) *entity.DeleteResourceRequest {
	id := monoflake.IDFromBase62(c.Params("id")).Int64()
	if id == 0 {
		return nil
	}
	return &entity.DeleteResourceRequest{ID: id}
}
