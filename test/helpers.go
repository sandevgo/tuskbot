package test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const (
	EmbedModelPath = "./models/stsb-bert-tiny-i1.gguf"
)

func GetEmbedModelPath(t *testing.T) string {
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)

	path := filepath.Join(testDir, EmbedModelPath)
	if _, err := os.Stat(path); err != nil {
		t.Skipf("Model not found at %s: %v", path, err)
	}
	return path
}
