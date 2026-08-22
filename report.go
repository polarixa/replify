package replify

import (
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
// "w_snapshot-*.json" and is removed automatically when Close is called.
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
			WithMessage("DumpJSON: wrapper is required")
	}
	d, err := dumpJSON(w.JSONBytes())
	if err != nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("DumpJSON: failed to create temp file")
	}
	return &Dump{syr: d}, New().
		WithHeader(OK).
		WithMessagef("DumpJSON: succeeded and written to temp file %s", d.Name())
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
			WithMessage("DumpJSONTo: wrapper is required")
	}
	if strutil.IsEmpty(dst) {
		return nil, New().
			WithHeader(BadRequest).
			WithMessage("DumpJSONTo: destination path must not be empty")
	}
	payload := w.JSONBytes()

	// Write to dst: append (with newline separator) when the file already has
	// content, write atomically when the file is new or empty.
	if err := sysx.AppendOrWriteBytes(dst, payload, []byte("\n")); err != nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessagef("DumpJSONTo: write to %s failed", dst)
	}
	// 2. Create an in-process seekable temp copy for streaming / re-reading.
	d, err := dumpJSON(payload)
	if err != nil {
		// Permanent file already written; surface only the temp-copy failure.
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("DumpJSONTo: failed to create in-process temp copy")
	}
	d.WithName(filepath.Base(dst)) // set the name for better error messages and debugging; the full path is in the wrapper message
	defer d.Close()                // ensure the temp copy is cleaned up if the caller forgets
	return &Dump{syr: d, filepath: dst},
		New().
			WithHeader(OK).
			WithMessagef("DumpJSONTo: succeeded, written to %s", dst)
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
// "w_snapshot_body-*.json" and is removed automatically when Close is called.
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

	d, err := dumpAny(body)
	if err != nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("DumpBody: failed to create temp file")
	}
	k := &Dump{syr: d}
	return k, New().
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

	d, err := dumpAny(body)
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
			WithMessagef("DumpBodyTo: write to %s failed", dst)
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
	defer d.Close()                // ensure the temp copy is cleaned up if the caller forgets
	return &Dump{syr: d, filepath: dst},
		New().
			WithHeader(OK).
			WithMessagef("DumpBodyTo: succeeded, written to %s", dst)
}

// DumpMDDoc serializes the [wrapper]'s diagnostic report as Markdown and writes
// it into a self-cleaning temporary file, returning a [Dump] that owns the file.
//
// The serialized content matches [wrapper.BasicDoc] — the full diagnostic report
// with summary, debug, headers, pagination, cursors, meta, and custom fields.
//
// Both return values are always non-nil:
//   - (*Dump, *wrapper) — Dump holds the seekable resource; wrapper carries
//     the outcome for fluent chaining.
//   - On error: Dump is nil, wrapper has the appropriate status + error detail.
func (w *wrapper) DumpMDDoc() (*Dump, *wrapper) {
	if !w.Available() {
		return nil, New().
			WithHeader(InternalServerError).
			WithMessage("DumpMDDoc: wrapper is required")
	}
	d, err := dumpMarkdown(w.BasicDoc().Bytes())
	if err != nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("DumpMDDoc: failed to create temp file")
	}
	return &Dump{syr: d}, New().
		WithHeader(OK).
		WithMessagef("DumpMDDoc: succeeded and written to temp file %s", d.Name())
}

// DumpMDDocTo serializes the [wrapper]'s diagnostic report as Markdown to dst and
// simultaneously returns a [Dump] holding an in-process seekable copy for
// streaming or re-reading.
//
// The serialized content matches [wrapper.BasicDoc] — the full diagnostic report
// with summary, debug, headers, pagination, cursors, meta, and custom fields.
//
// Both return values are always non-nil:
//   - (*Dump, *wrapper) — Dump holds the seekable copy and the dst path;
//     wrapper carries the outcome for fluent chaining.
//   - On error: Dump is nil, wrapper has the appropriate status + error detail.
func (w *wrapper) DumpMDDocTo(dst string) (*Dump, *wrapper) {
	if !w.Available() {
		return nil, New().
			WithHeader(InternalServerError).
			WithMessage("DumpMDDocTo: wrapper is required")
	}
	if strutil.IsEmpty(dst) {
		return nil, New().
			WithHeader(BadRequest).
			WithMessage("DumpMDDocTo: destination path must not be empty")
	}
	payload := w.BasicDoc().Bytes()

	if err := sysx.AppendOrWriteBytes(dst, payload, []byte("\n")); err != nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessagef("DumpMDDocTo: write to %s failed", dst)
	}
	d, err := dumpMarkdown(payload)
	if err != nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("DumpMDDocTo: failed to create in-process temp copy")
	}
	d.WithName(filepath.Base(dst)) // set the name for better error messages and debugging; the full path is in the wrapper message
	defer d.Close()                // ensure the temp copy is cleaned up if the caller forgets
	return &Dump{syr: d, filepath: dst},
		New().
			WithHeader(OK).
			WithMessagef("DumpMDDocTo: succeeded, written to %s", dst)
}
