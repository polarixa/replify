package replify_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	// WriteHeader must not have been called, so the body should be empty.
	if rr.Body.Len() != 0 {
		t.Errorf("expected no body when file is missing, got %q", rr.Body.String())
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

func TestWriteBinary_FallbackOctetStream(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	replify.New().WithHeader(replify.OK).Binary([]byte("unknown")).Filename("file.unknown_xyz99").WriteBinary(rr)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/octet-stream" {
		t.Errorf("expected application/octet-stream fallback, got %q", ct)
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

func TestWrite_BinaryFilename(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	w := replify.New().WithHeader(replify.OK).Binary([]byte("data")).Filename("result.json").Write(rr)

	if w.IsErrorPresent() {
		t.Fatalf("unexpected error: %v", w.Error())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "json") {
		t.Errorf("expected JSON content type from .json filename, got %q", ct)
	}
}

func TestWrite_CustomStatusBinary(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	replify.New().WithStatusCode(http.StatusCreated).Binary([]byte("created")).Write(rr)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
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
