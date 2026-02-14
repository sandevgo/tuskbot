//go:build linux && amd64

package llamacpp

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib/linux_amd64 -lllama -lggml -lggml-cpu -lggml-base -lc++ -lm
import "C"
