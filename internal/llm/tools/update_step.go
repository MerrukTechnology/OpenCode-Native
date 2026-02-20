package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MerrukTechnology/OpenCode-Native/internal/task"
)

type UpdateStepParams struct {
	TaskID    string `json:"task_id"`
	StepIndex int    `json:"step_index"`
	Status    string `json:"status"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
}

type UpdateStepResponseMetadata struct {
	TaskID         string `json:"task_id"`
	StepIndex      int    `json:"step_index"`
	PreviousStatus string `json:"previous_status"`
	CurrentStatus  string `json:"current_status"`
	NextStepIndex  int    `json:"next_step_index,omitempty"`
	IsTaskComplete bool   `json:"is_task_complete"`
}

type updateStepTool struct {
	taskService task.Service
}

const (
	UpdateStepToolName        = "update_step"
	updateStepToolDescription = `Updates the status of a task step, marking it as completed, failed, or running.

WHEN TO USE THIS TOOL:
- When you complete a step in a task plan
- When a step fails and you want to record the error
- When you start working on a step to mark it as running
- When you want to add output or results from a step

HOW TO USE:
- Provide the task_id from the plan_task tool
- Provide the step_index (0-based) of the step to update
- Set the status: "completed", "failed", or "running"
- Optionally provide output (results) or error messages

The update_step tool updates the persistent task record and automatically manages task progression.`
)

func NewUpdateStepTool(taskService task.Service) BaseTool {
	return &updateStepTool{
		taskService: taskService,
	}
}

func (u *updateStepTool) Info() ToolInfo {
	return ToolInfo{
		Name:        UpdateStepToolName,
		Description: updateStepToolDescription,
		Parameters: map[string]any{
			"task_id": map[string]any{
				"type":        "string",
				"description": "The task ID returned by plan_task",
			},
			"step_index": map[string]any{
				"type":        "number",
				"description": "The index of the step to update (0-based)",
			},
			"status": map[string]any{
				"type":        "string",
				"enum":        []string{"pending", "running", "completed", "failed"},
				"description": "The new status for the step",
			},
			"output": map[string]any{
				"type":        "string",
				"description": "Optional output or results from the step",
			},
			"error": map[string]any{
				"type":        "string",
				"description": "Optional error message if the step failed",
			},
		},
		Required: []string{"task_id", "step_index", "status"},
	}
}

func (u *updateStepTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
	var params UpdateStepParams
	if err := json.Unmarshal([]byte(call.Input), &params); err != nil {
		return NewTextErrorResponse(fmt.Sprintf("error parsing parameters: %s", err)), nil
	}

	if params.TaskID == "" {
		return NewTextErrorResponse("task_id is required"), nil
	}

	if params.StepIndex < 0 {
		return NewTextErrorResponse("step_index must be non-negative"), nil
	}

	// Get current task to check previous status
	t, err := u.taskService.GetTask(params.TaskID)
	if err != nil {
		return NewEmptyResponse(), fmt.Errorf("failed to get task: %w", err)
	}

	if params.StepIndex >= len(t.Steps) {
		return NewTextErrorResponse(fmt.Sprintf("step_index %d is out of range (task has %d steps)", params.StepIndex, len(t.Steps))), nil
	}

	previousStatus := t.Steps[params.StepIndex].Status

	// Convert string status to task.StepStatus
	var stepStatus task.StepStatus
	switch params.Status {
	case "pending":
		stepStatus = task.StepPending
	case "running":
		stepStatus = task.StepRunning
	case "completed":
		stepStatus = task.StepCompleted
	case "failed":
		stepStatus = task.StepFailed
	default:
		return NewTextErrorResponse(fmt.Sprintf("invalid status: %s (must be pending, running, completed, or failed)", params.Status)), nil
	}

	// Update the step
	if err := u.taskService.UpdateStep(params.TaskID, params.StepIndex, stepStatus, params.Output, params.Error); err != nil {
		return NewEmptyResponse(), fmt.Errorf("failed to update step: %w", err)
	}

	// Get updated task for response
	t, err = u.taskService.GetTask(params.TaskID)
	if err != nil {
		return NewEmptyResponse(), fmt.Errorf("failed to get updated task: %w", err)
	}

	metadata := UpdateStepResponseMetadata{
		TaskID:         params.TaskID,
		StepIndex:      params.StepIndex,
		PreviousStatus: string(previousStatus),
		CurrentStatus:  params.Status,
		NextStepIndex:  t.CurrentStepIndex,
		IsTaskComplete: t.Status == task.TaskCompleted,
	}

	var message string
	if t.Status == task.TaskCompleted {
		message = fmt.Sprintf("Step %d updated to '%s'. Task is now complete!", params.StepIndex, params.Status)
	} else if params.Status == "failed" {
		message = fmt.Sprintf("Step %d marked as failed. Task status: %s", params.StepIndex, t.Status)
	} else {
		message = fmt.Sprintf("Step %d updated to '%s'. Next step: %d", params.StepIndex, params.Status, t.CurrentStepIndex)
	}

	return WithResponseMetadata(
		NewTextResponse(message),
		metadata,
	), nil
}
