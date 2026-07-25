package crontask_test

import (
	"errors"
	"testing"
	"time"

	"github.com/polarixa/replify/pkg/crontask"
)

// TestParseFiveField exercises the standard five-field parser with a variety
// of valid and invalid expressions.
func TestParseFiveField(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"wildcard all", "* * * * *", false},
		{"specific minute", "30 * * * *", false},
		{"range", "0-30 * * * *", false},
		{"step", "*/5 * * * *", false},
		{"range with step", "0-30/5 * * * *", false},
		{"comma list", "1,15,30 * * * *", false},
		{"month name", "0 0 1 jan *", false},
		{"month name upper", "0 0 1 JAN *", false},
		{"dow name", "0 0 * * mon", false},
		{"dow range", "0 0 * * mon-fri", false},
		{"mixed comma and range", "0 0 * * mon,wed,fri", false},
		{"at midnight", "0 0 * * *", false},
		{"every 15 minutes", "*/15 * * * *", false},
		{"empty", "", true},
		{"too few fields", "* * * *", true},
		{"too many fields (7)", "* * * * * * *", true},
		{"minute out of range", "60 * * * *", true},
		{"hour out of range", "0 24 * * *", true},
		{"dom out of range", "0 0 32 * *", true},
		{"month out of range", "0 0 1 13 *", true},
		{"dow out of range", "0 0 * * 8", true},
		{"range start > end", "30-10 * * * *", true},
		{"invalid step", "*/0 * * * *", true},
		{"non-numeric", "abc * * * *", true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := crontask.Parse(tc.expr)
			if tc.wantErr && err == nil {
				t.Errorf("Parse(%q): expected error, got nil", tc.expr)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Parse(%q): unexpected error: %v", tc.expr, err)
			}
		})
	}
}

// TestParseSixField exercises the optional seconds field.
func TestParseSixField(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"six wildcard", "* * * * * *", false},
		{"every 30 seconds", "*/30 * * * * *", false},
		{"seconds out of range", "60 * * * * *", true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := crontask.Parse(tc.expr)
			if tc.wantErr && err == nil {
				t.Errorf("Parse(%q): expected error, got nil", tc.expr)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Parse(%q): unexpected error: %v", tc.expr, err)
			}
		})
	}
}

// TestParseAlias verifies that every recognised alias resolves correctly.
func TestParseAlias(t *testing.T) {
	t.Parallel()
	aliases := []string{
		"@yearly", "@annually", "@monthly", "@weekly",
		"@daily", "@midnight", "@hourly", "@minutely",
		"@weekdays", "@weekends",
		// Business aliases.
		"@businessDaily", "@businessHourly",
		"@quarterly", "@semiMonthly",
		"@workhours", "@marketOpen", "@marketClose",
	}
	for _, a := range aliases {
		a := a
		t.Run(a, func(t *testing.T) {
			t.Parallel()
			_, err := crontask.Parse(a)
			if err != nil {
				t.Errorf("Parse(%q): unexpected error: %v", a, err)
			}
		})
	}
}

// TestParseUnknownAlias ensures that unknown aliases are rejected with a
// typed ExpressionError.
func TestParseUnknownAlias(t *testing.T) {
	t.Parallel()
	_, err := crontask.Parse("@unknown")
	if err == nil {
		t.Fatal("expected error for unknown alias, got nil")
	}
	if !errors.Is(err, crontask.ErrInvalidExpression) {
		t.Errorf("expected ErrInvalidExpression, got %v", err)
	}
}

// TestParseEvery exercises the @every interval syntax.
func TestParseEvery(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"5 minutes", "@every 5m", false},
		{"30 seconds", "@every 30s", false},
		{"1 hour", "@every 1h", false},
		{"bad duration", "@every bad", true},
		{"zero duration", "@every 0s", true},
		{"negative duration", "@every -1m", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := crontask.Parse(tc.expr)
			if tc.wantErr && err == nil {
				t.Errorf("Parse(%q): expected error, got nil", tc.expr)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Parse(%q): unexpected error: %v", tc.expr, err)
			}
		})
	}
}

// TestParseTZ validates timezone-prefixed expressions.
func TestParseTZ(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"UTC", "TZ=UTC 0 * * * *", false},
		{"US Eastern", "TZ=America/New_York 0 9 * * 1-5", false},
		{"unknown TZ", "TZ=Mars/Olympus 0 * * * *", true},
		{"missing fields after TZ", "TZ=UTC", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := crontask.Parse(tc.expr)
			if tc.wantErr && err == nil {
				t.Errorf("Parse(%q): expected error, got nil", tc.expr)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Parse(%q): unexpected error: %v", tc.expr, err)
			}
		})
	}
}

// TestValidate mirrors Parse but via the public Validate wrapper.
func TestValidate(t *testing.T) {
	t.Parallel()
	if err := crontask.Validate("0 * * * *"); err != nil {
		t.Errorf("Validate: unexpected error for valid expression: %v", err)
	}
	if err := crontask.Validate("bad expression here"); err == nil {
		t.Error("Validate: expected error for invalid expression, got nil")
	}
}

// TestNextFiveField verifies Next returns the correct time for a five-field
// schedule with a known reference point.
func TestNextFiveField(t *testing.T) {
	t.Parallel()

	// "0 * * * *" — top of every hour.
	sched, err := crontask.Parse("0 * * * *")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ref := time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC)
	got := sched.Next(ref)
	want := time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("Next(%v) = %v, want %v", ref, got, want)
	}
}

// TestNextSixField verifies Next returns the correct time for a six-field
// (seconds-first) schedule.
func TestNextSixField(t *testing.T) {
	t.Parallel()

	// "30 * * * * *" — every minute at :30 seconds.
	sched, err := crontask.Parse("30 * * * * *")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ref := time.Date(2024, 6, 15, 10, 00, 00, 0, time.UTC)
	got := sched.Next(ref)
	want := time.Date(2024, 6, 15, 10, 00, 30, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("Next(%v) = %v, want %v", ref, got, want)
	}
}

// TestNextLeapYear checks that day-29 scheduling works in a leap year.
func TestNextLeapYear(t *testing.T) {
	t.Parallel()

	// "0 0 29 2 *" — midnight on 29 February.
	sched, err := crontask.Parse("0 0 29 2 *")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// 2024 is a leap year; from 1 Jan 2024 the next occurrence is 29 Feb 2024.
	ref := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := sched.Next(ref)
	want := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("Next(%v) = %v, want %v", ref, got, want)
	}
}

// TestNextIntervalSchedule verifies the intervalSchedule implementation.
func TestNextIntervalSchedule(t *testing.T) {
	t.Parallel()

	sched, err := crontask.Parse("@every 10m")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ref := time.Date(2024, 1, 1, 0, 3, 0, 0, time.UTC)
	got := sched.Next(ref)
	// Next 10-minute boundary after 00:03:00 is 00:10:00.
	want := time.Date(2024, 1, 1, 0, 10, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("Next(%v) = %v, want %v", ref, got, want)
	}
}

// TestExpressionErrorUnwrap verifies sentinel error wrapping.
func TestExpressionErrorUnwrap(t *testing.T) {
	t.Parallel()
	_, err := crontask.Parse("bad")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, crontask.ErrInvalidExpression) {
		t.Errorf("errors.Is(err, ErrInvalidExpression) = false, want true")
	}
}

// TestMonthNameCaseInsensitive verifies that month names are accepted in any
// case.
func TestMonthNameCaseInsensitive(t *testing.T) {
	t.Parallel()
	exprs := []string{
		"0 0 1 jan *",
		"0 0 1 JAN *",
		"0 0 1 Jan *",
	}
	for _, expr := range exprs {
		if err := crontask.Validate(expr); err != nil {
			t.Errorf("Validate(%q): unexpected error: %v", expr, err)
		}
	}
}

// TestDowSundayNormalisation checks that both 0 and 7 are accepted as Sunday.
func TestDowSundayNormalisation(t *testing.T) {
	t.Parallel()
	// "0 0 * * 7" — should parse the same as "0 0 * * 0".
	s7, err := crontask.Parse("0 0 * * 7")
	if err != nil {
		t.Fatalf("Parse(7): %v", err)
	}
	s0, err := crontask.Parse("0 0 * * 0")
	if err != nil {
		t.Fatalf("Parse(0): %v", err)
	}
	// Both should fire on the same day when given the same reference time.
	ref := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) // Monday 1 Jan 2024
	n7 := s7.Next(ref)
	n0 := s0.Next(ref)
	if !n7.Equal(n0) {
		t.Errorf("Next with dow=7 (%v) != Next with dow=0 (%v)", n7, n0)
	}
}

// TestRegisterAlias verifies the custom alias registration workflow.
func TestRegisterAlias(t *testing.T) {
	t.Parallel()

	const name = "@test-alias"

	// A valid five-field expression should register successfully.
	if err := crontask.RegisterAlias(name, "0 3 * * *"); err != nil {
		t.Fatalf("RegisterAlias: unexpected error: %v", err)
	}
	// Parse must now succeed for the custom alias.
	if _, err := crontask.Parse(name); err != nil {
		t.Errorf("Parse after RegisterAlias: %v", err)
	}
	// Registering with a new expression overwrites the old one.
	if err := crontask.RegisterAlias(name, "0 4 * * *"); err != nil {
		t.Fatalf("RegisterAlias overwrite: %v", err)
	}
	// Cleanup via the exported DeleteAlias function.
	if err := crontask.DeleteAlias(name); err != nil {
		t.Errorf("DeleteAlias: %v", err)
	}
}

// TestRegisterAliasErrors verifies rejection of malformed registrations.
func TestRegisterAliasErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		expr string
	}{
		// Name does not start with "@".
		{"no-at-prefix", "0 * * * *"},
		// Expression has wrong field count.
		{"@bad-count", "0 * * *"},
		// Expression has an invalid field value.
		{"@bad-value", "99 * * * *"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := crontask.RegisterAlias(tc.name, tc.expr); err == nil {
				t.Errorf("RegisterAlias(%q, %q): expected error, got nil", tc.name, tc.expr)
			}
		})
	}
}

// TestBusinessAliasesNextRun spot-checks that business alias schedules compute
// sensible next-run times.
func TestBusinessAliasesNextRun(t *testing.T) {
	t.Parallel()
	// @businessDaily fires at 09:00 on weekdays.
	// Use a Monday at 08:00 as the reference; next should be 09:00 same day.
	ref := time.Date(2024, 1, 8, 8, 0, 0, 0, time.UTC) // Monday
	sched, err := crontask.Parse(crontask.AliasBusinessDaily)
	if err != nil {
		t.Fatalf("Parse(%q): %v", crontask.AliasBusinessDaily, err)
	}
	next := sched.Next(ref)
	want := time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Errorf("AliasBusinessDaily Next = %v, want %v", next, want)
	}

	// @marketOpen fires at 09:30 on weekdays.
	sched2, err := crontask.Parse(crontask.AliasMarketOpen)
	if err != nil {
		t.Fatalf("Parse(%q): %v", crontask.AliasMarketOpen, err)
	}
	next2 := sched2.Next(ref)
	want2 := time.Date(2024, 1, 8, 9, 30, 0, 0, time.UTC)
	if !next2.Equal(want2) {
		t.Errorf("AliasMarketOpen Next = %v, want %v", next2, want2)
	}

	// @quarterly fires on 1 Apr after 1 Jan.
	refQ := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	schedQ, err := crontask.Parse(crontask.AliasQuarterly)
	if err != nil {
		t.Fatalf("Parse(%q): %v", crontask.AliasQuarterly, err)
	}
	nextQ := schedQ.Next(refQ)
	wantQ := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	if !nextQ.Equal(wantQ) {
		t.Errorf("AliasQuarterly Next = %v, want %v", nextQ, wantQ)
	}
}
