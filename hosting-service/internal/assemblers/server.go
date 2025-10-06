package assemblers

import (
	"fmt"
	"hosting-contracts/api"
	"hosting-service/internal/dto"
)

func ToServer(server dto.ServerPreview) api.Server {
	links := make(api.Links)

	links["self"] = api.Link{Href: to(fmt.Sprintf("/servers/%s", server.ID.String()))}

	actionsHref := to(fmt.Sprintf("/servers/%s/actions", server.ID.String()))

	switch server.Status {
	case "RUNNING":
		links["stop"] = api.Link{Href: actionsHref}
		links["reboot"] = api.Link{Href: actionsHref}

	case "STOPPED":
		links["start"] = api.Link{Href: actionsHref}
		links["delete"] = api.Link{Href: actionsHref}

	case "PENDING", "REBOOTING", "DELETING":
		break
	}

	return api.Server{Id: &server.ID,
		Name:            server.Name,
		PlanId:          server.PlanID,
		Status:          api.ServerStatus(server.Status),
		CreatedAt:       server.CreatedAt,
		UnderscoreLinks: &links,
	}
}

func ToServerCollectionResponse(result dto.ServerSearch, page, pageSize int) api.ServerCollectionResponse {
	embeddedServers := make([]api.Server, len(result.Data))
	for i, p := range result.Data {
		embeddedServers[i] = ToServer(*p)
	}

	collectionLinks := newPaginationLinks("/servers", page, pageSize, result.Meta.TotalPages, result.Meta.HasPrev, result.Meta.HasNext)

	pageMeta := &struct {
		Number        *int   `json:"number,omitempty"`
		Size          *int   `json:"size,omitempty"`
		TotalElements *int64 `json:"totalElements,omitempty"`
		TotalPages    *int   `json:"totalPages,omitempty"`
	}{
		Number:        to(page - 1),
		Size:          to(pageSize),
		TotalElements: to(result.Meta.TotalCount),
		TotalPages:    to(int(result.Meta.TotalPages)),
	}

	return api.ServerCollectionResponse{
		UnderscoreEmbedded: &struct {
			Servers *[]api.Server `json:"servers,omitempty"`
		}{Servers: &embeddedServers},
		UnderscoreLinks: &collectionLinks,
		Page:            pageMeta,
	}
}
