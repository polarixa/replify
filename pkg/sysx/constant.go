package sysx

import (
	"errors"
	"os"
)

const (
	// os.O_CREATE | os.O_WRONLY | os.O_APPEND
	// Flag for opening a file for writing with append mode and creating it if it doesn't exist.
	CWA FileOpenFlags = FileOpenFlags(os.O_CREATE | os.O_WRONLY | os.O_APPEND)

	// os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	// Flag for opening a file for writing with truncate mode and creating it if it doesn't exist.
	CWT FileOpenFlags = FileOpenFlags(os.O_CREATE | os.O_WRONLY | os.O_TRUNC)

	// os.O_CREATE | os.O_WRONLY
	// Flag for opening a file for writing and creating it if it doesn't exist.
	CW FileOpenFlags = FileOpenFlags(os.O_CREATE | os.O_WRONLY)

	// os.O_APPEND | os.O_WRONLY
	// Flag for opening a file for writing with append mode.
	AW FileOpenFlags = FileOpenFlags(os.O_APPEND | os.O_WRONLY)

	// os.O_TRUNC | os.O_WRONLY
	// Flag for opening a file for writing with truncate mode.
	TW FileOpenFlags = FileOpenFlags(os.O_TRUNC | os.O_WRONLY)

	// os.O_RDONLY
	// Flag for opening a file for reading only.
	RO FileOpenFlags = FileOpenFlags(os.O_RDONLY)

	// os.O_RDWR
	// Flag for opening a file for reading and writing.
	RW FileOpenFlags = FileOpenFlags(os.O_RDWR)

	// os.O_RDWR | os.O_CREATE
	// Flag for opening a file for reading and writing and creating it if it doesn't exist.
	CRW FileOpenFlags = FileOpenFlags(os.O_RDWR | os.O_CREATE)

	// os.O_RDWR | os.O_CREATE | os.O_TRUNC
	// Flag for opening a file for reading and writing with truncate mode and creating it if it doesn't exist.
	CRWT FileOpenFlags = FileOpenFlags(os.O_RDWR | os.O_CREATE | os.O_TRUNC)

	// os.O_RDWR | os.O_APPEND
	// Flag for opening a file for reading and writing with append mode.
	ARW FileOpenFlags = FileOpenFlags(os.O_RDWR | os.O_APPEND)

	// os.O_RDWR | os.O_APPEND | os.O_TRUNC
	// Flag for opening a file for reading and writing with append and truncate mode and creating it if it doesn't exist.
	TARW FileOpenFlags = FileOpenFlags(os.O_RDWR | os.O_APPEND | os.O_TRUNC)

	// os.O_RDWR | os.O_TRUNC
	// Flag for opening a file for reading and writing with truncate mode.
	TRW FileOpenFlags = FileOpenFlags(os.O_RDWR | os.O_TRUNC)

	// os.O_RDWR | os.O_CREATE | os.O_APPEND
	// Flag for opening a file for reading and writing with append mode and creating it if it doesn't exist.
	CRWA FileOpenFlags = FileOpenFlags(os.O_RDWR | os.O_CREATE | os.O_APPEND)
)

// IANA media-type constants used by Resource producers and consumers. They
// are declared as untyped string constants so callers may pass them in any
// string context without conversion.
const (
	MimeOctetStream = "application/octet-stream"
	MimeText        = "text/plain; charset=utf-8"
	MimeCSV         = "text/csv; charset=utf-8"
	MimeJSON        = "application/json"
	MimeXML         = "application/xml"
	MimeHTML        = "text/html; charset=utf-8"
	MimePDF         = "application/pdf"
	MimeZIP         = "application/zip"
	MimeGZIP        = "application/gzip"
	MimeSQL         = "application/x-sql"
)

var (
	// ErrNilResource is returned by Resource helpers when the receiver, its
	// Content, or a required argument is nil.
	ErrNilResource = errors.New("sysx: nil resource or nil content")
)

// DefaultSpillThreshold is the default in-memory ceiling used by Resource
// builders before spilling the remainder of a stream onto a temporary
// file. It balances avoidance of disk I/O for typical exports against
// protection from runaway memory usage when streaming arbitrary
// producers.
const DefaultSpillThreshold int64 = 8 << 20 // 8 MiB

// defaultTempPattern is the fall-back name pattern handed to os.CreateTemp
// when a Resource builder is asked to create a temp file without one.
const defaultTempPattern = "sysx-resource-*"
