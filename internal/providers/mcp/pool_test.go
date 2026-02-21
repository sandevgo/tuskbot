package mcp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/mark3labs/mcp-go/client"
)

// mockManagedClient creates a ManagedClient safe for testing (no real connection)
func mockManagedClient(name string) *ManagedClient {
	return &ManagedClient{
		Client: nil,
		name:   name,
	}
}

func mockTransportFactory(transport Transport, err error) TransportFactory {
	return func(t TransportType) (Transport, error) {
		if err != nil {
			return nil, err
		}
		return transport, nil
	}
}

func successTransport(ctx context.Context, cfg ServerConfig) (*client.Client, error) {
	return nil, nil
}

func failTransport(ctx context.Context, cfg ServerConfig) (*client.Client, error) {
	return nil, errors.New("connection failed")
}

func TestPool_Add(t *testing.T) {
	tests := []struct {
		name       string
		factory    TransportFactory
		serverName string
		serverCfg  ServerConfig
		wantErr    bool
		wantInPool bool
	}{
		{
			name:       "successful_add",
			factory:    mockTransportFactory(successTransport, nil),
			serverName: "server1",
			serverCfg:  ServerConfig{Command: "echo"},
			wantErr:    false,
			wantInPool: true,
		},
		{
			name:       "invalid_config_no_command_or_url",
			factory:    mockTransportFactory(successTransport, nil),
			serverName: "server1",
			serverCfg:  ServerConfig{}, // Empty config
			wantErr:    true,
			wantInPool: false,
		},
		{
			name:       "transport_factory_error",
			factory:    mockTransportFactory(nil, errors.New("unsupported transport")),
			serverName: "server1",
			serverCfg:  ServerConfig{Command: "echo"},
			wantErr:    true,
			wantInPool: false,
		},
		{
			name:       "transport_connection_error",
			factory:    mockTransportFactory(failTransport, nil),
			serverName: "server1",
			serverCfg:  ServerConfig{Command: "echo"},
			wantErr:    true,
			wantInPool: false,
		},
		{
			name:       "empty_server_name",
			factory:    mockTransportFactory(successTransport, nil),
			serverName: "",
			serverCfg:  ServerConfig{Command: "echo"},
			wantErr:    false,
			wantInPool: true,
		},
		{
			name:       "unicode_server_name",
			factory:    mockTransportFactory(successTransport, nil),
			serverName: "服务器",
			serverCfg:  ServerConfig{Command: "echo"},
			wantErr:    false,
			wantInPool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPoolWithFactory(tt.factory)
			ctx := context.Background()

			cli, err := p.Add(ctx, tt.serverName, tt.serverCfg)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if cli != nil {
					t.Error("expected nil client on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if cli == nil {
					t.Fatal("expected client, got nil")
				}
			}

			_, inPool := p.Get(tt.serverName)
			if inPool != tt.wantInPool {
				t.Errorf("in pool = %v, want %v", inPool, tt.wantInPool)
			}
		})
	}
}

func TestPool_Add_ReplacesExisting(t *testing.T) {
	var closeCount int32

	p := NewPoolWithFactory(mockTransportFactory(successTransport, nil))
	ctx := context.Background()

	// Add first client
	_, err := p.Add(ctx, "server", ServerConfig{Command: "first"})
	if err != nil {
		t.Fatalf("first add failed: %v", err)
	}

	// Manually track close
	first, _ := p.Get("server")
	go func() {
		// Simulate close being called on old client
		atomic.AddInt32(&closeCount, 1)
	}()

	// Add second client with same name
	_, err = p.Add(ctx, "server", ServerConfig{Command: "second"})
	if err != nil {
		t.Fatalf("second add failed: %v", err)
	}

	// Should only have one client
	if len(p.All()) != 1 {
		t.Errorf("count = %d, want 1", len(p.All()))
	}

	// New client should be different from first
	second, _ := p.Get("server")
	if first == second {
		t.Error("expected new client instance")
	}
}

func TestPool_Get(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(p *Pool)
		getName string
		wantOk  bool
	}{
		{
			name:    "get_from_empty",
			setup:   func(p *Pool) {},
			getName: "any",
			wantOk:  false,
		},
		{
			name: "get_existing",
			setup: func(p *Pool) {
				p.mu.Lock()
				p.clients["server1"] = mockManagedClient("server1")
				p.mu.Unlock()
			},
			getName: "server1",
			wantOk:  true,
		},
		{
			name: "get_nonexistent",
			setup: func(p *Pool) {
				p.mu.Lock()
				p.clients["server1"] = mockManagedClient("server1")
				p.mu.Unlock()
			},
			getName: "server2",
			wantOk:  false,
		},
		{
			name: "get_empty_name",
			setup: func(p *Pool) {
				p.mu.Lock()
				p.clients[""] = mockManagedClient("")
				p.mu.Unlock()
			},
			getName: "",
			wantOk:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPool()
			tt.setup(p)

			cli, ok := p.Get(tt.getName)

			if ok != tt.wantOk {
				t.Errorf("ok = %v, want %v", ok, tt.wantOk)
			}
			if tt.wantOk && cli == nil {
				t.Error("expected client, got nil")
			}
		})
	}
}

func TestPool_Del(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(p *Pool)
		delName   string
		wantErr   bool
		wantCount int
	}{
		{
			name:      "delete_from_empty",
			setup:     func(p *Pool) {},
			delName:   "any",
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "delete_existing",
			setup: func(p *Pool) {
				p.mu.Lock()
				p.clients["server1"] = mockManagedClient("server1")
				p.mu.Unlock()
			},
			delName:   "server1",
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "delete_nonexistent",
			setup: func(p *Pool) {
				p.mu.Lock()
				p.clients["server1"] = mockManagedClient("server1")
				p.mu.Unlock()
			},
			delName:   "server2",
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "delete_one_of_many",
			setup: func(p *Pool) {
				p.mu.Lock()
				p.clients["s1"] = mockManagedClient("s1")
				p.clients["s2"] = mockManagedClient("s2")
				p.clients["s3"] = mockManagedClient("s3")
				p.mu.Unlock()
			},
			delName:   "s2",
			wantErr:   false,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPool()
			tt.setup(p)

			err := p.Del(tt.delName)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if len(p.All()) != tt.wantCount {
				t.Errorf("count = %d, want %d", len(p.All()), tt.wantCount)
			}
		})
	}
}

func TestPool_All(t *testing.T) {
	tests := []struct {
		name  string
		setup func(p *Pool)
		want  int
	}{
		{
			name:  "empty_pool",
			setup: func(p *Pool) {},
			want:  0,
		},
		{
			name: "single_client",
			setup: func(p *Pool) {
				p.mu.Lock()
				p.clients["s1"] = mockManagedClient("s1")
				p.mu.Unlock()
			},
			want: 1,
		},
		{
			name: "multiple_clients",
			setup: func(p *Pool) {
				p.mu.Lock()
				p.clients["s1"] = mockManagedClient("s1")
				p.clients["s2"] = mockManagedClient("s2")
				p.clients["s3"] = mockManagedClient("s3")
				p.mu.Unlock()
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPool()
			tt.setup(p)

			all := p.All()

			if len(all) != tt.want {
				t.Errorf("count = %d, want %d", len(all), tt.want)
			}
		})
	}
}

func TestPool_All_ReturnsCopy(t *testing.T) {
	p := NewPool()
	p.mu.Lock()
	p.clients["server"] = mockManagedClient("server")
	p.mu.Unlock()

	all := p.All()
	all["hacked"] = mockManagedClient("hacked")

	if len(p.All()) != 1 {
		t.Error("All() should return a copy, not reference")
	}
}

func TestPool_Close(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(p *Pool)
		wantEmptyAfter bool
	}{
		{
			name:           "close_empty",
			setup:          func(p *Pool) {},
			wantEmptyAfter: true,
		},
		{
			name: "close_single",
			setup: func(p *Pool) {
				p.mu.Lock()
				p.clients["s1"] = mockManagedClient("s1")
				p.mu.Unlock()
			},
			wantEmptyAfter: true,
		},
		{
			name: "close_multiple",
			setup: func(p *Pool) {
				p.mu.Lock()
				p.clients["s1"] = mockManagedClient("s1")
				p.clients["s2"] = mockManagedClient("s2")
				p.mu.Unlock()
			},
			wantEmptyAfter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPool()
			tt.setup(p)

			err := p.Close()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantEmptyAfter && len(p.All()) != 0 {
				t.Error("pool should be empty after Close")
			}
		})
	}
}

func TestPool_Close_DoubleClose(t *testing.T) {
	p := NewPool()
	p.mu.Lock()
	p.clients["server"] = mockManagedClient("server")
	p.mu.Unlock()

	err1 := p.Close()
	err2 := p.Close()

	if err1 != nil {
		t.Errorf("first close error: %v", err1)
	}
	if err2 != nil {
		t.Errorf("second close error: %v", err2)
	}
}

func TestPool_ConcurrentAccess(t *testing.T) {
	tests := []struct {
		name       string
		readers    int
		writers    int
		deleters   int
		iterations int
	}{
		{
			name:       "light_load",
			readers:    5,
			writers:    2,
			deleters:   1,
			iterations: 20,
		},
		{
			name:       "heavy_reads",
			readers:    20,
			writers:    2,
			deleters:   1,
			iterations: 50,
		},
		{
			name:       "balanced",
			readers:    10,
			writers:    5,
			deleters:   3,
			iterations: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPoolWithFactory(mockTransportFactory(successTransport, nil))
			ctx := context.Background()
			var wg sync.WaitGroup

			// Writers
			for i := 0; i < tt.writers; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < tt.iterations; j++ {
						name := fmt.Sprintf("server-%d", j%5)
						_, _ = p.Add(ctx, name, ServerConfig{Command: "cmd"})
					}
				}(i)
			}

			// Readers
			for i := 0; i < tt.readers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tt.iterations; j++ {
						p.All()
						p.Get("server-0")
					}
				}()
			}

			// Deleters
			for i := 0; i < tt.deleters; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tt.iterations; j++ {
						name := fmt.Sprintf("server-%d", j%5)
						_ = p.Del(name)
					}
				}()
			}

			wg.Wait()
		})
	}
}

func TestPool_ContextCancellation(t *testing.T) {
	p := NewPoolWithFactory(func(tt TransportType) (Transport, error) {
		return func(ctx context.Context, cfg ServerConfig) (*client.Client, error) {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return nil, nil
			}
		}, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := p.Add(ctx, "server", ServerConfig{Command: "cmd"})
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestPool_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "add_same_name_twice",
			test: func(t *testing.T) {
				p := NewPoolWithFactory(mockTransportFactory(successTransport, nil))
				ctx := context.Background()

				first, _ := p.Add(ctx, "server", ServerConfig{Command: "first"})
				second, _ := p.Add(ctx, "server", ServerConfig{Command: "second"})

				if len(p.All()) != 1 {
					t.Errorf("count = %d, want 1", len(p.All()))
				}

				current, _ := p.Get("server")
				if current == first {
					t.Error("should have replaced with new client")
				}
				if current != second {
					t.Error("should return second client")
				}
			},
		},
		{
			name: "get_after_close",
			test: func(t *testing.T) {
				p := NewPool()
				p.mu.Lock()
				p.clients["server"] = mockManagedClient("server")
				p.mu.Unlock()

				_ = p.Close()

				// Pool is emptied after Close
				_, ok := p.Get("server")
				if ok {
					t.Error("Get should return false after Close")
				}
			},
		},
		{
			name: "double_close",
			test: func(t *testing.T) {
				p := NewPool()
				p.mu.Lock()
				p.clients["server"] = mockManagedClient("server")
				p.mu.Unlock()

				err1 := p.Close()
				err2 := p.Close()

				if err1 != nil {
					t.Errorf("first close error: %v", err1)
				}
				if err2 != nil {
					t.Errorf("second close error: %v", err2)
				}
			},
		},
		{
			name: "unicode_server_name",
			test: func(t *testing.T) {
				p := NewPoolWithFactory(mockTransportFactory(successTransport, nil))
				ctx := context.Background()

				_, err := p.Add(ctx, "服务器", ServerConfig{Command: "cmd"})
				if err != nil {
					t.Fatalf("add failed: %v", err)
				}

				cli, ok := p.Get("服务器")
				if !ok {
					t.Error("unicode name should work")
				}
				if cli == nil {
					t.Error("expected client")
				}
			},
		},
		{
			name: "special_characters_in_name",
			test: func(t *testing.T) {
				p := NewPoolWithFactory(mockTransportFactory(successTransport, nil))
				ctx := context.Background()

				names := []string{
					"server/with/slashes",
					"server.with.dots",
					"server:with:colons",
					"server with spaces",
				}

				for _, name := range names {
					if _, err := p.Add(ctx, name, ServerConfig{Command: "cmd"}); err != nil {
						t.Errorf("failed to add %q: %v", name, err)
					}
				}

				for _, name := range names {
					if _, ok := p.Get(name); !ok {
						t.Errorf("failed to get %q", name)
					}
				}
			},
		},
		{
			name: "empty_server_name",
			test: func(t *testing.T) {
				p := NewPoolWithFactory(mockTransportFactory(successTransport, nil))
				ctx := context.Background()

				_, err := p.Add(ctx, "", ServerConfig{Command: "cmd"})
				if err != nil {
					t.Fatalf("add failed: %v", err)
				}

				_, ok := p.Get("")
				if !ok {
					t.Error("empty name should be valid")
				}

				err = p.Del("")
				if err != nil {
					t.Errorf("del failed: %v", err)
				}
			},
		},
		{
			name: "del_then_add_same_name",
			test: func(t *testing.T) {
				p := NewPoolWithFactory(mockTransportFactory(successTransport, nil))
				ctx := context.Background()

				p.Add(ctx, "server", ServerConfig{Command: "first"})
				p.Del("server")
				p.Add(ctx, "server", ServerConfig{Command: "second"})

				if len(p.All()) != 1 {
					t.Errorf("count = %d, want 1", len(p.All()))
				}

				_, ok := p.Get("server")
				if !ok {
					t.Error("server should exist after re-add")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}
