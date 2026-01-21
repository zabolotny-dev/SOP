package rootgrp

import (
	"context"
	"hosting-service/cmd/server/rest/gen"
)

type RootHandlers struct {
	prefix string
}

func New(prefix string) *RootHandlers {
	return &RootHandlers{prefix: prefix}
}

func (r *RootHandlers) GetRoot(ctx context.Context, request gen.GetRootRequestObject) (gen.GetRootResponseObject, error) {
	return gen.GetRoot200ApplicationHalPlusJSONResponse(toRoot(r.prefix)), nil
}
