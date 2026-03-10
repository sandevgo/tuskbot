package eventbus

import (
	"context"
	"sync"
)

type Event interface {
	EventName() string
}

type Handler func(ctx context.Context, event Event)

type Bus struct {
	mu      sync.RWMutex
	subs    map[string][]chan Event
	bufSize int
}

func New(bufSize int) *Bus {
	return &Bus{
		subs:    make(map[string][]chan Event),
		bufSize: bufSize,
	}
}

func (b *Bus) Subscribe(eventName string) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, b.bufSize)
	b.subs[eventName] = append(b.subs[eventName], ch)
	return ch
}

func (b *Bus) Publish(ctx context.Context, event Event) {
	b.mu.RLock()
	subs := b.subs[event.EventName()]
	b.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
			// subscriber is lagging, drop or log the event
		}
	}
}

func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, subs := range b.subs {
		for _, ch := range subs {
			close(ch)
		}
	}
	b.subs = make(map[string][]chan Event)
}
