//go:build linux && arm64

package llamacpp

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib/linux_arm64 -lllama -lggml -lggml-cpu -lggml-base -lc++ -lm
import "C"
