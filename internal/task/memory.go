package task

import (
	"errors"
	"sync"
	"time"
)

// InMemoryTaskService is a thread-safe in-memory implementation of Service.
type InMemoryTaskService struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

// NewInMemoryTaskService creates a new in-memory task service.
func NewInMemoryTaskService() *InMemoryTaskService {
	return &InMemoryTaskService{tasks: make(map[string]*Task)}
}

// CreateTask creates a new task with the given title, session and steps.
func (s *InMemoryTaskService) CreateTask(title, sessionID string, steps []Step) (*Task, error) {
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}
	id := generateTaskID()
	now := time.Now()
	t := &Task{
		ID:               id,
		SessionID:        sessionID,
		Title:            title,
		Status:           TaskPending,
		CurrentStepIndex: 0,
		Steps:            stepsWithDefaults(steps),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[id] = t
	return t, nil
}

// GetTask retrieves a task by ID.
func (s *InMemoryTaskService) GetTask(id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, errors.New("task not found")
	}
	// Return a shallow copy to avoid external mutation.
	cp := *t
	cp.Steps = append([]Step(nil), t.Steps...)
	return &cp, nil
}

// UpdateStep updates a specific step within a task.
func (s *InMemoryTaskService) UpdateStep(taskID string, stepIndex int, newStatus StepStatus, output string, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[taskID]
	if !ok {
		return errors.New("task not found")
	}
	if stepIndex < 0 || stepIndex >= len(task.Steps) {
		return errors.New("invalid step index")
	}
	st := &task.Steps[stepIndex]
	st.Status = newStatus
	st.Output = output
	st.Error = errMsg

	// Update timestamps
	task.UpdatedAt = time.Now()

	// Recompute task status and next index
	allDone := true
	for i := range task.Steps {
		if task.Steps[i].Status != StepCompleted {
			allDone = false
			break
		}
	}
	if allDone {
		task.Status = TaskCompleted
		task.CurrentStepIndex = len(task.Steps)
		return nil
	}

	// Update current step index to first non-completed step
	next := 0
	for i := range task.Steps {
		if task.Steps[i].Status != StepCompleted {
			next = i
			break
		}
	}
	task.CurrentStepIndex = next
	// If any step is currently running, reflect that in task status
	if newStatus == StepRunning {
		task.Status = TaskRunning
	}
	return nil
}

// ListTasks returns all tasks.
func (s *InMemoryTaskService) ListTasks() ([]Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		// copy
		cp := *t
		cp.Steps = append([]Step(nil), t.Steps...)
		out = append(out, cp)
	}
	return out, nil
}

// Helpers
func generateTaskID() string {
	return "task-" + time.Now().Format("20060102150405.000000000")
}

func stepsWithDefaults(steps []Step) []Step {
	out := make([]Step, 0, len(steps))
	for _, s := range steps {
		st := s
		if st.Status == "" {
			st.Status = StepPending
		}
		out = append(out, st)
	}
	return out
}
