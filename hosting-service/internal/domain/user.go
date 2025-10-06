package domain

import (
	"fmt"
	"strings"
)

type User struct {
	BaseModel
	Name    string    `gorm:"type:varchar(255);not null"`
	Servers []*Server `gorm:"foreignKey:UserID"`
}

func NewUser(name string) (*User, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, fmt.Errorf("%w: user name cannot be empty", ErrValidation)
	}
	if len(trimmedName) < 2 {
		return nil, fmt.Errorf("%w: user name must be at least 2 characters long", ErrValidation)
	}

	user := &User{
		BaseModel: NewBaseModel(),
		Name:      trimmedName,
	}

	return user, nil
}
