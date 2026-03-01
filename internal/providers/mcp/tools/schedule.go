package tools

import (
	"context"
	"encoding/json"

	"github.com/sandevgo/tuskbot/internal/core"
)

const scheduleAddSchema = `
{
  "name": "schedule_add",
  "description": "Запланировать выполнение новой задачи (напоминание или регулярное действие).",
  "parameters": {
    "type": "object",
    "properties": {
      "type": {
        "type": "string",
        "enum": ["one_off", "interval"],
        "description": "Тип задачи: 'one_off' (один раз) или 'interval' (повторять)."
      },
      "time_spec": {
        "type": "string",
        "description": "Для 'one_off': дата ISO8601 (2023-10-27T15:00:00Z) или задержка (10m, 2h). Для 'interval': период повторения (30s, 10m, 24h)."
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

func (s *Schedule) Add(ctx context.Context, args json.RawMessage) (string, error) {
	return "", nil
}

func (s *Schedule) List(ctx context.Context, args json.RawMessage) (string, error) {
	return "", nil
}

func (s *Schedule) Cancel(ctx context.Context, args json.RawMessage) (string, error) {
	return "", nil
}

func (s *Schedule) GetDefinitions() map[string]struct {
	Description string
	Schema      string
	Handler     func(context.Context, json.RawMessage) (string, error)
} {
	return map[string]struct {
		Description string
		Schema      string
		Handler     func(context.Context, json.RawMessage) (string, error)
	}{
		"schedule_add":    {"Schedule job", fetchURLSchema, s.Add},
		"schedule_list":   {"List scheduled jobs", scheduleListSchema, s.List},
		"schedule_cancel": {"Cancel and remove a scheduled job by name", scheduleCancelSchema, s.Cancel},
	}
}
