package graph

import "hosting-service/internal/service"

//go:generate go run github.com/99designs/gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	PlanService   service.PlanService
	ServerService service.ServerService
}
