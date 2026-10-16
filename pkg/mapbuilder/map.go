package mapbuilder

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/polarixa/replify/pkg/strutil"
)

// MapBuilder is a fluent builder for map[string]any.
// It is not safe for concurrent use.
type MapBuilder struct {
	data map[string]any // The underlying map being built.
	err  error          // Any error encountered during the building process.
}

// New creates and returns a new instance of [MapBuilder].
func New() *MapBuilder {
	return &MapBuilder{
		data: make(map[string]any),
	}
}

// From creates a new [MapBuilder] initialized with a deep copy of the provided map.
// The source map is not mutated by subsequent builder operations.
func From(src map[string]any) *MapBuilder {
	b := New()
	if src != nil {
		b.data = deepCopy(src).(map[string]any)
	}
	return b
}

// Add adds a key-value pair to the map being built.
// If the key is empty, it sets an error and returns the builder.
func (b *MapBuilder) Add(key string, value any) *MapBuilder {
	if b.err != nil {
		return b
	}
	if strutil.IsEmpty(key) {
		b.err = errors.New("mapbuilder.Add: empty key")
		return b
	}
	b.data[key] = value
	return b
}

// AddAll adds all key-value pairs from the provided map to the map being built.
// If any key is empty, it sets an error and returns the builder.
func (b *MapBuilder) AddAll(values map[string]any) *MapBuilder {
	if b.err != nil {
		return b
	}
	for k := range values {
		if strutil.IsEmpty(k) {
			b.err = errors.New("mapbuilder.AddAll: empty key")
			return b
		}
	}
	maps.Copy(b.data, values)
	return b
}

// AddIfAbsent adds the key-value pair to the map being built only if the key is not already present.
// If the key is empty, it sets an error and returns the builder.
func (b *MapBuilder) AddIfAbsent(key string, value any) *MapBuilder {
	if b.err != nil {
		return b
	}
	if strutil.IsEmpty(key) {
		b.err = errors.New("mapbuilder.AddIfAbsent: empty key")
		return b
	}
	if _, exists := b.data[key]; !exists {
		b.data[key] = value
	}
	return b
}

// setPath sets a value at the specified dot-separated path within the map being built.
// If onlyIfAbsent is true, the value is only set if the path does not already exist.
// It returns the builder for chaining and sets an error if the path is invalid or conflicts with existing non-map values.
func (b *MapBuilder) setPath(op, path string, value any, onlyIfAbsent bool) *MapBuilder {
	if b.err != nil {
		return b
	}
	parts, err := parsePath(path)
	if err != nil {
		b.err = fmt.Errorf("mapbuilder.%s: %w", op, err)
		return b
	}

	// First pass: validate path without creating maps to avoid partial application
	current := b.data
	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]
		next, exists := current[key]
		if exists {
			nextMap, ok := next.(map[string]any)
			if !ok {
				b.err = fmt.Errorf("%s: path segment %q is not a map", op, key)
				return b
			}
			current = nextMap
		} else {
			// If a segment doesn't exist, all subsequent segments will be created
			// as new maps, so no scalar conflicts are possible deeper down.
			break
		}
	}

	// Second pass: create maps and set value
	current = b.data
	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]
		next, exists := current[key]
		if !exists {
			newMap := make(map[string]any)
			current[key] = newMap
			current = newMap
		} else {
			current = next.(map[string]any)
		}
	}

	lastKey := parts[len(parts)-1]
	if onlyIfAbsent {
		if _, exists := current[lastKey]; exists {
			return b
		}
	}
	current[lastKey] = value
	return b
}

// AddPath sets a value at the specified dot-separated path within the map being built.
// It overwrites any existing value at the path.
// It returns the builder for chaining and sets an error if the path is invalid or conflicts with existing non-map values.
func (b *MapBuilder) AddPath(path string, value any) *MapBuilder {
	return b.setPath("AddPath", path, value, false)
}

// AddPathIfAbsent sets a value at the specified dot-separated path within the map being built.
// It only sets the value if the path does not already exist.
// It returns the builder for chaining and sets an error if the path is invalid or conflicts with existing non-map values.
func (b *MapBuilder) AddPathIfAbsent(path string, value any) *MapBuilder {
	return b.setPath("AddPathIfAbsent", path, value, true)
}

// Append appends the given values to the slice at the specified key within the map being built.
// If the key does not exist, a new slice is created.
// It returns the builder for chaining and sets an error if the existing value at the key is not a []any.
func (b *MapBuilder) Append(key string, values ...any) *MapBuilder {
	if b.err != nil {
		return b
	}
	if strutil.IsEmpty(key) {
		b.err = errors.New("mapbuilder.Append: empty key")
		return b
	}
	if len(values) == 0 {
		return b
	}
	newSlice, err := appendToSlice(b.data[key], values)
	if err != nil {
		b.err = fmt.Errorf("mapbuilder.Append: %w", err)
		return b
	}
	b.data[key] = newSlice
	return b
}

// AppendPath appends the given values to the slice at the specified dot-separated path within the map being built.
// If the path does not exist, intermediate maps are created.
// It returns the builder for chaining and sets an error if the existing value at the path is not a []any.
func (b *MapBuilder) AppendPath(path string, values ...any) *MapBuilder {
	if b.err != nil {
		return b
	}
	if len(values) == 0 {
		return b
	}
	parts, err := parsePath(path)
	if err != nil {
		b.err = fmt.Errorf("mapbuilder.AppendPath: %w", err)
		return b
	}

	// First pass: validate path
	current := b.data
	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]
		next, exists := current[key]
		if exists {
			nextMap, ok := next.(map[string]any)
			if !ok {
				b.err = fmt.Errorf("mapbuilder.AppendPath: path segment %q is not a map", key)
				return b
			}
			current = nextMap
		} else {
			break
		}
	}

	// Second pass: create maps and append
	current = b.data
	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]
		next, exists := current[key]
		if !exists {
			newMap := make(map[string]any)
			current[key] = newMap
			current = newMap
		} else {
			current = next.(map[string]any)
		}
	}

	lastKey := parts[len(parts)-1]
	newSlice, err := appendToSlice(current[lastKey], values)
	if err != nil {
		b.err = fmt.Errorf("mapbuilder.AppendPath: %w", err)
		return b
	}
	current[lastKey] = newSlice
	return b
}

// Merge recursively merges the given map into the map being built.
// If a key exists in both maps and both values are maps, they are merged recursively.
// Otherwise, the value from the given map overwrites the value in the map being built.
// It returns the builder for chaining.
func (b *MapBuilder) Merge(values map[string]any) *MapBuilder {
	if b.err != nil {
		return b
	}
	if values == nil {
		return b
	}
	copySrc := deepCopy(values).(map[string]any)
	mergeMap(b.data, copySrc)
	return b
}

// MergePath recursively merges the given map into the map at the specified dot-separated path within the map being built.
// If a key exists in both maps and both values are maps, they are merged recursively.
// Otherwise, the value from the given map overwrites the value in the map at the path.
// Intermediate maps are created if the path does not exist.
// It returns the builder for chaining and sets an error if any path segment is not a map.
func (b *MapBuilder) MergePath(path string, values map[string]any) *MapBuilder {
	if b.err != nil {
		return b
	}
	if values == nil {
		return b
	}
	parts, err := parsePath(path)
	if err != nil {
		b.err = fmt.Errorf("mapbuilder.MergePath: %w", err)
		return b
	}

	// First pass: validate path (note: len(parts) because we merge INTO the final map)
	current := b.data
	for i := range parts {
		key := parts[i]
		next, exists := current[key]
		if exists {
			nextMap, ok := next.(map[string]any)
			if !ok {
				b.err = fmt.Errorf("mapbuilder.MergePath: path segment %q is not a map", key)
				return b
			}
			current = nextMap
		} else {
			break
		}
	}

	// Second pass: create maps and merge
	current = b.data
	for i := range parts {
		key := parts[i]
		next, exists := current[key]
		if !exists {
			newMap := make(map[string]any)
			current[key] = newMap
			current = newMap
		} else {
			current = next.(map[string]any)
		}
	}

	copySrc := deepCopy(values).(map[string]any)
	mergeMap(current, copySrc)
	return b
}

// Remove deletes the value associated with the given key from the map being built.
// It returns the builder for chaining.
func (b *MapBuilder) Remove(key string) *MapBuilder {
	if b.err != nil {
		return b
	}
	delete(b.data, key)
	return b
}

// RemovePath deletes the value at the specified dot-separated path from the map being built.
// If any intermediate maps become empty as a result, they are also removed.
// It returns the builder for chaining.
func (b *MapBuilder) RemovePath(path string) *MapBuilder {
	if b.err != nil {
		return b
	}
	parts, err := parsePath(path)
	if err != nil {
		b.err = fmt.Errorf("mapbuilder.RemovePath: %w", err)
		return b
	}

	current := b.data
	stack := make([]map[string]any, 0, len(parts))
	stackKeys := make([]string, 0, len(parts)-1)

	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]
		next, exists := current[key]
		if !exists {
			return b
		}
		nextMap, ok := next.(map[string]any)
		if !ok {
			return b
		}
		stack = append(stack, current)
		stackKeys = append(stackKeys, key)
		current = nextMap
	}

	lastKey := parts[len(parts)-1]
	delete(current, lastKey)

	for len(stack) > 0 && len(current) == 0 {
		parent := stack[len(stack)-1]
		parentKey := stackKeys[len(stackKeys)-1]
		delete(parent, parentKey)
		current = parent
		stack = stack[:len(stack)-1]
		stackKeys = stackKeys[:len(stackKeys)-1]
	}

	return b
}

// Get retrieves the value associated with the given key from the map being built.
// It returns the value and a boolean indicating whether the key exists.
func (b *MapBuilder) Get(key string) (any, bool) {
	if b.err != nil {
		return nil, false
	}
	val, exists := b.data[key]
	return val, exists
}

// GetPath retrieves the value at the specified dot-separated path from the map being built.
// It returns the value and a boolean indicating whether the path exists.
func (b *MapBuilder) GetPath(path string) (any, bool) {
	if b.err != nil {
		return nil, false
	}
	parts, err := parsePath(path)
	if err != nil {
		// Read operations do not mutate the error state
		return nil, false
	}
	current := b.data
	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]
		next, exists := current[key]
		if !exists {
			return nil, false
		}
		nextMap, ok := next.(map[string]any)
		if !ok {
			return nil, false
		}
		current = nextMap
	}
	lastKey := parts[len(parts)-1]
	val, exists := current[lastKey]
	return val, exists
}

// Has checks if a value exists for the given key in the map being built.
// It returns true if the key exists, false otherwise.
func (b *MapBuilder) Has(key string) bool {
	if b.err != nil {
		return false
	}
	_, exists := b.data[key]
	return exists
}

// HasPath checks if a value exists at the specified dot-separated path in the map being built.
// It returns true if the path exists, false otherwise.
func (b *MapBuilder) HasPath(path string) bool {
	_, exists := b.GetPath(path)
	return exists
}

// Build finalizes the map being built and returns a deep copy of it along with any error encountered during the building process.
// The returned map is independent of the builder's internal state, so further modifications to the builder do not affect the returned map.
// If an error occurred during the building process, the returned map will be nil.
func (b *MapBuilder) Build() (map[string]any, error) {
	if b.err != nil {
		return nil, b.err
	}
	m := deepCopy(b.data)
	return m.(map[string]any), nil
}

// MustBuild finalizes the map being built and returns a deep copy of it.
// If an error occurred during the building process, it panics.
//
// Consider using Build() if you want to handle errors explicitly instead of panicking.
func (b *MapBuilder) MustBuild() map[string]any {
	m, err := b.Build()
	if err != nil {
		panic(err)
	}
	return m
}

// Reset clears the builder's internal state, allowing it to be reused for building a new map.
// After calling Reset, the builder is in the same state as a newly created [MapBuilder].
func (b *MapBuilder) Reset() {
	b.data = make(map[string]any)
	b.err = nil
}

// IsError checks if the builder has encountered an error during the building process.
// It returns true if an error has occurred, false otherwise.
func (b *MapBuilder) IsError() bool {
	return b.err != nil
}

// Cause returns the underlying error that caused the builder to be in an error state.
// It returns nil if no error has occurred.
func (b *MapBuilder) Cause() error {
	if b.err == nil {
		// This ensures that calling Cause() without checking IsError() first does not return nil.
		return errors.New("")
	}
	return b.err
}

// Error returns the error message of the underlying cause.
// If no error has occurred, it returns an empty string.
func (b *MapBuilder) Error() string {
	return b.Cause().Error()
}

// deepCopy recursively copies map[string]any and []any.
// Other types (including typed maps/slices, pointers, and structs) are shallow-copied.
// Note: Cyclic map[string]any or []any structures are not supported and will cause a stack overflow.
func deepCopy(v any) any {
	switch val := v.(type) {
	case map[string]any:
		if val == nil {
			return nil
		}
		res := make(map[string]any, len(val))
		for k, v := range val {
			res[k] = deepCopy(v)
		}
		return res
	case []any:
		if val == nil {
			return nil
		}
		res := make([]any, len(val))
		for i, v := range val {
			res[i] = deepCopy(v)
		}
		return res
	default:
		return v
	}
}

// parsePath splits a dot-separated path into its individual segments.
// It returns an error if the path is empty or contains empty segments.
func parsePath(path string) ([]string, error) {
	if strutil.IsEmpty(path) {
		return nil, errors.New("empty path")
	}
	parts := strings.Split(path, ".")
	if slices.Contains(parts, "") {
		return nil, fmt.Errorf("invalid path %q: empty segment", path)
	}
	return parts, nil
}

// appendToSlice appends the given values to an existing slice.
// If the existing value is nil, a new slice is created.
// Returns an error if the existing value is not a []any.
func appendToSlice(existing any, values []any) ([]any, error) {
	if existing == nil {
		res := make([]any, 0, len(values))
		res = append(res, values...)
		return res, nil
	}

	exSlice, ok := existing.([]any)
	if !ok {
		return nil, errors.New("existing value is not a []any")
	}

	res := make([]any, len(exSlice), len(exSlice)+len(values))
	copy(res, exSlice)
	res = append(res, values...)
	return res, nil
}

// mergeMap recursively merges the source map into the destination map.
// If a key exists in both maps and both values are maps, they are merged recursively.
// Otherwise, the value from the source map overwrites the value in the destination map.
func mergeMap(dst, src map[string]any) {
	for k, v := range src {
		if vSrc, ok := v.(map[string]any); ok {
			if vDst, ok := dst[k].(map[string]any); ok {
				mergeMap(vDst, vSrc)
				continue
			}
		}
		dst[k] = v
	}
}
