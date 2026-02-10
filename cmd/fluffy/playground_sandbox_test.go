package main

import (
	"testing"
)

func TestValidateImports_Safe(t *testing.T) {
	code := `package main

import (
	"fmt"
	"context"
	"strings"
	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/widgets"
	"github.com/odvcencio/fluffyui/state"
)

func main() { fmt.Println("hi"); _ = context.Background(); _ = strings.ToUpper; _ = fluffy.Run; _ = widgets.NewButton; _ = state.New[int] }
`
	if err := validateCode([]byte(code)); err != nil {
		t.Fatal("safe code rejected:", err)
	}
}

func TestValidateImports_Dangerous(t *testing.T) {
	dangerous := []struct {
		name string
		code string
	}{
		{"os/exec", `package main; import "os/exec"; func main() { _ = exec.Command }`},
		{"syscall", `package main; import "syscall"; func main() { _ = syscall.Getpid }`},
		{"unsafe", `package main; import "unsafe"; func main() { _ = unsafe.Sizeof }`},
		{"plugin", `package main; import "plugin"; func main() { _ = plugin.Open }`},
		{"net", `package main; import "net"; func main() { _ = net.Dial }`},
		{"net/http", `package main; import "net/http"; func main() { _ = http.Get }`},
		{"net/rpc", `package main; import "net/rpc"; func main() { _ = rpc.Dial }`},
		{"C", `package main; import "C"; func main() {}`},
		{"runtime/cgo", `package main; import "runtime/cgo"; func main() { _ = cgo.Handle(0) }`},
		{"os/signal", `package main; import "os/signal"; func main() { _ = signal.Notify }`},
		{"runtime/debug", `package main; import "runtime/debug"; func main() { _ = debug.FreeOSMemory }`},
	}
	for _, tc := range dangerous {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateCode([]byte(tc.code)); err == nil {
				t.Errorf("dangerous import %q not caught", tc.name)
			}
		})
	}
}

func TestValidateCode_NoMain(t *testing.T) {
	code := `package foo

func Bar() {}
`
	if err := validateCode([]byte(code)); err == nil {
		t.Error("non-main package not rejected")
	}
}

func TestValidateCode_ParseError(t *testing.T) {
	code := `this is not valid go code`
	if err := validateCode([]byte(code)); err == nil {
		t.Error("parse error not caught")
	}
}

func TestValidateCode_StdlibAllowed(t *testing.T) {
	// Standard library packages that should be allowed
	allowed := []string{
		`package main; import "fmt"; func main() {}`,
		`package main; import "os"; func main() {}`,
		`package main; import "strings"; func main() {}`,
		`package main; import "strconv"; func main() {}`,
		`package main; import "math"; func main() {}`,
		`package main; import "time"; func main() {}`,
		`package main; import "context"; func main() {}`,
	}
	for _, code := range allowed {
		if err := validateCode([]byte(code)); err != nil {
			t.Errorf("allowed import rejected: %s: %v", code, err)
		}
	}
}
