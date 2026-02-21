package mcp

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/client"
)

// mockTransport creates a mock transport function for testing
func mockTransport(shouldFail bool, failErr error) Transport {
	return func(ctx context.Context, cfg ServerConfig) (*ManagedClient, error) {
		if shouldFail {
			return nil, failErr
		}
		return &ManagedClient{
			Client: &client.Client{},
			name:   "mock",
		}, nil
	}
}

func TestPool_Add(t *testing.T) {
	tests := []struct {
		name       string
		transport  Transport
		serverName string
		serverCfg  ServerConfig
		wantErr    bool
	}{
		{
			name:       "successful_add",
			transport:  mockTransport(false, nil),
			serverName: "server1",
			serverCfg:  ServerConfig{Command: "echo"},
			wantErr:    false,
		},
		{
			name:       "transport_failure",
			transport:  mockTransport(true, errors.New("connection failed")),
			serverName: "server1",
			serverCfg:  ServerConfig{Command: "echo"},
			wantErr:    true,
		},
		{
			name:       "empty_server_name",
			transport:  mockTransport(false, nil),
			serverName: "",
			serverCfg:  ServerConfig{Command: "echo"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPool()
			// Override transport for testing
			origTransport := StdioTransport
			StdioTransport = tt.transport
			defer func() { StdioTransport = origTransport }()

			ctx := context.Background()
			cli, err := p.Add(ctx, tt.serverName, tt.serverCfg)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cli == nil {
				t.Fatal("expected client, got nil")
			}
		})
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
				p.clients["server1"] = &ManagedClient{name: "server1"}
			},
			getName: "server1",
			wantOk:  true,
		},
		{
			name: "get_nonexistent",
			setup: func(p *Pool) {
				p.clients["server1"] = &ManagedClient{name: "server1"}
			},
			getName: "server2",
			wantOk:  false,
		},
		{
			name: "get_empty_name",
			setup: func(p *Pool) {
				p.clients[""] = &ManagedClient{name: ""}
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
			wantErr:   true,
			wantCount: 0,
		},
		{
			name: "delete_existing",
			setup: func(p *Pool) {
				p.clients["server1"] = &ManagedClient{
					Client: &client.Client{},
					name:   "server1",
				}
			},
			delName:   "server1",
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "delete_nonexistent",
			setup: func(p *Pool) {
				p.clients["server1"] = &ManagedClient{name: "server1"}
			},
			delName:   "server2",
			wantErr:   true,
			wantCount: 1,
		},
		{
			name: "delete_one_of_many",
			setup: func(p *Pool) {
				p.clients["s1"] = &ManagedClient{Client: &client.Client{}, name: "s1"}
				p.clients["s2"] = &ManagedClient{Client: &client.Client{}, name: "s2"}
				p.clients["s3"] = &ManagedClient{Client: &client.Client{}, name: "s3"}
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
				p.clients["s1"] = &ManagedClient{name: "s1"}
			},
			want: 1,
		},
		{
			name: "multiple_clients",
			setup: func(p *Pool) {
				p.clients["s1"] = &ManagedClient{name: "s1"}
				p.clients["s2"] = &ManagedClient{name: "s2"}
				p.clients["s3"] = &ManagedClient{name: "s3"}
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
	p.clients["server"] = &ManagedClient{name: "server"}

	all := p.All()
	all["hacked"] = &ManagedClient{name: "hacked"}

	if len(p.All()) != 1 {
		t.Error("All() should return a copy, not reference")
	}
}

func TestPool_Close(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(p *Pool)
		wantErr bool
	}{
		{
			name:    "close_empty",
			setup:   func(p *Pool) {},
			wantErr: false,
		},
		{
			name: "close_single",
			setup: func(p *Pool) {
				p.clients["s1"] = &ManagedClient{
					Client: &client.Client{},
					name:   "s1",
				}
			},
			wantErr: false,
		},
		{
			name: "close_multiple",
			setup: func(p *Pool) {
				p.clients["s1"] = &ManagedClient{Client: &client.Client{}, name: "s1"}
				p.clients["s2"] = &ManagedClient{Client: &client.Client{}, name: "s2"}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPool()
			tt.setup(p)

			err := p.Close()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
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
			p := NewPool()
			var wg sync.WaitGroup

			// Writers - directly add to map to avoid transport issues
			for i := 0; i < tt.writers; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < tt.iterations; j++ {
						p.mu.Lock()
						name := "server"
						p.clients[name] = &ManagedClient{
							Client: &client.Client{},
							name:   name,
						}
						p.mu.Unlock()
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
						p.Get("server")
					}
				}()
			}

			// Deleters
			for i := 0; i < tt.deleters; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tt.iterations; j++ {
						_ = p.Del("server")
					}
				}()
			}

			wg.Wait()
		})
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
				p := NewPool()
				p.clients["server"] = &ManagedClient{
					Client: &client.Client{},
					name:   "server",
				}

				// Add again with same name
				p.mu.Lock()
				p.clients["server"] = &ManagedClient{
					Client: &client.Client{},
					name:   "server-new",
				}
				p.mu.Unlock()

				if len(p.All()) != 1 {
					t.Errorf("count = %d, want 1", len(p.All()))
				}
			},
		},
		{
			name: "get_after_close",
			test: func(t *testing.T) {
				p := NewPool()
				p.clients["server"] = &ManagedClient{
					Client: &client.Client{},
					name:   "server",
				}

				_ = p.Close()

				// Should still be able to get (though client is closed)
				_, ok := p.Get("server")
				if !ok {
					t.Error("Get should return client even after Close")
				}
			},
		},
		{
			name: "double_close",
			test: func(t *testing.T) {
				p := NewPool()
				p.clients["server"] = &ManagedClient{
					Client: &client.Client{},
					name:   "server",
				}

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
				p := NewPool()
				p.clients["服务器"] = &ManagedClient{name: "服务器"}

				cli, ok := p.Get("服务器")
				if !ok {
					t.Error("unicode name should work")
				}
				if cli.name != "服务器" {
					t.Errorf("name = %s, want 服务器", cli.name)
				}
			},
		},
		{
			name: "special_characters_in_name",
			test: func(t *testing.T) {
				names := []string{
					"server/with/slashes",
					"server.with.dots",
					"server:with:colons",
					"server with spaces",
				}

				p := NewPool()
				for _, name := range names {
					p.clients[name] = &ManagedClient{name: name}
				}

				for _, name := range names {
					if _, ok := p.Get(name); !ok {
						t.Errorf("failed to get %q", name)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}
