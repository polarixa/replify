package replify

import (
	"os"

	"github.com/polarixa/replify/pkg/strutil"
	"github.com/polarixa/replify/pkg/sysx"
)

// Resource returns the underlying [sysx.Resource], which exposes the
// serialized payload as a seekable, closeable stream ([sysx.ReadSeekCloser]).
// Valid until Close is called; returns nil when the Dump itself is nil.
func (d *Dump) Resource() *sysx.Resource {
	if d == nil {
		return nil
	}
	return d.syr
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
