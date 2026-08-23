package replify

import (
	"io"
	"mime"
	"net/http"
	"os"
	gopath "path/filepath"
	"strconv"

	"github.com/polarixa/replify/pkg/strutil"
)

// WriteJSON writes the [wrapper] instance as a JSON response to the provided [http.ResponseWriter].
//
// This method checks if the [wrapper] is available and if the provided [http.ResponseWriter] is not nil.
// If the [wrapper] is not available or the [http.ResponseWriter] is nil, it returns the [wrapper] with an error acknowledgment.
// If the status code indicates "204 No Content", it writes the status code without a body.
// Otherwise, it serializes the [wrapper] to JSON, sets the appropriate Content-Type header, and writes the JSON payload to the ResponseWriter.
//
// Parameters:
//   - w: An [http.ResponseWriter] to which the JSON response will be written.
//
// Returns:
//   - A pointer to the modified [wrapper] instance, allowing for method chaining.
func (r *wrapper) WriteJSON(w http.ResponseWriter) *wrapper {
	if !r.Available() {
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("WriteJSON called with nil http.ResponseWriter"))
	}
	if r.EqualHeader(NoContent) {
		w.WriteHeader(r.StatusCode())
		return r
	}
	payload := r.JSONBytes()
	w.Header().Set(HeaderContentType.String(), MediaTypeApplicationJSONUTF8.String())
	w.WriteHeader(r.StatusCode())
	if _, err := w.Write(payload); err != nil {
		return r.WithErrorAck(err)
	}
	return r
}

// File sets the file path in the [wrapper] instance.
//
// This method allows the user to specify a file path associated with the response.
// It updates the `filepath` field of the [wrapper] instance with the provided path.
//
// Parameters:
//   - `path`: A string representing the file path to set.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (r *wrapper) File(path string) *wrapper {
	if !r.Available() {
		return r
	}
	r.filepath = path
	return r
}

// FileAttachment sets the file path and filename in the [wrapper] instance.
//
// This method allows the user to specify a file path and a filename associated with the response.
// It updates the `filepath` and `filename` fields of the [wrapper] instance with the provided values.
//
// Parameters:
//   - `path`: A string representing the file path to set.
//   - `filename`: A string representing the filename to set.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (r *wrapper) FileAttachment(path string, filename string) *wrapper {
	if !r.Available() {
		return r
	}
	r.filepath = path
	r.filename = filename
	return r
}

// Binary sets the binary data in the [wrapper] instance.
//
// This method allows the user to specify binary data associated with the response.
// It updates the `data` field of the [wrapper] instance with the provided byte slice.
//
// Parameters:
//   - `data`: A byte slice representing the binary data to set.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (r *wrapper) Binary(data []byte) *wrapper {
	if !r.Available() {
		return r
	}
	// Check if the data is binary, if not, return an error acknowledgment
	if !isBinaryValue(data) {
		r.WithErrorAck(NewErrorf("Binary: data is not binary, got %T", data))
		return r
	}

	r.data = data
	return r
}

// WithFilename sets the filename for the response in the [wrapper] instance.
//
// This method allows the user to specify a filename associated with the response.
// It updates the `filename` field of the [wrapper] instance with the provided value.
//
// Parameters:
//   - `filename`: A string representing the filename to set.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (r *wrapper) WithFilename(filename string) *wrapper {
	if !r.Available() {
		return r
	}
	r.filename = filename
	return r
}

// Write writes the response to the provided [http.ResponseWriter] based on the configuration of the [wrapper] instance.

// This method checks the availability of the [wrapper] instance and determines the appropriate response format to write.
// If a file path is set, it calls WriteFile; if binary data is set, it calls WriteBinary; otherwise, it defaults to WriteJSON.
// The method ensures that the response is written according to the specified configuration in the [wrapper] instance.
//
// Parameters:
//   - w: An [http.ResponseWriter] to which the response will be written.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (r *wrapper) Write(w http.ResponseWriter) *wrapper {
	if !r.Available() {
		return r
	}
	if strutil.IsNotEmpty(r.filepath) {
		return r.WriteFile(w)
	}
	if r.IsBinaryBody() {
		return r.WriteBinary(w)
	}
	return r.WriteJSON(w)
}

// WriteFile writes the file specified in the [wrapper] instance to the provided [http.ResponseWriter].

// This method checks the availability of the [wrapper] instance and ensures that a valid file path is set.
// It opens the specified file, retrieves its information, and writes it to the ResponseWriter with appropriate headers.
// If any errors occur during file operations, they are recorded in the [wrapper] instance.
//
// Parameters:
//   - w: An [http.ResponseWriter] to which the file response will be written.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (r *wrapper) WriteFile(w http.ResponseWriter) *wrapper {
	if !r.Available() {
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("WriteFile called with nil http.ResponseWriter"))
	}
	if strutil.IsEmpty(r.filepath) {
		return r.WithErrorAck(NewError("WriteFile called with empty filepath"))
	}
	if r.EqualHeader(NoContent) {
		w.WriteHeader(r.StatusCode())
		return r
	}
	// Open and stat before committing any response so errors can still set a non-200 status.
	f, err := os.Open(r.filepath)
	if err != nil {
		return r.WithErrorAck(err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return r.WithErrorAck(err)
	}
	if stat.IsDir() {
		return r.WithErrorAck(NewErrorf("WriteFile: %q is a directory", r.filepath))
	}
	ct := mime.TypeByExtension(gopath.Ext(r.filepath))
	if ct == "" {
		ct = MediaTypeApplicationOctetStream.String()
	}
	// All headers must be set before WriteHeader.
	w.Header().Set(HeaderContentType.String(), ct)
	if r.filename != "" {
		disp, dispErr := buildContentDisposition(r.filename)
		if dispErr != nil {
			return r.WithErrorAck(dispErr)
		}
		w.Header().Set(HeaderContentDisposition.String(), disp)
	}
	w.Header().Set(HeaderContentLength.String(), strconv.FormatInt(stat.Size(), 10))
	w.WriteHeader(r.StatusCode())
	if _, err := io.Copy(w, f); err != nil {
		return r.WithErrorAck(err)
	}
	return r
}

// WriteBinary writes the binary data stored in the wrapper to the provided http.ResponseWriter.
// The payload must have been set via Binary(); if data is not a []byte the error is recorded.
func (r *wrapper) WriteBinary(w http.ResponseWriter) *wrapper {
	if !r.Available() {
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("WriteBinary called with nil http.ResponseWriter"))
	}
	if r.EqualHeader(NoContent) {
		w.WriteHeader(r.StatusCode())
		return r
	}
	data, ok := r.data.([]byte)
	if !ok {
		return r.WithErrorAck(NewErrorf("WriteBinary: data is not []byte, got %T", r.data))
	}
	ct := ""
	if r.filename != "" {
		ct = mime.TypeByExtension(gopath.Ext(r.filename))
	}
	if ct == "" {
		ct = MediaTypeApplicationOctetStream.String()
	}
	// All headers must be set before WriteHeader.
	w.Header().Set(HeaderContentType.String(), ct)
	if r.filename != "" {
		disp, dispErr := buildContentDisposition(r.filename)
		if dispErr != nil {
			return r.WithErrorAck(dispErr)
		}
		w.Header().Set(HeaderContentDisposition.String(), disp)
	}
	w.WriteHeader(r.StatusCode())
	if _, err := w.Write(data); err != nil {
		return r.WithErrorAck(err)
	}
	return r
}

// buildContentDisposition returns a safe Content-Disposition attachment header value.
// Filenames with CR, LF, or null bytes are rejected to prevent header injection.
// All other characters are properly encoded by the mime package (RFC 5987 for non-ASCII).
func buildContentDisposition(filename string) (string, error) {
	for i, c := range filename {
		if c == '\r' || c == '\n' || c == 0 {
			return "", NewErrorf("filename contains unsafe character at byte position %d", i)
		}
	}
	return mime.FormatMediaType("attachment", map[string]string{"filename": filename}), nil
}
