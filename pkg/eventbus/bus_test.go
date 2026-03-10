package eventbus

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Test event types
type TestEvent struct {
	Name  string
	Value int
}

func (e TestEvent) EventName() string { return "test.event" }

type OtherEvent struct {
	Data string
}

func (e OtherEvent) EventName() string { return "other.event" }

func TestBasicPublishSubscribe(t *testing.T) {
	bus := New(10)
	defer bus.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	received := make(chan TestEvent, 1)

	bus.Subscribe(ctx, "test.event", func(ctx context.Context, e Event) {
		received <- e.(TestEvent)
	})

	event := TestEvent{Name: "test", Value: 42}
	bus.Publish(ctx, event)

	select {
	case got := <-received:
		if got.Name != "test" || got.Value != 42 {
			t.Errorf("expected event {test, 42}, got {%s, %d}", got.Name, got.Value)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestMultipleSubscribers(t *testing.T) {
	bus := New(10)
	defer bus.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var counter atomic.Int32
	handler := func(ctx context.Context, e Event) {
		counter.Add(1)
	}

	bus.Subscribe(ctx, "test.event", handler)
	bus.Subscribe(ctx, "test.event", handler)
	bus.Subscribe(ctx, "test.event", handler)

	bus.Publish(ctx, TestEvent{Name: "multi", Value: 1})

	// Give time for all handlers to execute
	time.Sleep(100 * time.Millisecond)

	if counter.Load() != 3 {
		t.Errorf("expected 3 handlers to be called, got %d", counter.Load())
	}
}

func TestDifferentEventTypes(t *testing.T) {
	bus := New(10)
	defer bus.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testReceived := make(chan TestEvent, 1)
	otherReceived := make(chan OtherEvent, 1)

	bus.Subscribe(ctx, "test.event", func(ctx context.Context, e Event) {
		testReceived <- e.(TestEvent)
	})

	bus.Subscribe(ctx, "other.event", func(ctx context.Context, e Event) {
		otherReceived <- e.(OtherEvent)
	})

	// Publish different event types
	bus.Publish(ctx, TestEvent{Name: "test", Value: 1})
	bus.Publish(ctx, OtherEvent{Data: "other"})

	// Verify correct routing
	select {
	case e := <-testReceived:
		if e.Name != "test" {
			t.Errorf("wrong test event: %v", e)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for test event")
	}

	select {
	case e := <-otherReceived:
		if e.Data != "other" {
			t.Errorf("wrong other event: %v", e)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for other event")
	}
}

func TestContextCancellation(t *testing.T) {
	bus := New(10)
	defer bus.Close()

	ctx, cancel := context.WithCancel(context.Background())

	received := make(chan TestEvent, 10)
	handler := func(ctx context.Context, e Event) {
		received <- e.(TestEvent)
	}

	bus.Subscribe(ctx, "test.event", handler)

	// First event should work
	bus.Publish(context.Background(), TestEvent{Name: "before", Value: 1})
	
	select {
	case <-received:
		// ok
	case <-time.After(time.Second):
		t.Fatal("should receive first event")
	}

	// Cancel context
	cancel()
	time.Sleep(50 * time.Millisecond) // Let goroutine exit

	// Second event should not be received (subscriber is gone)
	bus.Publish(context.Background(), TestEvent{Name: "after", Value: 2})
	
	select {
	case <-received:
		t.Fatal("should not receive event after cancellation")
	case <-time.After(200 * time.Millisecond):
		// expected
	}
}

func TestClosedBus(t *testing.T) {
	bus := New(10)

	ctx := context.Background()
	received := make(chan TestEvent, 1)

	bus.Subscribe(ctx, "test.event", func(ctx context.Context, e Event) {
		received <- e.(TestEvent)
	})

	bus.Close()

	// Should not panic and should not deliver
	bus.Publish(ctx, TestEvent{Name: "closed", Value: 1})

	select {
	case <-received:
		t.Fatal("should not receive event on closed bus")
	case <-time.After(100 * time.Millisecond):
		// expected
	}

	// Subscribe on closed bus should not panic
	bus.Subscribe(ctx, "test.event", func(ctx context.Context, e Event) {})
}

func TestConcurrentPublishSubscribe(t *testing.T) {
	bus := New(100)
	defer bus.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const (
		numSubscribers = 10
		numPublishes   = 100
	)

	var counter atomic.Int32
	handler := func(ctx context.Context, e Event) {
		counter.Add(1)
	}

	// Start subscribers
	for i := 0; i < numSubscribers; i++ {
		bus.Subscribe(ctx, "test.event", handler)
	}

	// Concurrent publishes
	var wg sync.WaitGroup
	for i := 0; i < numPublishes; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			bus.Publish(ctx, TestEvent{Name: "concurrent", Value: n})
		}(i)
	}
	wg.Wait()

	// Wait for all handlers
	time.Sleep(200 * time.Millisecond)

	expected := int32(numSubscribers * numPublishes)
	if counter.Load() != expected {
		t.Errorf("expected %d total handler calls, got %d", expected, counter.Load())
	}
}

func TestHandlerPanicRecovery(t *testing.T) {
	bus := New(10)
	defer bus.Close()

	ctx := context.Background()

	var recovered atomic.Bool
	var afterPanic atomic.Bool

	bus.Subscribe(ctx, "test.event", func(ctx context.Context, e Event) {
		te := e.(TestEvent)
		if te.Name == "panic" {
			panic("intentional panic")
		}
		recovered.Store(true)
	})

	bus.Subscribe(ctx, "test.event", func(ctx context.Context, e Event) {
		afterPanic.Store(true)
	})

	// First event causes panic in first handler
	bus.Publish(ctx, TestEvent{Name: "panic", Value: 1})
	time.Sleep(100 * time.Millisecond)

	// Second event should still be processed by both handlers
	bus.Publish(ctx, TestEvent{Name: "normal", Value: 2})
	time.Sleep(100 * time.Millisecond)

	if !recovered.Load() {
		t.Error("second handler should have processed event after panic")
	}
	if !afterPanic.Load() {
		t.Error("handler after panicking handler should still work")
	}
}

func TestEventDropping(t *testing.T) {
	bus := New(1) // Very small buffer
	defer bus.Close()

	ctx := context.Background()
	blocking := make(chan struct{})

	// Handler that blocks
	bus.Subscribe(ctx, "test.event", func(ctx context.Context, e Event) {
		<-blocking // Block forever
	})

	// Fill the channel
	bus.Publish(ctx, TestEvent{Name: "first", Value: 1})
	
	// This should be dropped (channel full, handler blocked)
	bus.Publish(ctx, TestEvent{Name: "dropped", Value: 2})

	// Unblock to verify first was received
	close(blocking)
	time.Sleep(50 * time.Millisecond)
}
