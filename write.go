package replify

import (
	"net/http"
	"reflect"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strutil"
	"github.com/polarixa/replify/pkg/sysx"
)

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
	if r.EqualHeader(NoContent) {
		w.WriteHeader(r.StatusCode())
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("WriteFile called with nil http.ResponseWriter"))
	}
	if strutil.IsEmpty(r.filepath) {
		r.WithHeader(BadRequest).
			WithMessage("failed to write file, filepath is empty").
			BindCause()
		return r.WriteJSON(w)
	}

	// Open the file specified in the [wrapper] instance.
	resource, err := sysx.NewResource().
		FromPath(r.filepath, false)
	if err != nil {
		r.WithHeader(InternalServerError).
			WithMessage("failed to open file").
			WithErrorAck(err).
			WithDebuggingKV("filepath", r.filepath)
		return r.WriteJSON(w)
	}

	// Check if the resource is a directory. If it is, return an error response.
	if resource.IsDir() {
		r.WithHeader(InternalServerError).
			WithMessage("failed to write file: path is a directory").
			BindCause().
			WithDebuggingKV("filepath", r.filepath).
			WithDebuggingKV("filename", r.filename)
		return r.WriteJSON(w)
	}

	// Use a deferred span for logging the WriteFile operation with the file path.
	defer New().Processing().Span("WriteFile",
		slogger.String("request_id", r.Meta().RequestID()),
		slogger.String("filepath", r.filepath),
		slogger.String("filename", r.filename),
		slogger.String("content_type", resource.ContentType()),
		slogger.String("content_length", resource.SizeHumanReadable()),
	)()

	// Set the Content-Type header based on the resource's content type.
	w.Header().Set(HeaderContentType.String(), resource.ContentType())

	// Set the Content-Length header based on the resource's size.
	w.Header().Set(HeaderContentLength.String(), conv.StringOrDefault(resource.Size(), "0B"))

	// If a filename is specified, set the Content-Disposition header to indicate an attachment with the given filename.
	if strutil.IsNotEmpty(r.filename) {
		name, err := assembleContentDisposition(r.filename)
		if err != nil {
			r.WithHeader(InternalServerError).
				WithMessagef("failed to assemble Content-Disposition for filename: '%s'", r.filename).
				WithErrorAck(err).
				WithDebuggingKV("filepath", r.filepath).
				WithDebuggingKV("filename", r.filename)
			return r.WriteJSON(w)
		}
		w.Header().Set(HeaderContentDisposition.String(), name)
	}

	// Write the status code to the ResponseWriter.
	w.WriteHeader(r.StatusCode())

	// Copy the resource's content to the ResponseWriter.
	// If an error occurs during the copy operation, it is recorded in the [wrapper] instance.
	if _, err := resource.CopyTo(w); err != nil {
		r.WithHeader(InternalServerError).
			WithMessage("file content could not be streamed to the response").
			WithErrorAck(err).
			WithDebuggingKV("filepath", r.filepath).
			WithDebuggingKV("filename", r.filename)
		return r.WriteJSON(w)
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
	if r.EqualHeader(NoContent) {
		w.WriteHeader(r.StatusCode())
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("WriteBinary called with nil http.ResponseWriter"))
	}
	if r.data == nil {
		r.WithHeader(BadRequest).
			WithMessage("failed to write binary data, data is nil").
			BindCause()
		return r.WriteJSON(w)
	}
	// Check if the data is binary, if not, return an error acknowledgment
	if !r.IsBinaryBody() {
		typename := reflect.TypeOf(r.data).String()

		r.WithHeader(UnprocessableEntity).
			WithMessage("response body must be a binary byte slice").
			WithErrorAck(NewErrorf("data is not binary, got %s", typename)).
			WithDebuggingKV("data_type", typename)
		return r.WriteJSON(w)
	}

	defer New().Processing().Span("WriteBinary",
		slogger.String("request_id", r.Meta().RequestID()),
		slogger.String("filename", r.filename),
	)()

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
			r.WithHeader(InternalServerError).
				WithMessagef("failed to assemble Content-Disposition for filename: '%s'", r.filename).
				WithErrorAck(err).
				WithDebuggingKV("filename", r.filename)
			return r.WriteJSON(w)
		}
		w.Header().Set(HeaderContentDisposition.String(), name)
	}

	// Write the status code to the ResponseWriter.
	w.WriteHeader(r.StatusCode())

	// Write the binary data to the ResponseWriter.
	if _, err := w.Write(data); err != nil {
		r.WithHeader(InternalServerError).
			WithMessage("binary data could not be written to the response").
			WithErrorAck(err).
			WithDebuggingKV("filename", r.filename)
		return r.WriteJSON(w)
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
	if r.EqualHeader(NoContent) {
		w.WriteHeader(r.StatusCode())
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("WriteJSON called with nil http.ResponseWriter"))
	}

	defer New().Processing().Span("WriteJSON",
		slogger.String("request_id", r.Meta().RequestID()),
		slogger.Int("status_code", r.StatusCode()),
		slogger.String("message", r.Message()),
	)()

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
