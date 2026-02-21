package task

import (
	"testing"
)

func TestInMemoryTask_CreateTask_EmptyTitle(t *testing.T) {
	svc := NewInMemoryTaskService()
	_, err := svc.CreateTask("", "sess1", nil)
	if err == nil {
		t.Fatalf("expected error when creating task with empty title")
	}
}

func TestInMemoryTask_GetTask_NotFound(t *testing.T) {
	svc := NewInMemoryTaskService()
	if _, err := svc.GetTask("not-found"); err == nil {
		// some implementations may return nil error; ensure error is returned
		t.Fatalf("expected error for not-found task")
	}
}

func TestInMemoryTask_UpdateStep_InvalidIndex(t *testing.T) {
	svc := NewInMemoryTaskService()
	tsk, _ := svc.CreateTask("Sample", "sess", []Step{{ID: "a", Description: "d"}})
	if err := svc.UpdateStep(tsk.ID, 5, StepPending, "", ""); err == nil {
		t.Fatalf("expected error for invalid step index")
	}
}

func TestInMemoryTask_ListTasks(t *testing.T) {
	svc := NewInMemoryTaskService()
	if _, err := svc.CreateTask("T1", "sess", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.CreateTask("T2", "sess", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tasks, err := svc.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
}
