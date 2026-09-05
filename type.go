package replify

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/polarixa/replify/pkg/sysx"
)

// ///////////////////////////
// Section exported types
// ///////////////////////////

// R represents a wrapper around the main [wrapper] struct. It is used as a high-level
// abstraction to provide a simplified interface for handling API responses.
// The `R` type allows for easier manipulation of the wrapped data, metadata, and other
// response components, while maintaining the flexibility of the underlying [wrapper] structure.
type R struct {
	*wrapper
}

// H represents a wrapper around the [header] struct. It is used to encapsulate
// the header information of an API response, providing a structured way to access
// and manipulate the status code, message, type, and description of the response.
// The "H" type allows for easier handling of response headers while maintaining
// the flexibility of the underlying [header] structure.
type H struct {
	*header
}

// P represents a wrapper around the [pagination] struct. It is used to encapsulate
// pagination details for paginated API responses, providing a structured way to access
// and manipulate the current page, items per page, total pages, total items, and whether
// the current page is the last one. The "P" type allows for easier handling of pagination
// information while maintaining the flexibility of the underlying [pagination] structure.
type P struct {
	*pagination
}

// M represents a wrapper around the [meta] struct. It is used to encapsulate
// metadata information for API responses, providing a structured way to access
// and manipulate the API version, request ID, locale, requested time, custom fields,
// delta value, and delta count. The "M" type allows for easier handling of metadata
// while maintaining the flexibility of the underlying [meta] structure.
type M struct {
	*meta
}

// C represents a wrapper around the [cursor] struct. It is used to encapsulate
// pagination cursor details for navigating through paginated API responses, providing
// a structured way to access and manipulate the next cursor, previous cursor, hasNext flag,
// hasPrevious flag, and limit. The "C" type allows for easier handling of pagination cursors
// while maintaining the flexibility of the underlying [cursor] structure.
type C struct {
	*cursor
}

// ROption is a functional option for configuring a [wrapper] instance.
// Functions of this type are passed to [Wrap] to apply settings in a
// declarative, composable way.
//
// Example:
//
//	w := replify.Wrap(
//	    replify.WithStatusCode(replify.OK),
//	    replify.WithMessage("Resource retrieved"),
//	    replify.WithBody(payload),
//	)
type ROption func(*wrapper)

// Locale represents an IETF-style locale identifier
// formatted as language_COUNTRY (e.g., en_US).
type Locale string

// StreamingStrategy defines how streaming is performed
// for large datasets or long-running operations.
type StreamingStrategy string

// CompressionType defines compression algorithm
// used for data transmission.
type CompressionType string

// Frame represents a program counter inside a stack frame.
// A `Frame` is essentially a single point in the stack trace,
// representing a program counter (the location in code) at the
// time of a function call. Historically, for compatibility reasons,
// a `Frame` is interpreted as a `uintptr`, but the value stored
// in the `Frame` represents the program counter + 1. This allows
// for distinguishing between an invalid program counter and a valid one.
//
// A `Frame` is typically used within a `StackTrace` to track the
// sequence of function calls leading to the current point in the program.
// A frame is a low-level representation of a specific place in the code,
// helping in debugging by pinpointing the exact line of execution that caused
// an error or event.
//
// Example usage:
//
//	var f Frame = Frame(0x1234567890)
//	fmt.Println(f) // Prints the value of the program counter + 1
type Frame uintptr

// StackTrace represents a stack of Frames, which are ordered from the
// innermost (newest) function call to the outermost (oldest) function call
// in the current stack. It provides a high-level representation of the
// sequence of function calls (the call stack) that led to the current execution
// point, typically used for debugging or error reporting. The `StackTrace`
// contains a slice of `Frame` values, which can be interpreted as program
// counters in the call stack.
//
// The `StackTrace` can be used to generate detailed stack traces when an
// error occurs, helping developers track down the sequence of function
// calls that resulted in the error. For example, it may be used in conjunction
// with the `underlying` and `underlyingStack` types to record where an error
// occurred in the code (using `Callers()` to populate the stack) and provide
// information on the call path leading to the error.
//
// Example usage:
//
//	var trace StackTrace = StackTrace{Frame(0x1234567890), Frame(0x0987654321)}
//	fmt.Println(trace) // Prints the stack trace with the frames
type StackTrace []Frame

// BufferPool for efficient buffer reuse
type BufferPool struct {
	buffers chan []byte
	size    int64
}

// StatusCode is a typed HTTP status code.
// Use the predeclared Status* constants (e.g. [StatusOK], [StatusBadRequest])
// rather than raw integers to benefit from compile-time safety when calling
// [WithStatusCode].
type StatusCode int

// StreamConfig contains streaming configuration
// options for handling large data transfers.
type StreamConfig struct {
	// ChunkSize defines size of each chunk in bytes (default: 64KB)
	ChunkSize int64 `json:"chunk_size"`

	// IsReceiving indicates if streaming is for receiving data
	// true if receiving data (decompress incoming), false if sending (compress outgoing)
	// it's used to determine direction of data flow
	IsReceiving bool `json:"is_receiving"`

	// Strategy for streaming (direct, buffered, chunked)
	Strategy StreamingStrategy `json:"strategy"`

	// Compression algorithm to use during streaming
	Compression CompressionType `json:"compression"`

	// UseBufferPool enables buffer pooling for efficiency
	UseBufferPool bool `json:"use_buffer_pool"`

	// MaxConcurrentChunks for parallel processing
	MaxConcurrentChunks int `json:"max_concurrent_chunks"`

	// ReadTimeout for read operations
	ReadTimeout time.Duration `json:"read_timeout,omitempty"`

	// WriteTimeout for write operations
	WriteTimeout time.Duration `json:"write_timeout,omitempty"`

	// ThrottleRate in bytes/second (0 = unlimited)
	// to limit bandwidth usage during streaming
	// useful for avoiding network congestion
	ThrottleRate int64 `json:"throttle_rate"`
}

// StreamProgress tracks streaming progress
// and statistics.
type StreamProgress struct {
	// TotalBytes total data to be streamed
	TotalBytes int64 `json:"total_bytes"`

	// TransferredBytes bytes transferred so far
	TransferredBytes int64 `json:"transferred_bytes"`

	// Percentage completion (0-100)
	Percentage int `json:"percentage"`

	// CurrentChunk chunk number being processed
	CurrentChunk int64 `json:"current_chunk"`

	// TotalChunks total number of chunks
	TotalChunks int64 `json:"total_chunks"`

	// ElapsedTime time since streaming started
	ElapsedTime time.Duration `json:"elapsed_time,omitempty"`

	// EstimatedTimeRemaining estimated time until completion
	EstimatedTimeRemaining time.Duration `json:"estimated_time_remaining,omitempty"`

	// TransferRate bytes per second
	TransferRate int64 `json:"transfer_rate"`

	// LastUpdate time of last progress update
	LastUpdate time.Time `json:"last_update,omitempty"`
}

// StreamingStats contains streaming statistics
// and performance metrics.
type StreamingStats struct {
	// Time when streaming started
	StartTime time.Time `json:"start_time,omitempty"`

	// Time when streaming ended
	EndTime time.Time `json:"end_time,omitempty"`

	// Total bytes processed
	TotalBytes int64 `json:"total_bytes"`

	// Bytes after compression
	CompressedBytes int64 `json:"compressed_bytes"`

	// Compression ratio achieved
	CompressionRatio float64 `json:"compression_ratio"`

	// Average size of each chunk
	AverageChunkSize int64 `json:"average_chunk_size"`

	// Total number of chunks processed
	TotalChunks int64 `json:"total_chunks"`

	// Number of chunks that failed
	FailedChunks int64 `json:"failed_chunks"`

	// Number of chunks that were retried
	RetriedChunks int64 `json:"retried_chunks"`

	// Average latency per chunk
	AverageLatency time.Duration `json:"average_latency,omitempty"`

	// Peak bandwidth observed
	PeakBandwidth int64 `json:"peak_bandwidth"`

	// Average bandwidth during streaming
	AverageBandwidth int64 `json:"average_bandwidth"`

	// List of errors encountered during streaming
	Errors []error `json:"-"`
}

// StreamingCallback function type for async notifications
// Streamer interface for streaming data with progress tracking
type StreamingCallback func(progress *StreamProgress, err error)

// StreamingHook is a function type used for asynchronous notifications
// that provides updates on the progress of a streaming operation along with
// a reference to the associated R wrapper. This allows the callback
// to access both the progress information and any relevant response data
// encapsulated within the R type.
type StreamingHook func(progress *StreamProgress, wrap *R)

// StreamingWrapper wraps response with streaming capabilities
// BufferPool represents a pool of reusable byte buffers to optimize memory usage during streaming.
type StreamingWrapper struct {
	*wrapper
	config         *StreamConfig      // Streaming configuration
	reader         io.Reader          // Source of data to be streamed
	writer         io.Writer          // Destination for streamed data
	progress       *StreamProgress    // Progress tracking information
	stats          *StreamingStats    // Streaming statistics and metrics
	callback       StreamingCallback  // Callback for progress updates
	hook           StreamingHook      // Callback with R wrapper for progress updates
	ctx            context.Context    // Context for managing streaming lifecycle
	cancel         context.CancelFunc // Function to cancel streaming
	currentChunk   int64              // Current chunk being processed
	errors         []error            // Errors encountered during streaming
	mu             sync.RWMutex       // Mutex for synchronizing access to shared fields
	isStreaming    bool               // Indicates if streaming is in progress
	compressionBuf []byte             // Compression buffer for data compression
	bufferPool     *BufferPool        // Pool of reusable buffers for efficient memory usage
}

// StreamChunk represents a single chunk of data
// in a streaming operation.
type StreamChunk struct {
	// SequenceNumber incremental chunk number
	SequenceNumber int64 `json:"sequence_number"`

	// Data chunk content
	Data []byte `json:"-"`

	// Size of chunk
	Size int64 `json:"size"`

	// Checksum for integrity verification
	Checksum uint32 `json:"checksum"`

	// Timestamp when chunk was created
	Timestamp time.Time `json:"timestamp,omitempty"`

	// Compressed indicates if chunk is compressed
	Compressed bool `json:"compressed"`

	// CompressionType used for this chunk
	CompressionType CompressionType `json:"compression_type"`

	// Error if any occurred during chunk processing
	Error error `json:"-"`
}

// StreamingMetadata extends wrapper metadata for streaming context
// Streamer defines methods for streaming data with progress tracking
type StreamingMetadata struct {
	// Streaming strategy used
	Strategy StreamingStrategy `json:"strategy"`

	// Compression algorithm used
	CompressionType CompressionType `json:"compression_type"`

	// Size of each chunk in bytes
	ChunkSize int64 `json:"chunk_size"`

	// Total number of chunks
	TotalChunks int64 `json:"total_chunks"`

	// Estimated total size of data
	EstimatedTotalSize int64 `json:"estimated_total_size"`

	// Timestamp when streaming started
	StartedAt time.Time `json:"started_at"`

	// Timestamp when streaming completed
	CompletedAt time.Time `json:"completed_at"`

	// Indicates if streaming can be paused
	IsPausable bool `json:"is_pausable"`

	// Indicates if streaming can be resumed
	IsResumable bool `json:"is_resumable"`
}

// Dump is the thread-safe result returned by [wrapper.Dump] and
// [wrapper.DumpTo]. It owns the serialized response payload as a
// seekable, re-readable stream and guarantees that the backing temporary
// file is removed exactly once — even when Close is called concurrently.
//
// Typical usage:
//
//	dump, w := w.Dump()
//	if w.IsError() {
//	    log.Fatal(w.Error())
//	}
//	defer dump.Close()
//
//	// pipe the JSON into an HTTP response writer:
//	io.Copy(httpResponseWriter, dump.Resource().Content())
type Dump struct {
	syr      *sysx.Resource
	once     sync.Once
	closeErr error

	// filepath is non-empty when DumpTo wrote a permanent on-disk copy.
	// That file is NOT removed on Close.
	filepath string
}

// HeaderType is a string type used to represent the type of a header in an API response.
// It is used to categorize headers into different types, such as informational, error, or success headers.
type HeaderType string

// MediaType is a string type used to represent the media type (also known as MIME type) of a resource or content in an API response.
// It is used to specify the format of the data being sent or received, such as "application/json", "text/html", or "image/png".
type MediaType string

// ///////////////////////////
// Section unexported types
// ///////////////////////////

// pagination represents pagination details for paginated API responses.
type pagination struct {
	page       int  // Current page number.
	perPage    int  // Number of items per page.
	totalPages int  // Total number of pages.
	totalItems int  // Total number of items available.
	isLast     bool // Indicates whether this is the last page.
}

// meta represents metadata information about an API response.
type meta struct {
	apiVersion    string         // API version used for the request.
	requestID     string         // Unique identifier for the request, useful for debugging.
	locale        string         // Locale used for the request, e.g., "en-US".
	requestedTime time.Time      // Timestamp when the request was made.
	customFields  map[string]any // Additional custom metadata fields.
	deltaValue    float64        // Represents the magnitude of change introduced by payload normalization or transformation.
	deltaCnt      int            // Count of normalization or transformation operations performed on the payload.
}

// header represents a structured header for API responses.
type header struct {
	code        int    // Application-specific status code.
	text        string // Human-readable status text.
	typez       string // Type or category of the status, e.g., "info", "error".
	description string // Detailed description of the status.
}

// cursor represents pagination cursors for navigating through paginated API responses.
type cursor struct {
	next        string // Cursor for the next page of results.
	previous    string // Cursor for the previous page of results.
	hasNext     bool   // Indicates if there is a next page.
	hasPrevious bool   // Indicates if there is a previous page.
	limit       int    // Limit on the number of items per page.
}

// issue is the minimal, API-facing surface of a failure. It is the public
// counterpart to the internal `errors` field carried by [wrapper]: every
// detail that could leak implementation internals (raw error chains, stack
// traces, panic values) is deliberately left out, leaving only what a client
// or support engineer needs to identify, group, and reference the failure.
//
//   - ID identifies this specific occurrence (one per failed request/panic).
//   - Fingerprint identifies the category of failure (stable across
//     occurrences originating from the same code path).
//   - Message is the root-cause message, safe to display to a caller.
type issue struct {
	ID          string `json:"id,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Message     string `json:"message,omitempty"`
}

// wrapper is the main structure for wrapping API responses, including metadata, data, and debugging information.
type wrapper struct {
	statusCode int            // HTTP status code for the response.
	total      int            // Total number of items (used in non-paginated responses).
	message    string         // A message providing additional context about the response.
	data       any            // The primary data payload of the response.
	path       string         // Request path for which the response is generated.
	header     *header        // Structured header details for the response.
	meta       *meta          // Metadata about the API response.
	pagination *pagination    // Pagination details, if applicable.
	cursor     *cursor        // Pagination cursors for navigating through results.
	issue      *issue         // API-facing issue information derived from internal errors.
	debug      map[string]any // Debugging information (useful for development).
	errors     error          // Internal errors (not exposed in JSON responses).
	skipBody   bool           // When true, the body payload is omitted from String(), build(), and Slogging() output.
	cachedWrap map[string]any // Cached response data for performance optimization.
	cacheHash  string         // Hash of the cached response, used for cache validation.
	cacheMutex sync.RWMutex   // Mutex for synchronizing access to the cached response data.
	filepath   string         // Filesystem path of a file to serve via WriteFile or Write.
	filename   string         // Download filename for Content-Disposition in file and binary responses.
	span       bool           // Indicates whether to include a span in the response for tracing purposes.
}

// sequenceParticipant represents a participant in a sequence diagram, typically used for visualizing interactions between different components or services in a system.
// It contains information about the participant's unique identifier, display name, call label, and whether it is part of the runtime or user-defined code.
// This information is used to generate sequence diagrams that illustrate the flow of messages and interactions between participants in a system.
type sequenceParticipant struct {
	// id is a unique identifier for the participant in the sequence diagram, used to reference the participant in arrows and messages.
	id string

	// displayName is the human-readable name of the participant, which will be shown in the sequence diagram.
	displayName string

	// callLabel is the label for the function or method call associated with the participant, which will be shown in the sequence diagram.
	callLabel string

	// isRuntime indicates whether the participant is part of the runtime (e.g., system-level functions) or user-defined code. This information can be used to differentiate between user and system participants in the sequence diagram.
	isRuntime bool

	// location is the shortened source file and line number (e.g. "main.go:42") where this frame appears in the stack trace.
	location string
}

// stack represents a stack of program counters. It is a slice of `uintptr`
// values that store the addresses (or locations) of program counters
// from a function call stack. This is useful for tracking where errors
// or events originated in the code, and it is used by the other error
// types to capture a stack trace at the time of error creation.
//
// The stack is typically populated using a function like `Callers()`,
// which retrieves the current call stack.
type stack []uintptr

// underlying represents an error with an associated message and a stack trace.
// It is used to encapsulate the message (i.e., what went wrong) and the
// program's call stack (i.e., where the error occurred) at the point when
// the error is created. This type is used internally by functions like
// `WithError` and `WithErrorf`.
//
// Fields:
//   - msg: The error message describing what went wrong. This is a string
//     representation of the error context.
//   - stack: A pointer to a stack of program counters that represent the
//     stack trace where the error was created.
type underlying struct {
	msg    string // Message that describes the error
	*stack        // Pointer to the stack trace where the error occurred
}

// underlyingStack is an error type that contains an existing error and its
// associated stack trace. It is used when we need to annotate an existing
// error with a stack trace, preserving the original error while adding
// additional information such as the place where the annotation occurred.
//
// This type is typically used in functions like `Wrap` and `Wrapf`.
//
// Fields:
//   - error: The original error that is being wrapped or annotated.
//   - stack: A pointer to the stack trace captured at the point this
//     annotation was added.
type underlyingStack struct {
	error  // The original error being wrapped or annotated
	*stack // Pointer to the stack trace where the annotation occurred
}

// underlyingMessage represents an error with a cause (another error)
// and a message. This type is used to create errors that have both
// an associated message and a "cause," which can be another error
// that led to the current one. It is commonly used to propagate errors
// and add context to the original error message.
//
// Fields:
//   - cause: The underlying error that caused this error. This could be
//     another error returned from a function, allowing us to trace back
//     the error chain.
//   - msg: A string message describing the context of the error, which
//     can be appended to the original cause message to provide more clarity.
type underlyingMessage struct {
	cause error  // The original error being wrapped or annotated
	msg   string // The message describing the additional context for the error
}

type tools struct{}

var Toolbox tools = tools{}
