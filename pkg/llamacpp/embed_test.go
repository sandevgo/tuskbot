package llamacpp

import (
	"context"
	"os"
	"testing"
)

func TestLlamaEmbedder(t *testing.T) {
	// Suppress llama.cpp logs for cleaner test output
	//SetSilentLogger()

	// 1. Determine model path
	modelPath := "../../test/models/stsb-bert-tiny.i1-Q6_K.gguf"

	// 2. Check if model exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Fatalf("Skipping TestLlamaEmbedder: model not found at %s", modelPath)
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
