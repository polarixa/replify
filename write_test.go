package replify_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/polarixa/replify"
)

// failingWriter is a minimal http.ResponseWriter whose Write always returns an error.
type failingWriter struct {
	header   http.Header
	writeErr error
}

func newFailingWriter(err error) *failingWriter {
	return &failingWriter{header: make(http.Header), writeErr: err}
}

func (f *failingWriter) Header() http.Header       { return f.header }
func (f *failingWriter) WriteHeader(int)           {}
func (f *failingWriter) Write([]byte) (int, error) { return 0, f.writeErr }

// TestWrite_SuccessfulResponse verifies status, body, and Content-Type for a 200 OK.
func TestWrite_SuccessfulResponse(t *testing.T) {
	t.Parallel()

	w := replify.New().
		WithHeader(replify.OK).
		WithBody(map[string]string{"id": "123", "name": "John Doe"}).
		WithMessage("User retrieved successfully")

	want := w.JSON() // snapshot before Write commits the response

	rr := httptest.NewRecorder()
	result := w.WriteJSON(rr)
	if result.IsErrorPresent() {
		t.Fatalf("Write stored unexpected error: %v", result.Error())
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("unexpected Content-Type: %q", got)
	}
	if rr.Body.String() != want {
		t.Errorf("body does not match JSON()\ngot:  %s\nwant: %s", rr.Body.String(), want)
	}
}

// TestWrite_ErrorResponse verifies a 404 Not Found response.
func TestWrite_ErrorResponse(t *testing.T) {
	t.Parallel()

	w := replify.New().
		WithHeader(replify.NotFound).
		WithMessage("resource not found")

	want := w.JSON()

	rr := httptest.NewRecorder()
	result := w.WriteJSON(rr)
	if result.IsErrorPresent() {
		t.Fatalf("Write stored unexpected error: %v", result.Error())
	}
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("unexpected Content-Type: %q", got)
	}
	if rr.Body.String() != want {
		t.Errorf("body does not match JSON()\ngot:  %s\nwant: %s", rr.Body.String(), want)
	}
}

// TestWrite_CustomStatus verifies that an explicitly configured status code is preserved.
func TestWrite_CustomStatus(t *testing.T) {
	t.Parallel()

	codes := []int{
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNoContent,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusInternalServerError,
	}

	for _, code := range codes {
		code := code
		t.Run(http.StatusText(code), func(t *testing.T) {
			t.Parallel()

			w := replify.New().WithStatusCode(code).WithMessage("test")
			rr := httptest.NewRecorder()
			w.WriteJSON(rr)
			if rr.Code != code {
				t.Errorf("expected status %d, got %d", code, rr.Code)
			}
		})
	}
}

// TestWrite_EmptyBody verifies that Write preserves existing JSON() semantics for a nil body.
func TestWrite_EmptyBody(t *testing.T) {
	t.Parallel()

	w := replify.New().WithHeader(replify.OK).WithMessage("no data")
	want := w.JSON()

	rr := httptest.NewRecorder()
	w.WriteJSON(rr)
	if rr.Body.String() != want {
		t.Errorf("body does not match JSON()\ngot:  %s\nwant: %s", rr.Body.String(), want)
	}
}

// TestWrite_WrapperReturnValue verifies that Write returns the same wrapper instance.
func TestWrite_WrapperReturnValue(t *testing.T) {
	t.Parallel()

	w := replify.New().WithHeader(replify.OK).WithMessage("same pointer")
	rr := httptest.NewRecorder()
	returned := w.WriteJSON(rr)
	if returned != w {
		t.Error("Write did not return the same wrapper instance")
	}
}

// TestWrite_ChainCompatibility verifies that Write participates in the fluent chain.
func TestWrite_ChainCompatibility(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	w := replify.New().
		WithHeader(replify.OK).
		WithBody(map[string]string{"key": "value"}).
		WriteJSON(rr) // chain operation — Write returns *wrapper

	if w == nil {
		t.Fatal("Write returned nil")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	// Wrapper remains usable after Write.
	if w.StatusCode() != http.StatusOK {
		t.Errorf("wrapper no longer holds status after Write")
	}
}

// TestWrite_ErrorPropagation verifies that a write failure is stored in the wrapper.
func TestWrite_ErrorPropagation(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("disk full")
	fw := newFailingWriter(sentinel)

	w := replify.New().WithHeader(replify.OK).WithMessage("test")
	result := w.WriteJSON(fw)

	if !result.IsErrorPresent() {
		t.Fatal("expected wrapper to hold a write error")
	}
	// errors.Is traverses Unwrap() — underlyingStack wraps the sentinel.
	if !errors.Is(result.Cause(), sentinel) {
		t.Errorf("expected sentinel in error chain, got: %v", result.Cause())
	}
}

// TestWrite_HeaderBeforeBody verifies that Content-Type is committed on the recorded response.
func TestWrite_HeaderBeforeBody(t *testing.T) {
	t.Parallel()

	w := replify.New().WithHeader(replify.OK).WithMessage("test")
	rr := httptest.NewRecorder()
	w.WriteJSON(rr)
	if got := rr.Result().Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type not set correctly: %q", got)
	}
}

// TestWrite_NilWriter verifies that a nil ResponseWriter stores an error in the wrapper.
func TestWrite_NilWriter(t *testing.T) {
	t.Parallel()

	w := replify.New().WithHeader(replify.OK)
	result := w.WriteJSON(nil)
	if !result.IsErrorPresent() {
		t.Fatal("expected wrapper to hold an error for nil ResponseWriter")
	}
}

// TestWrite_ThroughR verifies that Write is accessible via the public R type.
func TestWrite_ThroughR(t *testing.T) {
	t.Parallel()

	r := replify.New().
		WithHeader(replify.OK).
		WithMessage("via R").
		Reply()

	rr := httptest.NewRecorder()
	r.WriteJSON(rr)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
