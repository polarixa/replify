package replify_test

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/polarixa/replify"
)

var issueIDPattern = regexp.MustCompile(`^ISS-[0-9A-F]{8}$`)
var fingerprintPattern = regexp.MustCompile(`^IFP-[0-9A-F]{6}$`)

func TestNewIssueIDFormat(t *testing.T) {
	a := replify.NewIssueID()
	b := replify.NewIssueID()
	if !issueIDPattern.MatchString(a) {
		t.Fatalf("expected ID matching %s, got %q", issueIDPattern, a)
	}
	if a == b {
		t.Fatalf("expected two generated IDs to differ, got %q twice", a)
	}
}

func TestNewFingerprintDeterministic(t *testing.T) {
	fp1 := replify.NewFingerprint("panic", "UserService.GetProfile", "service/user_service.go", 81)
	fp2 := replify.NewFingerprint("panic", "UserService.GetProfile", "service/user_service.go", 81)
	if fp1 != fp2 {
		t.Fatalf("expected identical inputs to produce identical fingerprints, got %q vs %q", fp1, fp2)
	}
	if !fingerprintPattern.MatchString(fp1) {
		t.Fatalf("expected fingerprint matching %s, got %q", fingerprintPattern, fp1)
	}

	fp3 := replify.NewFingerprint("panic", "UserService.GetProfile", "service/user_service.go", 82)
	if fp1 == fp3 {
		t.Fatalf("expected different line numbers to produce different fingerprints")
	}
}

func TestNewIssueResolvesRootCause(t *testing.T) {
	wrapped := fmt.Errorf("query user: %w", sql.ErrNoRows)
	issue := replify.NewIssue(wrapped)
	if issue == nil {
		t.Fatal("expected non-nil issue")
	}
	if issue.Message != sql.ErrNoRows.Error() {
		t.Fatalf("expected message %q, got %q", sql.ErrNoRows.Error(), issue.Message)
	}
	if !issueIDPattern.MatchString(issue.ID) {
		t.Fatalf("expected ID matching %s, got %q", issueIDPattern, issue.ID)
	}
	if !fingerprintPattern.MatchString(issue.Fingerprint) {
		t.Fatalf("expected fingerprint matching %s, got %q", fingerprintPattern, issue.Fingerprint)
	}
}

func TestNewIssueNilError(t *testing.T) {
	if issue := replify.NewIssue(nil); issue != nil {
		t.Fatalf("expected nil issue for nil error, got %+v", issue)
	}
}

func TestWrappedErrorsSupportStandardErrorsAPI(t *testing.T) {
	wrapped := replify.AppendErrorAck(sql.ErrNoRows, "query user")
	if !errors.Is(wrapped, sql.ErrNoRows) {
		t.Fatal("expected errors.Is to find sql.ErrNoRows in the wrapped chain")
	}
	var target *underlyingMessageProbe
	if errors.As(wrapped, &target) {
		t.Fatal("errors.As should not match an unrelated type")
	}
	if got := errors.Unwrap(wrapped); got == nil {
		t.Fatal("expected errors.Unwrap to return the underlying underlyingMessage")
	}
}

// underlyingMessageProbe is an arbitrary type used only to exercise errors.As
// against a type that is deliberately not present in the wrapped chain.
type underlyingMessageProbe struct{ error }

func TestWrapperIssue(t *testing.T) {
	w := replify.New().WithErrorAck(sql.ErrNoRows)
	issue := w.Issue()
	if issue == nil {
		t.Fatal("expected non-nil issue from wrapper with an error")
	}
	if issue.Message != sql.ErrNoRows.Error() {
		t.Fatalf("expected message %q, got %q", sql.ErrNoRows.Error(), issue.Message)
	}

	noErr := replify.New()
	if got := noErr.Issue(); got != nil {
		t.Fatalf("expected nil issue for wrapper without an error, got %+v", got)
	}
}
