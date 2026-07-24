package runner

import (
	"testing"
)

func TestNewRunner(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal("expected non-nil runner instance")
	}
	if r.speedtester == nil {
		t.Error("expected non-nil speedtester field")
	}
	if r.scanner == nil {
		t.Error("expected non-nil scanner field")
	}
}
