package slogger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/polarixa/replify/pkg/sysx"
)

// ///////////////////////////////////////////////////////////////////////////
// RotationOptions accessors
// ///////////////////////////////////////////////////////////////////////////

// Directory returns the log directory.
//
// Returns:
//
// the dir value.
func (r *RotationOptions) Directory() string {
	if r == nil {
		return ""
	}
	return r.dir
}

// SetDirectory sets the log directory.
//
// Parameters:
//   - `dir`: the log directory path
func (r *RotationOptions) SetDirectory(dir string) {
	if r == nil {
		return
	}
	r.dir = dir
}

// MaxBytes returns the maximum file size before rotation.
//
// Returns:
//
// the maxBytes value.
func (r *RotationOptions) MaxBytes() int64 {
	if r == nil {
		return 0
	}
	return r.maxBytes
}

// SetMaxBytes sets the maximum file size before rotation.
//
// Parameters:
//   - `maxBytes`: the maximum size in bytes
func (r *RotationOptions) SetMaxBytes(maxBytes int64) {
	if r == nil {
		return
	}
	r.maxBytes = maxBytes
}

// MaxAge returns the maximum age before rotation.
//
// Returns:
//
// the maxAge value.
func (r *RotationOptions) MaxAge() time.Duration {
	if r == nil {
		return 0
	}
	return r.maxAge
}

// SetMaxAge sets the maximum age before rotation.
//
// Parameters:
//   - `maxAge`: the maximum age duration
func (r *RotationOptions) SetMaxAge(age time.Duration) {
	if r == nil {
		return
	}
	r.maxAge = age
}

// IsCompress returns whether compression is enabled.
//
// Returns:
//
// true if compression is enabled.
func (r *RotationOptions) IsCompress() bool {
	if r == nil {
		return false
	}
	return r.compress
}

// SetCompress enables or disables compression.
//
// Parameters:
//   - `compress`: whether to compress rotated files
func (r *RotationOptions) SetCompress(compress bool) {
	if r == nil {
		return
	}
	r.compress = compress
}

// WithDirectory sets the log directory and returns the receiver for chaining.
//
// Parameters:
//   - `dir`: the log directory path
//
// Returns:
//
// the receiver, for method chaining.
func (r *RotationOptions) WithDirectory(dir string) *RotationOptions {
	r.SetDirectory(dir)
	return r
}

// WithMaxBytes sets the maximum file size before rotation and returns the receiver for chaining.
//
// Parameters:
//   - `maxBytes`: the maximum size in bytes
//
// Returns:
//
// the receiver, for method chaining.
func (r *RotationOptions) WithMaxBytes(maxBytes int64) *RotationOptions {
	r.SetMaxBytes(maxBytes)
	return r
}

// WithMaxAge sets the maximum age before rotation and returns the receiver for chaining.
//
// Parameters:
//   - `maxAge`: the maximum age duration
//
// Returns:
//
// the receiver, for method chaining.
func (r *RotationOptions) WithMaxAge(maxAge time.Duration) *RotationOptions {
	r.SetMaxAge(maxAge)
	return r
}

// WithCompress enables or disables compression and returns the receiver for chaining.
//
// Parameters:
//   - `compress`: whether to compress rotated files
//
// Returns:
//
// the receiver, for method chaining.
func (r *RotationOptions) WithCompress(compress bool) *RotationOptions {
	r.SetCompress(compress)
	return r
}

// ///////////////////////////////////////////////////////////////////////////
// LevelFileWriter accessors
// ///////////////////////////////////////////////////////////////////////////

// Options returns the rotation options used by this writer.
//
// Returns:
//
// a copy of the RotationOptions.
func (lfw *LevelFileWriter) Options() RotationOptions {
	if lfw == nil {
		return RotationOptions{}
	}
	lfw.mu.Lock()
	defer lfw.mu.Unlock()
	return lfw.opts
}

// ///////////////////////////////////////////////////////////////////////////
// LevelWriterHook accessors
// ///////////////////////////////////////////////////////////////////////////

// Writer returns the underlying LevelFileWriter.
//
// Returns:
//
// the *LevelFileWriter used by this hook.
func (h *LevelWriterHook) Writer() *LevelFileWriter {
	if h == nil {
		return nil
	}
	return h.writer
}

// Formatter returns the formatter used by this hook.
//
// Returns:
//
// the Formatter used to serialise entries.
func (h *LevelWriterHook) Formatter() Formatter {
	if h == nil {
		return nil
	}
	return h.formatter
}

// SetFormatter sets the formatter used by this hook.
//
// Parameters:
//   - `formatter`: the formatter to use
func (h *LevelWriterHook) SetFormatter(formatter Formatter) {
	if h == nil {
		return
	}
	h.formatter = formatter
}

// Levels implements Hook by returning the set of log levels this hook handles.
//
// Returns:
//
// the slice of Level values for which this hook's Fire method will be called.
func (h *LevelWriterHook) Levels() []Level {
	return h.levels
}

// Fire implements Hook by serialising the entry with the configured Formatter
// and writing the result to the appropriate level-specific file.
//
// Parameters:
//   - `e`: the log entry to write
//
// Returns:
//
// an error if formatting or writing fails; nil on success.
func (h *LevelWriterHook) Fire(e *Entry) error {
	data, err := h.formatter.Format(e)
	if err != nil {
		return err
	}
	lvl := e.Level()
	if _, err = h.writer.WriteLevel(lvl, data); err != nil {
		return fmt.Errorf("slogger: LevelWriterHook write failed for level %s: %w", lvl, err)
	}
	return nil
}

// Write implements io.Writer by routing p to the InfoLevel file.
// This satisfies io.Writer so that a LevelFileWriter can be used as a generic
// writer destination.
//
// Parameters:
//   - `p`: the bytes to write
//
// Returns:
//
// the number of bytes written and any write error.
func (lfw *LevelFileWriter) Write(p []byte) (int, error) {
	return lfw.WriteLevel(InfoLevel, p)
}

// WriteLevel routes p to the file for the given level.
// If no exact file exists for level, the nearest coarser-grained file is used
// (Trace → Debug, Fatal/Panic → Error).
//
// Parameters:
//   - `level`: the severity level determining which file receives p
//   - `p`: the bytes to write
//
// Returns:
//
// the number of bytes written and any write error.
func (lfw *LevelFileWriter) WriteLevel(level Level, p []byte) (int, error) {
	lfw.mu.Lock()
	rf, ok := lfw.writers[level]
	if !ok {
		switch {
		case level <= DebugLevel:
			rf = lfw.writers[DebugLevel]
		case level <= InfoLevel:
			rf = lfw.writers[InfoLevel]
		case level <= WarnLevel:
			rf = lfw.writers[WarnLevel]
		default:
			rf = lfw.writers[ErrorLevel]
		}
	}
	if rf == nil {
		lfw.mu.Unlock()
		return 0, nil
	}
	// Hold lfw.mu until after the write completes so that a concurrent
	// Close cannot invalidate rf between the lookup and the write.
	defer lfw.mu.Unlock()
	return rf.write(p)
}

// Close closes all open file handles managed by this writer.
// After Close, the LevelFileWriter must not be used.
//
// Returns:
//
// the first error encountered while closing files, or nil if all succeed.
func (lfw *LevelFileWriter) Close() error {
	lfw.mu.Lock()
	defer lfw.mu.Unlock()
	var first error
	for _, rf := range lfw.writers {
		if err := rf.close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// Rotate forces immediate rotation of all level files, regardless of their
// current size or age. Useful for external rotation signals (e.g. SIGHUP).
//
// Returns:
//
// the first rotation error encountered, or nil if all files rotated successfully.
func (lfw *LevelFileWriter) Rotate() error {
	lfw.mu.Lock()
	defer lfw.mu.Unlock()
	var first error
	for _, rf := range lfw.writers {
		if err := rf.rotate(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// close flushes and releases the underlying file handle held by this
// rotatingFile. It acquires the instance mutex to ensure the operation is
// safe for concurrent callers. If the file is already closed (or was never
// opened), close is a no-op and returns nil. After a successful call the
// internal file reference is set to nil so that subsequent writes will
// return an error rather than attempt to write to a closed descriptor.
//
// Returns:
//
// any error returned by the OS when closing the file handle, or nil on success.
func (rf *rotatingFile) close() error {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.file != nil {
		err := rf.file.Close()
		rf.file = nil
		return err
	}
	return nil
}

// open creates or re-opens the log file at rf.path for appending and
// initialises the bookkeeping fields used by the rotation policy.
// The file is opened with O_CREATE|O_APPEND|O_WRONLY and mode 0644, so it is
// created if it does not already exist, and every write is appended to the end
// of any existing content.
// After a successful open, rf.file holds the active file handle, rf.size is
// set to the current on-disk file size (as reported by os.Stat), and
// rf.openedAt is set to the current wall-clock time so that age-based
// rotation can be evaluated correctly.
//
// Returns:
//
// an error if the file cannot be opened or stat'd; nil on success.
func (rf *rotatingFile) open() error {
	f, err := os.OpenFile(rf.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("slogger: cannot open log file %q: %w", rf.path, err)
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return fmt.Errorf("slogger: cannot stat log file %q: %w", rf.path, err)
	}
	rf.file = f
	rf.size = info.Size()
	rf.openedAt = time.Now()
	return nil
}

// shouldRotate reports whether the current log file should be rotated before
// writing an additional incoming bytes. Two independent policies are evaluated:
//
//   - Size policy: if rf.maxBytes is positive and the projected file size after
//     the write (rf.size + incoming) would exceed rf.maxBytes, rotation is required.
//   - Age policy: if rf.maxAge is positive and the time elapsed since the file
//     was last opened exceeds rf.maxAge, rotation is required.
//
// A policy is disabled when its threshold is zero or negative, so callers can
// opt out of either constraint independently. Both policies are checked on
// every call; if either is satisfied the method returns true immediately.
//
// Parameters:
//   - `incoming`: the number of bytes about to be written
//
// Returns:
//
// true if the file must be rotated before the write; false otherwise.
func (rf *rotatingFile) shouldRotate(incoming int64) bool {
	if rf.maxBytes > 0 && rf.size+incoming > rf.maxBytes {
		return true
	}
	if rf.maxAge > 0 && time.Since(rf.openedAt) > rf.maxAge {
		return true
	}
	return false
}

// write appends p to the active log file, rotating it first if the current
// rotation policy requires it. The method acquires the instance mutex for the
// duration of the call, making it safe to call concurrently.
//
// Before each write, needsRotation is consulted with the byte count of p. If
// rotation is needed, rotate is called to archive the current file and open a
// fresh one; any rotation error aborts the write and is returned immediately.
// After a successful write, rf.size is incremented by the number of bytes
// actually written so that subsequent size-policy checks remain accurate.
//
// Parameters:
//   - `p`: the bytes to append to the log file
//
// Returns:
//
// the number of bytes written and any error from the underlying file write.
func (rf *rotatingFile) write(p []byte) (int, error) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.shouldRotate(int64(len(p))) {
		if err := rf.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := rf.file.Write(p)
	rf.size += int64(n)
	return n, err
}

// rotate archives the current log file and opens a fresh one at rf.path.
// The sequence is: close → archive → open.
//
//  1. Close: if the file is currently open it is flushed and closed. Any close
//     error is returned immediately and no further steps are taken.
//  2. Archive directory: a date-stamped sub-directory is created (or reused)
//     under rf.dir/defaultArchivedDir using archiveDateFormat as the folder
//     name, with permission 0755.
//  3. Archive the file — two modes depending on rf.compress:
//     - Compressed (rf.compress == true): the current log is packed into a ZIP
//     file named "<timestamp>_<level>.zip" via compressToZip, then the
//     original log file is removed. If compression fails the error is returned
//     and the original file is left intact.
//     - Plain (rf.compress == false): the current log is renamed to
//     "<timestamp>_<level>.log" using os.Rename. If the rename fails the
//     error is returned.
//  4. Open: rf.open is called to create a new, empty log file at rf.path and
//     reset the size and timestamp bookkeeping fields.
//
// rotate does not acquire rf.mu; callers (write, Rotate) are responsible for
// holding the appropriate lock before calling this method.
//
// Returns:
//
// the first error encountered across any of the steps above; nil on success.
func (rf *rotatingFile) rotate() error {
	if rf.file != nil {
		if err := rf.file.Close(); err != nil {
			return fmt.Errorf("slogger: cannot close log file for rotation: %w", err)
		}
		rf.file = nil
	}

	now := time.Now()
	dateDir := filepath.Join(rf.dir, defaultArchivedDir, now.Format(archiveDateFormat))
	if !sysx.DirExists(dateDir) {
		if err := os.MkdirAll(dateDir, 0755); err != nil {
			return fmt.Errorf("slogger: cannot create archive dir %q: %w", dateDir, err)
		}
	}

	stamp := now.Format(archiveStampFormat)
	levelName := strings.ToLower(rf.level.String())

	if rf.compress {
		zipPath := filepath.Join(dateDir, fmt.Sprintf("%s_%s.zip", stamp, levelName))
		if err := compressToZip(rf.path, zipPath); err != nil {
			return fmt.Errorf("slogger: archive compression failed: %w", err)
		}
		_ = os.Remove(rf.path)
	} else {
		archivePath := filepath.Join(dateDir, fmt.Sprintf("%s_%s.log", stamp, levelName))
		if err := os.Rename(rf.path, archivePath); err != nil {
			return fmt.Errorf("slogger: cannot archive log file: %w", err)
		}
	}

	return rf.open()
}
