package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MerrukTechnology/OpenCode-Native/internal/task"
)

type PlanTaskParams struct {
	Title string      `json:"title"`
	Steps []task.Step `json:"steps"`
}

type PlanTaskResponseMetadata struct {
	TaskID string `json:"task_id"`
}

type planTaskTool struct {
	taskService task.Service
}

const (
	PlanTaskToolName        = "plan_task"
	planTaskToolDescription = `Creates a structured multi-step task plan that can be tracked and executed incrementally.

WHEN TO USE THIS TOOL:
- When you need to break down a complex goal into smaller, actionable steps
- When you want to track progress on a multi-stage task
- When you need to coordinate work across multiple subagents with a shared plan
- When you want to enable retry of failed steps without losing progress

HOW TO USE:
- Provide a clear, descriptive title for the overall task
- Break down the work into discrete steps with descriptions
- Each step should be actionable and have a clear completion criteria
- Steps are executed sequentially by default

The plan_task tool creates a persistent task record that survives restarts and can be visualized in the TUI.`
)

func NewPlanTaskTool(taskService task.Service) BaseTool {
	return &planTaskTool{
		taskService: taskService,
	}
}

func (p *planTaskTool) Info() ToolInfo {
	return ToolInfo{
		Name:        PlanTaskToolName,
		Description: planTaskToolDescription,
		Parameters: map[string]any{
			"title": map[string]any{
				"type":        "string",
				"description": "A clear, descriptive title for the task (up to 100 characters)",
			},
			"steps": map[string]any{
				"type":        "array",
				"description": "List of steps to complete the task",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{
							"type":        "string",
							"description": "Unique identifier for this step (e.g., 'step1', 'verify')",
						},
						"description": map[string]any{
							"type":        "string",
							"description": "Clear description of what this step should do",
						},
						"type": map[string]any{
							"type":        "string",
							"enum":        []string{"freeform", "structured"},
							"description": "Step type: 'freeform' for flexible work, 'structured' for well-defined tasks",
						},
					},
					"required": []string{"id", "description"},
				},
			},
		},
		Required: []string{"title", "steps"},
	}
}

func (p *planTaskTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
	var params PlanTaskParams
	if err := json.Unmarshal([]byte(call.Input), &params); err != nil {
		return NewTextErrorResponse(fmt.Sprintf("error parsing parameters: %s", err)), nil
	}

	if params.Title == "" {
		return NewTextErrorResponse("title is required"), nil
	}

	if len(params.Steps) == 0 {
		return NewTextErrorResponse("at least one step is required"), nil
	}

	sessionID, _ := GetContextValues(ctx)
	if sessionID == "" {
		return NewEmptyResponse(), fmt.Errorf("session ID is required to create a task")
	}

	t, err := p.taskService.CreateTask(params.Title, sessionID, params.Steps)
	if err != nil {
		return NewEmptyResponse(), fmt.Errorf("failed to create task: %w", err)
	}

	metadata := PlanTaskResponseMetadata{
		TaskID: t.ID,
	}

	return WithResponseMetadata(
		NewTextResponse(fmt.Sprintf("Task created successfully with %d steps. Task ID: %s", len(params.Steps), t.ID)),
		metadata,
	), nil
}
