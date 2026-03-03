package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sandevgo/tuskbot/internal/core"
)

func TestParseScheduleAdd(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		args      json.RawMessage
		wantErr   bool
		errMsg    string
		wantQuery *ScheduleAddQuery
	}{
		{
			name:    "valid input with all fields",
			args:    json.RawMessage(`{"type":"once","task_name":"test-task","time_spec":"10m","instruction":"do something"}`),
			wantErr: false,
			wantQuery: &ScheduleAddQuery{
				Type:        core.TriggerTypeOnce,
				TaskName:    "test-task",
				TimeSpec:    "10m",
				Instruction: "do something",
			},
		},
		{
			name:    "valid input with cron type",
			args:    json.RawMessage(`{"type":"cron","task_name":"daily-job","time_spec":"0 9 * * *","instruction":"send daily report"}`),
			wantErr: false,
			wantQuery: &ScheduleAddQuery{
				Type:        core.TriggerTypeCron,
				TaskName:    "daily-job",
				TimeSpec:    "0 9 * * *",
				Instruction: "send daily report",
			},
		},
		{
			name:    "valid input with interval type",
			args:    json.RawMessage(`{"type":"interval","task_name":"interval-task","time_spec":"30s","instruction":"check status"}`),
			wantErr: false,
			wantQuery: &ScheduleAddQuery{
				Type:        core.TriggerTypeInterval,
				TaskName:    "interval-task",
				TimeSpec:    "30s",
				Instruction: "check status",
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
			errMsg:  "invalid type",
		},
		{
			name:    "invalid type",
			args:    json.RawMessage(`{"type":"invalid","task_name":"test","time_spec":"10m","instruction":"do it"}`),
			wantErr: true,
			errMsg:  "invalid type",
		},
		{
			name:    "missing type",
			args:    json.RawMessage(`{"task_name":"test","time_spec":"10m","instruction":"do it"}`),
			wantErr: true,
			errMsg:  "invalid type",
		},
		{
			name:    "empty task_name",
			args:    json.RawMessage(`{"type":"once","task_name":"","time_spec":"10m","instruction":"do it"}`),
			wantErr: true,
			errMsg:  "task_name cannot be empty",
		},
		{
			name:    "whitespace only task_name",
			args:    json.RawMessage(`{"type":"once","task_name":"   ","time_spec":"10m","instruction":"do it"}`),
			wantErr: true,
			errMsg:  "task_name cannot be empty",
		},
		{
			name:    "task_name with spaces",
			args:    json.RawMessage(`{"type":"once","task_name":"test task","time_spec":"10m","instruction":"do it"}`),
			wantErr: true,
			errMsg:  "task_name must be a valid slug",
		},
		{
			name:    "task_name with special characters",
			args:    json.RawMessage(`{"type":"once","task_name":"test@task!","time_spec":"10m","instruction":"do it"}`),
			wantErr: true,
			errMsg:  "task_name must be a valid slug",
		},
		{
			name:    "empty time_spec",
			args:    json.RawMessage(`{"type":"once","task_name":"test-task","time_spec":"","instruction":"do it"}`),
			wantErr: true,
			errMsg:  "time_spec cannot be empty",
		},
		{
			name:    "whitespace only time_spec",
			args:    json.RawMessage(`{"type":"once","task_name":"test-task","time_spec":"   ","instruction":"do it"}`),
			wantErr: true,
			errMsg:  "time_spec cannot be empty",
		},
		{
			name:    "empty instruction",
			args:    json.RawMessage(`{"type":"once","task_name":"test-task","time_spec":"10m","instruction":""}`),
			wantErr: true,
			errMsg:  "instruction cannot be empty",
		},
		{
			name:    "whitespace only instruction",
			args:    json.RawMessage(`{"type":"once","task_name":"test-task","time_spec":"10m","instruction":"   "}`),
			wantErr: true,
			errMsg:  "instruction cannot be empty",
		},
		{
			name:    "task_name with underscores and numbers",
			args:    json.RawMessage(`{"type":"once","task_name":"test_task_123","time_spec":"10m","instruction":"do it"}`),
			wantErr: true,
			wantQuery: &ScheduleAddQuery{
				Type:        core.TriggerTypeOnce,
				TaskName:    "test_task_123",
				TimeSpec:    "10m",
				Instruction: "do it",
			},
		},
		{
			name:    "task_name with uppercase",
			args:    json.RawMessage(`{"type":"once","task_name":"Test-Task","time_spec":"10m","instruction":"do it"}`),
			wantErr: false,
			wantQuery: &ScheduleAddQuery{
				Type:        core.TriggerTypeOnce,
				TaskName:    "Test-Task",
				TimeSpec:    "10m",
				Instruction: "do it",
			},
		},
		{
			name:    "trimmed fields",
			args:    json.RawMessage(`{"type":"once","task_name":"  test-task  ","time_spec":"  10m  ","instruction":"  do it  "}`),
			wantErr: false,
			wantQuery: &ScheduleAddQuery{
				Type:        core.TriggerTypeOnce,
				TaskName:    "test-task",
				TimeSpec:    "10m",
				Instruction: "do it",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseScheduleAdd(ctx, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseScheduleAdd() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("parseScheduleAdd() error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("parseScheduleAdd() unexpected error = %v", err)
				return
			}

			if got == nil {
				t.Errorf("parseScheduleAdd() returned nil query without error")
				return
			}

			if got.Type != tt.wantQuery.Type {
				t.Errorf("parseScheduleAdd() Type = %v, want %v", got.Type, tt.wantQuery.Type)
			}
			if got.TaskName != tt.wantQuery.TaskName {
				t.Errorf("parseScheduleAdd() TaskName = %v, want %v", got.TaskName, tt.wantQuery.TaskName)
			}
			if got.TimeSpec != tt.wantQuery.TimeSpec {
				t.Errorf("parseScheduleAdd() TimeSpec = %v, want %v", got.TimeSpec, tt.wantQuery.TimeSpec)
			}
			if got.Instruction != tt.wantQuery.Instruction {
				t.Errorf("parseScheduleAdd() Instruction = %v, want %v", got.Instruction, tt.wantQuery.Instruction)
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
