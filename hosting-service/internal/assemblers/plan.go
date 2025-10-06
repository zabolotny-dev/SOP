package assemblers

import (
	"fmt"
	"hosting-contracts/api"
	"hosting-service/internal/dto"
)

func ToPlan(plan dto.PlanPreview) api.ServerPlan {
	links := make(api.Links)

	links["self"] = api.Link{Href: to(fmt.Sprintf("/plans/%s", plan.ID))}

	return api.ServerPlan{
		Id:              &plan.ID,
		Name:            plan.Name,
		CpuCores:        plan.CPUCores,
		RamMb:           plan.RAMMB,
		DiskGb:          plan.DiskGB,
		UnderscoreLinks: &links,
	}
}

func ToPlanCollectionResponse(result dto.PlanSearch, page, pageSize int) api.PlanCollectionResponse {
	embeddedPlans := make([]api.ServerPlan, len(result.Data))
	for i, p := range result.Data {
		embeddedPlans[i] = ToPlan(*p)
	}

	collectionLinks := newPaginationLinks("/plans", page, pageSize, result.Meta.TotalPages, result.Meta.HasPrev, result.Meta.HasNext)

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

	return api.PlanCollectionResponse{
		UnderscoreEmbedded: &struct {
			Plans *[]api.ServerPlan `json:"plans,omitempty"`
		}{Plans: &embeddedPlans},
		UnderscoreLinks: &collectionLinks,
		Page:            pageMeta,
	}
}
