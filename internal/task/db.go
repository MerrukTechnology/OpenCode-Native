package task

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// DBTaskService is a DB-backed implementation of the Task Service.
type DBTaskService struct {
	db *sql.DB
}

func NewDBTaskService(db *sql.DB) *DBTaskService {
	return &DBTaskService{db: db}
}

func (s *DBTaskService) CreateTask(title, sessionID string, steps []Step) (*Task, error) {
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}
	id := generateTaskID()
	now := time.Now()
	sSteps := stepsWithDefaults(steps)
	b, err := json.Marshal(sSteps)
	if err != nil {
		return nil, err
	}
	_, err = s.db.Exec(`INSERT INTO tasks (id, session_id, title, status, current_step_index, steps, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, id, sessionID, title, string(TaskPending), 0, string(b), now, now)
	if err != nil {
		return nil, err
	}
	return &Task{ID: id, SessionID: sessionID, Title: title, Status: TaskPending, CurrentStepIndex: 0, Steps: sSteps, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *DBTaskService) GetTask(id string) (*Task, error) {
	row := s.db.QueryRow(`SELECT id, session_id, title, status, current_step_index, steps, created_at, updated_at FROM tasks WHERE id = ?`, id)
	var tid, sessionID, title, statusStr string
	var current int
	var stepsJSON string
	var createdAt, updatedAt time.Time
	if err := row.Scan(&tid, &sessionID, &title, &statusStr, &current, &stepsJSON, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	var steps []Step
	if err := json.Unmarshal([]byte(stepsJSON), &steps); err != nil {
		return nil, err
	}
	t := &Task{ID: tid, SessionID: sessionID, Title: title, Status: TaskStatus(statusStr), CurrentStepIndex: current, Steps: steps, CreatedAt: createdAt, UpdatedAt: updatedAt}
	return t, nil
}

func (s *DBTaskService) UpdateStep(taskID string, stepIndex int, newStatus StepStatus, output string, errMsg string) error {
	// Load task first
	t, err := s.GetTask(taskID)
	if err != nil {
		return err
	}
	if stepIndex < 0 || stepIndex >= len(t.Steps) {
		return errors.New("invalid step index")
	}
	t.Steps[stepIndex].Status = newStatus
	t.Steps[stepIndex].Output = output
	t.Steps[stepIndex].Error = errMsg

	// Recompute task state
	allDone := true
	for _, st := range t.Steps {
		if st.Status != StepCompleted {
			allDone = false
			break
		}
	}
	now := time.Now()
	stepsBytes, err := json.Marshal(t.Steps)
	if err != nil {
		return fmt.Errorf("failed to marshal steps: %w", err)
	}
	if allDone {
		t.Status = TaskCompleted
		t.CurrentStepIndex = len(t.Steps)
	} else {
		t.Status = TaskRunning
		// next non-completed step index
		next := 0
		for i, st := range t.Steps {
			if st.Status != StepCompleted {
				next = i
				break
			}
		}
		t.CurrentStepIndex = next
	}
	t.UpdatedAt = now

	// Persist
	_, err = s.db.Exec(`UPDATE tasks SET status = ?, current_step_index = ?, steps = ?, updated_at = ? WHERE id = ?`, string(t.Status), t.CurrentStepIndex, string(stepsBytes), now, taskID)
	return err
}

func (s *DBTaskService) ListTasks() ([]Task, error) {
	rows, err := s.db.Query(`SELECT id, session_id, title, status, current_step_index, steps, created_at, updated_at FROM tasks`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var tid, sessionID, title, statusStr string
		var current int
		var stepsJSON string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&tid, &sessionID, &title, &statusStr, &current, &stepsJSON, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		var steps []Step
		if err := json.Unmarshal([]byte(stepsJSON), &steps); err != nil {
			return nil, err
		}
		tasks = append(tasks, Task{ID: tid, SessionID: sessionID, Title: title, Status: TaskStatus(statusStr), CurrentStepIndex: current, Steps: steps, CreatedAt: createdAt, UpdatedAt: updatedAt})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}
