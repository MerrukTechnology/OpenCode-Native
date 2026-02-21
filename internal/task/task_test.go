package task

import (
	"testing"
)

func TestInMemoryTaskLifecycle(t *testing.T) {
	svc := NewInMemoryTaskService()
	steps := []Step{
		{ID: "s1", Description: "first", Type: "structured"},
		{ID: "s2", Description: "second", Type: "structured"},
	}

	task, err := svc.CreateTask("Example Task", "sess1", steps)
	if err != nil {
		t.Fatalf("CreateTask error: %v", err)
	}
	if task.Status != TaskPending {
		t.Fatalf("expected pending, got %v", task.Status)
	}
	if task.CurrentStepIndex != 0 {
		t.Fatalf("expected currentStepIndex 0, got %d", task.CurrentStepIndex)
	}

	// Start first step
	if err := svc.UpdateStep(task.ID, 0, StepRunning, "in progress", ""); err != nil {
		t.Fatalf("UpdateStep running error: %v", err)
	}
	t2, _ := svc.GetTask(task.ID)
	if t2.Status != TaskRunning {
		t.Fatalf("expected task running after starting step, got %v", t2.Status)
	}
	if t2.CurrentStepIndex != 0 {
		t.Fatalf("expected currentStepIndex 0 after starting first step, got %d", t2.CurrentStepIndex)
	}

	// Complete first step
	if err := svc.UpdateStep(task.ID, 0, StepCompleted, "done", ""); err != nil {
		t.Fatalf("UpdateStep completed error: %v", err)
	}
	t3, _ := svc.GetTask(task.ID)
	if t3.CurrentStepIndex != 1 {
		t.Fatalf("expected currentStepIndex 1 after completing first step, got %d", t3.CurrentStepIndex)
	}
	if t3.Status == TaskCompleted {
		t.Fatalf("task should not be completed yet")
	}

	// Complete second step
	if err := svc.UpdateStep(task.ID, 1, StepCompleted, "done", ""); err != nil {
		t.Fatalf("UpdateStep completed second error: %v", err)
	}
	t4, _ := svc.GetTask(task.ID)
	if t4.Status != TaskCompleted {
		t.Fatalf("expected task completed, got %v", t4.Status)
	}
}
