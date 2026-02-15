package llamacpp

/*
#include <stdlib.h>
#include "llama.h"

// Silent callback to suppress logs
void silent_log_callback(enum ggml_log_level level, const char * text, void * user_data) {
    (void)level;
    (void)text;
    (void)user_data;
}

// Helper to set the silent logger
void set_silent_logger() {
    llama_log_set(silent_log_callback, NULL);
}

// Helper to restore default stderr logger
void set_default_logger() {
    llama_log_set(NULL, NULL);
}
*/
import "C"
import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

var (
	onceBackend sync.Once
)

// SetSilentLogger disables all internal llama.cpp logging (stderr).
func SetSilentLogger() {
	C.set_silent_logger()
}

// SetDefaultLogger restores the default llama.cpp logging to stderr.
func SetDefaultLogger() {
	C.set_default_logger()
}

func init() {
	SetSilentLogger()
}

// LlamaEmbedder wraps the C pointers for the model and context.
type LlamaEmbedder struct {
	mu        sync.RWMutex
	model     *C.struct_llama_model
	ctx       *C.struct_llama_context
	isEncoder bool
	nCtx      int
}

// NewLlamaEmbedder initializes the backend (once), loads the model, and creates a context.
func NewLlamaEmbedder(modelPath string) (*LlamaEmbedder, error) {
	onceBackend.Do(func() {
		C.llama_backend_init()
	})

	cPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cPath))

	mParams := C.llama_model_default_params()
	mParams.n_gpu_layers = 0 // CPU only

	model := C.llama_model_load_from_file(cPath, mParams)
	if model == nil {
		return nil, fmt.Errorf("failed to load model from %s", modelPath)
	}

	// Set context size
	nCtx := 512

	cParams := C.llama_context_default_params()
	cParams.embeddings = true
	cParams.n_ctx = C.uint32_t(nCtx)
	cParams.n_batch = C.uint32_t(nCtx)
	cParams.n_ubatch = C.uint32_t(nCtx)

	ctx := C.llama_init_from_model(model, cParams)
	if ctx == nil {
		C.llama_model_free(model)
		return nil, errors.New("failed to create llama context")
	}

	// Determine if this is an Encoder model (BERT, etc.) to avoid warnings.
	isEncoder := false
	if C.llama_model_has_encoder(model) {
		isEncoder = true
	} else {
		key := C.CString("general.architecture")
		buf := make([]byte, 64) // Enough for "bert", "roberta", etc.
		ret := C.llama_model_meta_val_str(model, key, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf)))
		C.free(unsafe.Pointer(key))

		if ret > 0 {
			arch := strings.ToLower(string(buf[:ret]))
			if strings.Contains(arch, "bert") || strings.Contains(arch, "t5") {
				isEncoder = true
			}
		}
	}

	return &LlamaEmbedder{
		model:     model,
		ctx:       ctx,
		isEncoder: isEncoder,
		nCtx:      nCtx,
	}, nil
}

// Free releases the C memory associated with the model and context.
func (l *LlamaEmbedder) Free() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.ctx != nil {
		C.llama_free(l.ctx)
		l.ctx = nil
	}
	if l.model != nil {
		C.llama_model_free(l.model)
		l.model = nil
	}
}

// Embed generates a vector for the given text.
func (l *LlamaEmbedder) Embed(text string) ([]float32, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.model == nil || l.ctx == nil {
		return nil, errors.New("embedder is not initialized or already freed")
	}

	// 1. Sanitize input (remove null bytes which break C strings)
	text = strings.ReplaceAll(text, "\x00", "")
	if text == "" {
		return nil, errors.New("text is empty")
	}

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	vocab := C.llama_model_get_vocab(l.model)
	tokens := make([]C.llama_token, l.nCtx)

	// 2. Tokenize with truncation
	nTokens := C.llama_tokenize(
		vocab,
		cText,
		C.int32_t(len(text)),
		(*C.llama_token)(unsafe.Pointer(&tokens[0])),
		C.int32_t(l.nCtx),
		true, // add_special
		true, // parse_special
	)

	if nTokens < 0 {
		return nil, fmt.Errorf("tokenization failed (code: %d)", nTokens)
	}

	// 3. Safety Clamp
	if int(nTokens) > l.nCtx {
		nTokens = C.int32_t(l.nCtx)
	}

	batch := C.llama_batch_get_one(
		(*C.llama_token)(unsafe.Pointer(&tokens[0])),
		nTokens,
	)

	// Use the correct inference function based on architecture
	if l.isEncoder {
		if res := C.llama_encode(l.ctx, batch); res != 0 {
			return nil, fmt.Errorf("llama_encode failed with code %d", res)
		}
	} else {
		if res := C.llama_decode(l.ctx, batch); res != 0 {
			return nil, fmt.Errorf("llama_decode failed with code %d", res)
		}
	}

	// Retrieve embeddings
	var embPtr *C.float
	if l.isEncoder {
		embPtr = C.llama_get_embeddings_seq(l.ctx, 0)
	} else {
		embPtr = C.llama_get_embeddings_ith(l.ctx, -1)
	}

	if embPtr == nil {
		embPtr = C.llama_get_embeddings(l.ctx)
	}
	if embPtr == nil {
		return nil, errors.New("failed to retrieve embeddings (pointer is nil)")
	}

	nEmbd := int(C.llama_model_n_embd(l.model))
	cSlice := unsafe.Slice((*float32)(unsafe.Pointer(embPtr)), nEmbd)
	result := make([]float32, nEmbd)
	copy(result, cSlice)

	return result, nil
}
