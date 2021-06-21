package job

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/egorka-gh/photocycle"
	"github.com/egorka-gh/photocycle/infrastructure/api"
)

func initFillBoxes(j *baseJob) error {
	b, err := api.CreateBuilder(j.repo)
	if err != nil {
		return fmt.Errorf("initFillBoxes error: %s", err.Error())
	}
	j.builder = b
	return nil
}

func fillBoxes(ctx context.Context, j *baseJob) error {
	//create api clients map
	var clients = make(map[int]api.FFService)
	var hasBox = make(map[int]bool)
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
		hasBox[u.ID] = u.HasBoxes
	}
	//fetch not processed groups
	grps, err := j.repo.GetNewPackages(ctx)
	if err != nil {
		return fmt.Errorf("repository.GetNewPackages error: %s", err.Error())
	}
	if len(grps) == 0 {
		return nil
	}
	//get boxes first
	filled := make([]*photocycle.Package, 0, len(grps))
	for _, g := range grps {
		cl, ok := clients[g.Source]
		if !ok {
			return fmt.Errorf("source %d not found", g.Source)
		}
		//load boxes from site (with hasBox or not )
		gbs, err := cl.GetBoxes(ctx, g.ID)
		//process err only if site hasBox
		if hasBox[g.Source] && (err != nil || gbs == nil || len(gbs.Boxes) == 0) {
			//boxes not filled or some error
			if err != nil {
				j.logger.Log("error", fmt.Sprintf("api.GetBoxes error: %s", err.Error()))

			}
			//increment err counter and skip
			g.Attempt++
			if g.Attempt < 3 {
				//maybe it's not ready
				//try next time
				j.repo.NewPackageUpdate(ctx, g)
				continue
			}
		}

		//j.logger.Log("boxes", fmt.Sprintf("%v", gbs))

		//get group (raw)
		raw, err := cl.GetGroup(ctx, g.ID)
		if err != nil {
			j.logger.Log("error", fmt.Sprintf("api.GetGroup error: %s", err.Error()))
			g.Attempt++
			j.repo.NewPackageUpdate(ctx, g)
			continue
		}
		group, err := j.builder.BuildPackage(g.Source, raw)
		if err != nil {
			j.logger.Log("error", fmt.Sprintf("api.BuildPackage error: %s", err.Error()))
			g.Attempt++
			j.repo.NewPackageUpdate(ctx, g)
			continue
		}

		//fill boxes
		group.Boxes = make([]photocycle.PackageBox, 0)
		// can be nil if site not suppert boxes
		if gbs != nil {
			for _, ba := range gbs.Boxes {
				bg := photocycle.PackageBox{
					Source:    group.Source,
					PackageID: group.ID,
					ID:        fmt.Sprintf("%d-%d", group.Source, ba.ID),
					Num:       ba.Number,
					Barcode:   ba.Barcode,
					Price:     ba.Price,
					Weight:    ba.Weight,
				}
				bg.Items = make([]photocycle.PackageBoxItem, 0, len(ba.Items))
				for _, bi := range ba.Items {
					i := photocycle.PackageBoxItem{
						BoxID:   bg.ID,
						OrderID: fmt.Sprintf("%d_%d", group.Source, bi.OrderID),
						Alias:   bi.Alias,
						Type:    bi.Type,
						From:    bi.From,
						To:      bi.To,
					}
					bg.Items = append(bg.Items, i)
				}
				group.Boxes = append(group.Boxes, bg)
			}
		}
		filled = append(filled, group)
	}
	//persist && del processed
	err = j.repo.PackageAddWithBoxes(ctx, filled)
	if err != nil {
		err = fmt.Errorf("repository.PackageAddWithBoxes error: %s", err.Error())
	} else {
		j.logger.Log("result", fmt.Sprintf("Groups found %d, added %d", len(grps), len(filled)))
	}
	return err
}
