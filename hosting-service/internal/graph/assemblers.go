package graph

import "hosting-service/internal/dto"

func toGraphQLPlan(p dto.PlanPreview) *Plan {
	return &Plan{
		ID:       p.ID.String(),
		Name:     p.Name,
		CPUCores: p.CPUCores,
		RAMMb:    p.RAMMB,
		DiskGb:   p.DiskGB,
	}
}
func toGraphQLServer(s dto.ServerPreview) *Server {
	return &Server{
		ID:        s.ID.String(),
		Name:      s.Name,
		Status:    ServerStatus(s.Status),
		PlanID:    s.PlanID.String(),
		CreatedAt: s.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
