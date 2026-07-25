package replify

import (
	"reflect"

	"github.com/polarixa/replify/pkg/encoding"
	"github.com/polarixa/replify/pkg/fj"
)

// JSONBodyParser parses the body of the wrapper as JSON and returns a fj.Context for the
// entire document. This is the entry point for all fj-based operations on the wrapper.
//
// If the body is nil or cannot be serialized, a zero-value fj.Context is returned.
// Callers can check presence with ctx.Exists().
//
// Example:
//
//	ctx := w.JSONBodyParser()
//	fmt.Println(ctx.Get("user.name").String())
func (w *wrapper) JSONBodyParser() fj.Context {
	return fj.Parse(jsonpass(w.data))
}

// JSONBodyCast casts the body of the wrapper to the given type.
//
// Parameters:
//   - v: The value to cast the body to. Must be a pointer.
//
// Returns:
//   - An error if the cast fails.
//
// Example:
//
//	err := w.JSONBodyCast(&myStruct)
func (w *wrapper) JSONBodyCast(v any) error {
	if v == nil {
		return NewError("value is required")
	}
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return NewError("value must be a pointer")
	}
	json := jsonpass(w.data)
	err := encoding.StrictUnmarshalJSONString(json, v)
	return NewErrorAck(err)
}

// QueryJSONBody retrieves the value at the given fj dot-notation path from the wrapper's
// body. The body is serialized to JSON on each call; for repeated queries on the same
// body, use BodyCtx() once and chain calls on the returned Context.
//
// Parameters:
//   - path: A fj dot-notation path (e.g. "user.name", "items.#.id", "roles.0").
//
// Returns:
//   - A fj.Context for the matched value. Call .Exists() to check presence.
//
// Example:
//
//	name := w.QueryJSONBody("user.name").String()
func (w *wrapper) QueryJSONBody(path string) fj.Context {
	return fj.Get(jsonpass(w.data), path)
}

// QueryJSONBodyMulti evaluates multiple fj paths against the body in a single pass and
// returns one fj.Context per path in the same order.
//
// Parameters:
//   - paths: One or more fj dot-notation paths.
//
// Returns:
//   - A slice of fj.Context values, one per path.
//
// Example:
//
//	results := w.QueryJSONBodyMulti("user.id", "user.email", "roles.#")
func (w *wrapper) QueryJSONBodyMulti(paths ...string) []fj.Context {
	return fj.GetMulti(jsonpass(w.data), paths...)
}

// ValidJSONBody reports whether the body of the wrapper is valid JSON.
//
// Returns:
//   - true if the body serializes to well-formed JSON; false otherwise.
//
// Example:
//
//	if !w.ValidJSONBody() {
//	    log.Println("body is not valid JSON")
//	}
func (w *wrapper) ValidJSONBody() bool {
	return fj.IsValidJSONString(jsonpass(w.data))
}

// SearchJSONBody performs a full-tree scan of the body JSON and returns all scalar
// leaf values whose string representation contains the given keyword (case-sensitive
// substring match).
//
// Parameters:
//   - keyword: The substring to search for. An empty keyword matches every leaf.
//
// Returns:
//   - A slice of fj.Context values whose string representation contains keyword.
//
// Example:
//
//	hits := w.SearchJSONBody("admin")
//	for _, h := range hits {
//	    fmt.Println(h.String())
//	}
func (w *wrapper) SearchJSONBody(keyword string) []fj.Context {
	return fj.Search(jsonpass(w.data), keyword)
}

// SearchJSONBodyMatch performs a full-tree wildcard scan of the body JSON and returns
// all scalar leaf values whose string representation matches the given pattern.
//
// The pattern supports '*' (any sequence) and '?' (single character) wildcards.
//
// Parameters:
//   - pattern: A wildcard pattern applied to leaf string values.
//
// Example:
//
//	hits := w.SearchJSONBodyMatch("admin*")
func (w *wrapper) SearchJSONBodyMatch(pattern string) []fj.Context {
	return fj.SearchMatch(jsonpass(w.data), pattern)
}

// SearchJSONBodyByKey performs a full-tree scan of the body JSON and returns all values
// stored under any of the given key names, regardless of nesting depth.
//
// Parameters:
//   - keys: One or more exact object key names to look up.
//
// Example:
//
//	emails := w.SearchJSONBodyByKey("email")
func (w *wrapper) SearchJSONBodyByKey(keys ...string) []fj.Context {
	return fj.SearchByKey(jsonpass(w.data), keys...)
}

// SearchJSONBodyByKeyPattern performs a full-tree wildcard scan of the body JSON and
// returns all values stored under object keys that match the given pattern.
//
// Parameters:
//   - keyPattern: A wildcard pattern applied to object key names.
//
// Example:
//
//	hits := w.SearchJSONBodyByKeyPattern("user*")
func (w *wrapper) SearchJSONBodyByKeyPattern(keyPattern string) []fj.Context {
	return fj.SearchByKeyPattern(jsonpass(w.data), keyPattern)
}

// JSONBodyContains reports whether the value at the given path inside the body contains
// the target substring (case-sensitive).
//
// Returns false when the path does not exist.
//
// Example:
//
//	w.JSONBodyContains("user.role", "admin")
func (w *wrapper) JSONBodyContains(path, target string) bool {
	return fj.Contains(jsonpass(w.data), path, target)
}

// JSONBodyContainsMatch reports whether the value at the given path inside the body
// matches the given wildcard pattern.
//
// Returns false when the path does not exist.
//
// Example:
//
//	w.JSONBodyContainsMatch("user.email", "*@example.com")
func (w *wrapper) JSONBodyContainsMatch(path, pattern string) bool {
	return fj.ContainsMatch(jsonpass(w.data), path, pattern)
}

// FindJSONBodyPath returns the first dot-notation path in the body at which a scalar
// value equals the given string (exact, case-sensitive match).
//
// Returns "" when no leaf matches.
//
// Example:
//
//	path := w.FindJSONBodyPath("alice@example.com")
func (w *wrapper) FindJSONBodyPath(value string) string {
	return fj.FindPath(jsonpass(w.data), value)
}

// FindJSONBodyPaths returns all dot-notation paths in the body at which a scalar value
// equals the given string.
//
// Example:
//
//	paths := w.FindJSONBodyPaths("active")
func (w *wrapper) FindJSONBodyPaths(value string) []string {
	return fj.FindPaths(jsonpass(w.data), value)
}

// FindJSONBodyPathMatch returns the first dot-notation path in the body at which a scalar
// value matches the given wildcard pattern.
//
// Example:
//
//	path := w.FindJSONBodyPathMatch("alice*")
func (w *wrapper) FindJSONBodyPathMatch(pattern string) string {
	return fj.FindPathMatch(jsonpass(w.data), pattern)
}

// FindJSONBodyPathsMatch returns all dot-notation paths in the body at which a scalar
// value matches the given wildcard pattern.
//
// Example:
//
//	paths := w.FindJSONBodyPathsMatch("err*")
func (w *wrapper) FindJSONBodyPathsMatch(pattern string) []string {
	return fj.FindPathsMatch(jsonpass(w.data), pattern)
}

// CountJSONBody returns the number of elements at the given path in the body.
// For an array result it returns the array length; for a scalar it returns 1;
// for a missing path it returns 0.
//
// Example:
//
//	n := w.CountJSONBody("items")
func (w *wrapper) CountJSONBody(path string) int {
	return fj.Count(jsonpass(w.data), path)
}

// SumJSONBody returns the sum of all numeric values at the given path in the body.
// Non-numeric elements are ignored. Returns 0 when no numbers are found.
//
// Example:
//
//	total := w.SumJSONBody("items.#.price")
func (w *wrapper) SumJSONBody(path string) float64 {
	return fj.Sum(jsonpass(w.data), path)
}

// MinJSONBody returns the minimum numeric value at the given path in the body.
// Returns (0, false) when no numeric values are found.
//
// Example:
//
//	v, ok := w.MinJSONBody("scores")
func (w *wrapper) MinJSONBody(path string) (float64, bool) {
	return fj.Min(jsonpass(w.data), path)
}

// MaxJSONBody returns the maximum numeric value at the given path in the body.
// Returns (0, false) when no numeric values are found.
//
// Example:
//
//	v, ok := w.MaxJSONBody("scores")
func (w *wrapper) MaxJSONBody(path string) (float64, bool) {
	return fj.Max(jsonpass(w.data), path)
}

// AvgJSONBody returns the arithmetic mean of all numeric values at the given path in the
// body. Returns (0, false) when no numeric values are found.
//
// Example:
//
//	avg, ok := w.AvgJSONBody("ratings")
func (w *wrapper) AvgJSONBody(path string) (float64, bool) {
	return fj.Avg(jsonpass(w.data), path)
}

// CollectJSONBodyFloat64 collects every value at the given path in the body that can be
// coerced to float64 (including string-encoded numbers). Non-numeric values are
// skipped.
//
// Example:
//
//	prices := w.CollectJSONBodyFloat64("items.#.price")
func (w *wrapper) CollectJSONBodyFloat64(path string) []float64 {
	return fj.CollectFloat64(jsonpass(w.data), path)
}

// FilterJSONBody evaluates the given path in the body, treats the result as an array,
// and returns only those elements for which fn returns true.
//
// Example:
//
//	active := w.FilterJSONBody("users", func(ctx fj.Context) bool {
//	    return ctx.Get("active").Bool()
//	})
func (w *wrapper) FilterJSONBody(path string, fn func(fj.Context) bool) []fj.Context {
	return fj.Filter(jsonpass(w.data), path, fn)
}

// FirstJSONBody evaluates the given path in the body and returns the first element for
// which fn returns true. Returns a zero-value fj.Context when not found.
//
// Example:
//
//	admin := w.FirstJSONBody("users", func(ctx fj.Context) bool {
//	    return ctx.Get("role").String() == "admin"
//	})
func (w *wrapper) FirstJSONBody(path string, fn func(fj.Context) bool) fj.Context {
	return fj.First(jsonpass(w.data), path, fn)
}

// DistinctJSONBody evaluates the given path in the body and returns a deduplicated slice
// of values using each element's string representation as the equality key.
// First-occurrence order is preserved.
//
// Example:
//
//	tags := w.DistinctJSONBody("tags")
func (w *wrapper) DistinctJSONBody(path string) []fj.Context {
	return fj.Distinct(jsonpass(w.data), path)
}

// PluckJSONBody evaluates the given path in the body (expected: array of objects) and
// returns a new object for each element containing only the specified fields.
//
// Example:
//
//	rows := w.PluckJSONBody("users", "id", "email")
func (w *wrapper) PluckJSONBody(path string, fields ...string) []fj.Context {
	return fj.Pluck(jsonpass(w.data), path, fields...)
}

// GroupByJSONBody groups the elements at the given path in the body by the string value
// of keyField, using conv.String for key normalization.
//
// Example:
//
//	byRole := w.GroupByJSONBody("users", "role")
func (w *wrapper) GroupByJSONBody(path, keyField string) map[string][]fj.Context {
	return fj.GroupBy(jsonpass(w.data), path, keyField)
}

// SortJSONBody sorts the elements at the given path in the body by the value of
// keyField. Numeric fields are compared as float64; all others fall back to string
// comparison.
//
// Parameters:
//   - path:      A fj path resolving to an array.
//   - keyField:  The field to sort by. Pass "" to sort scalar arrays.
//   - ascending: Sort direction.
//
// Example:
//
//	sorted := w.SortJSONBody("products", "price", true)
func (w *wrapper) SortJSONBody(path, keyField string, ascending bool) []fj.Context {
	return fj.SortBy(jsonpass(w.data), path, keyField, ascending)
}
