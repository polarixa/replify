package replify_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/polarixa/replify"
)

// The Dump*To family must return a Dump whose in-process stream is still open,
// so callers can re-read or rewind it before calling Close themselves.
func TestDumpTo_StreamRemainsReadable(t *testing.T) {
	w := replify.New().OK().WithBody(map[string]any{"id": 123})

	tests := []struct {
		name string
		dump func(dst string) (*replify.Dump, interface{ IsError() bool })
		want string
	}{
		{"DumpJSONTo", func(dst string) (*replify.Dump, interface{ IsError() bool }) { return w.DumpJSONTo(dst) }, `"id":123`},
		{"DumpBodyTo", func(dst string) (*replify.Dump, interface{ IsError() bool }) { return w.DumpBodyTo(dst) }, `"id":123`},
		{"DumpMDDocTo", func(dst string) (*replify.Dump, interface{ IsError() bool }) { return w.DumpMDDocTo(dst) }, "Diagnostic Report"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, res := tt.dump(filepath.Join(t.TempDir(), "out.dump"))
			if res.IsError() {
				t.Fatalf("dump failed")
			}
			defer d.Close()

			if err := d.Rewind(); err != nil {
				t.Fatalf("Rewind() = %v, want nil", err)
			}
			if got := d.ReadAll(); !strings.Contains(got, tt.want) {
				t.Fatalf("ReadAll() = %q, want it to contain %q", got, tt.want)
			}
		})
	}
}

// An error response must keep its error text after a JSON round trip.
func TestUnwrapJSON_RestoresErrorFromIssue(t *testing.T) {
	src := replify.New().Conflict().WithError("email taken")

	got, err := replify.UnwrapJSON(src.JSON())
	if err != nil {
		t.Fatalf("UnwrapJSON() error = %v", err)
	}
	if got.Error() != "email taken" {
		t.Fatalf("Error() = %q, want %q", got.Error(), "email taken")
	}
	if got.StatusCode() != 409 {
		t.Fatalf("StatusCode() = %d, want 409", got.StatusCode())
	}
}
