package api

import (
	"context"
)

// FFService describes the fabrika-fotoknigi.ru service.
type FFService interface {
	GetGroups(ctx context.Context, statuses []int, fromTS int64) ([]Group, error)
}

//Group dto (orders group)
type Group struct {
	ID        int    `json:"id"`
	Status    Status `json:"status"`
	CreatedTS int64  `json:"tstamp"`
	Boxes     []Box  `json:"boxes"`
	Npfactory bool   `json:"npfactory"`
}

//Box dto (group post box)
type Box struct {
	BoxNumber   int    `json:"number"`
	OrderNumber string `json:"orderNumber"`
}

//Status dto (group post box)
type Status struct {
	Value int `json:"value"`
}
