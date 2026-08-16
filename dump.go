package replify

import (
	"os"
	"path/filepath"

	"github.com/polarixa/replify/pkg/strutil"
	"github.com/polarixa/replify/pkg/sysx"
)

// DumpJSON serializes the full [wrapper] response as pretty-printed JSON and
// writes it into a self-cleaning temporary file. The returned [DumpJSON] owns the
// file — call Close (or defer it) to prevent disk leaks. Close is thread-safe
// and idempotent.
//
// The temporary file lives in the OS default temp directory with the pattern
// "replify-dump-*.json" and is removed automatically when Close is called.
//
// The serialized content matches [wrapper.JSONPretty] — the full response
// envelope (status, headers, body, meta, pagination, debug).
//
// Both return values are always non-nil:
//   - (*DumpJSON, *wrapper) — DumpJSON holds the file; wrapper carries the outcome so
//     the caller can continue a fluent chain.
//   - On error: DumpJSON is nil, wrapper has InternalServerError + error detail.
//
// Example:
//
//	dump, w := w.DumpJSON()
//	if w.IsError() {
//	    log.Fatal(w.Error())
//	}
//	defer dump.Close()
//
//	// pipe the JSON into an HTTP response writer:
//	io.Copy(rw, dump.Resource().Content())
func (w *wrapper) DumpJSON() (*Dump, *wrapper) {
	if !w.Available() {
		return nil, New().
			WithHeader(InternalServerError).
			WithMessage("Dump: wrapper is required")
	}
	d, err := dumpJSON(w.JSONBytes())
	if err != nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("Dump: failed to create temp file")
	}
	return &Dump{syr: d}, New().
		WithHeader(OK).
		WithMessagef("Dump: succeeded and written to temp file %s", d.Name())
}

// DumpJSONTo writes the full [wrapper] response as pretty-printed JSON to dst and
// simultaneously returns a [Dump] holding an in-process seekable copy for
// streaming or re-reading.
//
// Write strategy (append-or-create):
//   - If dst does not exist or is empty, data is written atomically via the
//     temp-file-and-rename pattern so readers never observe a partial write.
//   - If dst already exists and contains data, the new JSON entry is appended
//     separated by a single newline, producing a JSON-Lines / NDJSON file that
//     grows with each call. This makes DumpJSONTo safe to call repeatedly for the
//     same destination (e.g. one file per day that accumulates all responses).
//
// Parent directories are created automatically.
//
// Close on the returned Dump removes the in-process temp copy only. The
// permanent file at dst is the caller's responsibility and is never removed
// by this package.
//
// The serialized content matches [wrapper.JSONPretty] — the full response
// envelope (status, headers, body, meta, pagination, debug).
//
// Parameters:
//   - dst: destination file path; must not be empty.
//
// Both return values are always non-nil:
//   - (*Dump, *wrapper) — Dump holds the streamable copy and the dst path;
//     wrapper carries the outcome for continued fluent chaining.
//   - On error: Dump is nil, wrapper has InternalServerError + error detail.
//
// Example:
//
//	// Each call appends a new JSON entry to the same daily log file:
//	dump, w := w.DumpJSONTo("/var/log/app/responses-20260613.jsonl")
//	if w.IsError() {
//	    log.Fatal(w.Error())
//	}
//	defer dump.Close() // removes temp copy only; file at dst is kept
//
//	// re-read from the in-process temp copy:
//	io.Copy(os.Stdout, dump.Resource().Content())
func (w *wrapper) DumpJSONTo(dst string) (*Dump, *wrapper) {
	if !w.Available() {
		return nil, New().
			WithHeader(InternalServerError).
			WithMessage("DumpTo: wrapper is required")
	}
	if strutil.IsEmpty(dst) {
		return nil, New().
			WithHeader(BadRequest).
			WithMessage("DumpTo: destination path must not be empty")
	}
	payload := w.JSONBytes()

	// Write to dst: append (with newline separator) when the file already has
	// content, write atomically when the file is new or empty.
	if err := sysx.AppendOrWriteBytes(dst, payload, []byte("\n")); err != nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessagef("DumpTo: write to %q failed", dst)
	}
	// 2. Create an in-process seekable temp copy for streaming / re-reading.
	d, err := dumpJSON(payload)
	if err != nil {
		// Permanent file already written; surface only the temp-copy failure.
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("DumpTo: failed to create in-process temp copy")
	}
	d.WithName(filepath.Base(dst)) // set the name for better error messages and debugging; the full path is in the wrapper message
	return &Dump{syr: d, filepath: dst},
		New().
			WithHeader(OK).
			WithMessagef("DumpTo: succeeded, written to %q", dst)
}

// DumpBody serializes the [wrapper]'s body payload as JSON and writes it
// into a self-cleaning temporary file, returning a [Dump] that owns the file.
//
// Unlike [Dump] — which dumps the complete response envelope (status, headers,
// body, meta, pagination, debug) — DumpBody captures the raw body value only.
// The serialized content matches the JSON representation of the value passed
// to [WithBody] or [WithJSONBody].
//
// # Thread-safety
//
// DumpBody is safe for concurrent use. Each call creates an independent
// [io.Pipe] and goroutine pair; no shared mutable state is accessed after
// the body reference is read from the wrapper.
//
// # Large-body handling
//
// Serialization is streaming: the body is JSON-encoded via [json.Encoder]
// into a spill buffer ([sysx.DefaultSpillThreshold] = 8 MiB in memory,
// spilling automatically to a temp file beyond that). The full payload is
// never allocated as a single []byte.
//
// The temporary file lives in the OS default temp directory with the pattern
// "replify-dump-body-*.json" and is removed automatically when Close is called.
//
// Both return values are always non-nil:
//   - (*Dump, *wrapper) — Dump holds the seekable resource; wrapper carries
//     the outcome for fluent chaining.
//   - On error: Dump is nil, wrapper has the appropriate status + error detail.
//
// Example:
//
//	dump, w := response.DumpBody()
//	if w.IsError() {
//	    log.Fatal(w.Error())
//	}
//	defer dump.Close()
//
//	// pipe the raw body JSON into an HTTP response writer:
//	io.Copy(rw, dump.Resource().Content())
func (w *wrapper) DumpBody() (*Dump, *wrapper) {
	if !w.Available() {
		return nil, New().
			WithHeader(InternalServerError).
			WithMessage("DumpBody: wrapper is required")
	}
	if !w.IsBodyPresent() {
		return nil, New().
			WithHeader(NotFound).
			WithMessage("DumpBody: body is not present")
	}
	w.cacheMutex.RLock()
	body := w.data
	w.cacheMutex.RUnlock()
	d, err := dumpBodyStream(body)
	if err != nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("DumpBody: failed to create temp file")
	}
	return &Dump{syr: d}, New().
		WithHeader(OK).
		WithMessagef("DumpBody: succeeded and written to temp file %s", d.Name())
}

// DumpBodyTo serializes the [wrapper]'s body payload as JSON to dst and
// simultaneously returns a [Dump] holding an in-process seekable copy for
// streaming or re-reading.
//
// Unlike [DumpTo] — which dumps the complete response envelope — DumpBodyTo
// captures the raw body value only.
//
// # Write strategy (append-or-create)
//
//   - If dst does not exist or is empty, the JSON-encoded body is written
//     atomically via the temp-file-and-rename pattern so readers never observe
//     a partial write.
//   - If dst already exists and contains data, a newline separator followed
//     by the new JSON body is appended, producing a JSON-Lines / NDJSON file
//     that grows with each call.
//
// Parent directories are created automatically.
//
// # Thread-safety
//
// DumpBodyTo is safe for concurrent use. Each call creates an independent
// [io.Pipe], goroutine, and spill buffer. File-level atomicity for new files
// is guaranteed by temp-file-and-rename; concurrent appenders are serialized
// by the OS-level O_APPEND guarantee.
//
// # Large-body handling
//
// The body is JSON-encoded via [json.Encoder] into a spill buffer
// ([sysx.DefaultSpillThreshold] = 8 MiB in memory, then a temp file). The
// on-disk copy at dst is written by streaming from the spill buffer — no
// intermediate []byte allocation of the full payload.
//
// Close on the returned Dump removes the in-process temp copy only. The
// permanent file at dst is the caller's responsibility and is never removed
// by this package.
//
// Parameters:
//   - dst: destination file path; must not be empty.
//
// Both return values are always non-nil:
//   - (*Dump, *wrapper) — Dump holds the seekable copy and the dst path;
//     wrapper carries the outcome for fluent chaining.
//   - On error: Dump is nil, wrapper has the appropriate status + error detail.
//
// Example:
//
//	// Each call appends one JSON body entry to the same file:
//	dump, w := response.DumpBodyTo("/var/log/app/bodies-20260613.jsonl")
//	if w.IsError() {
//	    log.Fatal(w.Error())
//	}
//	defer dump.Close() // removes temp copy only; file at dst is kept
//
//	// re-read from the in-process temp copy:
//	io.Copy(os.Stdout, dump.Resource().Content())
func (w *wrapper) DumpBodyTo(dst string) (*Dump, *wrapper) {
	if !w.Available() {
		return nil, New().
			WithHeader(InternalServerError).
			WithMessage("DumpBodyTo: wrapper is required")
	}
	if strutil.IsEmpty(dst) {
		return nil, New().
			WithHeader(BadRequest).
			WithMessage("DumpBodyTo: destination path must not be empty")
	}
	if !w.IsBodyPresent() {
		return nil, New().
			WithHeader(NotFound).
			WithMessage("DumpBodyTo: body is not present")
	}
	w.cacheMutex.RLock()
	body := w.data
	w.cacheMutex.RUnlock()

	// Serialize body into a spill-buffered Resource (≤8 MiB in memory,
	// >8 MiB on a self-removing temp file).
	d, err := dumpBodyStream(body)
	if err != nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("DumpBodyTo: failed to serialize body")
	}
	// Stream from spill buffer to dst (append-or-create, no intermediate []byte alloc).
	if err := sysx.AppendOrCopyFrom(dst, d.Content(), []byte("\n")); err != nil {
		d.Close()
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessagef("DumpBodyTo: write to %q failed", dst)
	}
	// Rewind the in-process copy so the caller can read from offset 0.
	if err := d.Rewind(); err != nil {
		d.Close()
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("DumpBodyTo: failed to rewind in-process copy")
	}
	d.WithName(filepath.Base(dst)) // set the name for better error messages and debugging; the file itself is still the temp copy, not dst
	return &Dump{syr: d, filepath: dst},
		New().
			WithHeader(OK).
			WithMessagef("DumpBodyTo: succeeded, written to %q", dst)
}

// Resource returns the underlying [sysx.Resource], which exposes the
// serialized payload as a seekable, closeable stream ([sysx.ReadSeekCloser]).
// Valid until Close is called; returns nil when the Dump itself is nil.
func (d *Dump) Resource() *sysx.Resource {
	if d == nil {
		return nil
	}
	return d.syr
}

// Content returns the underlying [sysx.ReadSeekCloser] for the serialized
// payload. Valid until Close is called; returns nil when the Dump itself is nil.
//
// The returned stream is always seekable and closeable, but may be backed by
// either an in-memory buffer or a temporary file depending on the size of the
// payload and the spill threshold. Callers should not assume that the stream
// is always backed by a file.
func (d *Dump) Content() sysx.ReadSeekCloser {
	if d == nil || d.syr == nil {
		return nil
	}
	return d.syr.Content()
}

// ReadAll reads the entire serialized payload into a string.
// Returns an empty string when the Dump is nil or the underlying Resource is nil.
// Errors are ignored; callers should use [Dump.Content] to read the stream directly if
// they need to handle I/O errors.
func (d *Dump) ReadAll() string {
	r := d.Resource()
	if r == nil {
		return ""
	}
	c, err := r.ReadAll()
	if err != nil {
		return ""
	}
	return c
}

// ReadAllBytes reads the entire serialized payload into a byte slice.
// Returns nil when the Dump is nil or the underlying Resource is nil.
// Errors are ignored; callers should use [Dump.Content] to read the stream directly if
// they need to handle I/O errors.
func (d *Dump) ReadAllBytes() []byte {
	r := d.Resource()
	if r == nil {
		return nil
	}
	c, err := r.ReadAllBytes()
	if err != nil {
		return nil
	}
	return c
}

// MustReadAll reads the entire serialized payload into a string and returns
// a [wrapper] carrying the outcome. Returns an empty string and a wrapper with
// InternalServerError when the Dump is nil or the underlying Resource is nil.
// Errors are surfaced in the wrapper; callers should use [Dump.Content] to read
// the stream directly if they need to handle I/O errors.
func (d *Dump) MustReadAll() (c string, w *wrapper) {
	r := d.Resource()
	if r == nil {
		return "", New().
			WithHeader(InternalServerError).
			WithMessage("Dump: resource is nil")
	}
	c, err := r.ReadAll()
	if err != nil {
		return "", New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("Dump: failed to read all content")
	}
	return c, New().
		WithHeader(OK).
		WithMessagef("Dump: read %d bytes successfully", strutil.Len(c)).
		WithBody(c)
}

// MustReadAllBytes reads the entire serialized payload into a byte slice and
// returns a [wrapper] carrying the outcome. Returns nil and a wrapper with
// InternalServerError when the Dump is nil or the underlying Resource is nil.
// Errors are surfaced in the wrapper; callers should use [Dump.Content] to read
// the stream directly if they need to handle I/O errors.
func (d *Dump) MustReadAllBytes() (c []byte, w *wrapper) {
	r := d.Resource()
	if r == nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithMessage("Dump: resource is nil")
	}
	c, err := r.ReadAllBytes()
	if err != nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("Dump: failed to read all content")
	}
	return c, New().
		WithHeader(OK).
		WithMessagef("Dump: read %d bytes successfully", len(c)).
		WithBody(c)
}

// Filepath returns the destination path of the permanent on-disk file when
// the Dump was produced by [wrapper.DumpTo]. Returns an empty string for
// Dumps produced by [wrapper.Dump].
func (d *Dump) Filepath() string {
	if d == nil {
		return ""
	}
	return d.filepath
}

// Size returns the byte length of the serialized payload as reported by the
// underlying [sysx.Resource]. Returns 0 when the Dump is nil.
func (d *Dump) Size() int64 {
	if d == nil || d.syr == nil {
		return 0
	}
	return d.syr.Size()
}

// Rewind seeks the content stream back to offset 0, allowing the payload to
// be read again without creating a new Dump. Useful in retry paths or when
// the same body must be forwarded to multiple destinations.
//
// Returns [sysx.ErrNilResource] when the Resource content is nil.
func (d *Dump) Rewind() error {
	if d == nil || d.syr == nil {
		return nil
	}
	return d.syr.Rewind()
}

// File opens and returns a read-only [*os.File] for the on-disk file backing
// this Dump.
//
// Each call opens a fresh file handle; the caller is responsible for closing
// it. Because every call to [wrapper.Dump] writes through [sysx.FromTempFile],
// File always returns a non-nil handle for those Dumps. For [wrapper.DumpBody]
// the backing may be entirely in memory (body below 8 MiB and no spill
// occurred), in which case File returns (nil, nil).
//
// The returned handle becomes invalid once [Dump.Close] is called, as the
// underlying temp file is unlinked at that point.
//
// Returns:
//
//	A read-only [*os.File] and nil on success.
//	nil, nil when there is no on-disk backing.
//	nil, err when the file exists but cannot be opened.
func (d *Dump) File() (*os.File, error) {
	p := d.resolvePath()
	if strutil.IsEmpty(p) {
		return nil, nil
	}
	return os.Open(p)
}

// FileInfo returns the [os.FileInfo] for the on-disk file backing this Dump.
//
// It calls [os.Stat] on the resolved path and therefore does not consume
// or position the stream. Returns (nil, nil) when the Dump has no on-disk
// backing (in-memory payload).
//
// Returns:
//
//	An [os.FileInfo] and nil on success.
//	nil, nil when there is no on-disk backing.
//	nil, err when [os.Stat] fails.
func (d *Dump) FileInfo() (os.FileInfo, error) {
	p := d.resolvePath()
	if strutil.IsEmpty(p) {
		return nil, nil
	}
	return os.Stat(p)
}

// Close removes the backing temporary file and releases all held resources.
// It is safe to call from multiple goroutines simultaneously; the cleanup
// runs exactly once via [sync.Once] — subsequent calls are no-ops that
// return nil. The first caller receives any I/O error from the cleanup.
//
// When produced by [wrapper.DumpBodyToFile], only the in-process temp copy
// is removed. The permanent file at FilePath is never touched.
func (d *Dump) Close() error {
	if d == nil {
		return nil
	}
	d.once.Do(func() {
		if d.syr != nil {
			d.closeErr = d.syr.Close()
		}
	})
	return d.closeErr
}

// resolvePath returns the on-disk path for this Dump:
//   - permanent destination for Dumps produced by [wrapper.DumpTo] or
//     [wrapper.DumpBodyTo] (d.filepath).
//   - actual backing temp-file path for Dumps produced by [wrapper.Dump] or
//     [wrapper.DumpBody] (delegated to [sysx.Resource.ActualPath]).
//   - empty string when the payload is held entirely in memory (small body
//     below the spill threshold, no on-disk backing).
func (d *Dump) resolvePath() string {
	if d == nil {
		return ""
	}
	if strutil.IsNotEmpty(d.filepath) {
		return d.filepath
	}
	if d.syr != nil {
		return d.syr.ActualPath()
	}
	return ""
}
