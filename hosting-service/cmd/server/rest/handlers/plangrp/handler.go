package plangrp

import (
	"context"
	"errors"
	"hosting-kit/page"
	"hosting-service/cmd/server/rest/gen"
	"hosting-service/internal/plan"
)

type PlanHandlers struct {
	planBus plan.ExtBusiness
	prefix  string
}

func New(planBus plan.ExtBusiness, prefix string) *PlanHandlers {
	return &PlanHandlers{
		planBus: planBus,
		prefix:  prefix,
	}
}

func (p *PlanHandlers) ListPlans(ctx context.Context, request gen.ListPlansRequestObject) (gen.ListPlansResponseObject, error) {
	pageNum := 1
	pageSize := 10

	if request.Params.Page != nil {
		pageNum = *request.Params.Page
	}
	if request.Params.PageSize != nil {
		pageSize = *request.Params.PageSize
	}

	page := page.Parse(pageNum, pageSize)

	pages, count, err := p.planBus.Search(ctx, page)

	if err != nil {
		return nil, err
	}

	return gen.ListPlans200ApplicationHalPlusJSONResponse(toPlanCollectionResponse(pages, page, count, p.prefix)), nil
}

func (p *PlanHandlers) CreatePlan(ctx context.Context, request gen.CreatePlanRequestObject) (gen.CreatePlanResponseObject, error) {
	newPlan, err := p.planBus.Create(ctx, plan.CreatePlanParams{
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

	return gen.CreatePlan201JSONResponse(toPlan(newPlan, p.prefix)), nil
}

func (p *PlanHandlers) GetPlanById(ctx context.Context, request gen.GetPlanByIdRequestObject) (gen.GetPlanByIdResponseObject, error) {
	newPlan, err := p.planBus.FindByID(ctx, request.PlanId)
	if err != nil {
		if errors.Is(err, plan.ErrPlanNotFound) {
			return gen.GetPlanById404JSONResponse{
				Message: plan.ErrPlanNotFound.Error(),
			}, nil
		}
		return nil, err
	}

	return gen.GetPlanById200ApplicationHalPlusJSONResponse(toPlan(newPlan, p.prefix)), nil
}
