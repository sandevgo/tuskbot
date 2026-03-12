package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/sandevgo/tuskbot/internal/core"
)

type ScheduleOnceQuery struct {
	TaskName string `json:"name"`
	At       string `json:"at"`
	Prompt   string `json:"prompt"`
}

type ScheduleCancelQuery struct {
	TaskName string `json:"task_id"`
}

const scheduleOnce = `
{
  "name": "schedule_once",
  "description": "Schedule a background task to run exactly once at a specific time. Returns the ID of the created task.",
  "inputSchema": {
	  "parameters": {
		"type": "object",
		"properties": {
		  "name": {
			"type": "string",
			"description": "Identifier for the task. Must be a valid slug (lowercase letters, numbers, and hyphens only). Example: 'daily-report-task'."
		  },
		  "prompt": {
			"type": "string",
			"description": "A clear natural language prompt describing exactly what the AI should do when the task is triggered."
		  },
		  "at": {
			"type": "string",
			"description": "Execution time in RFC 3339 format (e.g., 2023-10-25T14:30:00Z). Use UTC if timezone is unknown."
		  }
		},
		"required": ["name", "prompt", "at"]
	  }
  }
}
`

const scheduleCron = `
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

func (s *Schedule) handleOnce(ctx context.Context, args json.RawMessage) (string, error) {
	query, err := parseScheduleOnce(ctx, args)
	if err != nil {
		return "", err
	}

	sessionID, ok := ctx.Value(core.CtxKeySessionID).(string)
	if !ok || sessionID == "" {
		return "", fmt.Errorf("sessionID not found in context")
	}

	err = s.swarm.ScheduleTask(ctx, sessionID, query.TaskName, core.TriggerTypeOnce, query.At, query.Prompt)
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
		"schedule_once": {
			Description: "Schedule background job to run it once",
			Schema:      scheduleOnce,
			Handler:     s.handleOnce,
		},
		//"schedule_cron": {
		//	Description: "Schedule inteval background job",
		//	Schema:      scheduleCron,
		//	Handler:     s.handleOnce,
		//},
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

func parseScheduleOnce(ctx context.Context, args json.RawMessage) (*ScheduleOnceQuery, error) {
	var input *ScheduleOnceQuery
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	if input == nil {
		return nil, fmt.Errorf("arguments cannot be null")
	}

	// Validate TaskName
	input.TaskName = strings.TrimSpace(input.TaskName)
	if input.TaskName == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if !slugRegex.MatchString(input.TaskName) {
		return nil, fmt.Errorf("name must be a valid slug (alphanumeric and hyphens only): %s", input.TaskName)
	}

	// Validate At
	input.At = strings.TrimSpace(input.At)
	if input.At == "" {
		return nil, fmt.Errorf("at cannot be empty")
	}

	// Validate Instruction
	input.Prompt = strings.TrimSpace(input.Prompt)
	if input.Prompt == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
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
