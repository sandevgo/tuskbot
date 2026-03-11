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
	Publish(ctx context.Context, event Event)
}

type EventSubscriber interface {
	Subscribe(ctx context.Context, eventName string, handler EventHandler)
}

type EventHandler func(ctx context.Context, event Event)

type ChatEvent struct {
	Type      string
	SessionID string
	Message   string
}

func NewChatEvent(t string, sessionID string, message string) *ChatEvent {
	return &ChatEvent{
		Type:      t,
		SessionID: sessionID,
		Message:   message,
	}
}

func (e ChatEvent) EventName() string {
	return e.Type
}
