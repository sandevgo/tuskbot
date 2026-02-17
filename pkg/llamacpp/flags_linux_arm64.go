package llamacpp

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo linux,arm64 LDFLAGS: -L${SRCDIR}/lib/linux_arm64 -lllama -lggml -lstdc++ -lm -fopenmp
import "C"
