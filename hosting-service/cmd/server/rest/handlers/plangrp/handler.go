package plangrp

import (
	"context"
	"errors"
	"hosting-contracts/api"
	"hosting-service/internal/plan"
	"hosting-service/internal/platform/page"
)

type Handlers struct {
	planBus *plan.Business
}

func New(planBus *plan.Business) *Handlers {
	return &Handlers{
		planBus: planBus,
	}
}

func (h *Handlers) ListPlans(ctx context.Context, request api.ListPlansRequestObject) (api.ListPlansResponseObject, error) {
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

	return api.ListPlans200ApplicationHalPlusJSONResponse(toPlanCollectionResponse(pages, page, count)), nil
}

func (h *Handlers) CreatePlan(ctx context.Context, request api.CreatePlanRequestObject) (api.CreatePlanResponseObject, error) {
	newPlan, err := h.planBus.Create(ctx, plan.CreatePlanParams{
		Name:     request.Body.Name,
		CPUCores: request.Body.CpuCores,
		RAMMB:    request.Body.RamMb,
		DiskGB:   request.Body.DiskGb,
	})

	if err != nil {
		if errors.Is(err, plan.ErrValidation) {
			return api.CreatePlan400JSONResponse{
				BadRequestJSONResponse: api.BadRequestJSONResponse{Message: err.Error()},
			}, nil
		}
		return nil, err
	}

	return api.CreatePlan201JSONResponse(toPlan(newPlan)), nil
}

func (h *Handlers) GetPlanById(ctx context.Context, request api.GetPlanByIdRequestObject) (api.GetPlanByIdResponseObject, error) {
	newPlan, err := h.planBus.FindByID(ctx, request.PlanId)
	if err != nil {
		if errors.Is(err, plan.ErrPlanNotFound) {
			return api.GetPlanById404JSONResponse{
				Message: err.Error(),
			}, nil
		}
		return nil, err
	}

	return api.GetPlanById200ApplicationHalPlusJSONResponse(toPlan(newPlan)), nil
}
