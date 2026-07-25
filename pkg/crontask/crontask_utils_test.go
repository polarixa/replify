package crontask_test

import (
	"errors"
	"testing"
	"time"

	"github.com/polarixa/replify/pkg/crontask"
)

// TestIsValidCronExpr verifies the package-level validity predicate.
func TestIsValidCronExpr(t *testing.T) {
	t.Parallel()
	cases := []struct {
		expr  string
		valid bool
	}{
		{"* * * * *", true},
		{"0 9 * * 1-5", true},
		{"@hourly", true},
		{"@every 5m", true},
		{"TZ=UTC 0 * * * *", true},
		{"", false},
		{"bad expression", false},
		{"60 * * * *", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.expr, func(t *testing.T) {
			t.Parallel()
			got := crontask.IsValidCronExpr(tc.expr)
			if got != tc.valid {
				t.Errorf("IsValidCronExpr(%q) = %v, want %v", tc.expr, got, tc.valid)
			}
		})
	}
}

// TestValidateCronExpr verifies the package-level validation function.
func TestValidateCronExpr(t *testing.T) {
	t.Parallel()
	if err := crontask.ValidateCronExpr("0 * * * *"); err != nil {
		t.Errorf("unexpected error for valid expression: %v", err)
	}
	err := crontask.ValidateCronExpr("invalid")
	if err == nil {
		t.Error("expected error for invalid expression, got nil")
	}
	if !errors.Is(err, crontask.ErrInvalidExpression) {
		t.Errorf("errors.Is(err, ErrInvalidExpression) = false, want true")
	}
}

// TestIsDue verifies the package-level IsDue predicate.
func TestIsDue(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		expr string
		at   time.Time
		want bool
	}{
		{
			// "* * * * *" is due at the start of every minute.
			name: "minutely at boundary",
			expr: "* * * * *",
			at:   time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC),
			want: true,
		},
		{
			// "* * * * *" is NOT due mid-minute.
			name: "minutely mid-minute",
			expr: "* * * * *",
			at:   time.Date(2024, 1, 1, 12, 5, 30, 0, time.UTC),
			want: false,
		},
		{
			// "0 * * * *" is due at the top of each hour.
			name: "hourly at boundary",
			expr: "0 * * * *",
			at:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			// "0 * * * *" is not due at :30.
			name: "hourly not at :30",
			expr: "0 * * * *",
			at:   time.Date(2024, 1, 1, 13, 30, 0, 0, time.UTC),
			want: false,
		},
		{
			// "@every 5m" is due at 5-minute boundaries (epoch-aligned).
			name: "every 5m at boundary",
			expr: "@every 5m",
			at:   time.Unix(0, 0).UTC().Add(5 * time.Minute),
			want: true,
		},
		{
			// "@every 5m" is not due 1 minute after a boundary.
			name: "every 5m off boundary",
			expr: "@every 5m",
			at:   time.Unix(0, 0).UTC().Add(6 * time.Minute),
			want: false,
		},
		{
			// Invalid expression returns false without panic.
			name: "invalid expression",
			expr: "bad expr",
			at:   time.Now(),
			want: false,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := crontask.IsDue(tc.expr, tc.at)
			if got != tc.want {
				t.Errorf("IsDue(%q, %v) = %v, want %v", tc.expr, tc.at, got, tc.want)
			}
		})
	}
}

// TestNextRun verifies the package-level NextRun helper.
func TestNextRun(t *testing.T) {
	t.Parallel()
	ref := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	got, err := crontask.NextRun("0 * * * *", ref)
	if err != nil {
		t.Fatalf("NextRun: unexpected error: %v", err)
	}
	want := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("NextRun = %v, want %v", got, want)
	}

	_, err = crontask.NextRun("bad", ref)
	if err == nil {
		t.Error("expected error for invalid expression, got nil")
	}
}

// TestPackageLevelNextRuns verifies the package-level NextRuns helper.
func TestPackageLevelNextRuns(t *testing.T) {
	t.Parallel()
	ref := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	runs, err := crontask.NextRuns("0 * * * *", ref, 3)
	if err != nil {
		t.Fatalf("NextRuns: unexpected error: %v", err)
	}
	if len(runs) != 3 {
		t.Fatalf("expected 3 runs, got %d", len(runs))
	}
	want := []time.Time{
		time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 3, 0, 0, 0, time.UTC),
	}
	for i, r := range runs {
		if !r.Equal(want[i]) {
			t.Errorf("runs[%d] = %v, want %v", i, r, want[i])
		}
	}

	// n <= 0 returns nil without error.
	empty, err := crontask.NextRuns("0 * * * *", ref, 0)
	if err != nil || empty != nil {
		t.Errorf("NextRuns with n=0: got (%v, %v), want (nil, nil)", empty, err)
	}

	// Invalid expression returns an error.
	_, err = crontask.NextRuns("bad", ref, 3)
	if err == nil {
		t.Error("expected error for invalid expression, got nil")
	}
}

// TestMustParse verifies that MustParse returns a valid Expression for a
// correct input and panics for an invalid one.
func TestMustParse(t *testing.T) {
	t.Parallel()

	e := crontask.MustParse("0 * * * *")
	if e.Raw() != "0 * * * *" {
		t.Errorf("Raw() = %q, want %q", e.Raw(), "0 * * * *")
	}

	// Confirm Next works.
	ref := time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC)
	next := e.Next(ref)
	want := time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Errorf("Expression.Next = %v, want %v", next, want)
	}

	// MustParse must panic for invalid input.
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse with invalid expression did not panic")
		}
	}()
	crontask.MustParse("not valid")
}

// TestExpressionNextN verifies Expression.NextN.
func TestExpressionNextN(t *testing.T) {
	t.Parallel()

	e := crontask.MustParse("0 * * * *")
	ref := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	runs := e.NextN(ref, 3)
	if len(runs) != 3 {
		t.Fatalf("expected 3 runs, got %d", len(runs))
	}

	// n <= 0 returns nil.
	if e.NextN(ref, 0) != nil {
		t.Error("NextN with n=0 should return nil")
	}
}

// TestExpressionIsDue verifies the Expression.IsDue method.
func TestExpressionIsDue(t *testing.T) {
	t.Parallel()

	e := crontask.MustParse("0 * * * *")
	at := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if !e.IsDue(at) {
		t.Error("IsDue should be true at hour boundary")
	}
	notAt := time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC)
	if e.IsDue(notAt) {
		t.Error("IsDue should be false at non-boundary")
	}
}

// TestIsDueSixField verifies IsDue with a six-field (seconds) expression.
func TestIsDueSixField(t *testing.T) {
	t.Parallel()

	at := time.Date(2024, 1, 1, 0, 0, 30, 0, time.UTC)
	if !crontask.IsDue("30 * * * * *", at) {
		t.Error("IsDue should be true at second boundary for six-field expr")
	}
	notAt := time.Date(2024, 1, 1, 0, 0, 31, 0, time.UTC)
	if crontask.IsDue("30 * * * * *", notAt) {
		t.Error("IsDue should be false for non-matching second")
	}
}
