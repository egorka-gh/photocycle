package job

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/egorka-gh/photocycle"
	"github.com/egorka-gh/photocycle/api"
	log "github.com/go-kit/kit/log"
)

//Job job to do
type Job interface {
	Init() error
	Do(ctx context.Context)
}

type baseJob struct {
	name     string
	repo     photocycle.Repository
	logger   log.Logger
	initFunc func(j *baseJob) error
	doFunc   func(ctx context.Context, j *baseJob) error
}

func (j *baseJob) Init() error {
	if j.logger == nil {
		j.logger = log.NewNopLogger()
	}
	j.logger = log.With(j.logger, "actor", "job-"+j.name)
	if j.initFunc != nil {
		return j.initFunc(j)
	}
	return nil
}

func (j *baseJob) Do(ctx context.Context) {
	if j.doFunc != nil {
		err := j.doFunc(ctx, j)
		if err != nil {
			j.logger.Log("Error", err)
		}
		return
	}
	return
}

//Runer job runer
type Runer struct {
	interval int
	logger   log.Logger
	jobs     []Job
}

//Run runs jobs periodicaly, blocks caller till get quit
func (r *Runer) Run(quit chan struct{}) error {
	if len(r.jobs) == 0 {
		err := errors.New("No jobs to do")
		return err
	}
	if r.logger == nil {
		r.logger = log.NewNopLogger()
	}
	r.logger = log.With(r.logger, "actor", "runner")
	//init jobs
	r.logger.Log("event", "Starting.Init jobs.")
	for _, job := range r.jobs {
		if err := job.Init(); err != nil {
			return err
		}
	}

	var timer *time.Timer
	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()
	start := make(chan int, 1)
	defer close(start)
	var wg sync.WaitGroup
	loop := true

	start <- 1
	for loop {
		select {
		case <-start:
			ctx, cancel := context.WithCancel(mainCtx)
			wg.Add(1)
			r.logger.Log("event", "Starting jobs.")
			go func() {
				defer wg.Done()
				defer cancel()
				//run sequentially
				for _, job := range r.jobs {
					job.Do(ctx)
					if ctx.Err() != nil {
						//contex canceled
						return
					}
				}
				timer = time.AfterFunc(time.Minute*time.Duration(r.interval), func() { start <- 1 })
			}()
		case <-quit:
			if timer != nil {
				timer.Stop()
			}
			mainCancel()
			loop = false
			wg.Wait()
		}
	}

	return nil
}

func fillBoxes(ctx context.Context, j *baseJob) error {
	//create api clients map
	var clients = make(map[int]api.FFService)
	su, err := j.repo.GetSourceUrls(ctx)
	if err != nil {
		return err
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
		return err
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
			//increment err counter and skip
			g.Attempt++
			j.repo.NewPackageUpdate(ctx, g)
			continue
		}
		//fill and save
		g.Boxes = make([]photocycle.PackageBox, 0, len(gbs.Boxes))
		for _, ba := range gbs.Boxes {
			bg := photocycle.PackageBox{
				ID:        fmt.Sprintf("%d-%d", g.Source, ba.ID),
				PackageID: g.ID,
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
	return j.repo.PackageAddWithBoxes(ctx, filled)
}
