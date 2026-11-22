package replify_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/polarixa/replify"
)

// requestedTimeJSON returns meta.requested_time from the serialized wrapper, and whether it exists.
func requestedTimeJSON(t *testing.T, w interface{ JSON() string }) (string, bool) {
	t.Helper()
	var env struct {
		Meta map[string]any `json:"meta"`
	}
	if err := json.Unmarshal([]byte(w.JSON()), &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	v, ok := env.Meta["requested_time"].(string)
	return v, ok
}

func TestWithRequestedTime_NormalizesToUTC(t *testing.T) {
	hcm, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Skipf("tzdata for Asia/Ho_Chi_Minh unavailable: %v", err)
	}

	tests := []struct {
		name string
		in   time.Time
		want string
	}{
		{"utc input", time.Date(2026, 11, 22, 8, 20, 0, 0, time.UTC), "2026-11-22T08:20:00Z"},
		{"named location", time.Date(2026, 11, 22, 15, 20, 0, 0, hcm), "2026-11-22T08:20:00Z"},
		{"fixed zone +07:00", time.Date(2026, 11, 22, 15, 20, 0, 0, time.FixedZone("ICT", 7*3600)), "2026-11-22T08:20:00Z"},
		{"fixed zone +08:00", time.Date(2026, 11, 22, 16, 20, 0, 0, time.FixedZone("X", 8*3600)), "2026-11-22T08:20:00Z"},
		{"fixed zone -05:00", time.Date(2026, 11, 22, 3, 20, 0, 0, time.FixedZone("EST", -5*3600)), "2026-11-22T08:20:00Z"},
		{"crosses midnight", time.Date(2026, 11, 23, 1, 30, 0, 0, time.FixedZone("ICT", 7*3600)), "2026-11-22T18:30:00Z"},
		{"fractional seconds", time.Date(2026, 11, 22, 15, 20, 1, 123456000, time.FixedZone("ICT", 7*3600)), "2026-11-22T08:20:01.123456Z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := replify.New().OK().WithRequestedTime(tt.in)

			got, ok := requestedTimeJSON(t, w)
			if !ok {
				t.Fatal("meta.requested_time missing")
			}
			if got != tt.want {
				t.Fatalf("requested_time = %q, want %q", got, tt.want)
			}
			if strings.Contains(w.JSON(), "+07:00") || strings.Contains(w.JSON(), "+08:00") || strings.Contains(w.JSON(), "-05:00") {
				t.Fatalf("JSON contains a local offset: %s", w.JSON())
			}

			stored := w.Meta().RequestedTime()
			if !stored.Equal(tt.in) {
				t.Fatalf("stored instant = %v, want the same instant as %v", stored, tt.in)
			}
			if stored.Location() != time.UTC {
				t.Fatalf("stored location = %v, want UTC", stored.Location())
			}
		})
	}
}

func TestWithRequestedTime_ZeroValueOmitted(t *testing.T) {
	// An explicit zero time means "not set": it is not present and not serialized.
	w := replify.New().OK().WithRequestedTime(time.Time{})

	if w.Meta().IsRequestedTimePresent() {
		t.Fatal("IsRequestedTimePresent() = true for a zero time")
	}
	if got, ok := requestedTimeJSON(t, w); ok {
		t.Fatalf("requested_time = %q, want the field omitted for a zero time", got)
	}
}

func TestRequestedTime_DefaultIsUTC(t *testing.T) {
	got, ok := requestedTimeJSON(t, replify.New().OK())
	if !ok {
		t.Fatal("meta.requested_time missing on a new wrapper")
	}
	parsed, err := time.Parse(time.RFC3339Nano, got)
	if err != nil {
		t.Fatalf("requested_time %q is not RFC 3339: %v", got, err)
	}
	if !strings.HasSuffix(got, "Z") || parsed.Location() != time.UTC {
		t.Fatalf("requested_time = %q, want UTC with Z suffix", got)
	}
}

func TestRequestedTime_UnwrapNormalizesToUTC(t *testing.T) {
	const payload = `{"status_code":200,"meta":{"requested_time":"2026-11-22T15:20:00+07:00"}}`

	w, err := replify.UnwrapJSON(payload)
	if err != nil {
		t.Fatalf("UnwrapJSON() error = %v", err)
	}
	got, ok := requestedTimeJSON(t, w)
	if !ok || got != "2026-11-22T08:20:00Z" {
		t.Fatalf("requested_time = %q (present=%v), want 2026-11-22T08:20:00Z", got, ok)
	}
}
