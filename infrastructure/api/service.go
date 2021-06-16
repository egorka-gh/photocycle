package api

import (
	"context"
)

// FFService describes the fabrika-fotoknigi.ru service.
type FFService interface {
	//netprint boxes (transit mail boxes)
	GetNPGroups(ctx context.Context, statuses []int, fromTS int64) ([]NPGroup, error)

	//common FF api
	GetBoxes(ctx context.Context, groupID int) (*GroupBoxes, error)
	GetGroup(ctx context.Context, groupID int) (map[string]interface{}, error)
}
