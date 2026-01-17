package servergrp

import (
	"context"
	"errors"
	"hosting-kit/page"
	"hosting-service/cmd/server/rest/gen"
	"hosting-service/internal/server"
)

type Handlers struct {
	serverBus server.ExtBusiness
	prefix    string
}

func New(serverBus server.ExtBusiness, prefix string) *Handlers {
	return &Handlers{
		serverBus: serverBus,
		prefix:    prefix,
	}
}

func (h *Handlers) ListServers(ctx context.Context, request gen.ListServersRequestObject) (gen.ListServersResponseObject, error) {
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

	return gen.ListServers200ApplicationHalPlusJSONResponse(toServerCollectionResponse(servers, page, count, h.prefix)), nil
}

func (h *Handlers) OrderServer(ctx context.Context, request gen.OrderServerRequestObject) (gen.OrderServerResponseObject, error) {
	newServer, err := h.serverBus.Create(ctx, request.Body.Name, request.Body.PlanId)

	if err != nil {
		if errors.Is(err, server.ErrValidation) {
			return gen.OrderServer400JSONResponse{
				BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: err.Error()},
			}, nil
		}
		if errors.Is(err, server.ErrInvalidPlan) {
			return gen.OrderServer400JSONResponse{
				BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: server.ErrInvalidPlan.Error()},
			}, nil
		}
		if errors.Is(err, server.ErrNoResources) {
			return gen.OrderServer409JSONResponse{
				ConflictJSONResponse: gen.ConflictJSONResponse{Message: server.ErrNoResources.Error()},
			}, nil
		}
		return nil, err
	}

	return gen.OrderServer202ApplicationHalPlusJSONResponse(toServer(newServer, h.prefix)), nil
}

func (h *Handlers) GetServerById(ctx context.Context, request gen.GetServerByIdRequestObject) (gen.GetServerByIdResponseObject, error) {
	serverFound, err := h.serverBus.FindByID(ctx, request.ServerId)

	if err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			return gen.GetServerById404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: server.ErrServerNotFound.Error()},
			}, nil
		}
		return nil, err
	}

	return gen.GetServerById200ApplicationHalPlusJSONResponse(toServer(serverFound, h.prefix)), nil
}

func (h *Handlers) PerformServerAction(ctx context.Context, request gen.PerformServerActionRequestObject) (gen.PerformServerActionResponseObject, error) {
	id := request.ServerId
	var err error
	var newServer server.Server

	switch request.Body.Action {
	case gen.START:
		newServer, err = h.serverBus.Start(ctx, id)
	case gen.STOP:
		newServer, err = h.serverBus.Stop(ctx, id)
	case gen.DELETE:
		newServer, err = h.serverBus.Delete(ctx, id)
	default:
		return gen.PerformServerAction400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{Message: "Unknown action"},
		}, nil
	}

	if err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			return gen.PerformServerAction404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: server.ErrServerNotFound.Error()},
			}, nil
		}
		if errors.Is(err, server.ErrValidation) {
			return gen.PerformServerAction409JSONResponse{
				Message: err.Error(),
			}, nil
		}
		return nil, err
	}

	return gen.PerformServerAction202JSONResponse(toServer(newServer, h.prefix)), nil
}
