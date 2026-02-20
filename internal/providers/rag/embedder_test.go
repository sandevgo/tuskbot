package rag

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// mockDualEncoder is a test double for the DualEncoder interface
type mockDualEncoder struct {
	encodeQueryFunc   func(ctx context.Context, text string) ([]float32, error)
	encodePassageFunc func(ctx context.Context, text string) ([]float32, error)
	shutdownFunc      func() error

	queryCalls   []string
	passageCalls []string
}

func (m *mockDualEncoder) EncodeQuery(ctx context.Context, text string) ([]float32, error) {
	m.queryCalls = append(m.queryCalls, text)
	if m.encodeQueryFunc != nil {
		return m.encodeQueryFunc(ctx, text)
	}
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *mockDualEncoder) EncodePassage(ctx context.Context, text string) ([]float32, error) {
	m.passageCalls = append(m.passageCalls, text)
	if m.encodePassageFunc != nil {
		return m.encodePassageFunc(ctx, text)
	}
	return []float32{0.4, 0.5, 0.6}, nil
}

func (m *mockDualEncoder) Shutdown() error {
	if m.shutdownFunc != nil {
		return m.shutdownFunc()
	}
	return nil
}

func TestEmbedder_EncodeQuery(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		timeout     time.Duration
		mockSetup   func(*mockDualEncoder)
		want        []float32
		wantErr     bool
		errContains string
	}{
		{
			name:    "successfully encodes query",
			text:    "test query",
			timeout: 5 * time.Second,
			mockSetup: func(m *mockDualEncoder) {
				m.encodeQueryFunc = func(ctx context.Context, text string) ([]float32, error) {
					return []float32{0.1, 0.2, 0.3}, nil
				}
			},
			want:    []float32{0.1, 0.2, 0.3},
			wantErr: false,
		},
		{
			name:    "returns error on model failure",
			text:    "failing query",
			timeout: 5 * time.Second,
			mockSetup: func(m *mockDualEncoder) {
				m.encodeQueryFunc = func(ctx context.Context, text string) ([]float32, error) {
					return nil, errors.New("model connection failed")
				}
			},
			want:        nil,
			wantErr:     true,
			errContains: "failed to encode query",
		},
		{
			name:    "handles empty text",
			text:    "",
			timeout: 5 * time.Second,
			mockSetup: func(m *mockDualEncoder) {
				m.encodeQueryFunc = func(ctx context.Context, text string) ([]float32, error) {
					return []float32{0.0}, nil
				}
			},
			want:    []float32{0.0},
			wantErr: false,
		},
		{
			name:    "respects timeout",
			text:    "slow query",
			timeout: 50 * time.Millisecond,
			mockSetup: func(m *mockDualEncoder) {
				m.encodeQueryFunc = func(ctx context.Context, text string) ([]float32, error) {
					select {
					case <-time.After(100 * time.Millisecond):
						return []float32{0.1}, nil
					case <-ctx.Done():
						return nil, ctx.Err()
					}
				}
			},
			want:        nil,
			wantErr:     true,
			errContains: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDualEncoder{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			embedder := &Embedder{
				model:     mock,
				timeout:   tt.timeout,
				chunkConf: E5BaseChunkerConfig(),
			}

			ctx := context.Background()
			got, err := embedder.EncodeQuery(ctx, tt.text)

			if tt.wantErr {
				if err == nil {
					t.Errorf("EncodeQuery() error = nil, wantErr true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("EncodeQuery() error = %v, should contain %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("EncodeQuery() unexpected error = %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("EncodeQuery() got %v, want %v", got, tt.want)
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("EncodeQuery() got %v, want %v", got, tt.want)
					return
				}
			}

			if len(mock.queryCalls) != 1 || mock.queryCalls[0] != tt.text {
				t.Errorf("EncodeQuery() did not call model with correct text, got calls: %v", mock.queryCalls)
			}
		})
	}
}

func TestEmbedder_EncodePassage(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		timeout       time.Duration
		chunkConf     ChunkerConfig
		mockSetup     func(*mockDualEncoder)
		wantChunks    int
		wantErr       bool
		errContains   string
		errIndexCheck bool
	}{
		{
			name:      "empty text returns empty embeddings",
			text:      "",
			timeout:   5 * time.Second,
			chunkConf: E5BaseChunkerConfig(),
			mockSetup: func(m *mockDualEncoder) {
				m.encodePassageFunc = func(ctx context.Context, text string) ([]float32, error) {
					return nil, errors.New("should not be called")
				}
			},
			wantChunks: 0,
			wantErr:    false,
		},
		{
			name:      "short text produces single chunk",
			text:      "Short text.",
			timeout:   5 * time.Second,
			chunkConf: E5BaseChunkerConfig(),
			mockSetup: func(m *mockDualEncoder) {
				m.encodePassageFunc = func(ctx context.Context, text string) ([]float32, error) {
					return []float32{0.1, 0.2}, nil
				}
			},
			wantChunks: 1,
			wantErr:    false,
		},
		{
			name:    "long text produces multiple chunks",
			text:    "First chunk of text. Second chunk of text here.",
			timeout: 5 * time.Second,
			chunkConf: ChunkerConfig{
				MaxTokens:     10,
				OverlapTokens: 2,
			},
			mockSetup: func(m *mockDualEncoder) {
				callCount := 0
				m.encodePassageFunc = func(ctx context.Context, text string) ([]float32, error) {
					callCount++
					return []float32{float32(callCount)}, nil
				}
			},
			wantChunks: 2,
			wantErr:    false,
		},
		{
			name:    "returns error on chunk failure",
			text:    "First chunk of text. Second chunk of text here.",
			timeout: 5 * time.Second,
			chunkConf: ChunkerConfig{
				MaxTokens:     10,
				OverlapTokens: 2,
			},
			mockSetup: func(m *mockDualEncoder) {
				callCount := 0
				m.encodePassageFunc = func(ctx context.Context, text string) ([]float32, error) {
					callCount++
					if callCount == 2 {
						return nil, errors.New("embedding failed")
					}
					return []float32{0.1}, nil
				}
			},
			wantChunks:    0,
			wantErr:       true,
			errContains:   "failed to embed chunk",
			errIndexCheck: true,
		},
		{
			name:    "timeout during chunk processing",
			text:    "First chunk of text. Second chunk of text here.",
			timeout: 50 * time.Millisecond,
			chunkConf: ChunkerConfig{
				MaxTokens:     10,
				OverlapTokens: 2,
			},
			mockSetup: func(m *mockDualEncoder) {
				callCount := 0
				m.encodePassageFunc = func(ctx context.Context, text string) ([]float32, error) {
					callCount++
					if callCount == 2 {
						select {
						case <-time.After(100 * time.Millisecond):
							return []float32{0.2}, nil
						case <-ctx.Done():
							return nil, ctx.Err()
						}
					}
					return []float32{0.1}, nil
				}
			},
			wantChunks:  0,
			wantErr:     true,
			errContains: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDualEncoder{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			embedder := &Embedder{
				model:     mock,
				timeout:   tt.timeout,
				chunkConf: tt.chunkConf,
			}

			ctx := context.Background()
			got, err := embedder.EncodePassage(ctx, tt.text)

			if tt.wantErr {
				if err == nil {
					t.Errorf("EncodePassage() error = nil, wantErr true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("EncodePassage() error = %v, should contain %v", err, tt.errContains)
				}
				if tt.errIndexCheck && !strings.Contains(err.Error(), "chunk 1") {
					t.Errorf("EncodePassage() error should contain chunk index, got: %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("EncodePassage() unexpected error = %v", err)
				return
			}

			if len(got) != tt.wantChunks {
				t.Errorf("EncodePassage() got %d chunks, want %d", len(got), tt.wantChunks)
			}

			if len(mock.passageCalls) != tt.wantChunks {
				t.Errorf("EncodePassage() called model %d times, want %d", len(mock.passageCalls), tt.wantChunks)
			}
		})
	}
}

func TestEmbedder_EncodeQuery_CallsModelWithCorrectContext(t *testing.T) {
	mock := &mockDualEncoder{
		encodeQueryFunc: func(ctx context.Context, text string) ([]float32, error) {
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Error("Expected context to have deadline")
			}

			expectedDeadline := time.Now().Add(100 * time.Millisecond)
			if deadline.After(expectedDeadline.Add(10*time.Millisecond)) || deadline.Before(expectedDeadline.Add(-10*time.Millisecond)) {
				t.Errorf("Deadline %v not within expected range of %v", deadline, expectedDeadline)
			}

			return []float32{0.1}, nil
		},
	}

	embedder := &Embedder{
		model:     mock,
		timeout:   100 * time.Millisecond,
		chunkConf: E5BaseChunkerConfig(),
	}

	ctx := context.Background()
	_, err := embedder.EncodeQuery(ctx, "test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestEmbedder_Shutdown(t *testing.T) {
	t.Run("shutdown propagates to model", func(t *testing.T) {
		shutdownCalled := false
		mock := &mockDualEncoder{
			shutdownFunc: func() error {
				shutdownCalled = true
				return nil
			},
		}

		embedder := NewEmbedder(mock)
		err := embedder.model.Shutdown()

		if err != nil {
			t.Errorf("Shutdown() error = %v", err)
		}
		if !shutdownCalled {
			t.Error("Shutdown() did not call model's Shutdown")
		}
	})
}
