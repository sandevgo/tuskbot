package llamacpp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLlamaEmbedder(t *testing.T) {
	// Suppress llama.cpp logs for cleaner test output
	//SetSilentLogger()

	// 1. Determine model path
	modelPath := os.Getenv("TUSKBOT_TEST_MODEL")
	if modelPath == "" {
		// Check default runtime location relative to this package
		candidates := []string{
			"/Users/sandevgo/.tuskbot/models/e5-base-v2-q8_0.gguf",
		}
		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				abs, _ := filepath.Abs(p)
				modelPath = abs
				break
			}
		}
	}

	// 2. Check if model exists
	if modelPath == "" {
		t.Skip("Skipping TestLlamaEmbedder: no model found. Set TUSKBOT_TEST_MODEL env var.")
		return
	}
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Skipf("Skipping TestLlamaEmbedder: model not found at %s", modelPath)
		return
	}

	// 3. Init
	embedder, err := NewLlamaEmbedder(modelPath)
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}
	defer embedder.Free()

	// 4. Embed
	text := "Hello TuskBot"
	vec, err := embedder.Embed(text)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	// 5. Assertions
	if len(vec) == 0 {
		t.Fatal("Generated vector is empty")
	}

	// Check dimensions (usually 384, 768, or 1024 depending on model)
	t.Logf("Vector dimensions: %d", len(vec))
	t.Logf("First 5 values: %v", vec[:5])

	// Sanity check: ensure not all zeros
	allZeros := true
	for _, v := range vec {
		if v != 0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		t.Fatal("Vector contains all zeros")
	}
}
