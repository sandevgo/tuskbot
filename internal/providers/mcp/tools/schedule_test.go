package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

func TestParseScheduleOnce(t *testing.T) {
	ctx := context.Background()

	// Create a valid RFC3339 timestamp for testing
	validTime := time.Now().Add(1 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		name      string
		args      json.RawMessage
		wantErr   bool
		errMsg    string
		wantQuery *ScheduleOnceQuery
	}{
		{
			name:    "valid input with all fields",
			args:    json.RawMessage(`{"name":"test-task","at":"` + validTime + `","prompt":"do something"}`),
			wantErr: false,
			wantQuery: &ScheduleOnceQuery{
				TaskName: "test-task",
				At:       validTime,
				Prompt:   "do something",
			},
		},
		{
			name:    "valid input with future time",
			args:    json.RawMessage(`{"name":"future-task","at":"2025-12-31T23:59:59Z","prompt":"end of year task"}`),
			wantErr: false,
			wantQuery: &ScheduleOnceQuery{
				TaskName: "future-task",
				At:       "2025-12-31T23:59:59Z",
				Prompt:   "end of year task",
			},
		},
		{
			name:    "invalid json",
			args:    json.RawMessage(`{invalid json}`),
			wantErr: true,
			errMsg:  "invalid arguments",
		},
		{
			name:    "null input",
			args:    json.RawMessage(`null`),
			wantErr: true,
			errMsg:  "arguments cannot be null",
		},
		{
			name:    "empty json object",
			args:    json.RawMessage(`{}`),
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name:    "missing name",
			args:    json.RawMessage(`{"at":"` + validTime + `","prompt":"do it"}`),
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name:    "empty name",
			args:    json.RawMessage(`{"name":"","at":"` + validTime + `","prompt":"do it"}`),
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name:    "whitespace only name",
			args:    json.RawMessage(`{"name":"   ","at":"` + validTime + `","prompt":"do it"}`),
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name:    "name with spaces",
			args:    json.RawMessage(`{"name":"test task","at":"` + validTime + `","prompt":"do it"}`),
			wantErr: true,
			errMsg:  "name must be a valid slug",
		},
		{
			name:    "name with special characters",
			args:    json.RawMessage(`{"name":"test@task!","at":"` + validTime + `","prompt":"do it"}`),
			wantErr: true,
			errMsg:  "name must be a valid slug",
		},
		{
			name:    "name with underscores",
			args:    json.RawMessage(`{"name":"test_task","at":"` + validTime + `","prompt":"do it"}`),
			wantErr: true,
			errMsg:  "name must be a valid slug",
		},
		{
			name:    "name with uppercase",
			args:    json.RawMessage(`{"name":"Test-Task","at":"` + validTime + `","prompt":"do it"}`),
			wantErr: false,
			wantQuery: &ScheduleOnceQuery{
				TaskName: "Test-Task",
				At:       validTime,
				Prompt:   "do it",
			},
		},
		{
			name:    "name with numbers and hyphens",
			args:    json.RawMessage(`{"name":"test-task-123","at":"` + validTime + `","prompt":"do it"}`),
			wantErr: false,
			wantQuery: &ScheduleOnceQuery{
				TaskName: "test-task-123",
				At:       validTime,
				Prompt:   "do it",
			},
		},
		{
			name:    "empty at",
			args:    json.RawMessage(`{"name":"test-task","at":"","prompt":"do it"}`),
			wantErr: true,
			errMsg:  "at cannot be empty",
		},
		{
			name:    "whitespace only at",
			args:    json.RawMessage(`{"name":"test-task","at":"   ","prompt":"do it"}`),
			wantErr: true,
			errMsg:  "at cannot be empty",
		},
		{
			name:    "invalid at format",
			args:    json.RawMessage(`{"name":"test-task","at":"not-a-date","prompt":"do it"}`),
			wantErr: true,
			errMsg:  "at must be a valid RFC3339 timestamp",
		},
		{
			name:    "empty prompt",
			args:    json.RawMessage(`{"name":"test-task","at":"` + validTime + `","prompt":""}`),
			wantErr: true,
			errMsg:  "prompt cannot be empty",
		},
		{
			name:    "whitespace only prompt",
			args:    json.RawMessage(`{"name":"test-task","at":"` + validTime + `","prompt":"   "}`),
			wantErr: true,
			errMsg:  "prompt cannot be empty",
		},
		{
			name:    "trimmed fields",
			args:    json.RawMessage(`{"name":"  test-task  ","at":"  ` + validTime + `  ","prompt":"  do it  "}`),
			wantErr: false,
			wantQuery: &ScheduleOnceQuery{
				TaskName: "test-task",
				At:       validTime,
				Prompt:   "do it",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseScheduleOnce(ctx, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseScheduleOnce() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("parseScheduleOnce() error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("parseScheduleOnce() unexpected error = %v", err)
				return
			}

			if got == nil {
				t.Errorf("parseScheduleOnce() returned nil query without error")
				return
			}

			if got.TaskName != tt.wantQuery.TaskName {
				t.Errorf("parseScheduleOnce() TaskName = %v, want %v", got.TaskName, tt.wantQuery.TaskName)
			}
			if got.At != tt.wantQuery.At {
				t.Errorf("parseScheduleOnce() At = %v, want %v", got.At, tt.wantQuery.At)
			}
			if got.Prompt != tt.wantQuery.Prompt {
				t.Errorf("parseScheduleOnce() Prompt = %v, want %v", got.Prompt, tt.wantQuery.Prompt)
			}
		})
	}
}

func TestParseScheduleCancel(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		args      json.RawMessage
		wantErr   bool
		errMsg    string
		wantQuery *ScheduleCancelQuery
	}{
		{
			name:    "valid input with task_id",
			args:    json.RawMessage(`{"task_id":"test-task"}`),
			wantErr: false,
			wantQuery: &ScheduleCancelQuery{
				TaskName: "test-task",
			},
		},
		{
			name:    "valid input with hyphens",
			args:    json.RawMessage(`{"task_id":"daily-cleanup-job"}`),
			wantErr: false,
			wantQuery: &ScheduleCancelQuery{
				TaskName: "daily-cleanup-job",
			},
		},
		{
			name:    "valid input with numbers",
			args:    json.RawMessage(`{"task_id":"task-123-456"}`),
			wantErr: false,
			wantQuery: &ScheduleCancelQuery{
				TaskName: "task-123-456",
			},
		},
		{
			name:    "valid input with uppercase",
			args:    json.RawMessage(`{"task_id":"Daily-Task"}`),
			wantErr: false,
			wantQuery: &ScheduleCancelQuery{
				TaskName: "Daily-Task",
			},
		},
		{
			name:    "trimmed fields",
			args:    json.RawMessage(`{"task_id":"  test-task  "}`),
			wantErr: false,
			wantQuery: &ScheduleCancelQuery{
				TaskName: "test-task",
			},
		},
		{
			name:    "invalid json",
			args:    json.RawMessage(`{invalid json}`),
			wantErr: true,
			errMsg:  "invalid arguments",
		},
		{
			name:    "null input",
			args:    json.RawMessage(`null`),
			wantErr: true,
			errMsg:  "arguments cannot be null",
		},
		{
			name:    "empty json object",
			args:    json.RawMessage(`{}`),
			wantErr: true,
			errMsg:  "task_id cannot be empty",
		},
		{
			name:    "empty task_id",
			args:    json.RawMessage(`{"task_id":""}`),
			wantErr: true,
			errMsg:  "task_id cannot be empty",
		},
		{
			name:    "whitespace only task_id",
			args:    json.RawMessage(`{"task_id":"   "}`),
			wantErr: true,
			errMsg:  "task_id cannot be empty",
		},
		{
			name:    "task_id with spaces",
			args:    json.RawMessage(`{"task_id":"test task"}`),
			wantErr: true,
			errMsg:  "task_id must be a valid slug",
		},
		{
			name:    "task_id with underscores",
			args:    json.RawMessage(`{"task_id":"test_task"}`),
			wantErr: true,
			errMsg:  "task_id must be a valid slug",
		},
		{
			name:    "task_id with special characters",
			args:    json.RawMessage(`{"task_id":"test@task!"}`),
			wantErr: true,
			errMsg:  "task_id must be a valid slug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseScheduleCancel(ctx, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseScheduleCancel() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("parseScheduleCancel() error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("parseScheduleCancel() unexpected error = %v", err)
				return
			}

			if got == nil {
				t.Errorf("parseScheduleCancel() returned nil query without error")
				return
			}

			if got.TaskName != tt.wantQuery.TaskName {
				t.Errorf("parseScheduleCancel() TaskName = %v, want %v", got.TaskName, tt.wantQuery.TaskName)
			}
		})
	}
}

func TestScheduleGetDefinitions(t *testing.T) {
	swarm := &mockSwarm{}
	s := NewSchedule(swarm)

	defs := s.GetDefinitions()

	// Check that all expected tools are defined
	expectedTools := []string{"schedule_once", "schedule_list", "schedule_cancel"}
	for _, tool := range expectedTools {
		if _, ok := defs[tool]; !ok {
			t.Errorf("GetDefinitions() missing tool: %s", tool)
		}
	}

	// Check schedule_once has required fields
	if defs["schedule_once"].Schema == "" {
		t.Error("schedule_once schema is empty")
	}
	if defs["schedule_once"].Handler == nil {
		t.Error("schedule_once handler is nil")
	}

	// Check schedule_list has required fields
	if defs["schedule_list"].Schema == "" {
		t.Error("schedule_list schema is empty")
	}
	if defs["schedule_list"].Handler == nil {
		t.Error("schedule_list handler is nil")
	}

	// Check schedule_cancel has required fields
	if defs["schedule_cancel"].Schema == "" {
		t.Error("schedule_cancel schema is empty")
	}
	if defs["schedule_cancel"].Handler == nil {
		t.Error("schedule_cancel handler is nil")
	}
}

// Mock implementation for testing
type mockSwarm struct{}

func (m *mockSwarm) ScheduleTask(ctx context.Context, sessionID, name, taskType, timeSpec, instruction string) error {
	return nil
}

func (m *mockSwarm) CancelTask(ctx context.Context, name string) error {
	return nil
}

func (m *mockSwarm) ListTasks(ctx context.Context) ([]core.Task, error) {
	return []core.Task{}, nil
}
