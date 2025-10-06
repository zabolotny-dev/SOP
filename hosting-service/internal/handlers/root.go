package handlers

import (
	"hosting-service/internal/service"
)

type ApiHandler struct {
	*PlanHandler
	*ServersHandler
}

func NewApiHandler(planService service.PlanService, serverService service.ServerService) *ApiHandler {
	return &ApiHandler{
		PlanHandler:    NewPlansHandler(planService),
		ServersHandler: NewServersHandler(serverService),
	}
}
