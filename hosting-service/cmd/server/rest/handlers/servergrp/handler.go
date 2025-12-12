package servergrp

import (
	"context"
	"errors"
	"hosting-contracts/api"
	"hosting-service/internal/platform/page"
	"hosting-service/internal/server"
)

type Handlers struct {
	serverBus *server.Business
}

func New(serverBus *server.Business) *Handlers {
	return &Handlers{
		serverBus: serverBus,
	}
}

func (h *Handlers) ListServers(ctx context.Context, request api.ListServersRequestObject) (api.ListServersResponseObject, error) {
	pageNum := 1
	pageSize := 10

	if request.Params.Page != nil {
		pageNum = *request.Params.Page
	}

	if request.Params.PageSize != nil {
		pageSize = *request.Params.PageSize
	}

	page := page.Parse(pageNum, pageSize)

	servers, count, err := h.serverBus.Search(ctx, page)
	if err != nil {
		return nil, err
	}

	return api.ListServers200ApplicationHalPlusJSONResponse(toServerCollectionResponse(servers, page, count)), nil
}

func (h *Handlers) OrderServer(ctx context.Context, request api.OrderServerRequestObject) (api.OrderServerResponseObject, error) {
	newServer, err := h.serverBus.Create(ctx, request.Body.Name, request.Body.PlanId)

	if err != nil {
		if errors.Is(err, server.ErrValidation) {
			return api.OrderServer400JSONResponse{
				BadRequestJSONResponse: api.BadRequestJSONResponse{Message: err.Error()},
			}, nil
		}
		return nil, err
	}

	return api.OrderServer202ApplicationHalPlusJSONResponse(toServer(newServer)), nil
}

func (h *Handlers) GetServerById(ctx context.Context, request api.GetServerByIdRequestObject) (api.GetServerByIdResponseObject, error) {
	serverFound, err := h.serverBus.FindByID(ctx, request.ServerId)

	if err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			return api.GetServerById404JSONResponse{
				NotFoundJSONResponse: api.NotFoundJSONResponse{Message: err.Error()},
			}, nil
		}
		return nil, err
	}

	return api.GetServerById200ApplicationHalPlusJSONResponse(toServer(serverFound)), nil
}

func (h *Handlers) PerformServerAction(ctx context.Context, request api.PerformServerActionRequestObject) (api.PerformServerActionResponseObject, error) {
	id := request.ServerId
	var err error
	var newServer server.Server

	switch request.Body.Action {
	case api.START:
		newServer, err = h.serverBus.Start(ctx, id)
	case api.STOP:
		newServer, err = h.serverBus.Stop(ctx, id)
	case api.DELETE:
		newServer, err = h.serverBus.Delete(ctx, id)
	default:
		return api.PerformServerAction400JSONResponse{
			BadRequestJSONResponse: api.BadRequestJSONResponse{Message: "Unknown action"},
		}, nil
	}

	if err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			return api.PerformServerAction404JSONResponse{
				NotFoundJSONResponse: api.NotFoundJSONResponse{Message: err.Error()},
			}, nil
		}
		if errors.Is(err, server.ErrValidation) {
			return api.PerformServerAction409JSONResponse{
				Message: err.Error(),
			}, nil
		}
		return nil, err
	}

	return api.PerformServerAction202JSONResponse(toServer(newServer)), nil
}
