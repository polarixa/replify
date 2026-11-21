package replify_test

import (
	"strings"
	"testing"

	"github.com/polarixa/replify"
)

func TestTotal_OmittedWhenUnset(t *testing.T) {
	w := replify.New().OK().WithMessage("ok")

	if w.IsTotalPresent() {
		t.Fatal("IsTotalPresent() = true, want false for a new wrapper")
	}
	if got := w.Total(); got != 0 {
		t.Fatalf("Total() = %d, want 0", got)
	}
	if strings.Contains(w.JSON(), `"total"`) {
		t.Fatalf("JSON() = %s, want no total field", w.JSON())
	}
}

func TestTotal_ExplicitZeroIsSerialized(t *testing.T) {
	w := replify.New().OK().WithTotal(0)

	if !w.IsTotalPresent() {
		t.Fatal("IsTotalPresent() = false, want true after WithTotal(0)")
	}
	if !strings.Contains(w.JSON(), `"total":0`) {
		t.Fatalf("JSON() = %s, want total:0", w.JSON())
	}
}

func TestTotal_NegativeMeansUnset(t *testing.T) {
	w := replify.New().OK().WithTotal(5).WithTotal(-1)

	if w.IsTotalPresent() || strings.Contains(w.JSON(), `"total"`) {
		t.Fatalf("negative total should be treated as unset, got %s", w.JSON())
	}
}

func TestTotal_RoundTrip(t *testing.T) {
	unset, err := replify.UnwrapJSON(replify.New().OK().JSON())
	if err != nil {
		t.Fatalf("UnwrapJSON() error = %v", err)
	}
	if unset.IsTotalPresent() {
		t.Fatal("unset total should stay unset after a round trip")
	}

	set, err := replify.UnwrapJSON(replify.New().OK().WithTotal(7).JSON())
	if err != nil {
		t.Fatalf("UnwrapJSON() error = %v", err)
	}
	if set.Total() != 7 || !set.IsTotalPresent() {
		t.Fatalf("Total() = %d present=%v, want 7 true", set.Total(), set.IsTotalPresent())
	}
}
