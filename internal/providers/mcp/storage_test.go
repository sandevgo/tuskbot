package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestFileStorage_Load_MissingDirectory(t *testing.T) {
	t.Parallel()
	fs := NewFileStorage("/nonexistent/path/config.json")
	ctx := context.Background()

	_, err := fs.Load(ctx)
	if err == nil {
		t.Fatal("expected error when directory does not exist")
	}
}

func TestFileStorage_Load_NullMCPServers(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mcp_config.json")

	// JSON with null instead of empty object
	content := `{"mcpServers": null}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	fs := NewFileStorage(path)
	ctx := context.Background()

	cfg, err := fs.Load(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MCPServers == nil {
		t.Fatal("expected MCPServers to be initialized, got nil")
	}
	if len(cfg.MCPServers) != 0 {
		t.Errorf("expected empty map, got %d entries", len(cfg.MCPServers))
	}
}

func TestFileStorage_Save_FilePermissions(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mcp_config.json")
	fs := NewFileStorage(path)
	ctx := context.Background()

	cfg := &Config{MCPServers: map[string]ServerConfig{"test": {Command: "echo"}}}
	if err := fs.Save(ctx, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	mode := info.Mode().Perm()
	expected := os.FileMode(0644)
	if mode != expected {
		t.Errorf("file permissions = %o, want %o", mode, expected)
	}
}

func TestFileStorage_Save_ReadOnlyDirectory(t *testing.T) {
	t.Parallel()
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mcp_config.json")
	fs := NewFileStorage(path)
	ctx := context.Background()

	// Make directory read-only
	if err := os.Chmod(tmpDir, 0555); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}
	// Cleanup is handled by t.TempDir cleanup, but we might need to restore perms to delete
	t.Cleanup(func() { os.Chmod(tmpDir, 0755) })

	cfg := &Config{MCPServers: map[string]ServerConfig{"test": {Command: "echo"}}}
	err := fs.Save(ctx, cfg)
	if err == nil {
		t.Fatal("expected error when saving to read-only directory")
	}
}

func TestFileStorage_Save_SpecialCharacters(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mcp_config.json")
	fs := NewFileStorage(path)
	ctx := context.Background()

	cfg := &Config{
		MCPServers: map[string]ServerConfig{
			"special": {
				Command: "echo",
				Args:    []string{"hello\nworld", "tab\there", `quotes"and'quotes`},
				Env: map[string]string{
					"NEWLINE": "value\nwith\nnewlines",
					"QUOTES":  `{"key": "value"}`,
					"UNICODE": "Hello, 世界 🌍",
				},
			},
		},
	}

	if err := fs.Save(ctx, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify by loading back
	loaded, err := fs.Load(ctx)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	srv := loaded.MCPServers["special"]
	if srv.Env["NEWLINE"] != "value\nwith\nnewlines" {
		t.Errorf("newline preservation failed: got %q", srv.Env["NEWLINE"])
	}
	if srv.Env["QUOTES"] != `{"key": "value"}` {
		t.Errorf("quotes preservation failed: got %q", srv.Env["QUOTES"])
	}
	if len(srv.Args) != 3 {
		t.Errorf("args count = %d, want 3", len(srv.Args))
	}
}

func TestFileStorage_Watch_AtomicWrite(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mcp_config.json")

	// Initial file
	if err := os.WriteFile(path, []byte(`{"mcpServers": {}}`), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	fs := NewFileStorage(path)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updates, err := fs.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Simulate atomic write (write to temp, rename)
	tmpPath := path + ".tmp"
	newContent := `{"mcpServers": {"atomic": {"command": "test"}}}`

	// Write to temp file
	if err := os.WriteFile(tmpPath, []byte(newContent), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Rename (atomic operation) - during this instant, file might not exist on some systems
	if err := os.Rename(tmpPath, path); err != nil {
		t.Fatalf("failed to rename: %v", err)
	}

	select {
	case cfg := <-updates:
		if _, ok := cfg.MCPServers["atomic"]; !ok {
			t.Error("did not detect atomic write update")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for atomic write detection")
	}
}

func TestFileStorage_Watch_FileDeleted(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mcp_config.json")

	// Initial file
	if err := os.WriteFile(path, []byte(`{"mcpServers": {}}`), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	fs := NewFileStorage(path)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updates, err := fs.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Delete file
	if err := os.Remove(path); err != nil {
		t.Fatalf("failed to remove file: %v", err)
	}

	// Recreate file with new content
	newContent := `{"mcpServers": {"recovered": {"command": "test"}}}`
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		t.Fatalf("failed to recreate file: %v", err)
	}

	// Loop until we receive the expected config
	timeout := time.After(3 * time.Second)
	for {
		select {
		case cfg, ok := <-updates:
			if !ok {
				t.Fatal("channel closed unexpectedly")
			}
			if _, ok := cfg.MCPServers["recovered"]; ok {
				return // Success
			}
			// Not the config we want, keep waiting
		case <-timeout:
			t.Fatal("timeout waiting for update after recreation")
		}
	}
}

func TestFileStorage_Watch_RapidUpdates(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mcp_config.json")

	// Initial file
	if err := os.WriteFile(path, []byte(`{"mcpServers": {}}`), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	fs := NewFileStorage(path)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updates, err := fs.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	// Send 5 rapid updates
	for i := 0; i < 5; i++ {
		content := `{"mcpServers": {"server` + string(rune('0'+i)) + `": {"command": "cmd` + string(rune('0'+i)) + `"}}}`
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write update %d: %v", i, err)
		}
		time.Sleep(100 * time.Millisecond) // Faster than poll interval
	}

	// Should eventually see the last update (server4)
	var lastCfg Config
	timeout := time.After(3 * time.Second)
	received := 0

	for received < 5 {
		select {
		case cfg := <-updates:
			lastCfg = cfg
			received++
		case <-timeout:
			goto done
		}
	}
done:

	// Verify we got at least the final state
	if _, ok := lastCfg.MCPServers["server4"]; !ok && received > 0 {
		t.Errorf("expected final update to contain server4, got %v", lastCfg.MCPServers)
	}

	// We might not get all 5 updates due to polling nature, but we should get at least one
	if received == 0 {
		t.Fatal("expected at least one update from rapid writes")
	}
}

func TestFileStorage_ConcurrentWatchAndSave(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mcp_config.json")

	// Initial file
	if err := os.WriteFile(path, []byte(`{"mcpServers": {}}`), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	fs := NewFileStorage(path)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	updates, err := fs.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	var updateCount int32

	// Consumer
	done := make(chan struct{})
	go func() {
		for range updates {
			atomic.AddInt32(&updateCount, 1)
		}
		close(done)
	}()

	// Producer - save while watching
	for i := 0; i < 10; i++ {
		cfg := &Config{
			MCPServers: map[string]ServerConfig{
				string(rune('a' + i)): {Command: "test"},
			},
		}
		if err := fs.Save(ctx, cfg); err != nil {
			t.Errorf("Save failed: %v", err)
		}
		// Reduced sleep to just cover the 1s tick interval (10 * 120ms = 1.2s)
		time.Sleep(120 * time.Millisecond)
	}

	cancel()
	<-done

	// Should have received some updates (exact count depends on timing)
	if atomic.LoadInt32(&updateCount) == 0 {
		t.Error("expected at least one update during concurrent watch and save")
	}
}

func TestFileStorage_Load_InvalidJSONTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "array_instead_of_object",
			content: `{"mcpServers": []}`,
			wantErr: true,
		},
		{
			name:    "string_instead_of_object",
			content: `{"mcpServers": "invalid"}`,
			wantErr: true,
		},
		{
			name:    "number_instead_of_string",
			content: `{"mcpServers": {"test": {"command": 123}}}`,
			wantErr: true,
		},
		{
			name:    "trailing_comma",
			content: `{"mcpServers": {"test": {"command": "echo",}}}`,
			wantErr: true,
		},
		{
			name:    "incomplete_json",
			content: `{"mcpServers": {"test": {`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "mcp_config.json")

			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			fs := NewFileStorage(path)
			ctx := context.Background()

			_, err := fs.Load(ctx)
			if tt.wantErr && err == nil {
				t.Error("expected error for invalid JSON, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestFileStorage_Save_EmptyConfig(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mcp_config.json")
	fs := NewFileStorage(path)
	ctx := context.Background()

	// Save nil map
	cfg := &Config{MCPServers: nil}
	if err := fs.Save(ctx, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load and verify
	loaded, err := fs.Load(ctx)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.MCPServers == nil {
		t.Fatal("MCPServers should be initialized after Load")
	}

	// Verify JSON structure
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `"mcpServers"`) {
		t.Error("JSON missing mcpServers key")
	}
}

// MockStorage implements Storage interface for testing error propagation
type MockStorage struct {
	loadFunc  func(ctx context.Context) (*Config, error)
	saveFunc  func(ctx context.Context, cfg *Config) error
	watchFunc func(ctx context.Context) (<-chan Config, error)
}

func (m *MockStorage) Load(ctx context.Context) (*Config, error) {
	if m.loadFunc != nil {
		return m.loadFunc(ctx)
	}
	return &Config{MCPServers: map[string]ServerConfig{}}, nil
}

func (m *MockStorage) Save(ctx context.Context, cfg *Config) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, cfg)
	}
	return nil
}

func (m *MockStorage) Watch(ctx context.Context) (<-chan Config, error) {
	if m.watchFunc != nil {
		return m.watchFunc(ctx)
	}
	return nil, nil
}

func TestFileStorage_ContextCancellation(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mcp_config.json")
	if err := os.WriteFile(path, []byte(`{}`), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	fs := NewFileStorage(path)

	// Test Load with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Load should still work (it doesn't check context during file read)
	_, err := fs.Load(ctx)
	if err != nil {
		t.Logf("Load with cancelled context returned: %v", err)
	}

	// Test Save with cancelled context
	ctx2, cancel2 := context.WithCancel(context.Background())
	cancel2()

	// Save should still work (file operations don't respect context)
	err = fs.Save(ctx2, &Config{MCPServers: map[string]ServerConfig{}})
	if err != nil {
		t.Logf("Save with cancelled context returned: %v", err)
	}
}
