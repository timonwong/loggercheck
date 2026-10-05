package runtimewitness

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"k8s.io/klog/v2"
)

type value struct{}

func (value) String() string { return "value" }

func generic[U any, P interface {
	*U
	String() string
}](p P) any {
	return p
}

// These witnesses establish reachability independently of SSA and diagnostic
// expectations. The log value is inspected without formatting its Stringer.
func TestReachableNil(t *testing.T) {
	old := klog.OsExit
	defer func() { klog.OsExit = old }()
	klog.LogToStderr(false)
	klog.SetOutput(&bytes.Buffer{})

	evidence := make(map[string]bool)
	for _, tc := range []struct {
		name string
		exit func()
	}{
		{"FlushAndExit", func() { klog.FlushAndExit(time.Nanosecond, 1) }},
		{"Fatal", func() { klog.Fatal("intercepted exit") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			klog.OsExit = func(int) { calls++ }
			var p *value
			if p == nil {
				tc.exit()
			}
			if calls == 0 || p != nil {
				t.Fatal("did not reach the logging point with nil after the exit hook")
			}
			evidence[tc.name] = true
		})
	}
	t.Run("GenericTypedNil", func(t *testing.T) {
		v := generic[value]((*value)(nil))
		p, ok := v.(*value)
		if !ok || p != nil || v == nil {
			t.Fatal("generic pointer did not retain a boxed typed-nil value")
		}
		evidence["GenericTypedNil"] = true
		panicked := false
		func() {
			defer func() { panicked = recover() != nil }()
			v.(interface{ String() string }).String()
		}()
		if !panicked {
			t.Fatal("boxed nil value-receiver Stringer did not panic")
		}
		evidence["GenericStringPanics"] = true
	})
	if path := os.Getenv("NILNESS_WITNESS_ARTIFACT"); path != "" {
		data, err := json.MarshalIndent(evidence, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
