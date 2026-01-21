package rest

import (
	"hosting-service/cmd/server/rest/handlers/plangrp"
	"hosting-service/cmd/server/rest/handlers/rootgrp"
	"hosting-service/cmd/server/rest/handlers/servergrp"
	"hosting-service/internal/plan"
	"hosting-service/internal/server"
)

type API struct {
	*plangrp.PlanHandlers
	*servergrp.ServerHandlers
	*rootgrp.RootHandlers
}

func New(planBus plan.ExtBusiness, serverBus server.ExtBusiness, prefix string) *API {
	return &API{
		PlanHandlers:   plangrp.New(planBus, prefix),
		ServerHandlers: servergrp.New(serverBus, prefix),
		RootHandlers:   rootgrp.New(prefix),
	}
}
