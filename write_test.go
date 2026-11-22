package replify_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	// A 404 with a message auto-derives its error from the message (see autoAdjust);
	// only a mismatched or write-failure error should fail this test.
	if result.IsErrorPresent() && result.Error() != "resource not found" {
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

// --- helpers ----------------------------------------------------------------

// writeTempFile creates a temporary file with the given content and returns its path.
func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return path
}

// --- WriteFile tests ---------------------------------------------------------

func TestWriteFile_SuccessfulResponse(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "hello.txt", "hello world")

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).File(path).WriteFile(rr)

	if w.IsErrorPresent() {
		t.Fatalf("unexpected error: %v", w.Error())
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if got := rr.Body.String(); got != "hello world" {
		t.Errorf("body = %q, want %q", got, "hello world")
	}
}

func TestWriteFile_CorrectStatusCode(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "data.txt", "data")

	rr := httptest.NewRecorder()
	replify.New().WithStatusCode(http.StatusPartialContent).File(path).WriteFile(rr)

	if rr.Code != http.StatusPartialContent {
		t.Errorf("expected %d, got %d", http.StatusPartialContent, rr.Code)
	}
}

func TestWriteFile_ExactBody(t *testing.T) {
	t.Parallel()
	content := "exact binary content \x00\x01\x02"
	path := writeTempFile(t, "blob.bin", content)

	rr := httptest.NewRecorder()
	replify.New().WithHeader(replify.OK).File(path).WriteFile(rr)

	if got := rr.Body.String(); got != content {
		t.Errorf("body does not match: got len=%d want len=%d", len(got), len(content))
	}
}

func TestWriteFile_CorrectContentType(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "doc.pdf", "%PDF-1.4")

	rr := httptest.NewRecorder()
	replify.New().WithHeader(replify.OK).File(path).WriteFile(rr)

	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/pdf") {
		t.Errorf("unexpected Content-Type %q, want application/pdf", ct)
	}
}

func TestWriteFile_UnknownExtension(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "file.unknown_xyz", "bytes")

	rr := httptest.NewRecorder()
	replify.New().WithHeader(replify.OK).File(path).WriteFile(rr)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/octet-stream" {
		t.Errorf("expected application/octet-stream for unknown extension, got %q", ct)
	}
}

func TestWriteFile_NonexistentFile(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).File("/nonexistent/path/that/does/not/exist.txt").WriteFile(rr)

	if !w.IsErrorPresent() {
		t.Error("expected error for nonexistent file, got none")
	}
	// The open failure is answered with a JSON error response (status 500), not an
	// empty implicit 200; see TestWriteFile_OpenErrorWritesJSON.
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when file is missing, got %d", rr.Code)
	}
	if rr.Body.Len() == 0 {
		t.Error("expected a JSON error body when file is missing, got none")
	}
}

func TestWriteFile_EmptyFilepath(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).WriteFile(rr)

	if !w.IsErrorPresent() {
		t.Error("expected error for empty filepath")
	}
}

func TestWriteFile_NilWriter(t *testing.T) {
	t.Parallel()

	w := replify.New().WithHeader(replify.OK).File("/tmp/file.txt").WriteFile(nil)
	if !w.IsErrorPresent() {
		t.Error("expected error for nil writer")
	}
}

func TestWriteFile_CustomHeaders(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "data.json", `{"ok":true}`)

	rr := httptest.NewRecorder()
	rr.Header().Set("X-Custom", "value")
	replify.New().WithHeader(replify.OK).File(path).WriteFile(rr)

	if got := rr.Header().Get("X-Custom"); got != "value" {
		t.Errorf("custom header lost: got %q", got)
	}
}

func TestWriteFile_FilenameAttachment(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "report.pdf", "%PDF-1.4")

	rr := httptest.NewRecorder()
	replify.New().WithHeader(replify.OK).FileAttachment(path, "report.pdf").WriteFile(rr)

	disp := rr.Header().Get("Content-Disposition")
	if !strings.Contains(disp, "attachment") {
		t.Errorf("Content-Disposition missing attachment: %q", disp)
	}
	if !strings.Contains(disp, "report.pdf") {
		t.Errorf("Content-Disposition missing filename: %q", disp)
	}
}

func TestWriteFile_NoContent(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "file.txt", "content")

	rr := httptest.NewRecorder()
	replify.New().WithStatusCode(http.StatusNoContent).File(path).WriteFile(rr)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Errorf("204 response must not have a body, got %q", rr.Body.String())
	}
}

// --- FileAttachment tests ----------------------------------------------------

func TestFileAttachment_Body(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "doc.pdf", "%PDF")

	rr := httptest.NewRecorder()
	replify.New().WithHeader(replify.OK).FileAttachment(path, "mydoc.pdf").WriteFile(rr)

	if rr.Body.String() != "%PDF" {
		t.Errorf("unexpected body: %q", rr.Body.String())
	}
}

func TestFileAttachment_ContentDisposition(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "report.pdf", "PDF")

	rr := httptest.NewRecorder()
	replify.New().WithHeader(replify.OK).FileAttachment(path, "report.pdf").WriteFile(rr)

	disp := rr.Header().Get("Content-Disposition")
	if !strings.HasPrefix(disp, "attachment") {
		t.Errorf("Content-Disposition should start with attachment, got %q", disp)
	}
}

func TestFileAttachment_Filename(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "x.pdf", "data")

	rr := httptest.NewRecorder()
	replify.New().WithHeader(replify.OK).FileAttachment(path, "output.pdf").WriteFile(rr)

	disp := rr.Header().Get("Content-Disposition")
	if !strings.Contains(disp, "output.pdf") {
		t.Errorf("Content-Disposition does not contain filename: %q", disp)
	}
}

func TestFileAttachment_MIMEType(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "data.pdf", "PDF")

	rr := httptest.NewRecorder()
	replify.New().WithHeader(replify.OK).FileAttachment(path, "data.pdf").WriteFile(rr)

	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/pdf") {
		t.Errorf("unexpected Content-Type %q, want application/pdf", ct)
	}
}

func TestFileAttachment_UnsafeFilename_CR(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "file.txt", "content")

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).FileAttachment(path, "evil\rname.txt").WriteFile(rr)

	if !w.IsErrorPresent() {
		t.Error("expected error for filename with CR")
	}
}

func TestFileAttachment_UnsafeFilename_LF(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "file.txt", "content")

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).FileAttachment(path, "evil\nname.txt").WriteFile(rr)

	if !w.IsErrorPresent() {
		t.Error("expected error for filename with LF")
	}
}

func TestFileAttachment_UnsafeFilename_Null(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "file.txt", "content")

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).FileAttachment(path, "evil\x00name.txt").WriteFile(rr)

	if !w.IsErrorPresent() {
		t.Error("expected error for filename with null byte")
	}
}

func TestFileAttachment_NonexistentFile(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).FileAttachment("/no/such/file.pdf", "out.pdf").WriteFile(rr)

	if !w.IsErrorPresent() {
		t.Error("expected error for nonexistent file")
	}
}

func TestFileAttachment_NilWriter(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "doc.pdf", "PDF")

	w := replify.New().WithHeader(replify.OK).FileAttachment(path, "doc.pdf").WriteFile(nil)
	if !w.IsErrorPresent() {
		t.Error("expected error for nil writer")
	}
}

// --- Binary tests ------------------------------------------------------------

func TestWriteBinary_Successful(t *testing.T) {
	t.Parallel()

	data := []byte("binary payload")
	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).Binary(data).WriteBinary(rr)

	if w.IsErrorPresent() {
		t.Fatalf("unexpected error: %v", w.Error())
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestWriteBinary_ExactBytes(t *testing.T) {
	t.Parallel()

	data := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
	rr := httptest.NewRecorder()
	replify.New().WithHeader(replify.OK).Binary(data).WriteBinary(rr)

	if got := rr.Body.Bytes(); string(got) != string(data) {
		t.Errorf("body mismatch: got %v, want %v", got, data)
	}
}

func TestWriteBinary_ContentType_Filename(t *testing.T) {
	t.Parallel()

	data := []byte("%PDF-1.4")
	rr := httptest.NewRecorder()
	replify.New().WithHeader(replify.OK).Binary(data).Filename("report.pdf").WriteBinary(rr)

	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/pdf") {
		t.Errorf("unexpected Content-Type %q, want application/pdf", ct)
	}
}

func TestWriteBinary_ContentType_WithoutFilename(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	replify.New().WithHeader(replify.OK).Binary([]byte("raw")).WriteBinary(rr)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/octet-stream" {
		t.Errorf("expected application/octet-stream, got %q", ct)
	}
}

func TestWriteBinary_ContentDisposition_Filename(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	replify.New().WithHeader(replify.OK).Binary([]byte("data")).Filename("result.bin").WriteBinary(rr)

	disp := rr.Header().Get("Content-Disposition")
	if !strings.Contains(disp, "attachment") {
		t.Errorf("Content-Disposition missing attachment: %q", disp)
	}
	if !strings.Contains(disp, "result.bin") {
		t.Errorf("Content-Disposition missing filename: %q", disp)
	}
}

func TestWriteBinary_InvalidDataType(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).WithBody("not bytes").WriteBinary(rr)

	if !w.IsErrorPresent() {
		t.Error("expected error when data is not []byte")
	}
}

func TestWriteBinary_EmptyByteSlice(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).Binary([]byte{}).WriteBinary(rr)

	if w.IsErrorPresent() {
		t.Fatalf("unexpected error: %v", w.Error())
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Error("expected empty body")
	}
}

func TestWriteBinary_NilByteSlice(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).Binary(nil).WriteBinary(rr)

	if w.IsErrorPresent() {
		t.Fatalf("unexpected error: %v", w.Error())
	}
	if rr.Body.Len() != 0 {
		t.Error("expected empty body for nil slice")
	}
}

func TestWriteBinary_NilWriter(t *testing.T) {
	t.Parallel()

	w := replify.New().WithHeader(replify.OK).Binary([]byte("data")).WriteBinary(nil)
	if !w.IsErrorPresent() {
		t.Error("expected error for nil writer")
	}
}

func TestWriteBinary_NoContent(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	replify.New().WithStatusCode(http.StatusNoContent).Binary([]byte("should not appear")).WriteBinary(rr)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Errorf("204 must not have a body, got %q", rr.Body.String())
	}
}

// --- Write dispatch tests ----------------------------------------------------

func TestWrite_DispatchToFile(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "hello.txt", "file content")

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).File(path).Write(rr)

	if w.IsErrorPresent() {
		t.Fatalf("unexpected error: %v", w.Error())
	}
	if got := rr.Body.String(); got != "file content" {
		t.Errorf("expected file content, got %q", got)
	}
	// Must not have JSON Content-Type.
	ct := rr.Header().Get("Content-Type")
	if ct == "application/json; charset=utf-8" {
		t.Error("Write dispatched to WriteJSON instead of WriteFile")
	}
}

func TestWrite_DispatchToBinary(t *testing.T) {
	t.Parallel()

	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).Binary(data).Write(rr)

	if w.IsErrorPresent() {
		t.Fatalf("unexpected error: %v", w.Error())
	}
	if got := rr.Body.Bytes(); string(got) != string(data) {
		t.Errorf("body mismatch: got %v, want %v", got, data)
	}
	ct := rr.Header().Get("Content-Type")
	if ct == "application/json; charset=utf-8" {
		t.Error("Write dispatched to WriteJSON instead of WriteBinary")
	}
}

func TestWrite_DispatchToJSON(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	w := replify.New().
		WithHeader(replify.OK).
		WithBody(map[string]string{"hello": "world"}).
		Write(rr)

	if w.IsErrorPresent() {
		t.Fatalf("unexpected error: %v", w.Error())
	}
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("expected JSON Content-Type, got %q", ct)
	}
}

func TestWrite_NilWriter_Binary(t *testing.T) {
	t.Parallel()

	w := replify.New().WithHeader(replify.OK).Binary([]byte("x")).Write(nil)
	if !w.IsErrorPresent() {
		t.Error("expected error for nil writer")
	}
}

func TestWrite_FileAttachmentChain(t *testing.T) {
	t.Parallel()
	path := writeTempFile(t, "report.pdf", "PDF content")

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).FileAttachment(path, "report.pdf").Write(rr)

	if w.IsErrorPresent() {
		t.Fatalf("unexpected error: %v", w.Error())
	}
	disp := rr.Header().Get("Content-Disposition")
	if !strings.Contains(disp, "attachment") {
		t.Errorf("Content-Disposition missing attachment: %q", disp)
	}
}

// --- HTTP semantics ----------------------------------------------------------

func TestWriteJSON_NoContent(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	replify.New().WithStatusCode(http.StatusNoContent).WriteJSON(rr)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Errorf("204 WriteJSON must not write a body, got %q", rr.Body.String())
	}
}

func TestWriteJSON_NoContent_NoContentType(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	replify.New().WithStatusCode(http.StatusNoContent).WriteJSON(rr)

	// Content-Type should not be set for a 204.
	if ct := rr.Header().Get("Content-Type"); ct != "" {
		t.Errorf("204 should not set Content-Type, got %q", ct)
	}
}

func TestWriteFile_StatusCodeRespected(t *testing.T) {
	t.Parallel()
	codes := []int{http.StatusOK, http.StatusCreated, http.StatusPartialContent}
	for _, code := range codes {
		code := code
		t.Run(http.StatusText(code), func(t *testing.T) {
			t.Parallel()
			path := writeTempFile(t, "f.txt", "x")
			rr := httptest.NewRecorder()
			replify.New().WithStatusCode(code).File(path).WriteFile(rr)
			if rr.Code != code {
				t.Errorf("expected %d, got %d", code, rr.Code)
			}
		})
	}
}

func TestWriteBinary_UnsafeFilename_CRLF(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).
		Binary([]byte("data")).
		Filename("inject\r\nX-Hdr: val").
		WriteBinary(rr)

	if !w.IsErrorPresent() {
		t.Error("expected error for CRLF in filename")
	}
	// The injected header must not appear.
	if rr.Header().Get("X-Hdr") != "" {
		t.Error("header injection succeeded — unsafe filename was accepted")
	}
}

// TestWrite_ReturnsSameWrapper ensures Write returns the same *wrapper instance.
func TestWrite_ReturnsSameWrapper(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	original := replify.New().WithHeader(replify.OK).WithBody(map[string]string{"k": "v"})
	returned := original.Write(rr)
	if returned != original {
		t.Error("Write did not return the same wrapper instance")
	}
}

// --- HTTP writer defect regressions -------------------------------------------

// commitTrackingWriter wraps a recorder and counts WriteHeader and Write calls.
// Write call number failOnWrite (1-based) fails; when failPartial is true that
// call first writes half of its input, simulating a mid-stream failure. Every
// other Write passes through, so any later write by the library is visible.
type commitTrackingWriter struct {
	rec         *httptest.ResponseRecorder
	failOnWrite int
	failPartial bool

	headerCalls int
	writeCalls  int
}

func newCommitTrackingWriter(failOnWrite int, partial bool) *commitTrackingWriter {
	return &commitTrackingWriter{rec: httptest.NewRecorder(), failOnWrite: failOnWrite, failPartial: partial}
}

func (c *commitTrackingWriter) Header() http.Header { return c.rec.Header() }

func (c *commitTrackingWriter) WriteHeader(code int) {
	c.headerCalls++
	c.rec.WriteHeader(code)
}

func (c *commitTrackingWriter) Write(p []byte) (int, error) {
	c.writeCalls++
	if c.writeCalls == c.failOnWrite {
		if c.failPartial {
			n, _ := c.rec.Write(p[:len(p)/2])
			return n, errors.New("simulated connection reset")
		}
		return 0, errors.New("simulated connection reset")
	}
	return c.rec.Write(p)
}

// T1 (Bug 2): failing to open the file must produce a JSON error response.
func TestWriteFile_OpenErrorWritesJSON(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).File("/nonexistent/path/that/does/not/exist.txt").WriteFile(rr)

	if !w.IsErrorPresent() {
		t.Error("expected the open error to be recorded on the wrapper")
	}
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("expected a JSON content type, got %q", ct)
	}
	if !strings.Contains(rr.Body.String(), `"status_code":500`) {
		t.Errorf("expected a JSON error envelope, got %q", rr.Body.String())
	}
}

// T2 (Bug 3): a nil writer must never panic, including for a NoContent response.
func TestWriteWriters_NilWriterNoContentDoesNotPanic(t *testing.T) {
	t.Parallel()

	cfg := replify.NewSignatureConfig("example-secret")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Each case builds a fresh 204 wrapper and reports whether an error was recorded.
	cases := []struct {
		name  string
		write func(rw http.ResponseWriter) bool
	}{
		{"WriteFile", func(rw http.ResponseWriter) bool {
			return replify.New().WithStatusCode(http.StatusNoContent).WriteFile(rw).IsErrorPresent()
		}},
		{"WriteBinary", func(rw http.ResponseWriter) bool {
			return replify.New().WithStatusCode(http.StatusNoContent).Binary([]byte("ignored")).WriteBinary(rw).IsErrorPresent()
		}},
		{"WriteJSON", func(rw http.ResponseWriter) bool {
			return replify.New().WithStatusCode(http.StatusNoContent).WriteJSON(rw).IsErrorPresent()
		}},
		{"WriteJSONSignature", func(rw http.ResponseWriter) bool {
			return replify.New().WithStatusCode(http.StatusNoContent).WriteJSONSignature(rw, cfg).IsErrorPresent()
		}},
		{"WriteJSONSignatureHeader", func(rw http.ResponseWriter) bool {
			return replify.New().WithStatusCode(http.StatusNoContent).WriteJSONSignatureHeader(rw, cfg).IsErrorPresent()
		}},
		{"WriteJSONSignatureFromRequest", func(rw http.ResponseWriter) bool {
			return replify.New().WithStatusCode(http.StatusNoContent).WriteJSONSignatureFromRequest(rw, req, cfg).IsErrorPresent()
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/nil_writer", func(t *testing.T) {
			t.Parallel()
			defer func() {
				if p := recover(); p != nil {
					t.Fatalf("%s panicked with a nil writer: %v", tc.name, p)
				}
			}()
			if !tc.write(nil) {
				t.Errorf("%s: expected an error to be recorded for a nil writer", tc.name)
			}
		})
		t.Run(tc.name+"/valid_writer", func(t *testing.T) {
			t.Parallel()
			rr := httptest.NewRecorder()
			tc.write(rr)
			if rr.Code != http.StatusNoContent {
				t.Errorf("%s: expected 204, got %d", tc.name, rr.Code)
			}
			if rr.Body.Len() != 0 {
				t.Errorf("%s: 204 must not have a body, got %q", tc.name, rr.Body.String())
			}
		})
	}
}

// T3 (Bug 1): once the response is committed, a body-write failure must only be
// recorded; it must not append a JSON envelope or issue a second WriteHeader.
func TestWriteWriters_PostCommitFailureOnlyRecorded(t *testing.T) {
	t.Parallel()

	large := strings.Repeat("0123456789abcdef", 8*1024) // 128 KiB: spans several copy chunks

	t.Run("WriteFile", func(t *testing.T) {
		t.Parallel()
		path := writeTempFile(t, "big.bin", large)
		cw := newCommitTrackingWriter(2, false) // first chunk succeeds, second fails

		w := replify.New().WithHeader(replify.Accepted).File(path).WriteFile(cw)

		if cw.headerCalls != 1 {
			t.Errorf("expected exactly one WriteHeader call, got %d", cw.headerCalls)
		}
		if cw.rec.Code != http.StatusAccepted {
			t.Errorf("expected the committed status 202 to stand, got %d", cw.rec.Code)
		}
		if got := cw.rec.Body.String(); got != large[:len(got)] || len(got) >= len(large) || strings.Contains(got, `"status_code"`) {
			t.Errorf("body must be the truncated file bytes only, got %d bytes (JSON appended: %v)", len(got), strings.Contains(got, `"status_code"`))
		}
		if cw.writeCalls != 2 {
			t.Errorf("expected no write after the failure, got %d write calls", cw.writeCalls)
		}
		if !w.IsErrorPresent() {
			t.Error("expected the write error to be recorded on the wrapper")
		}
	})

	t.Run("WriteBinary", func(t *testing.T) {
		t.Parallel()
		cw := newCommitTrackingWriter(1, true) // the single Write is partial, then fails

		w := replify.New().WithHeader(replify.Accepted).Binary([]byte(large)).WriteBinary(cw)

		if cw.headerCalls != 1 {
			t.Errorf("expected exactly one WriteHeader call, got %d", cw.headerCalls)
		}
		if cw.rec.Code != http.StatusAccepted {
			t.Errorf("expected the committed status 202 to stand, got %d", cw.rec.Code)
		}
		if got := cw.rec.Body.String(); got != large[:len(large)/2] {
			t.Errorf("body must be exactly the bytes written before the failure (%d), got %d bytes", len(large)/2, len(got))
		}
		if cw.writeCalls != 1 {
			t.Errorf("expected no write after the failure, got %d write calls", cw.writeCalls)
		}
		if !w.IsErrorPresent() {
			t.Error("expected the write error to be recorded on the wrapper")
		}
	})
}

// T4 (Bug 4a): Content-Length must always be a plain decimal number.
func TestWriteWriters_ContentLengthIsDigits(t *testing.T) {
	t.Parallel()

	digits := func(s string) bool {
		if s == "" {
			return false
		}
		for _, c := range s {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}

	t.Run("WriteFile", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		replify.New().WithHeader(replify.OK).File(writeTempFile(t, "a.txt", "hello")).WriteFile(rr)
		if got := rr.Header().Get("Content-Length"); !digits(got) || got != "5" {
			t.Errorf("expected Content-Length 5, got %q", got)
		}
	})
	t.Run("WriteFile/empty file", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		replify.New().WithHeader(replify.OK).File(writeTempFile(t, "empty.txt", "")).WriteFile(rr)
		if got := rr.Header().Get("Content-Length"); got != "0" {
			t.Errorf("expected Content-Length 0, got %q", got)
		}
	})
	t.Run("WriteBinary/empty", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		replify.New().WithHeader(replify.OK).Binary([]byte{}).WriteBinary(rr)
		if got := rr.Header().Get("Content-Length"); got != "0" {
			t.Errorf("expected Content-Length 0, got %q", got)
		}
	})
	t.Run("WriteBinary", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		replify.New().WithHeader(replify.OK).Binary([]byte("abc")).WriteBinary(rr)
		if got := rr.Header().Get("Content-Length"); !digits(got) || got != "3" {
			t.Errorf("expected Content-Length 3, got %q", got)
		}
	})
}

// T5 (Bug 4b): when Content-Disposition cannot be assembled, the JSON fallback
// must not carry the file's or binary's Content-Length or Content-Disposition.
func TestWriteWriters_DispositionFailureFallbackHasCleanHeaders(t *testing.T) {
	t.Parallel()

	payload := strings.Repeat("x", 4096) // much larger than the JSON error body

	cases := []struct {
		name  string
		write func(w http.ResponseWriter)
	}{
		{"WriteFile", func(w http.ResponseWriter) {
			path := filepath.Join(t.TempDir(), "file.bin")
			if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
				t.Errorf("write temp file: %v", err)
				return
			}
			replify.New().WithHeader(replify.OK).FileAttachment(path, "evil\r\nname.bin").WriteFile(w)
		}},
		{"WriteBinary", func(w http.ResponseWriter) {
			replify.New().WithHeader(replify.OK).Binary([]byte(payload)).Filename("evil\r\nname.bin").WriteBinary(w)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Header-level view.
			rr := httptest.NewRecorder()
			tc.write(rr)
			if got := rr.Header().Get("Content-Length"); got == "4096" {
				t.Errorf("stale Content-Length %q leaked into the JSON fallback", got)
			}
			if got := rr.Header().Get("Content-Disposition"); got != "" {
				t.Errorf("stale Content-Disposition %q leaked into the JSON fallback", got)
			}

			// Wire-level view: a real server enforces Content-Length.
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { tc.write(w) }))
			defer srv.Close()
			resp, err := http.Get(srv.URL)
			if err != nil {
				t.Fatalf("GET: %v", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("reading the response body failed (Content-Length does not match the body): %v", err)
			}
			if resp.StatusCode != http.StatusInternalServerError {
				t.Errorf("expected 500, got %d", resp.StatusCode)
			}
			// Either an exact Content-Length or chunked encoding (-1) is correct; a stale
			// file/binary length is not.
			if resp.ContentLength != -1 && resp.ContentLength != int64(len(body)) {
				t.Errorf("Content-Length %d does not match the %d-byte JSON body", resp.ContentLength, len(body))
			}
			if !strings.Contains(string(body), `"status_code":500`) {
				t.Errorf("expected a JSON error envelope, got %q", body)
			}
		})
	}
}

// T6 guards: behavior that must not change.
func TestWriteWriters_HappyPathGuards(t *testing.T) {
	t.Parallel()

	t.Run("WriteFile", func(t *testing.T) {
		t.Parallel()
		const content = "col1,col2\n1,2\n"
		path := writeTempFile(t, "report.csv", content)
		rr := httptest.NewRecorder()
		replify.New().WithHeader(replify.OK).FileAttachment(path, "report.csv").WriteFile(rr)

		if rr.Code != http.StatusOK || rr.Body.String() != content {
			t.Errorf("unexpected response: %d %q", rr.Code, rr.Body.String())
		}
		if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "csv") {
			t.Errorf("expected a CSV content type, got %q", ct)
		}
		if got, want := rr.Header().Get("Content-Length"), strconv.Itoa(len(content)); got != want {
			t.Errorf("Content-Length = %q, want %q", got, want)
		}
		if cd := rr.Header().Get("Content-Disposition"); !strings.Contains(cd, `filename=report.csv`) {
			t.Errorf("Content-Disposition = %q", cd)
		}
		if rr.Header().Get("X-Request-Id") == "" {
			t.Error("expected metadata headers to be attached")
		}
	})
	t.Run("WriteBinary", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		replify.New().WithHeader(replify.OK).Binary([]byte("PNGDATA")).Filename("image.png").WriteBinary(rr)

		if rr.Code != http.StatusOK || rr.Body.String() != "PNGDATA" {
			t.Errorf("unexpected response: %d %q", rr.Code, rr.Body.String())
		}
		if ct := rr.Header().Get("Content-Type"); ct != "image/png" {
			t.Errorf("Content-Type = %q, want image/png", ct)
		}
		if rr.Header().Get("Content-Length") != "7" {
			t.Errorf("Content-Length = %q, want 7", rr.Header().Get("Content-Length"))
		}
		if cd := rr.Header().Get("Content-Disposition"); !strings.Contains(cd, "image.png") {
			t.Errorf("Content-Disposition = %q", cd)
		}
	})
	t.Run("WriteJSON", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		replify.New().OK().WithMessage("fine").WriteJSON(rr)

		if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"message":"fine"`) {
			t.Errorf("unexpected response: %d %q", rr.Code, rr.Body.String())
		}
		if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("Content-Type = %q", ct)
		}
		for _, h := range []string{"X-Locale", "X-Api-Version", "X-Request-Id", "X-Request-Time"} {
			if rr.Header().Get(h) == "" {
				t.Errorf("expected metadata header %s", h)
			}
		}
	})
}
