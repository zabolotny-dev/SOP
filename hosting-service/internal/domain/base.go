package domain

import (
	"errors"

	"github.com/google/uuid"
)

type BaseModel struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;not null"`
}

func NewBaseModel() BaseModel {
	return BaseModel{
		ID: uuid.New(),
	}
}

var ErrValidation = errors.New("validation error")
