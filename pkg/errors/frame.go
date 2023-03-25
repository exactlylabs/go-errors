package errors

import (
	"runtime"
	"strings"
)

type Frame struct {
	runtime.Frame
}

func (f Frame) Package() string {
	// A prefix of "type." and "go." is a compiler-generated symbol that doesn't belong to any package.
	// See variable reservedimports in cmd/compile/internal/gc/subr.go
	if strings.HasPrefix(f.Function, "go.") || strings.HasPrefix(f.Function, "type.") {
		return ""
	}

	pathend := strings.LastIndex(f.Function, "/")

	if i := strings.Index(f.Function[pathend+1:], "."); i != -1 {
		return f.Function[pathend+1 : pathend+1+i]
	}
	return ""
}

func (f Frame) FuncName() string {
	if i := strings.LastIndex(f.Function, "."); i != -1 {
		return f.Function[i+1:]
	}
	return f.Function
}
