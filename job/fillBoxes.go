package job

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/egorka-gh/photocycle"
	"github.com/egorka-gh/photocycle/api"
)

func fillBoxes(ctx context.Context, j *baseJob) error {
	//create api clients map
	var clients = make(map[int]api.FFService)
	su, err := j.repo.GetSourceUrls(ctx)
	if err != nil {
		return fmt.Errorf("repository.GetSourceUrls error: %s", err.Error())
	}
	if len(su) == 0 {
		return nil
	}
	for _, u := range su {
		c := &http.Client{
			Timeout: time.Second * 40,
		}
		cl, err := api.NewClient(c, u.URL, u.AppKey)
		if err != nil {
			return err
		}

		clients[u.ID] = cl
	}
	//fetch not processed groups
	grps, err := j.repo.GetNewPackages(ctx)
	if err != nil {
		return fmt.Errorf("repository.GetNewPackages error: %s", err.Error())
	}
	if len(grps) == 0 {
		return nil
	}
	//get boxes
	filled := make([]photocycle.Package, 0, len(grps))
	for _, g := range grps {
		cl, ok := clients[g.Source]
		if !ok {
			return fmt.Errorf("Source %d not found", g.Source)
		}
		//loadfrom site
		gbs, err := cl.GetBoxes(ctx, g.ID)
		if err != nil || len(gbs.Boxes) == 0 {
			//boxes not filled or some error
			if err != nil {
				j.logger.Log("error", fmt.Sprintf("api.GetBoxes error: %s", err.Error()))

			}
			//increment err counter and skip
			g.Attempt++
			j.repo.NewPackageUpdate(ctx, g)
			continue
		}
		//fill and save
		g.Boxes = make([]photocycle.PackageBox, 0, len(gbs.Boxes))
		for _, ba := range gbs.Boxes {
			bg := photocycle.PackageBox{
				Source:    g.Source,
				PackageID: g.ID,
				ID:        fmt.Sprintf("%d-%d", g.Source, ba.ID),
				Num:       ba.Number,
				Barcode:   ba.Barcode,
				Price:     ba.Price,
				Weight:    ba.Weight,
			}
			bg.Items = make([]photocycle.PackageBoxItem, 0, len(ba.Items))
			for _, bi := range ba.Items {
				i := photocycle.PackageBoxItem{
					BoxID:   bg.ID,
					OrderID: fmt.Sprintf("%d-%d", g.Source, bi.OrderID),
					Alias:   bi.Alias,
					Type:    bi.Type,
					From:    bi.From,
					To:      bi.To,
				}
				bg.Items = append(bg.Items, i)
			}
			g.Boxes = append(g.Boxes, bg)
		}
		filled = append(filled, g)
	}
	//persist && del processed
	err = j.repo.PackageAddWithBoxes(ctx, filled)
	if err != nil {
		err = fmt.Errorf("repository.PackageAddWithBoxes error: %s", err.Error())
	}
	j.logger.Log("result", fmt.Sprintf("Groups found %d, added %d.", len(grps), len(filled)))
	return err
}
