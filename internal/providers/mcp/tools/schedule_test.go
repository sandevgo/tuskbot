package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseScheduleOnce(t *testing.T) {
	tests := []struct {
		name      string
		args      string
		wantQuery *ScheduleOnceQuery
		wantErr   bool
	}{
		{
			name: "valid once task",
			args: `{"name": "test-task", "at": "2023-10-25T14:30:00Z", "prompt": "do something"}`,
			wantQuery: &ScheduleOnceQuery{
				TaskName: "test-task",
				At:       "2023-10-25T14:30:00Z",
				Prompt:   "do something",
			},
			wantErr: false,
		},
		{
			name:    "empty name",
			args:    `{"name": "", "at": "2023-10-25T14:30:00Z", "prompt": "do something"}`,
			wantErr: true,
		},
		{
			name:    "invalid name slug",
			args:    `{"name": "test task!", "at": "2023-10-25T14:30:00Z", "prompt": "do something"}`,
			wantErr: true,
		},
		{
			name:    "invalid date",
			args:    `{"name": "test-task", "at": "invalid-date", "prompt": "do something"}`,
			wantErr: true,
		},
		{
			name:    "empty prompt",
			args:    `{"name": "test-task", "at": "2023-10-25T14:30:00Z", "prompt": ""}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseScheduleOnce(context.Background(), json.RawMessage(tt.args))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantQuery, got)
		})
	}
}

func TestParseScheduleCron(t *testing.T) {
	tests := []struct {
		name      string
		args      string
		wantQuery *ScheduleOnceQuery
		wantErr   bool
	}{
		{
			name: "valid cron task",
			args: `{"name": "cron-task", "at": "0 9 * * * *", "prompt": "daily report"}`,
			wantQuery: &ScheduleOnceQuery{
				TaskName: "cron-task",
				At:       "0 9 * * * *",
				Prompt:   "daily report",
			},
			wantErr: false,
		},
		{
			name:    "invalid cron expression",
			args:    `{"name": "cron-task", "at": "invalid-cron", "prompt": "daily report"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseScheduleCron(context.Background(), json.RawMessage(tt.args))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantQuery, got)
		})
	}
}

func TestParseScheduleCancel(t *testing.T) {
	tests := []struct {
		name      string
		args      string
		wantQuery *ScheduleCancelQuery
		wantErr   bool
	}{
		{
			name: "valid cancel task",
			args: `{"task_id": "550e8400-e29b-41d4-a716-446655440000"}`,
			wantQuery: &ScheduleCancelQuery{
				TaskID: "550e8400-e29b-41d4-a716-446655440000",
			},
			wantErr: false,
		},
		{
			name:    "empty task_id",
			args:    `{"task_id": ""}`,
			wantErr: true,
		},
		{
			name:    "missing task_id",
			args:    `{}`,
			wantErr: true,
		},
		{
			name:    "invalid json",
			args:    `{invalid}`,
			wantErr: true,
		},
		{
			name:    "null arguments",
			args:    `null`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseScheduleCancel(context.Background(), json.RawMessage(tt.args))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantQuery, got)
			assert.Equal(t, tt.wantQuery.TaskID, got.TaskID)
		})
	}
}
