package job

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

type jb struct {
	initErr error
	doFunc  func()
}

func (j *jb) Init() error {
	return j.initErr
}

func (j *jb) Do(ctx context.Context) {
	defer func() {
		if j.doFunc != nil {
			j.doFunc()
		}
	}()
	//exit on timer or context cancel
	timer := time.NewTimer(1 * time.Second)
	select {
	case <-timer.C:
		return
	case <-ctx.Done():
		return
	}
}

func TestGroupBoxes(t *testing.T) {
	var cnt uint64
	q := make(chan struct{})
	defer func() {
		close(q)
	}()
	//check init error
	j := &jb{
		initErr: errors.New("init err"),
		doFunc: func() {
			atomic.AddUint64(&cnt, 1)
		},
	}
	r := baseRuner{
		interval: 1,
		jobs:     []Job{j},
	}

	err := r.Run(q)
	fmt.Printf("error:  %v\n", err)
	if err == nil {
		t.Error("Expect error on jobs init, got nil")
		return
	}
	if cnt > 0 {
		t.Errorf("Expected job runs 0, got %q", cnt)
		return
	}

	j.initErr = nil
	r.jobs = []Job{j, j, j}
	time.AfterFunc(5*time.Second, func() { q <- struct{}{} })
	err = r.Run(q)
	if err != nil {
		t.Errorf("Error %q", err.Error())
		return
	}
	fmt.Printf("jobs:  %v\n", cnt)
	if cnt != 3 {
		t.Errorf("Expected job runs 3, got %q", cnt)
		return
	}

}
