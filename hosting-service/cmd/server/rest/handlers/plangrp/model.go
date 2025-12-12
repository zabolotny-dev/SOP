package plangrp

import (
	"fmt"
	"hosting-contracts/api"
	"hosting-service/cmd/server/rest/pagination"
	"hosting-service/internal/plan"
	"hosting-service/internal/platform/page"
)

func toPlan(p plan.Plan) api.ServerPlan {
	links := api.Links{
		"self": api.Link{Href: fmt.Sprintf("/api/plans/%s", p.ID)},
	}

	return api.ServerPlan{
		Id:              p.ID,
		Name:            p.Name,
		CpuCores:        p.CPUCores,
		RamMb:           p.RAMMB,
		DiskGb:          p.DiskGB,
		UnderscoreLinks: links,
	}
}

func toPlanCollectionResponse(plans []plan.Plan, pg page.Page, total int) api.PlanCollectionResponse {
	items := make([]api.ServerPlan, len(plans))
	for i, p := range plans {
		items[i] = toPlan(p)
	}

	return api.PlanCollectionResponse{
		UnderscoreEmbedded: &struct {
			Plans *[]api.ServerPlan `json:"plans,omitempty"`
		}{
			Plans: &items,
		},
		Page:            pagination.ToMetaData(pg, total),
		UnderscoreLinks: pagination.ToLinks("/api/plans", pg, total),
	}
}
