package rest

import (
	"context"

	"hosting-contracts/api"
	"hosting-service/cmd/server/rest/handlers/plangrp"
	"hosting-service/cmd/server/rest/handlers/servergrp"
	"hosting-service/internal/plan"
	"hosting-service/internal/server"
)

type API struct {
	Plans   *plangrp.Handlers
	Servers *servergrp.Handlers
}

func New(planBus *plan.Business, serverBus *server.Business) *API {
	return &API{
		Plans:   plangrp.New(planBus),
		Servers: servergrp.New(serverBus),
	}
}

func (a *API) ListPlans(ctx context.Context, request api.ListPlansRequestObject) (api.ListPlansResponseObject, error) {
	return a.Plans.ListPlans(ctx, request)
}

func (a *API) CreatePlan(ctx context.Context, request api.CreatePlanRequestObject) (api.CreatePlanResponseObject, error) {
	return a.Plans.CreatePlan(ctx, request)
}

func (a *API) GetPlanById(ctx context.Context, request api.GetPlanByIdRequestObject) (api.GetPlanByIdResponseObject, error) {
	return a.Plans.GetPlanById(ctx, request)
}

func (a *API) ListServers(ctx context.Context, request api.ListServersRequestObject) (api.ListServersResponseObject, error) {
	return a.Servers.ListServers(ctx, request)
}

func (a *API) OrderServer(ctx context.Context, request api.OrderServerRequestObject) (api.OrderServerResponseObject, error) {
	return a.Servers.OrderServer(ctx, request)
}

func (a *API) GetServerById(ctx context.Context, request api.GetServerByIdRequestObject) (api.GetServerByIdResponseObject, error) {
	return a.Servers.GetServerById(ctx, request)
}

func (a *API) PerformServerAction(ctx context.Context, request api.PerformServerActionRequestObject) (api.PerformServerActionResponseObject, error) {
	return a.Servers.PerformServerAction(ctx, request)
}
