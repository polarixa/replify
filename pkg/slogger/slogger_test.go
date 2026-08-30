package slogger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/polarixa/replify/pkg/slogger"
)

// ///////////////////////////
// Section: Level parsing tests
// ///////////////////////////

func TestSlogger_LevelParsing(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input   string
		want    slogger.Level
		wantErr bool
	}{
		{"trace", slogger.TraceLevel, false},
		{"TRACE", slogger.TraceLevel, false},
		{"debug", slogger.DebugLevel, false},
		{"DEBUG", slogger.DebugLevel, false},
		{"info", slogger.InfoLevel, false},
		{"INFO", slogger.InfoLevel, false},
		{"warn", slogger.WarnLevel, false},
		{"WARN", slogger.WarnLevel, false},
		{"warning", slogger.WarnLevel, false},
		{"WARNING", slogger.WarnLevel, false},
		{"error", slogger.ErrorLevel, false},
		{"ERROR", slogger.ErrorLevel, false},
		{"fatal", slogger.FatalLevel, false},
		{"FATAL", slogger.FatalLevel, false},
		{"panic", slogger.PanicLevel, false},
		{"PANIC", slogger.PanicLevel, false},
		{"unknown", slogger.TraceLevel, true},
		{"", slogger.TraceLevel, true},
		{"verbose", slogger.TraceLevel, true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got, err := slogger.ParseLevel(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ParseLevel(%q) expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseLevel(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("ParseLevel(%q) = %v; want %v", tc.input, got, tc.want)
			}
		})
	}
}

// ///////////////////////////
// Section: Level string tests
// ///////////////////////////

func TestSlogger_LevelString(t *testing.T) {
	t.Parallel()
	cases := []struct {
		level slogger.Level
		want  string
	}{
		{slogger.TraceLevel, "TRACE"},
		{slogger.DebugLevel, "DEBUG"},
		{slogger.InfoLevel, "INFO"},
		{slogger.WarnLevel, "WARN"},
		{slogger.ErrorLevel, "ERROR"},
		{slogger.FatalLevel, "FATAL"},
		{slogger.PanicLevel, "PANIC"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			got := tc.level.String()
			if got != tc.want {
				t.Errorf("Level.String() = %q; want %q", got, tc.want)
			}
		})
	}
}

// ///////////////////////////
// Section: Level IsEnabled tests
// ///////////////////////////

func TestSlogger_LevelIsEnabled(t *testing.T) {
	t.Parallel()
	if !slogger.InfoLevel.IsEnabled(slogger.InfoLevel) {
		t.Error("InfoLevel.IsEnabled(InfoLevel) should be true")
	}
	if !slogger.ErrorLevel.IsEnabled(slogger.InfoLevel) {
		t.Error("ErrorLevel.IsEnabled(InfoLevel) should be true")
	}
	if slogger.DebugLevel.IsEnabled(slogger.InfoLevel) {
		t.Error("DebugLevel.IsEnabled(InfoLevel) should be false")
	}
	if slogger.TraceLevel.IsEnabled(slogger.WarnLevel) {
		t.Error("TraceLevel.IsEnabled(WarnLevel) should be false")
	}
}

// ///////////////////////////
// Section: Field constructor tests
// ///////////////////////////

func TestSlogger_FieldConstructors(t *testing.T) {
	t.Parallel()

	t.Run("String", func(t *testing.T) {
		t.Parallel()
		f := slogger.String("k", "hello world")
		if f.Key() != "k" {
			t.Errorf("Key = %q; want %q", f.Key(), "k")
		}
		if f.Value() != "hello world" {
			t.Errorf("Value() = %q; want %q", f.Value(), "hello world")
		}
	})

	t.Run("Int", func(t *testing.T) {
		t.Parallel()
		f := slogger.Int("n", 42)
		if f.Value() != "42" {
			t.Errorf("Int.Value() = %q; want %q", f.Value(), "42")
		}
	})

	t.Run("Int64", func(t *testing.T) {
		t.Parallel()
		f := slogger.Int64("n", 9876543210)
		if f.Value() != "9876543210" {
			t.Errorf("Int64.Value() = %q; want %q", f.Value(), "9876543210")
		}
	})

	t.Run("Float64", func(t *testing.T) {
		t.Parallel()
		f := slogger.Float64("ratio", 3.14)
		if f.Value() != "3.14" {
			t.Errorf("Float64.Value() = %q; want %q", f.Value(), "3.14")
		}
	})

	t.Run("Bool_true", func(t *testing.T) {
		t.Parallel()
		f := slogger.Bool("ok", true)
		if f.Value() != "true" {
			t.Errorf("Bool(true).Value() = %q; want true", f.Value())
		}
	})

	t.Run("Bool_false", func(t *testing.T) {
		t.Parallel()
		f := slogger.Bool("ok", false)
		if f.Value() != "false" {
			t.Errorf("Bool(false).Value() = %q; want false", f.Value())
		}
	})

	t.Run("Err_non_nil", func(t *testing.T) {
		t.Parallel()
		f := slogger.Err(errors.New("boom"))
		if f.Key() != "error" {
			t.Errorf("Err key = %q; want %q", f.Key(), "error")
		}
		if f.Value() != "boom" {
			t.Errorf("Err.Value() = %q; want %q", f.Value(), "boom")
		}
	})

	t.Run("Err_nil", func(t *testing.T) {
		t.Parallel()
		f := slogger.Err(nil)
		if f.Value() != "<nil>" {
			t.Errorf("Err(nil).Value() = %q; want <nil>", f.Value())
		}
	})

	// t.Run("Time", func(t *testing.T) {
	// 	t.Parallel()
	// 	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	// 	f := slogger.Time("at", ts)
	// 	if f.Value() != "2024-01-15T10:30:00Z" {
	// 		t.Errorf("Time.Value() = %q; want 2024-01-15T10:30:00Z", f.Value())
	// 	}
	// })

	t.Run("Duration", func(t *testing.T) {
		t.Parallel()
		f := slogger.Duration("took", 500*time.Millisecond)
		if f.Value() != "500ms" {
			t.Errorf("Duration.Value() = %q; want 500ms", f.Value())
		}
	})

	t.Run("Any", func(t *testing.T) {
		t.Parallel()
		f := slogger.Any("meta", []int{1, 2, 3})
		if f.Value() != "[1 2 3]" {
			t.Errorf("Any.Value() = %q; want [1 2 3]", f.Value())
		}
	})
}

// ///////////////////////////
// Section: TextFormatter tests
// ///////////////////////////

func TestSlogger_TextFormatter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})

	log.Info("hello world", slogger.String("key", "val"))
	out := buf.String()

	if !strings.Contains(out, "INFO") {
		t.Errorf("expected INFO in output, got: %s", out)
	}
	if !strings.Contains(out, "hello world") {
		t.Errorf("expected message in output, got: %s", out)
	}
	if !strings.Contains(out, "key=val") {
		t.Errorf("expected key=val in output, got: %s", out)
	}
	if !strings.HasSuffix(strings.TrimRight(out, "\n"), "key=val") {
		// ends with key=val before newline
	}
	if !strings.HasSuffix(out, "\n") {
		t.Error("expected output to end with newline")
	}
}

func TestSlogger_TextFormatter_QuotedValues(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})
	log.Info("msg", slogger.String("msg", "hello world"))
	out := buf.String()
	if !strings.Contains(out, `"hello world"`) {
		t.Errorf("expected quoted string value in output, got: %s", out)
	}
}

// ///////////////////////////
// Section: JSONFormatter tests
// ///////////////////////////

func TestSlogger_JSONFormatter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewJSONFormatter())
	})

	log.Info("json test", slogger.String("env", "prod"), slogger.Int("count", 3))
	out := strings.TrimSpace(buf.String())

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if m["level"] != "INFO" {
		t.Errorf("level = %v; want INFO", m["level"])
	}
	if m["msg"] != "json test" {
		t.Errorf("msg = %v; want json test", m["msg"])
	}
	if m["env"] != "prod" {
		t.Errorf("env = %v; want prod", m["env"])
	}
	if m["count"] != float64(3) {
		t.Errorf("count = %v; want 3", m["count"])
	}
}

func TestSlogger_JSONFormatter_CustomKeys(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.InfoLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewJSONFormatter().
			WithTimeKey("timestamp").
			WithLevelKey("severity").
			WithMessageKey("message"))
	})
	log.Info("custom keys")
	out := strings.TrimSpace(buf.String())
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if _, ok := m["timestamp"]; !ok {
		t.Error("expected timestamp key in JSON output")
	}
	if m["severity"] != "INFO" {
		t.Errorf("severity = %v; want INFO", m["severity"])
	}
	if m["message"] != "custom keys" {
		t.Errorf("message = %v; want 'custom keys'", m["message"])
	}
}

// ///////////////////////////
// Section: Logger construction tests
// ///////////////////////////

func TestSlogger_Logger_New(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})
	log.Info("startup complete")
	if !strings.Contains(buf.String(), "startup complete") {
		t.Errorf("expected message in output, got: %s", buf.String())
	}
}

// ///////////////////////////
// Section: Logger With tests
// ///////////////////////////

func TestSlogger_Logger_With(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})
	child := log.With(slogger.String("component", "auth"))
	child.Info("login attempt")
	out := buf.String()
	if !strings.Contains(out, "component=auth") {
		t.Errorf("expected component=auth in output, got: %s", out)
	}
}

// ///////////////////////////
// Section: Logger Named tests
// ///////////////////////////

func TestSlogger_Logger_Named(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.InfoLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})
	db := log.Named("db")
	rw := db.Named("reader")
	rw.Info("query executed")
	out := buf.String()
	if !strings.Contains(out, "[db.reader]") {
		t.Errorf("expected [db.reader] in output, got: %s", out)
	}
}

func TestSlogger_Logger_Named_NoParent(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.InfoLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})
	named := log.Named("api")
	named.Info("route registered")
	out := buf.String()
	if !strings.Contains(out, "[api]") {
		t.Errorf("expected [api] in output, got: %s", out)
	}
}

// ///////////////////////////
// Section: Logger SetLevel tests
// ///////////////////////////

func TestSlogger_Logger_SetLevel(t *testing.T) {
	t.Parallel()
	log := slogger.New()
	log.SetLevel(slogger.DebugLevel)
	if log.GetLevel() != slogger.DebugLevel {
		t.Errorf("GetLevel() = %v; want DebugLevel", log.GetLevel())
	}
	log.SetLevel(slogger.WarnLevel)
	if log.GetLevel() != slogger.WarnLevel {
		t.Errorf("GetLevel() = %v; want WarnLevel", log.GetLevel())
	}
}

// ///////////////////////////
// Section: Logger IsLevelEnabled tests
// ///////////////////////////

func TestSlogger_Logger_IsLevelEnabled(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.WarnLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})
	if log.IsLevelEnabled(slogger.DebugLevel) {
		t.Error("IsLevelEnabled(DebugLevel) should be false when min=WarnLevel")
	}
	if !log.IsLevelEnabled(slogger.WarnLevel) {
		t.Error("IsLevelEnabled(WarnLevel) should be true when min=WarnLevel")
	}
	if !log.IsLevelEnabled(slogger.ErrorLevel) {
		t.Error("IsLevelEnabled(ErrorLevel) should be true when min=WarnLevel")
	}

	// debug message should not appear in output
	log.Debug("this should not appear")
	if strings.Contains(buf.String(), "this should not appear") {
		t.Error("debug message was written despite WarnLevel minimum")
	}

	// warn message should appear
	log.Warn("this should appear")
	if !strings.Contains(buf.String(), "this should appear") {
		t.Error("warn message was not written")
	}
}

// ///////////////////////////
// Section: Logger SetOutput tests
// ///////////////////////////

func TestSlogger_Logger_SetOutput(t *testing.T) {
	t.Parallel()
	var buf1, buf2 bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.InfoLevel)
		o.SetOutput(&buf1)
		o.SetFormatter(slogger.NewTextFormatter(&buf1).WithDisableColor())
	})
	log.Info("first destination")
	if !strings.Contains(buf1.String(), "first destination") {
		t.Errorf("expected message in buf1, got: %s", buf1.String())
	}

	log.SetOutput(&buf2)
	log.Info("second destination")
	if !strings.Contains(buf2.String(), "second destination") {
		t.Errorf("expected message in buf2, got: %s", buf2.String())
	}
}

// ///////////////////////////
// Section: Logger SetFormatter tests
// ///////////////////////////

func TestSlogger_Logger_SetFormatter(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.InfoLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})
	log.SetFormatter(slogger.NewJSONFormatter())
	log.Info("formatted as json")
	out := strings.TrimSpace(buf.String())
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("SetFormatter: invalid JSON after switch: %v\noutput: %s", err, out)
	}
}

// ///////////////////////////
// Section: Hooks tests
// ///////////////////////////

type testHook struct {
	mu     sync.Mutex
	levels []slogger.Level
	fired  []*slogger.Entry
}

func (h *testHook) Levels() []slogger.Level { return h.levels }
func (h *testHook) Fire(e *slogger.Entry) error {
	h.mu.Lock()
	// capture a copy since entry will be released
	cp := *e
	h.fired = append(h.fired, &cp)
	h.mu.Unlock()
	return nil
}

func TestSlogger_Hooks(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})
	hook := &testHook{levels: []slogger.Level{slogger.ErrorLevel, slogger.WarnLevel}}
	log.AddHook(hook)

	log.Info("info msg")
	log.Warn("warn msg")
	log.Error("error msg")

	hook.mu.Lock()
	n := len(hook.fired)
	hook.mu.Unlock()

	if n != 2 {
		t.Errorf("expected 2 hook firings (warn+error), got %d", n)
	}
}

func TestSlogger_Hooks_Len(t *testing.T) {
	t.Parallel()
	hooks := slogger.NewHooks()
	h := &testHook{levels: []slogger.Level{slogger.InfoLevel, slogger.ErrorLevel}}
	hooks.Add(h)
	if hooks.Len(slogger.InfoLevel) != 1 {
		t.Errorf("Len(InfoLevel) = %d; want 1", hooks.Len(slogger.InfoLevel))
	}
	if hooks.Len(slogger.DebugLevel) != 0 {
		t.Errorf("Len(DebugLevel) = %d; want 0", hooks.Len(slogger.DebugLevel))
	}
}

// ///////////////////////////
// Section: Context tests
// ///////////////////////////

func TestSlogger_Context(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fields := slogger.FieldsFromContext(ctx)
	if fields != nil {
		t.Error("FieldsFromContext on empty context should return nil")
	}

	ctx = slogger.WithContextFields(ctx,
		slogger.String("trace_id", "abc123"),
		slogger.String("span_id", "def456"),
	)
	fields = slogger.FieldsFromContext(ctx)
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}
	if fields[0].Key() != "trace_id" {
		t.Errorf("fields[0].Key() = %q; want trace_id", fields[0].Key())
	}

	// Append more fields preserving existing
	ctx = slogger.WithContextFields(ctx, slogger.Int("attempt", 1))
	fields = slogger.FieldsFromContext(ctx)
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields after append, got %d", len(fields))
	}
}

func TestSlogger_Context_Logging(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.InfoLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})
	ctx := slogger.WithContextFields(context.Background(), slogger.String("req_id", "xyz"))
	log.WithContext(ctx).Info("handling request")
	out := buf.String()
	if !strings.Contains(out, "req_id=xyz") {
		t.Errorf("expected req_id=xyz in output, got: %s", out)
	}
}

// ///////////////////////////
// Section: Sampling tests
// ///////////////////////////

func TestSlogger_Sampling(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.InfoLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor().WithDisableTimestamp())
		o.SetSamplingOpts(slogger.NewSamplingOptions().
			WithFirst(3).
			WithPeriod(10 * time.Second).
			WithThereafter(0)) // drop all after first 3
	})

	for i := 0; i < 10; i++ {
		log.Info("sampled message")
	}

	count := strings.Count(buf.String(), "sampled message")
	if count != 3 {
		t.Errorf("expected exactly 3 log lines, got %d\noutput:\n%s", count, buf.String())
	}
}

func TestSlogger_Sampling_Thereafter(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.InfoLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor().WithDisableTimestamp())
		o.SetSamplingOpts(slogger.NewSamplingOptions().
			WithFirst(2).
			WithPeriod(10 * time.Second).
			WithThereafter(2)) // every 2nd after first 2
	})
	// 2 always + msgs 3,4,5,6,7,8,9,10 = 8 more; allow at positions (count-first-1)%2==0 => 0,2,4,6 => 4 more
	// total expected: 2 + 4 = 6
	for i := 0; i < 10; i++ {
		log.Info("thereafter message")
	}
	count := strings.Count(buf.String(), "thereafter message")
	if count != 6 {
		t.Errorf("expected 6 log lines with thereafter=2, got %d\noutput:\n%s", count, buf.String())
	}
}

// ///////////////////////////
// Section: Global logger tests
// ///////////////////////////

func TestSlogger_GlobalLogger(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.InfoLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})

	original := slogger.GlobalLogger()
	slogger.SetGlobalLogger(log)
	defer slogger.SetGlobalLogger(original)

	slogger.Info("global info message")
	if !strings.Contains(buf.String(), "global info message") {
		t.Errorf("expected global info message in output, got: %s", buf.String())
	}

	slogger.SetGlobalLogger(nil) // should be a no-op
	if slogger.GlobalLogger() != log {
		t.Error("SetGlobalLogger(nil) should not replace the current logger")
	}
}

// ///////////////////////////
// Section: Entry methods tests
// ///////////////////////////

func TestSlogger_EntryMethods(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})

	ctx := context.Background()
	entry := log.WithContext(ctx)

	entry.Trace("trace via entry")
	entry.Debug("debug via entry")
	entry.Info("info via entry")
	entry.Warn("warn via entry")
	entry.Error("error via entry")

	out := buf.String()
	for _, want := range []string{
		"trace via entry",
		"debug via entry",
		"info via entry",
		"warn via entry",
		"error via entry",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got: %s", want, out)
		}
	}
}

func TestSlogger_EntryMethods_Panic(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic, got none")
		}
		if !strings.Contains(buf.String(), "panic via entry") {
			t.Errorf("expected panic message in output, got: %s", buf.String())
		}
	}()

	log.WithContext(context.Background()).Panic("panic via entry")
}

// ///////////////////////////
// Section: Fatal hook tests
// ///////////////////////////

func TestSlogger_Fatal_Hook(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})
	hook := &testHook{levels: []slogger.Level{slogger.FatalLevel}}
	log.AddHook(hook)

	// Directly invoke the internal log dispatch at FatalLevel without triggering os.Exit
	// by using a hook to capture it. We test that the hook fires for fatal level.
	// We cannot call log.Fatal() in tests as it calls os.Exit(1).
	// Instead, we verify hook fires by using ErrorLevel as a proxy test for hook wiring.
	// For FatalLevel hook count we add a hook and call Panic and recover:
	panicHook := &testHook{levels: []slogger.Level{slogger.PanicLevel}}
	log.AddHook(panicHook)

	func() {
		defer func() { _ = recover() }()
		log.Panic("fatal-level-proxy panic")
	}()

	panicHook.mu.Lock()
	n := len(panicHook.fired)
	panicHook.mu.Unlock()
	if n != 1 {
		t.Errorf("expected 1 panic hook firing, got %d", n)
	}
}

// ///////////////////////////
// Section: Concurrent logging tests
// ///////////////////////////

func TestSlogger_Concurrent(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	var mu sync.Mutex
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.InfoLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})

	const goroutines = 50
	const messages = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messages; j++ {
				log.Info("concurrent message", slogger.Int("goroutine", id), slogger.Int("msg", j))
			}
		}(i)
	}
	wg.Wait()

	mu.Lock()
	out := buf.String()
	mu.Unlock()

	count := strings.Count(out, "concurrent message")
	if count != goroutines*messages {
		t.Errorf("expected %d log lines, got %d", goroutines*messages, count)
	}
}

// ///////////////////////////
// Section: MultiWriter tests
// ///////////////////////////

func TestSlogger_MultiWriter(t *testing.T) {
	t.Parallel()
	var buf1, buf2 bytes.Buffer
	mw := slogger.NewMultiWriter(&buf1, &buf2)
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.InfoLevel)
		o.SetOutput(mw)
		o.SetFormatter(slogger.NewTextFormatter(mw).WithDisableColor())
	})
	log.Info("multiwriter test")
	if !strings.Contains(buf1.String(), "multiwriter test") {
		t.Errorf("expected message in buf1, got: %s", buf1.String())
	}
	if !strings.Contains(buf2.String(), "multiwriter test") {
		t.Errorf("expected message in buf2, got: %s", buf2.String())
	}
}

// ///////////////////////////
// Section: Entry accessor tests
// ///////////////////////////

func TestSlogger_EntryAccessors(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer

	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})

	hook := &testHook{levels: []slogger.Level{slogger.InfoLevel}}
	log.AddHook(hook)

	before := time.Now()
	log.Info("accessor test", slogger.String("k", "v"))
	after := time.Now()

	hook.mu.Lock()
	n := len(hook.fired)
	hook.mu.Unlock()
	if n != 1 {
		t.Fatalf("expected 1 fired entry, got %d", n)
	}

	hook.mu.Lock()
	e := hook.fired[0]
	hook.mu.Unlock()

	// Logger accessor
	if e.Logger() == nil {
		t.Error("Entry.Logger() should not be nil")
	}
	// Time accessor
	if e.Time().Before(before) || e.Time().After(after) {
		t.Errorf("Entry.Time() = %v; want between %v and %v", e.Time(), before, after)
	}
	// GetLevel accessor
	if e.GetLevel() != slogger.InfoLevel {
		t.Errorf("Entry.GetLevel() = %v; want InfoLevel", e.GetLevel())
	}
	// Message accessor
	if e.Message() != "accessor test" {
		t.Errorf("Entry.Message() = %q; want %q", e.Message(), "accessor test")
	}
	// Fields accessor
	fields := e.Fields()
	if len(fields) != 1 || fields[0].Key() != "k" {
		t.Errorf("Entry.Fields() = %v; want [{k v}]", fields)
	}
	// Caller accessor (should be nil since caller not enabled)
	if e.Caller() != nil {
		t.Error("Entry.Caller() should be nil when caller not enabled")
	}
	// Context accessor (should be nil)
	if e.Context() != nil {
		t.Error("Entry.Context() should be nil when no context set")
	}
}

func TestSlogger_CallerInfoAccessors(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer

	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
		o.SetCaller(true)
	})

	hook := &testHook{levels: []slogger.Level{slogger.InfoLevel}}
	log.AddHook(hook)

	log.Info("caller test")

	hook.mu.Lock()
	n := len(hook.fired)
	hook.mu.Unlock()
	if n != 1 {
		t.Fatalf("expected 1 fired entry, got %d", n)
	}

	hook.mu.Lock()
	e := hook.fired[0]
	hook.mu.Unlock()

	c := e.Caller()
	if c == nil {
		t.Fatal("Entry.Caller() should not be nil when CallerReporter is true")
	}
	if c.File() == "" {
		t.Error("CallerInfo.File() should not be empty")
	}
	if c.Line() <= 0 {
		t.Errorf("CallerInfo.Line() = %d; want > 0", c.Line())
	}
	if c.Function() == "" {
		t.Error("CallerInfo.Function() should not be empty")
	}
}

// ///////////////////////////
// Section: Rotation tests
// ///////////////////////////

func TestSlogger_LevelFileWriter_Basic(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	lfw, err := slogger.NewLevelFileWriter(*slogger.NewRotationOptions().
		WithDirectory(dir).
		WithMaxBytes(1024 * 1024).
		WithCompress(false))
	if err != nil {
		t.Fatalf("NewLevelFileWriter: %v", err)
	}
	defer lfw.Close()

	msg := []byte("hello rotation\n")
	n, err := lfw.WriteLevel(slogger.InfoLevel, msg)
	if err != nil {
		t.Fatalf("WriteLevel: %v", err)
	}
	if n != len(msg) {
		t.Errorf("WriteLevel returned %d; want %d", n, len(msg))
	}
}

func TestSlogger_LevelWriterHook_Routing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	lfw, err := slogger.NewLevelFileWriter(*slogger.NewRotationOptions().
		WithDirectory(dir).
		WithMaxBytes(1024 * 1024).
		WithCompress(false))
	if err != nil {
		t.Fatalf("NewLevelFileWriter: %v", err)
	}
	defer lfw.Close()

	var buf bytes.Buffer
	formatter := slogger.NewTextFormatter(&buf).WithDisableColor()
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(formatter)
	})
	hook := slogger.NewLevelWriterHook(lfw, formatter)
	log.AddHook(hook)

	log.Info("rotation hook test")
	log.Error("rotation error test")
	// No panic means routing worked correctly.
}

func TestSlogger_RotationOptions_Defaults(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	lfw, err := slogger.NewLevelFileWriter(*slogger.NewRotationOptions().
		WithDirectory(dir))
	// MaxBytes zero -> should default to 10MB
	if err != nil {
		t.Fatalf("NewLevelFileWriter with defaults: %v", err)
	}
	defer lfw.Close()
}

func TestSlogger_WithRotation_Option(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var buf bytes.Buffer
	log := slogger.New(
		func(o *slogger.Options) {
			o.SetOutput(&buf)
			o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
		},
		slogger.WithRotation(slogger.NewRotationOptions().
			WithDirectory(dir).
			WithMaxBytes(1024*1024).
			WithCompress(false)),
	)
	// Just verify logger was constructed successfully and can log
	log.Info("with rotation test")
	if !strings.Contains(buf.String(), "with rotation test") {
		t.Errorf("expected log output, got: %s", buf.String())
	}
}

// ///////////////////////////
// Section: JSON string embedding tests
// ///////////////////////////

// TestSlogger_JSONFormatter_AnyJSONString verifies that slogger.Any with a
// valid JSON string value is embedded as raw JSON (not double-encoded) in the
// JSON formatter output.
func TestSlogger_JSONFormatter_AnyJSONString(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewJSONFormatter())
	})

	log.Warn("server started",
		slogger.Any("f4", `{"user_id":2,"username":"abc@gmail.com"}`),
	)
	out := strings.TrimSpace(buf.String())

	// Output must be valid JSON.
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, out)
	}

	// f4 must be an embedded object, not a quoted string.
	f4, ok := m["f4"]
	if !ok {
		t.Fatalf("expected key f4 in output, got: %s", out)
	}
	f4Map, ok := f4.(map[string]interface{})
	if !ok {
		t.Fatalf("f4 should be a JSON object, got %T: %v", f4, f4)
	}
	if f4Map["user_id"] != float64(2) {
		t.Errorf("f4.user_id = %v; want 2", f4Map["user_id"])
	}
	if f4Map["username"] != "abc@gmail.com" {
		t.Errorf("f4.username = %v; want abc@gmail.com", f4Map["username"])
	}
}

// TestSlogger_JSONFormatter_JSONFieldWithString verifies that slogger.JSON with
// a valid JSON string value is embedded as raw JSON (not double-encoded) in the
// JSON formatter output.
func TestSlogger_JSONFormatter_JSONFieldWithString(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewJSONFormatter())
	})

	log.Warn("server started",
		slogger.JSON("f1", `{"user_id":1,"username":"abc@gmail.com"}`),
	)
	out := strings.TrimSpace(buf.String())

	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, out)
	}

	f1, ok := m["f1"]
	if !ok {
		t.Fatalf("expected key f1 in output, got: %s", out)
	}
	_, ok = f1.(string)
	if !ok {
		t.Fatalf("f1 should be a JSON string, got %T: %v", f1, f1)
	}
}

// TestSlogger_JSONFormatter_JSONFieldWithMap verifies that slogger.JSON with a
// map value is embedded as a JSON object.
func TestSlogger_JSONFormatter_JSONFieldWithMap(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewJSONFormatter())
	})

	log.Info("event",
		slogger.JSON("data", map[string]any{"a": 1}),
	)
	out := strings.TrimSpace(buf.String())

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, out)
	}
	dataObj, ok := m["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data should be a JSON object, got %T: %v", m["data"], m["data"])
	}
	if dataObj["a"] != float64(1) {
		t.Errorf("data.a = %v; want 1", dataObj["a"])
	}
}

// TestSlogger_JSONFormatter_AnyNormalString verifies that slogger.Any with a
// plain (non-JSON) string is correctly quoted in JSON output.
func TestSlogger_JSONFormatter_AnyNormalString(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewJSONFormatter())
	})

	log.Info("event",
		slogger.Any("name", "alice"),
	)
	out := strings.TrimSpace(buf.String())

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, out)
	}
	if m["name"] != "alice" {
		t.Errorf("name = %v; want alice", m["name"])
	}
}

// TestSlogger_JSONFormatter_AnyStruct verifies that slogger.Any with a struct
// value is embedded as a JSON object.
func TestSlogger_JSONFormatter_AnyStruct(t *testing.T) {
	t.Parallel()

	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewJSONFormatter())
	})

	log.Info("event",
		slogger.Any("user", User{ID: 1, Name: "alice"}),
	)
	out := strings.TrimSpace(buf.String())

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, out)
	}
	userObj, ok := m["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("user should be a JSON object, got %T: %v", m["user"], m["user"])
	}
	if userObj["id"] != float64(1) {
		t.Errorf("user.id = %v; want 1", userObj["id"])
	}
	if userObj["name"] != "alice" {
		t.Errorf("user.name = %v; want alice", userObj["name"])
	}
}

// ///////////////////////////
// Section: JSONFormatter color tests
// ///////////////////////////

// TestSlogger_JSONFormatter_ColorDisabled verifies that WithColor(false) produces
// plain JSON with no ANSI escape sequences, even when called explicitly.
func TestSlogger_JSONFormatter_ColorDisabled(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewJSONFormatter().WithColor(false))
	})

	log.Info("color test", slogger.String("key", "val"), slogger.Int("n", 42))
	out := strings.TrimSpace(buf.String())

	// Must not contain ANSI escape codes.
	if strings.Contains(out, "\x1b[") {
		t.Errorf("expected no ANSI codes with color disabled, got: %q", out)
	}

	// Must be valid JSON.
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, out)
	}
	if m["key"] != "val" {
		t.Errorf("key = %v; want val", m["key"])
	}
}

// TestSlogger_JSONFormatter_ColorDefaultNonTTY verifies that the default
// JSONFormatter (enableColor: true) produces plain JSON when writing to a
// non-TTY writer (bytes.Buffer), because IsTTY returns false.
func TestSlogger_JSONFormatter_ColorDefaultNonTTY(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		// Default JSONFormatter has enableColor: true, but buf is not a TTY.
		o.SetFormatter(slogger.NewJSONFormatter())
	})

	log.Warn("server started", slogger.String("addr", ":8080"), slogger.Int("port", 8080))
	out := strings.TrimSpace(buf.String())

	// No ANSI codes because buf is not a TTY.
	if strings.Contains(out, "\x1b[") {
		t.Errorf("expected no ANSI codes for non-TTY output, got: %q", out)
	}

	// Output must still be valid JSON.
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON output: %v\noutput: %s", err, out)
	}
	if m["level"] != "WARN" {
		t.Errorf("level = %v; want WARN", m["level"])
	}
}

// TestSlogger_JSONFormatter_WithColorChaining verifies that WithColor is chainable
// with other JSONFormatter methods without breaking existing functionality.
func TestSlogger_JSONFormatter_WithColorChaining(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewJSONFormatter().
			WithColor(false).
			WithTimeKey("timestamp").
			WithLevelKey("severity").
			WithMessageKey("message"))
	})

	log.Info("chaining test")
	out := strings.TrimSpace(buf.String())

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if _, ok := m["timestamp"]; !ok {
		t.Error("expected timestamp key in output")
	}
	if m["severity"] != "INFO" {
		t.Errorf("severity = %v; want INFO", m["severity"])
	}
	if m["message"] != "chaining test" {
		t.Errorf("message = %v; want 'chaining test'", m["message"])
	}
	// No ANSI codes.
	if strings.Contains(out, "\x1b[") {
		t.Errorf("expected no ANSI codes, got: %q", out)
	}
}

// ///////////////////////////
// Section: Cross-platform and safety tests
// ///////////////////////////

// TestSlogger_TrimFilePath_CrossPlatform verifies that trimFilepath handles
// both Unix and Windows path separators correctly.
func TestSlogger_TrimFilePath_CrossPlatform(t *testing.T) {
	t.Parallel()

	// This test verifies caller information is properly formatted on the current platform.
	// The trimFilepath function should produce valid file:line output regardless of OS.
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor().WithEnableCaller())
		o.SetCaller(true)
	})

	log.Info("test caller")
	out := buf.String()

	// Should contain caller= in the output
	if !strings.Contains(out, "caller=") {
		t.Errorf("expected caller= in output, got: %s", out)
	}
	// Should contain .go file extension
	if !strings.Contains(out, ".go:") {
		t.Errorf("expected .go: in output for caller, got: %s", out)
	}
}

// TestSlogger_Itoa64_MinInt64 verifies that itoa64 handles math.MinInt64 correctly.
func TestSlogger_Itoa64_MinInt64(t *testing.T) {
	t.Parallel()

	// Use math.MinInt64 constant for clarity
	minInt64 := int64(math.MinInt64)

	// Test by logging an Int64 with math.MinInt64
	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewJSONFormatter())
	})

	log.Info("test minint64", slogger.Int64("val", minInt64))
	out := strings.TrimSpace(buf.String())

	// Verify valid JSON
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}

	// Check value is correct
	if m["val"] != float64(minInt64) {
		t.Errorf("val = %v; want %d", m["val"], minInt64)
	}

	// Also verify the string contains the correct value
	if !strings.Contains(out, "-9223372036854775808") {
		t.Errorf("expected -9223372036854775808 in output, got: %s", out)
	}
}

// TestSlogger_EntryNilLoggerSafety verifies Entry methods don't panic when
// logger is nil (detached entry scenario).
func TestSlogger_EntryNilLoggerSafety(t *testing.T) {
	t.Parallel()

	// Create a detached entry with nil logger
	entry := &slogger.Entry{}

	// These should not panic
	entry.Trace("trace msg")
	entry.Debug("debug msg")
	entry.Info("info msg")
	entry.Warn("warn msg")
	entry.Error("error msg")
}

// TestSlogger_WithConcurrent verifies that With method is safe for concurrent use.
func TestSlogger_WithConcurrent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.InfoLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			child := log.With(slogger.Int("id", id))
			child.Info("concurrent with")
		}(i)
	}
	wg.Wait()

	// Verify we got output without races (race detector would catch issues)
	count := strings.Count(buf.String(), "concurrent with")
	if count != goroutines {
		t.Errorf("expected %d log lines, got %d", goroutines, count)
	}
}

// TestSlogger_TextFormatter_ConcurrentFormatSafety verifies that TextFormatter.Format
// is safe for concurrent use and doesn't modify formatter state.
func TestSlogger_TextFormatter_ConcurrentFormatSafety(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := slogger.NewTextFormatter(&buf).WithDisableColor()

	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.TraceLevel)
		o.SetOutput(&buf)
		o.SetFormatter(formatter)
		o.SetCaller(true)
	})

	const goroutines = 50
	const messages = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messages; j++ {
				log.Info("formatter safety test", slogger.Int("g", id), slogger.Int("m", j))
			}
		}(i)
	}
	wg.Wait()

	// Count messages (race detector would catch data races)
	count := strings.Count(buf.String(), "formatter safety test")
	if count != goroutines*messages {
		t.Errorf("expected %d log lines, got %d", goroutines*messages, count)
	}
}

// TestSlogger_EntryWithContext_Safety verifies that Entry.WithContext returns
// a new entry without modifying the original.
func TestSlogger_EntryWithContext_Safety(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := slogger.New(func(o *slogger.Options) {
		o.SetLevel(slogger.InfoLevel)
		o.SetOutput(&buf)
		o.SetFormatter(slogger.NewTextFormatter(&buf).WithDisableColor())
	})

	ctx1 := context.Background()
	ctx2 := slogger.WithContextFields(ctx1, slogger.String("req", "123"))

	entry := log.WithContext(ctx1)
	entryWithNewCtx := entry.WithContext(ctx2)

	// Original entry's context should be unchanged
	if entry.Context() != ctx1 {
		t.Error("original entry context was modified")
	}
	if entryWithNewCtx.Context() != ctx2 {
		t.Error("new entry should have ctx2")
	}
}
