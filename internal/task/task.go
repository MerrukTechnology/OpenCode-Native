package task

import (
	"time"
)

// TaskStatus represents the high-level status of a multi-step task.
type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskRunning   TaskStatus = "running"
	TaskCompleted TaskStatus = "completed"
	TaskFailed    TaskStatus = "failed"
)

// StepStatus represents the status of an individual task step.
type StepStatus string

const (
	StepPending   StepStatus = "pending"
	StepRunning   StepStatus = "running"
	StepCompleted StepStatus = "completed"
	StepFailed    StepStatus = "failed"
)

// Step describes a single actionable unit within a Task.
type Step struct {
	ID          string     // unique step identifier
	Description string     // what this step does
	Type        string     // "freeform" or "structured"
	Status      StepStatus // current status
	RetryCount  int        // how many retries have occurred
	Output      string     // textual output from execution
	Error       string     // error message if failed
}

// Task represents a multi-step plan that agents can execute.
type Task struct {
	ID               string // unique task identifier
	SessionID        string // session this task belongs to
	Title            string // human-readable title
	Status           TaskStatus
	CurrentStepIndex int       // index of the currently active step
	Steps            []Step    // ordered list of steps
	CreatedAt        time.Time // creation timestamp
	UpdatedAt        time.Time // last update timestamp
}

// Service defines the contract for a task planning and execution service.
type Service interface {
	// CreateTask creates a new task with the given title, session and steps.
	CreateTask(title, sessionID string, steps []Step) (*Task, error)
	// GetTask retrieves a task by ID.
	GetTask(id string) (*Task, error)
	// UpdateStep updates a specific step within a task.
	UpdateStep(taskID string, stepIndex int, newStatus StepStatus, output string, err string) error
	// ListTasks returns all tasks.
	ListTasks() ([]Task, error)
}
