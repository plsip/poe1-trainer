package game

import "context"

type Area struct {
	AreaCode string  `json:"areaCode"`
	Name     string  `json:"name"`
	Act      *int    `json:"act,omitempty"`
}

type AreaRepository interface {
	GetByCode(ctx context.Context, areaCode string) (*Area, error)
	GetByName(ctx context.Context, name string) ([]Area, error)
	GetAll(ctx context.Context) ([]Area, error)
}
