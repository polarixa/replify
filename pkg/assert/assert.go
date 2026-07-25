package assert

import (
	"reflect"
	"testing"

	"github.com/polarixa/replify/pkg/ref"
)

// AssertEqual compares two objects for equality and reports a test failure if they are not equal.
//
// This function uses the `reflect.DeepEqual` function to check if `object1` and `object2`
// are deeply equal. If they are not, the function marks the test as failed and reports an error message.
//
// Parameters:
//   - `t`: The testing instance (from `*testing.T`) used to report failures.
//   - `object1`: The first object to compare.
//   - `object2`: The second object to compare.
//
// Example:
//
//	AssertEqual(t, actualValue, expectedValue)
func AssertEqual(t *testing.T, object1, object2 any) {
	if !reflect.DeepEqual(object1, object2) {
		t.Helper()
		t.Errorf("expected %v and %v to be equal, but they are not equal", object1, object2)
	}
}

// AssertNil checks if the given object is nil and reports a test failure if it is not.
//
// This function uses the `IsNil` helper function to determine if the object is nil.
// If the object is not nil, it marks the test as failed and reports an error.
//
// Parameters:
//   - `t`: The testing instance (from `*testing.T`) used to report failures.
//   - `object`: The object to check for nil.
//
// Example:
//
//	AssertNil(t, err)
func AssertNil(t *testing.T, object any) {
	if !ref.IsNil(&object) {
		t.Helper()
		t.Errorf("expected %v to be nil, but wasn't nil", object)
	}
}

// AssertNotNil checks if the given object is not nil and reports a test failure if it is nil.
//
// This function uses the `IsNil` helper function to determine if the object is nil.
// If the object is nil, it marks the test as failed and reports an error.
//
// Parameters:
//   - `t`: The testing instance (from `*testing.T`) used to report failures.
//   - `object`: The object to check for non-nil.
//
// Example:
//
//	AssertNotNil(t, result)
func AssertNotNil(t *testing.T, object any) {
	if ref.IsNil(&object) {
		t.Helper()
		t.Errorf("expected %v not to be nil, but was nil", object)
	}
}

// AssertTrue checks if the given boolean is true and reports a test failure if it is not.
//
// This function internally calls `AssertEqual` to compare the boolean value `b` with `true`.
// If the comparison fails, it marks the test as failed and reports an error.
//
// Parameters:
//   - `t`: The testing instance (from `*testing.T`) used to report failures.
//   - `b`: The boolean value to check.
//
// Example:
//
//	AssertTrue(t, isValid)
func AssertTrue(t *testing.T, b bool) {
	t.Helper()
	AssertEqual(t, b, true)
}

// AssertFalse checks if the given boolean is false and reports a test failure if it is not.
//
// This function internally calls `AssertEqual` to compare the boolean value `b` with `false`.
// If the comparison fails, it marks the test as failed and reports an error.
//
// Parameters:
//   - `t`: The testing instance (from `*testing.T`) used to report failures.
//   - `b`: The boolean value to check.
//
// Example:
//
//	AssertFalse(t, hasError)
func AssertFalse(t *testing.T, b bool) {
	t.Helper()
	AssertEqual(t, b, false)
}
