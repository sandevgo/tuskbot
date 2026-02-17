//go:build linux && arm64

package sqlite

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo linux,arm64 LDFLAGS: -L${SRCDIR}/lib/linux_arm64 -lsqlite_vec -lm
import "C"
