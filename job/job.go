package job

import (
	"context"
	"errors"
	"sync"
	"time"

	log "github.com/go-kit/kit/log"
)

//Job job to do
type Job interface {
	Init() error
	Do(ctx context.Context)
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
