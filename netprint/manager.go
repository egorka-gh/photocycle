package netprint

import (
	"context"
	"sync"
	"time"

	"github.com/egorka-gh/photocycle"
	"github.com/egorka-gh/photocycle/infrastructure/api"
	log "github.com/go-kit/kit/log"
)

//New creates new sync manager
func New(source, offset int, client api.FFService, repo photocycle.Repository, logger log.Logger) *Manager {
	if offset < 0 {
		offset = 1
	}
	return &Manager{
		source: source,
		offset: offset,
		client: client,
		repo:   repo,
		logger: logger,
	}

}

//Manager netprint sync manager
type Manager struct {
	source int
	offset int
	client api.FFService
	repo   photocycle.Repository
	logger log.Logger
}

//Run calls sync periodicaly, blocks caller till get quit
func (m *Manager) Run(interval int, quit chan struct{}) {
	var timer *time.Timer
	min := 10
	if interval < min {
		interval = min
	}
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
			go func() {
				defer wg.Done()
				defer cancel()
				m.Sync(ctx)
				timer = time.AfterFunc(time.Minute*time.Duration(interval), func() { start <- 1 })
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

}

//Sync fetch and save new boxes.
//boxes are filled with 10-20 min gap (after group get 30 state), so sync uses some offset in hours
func (m *Manager) Sync(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	m.logger.Log("event", "start")
	//last sync tstamp
	lastSyncts, err := m.repo.GetLastNetprintSync(ctx, m.source)
	if err != nil {
		m.logger.Log("Error", err.Error())
		return
	}
	if lastSyncts == 0 {
		lastSyncts = time.Now().Unix()
	}
	//fetch from
	t := time.Unix(lastSyncts, 0).Add(-time.Hour * time.Duration(m.offset))
	//current sync timestamp
	syncts := time.Now().Unix()
	//fetch
	groups, err := m.client.GetNPGroups(ctx, []int{30, 40}, t.Unix())
	if err != nil {
		m.logger.Log("Error", err.Error())
		return
	}

	if len(groups) == 0 {
		m.logger.Log("event", "end", "groups", "0")
		return
	}
	nps := make([]photocycle.GroupNetprint, 0, len(groups))
	boxCount := 0
	for _, group := range groups {
		if !group.Npfactory {
			continue
		}
		hasBoxes := false
		for _, box := range group.Boxes {
			if box.OrderNumber == "" {
				continue
			}
			hasBoxes = true
			boxCount++
			nps = append(nps, photocycle.GroupNetprint{
				BoxNumber:  box.BoxNumber,
				GroupID:    group.ID,
				NetprintID: box.OrderNumber,
				Source:     m.source,
				State:      group.Status.Value,
			})
		}
		if !hasBoxes {
			//not filled group
			nps = append(nps, photocycle.GroupNetprint{
				BoxNumber:  0,
				GroupID:    group.ID,
				NetprintID: "notprocessed",
				Source:     m.source,
				State:      0,
			})
		}
	}
	defer m.logger.Log("event", "end", "groups", len(groups), "boxes", boxCount)
	//persists
	err = m.repo.AddNetprints(context.Background(), nps)
	if err != nil {
		m.logger.Log("Error", err.Error())
		return
	}
	//fix fetch timestamp
	err = m.repo.SetLastNetprintSync(ctx, m.source, syncts)
	if err != nil {
		m.logger.Log("Error", err.Error())
	}
}
