package replify

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"
)

// New creates a new instance of the [wrapper] struct.
//
// This function initializes a [wrapper] struct with its default values,
// including an empty map for the `Debug` field.
//
// Returns:
//   - A pointer to a newly created [wrapper] instance with initialized fields.
func New() *wrapper {
	w := &wrapper{
		meta:   defaultMetaValues(),
		header: Processing,
	}
	return w
}

// Pages creates a new instance of the [pagination] struct.
//
// This function initializes a [pagination] struct with its default values.
//
// Returns:
//   - A pointer to a newly created [pagination] instance.
func Pages() *pagination {
	p := &pagination{}
	return p
}

// FromPages creates a new instance of the [pagination] struct
// with specified total items and items per page.
//
// This function initializes a [pagination] struct and sets
// the total number of items and items per page using the provided parameters.
//
// Parameters:
//   - totalItems: The total number of items to be paginated.
//   - perPage: The number of items to be displayed per page.
//
// Returns:
//   - A pointer to a newly created [pagination] instance with the specified settings.
func FromPages(totalItems int, perPage int) *pagination {
	return Pages().WithPerPage(perPage).WithTotalItems(totalItems)
}

// Meta creates a new instance of the [meta] struct.
//
// This function initializes a [meta] struct with its default values,
// including an empty `CustomFields` map.
//
// Returns:
//   - A pointer to a newly created [meta] instance with initialized fields.
func Meta() *meta {
	m := &meta{
		customFields: map[string]any{},
	}
	return m
}

// Header creates a new instance of the [header] struct.
//
// This function initializes a [header] struct with its default values.
//
// Returns:
//   - A pointer to a newly created [header] instance.
func Header() *header {
	h := &header{}
	return h
}

// Cursors creates a new instance of the [cursor] struct.
//
// This function initializes a [cursor] struct with its default values.
//
// Returns:
//   - A pointer to a newly created [cursor] instance.
func Cursors() *cursor {
	c := &cursor{}
	return c
}

// NewBufferPool creates new buffer pool
//
// This function initializes a `BufferPool` struct with a specified buffer size
// and pool size.
//
// Parameters:
//   - bufferSize: The size of each buffer in bytes.
//   - poolSize: The maximum number of buffers to maintain in the pool.
//
// Returns:
//
//   - A pointer to a newly created `BufferPool` instance with the specified settings.
func NewBufferPool(bufferSize int64, poolSize int) *BufferPool {
	return &BufferPool{
		buffers: make(chan []byte, poolSize),
		size:    bufferSize,
	}
}

// NewStreamConfig creates default streaming configuration
//
// This function initializes a `StreamConfig` struct with default values
// suitable for typical streaming scenarios.
//
// Returns:
//   - A pointer to a newly created `StreamConfig` instance with default settings.
func NewStreamConfig() *StreamConfig {
	return &StreamConfig{
		ChunkSize:           65536, // 64KB default
		Strategy:            StrategyBuffered,
		Compression:         CompressNone,
		UseBufferPool:       true,
		MaxConcurrentChunks: 4,
		ReadTimeout:         30 * time.Second,
		WriteTimeout:        30 * time.Second,
		ThrottleRate:        0, // unlimited
	}
}

// NewStreaming creates a new instance of the `StreamingWrapper` struct.
//
// This function initializes a `StreamingWrapper` struct with the provided
// `reader`, and `config`. If the `config` is nil, it uses default
// streaming configuration.
//
// Parameters:
//   - `reader`: An `io.Reader` instance from which data will be streamed.
//   - `config`: A pointer to a `StreamConfig` struct containing streaming configuration.
//
// Returns:
//   - A pointer to a newly created `StreamingWrapper` instance with initialized fields.
func NewStreaming(reader io.Reader, config *StreamConfig) *StreamingWrapper {
	if config == nil {
		config = NewStreamConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	sw := &StreamingWrapper{
		wrapper:        New(),
		config:         config,
		reader:         reader,
		ctx:            ctx,
		cancel:         cancel,
		progress:       &StreamProgress{},
		stats:          &StreamingStats{StartTime: time.Now()},
		errors:         make([]error, 0),
		isStreaming:    false,
		compressionBuf: make([]byte, 0),
		mu:             sync.RWMutex{},
	}

	// Initialize wrapper defaults
	sw.wrapper.WithStatusCode(http.StatusOK)
	sw.wrapper.WithMessage("Streaming initialized")
	sw.wrapper.WithDebuggingKV("streaming", true)
	sw.wrapper.WithDebuggingKV("strategy", string(config.Strategy))
	sw.wrapper.WithDebuggingKV("compression", string(config.Compression))

	// Initialize buffer pool if enabled
	if config.UseBufferPool {
		sw.bufferPool = NewBufferPool(config.ChunkSize, 4)
	}

	return sw
}

// defaultMetaValues returns the default values for the [meta] struct.
//
// This function creates a new [meta] instance with the default values,
// including the locale, API version, requested time, and request ID.
//
// Returns:
//   - A pointer to a newly created [meta] instance with the default values.
func defaultMetaValues() *meta {
	return Meta().
		WithLocale("en_US"). // vi_VN, en_US
		WithApiVersion("v0.0.1").
		WithRequestedTime(time.Now()).
		WithDeltaValue(0).
		autoRequestID()
}
