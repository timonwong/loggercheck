package nilness

import (
	"testing"

	"k8s.io/klog/v2"
)

func testingFatalOnNil(t *testing.T, p *Time) {
	if p == nil {
		t.Fatal("missing time")
	}
	klog.InfoS("after t.Fatal guard", "time", p)
}

func testingFatalfOnNil(t *testing.T, p *Time) {
	if p == nil {
		t.Fatalf("missing %s", "time")
	}
	klog.InfoS("after t.Fatalf guard", "time", p)
}

func testingFailNowOnNil(t *testing.T, p *Time) {
	if p == nil {
		t.FailNow()
	}
	klog.InfoS("after t.FailNow guard", "time", p)
}

func testingSkipOnNil(t *testing.T, p *Time) {
	if p == nil {
		t.Skip("missing time")
	}
	klog.InfoS("after t.Skip guard", "time", p)
}

func testingSkipfOnNil(t *testing.T, p *Time) {
	if p == nil {
		t.Skipf("missing %s", "time")
	}
	klog.InfoS("after t.Skipf guard", "time", p)
}

func testingSkipNowOnNil(t *testing.T, p *Time) {
	if p == nil {
		t.SkipNow()
	}
	klog.InfoS("after t.SkipNow guard", "time", p)
}

func benchmarkFatalOnNil(b *testing.B, p *Time) {
	if p == nil {
		b.Fatal("missing time")
	}
	klog.InfoS("after b.Fatal guard", "time", p)
}

func testingErrorOnNil(t *testing.T, p *Time) {
	if p == nil {
		t.Error("missing time")
	}
	klog.InfoS("after t.Error", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
}

// Calls through an interface have no static callee, so a testing.TB guard
// does not end the nil branch (remaining false positive).
func testingTBFatalOnNil(tb testing.TB, p *Time) {
	if p == nil {
		tb.Fatal("missing time")
	}
	klog.InfoS("after tb.Fatal guard", "time", p) // want `logging value may panic when nil because its element type implements fmt.Stringer`
}
