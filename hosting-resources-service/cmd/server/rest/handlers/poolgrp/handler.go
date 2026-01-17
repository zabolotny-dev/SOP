package poolgrp

import (
	"context"
	"errors"
	"hosting-kit/page"
	"hosting-resources-service/cmd/server/rest/gen"
	"hosting-resources-service/internal/pool"
)

type Handlers struct {
	poolBus pool.ExtBusiness
	prefix  string
}

func New(poolBus pool.ExtBusiness, prefix string) *Handlers {
	return &Handlers{
		poolBus: poolBus,
		prefix:  prefix,
	}
}

func (h *Handlers) ListPools(ctx context.Context, request gen.ListPoolsRequestObject) (gen.ListPoolsResponseObject, error) {
	pageNum := 1
	pageSize := 10

	if request.Params.Page != nil {
		pageNum = *request.Params.Page
	}
	if request.Params.PageSize != nil {
		pageSize = *request.Params.PageSize
	}

	page := page.Parse(pageNum, pageSize)

	pools, count, err := h.poolBus.Search(ctx, page)

	if err != nil {
		return nil, err
	}

	return gen.ListPools200ApplicationHalPlusJSONResponse(toPoolCollectionResponse(pools, page, count, h.prefix)), nil
}

func (h *Handlers) CreatePool(ctx context.Context, request gen.CreatePoolRequestObject) (gen.CreatePoolResponseObject, error) {
	newPool, err := h.poolBus.CreatePool(ctx, pool.NewPool{
		Name:     request.Body.Name,
		CPUCores: request.Body.CpuCores,
		RAMMB:    request.Body.RamMb,
		DiskGB:   request.Body.DiskGb,
		IPCount:  request.Body.IpCount,
	})

	if err != nil {
		if errors.Is(err, pool.ErrValidation) {
			return gen.CreatePool400JSONResponse{
				BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: err.Error()},
			}, nil
		}
		return nil, err
	}

	return gen.CreatePool201ApplicationHalPlusJSONResponse(toPool(newPool, h.prefix)), nil
}

func (h *Handlers) AddResources(ctx context.Context, request gen.AddResourcesRequestObject) (gen.AddResourcesResponseObject, error) {
	newPool, err := h.poolBus.AddResources(ctx, pool.Resource{
		CPUCores: request.Body.CpuCores,
		RAMMB:    request.Body.RamMb,
		DiskGB:   request.Body.DiskGb,
		IPCount:  request.Body.IpCount,
	}, request.PoolId)

	if err != nil {
		if errors.Is(err, pool.ErrPoolNotFound) {
			return gen.AddResources404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: err.Error()},
			}, nil
		}
		if errors.Is(err, pool.ErrValidation) {
			return gen.AddResources400JSONResponse{
				BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: err.Error()},
			}, nil
		}
		return nil, err
	}

	return gen.AddResources200ApplicationHalPlusJSONResponse(toPool(newPool, h.prefix)), nil
}
