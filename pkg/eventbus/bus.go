package eventbus

import (
	"context"
	"sync"

	"github.com/sandevgo/tuskbot/pkg/log"
)

type Event interface {
	EventName() string
}

// Bus is the event bus implementation
type Bus struct {
	mu        sync.RWMutex
	subs      map[string][]*subscription
	bufSize   int
	closeOnce sync.Once
	closed    bool
}

type subscription struct {
	ch      chan Event
	handler func(context.Context, Event)
	ctx     context.Context
	cancel  context.CancelFunc
	// wg tracks the goroutine for this subscription
	wg *sync.WaitGroup
}

// New creates a new Bus with the given buffer size for channels
func New(bufSize int) *Bus {
	return &Bus{
		subs:    make(map[string][]*subscription),
		bufSize: bufSize,
	}
}

// Subscribe registers a handler for a specific event type.
// Generic type T ensures compile-time type safety for handlers (Point 9).
func Subscribe[T Event](b *Bus, ctx context.Context, eventName string, handler func(context.Context, T)) {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}

	subCtx, cancel := context.WithCancel(ctx)
	
	// Create a wrapped handler that type-asserts safely
	wrappedHandler := func(ctx context.Context, e Event) {
		if typed, ok := e.(T); ok {
			handler(ctx, typed)
		}
	}

	sub := &subscription{
		ch:      make(chan Event, b.bufSize),
		handler: wrappedHandler,
		ctx:     subCtx,
		cancel:  cancel,
		wg:      &sync.WaitGroup{},
	}

	b.subs[eventName] = append(b.subs[eventName], sub)
	b.mu.Unlock()

	sub.wg.Add(1)
	go func() {
		defer sub.wg.Done()
		
		// Point 2: Cleanup subscription from map when goroutine exits
		defer func() {
			b.mu.Lock()
			if b.closed {
				b.mu.Unlock()
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
			b.mu.Unlock()
		}()

		for {
			select {
			case <-subCtx.Done(): // Point 1: Use subCtx.Done() instead of ctx.Done()
				return
			case event, ok := <-sub.ch:
				if !ok {
					return
				}
				// Point 4: Panic recovery to prevent one bad handler from crashing the bus
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.FromCtx(subCtx).
								Error().
								Interface("panic", r).
								Str("event", event.EventName()).
								Msg("event handler panic")
						}
					}()
					// Point 6: Pass subCtx to handler to respect subscription lifecycle
					sub.handler(subCtx, event)
				}()
			}
		}
	}()
}

// Publish sends an event to all subscribers.
// Generic type T ensures compile-time type safety (Point 9).
func Publish[T Event](b *Bus, ctx context.Context, event T) {
	eventName := event.EventName()
	
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return
	}
	// Copy slice to avoid holding lock during send
	subs := b.subs[eventName]
	subsCopy := make([]*subscription, len(subs))
	copy(subsCopy, subs)
	b.mu.RUnlock()

	for _, sub := range subsCopy {
		// Skip if subscription is already canceled
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

// Close shuts down the bus and waits for all goroutines to finish (Point 5).
// It is safe to call Close multiple times (Point 3).
func (b *Bus) Close() {
	b.closeOnce.Do(func() {
		b.mu.Lock()
		b.closed = true
		
		// Collect all subscriptions
		var allSubs []*subscription
		for _, subs := range b.subs {
			allSubs = append(allSubs, subs...)
		}
		// Clear the map immediately to prevent new operations
		b.subs = make(map[string][]*subscription)
		b.mu.Unlock()
		
		// Cancel all subscription contexts
		for _, sub := range allSubs {
			sub.cancel()
		}
		
		// Wait for all goroutines to finish
		for _, sub := range allSubs {
			sub.wg.Wait()
			close(sub.ch)
		}
	})
}
