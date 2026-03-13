package eventbus

import (
	"context"
	"sync"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type Event interface {
	EventName() string
}

type Bus struct {
	mu        sync.RWMutex
	subs      map[string][]*subscription
	bufSize   int
	closeOnce sync.Once
	closed    bool
}

type subscription struct {
	ch      chan core.Event
	handler core.EventHandler
	ctx     context.Context
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
}

func New(bufSize int) *Bus {
	return &Bus{
		subs:    make(map[string][]*subscription),
		bufSize: bufSize,
	}
}

func (b *Bus) removeSubscription(eventName string, sub *subscription) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	subs := b.subs[eventName]
	for i, s := range subs {
		if s == sub {
			b.subs[eventName] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(b.subs[eventName]) == 0 {
		delete(b.subs, eventName)
	}
}

func (b *Bus) runHandler(sub *subscription, event Event) {
	defer func() {
		if r := recover(); r != nil {
			log.FromCtx(sub.ctx).
				Error().
				Interface("panic", r).
				Str("event", event.EventName()).
				Msg("event handler panic")
		}
	}()
	sub.handler(sub.ctx, event)
}

func (b *Bus) Subscribe(ctx context.Context, eventName string, handler core.EventHandler) {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}

	subCtx, cancel := context.WithCancel(ctx)

	sub := &subscription{
		ch:      make(chan core.Event, b.bufSize),
		handler: handler,
		ctx:     subCtx,
		cancel:  cancel,
		wg:      &sync.WaitGroup{},
	}

	b.subs[eventName] = append(b.subs[eventName], sub)
	b.mu.Unlock()

	sub.wg.Add(1)
	go func() {
		defer sub.wg.Done()
		defer b.removeSubscription(eventName, sub)

		for {
			select {
			case <-subCtx.Done():
				return
			case event, ok := <-sub.ch:
				if !ok {
					return
				}
				b.runHandler(sub, event)
			}
		}
	}()
}

func (b *Bus) Publish(ctx context.Context, event core.Event) {
	eventName := event.EventName()

	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return
	}
	subs := b.subs[eventName]
	subsCopy := make([]*subscription, len(subs))
	copy(subsCopy, subs)
	b.mu.RUnlock()

	for _, sub := range subsCopy {
		if sub.ctx.Err() != nil {
			continue
		}

		select {
		case sub.ch <- event:
		default:
			log.FromCtx(ctx).
				Warn().
				Str("event", eventName).
				Msg("subscription channel is full, dropping event")
		}
	}
}

func (b *Bus) Close() {
	b.closeOnce.Do(func() {
		b.mu.Lock()
		b.closed = true

		var allSubs []*subscription
		for _, subs := range b.subs {
			allSubs = append(allSubs, subs...)
		}
		b.subs = make(map[string][]*subscription)
		b.mu.Unlock()

		for _, sub := range allSubs {
			sub.cancel()
		}

		for _, sub := range allSubs {
			sub.wg.Wait()
			close(sub.ch)
		}
	})
}
