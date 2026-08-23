package sysx

import (
	"bytes"
	"io"
	"os"

	"github.com/polarixa/replify/pkg/strutil"
)

// NewResource creates an empty Resource ready to be configured through the
// chainable With* setters and one of the From* loaders.
//
// The returned Resource has Size() == 0, an empty Name and ContentType,
// nil Content, and DefaultSpillThreshold as its streaming threshold.
//
// Returns:
//
// A pointer to a freshly initialized Resource.
//
// Example:
//
//	res := sysx.NewResource().
//	    WithName("user-report.csv").
//	    WithContentType(sysx.MimeCSV).
//	    FromBytes(payload)
//	defer res.Close()
func NewResource() *Resource {
	return &Resource{spillThreshold: DefaultSpillThreshold, removeOnClose: true}
}

// extra builder-only configuration carried by Resource. These fields are
// declared on the original struct in type.go to keep all field-bearing
// type definitions there. The setters and accessors live in this file.
//
// (See type.go for the canonical Resource declaration.)

// Name returns the suggested filename advertised to consumers, including
// any extension. An empty string means the producer has not assigned one.
//
// Returns:
//
// The configured filename.
func (r *Resource) Name() string {
	if r == nil {
		return ""
	}
	return r.name
}

// Size returns the total number of bytes available from Content, or -1 if
// the size is not known up-front (typical for unbounded streaming
// backends).
//
// Returns:
//
// The configured payload size in bytes.
func (r *Resource) Size() int64 {
	if r == nil {
		return 0
	}
	return r.size
}

// ContentType returns the IANA media type configured on the Resource. An
// empty string means the producer could not determine one.
//
// Returns:
//
// The configured media type.
func (r *Resource) ContentType() string {
	if r == nil {
		return ""
	}
	return r.contentType
}

// Content returns the owned ReadSeekCloser stream backing the Resource.
// Callers must invoke Close exactly once when done reading, either
// directly on the returned value or through Resource.Close.
//
// Returns:
//
// The configured ReadSeekCloser, or nil if no content has been loaded.
func (r *Resource) Content() ReadSeekCloser {
	if r == nil {
		return nil
	}
	return r.content
}

// SpillThreshold returns the in-memory ceiling, in bytes, applied by
// FromReader before spilling to a temporary file. A value of 0 means the
// default (DefaultSpillThreshold) is used.
//
// Returns:
//
// The configured spill threshold in bytes.
func (r *Resource) SpillThreshold() int64 {
	if r == nil {
		return 0
	}
	return r.spillThreshold
}

// TempPattern returns the os.CreateTemp pattern used by FromTempFile and
// by FromReader's spill path. An empty string means the package default
// (defaultTempPattern) is used.
//
// Returns:
//
// The configured temp-file pattern.
func (r *Resource) TempPattern() string {
	if r == nil {
		return ""
	}
	return r.tempPattern
}

// TempDir returns the parent directory used by FromTempFile and by the
// spill path of FromReader. An empty string means the system default
// temp directory (os.TempDir) is used.
//
// Returns:
//
// The configured parent directory.
func (r *Resource) TempDir() string {
	if r == nil {
		return ""
	}
	return r.tempDir
}

// RemoveOnClose reports whether temp files created by FromTempFile and by
// the spill path of FromReader will be unlinked when the Resource is
// closed. Defaults to true.
//
// Returns:
//
// The configured auto-removal flag.
func (r *Resource) RemoveOnClose() bool {
	if r == nil {
		return false
	}
	return r.removeOnClose
}

// WithName sets the suggested filename advertised to consumers. The name
// should include any extension that matches ContentType so consumers can
// derive the correct headers (e.g. Content-Disposition).
//
// Parameters:
//   - `name`: the filename to advertise.
//
// Returns:
//
// The receiver, enabling method chaining.
func (r *Resource) WithName(name string) *Resource {
	r.name = name
	return r
}

// WithSize overrides the payload size in bytes. From* loaders set this
// automatically; calling WithSize is rarely necessary unless the producer
// already knows the total length up-front and is providing a custom
// Content via WithContent.
//
// Parameters:
//   - `size`: the payload size in bytes; -1 means unknown.
//
// Returns:
//
// The receiver, enabling method chaining.
func (r *Resource) WithSize(size int64) *Resource {
	r.size = size
	return r
}

// WithContentType sets the IANA media type explicitly, overriding any
// auto-detection performed by From* loaders.
//
// Parameters:
//   - `mime`: the IANA media type, e.g. sysx.MimeCSV.
//
// Returns:
//
// The receiver, enabling method chaining.
func (r *Resource) WithContentType(mime string) *Resource {
	r.contentType = mime
	return r
}

// WithContent attaches a custom ReadSeekCloser to the Resource. The
// caller is responsible for setting Size separately when known.
//
// Parameters:
//   - `c`: the ReadSeekCloser owning the payload bytes.
//
// Returns:
//
// The receiver, enabling method chaining.
func (r *Resource) WithContent(c ReadSeekCloser) *Resource {
	r.content = c
	return r
}

// WithSpillThreshold configures the in-memory ceiling, in bytes, applied
// by FromReader before spilling the rest of the stream to a temporary
// file. A non-positive value resets the threshold to
// DefaultSpillThreshold.
//
// Parameters:
//   - `bytes`: the in-memory ceiling.
//
// Returns:
//
// The receiver, enabling method chaining.
func (r *Resource) WithSpillThreshold(bytes int64) *Resource {
	if bytes <= 0 {
		bytes = DefaultSpillThreshold
	}
	r.spillThreshold = bytes
	return r
}

// WithTempPattern configures the os.CreateTemp pattern used by
// FromTempFile and by the spill path of FromReader. Use "*" as the random
// suffix placeholder, e.g. "user-report-*.csv".
//
// Parameters:
//   - `pattern`: the os.CreateTemp pattern.
//
// Returns:
//
// The receiver, enabling method chaining.
func (r *Resource) WithTempPattern(pattern string) *Resource {
	r.tempPattern = pattern
	return r
}

// WithTempDir configures the parent directory used by FromTempFile and by
// the spill path of FromReader. An empty string falls back to
// os.TempDir.
//
// Parameters:
//   - `dir`: the parent directory.
//
// Returns:
//
// The receiver, enabling method chaining.
func (r *Resource) WithTempDir(dir string) *Resource {
	r.tempDir = dir
	return r
}

// WithRemoveOnClose configures whether temporary files created on behalf
// of this Resource (by FromTempFile and the spill path of FromReader)
// will be unlinked when the Resource is closed.
//
// Parameters:
//   - `remove`: true to auto-remove temp files; false to retain them.
//
// Returns:
//
// The receiver, enabling method chaining.
func (r *Resource) WithRemoveOnClose(remove bool) *Resource {
	r.removeOnClose = remove
	return r
}

// FromBytes installs an in-memory MemBlob backing populated from data and
// updates Size accordingly. The slice is retained, not copied; callers
// must not mutate it for the lifetime of the Resource.
//
// If ContentType has not already been set, it is inferred from Name via
// MimeFromName.
//
// Parameters:
//   - `data`: the payload bytes.
//
// Returns:
//
// The receiver, enabling method chaining.
func (r *Resource) FromBytes(data []byte) *Resource {
	r.content = NewMemBlob(data)
	r.size = int64(len(data))
	r.fillMime()
	return r
}

// FromString is a convenience wrapper around FromBytes for textual
// payloads.
//
// Parameters:
//   - `s`: the payload text.
//
// Returns:
//
// The receiver, enabling method chaining.
func (r *Resource) FromString(s string) *Resource {
	return r.FromBytes([]byte(s))
}

// FromFile adopts an existing on-disk file as the Resource backing. The
// file is rewound and stat-ed; Size and (when not yet set) Name are
// populated from the file's metadata. Whether the file is unlinked on
// Close is controlled by WithRemoveOnClose (defaults to true).
//
// FromFile is the only Resource loader that accepts a raw *os.File; this
// is intentional, as sysx is the package's infrastructure layer.
// Callers above sysx must depend on Resource exclusively.
//
// Parameters:
//   - `f`: an open file with read permission.
//
// Returns:
//
// The receiver and any error encountered while stat-ing or rewinding
// the file.
func (r *Resource) FromFile(f *os.File) (*Resource, error) {
	if f == nil {
		return r, ErrNilResource
	}
	stat, err := f.Stat()
	if err != nil {
		return r, err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return r, err
	}
	if strutil.IsEmpty(r.name) {
		r.name = stat.Name()
	}
	r.size = stat.Size()
	r.content = wrapTempFile(f, r.removeOnClose)
	r.fillMime()
	return r, nil
}

// FromTempFile creates a fresh temporary file using the configured
// TempPattern and TempDir, invokes write with an io.Writer view of the
// file, and installs the result as the Resource backing. The file is
// rewound and stat-ed before returning, so consumers can read it
// immediately.
//
// On any failure the temp file is closed and removed before returning.
//
// Parameters:
//   - `write`: producer callback that serialize the payload. The
//     argument is intentionally io.Writer (not *os.File) so producers
//     cannot smuggle out the underlying handle and break the abstraction.
//
// Returns:
//
// The receiver and any error encountered during creation, writing,
// stat-ing, or rewinding.
func (r *Resource) FromTempFile(write func(io.Writer) error) (*Resource, error) {
	if write == nil {
		return r, ErrNilResource
	}
	tf, err := newTempFileWith(r.tempDir, r.tempPattern, r.removeOnClose)
	if err != nil {
		return r, err
	}
	if err := write(tf); err != nil {
		_ = tf.Close()
		return r, err
	}
	stat, err := tf.file.Stat()
	if err != nil {
		_ = tf.Close()
		return r, err
	}
	if _, err := tf.file.Seek(0, io.SeekStart); err != nil {
		_ = tf.Close()
		return r, err
	}
	if strutil.IsEmpty(r.name) {
		r.name = stat.Name()
	}
	r.size = stat.Size()
	r.content = tf
	r.fillMime()
	return r, nil
}

// FromReader drains an arbitrary io.Reader into a hybrid buffer that
// starts in memory and spills to a temporary file once SpillThreshold
// bytes have been accumulated. The resulting backing satisfies
// ReadSeekCloser, allowing consumers (S3 multipart, archive writers, …)
// to seek freely even though the original producer was not seekable.
//
// Parameters:
//   - `src`: the producer stream; must not be nil.
//
// Returns:
//
// The receiver and any error encountered while draining or sealing the
// buffer.
func (r *Resource) FromReader(src io.Reader) (*Resource, error) {
	sb, err := r.drainToSpill(src, r.spillThreshold, r.tempDir, r.tempPattern, r.removeOnClose)
	if err != nil {
		return r, err
	}
	r.size = sb.size
	r.content = sb
	r.fillMime()
	return r, nil
}

// FromPath opens the file at path and installs it as the Resource backing.
// The file is rewound and stat-ed; Size and (when not yet set) Name are
// populated from the file's metadata. Whether the file is unlinked on
// Close is controlled by WithRemoveOnClose (defaults to true).
//
// Parameters:
//   - `path`: the filesystem path of the file to adopt.
//
// Returns:
//
// The receiver and any error encountered while opening, stat-ing, or
// rewinding the file.
func (r *Resource) FromPath(path string) (*Resource, error) {
	if strutil.IsEmpty(path) {
		return r, ErrNilResource
	}
	f, err := os.Open(path)
	if err != nil {
		return r, err
	}
	defer f.Close()
	return r.FromFile(f)
}

// Close releases the underlying ReadSeekCloser, tolerating a nil
// Resource or a nil Content. Calling Close more than once is safe; the
// underlying backings are required to be idempotent.
//
// Returns:
//
// Any error returned by the underlying Close, or nil when there is
// nothing to close.
func (r *Resource) Close() error {
	if r == nil || r.content == nil {
		return nil
	}
	return r.content.Close()
}

// Rewind seeks Content back to offset 0. It is intended for consumers
// that need to re-read a Resource — for example, an HTTP retry path or
// an upload that must recompute a checksum.
//
// Returns:
//
// Any error returned by Seek, or ErrNilResource when Content is nil.
func (r *Resource) Rewind() error {
	if r == nil || r.content == nil {
		return ErrNilResource
	}
	_, err := r.content.Seek(0, io.SeekStart)
	return err
}

// CopyTo streams the Resource payload into dst and returns the number of
// bytes written. The Resource is NOT closed; the caller retains
// ownership and must invoke Close once consumption is complete.
//
// Parameters:
//   - `dst`: the destination writer.
//
// Returns:
//
// The number of bytes copied and any error encountered during the copy.
func (r *Resource) CopyTo(dst io.Writer) (int64, error) {
	if r == nil || r.content == nil {
		return 0, ErrNilResource
	}
	return io.Copy(dst, r.content)
}

// Drain reads and discards the entire Resource payload. It is useful for
// integration tests, dry-run pipelines, and benchmark shaping where the
// caller wants to exercise the producer end-to-end without preserving
// the output.
//
// Returns:
//
// The number of bytes discarded and any error encountered during the
// read.
func (r *Resource) Drain() (int64, error) {
	return r.CopyTo(io.Discard)
}

// ActualPath returns the on-disk path of the file backing this Resource.
//
// The returned path is non-empty only when the Resource is backed by a
// concrete file:
//
//   - [FromTempFile] — always returns the actual temp file path.
//   - [FromReader] with spill — returns the spill temp file path once the
//     in-memory threshold has been exceeded.
//   - [FromReader] without spill / [FromBytes] / [FromString] — returns ""
//     because the payload is held entirely in memory.
//   - [FromFile] — returns the path of the adopted file.
//
// The returned string is valid as long as the Resource has not been closed.
// Callers must not use the path after Close, as the underlying file may have
// been deleted.
//
// Returns:
//
// The absolute on-disk path, or an empty string for memory-backed Resources.
func (r *Resource) ActualPath() string {
	if r == nil {
		return ""
	}
	switch c := r.content.(type) {
	case *TempFile:
		return c.Path()
	case *spillBuffer:
		if c.file != nil {
			return c.file.Path()
		}
	}
	return ""
}

// ReadAll rewinds Content to offset 0, reads every byte into memory, and
// returns the payload as a string. The operation is serialized by an
// internal mutex so it is safe to call from multiple goroutines sharing
// the same Resource, though callers should still avoid interleaving reads
// with concurrent From* calls or Close.
//
// Returns:
//
// The full payload as a string and any I/O error encountered.
func (r *Resource) ReadAll() (string, error) {
	if r == nil || r.content == nil {
		return "", ErrNilResource
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, err := r.content.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.content); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ReadAllBytes rewinds Content to offset 0, reads every byte into memory, and
// returns the payload as a byte slice. The operation is serialized by an
// internal mutex so it is safe to call from multiple goroutines sharing
// the same Resource, though callers should still avoid interleaving reads
// with concurrent From* calls or Close.
//
// Returns:
//
// A byte slice containing the full payload and any I/O error encountered.
func (r *Resource) ReadAllBytes() ([]byte, error) {
	if r == nil || r.content == nil {
		return nil, ErrNilResource
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, err := r.content.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.content); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SizeHumanReadable returns Size formatted as a human-readable string
// using binary prefixes (KiB, MiB, GiB, etc.), as produced by
// HumanizeBytes.
//
// Returns:
//
// The configured payload size in bytes, humanized.
func (r *Resource) SizeHumanReadable() string {
	if r == nil {
		return HumanizeBytes(0)
	}
	return HumanizeBytes(r.Size())
}

// IsDir reports whether the [Resource]'s underlying content handle is
// actually an open directory rather than a regular file or in-memory
// buffer.
//
// This can only be true for filesystem-backed content adopted via
// FromFile / FromPath (both use [os.Open], which succeeds on directories on
// most platforms) or attached directly via WithContent with a caller-
// supplied *[os.File]. Purely in-memory backings ([MemBlob]) and the
// in-memory phase of a [spillBuffer] can never be directories.
//
// Detection is done via Stat on the already-open file descriptor rather
// than by re-stat-ing a path string, so the result reflects the actual
// handle the Resource holds — it is unaffected by TOCTOU races (the path
// being replaced on disk between calls) and, on Unix, keeps working even
// if the directory entry was unlinked after opening. If the handle has
// already been closed, or Stat fails for any other reason, IsDir
// conservatively returns false rather than propagating an error: this
// method is a boolean predicate, not a validator, and a Resource that
// can't prove it's a directory should be treated as if it isn't one.
//
// A nil Resource, a Resource with no content loaded, and any
// custom [ReadSeekCloser] supplied via WithContent that isn't *[os.File] or a
// sysx-internal file-backed type all safely return false.
//
// Returns:
//
// True only when the content handle is confirmed, via Stat, to be a
// directory.
func (r *Resource) IsDir() bool {
	if r == nil {
		return false
	}
	r.mu.Lock()
	content := r.content
	r.mu.Unlock()
	return contentIsDir(content)
}

// fillMime populates ContentType from Name when ContentType is empty. It
// is invoked by every From* loader after Name and Content are settled.
func (r *Resource) fillMime() {
	if strutil.IsEmpty(r.contentType) {
		r.contentType = MimeFromName(r.name)
	}
}

// drainToSpill is the constructor for the private spillBuffer backing
// used by Resource.FromReader. It reads src to EOF, buffering the first
// threshold bytes in memory and overflowing the rest onto a temporary
// file created with the supplied pattern/dir/removeOnClose options.
//
// A non-positive threshold falls back to DefaultSpillThreshold.
func (r *Resource) drainToSpill(src io.Reader, threshold int64, tempDir, tempPattern string, remove bool) (*spillBuffer, error) {
	if src == nil {
		return nil, ErrNilResource
	}
	if threshold <= 0 {
		threshold = DefaultSpillThreshold
	}
	sb := &spillBuffer{
		mem:       new(bytes.Buffer),
		threshold: threshold,
	}
	if err := sb.drain(src, tempDir, tempPattern, remove); err != nil {
		_ = sb.Close()
		return nil, err
	}
	if err := sb.seal(); err != nil {
		_ = sb.Close()
		return nil, err
	}
	return sb, nil
}
