package handlers

import (
	"context"
	"errors"
	"hosting-contracts/api"
	"hosting-service/internal/assemblers"
	"hosting-service/internal/domain"
	"hosting-service/internal/service"
)

type PlanHandler struct {
	planService service.PlanService
}

func NewPlansHandler(planService service.PlanService) *PlanHandler {
	return &PlanHandler{planService: planService}
}

func (h *PlanHandler) ListPlans(ctx context.Context, request api.ListPlansRequestObject) (api.ListPlansResponseObject, error) {
	page := 1
	if request.Params.Page != nil {
		page = int(*request.Params.Page)
	}

	pageSize := 10
	if request.Params.PageSize != nil {
		pageSize = int(*request.Params.PageSize)
	}

	paginatedResult, err := h.planService.Search(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	return api.ListPlans200ApplicationHalPlusJSONResponse(assemblers.ToPlanCollectionResponse(*paginatedResult, page, pageSize)), nil
}

func (h *PlanHandler) CreatePlan(ctx context.Context, request api.CreatePlanRequestObject) (api.CreatePlanResponseObject, error) {
	plan, err := h.planService.Save(ctx, service.CreatePlanParams{
		Name:     request.Body.Name,
		CPUCores: request.Body.CpuCores,
		RAMMB:    request.Body.RamMb,
		DiskGB:   request.Body.DiskGb,
	})

	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			return api.CreatePlan400JSONResponse{
				BadRequestJSONResponse: api.BadRequestJSONResponse{Message: err.Error()},
			}, nil
		}
		return nil, err
	}

	return api.CreatePlan201JSONResponse(assemblers.ToPlan(*plan)), nil
}

func (h *PlanHandler) GetPlanById(ctx context.Context, request api.GetPlanByIdRequestObject) (api.GetPlanByIdResponseObject, error) {
	plan, err := h.planService.FindByID(ctx, request.PlanId)
	if err != nil {
		if errors.Is(err, service.ErrPlanNotFound) {
			return api.GetPlanById404JSONResponse{
				Message: err.Error(),
			}, nil
		}
		return nil, err
	}

	return api.GetPlanById200ApplicationHalPlusJSONResponse(assemblers.ToPlan(*plan)), nil
}
