package sysx

import (
	"io"
	"os"

	"github.com/polarixa/replify/pkg/strutil"
)

// NewTempFile creates a new temporary file in the system default temp
// directory using defaultTempPattern. The returned TempFile is armed for
// auto-removal on Close.
//
// Returns:
//
// A pointer to the new TempFile and any error returned by os.CreateTemp.
//
// Example:
//
//	tf, err := sysx.NewTempFile()
//	if err != nil {
//	    return err
//	}
//	defer tf.Close()
func NewTempFile() (*TempFile, error) {
	return newTempFileWith("", "", true)
}

// NewTempFileAt creates a new temporary file using pattern inside dir.
// An empty dir falls back to the system default temp directory; an empty
// pattern falls back to defaultTempPattern. The returned TempFile is
// armed for auto-removal on Close.
//
// Parameters:
//   - `dir`: the parent directory; empty for the default temp dir.
//   - `pattern`: the os.CreateTemp pattern.
//
// Returns:
//
// A pointer to the new TempFile and any error returned by
// os.CreateTemp.
func NewTempFileAt(dir, pattern string) (*TempFile, error) {
	return newTempFileWith(dir, pattern, true)
}

// NewTempFilename creates a new temporary file using the supplied
// os.CreateTemp pattern (e.g. "user-report-*.csv") in the default temp
// directory. The returned TempFile is armed for auto-removal on Close.
//
// Parameters:
//   - `pattern`: the os.CreateTemp pattern.
//
// Returns:
//
// A pointer to the new TempFile and any error returned by
// os.CreateTemp.
func NewTempFilename(pattern string) (*TempFile, error) {
	return newTempFileWith("", pattern, true)
}

// Path returns the path of the temporary file as reported by the
// underlying *os.File.
//
// Returns:
//
// The on-disk path of the temp file.
func (t *TempFile) Path() string {
	if t == nil || t.file == nil {
		return ""
	}
	return t.file.Name()
}

// Stat returns the FileInfo describing the temporary file.
//
// Returns:
//
// The os.FileInfo and any error returned by the underlying Stat call.
func (t *TempFile) Stat() (os.FileInfo, error) {
	if t == nil || t.file == nil {
		return nil, ErrNilResource
	}
	return t.file.Stat()
}

// RemoveOnClose reports whether the underlying file will be unlinked the
// next time Close is called.
//
// Returns:
//
// true when auto-removal is armed.
func (t *TempFile) RemoveOnClose() bool {
	if t == nil {
		return false
	}
	return t.removeOnClose
}

// Closed reports whether Close has already been invoked on the TempFile.
// Subsequent Close calls are no-ops once Closed returns true.
//
// Returns:
//
// true after the first successful Close.
func (t *TempFile) Closed() bool {
	if t == nil {
		return true
	}
	return t.closed
}

// WithRemoveOnClose configures whether the underlying file will be
// unlinked the next time Close is called. Use this to disarm auto-
// removal before handing the file off to another owner (for example, by
// renaming it into a permanent location).
//
// Parameters:
//   - `remove`: true to arm auto-removal; false to disarm it.
//
// Returns:
//
// The receiver, enabling method chaining.
func (t *TempFile) WithRemoveOnClose(remove bool) *TempFile {
	t.removeOnClose = remove
	return t
}

// Read implements io.Reader by delegating to the underlying *os.File.
func (t *TempFile) Read(p []byte) (int, error) {
	return t.file.Read(p)
}

// Write implements io.Writer by delegating to the underlying *os.File.
// It is exposed so FromTempFile can hand a write-only view to producer
// callbacks without leaking the *os.File itself.
func (t *TempFile) Write(p []byte) (int, error) {
	return t.file.Write(p)
}

// Seek implements io.Seeker by delegating to the underlying *os.File.
func (t *TempFile) Seek(offset int64, whence int) (int64, error) {
	return t.file.Seek(offset, whence)
}

// Close closes the underlying file and, when armed, removes it from
// disk. It is safe to call multiple times; only the first invocation
// performs any work, so the value remains safe to defer.
//
// Returns:
//
// The first non-nil error encountered between closing and unlinking, or
// nil when both succeed (or when Close has already been called).
func (t *TempFile) Close() error {
	if t == nil || t.file == nil || t.closed {
		return nil
	}
	t.closed = true
	name := t.file.Name()
	err := t.file.Close()
	if t.removeOnClose {
		if rmErr := os.Remove(name); rmErr != nil && !os.IsNotExist(rmErr) && err == nil {
			err = rmErr
		}
	}
	return err
}

// isDir reports whether t wraps an open file that Stat confirms is
// a directory.
//
// It delegates to [TempFile.Stat], which is already nil-safe
// for a nil *[TempFile] or a TempFile with no underlying *[os.File], so any
// Stat error (nil handle, closed fd, revoked handle, permission change,
// ...) is treated conservatively as "not a directory" rather than
// propagated.
func (t *TempFile) isDir() bool {
	info, err := t.Stat()
	if err != nil || info == nil {
		return false
	}
	return info.IsDir()
}

// newTempFileWith is the single shared constructor used by every public
// TempFile factory. It applies the package's pattern fallbacks and
// returns a TempFile whose removeOnClose flag is set to remove.
func newTempFileWith(dir, pattern string, remove bool) (*TempFile, error) {
	if strutil.IsEmpty(pattern) {
		pattern = defaultTempPattern
	}
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return nil, err
	}
	return &TempFile{file: f, removeOnClose: remove}, nil
}

// wrapTempFile wraps an existing *os.File as a TempFile without creating
// a new on-disk entry. It is used by Resource.FromFile to adopt a file
// the caller already owns.
func wrapTempFile(f *os.File, remove bool) *TempFile {
	return &TempFile{file: f, removeOnClose: remove}
}

// compile-time interface assertions.
var _ ReadSeekCloser = (*TempFile)(nil)
var _ io.Writer = (*TempFile)(nil)
