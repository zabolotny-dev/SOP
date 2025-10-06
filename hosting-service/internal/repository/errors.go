package repository

import "errors"

var (
	ErrServerNotFound = errors.New("server not found")
	ErrPlanNotFound   = errors.New("plan not found")
)
