package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/sandevgo/tuskbot/internal/core"
)

type ScheduleAddQuery struct {
	Type        string `json:"type"`
	TaskName    string `json:"task_name"`
	TimeSpec    string `json:"time_spec"`
	Instruction string `json:"instruction"`
}

type ScheduleCancelQuery struct {
	TaskName string `json:"task_id"`
}

const scheduleAddSchema = `
{
  "name": "schedule_add",
  "description": "Schedules a task for background execution. All parameters are strictly required. Returns the ID of the created task.",
  "inputSchema": {
	  "parameters": {
		"type": "object",
		"properties": {
		  "task_name": {
			"type": "string",
			"description": "Unique identifier for the task. Must be a valid slug (lowercase letters, numbers, and hyphens only). Example: 'daily-report-task'."
		  },
		  "instruction": {
			"type": "string",
			"description": "A clear natural language prompt describing exactly what the AI should do when the task is triggered."
		  },
		  "type": {
			"type": "string",
			"enum": ["once", "cron"],
			"description": "The scheduling strategy. Use 'once' for a single execution, 'cron' for recurring tasks."
		  },
		  "time_spec": {
			"type": "string",
			"description": "For 'once': a duration string (e.g., '30s', '5m', '2h', '1d') or an RFC3339 timestamp. For 'cron': a standard cron expression (e.g., '0 9 * * *')."
		  }
		},
		"required": ["task_name", "instruction", "type", "time_spec"]
	  }
  }
}
`

const scheduleListSchema = `
{
  "name": "schedule_list",
  "description": "Retrieves information about all currently scheduled background tasks. Returns an array of task objects, including their ID, name, current status, type, and the original instruction."
}
`

const scheduleCancelSchema = `
{
  "name": "schedule_cancel",
  "description": "Cancels and permanently removes a scheduled task by its unique identifier. This stops future executions, including recurring 'cron' tasks.",
  "parameters": {
    "type": "object",
    "properties": {
      "task_name": {
        "type": "string",
        "description": "The unique slug-formatted name of the task to be cancelled (e.g., 'write-hokku-task')."
      }
    },
    "required": ["task_name"]
  }
}
`

type Schedule struct {
	swarm core.Swarm
}

func NewSchedule(swarm core.Swarm) *Schedule {
	return &Schedule{
		swarm: swarm,
	}
}

func (s *Schedule) handleAdd(ctx context.Context, args json.RawMessage) (string, error) {
	query, err := parseScheduleAdd(ctx, args)
	if err != nil {
		return "", err
	}

	// Get sessionID from context
	sessionID, ok := ctx.Value(core.CtxKeySessionID).(string)
	if !ok || sessionID == "" {
		return "", fmt.Errorf("sessionID not found in context")
	}

	err = s.swarm.ScheduleTask(ctx, sessionID, query.TaskName, query.Type, query.TimeSpec, query.Instruction)
	if err != nil {
		return "", err
	}

	return "", nil
}

func (s *Schedule) handleList(ctx context.Context, args json.RawMessage) (string, error) {
	tasks, err := s.swarm.ListTasks(ctx)
	if err != nil {
		return "", err
	}

	sb := strings.Builder{}
	for _, task := range tasks {
		sb.WriteString(fmt.Sprintf("Task ID: %s Name: %s\n", task.ID, task.Name))
	}
	return sb.String(), nil
}

func (s *Schedule) handleCancel(ctx context.Context, args json.RawMessage) (string, error) {
	query, err := parseScheduleCancel(ctx, args)
	if err != nil {
		return "", err
	}

	err = s.swarm.CancelTask(ctx, query.TaskName)
	if err != nil {
		return "", err
	}

	return "", nil
}

func (s *Schedule) GetDefinitions() map[string]struct {
	Description string
	Schema      string
	Handler     core.NativeHandler
} {
	return map[string]struct {
		Description string
		Schema      string
		Handler     core.NativeHandler
	}{
		"schedule_add": {
			Description: "Schedule background job",
			Schema:      scheduleAddSchema,
			Handler:     s.handleAdd,
		},
		"schedule_list": {
			Description: "Get list of scheduled jobs",
			Schema:      scheduleListSchema,
			Handler:     s.handleList,
		},
		"schedule_cancel": {
			Description: "Cancel and remove a scheduled job by name",
			Schema:      scheduleCancelSchema,
			Handler:     s.handleCancel,
		},
	}
}

// slugRegex matches valid task names: alphanumeric, hyphens, and underscores only
var slugRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

func parseScheduleAdd(ctx context.Context, args json.RawMessage) (*ScheduleAddQuery, error) {
	var input *ScheduleAddQuery
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	if input == nil {
		return nil, fmt.Errorf("arguments cannot be null")
	}

	validType := []string{
		core.TriggerTypeCron,
		core.TriggerTypeOnce,
		core.TriggerTypeInterval,
	}

	if input.Type == "" {
		return nil, fmt.Errorf("type is required")
	}

	if !slices.Contains(validType, input.Type) {
		return nil, fmt.Errorf("invalid type: %s", input.Type)
	}

	// Validate TaskName
	input.TaskName = strings.TrimSpace(input.TaskName)
	if input.TaskName == "" {
		return nil, fmt.Errorf("task_name cannot be empty")
	}
	if !slugRegex.MatchString(input.TaskName) {
		return nil, fmt.Errorf("task_name must be a valid slug (alphanumeric and hyphens only): %s", input.TaskName)
	}

	// Validate TimeSpec
	input.TimeSpec = strings.TrimSpace(input.TimeSpec)
	if input.TimeSpec == "" {
		return nil, fmt.Errorf("time_spec cannot be empty")
	}

	// Validate Instruction
	input.Instruction = strings.TrimSpace(input.Instruction)
	if input.Instruction == "" {
		return nil, fmt.Errorf("instruction cannot be empty")
	}

	return input, nil
}

func parseScheduleCancel(ctx context.Context, args json.RawMessage) (*ScheduleCancelQuery, error) {
	var input *ScheduleCancelQuery
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	if input == nil {
		return nil, fmt.Errorf("arguments cannot be null")
	}

	// Validate TaskName (task_id in JSON)
	input.TaskName = strings.TrimSpace(input.TaskName)
	if input.TaskName == "" {
		return nil, fmt.Errorf("task_id cannot be empty")
	}
	if !slugRegex.MatchString(input.TaskName) {
		return nil, fmt.Errorf("task_id must be a valid slug (alphanumeric and hyphens only): %s", input.TaskName)
	}

	return input, nil
}
