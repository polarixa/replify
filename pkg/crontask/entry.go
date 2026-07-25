package crontask

import (
	"fmt"
	"strings"
	"time"

	"github.com/polarixa/replify/pkg/strutil"
)

// Parse converts a cron expression string into a Schedule that can be used
// to compute successive activation times.
//
// Supported formats
//
//   - Five fields:  "minute hour day-of-month month day-of-week"
//   - Six fields:   "second minute hour day-of-month month day-of-week"
//   - @alias:       see the Alias constants for the full list
//   - @every <d>:   interval expression, e.g. "@every 5m"
//
// Field syntax (per field)
//
//   - *                — every value in the valid range
//   - n                — exact value
//   - n-m              — inclusive range
//   - n-m/step         — range with step
//   - */step           — every step values across the full range
//   - a,b,c            — comma-separated list (each element may itself use
//     any of the above forms)
//
// Month and day-of-week fields additionally accept three-letter English
// abbreviations (jan-dec and sun-sat respectively), case-insensitive.
//
// # Timezone
//
// An optional IANA timezone specifier may appear at the front of the
// expression, separated from the fields by a space:
//
//	"TZ=America/New_York 0 9 * * 1-5"
//
// When a timezone is provided, the returned Schedule activates at the
// specified local time. When omitted, UTC is used.
func Parse(expr string) (Schedule, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, &ExpressionError{Expression: expr, Field: -1, Reason: "expression is empty"}
	}

	// Extract optional leading TZ= specifier.
	loc := time.UTC
	if strings.HasPrefix(expr, "TZ=") {
		idx := strings.Index(expr, " ")
		if idx < 0 {
			return nil, &ExpressionError{Expression: expr, Field: -1, Reason: "missing fields after TZ specifier"}
		}
		tzName := expr[3:idx]
		var err error
		loc, err = time.LoadLocation(tzName)
		if err != nil {
			return nil, &ExpressionError{Expression: expr, Field: -1, Reason: fmt.Sprintf("unknown timezone %q: %v", tzName, err)}
		}
		expr = strings.TrimSpace(expr[idx+1:])
	}

	// Handle @every interval expressions.
	if strings.HasPrefix(expr, "@every ") {
		durationStr := strings.TrimPrefix(expr, "@every ")
		d, err := time.ParseDuration(strings.TrimSpace(durationStr))
		if err != nil {
			return nil, &ExpressionError{Expression: expr, Field: -1, Reason: fmt.Sprintf("invalid duration %q: %v", durationStr, err)}
		}
		if d <= 0 {
			return nil, &ExpressionError{Expression: expr, Field: -1, Reason: "interval duration must be positive"}
		}
		return &intervalSchedule{interval: d}, nil
	}

	// Handle @alias expressions.
	if strings.HasPrefix(expr, "@") {
		expanded, ok := lookupAlias(strings.ToLower(expr))
		if !ok {
			return nil, &ExpressionError{Expression: expr, Field: -1, Reason: fmt.Sprintf("unknown alias %q", expr)}
		}
		expr = expanded
	}

	// Split into fields.
	fields := strings.Fields(expr)
	switch len(fields) {
	case 5:
		return parseFiveField(expr, fields, loc)
	case 6:
		return parseSixField(expr, fields, loc)
	default:
		return nil, &ExpressionError{
			Expression: expr, Field: -1,
			Reason: fmt.Sprintf("expected 5 or 6 fields, got %d", len(fields)),
		}
	}
}

// MustParse is like Parse but panics instead of returning an error when the
// expression is invalid. It is intended for use in package-level variable
// initializers where the expression is a compile-time constant.
//
// Example:
//
//	var reportSchedule = crontask.MustParse("0 9 * * 1-5")
func MustParse(expr string) Expression {
	s, err := Parse(expr)
	if err != nil {
		panic(fmt.Sprintf("crontask.MustParse: %v", err))
	}
	return Expression{raw: expr, schedule: s}
}

// Validate reports whether the given expression is syntactically and
// semantically valid. It returns nil on success or a typed *ExpressionError
// (which also satisfies errors.Is(err, ErrInvalidExpression)).
func Validate(expr string) error {
	if strutil.IsEmpty(expr) {
		return &ExpressionError{Expression: expr, Field: -1, Reason: "expression is empty"}
	}
	_, err := Parse(expr)
	return err
}

// IsValidCronExpr reports whether expr is a syntactically and semantically
// valid cron expression recognised by this package. It is equivalent to
// Validate(expr) == nil.
//
// Example:
//
//	valid := crontask.IsValidCronExpr("0 9 * * 1-5")
//
// Thread-safe; does not require a Scheduler instance.
func IsValidCronExpr(expr string) bool {
	return Validate(expr) == nil
}

// ValidateCronExpr validates expr and returns a descriptive error when it is
// invalid. The returned error, if non-nil, is an *ExpressionError that also
// satisfies errors.Is(err, ErrInvalidExpression).
//
// ValidateCronExpr is a convenience wrapper around Validate and is provided
// for callers who prefer the longer, self-documenting name.
//
// Example:
//
//	err := crontask.ValidateCronExpr("0 9 * * 1-5")
//
// Thread-safe; does not require a Scheduler instance.
func ValidateCronExpr(expr string) error {
	return Validate(expr)
}

// IsDue reports whether the cron expression expr was due at the given time at.
//
// The check uses second-level granularity: IsDue returns true when the first
// activation of the schedule after (at − 1 second) is at or before at. For
// standard five-field expressions this means the function returns true exactly
// at the start of a matching minute; for six-field or @every expressions it
// returns true at the matching second.
//
// IsDue returns false for invalid expressions without panicking.
//
// Thread-safe; does not require a Scheduler instance.
//
// Example:
//
//	now := time.Now().Truncate(time.Minute)
//	if crontask.IsDue("0 9 * * 1-5", now) {
//	    sendDailyReport()
//	}
func IsDue(expr string, at time.Time) bool {
	sched, err := Parse(expr)
	if err != nil {
		return false
	}
	return isDue(sched, at)
}

// NextRun returns the first activation time for expr after the reference time
// from. It is a convenience wrapper around Parse and Schedule.Next that avoids
// the need to create a Scheduler or call Parse manually.
//
// Thread-safe; does not require a Scheduler instance.
//
// Example:
//
//	next, err := crontask.NextRun("0 9 * * 1-5", time.Now())
func NextRun(expr string, from time.Time) (time.Time, error) {
	sched, err := Parse(expr)
	if err != nil {
		return time.Time{}, err
	}
	return sched.Next(from), nil
}

// NextRuns returns the next n activation times for expr starting from from.
// It is a convenience wrapper around Parse and Schedule.Next.
//
// When n ≤ 0, NextRuns returns a nil slice without an error. When the schedule
// has fewer than n future activations, the returned slice is shorter than n.
//
// Thread-safe; does not require a Scheduler instance.
//
// Example:
//
//	runs, err := crontask.NextRuns("0 9 * * 1-5", time.Now(), 5)
func NextRuns(expr string, from time.Time, n int) ([]time.Time, error) {
	sched, err := Parse(expr)
	if err != nil {
		return nil, err
	}
	if n <= 0 {
		return nil, nil
	}
	out := make([]time.Time, 0, n)
	cur := from
	for len(out) < n {
		next := sched.Next(cur)
		if next.IsZero() {
			break
		}
		out = append(out, next)
		cur = next
	}
	return out, nil
}

// Explain converts a cron expression into a natural English description.
//
// Supported input forms:
//
//   - "@every 5m"               → "Every 5 minutes"
//   - "@daily", "@hourly", etc  → predefined descriptions for built-in aliases
//   - "*/30 * * * * *"          → "Every 30 seconds"
//   - "0 9 * * 1-5"             → "At 09:00, Monday through Friday"
//   - "TZ=..." prefixes         → described without the timezone qualifier
//
// Custom aliases registered via RegisterAlias are described by expanding them
// to their underlying expression and applying the field-based explainer.
//
// Explain returns ErrInvalidExpression (wrapped) for invalid input and never
// panics.
//
// Example:
//
//	desc, err := crontask.Explain("0 0 * * 1-5")
//	// desc == "At 00:00, Monday through Friday"
func Explain(expr string) (string, error) {
	if strutil.IsEmpty(expr) {
		return "", ErrInvalidExpression
	}

	trimmed := strings.TrimSpace(expr)
	// Validate first — reuse Parse to avoid duplicating validation logic.
	if _, err := Parse(trimmed); err != nil {
		return "", err
	}

	// Strip optional leading TZ= specifier; timezone does not affect the
	// "when" part of the description.
	clean := trimmed
	if strings.HasPrefix(clean, "TZ=") {
		idx := strings.Index(clean, " ")
		if idx >= 0 {
			clean = strings.TrimSpace(clean[idx+1:])
		}
	}

	// @every interval expressions.
	if strings.HasPrefix(clean, "@every ") {
		durStr := strings.TrimSpace(strings.TrimPrefix(clean, "@every "))
		d, _ := time.ParseDuration(durStr)
		return explainIntervalDuration(d), nil
	}

	// @alias expressions.
	if strings.HasPrefix(clean, "@") {
		lower := strings.ToLower(clean)
		// Check predefined descriptions first for the best output.
		if desc, ok := aliasDescriptions[lower]; ok {
			return desc, nil
		}
		// Custom alias — expand and fall through to the field-based explainer.
		if expanded, ok := lookupAlias(lower); ok {
			return explainFields(strings.Fields(expanded)), nil
		}
		return "", ErrInvalidExpression
	}

	return explainFields(strings.Fields(clean)), nil
}

// RegisterAlias registers a custom alias that can subsequently be used
// anywhere a cron expression is accepted (Register, Parse, IsDue, etc.).
//
// name must begin with "@". expr must be a valid five-field or six-field cron
// expression; @every and nested alias expressions are not accepted as the
// right-hand side of a registration.
//
// If name already exists (built-in or previously registered), it is silently
// overwritten with the new expression. Names are matched case-insensitively.
//
// Example:
//
//	err := crontask.RegisterAlias("@nightly", "0 2 * * *")
func RegisterAlias(name, expr string) error {
	if strutil.IsEmpty(name) || strutil.IsEmpty(expr) {
		return fmt.Errorf("crontask: alias name %q and expression %q cannot be empty", name, expr)
	}

	// name must begin with "@"
	if !strings.HasPrefix(name, "@") {
		return fmt.Errorf("crontask: alias name %q must begin with \"@\"", name)
	}
	// Validate that expr is a plain 5 or 6 field expression so that aliases
	// cannot create indirect chains.
	fields := strings.Fields(strings.TrimSpace(expr))
	if len(fields) != 5 && len(fields) != 6 {
		return fmt.Errorf("crontask: alias expression must have 5 or 6 fields, got %q", expr)
	}

	// Validate that expr is a valid cron expression
	if _, err := Parse(expr); err != nil {
		return err
	}
	key := strings.ToLower(name)
	aliasMapMu.Lock()
	aliasMap[key] = expr
	aliasMapMu.Unlock()
	return nil
}

// DeleteAlias removes a previously registered alias by name. The name
// comparison is case-insensitive and the "@" prefix is required.
//
// Built-in aliases can be deleted if needed, though doing so is not
// recommended for production code. An error is returned when the alias
// does not exist in the registry.
//
// Example:
//
//	crontask.RegisterAlias("@nightly", "0 2 * * *")
//	// ... use @nightly ...
//	crontask.DeleteAlias("@nightly")
func DeleteAlias(name string) error {
	if strutil.IsEmpty(name) {
		return fmt.Errorf("crontask: alias name %q cannot be empty", name)
	}

	// name must begin with "@"
	if !strings.HasPrefix(name, "@") {
		return fmt.Errorf("crontask: alias name %q must begin with \"@\"", name)
	}
	key := strings.ToLower(name)
	aliasMapMu.Lock()
	defer aliasMapMu.Unlock()

	// Check if alias exists
	if _, ok := aliasMap[key]; !ok {
		return fmt.Errorf("crontask: alias %q not found", name)
	}
	delete(aliasMap, key)
	return nil
}
