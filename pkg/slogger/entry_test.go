package slogger

import (
	"context"
	"testing"
	"time"
)

// =============================================================================
// Entry Accessor Tests
// =============================================================================

func TestEntry_Logger(t *testing.T) {
	t.Parallel()

	t.Run("returns logger", func(t *testing.T) {
		t.Parallel()
		log := New()
		entry := &Entry{logger: log}
		assertEqual(t, log, entry.Logger())
	})

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var entry *Entry
		assertNil(t, entry.Logger())
	})
}

func TestEntry_Time(t *testing.T) {
	t.Parallel()

	t.Run("returns time", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		entry := &Entry{time: now}
		assertEqual(t, now, entry.Time())
	})

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var entry *Entry
		assertEqual(t, time.Time{}, entry.Time())
	})

	t.Run("zero time", func(t *testing.T) {
		t.Parallel()
		entry := &Entry{}
		assertTrue(t, entry.Time().IsZero())
	})
}

func TestEntry_Level(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level Level
	}{
		{name: "trace", level: TraceLevel},
		{name: "debug", level: DebugLevel},
		{name: "info", level: InfoLevel},
		{name: "warn", level: WarnLevel},
		{name: "error", level: ErrorLevel},
		{name: "fatal", level: FatalLevel},
		{name: "panic", level: PanicLevel},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			entry := &Entry{level: tt.level}
			assertEqual(t, tt.level, entry.Level())
		})
	}

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var entry *Entry
		assertEqual(t, TraceLevel, entry.Level())
	})
}

func TestEntry_GetLevel(t *testing.T) {
	t.Parallel()

	entry := &Entry{level: WarnLevel}
	assertEqual(t, WarnLevel, entry.GetLevel())
}

func TestEntry_Message(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
	}{
		{name: "simple message", message: "hello world"},
		{name: "empty message", message: ""},
		{name: "unicode message", message: "日本語🎉"},
		{name: "special chars", message: "line1\nline2\ttab"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			entry := &Entry{message: tt.message}
			assertEqual(t, tt.message, entry.Message())
		})
	}

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var entry *Entry
		assertEqual(t, "", entry.Message())
	})
}

func TestEntry_Fields(t *testing.T) {
	t.Parallel()

	t.Run("returns fields", func(t *testing.T) {
		t.Parallel()
		fields := []Field{String("a", "1"), Int("b", 2)}
		entry := &Entry{fields: fields}
		got := entry.Fields()
		assertLen(t, got, 2)
		assertEqual(t, "a", got[0].Key())
	})

	t.Run("returns copy", func(t *testing.T) {
		t.Parallel()
		fields := []Field{String("original", "value")}
		entry := &Entry{fields: fields}
		got := entry.Fields()
		got[0] = String("modified", "value")
		// Original should be unchanged
		assertEqual(t, "original", entry.Fields()[0].Key())
	})

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var entry *Entry
		assertNil(t, entry.Fields())
	})

	t.Run("nil fields", func(t *testing.T) {
		t.Parallel()
		entry := &Entry{}
		assertNil(t, entry.Fields())
	})
}

func TestEntry_Caller(t *testing.T) {
	t.Parallel()

	t.Run("returns caller", func(t *testing.T) {
		t.Parallel()
		caller := &CallerInfo{file: "test.go", line: 42, function: "testFunc"}
		entry := &Entry{caller: caller}
		assertEqual(t, caller, entry.Caller())
	})

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var entry *Entry
		assertNil(t, entry.Caller())
	})

	t.Run("nil caller", func(t *testing.T) {
		t.Parallel()
		entry := &Entry{}
		assertNil(t, entry.Caller())
	})
}

func TestEntry_Context(t *testing.T) {
	t.Parallel()

	t.Run("returns context", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		entry := &Entry{ctx: ctx}
		assertEqual(t, ctx, entry.Context())
	})

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var entry *Entry
		assertNil(t, entry.Context())
	})

	t.Run("nil context", func(t *testing.T) {
		t.Parallel()
		entry := &Entry{}
		assertNil(t, entry.Context())
	})
}

// =============================================================================
// Entry WithContext Tests
// =============================================================================

func TestEntry_WithContext(t *testing.T) {
	t.Parallel()

	t.Run("creates copy with context", func(t *testing.T) {
		t.Parallel()
		ctx := context.WithValue(context.Background(), "key", "value")
		original := &Entry{message: "original", level: InfoLevel}
		copy := original.WithContext(ctx)

		// Copy should have the context
		assertEqual(t, ctx, copy.Context())
		// Copy should have original values
		assertEqual(t, "original", copy.Message())
		assertEqual(t, InfoLevel, copy.Level())
		// Original should not have context
		assertNil(t, original.Context())
	})
}

// =============================================================================
// Entry reset Tests
// =============================================================================

func TestEntry_reset(t *testing.T) {
	t.Parallel()

	entry := &Entry{
		logger:  New(),
		time:    time.Now(),
		level:   ErrorLevel,
		message: "test message",
		fields:  []Field{String("a", "1")},
		caller:  &CallerInfo{file: "test.go", line: 42},
		ctx:     context.Background(),
	}

	entry.reset()

	assertNil(t, entry.logger)
	assertTrue(t, entry.time.IsZero())
	assertEqual(t, TraceLevel, entry.level)
	assertEqual(t, "", entry.message)
	assertLen(t, entry.fields, 0)
	assertNil(t, entry.caller)
	assertNil(t, entry.ctx)
}

// =============================================================================
// Entry Logging Methods Tests
// =============================================================================

func TestEntry_LoggingMethods_NilLogger(t *testing.T) {
	t.Parallel()

	entry := &Entry{} // No logger set

	// Should not panic
	assertNotPanics(t, func() { entry.Trace("test") })
	assertNotPanics(t, func() { entry.Debug("test") })
	assertNotPanics(t, func() { entry.Info("test") })
	assertNotPanics(t, func() { entry.Warn("test") })
	assertNotPanics(t, func() { entry.Error("test") })
}

func TestEntry_Panic_NilLogger(t *testing.T) {
	t.Parallel()

	entry := &Entry{} // No logger set

	defer func() {
		r := recover()
		assertNotNil(t, r)
		assertEqual(t, "test panic", r)
	}()

	entry.Panic("test panic")
}

// =============================================================================
// CallerInfo Tests
// =============================================================================

func TestCallerInfo_File(t *testing.T) {
	t.Parallel()

	t.Run("returns file", func(t *testing.T) {
		t.Parallel()
		c := &CallerInfo{file: "pkg/test.go"}
		assertEqual(t, "pkg/test.go", c.File())
	})

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var c *CallerInfo
		assertEqual(t, "", c.File())
	})
}

func TestCallerInfo_Line(t *testing.T) {
	t.Parallel()

	t.Run("returns line", func(t *testing.T) {
		t.Parallel()
		c := &CallerInfo{line: 42}
		assertEqual(t, 42, c.Line())
	})

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var c *CallerInfo
		assertEqual(t, 0, c.Line())
	})
}

func TestCallerInfo_Function(t *testing.T) {
	t.Parallel()

	t.Run("returns function", func(t *testing.T) {
		t.Parallel()
		c := &CallerInfo{function: "github.com/polarixa/replify/pkg/slogger.TestCallerInfo"}
		assertEqual(t, "github.com/polarixa/replify/pkg/slogger.TestCallerInfo", c.Function())
	})

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var c *CallerInfo
		assertEqual(t, "", c.Function())
	})
}

// =============================================================================
// CallerInfo Edge Cases
// =============================================================================

func TestCallerInfo_ZeroValue(t *testing.T) {
	t.Parallel()

	c := &CallerInfo{}
	assertEqual(t, "", c.File())
	assertEqual(t, 0, c.Line())
	assertEqual(t, "", c.Function())
}

// =============================================================================
// Entry Pool Tests
// =============================================================================

func TestEntryPool(t *testing.T) {
	t.Parallel()

	t.Run("acquires and releases", func(t *testing.T) {
		t.Parallel()
		log := New()

		entry := acquireEntry(log)
		assertNotNil(t, entry)
		assertEqual(t, log, entry.logger)

		releaseEntry(entry)
		// After release, entry should be reset
		assertNil(t, entry.logger)
	})

	t.Run("reuses entries", func(t *testing.T) {
		t.Parallel()
		log := New()

		// Acquire and release multiple times
		for i := 0; i < 100; i++ {
			entry := acquireEntry(log)
			entry.message = "test"
			releaseEntry(entry)
		}

		// Pool should still work
		entry := acquireEntry(log)
		assertNotNil(t, entry)
		releaseEntry(entry)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkEntry_Fields(b *testing.B) {
	entry := &Entry{
		fields: []Field{String("a", "1"), Int("b", 2), Bool("c", true)},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = entry.Fields()
	}
}

func BenchmarkEntry_reset(b *testing.B) {
	log := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry := acquireEntry(log)
		entry.message = "test"
		entry.level = InfoLevel
		releaseEntry(entry)
	}
}

func BenchmarkAcquireReleaseEntry(b *testing.B) {
	log := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry := acquireEntry(log)
		releaseEntry(entry)
	}
}
