package rag

import (
	"context"

	"github.com/sandevgo/tuskbot/pkg/llamacpp"
)

const (
	ModelNameE5BaseQ8 = "multilingual-e5-base-q8.gguf"
	ModelUrlE5BaseQ8  = "https://huggingface.co/dinab/multilingual-e5-base-Q8_0-GGUF/resolve/main/multilingual-e5-base-q8_0.gguf"
)

type E5BaseModel struct {
	emb *llamacpp.LlamaEmbedder
}

func NewE5BaseModel(emb *llamacpp.LlamaEmbedder) *E5BaseModel {
	return &E5BaseModel{
		emb: emb,
	}
}

func (m *E5BaseModel) EncodeQuery(ctx context.Context, text string) ([]float32, error) {
	return m.emb.Embed(ctx, "query: "+text)
}

func (m *E5BaseModel) EncodePassage(ctx context.Context, text string) ([]float32, error) {
	return m.emb.Embed(ctx, "passage: "+text)
}

func (m *E5BaseModel) Dims() int {
	return 768
}

func (m *E5BaseModel) GetModelName() string {
	return ModelNameE5BaseQ8
}

func (m *E5BaseModel) GetURL() string {
	return ModelUrlE5BaseQ8
}

func (m *E5BaseModel) Shutdown() error {
	m.emb.Free()
	return nil
}
