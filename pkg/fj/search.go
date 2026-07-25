package fj

import (
	"math"
	"sort"
	"strings"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/match"
)

// Search performs a full-tree scan of the JSON document and returns all leaf values
// whose string representation contains the given keyword (case-sensitive substring
// match). Both string values and non-string scalar values (numbers, booleans, null)
// are compared against the keyword.
//
// The traversal is depth-first; array elements and object values are each visited
// recursively. Compound values (objects and arrays) are never returned as matches —
// only scalar leaves are considered.
//
// Parameters:
//   - json:    A well-formed JSON string to scan.
//   - keyword: The substring to search for. An empty keyword matches every leaf.
//
// Returns:
//   - A slice of Context values whose string representation contains keyword.
//     Returns an empty (non-nil) slice when there are no matches.
//
// Example:
//
//	json := `{"users":[{"name":"Alice"},{"name":"Bob"},{"name":"Charlie"}]}`
//	results := fj.Search(json, "Ali")
//	// results[0].String() == "Alice"
func Search(json, keyword string) []Context {
	return scanLeaves(nil, Parse(json), keyword)
}

// SearchByKey performs a full-tree scan of the JSON document and returns all Context
// values that are stored under any of the given key names. The search is recursive —
// it descends into nested objects and arrays at every depth level.
//
// Key matching is exact and case-sensitive. Multiple key names may be supplied; any
// value whose immediate parent key matches at least one of them is included.
//
// Parameters:
//   - json: A well-formed JSON string to scan.
//   - keys: One or more object key names to look up. If no keys are provided the
//     function returns an empty slice.
//
// Returns:
//   - A slice of Context values stored under the given keys, in depth-first order.
//
// Example:
//
//	json := `{"a":{"title":"Go"},"b":{"title":"Rust"},"c":{"other":"x"}}`
//	results := fj.SearchByKey(json, "title")
//	// len(results) == 2, results[0].String() == "Go"
func SearchByKey(json string, keys ...string) []Context {
	if len(keys) == 0 {
		return []Context{}
	}
	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[k] = struct{}{}
	}
	return scanByKey(nil, Parse(json), keySet)
}

// Contains reports whether the result of querying json at path contains the given
// target substring (case-sensitive). If the path does not exist, Contains returns
// false.
//
// Parameters:
//   - json:   A well-formed JSON string.
//   - path:   A fj dot-notation path.
//   - target: The substring to look for within the string representation of the
//     value found at path.
//
// Returns:
//   - true if the value exists and its string representation contains target.
//
// Example:
//
//	json := `{"msg":"hello world"}`
//	fj.Contains(json, "msg", "world") // true
//	fj.Contains(json, "msg", "xyz")   // false
func Contains(json, path, target string) bool {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return false
	}
	return strings.Contains(ctx.String(), target)
}

// FindPath returns the first dot-notation path in the JSON document at which the
// given scalar value can be found (case-sensitive, exact string match against
// Context.String()). Object keys and array indices are joined by ".".
//
// FindPath only searches leaf (scalar) values. If no leaf matches, an empty string
// is returned.
//
// Parameters:
//   - json:  A well-formed JSON string.
//   - value: The scalar string value to locate.
//
// Returns:
//   - The dot-notation path of the first matching leaf, or "" when not found.
//
// Example:
//
//	json := `{"user":{"name":"Alice","age":30}}`
//	fj.FindPath(json, "Alice") // "user.name"
func FindPath(json, value string) string {
	path, _ := scanPath(Parse(json), value, "")
	return path
}

// FindPaths returns the dot-notation paths for every leaf in the JSON document
// whose string representation exactly equals value (case-sensitive).
//
// Parameters:
//   - json:  A well-formed JSON string.
//   - value: The scalar string value to locate.
//
// Returns:
//   - All matching paths in depth-first order. Returns an empty slice when there
//     are no matches.
//
// Example:
//
//	json := `{"a":"x","b":{"c":"x","d":"y"}}`
//	fj.FindPaths(json, "x") // ["a", "b.c"]
func FindPaths(json, value string) []string {
	return scanPaths(nil, Parse(json), value, "")
}

// Filter evaluates path against json, treats the result as an array, and returns
// only those elements for which fn returns true.
//
// If path resolves to a non-array value (including a single scalar), that single
// value is treated as a one-element collection. If the path does not exist, an
// empty slice is returned.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//   - fn:   A predicate applied to each element. Returning true keeps the element.
//
// Returns:
//   - A slice of matching Context values.
//
// Example:
//
//	json := `{"items":[1,2,3,4,5]}`
//	results := fj.Filter(json, "items", func(ctx fj.Context) bool {
//	    return ctx.Float64() > 2
//	})
//	// results holds Context values for 3, 4, 5
func Filter(json, path string, fn func(Context) bool) []Context {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return []Context{}
	}
	items := ctx.Array()
	out := make([]Context, 0, len(items))
	for _, item := range items {
		if fn(item) {
			out = append(out, item)
		}
	}
	return out
}

// First evaluates path against json and returns the first element for which fn
// returns true. If no element matches, a zero-value Context is returned (use
// Context.Exists() to check).
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//   - fn:   A predicate applied to each element.
//
// Returns:
//   - The first matching Context, or a zero-value Context when not found.
//
// Example:
//
//	json := `{"items":[1,2,3,4,5]}`
//	ctx := fj.First(json, "items", func(c fj.Context) bool {
//	    return c.Float64() > 3
//	})
//	ctx.Float64() // 4
func First(json, path string, fn func(Context) bool) Context {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return Context{}
	}
	for _, item := range ctx.Array() {
		if fn(item) {
			return item
		}
	}
	return Context{}
}

// Distinct evaluates path against json and returns a deduplicated slice of values
// using the string representation of each element as the equality key. Order is
// preserved (first occurrence wins).
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - A slice of unique Context values. Returns an empty slice when the path does
//     not exist.
//
// Example:
//
//	json := `{"tags":["go","json","go","fast","json"]}`
//	results := fj.Distinct(json, "tags")
//	// len(results) == 3: "go", "json", "fast"
func Distinct(json, path string) []Context {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return []Context{}
	}
	seen := make(map[string]struct{})
	var out []Context
	for _, item := range ctx.Array() {
		key := item.String()
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			out = append(out, item)
		}
	}
	if out == nil {
		return []Context{}
	}
	return out
}

// Pluck evaluates path against json (expecting a JSON array of objects), and for
// each element builds a new JSON object containing only the specified fields. Fields
// absent from an element are omitted from its output object.
//
// Parameters:
//   - json:   A well-formed JSON string.
//   - path:   A fj dot-notation path resolving to an array of objects.
//   - fields: The field names to extract from each object.
//
// Returns:
//   - A slice of Context values, one per element in the source array, each
//     wrapping the projected JSON object. Returns an empty slice when path does
//     not exist or no fields are provided.
//
// Example:
//
//	json := `{"users":[
//	    {"id":1,"name":"Alice","email":"a@example.com"},
//	    {"id":2,"name":"Bob","email":"b@example.com"}
//	]}`
//	results := fj.Pluck(json, "users", "id", "name")
//	// results[0].String() == `{"id":1,"name":"Alice"}`
//	// results[1].String() == `{"id":2,"name":"Bob"}`
func Pluck(json, path string, fields ...string) []Context {
	if len(fields) == 0 {
		return []Context{}
	}
	ctx := Get(json, path)
	if !ctx.Exists() {
		return []Context{}
	}
	items := ctx.Array()
	out := make([]Context, 0, len(items))
	for _, item := range items {
		if !item.IsObject() {
			continue
		}
		var b strings.Builder
		b.WriteByte('{')
		wrote := 0
		for _, f := range fields {
			v := item.Get(f)
			if !v.Exists() {
				continue
			}
			if wrote > 0 {
				b.WriteByte(',')
			}
			b.WriteString(`"` + f + `":`)
			raw := v.raw
			if raw == "" {
				raw = v.String()
			}
			b.WriteString(raw)
			wrote++
		}
		b.WriteByte('}')
		projected := Parse(b.String())
		out = append(out, projected)
	}
	return out
}

// SearchMatch performs a full-tree scan of the JSON document and returns all scalar
// leaf values whose string representation matches the given wildcard pattern.
//
// The pattern follows the same syntax as match.Match:
//   - '*' matches any sequence of characters (including empty).
//   - '?' matches exactly one character.
//   - Any other character is matched literally.
//
// Both string and non-string scalar values (numbers, booleans, null) are tested
// against the pattern using their string representation.
//
// Parameters:
//   - json:    A well-formed JSON string to scan.
//   - pattern: A wildcard pattern. An empty pattern matches only an empty string.
//
// Returns:
//   - A slice of Context values whose string representation matches pattern.
//     Returns an empty (non-nil) slice when there are no matches.
//
// Example:
//
//	json := `{"users":[{"name":"Alice"},{"name":"Bob"},{"name":"Albany"}]}`
//	results := fj.SearchMatch(json, "Al*")
//	// len(results) == 2: "Alice", "Albany"
func SearchMatch(json, pattern string) []Context {
	return scanLeavesMatch(nil, Parse(json), pattern)
}

// SearchByKeyPattern performs a full-tree scan of the JSON document and returns all
// Context values whose immediate parent object key matches the given wildcard pattern.
//
// Key matching uses match.Match, supporting '*' (any sequence) and '?' (one character)
// wildcards. The scan is recursive, descending into nested objects and arrays.
//
// Parameters:
//   - json:       A well-formed JSON string to scan.
//   - keyPattern: A wildcard pattern applied to object key names.
//
// Returns:
//   - A slice of Context values stored under matching keys, in depth-first order.
//
// Example:
//
//	json := `{"author":"Donovan","authority":"admin","title":"Go"}`
//	results := fj.SearchByKeyPattern(json, "auth*")
//	// len(results) == 2: "Donovan" (author) and "admin" (authority)
func SearchByKeyPattern(json, keyPattern string) []Context {
	return scanByKeyPattern(nil, Parse(json), keyPattern)
}

// ContainsMatch reports whether the value at path in json matches the given wildcard
// pattern. If the path does not exist, ContainsMatch returns false.
//
// The pattern follows the same syntax as match.Match:
//   - '*' matches any sequence of characters.
//   - '?' matches exactly one character.
//
// Parameters:
//   - json:    A well-formed JSON string.
//   - path:    A fj dot-notation path.
//   - pattern: A wildcard pattern applied to the string representation of the value.
//
// Returns:
//   - true if the value exists and its string representation matches pattern.
//
// Example:
//
//	json := `{"email":"alice@example.com"}`
//	fj.ContainsMatch(json, "email", "*@example.com") // true
//	fj.ContainsMatch(json, "email", "*@other.com")   // false
func ContainsMatch(json, path, pattern string) bool {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return false
	}
	return match.Match(ctx.String(), pattern)
}

// FindPathMatch returns the first dot-notation path in the JSON document at which a
// scalar value matches the given wildcard pattern. Object keys and array indices are
// joined with ".".
//
// Only leaf (scalar) values are tested. If no leaf matches, an empty string is
// returned.
//
// Parameters:
//   - json:         A well-formed JSON string.
//   - valuePattern: A wildcard pattern applied to the string representation of each
//     scalar leaf.
//
// Returns:
//   - The dot-notation path of the first matching leaf, or "" when not found.
//
// Example:
//
//	json := `{"users":[{"name":"Alice"},{"name":"Bob"}]}`
//	fj.FindPathMatch(json, "Ali*") // "users.0.name"
func FindPathMatch(json, valuePattern string) string {
	path, _ := scanPathMatch(Parse(json), valuePattern, "")
	return path
}

// FindPathsMatch returns the dot-notation paths for every scalar leaf in the JSON
// document whose string representation matches the given wildcard pattern.
//
// Parameters:
//   - json:         A well-formed JSON string.
//   - valuePattern: A wildcard pattern applied to the string representation of each
//     scalar leaf.
//
// Returns:
//   - All matching paths in depth-first order. Returns an empty slice when there
//     are no matches.
//
// Example:
//
//	json := `{"a":"Alice","b":{"c":"Albany","d":"Bob"}}`
//	fj.FindPathsMatch(json, "Al*") // ["a", "b.c"]
func FindPathsMatch(json, valuePattern string) []string {
	return scanPathsMatch(nil, Parse(json), valuePattern, "")
}

// CoerceTo converts the JSON value held in ctx into the Go variable pointed to by
// into, using the conv.Infer conversion engine. into must be a non-nil pointer to
// a supported type (bool, int*, uint*, float*, string, time.Time, slices, maps, or
// any struct with JSON-compatible fields).
//
// This function provides a bridge between fj's Context values and Go's type system,
// enabling ergonomic extraction of typed values without manual type-assertion chains.
//
// Parameters:
//   - ctx:  The Context whose value should be coerced.
//   - into: A non-nil pointer to the target variable.
//
// Returns:
//   - An error if the context has no value, or if the conversion fails.
//
// Example:
//
//	ctx := fj.Get(json, "user.age")
//	var age int
//	if err := fj.CoerceTo(ctx, &age); err == nil {
//	    fmt.Println(age) // 30
//	}
//
//	ctx = fj.Get(json, "user.active")
//	var active bool
//	_ = fj.CoerceTo(ctx, &active) // active == true
func CoerceTo(ctx Context, into any) error {
	if !ctx.Exists() {
		return conv.Infer(into, nil)
	}
	return conv.Infer(into, ctx.Value())
}

// CollectFloat64 evaluates path against json and returns a slice of float64 values
// for every element that can be coerced to a number by conv.Float64. This includes
// both JSON Number values and JSON strings that represent valid numbers (e.g.,
// "42", "3.14").
//
// Non-numeric elements for which conv.Float64 returns an error are silently skipped.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path resolving to an array or a single scalar.
//
// Returns:
//   - A slice of float64 values. Returns an empty (non-nil) slice when no elements
//     can be coerced to float64.
//
// Example:
//
//	json := `{"data":["10","20.5",30,null,"skip"]}`
//	vals := fj.CollectFloat64(json, "data")
//	// vals == []float64{10, 20.5, 30}
func CollectFloat64(json, path string) []float64 {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return []float64{}
	}
	var out []float64
	items := ctx.Array()
	if len(items) == 0 && ctx.Exists() {
		// scalar case
		if v, err := conv.Float64(ctx.Value()); err == nil {
			return []float64{v}
		}
		return []float64{}
	}
	for _, item := range items {
		if v, err := conv.Float64(item.Value()); err == nil {
			out = append(out, v)
		}
	}
	if out == nil {
		return []float64{}
	}
	return out
}

// GroupBy evaluates path against json, treats the result as an array of objects, and
// groups the elements by the string value of the specified keyField using conv.String
// for key normalization.
//
// Elements that do not contain keyField, or for which the key cannot be converted to
// a string, are placed under the empty-string group "".
//
// Parameters:
//   - json:     A well-formed JSON string.
//   - path:     A fj dot-notation path resolving to an array of objects.
//   - keyField: The object field whose value is used as the group key.
//
// Returns:
//   - A map from group-key string to a slice of Context values in that group.
//     Returns an empty map when path does not exist or the result is not an array.
//
// Example:
//
//	json := `{"books":[
//	    {"title":"Clean Code","genre":"tech"},
//	    {"title":"Dune","genre":"fiction"},
//	    {"title":"The Go Book","genre":"tech"}
//	]}`
//	groups := fj.GroupBy(json, "books", "genre")
//	// groups["tech"]    → 2 elements
//	// groups["fiction"] → 1 element
func GroupBy(json, path, keyField string) map[string][]Context {
	ctx := Get(json, path)
	out := make(map[string][]Context)
	if !ctx.Exists() || !ctx.IsArray() {
		return out
	}
	ctx.Foreach(func(_, item Context) bool {
		keyCtx := item.Get(keyField)
		var groupKey string
		if keyCtx.Exists() {
			if s, err := conv.String(keyCtx.Value()); err == nil {
				groupKey = s
			}
		}
		out[groupKey] = append(out[groupKey], item)
		return true
	})
	return out
}

// SortBy evaluates path against json, treats the result as an array, and returns a
// new slice sorted by the value of keyField on each element. Sorting uses
// conv.Float64 for numeric fields and falls back to the string representation for
// non-numeric or missing fields.
//
// When keyField is empty, elements are sorted by their own top-level string
// representation, which is useful for arrays of scalars.
//
// Parameters:
//   - json:      A well-formed JSON string.
//   - path:      A fj dot-notation path resolving to an array.
//   - keyField:  The object field to sort by. Pass "" to sort scalar arrays directly.
//   - ascending: If true, sort in ascending order; if false, descending.
//
// Returns:
//   - A new sorted slice of Context values. Returns an empty slice when path does
//     not exist.
//
// Example:
//
//	json := `{"items":[{"n":3},{"n":1},{"n":2}]}`
//	sorted := fj.SortBy(json, "items", "n", true)
//	// sorted[0].Get("n").Int64() == 1
//	// sorted[1].Get("n").Int64() == 2
//	// sorted[2].Get("n").Int64() == 3
func SortBy(json, path, keyField string, ascending bool) []Context {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return []Context{}
	}
	items := append([]Context(nil), ctx.Array()...)
	sort.SliceStable(items, func(i, j int) bool {
		vi := sortField(items[i], keyField)
		vj := sortField(items[j], keyField)
		less := sortOrdered(vi, vj)
		if ascending {
			return less
		}
		return !less
	})
	return items
}

// Count returns the number of elements returned by evaluating path against json.
// For a path that produces a JSON array the count equals the array length. For a
// path that produces a single scalar value, Count returns 1. For a missing or null
// result, Count returns 0.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The count of matching elements (≥ 0).
//
// Example:
//
//	json := `{"tags":["go","json","fast"]}`
//	fj.Count(json, "tags")   // 3
//	fj.Count(json, "tags.0") // 1
//	fj.Count(json, "missing")// 0
func Count(json, path string) int {
	ctx := Get(json, path)
	if !ctx.Exists() {
		return 0
	}
	if ctx.IsArray() {
		return len(ctx.Array())
	}
	return 1
}

// Sum returns the sum of all numeric values produced by evaluating path against
// json. Non-numeric results are silently ignored. Returns 0 when no numeric values
// are found.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path. The path may resolve to a JSON array of numbers
//     (e.g. "scores") or a single number (e.g. "scores.0").
//
// Returns:
//   - The sum as float64.
//
// Example:
//
//	json := `{"scores":[10,20,30]}`
//	fj.Sum(json, "scores") // 60.0
func Sum(json, path string) float64 {
	var total float64
	scanFloat64(json, path, func(n float64) { total += n })
	return total
}

// SumFloat32 returns the sum of all numeric values produced by evaluating path against
// json. Non-numeric results are silently ignored. Returns 0 when no numeric values
// are found.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path. The path may resolve to a JSON array of numbers
//     (e.g. "scores") or a single number (e.g. "scores.0").
//
// Returns:
//   - The sum as float32.
//
// Example:
//
//	json := `{"scores":[10,20,30]}`
//	fj.SumFloat32(json, "scores") // 60.0
func SumFloat32(json, path string) float32 {
	var total float32
	scanFloat32(json, path, func(n float32) { total += n })
	return total
}

// SumInt returns the sum of all numeric values produced by evaluating path against
// json. Non-numeric results are silently ignored. Returns 0 when no numeric values
// are found.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path. The path may resolve to a JSON array of numbers
//     (e.g. "scores") or a single number (e.g. "scores.0").
//
// Returns:
//   - The sum as int.
//
// Example:
//
//	json := `{"scores":[10,20,30]}`
//	fj.SumInt(json, "scores") // 60
func SumInt(json, path string) int {
	var total int
	scanInt(json, path, func(n int) { total += n })
	return total
}

// SumInt8 returns the sum of all numeric values produced by evaluating path against
// json. Non-numeric results are silently ignored. Returns 0 when no numeric values
// are found.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path. The path may resolve to a JSON array of numbers
//     (e.g. "scores") or a single number (e.g. "scores.0").
//
// Returns:
//   - The sum as int8.
//
// Example:
//
//	json := `{"scores":[10,20,30]}`
//	fj.SumInt8(json, "scores") // 60
func SumInt8(json, path string) int8 {
	var total int8
	scanInt8(json, path, func(n int8) { total += n })
	return total
}

// SumInt16 returns the sum of all numeric values produced by evaluating path against
// json. Non-numeric results are silently ignored. Returns 0 when no numeric values
// are found.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path. The path may resolve to a JSON array of numbers
//     (e.g. "scores") or a single number (e.g. "scores.0").
//
// Returns:
//   - The sum as int16.
//
// Example:
//
//	json := `{"scores":[10,20,30]}`
//	fj.SumInt16(json, "scores") // 60
func SumInt16(json, path string) int16 {
	var total int16
	scanInt16(json, path, func(n int16) { total += n })
	return total
}

// SumInt32 returns the sum of all numeric values produced by evaluating path against
// json. Non-numeric results are silently ignored. Returns 0 when no numeric values
// are found.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path. The path may resolve to a JSON array of numbers
//     (e.g. "scores") or a single number (e.g. "scores.0").
//
// Returns:
//   - The sum as int32.
//
// Example:
//
//	json := `{"scores":[10,20,30]}`
//	fj.SumInt32(json, "scores") // 60
func SumInt32(json, path string) int32 {
	var total int32
	scanInt32(json, path, func(n int32) { total += n })
	return total
}

// SumInt64 returns the sum of all numeric values produced by evaluating path against
// json. Non-numeric results are silently ignored. Returns 0 when no numeric values
// are found.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path. The path may resolve to a JSON array of numbers
//     (e.g. "scores") or a single number (e.g. "scores.0").
//
// Returns:
//   - The sum as int64.
//
// Example:
//
//	json := `{"scores":[10,20,30]}`
//	fj.SumInt64(json, "scores") // 60
func SumInt64(json, path string) int64 {
	var total int64
	scanInt64(json, path, func(n int64) { total += n })
	return total
}

// SumUint returns the sum of all numeric values produced by evaluating path against
// json. Non-numeric results are silently ignored. Returns 0 when no numeric values
// are found.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path. The path may resolve to a JSON array of numbers
//     (e.g. "scores") or a single number (e.g. "scores.0").
//
// Returns:
//   - The sum as uint.
//
// Example:
//
//	json := `{"scores":[10,20,30]}`
//	fj.SumUint(json, "scores") // 60
func SumUint(json, path string) uint {
	var total uint
	scanUint(json, path, func(n uint) { total += n })
	return total
}

// SumUint8 returns the sum of all numeric values produced by evaluating path against
// json. Non-numeric results are silently ignored. Returns 0 when no numeric values
// are found.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path. The path may resolve to a JSON array of numbers
//     (e.g. "scores") or a single number (e.g. "scores.0").
//
// Returns:
//   - The sum as uint8.
//
// Example:
//
//	json := `{"scores":[10,20,30]}`
//	fj.SumUint8(json, "scores") // 60
func SumUint8(json, path string) uint8 {
	var total uint8
	scanUint8(json, path, func(n uint8) { total += n })
	return total
}

// SumUint16 returns the sum of all numeric values produced by evaluating path against
// json. Non-numeric results are silently ignored. Returns 0 when no numeric values
// are found.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path. The path may resolve to a JSON array of numbers
//     (e.g. "scores") or a single number (e.g. "scores.0").
//
// Returns:
//   - The sum as uint16.
//
// Example:
//
//	json := `{"scores":[10,20,30]}`
//	fj.SumUint16(json, "scores") // 60
func SumUint16(json, path string) uint16 {
	var total uint16
	scanUint16(json, path, func(n uint16) { total += n })
	return total
}

// SumUint32 returns the sum of all numeric values produced by evaluating path against
// json. Non-numeric results are silently ignored. Returns 0 when no numeric values
// are found.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path. The path may resolve to a JSON array of numbers
//     (e.g. "scores") or a single number (e.g. "scores.0").
//
// Returns:
//   - The sum as uint32.
//
// Example:
//
//	json := `{"scores":[10,20,30]}`
//	fj.SumUint32(json, "scores") // 60
func SumUint32(json, path string) uint32 {
	var total uint32
	scanUint32(json, path, func(n uint32) { total += n })
	return total
}

// SumUint64 returns the sum of all numeric values produced by evaluating path against
// json. Non-numeric results are silently ignored. Returns 0 when no numeric values
// are found.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path. The path may resolve to a JSON array of numbers
//     (e.g. "scores") or a single number (e.g. "scores.0").
//
// Returns:
//   - The sum as uint64.
//
// Example:
//
//	json := `{"scores":[10,20,30]}`
//	fj.SumUint64(json, "scores") // 60
func SumUint64(json, path string) uint64 {
	var total uint64
	scanUint64(json, path, func(n uint64) { total += n })
	return total
}

// Min returns the minimum numeric value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The minimum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.Min(json, "scores") // 5.0, true
func Min(json, path string) (float64, bool) {
	min := math.MaxFloat64
	found := false
	scanFloat64(json, path, func(n float64) {
		if n < min {
			min = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return min, true
}

// MinFloat32 returns the minimum numeric value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The minimum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MinFloat32(json, "scores") // 5.0, true
func MinFloat32(json, path string) (float32, bool) {
	min := float32(math.MaxFloat32)
	found := false
	scanFloat32(json, path, func(n float32) {
		if n < min {
			min = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return min, true
}

// MinInt returns the minimum integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The minimum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MinInt(json, "scores") // 5, true
func MinInt(json, path string) (int, bool) {
	min := math.MaxInt
	found := false
	scanInt(json, path, func(n int) {
		if n < min {
			min = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return min, true
}

// MinInt8 returns the minimum integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The minimum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MinInt8(json, "scores") // 5, true
func MinInt8(json, path string) (int8, bool) {
	min := int8(math.MaxInt8)
	found := false
	scanInt8(json, path, func(n int8) {
		if n < min {
			min = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return min, true
}

// MinInt16 returns the minimum integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The minimum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MinInt16(json, "scores") // 5, true
func MinInt16(json, path string) (int16, bool) {
	min := int16(math.MaxInt16)
	found := false
	scanInt16(json, path, func(n int16) {
		if n < min {
			min = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return min, true
}

// MinInt32 returns the minimum integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The minimum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MinInt32(json, "scores") // 5, true
func MinInt32(json, path string) (int32, bool) {
	min := int32(math.MaxInt32)
	found := false
	scanInt32(json, path, func(n int32) {
		if n < min {
			min = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return min, true
}

// MinInt64 returns the minimum integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The minimum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MinInt64(json, "scores") // 5, true
func MinInt64(json, path string) (int64, bool) {
	min := int64(math.MaxInt64)
	found := false
	scanInt64(json, path, func(n int64) {
		if n < min {
			min = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return min, true
}

// MinUint returns the minimum unsigned integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The minimum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MinUint(json, "scores") // 5, true
func MinUint(json, path string) (uint, bool) {
	min := uint(math.MaxUint)
	found := false
	scanUint(json, path, func(n uint) {
		if n < min {
			min = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return min, true
}

// MinUint8 returns the minimum unsigned integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The minimum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MinUint8(json, "scores") // 5, true
func MinUint8(json, path string) (uint8, bool) {
	min := uint8(math.MaxUint8)
	found := false
	scanUint8(json, path, func(n uint8) {
		if n < min {
			min = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return min, true
}

// MinUint16 returns the minimum unsigned integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The minimum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MinUint16(json, "scores") // 5, true
func MinUint16(json, path string) (uint16, bool) {
	min := uint16(math.MaxUint16)
	found := false
	scanUint16(json, path, func(n uint16) {
		if n < min {
			min = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return min, true
}

// MinUint32 returns the minimum unsigned integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The minimum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MinUint32(json, "scores") // 5, true
func MinUint32(json, path string) (uint32, bool) {
	min := uint32(math.MaxUint32)
	found := false
	scanUint32(json, path, func(n uint32) {
		if n < min {
			min = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return min, true
}

// MinUint64 returns the minimum unsigned integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The minimum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MinUint64(json, "scores") // 5, true
func MinUint64(json, path string) (uint64, bool) {
	min := uint64(math.MaxUint64)
	found := false
	scanUint64(json, path, func(n uint64) {
		if n < min {
			min = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return min, true
}

// Max returns the maximum numeric value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The maximum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.Max(json, "scores") // 30.0, true
func Max(json, path string) (float64, bool) {
	max := -math.MaxFloat64
	found := false
	scanFloat64(json, path, func(n float64) {
		if n > max {
			max = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return max, true
}

// MaxFloat32 returns the maximum numeric value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The maximum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MaxFloat32(json, "scores") // 30.0, true
func MaxFloat32(json, path string) (float32, bool) {
	max := -float32(math.MaxFloat32)
	found := false
	scanFloat32(json, path, func(n float32) {
		if n > max {
			max = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return max, true
}

// MaxInt returns the maximum integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The maximum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MaxInt(json, "scores") // 30, true
func MaxInt(json, path string) (int, bool) {
	max := -math.MaxInt
	found := false
	scanInt(json, path, func(n int) {
		if n > max {
			max = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return max, true
}

// MaxInt8 returns the maximum integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The maximum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MaxInt8(json, "scores") // 30, true
func MaxInt8(json, path string) (int8, bool) {
	max := -int8(math.MaxInt8)
	found := false
	scanInt8(json, path, func(n int8) {
		if n > max {
			max = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return max, true
}

// MaxInt16 returns the maximum integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The maximum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MaxInt16(json, "scores") // 30, true
func MaxInt16(json, path string) (int16, bool) {
	max := -int16(math.MaxInt16)
	found := false
	scanInt16(json, path, func(n int16) {
		if n > max {
			max = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return max, true
}

// MaxInt32 returns the maximum integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The maximum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MaxInt32(json, "scores") // 30, true
func MaxInt32(json, path string) (int32, bool) {
	max := -int32(math.MaxInt32)
	found := false
	scanInt32(json, path, func(n int32) {
		if n > max {
			max = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return max, true
}

// MaxInt64 returns the maximum integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The maximum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MaxInt64(json, "scores") // 30, true
func MaxInt64(json, path string) (int64, bool) {
	max := -int64(math.MaxInt64)
	found := false
	scanInt64(json, path, func(n int64) {
		if n > max {
			max = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return max, true
}

// MaxUint returns the maximum unsigned integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The maximum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MaxUint(json, "scores") // 30, true
func MaxUint(json, path string) (uint, bool) {
	max := uint(0)
	found := false
	scanUint(json, path, func(n uint) {
		if n > max {
			max = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return max, true
}

// MaxUint8 returns the maximum unsigned integer value among all results produced by evaluating
// path against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The maximum value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,5,30]}`
//	v, ok := fj.MaxUint8(json, "scores") // 30, true
func MaxUint8(json, path string) (uint8, bool) {
	max := uint8(0)
	found := false
	scanUint8(json, path, func(n uint8) {
		if n > max {
			max = n
		}
		found = true
	})
	if !found {
		return 0, false
	}
	return max, true
}

// Avg returns the arithmetic mean of all numeric values produced by evaluating path
// against json. Non-numeric results are silently ignored.
//
// Parameters:
//   - json: A well-formed JSON string.
//   - path: A fj dot-notation path.
//
// Returns:
//   - The average value and true when at least one number is found.
//   - 0 and false when no numeric values are found.
//
// Example:
//
//	json := `{"scores":[10,20,30]}`
//	v, ok := fj.Avg(json, "scores") // 20.0, true
func Avg(json, path string) (float64, bool) {
	var total float64
	var n int
	scanFloat64(json, path, func(v float64) {
		total += v
		n++
	})
	if n == 0 {
		return 0, false
	}
	return total / float64(n), true
}
