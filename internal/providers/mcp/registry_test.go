package mcp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

type mockStorage struct {
	mu        sync.Mutex
	config    *Config
	loadErr   error
	saveErr   error
	watchErr  error
	loadFunc  func(ctx context.Context) (*Config, error)
	saveFunc  func(ctx context.Context, cfg *Config) error
	watchFunc func(ctx context.Context) (<-chan Config, error)
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		config: &Config{MCPServers: make(map[string]ServerConfig)},
	}
}

func (m *mockStorage) Load(ctx context.Context) (*Config, error) {
	if m.loadFunc != nil {
		return m.loadFunc(ctx)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	// Return copy
	cfg := &Config{MCPServers: make(map[string]ServerConfig)}
	for k, v := range m.config.MCPServers {
		cfg.MCPServers[k] = v
	}
	return cfg, nil
}

func (m *mockStorage) Save(ctx context.Context, cfg *Config) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, cfg)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveErr != nil {
		return m.saveErr
	}
	// Store copy
	m.config = &Config{MCPServers: make(map[string]ServerConfig)}
	for k, v := range cfg.MCPServers {
		m.config.MCPServers[k] = v
	}
	return nil
}

func (m *mockStorage) Watch(ctx context.Context) (<-chan Config, error) {
	if m.watchFunc != nil {
		return m.watchFunc(ctx)
	}
	if m.watchErr != nil {
		return nil, m.watchErr
	}
	return nil, nil
}

func (m *mockStorage) setConfig(cfg *Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = cfg
}

func (m *mockStorage) getConfig() *Config {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.config
}

func TestRegistry_Load(t *testing.T) {
	tests := []struct {
		name        string
		storageCfg  *Config
		storageErr  error
		wantErr     bool
		wantServers int
	}{
		{
			name:        "empty_config",
			storageCfg:  &Config{MCPServers: map[string]ServerConfig{}},
			wantErr:     false,
			wantServers: 0,
		},
		{
			name: "single_server",
			storageCfg: &Config{
				MCPServers: map[string]ServerConfig{
					"server1": {Command: "cmd1", Args: []string{"arg1"}},
				},
			},
			wantErr:     false,
			wantServers: 1,
		},
		{
			name: "multiple_servers",
			storageCfg: &Config{
				MCPServers: map[string]ServerConfig{
					"server1": {Command: "cmd1"},
					"server2": {Command: "cmd2"},
					"server3": {Command: "cmd3"},
				},
			},
			wantErr:     false,
			wantServers: 3,
		},
		{
			name:        "storage_error",
			storageCfg:  nil,
			storageErr:  errors.New("storage failure"),
			wantErr:     true,
			wantServers: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			if tt.storageCfg != nil {
				storage.setConfig(tt.storageCfg)
			}
			storage.loadErr = tt.storageErr

			r := NewRegistry(storage)
			ctx := context.Background()

			err := r.Load(ctx)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			servers := r.List()
			if len(servers) != tt.wantServers {
				t.Errorf("servers count = %d, want %d", len(servers), tt.wantServers)
			}
		})
	}
}

func TestRegistry_Add(t *testing.T) {
	tests := []struct {
		name      string
		initial   map[string]ServerConfig
		addName   string
		addConfig ServerConfig
		saveErr   error
		wantErr   bool
		wantCount int
	}{
		{
			name:      "add_to_empty",
			initial:   map[string]ServerConfig{},
			addName:   "new-server",
			addConfig: ServerConfig{Command: "echo", Args: []string{"hello"}},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "add_to_existing",
			initial: map[string]ServerConfig{
				"existing": {Command: "cmd1"},
			},
			addName:   "new-server",
			addConfig: ServerConfig{Command: "cmd2"},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "overwrite_existing",
			initial: map[string]ServerConfig{
				"server": {Command: "old"},
			},
			addName:   "server",
			addConfig: ServerConfig{Command: "new"},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:      "save_error",
			initial:   map[string]ServerConfig{},
			addName:   "server",
			addConfig: ServerConfig{Command: "cmd"},
			saveErr:   errors.New("save failed"),
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			storage.setConfig(&Config{MCPServers: tt.initial})
			storage.saveErr = tt.saveErr

			r := NewRegistry(storage)
			ctx := context.Background()

			// Load initial state
			if err := r.Load(ctx); err != nil {
				t.Fatalf("Load failed: %v", err)
			}

			err := r.Add(ctx, tt.addName, tt.addConfig)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			servers := r.List()
			if len(servers) != tt.wantCount {
				t.Errorf("servers count = %d, want %d", len(servers), tt.wantCount)
			}

			// Verify storage was updated
			saved := storage.getConfig()
			if len(saved.MCPServers) != tt.wantCount {
				t.Errorf("saved count = %d, want %d", len(saved.MCPServers), tt.wantCount)
			}
		})
	}
}

func TestRegistry_Remove(t *testing.T) {
	tests := []struct {
		name       string
		initial    map[string]ServerConfig
		removeName string
		saveErr    error
		wantErr    bool
		wantCount  int
	}{
		{
			name: "remove_existing",
			initial: map[string]ServerConfig{
				"server1": {Command: "cmd1"},
				"server2": {Command: "cmd2"},
			},
			removeName: "server1",
			wantErr:    false,
			wantCount:  1,
		},
		{
			name: "remove_nonexistent",
			initial: map[string]ServerConfig{
				"server1": {Command: "cmd1"},
			},
			removeName: "nonexistent",
			wantErr:    false,
			wantCount:  1,
		},
		{
			name: "remove_last",
			initial: map[string]ServerConfig{
				"only": {Command: "cmd"},
			},
			removeName: "only",
			wantErr:    false,
			wantCount:  0,
		},
		{
			name:       "remove_from_empty",
			initial:    map[string]ServerConfig{},
			removeName: "any",
			wantErr:    false,
			wantCount:  0,
		},
		{
			name: "save_error",
			initial: map[string]ServerConfig{
				"server": {Command: "cmd"},
			},
			removeName: "server",
			saveErr:    errors.New("save failed"),
			wantErr:    true,
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			storage.setConfig(&Config{MCPServers: tt.initial})

			r := NewRegistry(storage)
			ctx := context.Background()

			if err := r.Load(ctx); err != nil {
				t.Fatalf("Load failed: %v", err)
			}

			storage.saveErr = tt.saveErr

			err := r.Remove(ctx, tt.removeName)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				// On error, state should be unchanged
				if len(r.List()) != len(tt.initial) {
					t.Error("state changed despite error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			servers := r.List()
			if len(servers) != tt.wantCount {
				t.Errorf("servers count = %d, want %d", len(servers), tt.wantCount)
			}
		})
	}
}

func TestRegistry_Get(t *testing.T) {
	tests := []struct {
		name    string
		initial map[string]ServerConfig
		getName string
		wantOk  bool
		wantCmd string
	}{
		{
			name: "get_existing",
			initial: map[string]ServerConfig{
				"server1": {Command: "cmd1"},
			},
			getName: "server1",
			wantOk:  true,
			wantCmd: "cmd1",
		},
		{
			name: "get_nonexistent",
			initial: map[string]ServerConfig{
				"server1": {Command: "cmd1"},
			},
			getName: "nonexistent",
			wantOk:  false,
			wantCmd: "",
		},
		{
			name:    "get_from_empty",
			initial: map[string]ServerConfig{},
			getName: "any",
			wantOk:  false,
			wantCmd: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			storage.setConfig(&Config{MCPServers: tt.initial})

			r := NewRegistry(storage)
			ctx := context.Background()

			if err := r.Load(ctx); err != nil {
				t.Fatalf("Load failed: %v", err)
			}

			cfg, ok := r.Get(tt.getName)

			if ok != tt.wantOk {
				t.Errorf("ok = %v, want %v", ok, tt.wantOk)
			}
			if cfg.Command != tt.wantCmd {
				t.Errorf("command = %s, want %s", cfg.Command, tt.wantCmd)
			}
		})
	}
}

func TestRegistry_List(t *testing.T) {
	tests := []struct {
		name    string
		initial map[string]ServerConfig
		want    int
	}{
		{
			name:    "empty",
			initial: map[string]ServerConfig{},
			want:    0,
		},
		{
			name: "single",
			initial: map[string]ServerConfig{
				"s1": {Command: "c1"},
			},
			want: 1,
		},
		{
			name: "multiple",
			initial: map[string]ServerConfig{
				"s1": {Command: "c1"},
				"s2": {Command: "c2"},
				"s3": {Command: "c3"},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			storage.setConfig(&Config{MCPServers: tt.initial})

			r := NewRegistry(storage)
			ctx := context.Background()

			if err := r.Load(ctx); err != nil {
				t.Fatalf("Load failed: %v", err)
			}

			servers := r.List()

			if len(servers) != tt.want {
				t.Errorf("count = %d, want %d", len(servers), tt.want)
			}

			// Verify it's a copy
			servers["mutated"] = ServerConfig{Command: "hacked"}
			if len(r.List()) != tt.want {
				t.Error("List returned reference, not copy")
			}
		})
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	tests := []struct {
		name       string
		readers    int
		writers    int
		removers   int
		iterations int
	}{
		{
			name:       "light_load",
			readers:    5,
			writers:    2,
			removers:   1,
			iterations: 50,
		},
		{
			name:       "heavy_reads",
			readers:    20,
			writers:    2,
			removers:   1,
			iterations: 100,
		},
		{
			name:       "heavy_writes",
			readers:    5,
			writers:    10,
			removers:   5,
			iterations: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			r := NewRegistry(storage)
			ctx := context.Background()

			if err := r.Load(ctx); err != nil {
				t.Fatalf("Load failed: %v", err)
			}

			var wg sync.WaitGroup

			// Writers
			for i := 0; i < tt.writers; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < tt.iterations; j++ {
						name := fmt.Sprintf("server-%d-%d", id, j)
						_ = r.Add(ctx, name, ServerConfig{Command: "cmd"})
					}
				}(i)
			}

			// Readers
			for i := 0; i < tt.readers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tt.iterations; j++ {
						r.List()
						r.Get("any")
					}
				}()
			}

			// Removers
			for i := 0; i < tt.removers; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < tt.iterations; j++ {
						name := fmt.Sprintf("server-%d-%d", id, j)
						_ = r.Remove(ctx, name)
					}
				}(i)
			}

			wg.Wait()
		})
	}
}

func TestRegistry_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		setup func(r *Registry, storage *mockStorage, ctx context.Context)
		check func(t *testing.T, r *Registry)
	}{
		{
			name: "empty_server_name",
			setup: func(r *Registry, storage *mockStorage, ctx context.Context) {
				r.Load(ctx)
				r.Add(ctx, "", ServerConfig{Command: "cmd"})
			},
			check: func(t *testing.T, r *Registry) {
				cfg, ok := r.Get("")
				if !ok {
					t.Error("empty name should be valid key")
				}
				if cfg.Command != "cmd" {
					t.Errorf("command = %s, want cmd", cfg.Command)
				}
			},
		},
		{
			name: "special_characters_in_name",
			setup: func(r *Registry, storage *mockStorage, ctx context.Context) {
				r.Load(ctx)
				r.Add(ctx, "server/with/slashes", ServerConfig{Command: "c1"})
				r.Add(ctx, "server.with.dots", ServerConfig{Command: "c2"})
				r.Add(ctx, "server:with:colons", ServerConfig{Command: "c3"})
			},
			check: func(t *testing.T, r *Registry) {
				if _, ok := r.Get("server/with/slashes"); !ok {
					t.Error("slash name failed")
				}
				if _, ok := r.Get("server.with.dots"); !ok {
					t.Error("dot name failed")
				}
				if _, ok := r.Get("server:with:colons"); !ok {
					t.Error("colon name failed")
				}
			},
		},
		{
			name: "unicode_server_name",
			setup: func(r *Registry, storage *mockStorage, ctx context.Context) {
				r.Load(ctx)
				r.Add(ctx, "服务器", ServerConfig{Command: "cmd"})
			},
			check: func(t *testing.T, r *Registry) {
				cfg, ok := r.Get("服务器")
				if !ok {
					t.Error("unicode name should work")
				}
				if cfg.Command != "cmd" {
					t.Errorf("command = %s, want cmd", cfg.Command)
				}
			},
		},
		{
			name: "config_with_all_fields",
			setup: func(r *Registry, storage *mockStorage, ctx context.Context) {
				r.Load(ctx)
				r.Add(ctx, "full", ServerConfig{
					Command: "npx",
					Args:    []string{"-y", "@modelcontextprotocol/server"},
					Env:     map[string]string{"KEY": "value", "OTHER": "data"},
					URL:     "http://localhost:8080",
				})
			},
			check: func(t *testing.T, r *Registry) {
				cfg, ok := r.Get("full")
				if !ok {
					t.Fatal("server not found")
				}
				if cfg.Command != "npx" {
					t.Errorf("command = %s, want npx", cfg.Command)
				}
				if len(cfg.Args) != 2 {
					t.Errorf("args count = %d, want 2", len(cfg.Args))
				}
				if cfg.Env["KEY"] != "value" {
					t.Errorf("env KEY = %s, want value", cfg.Env["KEY"])
				}
				if cfg.URL != "http://localhost:8080" {
					t.Errorf("url = %s, want http://localhost:8080", cfg.URL)
				}
			},
		},
		{
			name: "load_without_prior_state",
			setup: func(r *Registry, storage *mockStorage, ctx context.Context) {
				storage.setConfig(&Config{
					MCPServers: map[string]ServerConfig{
						"preloaded": {Command: "cmd"},
					},
				})
				r.Load(ctx)
			},
			check: func(t *testing.T, r *Registry) {
				if _, ok := r.Get("preloaded"); !ok {
					t.Error("preloaded server should exist")
				}
			},
		},
		{
			name: "reload_overwrites_state",
			setup: func(r *Registry, storage *mockStorage, ctx context.Context) {
				storage.setConfig(&Config{
					MCPServers: map[string]ServerConfig{
						"first": {Command: "cmd1"},
					},
				})
				r.Load(ctx)
				r.Add(ctx, "added", ServerConfig{Command: "cmd2"})

				// Simulate external change
				storage.setConfig(&Config{
					MCPServers: map[string]ServerConfig{
						"external": {Command: "cmd3"},
					},
				})
				r.Load(ctx)
			},
			check: func(t *testing.T, r *Registry) {
				servers := r.List()
				if len(servers) != 1 {
					t.Errorf("count = %d, want 1", len(servers))
				}
				if _, ok := r.Get("external"); !ok {
					t.Error("external server should exist after reload")
				}
				if _, ok := r.Get("added"); ok {
					t.Error("added server should be gone after reload")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			r := NewRegistry(storage)
			ctx := context.Background()

			tt.setup(r, storage, ctx)
			tt.check(t, r)
		})
	}
}

func TestRegistry_Watch(t *testing.T) {
	tests := []struct {
		name     string
		watchErr error
		updates  []Config
		wantErr  bool
	}{
		{
			name:     "watch_error",
			watchErr: errors.New("watch failed"),
			wantErr:  true,
		},
		{
			name: "receive_updates",
			updates: []Config{
				{MCPServers: map[string]ServerConfig{"s1": {Command: "c1"}}},
				{MCPServers: map[string]ServerConfig{"s1": {Command: "c1"}, "s2": {Command: "c2"}}},
			},
			wantErr: false,
		},
		{
			name:    "channel_closed",
			updates: []Config{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			storage.watchErr = tt.watchErr

			// Channel to synchronize updates to prevent race conditions in tests
			nextUpdate := make(chan struct{})

			if tt.watchErr == nil {
				storage.watchFunc = func(ctx context.Context) (<-chan Config, error) {
					ch := make(chan Config)
					go func() {
						defer close(ch)
						for i, update := range tt.updates {
							// Wait for signal before sending subsequent updates
							if i > 0 {
								select {
								case <-nextUpdate:
								case <-ctx.Done():
									return
								}
							}

							// Deep copy to simulate storage behavior
							cfg := Config{MCPServers: make(map[string]ServerConfig)}
							for k, v := range update.MCPServers {
								cfg.MCPServers[k] = v
							}

							select {
							case ch <- cfg:
							case <-ctx.Done():
								return
							}
						}
					}()
					return ch, nil
				}
			}

			r := NewRegistry(storage)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ch, err := r.Watch(ctx)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify updates
			for i, wantCfg := range tt.updates {
				if i > 0 {
					// Signal that we are ready for the next update
					select {
					case nextUpdate <- struct{}{}:
					case <-ctx.Done():
						t.Fatal("context cancelled")
					}
				}

				gotCfg, ok := <-ch
				if !ok {
					t.Fatalf("channel closed unexpectedly at index %d", i)
				}

				if len(gotCfg.MCPServers) != len(wantCfg.MCPServers) {
					t.Errorf("update %d: servers count = %d, want %d", i, len(gotCfg.MCPServers), len(wantCfg.MCPServers))
				}

				// Verify internal state was updated
				rServers := r.List()
				if len(rServers) != len(wantCfg.MCPServers) {
					t.Errorf("update %d: internal state count = %d, want %d", i, len(rServers), len(wantCfg.MCPServers))
				}
			}

			// Verify channel closes
			_, ok := <-ch
			if ok {
				t.Error("channel should be closed")
			}
		})
	}
}
