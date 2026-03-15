package command

import (
	"context"
	"fmt"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

type TasksCommand struct {
	swarm     core.Swarm
	formatter *ResponseFormatter
}

func NewTasksCommand(swarm core.Swarm) core.Command {
	return &TasksCommand{
		swarm:     swarm,
		formatter: NewResponseFormatter(),
	}
}

func (c *TasksCommand) Name() string {
	return "tasks"
}

func (c *TasksCommand) Description() string {
	return "List active scheduled tasks"
}

func (c *TasksCommand) Execute(ctx context.Context, sessionID string, args []string) (string, error) {
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
