package eventbus

import (
	"context"
	"sync"

	"github.com/sandevgo/tuskbot/pkg/log"
)

type Event interface {
	EventName() string
}

type Handler func(ctx context.Context, event Event)

type subscription struct {
	ch      chan Event
	handler Handler
	ctx     context.Context
	cancel  context.CancelFunc
}

type Bus struct {
	mu      sync.RWMutex
	subs    map[string][]*subscription
	bufSize int
}

func New(bufSize int) *Bus {
	return &Bus{
		subs:    make(map[string][]*subscription),
		bufSize: bufSize,
	}
}

func (b *Bus) Subscribe(ctx context.Context, eventName string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subCtx, cancel := context.WithCancel(ctx)
	sub := &subscription{
		ch:      make(chan Event, b.bufSize),
		handler: handler,
		ctx:     subCtx,
		cancel:  cancel,
	}

	b.subs[eventName] = append(b.subs[eventName], sub)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-sub.ch:
				if !ok {
					return
				}
				handler(ctx, event)
			}
		}
	}()
}

func (b *Bus) Publish(ctx context.Context, event Event) {
	b.mu.RLock()
	subs := b.subs[event.EventName()]
	b.mu.RUnlock()

	for _, sub := range subs {
		select {
		case sub.ch <- event:
		default:
			log.FromCtx(ctx).
				Warn().
				Str("event", event.EventName()).
				Msg("subscription channel is full, dropping event")
		}
	}
}

func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, subs := range b.subs {
		for _, sub := range subs {
			sub.cancel()
			close(sub.ch)
		}
	}
	b.subs = make(map[string][]*subscription)
}
