package rest

import (
	"context"
	"hosting-service/cmd/server/rest/gen"
	"hosting-service/cmd/server/rest/handlers/plangrp"
	"hosting-service/cmd/server/rest/handlers/rootgrp"
	"hosting-service/cmd/server/rest/handlers/servergrp"
	"hosting-service/internal/plan"
	"hosting-service/internal/server"
)

type API struct {
	plans   *plangrp.Handlers
	servers *servergrp.Handlers
	root    *rootgrp.Handlers
}

func New(planBus plan.ExtBusiness, serverBus server.ExtBusiness, prefix string) *API {
	return &API{
		plans:   plangrp.New(planBus, prefix),
		servers: servergrp.New(serverBus, prefix),
		root:    rootgrp.New(prefix),
	}
}

func (a *API) ListPlans(ctx context.Context, request gen.ListPlansRequestObject) (gen.ListPlansResponseObject, error) {
	return a.plans.ListPlans(ctx, request)
}

func (a *API) CreatePlan(ctx context.Context, request gen.CreatePlanRequestObject) (gen.CreatePlanResponseObject, error) {
	return a.plans.CreatePlan(ctx, request)
}

func (a *API) GetPlanById(ctx context.Context, request gen.GetPlanByIdRequestObject) (gen.GetPlanByIdResponseObject, error) {
	return a.plans.GetPlanById(ctx, request)
}

func (a *API) ListServers(ctx context.Context, request gen.ListServersRequestObject) (gen.ListServersResponseObject, error) {
	return a.servers.ListServers(ctx, request)
}

func (a *API) OrderServer(ctx context.Context, request gen.OrderServerRequestObject) (gen.OrderServerResponseObject, error) {
	return a.servers.OrderServer(ctx, request)
}

func (a *API) GetServerById(ctx context.Context, request gen.GetServerByIdRequestObject) (gen.GetServerByIdResponseObject, error) {
	return a.servers.GetServerById(ctx, request)
}

func (a *API) PerformServerAction(ctx context.Context, request gen.PerformServerActionRequestObject) (gen.PerformServerActionResponseObject, error) {
	return a.servers.PerformServerAction(ctx, request)
}

func (a *API) GetRoot(ctx context.Context, request gen.GetRootRequestObject) (gen.GetRootResponseObject, error) {
	return a.root.GetRoot(ctx, request)
}
