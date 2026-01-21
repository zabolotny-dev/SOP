package servergrp

import (
	"context"
	"errors"
	"hosting-kit/auth"
	"hosting-kit/page"
	"hosting-service/cmd/server/rest/gen"
	"hosting-service/internal/server"
)

type ServerHandlers struct {
	serverBus server.ExtBusiness
	prefix    string
}

func New(serverBus server.ExtBusiness, prefix string) *ServerHandlers {
	return &ServerHandlers{
		serverBus: serverBus,
		prefix:    prefix,
	}
}

func (s *ServerHandlers) ListServers(ctx context.Context, request gen.ListServersRequestObject) (gen.ListServersResponseObject, error) {
	pageNum := 1
	pageSize := 10

	if request.Params.Page != nil {
		pageNum = *request.Params.Page
	}

	if request.Params.PageSize != nil {
		pageSize = *request.Params.PageSize
	}

	page := page.Parse(pageNum, pageSize)

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return nil, err
	}

	servers, count, err := s.serverBus.Search(ctx, page, claims.UserID)
	if err != nil {
		return nil, err
	}

	return gen.ListServers200ApplicationHalPlusJSONResponse(toServerCollectionResponse(servers, page, count, s.prefix)), nil
}

func (s *ServerHandlers) OrderServer(ctx context.Context, request gen.OrderServerRequestObject) (gen.OrderServerResponseObject, error) {
	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return nil, err
	}

	newServer, err := s.serverBus.Create(ctx, request.Body.Name, request.Body.PlanId, claims.UserID)

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

	return gen.OrderServer202ApplicationHalPlusJSONResponse(toServer(newServer, s.prefix)), nil
}

func (s *ServerHandlers) GetServerById(ctx context.Context, request gen.GetServerByIdRequestObject) (gen.GetServerByIdResponseObject, error) {
	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return nil, err
	}

	serverFound, err := s.serverBus.FindByID(ctx, request.ServerId, claims.UserID)

	if err != nil {
		if errors.Is(err, server.ErrServerNotFound) {
			return gen.GetServerById404JSONResponse{
				NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: server.ErrServerNotFound.Error()},
			}, nil
		}
		return nil, err
	}

	return gen.GetServerById200ApplicationHalPlusJSONResponse(toServer(serverFound, s.prefix)), nil
}

func (s *ServerHandlers) PerformServerAction(ctx context.Context, request gen.PerformServerActionRequestObject) (gen.PerformServerActionResponseObject, error) {
	id := request.ServerId
	var err error
	var newServer server.Server

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return nil, err
	}

	switch request.Body.Action {
	case gen.START:
		newServer, err = s.serverBus.Start(ctx, id, claims.UserID)
	case gen.STOP:
		newServer, err = s.serverBus.Stop(ctx, id, claims.UserID)
	case gen.DELETE:
		newServer, err = s.serverBus.Delete(ctx, id, claims.UserID)
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

	return gen.PerformServerAction202JSONResponse(toServer(newServer, s.prefix)), nil
}
