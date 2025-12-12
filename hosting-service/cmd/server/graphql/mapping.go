package graph

import (
	"hosting-service/internal/plan"
	"hosting-service/internal/platform/page"
	"hosting-service/internal/server"
)

func toPlan(p plan.Plan) *Plan {
	return &Plan{
		ID:       p.ID.String(),
		Name:     p.Name,
		CPUCores: p.CPUCores,
		RAMMb:    p.RAMMB,
		DiskGb:   p.DiskGB,
	}
}

func toServer(s server.Server) *Server {
	return &Server{
		ID:          s.ID.String(),
		Name:        s.Name,
		Status:      ServerStatus(s.Status),
		PlanID:      s.PlanID.String(),
		IPv4Address: s.IPv4Address,
		CreatedAt:   s.CreatedAt.String(),
	}
}

func toPlanCollection(plans []plan.Plan, p page.Page, count int) *PlanCollection {
	items := make([]*Plan, len(plans))
	for i, p := range plans {
		items[i] = toPlan(p)
	}

	doc := page.NewDocument(p, count)

	return &PlanCollection{
		Plans: items,
		Meta: &CollectionMeta{
			Number:        doc.Page,
			Size:          doc.PageSize,
			TotalElements: doc.TotalCount,
			TotalPages:    doc.TotalPages,
			HasNextPage:   doc.HasNext,
			HasPrevPage:   doc.HasPrev,
		},
	}
}

func toServerCollection(servers []server.Server, p page.Page, count int) *ServerCollection {
	items := make([]*Server, len(servers))
	for i, s := range servers {
		items[i] = toServer(s)
	}

	doc := page.NewDocument(p, count)

	return &ServerCollection{
		Servers: items,
		Meta: &CollectionMeta{
			Number:        doc.Page,
			Size:          doc.PageSize,
			TotalElements: doc.TotalCount,
			TotalPages:    doc.TotalPages,
			HasNextPage:   doc.HasNext,
			HasPrevPage:   doc.HasPrev,
		},
	}
}
