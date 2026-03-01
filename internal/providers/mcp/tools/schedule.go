package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/sandevgo/tuskbot/internal/core"
)

type ScheduleAddQuery struct {
	Type        string `json:"type"`
	TaskName    string `json:"task_name"`
	TimeSpec    string `json:"time_spec"`
	Instruction string `json:"instruction"`
}

const scheduleAddSchema = `
{
  "name": "schedule_add",
  "description": "Запланировать выполнение новой задачи (напоминание или регулярное действие).",
  "parameters": {
    "type": "object",
    "properties": {
      "type": {
        "type": "string",
        "enum": ["once", "interval"],
        "description": "Тип задачи: 'once' (один раз) или 'interval' (повторять)."
      },
      "time_spec": {
        "type": "string",
        "description": "Для 'once': дата ISO8601 (2023-10-27T15:00:00Z) или задержка (10m, 2h). Для 'interval': период повторения (30s, 10m, 24h)."
      },
      "task_name": {
        "type": "string",
        "description": "Уникальное короткое название задачи (slug), без пробелов (напр: 'daily-digest', 'remind-milk')."
      },
      "instruction": {
        "type": "string",
        "description": "Подробная инструкция для агента, который проснется в будущем. Что именно нужно сделать?"
      }
    },
    "required": ["type", "time_spec", "task_name", "instruction"]
  }
}
`

const scheduleListSchema = `
{
  "name": "schedule_list",
  "description": "Получить список всех активных запланированных задач.",
  "parameters": {
    "type": "object",
    "properties": {},
    "required": []
  }
}
`

const scheduleCancelSchema = `
{
  "name": "schedule_cancel",
  "description": "Отменить и удалить запланированную задачу по её названию.",
  "parameters": {
    "type": "object",
    "properties": {
      "task_id": {
        "type": "string",
        "description": "Название задачи (task_name), которую нужно удалить."
      }
    },
    "required": ["task_id"]
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

	if query.Type == "" {

	}

	return "", nil
}

func (s *Schedule) handleList(ctx context.Context, args json.RawMessage) (string, error) {
	return "", nil
}

func (s *Schedule) handleCancel(ctx context.Context, args json.RawMessage) (string, error) {
	return "", nil
}

func (m *Schedule) GetDefinitions() map[string]struct {
	Description string
	Schema      string
	Handler     func(context.Context, json.RawMessage) (string, error)
} {
	return map[string]struct {
		Description string
		Schema      string
		Handler     func(context.Context, json.RawMessage) (string, error)
	}{
		"schedule_add": {
			Description: "Schedule background job",
			Schema:      scheduleAddSchema,
			Handler:     m.handleAdd,
		},
		"schedule_list": {
			Description: "handleList scheduled jobs",
			Schema:      scheduleListSchema,
			Handler:     m.handleList,
		},
		"schedule_cancel": {
			Description: "handleCancel and remove a scheduled job by name",
			Schema:      scheduleCancelSchema,
			Handler:     m.handleCancel,
		},
	}
}

func parseScheduleAdd(ctx context.Context, args json.RawMessage) (*ScheduleAddQuery, error) {
	var input *ScheduleAddQuery
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	validType := []string{
		core.TriggerTypeCron,
		core.TriggerTypeOnce,
		core.TriggerTypeInterval,
	}

	if !slices.Contains(validType, input.Type) {
		return nil, fmt.Errorf("invalid type: %s", input.Type)
	}

	return input, nil
}
