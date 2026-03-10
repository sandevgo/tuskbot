package core

import (
	"context"
)

const (
	EventTypeTaskCompleted = "task.completed"
)

type Event interface {
	EventName() string
}

type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}
type EventSubscriber interface {
	Subscribe(eventName string, handler EventHandler)
}
type EventHandler func(ctx context.Context, event Event) error

type ChatEvent struct {
	Type    string
	ChatID  int64
	Message string
}

func NewChatEvent(t string, chatID int64, message string) *ChatEvent {
	return &ChatEvent{
		Type:    t,
		ChatID:  chatID,
		Message: message,
	}
}

func (e ChatEvent) EventName() string {
	return e.Type
}
