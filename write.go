package replify

import (
	"net/http"
	"os"

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
// Failures before the response is committed (empty path, file cannot be opened, path is a directory, unsafe
// filename) are answered with a JSON error response carrying only JSON headers; no file Content-Length or
// Content-Disposition is sent with it. Once the status line and headers have been written, the response is
// committed and cannot be replaced: a failure while streaming the body is only recorded on the [wrapper]
// (error, message, debugging details) and nothing more is written. The client sees a body shorter than the
// advertised Content-Length, which is the correct HTTP signal for a truncated transfer.
//
// A nil w never panics, even for a "204 No Content" response; the error is recorded on the [wrapper] instead.
//
// WriteFile always streams the whole file with the wrapper's status code; it has no [http.Request], so it
// cannot honor Range, If-Range, If-Modified-Since, If-None-Match or HEAD requests. Use [wrapper.ServeFile]
// when the caller has the request and wants resumable downloads and client-side caching.
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
		return r.WithErrorAck(NewError("WriteFile requires a non-nil http.ResponseWriter to stream the file response"))
	}
	if r.EqualHeader(NoContent) {
		w.WriteHeader(r.StatusCode())
		return r
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
	if r.isSpan() {
		defer New().Processing().Span("WriteFile",
			slogger.String("request_id", r.Meta().RequestID()),
			slogger.String("filepath", r.filepath),
			slogger.String("filename", r.filename),
			slogger.String("content_type", resource.ContentType()),
			slogger.String("content_length", resource.SizeHumanReadable()),
		)()
	}

	// Assemble Content-Disposition first: if it fails, nothing file-specific has been
	// put on the header map yet, so the JSON fallback below is clean (no stale
	// Content-Length or Content-Disposition).
	var disposition string
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
		disposition = name
	}

	// Set the Content-Type header based on the resource's content type.
	w.Header().Set(HeaderContentType.String(), resource.ContentType())

	// Set the Content-Length header based on the resource's size (a plain decimal number).
	w.Header().Set(HeaderContentLength.String(), conv.StringOrDefault(resource.Size(), "0"))
	r.attachMetadata(w) // Attach the wrapper's metadata headers to the [http.ResponseWriter].

	// If a filename is specified, mark the response as an attachment with the given filename.
	if disposition != "" {
		w.Header().Set(HeaderContentDisposition.String(), disposition)
	}

	// Write the status code to the ResponseWriter. From here on the response is committed.
	w.WriteHeader(r.StatusCode())

	// Copy the resource's content to the ResponseWriter.
	// The status line and headers (including Content-Length) are already on the wire, so
	// a failure here can no longer be turned into a JSON error response: it is only
	// recorded in the [wrapper]. The truncated body is the client's error signal.
	if _, err := resource.CopyTo(w); err != nil {
		r.WithHeader(InternalServerError).
			WithMessage("file content could not be streamed to the response").
			WithErrorAck(err).
			WithDebuggingKV("filepath", r.filepath).
			WithDebuggingKV("filename", r.filename)
		return r
	}
	return r
}

// ServeFile serves the file set on the [wrapper] (see [wrapper.File] and [wrapper.FileAttachment]) through
// [http.ServeContent], so the response honors Range and If-Range (206 partial content, 416 for an unsatisfiable
// range), conditional requests (If-Modified-Since, If-None-Match, If-Match, If-Unmodified-Since: 304 and 412)
// and HEAD requests.
//
// It exists next to [wrapper.WriteFile] because [http.ServeContent] needs a non-nil [*http.Request], which
// WriteFile does not have. WriteFile keeps its simple streaming behavior and is unchanged, and [wrapper.Write]
// still dispatches to it; call ServeFile explicitly when you have the request.
//
// Flow:
//  1. A nil w or a nil req records an error on the [wrapper] and returns without writing (never panics).
//     A "204 No Content" wrapper writes only the 204 status. An empty file path is answered with a JSON 400.
//  2. The file is opened and stat-ed. A file that cannot be opened or stat-ed, or a directory, is answered
//     with a JSON 500 error response (the error is recorded on the [wrapper]).
//  3. Headers are prepared before [http.ServeContent] runs: Content-Type (derived from the file name with the
//     same logic as WriteFile), the wrapper's metadata headers, Content-Disposition when a filename is set,
//     and an ETag. Content-Disposition is assembled first, so an unsafe filename falls back to a JSON 500
//     error response with none of the file headers set.
//  4. [http.ServeContent] is called. Content-Length, Last-Modified and Accept-Ranges are left to it, because
//     they differ for 200, 206, 304 and 416 responses.
//
// ETag: [http.ServeContent] does not generate one, so ServeFile sets a strong validator derived from the file's
// size and modification time, formatted as "<size in hex>-<modtime in unix nanoseconds, hex>" (quoted). It
// changes whenever either changes, which enables If-None-Match and If-Range handling.
//
// Status ownership: on this path [http.ServeContent] chooses the status code (200, 206, 304, 412 or 416); a
// status preset on the [wrapper] is ignored, except "204 No Content", which is handled before the file is opened.
//
// Mid-stream errors: [http.ServeContent] returns nothing, so a failure while copying the body (for example a
// client disconnect) is not observable here. Unlike [wrapper.WriteFile], ServeFile therefore records no error
// for it and returns the [wrapper] unchanged.
//
// Parameters:
//   - w: An [http.ResponseWriter] to which the file response will be written.
//   - req: The incoming request; its method and Range and conditional headers drive the response.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (r *wrapper) ServeFile(w http.ResponseWriter, req *http.Request) *wrapper {
	if !r.Available() {
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("ServeFile requires a non-nil http.ResponseWriter to serve the file response"))
	}
	if req == nil {
		return r.WithErrorAck(NewError("ServeFile requires a non-nil *http.Request to serve the file response"))
	}
	if r.EqualHeader(NoContent) {
		w.WriteHeader(r.StatusCode())
		return r
	}
	if strutil.IsEmpty(r.filepath) {
		r.BadRequest().
			WithMessage("failed to serve file, filepath is empty").
			BindCause()
		return r.WriteJSON(w)
	}

	file, err := os.Open(r.filepath)
	if err != nil {
		r.InternalServerError().
			WithMessage("failed to open file").
			WithErrorAck(err).
			WithDebuggingKV("filepath", r.filepath)
		return r.WriteJSON(w)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		r.InternalServerError().
			WithMessage("failed to read file information").
			WithErrorAck(err).
			WithDebuggingKV("filepath", r.filepath)
		return r.WriteJSON(w)
	}
	if info.IsDir() {
		r.InternalServerError().
			WithMessage("failed to serve file: path is a directory").
			BindCause().
			WithDebuggingKV("filepath", r.filepath).
			WithDebuggingKV("filename", r.filename)
		return r.WriteJSON(w)
	}

	// Assemble Content-Disposition first so that a failure leaves no file-specific
	// headers on the response (clean JSON fallback).
	var disposition string
	if strutil.IsNotEmpty(r.filename) {
		name, err := assembleContentDisposition(r.filename)
		if err != nil {
			r.InternalServerError().
				WithMessagef("failed to assemble Content-Disposition for filename: '%s'", r.filename).
				WithErrorAck(err).
				WithDebuggingKV("filepath", r.filepath).
				WithDebuggingKV("filename", r.filename)
			return r.WriteJSON(w)
		}
		disposition = name
	}

	// Content-Type uses the same name-based logic as WriteFile (sysx.Resource derives it from the base name).
	w.Header().Set(HeaderContentType.String(), sysx.MimeFromName(info.Name()))
	r.attachMetadata(w) // Attach the wrapper's metadata headers to the [http.ResponseWriter].
	if disposition != "" {
		w.Header().Set(HeaderContentDisposition.String(), disposition)
	}
	// http.ServeContent does not generate an ETag; set one so If-None-Match and If-Range work.
	w.Header().Set("ETag", fileETag(info.Size(), info.ModTime()))

	name := r.filename
	if strutil.IsEmpty(name) {
		name = r.filepath
	}
	http.ServeContent(w, req, name, info.ModTime(), file)
	return r
}

// WriteBinary writes the binary data specified in the [wrapper] instance to the provided [http.ResponseWriter].
//
// This method checks the availability of the [wrapper] instance and ensures that binary data is set.
// It sets the appropriate headers, including Content-Type and Content-Length, and writes the binary data to the ResponseWriter.
// If any errors occur during writing, they are recorded in the [wrapper] instance.
//
// Failures before the response is committed (nil or non-binary data, unsafe filename) are answered with a JSON
// error response carrying only JSON headers; no binary Content-Length or Content-Disposition is sent with it.
// Once the status line and headers have been written, the response is committed and cannot be replaced: a write
// failure is only recorded on the [wrapper] and nothing more is written, so the client sees a truncated body.
//
// A nil w never panics, even for a "204 No Content" response; the error is recorded on the [wrapper] instead.
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
		return r.WithErrorAck(NewError("WriteBinary requires a non-nil http.ResponseWriter to write the binary response"))
	}
	if r.EqualHeader(NoContent) {
		w.WriteHeader(r.StatusCode())
		return r
	}
	// r.data is nil (unset) vs []byte(nil) stored by Binary(nil) — only the former is an error
	if r.data == nil {
		r.WithHeader(BadRequest).
			WithMessage("failed to write binary data, data is nil").
			BindCause()
		return r.WriteJSON(w)
	}
	// Safely cast the data to a byte slice for writing.
	data, ok := r.data.([]byte)
	if !ok {
		typename := r.BodyType()

		r.WithHeader(UnprocessableEntity).
			WithMessage("response body must be a binary byte slice").
			WithErrorAck(NewErrorf("data is not binary, got %s", typename)).
			WithDebuggingKV("data_type", typename)
		return r.WriteJSON(w)
	}

	if r.isSpan() {
		defer New().Processing().Span("WriteBinary",
			slogger.String("request_id", r.Meta().RequestID()),
			slogger.String("filename", r.filename),
		)()
	}

	// Assemble Content-Disposition first: if it fails, nothing binary-specific has been
	// put on the header map yet, so the JSON fallback below is clean (no stale
	// Content-Length or Content-Disposition).
	var disposition string
	if strutil.IsNotEmpty(r.filename) {
		name, err := assembleContentDisposition(r.filename)
		if err != nil {
			r.WithHeader(InternalServerError).
				WithMessagef("failed to assemble Content-Disposition for filename: '%s'", r.filename).
				WithErrorAck(err).
				WithDebuggingKV("filename", r.filename)
			return r.WriteJSON(w)
		}
		disposition = name
	}

	// Set the Content-Type header based on the filename's MIME type.
	contentType := sysx.MimeFromName(r.filename)
	w.Header().Set(HeaderContentType.String(), contentType)

	// Set the Content-Length header based on the length of the binary data (a plain decimal number).
	w.Header().Set(HeaderContentLength.String(), conv.StringOrDefault(len(data), "0"))
	r.attachMetadata(w) // Attach the wrapper's metadata headers to the [http.ResponseWriter].

	// If a filename is specified, mark the response as an attachment with the given filename.
	if disposition != "" {
		w.Header().Set(HeaderContentDisposition.String(), disposition)
	}

	// Write the status code to the ResponseWriter. From here on the response is committed.
	w.WriteHeader(r.StatusCode())

	// Write the binary data to the ResponseWriter.
	// The status line and headers (including Content-Length) are already on the wire, so
	// a failure here can no longer be turned into a JSON error response: it is only
	// recorded in the [wrapper]. The truncated body is the client's error signal.
	if _, err := w.Write(data); err != nil {
		r.WithHeader(InternalServerError).
			WithMessage("binary data could not be written to the response").
			WithErrorAck(err).
			WithDebuggingKV("filename", r.filename)
		return r
	}
	return r
}

// WriteJSON writes the [wrapper] instance as a JSON response to the provided [http.ResponseWriter].
//
// This method checks if the [wrapper] is available and if the provided [http.ResponseWriter] is not nil.
// If the [wrapper] is not available or the [http.ResponseWriter] is nil, it returns the [wrapper] with an error acknowledgment.
// The nil check comes before everything else, so a nil writer never panics, including for a "204 No Content" response.
// If the status code indicates "204 No Content", it writes the status code without a body.
// Otherwise, it serializes the [wrapper] to JSON, sets the appropriate Content-Type header, and writes the JSON payload to the ResponseWriter.
//
// Parameters:
//   - w: An [http.ResponseWriter] to which the JSON response will be written.
//   - ignoringJSONfields: A variadic list of field names to be ignored when serializing the [wrapper] to JSON.
//
// Returns:
//   - A pointer to the modified [wrapper] instance, allowing for method chaining.
func (r *wrapper) WriteJSON(w http.ResponseWriter, ignoringJSONfields ...string) *wrapper {
	if !r.Available() {
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("WriteJSON requires a non-nil http.ResponseWriter to write the JSON response"))
	}
	if r.EqualHeader(NoContent) {
		w.WriteHeader(r.StatusCode())
		return r
	}

	if r.isSpan() {
		defer New().Processing().Span("WriteJSON",
			slogger.String("request_id", r.Meta().RequestID()),
			slogger.Int("status_code", r.StatusCode()),
			slogger.String("message", r.Message()),
		)()
	}

	data := r.JSONBytesIgnoring(ignoringJSONfields...)

	// Set the Content-Type header to indicate that the response is JSON with UTF-8 encoding.
	w.Header().Set(HeaderContentType.String(), MediaTypeApplicationJSONUTF8.String())
	r.attachMetadata(w) // Attach the wrapper's metadata headers to the [http.ResponseWriter].

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
//   - ignoringJSONfields: A variadic list of JSON field names to be ignored when writing the JSON response.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (r *wrapper) Write(w http.ResponseWriter, ignoringJSONfields ...string) *wrapper {
	if !r.Available() {
		return r
	}
	if strutil.IsNotEmpty(r.filepath) {
		return r.WriteFile(w)
	}
	if _, ok := r.data.([]byte); ok || r.IsBinaryBody() {
		return r.WriteBinary(w)
	}
	return r.WriteJSON(w, ignoringJSONfields...)
}

/////////////////////////////////////////////////////////////////////////////////
//
// Section: Signature-aware JSON writers
// These methods combine signature generation ([ApplySignature], [ApplySignatureFromRequest],
// [ApplySignatureHeader]) with [WriteJSON] so the signed response can be written in one call.
//
// The fields ignored when computing the signature and the fields ignored when serializing the
// JSON sent to the client are independently configurable via [SignatureConfig.WithIgnoringSignatureFields]
// and [SignatureConfig.WithIgnoringResponseFields] — sharing a single ignore-list between the two
// would otherwise strip the freshly generated "signature" field from the client response.
//
/////////////////////////////////////////////////////////////////////////////////

// WriteJSONSignature generates a signature for the wrapper's JSON body using the provided
// [SignatureConfig], applies it to the wrapper via [ApplySignature], and writes the resulting
// JSON response to the [http.ResponseWriter].
//
// Fields ignored when computing the signature are taken from [SignatureConfig.IgnoringSignatureFields];
// fields ignored when serializing the response are taken from [SignatureConfig.IgnoringResponseFields].
//
// Parameters:
//   - w: An [http.ResponseWriter] to which the JSON response will be written.
//   - config: The [SignatureConfig] instance containing the signature configuration.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
//
// Example:
//
//	config := replify.NewSignatureConfig("abc@123").
//	    WithIgnoringSignatureFields("signature").
//	    WithIgnoringResponseFields("debug")
//	w.WriteJSONSignature(rw, config)
func (r *wrapper) WriteJSONSignature(w http.ResponseWriter, config *SignatureConfig) *wrapper {
	if !r.Available() {
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("WriteJSONSignature requires a non-nil http.ResponseWriter to write the signed JSON response"))
	}
	if r.EqualHeader(NoContent) {
		return r.WriteJSON(w)
	}
	r.ApplySignature(config, config.IgnoringSignatureFields()...)
	return r.WriteJSON(w, config.IgnoringResponseFields()...)
}

// WriteJSONSignatureHeader generates a signature for the wrapper's JSON body using the provided
// [SignatureConfig], adds the signature headers (X-Signature, X-Signature-Algorithm and, when
// present, X-Signature-Timestamp) to the [http.ResponseWriter] via [ApplySignatureHeader], and
// writes the resulting JSON response.
//
// Fields ignored when computing the signature are taken from [SignatureConfig.IgnoringSignatureFields];
// fields ignored when serializing the response are taken from [SignatureConfig.IgnoringResponseFields].
//
// Parameters:
//   - w: An [http.ResponseWriter] to which the signature headers and JSON response will be written.
//   - config: The [SignatureConfig] instance containing the signature configuration.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (r *wrapper) WriteJSONSignatureHeader(w http.ResponseWriter, config *SignatureConfig) *wrapper {
	if !r.Available() {
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("WriteJSONSignatureHeader requires a non-nil http.ResponseWriter to write the signature headers and JSON response"))
	}
	if r.EqualHeader(NoContent) {
		return r.WriteJSON(w)
	}
	r.ApplySignatureHeader(w, config, config.IgnoringSignatureFields()...)
	return r.WriteJSON(w, config.IgnoringResponseFields()...)
}

// WriteJSONSignatureFromRequest generates a signature for the wrapper's JSON body based on the
// provided HTTP request and [SignatureConfig] via [ApplySignatureFromRequest], then writes the
// resulting JSON response to the [http.ResponseWriter].
//
// Fields ignored when computing the signature are taken from [SignatureConfig.IgnoringSignatureFields];
// fields ignored when serializing the response are taken from [SignatureConfig.IgnoringResponseFields].
//
// Parameters:
//   - w: An [http.ResponseWriter] to which the JSON response will be written.
//   - request: The HTTP request containing the headers to be signed.
//   - config: The [SignatureConfig] instance containing the signature configuration.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (r *wrapper) WriteJSONSignatureFromRequest(w http.ResponseWriter, request *http.Request, config *SignatureConfig) *wrapper {
	if !r.Available() {
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("WriteJSONSignatureFromRequest requires a non-nil http.ResponseWriter to write the signed JSON response"))
	}
	if r.EqualHeader(NoContent) {
		return r.WriteJSON(w)
	}
	r.ApplySignatureFromRequest(request, config, config.IgnoringSignatureFields()...)
	return r.WriteJSON(w, config.IgnoringResponseFields()...)
}

// attachMetadata writes the wrapper's metadata headers to the provided [http.ResponseWriter].
//
// Parameters:
//   - w: An [http.ResponseWriter] to which the metadata headers will be written.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (r *wrapper) attachMetadata(w http.ResponseWriter) *wrapper {
	if !r.Available() {
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("attachMetadata requires a non-nil http.ResponseWriter to write the metadata headers"))
	}
	// Attach the wrapper's metadata headers to the [http.ResponseWriter].
	w.Header().Set(HeaderXLocale.String(), r.Meta().Locale())
	w.Header().Set(HeaderXAPIVersion.String(), r.Meta().ApiVersion())
	w.Header().Set(HeaderXRequestID.String(), r.Meta().RequestID())
	w.Header().Set(HeaderXRequestTime.String(), r.Meta().RequestedTimeFormat())
	return r
}
