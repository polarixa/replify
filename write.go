package replify

import (
	"net/http"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strutil"
	"github.com/polarixa/replify/pkg/sysx"
)

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

// Filename sets the filename for the response in the [wrapper] instance.
//
// This method allows the user to specify a filename associated with the response.
// It updates the `filename` field of the [wrapper] instance with the provided value.
//
// Parameters:
//   - `filename`: A string representing the filename to set.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (r *wrapper) Filename(filename string) *wrapper {
	if !r.Available() {
		return r
	}
	r.filename = filename
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

// WriteFile writes the file specified in the [wrapper] instance to the provided [http.ResponseWriter].
//
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

	// Use a deferred span for logging the WriteFile operation with the file path.
	defer New().Processing().Span("WriteFile", slogger.String("filepath", r.filepath))()

	// Open the file specified in the [wrapper] instance.
	resource, err := sysx.NewResource().FromPath(r.filepath)
	if err != nil {
		return r.WithErrorAck(err)
	}
	if resource.IsDir() {
		return r.WithErrorAck(NewErrorf("WriteFile: '%s' is a directory", r.filepath))
	}
	defer resource.Close() // Ensure the resource is closed after writing.

	// Set the Content-Type header based on the resource's content type.
	w.Header().Set(HeaderContentType.String(), resource.ContentType())

	// Set the Content-Length header based on the resource's size.
	w.Header().Set(HeaderContentLength.String(), conv.StringOrDefault(resource.Size(), "0B"))

	// If a filename is specified, set the Content-Disposition header to indicate an attachment with the given filename.
	if strutil.IsNotEmpty(r.filename) {
		name, err := assembleContentDisposition(r.filename)
		if err != nil {
			return r.WithErrorAck(err)
		}
		w.Header().Set(HeaderContentDisposition.String(), name)
	}

	// Write the status code to the ResponseWriter.
	w.WriteHeader(r.StatusCode())

	// Copy the resource's content to the ResponseWriter.
	// If an error occurs during the copy operation, it is recorded in the [wrapper] instance.
	if _, err := resource.CopyTo(w); err != nil {
		return r.WithErrorAck(err)
	}
	return r
}

// WriteBinary writes the binary data specified in the [wrapper] instance to the provided [http.ResponseWriter].
//
// This method checks the availability of the [wrapper] instance and ensures that binary data is set.
// It sets the appropriate headers, including Content-Type and Content-Length, and writes the binary data to the ResponseWriter.
// If any errors occur during writing, they are recorded in the [wrapper] instance.
//
// Parameters:
//   - w: An [http.ResponseWriter] to which the binary response will be written.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (r *wrapper) WriteBinary(w http.ResponseWriter) *wrapper {
	if !r.Available() {
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("WriteBinary called with nil http.ResponseWriter"))
	}
	if !r.IsBinaryBody() {
		return r.WithErrorAck(NewErrorf("WriteBinary: data is not binary, got %T", r.data))
	}
	if r.EqualHeader(NoContent) {
		w.WriteHeader(r.StatusCode())
		return r
	}

	defer New().Processing().Span("WriteBinary", slogger.String("filename", r.filename))()

	// Safely cast the data to a byte slice for writing.
	data, _ := r.data.([]byte)

	// Set the Content-Disposition header to indicate an attachment with the given filename.
	contentType := sysx.MimeFromName(r.filename)
	w.Header().Set(HeaderContentType.String(), contentType)

	// Set the Content-Length header based on the length of the binary data.
	w.Header().Set(HeaderContentLength.String(), conv.StringOrDefault(len(data), "0B"))

	// If a filename is specified, set the Content-Disposition header to indicate an attachment with the given filename.
	if strutil.IsNotEmpty(r.filename) {
		name, err := assembleContentDisposition(r.filename)
		if err != nil {
			return r.WithErrorAck(err)
		}
		w.Header().Set(HeaderContentDisposition.String(), name)
	}

	// Write the status code to the ResponseWriter.
	w.WriteHeader(r.StatusCode())

	// Write the binary data to the ResponseWriter.
	if _, err := w.Write(data); err != nil {
		return r.WithErrorAck(err)
	}
	return r
}

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
	data := r.JSONBytes()

	// Set the Content-Type header to indicate that the response is JSON with UTF-8 encoding.
	w.Header().Set(HeaderContentType.String(), MediaTypeApplicationJSONUTF8.String())

	// Write the status code to the ResponseWriter.
	w.WriteHeader(r.StatusCode())

	// Write the JSON payload to the ResponseWriter. If an error occurs during writing, it is recorded in the [wrapper] instance.
	if _, err := w.Write(data); err != nil {
		return r.WithErrorAck(err)
	}
	return r
}

// Write writes the response to the provided [http.ResponseWriter] based on the configuration of the [wrapper] instance.
//
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
