package crontask

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/strutil"
)

// isDue is the internal implementation shared by the package-level IsDue and
// Expression.IsDue. It returns true when the first activation of sched after
// (at − 1 second) is at or before at.
//
// Example:
//
//	now := time.Now().Truncate(time.Minute)
//	if expr.IsDue(now) {
//		sendDailyReport()
//	}
func isDue(sched Schedule, at time.Time) bool {
	at = at.Truncate(time.Second)
	next := sched.Next(at.Add(-time.Second))
	return !next.IsZero() && !next.After(at)
}

// isWild reports whether a cron field token represents "all values".
//
// Example:
//
//	isWild("*") → true
//	isWild("?") → true
//	isWild("1") → false
//	isWild("*/2") → false
func isWild(f string) bool {
	return f == "*" || f == "?"
}

// extractStep extracts the step value from a "*/n" field token. It returns 0
// when the field is not a step expression or the step is invalid.
//
// Example:
//
//	extractStep("*/2") → 2
//	extractStep("*/") → 0
//	extractStep("1") → 0
func extractStep(f string) int {
	if strutil.IsEmpty(f) {
		return 0
	}
	if !strings.HasPrefix(f, "*/") {
		return 0
	}
	n := conv.IntOrDefault(f[2:], 0)
	if n <= 0 {
		return 0
	}
	return n
}

// parseInt parses a bare integer field token.
//
// Example:
//
//	parseInt("1") → (1, true)
//	parseInt("*") → (0, false)
//	parseInt("*/2") → (0, false)
func parseInt(f string) (int, bool) {
	if strutil.IsEmpty(f) {
		return 0, false
	}
	n, err := conv.Int(f)
	return n, err == nil
}

// parseRange parses a "lo-hi" field token into its two components.
//
// Example:
//
//	parseRange("1-5") → ([1, 5], true)
//	parseRange("*") → ([0, 0], false)
//	parseRange("*/2") → ([0, 0], false)
func parseRange(f string) ([2]int, bool) {
	if strutil.IsEmpty(f) {
		return [2]int{}, false
	}
	idx := strings.Index(f, "-")
	if idx < 0 {
		return [2]int{}, false
	}
	lo, err1 := conv.Int(f[:idx])
	hi, err2 := conv.Int(f[idx+1:])
	if err1 != nil || err2 != nil {
		return [2]int{}, false
	}
	return [2]int{lo, hi}, true
}

// dowLabel returns the English day name for a 0-based weekday index.
// Day of week: 0 = Sunday, 1 = Monday, ..., 6 = Saturday.
//
// Example:
//
//	dowLabel(0) → "Sunday"
//	dowLabel(1) → "Monday"
//	dowLabel(6) → "Saturday"
func dowLabel(n int) string {
	labels := [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	if n >= 0 && n < 7 {
		return labels[n]
	}
	return strconv.Itoa(n)
}

// ordinal converts a positive integer to its English ordinal string
// (1 → "1st", 2 → "2nd", 3 → "3rd", 4 → "4th", …).
//
// Example:
//
//	ordinal(1) → "1st"
//	ordinal(2) → "2nd"
//	ordinal(3) → "3rd"
//	ordinal(4) → "4th"
func ordinal(n int) string {
	switch n % 100 {
	case 11, 12, 13:
		return fmt.Sprintf("%dth", n)
	}
	switch n % 10 {
	case 1:
		return fmt.Sprintf("%dst", n)
	case 2:
		return fmt.Sprintf("%dnd", n)
	case 3:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}

// joinEnglish joins a slice of strings with commas and an Oxford "and".
//
// Example:
//
//	[]string{"a"} → "a"
//	[]string{"a", "b"} → "a and b"
//	[]string{"a", "b", "c"} → "a, b, and c"
func joinEnglish(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
	}
}

// describeMonthField returns an English description of a month field.
//
// Example:
//
//	describeMonthField("1") → "January"
//	describeMonthField("1-3") → "January through March"
//	describeMonthField("1,3,5") → "January, March, and May"
func describeMonthField(field string) string {
	if strutil.IsEmpty(field) {
		return ""
	}

	lower := strings.ToLower(field)
	if n, ok := monthNames[lower]; ok {
		return monthLabels[n]
	}
	if n, ok := parseInt(field); ok && n >= 1 && n <= 12 {
		return monthLabels[n]
	}
	if r, ok := parseRange(field); ok && r[0] >= 1 && r[0] <= 12 && r[1] >= 1 && r[1] <= 12 {
		return monthLabels[r[0]] + " through " + monthLabels[r[1]]
	}
	parts := strings.Split(field, ",")
	if len(parts) > 1 {
		names := make([]string, len(parts))
		for i, p := range parts {
			p = strings.TrimSpace(p)
			if n, ok2 := monthNames[strings.ToLower(p)]; ok2 {
				names[i] = monthLabels[n]
			} else if n, err := strconv.Atoi(p); err == nil && n >= 1 && n <= 12 {
				names[i] = monthLabels[n]
			} else {
				names[i] = p
			}
		}
		return joinEnglish(names)
	}
	return field
}

// describeDOMField returns an English description of a day-of-month field.
//
// Example:
//
//	describeDOMField("1") → "1st"
//	describeDOMField("1-3") → "1st through 3rd"
//	describeDOMField("1,3,5") → "1st, 3rd, and 5th"
func describeDOMField(field string) string {
	if strutil.IsEmpty(field) {
		return ""
	}

	if n, ok := parseInt(field); ok {
		return ordinal(n)
	}
	if r, ok := parseRange(field); ok {
		return ordinal(r[0]) + " through the " + ordinal(r[1])
	}
	parts := strings.Split(field, ",")
	if len(parts) > 1 {
		ords := make([]string, len(parts))
		for i, p := range parts {
			if n, err := conv.Int(strings.TrimSpace(p)); err == nil {
				ords[i] = ordinal(n)
			} else {
				ords[i] = p
			}
		}
		return joinEnglish(ords)
	}
	return field
}

// describeDOWField returns an English description of a day-of-week field value.
// Common patterns are handled with fast-path switch cases; the general path
// parses the field as a list of numeric or named values.
//
// Example:
//
//	describeDOWField("1") → "Monday"
//	describeDOWField("1-3") → "Monday through Wednesday"
//	describeDOWField("1,3,5") → "Monday, Wednesday, and Friday"
func describeDOWField(field string) string {
	if strutil.IsEmpty(field) {
		return ""
	}

	lower := strings.ToLower(field)
	switch lower {
	case "1-5", "mon-fri":
		return "Monday through Friday"
	case "0", "7", "sun":
		return "Sunday"
	case "1", "mon":
		return "Monday"
	case "2", "tue":
		return "Tuesday"
	case "3", "wed":
		return "Wednesday"
	case "4", "thu":
		return "Thursday"
	case "5", "fri":
		return "Friday"
	case "6", "sat":
		return "Saturday"
	}
	// Numeric range "lo-hi".
	if r, ok := parseRange(field); ok && r[0] >= 0 && r[0] <= 6 && r[1] >= 0 && r[1] <= 6 {
		return dowLabel(r[0]) + " through " + dowLabel(r[1])
	}
	// Comma-separated list — process in input order so the caller controls ordering.
	parts := strings.Split(field, ",")
	if len(parts) > 1 {
		names := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if n, err := conv.Int(p); err == nil {
				if n == 7 {
					n = 0
				}
				names = append(names, dowLabel(n))
			} else {
				// Named value: capitalise and use as-is.
				names = append(names, strutil.Capitalize(p))
			}
		}
		return joinEnglish(names)
	}
	return field
}

// explainIntervalDuration converts a Duration into an English phrase such as
// "Every 5 minutes" or "Every 30 seconds".
//
// Example:
//
//	explainIntervalDuration(5 * time.Minute) → "Every 5 minutes"
//	explainIntervalDuration(time.Hour) → "Every hour"
//	explainIntervalDuration(24 * time.Hour) → "Every day"
//	explainIntervalDuration(2 * 24 * time.Hour) → "Every 2 days"
func explainIntervalDuration(d time.Duration) string {
	if d < time.Minute {
		secs := int(d.Seconds())
		if secs == 1 {
			return "Every second"
		}
		return fmt.Sprintf("Every %d seconds", secs)
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		if mins == 1 {
			return "Every minute"
		}
		return fmt.Sprintf("Every %d minutes", mins)
	}
	if d < 24*time.Hour {
		hrs := int(d.Hours())
		if hrs == 1 {
			return "Every hour"
		}
		return fmt.Sprintf("Every %d hours", hrs)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "Every day"
	}
	return fmt.Sprintf("Every %d days", days)
}

// explainTimePart builds the time-of-day sentence fragment from the minute and
// hour fields. It covers the most common patterns and falls back to a generic
// description for unusual combinations.
//
// Example:
//
//	explainTimePart("0", "9") → "Every hour at 09:00"
//	explainTimePart("0", "9-17") → "Every hour at 09:00 through 17:00"
//	explainTimePart("*/15", "9") → "Every 15 minutes at 09:00"
func explainTimePart(min, hr string) string {
	if strutil.IsEmpty(min) || strutil.IsEmpty(hr) {
		return ""
	}

	minWild := isWild(min)
	hrWild := isWild(hr)
	minStep := extractStep(min)
	hrStep := extractStep(hr)

	// All wildcards: "Every minute".
	if minWild && hrWild {
		return "Every minute"
	}
	// "*/N * ..." — every N minutes.
	if minStep > 0 && hrWild {
		if minStep == 1 {
			return "Every minute"
		}
		return fmt.Sprintf("Every %d minutes", minStep)
	}
	// "0 */N ..." — every N hours (only when minute is exactly 0).
	if min == "0" && hrStep > 0 {
		if hrStep == 1 {
			return "Every hour"
		}
		return fmt.Sprintf("Every %d hours", hrStep)
	}
	// "0 * ..." — every hour.
	if min == "0" && hrWild {
		return "Every hour"
	}
	// Both minute and hour are single integers — "At HH:MM".
	if mInt, mOK := parseInt(min); mOK {
		if hInt, hOK := parseInt(hr); hOK {
			return fmt.Sprintf("At %02d:%02d", hInt, mInt)
		}
		// Single minute, wild hour — "Every hour at :MM".
		if hrWild {
			return fmt.Sprintf("Every hour at :%02d", mInt)
		}
	}
	// Wild minute, single hour — "Every minute of hour HH".
	if hInt, hOK := parseInt(hr); hOK && minWild {
		return fmt.Sprintf("Every minute of hour %02d", hInt)
	}
	// Step minute, single hour — "Every N minutes during hour HH".
	if minStep > 0 {
		if hInt, hOK := parseInt(hr); hOK {
			return fmt.Sprintf("Every %d minutes during hour %02d", minStep, hInt)
		}
	}
	// Hour range with minute 0 — "Every hour from HH:00 to HH:00".
	if hrRange, ok := parseRange(hr); ok && min == "0" {
		return fmt.Sprintf("Every hour from %02d:00 to %02d:00", hrRange[0], hrRange[1])
	}
	// Generic fallback.
	return fmt.Sprintf("At minute %s of hour %s", min, hr)
}

// explainFiveFields describes a standard five-field expression by combining a
// time-of-day description with optional day-of-week, day-of-month, and month
// qualifiers.
//
// Example:
//
//	explainFiveFields("0", "9", "1", "1", "*") → "Every hour at 09:00 on the 1st of January"
//	explainFiveFields("0", "9", "1", "1", "1") → "Every hour at 09:00 on the 1st of January, every Monday"
//	explainFiveFields("0", "9", "1", "1", "1-5") → "Every hour at 09:00 on the 1st of January, every Monday through Friday"
func explainFiveFields(min, hr, dom, mon, dow string) string {
	timeDesc := explainTimePart(min, hr)

	var qualifiers []string
	if !isWild(dom) {
		qualifiers = append(qualifiers, "on the "+describeDOMField(dom)+" of each month")
	}
	if !isWild(mon) {
		qualifiers = append(qualifiers, "in "+describeMonthField(mon))
	}
	if !isWild(dow) {
		qualifiers = append(qualifiers, describeDOWField(dow))
	}

	if len(qualifiers) == 0 {
		return timeDesc
	}
	return timeDesc + ", " + strings.Join(qualifiers, ", ")
}

// explainSixFields describes a six-field (seconds-first) expression.
//
// Example:
//
//	explainSixFields("0", "0", "9", "1", "1", "*") → "Every hour at 09:00 on the 1st of January"
//	explainSixFields("0", "0", "9", "1", "1", "1") → "Every hour at 09:00 on the 1st of January, every Monday"
func explainSixFields(sec, min, hr, dom, mon, dow string) string {
	// All wildcards — fires every second.
	if isWild(sec) && isWild(min) && isWild(hr) && isWild(dom) && isWild(mon) && isWild(dow) {
		return "Every second"
	}
	// Pure second-based patterns.
	if step := extractStep(sec); step > 0 &&
		isWild(min) && isWild(hr) && isWild(dom) && isWild(mon) && isWild(dow) {
		if step == 1 {
			return "Every second"
		}
		return fmt.Sprintf("Every %d seconds", step)
	}
	if n, ok := parseInt(sec); ok &&
		isWild(min) && isWild(hr) && isWild(dom) && isWild(mon) && isWild(dow) {
		return fmt.Sprintf("At second :%02d of every minute", n)
	}

	// Fall through: describe the five remaining fields and annotate the second
	// field only when it is not zero or wildcard.
	base := explainFiveFields(min, hr, dom, mon, dow)
	if sec != "0" && !isWild(sec) {
		return base + fmt.Sprintf(" (at second %s)", sec)
	}
	return base
}

// explainFields produces an English description from a pre-split field slice.
//
// Example:
//
//	explainFields([]string{"0", "9", "1", "1", "*"}) → "Every hour at 09:00 on the 1st of January"
//	explainFields([]string{"0", "0", "9", "1", "1", "1"}) → "Every hour at 09:00 on the 1st of January, every Monday"
//	explainFields([]string{"0", "0", "9", "1", "1", "1", "1"}) → "Every hour at 09:00 on the 1st of January, every Monday"
func explainFields(fields []string) string {
	switch len(fields) {
	case 5:
		return explainFiveFields(fields[0], fields[1], fields[2], fields[3], fields[4])
	case 6:
		return explainSixFields(fields[0], fields[1], fields[2], fields[3], fields[4], fields[5])
	default:
		return "Custom schedule"
	}
}

// recordResult stores the outcome of a single execution into the entry's
// mutable state.
func recordResult(e *entry, t time.Time, err error) {
	e.mu.Lock()
	e.lastRun = t
	e.lastErr = err
	e.runCount++
	e.mu.Unlock()
}

// updateNextRun recomputes the next activation time for e using the
// scheduler's reference time as the base.
func updateNextRun(e *entry, now time.Time) {
	e.mu.Lock()
	e.nextRun = e.schedule.Next(now)
	e.mu.Unlock()
}

// parseNamedOrInt attempts to interpret s as a named month/day-of-week
// abbreviation first, then falls back to integer parsing.
//
// Example:
//
//	parseNamedOrInt("jan", true, false) → 1, nil
//	parseNamedOrInt("1", true, false) → 1, nil
//	parseNamedOrInt("foo", true, false) → 0, error
func parseNamedOrInt(s string, isMonth, isDow bool) (int, error) {
	if strutil.IsEmpty(s) {
		return 0, fmt.Errorf("empty string")
	}
	lower := strings.ToLower(s)
	if isMonth {
		if n, ok := monthNames[lower]; ok {
			return n, nil
		}
	}
	if isDow {
		if n, ok := dowNames[lower]; ok {
			return n, nil
		}
	}
	n, err := conv.Int(s)
	if err != nil {
		return 0, fmt.Errorf("not a number or recognised name: %q", s)
	}
	return n, nil
}

// setBit marks index v as active in the bits slice. For fields with a
// minimum of 1 (day-of-month, month) the index is stored as-is; for
// day-of-week, value 7 is normalised to 0 (both represent Sunday).
//
// Example:
//
//	setBit(&bits, 1, fieldSpec{name: "day-of-month"})
//	setBit(&bits, 7, fieldSpec{name: "day-of-week"})
func setBit(bits *[]bool, v int, spec fieldSpec) {
	if spec.name == "day-of-week" && v == 7 {
		v = 0
	}
	if v >= 0 && v < len(*bits) {
		(*bits)[v] = true
	}
}

// parsePart handles a single, comma-free element of a cron field: *, n, n-m,
// n-m/step, */step.
//
// Example:
//
//	parsePart("0 9 * * 1-5", "0", 0, cronFields[0], &s.minute)
func parsePart(expr, part string, fieldIndex int, spec fieldSpec, bits *[]bool) error {
	// Step suffix.
	step := 1
	if idx := strings.Index(part, "/"); idx >= 0 {
		stepStr := part[idx+1:]
		var err error
		step, err = strconv.Atoi(stepStr)
		if err != nil || step <= 0 {
			return newExpressionError(expr, fieldIndex, "invalid step %q in %q", stepStr, part)
		}
		part = part[:idx]
	}

	isMonth := spec.name == "month"
	isDow := spec.name == "day-of-week"

	// Wildcard.
	if part == "*" || part == "?" {
		for v := spec.min; v <= spec.max; v += step {
			setBit(bits, v, spec)
		}
		return nil
	}

	// Range or single value.
	rangeMin, rangeMax := spec.min, spec.max
	if idx := strings.Index(part, "-"); idx >= 0 {
		loStr, hiStr := part[:idx], part[idx+1:]
		lo, err := parseNamedOrInt(loStr, isMonth, isDow)
		if err != nil {
			return newExpressionError(expr, fieldIndex, "invalid range start %q: %v", loStr, err)
		}
		hi, err := parseNamedOrInt(hiStr, isMonth, isDow)
		if err != nil {
			return newExpressionError(expr, fieldIndex, "invalid range end %q: %v", hiStr, err)
		}
		if lo < spec.min || lo > spec.max {
			return newExpressionError(expr, fieldIndex, "range start %d out of [%d,%d]", lo, spec.min, spec.max)
		}
		if hi < spec.min || hi > spec.max {
			return newExpressionError(expr, fieldIndex, "range end %d out of [%d,%d]", hi, spec.min, spec.max)
		}
		if lo > hi {
			return newExpressionError(expr, fieldIndex, "range start %d is greater than end %d", lo, hi)
		}
		rangeMin, rangeMax = lo, hi
	} else {
		// Single value.
		v, err := parseNamedOrInt(part, isMonth, isDow)
		if err != nil {
			return newExpressionError(expr, fieldIndex, "invalid value %q: %v", part, err)
		}
		if v < spec.min || v > spec.max {
			return newExpressionError(expr, fieldIndex, "value %d out of [%d,%d]", v, spec.min, spec.max)
		}
		rangeMin, rangeMax = v, v
	}

	for v := rangeMin; v <= rangeMax; v += step {
		setBit(bits, v, spec)
	}
	return nil
}

// parseField populates the bits slice for a single cron field expression.
// The fieldIndex parameter is used only for error reporting.
//
// Example:
//
//	parseField("0 9 * * 1-5", "0", 0, cronFields[0], &s.minute)
func parseField(expr, field string, fieldIndex int, spec fieldSpec, bits *[]bool) error {
	// Comma-separated list.
	parts := strings.Split(field, ",")
	for _, part := range parts {
		if err := parsePart(expr, part, fieldIndex, spec, bits); err != nil {
			return err
		}
	}
	return nil
}

// parseSixField parses a six-field (seconds-first) cron expression.
//
// Example:
//
//	parseSixField("0 9 * * 1-5 1", "0", 0, cronFields[0], &s.minute)
func parseSixField(raw string, fields []string, loc *time.Location) (*cronSchedule, error) {
	s := &cronSchedule{
		second:     make([]bool, 60),
		minute:     make([]bool, 60),
		hour:       make([]bool, 24),
		dayOfMonth: make([]bool, 32),
		month:      make([]bool, 13),
		dayOfWeek:  make([]bool, 8), // indices 0..6; 7 is normalised to 0
		loc:        loc,
	}
	specs := cronFieldsWithSeconds
	targets := [6]*[]bool{&s.second, &s.minute, &s.hour, &s.dayOfMonth, &s.month, &s.dayOfWeek}
	for i, f := range fields {
		if err := parseField(raw, f, i, specs[i], targets[i]); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// parseFiveField parses a standard five-field cron expression.
//
// Example:
//
//	parseFiveField("0 9 * * 1-5", "0", 0, cronFields[0], &s.minute)
func parseFiveField(raw string, fields []string, loc *time.Location) (*cronSchedule, error) {
	s := &cronSchedule{
		minute:     make([]bool, 60),
		hour:       make([]bool, 24),
		dayOfMonth: make([]bool, 32), // index 0 unused; days 1..31
		month:      make([]bool, 13), // index 0 unused; months 1..12
		dayOfWeek:  make([]bool, 8),  // indices 0..6; 7 is normalised to 0
		loc:        loc,
	}
	specs := cronFields
	targets := [5]*[]bool{&s.minute, &s.hour, &s.dayOfMonth, &s.month, &s.dayOfWeek}
	for i, f := range fields {
		if err := parseField(raw, f, i, specs[i], targets[i]); err != nil {
			return nil, err
		}
	}
	return s, nil
}
