package testutils

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func RunTestWithName(t *testing.T, name string, fn func(t *testing.T)) {
	t.Helper()

	// Get caller info
	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	var caller string
	if ok && details != nil {
		caller = details.Name()
		parts := strings.Split(caller, ".")
		if len(parts) > 0 {
			caller = parts[len(parts)-1]
		}
	}

	t.Run(name, func(t *testing.T) {
		t.Log(fmt.Sprintf("\n=== RUN %s/%s", caller, name))
		fn(t)
		if !t.Failed() {
			t.Log(fmt.Sprintf("--- PASS: %s/%s", caller, name))
		}
	})
}
