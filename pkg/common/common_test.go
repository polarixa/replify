package common_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/polarixa/replify/pkg/common"
)

// ─── builtin.go ─────────────────────────────────────────────────────────────

// TestIsEmptyValue_Complex verifies that complex zero values are correctly
// recognised as empty (regression for missing complex64/complex128 case).
func TestIsEmptyValue_Complex(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"complex64 zero", complex64(0), true},
		{"complex128 zero", complex128(0), true},
		{"complex64 nonzero", complex64(1 + 2i), false},
		{"complex128 nonzero", complex128(0 + 1i), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := common.IsEmptyValue(reflect.ValueOf(tc.value))
			if got != tc.want {
				t.Errorf("IsEmptyValue(%v) = %v; want %v", tc.value, got, tc.want)
			}
		})
	}
}

// ─── reader.go ───────────────────────────────────────────────────────────────

// TestSlurpLines_PartialLastLine verifies that a final line without a trailing
// newline is preserved (regression for the silent data-loss bug).
func TestSlurpLines_PartialLastLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "newline terminated",
			input: "line1\nline2\n",
			want:  []string{"line1\n", "line2\n"},
		},
		{
			name:  "no trailing newline",
			input: "line1\nline2",
			want:  []string{"line1\n", "line2"},
		},
		{
			name:  "single line no newline",
			input: "hello",
			want:  []string{"hello"},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := common.SlurpLines(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("SlurpLines(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

// TestSlurpLine_PartialLastLine verifies that a final line without a trailing
// newline is preserved.
func TestSlurpLine_PartialLastLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "newline terminated",
			input: "line1\nline2\n",
			want:  "line1\nline2\n",
		},
		{
			name:  "no trailing newline",
			input: "line1\nline2",
			want:  "line1\nline2",
		},
		{
			name:  "single word no newline",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := common.SlurpLine(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("SlurpLine(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ─── generic.go ──────────────────────────────────────────────────────────────

// TestTransform_EmptySlice verifies that Transform does not panic and returns
// an empty result when called with an empty slice.
func TestTransform_EmptySlice(t *testing.T) {
	result := common.Transform([]int{}, func(v any) any { return v.(int) * 2 })
	if result == nil {
		t.Fatal("expected non-nil result for empty slice")
	}
	rv := reflect.ValueOf(result)
	if rv.Len() != 0 {
		t.Errorf("expected length 0, got %d", rv.Len())
	}
}

// TestTransform_NonEmptySlice verifies basic Transform behaviour on a non-empty slice.
func TestTransform_NonEmptySlice(t *testing.T) {
	result := common.Transform([]int{1, 2, 3}, func(v any) any { return v.(int) * 10 })
	got, ok := result.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", result)
	}
	want := []int{10, 20, 30}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Transform = %v; want %v", got, want)
	}
}

// TestTransform_Map verifies that Transform works on a non-empty map without
// panicking (regression for the v.Index(0) on map bug).
func TestTransform_Map(t *testing.T) {
	input := map[string]int{"a": 1}
	result := common.Transform(input, func(v any) any {
		switch val := v.(type) {
		case string:
			return strings.ToUpper(val)
		case int:
			return val * 2
		default:
			return v
		}
	})
	// Maps interleave keys and values; result type is []any.
	got, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any for map Transform, got %T", result)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 elements (key + value), got %d", len(got))
	}
}

// TestTransform_EmptyMap verifies that Transform on an empty map returns an
// empty result without panicking.
func TestTransform_EmptyMap(t *testing.T) {
	result := common.Transform(map[string]int{}, func(v any) any { return v })
	if result == nil {
		t.Fatal("expected non-nil result for empty map")
	}
}

// TestFilter_Array verifies that Filter does not panic when given a fixed-size
// array (regression for reflect.MakeSlice on array type).
func TestFilter_Array(t *testing.T) {
	arr := [5]int{1, 2, 3, 4, 5}
	result := common.Filter(arr, func(v any) bool { return v.(int)%2 == 0 })
	got, ok := result.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", result)
	}
	want := []int{2, 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Filter(array) = %v; want %v", got, want)
	}
}

// TestRotateLeft_EmptySlice verifies that RotateLeft does not panic when the
// collection is empty (regression for divide-by-zero).
func TestRotateLeft_EmptySlice(t *testing.T) {
	result := common.RotateLeft([]int{}, 3)
	got, ok := result.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", result)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

// TestRotateRight_EmptySlice verifies that RotateRight does not panic when the
// collection is empty.
func TestRotateRight_EmptySlice(t *testing.T) {
	result := common.RotateRight([]int{}, 3)
	got, ok := result.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", result)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

// TestRotateLeft_NonEmpty verifies basic RotateLeft correctness.
func TestRotateLeft_NonEmpty(t *testing.T) {
	result := common.RotateLeft([]int{1, 2, 3, 4, 5}, 2)
	got, ok := result.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", result)
	}
	want := []int{3, 4, 5, 1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("RotateLeft = %v; want %v", got, want)
	}
}

// TestRotateRight_NonEmpty verifies basic RotateRight correctness.
func TestRotateRight_NonEmpty(t *testing.T) {
	result := common.RotateRight([]int{1, 2, 3, 4, 5}, 2)
	got, ok := result.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", result)
	}
	want := []int{4, 5, 1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("RotateRight = %v; want %v", got, want)
	}
}

// TestUnique_Array verifies Unique does not panic on a fixed-size array input.
func TestUnique_Array(t *testing.T) {
	arr := [6]int{1, 2, 2, 3, 3, 4}
	result := common.Unique(arr)
	got, ok := result.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", result)
	}
	want := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Unique(array) = %v; want %v", got, want)
	}
}

// TestPartition_Array verifies Partition does not panic on a fixed-size array input.
func TestPartition_Array(t *testing.T) {
	arr := [6]int{1, 2, 3, 4, 5, 6}
	trueResult, falseResult := common.Partition(arr, func(v any) bool { return v.(int)%2 == 0 })
	evens, ok := trueResult.([]int)
	if !ok {
		t.Fatalf("expected []int for evens, got %T", trueResult)
	}
	odds, ok := falseResult.([]int)
	if !ok {
		t.Fatalf("expected []int for odds, got %T", falseResult)
	}
	if !reflect.DeepEqual(evens, []int{2, 4, 6}) {
		t.Errorf("evens = %v; want [2 4 6]", evens)
	}
	if !reflect.DeepEqual(odds, []int{1, 3, 5}) {
		t.Errorf("odds = %v; want [1 3 5]", odds)
	}
}
