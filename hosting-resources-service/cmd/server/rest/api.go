package rest

import (
	"context"
	"hosting-resources-service/cmd/server/rest/gen"
	"hosting-resources-service/cmd/server/rest/handlers/poolgrp"
	"hosting-resources-service/cmd/server/rest/handlers/rootgrp"
	"hosting-resources-service/internal/pool"
)

type API struct {
	pools *poolgrp.Handlers
	root  *rootgrp.Handlers
}

func New(poolBus pool.ExtBusiness, prefix string) *API {
	return &API{
		pools: poolgrp.New(poolBus, prefix),
		root:  rootgrp.New(prefix),
	}
}

func (a *API) ListPools(ctx context.Context, request gen.ListPoolsRequestObject) (gen.ListPoolsResponseObject, error) {
	return a.pools.ListPools(ctx, request)
}

func (a *API) CreatePool(ctx context.Context, request gen.CreatePoolRequestObject) (gen.CreatePoolResponseObject, error) {
	return a.pools.CreatePool(ctx, request)
}

func (a *API) AddResources(ctx context.Context, request gen.AddResourcesRequestObject) (gen.AddResourcesResponseObject, error) {
	return a.pools.AddResources(ctx, request)
}

func (a *API) GetRoot(ctx context.Context, request gen.GetRootRequestObject) (gen.GetRootResponseObject, error) {
	return a.root.GetRoot(ctx, request)
}
