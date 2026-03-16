package command

import (
	"context"
	"fmt"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

type TaskCommand struct {
	swarm     core.Swarm
	formatter *ResponseFormatter
}

func NewTaskCommand(swarm core.Swarm) core.Command {
	return &TaskCommand{
		swarm:     swarm,
		formatter: NewResponseFormatter(),
	}
}

func (c *TaskCommand) Name() string {
	return "task"
}

func (c *TaskCommand) Description() string {
	return "List active scheduled tasks"
}

func (c *TaskCommand) Execute(ctx context.Context, sessionID string, args []string) (string, error) {
	tasks, err := c.swarm.ListTasks(ctx)
	if err != nil {
		return "", err
	}

	if len(tasks) == 0 {
		return c.formatter.Combine(
			c.formatter.Info("Active Tasks"),
			c.formatter.Label("Status", "No active tasks found."),
		), nil
	}

	var taskList []string
	for _, t := range tasks {
		nextRun := "N/A"
		if t.Trigger != nil {
			nextRun = t.NextFireTime(time.Now()).Format("2006-01-02 15:04")
		}
		taskList = append(taskList, fmt.Sprintf("**%s** (ID: `%s`) - Next: %s", t.Name, t.ID, nextRun))
	}

	return c.formatter.Combine(
		c.formatter.Info("Active Tasks"),
		c.formatter.List(taskList),
	), nil
}
