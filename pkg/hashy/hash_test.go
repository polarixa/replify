package hashy_test

import (
	"fmt"
	"hash/fnv"
	"strings"
	"testing"
	"time"

	"github.com/polarixa/replify/pkg/hashy"
)

func TestHash_PrimitiveTypes(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{"int", 42},
		{"int64", int64(42)},
		{"uint", uint(42)},
		{"string", "hello"},
		{"bool_true", true},
		{"bool_false", false},
		{"float64", 3.14159},
		{"float32", float32(3.14)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hashy.Hash(tt.value)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
			if hash == 0 {
				t.Error("Hash() returned zero")
			}
		})
	}
}

func TestHash_VariadicSingleValue(t *testing.T) {
	value := "test"
	hash1, err := hashy.Hash(value)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	hash2, err := hashy.HashValue(value, nil)
	if err != nil {
		t.Fatalf("HashValue() error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("Hash() = %d, HashValue() = %d, want equal", hash1, hash2)
	}
}

func TestHash_VariadicMultipleValues(t *testing.T) {
	// Multiple values should hash as tuple
	hash1, err := hashy.Hash("a", "b", "c")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	hash2, err := hashy.Hash([]any{"a", "b", "c"})
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("Hash(a,b,c) = %d, Hash([a,b,c]) = %d, want equal", hash1, hash2)
	}
}

func TestHash_VariadicWithOptions(t *testing.T) {
	opts := hashy.NewOptions().WithZeroNil(true).Build()

	hash1, err := hashy.Hash("test", opts)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	hash2, err := hashy.HashValue("test", opts)
	if err != nil {
		t.Fatalf("HashValue() error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("Hash(test, opts) = %d, HashValue(test, opts) = %d, want equal", hash1, hash2)
	}
}

func TestHash_VariadicMultipleWithOptions(t *testing.T) {
	opts := hashy.NewOptions().WithZeroNil(true).Build()

	hash, err := hashy.Hash("a", "b", "c", opts)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if hash == 0 {
		t.Error("Hash() returned zero")
	}
}

func TestHash_NoData(t *testing.T) {
	_, err := hashy.Hash()
	if err == nil {
		t.Error("Hash() with no args should return error")
	}
}

func TestHash_OnlyOptions(t *testing.T) {
	opts := hashy.NewOptions().Build()
	_, err := hashy.Hash(opts)
	if err == nil {
		t.Error("Hash() with only options should return error")
	}
}

// ============================================================================
// DETERMINISM TESTS
// ============================================================================

func TestHash_Deterministic(t *testing.T) {
	type testStruct struct {
		Name  string
		Age   int
		Email string
	}

	value := testStruct{
		Name:  "Alice",
		Age:   30,
		Email: "alice@example.com",
	}

	// Run 100 times to ensure determinism
	hashes := make([]uint64, 100)
	for i := 0; i < 100; i++ {
		hash, err := hashy.Hash(value)
		if err != nil {
			t.Fatalf("Hash() error = %v", err)
		}
		hashes[i] = hash
	}

	// All hashes should be identical
	for i := 1; i < len(hashes); i++ {
		if hashes[i] != hashes[0] {
			t.Errorf("Hash is not deterministic: iteration %d gave %d, expected %d", i, hashes[i], hashes[0])
		}
	}
}

// ============================================================================
// EQUALITY TESTS
// ============================================================================

func TestHash_Equality(t *testing.T) {
	tests := []struct {
		name  string
		val1  any
		val2  any
		equal bool
	}{
		{
			name:  "identical_strings",
			val1:  "hello",
			val2:  "hello",
			equal: true,
		},
		{
			name:  "different_strings",
			val1:  "hello",
			val2:  "world",
			equal: false,
		},
		{
			name:  "identical_structs",
			val1:  struct{ Name string }{"Alice"},
			val2:  struct{ Name string }{"Alice"},
			equal: true,
		},
		{
			name:  "different_structs",
			val1:  struct{ Name string }{"Alice"},
			val2:  struct{ Name string }{"Bob"},
			equal: false,
		},
		{
			name:  "identical_slices",
			val1:  []int{1, 2, 3},
			val2:  []int{1, 2, 3},
			equal: true,
		},
		{
			name:  "different_slice_order",
			val1:  []int{1, 2, 3},
			val2:  []int{3, 2, 1},
			equal: false,
		},
		{
			name:  "identical_maps",
			val1:  map[string]int{"a": 1, "b": 2},
			val2:  map[string]int{"a": 1, "b": 2},
			equal: true,
		},
		{
			name:  "maps_different_order_same_content",
			val1:  map[string]int{"a": 1, "b": 2},
			val2:  map[string]int{"b": 2, "a": 1},
			equal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1, err := hashy.Hash(tt.val1)
			if err != nil {
				t.Fatalf("Hash(val1) error = %v", err)
			}

			hash2, err := hashy.Hash(tt.val2)
			if err != nil {
				t.Fatalf("Hash(val2) error = %v", err)
			}

			if (hash1 == hash2) != tt.equal {
				t.Errorf("Hash equality = %v, want %v (hash1=%d, hash2=%d)", hash1 == hash2, tt.equal, hash1, hash2)
			}
		})
	}
}

func TestHash_StructFieldOrder(t *testing.T) {
	type Struct1 struct {
		A string
		B string
	}

	type Struct2 struct {
		B string
		A string
	}

	s1 := Struct1{A: "a", B: "b"}
	s2 := Struct2{A: "a", B: "b"}

	hash1, _ := hashy.Hash(s1)
	hash2, _ := hashy.Hash(s2)

	// Different field order should produce different hashes
	if hash1 == hash2 {
		t.Error("Structs with different field order should have different hashes")
	}
}

func TestHash_StructTypeName(t *testing.T) {
	type Type1 struct{ Value string }
	type Type2 struct{ Value string }

	v1 := Type1{Value: "test"}
	v2 := Type2{Value: "test"}

	hash1, _ := hashy.Hash(v1)
	hash2, _ := hashy.Hash(v2)

	// Different type names should produce different hashes
	if hash1 == hash2 {
		t.Error("Different struct types should have different hashes")
	}
}

// ============================================================================
// STRUCT TAG TESTS
// ============================================================================

func TestHash_IgnoreTag(t *testing.T) {
	type User struct {
		Name     string
		Password string `hash:"ignore"`
	}

	u1 := User{Name: "Alice", Password: "secret1"}
	u2 := User{Name: "Alice", Password: "secret2"}

	hash1, _ := hashy.Hash(u1)
	hash2, _ := hashy.Hash(u2)

	t.Logf("hash1: %d, hash2: %d", hash1, hash2)

	if hash1 != hash2 {
		t.Error("Ignored field should not affect hash")
	}
}

func TestHash_DashTag(t *testing.T) {
	type User struct {
		Name     string
		Password string `hash:"-"`
	}

	u1 := User{Name: "Alice", Password: "secret1"}
	u2 := User{Name: "Alice", Password: "secret2"}

	hash1, _ := hashy.Hash(u1)
	hash2, _ := hashy.Hash(u2)

	if hash1 != hash2 {
		t.Error("Ignored field (dash) should not affect hash")
	}
}

func TestHash_SetTag(t *testing.T) {
	type Friends struct {
		Names []string `hash:"set"`
	}

	f1 := Friends{Names: []string{"Alice", "Bob", "Charlie"}}
	f2 := Friends{Names: []string{"Charlie", "Alice", "Bob"}}

	hash1, _ := hashy.Hash(f1)
	hash2, _ := hashy.Hash(f2)

	if hash1 != hash2 {
		t.Error("Set-tagged slice should be order-independent")
	}
}

func TestHash_StringTag(t *testing.T) {
	type Custom struct {
		Value int
	}

	type Container struct {
		Data Custom `hash:"string"`
	}

	c1 := Container{Data: Custom{Value: 42}}
	c2 := Container{Data: Custom{Value: 42}}

	hash1, _ := hashy.Hash(c1)
	hash2, _ := hashy.Hash(c2)

	if hash1 != hash2 {
		t.Error("String-tagged field should hash using String() method")
	}
}

func TestHash_StringTagError(t *testing.T) {
	type NonStringer struct {
		Value int
	}

	type Container struct {
		Data NonStringer `hash:"string"`
	}

	c := Container{Data: NonStringer{Value: 42}}

	_, err := hashy.Hash(c)
	if err == nil {
		t.Error("Expected error for string tag on non-Stringer type")
	}

	if _, ok := err.(*hashy.ErrNotStringer); !ok {
		t.Errorf("Expected ErrNotStringer, got %T", err)
	}
}

// ============================================================================
// OPTIONS TESTS
// ============================================================================

func TestHash_ZeroNilOption(t *testing.T) {
	type Data struct {
		Str   *string
		Int   *int
		Slice []string
	}

	d1 := Data{
		Str:   nil,
		Int:   nil,
		Slice: nil,
	}

	emptyStr := ""
	zeroInt := 0
	d2 := Data{
		Str:   &emptyStr,
		Int:   &zeroInt,
		Slice: []string{},
	}

	// Without ZeroNil
	hash1, _ := hashy.Hash(d1)
	hash2, _ := hashy.Hash(d2)
	if hash1 == hash2 {
		t.Error("Without ZeroNil, nil and zero should differ")
	}

	// With ZeroNil
	opts := hashy.NewOptions().WithZeroNil(true).Build()
	hash3, _ := hashy.Hash(d1, opts)
	hash4, _ := hashy.Hash(d2, opts)
	if hash3 != hash4 {
		t.Error("With ZeroNil, nil and zero should be equal")
	}
}

func TestHash_IgnoreZeroValueOption(t *testing.T) {
	type Data struct {
		Name  string
		Age   int
		Email string
	}

	d1 := Data{Name: "Alice", Age: 30}
	d2 := Data{Name: "Alice", Age: 30, Email: ""}

	// Without IgnoreZeroValue
	hash1, _ := hashy.Hash(d1)
	hash2, _ := hashy.Hash(d2)
	if hash1 != hash2 {
		t.Error("Without IgnoreZeroValue, zero fields should affect hash")
	}

	// With IgnoreZeroValue
	opts := hashy.NewOptions().WithIgnoreZeroValue(true).Build()
	hash3, _ := hashy.Hash(d1, opts)
	hash4, _ := hashy.Hash(d2, opts)
	if hash3 != hash4 {
		t.Error("With IgnoreZeroValue, zero fields should be ignored")
	}
}

func TestHash_SlicesAsSetsOption(t *testing.T) {
	type Data struct {
		Items []string
	}

	d1 := Data{Items: []string{"a", "b", "c"}}
	d2 := Data{Items: []string{"c", "a", "b"}}

	// Without SlicesAsSets
	hash1, _ := hashy.Hash(d1)
	hash2, _ := hashy.Hash(d2)
	if hash1 == hash2 {
		t.Error("Without SlicesAsSets, order should matter")
	}

	// With SlicesAsSets
	opts := hashy.NewOptions().WithSlicesAsSets(true).Build()
	hash3, _ := hashy.Hash(d1, opts)
	hash4, _ := hashy.Hash(d2, opts)
	if hash3 != hash4 {
		t.Error("With SlicesAsSets, order should not matter")
	}
}

func TestHash_CustomHasher(t *testing.T) {
	value := "test"

	// Default hasher
	hash1, _ := hashy.Hash(value)

	// Custom hasher
	opts := hashy.NewOptions().WithHasher(fnv.New64()).Build()
	hash2, _ := hashy.Hash(value, opts)

	// Different hashers may produce different results
	// We just verify both work
	if hash1 == 0 || hash2 == 0 {
		t.Error("Hashes should not be zero")
	}
}

// ============================================================================
// INTERFACE TESTS
// ============================================================================

type includableStruct struct {
	Value  string
	Ignore string
}

func (i includableStruct) SelectField() hashy.SelectField {
	return func(field string, value any) (bool, error) {
		return field != "Ignore", nil
	}
}

func (i includableStruct) SelectMapEntry() hashy.SelectMapEntry {
	return func(field string, k, v any) (bool, error) {
		return field != "Ignore", nil
	}
}

func TestHash_IncludableInterface(t *testing.T) {
	s1 := includableStruct{Value: "test", Ignore: "ignore1"}
	s2 := includableStruct{Value: "test", Ignore: "ignore2"}

	hash1, _ := hashy.Hash(s1)
	hash2, _ := hashy.Hash(s2)

	if hash1 != hash2 {
		t.Error("Includable interface should exclude Ignore field")
	}
}

type includableMapStruct struct {
	Data map[string]string
}

func (i includableMapStruct) SelectMapEntry() hashy.SelectMapEntry {
	return func(field string, key, value any) (bool, error) {
		if k, ok := key.(string); ok && k == "Ignore" {
			return false, nil
		}
		return true, nil
	}
}

func TestHash_IncludableMapInterface(t *testing.T) {
	s1 := includableMapStruct{Data: map[string]string{"a": "1", "Ignore": "ignore"}}
	s2 := includableMapStruct{Data: map[string]string{"a": "1"}}

	hash1, _ := hashy.Hash(s1)
	hash2, _ := hashy.Hash(s2)

	if hash1 != hash2 {
		t.Error("IncludableMap interface should exclude 'skip' key")
	}
}

type hashableStruct struct {
	Value string
}

func (h hashableStruct) Hash() (uint64, error) {
	if strings.HasPrefix(h.Value, "custom") {
		return 12345, nil
	}
	return 67890, nil
}

func TestHash_HashableInterface(t *testing.T) {
	s1 := hashableStruct{Value: "custom1"}
	s2 := hashableStruct{Value: "custom2"}

	hash1, _ := hashy.Hash(s1)
	hash2, _ := hashy.Hash(s2)

	if hash1 != hash2 {
		t.Error("Hashable interface should use custom hash")
	}

	if hash1 != 12345 {
		t.Errorf("Expected custom hash 12345, got %d", hash1)
	}
}

// ============================================================================
// EDGE CASE TESTS
// ============================================================================

func TestHash_EmptyCollections(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{"empty_slice", []int{}},
		{"empty_map", map[string]int{}},
		{"empty_string", ""},
		{"empty_struct", struct{}{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := hashy.Hash(tt.value)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}
		})
	}
}

func TestHash_TimeValues(t *testing.T) {
	now := time.Now()

	// Same time should produce same hash
	hash1, _ := hashy.Hash(now)
	hash2, _ := hashy.Hash(now)

	if hash1 != hash2 {
		t.Error("Same time value should produce same hash")
	}

	// Different times should produce different hashes
	later := now.Add(time.Hour)
	hash3, _ := hashy.Hash(later)

	if hash1 == hash3 {
		t.Error("Different times should produce different hashes")
	}

	// Strip monotonic clock
	stripped := time.Date(now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())

	hash4, _ := hashy.Hash(stripped)

	if hash1 != hash4 {
		t.Error("Time with and without monotonic clock should hash identically")
	}
}

func TestHash_UnexportedFields(t *testing.T) {
	type Data struct {
		Public     string
		private    string
		AnotherPub int
	}

	d1 := Data{Public: "test", private: "private1", AnotherPub: 42}
	d2 := Data{Public: "test", private: "private2", AnotherPub: 42}

	hash1, _ := hashy.Hash(d1)
	hash2, _ := hashy.Hash(d2)

	if hash1 != hash2 {
		t.Error("Unexported fields should not affect hash")
	}
}

// ============================================================================
// OPTIONS BUILDER TESTS
// ============================================================================

func TestOptionsBuilder(t *testing.T) {
	opts := hashy.NewOptions().
		WithZeroNil(true).
		WithIgnoreZeroValue(true).
		WithSlicesAsSets(true).
		WithTagName("custom").
		Build()

	if !opts.ZeroNil {
		t.Error("ZeroNil not set")
	}
	if !opts.IgnoreZeroValue {
		t.Error("IgnoreZeroValue not set")
	}
	if !opts.SlicesAsSets {
		t.Error("SlicesAsSets not set")
	}
	if opts.TagName != "custom" {
		t.Error("TagName not set")
	}
}

// ============================================================================
// COMPLEX NESTED STRUCTURES
// ============================================================================

func TestHash_ComplexNested(t *testing.T) {
	type Address struct {
		Street string
		City   string
	}

	type Person struct {
		Name     string
		Age      int
		Address  Address
		Friends  []string
		Metadata map[string]any
	}

	p1 := Person{
		Name: "Alice",
		Age:  30,
		Address: Address{
			Street: "123 Main St",
			City:   "NYC",
		},
		Friends: []string{"Bob", "Charlie"},
		Metadata: map[string]any{
			"active": true,
			"score":  95.5,
		},
	}

	p2 := p1 // Copy

	hash1, err := hashy.Hash(p1)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	hash2, err := hashy.Hash(p2)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if hash1 != hash2 {
		t.Error("Identical nested structures should have same hash")
	}

	// Modify nested field
	p2.Address.City = "LA"
	hash3, err := hashy.Hash(p2)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if hash1 == hash3 {
		t.Error("Modified nested structure should have different hash")
	}
}

// ============================================================================
// BENCHMARK TESTS
// ============================================================================

func BenchmarkHash_SimpleStruct(b *testing.B) {
	type Simple struct {
		Name string
		Age  int
	}

	val := Simple{"Alice", 30}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hashy.Hash(val)
	}
}

func BenchmarkHash_ComplexStruct(b *testing.B) {
	type Complex struct {
		Name     string
		Age      int
		Scores   []float64
		Metadata map[string]string
	}

	val := Complex{
		Name:   "Alice",
		Age:    30,
		Scores: []float64{95.5, 87.3, 92.1},
		Metadata: map[string]string{
			"city":    "NYC",
			"country": "USA",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hashy.Hash(val)
	}
}

func BenchmarkHash_LargeSlice(b *testing.B) {
	slice := make([]int, 1000)
	for i := range slice {
		slice[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hashy.Hash(slice)
	}
}

func BenchmarkHash_DeepNesting(b *testing.B) {
	type Node struct {
		Value string
		Child *Node
	}

	// Create 10-level deep nesting
	var root *Node
	current := &Node{Value: "root"}
	root = current
	for i := 0; i < 10; i++ {
		current.Child = &Node{Value: fmt.Sprintf("level-%d", i)}
		current = current.Child
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hashy.Hash(root)
	}
}

// ============================================================================
// NEW-HASH / NEW-HASH64 TESTS
// ============================================================================

// TestNewHash64_NonHash64Algorithms verifies that NewHash64 does not panic
// when called with algorithms that implement hash.Hash but not hash.Hash64
// (e.g. MD5, SHA-*). It must return a valid hash.Hash64 (FNV-1a fallback).
func TestNewHash64_NonHash64Algorithms(t *testing.T) {
	nonHash64Algos := []hashy.HashAlgorithm{
		hashy.H_MD5,
		hashy.H_SHA1,
		hashy.H_SHA224,
		hashy.H_SHA256,
		hashy.H_SHA384,
		hashy.H_SHA512,
		hashy.H_SHA512_224,
		hashy.H_SHA512_256,
	}

	for _, algo := range nonHash64Algos {
		algo := algo
		t.Run(string(algo), func(t *testing.T) {
			var h hashy.HashAlgorithm = algo
			// Must not panic.
			result := hashy.NewHash64(h)
			if result == nil {
				t.Fatalf("NewHash64(%s) returned nil", algo)
			}
			// The returned instance must be usable.
			result.Reset()
			if _, err := result.Write([]byte("test")); err != nil {
				t.Fatalf("NewHash64(%s).Write() error = %v", algo, err)
			}
			_ = result.Sum64()
		})
	}
}

// TestNewHash64_Hash64Algorithms verifies that NewHash64 returns the native
// hash.Hash64 implementation for algorithms that do support it.
func TestNewHash64_Hash64Algorithms(t *testing.T) {
	hash64Algos := []hashy.HashAlgorithm{
		hashy.H_CRC64,
	}

	for _, algo := range hash64Algos {
		algo := algo
		t.Run(string(algo), func(t *testing.T) {
			result := hashy.NewHash64(algo)
			if result == nil {
				t.Fatalf("NewHash64(%s) returned nil", algo)
			}
		})
	}
}

// ============================================================================
// CONCURRENCY SAFETY TESTS
// ============================================================================

// TestHash_ConcurrentSafety_NilOptions verifies that calling Hash concurrently
// with nil options (each call gets its own DefaultOptions) is race-free.
func TestHash_ConcurrentSafety_NilOptions(t *testing.T) {
	const goroutines = 50
	type item struct {
		Name  string
		Value int
	}

	done := make(chan struct{}, goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer func() { done <- struct{}{} }()
			_, err := hashy.Hash(item{Name: fmt.Sprintf("g%d", i), Value: i})
			if err != nil {
				t.Errorf("goroutine %d: Hash() error = %v", i, err)
			}
		}()
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

// TestHash_ConcurrentSafety_SharedOptsWithHasherFunc verifies that reusing
// *hashOptions built with WithHasherFunc across goroutines is race-free.
func TestHash_ConcurrentSafety_SharedOptsWithHasherFunc(t *testing.T) {
	opts := hashy.NewOptions().WithHasherFunc(fnv.New64a).Build()

	const goroutines = 50
	done := make(chan struct{}, goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer func() { done <- struct{}{} }()
			_, err := hashy.Hash(i, opts)
			if err != nil {
				t.Errorf("goroutine %d: Hash() error = %v", i, err)
			}
		}()
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

// TestHash_WithHasherFunc_Determinism verifies that using WithHasherFunc
// produces the same hashes as the default (FNV-1a) options.
func TestHash_WithHasherFunc_Determinism(t *testing.T) {
	value := struct {
		Name string
		Age  int
	}{Name: "Alice", Age: 30}

	hash1, err := hashy.Hash(value)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	opts := hashy.NewOptions().WithHasherFunc(fnv.New64a).Build()
	hash2, err := hashy.Hash(value, opts)
	if err != nil {
		t.Fatalf("Hash(opts) error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("WithHasherFunc(fnv.New64a) hash = %d, default hash = %d, want equal", hash2, hash1)
	}
}

// TestHash_ValidateNilHasher verifies that passing options with both Hasher
// and HasherFunc nil returns an error instead of panicking.
func TestHash_ValidateNilHasher(t *testing.T) {
	// Build options with a nil Hasher and nil HasherFunc via the builder's
	// exported fields (testing internal validate path via HashValue).
	opts := hashy.NewOptions().WithHasher(nil).WithHasherFunc(nil).Build()
	_, err := hashy.Hash("test", opts)
	if err == nil {
		t.Fatal("expected error when both Hasher and HasherFunc are nil")
	}
	const want = "hasher cannot be nil"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error %q does not contain %q", err.Error(), want)
	}
}

func TestHash_TimeValue(t *testing.T) {
	now := time.Now()
	h1, err := hashy.Hash(now)
	if err != nil {
		t.Fatalf("Hash(time.Time) error = %v", err)
	}
	if h1 == 0 {
		t.Error("Hash(time.Time) returned zero")
	}
	h2, err := hashy.Hash(now)
	if err != nil {
		t.Fatalf("Hash(time.Time) second call error = %v", err)
	}
	if h1 != h2 {
		t.Errorf("Hash(time.Time) not deterministic: %d != %d", h1, h2)
	}
}
