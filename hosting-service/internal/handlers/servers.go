package handlers

import (
	"context"
	"errors"
	"hosting-service/internal/assemblers"
	"hosting-service/internal/domain"
	"hosting-service/internal/service"

	"hosting-contracts/api"
)

type ServersHandler struct {
	serverService service.ServerService
}

func NewServersHandler(serverService service.ServerService) *ServersHandler {
	return &ServersHandler{serverService: serverService}
}

func (h *ServersHandler) ListServers(ctx context.Context, request api.ListServersRequestObject) (api.ListServersResponseObject, error) {
	page := 1
	if request.Params.Page != nil {
		page = int(*request.Params.Page)
	}

	pageSize := 10
	if request.Params.PageSize != nil {
		pageSize = int(*request.Params.PageSize)
	}

	paginatedResult, err := h.serverService.Search(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	return api.ListServers200ApplicationHalPlusJSONResponse(assemblers.ToServerCollectionResponse(*paginatedResult, page, pageSize)), nil
}

func (h *ServersHandler) OrderServer(ctx context.Context, request api.OrderServerRequestObject) (api.OrderServerResponseObject, error) {
	server, err := h.serverService.Save(ctx, service.CreateServerParams{
		Name:   request.Body.Name,
		PlanID: request.Body.PlanId,
	})

	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			return api.OrderServer400JSONResponse{
				BadRequestJSONResponse: api.BadRequestJSONResponse{Message: err.Error()},
			}, nil
		}
		return nil, err
	}

	return api.OrderServer202ApplicationHalPlusJSONResponse(assemblers.ToServer(*server)), nil
}

func (h *ServersHandler) GetServerById(ctx context.Context, request api.GetServerByIdRequestObject) (api.GetServerByIdResponseObject, error) {
	server, err := h.serverService.FindByID(ctx, request.ServerId)

	if err != nil {
		if errors.Is(err, service.ErrServerNotFound) {
			return api.GetServerById404JSONResponse{
				NotFoundJSONResponse: api.NotFoundJSONResponse{Message: err.Error()},
			}, nil
		}
		return nil, err
	}

	return api.GetServerById200ApplicationHalPlusJSONResponse(assemblers.ToServer(*server)), nil
}

func (h *ServersHandler) PerformServerAction(ctx context.Context, request api.PerformServerActionRequestObject) (api.PerformServerActionResponseObject, error) {
	server, err := h.serverService.PerformAction(ctx, service.PerformActionParams{
		ServerID: request.ServerId,
		Action:   service.ActionType(request.Body.Action),
	})

	if err != nil {
		if errors.Is(err, service.ErrServerNotFound) {
			return api.PerformServerAction404JSONResponse{
				NotFoundJSONResponse: api.NotFoundJSONResponse{Message: err.Error()},
			}, nil
		}
		if errors.Is(err, service.ErrInvalidAction) {
			return api.PerformServerAction400JSONResponse{
				BadRequestJSONResponse: api.BadRequestJSONResponse{Message: err.Error()},
			}, nil
		}
		if errors.Is(err, domain.ErrValidation) {
			return api.PerformServerAction409JSONResponse{
				Message: err.Error(),
			}, nil
		}
		return nil, err
	}

	return api.PerformServerAction202JSONResponse(assemblers.ToServer(*server)), nil
}
