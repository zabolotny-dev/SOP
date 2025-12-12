package servergrp

import (
	"fmt"
	"hosting-contracts/api"
	"hosting-service/cmd/server/rest/pagination"
	"hosting-service/internal/platform/page"
	"hosting-service/internal/server"
)

func toServer(s server.Server) api.Server {
	links := make(api.Links)

	selfLink := fmt.Sprintf("/api/servers/%s", s.ID)
	actionsLink := fmt.Sprintf("/api/servers/%s/actions", s.ID)

	links["self"] = api.Link{Href: selfLink}

	switch s.Status {
	case server.StatusRunning:
		links["stop"] = api.Link{Href: actionsLink}
	case server.StatusStopped:
		links["start"] = api.Link{Href: actionsLink}
		links["delete"] = api.Link{Href: actionsLink}
	}

	return api.Server{
		Id:              s.ID,
		Name:            s.Name,
		PlanId:          s.PlanID,
		IPv4Address:     s.IPv4Address,
		Status:          api.ServerStatus(s.Status),
		CreatedAt:       s.CreatedAt,
		UnderscoreLinks: links,
	}
}

func toServerCollectionResponse(servers []server.Server, pg page.Page, total int) api.ServerCollectionResponse {
	items := make([]api.Server, len(servers))
	for i, s := range servers {
		items[i] = toServer(s)
	}

	return api.ServerCollectionResponse{
		UnderscoreEmbedded: &struct {
			Servers *[]api.Server `json:"servers,omitempty"`
		}{
			Servers: &items,
		},
		Page:            pagination.ToMetaData(pg, total),
		UnderscoreLinks: pagination.ToLinks("/api/servers", pg, total),
	}
}
