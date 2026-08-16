package sysx

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// FileOpenFlags is a convenience type for file open flags that provides a
// type-safe way to specify common flag combinations. When a function accepts
// this type, IDEs will suggest the predefined constants (e.g. CWA, RO, RW).
type FileOpenFlags int

// Command holds the fully resolved configuration for a single external process.
// Use NewCommand to create a Command, then chain With* methods to configure it
// before calling Execute, Run, or Output.
//
// Command is not safe for concurrent mutation; build a new value per goroutine
// when the same logical command must be launched from multiple goroutines.
type Command struct {
	name    string
	args    []string
	dir     string
	env     []string
	timeout time.Duration // zero means no timeout
	ctx     context.Context
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
}

// CommandResult holds the structured outcome of a completed command execution.
// All fields are always populated; zero values indicate the field was not
// applicable (e.g. ExitCode() == 0 means success, Duration() == 0 on immediate failure).
//
// Use the accessor methods to read individual result fields.
type CommandResult struct {
	stdout   string
	stderr   string
	exitCode int
	duration time.Duration
	err      error
}

// SafeFileWriter provides concurrency-safe append and overwrite operations
// targeting a single file path. A single SafeFileWriter instance can be shared
// across goroutines; all write operations are serialize by an internal mutex.
//
// Create a SafeFileWriter with NewSafeFileWriter and optionally adjust the
// file permission with WithPerm before sharing across goroutines.
type SafeFileWriter struct {
	mu   sync.Mutex
	path string
	perm os.FileMode
}

// ReadSeekCloser is the I/O contract every Resource payload must satisfy.
// It composes the three standard library interfaces required by downstream
// consumers (HTTP responses, S3 uploader, archive writers, Telegram bots,
// …) without binding the producer to a concrete type such as *os.File.
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// Resource is the storage-agnostic envelope returned by every data-exporting
// workflow (reports, dumps, archives, attachments, backups, media,
// rendered documents). All fields are unexported; use the constructor
// NewResource and the chainable With* methods to populate the instance,
// and the accessor methods to read it.
//
// A Resource owns its underlying ReadSeekCloser; consumers must invoke
// Close exactly once to release file handles, delete temporary files, or
// free buffers as appropriate to the backing implementation.
type Resource struct {
	mu          sync.Mutex
	name        string
	size        int64
	contentType string
	content     ReadSeekCloser

	// spillThreshold is the maximum size of in-memory buffers before spilling to disk.
	// It is applied by Resource.FromReader and is ignored by other builders that produce a concrete ReadSeekCloser.
	spillThreshold int64

	// defaultTempPattern is the filename pattern used by spillBuffer when creating temporary files.
	// It is ignored by other builders that produce a concrete ReadSeekCloser.
	tempPattern string

	// tempDir is the directory used by spillBuffer when creating temporary files.
	// It is ignored by other builders that produce a concrete ReadSeekCloser.
	tempDir string

	// removeOnClose indicates whether the Resource's content should be removed from disk when Close is called.
	// It is applied by spillBuffer and TempFile and ignored by MemBlob.
	removeOnClose bool
}

// MemBlob is an in-memory ReadSeekCloser backed by a byte slice. It is the
// cheapest backing implementation and is ideal for small payloads that
// comfortably fit in process memory. Close is a no-op; the underlying
// slice is released when the MemBlob becomes unreachable.
type MemBlob struct {
	data   []byte
	reader *bytes.Reader
}

// TempFile is a ReadSeekCloser backed by an on-disk temporary file. It is
// the only public type in sysx that wraps an *os.File for the purpose of
// satisfying the Resource contract; outside of sysx, callers must depend
// on Resource and ReadSeekCloser instead.
//
// When removeOnClose is true the underlying file is unlinked on the first
// Close call. Subsequent Close calls are no-ops so the value remains safe
// to defer.
type TempFile struct {
	file          *os.File
	removeOnClose bool
	closed        bool
}

// spillBuffer is a private ReadSeekCloser that begins life as a memory
// buffer and transparently spills to a temporary file once the configured
// threshold is exceeded. It promotes a plain io.Reader into a seekable,
// closeable stream — required by Resource — without committing to either
// memory or disk up-front.
type spillBuffer struct {
	mem       *bytes.Buffer
	threshold int64
	file      *TempFile
	reader    io.ReadSeeker
	size      int64
}

// commandBuffer is a minimal zero-allocation-friendly byte accumulator used
// internally by Execute to capture stdout and stderr. It satisfies io.Writer.
type commandBuffer struct {
	buf strings.Builder
}

// fileMutexes is the package-level registry of per-path in-process mutexes used
// by WriteFileLocked to serialize concurrent writes to the same file path.
// It is populated lazily by getFileMutex and is safe for concurrent access.
var fileMutexes sync.Map // map[string]*sync.Mutex
