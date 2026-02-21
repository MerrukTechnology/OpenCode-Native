package task

import (
	"database/sql"
	dbsql "github.com/MerrukTechnology/OpenCode-Native/internal/db/sql"
	_ "modernc.org/sqlite"
	"testing"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:memdb_task?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}
	// initialize schema
	if err := dbsql.InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	return db
}

func TestDBTaskLifecycle(t *testing.T) {
	db := newTestDB(t)
	svc := NewDBTaskService(db)
	// create task
	steps := []Step{{ID: "s1", Description: "first", Type: "structured"}, {ID: "s2", Description: "second", Type: "structured"}}
	tsk, err := svc.CreateTask("DB Task", "sess1", steps)
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if tsk.Status != TaskPending {
		t.Fatalf("expected pending, got %v", tsk.Status)
	}

	// update first step to running
	if err := svc.UpdateStep(tsk.ID, 0, StepRunning, "in progress", ""); err != nil {
		t.Fatalf("UpdateStep: %v", err)
	}
	// complete first step
	if err := svc.UpdateStep(tsk.ID, 0, StepCompleted, "done", ""); err != nil {
		t.Fatalf("UpdateStep: %v", err)
	}
	// complete second step
	if err := svc.UpdateStep(tsk.ID, 1, StepCompleted, "done", ""); err != nil {
		t.Fatalf("UpdateStep: %v", err)
	}
	// fetch final state
	tsk2, err := svc.GetTask(tsk.ID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if tsk2.Status != TaskCompleted {
		t.Fatalf("expected completed, got %v", tsk2.Status)
	}
}
