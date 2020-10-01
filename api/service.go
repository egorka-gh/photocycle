package api

import (
	"context"
)

// FFService describes the fabrika-fotoknigi.ru service.
type FFService interface {
	GetNPGroups(ctx context.Context, statuses []int, fromTS int64) ([]NPGroup, error)
}

//NPGroup netprint group dto (orders group)
type NPGroup struct {
	ID        int     `json:"id"`
	Status    Status  `json:"status"`
	CreatedTS int64   `json:"tstamp"`
	Boxes     []NPBox `json:"boxes"`
	Npfactory bool    `json:"npfactory"`
}

//NPBox dto (netprint maip box)
type NPBox struct {
	BoxNumber   int    `json:"number"`
	OrderNumber string `json:"orderNumber"`
}

//Status dto (group post box)
type Status struct {
	Value int `json:"value"`
}
