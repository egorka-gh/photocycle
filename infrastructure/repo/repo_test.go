package repo

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

type exec struct {
}

var execRuns int

func (e *exec) Exec(query string, args ...interface{}) (sql.Result, error) {
	execRuns++
	i := strings.Count(query, "?")
	if i != len(args) {
		return nil, fmt.Errorf("wrong args len, query args count %d got %d", i, len(args))
	}
	return nil, nil
}

func TestInsertBatch(t *testing.T) {
	execRuns = 0
	args := make([]interface{}, 0, 1000)
	for i := 0; i < 25; i++ {
		j := i
		args = append(args, j)
	}
	err := insertBatch(new(exec), "INSERT", "(?,?,?,?,?)", args)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if execRuns != 1 {
		t.Error(fmt.Sprintf("expected runs %d got %d", 1, execRuns))
	}

	execRuns = 0
	maxParamsPerBatch = 30
	args = args[:0]
	for i := 0; i < 50; i++ {
		j := i
		args = append(args, j)
	}
	err = insertBatch(new(exec), "INSERT", "(?,?,?,?,?)", args)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if execRuns != 2 {
		t.Error(fmt.Sprintf("expected runs %d got %d", 1, execRuns))
	}

	execRuns = 0
	maxParamsPerBatch = 10
	err = insertBatch(new(exec), "INSERT", "(?,?,?,?,?)", args)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if execRuns != 5 {
		t.Error(fmt.Sprintf("expected runs %d got %d", 1, execRuns))
	}

}
