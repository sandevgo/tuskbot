//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/sandevgo/tuskbot/pkg/llamacpp"
	"github.com/sandevgo/tuskbot/test"
)

func TestLlamaEmbedder(t *testing.T) {
	// llamacpp.SetDefaultLogger()

	modelPath := test.GetEmbedModelPath(t)
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Fatalf("Skipping TestLlamaEmbedder: model not found at %s", modelPath)
		return
	}

	embedder, err := llamacpp.NewLlamaEmbedder(modelPath)
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}
	defer embedder.Free()

	text := "Hello TuskBot"
	vec, err := embedder.Embed(context.Background(), text)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(vec) == 0 {
		t.Fatal("Generated vector is empty")
	}

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
