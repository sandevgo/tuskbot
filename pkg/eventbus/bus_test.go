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

	Subscribe(bus, ctx, "test.event", func(ctx context.Context, e TestEvent) {
		received <- e
	})

	event := TestEvent{Name: "test", Value: 42}
	Publish(bus, ctx, event)

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
	handler := func(ctx context.Context, e TestEvent) {
		counter.Add(1)
	}

	Subscribe(bus, ctx, "test.event", handler)
	Subscribe(bus, ctx, "test.event", handler)
	Subscribe(bus, ctx, "test.event", handler)

	Publish(bus, ctx, TestEvent{Name: "multi", Value: 1})

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

	Subscribe(bus, ctx, "test.event", func(ctx context.Context, e TestEvent) {
		testReceived <- e
	})

	Subscribe(bus, ctx, "other.event", func(ctx context.Context, e OtherEvent) {
		otherReceived <- e
	})

	// Publish different event types
	Publish(bus, ctx, TestEvent{Name: "test", Value: 1})
	Publish(bus, ctx, OtherEvent{Data: "other"})

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
	handler := func(ctx context.Context, e TestEvent) {
		received <- e
	}

	Subscribe(bus, ctx, "test.event", handler)

	// First event should work
	Publish(bus, context.Background(), TestEvent{Name: "before", Value: 1})
	
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
	Publish(bus, context.Background(), TestEvent{Name: "after", Value: 2})
	
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

	Subscribe(bus, ctx, "test.event", func(ctx context.Context, e TestEvent) {
		received <- e
	})

	bus.Close()

	// Should not panic and should not deliver
	Publish(bus, ctx, TestEvent{Name: "closed", Value: 1})

	select {
	case <-received:
		t.Fatal("should not receive event on closed bus")
	case <-time.After(100 * time.Millisecond):
		// expected
	}

	// Subscribe on closed bus should not panic
	Subscribe(bus, ctx, "test.event", func(ctx context.Context, e TestEvent) {})
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
	handler := func(ctx context.Context, e TestEvent) {
		counter.Add(1)
	}

	// Start subscribers
	for i := 0; i < numSubscribers; i++ {
		Subscribe(bus, ctx, "test.event", handler)
	}

	// Concurrent publishes
	var wg sync.WaitGroup
	for i := 0; i < numPublishes; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			Publish(bus, ctx, TestEvent{Name: "concurrent", Value: n})
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

	Subscribe(bus, ctx, "test.event", func(ctx context.Context, e TestEvent) {
		if e.Name == "panic" {
			panic("intentional panic")
		}
		recovered.Store(true)
	})

	Subscribe(bus, ctx, "test.event", func(ctx context.Context, e TestEvent) {
		afterPanic.Store(true)
	})

	// First event causes panic in first handler
	Publish(bus, ctx, TestEvent{Name: "panic", Value: 1})
	time.Sleep(100 * time.Millisecond)

	// Second event should still be processed by both handlers
	Publish(bus, ctx, TestEvent{Name: "normal", Value: 2})
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
	Subscribe(bus, ctx, "test.event", func(ctx context.Context, e TestEvent) {
		<-blocking // Block forever
	})

	// Fill the channel
	Publish(bus, ctx, TestEvent{Name: "first", Value: 1})
	
	// This should be dropped (channel full, handler blocked)
	Publish(bus, ctx, TestEvent{Name: "dropped", Value: 2})

	// Unblock to verify first was received
	close(blocking)
	time.Sleep(50 * time.Millisecond)
}
