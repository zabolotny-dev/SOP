package plangrp

import (
	"context"
	"errors"
	"hosting-kit/page"
	"hosting-service/cmd/server/rest/gen"
	"hosting-service/internal/plan"
)

type Handlers struct {
	planBus plan.ExtBusiness
	prefix  string
}

func New(planBus plan.ExtBusiness, prefix string) *Handlers {
	return &Handlers{
		planBus: planBus,
		prefix:  prefix,
	}
}

func (h *Handlers) ListPlans(ctx context.Context, request gen.ListPlansRequestObject) (gen.ListPlansResponseObject, error) {
	pageNum := 1
	pageSize := 10

	if request.Params.Page != nil {
		pageNum = *request.Params.Page
	}
	if request.Params.PageSize != nil {
		pageSize = *request.Params.PageSize
	}

	page := page.Parse(pageNum, pageSize)

	pages, count, err := h.planBus.Search(ctx, page)

	if err != nil {
		return nil, err
	}

	return gen.ListPlans200ApplicationHalPlusJSONResponse(toPlanCollectionResponse(pages, page, count, h.prefix)), nil
}

func (h *Handlers) CreatePlan(ctx context.Context, request gen.CreatePlanRequestObject) (gen.CreatePlanResponseObject, error) {
	newPlan, err := h.planBus.Create(ctx, plan.CreatePlanParams{
		Name:     request.Body.Name,
		CPUCores: request.Body.CpuCores,
		RAMMB:    request.Body.RamMb,
		DiskGB:   request.Body.DiskGb,
		IpCount:  request.Body.IpCount,
	})

	if err != nil {
		if errors.Is(err, plan.ErrValidation) {
			return gen.CreatePlan400JSONResponse{
				BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: err.Error()},
			}, nil
		}
		return nil, err
	}

	return gen.CreatePlan201JSONResponse(toPlan(newPlan, h.prefix)), nil
}

func (h *Handlers) GetPlanById(ctx context.Context, request gen.GetPlanByIdRequestObject) (gen.GetPlanByIdResponseObject, error) {
	newPlan, err := h.planBus.FindByID(ctx, request.PlanId)
	if err != nil {
		if errors.Is(err, plan.ErrPlanNotFound) {
			return gen.GetPlanById404JSONResponse{
				Message: plan.ErrPlanNotFound.Error(),
			}, nil
		}
		return nil, err
	}

	return gen.GetPlanById200ApplicationHalPlusJSONResponse(toPlan(newPlan, h.prefix)), nil
}
