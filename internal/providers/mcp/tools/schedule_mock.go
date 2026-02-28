package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type MockScheduler struct{}

func NewMockScheduler() *MockScheduler {
	return &MockScheduler{}
}

func (m *MockScheduler) GetDefinitions() map[string]struct {
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
			Description: "Mock: Запланировать выполнение новой задачи (напоминание или регулярное действие).",
			Schema:      scheduleAddSchema,
			Handler:     m.handleAdd,
		},
		"schedule_list": {
			Description: "Mock: Получить список всех активных запланированных задач.",
			Schema:      scheduleListSchema,
			Handler:     m.handleList,
		},
		"schedule_cancel": {
			Description: "Mock: Отменить и удалить запланированную задачу по её названию.",
			Schema:      scheduleCancelSchema,
			Handler:     m.handleCancel,
		},
	}
}

func (m *MockScheduler) handleAdd(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		Type        string `json:"type"`
		TimeSpec    string `json:"time_spec"`
		TaskName    string `json:"task_name"`
		Instruction string `json:"instruction"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	return fmt.Sprintf("Mock: Task '%s' scheduled (type: %s, spec: %s)", input.TaskName, input.Type, input.TimeSpec), nil
}

func (m *MockScheduler) handleList(ctx context.Context, args json.RawMessage) (string, error) {
	return "Mock: No scheduled tasks (mock mode)", nil
}

func (m *MockScheduler) handleCancel(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	return fmt.Sprintf("Mock: Task '%s' cancelled", input.TaskID), nil
}
