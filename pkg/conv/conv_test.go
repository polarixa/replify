package conv_test

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/polarixa/replify/pkg/conv"
)

// ─── ConvError / error wrapping ──────────────────────────────────────────────

// TestConvError_Unwrap verifies that ConvError.Unwrap exposes the wrapped cause
// error so that errors.Is / errors.As can traverse the chain.
func TestConvError_Unwrap(t *testing.T) {
	sentinel := errors.New("sentinel cause")
	wrapped := fmt.Errorf("outer: %w", &conv.ConvError{
		From:    "bad",
		To:      "int",
		Message: "wrap test",
		Cause:   sentinel,
	})

	if !errors.Is(wrapped, sentinel) {
		t.Errorf("errors.Is should find sentinel through ConvError.Unwrap chain")
	}
}

// TestIsConvError_Chain verifies that IsConvError traverses wrapped errors.
func TestIsConvError_Chain(t *testing.T) {
	inner := &conv.ConvError{From: "x", To: "int"}
	outer := fmt.Errorf("context: %w", inner)

	if !conv.IsConvError(outer) {
		t.Errorf("IsConvError should return true for a wrapped *ConvError")
	}
	if conv.IsConvError(errors.New("plain error")) {
		t.Errorf("IsConvError should return false for a non-ConvError")
	}
}

// TestAsConvError_Chain verifies that AsConvError unwraps correctly.
func TestAsConvError_Chain(t *testing.T) {
	inner := &conv.ConvError{From: "y", To: "bool", Message: "test msg"}
	outer := fmt.Errorf("layer: %w", inner)

	ce, ok := conv.AsConvError(outer)
	if !ok {
		t.Fatal("AsConvError should succeed for a wrapped *ConvError")
	}
	if ce.To != "bool" {
		t.Errorf("AsConvError: got To=%q, want %q", ce.To, "bool")
	}
}

// ─── strictMode – integer overflow ───────────────────────────────────────────

// TestStrictMode_Int8_Overflow verifies that Int8 returns an error in strict
// mode when the input overflows int8, and saturates in lenient mode.
func TestStrictMode_Int8_Overflow(t *testing.T) {
	strict := conv.NewConverter().WithStrictMode(true)
	lenient := conv.NewConverter().WithStrictMode(false)

	tests := []struct {
		name      string
		input     any
		wantErr   bool
		wantValue int8
	}{
		{"in-range positive", int64(100), false, 100},
		{"in-range negative", int64(-100), false, -100},
		{"overflow positive", int64(200), true, 0},
		{"overflow negative", int64(-200), true, 0},
		{"max boundary", int64(math.MaxInt8), false, math.MaxInt8},
		{"min boundary", int64(math.MinInt8), false, math.MinInt8},
	}

	for _, tc := range tests {
		t.Run("strict/"+tc.name, func(t *testing.T) {
			got, err := strict.Int8(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for %v, got nil (value=%d)", tc.input, got)
				}
				if !conv.IsConvError(err) {
					t.Errorf("expected *ConvError, got %T: %v", err, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if got != tc.wantValue {
					t.Errorf("got %d, want %d", got, tc.wantValue)
				}
			}
		})
	}

	t.Run("lenient saturates positive overflow", func(t *testing.T) {
		got, err := lenient.Int8(int64(200))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != math.MaxInt8 {
			t.Errorf("got %d, want %d (saturated)", got, math.MaxInt8)
		}
	})

	t.Run("lenient saturates negative overflow", func(t *testing.T) {
		got, err := lenient.Int8(int64(-200))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != math.MinInt8 {
			t.Errorf("got %d, want %d (saturated)", got, math.MinInt8)
		}
	})
}

// TestStrictMode_Int16_Overflow exercises the int16 boundary.
func TestStrictMode_Int16_Overflow(t *testing.T) {
	strict := conv.NewConverter().WithStrictMode(true)
	lenient := conv.NewConverter()

	if _, err := strict.Int16(int64(math.MaxInt16 + 1)); err == nil {
		t.Error("expected overflow error from strict Int16")
	}
	got, err := lenient.Int16(int64(math.MaxInt16 + 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != math.MaxInt16 {
		t.Errorf("got %d, want %d (saturated)", got, math.MaxInt16)
	}
}

// TestStrictMode_Int32_Overflow exercises the int32 boundary.
func TestStrictMode_Int32_Overflow(t *testing.T) {
	strict := conv.NewConverter().WithStrictMode(true)
	lenient := conv.NewConverter()

	if _, err := strict.Int32(int64(math.MaxInt32 + 1)); err == nil {
		t.Error("expected overflow error from strict Int32")
	}
	got, err := lenient.Int32(int64(math.MaxInt32 + 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != math.MaxInt32 {
		t.Errorf("got %d, want %d (saturated)", got, math.MaxInt32)
	}
}

// ─── strictMode – float-to-int truncation ────────────────────────────────────

// TestStrictMode_Int64_FractionalFloat verifies that converting a float with a
// fractional part to int64 errors in strict mode and truncates in lenient mode.
func TestStrictMode_Int64_FractionalFloat(t *testing.T) {
	strict := conv.NewConverter().WithStrictMode(true)
	lenient := conv.NewConverter()

	// strict must error on fractional float
	if _, err := strict.Int64(float64(3.7)); err == nil {
		t.Error("expected error converting 3.7 to int64 in strict mode")
	}
	if _, err := strict.Int64(float32(1.5)); err == nil {
		t.Error("expected error converting float32(1.5) to int64 in strict mode")
	}
	if _, err := strict.Int64("2.9"); err == nil {
		t.Error("expected error converting \"2.9\" to int64 in strict mode")
	}

	// strict must NOT error on whole-number floats
	if v, err := strict.Int64(float64(42.0)); err != nil || v != 42 {
		t.Errorf("unexpected error on whole float: err=%v, v=%d", err, v)
	}

	// lenient must truncate
	if v, err := lenient.Int64(float64(3.7)); err != nil || v != 3 {
		t.Errorf("lenient Int64(3.7) = %d, %v; want 3, nil", v, err)
	}
}

// ─── strictMode – uint overflow ──────────────────────────────────────────────

// TestStrictMode_Uint8_Overflow verifies that Uint8 errors on overflow in strict mode.
func TestStrictMode_Uint8_Overflow(t *testing.T) {
	strict := conv.NewConverter().WithStrictMode(true)
	lenient := conv.NewConverter()

	if _, err := strict.Uint8(uint64(256)); err == nil {
		t.Error("expected overflow error from strict Uint8(256)")
	}
	got, err := lenient.Uint8(uint64(256))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != math.MaxUint8 {
		t.Errorf("got %d, want %d (saturated)", got, math.MaxUint8)
	}
}

// TestStrictMode_Uint16_Overflow verifies that Uint16 errors on overflow in strict mode.
func TestStrictMode_Uint16_Overflow(t *testing.T) {
	strict := conv.NewConverter().WithStrictMode(true)

	if _, err := strict.Uint16(uint64(math.MaxUint16 + 1)); err == nil {
		t.Error("expected overflow error from strict Uint16")
	}
}

// TestStrictMode_Uint32_Overflow verifies that Uint32 errors on overflow in strict mode.
func TestStrictMode_Uint32_Overflow(t *testing.T) {
	strict := conv.NewConverter().WithStrictMode(true)

	if _, err := strict.Uint32(uint64(math.MaxUint32 + 1)); err == nil {
		t.Error("expected overflow error from strict Uint32")
	}
}

// ─── strictMode – float32 overflow ───────────────────────────────────────────

// TestStrictMode_Float32_Overflow verifies that Float32 errors in strict mode
// when the float64 value exceeds the float32 range.
func TestStrictMode_Float32_Overflow(t *testing.T) {
	strict := conv.NewConverter().WithStrictMode(true)
	lenient := conv.NewConverter()

	big := math.MaxFloat32 * 2 // clearly out of float32 range

	if _, err := strict.Float32(big); err == nil {
		t.Errorf("expected overflow error converting %v to float32 in strict mode", big)
	}

	got, err := lenient.Float32(big)
	if err != nil {
		t.Fatalf("unexpected error in lenient mode: %v", err)
	}
	if got != math.MaxFloat32 {
		t.Errorf("got %v, want %v (saturated)", got, math.MaxFloat32)
	}
}

// ─── time.Time – uint64 overflow protection ──────────────────────────────────

// TestTime_Uint64_Overflow verifies that values larger than math.MaxInt64 do
// not wrap around to a negative Unix timestamp.
func TestTime_Uint64_Overflow(t *testing.T) {
	c := conv.NewConverter()

	huge := uint64(math.MaxInt64) + 1 // would overflow int64 silently before the fix
	got, err := c.Time(huge)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Must NOT be negative (which would indicate silent int64 wrap-around).
	if got.Unix() < 0 {
		t.Errorf("Time(%d) = %v; expected non-negative Unix timestamp (overflow clamped)", huge, got)
	}
	// The clamped value should equal time.Unix(math.MaxInt64, 0).
	wantSec := int64(math.MaxInt64)
	if got.Unix() != wantSec {
		t.Errorf("got Unix=%d, want %d (clamped to MaxInt64)", got.Unix(), wantSec)
	}
}

// TestTime_Uint64_Normal verifies that normal (in-range) uint64 values still
// produce correct results after the overflow guard.
func TestTime_Uint64_Normal(t *testing.T) {
	c := conv.NewConverter()
	input := uint64(1_700_000_000) // a realistic epoch timestamp
	got, err := c.Time(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Unix() != int64(input) {
		t.Errorf("got Unix=%d, want %d", got.Unix(), input)
	}
}

// ─── Duration – float64 overflow protection ──────────────────────────────────

// TestDuration_Float64_Overflow verifies that very large float64 values do not
// silently wrap around when converted to time.Duration.
func TestDuration_Float64_Overflow(t *testing.T) {
	c := conv.NewConverter()

	// A value in seconds that clearly exceeds the int64 ns range (~292 years).
	hugeSeconds := 1e20

	got, err := c.Duration(hugeSeconds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The result must be clamped to the maximum duration, not a negative wrap.
	if got < 0 {
		t.Errorf("Duration(%v) = %v; expected positive (clamped) value", hugeSeconds, got)
	}
	if got != time.Duration(math.MaxInt64) {
		t.Errorf("got %v, want %v (clamped to MaxInt64 ns)", got, time.Duration(math.MaxInt64))
	}
}

// TestDuration_Float64_NegativeOverflow verifies clamping for very negative values.
func TestDuration_Float64_NegativeOverflow(t *testing.T) {
	c := conv.NewConverter()

	got, err := c.Duration(-1e20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got > 0 {
		t.Errorf("Duration(-1e20) = %v; expected negative (clamped) value", got)
	}
	if got != time.Duration(math.MinInt64) {
		t.Errorf("got %v, want %v (clamped to MinInt64 ns)", got, time.Duration(math.MinInt64))
	}
}

// TestDuration_Float64_Normal verifies that normal float64 second values
// produce correct durations after the overflow guard.
func TestDuration_Float64_Normal(t *testing.T) {
	c := conv.NewConverter()

	got, err := c.Duration(float64(1.5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := 1500 * time.Millisecond
	if got != want {
		t.Errorf("Duration(1.5) = %v, want %v", got, want)
	}
}

// ─── Concurrent access to dateFormats ────────────────────────────────────────

// TestConverter_ConcurrentDateFormats verifies that concurrent calls to
// WithDateFormats and Time do not race.  Run with -race.
func TestConverter_ConcurrentDateFormats(t *testing.T) {
	c := conv.NewConverter()
	var wg sync.WaitGroup

	const goroutines = 20
	wg.Add(goroutines * 2)

	for i := 0; i < goroutines; i++ {
		// Writer goroutine
		go func() {
			defer wg.Done()
			c.WithDateFormats(time.RFC3339, "2006-01-02")
		}()
		// Reader goroutine
		go func() {
			defer wg.Done()
			_, _ = c.Time("2024-06-15")
		}()
	}

	wg.Wait()
}

// ─── Regression: To[T] generic helper uses correct types ─────────────────────

// TestTo_Generic verifies that the To[T] generic helper returns correct values
// for the built-in supported types.
func TestTo_Generic(t *testing.T) {
	tests := []struct {
		name  string
		input any
		fn    func(any) (any, error)
		want  any
	}{
		{"int", "42", func(v any) (any, error) { return conv.To[int](v) }, 42},
		{"int8", int64(100), func(v any) (any, error) { return conv.To[int8](v) }, int8(100)},
		{"int16", int64(1000), func(v any) (any, error) { return conv.To[int16](v) }, int16(1000)},
		{"int32", int64(100000), func(v any) (any, error) { return conv.To[int32](v) }, int32(100000)},
		{"int64", "9000000000", func(v any) (any, error) { return conv.To[int64](v) }, int64(9000000000)},
		{"uint", "100", func(v any) (any, error) { return conv.To[uint](v) }, uint(100)},
		{"float64", "3.14", func(v any) (any, error) { return conv.To[float64](v) }, 3.14},
		{"bool", "true", func(v any) (any, error) { return conv.To[bool](v) }, true},
		{"string", 42, func(v any) (any, error) { return conv.To[string](v) }, "42"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.fn(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %v (%T), want %v (%T)", got, got, tc.want, tc.want)
			}
		})
	}
}

// ─── IsConvError nil-safety ───────────────────────────────────────────────────

// TestIsConvError_Nil ensures IsConvError does not panic on a nil error.
func TestIsConvError_Nil(t *testing.T) {
	if conv.IsConvError(nil) {
		t.Error("IsConvError(nil) should return false")
	}
}

// TestAsConvError_Nil ensures AsConvError does not panic on a nil error.
func TestAsConvError_Nil(t *testing.T) {
	ce, ok := conv.AsConvError(nil)
	if ok || ce != nil {
		t.Errorf("AsConvError(nil) should return (nil, false)")
	}
}

func TestToTime_FromTimeValue(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	got, err := conv.Time(now)
	if err != nil {
		t.Fatalf("Time(time.Time) error = %v", err)
	}
	if !got.Equal(now) {
		t.Errorf("Time(time.Time) = %v; want %v", got, now)
	}
}
