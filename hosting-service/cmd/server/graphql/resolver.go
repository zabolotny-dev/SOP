package graph

import (
	"hosting-service/internal/plan"
	"hosting-service/internal/server"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	PlanBus   *plan.Business
	ServerBus *server.Business
}
