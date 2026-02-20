package llamacpp

import (
	"context"
	"os"
	"testing"
)

// TODO: download https://huggingface.co/LLukas22/all-MiniLM-L6-v2-GGUF/resolve/main/all-minilm-l6-v2_q8_0.gguf
func TestLlamaEmbedder(t *testing.T) {
	// Suppress llama.cpp logs for cleaner test output
	//SetSilentLogger()

	// 1. Determine model path
	modelPath := os.Getenv("TUSKBOT_TEST_MODEL")

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
	vec, err := embedder.Embed(context.Background(), text)
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
