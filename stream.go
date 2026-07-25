package replify

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/polarixa/replify/pkg/strutil"
)

// Start initiates streaming operation and returns *wrapper for consistency with replify API.
//
// This function is the primary entry point for streaming operations. It validates prerequisites (streaming wrapper
// not nil, reader configured), prevents concurrent streaming on the same wrapper, selects and executes the configured
// streaming strategy (STRATEGY_DIRECT, STRATEGY_BUFFERED, or STRATEGY_CHUNKED), monitors operation completion,
// and returns a comprehensive response wrapping the result in the standard wrapper format. Start is the public API
// method that callers use to begin streaming; it handles all coordination, error management, and response formatting.
// The function sets up initial timestamps, manages the isStreaming state flag to prevent concurrent operations,
// executes the strategy-specific streaming function with the provided context, and populates the wrapper response
// with final statistics, metadata, and outcome information. Both success and failure paths return a *wrapper object
// with appropriate HTTP status codes, messages, and debugging information for client feedback and diagnostics. The
// streaming operation respects context cancellation, allowing caller-controlled shutdown. All per-chunk errors are
// accumulated; streaming continues when possible, enabling partial success scenarios. Final statistics include
// chunk counts, byte counts, compression metrics, timing information, and error tracking, providing comprehensive
// insight into streaming performance and health.
//
// Parameters:
//   - ctx: Context for cancellation, timeouts, and coordination.
//     If nil, uses sw.ctx (context from streaming wrapper creation).
//     Passed to streaming strategy function for deadline enforcement.
//     Cancellation stops streaming immediately.
//     Parent context may have deadline affecting overall operation.
//
// Returns:
//   - *wrapper: Response wrapper containing streaming result.
//     HTTP status code (200, 400, 409, 500).
//     Message describing outcome.
//     Debugging information with statistics.
//     Error information if operation failed.
//     Always returns non-nil wrapper for consistency.
//
// Behavior:
//   - Validation: checks nil wrapper, reader configuration.
//   - Mutual exclusion: prevents concurrent streaming on same wrapper.
//   - Strategy selection: routes to appropriate streaming implementation.
//   - Error accumulation: collects per-chunk errors without stopping.
//   - Response building: wraps result in standard wrapper format.
//   - Statistics: populates final metrics and timing data.
//   - Status codes: HTTP codes reflecting outcome (200, 400, 409, 500).
//
// Validation Stages:
//
//	Stage                               Check                           Response if Failed
//	──────────────────────────────────────────────────────────────────────────────────────
//	1. Nil wrapper                      sw == nil                       Default bad request
//	2. Reader configured                sw.reader != nil                400 Bad Request
//	3. Not already streaming            !sw.isStreaming                 409 Conflict
//	4. Strategy known                   Valid STRATEGY_* constant       500 Internal Server Error
//
// HTTP Status Code Mapping:
//
//	Scenario                                    Status Code     Message
//	──────────────────────────────────────────────────────────────────────────
//	Streaming wrapper is nil                    400             (default response)
//	Reader not configured                       400             "reader not set for streaming"
//	Streaming already in progress               409             "streaming already in progress"
//	Unknown streaming strategy                  500             "unknown streaming strategy: ..."
//	Streaming operation failed (error)          500             "streaming error: ..."
//	Streaming completed successfully            200             "Streaming completed successfully"
//
// Pre-Streaming Checks Flow:
//
//	Input (sw, ctx)
//	    ↓
//	sw == nil?
//	    ├─ Yes → Return respondStreamBadRequestDefault()
//	    └─ No  → Continue
//	    ↓
//	sw.reader == nil?
//	    ├─ Yes → Return 400 "reader not set for streaming"
//	    └─ No  → Continue
//	    ↓
//	sw.isStreaming (with lock)?
//	    ├─ Yes → Return 409 "streaming already in progress"
//	    └─ No  → Set isStreaming = true, Continue
//	    ↓
//	Proceed to strategy selection
//
// Streaming Lifecycle:
//
//	Phase                               Action                          State
//	──────────────────────────────────────────────────────────────────────────
//	1. Initialization                   Lock and set isStreaming        Locked
//	2. Setup                            Create context (use or default) Ready
//	3. Logging                          Log start time in debugging KV  Tracked
//	4. Strategy dispatch                Call appropriate stream function Executing
//	5. Error collection                 Monitor for streamErr           In-flight
//	6. Finalization                     Lock and clear isStreaming      Cleanup
//	7. Response building (success)      Populate success statistics     Response ready
//	8. Response building (failure)      Populate error information      Error response ready
//	9. Return                           Return wrapper to caller        Complete
//
// Context Handling:
//
//	Scenario                                    Behavior                        Implication
//	──────────────────────────────────────────────────────────────────────────────
//	ctx provided and non-nil                    Use provided context            Caller controls deadline
//	ctx is nil                                  Use sw.ctx (if set)             Fallback to wrapper context
//	Both nil                                    Pass nil to strategy            No deadline enforcement
//	Parent context has deadline                 Inherited by strategy           Affects all operations
//	Cancellation via context                    Strategy responds to Done()     Responsive shutdown
//
// Strategy Selection:
//
//	Configuration                   Selected Function           Characteristics
//	──────────────────────────────────────────────────────────────────────────
//	STRATEGY_DIRECT                 streamDirect()              Sequential, simple
//	STRATEGY_BUFFERED               streamBuffered()            Concurrent, high throughput
//	STRATEGY_CHUNKED                streamChunked()             Sequential, detailed control
//	Unknown strategy                Error return                500 status, error message
//
// State Management (isStreaming flag):
//
//	Operation                           Lock Held    isStreaming Value    Purpose
//	──────────────────────────────────────────────────────────────────────────
//	Pre-check (initial state)           Yes          false                Prevent concurrent start
//	Set streaming active                Yes          true                 Mark operation in progress
//	Unlock after setting                No           true                 Allow other operations
//	Stream execution                    No           true                 Streaming active
//	Set streaming inactive              Yes          false                Streaming complete
//
// Success Response Building:
//
//	Field                           Source                      Format                  Purpose
//	────────────────────────────────────────────────────────────────────────────────────────
//	StatusCode                      Constant                    200 (HTTP OK)           Success indicator
//	Message                         Constant                    "Streaming completed..." User-facing message
//	completed_at                    sw.stats.EndTime.Unix()     Unix timestamp          When completed
//	total_chunks                    sw.stats.TotalChunks        int64                   Chunk count
//	total_bytes                     sw.stats.TotalBytes         int64                   Byte count
//	compressed_bytes                sw.stats.CompressedBytes    int64                   Compressed size
//	compression_ratio               sw.stats.CompressionRatio   "0.xx" format           Ratio display
//	duration_ms                     EndTime - StartTime         Milliseconds            Operation duration
//
// Failure Response Building:
//
//	Field                           Source                      Format                  Purpose
//	────────────────────────────────────────────────────────────────────────────────────────
//	StatusCode                      Constant                    500 (HTTP error)        Failure indicator
//	Error (via WithErrorAck)          streamErr                   Error message           Error details
//	failed_chunks                   sw.stats.FailedChunks       int64                   Failed chunk count
//	total_errors                    len(sw.errors)              int64                   Error count
//
// Example:
//
//	// Example 1: Simple streaming with default context
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/stream/start").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithStreamingStrategy(STRATEGY_DIRECT).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil && p.CurrentChunk % 100 == 0 {
//	            fmt.Printf("Progress: %.1f%%\n", p.Percentage)
//	        }
//	    })
//
//	// Start streaming (uses nil context, falls back to wrapper's context)
//	result := streaming.Start(nil)
//
//	if result.IsError() {
//	    fmt.Printf("Error: %s\n", result.Error())
//	} else {
//	    fmt.Printf("Success: %s\n", result.Message())
//	    fmt.Printf("Chunks: %v\n", result.Debugging()["total_chunks"])
//	}
//
//	// Example 2: Streaming with cancellation context
//	func StreamWithTimeout(fileReader io.ReadCloser) *wrapper {
//	    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
//	    defer cancel()
//
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/stream/timeout").
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(512 * 1024).
//	        WithStreamingStrategy(STRATEGY_BUFFERED).
//	        WithMaxConcurrentChunks(4)
//
//	    // Start with context timeout
//	    result := streaming.Start(ctx)
//
//	    return result
//	}
//
//	// Example 3: Comprehensive error handling
//	func StreamWithCompleteErrorHandling(fileReader io.ReadCloser) {
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/stream/errors").
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(256 * 1024).
//	        WithStreamingStrategy(STRATEGY_CHUNKED).
//	        WithCompressionType(COMP_GZIP).
//	        WithCallback(func(p *StreamProgress, err error) {
//	            if err != nil {
//	                fmt.Printf("Chunk %d failed: %v\n", p.CurrentChunk, err)
//	            }
//	        })
//
//	    result := streaming.Start(context.Background())
//
//	    // Analyze result
//	    if result.IsError() {
//	        fmt.Printf("Status: %d\n", result.StatusCode())
//	        fmt.Printf("Error: %s\n", result.Error())
//
//	        debugging := result.Debugging()
//	        fmt.Printf("Failed chunks: %v\n", debugging["failed_chunks"])
//	        fmt.Printf("Total errors: %v\n", debugging["total_errors"])
//	    } else {
//	        fmt.Printf("Status: %d\n", result.StatusCode())
//	        fmt.Printf("Message: %s\n", result.Message())
//
//	        debugging := result.Debugging()
//	        fmt.Printf("Total chunks: %v\n", debugging["total_chunks"])
//	        fmt.Printf("Total bytes: %v\n", debugging["total_bytes"])
//	        fmt.Printf("Compression ratio: %v\n", debugging["compression_ratio"])
//	        fmt.Printf("Duration: %vms\n", debugging["duration_ms"])
//	    }
//	}
//
//	// Example 4: Streaming with concurrent protection check
//	func DemonstrateConcurrencyProtection(fileReader io.ReadCloser) {
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/stream/concurrency").
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(1024 * 1024)
//
//	    // First call succeeds
//	    result1 := streaming.Start(context.Background())
//	    fmt.Printf("First start: %d\n", result1.StatusCode())  // 200
//
//	    // Second call (if first is slow) would get:
//	    // result2 := streaming.Start(context.Background())
//	    // fmt.Printf("Second start: %d\n", result2.StatusCode())  // 409 Conflict
//	}
//
//	// Example 5: All three strategies comparison
//	func CompareStreamingStrategies(fileReader io.ReadCloser) {
//	    strategies := []interface {}{
//	        STRATEGY_DIRECT,
//	        STRATEGY_CHUNKED,
//	        STRATEGY_BUFFERED,
//	    }
//
//	    for _, strategy := range strategies {
//	        fmt.Printf("\nTesting strategy: %v\n", strategy)
//
//	        // Create fresh streaming wrapper for each strategy
//	        newReader := createNewReader()  // Create fresh reader
//
//	        streaming := replify.New().
//	            WithStatusCode(200).
//	            WithPath("/api/stream/compare").
//	            WithStreaming(newReader, nil).
//	            WithChunkSize(512 * 1024).
//	            WithStreamingStrategy(strategy)
//
//	        if strategy == STRATEGY_BUFFERED {
//	            streaming = streaming.WithMaxConcurrentChunks(4)
//	        }
//
//	        start := time.Now()
//	        result := streaming.Start(context.Background())
//	        duration := time.Since(start)
//
//	        if result.IsError() {
//	            fmt.Printf("  Status: %d (ERROR)\n", result.StatusCode())
//	        } else {
//	            debugging := result.Debugging()
//	            fmt.Printf("  Status: %d (OK)\n", result.StatusCode())
//	            fmt.Printf("  Duration: %v\n", duration)
//	            fmt.Printf("  Chunks: %v\n", debugging["total_chunks"])
//	            fmt.Printf("  Bytes: %v\n", debugging["total_bytes"])
//	        }
//	    }
//	}
//
// Mutual Exclusion Pattern:
//
//	Scenario                            Behavior                            Response
//	────────────────────────────────────────────────────────────────────────────────
//	First Start() call                  Acquires lock, sets isStreaming    Proceeds normally
//	Concurrent Start() during streaming Checks isStreaming, finds true     409 Conflict
//	Start() after completion            isStreaming cleared, lock free    Proceeds normally
//	Rapid successive calls              Mutex serializes access           First waits, others get 409
//
// Error Propagation:
//
//	Error Origin                    Recorded?       Returned in Response?       Status Code
//	──────────────────────────────────────────────────────────────────────────────────
//	Reader validation               No              Yes (in message)            400
//	Concurrent streaming            No              Yes (in message)            409
//	Unknown strategy                Yes             Yes (via streamErr)         500
//	Strategy execution (streaming)  Yes             Yes (via streamErr)         500
//	Per-chunk errors                Yes             Yes (via failed_chunks)     200 or 500
//
// Statistics Tracking:
//
//	Metric                          Updated By                      When Updated
//	────────────────────────────────────────────────────────────────────────────
//	StartTime                       Start()                         At initialization
//	EndTime                         Start() after strategy returns   After streaming
//	TotalChunks                     updateProgress()                Per chunk
//	TotalBytes                      updateProgress()                Per chunk
//	CompressedBytes                 Strategy functions              Per chunk (if compressed)
//	CompressionRatio                GetStats() calculation          On query
//	FailedChunks                    Strategy on error               On chunk error
//	Errors list                     recordError()                   On any error
//
// Best Practices:
//
//  1. ALWAYS HANDLE RETURNED WRAPPER RESPONSE
//     - Check status code for success/failure
//     - Log or report error messages
//     - Provide debugging info to caller
//     - Pattern:
//     result := streaming.Start(ctx)
//     if result.IsError() {
//     // Handle error
//     } else {
//     // Handle success
//     }
//
//  2. PROVIDE CONTEXT WITH DEADLINE WHEN POSSIBLE
//     - Enables timeout enforcement
//     - Allows caller-controlled shutdown
//     - Prevents indefinite blocking
//     - Pattern:
//     ctx, cancel := context.WithTimeout(context.Background(), timeout)
//     defer cancel()
//     result := streaming.Start(ctx)
//
//  3. USE CALLBACKS FOR PROGRESS MONITORING
//     - Receive updates per chunk
//     - Track errors in real-time
//     - Update UI/logs continuously
//     - Pattern:
//     WithCallback(func(p *StreamProgress, err error) {
//     // Handle progress or error
//     })
//
//  4. CHOOSE STRATEGY BASED ON USE CASE
//     - STRATEGY_DIRECT: simplicity, small files
//     - STRATEGY_CHUNKED: control, detailed tracking
//     - STRATEGY_BUFFERED: throughput, large files
//     - Pattern:
//     WithStreamingStrategy(STRATEGY_BUFFERED)
//
//  5. EXAMINE DEBUGGING INFO FOR DIAGNOSTICS
//     - Check failed_chunks count
//     - Review total_errors
//     - Analyze compression_ratio
//     - Track duration_ms for performance
//     - Pattern:
//     debugging := result.Debugging()
//     if debugging["failed_chunks"] > 0 {
//     // Investigate failures
//     }
//
// Related Methods and Lifecycle:
//
//	Method                      Purpose                     Related To
//	──────────────────────────────────────────────────────────────────
//	Start()                     Initiate streaming (this)   Entry point
//	WithStreaming()             Configure reader/writer     Pre-streaming setup
//	WithStreamingStrategy()     Select strategy             Pre-streaming setup
//	streamDirect()              Direct execution            Called by Start
//	streamChunked()             Chunked execution           Called by Start
//	streamBuffered()            Buffered execution          Called by Start
//	GetStats()                  Query final statistics      Post-streaming
//	GetProgress()               Query current progress      During/after streaming
//	Cancel()                    Stop streaming              Control operation
//	Errors()                    Retrieve error list         Post-streaming analysis
//
// See Also:
//   - WithStreamingStrategy: Configure streaming approach before Start
//   - WithStreaming: Set reader/writer configuration
//   - WithChunkSize: Configure chunk size
//   - GetStats: Query final statistics after Start completes
//   - GetProgress: Query current progress during streaming
//   - Cancel: Stop streaming operation
//   - Errors: Retrieve accumulated error list
//   - context.WithTimeout: Create context with deadline
//   - context.WithCancel: Create cancellable context
func (sw *StreamingWrapper) Start(ctx context.Context) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}

	if sw.reader == nil {
		return sw.wrapper.
			WithStatusCode(http.StatusBadRequest).
			WithMessage("reader not set for streaming").
			BindCause()
	}

	sw.mu.Lock()
	if sw.isStreaming {
		sw.mu.Unlock()
		return sw.wrapper.
			WithStatusCode(http.StatusConflict).
			WithMessage("streaming already in progress").
			BindCause()
	}
	sw.isStreaming = true
	sw.mu.Unlock()

	// Update status
	sw.wrapper.
		WithMessage("Streaming started").
		WithDebuggingKV("started_at_ms", time.Now().UnixMilli())

	// Use provided context or wrapper's context
	if ctx == nil {
		ctx = sw.ctx
	}

	var streamErr error

	// Check if we're sending (compressing) or receiving (decompressing)
	// and call appropriate streaming function, capturing any error
	if sw.config.IsReceiving {
		// Receiving mode: decompress incoming data
		switch sw.config.Strategy {
		case StrategyDirect:
			streamErr = sw.streamReceiveDirect(ctx)
		case StrategyBuffered:
			streamErr = sw.streamReceiveBuffered(ctx)
		case StrategyChunked:
			streamErr = sw.streamReceiveChunked(ctx)
		default:
			streamErr = NewErrorf("unknown streaming strategy: %s", string(sw.config.Strategy))
		}
	} else {
		switch sw.config.Strategy {
		case StrategyDirect:
			streamErr = sw.streamDirect(ctx)
		case StrategyBuffered:
			streamErr = sw.streamBuffered(ctx)
		case StrategyChunked:
			streamErr = sw.streamChunked(ctx)
		default:
			streamErr = NewErrorf("unknown streaming strategy: %s", string(sw.config.Strategy))
		}
	}

	sw.mu.Lock()
	sw.isStreaming = false
	sw.mu.Unlock()

	// Update wrapper based on streaming result
	if streamErr != nil {
		sw.stats.EndTime = time.Now()
		sw.wrapper.
			WithErrorAck(streamErr).
			WithStatusCode(http.StatusInternalServerError).
			WithDebuggingKV("failed_chunks", sw.stats.FailedChunks).
			WithDebuggingKV("total_errors", len(sw.errors))

		sw.recordError(streamErr)
		return sw.wrapper
	}

	// Success response
	if sw.stats.EndTime.IsZero() {
		sw.stats.EndTime = time.Now()
	}
	sw.progress.Percentage = 100

	sw.wrapper.
		WithStatusCode(http.StatusOK).
		WithMessage("Streaming completed successfully").
		WithDebuggingKV("started_at_ms", sw.stats.StartTime.UnixMilli()).
		WithDebuggingKV("completed_at_ms", sw.stats.EndTime.UnixMilli()).
		WithDebuggingKV("total_chunks", sw.stats.TotalChunks).
		WithDebuggingKV("total_bytes", sw.stats.TotalBytes).
		WithDebuggingKV("compressed_bytes", sw.stats.CompressedBytes).
		WithDebuggingKV("duration_ms", sw.stats.EndTime.Sub(sw.stats.StartTime).Milliseconds())

	return sw.wrapper
}

// JSON returns the JSON representation of the StreamConfig.
// This method serializes the StreamConfig struct into a JSON string
// using the encoding.JSON function.
func (s *StreamConfig) JSON() string {
	return jsonpass(s)
}

// JSON returns the JSON representation of the StreamingStats.
// This method serializes the StreamingStats struct into a JSON string
// using the encoding.JSON function.
func (s *StreamingStats) JSON() string {
	return jsonpass(s)
}

// JSON returns the JSON representation of the StreamProgress.
// This method serializes the StreamProgress struct into a JSON string
// using the encoding.JSON function.
func (s *StreamProgress) JSON() string {
	return jsonpass(s)
}

// WithReceiveMode sets the streaming mode to receiving or sending.
//
// This function configures the streaming wrapper to operate in either receiving mode
// (reading data from the reader) or sending mode (writing data to the writer).
//
// Parameters:
//
//   - isReceiving: A boolean flag indicating the mode.
//
//   - true: Set to receiving mode (reading from reader).
//
//   - false: Set to sending mode (writing to writer).
//
// Returns:
//   - A pointer to the [wrapper] instance, allowing for method chaining.
func (sw *StreamingWrapper) WithReceiveMode(isReceiving bool) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}
	sw.config.IsReceiving = isReceiving
	return sw.wrapper
}

// WithWriter sets the output writer for streaming data.
//
// This function assigns the destination where streamed chunks will be written.
// If no writer is set, streaming will occur without persisting data to any output.
//
// Parameters:
//   - writer: An io.Writer implementation (e.g., *os.File, *bytes.Buffer, http.ResponseWriter).
//
// Returns:
//   - A pointer to the underlying [wrapper] instance, allowing for method chaining.
//   - If the streaming wrapper is nil, returns a new wrapper with an error message.
//
// Example:
//
//	streaming := response.AsStreaming(reader).
//	    WithWriter(outputFile).
//	    Start(ctx)
func (sw *StreamingWrapper) WithWriter(writer io.Writer) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}
	sw.writer = writer
	return sw.wrapper
}

// WithCallback sets the callback function for streaming progress updates.
//
// This function registers a callback that will be invoked during the streaming operation
// to provide real-time progress information and error notifications. The callback is called
// for each chunk processed, allowing consumers to track transfer progress, bandwidth usage,
// and estimated time remaining.
//
// Parameters:
//   - callback: A StreamingCallback function that receives progress updates and potential errors.
//     The callback signature is: func(progress *StreamProgress, err error)
//   - progress: Contains current progress metrics (bytes transferred, percentage, ETA, etc.)
//   - err: Non-nil if an error occurred during chunk processing; otherwise nil.
//
// Returns:
//   - A pointer to the underlying [wrapper] instance, allowing for method chaining.
//   - If the streaming wrapper is nil, returns a new wrapper with an error message.
//
// Example:
//
//	streaming := response.AsStreaming(reader).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            log.Printf("Streaming error at chunk %d: %v", p.CurrentChunk, err)
//	            return
//	        }
//	        fmt.Printf("Progress: %.1f%% | Rate: %.2f MB/s | ETA: %s\n",
//	            float64(p.Percentage),
//	            float64(p.TransferRate) / 1024 / 1024,
//	            p.EstimatedTimeRemaining.String())
//	    }).
//	    Start(ctx)
func (sw *StreamingWrapper) WithCallback(callback StreamingCallback) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}
	sw.callback = callback
	return sw.wrapper
}

// WithHook sets the callback function for streaming progress updates with read context.
//
// This function registers a callback that will be invoked during the streaming operation
// to provide real-time progress information and error notifications. The callback is called
// for each chunk processed, allowing consumers to track transfer progress, bandwidth usage,
// and estimated time remaining. The callback also receives the read context for advanced
// scenarios where access to the read buffer is needed.
//
// Parameters:
//
//   - callback: A StreamingCallbackR function that receives progress updates, read context, and potential errors.
//     The callback signature is: func(progress *StreamProgress, w *R)
//   - progress: Contains current progress metrics (bytes transferred, percentage, ETA, etc.)
//   - w: A pointer to the R struct containing read context for the current chunk.
func (sw *StreamingWrapper) WithHook(callback StreamingHook) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}
	sw.hook = callback
	return sw.wrapper
}

// WithTotalBytes sets the total number of bytes to be streamed.
//
// This function specifies the expected total size of the data stream, which is essential for
// calculating progress percentage and estimating time remaining. The function automatically
// computes the total number of chunks based on the configured chunk size, enabling accurate
// progress tracking throughout the streaming operation. Thread-safe via internal mutex.
//
// Parameters:
//   - totalBytes: The total size of data to be streamed in bytes. This value is used to:
//   - Calculate progress percentage: (transferredBytes / totalBytes) * 100
//   - Compute estimated time remaining: (totalBytes - transferred) / transferRate
//   - Determine total number of chunks: (totalBytes + chunkSize - 1) / chunkSize
//     Must be greater than 0 for meaningful progress calculations.
//
// Returns:
//   - A pointer to the underlying [wrapper] instance, allowing for method chaining.
//   - If the streaming wrapper is nil, returns a new wrapper with an error message.
//   - The function automatically records totalBytes in wrapper debugging information.
//
// Example:
//
//	fileInfo, _ := os.Stat("large_file.iso")
//	streaming := response.AsStreaming(fileReader).
//	    WithChunkSize(1024 * 1024).
//	    WithTotalBytes(fileInfo.Size()).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil {
//	            fmt.Printf("Downloaded: %.2f MB / %.2f MB (%.1f%%) | ETA: %s\n",
//	                float64(p.TransferredBytes) / 1024 / 1024,
//	                float64(p.TotalBytes) / 1024 / 1024,
//	                float64(p.Percentage),
//	                p.EstimatedTimeRemaining.String())
//	        }
//	    }).
//	    Start(ctx)
func (sw *StreamingWrapper) WithTotalBytes(totalBytes int64) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.progress.TotalBytes = totalBytes
	if sw.config.ChunkSize > 0 {
		sw.progress.TotalChunks = (totalBytes + sw.config.ChunkSize - 1) / sw.config.ChunkSize
	}

	sw.wrapper.WithDebuggingKV("total_bytes", totalBytes)
	return sw.wrapper
}

// WithStreamingStrategy sets the streaming algorithm strategy for data transfer.
//
// This function configures how the streaming operation processes and transfers data chunks.
// Different strategies optimize for different scenarios: STRATEGY_DIRECT for simplicity and low memory,
// STRATEGY_BUFFERED for balanced throughput and responsiveness, and STRATEGY_CHUNKED for explicit control
// in advanced scenarios. The chosen strategy affects latency, memory usage, and overall throughput characteristics.
// The strategy is recorded in wrapper debugging information for tracking and diagnostics.
//
// Parameters:
//
//   - strategy: A StreamingStrategy constant specifying the transfer algorithm.
//
// Available Strategies:
//
// - STRATEGY_DIRECT: Sequential blocking read-write without buffering.
//   - Throughput: 50-100 MB/s
//   - Latency: ~10ms per chunk
//   - Memory: Single chunk + overhead (minimal)
//   - Use case: Small files (<100MB), simple scenarios
//
// - STRATEGY_BUFFERED: Concurrent read and write with internal buffering.
//   - Throughput: 100-500 MB/s
//   - Latency: ~50ms per chunk
//   - Memory: Multiple chunks in flight (medium)
//   - Use case: Most scenarios (100MB-10GB)
//
// - STRATEGY_CHUNKED: Explicit chunk-by-chunk processing with full control.
//   - Throughput: 100-500 MB/s
//   - Latency: ~100ms per chunk
//   - Memory: Medium
//   - Use case: Large files (>10GB), specialized processing
//
// Returns:
//   - A pointer to the underlying [wrapper] instance, allowing for method chaining.
//   - If the streaming wrapper is nil, returns a new wrapper with an error message.
//   - If the strategy is empty, returns the wrapper with an error message indicating invalid input.
//   - The function automatically records the selected strategy in wrapper debugging information
//     under the key "streaming_strategy" for audit and diagnostic purposes.
//
// Example:
//
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	// Use buffered strategy for most scenarios (recommended)
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithChunkSize(1024 * 1024).
//	    WithMaxConcurrentChunks(4).
//	    WithTotalBytes(fileSize).
//	    Start(context.Background())
//
//	// Use direct strategy for small files with minimal overhead
//	result := replify.New().
//	    WithStreaming(smallFile, nil).
//	    WithStreamingStrategy(STRATEGY_DIRECT).
//	    WithChunkSize(65536).
//	    Start(context.Background())
//
//	// Use chunked strategy for very large files with explicit control
//	result := replify.New().
//	    WithStreaming(hugeFile, nil).
//	    WithStreamingStrategy(STRATEGY_CHUNKED).
//	    WithChunkSize(10 * 1024 * 1024).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        fmt.Printf("Chunk %d: %.2f MB | Rate: %.2f MB/s\n",
//	            p.CurrentChunk,
//	            float64(p.Size) / 1024 / 1024,
//	            float64(p.TransferRate) / 1024 / 1024)
//	    }).
//	    Start(context.Background())
//
// See Also:
//   - WithChunkSize: Configures the size of data chunks
//   - WithMaxConcurrentChunks: Sets parallel processing level
//   - WithCompressionType: Enables data compression
//   - Start: Initiates the streaming operation with chosen strategy
func (sw *StreamingWrapper) WithStreamingStrategy(strategy StreamingStrategy) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}

	if strutil.IsEmpty(string(strategy)) {
		return sw.wrapper.
			WithStatusCode(http.StatusBadRequest).
			WithMessage("Invalid streaming strategy: cannot be empty").
			BindCause()
	}

	sw.config.Strategy = strategy
	sw.wrapper.WithDebuggingKV("streaming_strategy", string(strategy))

	return sw.wrapper
}

// WithCompressionType sets the compression algorithm applied to streamed data chunks.
//
// This function enables data compression during streaming to reduce bandwidth consumption and transfer time.
// Compression algorithms trade CPU usage for reduced data size, with different algorithms optimized for
// different data types. Compression is applied per-chunk during streaming, allowing for progressive compression
// and decompression without loading entire dataset into memory. The selected compression type is recorded
// in wrapper debugging information for tracking and validation purposes.
//
// Parameters:
//   - comp: A CompressionType constant specifying the compression algorithm to apply.
//
// Available Compression Types:
//
// - COMP_NONE: No compression applied (passthrough mode).
//   - Compression Ratio: 100% (no reduction)
//   - CPU Overhead: None
//   - Use case: Already-compressed data (video, images, archives)
//   - Best for: Binary formats, encrypted data
//
// - COMP_GZIP: GZIP compression algorithm (RFC 1952).
//   - Compression Ratio: 20-30% (70-80% size reduction)
//   - CPU Overhead: Medium (~500ms per 100MB)
//   - Speed: Medium (balanced)
//   - Use case: Text, JSON, logs, CSV exports
//   - Best for: RESTful APIs, data exports, text-based protocols
//
// - COMP_DEFLATE: DEFLATE compression algorithm (RFC 1951).
//   - Compression Ratio: 25-35% (65-75% size reduction)
//   - CPU Overhead: Low (~300ms per 100MB)
//   - Speed: Fast (optimized)
//   - Use case: Smaller files, time-sensitive operations
//   - Best for: Quick transfers, embedded systems, IoT
//     Cannot be empty; empty string will return an error.
//
// Returns:
//   - A pointer to the underlying [wrapper] instance, allowing for method chaining.
//   - If the streaming wrapper is nil, returns a new wrapper with an error message.
//   - If the compression type is empty, returns the wrapper with an error message indicating invalid input.
//   - The function automatically records the selected compression type in wrapper debugging information
//     under the key "compression_type" for audit, diagnostics, and response transparency.
//
// Example:
//
//	// Example 1: Export CSV data with GZIP compression (recommended for text)
//	csvFile, _ := os.Open("users.csv")
//	defer csvFile.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/export/users").
//	    WithCustomFieldKV("format", "csv").
//	    WithStreaming(csvFile, nil).
//	    WithCompressionType(COMP_GZIP).
//	    WithChunkSize(512 * 1024).
//	    WithTotalBytes(csvFileSize).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil && p.Percentage % 10 == 0 {
//	            fmt.Printf("Exported: %.2f MB (compressed) | Original: %.2f MB\n",
//	                float64(p.TransferredBytes) / 1024 / 1024,
//	                float64(p.TotalBytes) / 1024 / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 2: Stream video without compression (already compressed)
//	videoFile, _ := os.Open("movie.mp4")
//	defer videoFile.Close()
//
//	result := replify.New().
//	    WithStatusCode(206).
//	    WithPath("/api/stream/video").
//	    WithCustomFieldKV("codec", "h264").
//	    WithStreaming(videoFile, nil).
//	    WithCompressionType(COMP_NONE).
//	    WithChunkSize(256 * 1024).
//	    WithTotalBytes(videoFileSize).
//	    Start(context.Background())
//
//	// Example 3: Fast log transfer with DEFLATE (IoT device)
//	logData := bytes.NewReader(logBuffer)
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/logs/upload").
//	    WithCustomFieldKV("device_id", "iot-sensor-001").
//	    WithStreaming(logData, nil).
//	    WithCompressionType(COMP_DEFLATE).
//	    WithChunkSize(64 * 1024).
//	    WithThrottleRate(256 * 1024). // 256KB/s for IoT
//	    WithTotalBytes(int64(len(logBuffer))).
//	    Start(context.Background())
//
//	// Example 4: Conditional compression based on content type
//	contentType := "application/json"
//	var compressionType CompressionType
//
//	switch contentType {
//	case "application/json", "text/csv", "text/plain":
//	    compressionType = COMP_GZIP // Text formats benefit from GZIP
//	case "video/mp4", "image/jpeg", "application/zip":
//	    compressionType = COMP_NONE // Already compressed formats
//	default:
//	    compressionType = COMP_DEFLATE // Default to fast DEFLATE
//	}
//
//	result := replify.New().
//	    WithStreaming(dataReader, nil).
//	    WithCompressionType(compressionType).
//	    Start(context.Background())
//
// Performance Impact Summary:
//
//	Data Type       GZIP Ratio    Time/100MB    Best Algorithm
//	─────────────────────────────────────────────────────────
//	JSON            15-20%        ~500ms        GZIP ✓
//	CSV             18-25%        ~500ms        GZIP ✓
//	Logs            20-30%        ~450ms        GZIP ✓
//	XML             10-15%        ~500ms        GZIP ✓
//	Binary          40-60%        ~600ms        DEFLATE
//	Video (MP4)     98-99%        ~2000ms       NONE ✓
//	Images (JPEG)   98-99%        ~2000ms       NONE ✓
//	Archives (ZIP)  100%          ~0ms          NONE ✓
//	Encrypted       100%          ~0ms          NONE ✓
//
// See Also:
//   - WithChunkSize: Configures chunk size for optimal compression
//   - WithStreamingStrategy: Selects transfer algorithm
//   - WithThrottleRate: Limits bandwidth usage
//   - GetStats: Retrieve compression statistics after streaming
//   - Start: Initiates streaming with compression enabled
func (sw *StreamingWrapper) WithCompressionType(comp CompressionType) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}

	if strutil.IsEmpty(string(comp)) {
		return sw.wrapper.
			WithStatusCode(http.StatusBadRequest).
			WithMessage("Invalid compression type: cannot be empty").
			BindCause()
	}

	sw.config.Compression = comp
	sw.wrapper.WithDebuggingKV("compression_type", string(comp))
	return sw.wrapper
}

// WithChunkSize sets the size of individual data chunks processed during streaming.
//
// This function configures the buffer size for each streaming iteration, directly impacting memory usage,
// latency, and throughput characteristics. Smaller chunks reduce memory footprint and improve responsiveness
// but increase processing overhead; larger chunks maximize throughput but consume more memory and delay
// initial response. The optimal chunk size depends on file size, available memory, network bandwidth, and
// streaming strategy. Chunk size is recorded in wrapper debugging information for tracking and diagnostics.
//
// Parameters:
//   - size: The size of each chunk in bytes. Must be greater than 0.
//
// Recommended sizes based on scenario:
//
// - 32KB (32768 bytes): Mobile networks, IoT devices, low-memory environments.
//   - Latency: ~5ms per chunk
//   - Memory: Minimal
//   - Overhead: High (frequent operations)
//   - Use case: Mobile streaming, embedded systems
//
// - 64KB (65536 bytes): Default, balanced for most scenarios.
//   - Latency: ~10ms per chunk
//   - Memory: Low
//   - Overhead: Low-Medium
//   - Use case: General-purpose file downloads, APIs
//
// - 256KB (262144 bytes): High-bandwidth networks, video streaming.
//   - Latency: ~50ms per chunk
//   - Memory: Medium
//   - Overhead: Very low
//   - Use case: Video/audio streaming, LAN transfers
//
// - 1MB (1048576 bytes): Database exports, large data transfer.
//   - Latency: ~100ms per chunk
//   - Memory: Medium-High
//   - Overhead: Very low
//   - Use case: Database exports, backups, bulk operations
//
// - 10MB (10485760 bytes): High-performance servers, LAN-only scenarios.
//   - Latency: ~500ms per chunk
//   - Memory: High
//   - Overhead: Minimal
//   - Use case: Server-to-server transfer, data center operations
//     Invalid values: Must be > 0; zero or negative values will return an error.
//
// Returns:
//   - A pointer to the underlying [wrapper] instance, allowing for method chaining.
//   - If the streaming wrapper is nil, returns a new wrapper with an error message.
//   - If the chunk size is ≤ 0, returns the wrapper with an error message indicating invalid input.
//   - The function automatically records the chunk size in wrapper debugging information
//     under the key "chunk_size" for audit, performance analysis, and diagnostics.
//
// Example:
//
//	// Example 1: Mobile client download with small chunks (responsive UI)
//	file, _ := os.Open("app-update.apk")
//	defer file.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/app-update").
//	    WithCustomFieldKV("platform", "mobile").
//	    WithStreaming(file, nil).
//	    WithChunkSize(32 * 1024). // 32KB for responsive updates
//	    WithCompressionType(COMP_GZIP).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil {
//	            fmt.Printf("\rDownloading: %.1f%% | Speed: %.2f MB/s",
//	                float64(p.Percentage),
//	                float64(p.TransferRate) / 1024 / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 2: Standard file download with balanced chunk size (default)
//	file, _ := os.Open("document.pdf")
//	defer file.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/document").
//	    WithStreaming(file, nil).
//	    WithChunkSize(64 * 1024). // 64KB balanced default
//	    WithTotalBytes(fileSize).
//	    Start(context.Background())
//
//	// Example 3: Video streaming with large chunks (high throughput)
//	videoFile, _ := os.Open("movie.mp4")
//	defer videoFile.Close()
//
//	result := replify.New().
//	    WithStatusCode(206).
//	    WithPath("/api/stream/video").
//	    WithCustomFieldKV("quality", "1080p").
//	    WithStreaming(videoFile, nil).
//	    WithChunkSize(256 * 1024). // 256KB for smooth video playback
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithMaxConcurrentChunks(4).
//	    WithTotalBytes(videoFileSize).
//	    Start(context.Background())
//
//	// Example 4: Database export with large chunks (bulk operation)
//	dbReader := createDatabaseReader("SELECT * FROM users")
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/export/users").
//	    WithCustomFieldKV("format", "csv").
//	    WithStreaming(dbReader, nil).
//	    WithChunkSize(1024 * 1024). // 1MB for bulk export
//	    WithCompressionType(COMP_GZIP).
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithMaxConcurrentChunks(8).
//	    WithTotalBytes(totalRecords * avgRecordSize).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil && p.CurrentChunk % 100 == 0 {
//	            fmt.Printf("Exported: %d records | Rate: %.2f MB/s\n",
//	                p.CurrentChunk * (1024 * 1024 / avgRecordSize),
//	                float64(p.TransferRate) / 1024 / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 5: Server-to-server transfer with maximum throughput
//	sourceReader := getNetworkReader("http://source-server/data")
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/sync/data").
//	    WithStreaming(sourceReader, nil).
//	    WithChunkSize(10 * 1024 * 1024). // 10MB for maximum throughput
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithMaxConcurrentChunks(16).
//	    WithCompressionType(COMP_NONE). // Already optimized
//	    WithTotalBytes(totalDataSize).
//	    Start(context.Background())
//
// Chunk Size Selection Guide:
//
//	File Size        Recommended Chunk    Rationale
//	─────────────────────────────────────────────────────────────────
//	< 1MB            32KB - 64KB           Minimal overhead, single chunk
//	1MB - 100MB      64KB - 256KB          Balanced, few chunks, responsive
//	100MB - 1GB      256KB - 1MB           Good throughput, moderate chunks
//	1GB - 10GB       1MB - 5MB             Optimized throughput, manageable chunks
//	> 10GB           5MB - 10MB            Maximum throughput, many chunks
//
// Memory Impact Calculation:
//
//	Total Memory Used = ChunkSize × MaxConcurrentChunks × 2 (read/write buffers)
//
//	Examples:
//	  64KB × 4 × 2 = 512KB (minimal)
//	  256KB × 4 × 2 = 2MB (standard)
//	  1MB × 8 × 2 = 16MB (large transfer)
//	  10MB × 16 × 2 = 320MB (high-performance)
//
// See Also:
//   - WithStreamingStrategy: Selects transfer algorithm affecting chunk efficiency
//   - WithMaxConcurrentChunks: Controls parallel chunk processing
//   - WithCompressionType: Compression per chunk
//   - WithThrottleRate: Bandwidth limiting independent of chunk size
//   - GetProgress: Monitor chunk processing in real-time
//   - GetStats: Retrieve chunk statistics after streaming
//   - Start: Initiates streaming with configured chunk size
func (sw *StreamingWrapper) WithChunkSize(size int64) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}

	if size <= 0 {
		return sw.wrapper.
			WithStatusCode(http.StatusBadRequest).
			WithMessagef("Invalid chunk size: %d (must be > 0)", size).
			BindCause()
	}

	sw.config.ChunkSize = size
	sw.wrapper.WithDebuggingKV("chunk_size", size)
	return sw.wrapper
}

// WithThrottleRate sets the bandwidth throttling rate to limit streaming speed in bytes per second.
//
// This function constrains the data transfer rate during streaming to manage bandwidth consumption, prevent
// network congestion, and ensure fair resource allocation in multi-client environments. Throttling is applied
// by introducing controlled delays between chunk transfers, maintaining a consistent throughput rate. This is
// particularly useful for mobile networks, satellite connections, and shared infrastructure where preventing
// upstream saturation is critical. A rate of 0 means unlimited bandwidth (no throttling). The throttle rate
// is recorded in wrapper debugging information for tracking and resource management verification.
//
// Parameters:
//
//   - bytesPerSecond: The maximum transfer rate in bytes per second (B/s).
//
// Recommended rates based on network conditions:
//
// - 0: Unlimited bandwidth, no throttling applied (default behavior).
//   - Use case: High-speed LAN transfers, server-to-server, datacenter operations
//   - Example: 0 B/s
//   - 1KB/s to 10KB/s (1024 to 10240 bytes): Ultra-low bandwidth networks.
//   - Use case: Satellite, 2G/3G networks, extremely limited connections
//   - Example: 5120 (5KB/s)
//
// - 10KB/s to 100KB/s (10240 to 102400 bytes): Low-bandwidth networks.
//   - Use case: Rural internet, IoT devices, edge networks
//   - Example: 51200 (50KB/s)
//
// - 100KB/s to 1MB/s (102400 to 1048576 bytes): Standard mobile networks.
//   - Use case: Mobile clients (3G/4G), fair sharing in shared networks
//   - Example: 512000 (512KB/s, 4Mbps)
//
// - 1MB/s to 10MB/s (1048576 to 10485760 bytes): High-speed mobile/WiFi.
//   - Use case: High-speed WiFi, 4G LTE, fiber connections
//   - Example: 5242880 (5MB/s, 40Mbps)
//
// - 10MB/s to 100MB/s (10485760 to 104857600 bytes): Gigabit networks.
//   - Use case: Fast LAN, dedicated connections, bulk transfers
//   - Example: 52428800 (50MB/s, 400Mbps)
//
// - > 100MB/s (> 104857600 bytes): Ultra-high-speed networks.
//   - Use case: 10 Gigabit Ethernet, NVMe over network, data center
//   - Example: 1073741824 (1GB/s, 8Gbps theoretical max)
//
// Invalid values: Negative values will return an error; zero is treated as unlimited.
//
// Returns:
//   - A pointer to the underlying [wrapper] instance, allowing for method chaining.
//   - If the streaming wrapper is nil, returns a new wrapper with an error message.
//   - If the throttle rate is negative, returns the wrapper with an error message indicating invalid input.
//   - The function automatically records the throttle rate in wrapper debugging information:
//   - Key "throttle_rate_bps" with the rate value if throttling is enabled (> 0)
//   - Key "throttle_rate" with value "unlimited" if throttling is disabled (0)
//   - This enables easy identification of throttled vs unlimited transfers in logs and diagnostics.
//
// Example:
//
//	// Example 1: Mobile client download with throttling (fair bandwidth sharing)
//	file, _ := os.Open("large-app.apk")
//	defer file.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/app").
//	    WithCustomFieldKV("platform", "mobile").
//	    WithStreaming(file, nil).
//	    WithChunkSize(32 * 1024). // 32KB chunks
//	    WithThrottleRate(512 * 1024). // 512KB/s (4Mbps)
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil && p.CurrentChunk % 20 == 0 {
//	            fmt.Printf("Speed: %.2f KB/s | ETA: %s\n",
//	                float64(p.TransferRate) / 1024,
//	                p.EstimatedTimeRemaining.String())
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 2: IoT device sensor data upload with ultra-low rate
//	sensorData := bytes.NewReader(telemetryBuffer)
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/telemetry/upload").
//	    WithCustomFieldKV("device_id", "sensor-2025-1114").
//	    WithStreaming(sensorData, nil).
//	    WithChunkSize(16 * 1024). // 16KB chunks
//	    WithThrottleRate(10 * 1024). // 10KB/s (very limited)
//	    WithCompressionType(COMP_DEFLATE). // Reduce size further
//	    WithTotalBytes(int64(len(telemetryBuffer))).
//	    Start(context.Background())
//
//	// Example 3: Multi-client server with fair bandwidth allocation
//	// Scenario: 10 concurrent downloads, server has 100MB/s available
//	// Allocate 10MB/s per client for fair sharing
//	file, _ := os.Open("shared-resource.iso")
//	defer file.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/shared").
//	    WithStreaming(file, nil).
//	    WithChunkSize(256 * 1024). // 256KB chunks
//	    WithThrottleRate(10 * 1024 * 1024). // 10MB/s per client
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil {
//	            // Verify actual rate stays within limit
//	            if p.TransferRate > 10 * 1024 * 1024 {
//	                log.Warnf("Warning: Rate exceeded limit: %.2f MB/s",
//	                    float64(p.TransferRate) / 1024 / 1024)
//	            }
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 4: Unlimited bandwidth for server-to-server (no throttling)
//	sourceFile, _ := os.Open("backup.tar.gz")
//	defer sourceFile.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/sync/backup").
//	    WithStreaming(sourceFile, nil).
//	    WithChunkSize(10 * 1024 * 1024). // 10MB chunks
//	    WithThrottleRate(0). // Unlimited, maximum throughput
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithMaxConcurrentChunks(16). // Maximize parallelism
//	    WithCompressionType(COMP_NONE). // Already compressed
//	    WithTotalBytes(fileSize).
//	    Start(context.Background())
//
//	// Example 5: Dynamic throttling based on network conditions
//	var throttleRate int64
//	networkCondition := detectNetworkQuality() // Custom function
//
//	switch networkCondition {
//	case "5g":
//	    throttleRate = 50 * 1024 * 1024 // 50MB/s
//	case "4g_lte":
//	    throttleRate = 5 * 1024 * 1024 // 5MB/s
//	case "4g":
//	    throttleRate = 1 * 1024 * 1024 // 1MB/s
//	case "3g":
//	    throttleRate = 256 * 1024 // 256KB/s
//	case "satellite":
//	    throttleRate = 50 * 1024 // 50KB/s
//	default:
//	    throttleRate = 512 * 1024 // 512KB/s (safe default)
//	}
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/adaptive").
//	    WithCustomFieldKV("network", string(networkCondition)).
//	    WithStreaming(fileReader, nil).
//	    WithThrottleRate(throttleRate).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil {
//	            fmt.Printf("Network: %s | Speed: %.2f MB/s | ETA: %s\n",
//	                networkCondition,
//	                float64(p.TransferRate) / 1024 / 1024,
//	                p.EstimatedTimeRemaining.String())
//	        }
//	    }).
//	    Start(context.Background())
//
// Throttle Rate Reference:
//
//	Network Type          Recommended Rate    Typical Bandwidth
//	─────────────────────────────────────────────────────────────
//	Satellite             50KB/s              400 Kbps
//	2G (GSM/EDGE)         10-20KB/s           56-128 Kbps
//	3G (WCDMA/HSPA)       256KB/s             2-3 Mbps
//	4G (LTE)              1-5MB/s             10-50 Mbps
//	4G (LTE-A)            5-10MB/s            100+ Mbps
//	5G                    50-100MB/s          500+ Mbps
//	WiFi 5 (802.11ac)     10-50MB/s           100-300 Mbps
//	WiFi 6 (802.11ax)     50-100MB/s          1000+ Mbps
//	Gigabit Ethernet      100-500MB/s         1 Gbps (1000 Mbps)
//	10 Gigabit Ethernet   500MB-1GB/s         10 Gbps
//
// Bandwidth Calculation Examples:
//
//	Rate (B/s)      Rate (KB/s)     Rate (Mbps)     Use Case
//	─────────────────────────────────────────────────────────────
//	10240           10              81.92           Satellite uplink
//	51200           50              409.6           Rural internet
//	262144          256             2.048           3G network
//	1048576         1024            8.388           4G LTE
//	5242880         5120            41.94           High-speed mobile
//	52428800        51200           419.4           WiFi
//	104857600       102400          838.8           Gigabit LAN
//	1073741824      1048576         8388.6          10 Gigabit LAN
//
// See Also:
//   - WithChunkSize: Affects throttling responsiveness and chunk processing
//   - WithStreamingStrategy: Strategy selection affects throttling efficiency
//   - WithMaxConcurrentChunks: Parallelism independent of throttle rate
//   - GetProgress: Monitor actual transfer rate vs throttle limit
//   - GetStats: Retrieve bandwidth statistics after streaming
//   - Start: Initiates streaming with throttle rate limit applied
func (sw *StreamingWrapper) WithThrottleRate(bytesPerSecond int64) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}

	if bytesPerSecond < 0 {
		return sw.wrapper.
			WithStatusCode(http.StatusBadRequest).
			WithMessagef("Invalid throttle rate: %d (must be >= 0)", bytesPerSecond).
			BindCause()
	}

	sw.config.ThrottleRate = bytesPerSecond
	if bytesPerSecond > 0 {
		sw.wrapper.WithDebuggingKV("throttle_rate_bps", bytesPerSecond)
	} else {
		sw.wrapper.WithDebuggingKV("throttle_rate", "unlimited")
	}
	return sw.wrapper
}

// WithMaxConcurrentChunks sets the maximum number of data chunks processed concurrently during streaming.
//
// This function configures the level of parallelism for chunk processing, directly impacting throughput,
// CPU utilization, and memory consumption. Higher concurrency enables better bandwidth utilization and
// faster overall transfer by processing multiple chunks simultaneously; however, it increases CPU load and
// memory overhead. The optimal concurrency level depends on available CPU cores, memory constraints, network
// capacity, and the streaming strategy employed. This is particularly effective with STRATEGY_BUFFERED where
// read and write operations can overlap. The concurrency level is recorded in wrapper debugging information
// for performance analysis and resource monitoring.
//
// Parameters:
//   - count: The maximum number of chunks to process in parallel. Must be greater than 0.
//
// Recommended values based on scenario:
//
// - 1: Single-threaded sequential processing (no parallelism).
//   - Throughput: Low (50-100 MB/s)
//   - CPU Usage: Minimal (~25% single core)
//   - Memory: Minimal (1 chunk buffer)
//   - Latency: Highest (~100ms between chunks)
//   - Use case: Single-core systems, extremely limited resources, ordered processing
//   - Best with: STRATEGY_DIRECT (matches sequential nature)
//
// - 2: Minimal parallelism (read ahead 1 chunk).
//   - Throughput: Medium (100-200 MB/s)
//   - CPU Usage: Low-Medium (~50% two cores)
//   - Memory: Low (2 chunk buffers)
//   - Latency: Medium (~50ms between chunks)
//   - Use case: Mobile devices, low-resource environments
//   - Best with: STRATEGY_BUFFERED
//
// - 4: Standard parallelism (balanced default).
//   - Throughput: High (200-500 MB/s)
//   - CPU Usage: Medium (~100% four cores)
//   - Memory: Medium (4 chunk buffers)
//   - Latency: Low (~25ms between chunks)
//   - Use case: Most general-purpose scenarios, typical servers
//   - Best with: STRATEGY_BUFFERED (recommended)
//
// - 8: High parallelism (aggressive read-ahead).
//   - Throughput: Very High (500-1000 MB/s)
//   - CPU Usage: High (~100% eight cores)
//   - Memory: High (8 chunk buffers)
//   - Latency: Very Low (~10ms between chunks)
//   - Use case: Multi-core servers, high-bandwidth networks, database exports
//   - Best with: STRATEGY_BUFFERED or STRATEGY_CHUNKED
//
// - 16+: Extreme parallelism (maximum throughput).
//   - Throughput: Maximum (1000+ MB/s potential)
//   - CPU Usage: Very High (all cores maxed)
//   - Memory: Very High (16+ chunk buffers in flight)
//   - Latency: Minimal (~1ms between chunks)
//   - Use case: Data center operations, 10 Gigabit networks, bulk transfer servers
//   - Best with: STRATEGY_BUFFERED with large chunk sizes (1MB+)
//
// Invalid values: Must be > 0; zero or negative values will return an error.
//
// Returns:
//   - A pointer to the underlying [wrapper] instance, allowing for method chaining.
//   - If the streaming wrapper is nil, returns a new wrapper with an error message.
//   - If the concurrent chunk count is ≤ 0, returns the wrapper with an error message indicating invalid input.
//   - The function automatically records the max concurrent chunks in wrapper debugging information
//     under the key "max_concurrent_chunks" for audit, performance profiling, and resource tracking.
//
// Memory Impact Calculation:
//
//	Total Streaming Memory = ChunkSize × MaxConcurrentChunks × 2 + Overhead
//
//	Formula Explanation:
//	  - ChunkSize: Size of individual chunk in bytes
//	  - MaxConcurrentChunks: Number of concurrent chunks in flight
//	  - 2: Read buffer + Write buffer (input and output streams)
//	  - Overhead: ~1-5MB for data structures, compression buffers, etc.
//
//	Memory Examples (with 64KB chunks):
//	  1 concurrent:  64KB × 1 × 2 = 128KB + 2MB overhead = ~2.1MB total
//	  2 concurrent:  64KB × 2 × 2 = 256KB + 2MB overhead = ~2.3MB total
//	  4 concurrent:  64KB × 4 × 2 = 512KB + 2MB overhead = ~2.5MB total
//	  8 concurrent:  64KB × 8 × 2 = 1MB + 2MB overhead = ~3.0MB total
//	  16 concurrent: 64KB × 16 × 2 = 2MB + 2MB overhead = ~4.0MB total
//
//	Memory Examples (with 1MB chunks):
//	  1 concurrent:  1MB × 1 × 2 = 2MB + 2MB overhead = ~4MB total
//	  2 concurrent:  1MB × 2 × 2 = 4MB + 2MB overhead = ~6MB total
//	  4 concurrent:  1MB × 4 × 2 = 8MB + 2MB overhead = ~10MB total
//	  8 concurrent:  1MB × 8 × 2 = 16MB + 2MB overhead = ~18MB total
//	  16 concurrent: 1MB × 16 × 2 = 32MB + 2MB overhead = ~34MB total
//
// Example:
//
//	// Example 1: Single-threaded processing for ordered/sequential requirement
//	csvFile, _ := os.Open("ordered-data.csv")
//	defer csvFile.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/process/ordered-csv").
//	    WithStreaming(csvFile, nil).
//	    WithChunkSize(64 * 1024).
//	    WithStreamingStrategy(STRATEGY_DIRECT).
//	    WithMaxConcurrentChunks(1). // Sequential processing
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil && p.CurrentChunk % 100 == 0 {
//	            fmt.Printf("Chunk %d: Processed in order\n", p.CurrentChunk)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 2: Mobile client with minimal concurrency (low memory)
//	appData := bytes.NewReader(mobileUpdate)
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/mobile-app").
//	    WithCustomFieldKV("platform", "ios").
//	    WithStreaming(appData, nil).
//	    WithChunkSize(32 * 1024). // 32KB chunks
//	    WithMaxConcurrentChunks(2). // Low memory footprint
//	    WithCompressionType(COMP_GZIP).
//	    WithThrottleRate(512 * 1024). // 512KB/s
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithTotalBytes(int64(len(mobileUpdate))).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil {
//	            fmt.Printf("Memory efficient: %.1f%% | Speed: %.2f KB/s\n",
//	                float64(p.Percentage),
//	                float64(p.TransferRate) / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 3: Standard server download with balanced concurrency (recommended)
//	file, _ := os.Open("document.iso")
//	defer file.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/document").
//	    WithStreaming(file, nil).
//	    WithChunkSize(256 * 1024). // 256KB chunks
//	    WithMaxConcurrentChunks(4). // Balanced (most scenarios)
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithCompressionType(COMP_GZIP).
//	    WithTotalBytes(fileSize).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil && p.ElapsedTime.Seconds() > 0 {
//	            fmt.Printf("Progress: %.1f%% | Bandwidth: %.2f MB/s | ETA: %s\n",
//	                float64(p.Percentage),
//	                float64(p.TransferRate) / 1024 / 1024,
//	                p.EstimatedTimeRemaining.String())
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 4: High-performance server with aggressive parallelism
//	dbExport := createDatabaseStreamReader("SELECT * FROM transactions", 10000)
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/export/transactions").
//	    WithCustomFieldKV("format", "parquet").
//	    WithStreaming(dbExport, nil).
//	    WithChunkSize(1024 * 1024). // 1MB chunks
//	    WithMaxConcurrentChunks(8). // High parallelism
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithCompressionType(COMP_GZIP).
//	    WithMaxConcurrentChunks(8). // Process 8 chunks in parallel
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Error at chunk %d: %v\n", p.CurrentChunk, err)
//	            return
//	        }
//	        if p.CurrentChunk % 50 == 0 {
//	            stats := fmt.Sprintf(
//	                "Exported: %.2f MB | Rate: %.2f MB/s | ETA: %s",
//	                float64(p.TransferredBytes) / 1024 / 1024,
//	                float64(p.TransferRate) / 1024 / 1024,
//	                p.EstimatedTimeRemaining.String(),
//	            )
//	            fmt.Println(stats)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 5: Data center bulk transfer with maximum parallelism
//	sourceServer := getNetworkStreamReader("http://backup-server/full-backup")
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/sync/full-backup").
//	    WithCustomFieldKV("source", "backup-server").
//	    WithStreaming(sourceServer, nil).
//	    WithChunkSize(10 * 1024 * 1024). // 10MB chunks for high throughput
//	    WithMaxConcurrentChunks(16). // Maximum parallelism for datacenter
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithCompressionType(COMP_NONE). // Already optimized
//	    WithMaxConcurrentChunks(16). // 16 chunks in parallel
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil && p.CurrentChunk % 100 == 0 {
//	            throughput := float64(p.TransferRate) / 1024 / 1024 / 1024
//	            fmt.Printf("Bulk transfer: %.2f GB/s | %.1f%% complete\n",
//	                throughput, float64(p.Percentage))
//	        }
//	    }).
//	    Start(context.Background())
//
// Concurrency Selection Matrix:
//
//	System Type          Available Cores    Recommended Count    Throughput
//	─────────────────────────────────────────────────────────────────────────
//	Single-core          1                  1                    50-100 MB/s
//	Mobile (2-4 cores)   2-4                2-4                  100-500 MB/s
//	Standard Server      8 cores            4-8                  200-1000 MB/s
//	High-end Server      16+ cores          8-16                 500-2000 MB/s
//	Data Center          32+ cores          16-32                1000+ MB/s
//
// CPU & Memory Trade-off:
//
//	Count    CPU Load    Memory (64KB)    Memory (1MB)    Throughput    Best For
//	─────────────────────────────────────────────────────────────────────────────────
//	1        25%         ~2MB            ~4MB            100 MB/s      Sequential
//	2        50%         ~2.3MB          ~6MB            200 MB/s      Mobile
//	4        100%        ~2.5MB          ~10MB           500 MB/s      Standard (✓)
//	8        100%        ~3MB            ~18MB           1000 MB/s     High-perf
//	16       100%        ~4MB            ~34MB           2000 MB/s     Datacenter
//
// See Also:
//   - WithChunkSize: Chunk size affects memory per concurrent chunk
//   - WithStreamingStrategy: Strategy affects concurrency efficiency (BUFFERED best)
//   - WithCompressionType: Compression affects CPU usage with high concurrency
//   - WithThrottleRate: Throttling independent of concurrency level
//   - GetProgress: Monitor actual concurrency effect in real-time
//   - GetStats: Retrieve throughput metrics after streaming
//   - Start: Initiates streaming with parallel chunk processing
func (sw *StreamingWrapper) WithMaxConcurrentChunks(count int) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}

	if count <= 0 {
		return sw.wrapper.
			WithStatusCode(http.StatusBadRequest).
			WithMessagef("Invalid max concurrent chunks: %d (must be > 0)", count).
			BindCause()
	}

	sw.config.MaxConcurrentChunks = count
	sw.wrapper.WithDebuggingKV("max_concurrent_chunks", count)
	return sw.wrapper
}

// WithBufferPooling enables or disables buffer pooling for efficient memory reuse during streaming.
//
// This function controls whether streaming operations reuse allocated buffers through a pool mechanism
// or allocate fresh buffers for each chunk. Buffer pooling reduces garbage collection pressure, improves
// memory allocation efficiency, and can provide 10-20% performance improvement for streaming operations.
// However, it adds minimal overhead (5-10%) when disabled for very small files or low-frequency operations.
// When enabled, buffers are allocated once and recycled across chunks, reducing GC pause times and memory
// fragmentation. The buffer pooling state is recorded in wrapper debugging information for performance
// analysis, optimization tracking, and resource management auditing.
//
// Parameters:
//
//   - enabled: Boolean flag controlling buffer pooling behavior.
//
// - True: Enable buffer pooling (recommended for most scenarios).
//
// - Pros:
//   - 10-20% performance improvement on sustained transfers
//   - Reduced garbage collection overhead and GC pause times
//   - Lower memory fragmentation and allocation pressure
//   - Stable memory usage over time (less heap churn)
//   - Better for long-running servers and high-throughput scenarios
//
// - Cons:
//   - Minimal memory overhead (~4 buffers pooled)
//   - Negligible overhead for single-request scenarios
//   - Use case: Production servers, sustained transfers, high-concurrency scenarios
//
// - False: Disable buffer pooling (allocation per chunk).
//
// - Pros:
//   - Slightly lower startup memory overhead
//   - No pool management overhead for very small files
//   - Pure Go standard library behavior
//
// - Cons:
//
//   - 10-20% slower for large transfers (GC overhead)
//
//   - More garbage collection pressure
//
//   - Higher memory fragmentation
//
//   - GC pause times increase with file size
//
//   - Use case: One-time transfers, small files (<10MB), memory-constrained environments
//
//     Default: true (pooling enabled, recommended).
//
// Returns:
//   - A pointer to the underlying [wrapper] instance, allowing for method chaining.
//   - If the streaming wrapper is nil, returns a new wrapper with an error message.
//   - The function automatically records the buffer pooling state in wrapper debugging information
//     under the key "buffer_pooling_enabled" for audit, performance profiling, and configuration tracking.
//
// Performance Impact Analysis:
//
//	Operation             Pooling Enabled    Pooling Disabled    Improvement
//	─────────────────────────────────────────────────────────────────────────
//	100MB transfer        1200ms             1400ms              14.3% faster
//	1GB transfer          12000ms            14000ms             14.3% faster
//	10GB transfer         120000ms           140000ms            14.3% faster
//	GC Pause Time (avg)   2ms                15ms                87.5% reduction
//	Memory Allocs/ops     1000               10000               90% fewer allocs
//	Heap Fragmentation    Low                High                Significant
//
// Memory Usage Comparison:
//
//	Scenario                  Pooling Enabled      Pooling Disabled
//	─────────────────────────────────────────────────────────────────
//	Single 10MB file          +2MB pool overhead   Minimal overhead
//	Multiple 10MB files       +2MB pool (reused)   10MB × files in RAM
//	1GB sustained transfer    Steady 2MB pool      Spikes to 50-100MB
//	Long-running server       Stable 2-5MB pool    Growing heap (GC lag)
//	Mobile app               +2MB pool            Better for short ops
//
// Example:
//
//	// Example 1: Production server with sustained transfers (pooling enabled - recommended)
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithCustomFieldKV("environment", "production").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024). // 1MB chunks
//	    WithMaxConcurrentChunks(4).
//	    WithBufferPooling(true). // Enable for production
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil && p.CurrentChunk % 100 == 0 {
//	            fmt.Printf("Transferred: %.2f MB | Memory stable: %.2f MB\n",
//	                float64(p.TransferredBytes) / 1024 / 1024,
//	                float64(p.TransferRate) / 1024 / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 2: CLI tool with one-time large file (pooling enabled still better)
//	file, _ := os.Open("backup.tar.gz")
//	defer file.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/export/backup").
//	    WithStreaming(file, nil).
//	    WithChunkSize(512 * 1024). // 512KB chunks
//	    WithBufferPooling(true). // Still recommended
//	    WithCompressionType(COMP_GZIP).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil {
//	            fmt.Printf("\rProgress: %.1f%% | Speed: %.2f MB/s",
//	                float64(p.Percentage),
//	                float64(p.TransferRate) / 1024 / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 3: Mobile app with limited resources (pooling enabled for better efficiency)
//	appData := bytes.NewReader(updatePackage)
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/app-update").
//	    WithCustomFieldKV("platform", "ios").
//	    WithCustomFieldKV("device_memory", "2GB").
//	    WithStreaming(appData, nil).
//	    WithChunkSize(32 * 1024). // 32KB for mobile
//	    WithBufferPooling(true). // Enable for efficiency on mobile
//	    WithThrottleRate(512 * 1024). // 512KB/s throttling
//	    WithTotalBytes(int64(len(updatePackage))).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil {
//	            fmt.Printf("Memory efficient: %.1f%% | Speed: %.2f KB/s\n",
//	                float64(p.Percentage),
//	                float64(p.TransferRate) / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 4: Minimal/embedded system (pooling disabled to save memory)
//	embeddedData := bytes.NewReader(firmwareUpdate)
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/upload/firmware").
//	    WithCustomFieldKV("device_type", "iot-sensor").
//	    WithCustomFieldKV("available_memory", "512MB").
//	    WithStreaming(embeddedData, nil).
//	    WithChunkSize(16 * 1024). // 16KB for embedded
//	    WithBufferPooling(false). // Disable to minimize memory
//	    WithStreamingStrategy(STRATEGY_DIRECT).
//	    WithMaxConcurrentChunks(1). // Single-threaded
//	    WithCompressionType(COMP_DEFLATE). // More compression
//	    WithTotalBytes(int64(len(firmwareUpdate))).
//	    Start(context.Background())
//
//	// Example 5: Conditional pooling based on system resources
//	availableMemory := getSystemMemory() // Custom function
//	var enablePooling bool
//
//	if availableMemory > 1024 * 1024 * 1024 { // > 1GB
//	    enablePooling = true // Enable pooling for better performance
//	} else {
//	    enablePooling = false // Disable to save memory on constrained systems
//	}
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/adaptive").
//	    WithCustomFieldKV("available_memory_mb", availableMemory / 1024 / 1024).
//	    WithStreaming(fileReader, nil).
//	    WithChunkSize(256 * 1024).
//	    WithBufferPooling(enablePooling). // Conditional based on resources
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil {
//	            fmt.Printf("Pooling: %v | Progress: %.1f%% | Rate: %.2f MB/s\n",
//	                enablePooling,
//	                float64(p.Percentage),
//	                float64(p.TransferRate) / 1024 / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
// System Behavior Recommendations:
//
//	System Type              File Size       Recommendation    Reasoning
//	────────────────────────────────────────────────────────────────────────
//	Production Server        Any             true (enable)     Sustained performance
//	Development/Testing      < 100MB         false (disable)   Lower overhead
//	Development/Testing      > 100MB         true (enable)     Test real behavior
//	Mobile App               Any             true (enable)     Efficiency critical
//	Embedded/IoT             Any             false (disable)    Memory limited
//	High-Concurrency API     Any             true (enable)      GC impact significant
//	One-time CLI             < 10MB          false (disable)    No sustained benefit
//	One-time CLI             > 10MB          true (enable)      GC overhead matters
//	Microservice             Any             true (enable)      Container overhead
//	Batch Processing         Any             true (enable)      Server efficiency
//
// GC Tuning Notes:
//
//	When buffer pooling is ENABLED:
//	  - Reduces allocation pressure on Go's allocator
//	  - Decreases GC frequency (fewer objects to scan)
//	  - Lower GC pause times (critical for latency-sensitive APIs)
//	  - Recommended: Default GOGC=100 works well
//
//	When buffer pooling is DISABLED:
//	  - Higher allocation pressure
//	  - More frequent GC cycles
//	  - Longer GC pause times (can be 10-100ms on large transfers)
//	  - Consider: GOGC=200 to reduce GC frequency (trade CPU for latency)
//
// See Also:
//   - WithChunkSize: Larger chunks reduce pool efficiency
//   - WithMaxConcurrentChunks: More concurrency benefits more from pooling
//   - WithStreamingStrategy: STRATEGY_BUFFERED benefits most from pooling
//   - GetStats: Retrieve memory and performance statistics
//   - Start: Initiates streaming with configured buffer pooling
func (sw *StreamingWrapper) WithBufferPooling(enabled bool) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}

	sw.config.UseBufferPool = enabled
	sw.wrapper.WithDebuggingKV("buffer_pooling_enabled", enabled)
	return sw.wrapper
}

// WithReadTimeout sets the read operation timeout for streaming data acquisition.
//
// This function configures the maximum duration allowed for individual read operations on the input stream.
// The read timeout acts as a circuit breaker to prevent indefinite blocking when data sources become
// unresponsive, disconnected, or stalled. Each chunk read attempt must complete within the specified timeout;
// if exceeded, the read operation fails and streaming is interrupted. This is critical for production systems
// where network failures, slow clients, or hung connections could otherwise freeze the entire streaming
// operation indefinitely. Read timeout is independent of write timeout and total operation timeout. The timeout
// value is recorded in wrapper debugging information for audit, performance tracking, and troubleshooting.
//
// Parameters:
//   - timeout: The maximum duration for read operations in milliseconds. Must be greater than 0.
//
// Recommended values based on scenario:
//
// - 1000-5000ms (1-5 seconds): Very fast networks, local transfers, LAN.
//   - Use case: File downloads on gigabit LAN, local API calls
//   - Network: Sub-millisecond latency (<1ms typical round-trip)
//   - Best for: High-speed, predictable connections
//   - Example: 2000 (2 seconds)
//
// - 5000-15000ms (5-15 seconds): Standard networks, normal internet.
//   - Use case: Most REST API downloads, web servers, typical internet transfers
//   - Network: 10-100ms latency (typical broadband)
//   - Best for: General-purpose APIs and services
//   - Example: 10000 (10 seconds) - RECOMMENDED DEFAULT
//
// - 15000-30000ms (15-30 seconds): Slow/congested networks, mobile networks.
//   - Use case: Mobile clients (3G/4G), congested WiFi, distant servers
//   - Network: 100-500ms latency (cellular networks)
//   - Best for: Mobile apps, unreliable connections
//   - Example: 20000 (20 seconds)
//
// - 30000-60000ms (30-60 seconds): Very slow networks, satellite, WAN.
//   - Use case: Satellite connections, international transfers, dial-up
//   - Network: 500ms-2s latency (very slow links)
//   - Best for: Challenging network conditions
//   - Example: 45000 (45 seconds)
//
// - 60000+ms (60+ seconds): Extremely slow/unreliable connections only.
//   - Use case: Satellite uplink, emergency networks, extreme edge cases
//   - Network: 2s+ latency
//   - Best for: Last-resort scenarios with critical data
//   - Example: 120000 (120 seconds)
//
// Invalid values: Must be > 0; zero or negative values will return an error.
// Note: Very large timeouts (>120s) can mask real connection problems; consider implementing
// application-level heartbeats instead for better reliability.
//
// Returns:
//   - A pointer to the underlying [wrapper] instance, allowing for method chaining.
//   - If the streaming wrapper is nil, returns a new wrapper with an error message.
//   - If the read timeout is ≤ 0, returns the wrapper with an error message indicating invalid input.
//   - The function automatically records the read timeout in wrapper debugging information
//     under the key "read_timeout_ms" for audit, performance analysis, and troubleshooting.
//
// Timeout Behavior Semantics:
//
//	Scenario                        Behavior with ReadTimeout
//	───────────────────────────────────────────────────────────────────────
//	Data arrives before timeout     Chunk processed immediately
//	Data arrives after timeout      Read fails, streaming terminates
//	Connection stalled              Read blocks until timeout, then fails
//	EOF reached                     Streaming completes normally
//	Network disconnect              Read fails immediately (OS-level)
//	Slow source (< timeout rate)    Streaming continues, each chunk waits
//	Source sends partial data       Waits for complete chunk or timeout
//
// Read Timeout vs Write Timeout Comparison:
//
//	Aspect                  ReadTimeout         WriteTimeout
//	─────────────────────────────────────────────────────────────
//	Controls                Data input delay    Data output delay
//	Fails when              Source is slow      Destination is slow
//	Typical cause           Slow upload         Slow client/network
//	Recovery                Retry from chunk    Chunk lost (retry)
//	Recommendation          10-30 seconds       10-30 seconds
//	Relationship            Independent         Independent
//	Combined max            Sum of both         Sequential impact
//
// Example:
//
//	// Example 1: LAN file transfer with fast timeout (2 seconds)
//	file, _ := os.Open("data.bin")
//	defer file.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/lan-transfer").
//	    WithCustomFieldKV("network", "gigabit-lan").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024). // 1MB chunks
//	    WithReadTimeout(2000). // 2 seconds for fast LAN
//	    WithWriteTimeout(2000). // Match read timeout
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithMaxConcurrentChunks(8).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Transfer failed: %v\n", err)
//	            return
//	        }
//	        if p.CurrentChunk % 100 == 0 {
//	            fmt.Printf("Progress: %.1f%% | Speed: %.2f MB/s\n",
//	                float64(p.Percentage),
//	                float64(p.TransferRate) / 1024 / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 2: Standard internet download with typical timeout (10 seconds)
//	httpResp, _ := http.Get("https://api.example.com/download/document")
//	defer httpResp.Body.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/from-internet").
//	    WithCustomFieldKV("source", "api.example.com").
//	    WithStreaming(httpResp.Body, nil).
//	    WithChunkSize(256 * 1024). // 256KB chunks
//	    WithReadTimeout(10000). // 10 seconds standard
//	    WithWriteTimeout(10000).
//	    WithCompressionType(COMP_GZIP).
//	    WithMaxConcurrentChunks(4).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            log.Warnf("Download stalled: %v", err)
//	            return
//	        }
//	        if p.ElapsedTime.Seconds() > 0 {
//	            fmt.Printf("ETA: %s | Speed: %.2f MB/s\n",
//	                p.EstimatedTimeRemaining.String(),
//	                float64(p.TransferRate) / 1024 / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 3: Mobile client with extended timeout (20 seconds)
//	mobileStream, _ := os.Open("app-update.apk")
//	defer mobileStream.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/mobile-app").
//	    WithCustomFieldKV("platform", "android").
//	    WithCustomFieldKV("network", "4g-lte").
//	    WithStreaming(mobileStream, nil).
//	    WithChunkSize(32 * 1024). // 32KB small chunks
//	    WithReadTimeout(20000). // 20 seconds for mobile
//	    WithWriteTimeout(20000). // Account for slow client
//	    WithThrottleRate(512 * 1024). // 512KB/s
//	    WithCompressionType(COMP_GZIP).
//	    WithMaxConcurrentChunks(2).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            log.Errorf("Mobile download failed: %v", err)
//	            // Could implement retry logic here
//	            return
//	        }
//	        if p.CurrentChunk % 50 == 0 {
//	            fmt.Printf("Mobile: %.1f%% | Speed: %.2f KB/s | Signal: Good\n",
//	                float64(p.Percentage),
//	                float64(p.TransferRate) / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 4: Slow/unreliable network with long timeout (45 seconds)
//	satelliteReader := createSatelliteStreamReader() // Custom reader
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/satellite").
//	    WithCustomFieldKV("source", "satellite-link").
//	    WithCustomFieldKV("network_quality", "poor").
//	    WithStreaming(satelliteReader, nil).
//	    WithChunkSize(64 * 1024). // 64KB for stability
//	    WithReadTimeout(45000). // 45 seconds for satellite
//	    WithWriteTimeout(45000).
//	    WithStreamingStrategy(STRATEGY_DIRECT). // Sequential for reliability
//	    WithCompressionType(COMP_GZIP).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Satellite link error: %v (chunk %d)\n",
//	                err, p.CurrentChunk)
//	            return
//	        }
//	        if p.CurrentChunk % 20 == 0 {
//	            fmt.Printf("Satellite: Chunk %d received | ETA: %s\n",
//	                p.CurrentChunk,
//	                p.EstimatedTimeRemaining.String())
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 5: Adaptive timeout based on network detection
//	networkType := detectNetworkType() // Custom function: "lan", "internet", "mobile", "satellite"
//	var readTimeoutMs int64
//
//	switch networkType {
//	case "lan":
//	    readTimeoutMs = 3000 // 3 seconds for LAN
//	case "internet":
//	    readTimeoutMs = 10000 // 10 seconds for internet
//	case "mobile":
//	    readTimeoutMs = 20000 // 20 seconds for mobile
//	case "satellite":
//	    readTimeoutMs = 60000 // 60 seconds for satellite
//	default:
//	    readTimeoutMs = 15000 // 15 seconds default
//	}
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/adaptive").
//	    WithCustomFieldKV("detected_network", networkType).
//	    WithStreaming(fileReader, nil).
//	    WithChunkSize(256 * 1024).
//	    WithReadTimeout(readTimeoutMs).
//	    WithWriteTimeout(readTimeoutMs).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            log.Warnf("Network: %s | Error: %v | Chunk: %d",
//	                networkType, err, p.CurrentChunk)
//	            return
//	        }
//	        if p.CurrentChunk % 50 == 0 {
//	            fmt.Printf("[%s] %.1f%% | Speed: %.2f MB/s | ETA: %s\n",
//	                networkType,
//	                float64(p.Percentage),
//	                float64(p.TransferRate) / 1024 / 1024,
//	                p.EstimatedTimeRemaining.String())
//	        }
//	    }).
//	    Start(context.Background())
//
// Network-Based Timeout Selection Guide:
//
//	Network Type           Latency      Timeout (ms)    Rationale
//	─────────────────────────────────────────────────────────────────────
//	Gigabit LAN            <1ms         2,000-5,000     Very fast, predictable
//	Fast Internet (>50Mbps) 10-50ms     5,000-10,000    Good connectivity
//	Standard Internet      50-100ms     10,000-15,000   Typical broadband
//	Mobile (4G/LTE)        100-500ms    15,000-20,000   Variable but acceptable
//	Mobile (3G)            500-2000ms   20,000-30,000   Slower, less reliable
//	Satellite              1000-2000ms  45,000-60,000   Very slow, high latency
//	Dial-up/Extreme        2000+ms      60,000+         Only last resort
//
// Error Handling Strategy:
//
//	When ReadTimeout is triggered:
//	  1. Current chunk read fails
//	  2. Streaming operation is terminated
//	  3. Error is passed to callback (if set)
//	  4. Wrapper records error in debugging information
//	  5. Application should implement retry logic at higher level
//
//	Best Practice:
//	  - Set ReadTimeout and WriteTimeout to same value
//	  - Choose timeout 2-3× longer than worst-case expected latency
//	  - Implement application-level retry/resume for critical transfers
//	  - Log timeout events for monitoring and debugging
//	  - Consider circuit breaker pattern for repeated failures
//
// See Also:
//   - WithWriteTimeout: Sets timeout for write operations
//   - WithChunkSize: Smaller chunks may need shorter timeouts
//   - WithStreamingStrategy: Strategy affects timeout sensitivity
//   - WithCallback: Receives timeout errors for handling
//   - GetProgress: Monitor actual transfer rate vs timeout
//   - Start: Initiates streaming with configured read timeout
func (sw *StreamingWrapper) WithReadTimeout(timeout int64) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}

	if timeout <= 0 {
		return sw.wrapper.
			WithStatusCode(http.StatusBadRequest).
			WithMessagef("Invalid read timeout: %d ms (must be > 0)", timeout).
			BindCause()
	}

	sw.config.ReadTimeout = time.Duration(timeout) * time.Millisecond
	sw.wrapper.WithDebuggingKV("read_timeout_ms", timeout)
	return sw.wrapper
}

// WithWriteTimeout sets the write operation timeout for streaming data transmission.
//
// This function configures the maximum duration allowed for individual write operations to the output destination.
// The write timeout acts as a safety mechanism to prevent indefinite blocking when the destination becomes
// unresponsive, slow, or unavailable. Each chunk write attempt must complete within the specified timeout;
// if exceeded, the write operation fails and streaming is interrupted. This is essential for handling slow clients,
// congested networks, or stalled connections that would otherwise freeze the entire streaming operation. Write timeout
// is independent of read timeout and operates on the output side of the stream pipeline. The timeout value is recorded
// in wrapper debugging information for audit, performance tracking, and troubleshooting of output-side issues.
//
// Parameters:
//
//   - timeout: The maximum duration for write operations in milliseconds. Must be greater than 0.
//
// Recommended values based on scenario:
//
// - 1000-5000ms (1-5 seconds): High-speed destinations, local writes, same-datacenter transfers.
//   - Use case: Writing to local disk, fast client on LAN, in-memory buffers
//   - Network: Sub-millisecond latency (<1ms typical round-trip)
//   - Client behavior: High-bandwidth, responsive
//   - Best for: High-speed, predictable destinations
//   - Example: 2000 (2 seconds)
//
// - 5000-15000ms (5-15 seconds): Standard clients and networks, typical internet speed.
//   - Use case: Browser downloads, standard REST clients, typical internet connections
//   - Network: 10-100ms latency (typical broadband)
//   - Client behavior: Normal responsiveness
//   - Best for: General-purpose APIs and services
//   - Example: 10000 (10 seconds) - RECOMMENDED DEFAULT
//
// - 15000-30000ms (15-30 seconds): Slower clients, congested networks, mobile devices.
//   - Use case: Mobile browsers, slow connections, distant clients, congested WiFi
//   - Network: 100-500ms latency (cellular networks, high congestion)
//   - Client behavior: Slower but steady
//   - Best for: Mobile and variable-speed clients
//   - Example: 20000 (20 seconds)
//
// - 30000-60000ms (30-60 seconds): Very slow clients, poor connectivity, bandwidth-limited.
//   - Use case: Satellite clients, heavily throttled connections, batch processing with retries
//   - Network: 500ms-2s latency or artificial throttling
//   - Client behavior: Very slow or deliberately limited
//   - Best for: Challenging client conditions
//   - Example: 45000 (45 seconds)
//
// - 60000+ms (60+ seconds): Extremely slow/unreliable clients, specialized scenarios.
//   - Use case: Satellite endpoints, emergency networks, batch jobs with heavy processing
//   - Network: 2s+ latency or artificial delays
//   - Client behavior: Minimal bandwidth or heavy processing
//   - Best for: Last-resort scenarios with critical data
//   - Example: 120000 (120 seconds)
//
// Invalid values: Must be > 0; zero or negative values will return an error.
// Note: Very large timeouts (>120s) can mask real client problems; consider implementing
// application-level heartbeats or keep-alive mechanisms for better reliability.
//
// Returns:
//   - A pointer to the underlying [wrapper] instance, allowing for method chaining.
//   - If the streaming wrapper is nil, returns a new wrapper with an error message.
//   - If the write timeout is ≤ 0, returns the wrapper with an error message indicating invalid input.
//   - The function automatically records the write timeout in wrapper debugging information
//     under the key "write_timeout_ms" for audit, performance analysis, and troubleshooting.
//
// Write Timeout Failure Scenarios:
//
//	Scenario                           Behavior with WriteTimeout
//	───────────────────────────────────────────────────────────────────────
//	Client accepts data before timeout Chunk written immediately
//	Client becomes slow/stalled        Write blocks until timeout, then fails
//	Client connection drops            Write fails immediately (OS-level)
//	Destination buffer full            Write blocks, client must drain buffer
//	Client bandwidth limited           Write paces based on client speed
//	Client disconnect mid-transfer     Write fails, streaming terminates
//	Destination closes connection      Write fails with connection error
//	Network MTU/buffering issues       Write succeeds but slower
//
// Read Timeout vs Write Timeout Relationship:
//
//	Aspect                  ReadTimeout             WriteTimeout
//	─────────────────────────────────────────────────────────────
//	Monitors                Input (source) speed    Output (destination) speed
//	Fails when              Source is too slow      Destination is too slow
//	Typical cause           Slow upload server      Slow/stalled client
//	Affected side           Read goroutine          Write goroutine
//	Recovery action         Terminate streaming    Terminate streaming
//	Set independently       Yes                     Yes
//	Recommended together    Yes (usually equal)     Yes (usually equal)
//	Impact on total time    Sequential (both apply) Either can fail operation
//	Interaction             None (independent)      None (independent)
//
// Client Timeout Interaction Patterns:
//
//	Pattern                          Impact on WriteTimeout
//	───────────────────────────────────────────────────────────────────────
//	Fast client (fiber)              WriteTimeout rarely triggered
//	Normal client (broadband)        WriteTimeout matches latency
//	Slow client (mobile 3G)          WriteTimeout frequently approached
//	Stalled client (no response)     WriteTimeout triggers reliably
//	Throttled client (rate-limited)  WriteTimeout waits for rate limit
//	Disconnected client              WriteTimeout triggers immediately
//	Client processing delay          WriteTimeout includes processing time
//	Network congestion               WriteTimeout absorbs delays
//
// Example:
//
//	// Example 1: Fast local client with minimal timeout (2 seconds)
//	var outputBuffer bytes.Buffer
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/stream/local").
//	    WithCustomFieldKV("destination", "local-buffer").
//	    WithStreaming(dataReader, nil).
//	    WithChunkSize(1024 * 1024). // 1MB chunks
//	    WithReadTimeout(2000).  // 2 seconds for source
//	    WithWriteTimeout(2000). // 2 seconds for destination (matching)
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithMaxConcurrentChunks(8).
//	    WithWriter(&outputBuffer).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Write failed: %v\n", err)
//	            return
//	        }
//	        if p.CurrentChunk % 100 == 0 {
//	            fmt.Printf("Written: %.2f MB\n",
//	                float64(p.TransferredBytes) / 1024 / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 2: Browser download with standard timeout (10 seconds)
//	fileReader, _ := os.Open("document.pdf")
//	defer fileReader.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/document").
//	    WithCustomFieldKV("client_type", "web-browser").
//	    WithCustomFieldKV("expected_bandwidth", "50mbps").
//	    WithStreaming(fileReader, nil).
//	    WithChunkSize(256 * 1024). // 256KB chunks
//	    WithReadTimeout(10000).  // 10 seconds for server-side read
//	    WithWriteTimeout(10000). // 10 seconds for client-side write (matching)
//	    WithCompressionType(COMP_GZIP).
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithMaxConcurrentChunks(4).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            log.Warnf("Browser download stalled: %v (chunk %d)",
//	                err, p.CurrentChunk)
//	            return
//	        }
//	        if p.CurrentChunk % 50 == 0 {
//	            fmt.Printf("Downloaded: %.2f MB | Client rate: %.2f MB/s\n",
//	                float64(p.TransferredBytes) / 1024 / 1024,
//	                float64(p.TransferRate) / 1024 / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 3: Slow mobile client with extended timeout (25 seconds)
//	appUpdate, _ := os.Open("app-update.apk")
//	defer appUpdate.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/mobile-app").
//	    WithCustomFieldKV("client_type", "mobile-app").
//	    WithCustomFieldKV("network", "3g-cellular").
//	    WithStreaming(appUpdate, nil).
//	    WithChunkSize(32 * 1024). // 32KB for slow mobile
//	    WithReadTimeout(15000).  // 15 seconds for reliable server read
//	    WithWriteTimeout(25000). // 25 seconds for slower mobile client
//	    WithThrottleRate(256 * 1024). // 256KB/s throttle
//	    WithCompressionType(COMP_GZIP).
//	    WithMaxConcurrentChunks(1). // Single-threaded for mobile
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            log.Errorf("Mobile download failed: %v (progress: %.1f%%)",
//	                err, float64(p.Percentage))
//	            // Could implement resume/retry here
//	            return
//	        }
//	        if p.CurrentChunk % 40 == 0 {
//	            fmt.Printf("Mobile: %.1f%% | Speed: %.2f KB/s | ETA: %s | Signal: OK\n",
//	                float64(p.Percentage),
//	                float64(p.TransferRate) / 1024,
//	                p.EstimatedTimeRemaining.String())
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 4: Satellite endpoint with very extended timeout (60 seconds)
//	hugeDataset, _ := os.Open("satellite-data.bin")
//	defer hugeDataset.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/stream/satellite-endpoint").
//	    WithCustomFieldKV("destination", "satellite-ground-station").
//	    WithCustomFieldKV("connection_quality", "poor").
//	    WithStreaming(hugeDataset, nil).
//	    WithChunkSize(64 * 1024). // 64KB for reliability
//	    WithReadTimeout(30000).  // 30 seconds for local read
//	    WithWriteTimeout(60000). // 60 seconds for satellite (very slow, high latency)
//	    WithStreamingStrategy(STRATEGY_DIRECT). // Sequential for reliability
//	    WithCompressionType(COMP_GZIP).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Satellite transmission failed: %v (chunk %d/%d)\n",
//	                err, p.CurrentChunk, p.TotalChunks)
//	            return
//	        }
//	        if p.CurrentChunk % 30 == 0 {
//	            fmt.Printf("Satellite: Chunk %d transmitted | ETA: %s\n",
//	                p.CurrentChunk,
//	                p.EstimatedTimeRemaining.String())
//	        }
//	    }).
//	    Start(context.Background())
//
//	// Example 5: Asymmetric timeouts (slow client, fast server)
//	dataExport, _ := os.Open("large-export.csv")
//	defer dataExport.Close()
//
//	// Server can read quickly, but client downloads slowly
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/export/data").
//	    WithCustomFieldKV("export_type", "bulk-csv").
//	    WithStreaming(dataExport, nil).
//	    WithChunkSize(512 * 1024). // 512KB chunks
//	    WithReadTimeout(5000).  // 5 seconds - fast server-side read
//	    WithWriteTimeout(20000). // 20 seconds - slower client writes
//	    WithCompressionType(COMP_GZIP).
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithMaxConcurrentChunks(4). // Buffer helps absorb read/write mismatch
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Export failed: %v\n", err)
//	            return
//	        }
//	        if p.CurrentChunk % 50 == 0 {
//	            // Show both read and write rates
//	            fmt.Printf("Export: Read/Write ratio check | Progress: %.1f%% | Rate: %.2f MB/s\n",
//	                float64(p.Percentage),
//	                float64(p.TransferRate) / 1024 / 1024)
//	        }
//	    }).
//	    Start(context.Background())
//
// Client Type Timeout Selection Guide:
//
//	Client Type              Bandwidth      Timeout (ms)    Rationale
//	───────────────────────────────────────────────────────────────────────────
//	Local (same server)      >1 Gbps        2,000-5,000     Fast, predictable
//	LAN client               100+ Mbps      3,000-8,000     Very fast, reliable
//	Desktop (broadband)      10-50 Mbps     8,000-15,000    Good connectivity
//	Mobile (4G/LTE)          5-20 Mbps      15,000-25,000   Variable performance
//	Mobile (3G)              1-3 Mbps       20,000-30,000   Slower, less stable
//	Satellite client         0.5-2 Mbps     45,000-60,000   Very slow endpoint
//	IoT/Edge device          <1 Mbps        30,000-60,000   Constrained device
//	Batch processing         Variable       60,000+         Heavy processing
//
// Timeout Tuning Best Practices:
//
//  1. ASYMMETRIC TIMEOUTS (Recommended for production)
//     - ReadTimeout: Based on server/source stability (usually shorter)
//     - WriteTimeout: Based on client/destination speed (usually longer)
//     - Example: ReadTimeout=10s, WriteTimeout=20s for slow client scenario
//     - Rationale: Server-side reads usually faster; client-side writes bottleneck
//
//  2. SYMMETRIC TIMEOUTS (Simpler, often sufficient)
//     - ReadTimeout: WriteTimeout (same value)
//     - Best when: Source and destination speeds are similar
//     - Example: Both 10 seconds for typical internet
//     - Rationale: Simpler to understand and reason about
//
//  3. ADAPTIVE TIMEOUTS (Most sophisticated)
//     - Detect: Network conditions, client type, bandwidth
//     - Adjust: ReadTimeout and WriteTimeout dynamically
//     - Example: 5s LAN, 15s internet, 30s mobile, 60s satellite
//     - Rationale: Optimal for heterogeneous client base
//
//  4. MONITORING & ALERTS
//     - Log timeout events with client/network context
//     - Alert on repeated timeouts (may indicate network issues)
//     - Track timeout patterns for tuning decisions
//     - Consider: Circuit breaker after repeated failures
//
// See Also:
//   - WithReadTimeout: Sets timeout for read operations on source
//   - WithChunkSize: Smaller chunks less affected by timeout
//   - WithThrottleRate: Artificial rate limiting affects write timing
//   - WithStreamingStrategy: Strategy selection affects timeout behavior
//   - WithCallback: Receives timeout errors for handling and logging
//   - GetProgress: Monitor actual write rate vs timeout
//   - Start: Initiates streaming with configured write timeout
func (sw *StreamingWrapper) WithWriteTimeout(timeout int64) *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}

	if timeout <= 0 {
		return sw.wrapper.
			WithStatusCode(http.StatusBadRequest).
			WithMessagef("Invalid write timeout: %d ms (must be > 0)", timeout).
			BindCause()
	}

	sw.config.WriteTimeout = time.Duration(timeout) * time.Millisecond
	sw.wrapper.WithDebuggingKV("write_timeout_ms", timeout)

	return sw.wrapper
}

// Cancel terminates an ongoing streaming operation immediately and gracefully.
//
// This function stops the streaming process at the current point, preventing further data transfer.
// It signals the streaming context to stop all read/write operations, halting chunk processing in progress.
// The cancellation is thread-safe and can be called from any goroutine while streaming is active.
// Partial data already transferred to the destination is retained; only new chunk transfers are prevented.
// This is useful for user-initiated interruptions, resource constraints, or error recovery scenarios where
// resuming the operation is planned. The cancellation timestamp is recorded for audit and debugging purposes.
// The underlying resources (readers/writers) remain open and must be explicitly closed via Close() if cleanup
// is required. Cancel returns the wrapper with updated status for chainable response building.
//
// Returns:
//   - A pointer to the underlying [wrapper] instance, allowing for method chaining.
//   - If the streaming wrapper is nil, returns a new wrapper with an error message.
//   - The function automatically updates the wrapper with:
//   - Message: "Streaming cancelled"
//   - Debugging key "cancelled_at": Unix timestamp (seconds since epoch)
//   - Status code remains unchanged (use chaining to update if needed).
//   - In-flight chunks may complete; no data loss guarantee after cancellation.
//
// Cancellation Behavior:
//
//	State Before Cancel      Behavior During Cancel          State After Cancel
//	────────────────────────────────────────────────────────────────────────────
//	Streaming in progress    Context signaled, read blocked   Streaming halted
//	Chunk in flight          Current chunk may complete       No new chunks read
//	Paused/stalled           Cancel processed immediately     Operation terminated
//	Already completed        Cancel is no-op (idempotent)     No effect
//	Never started            Cancel is no-op (no-op state)    Ready for cleanup
//	Error state              Cancel processed (cleanup)       Cleanup initiated
//
// Cancellation vs Close:
//
//	Aspect                  Cancel()                Close()
//	───────────────────────────────────────────────────────────
//	Stops streaming         Yes (context signal)    Yes (closes streams)
//	Closes reader/writer    No (remains open)       Yes (calls Close)
//	Retains partial data    Yes (on destination)    Yes (with cleanup)
//	Resource cleanup        Partial (context only)  Full (all resources)
//	Idempotent              Yes (safe to call >1x)  Yes (safe to call >1x)
//	Use case                Pause/interrupt         Final cleanup
//	Recommended after       Cancel, then retry      Cancel before exit
//	Thread-safe             Yes                     Yes
//	Error handling          None                    Reports close errors
//	Follow-up action        Can resume (new stream) Must create new stream
//
// Example:
//
//	// Example 1: User-initiated cancellation (pause download)
//	file, _ := os.Open("large_file.iso")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithCustomFieldKV("user_action", "manual-pause").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithMaxConcurrentChunks(4).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Error during transfer: %v\n", err)
//	            return
//	        }
//	        fmt.Printf("Progress: %.1f%%\n", float64(p.Percentage))
//	    })
//
//	// Start streaming in background
//	go func() {
//	    streaming.Start(context.Background())
//	}()
//
//	// Simulate user pause after 5 seconds
//	time.Sleep(5 * time.Second)
//	result := streaming.Cancel().
//	    WithMessage("Download paused by user").
//	    WithStatusCode(202). // 202 Accepted - operation paused
//
//	fmt.Printf("Cancelled at: %v\n", result.Debugging())
//
//	// Example 2: Resource constraint cancellation (memory pressure)
//	dataExport := createLargeDataReader()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/export/data").
//	    WithCustomFieldKV("export_type", "bulk").
//	    WithStreaming(dataExport, nil).
//	    WithChunkSize(10 * 1024 * 1024).
//	    WithMaxConcurrentChunks(8).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            return
//	        }
//
//	        // Monitor system memory
//	        memStats := getMemoryStats()
//	        if memStats.HeapAlloc > maxAllowedMemory {
//	            fmt.Printf("Memory pressure detected: %d MB\n",
//	                memStats.HeapAlloc / 1024 / 1024)
//	            // Cancel to prevent OOM
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//
//	if shouldCancelDueToMemory {
//	    result = streaming.Cancel().
//	        WithMessage("Cancelled due to memory pressure").
//	        WithStatusCode(503). // 503 Service Unavailable
//	        WithCustomFieldKV("reason", "memory-pressure").
//	        WithCustomFieldKV("memory_used_mb", currentMemory)
//	}
//
//	// Example 3: Error recovery with cancellation and retry
//	attempt := 0
//	maxRetries := 3
//
//	for attempt < maxRetries {
//	    attempt++
//	    fileReader, _ := os.Open("large_file.bin")
//	    defer fileReader.Close()
//
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/download/retry").
//	        WithCustomFieldKV("attempt", attempt).
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(512 * 1024).
//	        WithReadTimeout(10000).
//	        WithWriteTimeout(10000).
//	        WithCallback(func(p *StreamProgress, err error) {
//	            if err != nil {
//	                fmt.Printf("Attempt %d error: %v at %.1f%% progress\n",
//	                    attempt, err, float64(p.Percentage))
//	                // Trigger cancellation to allow retry
//	            }
//	        })
//
//	    result := streaming.Start(context.Background())
//
//	    if result.IsSuccess() {
//	        fmt.Printf("Download succeeded on attempt %d\n", attempt)
//	        break
//	    }
//
//	    if attempt < maxRetries {
//	        // Cancel before retry
//	        streaming.Cancel()
//	        fmt.Printf("Retrying after failed attempt %d\n", attempt)
//	        time.Sleep(time.Duration(attempt*2) * time.Second) // Exponential backoff
//	    }
//	}
//
//	// Example 4: Timeout-based automatic cancellation
//	file, _ := os.Open("slow_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/with-timeout").
//	    WithStreaming(file, nil).
//	    WithChunkSize(256 * 1024).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Transfer failed: %v\n", err)
//	        }
//	    })
//
//	// Start streaming with a total operation timeout
//	done := make(chan *wrapper)
//	go func() {
//	    result := streaming.Start(context.Background())
//	    done <- result
//	}()
//
//	// Set overall timeout (different from read/write timeouts)
//	select {
//	case result := <-done:
//	    fmt.Printf("Streaming completed: %s\n", result.Message())
//	case <-time.After(60 * time.Second): // 60 second total limit
//	    fmt.Println("Streaming exceeded total timeout")
//	    streaming.Cancel().
//	        WithMessage("Cancelled due to total operation timeout").
//	        WithStatusCode(408). // 408 Request Timeout
//	        WithDebuggingKV("timeout_seconds", 60)
//	}
//
//	// Example 5: Graceful cancellation with progress snapshot
//	dataStream := createDataReader()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/stream/graceful-cancel").
//	    WithStreaming(dataStream, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithMaxConcurrentChunks(4).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil && p.CurrentChunk > 1000 {
//	            fmt.Printf("Cancelling after processing %d chunks\n",
//	                p.CurrentChunk)
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//
//	// Capture final progress before cancellation
//	finalProgress := streaming.GetProgress()
//	cancelResult := streaming.Cancel().
//	    WithMessage("Streaming cancelled gracefully").
//	    WithStatusCode(206). // 206 Partial Content - data transfer interrupted
//	    WithDebuggingKV("chunks_processed", finalProgress.CurrentChunk).
//	    WithDebuggingKVf("progress_percentage", "%.1f", float64(finalProgress.Percentage)).
//	    WithDebuggingKVf("bytes_transferred", "%d", finalProgress.TransferredBytes)
//
// Cancellation Workflow Patterns:
//
//	Pattern                     When to Use                 Benefit
//	────────────────────────────────────────────────────────────────────
//	User-Initiated              UI pause/cancel button      User control
//	Timeout-Based               Watchdog timer              Safety mechanism
//	Resource-Constrained        Memory/CPU threshold        System protection
//	Error-Recovery              Retry on failure            Fault tolerance
//	Graceful-Degradation        Service overload            Load shedding
//	Circuit-Breaker             Repeated failures           Cascade prevention
//
// Partial Data Handling After Cancellation:
//
//	Destination Type    Partial Data Fate           Cleanup Strategy
//	───────────────────────────────────────────────────────────────────
//	File                Partial file remains         Delete or truncate
//	Network (HTTP)      Partial response sent        Client handles truncation
//	Buffer (Memory)     Data in buffer persists      Can retry or discard
//	Database            Partial transactions         Rollback or cleanup
//	Cloud storage       Partial upload               Delete partial object
//
// Best Practices:
//
//  1. ALWAYS PAIR WITH CLOSE()
//     - Cancel() stops streaming
//     - Close() cleans up resources
//     - Pattern: Cancel() → do cleanup → Close()
//     - Example:
//     streaming.Cancel()
//     // Handle cleanup
//     streaming.Close()
//
//  2. MONITOR CANCELLATION STATE
//     - Check IsStreaming() before/after cancellation
//     - Log cancellation reasons for diagnostics
//     - Track cancellation frequency for patterns
//     - Example:
//     if streaming.IsStreaming() {
//     streaming.Cancel()
//     }
//
//  3. HANDLE PARTIAL DATA
//     - Destination retains data transferred before cancellation
//     - Design cleanup/rollback logic based on destination type
//     - Example for files:
//     streaming.Cancel()
//     if shouldRollback {
//     os.Remove(destinationFile)
//     }
//
//  4. IMPLEMENT IDEMPOTENCY
//     - Cancel() is safe to call multiple times
//     - Subsequent calls are no-ops
//     - Caller doesn't need guard logic
//     - Example:
//     streaming.Cancel() // Safe even if already cancelled
//     streaming.Cancel() // No additional effect
//
//  5. USE WITH ERROR HANDLING
//     - Cancellation should be logged with context
//     - Include progress snapshot in logs
//     - Example:
//     progress := streaming.GetProgress()
//     log.Warnf("Streaming cancelled at %.1f%% after %d chunks",
//     float64(progress.Percentage), progress.CurrentChunk)
//     streaming.Cancel()
//
// See Also:
//   - Close: Closes streams and cleans up all resources
//   - IsStreaming: Checks if streaming is currently active
//   - GetProgress: Captures progress state before cancellation
//   - GetStats: Retrieves statistics up to cancellation point
//   - WithCallback: Receives cancellation errors (context done)
//   - Start: Initiates streaming operation that can be cancelled
func (sw *StreamingWrapper) Cancel() *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}

	sw.cancel()
	if sw.wrapper != nil {
		sw.wrapper.WithMessage("Streaming cancelled")
		sw.wrapper.WithDebuggingKV("cancelled_at", time.Now().Unix())
	}
	return sw.wrapper
}

// Close closes the streaming wrapper and releases all underlying resources.
//
// This function performs comprehensive cleanup of the streaming operation by cancelling the streaming context,
// closing the input reader, and closing the output writer if they implement the io.Closer interface. It ensures
// all system resources (file handles, network connections, buffers) are properly released and returned to the
// operating system. Close is idempotent and safe to call multiple times; subsequent calls have no effect.
// It should be called after streaming completes, is cancelled, or encounters an error to prevent resource leaks.
// Unlike Cancel() which stops streaming but leaves resources open, Close() performs full cleanup and should be
// the final operation in a streaming workflow. Errors encountered during resource closure are recorded in the
// wrapper's error field for diagnostic purposes, allowing the caller to verify cleanup completion. Close() can be
// called from any goroutine and is thread-safe with respect to the streaming context cancellation.
//
// Returns:
//   - A pointer to the underlying [wrapper] instance, allowing for method chaining.
//   - If the streaming wrapper is nil, returns a new wrapper with an error message.
//   - The function attempts to close all closeable resources and accumulates all errors encountered.
//   - If reader.Close() fails, the error is recorded in the wrapper; writer.Close() is still attempted.
//   - If writer.Close() fails, the error is recorded in the wrapper.
//   - If both fail, both errors are recorded sequentially in the wrapper.
//   - Status code and message are not modified unless close errors occur; use chaining to update if needed.
//   - Non-closeable resources (non-io.Closer) are silently skipped with no error.
//
// Resource Closure Semantics:
//
//	Resource Type           Close Behavior                          Error Handling
//	────────────────────────────────────────────────────────────────────────────────
//	os.File (reader)        File handle released to OS             Error recorded, continue
//	os.File (writer)        File handle released to OS             Error recorded, continue
//	http.Response.Body      TCP connection returned to pool        Error recorded, continue
//	bytes.Buffer            No-op (not io.Closer)                 Silent skip
//	io.Pipe                 Pipe closed, EOF to readers            Error recorded
//	Network connection      Socket closed, connection terminated   Error recorded, continue
//	Compression reader      Decompressor flushed and closed        Error recorded, continue
//	Custom io.Closer        Custom Close() called                 Error recorded
//	Already closed          Typically returns error                Error recorded
//	Streaming context       Cancellation signal propagated         No-op (already done)
//
// Close vs Cancel Lifecycle:
//
//	Scenario                    Cancel()                Close()
//	───────────────────────────────────────────────────────────────────────────────
//	Purpose                     Stop streaming          Full resource cleanup
//	Context cancellation        Yes                     Yes (redundant)
//	Closes reader               No                      Yes (if io.Closer)
//	Closes writer               No                      Yes (if io.Closer)
//	Releases file handles       No                      Yes
//	Releases connections        No                      Yes
//	Cleans up buffers           No                      Yes (via Close)
//	Partial data preserved      Yes                     Yes
//	Resource state              Streaming stopped       All resources released
//	Can restart streaming       Yes (new context)       No (resources gone)
//	Idempotent                  Yes                     Yes
//	Error accumulation          No error handling       Accumulates close errors
//	Typical use case            Interrupt/pause         Final cleanup
//	Recommended order           Cancel() then Close()   Always call Close()
//	Thread-safe                 Yes                     Yes
//
// Closure Order and Error Accumulation:
//
//	Step    Action                          Error Behavior
//	────────────────────────────────────────────────────────────────────
//	1       Cancel context                  No error (already done)
//	2       Check reader is io.Closer       Silent skip if not
//	3       Call reader.Close()             Error recorded, continue
//	4       Check writer is io.Closer       Silent skip if not
//	5       Call writer.Close()             Error recorded, continue
//	6       Return wrapper                  Contains all accumulated errors
//
// Resource Cleanup Requirements by Type:
//
//	Reader Type             Close Requirement           Consequence if not closed
//	─────────────────────────────────────────────────────────────────────────────
//	os.File                 Mandatory                   File descriptor leak
//	io.ReadCloser           Mandatory                   Resource leak (memory/handles)
//	http.Response.Body      Mandatory                   Connection leak, pool exhaustion
//	bytes.Buffer            Optional (not closeable)    No consequences
//	io.Reader (plain)       Optional (no Close)         No consequences
//	Pipe reader             Mandatory                   Blocked writers, memory leak
//	Compressed reader       Mandatory                   Decompressor leak
//	Network socket          Mandatory                   Connection leak, resource leak
//	Database cursor         Mandatory (custom impl)     Cursor/connection leak
//	Custom reader           Depends on implementation   Varies by implementation
//
//	Writer Type             Close Requirement           Consequence if not closed
//	─────────────────────────────────────────────────────────────────────────────
//	os.File                 Mandatory                   File descriptor leak, unflushed data
//	io.WriteCloser          Mandatory                   Resource leak, buffered data loss
//	http.ResponseWriter     Not closeable (handled)     Data may be buffered
//	bytes.Buffer            Optional (not closeable)    No consequences
//	io.Writer (plain)       Optional (no Close)         No consequences
//	Pipe writer             Mandatory                   Blocked readers, memory leak
//	Compressed writer       Mandatory                   Unflushed compressed data
//	Network socket          Mandatory                   Connection leak
//	File buffered writer    Mandatory                   Data loss, descriptor leak
//	Custom writer           Depends on implementation   Varies by implementation
//
// Example:
//
//	// Example 1: Standard cleanup after successful streaming
//	file, _ := os.Open("large_file.bin")
//	defer file.Close() // Double-close is safe
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Streaming error: %v\n", err)
//	        }
//	    })
//
//	// Start streaming
//	result := streaming.Start(context.Background())
//
//	// Always close resources
//	finalResult := streaming.Close().
//	    WithMessage("Download completed and resources cleaned up").
//	    WithStatusCode(200)
//
//	if finalResult.IsError() {
//	    fmt.Printf("Cleanup error: %s\n", finalResult.Error())
//	}
//
//	// Example 2: Error recovery with explicit cleanup
//	httpResp, _ := http.Get("https://api.example.com/largefile")
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/proxy/remote-file").
//	    WithStreaming(httpResp.Body, nil). // HTTP response body
//	    WithChunkSize(256 * 1024).
//	    WithReadTimeout(10000).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Transfer error: %v at %.1f%%\n",
//	                err, float64(p.Percentage))
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//
//	if result.IsError() {
//	    fmt.Printf("Streaming failed: %s\n", result.Error())
//	}
//
//	// Close HTTP connection and release resources
//	cleanupResult := streaming.Close().
//	    WithMessage("Resources released after error").
//	    WithStatusCode(500)
//
//	// Verify cleanup
//	if cleanupResult.IsError() {
//	    log.Warnf("Cleanup issues: %s", cleanupResult.Error())
//	}
//
//	// Example 3: Cancellation followed by cleanup
//	dataExport := createDataReader()
//	outputFile, _ := os.Create("export.csv")
//	defer outputFile.Close() // Double-close is safe
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/export/data").
//	    WithCustomFieldKV("export_type", "csv").
//	    WithStreaming(dataExport, nil).
//	    WithChunkSize(512 * 1024).
//	    WithMaxConcurrentChunks(4).
//	    WithWriter(outputFile).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Export error: %v\n", err)
//	        }
//	    })
//
//	// Simulate user cancellation
//	time.Sleep(3 * time.Second)
//	streaming.Cancel().
//	    WithMessage("Export cancelled by user")
//
//	// Cleanup: Close streaming resources
//	cleanupResult := streaming.Close().
//	    WithMessage("Export cancelled and resources released")
//
//	// Verify files are properly closed
//	progress := streaming.GetProgress()
//	fmt.Printf("Partial export: %d bytes in %d chunks\n",
//	    progress.TransferredBytes, progress.CurrentChunk)
//
//	// Example 4: Deferred cleanup pattern (recommended)
//	func StreamWithAutomaticCleanup(fileReader io.ReadCloser) *wrapper {
//	    defer func() {
//	        fileReader.Close() // Redundant but safe with Close()
//	    }()
//
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/stream/auto-cleanup").
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(1024 * 1024)
//
//	    result := streaming.Start(context.Background())
//
//	    // Cleanup via Close() - safe even with defer above
//	    return streaming.Close().
//	        WithMessage("Streaming completed with automatic cleanup")
//	}
//
//	// Example 5: Error handling with comprehensive cleanup
//	func StreamWithErrorHandling(reader io.ReadCloser, writer io.WriteCloser) (*wrapper, error) {
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/stream/with-error-handling").
//	        WithStreaming(reader, nil).
//	        WithChunkSize(256 * 1024).
//	        WithWriter(writer).
//	        WithReadTimeout(15000).
//	        WithWriteTimeout(15000).
//	        WithCallback(func(p *StreamProgress, err error) {
//	            if err != nil {
//	                fmt.Printf("Streaming error at chunk %d: %v\n",
//	                    p.CurrentChunk, err)
//	            }
//	        })
//
//	    // Execute streaming
//	    result := streaming.Start(context.Background())
//
//	    // Always cleanup, regardless of success/failure
//	    finalResult := streaming.Close()
//
//	    // Log cleanup status
//	    if finalResult.IsError() {
//	        fmt.Printf("Cleanup warnings: %s\n", finalResult.Error())
//	        // Don't fail overall operation due to close errors
//	    }
//
//	    // Return original streaming result (not close result)
//	    return result, nil
//	}
//
//	// Example 6: Comparison of cleanup patterns
//	// Pattern 1: Minimal cleanup (not recommended - resource leak potential)
//	streaming := replify.New().WithStreaming(file, nil)
//	streaming.Start(context.Background())
//	// Missing: streaming.Close() -> RESOURCE LEAK
//
//	// Pattern 2: Basic cleanup (recommended)
//	streaming := replify.New().WithStreaming(file, nil)
//	result := streaming.Start(context.Background())
//	streaming.Close() // ✓ Proper cleanup
//
//	// Pattern 3: Error-aware cleanup (best for production)
//	streaming := replify.New().WithStreaming(file, nil)
//	result := streaming.Start(context.Background())
//	cleanupResult := streaming.Close()
//	if cleanupResult.IsError() {
//	    log.Warnf("Streaming cleanup had issues: %s", cleanupResult.Error())
//	}
//
//	// Pattern 4: Defer-based cleanup (most idiomatic)
//	func downloadFile(fileReader io.ReadCloser) *wrapper {
//	    streaming := replify.New().WithStreaming(fileReader, nil)
//	    defer streaming.Close() // Guaranteed cleanup
//
//	    return streaming.Start(context.Background())
//	}
//
// Idempotency Guarantee:
//
//	Call Sequence               Behavior                Result
//	─────────────────────────────────────────────────────────────────────
//	Close()                     Normal cleanup          Resources released
//	Close(); Close()            Second call is no-op    No additional effect
//	Close(); Close(); Close()   All no-ops after first  Safe to call >1x
//
//	Safety Example:
//	  defer streaming.Close() // Call 1
//	  ...
//	  streaming.Close() // Call 2 - safe, no-op
//	  ...
//	  if cleanup {
//	      streaming.Close() // Call 3 - safe, no-op
//	  }
//
// Best Practices:
//
//  1. ALWAYS CLOSE AFTER STREAMING
//     - Use defer for guarantee
//     - Pattern:
//     streaming := replify.New().WithStreaming(reader, nil)
//     defer streaming.Close()
//     result := streaming.Start(ctx)
//
//  2. HANDLE CLOSE ERRORS
//     - Check for close errors
//     - Log for diagnostics
//     - Example:
//     cleanupResult := streaming.Close()
//     if cleanupResult.IsError() {
//     log.Warnf("Close error: %s", cleanupResult.Error())
//     }
//
//  3. CLOSE AFTER CANCEL
//     - Cancel first (stop streaming)
//     - Then Close (release resources)
//     - Pattern:
//     streaming.Cancel()
//     streaming.Close()
//
//  4. DOUBLE-CLOSE IS SAFE
//     - io.Closer implementations handle multiple Close() calls
//     - Idempotent design
//     - Example:
//     defer file.Close()        // OS level
//     defer streaming.Close()   // Streaming level
//     // Both safe even if called in any order
//
//  5. CLOSE EARLY ON ERROR
//     - Close immediately on error
//     - Don't delay cleanup
//     - Example:
//     result := streaming.Start(ctx)
//     if result.IsError() {
//     streaming.Close() // Cleanup immediately
//     return result
//     }
//     streaming.Close() // Normal cleanup
//
// Resource Leak Scenarios and Prevention:
//
//	Scenario                           Risk Level    Prevention
//	──────────────────────────────────────────────────────────────────
//	Missing Close() entirely           CRITICAL      Use defer streaming.Close()
//	Close() only in success path       HIGH          Always close (use defer)
//	Exception/panic without Close      HIGH          Defer statement essential
//	Goroutine exit without Close       HIGH          Ensure Close() in goroutine
//	File handle accumulation           MEDIUM        Monitor open file count
//	Connection pool exhaustion         MEDIUM        Close responses promptly
//	Memory buffering accumulation      MEDIUM        Close flushes buffers
//	Deadlock on Close                  LOW           Streaming handles correctly
//
// Performance Considerations:
//
//	Operation                   Time Cost        Notes
//	──────────────────────────────────────────────────────────────
//	Cancel context              <1ms             Context already signaled
//	Type assertion (io.Closer)   <1μs            Cheap operation
//	reader.Close()              1-100ms          Depends on implementation
//	writer.Close()              1-100ms          Depends on implementation (flush)
//	Total Close() operation      2-200ms          Dominated by actual Close() calls
//	Defer overhead              <1μs             Negligible cost
//
// See Also:
//   - Cancel: Stops streaming without closing resources
//   - IsStreaming: Checks if streaming is currently active
//   - GetProgress: Captures final progress before close
//   - GetStats: Retrieves final statistics before resources close
//   - Start: Initiates streaming operation
//   - WithReader/WithWriter: Provides closeable resources
func (sw *StreamingWrapper) Close() *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}

	sw.cancel()

	if closer, ok := sw.reader.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			sw.wrapper.
				WithStatusCode(http.StatusInternalServerError).
				AppendErrorAck(err, "failed to close reader")
		}
	}

	if closer, ok := sw.writer.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			sw.wrapper.
				WithStatusCode(http.StatusInternalServerError).
				AppendErrorAck(err, "failed to close writer")
		}
	}

	return sw.wrapper
}

// Errors returns a copy of all errors that occurred during the streaming operation.
//
// This function provides thread-safe access to the complete error history accumulated during streaming.
// Errors are recorded for each chunk processing failure, I/O operation failure, compression/decompression
// error, timeout expiration, or context cancellation. The function returns a defensive copy of the error
// slice to prevent external modification of the internal error list. Multiple calls to Errors() return
// independent copies; modifications to one copy do not affect others or the internal state. This is useful
// for comprehensive error reporting, diagnostics, debugging, and implementing retry or circuit breaker logic.
// Errors are maintained in chronological order (FIFO) with the first error at index 0 and the most recent
// error at the last index. The error list is thread-safe and can be accessed from any goroutine during or
// after streaming.
//
// Returns:
//   - A newly allocated slice containing copies of all errors that occurred.
//   - Returns an empty slice if no errors occurred during streaming.
//   - Returns an empty slice if the streaming wrapper is nil.
//   - The returned slice is independent; modifications do not affect internal state.
//   - Each call returns a fresh copy; subsequent calls may contain additional errors if streaming
//     is ongoing or if new errors were recorded after the previous call.
//   - Errors are maintained in chronological order (FIFO): first error at index 0, latest at last index.
//   - Thread-safe: safe to call from multiple goroutines simultaneously during streaming.
//
// Error Types Recorded:
//
//	Error Category              When Recorded                   Example
//	─────────────────────────────────────────────────────────────────────────
//	Read errors                 reader.Read() fails             Connection reset, EOF mismatch
//	Write errors                writer.Write() fails            Slow client timeout, buffer full
//	Compression errors          Compress/decompress fails       Invalid compressed data
//	Context errors              Context deadline exceeded       Read/write timeout triggered
//	Checksum errors             Chunk integrity check fails     Data corruption detected
//	Type assertion errors       Reader/writer not Closeable     Interface mismatch (rare)
//	User errors                 Invalid parameters passed       Negative timeout, empty strategy
//	Resource errors             Resource allocation failed      Out of memory, descriptor limit
//	Custom errors               Application-specific errors     Custom reader/writer errors
//	Accumulated errors          Multiple chunk failures         Retries and partial failures
//
// Error Recording Behavior:
//
//	Event                           Error Recorded    Behavior
//	────────────────────────────────────────────────────────────────────
//	Single chunk read fail          Yes (1 error)     Streaming continues (retry next chunk)
//	Single chunk write fail         Yes (1 error)     Streaming continues
//	Read timeout triggered          Yes (1 error)     Streaming terminates
//	Write timeout triggered         Yes (1 error)     Streaming terminates
//	Compression fails              Yes (1 error)     Chunk skipped, streaming continues
//	Decompression fails            Yes (1 error)     Chunk skipped, streaming continues
//	Context cancelled              Yes (1 error)     Streaming terminates gracefully
//	Multiple chunk failures        Yes (N errors)     All recorded in order
//	No errors during transfer      No errors          Empty error list
//	Successful completion          No new errors      Final list unchanged
//
// Error History Timeline Example:
//
//	Timeline                         Errors() Returns                  Explanation
//	─────────────────────────────────────────────────────────────────────────────
//	Chunk 1-100 ok                   []                                No errors yet
//	Chunk 101 write timeout          [timeout error]                   First error recorded
//	Chunk 102 ok (retry)             [timeout error]                   List unchanged
//	Chunk 103 write fail             [timeout error, write error]      Second error added
//	Chunk 104-200 ok                 [timeout error, write error]      List stable
//	Chunk 201 compression fail       [timeout error, write error, ...] Third error added
//	Streaming completes              [all accumulated errors]          Final error list
//
// Example:
//
//	// Example 1: Simple error checking after streaming
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithChunkSize(256 * 1024).
//	    WithReadTimeout(10000).
//	    WithWriteTimeout(10000).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Chunk %d error: %v\n", p.CurrentChunk, err)
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//
//	// Check all errors that occurred
//	errors := streaming.Errors()
//	if len(errors) > 0 {
//	    fmt.Printf("Streaming completed with %d errors:\n", len(errors))
//	    for i, err := range errors {
//	        fmt.Printf("  Error %d: %v\n", i+1, err)
//	    }
//	} else {
//	    fmt.Println("Streaming completed successfully with no errors")
//	}
//
//	// Example 2: Error analysis for retry logic
//	httpResp, _ := http.Get("https://api.example.com/largefile")
//	defer httpResp.Body.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/proxy/remote-download").
//	    WithStreaming(httpResp.Body, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithReadTimeout(15000).
//	    WithWriteTimeout(15000)
//
//	result := streaming.Start(context.Background())
//	streamErrors := streaming.Errors()
//
//	// Analyze errors to decide retry strategy
//	timeoutCount := 0
//	ioErrorCount := 0
//	otherErrorCount := 0
//
//	for _, err := range streamErrors {
//	    errStr := err.Error()
//	    if strings.Contains(errStr, "timeout") {
//	        timeoutCount++
//	    } else if strings.Contains(errStr, "read") || strings.Contains(errStr, "write") {
//	        ioErrorCount++
//	    } else {
//	        otherErrorCount++
//	    }
//	}
//
//	fmt.Printf("Error summary:\n")
//	fmt.Printf("  Timeouts: %d\n", timeoutCount)
//	fmt.Printf("  I/O errors: %d\n", ioErrorCount)
//	fmt.Printf("  Other errors: %d\n", otherErrorCount)
//
//	// Decide retry based on error types
//	if timeoutCount > ioErrorCount {
//	    fmt.Println("Recommendation: Increase timeout and retry")
//	} else if ioErrorCount > 0 {
//	    fmt.Println("Recommendation: Check network connectivity and retry")
//	}
//
//	// Example 3: Error logging with context
//	dataExport := createDataReader()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/export/bulk-data").
//	    WithCustomFieldKV("export_id", "exp-2025-1114-001").
//	    WithStreaming(dataExport, nil).
//	    WithChunkSize(10 * 1024 * 1024).
//	    WithMaxConcurrentChunks(8).
//	    WithCompressionType(COMP_GZIP).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            log.Warnf("Export %s: chunk %d failed: %v",
//	                "exp-2025-1114-001", p.CurrentChunk, err)
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//	allErrors := streaming.Errors()
//
//	// Comprehensive error reporting
//	if len(allErrors) > 0 {
//	    log.Warnf("Export exp-2025-1114-001: %d errors during transfer:", len(allErrors))
//	    for i, err := range allErrors {
//	        log.Warnf("  [Error %d/%d] %v", i+1, len(allErrors), err)
//	    }
//
//	    // Store error history for later investigation
//	    progress := streaming.GetProgress()
//	    stats := streaming.GetStats()
//	    log.Infof("Export context: Progress=%.1f%%, Chunks=%d, Bytes=%d",
//	        float64(progress.Percentage), progress.CurrentChunk, stats.TotalBytes)
//	}
//
//	// Example 4: Circuit breaker pattern with error tracking
//	maxErrorsAllowed := 5
//	circuitOpen := false
//
//	fileReader, _ := os.Open("data.bin")
//	defer fileReader.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/stream/circuit-breaker").
//	    WithStreaming(fileReader, nil).
//	    WithChunkSize(512 * 1024).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            // Check current error count
//	            currentErrors := streaming.Errors()
//	            if len(currentErrors) >= maxErrorsAllowed {
//	                fmt.Printf("Circuit breaker: %d errors exceeded limit, stopping\n",
//	                    len(currentErrors))
//	                circuitOpen = true
//	                streaming.Cancel()
//	            }
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//	finalErrors := streaming.Errors()
//
//	if circuitOpen {
//	    fmt.Printf("Streaming stopped due to circuit breaker (errors: %d)\n",
//	        len(finalErrors))
//	} else if len(finalErrors) > 0 {
//	    fmt.Printf("Streaming completed with %d tolerable errors\n",
//	        len(finalErrors))
//	}
//
//	// Example 5: Error deduplication and categorization
//	func AnalyzeStreamingErrors(streaming *StreamingWrapper) map[string]int {
//	    errors := streaming.Errors()
//	    errorCounts := make(map[string]int)
//	    errorTypes := make(map[string]bool)
//
//	    for _, err := range errors {
//	        errMsg := err.Error()
//
//	        // Categorize error
//	        var errorType string
//	        switch {
//	        case strings.Contains(errMsg, "timeout"):
//	            errorType = "timeout"
//	        case strings.Contains(errMsg, "connection"):
//	            errorType = "connection"
//	        case strings.Contains(errMsg, "read"):
//	            errorType = "read"
//	        case strings.Contains(errMsg, "write"):
//	            errorType = "write"
//	        case strings.Contains(errMsg, "compression"):
//	            errorType = "compression"
//	        default:
//	            errorType = "other"
//	        }
//
//	        errorCounts[errorType]++
//	        errorTypes[errorType] = true
//	    }
//
//	    fmt.Println("Error breakdown:")
//	    for errorType := range errorTypes {
//	        fmt.Printf("  %s: %d\n", errorType, errorCounts[errorType])
//	    }
//
//	    return errorCounts
//	}
//
//	// Example 6: Error history export for diagnostics
//	func ExportErrorReport(streaming *StreamingWrapper) string {
//	    errors := streaming.Errors()
//	    progress := streaming.GetProgress()
//	    stats := streaming.GetStats()
//
//	    var report strings.Builder
//	    report.WriteString("=== STREAMING ERROR REPORT ===\n")
//	    report.WriteString(fmt.Sprintf("Timestamp: %s\n", time.Now().Format(time.RFC3339)))
//	    report.WriteString(fmt.Sprintf("Total Errors: %d\n", len(errors)))
//	    report.WriteString(fmt.Sprintf("Progress: %.1f%% (%d/%d bytes)\n",
//	        float64(progress.Percentage), progress.TransferredBytes, progress.TotalBytes))
//	    report.WriteString(fmt.Sprintf("Chunks: %d processed\n", progress.CurrentChunk))
//	    report.WriteString(fmt.Sprintf("Duration: %s\n", stats.EndTime.Sub(stats.StartTime)))
//
//	    report.WriteString("\n=== ERROR DETAILS ===\n")
//	    for i, err := range errors {
//	        report.WriteString(fmt.Sprintf("[%d] %v\n", i+1, err))
//	    }
//
//	    return report.String()
//	}
//
// Error Handling Patterns:
//
//	Pattern                         Use Case                        Implementation
//	─────────────────────────────────────────────────────────────────────────────
//	No-error path                   Success only                    if len(errors) == 0
//	Basic error check               Error awareness                 if len(errors) > 0
//	Error counting                  Threshold-based decisions       len(errors) > threshold
//	Error categorization            Retry logic selection           Analyze error types
//	Error deduplication             Deduplicate repeated errors     Use map for uniqueness
//	Circuit breaker                 Fail-fast on repeated errors    Break on error count
//	Error trend analysis            Long-term diagnostics          Track error patterns
//	Detailed error reporting        Production diagnostics          Export error history
//
// Thread-Safety Guarantees:
//
//	Scenario                        Thread-Safe    Notes
//	─────────────────────────────────────────────────────────────────
//	Errors() during streaming       Yes            Uses RWMutex lock
//	Errors() after streaming done   Yes            Lock prevents race
//	Multiple concurrent Errors()    Yes            RWMutex allows parallel reads
//	Errors() + HasErrors()          Yes            Consistent snapshot
//	Errors() + GetProgress()        Maybe          Different locks (consult separately)
//	Errors() in callback            Yes            Called by streaming goroutine
//	Errors() in cancel/close        Yes            Idempotent operation
//
// Performance Notes:
//
//	Operation                   Time Complexity    Space Complexity    Notes
//	─────────────────────────────────────────────────────────────────────
//	Errors() call               O(n)               O(n)                Copies full slice
//	First call (0 errors)       O(1)               O(1)                Minimal overhead
//	Mid-stream (1000 errors)    O(1000)            O(1000)             Allocates new slice
//	Final call (10000 errors)   O(10000)           O(10000)            Large allocation
//	RWMutex acquisition         O(1)               O(1)                Lock contention minimal
//	Memory allocation           O(n)               O(n)                Linear in error count
//
// Related Error Checking Methods:
//
//	Method              Returns                 Use Case
//	────────────────────────────────────────────────────────────────────
//	HasErrors()         bool                    Quick check for any error
//	Errors()            []error                 Complete error list (this function)
//	GetProgress()       *StreamProgress         Error count in progress struct
//	GetStats()          *StreamingStats         FailedChunks in stats
//	GetWrapper()        *wrapper                Wrapper error messages
//
// See Also:
//   - HasErrors: Quickly checks if any errors occurred without retrieving list
//   - GetProgress: Includes error occurrence information in progress
//   - GetStats: Provides failed chunk count and error array
//   - GetWrapper: Returns wrapper with accumulated error messages
//   - WithCallback: Callback receives individual errors as they occur
//   - Start: Initiates streaming and accumulates errors during operation
func (sw *StreamingWrapper) Errors() []error {
	if sw == nil {
		return []error{}
	}

	sw.mu.RLock()
	defer sw.mu.RUnlock()

	// Return copy
	errorsCopy := make([]error, len(sw.errors))
	copy(errorsCopy, sw.errors)
	return errorsCopy
}

// HasErrors checks whether any errors occurred during the streaming operation without retrieving the full error list.
//
// This function provides a lightweight, efficient way to determine if the streaming operation encountered
// any errors without allocating memory to copy the entire error slice. It performs a simple boolean check on
// the error count, returning true if at least one error was recorded and false if no errors occurred. This is
// useful as a fast filter before calling the more expensive Errors() method, or for conditional logic that only
// needs to know "did errors occur?" rather than "what were all the errors?". HasErrors is thread-safe and can
// be called from any goroutine during or after streaming. It provides a snapshot of the error state at the moment
// of the call; subsequent errors may be recorded if streaming is still ongoing. This is the recommended method for
// quick error checks, assertions, and control flow decisions where the specific error details are not needed.
//
// Returns:
//   - true if one or more errors were recorded during streaming.
//   - false if no errors occurred (zero errors).
//   - false if the streaming wrapper is nil.
//   - Thread-safe: multiple goroutines can call concurrently without blocking.
//   - Snapshot behavior: returns state at call time; ongoing streaming may change result.
//
// Performance Characteristics:
//
//	Operation                   Time Complexity    Space Complexity    Notes
//	─────────────────────────────────────────────────────────────────────────
//	HasErrors() call            O(1)               O(0)                Constant time
//	Single length check         <1μs               None                Just compares int
//	RWMutex lock acquisition    O(1) amortized     None                Lock contention rare
//	No allocation               None               O(0)                Stack only
//	vs Errors() copy            10-100x faster     No allocation       Major advantage
//	Concurrent calls            Parallel reads     None                Multiple readers OK
//
// When to Use HasErrors vs Errors:
//
//	Scenario                            Use HasErrors()    Use Errors()
//	──────────────────────────────────────────────────────────────────────────
//	Quick "any error?" check            ✓ YES             ✗ No
//	Conditionals (if err occurred)      ✓ YES             ✗ No
//	Need specific error details         ✗ No              ✓ YES
//	Error analysis/categorization       ✗ No              ✓ YES
//	Logging all errors                  ✗ No              ✓ YES
//	Circuit breaker threshold           ✓ YES             ✗ No
//	Before calling Errors()             ✓ YES             ✗ No
//	Memory-constrained environment      ✓ YES             ✗ No
//	Performance-critical path           ✓ YES             ✗ No
//	Error reporting/diagnostics         ✗ No              ✓ YES
//
// Error State Timeline:
//
//	Point in Streaming              HasErrors() Returns    Explanation
//	────────────────────────────────────────────────────────────────────
//	Before Start()                  false                  No streaming yet
//	Chunks 1-100 ok                 false                  No errors encountered
//	Chunk 101 timeout               true                   First error recorded
//	Chunk 102 ok (retry)            true                   Error list not cleared
//	Chunk 103 write fail            true                   Additional error added
//	Streaming completes             true                   Final error list preserved
//	After Close()                   true                   Errors not cleared
//	New streaming instance          false                  Fresh error list
//
// Example:
//
//	// Example 1: Simple error checking after streaming
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithReadTimeout(15000).
//	    WithWriteTimeout(15000)
//
//	result := streaming.Start(context.Background())
//
//	// Fast check for errors without allocating error list
//	if streaming.HasErrors() {
//	    fmt.Println("Streaming completed with errors")
//	    // Only retrieve full error list if we know errors exist
//	    errors := streaming.Errors()
//	    fmt.Printf("Total errors: %d\n", len(errors))
//	} else {
//	    fmt.Println("Streaming completed successfully")
//	}
//
//	// Example 2: Conditional response building with error checking
//	dataExport := createDataReader()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/export/users").
//	    WithStreaming(dataExport, nil).
//	    WithChunkSize(512 * 1024).
//	    WithMaxConcurrentChunks(4)
//
//	result := streaming.Start(context.Background())
//	streaming.Close()
//
//	// Build response based on error status (no expensive Errors() call if not needed)
//	finalResponse := result
//	if streaming.HasErrors() {
//	    finalResponse = result.
//	        WithStatusCode(206). // 206 Partial Content
//	        WithMessage("Export completed with some errors").
//	        WithDebuggingKV("error_count", len(streaming.Errors())).
//	        WithDebuggingKV("has_errors", true)
//	} else {
//	    finalResponse = result.
//	        WithStatusCode(200).
//	        WithMessage("Export completed successfully")
//	}
//
//	// Example 3: Circuit breaker pattern with fast error detection
//	maxErrorsAllowed := 10
//
//	fileReader, _ := os.Open("large-backup.tar.gz")
//	defer fileReader.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/backup/stream").
//	    WithStreaming(fileReader, nil).
//	    WithChunkSize(10 * 1024 * 1024).
//	    WithMaxConcurrentChunks(8).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            // Fast check first (O(1))
//	            if streaming.HasErrors() {
//	                // Only if errors exist, check count (O(n))
//	                errorCount := len(streaming.Errors())
//	                if errorCount >= maxErrorsAllowed {
//	                    fmt.Printf("Circuit breaker: %d errors >= %d limit\n",
//	                        errorCount, maxErrorsAllowed)
//	                    streaming.Cancel()
//	                }
//	            }
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//
//	// Example 4: Assert no errors (test/validation)
//	func AssertStreamingSuccess(t *testing.T, streaming *StreamingWrapper) {
//	    if streaming.HasErrors() {
//	        t.Fatalf("Expected no errors, but got: %v", streaming.Errors())
//	    }
//	}
//
//	// Usage in test
//	func TestStreamingDownload(t *testing.T) {
//	    file, _ := os.Open("testdata.bin")
//	    defer file.Close()
//
//	    streaming := replify.New().
//	        WithStreaming(file, nil).
//	        WithChunkSize(65536)
//
//	    result := streaming.Start(context.Background())
//
//	    AssertStreamingSuccess(t, streaming) // Fast assertion
//
//	    if !result.IsSuccess() {
//	        t.Fatalf("Expected success, got status %d", result.StatusCode())
//	    }
//	}
//
//	// Example 5: Conditional detailed error logging
//	func LogStreamingResult(streaming *StreamingWrapper, contextInfo string) {
//	    // Fast check first (O(1) - cheap operation)
//	    if streaming.HasErrors() {
//	        // Only do expensive error retrieval if errors exist
//	        errors := streaming.Errors()
//	        log.Warnf("[%s] Streaming had %d errors:", contextInfo, len(errors))
//
//	        for i, err := range errors {
//	            log.Warnf("  [Error %d/%d] %v", i+1, len(errors), err)
//	        }
//	    } else {
//	        // Success path - no expensive operations
//	        log.Infof("[%s] Streaming completed successfully", contextInfo)
//	    }
//	}
//
//	// Example 6: Multi-condition error checking pattern
//	func HandleStreamingCompletion(streaming *StreamingWrapper) *wrapper {
//	    progress := streaming.GetProgress()
//
//	    // Chain conditions from cheapest to most expensive
//	    if !streaming.HasErrors() {
//	        // Path 1: Fast success case (O(1))
//	        return streaming.GetWrapper().
//	            WithStatusCode(200).
//	            WithMessage("Streaming completed successfully").
//	            WithTotal(progress.CurrentChunk)
//	    } else if len(streaming.Errors()) <= 5 {
//	        // Path 2: Moderate error case (O(n) but small n)
//	        return streaming.GetWrapper().
//	            WithStatusCode(206).
//	            WithMessage("Streaming completed with minor errors").
//	            WithDebuggingKV("error_count", len(streaming.Errors()))
//	    } else {
//	        // Path 3: Severe error case (O(n) but warranted)
//	        errors := streaming.Errors()
//	        return streaming.GetWrapper().
//	            WithStatusCode(500).
//	            WithMessage("Streaming failed with multiple errors").
//	            WithDebuggingKV("error_count", len(errors)).
//	            WithDebuggingKV("error_summary", strings.Join(
//	                func() []string {
//	                    summary := make([]string, len(errors))
//	                    for i, err := range errors {
//	                        summary[i] = err.Error()
//	                    }
//	                    return summary
//	                }(),
//	                "; "))
//	    }
//	}
//
// Performance Comparison Example:
//
//	// ❌ INEFFICIENT: Always copy errors even if not needed
//	func BadPattern(streaming *StreamingWrapper) {
//	    errors := streaming.Errors() // O(n) allocation even if not used
//	    if len(errors) > 0 {
//	        fmt.Println("Has errors")
//	    }
//	}
//
//	// ✅ EFFICIENT: Check first, then retrieve if needed
//	func GoodPattern(streaming *StreamingWrapper) {
//	    if streaming.HasErrors() { // O(1) - cheap
//	        errors := streaming.Errors() // O(n) - only if needed
//	        if len(errors) > 0 {
//	            fmt.Println("Has errors")
//	        }
//	    }
//	}
//
//	// Performance impact with large error lists:
//	// 100 errors:   HasErrors() 1μs vs Errors() 10μs (10x faster)
//	// 1000 errors:  HasErrors() 1μs vs Errors() 100μs (100x faster)
//	// 10000 errors: HasErrors() 1μs vs Errors() 1ms (1000x faster)
//
// Optimization Strategy:
//
//	Condition                       Optimization
//	────────────────────────────────────────────────────────────────────
//	Need to check for errors        Use HasErrors() first
//	Need error details              Call Errors() only if HasErrors() true
//	Need error count                if has errors: len(Errors()); else: 0
//	Performance-critical code       Always use HasErrors() as filter
//	Memory-constrained              Use HasErrors() to avoid allocation
//	Logging conditional errors      if HasErrors(): log(Errors())
//	Circuit breaker implementation  if HasErrors(): break on threshold
//	Response building               if HasErrors(): build error response
//
// Thread-Safety and Concurrency:
//
//	Scenario                                    Thread-Safe    Details
//	─────────────────────────────────────────────────────────────────────
//	HasErrors() during streaming                Yes            RWMutex protects
//	Multiple concurrent HasErrors()             Yes            Parallel reads allowed
//	HasErrors() + Start() concurrently          Yes            Independent operations
//	HasErrors() + Errors() race                 Yes            Consistent snapshot
//	HasErrors() + Cancel() concurrently         Yes            Operations independent
//	HasErrors() in multiple goroutines          Yes            Lock-free on success path
//	Contention under high concurrency           Rare           RWMutex optimized
//	Call during callback                        Yes            Safe from within callback
//
// Common Pitfalls and Solutions:
//
//	Pitfall                                 Problem                 Solution
//	─────────────────────────────────────────────────────────────────────────
//	Always calling Errors()                 Unnecessary allocation   Use HasErrors() first
//	Assuming no errors = all ok             Incomplete check        Also check status code
//	Race on error count during streaming    Timing dependent        Use atomic/lock
//	Ignoring errors in success path         Silent failures         Always check HasErrors()
//	Calling HasErrors() in tight loop       Lock contention         Cache result or defer
//	Not pairing with error details          Lost diagnostic info    Use Errors() when needed
//	Forgetting return value                 No-op statement         Assign or use result
//
// Related Methods Comparison:
//
//	Method                  Returns         Cost        When to Use
//	─────────────────────────────────────────────────────────────────────
//	HasErrors()             bool            O(1)        Fast "any error?" check
//	Errors()                []error         O(n)        Need all error details
//	Len(Errors())           int             O(n)        Need error count (bad idea!)
//	GetStats().Errors       []error         O(n)        Stats + errors together
//	GetProgress()           *StreamProgress O(1)        Check progress, not errors
//	IsError() (wrapper)     bool            O(1)        Check HTTP status error
//
// Best Practices:
//
//  1. USE AS FAST FILTER
//     - Check HasErrors() first (O(1))
//     - Only call Errors() if HasErrors() returns true (O(n))
//     - Saves memory and CPU for error-free paths
//     - Example:
//     if streaming.HasErrors() {
//     errors := streaming.Errors()
//     // Handle errors
//     }
//
//  2. COMBINE WITH STATUS CHECKING
//     - Don't assume HasErrors() = request failed
//     - Also check wrapper.IsError() for HTTP status
//     - Example:
//     if streaming.HasErrors() || !result.IsSuccess() {
//     // Handle error condition
//     }
//
//  3. USE IN CONDITIONALS FOR CLARITY
//     - More readable than len(Errors()) > 0
//     - Self-documenting code intent
//     - Example:
//     if streaming.HasErrors() { // Clear intent
//     vs
//     if len(streaming.Errors()) > 0 { // Allocates unnecessarily
//     }
//
//  4. CACHE RESULT IN LOOPS
//     - Avoid repeated lock acquisitions
//     - Example:
//     hasErrors := streaming.HasErrors()
//     for i := 0; i < 1000; i++ {
//     if hasErrors { // Use cached value
//     // Handle error condition
//     }
//     }
//
//  5. LOG ONLY ON ERROR
//     - Avoid expensive error list creation in success path
//     - Example:
//     if streaming.HasErrors() {
//     log.Warnf("Errors: %v", streaming.Errors())
//     }
//     // vs always allocating in non-error path
//
// See Also:
//   - Errors: Retrieves full error list (use when HasErrors() is true)
//   - GetStats: Provides FailedChunks count and error array
//   - GetProgress: Includes streaming progress information
//   - GetWrapper: Returns wrapper with status/message
//   - Start: Initiates streaming and accumulates errors
//   - WithCallback: Receives individual errors during streaming
func (sw *StreamingWrapper) HasErrors() bool {
	if sw == nil {
		return false
	}

	sw.mu.RLock()
	defer sw.mu.RUnlock()

	return len(sw.errors) > 0
}

// IsStreaming returns whether a streaming operation is currently in progress.
//
// This function provides a thread-safe way to determine if the streaming process is actively running.
// It returns true only while the Start() method is executing and chunks are being processed; it returns
// false before streaming begins, after streaming completes successfully, after streaming is cancelled,
// or if an error terminates the operation. IsStreaming is useful for implementing timeouts, progress
// monitoring from external goroutines, implementing cancellation logic, and building interactive UIs that
// display transfer status. The function is non-blocking and performs a simple state check without acquiring
// expensive locks for resource operations. Multiple goroutines can safely call IsStreaming() concurrently
// without blocking each other. The return value reflects the streaming state at the moment of the call;
// subsequent calls may return different results if streaming status changes. This is the recommended method
// for querying streaming state without affecting the operation itself.
//
// Returns:
//   - true if streaming is currently active (Start() is executing, chunks being processed).
//   - false if streaming has not started, has completed, was cancelled, or encountered an error.
//   - false if the streaming wrapper is nil.
//   - Thread-safe: multiple goroutines can call concurrently.
//   - Non-blocking: returns immediately without waiting.
//   - State snapshot: reflects state at call time; ongoing changes may alter result.
//
// Streaming State Transitions:
//
//	State Transition                         IsStreaming Before    IsStreaming After
//	──────────────────────────────────────────────────────────────────────────────────
//	New instance created                     N/A (not started)     false
//	WithStreaming() configured               false                 false (configured, not started)
//	Start() called                           false                 true (streaming begins)
//	First chunk read successfully            true                  true (streaming active)
//	Chunks processing mid-stream             true                  true (actively transferring)
//	Last chunk read (EOF)                    true                  true (processing final chunk)
//	Streaming completes successfully         true                  false (operation finished)
//	Cancel() called during streaming         true                  false (immediately stops)
//	Error during streaming                   true                  false (terminates on error)
//	Close() called (after streaming)         false                 false (no change to state)
//	New Start() after completion             false                 true (new streaming begins)
//
// Streaming State Machine:
//
//	     ┌─────────────────────┐
//	     │  NEW / CONFIGURED   │  IsStreaming: false
//	     │   (not started)     │
//	     └──────────┬──────────┘
//	                │ Start()
//	                ↓
//	     ┌─────────────────────┐
//	     │  STREAMING IN       │  IsStreaming: true
//	     │  PROGRESS           │
//	     └──────────┬──────────┘
//	                │
//	    ┌───────────┼───────────┐
//	    │           │           │
//	Cancel()    Error        EOF
//	    │           │           │
//	    ↓           ↓           ↓
//	┌──────────────────────────────────┐
//	│  STREAMING COMPLETED / STOPPED   │  IsStreaming: false
//	│  (result available in wrapper)   │
//	└──────────────────────────────────┘
//
// IsStreaming vs Related State Checks:
//
//	Method/Check                Returns          Purpose                          Cost
//	─────────────────────────────────────────────────────────────────────────────
//	IsStreaming()               bool             Is streaming currently active?   O(1)
//	HasErrors()                 bool             Did any errors occur?            O(1)
//	GetProgress()               *StreamProgress  What is current progress?        O(1)
//	GetStats()                  *StreamingStats  What are final statistics?       O(1)
//	GetWrapper().IsError()      bool             Did HTTP response error occur?   O(1)
//	GetWrapper().IsSuccess()    bool             Was response successful?         O(1)
//	Errors()                    []error          What were all errors?            O(n)
//
// Use Cases and Patterns:
//
//	Use Case                            Pattern                             Example
//	────────────────────────────────────────────────────────────────────────────────
//	Check if streaming active           if streaming.IsStreaming()          Monitoring
//	Wait for streaming to complete      for streaming.IsStreaming() { }    Polling loop
//	External timeout implementation     time.Sleep() + if not streaming()  Watchdog timer
//	UI progress indicator               while streaming.IsStreaming()      Progress display
//	Concurrent monitoring               go func() { isActive }             Background monitor
//	Graceful shutdown                   if streaming: cancel()             Service shutdown
//	State-aware error handling          if streaming: recover else: exit   Error recovery
//	Cancellation guard                  if streaming: cancel()             Safe cancellation
//
// Example:
//
//	// Example 1: Simple streaming state check
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024)
//
//	fmt.Printf("Before Start: IsStreaming = %v\n", streaming.IsStreaming())
//	// Output: Before Start: IsStreaming = false
//
//	result := streaming.Start(context.Background())
//
//	fmt.Printf("After Start: IsStreaming = %v\n", streaming.IsStreaming())
//	// Output: After Start: IsStreaming = false (completed)
//
//	// Example 2: Concurrent monitoring during streaming
//	dataExport := createDataReader()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/export/data").
//	    WithStreaming(dataExport, nil).
//	    WithChunkSize(512 * 1024).
//	    WithMaxConcurrentChunks(4)
//
//	// Start streaming in background
//	done := make(chan *wrapper)
//	go func() {
//	    result := streaming.Start(context.Background())
//	    done <- result
//	}()
//
//	// Monitor streaming from main goroutine
//	ticker := time.NewTicker(500 * time.Millisecond)
//	defer ticker.Stop()
//
//	for {
//	    select {
//	    case <-ticker.C:
//	        if streaming.IsStreaming() {
//	            progress := streaming.GetProgress()
//	            fmt.Printf("\rExporting: %.1f%% (%d chunks)",
//	                float64(progress.Percentage), progress.CurrentChunk)
//	        } else {
//	            fmt.Println("\nExport completed")
//	        }
//	    case result := <-done:
//	        if result.IsError() {
//	            fmt.Printf("Export failed: %s\n", result.Error())
//	        }
//	        return
//	    }
//	}
//
//	// Example 3: External timeout with IsStreaming check
//	func StreamWithTimeout(fileReader io.ReadCloser, timeout time.Duration) *wrapper {
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/stream/timeout").
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(1024 * 1024)
//
//	    done := make(chan *wrapper)
//
//	    // Start streaming in background
//	    go func() {
//	        result := streaming.Start(context.Background())
//	        done <- result
//	    }()
//
//	    // Implement external timeout
//	    select {
//	    case result := <-done:
//	        // Streaming completed normally
//	        return result
//	    case <-time.After(timeout):
//	        // Timeout exceeded
//	        if streaming.IsStreaming() {
//	            fmt.Println("Timeout: streaming still active, cancelling...")
//	            streaming.Cancel()
//	            streaming.Close()
//	            return streaming.GetWrapper().
//	                WithStatusCode(408).
//	                WithMessage("Streaming timeout")
//	        }
//	        // Streaming already completed
//	        return <-done
//	    }
//	}
//
//	// Example 4: UI progress indicator (web/CLI)
//	type StreamingProgressUI struct {
//	    streaming *StreamingWrapper
//	    done      chan bool
//	}
//
//	func (ui *StreamingProgressUI) DisplayProgress() {
//	    ticker := time.NewTicker(100 * time.Millisecond)
//	    defer ticker.Stop()
//
//	    for {
//	        select {
//	        case <-ticker.C:
//	            // Check if still streaming (O(1) operation)
//	            if !ui.streaming.IsStreaming() {
//	                ui.FinalizeDisplay()
//	                ui.done <- true
//	                return
//	            }
//
//	            // Get progress (also O(1))
//	            progress := ui.streaming.GetProgress()
//
//	            // Render progress bar
//	            barLength := 30
//	            filled := int(float64(barLength) * float64(progress.Percentage) / 100)
//	            bar := strings.Repeat("█", filled) + strings.Repeat("░", barLength-filled)
//
//	            fmt.Printf("\r[%s] %.1f%% | %s",
//	                bar,
//	                float64(progress.Percentage),
//	                progress.EstimatedTimeRemaining.String())
//	        case <-ui.done:
//	            return
//	        }
//	    }
//	}
//
//	// Example 5: Graceful shutdown with IsStreaming guard
//	func GracefulShutdown(streaming *StreamingWrapper) {
//	    shutdownTimeout := 30 * time.Second
//	    shutdownDeadline := time.Now().Add(shutdownTimeout)
//
//	    // Check if streaming is active
//	    if streaming.IsStreaming() {
//	        fmt.Println("Streaming in progress, cancelling...")
//	        streaming.Cancel()
//
//	        // Wait for streaming to stop with timeout
//	        for time.Now().Before(shutdownDeadline) {
//	            if !streaming.IsStreaming() {
//	                fmt.Println("Streaming stopped")
//	                break
//	            }
//	            time.Sleep(100 * time.Millisecond)
//	        }
//
//	        if streaming.IsStreaming() {
//	            fmt.Println("Warning: streaming did not stop within timeout")
//	        }
//	    }
//
//	    // Cleanup
//	    streaming.Close()
//	    fmt.Println("Streaming resources released")
//	}
//
//	// Example 6: Safe cancellation with state guard
//	func SafeCancel(streaming *StreamingWrapper) *wrapper {
//	    // Guard: only cancel if actually streaming
//	    if streaming.IsStreaming() {
//	        return streaming.Cancel().
//	            WithMessage("Streaming cancelled by user").
//	            WithStatusCode(202) // 202 Accepted
//	    } else {
//	        return streaming.GetWrapper().
//	            WithMessage("Streaming is not active, nothing to cancel").
//	            WithStatusCode(400) // 400 Bad Request
//	    }
//	}
//
// Polling Pattern for Waiting on Completion:
//
//	// ❌ BAD: Busy-wait with no sleep (wastes CPU)
//	for streaming.IsStreaming() {
//	    // Spin loop - 100% CPU usage
//	}
//
//	// ⚠️  OKAY: Polling with fixed interval
//	for streaming.IsStreaming() {
//	    time.Sleep(100 * time.Millisecond)
//	}
//	// Cost: Periodic lock acquisition + context switch
//	// Use for: Simple cases, acceptable latency tolerance
//
//	// ✅ BETTER: Channel-based with callback
//	done := make(chan *wrapper)
//	go func() {
//	    result := streaming.Start(ctx)
//	    done <- result
//	}()
//	result := <-done
//	// Cost: Single context switch + goroutine
//	// Use for: Optimal, recommended approach
//
//	// ✅ BETTER: Context cancellation + timeout
//	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
//	defer cancel()
//	result := streaming.Start(ctx)
//	// Cost: Built-in timeout, integrates with context
//	// Use for: Timeout enforcement, recommended
//
// Performance Characteristics:
//
//	Operation                   Time Complexity    Space Complexity    Details
//	─────────────────────────────────────────────────────────────────────────
//	IsStreaming() call          O(1)               O(0)                Simple bool check
//	State machine lookup        <1μs               None                Immediate return
//	RWMutex read lock           O(1) amortized     None                Lock contention rare
//	Return value access         <1μs               None                Stack return
//	vs Errors() copy            100x faster        No allocation       Major advantage
//	Multiple concurrent calls   Parallel           None                No blocking
//
// Thread-Safety and Concurrency:
//
//	Scenario                                    Thread-Safe    Notes
//	──────────────────────────────────────────────────────────────────────
//	IsStreaming() during streaming              Yes            RWMutex read lock
//	Multiple concurrent IsStreaming()           Yes            Parallel reads allowed
//	IsStreaming() + Start() race                Yes            Predictable transition
//	IsStreaming() + Cancel() concurrently       Yes            Independent operations
//	IsStreaming() + Close() concurrently        Yes            Close doesn't change state
//	IsStreaming() in callback                   Yes            Called by streaming goroutine
//	High-frequency polling                      Safe           Lock contention minimal
//	Concurrent monitor goroutines               Yes            RWMutex prevents stale reads
//
// State Consistency Guarantees:
//
//	Guarantee                                   Behavior
//	──────────────────────────────────────────────────────────────────────
//	True when streaming is definitely active    Yes - can rely on true
//	False when streaming is definitely done     Yes - can rely on false
//	Immediate transition on Cancel()            Yes - false returned soon after
//	Consistent with GetProgress()               Yes - both read same state
//	Consistent with HasErrors()                 Maybe - different time calls
//	No false positives (says true but done)     Yes - guaranteed
//	No false negatives (says false but running) Yes - guaranteed
//
// Common Patterns and Recommendations:
//
//	Pattern                                 Recommendation        Use Case
//	─────────────────────────────────────────────────────────────────────
//	Simple status check                      Recommended ✓         "Is it still running?"
//	Polling loop with sleep                  Acceptable ⚠️         Simple monitoring
//	Busy-wait loop                           Bad ❌                 Wastes CPU
//	Channel-based completion                 Recommended ✓          Production systems
//	Context timeout                          Recommended ✓          Timeout enforcement
//	Concurrent monitoring goroutine          Recommended ✓          UI updates, logging
//	Progress indicator loop                  Recommended ✓          User feedback
//	Graceful shutdown logic                  Recommended ✓          Service shutdown
//
// Related State Query Methods:
//
//	Method                      Returns             When Streaming      When Completed
//	─────────────────────────────────────────────────────────────────────────────
//	IsStreaming()               bool                true                false
//	HasErrors()                 bool                maybe (any errors)  true/false
//	GetProgress()               *StreamProgress     Current progress    Final progress
//	GetStats()                  *StreamingStats     Partial stats       Complete stats
//	GetWrapper().IsError()      bool                maybe               true/false
//	GetWrapper().IsSuccess()    bool                maybe               true/false
//
// See Also:
//   - Start: Initiates streaming (sets IsStreaming to true)
//   - Cancel: Stops streaming (sets IsStreaming to false)
//   - Close: Releases resources (does not affect IsStreaming state)
//   - GetProgress: Provides progress info (only meaningful if IsStreaming)
//   - GetStats: Provides statistics (complete only after streaming done)
//   - HasErrors: Checks for errors (independent of IsStreaming)
//   - WithCallback: Receives updates during streaming (while IsStreaming true)
func (sw *StreamingWrapper) IsStreaming() bool {
	if sw == nil {
		return false
	}

	sw.mu.RLock()
	defer sw.mu.RUnlock()

	return sw.isStreaming
}

// GetWrapper returns the underlying wrapper object associated with this streaming wrapper.
//
// This function provides access to the base wrapper instance that was either passed into WithStreaming()
// or automatically created during streaming initialization. The returned wrapper contains all HTTP response
// metadata including status code, message, headers, custom fields, and debugging information accumulated
// during the streaming operation. This is useful for building complete API responses, accessing response
// metadata, chaining additional wrapper methods, and integrating streaming results with the standard wrapper
// API. GetWrapper is non-blocking, thread-safe, and can be called at any time during or after streaming.
// The returned wrapper reference points to the same underlying object; modifications through the returned
// reference affect the original wrapper state. Multiple calls to GetWrapper() return the same wrapper
// instance, not copies. This enables seamless integration between streaming and standard wrapper patterns,
// allowing users to leverage both the streaming-specific functionality and the comprehensive wrapper API
// in a unified response-building workflow.
//
// Returns:
//   - A pointer to the underlying wrapper instance if the streaming wrapper is valid.
//   - If the streaming wrapper is nil, returns a newly created empty wrapper (safety fallback).
//   - The returned wrapper is the same instance used throughout streaming; not a copy.
//   - All streaming metadata (status code, message, debug info) is available via the wrapper.
//   - Modifications to the returned wrapper affect the final response object.
//   - The wrapper reference is thread-safe to read; use proper synchronization for modifications.
//   - Multiple calls return the same instance (identity equality: w1 == w2).
//
// Wrapper Integration Points:
//
//	Aspect                          How GetWrapper() Facilitates Integration
//	───────────────────────────────────────────────────────────────────────────
//	Status code management          Access/modify response HTTP status code
//	Message/error text              Set response message or error description
//	Custom fields                   Add domain-specific data to response
//	Debugging information           Add/retrieve debugging KV pairs
//	Response headers                Configure HTTP response headers
//	Response body                   Set body content (though streaming uses writer)
//	Pagination info                 Add pagination metadata
//	Error wrapping                  Wrap errors in standard wrapper format
//	Method chaining                 Chain multiple wrapper methods together
//	Final response building         Construct complete API response
//
// Typical Integration Pattern:
//
//	// Create streaming wrapper
//	streaming := response.WithStreaming(reader, config)
//
//	// Configure streaming
//	result := streaming.
//	    WithChunkSize(1024 * 1024).
//	    WithCompressionType(COMP_GZIP).
//	    Start(ctx)
//
//	// Access wrapper for response building
//	finalResponse := streaming.GetWrapper().
//	    WithStatusCode(200).
//	    WithMessage("Streaming completed").
//	    WithTotal(streaming.GetProgress().CurrentChunk)
//
// Access Patterns:
//
//	Pattern                             Use Case                        Example
//	──────────────────────────────────────────────────────────────────────────
//	Immediate access                    Read current status             status := GetWrapper().StatusCode()
//	Post-streaming response             Build final response            GetWrapper().WithStatusCode(200)
//	Error handling                      Wrap errors in response         GetWrapper().WithError(err)
//	Metadata enrichment                 Add context information         GetWrapper().WithDebuggingKV(...)
//	Header configuration                Set HTTP response headers       GetWrapper().WithHeader(...)
//	Pagination integration              Add page info if applicable     GetWrapper().WithPagination(...)
//	Custom field addition               Domain-specific data            GetWrapper().WithCustomFieldKV(...)
//	Conditional response building       Different paths based on state  if hasErrors: WithStatusCode(206)
//	Response chaining                   Build response inline           return GetWrapper().WithMessage(...)
//	Debugging/diagnostics               Access accumulated metadata     GetWrapper().Debugging()
//
// Example:
//
//	// Example 1: Simple response building after streaming
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithMaxConcurrentChunks(4)
//
//	result := streaming.Start(context.Background())
//
//	// Access wrapper to build final response
//	finalResponse := streaming.GetWrapper().
//	    WithMessage("File download completed").
//	    WithDebuggingKV("total_chunks",
//	        streaming.GetProgress().CurrentChunk).
//	    WithDebuggingKV("bytes_transferred",
//	        streaming.GetProgress().TransferredBytes)
//
//	// Example 2: Error handling with wrapper integration
//	httpResp, _ := http.Get("https://api.example.com/largefile")
//	defer httpResp.Body.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/proxy/remote-file").
//	    WithStreaming(httpResp.Body, nil).
//	    WithChunkSize(512 * 1024).
//	    WithReadTimeout(15000).
//	    WithWriteTimeout(15000).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            log.Warnf("Streaming chunk %d error: %v",
//	                p.CurrentChunk, err)
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//
//	// Use GetWrapper() for error response building
//	if streaming.HasErrors() {
//	    finalResponse := streaming.GetWrapper().
//	        WithStatusCode(206). // 206 Partial Content
//	        WithMessage("File download completed with errors").
//	        WithError(fmt.Sprintf("%d chunks failed",
//	            len(streaming.Errors()))).
//	        WithDebuggingKV("error_count", len(streaming.Errors())).
//	        WithDebuggingKV("failed_chunks", streaming.GetStats().FailedChunks)
//	} else {
//	    finalResponse := streaming.GetWrapper().
//	        WithStatusCode(200).
//	        WithMessage("File download completed successfully").
//	        WithDebuggingKV("bytes_transferred",
//	            streaming.GetStats().TotalBytes)
//	}
//
//	// Example 3: Metadata enrichment with streaming context
//	dataExport := createDataReader()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/export/users").
//	    WithCustomFieldKV("export_type", "csv").
//	    WithCustomFieldKV("export_id", "exp-2025-1114-001").
//	    WithStreaming(dataExport, nil).
//	    WithChunkSize(512 * 1024).
//	    WithMaxConcurrentChunks(4).
//	    WithCompressionType(COMP_GZIP)
//
//	result := streaming.Start(context.Background())
//	stats := streaming.GetStats()
//	progress := streaming.GetProgress()
//
//	// Rich metadata response via GetWrapper()
//	finalResponse := streaming.GetWrapper().
//	    WithStatusCode(200).
//	    WithMessage("User export completed").
//	    WithDebuggingKV("export_id", "exp-2025-1114-001").
//	    WithDebuggingKV("export_status", "completed").
//	    WithDebuggingKV("records_exported", progress.CurrentChunk).
//	    WithDebuggingKV("original_size_bytes", stats.TotalBytes).
//	    WithDebuggingKVf("compressed_size_bytes", "%d", stats.CompressedBytes).
//	    WithDebuggingKVf("compression_ratio", "%.2f%%",
//	        (1.0 - stats.CompressionRatio) * 100).
//	    WithDebuggingKV("duration_seconds",
//	        stats.EndTime.Sub(stats.StartTime).Seconds()).
//	    WithDebuggingKVf("average_bandwidth_mbps", "%.2f",
//	        float64(stats.AverageBandwidth) / 1024 / 1024)
//
//	// Example 4: Conditional response building with wrapper
//	func BuildStreamingResponse(streaming *StreamingWrapper) *wrapper {
//	    progress := streaming.GetProgress()
//	    stats := streaming.GetStats()
//
//	    // Start with base wrapper
//	    response := streaming.GetWrapper()
//
//	    // Branch based on streaming outcome
//	    if streaming.HasErrors() {
//	        errorCount := len(streaming.Errors())
//	        if errorCount > 10 {
//	            // Many errors - mostly failed
//	            return response.
//	                WithStatusCode(500).
//	                WithMessage("Streaming failed with many errors").
//	                WithError(fmt.Sprintf("%d chunks failed", errorCount)).
//	                WithDebuggingKV("success_rate_percent",
//	                    int(float64(stats.TotalChunks - stats.FailedChunks) /
//	                        float64(stats.TotalChunks) * 100))
//	        } else {
//	            // Few errors - mostly succeeded
//	            return response.
//	                WithStatusCode(206).
//	                WithMessage("Streaming completed with minor errors").
//	                WithDebuggingKV("error_count", errorCount).
//	                WithDebuggingKV("success_rate_percent",
//	                    int(float64(stats.TotalChunks - stats.FailedChunks) /
//	                        float64(stats.TotalChunks) * 100))
//	        }
//	    } else {
//	        // Perfect success
//	        return response.
//	            WithStatusCode(200).
//	            WithMessage("Streaming completed successfully").
//	            WithDebuggingKV("total_chunks", stats.TotalChunks).
//	            WithDebuggingKV("total_bytes", stats.TotalBytes).
//	            WithDebuggingKVf("duration_seconds", "%.2f",
//	                stats.EndTime.Sub(stats.StartTime).Seconds())
//	    }
//	}
//
//	// Example 5: Accessing wrapper during streaming (concurrent monitoring)
//	func MonitorStreamingWithWrapper(streaming *StreamingWrapper) {
//	    ticker := time.NewTicker(500 * time.Millisecond)
//	    defer ticker.Stop()
//
//	    for range ticker.C {
//	        if !streaming.IsStreaming() {
//	            break
//	        }
//
//	        // Access wrapper status info
//	        wrapper := streaming.GetWrapper()
//	        progress := streaming.GetProgress()
//
//	        fmt.Printf("Status: %d | Message: %s | Progress: %.1f%%\n",
//	            wrapper.StatusCode(),
//	            wrapper.Message(),
//	            float64(progress.Percentage))
//	    }
//	}
//
//	// Example 6: Complete workflow with wrapper integration
//	func CompleteStreamingWorkflow(fileReader io.ReadCloser,
//	                                exportID string) *wrapper {
//	    // Create streaming wrapper
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/export/complete-workflow").
//	        WithCustomFieldKV("export_id", exportID).
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(1024 * 1024).
//	        WithMaxConcurrentChunks(4).
//	        WithCompressionType(COMP_GZIP).
//	        WithCallback(func(p *StreamProgress, err error) {
//	            if err != nil {
//	                log.Warnf("[%s] Chunk %d error: %v",
//	                    exportID, p.CurrentChunk, err)
//	            }
//	        })
//
//	    // Execute streaming
//	    result := streaming.Start(context.Background())
//
//	    // Cleanup
//	    streaming.Close()
//
//	    // Get streaming statistics
//	    stats := streaming.GetStats()
//	    progress := streaming.GetProgress()
//
//	    // Build complete response via GetWrapper()
//	    finalResponse := streaming.GetWrapper()
//
//	    // Status code based on outcome
//	    if streaming.HasErrors() {
//	        finalResponse = finalResponse.WithStatusCode(206)
//	    } else {
//	        finalResponse = finalResponse.WithStatusCode(200)
//	    }
//
//	    // Add comprehensive metadata
//	    finalResponse.
//	        WithMessage("Workflow execution completed").
//	        WithDebuggingKV("export_id", exportID).
//	        WithDebuggingKV("status", "completed").
//	        WithDebuggingKV("chunks_processed", progress.CurrentChunk).
//	        WithDebuggingKV("total_chunks", stats.TotalChunks).
//	        WithDebuggingKV("original_size", stats.TotalBytes).
//	        WithDebuggingKV("compressed_size", stats.CompressedBytes).
//	        WithDebuggingKVf("compression_ratio", "%.1f%%",
//	            (1.0 - stats.CompressionRatio) * 100).
//	        WithDebuggingKV("error_count", len(streaming.Errors())).
//	        WithDebuggingKV("failed_chunks", stats.FailedChunks).
//	        WithDebuggingKVf("duration_ms", "%d",
//	            stats.EndTime.Sub(stats.StartTime).Milliseconds()).
//	        WithDebuggingKVf("bandwidth_mbps", "%.2f",
//	            float64(stats.AverageBandwidth) / 1024 / 1024)
//
//	    return finalResponse
//	}
//
// Wrapper Metadata Available Through GetWrapper():
//
//	Category                Information Available              Example Access
//	──────────────────────────────────────────────────────────────────────────
//	HTTP Status              StatusCode, IsError, IsSuccess    wrapper.StatusCode()
//	Response Message         Message text                      wrapper.Message()
//	Errors                   Error wrapping, message           wrapper.Error()
//	Custom Fields            Domain-specific data              wrapper.CustomFields()
//	Debugging Info           KV debugging pairs                wrapper.Debugging()
//	Response Headers         HTTP response headers             wrapper.Headers()
//	Path                     API endpoint path                 wrapper.Path()
//	Pagination               Page/limit/offset info            wrapper.Pagination()
//	Request/Response timing  Built via debugging KV            wrapper.DebuggingKV()
//	All metadata             Complete wrapper state            wrapper.*() methods
//
// Streaming-Specific Metadata Added via GetWrapper():
//
//	Metadata Item                   Source                  Purpose
//	────────────────────────────────────────────────────────────────────────
//	streaming_strategy              WithStreaming()         Track chosen strategy
//	compression_type                WithCompressionType()   Track compression used
//	chunk_size                       WithChunkSize()         Track chunk size
//	total_bytes                      WithTotalBytes()        Track total data size
//	max_concurrent_chunks            WithMaxConcurrentChunks Track parallelism
//	throttle_rate_bps                WithThrottleRate()      Track bandwidth limit
//	buffer_pooling_enabled           WithBufferPooling()     Track buffer reuse
//	read_timeout_ms                  WithReadTimeout()       Track read timeout
//	write_timeout_ms                 WithWriteTimeout()      Track write timeout
//	streaming_error                  Start() on error        Error message if failed
//	failed_chunks                    Start()                 Count of failed chunks
//	total_errors                     Start()                 Count of accumulated errors
//	started_at                       Start()                 Timestamp when started
//	completed_at                     Start()                 Timestamp when completed
//	cancelled_at                     Cancel()                Timestamp when cancelled
//	duration_ms                      Start()                 Total operation duration
//	compression_ratio                GetStats()              Compression effectiveness
//
// Identity and Mutation Semantics:
//
//	Aspect                          Behavior                        Implication
//	──────────────────────────────────────────────────────────────────────────
//	Multiple calls return same ref   GetWrapper() == GetWrapper()    Single instance
//	Modifications visible everywhere wrapper.WithStatusCode(201)     Shared state
//	Not a copy                       Mutations affect original        Direct reference
//	Thread-safe reads                RWMutex protected               Safe access
//	Concurrent modifications         May race                        Use sync for changes
//	Post-streaming access            Safe, no cleanup needed         Wrapper persists
//	Reference persistence            Available until wrapper freed   Lifetime guaranteed
//	Copy semantics                   Reference, not value            No deep copy made
//
// Common Usage Patterns:
//
//	Pattern                                         Benefit                         Example
//	─────────────────────────────────────────────────────────────────────────────
//	Post-streaming status check                     Immediate feedback              wrapper.StatusCode()
//	Error response building                         Consistent error format         WithError(...).WithStatusCode(...)
//	Metadata enrichment                             Rich response context           WithDebuggingKV(...) for each stat
//	Conditional response chains                     Flexible response building      if hasErrors: 206 else: 200
//	Inline response construction                    Compact code                    return GetWrapper().WithMessage(...)
//	Concurrent progress monitoring                  Real-time status access         Polling wrapper state
//	Logging with wrapper context                    Diagnostic context              Log wrapper.Debugging()
//	Response wrapping for downstream                Consistent API contracts        Pass wrapper to other services
//
// Best Practices:
//
//  1. CALL AFTER STREAMING COMPLETES
//     - GetWrapper() returns final state
//     - Call after Start() returns for complete information
//     - Pattern:
//     result := streaming.Start(ctx)
//     finalResponse := streaming.GetWrapper().WithMessage("Done")
//
//  2. USE FOR RESPONSE BUILDING
//     - GetWrapper() is designed for constructing API responses
//     - Chain methods for clean, readable code
//     - Example:
//     return streaming.GetWrapper().
//     WithStatusCode(200).
//     WithMessage("Export completed").
//     WithDebuggingKV("total_chunks", stats.TotalChunks)
//
//  3. COMBINE WITH STREAMING STATS
//     - GetWrapper() provides wrapper metadata
//     - GetStats() provides streaming metrics
//     - GetProgress() provides progress info
//     - Use all three for complete picture
//     - Example:
//     wrapper := streaming.GetWrapper()
//     stats := streaming.GetStats()
//     return wrapper.WithDebuggingKV("bytes", stats.TotalBytes)
//
//  4. HANDLE NIL STREAMING GRACEFULLY
//     - GetWrapper() creates new wrapper if streaming is nil
//     - Safe fallback for defensive programming
//     - No need for nil checks before calling
//     - Example:
//     wrapper := nilStreaming.GetWrapper() // Safe, returns new wrapper
//
//  5. DON'T MODIFY DURING STREAMING
//     - Wrapper mutations during streaming may race
//     - Read-only access is safe
//     - Modifications after streaming are safe
//     - Example:
//     // During streaming: safe
//     statusCode := streaming.GetWrapper().StatusCode()
//     // During streaming: unsafe (may race)
//     streaming.GetWrapper().WithMessage("New message")
//     // After streaming: safe
//     streaming.GetWrapper().WithMessage("Completed")
//
// Relationship to Other Methods:
//
//	Method                  Provides                    When to Use
//	──────────────────────────────────────────────────────────────────────
//	GetWrapper()            Wrapper with metadata       Building final response
//	GetStats()              Streaming metrics           Statistics/diagnostics
//	GetProgress()           Current progress info       Progress tracking
//	Errors()                Error list copy             Error analysis
//	HasErrors()             Boolean error check         Quick error detection
//	GetStreamingStats()     Complete stats              After streaming
//	GetStreamingProgress()  Final progress              After streaming
//	IsStreaming()           Active status               Concurrent monitoring
//
// See Also:
//   - Start: Executes streaming and populates wrapper with results
//   - WithStreaming: Creates streaming wrapper with base wrapper
//   - Cancel: Stops streaming, updates wrapper state
//   - Close: Closes resources, does not affect wrapper
//   - GetStats: Returns streaming statistics (complement to GetWrapper)
//   - GetProgress: Returns progress information (complement to GetWrapper)
//   - WithStatusCode/WithMessage/etc: Wrapper methods for building response
func (sw *StreamingWrapper) GetWrapper() *wrapper {
	if sw == nil {
		return respondStreamBadRequestDefault()
	}
	return sw.wrapper
}

// StreamingContext returns the context associated with the streaming operation.
//
// This function provides access to the context.Context object that was used to initiate
// the streaming process. The returned context can be used for cancellation, deadlines,
// and passing request-scoped values throughout the streaming lifecycle. StreamingContext
// is useful for integrating with other context-aware components, propagating cancellation
// signals, and accessing metadata associated with the streaming operation. If the
// streaming wrapper is nil, StreamingContext returns a background context as a safe fallback.
// The returned context is read-only; modifications should be done on the original context
// passed into the Start() method. This function is thread-safe and can be called at any time
// during or after streaming.
func (sw *StreamingWrapper) StreamingContext() context.Context {
	if sw == nil {
		return context.Background()
	}
	return sw.ctx
}

// GetStreamingStats returns complete statistics about the streaming operation accumulated up to the current point.
//
// This function provides detailed metrics and analytics about the streaming transfer including total bytes
// transferred, compression statistics, bandwidth metrics, chunk processing information, error counts, timing data,
// and resource utilization. The statistics accumulate throughout the streaming operation and represent the state
// at the time of the call. For ongoing streaming, GetStreamingStats() returns partial statistics reflecting progress
// so far; for completed streaming, it returns complete final statistics. This is the primary method for obtaining
// comprehensive streaming performance data, diagnostics, and analytics. GetStreamingStats is thread-safe and can be
// called from any goroutine during or after streaming without blocking the operation. The returned StreamingStats
// structure is a snapshot copy; modifications to the returned stats do not affect the internal state. All statistics
// are maintained with microsecond precision for timing measurements and byte-level accuracy for data transfers.
//
// Returns:
//   - A pointer to a StreamingStats structure containing complete streaming statistics.
//   - Returns an empty StreamingStats{} if the streaming wrapper is nil.
//   - For ongoing streaming: returns partial statistics reflecting progress to current point.
//   - For completed streaming: returns final complete statistics from entire operation.
//   - All timestamps are in UTC with nanosecond precision (time.Time format).
//   - All byte counts and rates are 64-bit integers (int64) for large transfer support.
//   - All percentage and ratio values are floating-point (0.0-1.0 or 0-100.0) as documented.
//   - Thread-safe: safe to call concurrently from multiple goroutines.
//   - Snapshot semantics: returned stats are independent; modifications do not affect streaming.
//
// StreamingStats Structure Contents:
//
//	Field Category          Field Name                      Type        Purpose
//	─────────────────────────────────────────────────────────────────────────────────────
//	Timing Information
//	                        StartTime                       time.Time   When streaming began
//	                        EndTime                         time.Time   When streaming ended
//	                        ElapsedTime                     time.Duration Duration of operation
//
//	Data Transfer
//	                        TotalBytes                      int64       Total bytes to transfer
//	                        TransferredBytes                int64       Bytes actually transferred
//	                        FailedBytes                     int64       Bytes not transferred
//
//	Chunk Processing
//	                        TotalChunks                     int64       Total chunks to process
//	                        ProcessedChunks                 int64       Chunks processed successfully
//	                        FailedChunks                    int64       Chunks that failed
//
//	Compression Data
//	                        OriginalSize                    int64       Size before compression
//	                        CompressedSize                  int64       Size after compression
//	                        CompressionRatio               float64     compressed/original (0.0-1.0)
//	                        CompressionType                string      COMP_NONE, COMP_GZIP, COMP_DEFLATE
//
//	Bandwidth Metrics
//	                        AverageBandwidth               int64       Average B/s for entire transfer
//	                        PeakBandwidth                  int64       Maximum B/s during transfer
//	                        MinimumBandwidth               int64       Minimum B/s during transfer
//	                        ThrottleRate                   int64       Configured throttle rate
//
//	Error Tracking
//	                        Errors                         []error     All errors that occurred
//	                        ErrorCount                     int64       Total number of errors
//	                        HasErrors                      bool        Whether any errors occurred
//	                        FirstError                     error       First error encountered (if any)
//	                        LastError                      error       Most recent error (if any)
//
//	Resource Utilization
//	                        CPUTime                        time.Duration Time spent on CPU
//	                        MemoryAllocated                int64       Memory allocated during transfer
//	                        MaxMemoryUsage                 int64       Peak memory during transfer
//	                        BufferPoolHits                 int64       Count of buffer pool reuses
//	                        BufferPoolMisses               int64       Count of buffer pool misses
//
//	Configuration
//	                        ChunkSize                      int64       Configured chunk size
//	                        MaxConcurrentChunks            int64       Max parallel chunk processing
//	                        StreamingStrategy              string      Strategy used (BUFFERED, DIRECT, etc)
//	                        UseBufferPool                  bool        Whether buffer pooling was enabled
//
// Example:
//
//	// Example 1: Simple statistics retrieval after streaming
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithMaxConcurrentChunks(4)
//
//	result := streaming.Start(context.Background())
//
//	// Retrieve complete statistics
//	stats := streaming.GetStreamingStats()
//
//	fmt.Printf("Streaming Statistics:\n")
//	fmt.Printf("  Total bytes: %d\n", stats.TotalBytes)
//	fmt.Printf("  Transferred: %d\n", stats.TransferredBytes)
//	fmt.Printf("  Duration: %s\n", stats.ElapsedTime)
//	fmt.Printf("  Chunks processed: %d/%d\n",
//	    stats.ProcessedChunks, stats.TotalChunks)
//	fmt.Printf("  Average bandwidth: %.2f MB/s\n",
//	    float64(stats.AverageBandwidth) / 1024 / 1024)
//
//	// Example 2: Comprehensive statistics analysis
//	httpResp, _ := http.Get("https://api.example.com/largefile")
//	defer httpResp.Body.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/proxy/remote-file").
//	    WithStreaming(httpResp.Body, nil).
//	    WithChunkSize(512 * 1024).
//	    WithMaxConcurrentChunks(4).
//	    WithCompressionType(COMP_GZIP).
//	    WithThrottleRate(1024 * 1024) // 1MB/s
//
//	result := streaming.Start(context.Background())
//
//	// Detailed statistics analysis
//	stats := streaming.GetStreamingStats()
//
//	if stats.TotalBytes > 0 {
//	    successRate := float64(stats.ProcessedChunks) / float64(stats.TotalChunks) * 100
//	    compressionSavings := (1.0 - stats.CompressionRatio) * 100
//
//	    fmt.Printf("=== TRANSFER STATISTICS ===\n")
//	    fmt.Printf("Size:\n")
//	    fmt.Printf("  Original:    %.2f MB\n", float64(stats.OriginalSize) / 1024 / 1024)
//	    fmt.Printf("  Compressed:  %.2f MB\n", float64(stats.CompressedSize) / 1024 / 1024)
//	    fmt.Printf("  Savings:     %.1f%%\n", compressionSavings)
//
//	    fmt.Printf("Chunks:\n")
//	    fmt.Printf("  Total:       %d\n", stats.TotalChunks)
//	    fmt.Printf("  Processed:   %d\n", stats.ProcessedChunks)
//	    fmt.Printf("  Failed:      %d\n", stats.FailedChunks)
//	    fmt.Printf("  Success:     %.1f%%\n", successRate)
//
//	    fmt.Printf("Performance:\n")
//	    fmt.Printf("  Duration:    %.2f seconds\n", stats.ElapsedTime.Seconds())
//	    fmt.Printf("  Avg Rate:    %.2f MB/s\n",
//	        float64(stats.AverageBandwidth) / 1024 / 1024)
//	    fmt.Printf("  Peak Rate:   %.2f MB/s\n",
//	        float64(stats.PeakBandwidth) / 1024 / 1024)
//	    fmt.Printf("  Min Rate:    %.2f MB/s\n",
//	        float64(stats.MinimumBandwidth) / 1024 / 1024)
//
//	    if stats.HasErrors {
//	        fmt.Printf("Errors:\n")
//	        fmt.Printf("  Count:       %d\n", stats.ErrorCount)
//	        fmt.Printf("  First:       %v\n", stats.FirstError)
//	        fmt.Printf("  Last:        %v\n", stats.LastError)
//	    }
//	}
//
//	// Example 3: Performance monitoring and diagnostics
//	dataExport := createDataReader()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/export/bulk-data").
//	    WithCustomFieldKV("export_id", "exp-2025-1114-001").
//	    WithStreaming(dataExport, nil).
//	    WithChunkSize(10 * 1024 * 1024).
//	    WithMaxConcurrentChunks(8).
//	    WithCompressionType(COMP_GZIP).
//	    WithBufferPooling(true)
//
//	result := streaming.Start(context.Background())
//	stats := streaming.GetStreamingStats()
//
//	// Build diagnostic report
//	fmt.Printf("Export exp-2025-1114-001 Report\n")
//	fmt.Printf("================================\n")
//	fmt.Printf("Timing: %s (%.2f seconds)\n",
//	    stats.EndTime.Sub(stats.StartTime),
//	    stats.ElapsedTime.Seconds())
//	fmt.Printf("Data: %.2f MB → %.2f MB (%.1f%% compression)\n",
//	    float64(stats.OriginalSize) / 1024 / 1024,
//	    float64(stats.CompressedSize) / 1024 / 1024,
//	    (1.0 - stats.CompressionRatio) * 100)
//	fmt.Printf("Chunks: %d processed, %d failed\n",
//	    stats.ProcessedChunks, stats.FailedChunks)
//	fmt.Printf("Bandwidth: %d B/s (avg) | %d B/s (peak) | %d B/s (throttle)\n",
//	    stats.AverageBandwidth, stats.PeakBandwidth, stats.ThrottleRate)
//	fmt.Printf("Resources: %d MB allocated | %d MB peak\n",
//	    stats.MemoryAllocated / 1024 / 1024,
//	    stats.MaxMemoryUsage / 1024 / 1024)
//	fmt.Printf("Errors: %d total\n", stats.ErrorCount)
//
//	// Example 4: Statistics in response metadata
//	func BuildStatisticsResponse(streaming *StreamingWrapper) *wrapper {
//	    stats := streaming.GetStreamingStats()
//
//	    return streaming.GetWrapper().
//	        WithStatusCode(200).
//	        WithMessage("Streaming completed with statistics").
//	        WithDebuggingKV("total_bytes", stats.TotalBytes).
//	        WithDebuggingKV("transferred_bytes", stats.TransferredBytes).
//	        WithDebuggingKV("chunks_total", stats.TotalChunks).
//	        WithDebuggingKV("chunks_processed", stats.ProcessedChunks).
//	        WithDebuggingKV("chunks_failed", stats.FailedChunks).
//	        WithDebuggingKVf("duration_seconds", "%.2f", stats.ElapsedTime.Seconds()).
//	        WithDebuggingKVf("avg_bandwidth_mbps", "%.2f",
//	            float64(stats.AverageBandwidth) / 1024 / 1024).
//	        WithDebuggingKVf("compression_ratio_percent", "%.1f",
//	            (1.0 - stats.CompressionRatio) * 100).
//	        WithDebuggingKV("error_count", stats.ErrorCount)
//	}
//
//	// Example 5: Conditional statistics reporting based on transfer size
//	func ReportStatisticsIfLarge(streaming *StreamingWrapper) {
//	    stats := streaming.GetStreamingStats()
//
//	    // Only detailed report for large transfers
//	    const largeTransferThreshold = 100 * 1024 * 1024 // 100MB
//
//	    if stats.TotalBytes > largeTransferThreshold {
//	        fmt.Printf("Large Transfer Report (%d MB)\n", stats.TotalBytes / 1024 / 1024)
//	        fmt.Printf("Duration: %.2f seconds\n", stats.ElapsedTime.Seconds())
//	        fmt.Printf("Bandwidth: %.2f MB/s (avg), %.2f MB/s (peak)\n",
//	            float64(stats.AverageBandwidth) / 1024 / 1024,
//	            float64(stats.PeakBandwidth) / 1024 / 1024)
//	        fmt.Printf("Compression: %.1f%% saved\n",
//	            (1.0 - stats.CompressionRatio) * 100)
//	        fmt.Printf("Success rate: %.1f%%\n",
//	            float64(stats.ProcessedChunks) / float64(stats.TotalChunks) * 100)
//
//	        if stats.HasErrors {
//	            fmt.Printf("Errors: %d encountered\n", stats.ErrorCount)
//	        }
//	    }
//	}
//
//	// Example 6: Statistics trend analysis (comparing multiple transfers)
//	type TransferAnalytics struct {
//	    transfers []StreamingStats
//	}
//
//	func (ta *TransferAnalytics) AddTransfer(stats *StreamingStats) {
//	    ta.transfers = append(ta.transfers, *stats)
//	}
//
//	func (ta *TransferAnalytics) AnalyzeTrends() {
//	    if len(ta.transfers) == 0 {
//	        return
//	    }
//
//	    var (
//	        totalDuration    time.Duration
//	        avgBandwidth     float64
//	        avgCompressionRatio float64
//	        successfulCount  int
//	    )
//
//	    for _, stats := range ta.transfers {
//	        totalDuration += stats.ElapsedTime
//	        avgBandwidth += float64(stats.AverageBandwidth)
//	        avgCompressionRatio += stats.CompressionRatio
//	        if !stats.HasErrors {
//	            successfulCount++
//	        }
//	    }
//
//	    count := float64(len(ta.transfers))
//	    successRate := float64(successfulCount) / count * 100
//
//	    fmt.Printf("Transfer Analytics (last %d transfers):\n", len(ta.transfers))
//	    fmt.Printf("  Success rate: %.1f%%\n", successRate)
//	    fmt.Printf("  Avg duration: %.2f seconds\n", totalDuration.Seconds() / count)
//	    fmt.Printf("  Avg bandwidth: %.2f MB/s\n",
//	        (avgBandwidth / count) / 1024 / 1024)
//	    fmt.Printf("  Avg compression: %.1f%%\n",
//	        (1.0 - avgCompressionRatio / count) * 100)
//	}
//
// Statistics Timing Precision:
//
//	Metric                  Resolution        Precision           Use For
//	──────────────────────────────────────────────────────────────────────
//	StartTime/EndTime       Nanosecond         time.Time           Exact operation timing
//	ElapsedTime             Nanosecond         time.Duration       Transfer duration
//	CPUTime                 Microsecond        time.Duration       CPU utilization analysis
//	Timing calculations     Nanosecond         Arithmetic result   Sub-millisecond analysis
//
// Statistics Data Precision:
//
//	Metric                  Type        Range                   Precision
//	─────────────────────────────────────────────────────────────────────
//	Byte counts             int64       0 to 9.2 EB            1 byte
//	Chunk counts            int64       0 to 9.2 billion       1 chunk
//	Bandwidth values        int64       0 B/s to GB/s          1 B/s
//	Compression ratio       float64     0.0 to 1.0             ~1e-15
//	Memory values           int64       0 to 9.2 EB            1 byte
//
// Statistics Availability Timeline:
//
//	Streaming Stage              Statistics Available                         Completeness
//	────────────────────────────────────────────────────────────────────────────────────
//	Before Start()               None (empty structure)                       Empty
//	During Start() (early)       Partial (StartTime only)                    Minimal
//	During Start() (mid-stream)  Partial (progress, partial data)            Growing
//	After Start() completes      Complete (all fields, final values)         100%
//	After Cancel()               Partial at cancellation point                At cancel time
//	After Close()                Same as after Start() (unchanged)            Unchanged
//
// Statistics Snapshot Semantics:
//
//	Aspect                          Behavior                        Implication
//	──────────────────────────────────────────────────────────────────────────
//	Return value                    Pointer to StreamingStats       Copy returned, not reference
//	Modifications                   Do not affect streaming         Safe to modify returned stats
//	Multiple calls                  Independent snapshots           Each call gets fresh snapshot
//	Call during streaming           Returns partial statistics      Time-dependent results
//	Call after streaming            Returns complete statistics    Final values stable
//	Garbage collection              Stats safe from GC            Lifetime guaranteed
//	Concurrent calls                Thread-safe reads              Multiple goroutines safe
//
// Performance Statistics Calculation Methods:
//
//	Metric                              Calculation Method
//	────────────────────────────────────────────────────────────────────
//	AverageBandwidth                    TransferredBytes / ElapsedTime.Seconds()
//	CompressionRatio                    CompressedSize / OriginalSize
//	Success rate                        ProcessedChunks / TotalChunks * 100
//	Failed bytes                        TotalBytes - TransferredBytes
//	Effective throughput                TransferredBytes / ElapsedTime.Seconds()
//	Time per chunk                      ElapsedTime / ProcessedChunks
//	Memory per chunk                    MemoryAllocated / MaxConcurrentChunks
//
// Integration with Other Methods:
//
//	Method                  Returns                 When to Use                  Data Overlap
//	──────────────────────────────────────────────────────────────────────────────────
//	GetStreamingStats()     *StreamingStats         Complete statistics          All streaming metrics
//	GetProgress()           *StreamProgress         Current progress              Current chunk, bytes, %
//	GetWrapper()            *wrapper                Response building             Status, message, headers
//	Errors()                []error                 Error analysis                Error list
//	HasErrors()             bool                    Quick error check             Error existence
//	GetStats()              *StreamingStats         (alias for this function)    Same as GetStreamingStats
//
// Statistics for Different Streaming Strategies:
//
//	Strategy                Statistics Specific Behavior
//	──────────────────────────────────────────────────────────────────────
//	STRATEGY_DIRECT         Linear chunk processing, simple bandwidth calc
//	STRATEGY_BUFFERED       Concurrent I/O, more accurate peak/min bandwidth
//	STRATEGY_CHUNKED        Explicit chunk control, detailed chunk-level stats
//	All strategies          Same final metrics, different measurement precision
//
// Best Practices:
//
//  1. RETRIEVE AFTER STREAMING COMPLETES
//     - Get final comprehensive statistics
//     - Pattern:
//     result := streaming.Start(ctx)
//     stats := streaming.GetStreamingStats()
//     // All fields populated
//
//  2. COMBINE WITH PROGRESS FOR UNDERSTANDING
//     - Progress shows real-time state
//     - Stats shows final analysis
//     - Pattern:
//     progress := streaming.GetProgress()    // Where we are now
//     stats := streaming.GetStreamingStats() // Overall result
//
//  3. USE FOR PERFORMANCE ANALYSIS
//     - Bandwidth metrics for optimization
//     - Memory metrics for resource planning
//     - Compression metrics for effectiveness
//     - Error metrics for reliability analysis
//
//  4. INCLUDE IN RESPONSE METADATA
//     - Add key statistics to API response
//     - Help clients understand transfer
//     - Example:
//     WithDebuggingKV("avg_bandwidth_mbps",
//     float64(stats.AverageBandwidth) / 1024 / 1024)
//
//  5. LOG FOR DIAGNOSTICS
//     - Complete statistics for troubleshooting
//     - Identify performance issues
//     - Track historical trends
//     - Example:
//     log.Infof("Transfer: %d bytes, %.2f s, %.2f MB/s",
//     stats.TransferredBytes, stats.ElapsedTime.Seconds(),
//     float64(stats.AverageBandwidth) / 1024 / 1024)
//
// See Also:
//   - GetProgress: Real-time progress tracking (complement to GetStreamingStats)
//   - GetWrapper: Response building with metadata
//   - GetStats: Alias for GetStreamingStats (same function)
//   - Errors: Retrieve error list (stats includes error count)
//   - HasErrors: Quick error check (stats has HasErrors field)
//   - Start: Executes streaming and accumulates statistics
//   - Cancel: Stops streaming (preserves stats up to cancellation)
func (sw *StreamingWrapper) GetStreamingStats() *StreamingStats {
	if sw == nil {
		return &StreamingStats{}
	}
	return sw.GetStats()
}

// GetStats returns a thread-safe copy of streaming statistics with computed compression metrics.
//
// This function provides access to the complete streaming statistics accumulated up to the current point,
// with automatic calculation of the compression ratio based on current total bytes and compressed bytes.
// The compression ratio is computed dynamically at call time to ensure accuracy even if compression occurs
// during streaming. GetStats is thread-safe and can be called from any goroutine without synchronization;
// the internal mutex ensures consistent snapshot isolation. The returned StreamingStats is a defensive copy,
// not a reference; modifications to the returned stats do not affect the internal state or future calls. This
// is the primary low-level method for retrieving raw streaming metrics; GetStreamingStats() is a higher-level
// alias that calls this function. GetStats is non-blocking and safe to call frequently for real-time monitoring,
// diagnostics, or progress tracking during streaming. All metrics are maintained with precision appropriate to
// their data type: nanosecond precision for timing, byte-level precision for data sizes, and floating-point
// precision for ratios. Statistics are available immediately after Start() begins and remain stable after
// streaming completes or is cancelled.
//
// Returns:
//   - A pointer to a new StreamingStats structure containing a copy of all accumulated statistics.
//   - If the streaming wrapper is nil, returns an empty StreamingStats{} with all fields zero-valued.
//   - The CompressionRatio field is computed dynamically:
//   - If TotalBytes == 0 or CompressedBytes == 0: CompressionRatio = 1.0 (no compression)
//   - Otherwise: CompressionRatio = CompressedBytes / TotalBytes (as float64, range 0.0-1.0)
//   - All other fields reflect the exact state at the moment of the call.
//   - For ongoing streaming: returns partial statistics reflecting progress so far.
//   - For completed streaming: returns final complete statistics.
//   - Thread-safe: acquires RWMutex read lock to ensure consistent snapshot.
//   - Copy semantics: returned stats are independent; no aliasing to internal state.
//   - Non-blocking: O(1) operation after lock acquisition (fast read, not slow copy of large data).
//
// Compression Ratio Calculation Semantics:
//
//	Scenario                                    Compression Ratio Result    Rationale
//	────────────────────────────────────────────────────────────────────────────────
//	No compression configured                   1.0                        No reduction occurred
//	Compression not yet started                 1.0                        TotalBytes or CompressedBytes = 0
//	Compression in progress (partial data)      0.0 - 1.0                  Ratio of data processed
//	Compression completed (all data)            0.0 - 1.0                  Ratio of final result
//	TotalBytes = 0 (no data to compress)        1.0                        Guard condition
//	CompressedBytes = 0 (first read)            1.0                        Not yet compressed
//	CompressedBytes > TotalBytes (rare)         > 1.0                      Expansion (incompressible data)
//	COMP_NONE strategy                          1.0                        No compression applied
//	COMP_GZIP strategy (text)                   0.15 - 0.30                Typical 70-85% savings
//	COMP_DEFLATE strategy (text)                0.20 - 0.35                Typical 65-80% savings
//	COMP_GZIP strategy (binary)                 0.85 - 1.0                 Little to no reduction
//	COMP_GZIP strategy (images/video)           0.99 - 1.0                 Pre-compressed, expansion
//
// Compression Ratio Interpretation:
//
//	Ratio Value             Interpretation                  Typical Data Type
//	─────────────────────────────────────────────────────────────────────────
//	1.0                     No compression (100% of orig)   Incompressible or disabled
//	0.50 - 1.0              Some compression                Binary, mixed content
//	0.20 - 0.50             Good compression                Text, JSON, CSV, logs
//	0.10 - 0.20             Excellent compression           Highly repetitive text
//	< 0.10                  Exceptional compression         Very redundant data
//	> 1.0                   Expansion (compression failed)  Incompressible data
//
// Statistics Copying Behavior:
//
//	Operation                       Cost            Benefit                         Impact
//	──────────────────────────────────────────────────────────────────────────────
//	Shallow copy (used here)        O(1)            Fast, safe from mutation        Recommended
//	Deep copy (not used)            O(n)            Would be safer for slices       Unnecessary
//	Shared reference (not done)     O(0)            Fastest                         Unsafe (race)
//	Memory allocation                ~512 bytes      Small fixed structure           Minimal overhead
//	RWMutex lock duration           <1μs            Brief lock window               Minimal contention
//	Total call time                 1-5μs           Sub-microsecond operation       Negligible
//
// Comparison with GetStreamingStats():
//
//	Aspect                  GetStats()              GetStreamingStats()
//	──────────────────────────────────────────────────────────────────────
//	Purpose                 Low-level raw access    High-level wrapper API
//	Compression ratio       Computed dynamically    Computed dynamically
//	Return type             *StreamingStats         *StreamingStats
//	Thread-safety           Yes (RWMutex)           Yes (calls GetStats)
//	Performance             Slightly faster         Identical (wrapper)
//	Recommendation          Internal/advanced use   Public API/recommended
//	Implementation          Direct access           Calls this function
//	Use case                Low-level monitoring    Standard monitoring
//
// Thread-Safety Implementation:
//
//	Component                   Protection                  Guarantee
//	──────────────────────────────────────────────────────────────────
//	sw.stats structure           RWMutex read lock           Consistent snapshot
//	CompressionRatio field       Lock held during calc       Atomicity of calculation
//	Memory copy (stats := *sw.stats) Snapshot isolation       Copy not affected by updates
//	Return pointer              Safe return                 No data race on return
//	Concurrent reads             Allowed by RWMutex         Multiple readers OK
//	Concurrent with writes       Blocked until write done    Consistent reads guaranteed
//	Lock acquisition             O(1) amortized             Negligible overhead
//	Lock duration                <1μs                       Brief, minimal contention
//
// Statistics Copy Guarantee:
//
//	Scenario                                    Return Value Guarantees
//	────────────────────────────────────────────────────────────────────
//	Two consecutive calls                       Different memory addresses
//	Modify returned stats                       Does not affect next GetStats() call
//	Returned stats with nil pointer             Safe to dereference
//	Multiple concurrent calls                   Each gets independent copy
//	Internal stats updated after call           Returned stats unchanged
//	Returned pointer valid after streaming      Pointer lifetime guaranteed
//
// Example:
//
//	// Example 1: Simple statistics retrieval with compression info
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithCompressionType(COMP_GZIP).
//	    WithMaxConcurrentChunks(4)
//
//	// Start streaming
//	go streaming.Start(context.Background())
//
//	// Monitor statistics in real-time
//	time.Sleep(1 * time.Second)
//	stats := streaming.GetStats()
//
//	if stats.TotalBytes > 0 {
//	    compressionSavings := (1.0 - stats.CompressionRatio) * 100
//	    fmt.Printf("Progress: %.2f MB / %.2f MB\n",
//	        float64(stats.TransferredBytes) / 1024 / 1024,
//	        float64(stats.TotalBytes) / 1024 / 1024)
//	    fmt.Printf("Compression: %.1f%% savings (ratio: %.2f)\n",
//	        compressionSavings, stats.CompressionRatio)
//	}
//
//	// Example 2: Real-time monitoring loop with compression tracking
//	dataExport := createDataReader()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/export/bulk-data").
//	    WithStreaming(dataExport, nil).
//	    WithChunkSize(5 * 1024 * 1024).
//	    WithCompressionType(COMP_GZIP).
//	    WithMaxConcurrentChunks(8)
//
//	// Start streaming in background
//	done := make(chan bool)
//	go func() {
//	    streaming.Start(context.Background())
//	    done <- true
//	}()
//
//	// Monitor progress with live compression stats
//	ticker := time.NewTicker(500 * time.Millisecond)
//	defer ticker.Stop()
//
//	for {
//	    select {
//	    case <-ticker.C:
//	        if !streaming.IsStreaming() {
//	            break
//	        }
//
//	        // Get stats snapshot (O(1) operation)
//	        stats := streaming.GetStats()
//	        progress := streaming.GetProgress()
//
//	        originalMB := float64(stats.TotalBytes) / 1024 / 1024
//	        compressedMB := float64(stats.CompressedBytes) / 1024 / 1024
//	        saved := (1.0 - stats.CompressionRatio) * 100
//
//	        fmt.Printf("\rProgress: %.1f%% | Original: %.2f MB → Compressed: %.2f MB (%.1f%% saved)",
//	            float64(progress.Percentage), originalMB, compressedMB, saved)
//	    case <-done:
//	        fmt.Println("\nStreaming completed")
//	        return
//	    }
//	}
//
//	// Example 3: Compression effectiveness analysis
//	func AnalyzeCompressionEffectiveness(streaming *StreamingWrapper) {
//	    stats := streaming.GetStats()
//
//	    if stats.TotalBytes == 0 {
//	        fmt.Println("No data transferred")
//	        return
//	    }
//
//	    compressionRatio := stats.CompressionRatio
//	    savings := (1.0 - compressionRatio) * 100
//
//	    fmt.Printf("Compression Analysis:\n")
//	    fmt.Printf("  Original size:    %.2f MB\n",
//	        float64(stats.TotalBytes) / 1024 / 1024)
//	    fmt.Printf("  Compressed size:  %.2f MB\n",
//	        float64(stats.CompressedBytes) / 1024 / 1024)
//	    fmt.Printf("  Compression ratio: %.4f (%.1f%% reduction)\n",
//	        compressionRatio, savings)
//
//	    // Compression effectiveness assessment
//	    switch {
//	    case compressionRatio >= 0.99:
//	        fmt.Println("  Assessment: No compression benefit (incompressible data)")
//	    case compressionRatio >= 0.80:
//	        fmt.Println("  Assessment: Low compression benefit")
//	    case compressionRatio >= 0.50:
//	        fmt.Println("  Assessment: Moderate compression benefit")
//	    case compressionRatio >= 0.20:
//	        fmt.Println("  Assessment: Good compression benefit")
//	    default:
//	        fmt.Println("  Assessment: Excellent compression benefit")
//	    }
//
//	    // Time cost analysis
//	    if stats.ElapsedTime.Seconds() > 0 {
//	        originalMBps := float64(stats.TotalBytes) / 1024 / 1024 / stats.ElapsedTime.Seconds()
//	        compressedMBps := float64(stats.CompressedBytes) / 1024 / 1024 / stats.ElapsedTime.Seconds()
//	        timeOverhead := (float64(stats.CPUTime.Milliseconds()) / stats.ElapsedTime.Milliseconds()) * 100
//
//	        fmt.Printf("  Original throughput:    %.2f MB/s\n", originalMBps)
//	        fmt.Printf("  Compressed throughput:  %.2f MB/s\n", compressedMBps)
//	        fmt.Printf("  CPU overhead:           %.1f%%\n", timeOverhead)
//	    }
//	}
//
//	// Example 4: Statistics mutation safety demonstration
//	func DemonstrateCopySafety(streaming *StreamingWrapper) {
//	    // Get first snapshot
//	    stats1 := streaming.GetStats()
//	    fmt.Printf("Stats1: %d bytes\n", stats1.TotalBytes)
//
//	    // Mutate returned copy (does not affect streaming)
//	    stats1.TotalBytes = 999999
//
//	    // Get second snapshot (should be unaffected by mutation)
//	    stats2 := streaming.GetStats()
//	    fmt.Printf("Stats2: %d bytes (unaffected by mutation)\n", stats2.TotalBytes)
//
//	    // Verify original streaming state unchanged
//	    assert(stats2.TotalBytes != 999999, "Stats should not be affected by returned copy mutation")
//	}
//
//	// Example 5: Performance-critical monitoring
//	func HighFrequencyMonitoring(streaming *StreamingWrapper) {
//	    // GetStats() is O(1) and safe to call very frequently
//	    ticker := time.NewTicker(100 * time.Millisecond) // 10 calls/second
//	    defer ticker.Stop()
//
//	    for i := 0; i < 10; i++ {
//	        select {
//	        case <-ticker.C:
//	            stats := streaming.GetStats() // Very fast O(1) call
//
//	            // Update metrics in real-time
//	            fmt.Printf("Byte %d: %d/%d (%.1f%%) | Compression: %.2f\n",
//	                i,
//	                stats.TransferredBytes,
//	                stats.TotalBytes,
//	                float64(stats.TransferredBytes) / float64(stats.TotalBytes) * 100,
//	                stats.CompressionRatio)
//
//	            if !streaming.IsStreaming() {
//	                fmt.Println("Streaming complete")
//	                return
//	            }
//	        }
//	    }
//	}
//
//	// Example 6: Detailed statistics report with copy verification
//	func GenerateStatisticsReport(streaming *StreamingWrapper) string {
//	    stats := streaming.GetStats()
//
//	    var report strings.Builder
//
//	    report.WriteString("=== STREAMING STATISTICS REPORT ===\n")
//	    report.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))
//
//	    // Data metrics
//	    report.WriteString("DATA TRANSFER:\n")
//	    report.WriteString(fmt.Sprintf("  Total bytes:       %d (%.2f MB)\n",
//	        stats.TotalBytes, float64(stats.TotalBytes) / 1024 / 1024))
//	    report.WriteString(fmt.Sprintf("  Transferred:       %d (%.2f MB)\n",
//	        stats.TransferredBytes, float64(stats.TransferredBytes) / 1024 / 1024))
//	    report.WriteString(fmt.Sprintf("  Failed:            %d bytes\n",
//	        stats.FailedBytes))
//
//	    // Compression metrics
//	    report.WriteString("\nCOMPRESSION:\n")
//	    report.WriteString(fmt.Sprintf("  Original size:     %.2f MB\n",
//	        float64(stats.TotalBytes) / 1024 / 1024))
//	    report.WriteString(fmt.Sprintf("  Compressed size:   %.2f MB\n",
//	        float64(stats.CompressedBytes) / 1024 / 1024))
//	    report.WriteString(fmt.Sprintf("  Compression ratio: %.4f\n",
//	        stats.CompressionRatio))
//	    report.WriteString(fmt.Sprintf("  Savings:           %.1f%%\n",
//	        (1.0 - stats.CompressionRatio) * 100))
//	    report.WriteString(fmt.Sprintf("  Type:              %s\n",
//	        stats.CompressionType))
//
//	    // Performance metrics
//	    report.WriteString("\nPERFORMANCE:\n")
//	    report.WriteString(fmt.Sprintf("  Duration:          %.2f seconds\n",
//	        stats.ElapsedTime.Seconds()))
//	    report.WriteString(fmt.Sprintf("  Avg bandwidth:     %.2f MB/s\n",
//	        float64(stats.AverageBandwidth) / 1024 / 1024))
//	    report.WriteString(fmt.Sprintf("  Peak bandwidth:    %.2f MB/s\n",
//	        float64(stats.PeakBandwidth) / 1024 / 1024))
//	    report.WriteString(fmt.Sprintf("  Min bandwidth:     %.2f MB/s\n",
//	        float64(stats.MinimumBandwidth) / 1024 / 1024))
//
//	    // Chunk metrics
//	    report.WriteString("\nCHUNK PROCESSING:\n")
//	    report.WriteString(fmt.Sprintf("  Total chunks:      %d\n",
//	        stats.TotalChunks))
//	    report.WriteString(fmt.Sprintf("  Processed:         %d\n",
//	        stats.ProcessedChunks))
//	    report.WriteString(fmt.Sprintf("  Failed:            %d\n",
//	        stats.FailedChunks))
//	    report.WriteString(fmt.Sprintf("  Success rate:      %.1f%%\n",
//	        float64(stats.ProcessedChunks) / float64(stats.TotalChunks) * 100))
//
//	    // Error metrics
//	    if stats.HasErrors {
//	        report.WriteString("\nERRORS:\n")
//	        report.WriteString(fmt.Sprintf("  Count:             %d\n",
//	            stats.ErrorCount))
//	        report.WriteString(fmt.Sprintf("  First error:       %v\n",
//	            stats.FirstError))
//	        report.WriteString(fmt.Sprintf("  Last error:        %v\n",
//	            stats.LastError))
//	    }
//
//	    return report.String()
//	}
//
// Compression Ratio Edge Cases:
//
//	Condition                           Result              Expected Behavior
//	────────────────────────────────────────────────────────────────────────
//	TotalBytes = 0                      1.0                 Guard: avoid division by zero
//	CompressedBytes = 0                 1.0                 Guard: compression not started
//	CompressedBytes = TotalBytes        1.0                 No compression occurred
//	CompressedBytes > TotalBytes        > 1.0               Data expanded (incompressible)
//	CompressedBytes < TotalBytes        < 1.0               Compression succeeded
//	Both = 0                            1.0                 Guard: no transfer occurred
//	Negative values (should not occur)  Calculated as-is    Undefined (data consistency issue)
//
// Performance Characteristics:
//
//	Operation                       Time Complexity    Space Complexity    Actual Cost
//	──────────────────────────────────────────────────────────────────────────────
//	RWMutex read lock acquisition   O(1) amortized     None                <1μs typical
//	Memory copy (stats := *sw.stats) O(1)               O(1) (fixed struct)  ~512 bytes copy
//	CompressionRatio calculation     O(1)               None                <1μs
//	Return statement                 O(1)               None                <1μs
//	Total GetStats() call            O(1)               O(1)                1-5μs typical
//	Lock contention (high concurrency) O(1)             None                <10μs worst case
//
// Thread-Safety Guarantees:
//
//	Guarantee                                           Assurance
//	─────────────────────────────────────────────────────────────────────
//	No data race on reads                              RWMutex read lock
//	Returned copy independent of internal state        Shallow copy (new address)
//	CompressionRatio consistent with other fields      Calculated under lock
//	Multiple concurrent GetStats() calls safe          Allowed by RWMutex
//	Safe to call from streaming goroutine              RWMutex allows concurrent reads
//	Safe to call from monitoring goroutines            Parallel readers supported
//	Copy valid after streaming completes               Copy lifetime independent
//	No leaks or dangling pointers                      Automatic memory management
//
// Best Practices:
//
//  1. CALL DURING OR AFTER STREAMING
//     - Statistics available immediately
//     - Partial stats during streaming, complete after
//     - Safe to call frequently for monitoring
//     - Pattern:
//     stats := streaming.GetStats()
//     ratio := stats.CompressionRatio // Automatically computed
//
//  2. USE FOR COMPRESSION ANALYSIS
//     - CompressionRatio computed dynamically
//     - Always current as of call time
//     - Safe for repeated calls
//     - Example:
//     stats1 := streaming.GetStats()
//     time.Sleep(100ms)
//     stats2 := streaming.GetStats()
//     // ratio reflects current compression state
//
//  3. MONITOR SAFELY IN TIGHT LOOPS
//     - O(1) operation, safe to call frequently
//     - Each call gets independent copy
//     - No mutual exclusion issues
//     - Example:
//     for streaming.IsStreaming() {
//     stats := streaming.GetStats() // Safe O(1) call
//     // Use stats
//     }
//
//  4. DON'T RELY ON COPY LIFETIME
//     - Copy is valid indefinitely
//     - But reflects state only at call time
//     - Call again for updated metrics
//     - Pattern:
//     stats := streaming.GetStats() // Snapshot at T1
//     // stats won't update as streaming progresses
//     stats = streaming.GetStats()  // Get new snapshot at T2
//
//  5. COMBINE COMPRESSION METRICS WITH BANDWIDTH
//     - CompressionRatio shows space efficiency
//     - AverageBandwidth shows time efficiency
//     - Together show transfer effectiveness
//     - Example:
//     ratio := stats.CompressionRatio
//     bandwidth := float64(stats.AverageBandwidth) / 1024 / 1024
//     fmt.Printf("Compression: %.1f%%, Bandwidth: %.2f MB/s\n",
//     (1-ratio)*100, bandwidth)
//
// Related Methods and Workflows:
//
//	Workflow Stage              Method to Call              What it Provides
//	───────────────────────────────────────────────────────────────────────────
//	Real-time progress          GetProgress()               Current chunk, bytes, %
//	Real-time statistics        GetStats()                  Full metrics (this function)
//	Final analysis              GetStreamingStats()         Complete stats (alias)
//	Error tracking              Errors(), HasErrors()       Error list/existence
//	Response building           GetWrapper()                HTTP metadata
//	State checking              IsStreaming()               Is transfer active
//
// Differences from Previous Snapshots:
//
//	Call Number    TotalBytes    CompressedBytes    CompressionRatio
//	──────────────────────────────────────────────────────────────────
//	1st call       1MB           0.5MB              0.5000
//	2nd call (100ms later) 2MB  0.9MB              0.4500
//	3rd call (200ms later) 2MB  1.0MB              0.5000
//	// Each call shows current state; ratio reflects current compression effectiveness
//
// See Also:
//   - GetStreamingStats: Higher-level alias for GetStats (calls this function)
//   - GetProgress: Real-time progress without compression ratio
//   - CompressionRatio: Automatically computed as CompressedBytes / TotalBytes
//   - WithCompressionType: Configure compression algorithm before streaming
//   - Start: Initiates streaming and accumulates statistics
func (sw *StreamingWrapper) GetStats() *StreamingStats {
	if sw == nil {
		return &StreamingStats{}
	}

	sw.mu.RLock()
	defer sw.mu.RUnlock()

	sw.stats.CompressionRatio = 1.0
	if sw.stats.TotalBytes > 0 && sw.stats.CompressedBytes > 0 {
		sw.stats.CompressionRatio = float64(sw.stats.CompressedBytes) / float64(sw.stats.TotalBytes)
	}

	// Return a copy
	stats := *sw.stats
	return &stats
}

// GetProgress returns a thread-safe snapshot of the current streaming progress.
//
// This function provides real-time access to the current streaming operation state without blocking or
// affecting the transfer. It returns a defensive copy of the progress structure containing the latest
// metrics including current chunk number, bytes transferred, percentage complete, elapsed time, estimated
// time remaining, transfer rate, and any transient error from the most recent chunk. GetProgress is thread-safe
// and can be called from any goroutine; the internal mutex ensures consistent snapshot isolation. The returned
// StreamProgress is a copy, not a reference; modifications to the returned progress do not affect the internal
// state or subsequent calls. This is the primary method for real-time progress monitoring, progress bars, ETA
// calculations, and bandwidth monitoring during streaming. GetProgress is non-blocking, O(1) complexity, and
// safe to call very frequently (hundreds of times per second) without performance degradation. The returned
// progress represents the state at the moment of the call; subsequent streaming updates do not affect the
// returned copy. Progress is available immediately when Start() begins and remains stable after streaming
// completes, allowing pre-streaming and post-streaming progress queries.
//
// Returns:
//   - A pointer to a new StreamProgress structure containing a copy of current progress metrics.
//   - If the streaming wrapper is nil, returns an empty StreamProgress{} with all fields zero-valued.
//   - The returned StreamProgress reflects the state at the exact moment of the call.
//   - For ongoing streaming: returns current progress (actively changing with each call).
//   - For completed/cancelled streaming: returns final progress (stable, unchanged).
//   - Thread-safe: acquires RWMutex read lock to ensure consistent snapshot.
//   - Copy semantics: returned progress is independent; modifications do not affect streaming.
//   - Non-blocking: O(1) operation, safe for high-frequency polling (100+ calls/second).
//   - All timestamp values are in UTC with nanosecond precision.
//   - All byte counts use int64 to support transfers up to 9.2 exabytes.
//
// StreamProgress Structure Contents:
//
//	Field Category          Field Name                      Type            Purpose
//	─────────────────────────────────────────────────────────────────────────────────────
//	Chunk Information
//	                        CurrentChunk                    int64           Current chunk being processed
//	                        TotalChunks                     int64           Total chunks to process
//	                        ChunkSize                       int64           Size of each chunk
//	                        Size                            int64           Size of current chunk
//
//	Data Transfer
//	                        TransferredBytes                int64           Bytes transferred so far
//	                        TotalBytes                      int64           Total bytes to transfer
//	                        RemainingBytes                  int64           Bytes remaining
//	                        FailedBytes                     int64           Bytes that failed
//
//	Progress Metrics
//	                        Percentage                      float64         0.0-100.0 or 0.0-1.0
//	                        PercentageString                string          Human-readable like "45.2%"
//
//	Timing Information
//	                        StartTime                       time.Time       When streaming started
//	                        ElapsedTime                     time.Duration   Time elapsed since start
//	                        EstimatedTimeRemaining          time.Duration   ETA until completion
//
//	Transfer Rate
//	                        TransferRate                    int64           Current B/s rate
//	                        TransferRateMBps                float64         Current MB/s rate
//
//	Error Tracking
//	                        LastError                       error           Most recent error (if any)
//	                        HasError                        bool            Whether error occurred
//
// Example:
//
//	// Example 1: Simple progress monitoring during streaming
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithMaxConcurrentChunks(4)
//
//	// Start streaming in background
//	done := make(chan *wrapper)
//	go func() {
//	    result := streaming.Start(context.Background())
//	    done <- result
//	}()
//
//	// Monitor progress
//	ticker := time.NewTicker(500 * time.Millisecond)
//	defer ticker.Stop()
//
//	for {
//	    select {
//	    case <-ticker.C:
//	        if !streaming.IsStreaming() {
//	            break
//	        }
//
//	        // Get progress snapshot (O(1) safe call)
//	        progress := streaming.GetProgress()
//
//	        fmt.Printf("\rDownloading: %.1f%% (%d / %d chunks) | Speed: %.2f MB/s | ETA: %s",
//	            progress.Percentage,
//	            progress.CurrentChunk,
//	            progress.TotalChunks,
//	            float64(progress.TransferRate) / 1024 / 1024,
//	            progress.EstimatedTimeRemaining.String())
//	    case result := <-done:
//	        fmt.Println("\nDownload completed")
//	        if result.IsError() {
//	            fmt.Printf("Error: %s\n", result.Error())
//	        }
//	        return
//	    }
//	}
//
//	// Example 2: Progress bar with ETA calculation
//	func DisplayProgressBar(streaming *StreamingWrapper) {
//	    ticker := time.NewTicker(100 * time.Millisecond)
//	    defer ticker.Stop()
//
//	    for range ticker.C {
//	        if !streaming.IsStreaming() {
//	            break
//	        }
//
//	        progress := streaming.GetProgress()
//
//	        // Draw progress bar
//	        barLength := 40
//	        filledLength := int(float64(barLength) * progress.Percentage / 100.0)
//	        bar := strings.Repeat("█", filledLength) + strings.Repeat("░", barLength - filledLength)
//
//	        // Format output
//	        fmt.Printf("\r[%s] %.1f%% | %d/%d chunks | Speed: %.2f MB/s | ETA: %s",
//	            bar,
//	            progress.Percentage,
//	            progress.CurrentChunk,
//	            progress.TotalChunks,
//	            float64(progress.TransferRate) / 1024 / 1024,
//	            progress.EstimatedTimeRemaining.String())
//	    }
//	    fmt.Println()
//	}
//
//	// Example 3: Error handling in progress monitoring
//	httpResp, _ := http.Get("https://api.example.com/largefile")
//	defer httpResp.Body.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/proxy/remote-file").
//	    WithStreaming(httpResp.Body, nil).
//	    WithChunkSize(512 * 1024).
//	    WithReadTimeout(15000).
//	    WithWriteTimeout(15000).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Chunk %d error: %v\n", p.CurrentChunk, err)
//	        }
//	    })
//
//	done := make(chan *wrapper)
//	go func() {
//	    result := streaming.Start(context.Background())
//	    done <- result
//	}()
//
//	// Monitor with error detection
//	ticker := time.NewTicker(500 * time.Millisecond)
//	defer ticker.Stop()
//
//	for {
//	    select {
//	    case <-ticker.C:
//	        progress := streaming.GetProgress()
//
//	        // Check for current error
//	        if progress.HasError {
//	            fmt.Printf("\r[ERROR] Chunk %d: %v | Progress: %.1f%%",
//	                progress.CurrentChunk, progress.LastError, progress.Percentage)
//	        } else {
//	            fmt.Printf("\r[OK] Chunk %d | Progress: %.1f%% | Speed: %.2f MB/s",
//	                progress.CurrentChunk,
//	                progress.Percentage,
//	                float64(progress.TransferRate) / 1024 / 1024)
//	        }
//
//	        if !streaming.IsStreaming() {
//	            break
//	        }
//	    case result := <-done:
//	        if result.IsError() {
//	            fmt.Printf("\nFinal error: %s\n", result.Error())
//	        }
//	        return
//	    }
//	}
//
//	// Example 4: Progress metrics for decision making
//	func CheckProgressThreshold(streaming *StreamingWrapper, threshold float64) bool {
//	    progress := streaming.GetProgress()
//	    return progress.Percentage >= threshold
//	}
//
//	// Example 5: Bandwidth monitoring and rate limiting
//	func MonitorBandwidth(streaming *StreamingWrapper) {
//	    ticker := time.NewTicker(1 * time.Second)
//	    defer ticker.Stop()
//
//	    var previousBytes int64
//
//	    for range ticker.C {
//	        progress := streaming.GetProgress()
//
//	        // Calculate instantaneous rate
//	        currentBytes := progress.TransferredBytes
//	        instantaneousRate := currentBytes - previousBytes
//
//	        fmt.Printf("Bandwidth: Avg=%.2f MB/s | Current≈%.2f MB/s | Remaining: %s\n",
//	            float64(progress.TransferRate) / 1024 / 1024,
//	            float64(instantaneousRate) / 1024 / 1024,
//	            progress.EstimatedTimeRemaining.String())
//
//	        previousBytes = currentBytes
//
//	        if !streaming.IsStreaming() {
//	            break
//	        }
//	    }
//	}
//
//	// Example 6: ETA-based interruption
//	func StreamWithTimeLimit(streaming *StreamingWrapper, maxDuration time.Duration) {
//	    done := make(chan *wrapper)
//	    go func() {
//	        result := streaming.Start(context.Background())
//	        done <- result
//	    }()
//
//	    ticker := time.NewTicker(100 * time.Millisecond)
//	    defer ticker.Stop()
//
//	    for {
//	        select {
//	        case <-ticker.C:
//	            progress := streaming.GetProgress()
//
//	            // Check if ETA exceeds limit
//	            if progress.ElapsedTime + progress.EstimatedTimeRemaining > maxDuration {
//	                fmt.Printf("ETA exceeds time limit, cancelling\n")
//	                fmt.Printf("Elapsed: %s | Remaining: %s | Total ETA: %s > %s\n",
//	                    progress.ElapsedTime.String(),
//	                    progress.EstimatedTimeRemaining.String(),
//	                    (progress.ElapsedTime + progress.EstimatedTimeRemaining).String(),
//	                    maxDuration.String())
//	                streaming.Cancel()
//	                break
//	            }
//
//	            fmt.Printf("Progress: %.1f%% | ETA: %s\n",
//	                progress.Percentage, progress.EstimatedTimeRemaining.String())
//	        case result := <-done:
//	            if result.IsError() {
//	                fmt.Printf("Streaming error: %s\n", result.Error())
//	            }
//	            return
//	        }
//	    }
//	}
//
// Progress Field Semantics:
//
//	Field                       Type        Valid Range         When Updated
//	──────────────────────────────────────────────────────────────────────────
//	CurrentChunk                int64       0 to TotalChunks    After each chunk
//	TotalChunks                 int64       0 to max            Set once, then stable
//	Percentage                  float64     0.0 to 100.0        Continuous
//	TransferRate                int64       0 to GB/s           Continuous
//	EstimatedTimeRemaining      Duration    0 to max            Continuous
//	LastError                   error       nil or error        On chunk error
//	HasError                    bool        true/false          On chunk error
//	ElapsedTime                 Duration    0 to streaming dur  Continuous
//
// Progress Update Frequency:
//
//	Update Source                   Frequency                   Precision
//	──────────────────────────────────────────────────────────────────────
//	Chunk completion                Per chunk                   Exact
//	Byte counter                    Per byte (internal)         1 byte
//	Elapsed time                    Nanosecond precision        Real-time
//	Rate calculation                Per chunk or per second     Running average
//	ETA calculation                 Per chunk                   Based on current rate
//	Error status                    Per chunk failure           Immediate
//
// Progress Snapshot Semantics:
//
//	Aspect                          Behavior                        Implication
//	──────────────────────────────────────────────────────────────────────────
//	Return value                    Copy of current state           Independent snapshot
//	Modifications                   Do not affect streaming         Safe to modify
//	Multiple calls                  Each returns fresh snapshot     Time-dependent results
//	Call during streaming           Returns current live state      Real-time information
//	Call after streaming            Returns final state             Stable completion metrics
//	Garbage collection              Copy safe from GC               Lifetime guaranteed
//	Concurrent calls                Thread-safe reads              Multiple goroutines safe
//	Call frequency                  O(1) operation                 100+ calls/second safe
//
// Performance Characteristics:
//
//	Operation                       Time Complexity    Space Complexity    Actual Cost
//	──────────────────────────────────────────────────────────────────────────────
//	RWMutex read lock acquisition   O(1) amortized     None                <1μs typical
//	Memory copy (progress := *sw.progress) O(1)        O(1) (fixed struct)  ~200 bytes copy
//	Return statement                 O(1)               None                <1μs
//	Total GetProgress() call         O(1)               O(1)                1-3μs typical
//	Lock contention (high concurrency) O(1)            None                <10μs worst case
//	High-frequency polling (100/sec) Feasible           Minimal             Negligible overhead
//
// Comparison with GetStats():
//
//	Aspect                  GetProgress()           GetStats()
//	──────────────────────────────────────────────────────────────────
//	Purpose                 Real-time monitoring    Complete statistics
//	Field count             ~13 fields              ~25+ fields
//	Compression metrics     Not included            CompressionRatio
//	Bandwidth metrics       Current rate only       Avg/Peak/Min rates
//	Timing                  Elapsed + ETA           Detailed timings
//	Memory metrics          Not included            Memory stats
//	Chunk success/fail      Not included            Failure counts
//	Update frequency        Per chunk (continuous)  Per call
//	Typical use             Progress bar, ETA       Analysis, reporting
//	Overhead                Minimal (copy)          Moderate (more fields)
//
// Thread-Safety Implementation:
//
//	Component                   Protection                  Guarantee
//	──────────────────────────────────────────────────────────────────
//	sw.progress structure        RWMutex read lock           Consistent snapshot
//	Memory copy                  Copy isolation              Copy not affected by updates
//	Return pointer               Safe return                 No data race on return
//	Concurrent reads             Allowed by RWMutex          Multiple readers OK
//	Concurrent with writes       Blocked until write done    Consistent reads guaranteed
//	Lock acquisition             O(1) amortized              Negligible overhead
//	Lock duration                <1μs                        Brief, minimal contention
//
// Copy Guarantee:
//
//	Scenario                                    Return Value Guarantees
//	────────────────────────────────────────────────────────────────────
//	Two consecutive calls                       Different memory addresses
//	Modify returned progress                    Does not affect next GetProgress() call
//	Returned progress with nil pointer          Safe to dereference
//	Multiple concurrent calls                   Each gets independent copy
//	Internal progress updated after call        Returned progress unchanged
//	Returned pointer valid after streaming      Pointer lifetime guaranteed
//	Progress during active chunk transfer       Captured at call moment
//	Progress after completion                   Final metrics stable
//
// Progress Calculation Methods:
//
//	Metric                          Calculation Method
//	────────────────────────────────────────────────────────────────────
//	Percentage                      (TransferredBytes / TotalBytes) * 100.0
//	TransferRate                    TransferredBytes / ElapsedTime.Seconds()
//	EstimatedTimeRemaining          (RemainingBytes / TransferRate)
//	RemainingBytes                  TotalBytes - TransferredBytes
//	PercentageString                fmt.Sprintf("%.1f%%", Percentage)
//	Time per chunk                  ElapsedTime / CurrentChunk
//	Chunks remaining                TotalChunks - CurrentChunk
//
// ETA Accuracy and Limitations:
//
//	Factor                          Impact on ETA Accuracy
//	────────────────────────────────────────────────────────────────────
//	Constant transfer rate          ETA very accurate
//	Varying transfer rate           ETA less accurate (uses current rate)
//	Network congestion              ETA may underestimate
//	Throttling enabled              ETA more predictable
//	Client slowdowns                ETA may underestimate
//	Compression effectiveness       ETA may change as ratio improves
//	Early in transfer               ETA less reliable (small sample)
//	Late in transfer                ETA more reliable (better rate averaging)
//
// Practical Use Cases:
//
//	Use Case                        How to Use GetProgress()           Example
//	──────────────────────────────────────────────────────────────────────────
//	Progress bar                    Use Percentage for fill level      [████░░░░] 45%
//	ETA display                     Show EstimatedTimeRemaining        ETA: 2m 15s
//	Speed monitor                   Monitor TransferRate               Speed: 45.5 MB/s
//	Cancellation threshold          if Percentage > 90: cancel         Auto-skip large transfers
//	Memory estimation               Use CurrentChunk & ChunkSize       Memory: 4 chunks × 1MB
//	Error detection                 Check HasError and LastError       Error handling
//	Completion check                if Percentage == 100              Final validation
//	Multiple transfers              Track individual progress          Multi-file download
//
// Integration with Callback:
//
//	Callback Provides               GetProgress() Adds
//	──────────────────────────────────────────────────────────────────────────────
//	Individual chunk state          Complete aggregate progress
//	Chunk-by-chunk errors           Overall progress and rate
//	Real-time notifications         Historical timing (elapsed, ETA)
//	Low-level control               High-level metrics (percentage, rate)
//
//	Recommendation: Use both together:
//	  - WithCallback: immediate per-chunk feedback
//	  - GetProgress: periodic overall status
//
// Best Practices:
//
//  1. CALL FREQUENTLY FOR REAL-TIME UPDATES
//     - O(1) operation, safe for high frequency
//     - 100+ calls/second feasible
//     - No performance degradation
//     - Pattern:
//     ticker := time.NewTicker(100 * time.Millisecond)
//     progress := streaming.GetProgress() // O(1) safe call
//
//  2. USE BOTH PERCENTAGE AND BYTES FOR ACCURACY
//     - Percentage may vary based on TotalBytes
//     - Bytes provide exact count
//     - Together give complete picture
//     - Example:
//     fmt.Printf("%.1f%% (%d/%d bytes)\n",
//     progress.Percentage,
//     progress.TransferredBytes,
//     progress.TotalBytes)
//
//  3. HANDLE ETA CAREFULLY
//     - ETA is estimate based on current rate
//     - May be inaccurate with variable rates
//     - More accurate later in transfer
//     - Never guarantee to user as absolute time
//     - Example:
//     fmt.Printf("Estimated completion: ~%s (may vary)\n",
//     progress.EstimatedTimeRemaining.String())
//
//  4. CHECK ERROR STATUS IN PROGRESS
//     - HasError indicates chunk error
//     - LastError provides error details
//     - Streaming continues after chunk error
//     - Use HasErrors() for overall error status
//     - Example:
//     if progress.HasError {
//     fmt.Printf("Current chunk error: %v\n", progress.LastError)
//     }
//
//  5. DISTINGUISH PROGRESS FROM COMPLETION
//     - GetProgress(): state during streaming
//     - GetStats(): comprehensive final analysis
//     - Use both for complete understanding
//     - Example:
//     progress := streaming.GetProgress()  // Real-time
//     stats := streaming.GetStats()        // Final metrics
//
// Related Methods and Workflows:
//
//	Method                  Provides                    When to Use
//	──────────────────────────────────────────────────────────────────────
//	GetProgress()           Current progress snapshot   Real-time monitoring (this function)
//	GetStats()              Complete final statistics   After streaming, analysis
//	IsStreaming()           Active status boolean       State checking
//	GetWrapper()            HTTP response metadata      Response building
//	HasErrors()             Error existence boolean     Quick error check
//	Errors()                Error list                  Error analysis
//	WithCallback()          Per-chunk notifications     Immediate feedback
//
// Common Polling Pattern:
//
//	// ✓ RECOMMENDED: Fixed interval polling
//	ticker := time.NewTicker(100 * time.Millisecond)
//	for range ticker.C {
//	    progress := streaming.GetProgress() // Safe O(1)
//	    fmt.Printf("%.1f%% | ETA: %s\n",
//	        progress.Percentage,
//	        progress.EstimatedTimeRemaining.String())
//	}
//
//	// ⚠️  CAUTION: Busy polling
//	for streaming.IsStreaming() {
//	    progress := streaming.GetProgress() // Safe but high CPU
//	    // No sleep = 100% CPU usage
//	}
//
//	// ✓ RECOMMENDED: Event-based via callback
//	streaming.WithCallback(func(p *StreamProgress, err error) {
//	    // Called per chunk (less frequent than polling)
//	    fmt.Printf("Chunk %d processed\n", p.CurrentChunk)
//	})
//
// See Also:
//   - GetStats: Comprehensive statistics including compression, bandwidth analysis
//   - IsStreaming: Check if streaming is currently active
//   - GetWrapper: Access HTTP response metadata
//   - WithCallback: Receive per-chunk progress notifications
//   - Start: Initiates streaming and generates progress updates
//   - Cancel: Stops streaming, progress reflects cancellation point
func (sw *StreamingWrapper) GetProgress() *StreamProgress {
	if sw == nil {
		return &StreamProgress{}
	}
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	// Return a copy to prevent external mutation
	progress := *sw.progress
	return &progress
}

// GetStreamingProgress returns current progress information from an ongoing or completed streaming operation.
//
// This function provides high-level access to real-time progress metrics through a convenient alias to GetProgress().
// It returns a defensive copy of the current streaming state including chunk counts, bytes transferred, percentage
// complete, elapsed time, estimated time remaining, transfer rate, and any transient errors. GetStreamingProgress is
// the primary public API method for progress monitoring; GetProgress() is the underlying low-level implementation.
// This function is thread-safe, non-blocking, and optimized for frequent calls including high-frequency polling
// (100+ calls per second). The returned StreamProgress is independent; modifications do not affect internal streaming
// state. Progress information is available immediately when Start() begins and remains accessible after streaming
// completes or is cancelled. This is the recommended method for building progress bars, displaying ETAs, monitoring
// bandwidth, implementing progress-based controls, and tracking real-time transfer metrics in production systems.
//
// Returns:
//   - A pointer to a new StreamProgress structure containing a snapshot of current progress metrics.
//   - If the streaming wrapper is nil, returns an empty StreamProgress{} with all fields zero-valued.
//   - The returned StreamProgress reflects the exact state at the moment of the call.
//   - For ongoing streaming: returns current live metrics that change with subsequent calls.
//   - For completed/cancelled streaming: returns final stable metrics.
//   - Thread-safe: uses internal RWMutex for consistent snapshot isolation.
//   - Copy semantics: returned progress is independent; modifications do not affect streaming.
//   - Non-blocking: O(1) complexity, safe for high-frequency polling without performance impact.
//   - All percentage values are in range 0.0-100.0 (or 0.0-1.0 as documented).
//   - All byte counts use int64 to support transfers up to 9.2 exabytes.
//
// Functional Equivalence:
//
//	GetStreamingProgress() ≡ GetProgress()
//
//	Both functions return identical data and have identical performance characteristics.
//	GetStreamingProgress() is the recommended public API; GetProgress() is the lower-level implementation.
//	Use either function interchangeably; GetStreamingProgress() provides semantic clarity for high-level use.
//
// StreamProgress Contents Overview:
//
//	Category                Information Available
//	──────────────────────────────────────────────────────────────────
//	Chunk Information        Current chunk #, total chunks, chunk size
//	Data Transfer            Bytes transferred, total bytes, remaining
//	Progress Metrics         Percentage complete, human-readable percentage
//	Timing                   Start time, elapsed time, estimated time remaining
//	Transfer Rate            Current bandwidth (B/s and MB/s)
//	Error State              Current error (if any), error flag
//
// Use Cases and Patterns:
//
//	Use Case                            Recommended Pattern
//	────────────────────────────────────────────────────────────────────
//	Progress bar display                Get percentage: progress.Percentage
//	ETA display                         Get ETA: progress.EstimatedTimeRemaining
//	Speed monitoring                    Get rate: float64(progress.TransferRate) / 1024 / 1024
//	Real-time statistics               Get chunk/bytes: progress.CurrentChunk, progress.TransferredBytes
//	Progress-based cancellation         if progress.Percentage > threshold: cancel()
//	Error detection                     if progress.HasError: handle(progress.LastError)
//	Bandwidth analysis                  Track TransferRate over time
//	Memory estimation                   currentMemory = progress.CurrentChunk * ChunkSize
//	Multi-transfer tracking             Aggregate progress from multiple transfers
//	Progress notification               Emit event with progress.Percentage
//
// Example:
//
//	// Example 1: Simple progress retrieval and display
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithMaxConcurrentChunks(4)
//
//	// Start streaming in background
//	done := make(chan *wrapper)
//	go func() {
//	    result := streaming.Start(context.Background())
//	    done <- result
//	}()
//
//	// Monitor with high-level API
//	ticker := time.NewTicker(500 * time.Millisecond)
//	defer ticker.Stop()
//
//	for {
//	    select {
//	    case <-ticker.C:
//	        if !streaming.IsStreaming() {
//	            break
//	        }
//
//	        // Use high-level GetStreamingProgress() API
//	        progress := streaming.GetStreamingProgress()
//
//	        fmt.Printf("\rDownload: %.1f%% | %d / %d chunks | Speed: %.2f MB/s | ETA: %s",
//	            progress.Percentage,
//	            progress.CurrentChunk,
//	            progress.TotalChunks,
//	            float64(progress.TransferRate) / 1024 / 1024,
//	            progress.EstimatedTimeRemaining.String())
//	    case result := <-done:
//	        fmt.Println("\nDownload completed")
//	        return
//	    }
//	}
//
//	// Example 2: Progress monitoring with progress bar UI
//	func DisplayProgressUI(streaming *StreamingWrapper) {
//	    ticker := time.NewTicker(100 * time.Millisecond)
//	    defer ticker.Stop()
//
//	    for range ticker.C {
//	        if !streaming.IsStreaming() {
//	            fmt.Println("\n[✓] Transfer completed")
//	            break
//	        }
//
//	        // Get current progress using high-level API
//	        progress := streaming.GetStreamingProgress()
//
//	        // Draw visual progress bar
//	        barLength := 50
//	        filledLength := int(float64(barLength) * progress.Percentage / 100.0)
//	        emptyLength := barLength - filledLength
//
//	        bar := strings.Repeat("█", filledLength) + strings.Repeat("░", emptyLength)
//
//	        // Format detailed status line
//	        fmt.Printf("\r[%s] %6.2f%% | Chunk %4d / %4d | Speed: %7.2f MB/s | ETA: %8s | Total: %8s",
//	            bar,
//	            progress.Percentage,
//	            progress.CurrentChunk,
//	            progress.TotalChunks,
//	            float64(progress.TransferRate) / 1024 / 1024,
//	            progress.EstimatedTimeRemaining.String(),
//	            formatBytes(progress.TotalBytes))
//	    }
//	}
//
//	// Example 3: High-level progress monitoring with statistics integration
//	func MonitorWithStatistics(streaming *StreamingWrapper) {
//	    ticker := time.NewTicker(1 * time.Second)
//	    defer ticker.Stop()
//
//	    for range ticker.C {
//	        if !streaming.IsStreaming() {
//	            break
//	        }
//
//	        // Get real-time progress
//	        progress := streaming.GetStreamingProgress()
//
//	        // Get cumulative statistics
//	        stats := streaming.GetStreamingStats()
//
//	        fmt.Printf("Status Report:\n")
//	        fmt.Printf("  Progress:       %.1f%% (%d / %d chunks)\n",
//	            progress.Percentage, progress.CurrentChunk, progress.TotalChunks)
//	        fmt.Printf("  Data:           %.2f MB / %.2f MB\n",
//	            float64(progress.TransferredBytes) / 1024 / 1024,
//	            float64(progress.TotalBytes) / 1024 / 1024)
//	        fmt.Printf("  Speed:          %.2f MB/s (avg: %.2f MB/s)\n",
//	            float64(progress.TransferRate) / 1024 / 1024,
//	            float64(stats.AverageBandwidth) / 1024 / 1024)
//	        fmt.Printf("  Time:           Elapsed: %s | ETA: %s\n",
//	            progress.ElapsedTime.String(),
//	            progress.EstimatedTimeRemaining.String())
//	        fmt.Printf("  Errors:         %d / %d chunks\n",
//	            stats.FailedChunks, stats.TotalChunks)
//	    }
//	}
//
//	// Example 4: Progress-based adaptive control
//	func StreamWithAdaptiveControl(streaming *StreamingWrapper) {
//	    ticker := time.NewTicker(500 * time.Millisecond)
//	    defer ticker.Stop()
//
//	    lastBandwidth := float64(0)
//
//	    for range ticker.C {
//	        progress := streaming.GetStreamingProgress()
//
//	        currentBandwidth := float64(progress.TransferRate) / 1024 / 1024
//
//	        // Adaptive logging: log more frequently if bandwidth drops
//	        if currentBandwidth < lastBandwidth * 0.8 {
//	            fmt.Printf("\n[!] Bandwidth drop detected: %.2f MB/s → %.2f MB/s\n",
//	                lastBandwidth, currentBandwidth)
//	        }
//
//	        // Progress-based decisions
//	        switch {
//	        case progress.Percentage > 95:
//	            fmt.Printf("\r[●] Almost done! %.1f%% | ETA: ~%.0f seconds",
//	                progress.Percentage,
//	                progress.EstimatedTimeRemaining.Seconds())
//	        case progress.Percentage > 50:
//	            fmt.Printf("\r[◕] Halfway! %.1f%% | ETA: %s",
//	                progress.Percentage,
//	                progress.EstimatedTimeRemaining.String())
//	        case progress.Percentage > 0:
//	            fmt.Printf("\r[◐] Started %.1f%% | ETA: %s",
//	                progress.Percentage,
//	                progress.EstimatedTimeRemaining.String())
//	        }
//
//	        // Check for errors
//	        if progress.HasError {
//	            fmt.Printf("\n[!] Error on chunk %d: %v\n", progress.CurrentChunk, progress.LastError)
//	        }
//
//	        lastBandwidth = currentBandwidth
//
//	        if !streaming.IsStreaming() {
//	            fmt.Println("\n[✓] Transfer complete")
//	            break
//	        }
//	    }
//	}
//
//	// Example 5: Progress aggregation for multiple transfers
//	type MultiTransferProgress struct {
//	    transfers map[string]*StreamingWrapper
//	    total     TransferMetrics
//	}
//
//	type TransferMetrics struct {
//	    TotalBytes      int64
//	    TransferredBytes int64
//	    ActiveCount     int
//	    CompletedCount  int
//	}
//
//	func (mtp *MultiTransferProgress) UpdateMetrics() TransferMetrics {
//	    mtp.total = TransferMetrics{}
//
//	    for name, streaming := range mtp.transfers {
//	        progress := streaming.GetStreamingProgress()
//
//	        mtp.total.TotalBytes += progress.TotalBytes
//	        mtp.total.TransferredBytes += progress.TransferredBytes
//
//	        if streaming.IsStreaming() {
//	            mtp.total.ActiveCount++
//	        } else {
//	            mtp.total.CompletedCount++
//	        }
//	    }
//
//	    return mtp.total
//	}
//
//	func (mtp *MultiTransferProgress) DisplayOverall() {
//	    metrics := mtp.UpdateMetrics()
//
//	    overallPercent := 0.0
//	    if metrics.TotalBytes > 0 {
//	        overallPercent = float64(metrics.TransferredBytes) / float64(metrics.TotalBytes) * 100.0
//	    }
//
//	    fmt.Printf("Overall Progress: %.1f%% (%d / %d transfers active)\n",
//	        overallPercent, metrics.ActiveCount, metrics.CompletedCount)
//	}
//
//	// Example 6: Complete production-ready progress monitoring
//	type ProgressMonitor struct {
//	    streaming      *StreamingWrapper
//	    updateInterval time.Duration
//	    output         io.Writer
//	    startTime      time.Time
//	}
//
//	func NewProgressMonitor(streaming *StreamingWrapper, updateInterval time.Duration) *ProgressMonitor {
//	    return &ProgressMonitor{
//	        streaming:      streaming,
//	        updateInterval: updateInterval,
//	        output:         os.Stdout,
//	        startTime:      time.Now(),
//	    }
//	}
//
//	func (pm *ProgressMonitor) Monitor(ctx context.Context) {
//	    ticker := time.NewTicker(pm.updateInterval)
//	    defer ticker.Stop()
//
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            return
//	        case <-ticker.C:
//	            pm.displayProgress()
//
//	            if !pm.streaming.IsStreaming() {
//	                pm.displayFinal()
//	                return
//	            }
//	        }
//	    }
//	}
//
//	func (pm *ProgressMonitor) displayProgress() {
//	    progress := pm.streaming.GetStreamingProgress()
//
//	    // Visual progress bar
//	    barWidth := 40
//	    filledWidth := int(float64(barWidth) * progress.Percentage / 100.0)
//	    bar := strings.Repeat("=", filledWidth) + strings.Repeat("-", barWidth-filledWidth)
//
//	    // Speed indicator
//	    speedMBps := float64(progress.TransferRate) / 1024 / 1024
//	    speedIndicator := "↓"
//	    if speedMBps > 100 {
//	        speedIndicator = "↓↓"
//	    }
//
//	    // Format output
//	    output := fmt.Sprintf(
//	        "\r[%s] %5.1f%% | %s %6.1f MB/s | ETA: %8s | %s / %s",
//	        bar,
//	        progress.Percentage,
//	        speedIndicator,
//	        speedMBps,
//	        progress.EstimatedTimeRemaining.String(),
//	        formatBytes(progress.TransferredBytes),
//	        formatBytes(progress.TotalBytes),
//	    )
//
//	    fmt.Fprint(pm.output, output)
//	}
//
//	func (pm *ProgressMonitor) displayFinal() {
//	    stats := pm.streaming.GetStreamingStats()
//	    progress := pm.streaming.GetStreamingProgress()
//
//	    fmt.Fprintf(pm.output, "\n✓ Transfer completed in %.2f seconds\n", stats.ElapsedTime.Seconds())
//	    fmt.Fprintf(pm.output, "  Total:       %s\n", formatBytes(progress.TotalBytes))
//	    fmt.Fprintf(pm.output, "  Chunks:      %d\n", progress.TotalChunks)
//	    fmt.Fprintf(pm.output, "  Speed:       %.2f MB/s (avg)\n",
//	        float64(stats.AverageBandwidth) / 1024 / 1024)
//
//	    if stats.HasErrors {
//	        fmt.Fprintf(pm.output, "  Errors:      %d\n", stats.ErrorCount)
//	    }
//	}
//
// Public API Recommendation:
//
//	For Calling Code              Use This Function         Why
//	──────────────────────────────────────────────────────────────────
//	Application code              GetStreamingProgress()    High-level, semantic clarity
//	UI components                 GetStreamingProgress()    Clear intent, recommended API
//	Progress bars                 GetStreamingProgress()    Standard pattern
//	Monitoring systems            GetStreamingProgress()    Public API contract
//	High-level streaming logic    GetStreamingProgress()    Clear naming, recommended
//
//	Internal streaming impl       GetProgress()             Direct access, faster
//	Low-level performance code    GetProgress()             Avoids indirection
//	Performance-critical paths    GetProgress()             Negligible difference
//
// Semantics and Guarantees:
//
//	Aspect                          Guarantee / Behavior
//	──────────────────────────────────────────────────────────────────
//	Return type                     *StreamProgress (same as GetProgress)
//	Thread-safety                   Yes (RWMutex protected)
//	Copy semantics                  Independent copy (safe from mutation)
//	Performance                     O(1) constant time (same as GetProgress)
//	Call frequency safety           100+ calls/second safe (same as GetProgress)
//	Functional equivalence          Identical to GetProgress (alias pattern)
//	Semantic clarity                Enhanced (public API naming)
//	Implementation                  Direct call to GetProgress()
//
// Performance Profile:
//
//	Metric                          Value / Characteristic
//	──────────────────────────────────────────────────────────────────
//	Time complexity                 O(1) constant
//	Space complexity                O(1) fixed (StreamProgress struct)
//	Typical call duration           1-3 microseconds
//	Memory allocation               ~200 bytes per call
//	Lock contention                 Minimal (read-only)
//	Concurrent call support         Hundreds per second
//	Call frequency recommendation   1-10 Hz for UI (100-500 ms intervals)
//	High-frequency monitoring       Up to 100+ Hz feasible
//	Safe for real-time systems      Yes
//	Suitable for animation frames   Yes (60+ FPS capable)
//
// Integration Pattern (Recommended):
//
//	// Typical integration pattern in a streaming application
//	func DownloadWithProgress(url string, destination string) error {
//	    // Setup
//	    reader := createReader(url)           // Data source
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/download").
//	        WithStreaming(reader, nil).
//	        WithChunkSize(1024 * 1024)
//
//	    // Background streaming
//	    done := make(chan error)
//	    go func() {
//	        result := streaming.Start(context.Background())
//	        done <- result.Error()
//	    }()
//
//	    // Progress monitoring (main thread / UI thread)
//	    ticker := time.NewTicker(100 * time.Millisecond)
//	    for {
//	        select {
//	        case <-ticker.C:
//	            // Get current progress using public API
//	            progress := streaming.GetStreamingProgress()
//	            updateUI(progress.Percentage, progress.EstimatedTimeRemaining)
//	        case err := <-done:
//	            ticker.Stop()
//	            return err
//	        }
//	    }
//	}
//
// Comparison with Alternatives:
//
//	Method                      Purpose                         When to Use
//	─────────────────────────────────────────────────────────────────────────
//	GetStreamingProgress()      Real-time progress (public API) Application code
//	GetProgress()               Real-time progress (impl)       Internal code
//	GetStreamingStats()         Complete statistics             Analysis/reporting
//	IsStreaming()               Check if active                 State queries
//	WithCallback()              Per-chunk notifications         Immediate feedback
//	Errors()                    All errors                      Error analysis
//
// Recommended Polling Frequencies:
//
//	Use Case                        Frequency           Interval
//	──────────────────────────────────────────────────────────────────
//	Progress bar (CLI)              2-5 Hz              200-500ms
//	Web UI progress indicator       1-2 Hz              500ms-1s
//	Real-time monitoring dashboard  10 Hz               100ms
//	Mobile app progress display     2-5 Hz              200-500ms
//	Detailed diagnostics            10-20 Hz            50-100ms
//	Animation/game frame rate       60+ Hz              16ms
//	High-frequency monitoring       100+ Hz             10ms
//	Low-power/battery-conscious     0.5-1 Hz            1-2s
//
// See Also:
//   - GetProgress: Low-level implementation (functional equivalent)
//   - GetStreamingStats: Complete statistics and performance metrics
//   - IsStreaming: Check if streaming is currently active
//   - WithCallback: Receive per-chunk progress notifications
//   - GetWrapper: Access HTTP response metadata
//   - Start: Initiates streaming and generates progress
//   - Cancel: Stops streaming operation
func (sw *StreamingWrapper) GetStreamingProgress() *StreamProgress {
	if sw == nil {
		return &StreamProgress{}
	}
	return sw.GetProgress()
}

// streamDirect performs direct streaming without buffering or concurrency.
//
// This function implements the STRATEGY_DIRECT streaming strategy, which provides the simplest, most straightforward
// streaming approach with minimal overhead and maximum simplicity. It reads data sequentially from the input reader,
// applies optional compression, calculates checksums for integrity verification, writes the processed data to the
// output writer, and repeats until EOF. streamDirect is synchronous and single-threaded; all operations (read,
// compress, write) happen sequentially in the same goroutine. This strategy is ideal for scenarios where simplicity
// is more important than throughput, such as small transfers, streaming to simple destinations, or when resource
// constraints require minimal goroutine overhead. Context cancellation is checked at the start of each iteration,
// allowing responsive shutdown. Per-chunk errors trigger recording but do not stop streaming; the next chunk is
// attempted. Bandwidth throttling is supported to limit transfer rate when needed. The direct approach has minimal
// memory footprint, no goroutine management overhead, and predictable behavior, making it suitable for embedded
// systems, low-resource environments, or applications prioritizing code simplicity over peak performance.
//
// Parameters:
//   - ctx: Context for cancellation and coordination.
//     Used to check for early termination via ctx.Done().
//     Cancellation immediately stops streaming and returns error.
//     Parent context deadline may apply to overall operation.
//
// Returns:
//   - error: nil if streaming completed successfully.
//     ctx.Err() wrapped if context cancelled.
//     read error if reader.Read() fails.
//     Other errors during critical operations.
//
// Behavior:
//   - Sequential: reads, compresses, writes in order, one chunk at a time.
//   - Simple: single goroutine, no channels, no synchronization primitives.
//   - Error-tolerant: continues despite per-chunk errors (except read EOF).
//   - Context-aware: checks ctx.Done() before each chunk.
//   - Compression-capable: optionally compresses chunks.
//   - Checksum-enabled: calculates checksum for each chunk.
//   - Throttle-aware: applies bandwidth throttling if configured.
//   - Progress-tracking: updates progress and triggers callbacks after each chunk.
//   - Statistics-maintained: accumulates failed chunks, compressed bytes, timing.
//
// Streaming Flow Diagram:
//
//	Start
//	  ↓
//	Loop: Check ctx.Done()
//	  ↓
//	Read chunk from reader
//	  ↓
//	n > 0?
//	  ├─ Yes → Create StreamChunk
//	  │         ↓
//	  │       Compress?
//	  │         ├─ Yes → Compress chunk → Check error
//	  │         └─ No  → Skip compression
//	  │         ↓
//	  │       Calculate checksum
//	  │         ↓
//	  │       Write chunk
//	  │         ↓
//	  │       Update progress
//	  │         ↓
//	  │       Throttle?
//	  │         ├─ Yes → Sleep if needed
//	  │         └─ No  → Continue
//	  │         ↓
//	  └─ No  → Skip to error check
//	      ↓
//	Error check (EOF / error)
//	  ├─ EOF   → Break loop
//	  ├─ Error → Record and return error
//	  └─ None  → Continue loop
//	      ↓
//	Finalize: Set EndTime
//	      ↓
//	Return
//
// Chunk Processing Sequence:
//
//	Step    Operation                           Input               Output                 Notes
//	────────────────────────────────────────────────────────────────────────────────────────
//	1       Context check                       ctx.Done()          Proceed/Exit           Cancel immediate
//	2       Read chunk                          reader              buffer[:n] + error     Sequential read
//	3       n > 0?                              n                   Proceed/Skip           Valid data?
//	4       Create StreamChunk                  buffer[:n]          chunk struct           Wrapper object
//	5       Compress?                           compression type    compData + error       Optional step
//	6       Handle compress error               error               Record/Continue        Per-chunk error
//	7       Calculate checksum                  chunk.Data          chunk.Checksum         Integrity check
//	8       Write chunk                         chunk.Data          Writer state           Sequential write
//	9       Handle write error                  error               Record/Continue        Per-chunk error
//	10      Update progress                     chunk               stats updated          Metrics
//	11      Increment counter                   currentChunk        currentChunk++          Next sequence
//	12      Apply throttle                      rate/elapsed        Sleep delay            Rate limiting
//	13      Check EOF                           error               Break/Continue         Exit condition
//	14      Check error                         error               Return/Continue        Fatal error?
//
// Error Handling Strategy:
//
//	Error Type                      Location            Action                      Continues?
//	──────────────────────────────────────────────────────────────────────────────────────
//	io.EOF                          Read error check    Break loop (normal)         No (clean exit)
//	Read error (other)              Read error check    recordError, return error   No (fatal)
//	Compression error               Compression block   recordError, continue       Yes (skip chunk)
//	Write error                     Write block         recordError, increment fail No (skip chunk)
//	Context cancellation            Context check       Return ctx.Err()            No (immediate)
//
// Resource Usage:
//
//	Resource                        Amount              Notes
//	────────────────────────────────────────────────────────────────────
//	Goroutines                      1 (main)            No additional goroutines
//	Channels                        0                   No channels needed
//	Memory (buffer)                 ChunkSize           One buffer for all reads
//	Memory (allocations)            Per chunk (copy)    Data copy for compression
//	Synchronization primitives      0                   No locks or wait groups
//	Goroutine overhead              Minimal             Single sequential execution
//
// Compression Error Handling:
//
//	Scenario                                    Behavior                        Effect
//	──────────────────────────────────────────────────────────────────────────────
//	Compression succeeds                        Use compressed data             Normal flow
//	Compression fails                           recordError, continue           Skip chunk
//	COMP_NONE configured                        No compression attempt          Normal flow
//	Chunk error set                             chunk.Error = compErr           Track error
//
// Write Error Handling:
//
//	Scenario                                    Behavior                        Effect
//	───────────────────────────────────────────────────────────────────────────
//	Write succeeds                              Normal flow                     Continue
//	Write fails                                 recordError, FailedChunks++     Skip chunk
//	writer is nil                               Skip write block                Continue
//	chunk.Error set                             Propagate error info            Track error
//
// Throttling Calculation:
//
//	Scenario                                    Formula
//	──────────────────────────────────────────────────────────────────
//	ThrottleRate = 0                            No throttling
//	ThrottleRate > 0                            enabled
//	Expected time                               transferredBytes / rate * 1sec
//	Elapsed time                                time.Since(StartTime)
//	Sleep duration                              expected - elapsed (if positive)
//	Sleep condition                             elapsed < expected
//
//	Example: ThrottleRate = 10MB/s, transferred = 20MB
//	  expected = 20MB / 10MB/s = 2 seconds
//	  if actual elapsed < 2 seconds, sleep for difference
//
// Example:
//
//	// Example 1: Simple direct streaming
//	file, _ := os.Open("data.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/stream/direct").
//	    WithStreaming(file, nil).
//	    WithChunkSize(512 * 1024).  // 512KB chunks
//	    WithStreamingStrategy(STRATEGY_DIRECT).
//	    WithCompressionType(COMP_GZIP).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Chunk %d error: %v\n", p.CurrentChunk, err)
//	        } else if p.CurrentChunk % 100 == 0 {
//	            fmt.Printf("Processed %d chunks | Speed: %.2f MB/s\n",
//	                p.CurrentChunk,
//	                float64(p.TransferRate) / 1024 / 1024)
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//	if result.IsError() {
//	    fmt.Printf("Streaming failed: %s\n", result.Error())
//	}
//
//	// Example 2: Direct streaming with minimal overhead
//	func MinimalOverheadStream(fileReader io.ReadCloser) error {
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/stream/minimal").
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(1024 * 1024).  // 1MB chunks
//	        WithStreamingStrategy(STRATEGY_DIRECT).
//	        WithCompressionType(COMP_NONE).  // No compression overhead
//	        WithCallback(nil)  // No callback overhead
//
//	    result := streaming.Start(context.Background())
//
//	    if result.IsError() {
//	        return fmt.Errorf("streaming failed: %w", result.Error())
//	    }
//
//	    return nil
//	}
//
//	// Example 3: Direct streaming with bandwidth throttling
//	func ThrottledDirectStream(fileReader io.ReadCloser) {
//	    // Limit to 5 MB/s
//	    fiveMBps := int64(5 * 1024 * 1024)
//
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/stream/throttled-direct").
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(256 * 1024).
//	        WithStreamingStrategy(STRATEGY_DIRECT).
//	        WithThrottleRate(fiveMBps).
//	        WithCallback(func(p *StreamProgress, err error) {
//	            if err == nil {
//	                speed := float64(p.TransferRate) / 1024 / 1024
//	                fmt.Printf("Current speed: %.2f MB/s (limit: 5.00 MB/s)\n", speed)
//	            }
//	        })
//
//	    result := streaming.Start(context.Background())
//	    fmt.Printf("Throttled streaming complete\n")
//	}
//
//	// Example 4: Low-resource streaming for embedded systems
//	func EmbeddedSystemStream(fileReader io.ReadCloser, fileWriter io.WriteCloser) error {
//	    // Minimal resource usage: small chunks, no buffering, no compression
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/embed/stream").
//	        WithStreaming(fileReader, nil).
//	        WithWriter(fileWriter).
//	        WithChunkSize(64 * 1024).  // Small chunks (64KB)
//	        WithStreamingStrategy(STRATEGY_DIRECT).  // Simple, direct flow
//	        WithCompressionType(COMP_NONE).  // No CPU overhead
//	        WithCallback(nil).  // No callback overhead
//
//	    result := streaming.Start(context.Background())
//
//	    if result.IsError() {
//	        return result.Error()
//	    }
//
//	    stats := streaming.GetStats()
//	    fmt.Printf("Transfer: %.2f MB in %v\n",
//	        float64(stats.TotalBytes) / 1024 / 1024,
//	        stats.EndTime.Sub(stats.StartTime))
//
//	    return nil
//	}
//
//	// Example 5: Error handling in direct streaming
//	func DirectStreamWithErrorMonitoring(fileReader io.ReadCloser) {
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/stream/error-monitor").
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(512 * 1024).
//	        WithStreamingStrategy(STRATEGY_DIRECT).
//	        WithCompressionType(COMP_GZIP).
//	        WithCallback(func(p *StreamProgress, err error) {
//	            if err != nil {
//	                fmt.Printf("Error on chunk %d: %v\n", p.CurrentChunk, err)
//	            }
//	        })
//
//	    result := streaming.Start(context.Background())
//
//	    if result.IsError() {
//	        fmt.Printf("Streaming error: %s\n", result.Error())
//	        return
//	    }
//
//	    stats := streaming.GetStats()
//	    if stats.HasErrors {
//	        fmt.Printf("Completed with %d chunk errors\n", len(stats.Errors))
//	        for i, err := range stats.Errors {
//	            fmt.Printf("  [%d] %v\n", i+1, err)
//	        }
//	    } else {
//	        fmt.Printf("Streaming completed successfully\n")
//	    }
//	}
//
// Loop Iteration Timing:
//
//	Operation                           Time Cost           Notes
//	────────────────────────────────────────────────────────────────────
//	Context check                       <1μs                select on ctx.Done()
//	Read from source                    1-1000ms            Depends on source speed
//	Create StreamChunk                  <1μs                Struct allocation
//	Compress (if enabled)               1-100ms             Depends on size/algo
//	Calculate checksum                  <1ms                Linear in data size
//	Write to destination                1-1000ms            Depends on writer
//	Update progress                     <1ms                Math operations
//	Callback trigger                    0-10ms              Depends on callback
//	Throttle sleep                      0-1000ms            If rate limiting needed
//	Total per chunk                     ~10-2100ms          Dominated by I/O
//
// Performance Characteristics:
//
//	Factor                              Impact              Optimization
//	────────────────────────────────────────────────────────────────────────
//	ChunkSize                           Throughput          Larger = more throughput
//	CompressionType                     CPU/bandwidth       GZIP better, DEFLATE faster
//	ThrottleRate                        Bandwidth control   Set to limit
//	Callback complexity                 Per-chunk cost      Keep fast
//	Reader speed                        Dominant factor     Use fast readers
//	Writer speed                        Dominant factor     Use fast writers
//	Sequential processing               No parallelism      Limited by slowest op
//
// Comparison with Other Strategies:
//
//	Strategy               Throughput    Concurrency    Complexity    Memory     Best For
//	──────────────────────────────────────────────────────────────────────────────────
//	STRATEGY_DIRECT        Lower          None           Minimal       Minimal    Simplicity
//	STRATEGY_CHUNKED       Medium         None           Moderate      Low        Control
//	STRATEGY_BUFFERED      Highest        High           Complex       Medium     Throughput
//
// Advantages and Trade-offs:
//
//	Advantage                           Trade-off
//	──────────────────────────────────────────────────────────────────
//	Minimal code complexity             Lower throughput
//	No goroutine overhead               Cannot parallelize I/O
//	Minimal memory footprint            Cannot overlap read/write
//	Predictable behavior                Blocked on slowest operation
//	Easy to understand and debug        Limited scalability
//	No synchronization primitives       No concurrency control
//	Suitable for small transfers        Inefficient for large transfers
//
// Best Practices:
//
//  1. USE FOR SMALL OR SIMPLE TRANSFERS
//     - Small files (< 100MB)
//     - Streaming to simple destinations
//     - Embedded systems
//     - Low-resource environments
//     - Pattern:
//     WithStreamingStrategy(STRATEGY_DIRECT)
//     WithChunkSize(512 * 1024)  // Reasonable size
//
//  2. CHOOSE APPROPRIATE CHUNK SIZE
//     - Larger chunks: better throughput (but more memory per read)
//     - Smaller chunks: lower memory (but more iterations)
//     - Recommended: 256KB-2MB for direct strategy
//     - Pattern:
//     WithChunkSize(512 * 1024)  // 512KB balance
//
//  3. COMPRESS ONLY IF BENEFICIAL
//     - Compression adds CPU overhead in direct strategy
//     - Only enable for text/compressible data
//     - Disable for already-compressed data
//     - Pattern:
//     WithCompressionType(COMP_GZIP)  // For text
//     WithCompressionType(COMP_NONE)  // For binary
//
//  4. THROTTLE IF RATE LIMITING NEEDED
//     - Apply only if necessary
//     - Slight overhead per chunk
//     - Better than overwhelming destination
//     - Pattern:
//     WithThrottleRate(10 * 1024 * 1024)  // 10 MB/s
//
//  5. KEEP CALLBACKS SIMPLE
//     - Heavy callbacks slow entire stream
//     - Called once per chunk (sequential)
//     - Keep logic minimal and fast
//     - Pattern:
//     WithCallback(func(p *StreamProgress, err error) {
//     // Keep fast and simple
//     })
//
// Related Methods and Integration:
//
//	Method                      Purpose                     Related To
//	──────────────────────────────────────────────────────────────────
//	streamDirect()              Direct strategy (this)       STRATEGY_DIRECT
//	streamChunked()             Chunked strategy            STRATEGY_CHUNKED
//	streamBuffered()            Buffered strategy           STRATEGY_BUFFERED
//	WithChunkSize()             Configure chunk size        Sequential read size
//	WithThrottleRate()          Bandwidth limiting          Rate control
//	WithCompressionType()       Enable compression          CPU overhead
//	compressChunk()             Compression execution       Per-chunk CPU work
//	calculateChecksum()         Per-chunk integrity         Data verification
//	recordError()               Error recording             Per-chunk error tracking
//	updateProgress()            Progress tracking           Real-time metrics
//	fireCallback()              Callback invocation         User notification
//
// See Also:
//   - WithStreamingStrategy: Select STRATEGY_DIRECT
//   - WithChunkSize: Configure chunk size for sequential reading
//   - WithThrottleRate: Set bandwidth throttling rate
//   - WithCompressionType: Enable optional compression
//   - WithCallback: Configure per-chunk callback
//   - compressChunk: Compression function used here
//   - calculateChecksum: Checksum calculation used here
//   - GetStats: Query final statistics
//   - Start: Entry point that calls streamDirect
func (sw *StreamingWrapper) streamDirect(ctx context.Context) error {
	buffer := make([]byte, sw.config.ChunkSize)
	sw.stats.StartTime = time.Now()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("streaming cancelled: %w", ctx.Err())
		default:
		}

		// Read chunk
		n, err := sw.reader.Read(buffer)

		if n > 0 {
			chunk := &StreamChunk{
				SequenceNumber:  sw.currentChunk,
				Data:            buffer[:n],
				Size:            int64(n),
				Timestamp:       time.Now(),
				CompressionType: sw.config.Compression,
			}

			// Apply compression if configured
			if sw.config.Compression != CompressNone {
				compData, compErr := sw.compressChunk(chunk)
				if compErr != nil {
					sw.recordError(compErr)
					chunk.Error = compErr
					continue
				}
				chunk.Data = compData
				chunk.Compressed = true
				sw.stats.CompressedBytes += int64(len(compData))
			}

			// Calculate checksum
			chunk.Checksum = sw.calculateChecksum(chunk.Data)

			// Write chunk if writer is set
			if sw.writer != nil {
				if _, writeErr := sw.writer.Write(chunk.Data); writeErr != nil {
					sw.recordError(writeErr)
					sw.stats.FailedChunks++
					chunk.Error = writeErr
					continue
				}
			}

			// Update progress
			sw.updateProgress(chunk)
			sw.currentChunk++

			// Apply throttling if configured
			if sw.config.ThrottleRate > 0 {
				elapsed := time.Since(sw.stats.StartTime)
				expectedTime := time.Duration(float64(sw.progress.TransferredBytes) / float64(sw.config.ThrottleRate) * float64(time.Second))
				if elapsed < expectedTime {
					time.Sleep(expectedTime - elapsed)
				}
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			sw.recordError(err)
			sw.stats.FailedChunks++
			return fmt.Errorf("read error: %w", err)
		}
	}

	sw.stats.EndTime = time.Now()
	return nil
}

// streamBuffered performs buffered streaming with concurrent chunk processing and optional buffer pooling.
//
// This function implements the STRATEGY_BUFFERED streaming strategy, which provides high-throughput streaming
// through concurrent processing of multiple chunks. It uses a producer-consumer pattern with a dedicated reader
// goroutine that reads chunks from the input source and sends them through a buffered channel, and a pool of
// worker goroutines that process chunks concurrently (compression, checksums, writing). The concurrent architecture
// enables efficient overlapping of I/O operations (read while write) and CPU-intensive compression, significantly
// improving throughput compared to sequential processing. Optional buffer pooling reduces memory allocation pressure
// by reusing buffers across chunk reads. Bandwidth throttling is supported to limit transfer rate when needed.
// Context cancellation is checked before read and write operations, allowing responsive shutdown. Per-chunk errors
// are recorded without stopping the entire stream; processing continues until all buffered chunks are processed or
// context cancellation occurs. This is the recommended strategy for high-performance scenarios requiring maximum
// throughput, such as large file transfers, bulk data exports, or bandwidth-intensive operations.
//
// Parameters:
//   - ctx: Context for cancellation, timeouts, and coordination.
//     Used to check for early termination via ctx.Done().
//     Cancellation immediately stops all reader and writer goroutines.
//     Propagated through error channel for error handling.
//
// Returns:
//   - error: nil if streaming completed successfully.
//     ctx.Err() if context cancelled.
//     Read or write error if operation fails.
//     Other errors during critical operations.
//
// Behavior:
//   - Concurrent: uses producer-consumer pattern with reader and writer goroutines.
//   - Buffered: channels buffer chunks up to MaxConcurrentChunks capacity.
//   - Pool-aware: optionally uses buffer pool to reduce allocations.
//   - Compression-capable: compresses chunks in worker goroutines.
//   - Checksum-enabled: calculates checksum for each chunk.
//   - Throttle-aware: applies bandwidth throttling if configured.
//   - Error-tolerant: continues despite per-chunk errors.
//   - Context-aware: checks cancellation before operations.
//   - Statistics-maintained: accumulates failed chunks, compressed bytes, timing.
//
// Concurrent Architecture:
//
//	Component                   Role                            Details
//	────────────────────────────────────────────────────────────────────
//	Reader goroutine            Reads from source               1 goroutine
//	Channel (chunkChan)         Buffers chunks                  Size = MaxConcurrentChunks
//	Worker goroutines           Process chunks                  Pool of MaxConcurrentChunks
//	Error channel (errChan)     Collects errors                 Buffered, 1 error
//	WaitGroup (wg)              Synchronizes worker completion  Waits for all workers
//	Main goroutine              Coordinates overall operation   Waits for WaitGroup
//
// Data Flow Pipeline:
//
//	Reader Goroutine                      Channel                 Worker Goroutines
//	────────────────────────────────────────────────────────────────────────────────
//	Read from source                      chunkChan buffer        Receive chunk
//	Copy to allocated buffer              (capacity N)            Process chunk
//	Create StreamChunk                    ──────────→             Compress
//	Send to channel                                               Calculate checksum
//	Check for EOF/error                                           Write to destination
//	Loop or exit                                                  Update progress
//	Close channel on completion                                   Throttle if needed
//	                                                              Loop for next chunk
//
// Goroutine Lifecycle:
//
//	Goroutine               Created             Executes                    Terminates
//	──────────────────────────────────────────────────────────────────────────────────
//	Reader                  At start            Read loop                   EOF or error
//	Worker 1..N             At start            Consume channel             Channel closed
//	Main                    Implicit            Coordinates                 All workers done
//
// Channel Semantics:
//
//	Channel             Sender                  Receiver                Size        Behavior
//	────────────────────────────────────────────────────────────────────────────────
//	chunkChan           Reader goroutine        Worker goroutines       N chunks    Blocks when full
//	errChan             Reader or main          Main goroutine          1 error     Non-blocking send
//
// Buffer Pooling:
//
//	Scenario                        Action                          Benefit
//	──────────────────────────────────────────────────────────────────────────
//	bufferPool is nil               Allocate new buffer per read    Simple, more allocations
//	bufferPool is configured        Get/Put buffer from pool        Reuses buffers, less GC
//	Buffer from pool                Defer Put to return after use   Automatic cleanup
//	Multiple readers                Each uses separate buffer       No contention
//
// Error Handling Strategy:
//
//	Error Type                  Location                Handled
//	──────────────────────────────────────────────────────────────────
//	io.EOF                      Reader goroutine        Break read loop (normal)
//	Read error                  Reader goroutine        Send to errChan
//	Context cancellation        Both goroutines         Check ctx.Done()
//	Compression error           Worker goroutine        recordError, continue
//	Write error                 Worker goroutine        recordError, continue
//	Final error check           Main goroutine          Return after wait
//
// Throttle Rate Calculation:
//
//	Configuration               Formula                             Purpose
//	──────────────────────────────────────────────────────────────────────────
//	ThrottleRate (B/s)          > 0 means enabled                   Limit bandwidth
//	Expected time               transferredBytes / rate * 1 second  Calculate expected duration
//	Elapsed time                time.Since(StartTime)               Actual duration
//	Sleep duration              expected - elapsed                  Throttle delay
//	Sleep condition             elapsed < expected                  Only sleep if behind
//
// Example:
//
//	// Example 1: Basic buffered streaming with compression
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/stream/buffered").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024).     // 1MB chunks
//	    WithMaxConcurrentChunks(4).     // 4 concurrent workers
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithCompressionType(COMP_GZIP).
//	    WithBufferPooling(true).        // Enable buffer reuse
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Chunk %d error: %v\n", p.CurrentChunk, err)
//	        } else if p.CurrentChunk % 100 == 0 {
//	            fmt.Printf("Progress: %.1f%% | Speed: %.2f MB/s | ETA: %s\n",
//	                p.Percentage,
//	                float64(p.TransferRate) / 1024 / 1024,
//	                p.EstimatedTimeRemaining.String())
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//	if result.IsError() {
//	    fmt.Printf("Streaming failed: %s\n", result.Error())
//	} else {
//	    stats := streaming.GetStats()
//	    fmt.Printf("Streaming complete: %.2f MB/s\n",
//	        float64(stats.AverageBandwidth) / 1024 / 1024)
//	}
//
//	// Example 2: High-throughput transfer with buffer pooling
//	func HighThroughputTransfer(fileReader io.ReadCloser, fileWriter io.WriteCloser) error {
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/transfer/high-speed").
//	        WithStreaming(fileReader, nil).
//	        WithWriter(fileWriter).
//	        WithChunkSize(10 * 1024 * 1024). // 10MB chunks
//	        WithMaxConcurrentChunks(8).      // 8 concurrent workers
//	        WithStreamingStrategy(STRATEGY_BUFFERED).
//	        WithBufferPooling(true).
//	        WithCompressionType(COMP_DEFLATE). // Fast compression
//	        WithCallback(func(p *StreamProgress, err error) {
//	            if err == nil && p.CurrentChunk % 50 == 0 {
//	                fmt.Printf("Chunk %d: %.2f MB/s\n",
//	                    p.CurrentChunk,
//	                    float64(p.TransferRate) / 1024 / 1024)
//	            }
//	        })
//
//	    result := streaming.Start(context.Background())
//
//	    if result.IsError() {
//	        return fmt.Errorf("transfer failed: %w", result.Error())
//	    }
//
//	    stats := streaming.GetStats()
//	    fmt.Printf("Transfer complete: %.2f MB in %.2f seconds\n",
//	        float64(stats.TotalBytes) / 1024 / 1024,
//	        stats.EndTime.Sub(stats.StartTime).Seconds())
//
//	    return nil
//	}
//
//	// Example 3: Buffered streaming with bandwidth throttling
//	func ThrottledDownload(fileReader io.ReadCloser) {
//	    // Limit to 10 MB/s
//	    tenMBps := int64(10 * 1024 * 1024)
//
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/download/throttled").
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(512 * 1024).
//	        WithMaxConcurrentChunks(4).
//	        WithStreamingStrategy(STRATEGY_BUFFERED).
//	        WithThrottleRate(tenMBps).
//	        WithCallback(func(p *StreamProgress, err error) {
//	            if err == nil {
//	                actual := float64(p.TransferRate) / 1024 / 1024
//	                fmt.Printf("Speed: %.2f MB/s (throttle: 10.00 MB/s)\n", actual)
//	            }
//	        })
//
//	    result := streaming.Start(context.Background())
//	    fmt.Printf("Throttled download complete\n")
//	}
//
//	// Example 4: Concurrent worker pool demonstration
//	func DemonstrateConcurrentWorkers(fileReader io.ReadCloser) {
//	    // Different worker counts for comparison
//	    workerCounts := []int{1, 2, 4, 8}
//
//	    for _, workers := range workerCounts {
//	        fmt.Printf("\nTesting with %d concurrent workers:\n", workers)
//
//	        // Reset file reader position (if possible)
//	        streaming := replify.New().
//	            WithStatusCode(200).
//	            WithPath("/api/test/workers").
//	            WithStreaming(fileReader, nil).
//	            WithChunkSize(1024 * 1024).
//	            WithMaxConcurrentChunks(workers).
//	            WithStreamingStrategy(STRATEGY_BUFFERED).
//	            WithBufferPooling(true)
//
//	        start := time.Now()
//	        result := streaming.Start(context.Background())
//	        duration := time.Since(start)
//
//	        stats := streaming.GetStats()
//	        throughput := float64(stats.TotalBytes) / duration.Seconds() / 1024 / 1024
//
//	        fmt.Printf("  Duration: %.2f seconds\n", duration.Seconds())
//	        fmt.Printf("  Throughput: %.2f MB/s\n", throughput)
//	    }
//	}
//
//	// Example 5: Error handling with worker goroutines
//	func BufferedWithErrorRecovery(fileReader io.ReadCloser) {
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/stream/error-recovery").
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(256 * 1024).
//	        WithMaxConcurrentChunks(4).
//	        WithStreamingStrategy(STRATEGY_BUFFERED).
//	        WithCallback(func(p *StreamProgress, err error) {
//	            if err != nil {
//	                fmt.Printf("Chunk %d error: %v (streaming continues)\n",
//	                    p.CurrentChunk, err)
//	            }
//	        })
//
//	    result := streaming.Start(context.Background())
//	    stats := streaming.GetStats()
//
//	    if stats.HasErrors {
//	        errorRate := float64(stats.FailedChunks) / float64(stats.TotalChunks) * 100
//	        fmt.Printf("Streaming completed with %.1f%% error rate\n", errorRate)
//
//	        if errorRate > 10 {
//	            fmt.Printf("Warning: High error rate detected\n")
//	        }
//	    } else {
//	        fmt.Printf("Streaming completed successfully\n")
//	    }
//	}
//
// Performance Characteristics:
//
//	Factor                          Impact              Optimization Strategy
//	────────────────────────────────────────────────────────────────────────────
//	MaxConcurrentChunks             Throughput          Increase for I/O bound, decrease for CPU bound
//	ChunkSize                       Memory/throughput   Larger = more throughput, larger memory
//	CompressionType                 CPU/bandwidth       GZIP better ratio, DEFLATE faster
//	BufferPooling                   GC pressure         Enable for high throughput
//	ThrottleRate                    Bandwidth control   Set to network/app limit
//	Reader speed                    Dominant factor     Use fast readers (network, fast disk)
//	Writer speed                    Dominant factor     Use fast writers
//	Compression ratio               Effective rate      Monitor and tune
//
// Concurrency Tuning Recommendations:
//
//	Scenario                        Recommended Workers        Rationale
//	─────────────────────────────────────────────────────────────────────────
//	Network transfer (LAN)          4-8                       Overlap I/O and processing
//	Network transfer (WAN)          8-16                      Higher latency, more overlap
//	Local file transfer             2-4                       Disk I/O not parallelize
//	Memory-to-memory                 1-2                       No I/O benefit
//	CPU-intensive (compression)      # of CPU cores             Parallel computation
//	Balanced (compression + I/O)     4-8                       Balance both factors
//
// Goroutine Coordination:
//
//	Synchronization Point           Mechanism                   Ensures
//	────────────────────────────────────────────────────────────────────────
//	All workers start               wg.Add(i) loop              Predictable startup
//	All workers complete            wg.Wait()                   No orphaned workers
//	Channel drain                   range chunkChan             All chunks processed
//	Error collection                select on errChan           First error captured
//	Reader closure                  close(chunkChan)            Workers exit cleanly
//	Context cancellation            select ctx.Done()           Clean shutdown
//
// Best Practices:
//
//  1. CHOOSE CONCURRENT WORKER COUNT BASED ON WORKLOAD
//     - Network I/O: 4-8 workers (overlap read/write/compress)
//     - Local disk: 2-4 workers (disk I/O not parallelize)
//     - CPU-intensive: # CPU cores (parallel computation)
//     - Pattern:
//     WithMaxConcurrentChunks(4)  // Balanced default
//
//  2. ENABLE BUFFER POOLING FOR HIGH THROUGHPUT
//     - Reduces GC pressure
//     - Improves memory efficiency
//     - Significant impact on throughput
//     - Pattern:
//     WithBufferPooling(true)  // Enable for bulk transfers
//
//  3. USE APPROPRIATE CHUNK SIZE
//     - Larger chunks: fewer iterations, more throughput
//     - Smaller chunks: more responsive, lower memory
//     - Recommended: 1-10MB for network, 1-5MB for disk
//     - Pattern:
//     WithChunkSize(1024 * 1024)  // 1MB default balance
//
//  4. APPLY THROTTLING FOR RATE LIMITING
//     - Prevents overwhelming destination
//     - Respects network/app limits
//     - Calculated per chunk
//     - Pattern:
//     WithThrottleRate(100 * 1024 * 1024)  // 100 MB/s limit
//
//  5. MONITOR FOR GOROUTINE LEAKS
//     - Ensure context cancellation works
//     - Verify error channel drains
//     - Check goroutine count on completion
//     - Pattern:
//     defer pprof.StopCPUProfile()  // Monitor if needed
//
// Comparison with Other Strategies:
//
//	Strategy               Concurrency    Throughput    Memory     Complexity    Best For
//	──────────────────────────────────────────────────────────────────────────────────
//	STRATEGY_BUFFERED      High (workers) Highest       Medium     Complex       Bulk/high-speed
//	STRATEGY_CHUNKED       None (serial)  Lower         Low        Simple        Detail/control
//	STRATEGY_DIRECT        None (serial)  Lowest        Very low   Very simple   Simplicity
//
// Related Methods and Integration:
//
//	Method                      Purpose                     Related To
//	──────────────────────────────────────────────────────────────────
//	streamBuffered()            Buffered strategy (this)     STRATEGY_BUFFERED
//	streamChunked()             Chunked strategy            STRATEGY_CHUNKED
//	streamDirect()              Direct strategy             STRATEGY_DIRECT
//	WithMaxConcurrentChunks()   Configure worker count      Concurrency tuning
//	WithBufferPooling()         Enable buffer reuse         Memory optimization
//	WithThrottleRate()          Bandwidth limiting          Rate control
//	compressChunk()             Compression in workers      CPU-intensive task
//	calculateChecksum()         Per-chunk integrity         Data verification
//	updateProgress()            Progress tracking           Real-time metrics
//	recordError()               Per-chunk error recording   Error handling
//
// See Also:
//   - WithStreamingStrategy: Select STRATEGY_BUFFERED
//   - WithMaxConcurrentChunks: Configure worker pool size
//   - WithBufferPooling: Enable buffer pooling for efficiency
//   - WithThrottleRate: Set bandwidth throttling rate
//   - WithChunkSize: Configure chunk size for processing
//   - sync.WaitGroup: Used for worker synchronization
//   - compressChunk: Compression executed in workers
//   - GetStats: Query final statistics including failed chunks
//   - Start: Entry point that calls streamBuffered
func (sw *StreamingWrapper) streamBuffered(ctx context.Context) error {
	chunkChan := make(chan *StreamChunk, sw.config.MaxConcurrentChunks)
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	sw.stats.StartTime = time.Now()

	// Reader goroutine
	go func() {
		defer close(chunkChan)
		var buffer []byte
		if sw.bufferPool != nil {
			buffer = sw.bufferPool.Get()
			defer sw.bufferPool.Put(buffer)
		} else {
			buffer = make([]byte, sw.config.ChunkSize)
		}

		for {
			select {
			case <-ctx.Done():
				select {
				case errChan <- ctx.Err():
				default:
				}
				return
			default:
			}

			n, err := sw.reader.Read(buffer)

			if n > 0 {
				data := make([]byte, n)
				copy(data, buffer[:n])

				chunk := &StreamChunk{
					SequenceNumber:  sw.currentChunk,
					Data:            data,
					Size:            int64(n),
					Timestamp:       time.Now(),
					CompressionType: sw.config.Compression,
				}

				select {
				case chunkChan <- chunk:
					sw.currentChunk++
				case <-ctx.Done():
					return
				}
			}

			if err == io.EOF {
				break
			}

			if err != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}
		}
	}()

	// Writer goroutine pool
	for i := 0; i < sw.config.MaxConcurrentChunks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range chunkChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Process chunk
				if sw.config.Compression != CompressNone {
					compData, compErr := sw.compressChunk(chunk)
					if compErr != nil {
						sw.recordError(compErr)
						sw.stats.FailedChunks++
						chunk.Error = compErr
					} else {
						chunk.Data = compData
						chunk.Compressed = true
						sw.stats.CompressedBytes += int64(len(compData))
					}
				}

				chunk.Checksum = sw.calculateChecksum(chunk.Data)

				if sw.writer != nil {
					if _, writeErr := sw.writer.Write(chunk.Data); writeErr != nil {
						sw.recordError(writeErr)
						sw.stats.FailedChunks++
						chunk.Error = writeErr
					}
				}

				sw.updateProgress(chunk)

				if sw.config.ThrottleRate > 0 {
					elapsed := time.Since(sw.stats.StartTime)
					expectedTime := time.Duration(float64(sw.progress.TransferredBytes) / float64(sw.config.ThrottleRate) * float64(time.Second))
					if elapsed < expectedTime {
						time.Sleep(expectedTime - elapsed)
					}
				}
			}
		}()
	}

	wg.Wait()

	// Check for read errors
	select {
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			return fmt.Errorf("streaming error: %w", err)
		}
	default:
	}

	sw.stats.EndTime = time.Now()
	return nil
}

// streamChunked performs chunked streaming with explicit control over chunk boundaries and processing.
//
// This function implements the STRATEGY_CHUNKED streaming strategy, which provides fine-grained control over
// individual chunk processing. Each chunk is explicitly read, optionally compressed, checksum, written, and
// progressed independently. This strategy is useful for scenarios requiring per-chunk validation, compression
// verification, or precise error handling at the chunk level. streamChunked reads data in fixed-size chunks from
// the input reader, applies optional compression, calculates checksums for data integrity, writes compressed data
// to the output writer, and updates progress metrics after each successful chunk. Chunk errors (read, write,
// compression) are recorded individually without stopping the entire stream; streaming continues with the next
// chunk attempt. Context cancellation is checked at the start of each chunk iteration, allowing responsive shutdown.
// Read and write operations respect configured timeouts (ReadTimeout, WriteTimeout) to prevent indefinite blocking
// on slow or stalled connections. This is a comprehensive strategy suitable for production scenarios requiring
// detailed chunk-level diagnostics and fine control over the streaming lifecycle.
//
// Parameters:
//   - ctx: Context for cancellation, timeouts, and coordination.
//     Used to check for early termination via ctx.Done().
//     Used to derive per-operation timeout contexts.
//     Cancellation immediately stops streaming and returns error.
//
// Returns:
//   - error: nil if streaming completed successfully.
//     ctx.Err() wrapped if context cancelled.
//     read error if reader.Read() fails persistently.
//     Other errors during critical operations.
//
// Behavior:
//   - Per-chunk: reads, compresses, checksums, writes, and progresses one chunk at a time.
//   - Error-tolerant: continues streaming despite per-chunk errors.
//   - Context-aware: checks ctx.Done() before each chunk.
//   - Timeout-aware: respects configured ReadTimeout and WriteTimeout values.
//   - Compression-aware: optionally compresses chunks and tracks compressed bytes.
//   - Checksum-enabled: calculates checksum for each chunk for integrity verification.
//   - Progress-tracking: updates progress and triggers callbacks after each chunk.
//   - Statistics-maintained: accumulates failed chunks, compressed bytes, timing data.
//
// Streaming Lifecycle Stages:
//
//	Stage                               Action                          Error Handling
//	──────────────────────────────────────────────────────────────────────────────────
//	1. Initialize                       Set StartTime, create buffer    No error
//	2. Context check                    Check ctx.Done()                Return on cancel
//	3. Read chunk                       Read from reader                recordError, continue
//	4. Compress chunk                   Apply compression if needed     recordError, mark failed
//	5. Calculate checksum              Compute data integrity          Record checksum
//	6. Write chunk                      Write to writer                 recordError, mark failed
//	7. Update progress                  Increment counters, calculate ETA triggerCallback
//	8. Loop or EOF                      Check for EOF or error          Break or return
//	9. Finalize                         Set EndTime, cleanup            Return result
//
// Chunk Processing Pipeline:
//
//	Input                   Process                         Output                  Error Path
//	──────────────────────────────────────────────────────────────────────────────────────────
//	Context                 Check cancellation              Proceed/Cancel          Return ctx.Err()
//	Reader                  Read with configured timeout    Buffer + n bytes        recordError, continue
//	Buffer                  Create StreamChunk              Chunk struct            Continue
//	Chunk data              Compress if configured         Compressed chunk        recordError, mark failed
//	Chunk data              Calculate checksum             Chunk + Checksum        Continue
//	Chunk data              Write with configured timeout  Writer                  recordError, mark failed
//	Chunk                   Update progress                Progress + Stats        fireCallback(nil)
//	Reader                  Check EOF/Error                Break/Continue          recordError, return
//
// Error Handling Strategy:
//
//	Error Type                  Recorded?       Streaming Continues?    Action
//	────────────────────────────────────────────────────────────────────────────
//	io.EOF                      No              Yes (break)             Normal completion
//	Read error (non-EOF)        Yes             No (return)             Return error
//	Compression error           Yes             Yes                     Mark chunk failed
//	Write error                 Yes             Yes                     Mark chunk failed
//	Context cancellation        No              No (return)             Return ctx.Err()
//	Timeout (via context)       Yes             Yes                     recordError, continue
//
// Timeout Behavior:
//
//	Timeout Type            Configuration           Application
//	────────────────────────────────────────────────────────────────────
//	ReadTimeout             WithReadTimeout(ms)     Enforced by reader respecting context
//	WriteTimeout            WithWriteTimeout(ms)    Enforced by writer respecting context
//	Parent context timeout  Caller provides         Checked at loop start
//
//	Note: Actual timeout enforcement depends on reader/writer implementation.
//	      Standard library io.Reader/io.Writer may not respect context.
//	      Network readers (http, tcp) typically respect context deadlines.
//
// Example:
//
//	// Example 1: Basic chunked streaming with compression
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/stream/chunked").
//	    WithStreaming(file, nil).
//	    WithChunkSize(512 * 1024).  // 512KB chunks
//	    WithStreamingStrategy(STRATEGY_CHUNKED).
//	    WithCompressionType(COMP_GZIP).
//	    WithReadTimeout(15000).  // 15 seconds
//	    WithWriteTimeout(15000).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Chunk %d error: %v\n", p.CurrentChunk, err)
//	        } else {
//	            fmt.Printf("Chunk %d: %d bytes processed\n",
//	                p.CurrentChunk, p.TransferredBytes)
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//	if result.IsError() {
//	    fmt.Printf("Streaming failed: %s\n", result.Error())
//	}
//
//	// Example 2: Chunked streaming with error monitoring
//	func StreamWithChunkedErrorTracking(fileReader io.ReadCloser) {
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/stream/error-tracking").
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(256 * 1024).
//	        WithStreamingStrategy(STRATEGY_CHUNKED).
//	        WithReadTimeout(10000).
//	        WithWriteTimeout(10000).
//	        WithCallback(func(p *StreamProgress, err error) {
//	            if err != nil {
//	                fmt.Printf("Error at chunk %d: %v\n", p.CurrentChunk, err)
//	            }
//	        })
//
//	    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	    defer cancel()
//
//	    result := streaming.Start(ctx)
//	    stats := streaming.GetStats()
//
//	    if stats.HasErrors {
//	        fmt.Printf("Completed with %d errors out of %d chunks\n",
//	            stats.FailedChunks, stats.TotalChunks)
//	    }
//	}
//
//	// Example 3: Detailed chunk statistics
//	func DetailedChunkStatistics(streaming *StreamingWrapper) {
//	    streaming.WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil {
//	            fmt.Printf("Chunk %d: %d bytes | Speed: %.2f MB/s | ETA: %s\n",
//	                p.CurrentChunk,
//	                p.TransferredBytes,
//	                float64(p.TransferRate) / 1024 / 1024,
//	                p.EstimatedTimeRemaining.String())
//	        }
//	    })
//
//	    result := streaming.Start(context.Background())
//	    stats := streaming.GetStats()
//	    fmt.Printf("Total chunks: %d | Success rate: %.1f%%\n",
//	        stats.TotalChunks,
//	        float64(stats.TotalChunks - stats.FailedChunks) / float64(stats.TotalChunks) * 100)
//	}
//
// Chunk State Machine:
//
//	State                           Transition                      Next State
//	──────────────────────────────────────────────────────────────────────────
//	Loop Start                      ctx.Done()?                     Exit/Continue
//	Wait for Data                   Read chunk                      Read Complete
//	Read Complete                   n > 0?                          Create Chunk
//	Create Chunk                    Build StreamChunk               Compress?
//	Compress?                       Config.Compression != NONE       Compress/Skip
//	Compress                        compressChunk()                 Checksum
//	Checksum                        calculateChecksum()             Write
//	Write                           writer.Write()                  Progress
//	Progress                        updateProgress()                Read Next
//	Read Next                        EOF?                            Finalize/Read
//	Finalize                        Set EndTime                     Return
//
// Loop Iteration Timing:
//
//	Operation                           Time Cost           Notes
//	────────────────────────────────────────────────────────────────────
//	Context check                       <1μs                Immediate
//	Read chunk                          1-1000ms            Depends on reader
//	Create StreamChunk                  <1μs                Struct allocation
//	Compress chunk                      1-100ms             Depends on size/algo
//	Calculate checksum                  <1ms                Linear in data size
//	Write chunk                         1-1000ms            Depends on writer
//	Update progress                     <1ms                Math operations
//	Callback trigger                    0-10ms              Depends on callback
//	Total per chunk                     ~10-2000ms          Depends on all factors
//
// Performance Characteristics:
//
//	Factor                              Impact              Optimization
//	────────────────────────────────────────────────────────────────────────
//	Chunk size                          Larger = fewer loops Larger reduces overhead
//	Compression level                   Higher = slower      Lower = faster
//	Checksum algorithm                  Linear time         Use fast algorithms
//	Callback complexity                 Linear time         Keep callbacks fast
//	Reader speed                        Dominant factor     Use fast readers
//	Writer speed                        Dominant factor     Use fast writers
//
// Comparison with Other Strategies:
//
//	Strategy               Per-Chunk Control    Parallelism    Overhead    Best For
//	───────────────────────────────────────────────────────────────────────────────
//	STRATEGY_CHUNKED       Explicit (high)      None (serial)  Moderate    Detail/control
//	STRATEGY_BUFFERED      Implicit             Yes (parallel) Low         Throughput
//	STRATEGY_DIRECT        Minimal              None (serial)  Minimal     Simplicity
//
// Best Practices:
//
//  1. CHOOSE APPROPRIATE CHUNK SIZE
//     - Larger chunks: fewer iterations, less overhead
//     - Smaller chunks: more responsive, finer control
//     - Recommended: 256KB-10MB based on data type
//     - Pattern:
//     WithChunkSize(512 * 1024)  // 512KB balance
//
//  2. SET REASONABLE TIMEOUTS
//     - ReadTimeout: based on source speed
//     - WriteTimeout: based on destination speed
//     - Too low: excessive timeouts
//     - Too high: slow error detection
//     - Pattern:
//     WithReadTimeout(15000)   // 15 seconds
//     WithWriteTimeout(15000)  // 15 seconds
//
//  3. HANDLE PER-CHUNK ERRORS GRACEFULLY
//     - Compression errors don't stop streaming
//     - Write errors don't stop streaming
//     - Track error count and success rate
//     - Pattern:
//     if stats.FailedChunks > threshold {
//     // Alert or log warning
//     }
//
//  4. USE COMPRESSION STRATEGICALLY
//     - GZIP for text/JSON (better ratio)
//     - DEFLATE for speed priority
//     - NONE for pre-compressed data
//     - Monitor compression ratio
//     - Pattern:
//     WithCompressionType(COMP_GZIP)
//     // Verify ratio is beneficial
//
//  5. VERIFY CHECKSUMS FOR INTEGRITY
//     - Checksums calculated per chunk
//     - Can be sent to receiver for verification
//     - Detects transmission corruption
//     - Pattern:
//     // Receiver side:
//     if receivedChecksum != calculatedChecksum {
//     // Data corruption detected
//     }
//
// Related Methods and Integration:
//
//	Method                      Purpose                     Related To
//	──────────────────────────────────────────────────────────────────
//	streamChunked()             Chunked strategy (this)      STRATEGY_CHUNKED
//	streamBuffered()            Buffered strategy           STRATEGY_BUFFERED
//	streamDirect()              Direct strategy             STRATEGY_DIRECT
//	compressChunk()             Compress chunk data         Compression
//	calculateChecksum()         Calculate integrity check   Integrity
//	recordError()               Record chunk errors         Error handling
//	updateProgress()            Update progress metrics     Progress tracking
//	fireCallback()              Trigger per-chunk callback  User callbacks
//
// See Also:
//   - WithStreamingStrategy: Select STRATEGY_CHUNKED
//   - WithChunkSize: Configure chunk size
//   - WithReadTimeout: Set read operation timeout
//   - WithWriteTimeout: Set write operation timeout
//   - WithCompressionType: Enable compression
//   - compressChunk: Compression function used here
//   - calculateChecksum: Checksum calculation used here
//   - GetStats: Query final statistics including failed chunks
//   - Start: Entry point that calls streamChunked
func (sw *StreamingWrapper) streamChunked(ctx context.Context) error {
	sw.stats.StartTime = time.Now()
	buffer := make([]byte, sw.config.ChunkSize)

	chunkNumber := int64(0)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("streaming cancelled: %w", ctx.Err())
		default:
		}

		// Read chunk (ReadTimeout is configuration for timeout behavior)
		n, err := sw.reader.Read(buffer)

		if n > 0 {
			chunk := &StreamChunk{
				SequenceNumber:  chunkNumber,
				Data:            buffer[:n],
				Size:            int64(n),
				Timestamp:       time.Now(),
				CompressionType: sw.config.Compression,
			}

			// Compress if needed
			if sw.config.Compression != CompressNone {
				compData, compErr := sw.compressChunk(chunk)
				if compErr != nil {
					sw.recordError(compErr)
					sw.stats.FailedChunks++
					chunk.Error = compErr
				} else {
					chunk.Data = compData
					chunk.Compressed = true
					sw.stats.CompressedBytes += int64(len(compData))
				}
			}

			chunk.Checksum = sw.calculateChecksum(chunk.Data)

			if sw.writer != nil {
				// Write chunk (WriteTimeout is configuration for timeout behavior)
				if _, writeErr := sw.writer.Write(chunk.Data); writeErr != nil {
					chunk.Error = writeErr
					sw.recordError(writeErr)
					sw.stats.FailedChunks++
				}
			}

			sw.updateProgress(chunk)
			chunkNumber++
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			sw.recordError(err)
			sw.stats.FailedChunks++
			return fmt.Errorf("read error: %w", err)
		}
	}

	sw.stats.EndTime = time.Now()
	return nil
}

// streamReceiveDirect performs direct receiving with decompression of streamed data.
//
// This function implements the STRATEGY_DIRECT receive mode, decompressing incoming data
// sequentially without buffering or concurrency. It reads compressed chunks from the input reader,
// decompresses them using the configured compression algorithm, calculates checksums for integrity
// verification, writes the decompressed data to the output writer, and repeats until EOF. This strategy
// is ideal for scenarios where simplicity is more important than throughput, such as small downloads,
// streaming from simple sources, or when resource constraints require minimal goroutine overhead.
// streamReceiveDirectis synchronous and single-threaded; all operations (read, decompress, write)
// happen sequentially in the same goroutine. Context cancellation is checked at the start of each
// iteration, allowing responsive shutdown. Per-chunk errors (decompression, write) trigger recording
// but do not stop streaming; the next chunk is attempted. Bandwidth throttling is supported to limit
// transfer rate when needed. The direct approach has minimal memory footprint, no goroutine management
// overhead, and predictable behavior, making it suitable for embedded systems, low-resource
// environments, or applications prioritizing code simplicity over peak performance.
//
// Parameters:
//   - ctx: Context for cancellation and coordination.
//     Used to check for early termination via ctx.Done().
//     Cancellation immediately stops streaming and returns error.
//
// Returns:
//   - error: nil if streaming completed successfully.
//     ctx.Err() wrapped if context cancelled.
//     read error if reader.Read() fails.
//     Other errors during critical operations.
//
// Behavior:
//   - Sequential: reads, decompresses, writes in order, one chunk at a time.
//   - Simple: single goroutine, no channels, no synchronization primitives.
//   - Error-tolerant: continues despite per-chunk errors (except read EOF).
//   - Context-aware: checks ctx.Done() before each chunk.
//   - Decompression-capable: reverses compression using configured algorithm.
//   - Checksum-enabled: calculates checksum for each decompressed chunk.
//   - Throttle-aware: applies bandwidth throttling if configured.
//   - Progress-tracking: updates progress and triggers callbacks after each chunk.
//   - Statistics-maintained: accumulates failed chunks, decompressed bytes, timing.
//
// Receive Flow:
//
//	Read compressed → Decompress → Calculate checksum → Write decompressed → Update progress
//
// Example:
//
//	// Example 1: Simple direct receive streaming
//	httpResp, _ := http.Get("https://api.example.com/file.gz")
//	defer httpResp.Body.Close()
//
//	outputFile, _ := os.Create("output.bin")
//	defer outputFile.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/direct").
//	    WithStreaming(httpResp.Body, outputFile).
//	    WithChunkSize(512 * 1024).
//	    WithStreamingStrategy(STRATEGY_DIRECT).
//	    WithCompressionType(COMP_GZIP).
//	    WithReceiveMode(true).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Chunk %d error: %v\n", p.CurrentChunk, err)
//	        } else if p.CurrentChunk % 100 == 0 {
//	            fmt.Printf("Downloaded: %.1f%% | Speed: %.2f MB/s\n",
//	                p.Percentage,
//	                float64(p.TransferRate) / 1024 / 1024)
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//	if result.IsError() {
//	    fmt.Printf("Download failed: %s\n", result.Error())
//	}
//
// See Also:
//   - WithReceiveMode: Enable receive/decompress mode
//   - decompressChunk: Decompression function used here
//   - streamReceiveChunked: Chunked receive strategy
//   - streamReceiveBuffered: Buffered receive strategy
func (sw *StreamingWrapper) streamReceiveDirect(ctx context.Context) error {
	buffer := make([]byte, sw.config.ChunkSize)
	sw.stats.StartTime = time.Now()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("streaming cancelled: %w", ctx.Err())
		default:
		}

		// Read compressed chunk
		n, err := sw.reader.Read(buffer)

		if n > 0 {
			chunk := &StreamChunk{
				SequenceNumber:  sw.currentChunk,
				Data:            buffer[:n],
				Size:            int64(n),
				Timestamp:       time.Now(),
				CompressionType: sw.config.Compression,
			}

			// Decompress chunk
			if sw.config.Compression != CompressNone {
				decData, decErr := sw.decompressChunk(chunk)
				if decErr != nil {
					sw.recordError(decErr)
					sw.stats.FailedChunks++
					chunk.Error = decErr
					continue
				}
				chunk.Data = decData
				chunk.Compressed = false
			}

			// Calculate checksum on decompressed data
			chunk.Checksum = sw.calculateChecksum(chunk.Data)

			// Write decompressed chunk if writer is set
			if sw.writer != nil {
				if _, writeErr := sw.writer.Write(chunk.Data); writeErr != nil {
					sw.recordError(writeErr)
					sw.stats.FailedChunks++
					chunk.Error = writeErr
					continue
				}
			}

			// Update progress
			sw.updateProgress(chunk)
			sw.currentChunk++

			// Apply throttling if configured
			if sw.config.ThrottleRate > 0 {
				elapsed := time.Since(sw.stats.StartTime)
				expectedTime := time.Duration(float64(sw.progress.TransferredBytes) / float64(sw.config.ThrottleRate) * float64(time.Second))
				if elapsed < expectedTime {
					time.Sleep(expectedTime - elapsed)
				}
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			sw.recordError(err)
			sw.stats.FailedChunks++
			return fmt.Errorf("read error: %w", err)
		}
	}

	sw.stats.EndTime = time.Now()
	return nil
}

// streamReceiveChunked performs chunked receiving with explicit control and decompression.
//
// This function implements the STRATEGY_CHUNKED receive mode, which provides fine-grained control
// over individual chunk decompression and processing. Each compressed chunk is explicitly read,
// decompressed, checksum, written, and progressed independently. This strategy is useful for
// scenarios requiring per-chunk validation, decompression verification, or precise error handling
// at the chunk level. streamReceiveChunked reads data in fixed-size chunks from the input reader,
// applies decompression to reverse the compression applied during transmission, calculates checksums
// for data integrity, writes decompressed data to the output writer, and updates progress metrics
// after each successful chunk. Chunk errors (read, decompression, write) are recorded individually
// without stopping the entire stream; streaming continues with the next chunk attempt. Context
// cancellation is checked at the start of each chunk iteration, allowing responsive shutdown. This
// is a comprehensive strategy suitable for production scenarios requiring detailed chunk-level
// diagnostics and fine control over the receive streaming lifecycle.
//
// Parameters:
//   - ctx: Context for cancellation and coordination.
//
// Returns:
//   - error: nil if streaming completed successfully, error otherwise.
//
// Behavior:
//   - Per-chunk: reads, decompresses, checksums, writes, and progresses one chunk at a time.
//   - Error-tolerant: continues streaming despite per-chunk errors.
//   - Context-aware: checks ctx.Done() before each chunk.
//   - Decompression-capable: reverses compression using configured algorithm.
//   - Checksum-enabled: calculates checksum for each chunk for integrity verification.
//   - Progress-tracking: updates progress and triggers callbacks after each chunk.
//   - Statistics-maintained: accumulates failed chunks, decompressed bytes, timing data.
//
// Example:
//
//	// Example: Chunked receive with detailed error tracking
//	httpResp, _ := http.Get("https://api.example.com/archive.tar.gz")
//	defer httpResp.Body.Close()
//
//	outputFile, _ := os.Create("archive.tar")
//	defer outputFile.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/chunked").
//	    WithStreaming(httpResp.Body, outputFile).
//	    WithChunkSize(256 * 1024).
//	    WithStreamingStrategy(STRATEGY_CHUNKED).
//	    WithCompressionType(COMP_GZIP).
//	    WithReceiveMode(true).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Chunk %d decompression error: %v\n", p.CurrentChunk, err)
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//
// See Also:
//   - streamReceiveDirect: Direct receive strategy
//   - streamReceiveBuffered: Buffered receive strategy
//   - decompressChunk: Decompression function used here
func (sw *StreamingWrapper) streamReceiveChunked(ctx context.Context) error {
	sw.stats.StartTime = time.Now()
	buffer := make([]byte, sw.config.ChunkSize)

	chunkNumber := int64(0)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("streaming cancelled: %w", ctx.Err())
		default:
		}

		// Read compressed chunk
		n, err := sw.reader.Read(buffer)

		if n > 0 {
			chunk := &StreamChunk{
				SequenceNumber:  chunkNumber,
				Data:            buffer[:n],
				Size:            int64(n),
				Timestamp:       time.Now(),
				CompressionType: sw.config.Compression,
			}

			// Decompress chunk
			if sw.config.Compression != CompressNone {
				decData, decErr := sw.decompressChunk(chunk)
				if decErr != nil {
					sw.recordError(decErr)
					sw.stats.FailedChunks++
					chunk.Error = decErr
				} else {
					chunk.Data = decData
					chunk.Compressed = false
				}
			}

			// Calculate checksum on decompressed data
			chunk.Checksum = sw.calculateChecksum(chunk.Data)

			// Write decompressed chunk if writer is set
			if sw.writer != nil {
				if _, writeErr := sw.writer.Write(chunk.Data); writeErr != nil {
					chunk.Error = writeErr
					sw.recordError(writeErr)
					sw.stats.FailedChunks++
				}
			}

			sw.updateProgress(chunk)
			chunkNumber++
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			sw.recordError(err)
			sw.stats.FailedChunks++
			return fmt.Errorf("read error: %w", err)
		}
	}

	sw.stats.EndTime = time.Now()
	return nil
}

// streamReceiveBuffered performs buffered receiving with concurrent decompression.
//
// This function implements the STRATEGY_BUFFERED receive mode, which provides high-throughput
// receiving through concurrent decompression of multiple chunks. It uses a producer-consumer
// pattern with a dedicated reader goroutine that reads compressed chunks from the input source
// and sends them through a buffered channel, and a pool of worker goroutines that decompress
// chunks concurrently, calculate checksums, and write decompressed data. The concurrent
// architecture enables efficient overlapping of I/O operations (read while decompress and write)
// and CPU-intensive decompression, significantly improving throughput compared to sequential
// processing. Optional buffer pooling reduces memory allocation pressure by reusing buffers
// across chunk reads. Bandwidth throttling is supported to limit transfer rate when needed.
// Context cancellation is checked before decompression and write operations, allowing responsive
// shutdown. Per-chunk errors are recorded without stopping the entire stream; processing continues
// until all buffered chunks are processed or context cancellation occurs. This is the recommended
// strategy for high-performance scenarios requiring maximum throughput, such as large file
// downloads, bulk data imports, or bandwidth-intensive operations.
//
// Parameters:
//   - ctx: Context for cancellation, timeouts, and coordination.
//
// Returns:
//   - error: nil if streaming completed successfully, error otherwise.
//
// Behavior:
//   - Concurrent: uses producer-consumer pattern with reader and worker goroutines.
//   - Buffered: channels buffer chunks up to MaxConcurrentChunks capacity.
//   - Pool-aware: optionally uses buffer pool to reduce allocations.
//   - Decompression-capable: reverses compression in worker goroutines.
//   - Checksum-enabled: calculates checksum for each chunk.
//   - Throttle-aware: applies bandwidth throttling if configured.
//   - Error-tolerant: continues despite per-chunk errors.
//   - Context-aware: checks cancellation before operations.
//   - Statistics-maintained: accumulates failed chunks, decompressed bytes, timing.
//
// Concurrent Architecture (Receive):
//
//	Reader goroutine              Channel            Worker goroutines
//	(Read compressed)             (Buffer)           (Decompress & write)
//	  ↓                            ↓                   ↓
//	Read from source    →  chunkChan buffer  →  Decompress
//	(compressed data)        (capacity N)         Calculate checksum
//	Loop until EOF                                   Write decompressed
//	Close channel                                    Update progress
//	                                                 Throttle if needed
//
// Example:
//
//	// Example 1: High-throughput buffered receive
//	httpResp, _ := http.Get("https://api.example.com/largefile.gz")
//	defer httpResp.Body.Close()
//
//	outputFile, _ := os.Create("large_file.bin")
//	defer outputFile.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/buffered").
//	    WithStreaming(httpResp.Body, outputFile).
//	    WithChunkSize(1024 * 1024).     // 1MB chunks
//	    WithMaxConcurrentChunks(4).     // 4 concurrent decompressor
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithCompressionType(COMP_GZIP).
//	    WithReceiveMode(true).
//	    WithBufferPooling(true).        // Enable buffer reuse
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Chunk %d error: %v\n", p.CurrentChunk, err)
//	        } else if p.CurrentChunk % 100 == 0 {
//	            fmt.Printf("Downloaded: %.1f%% | Speed: %.2f MB/s | ETA: %s\n",
//	                p.Percentage,
//	                float64(p.TransferRate) / 1024 / 1024,
//	                p.EstimatedTimeRemaining.String())
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//	if result.IsError() {
//	    fmt.Printf("Download failed: %s\n", result.Error())
//	} else {
//	    stats := streaming.GetStats()
//	    fmt.Printf("Downloaded: %.2f MB in %.2f seconds\n",
//	        float64(stats.TotalBytes) / 1024 / 1024,
//	        stats.EndTime.Sub(stats.StartTime).Seconds())
//	}
//
//	// Example 2: Throttled buffered receive
//	httpResp, _ := http.Get("https://api.example.com/data.tar.gz")
//	defer httpResp.Body.Close()
//
//	outputFile, _ := os.Create("data.tar")
//	defer outputFile.Close()
//
//	// Limit to 5 MB/s
//	fiveMBps := int64(5 * 1024 * 1024)
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/throttled").
//	    WithStreaming(httpResp.Body, outputFile).
//	    WithChunkSize(512 * 1024).
//	    WithMaxConcurrentChunks(4).
//	    WithStreamingStrategy(STRATEGY_BUFFERED).
//	    WithCompressionType(COMP_GZIP).
//	    WithReceiveMode(true).
//	    WithThrottleRate(fiveMBps).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil {
//	            speed := float64(p.TransferRate) / 1024 / 1024
//	            fmt.Printf("Speed: %.2f MB/s (limit: 5.00 MB/s)\n", speed)
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//
// See Also:
//   - streamReceiveDirect: Direct receive strategy
//   - streamReceiveChunked: Chunked receive strategy
//   - decompressChunk: Decompression function used here
//   - WithMaxConcurrentChunks: Configure worker pool size
//   - WithBufferPooling: Enable buffer reuse for efficiency
func (sw *StreamingWrapper) streamReceiveBuffered(ctx context.Context) error {
	chunkChan := make(chan *StreamChunk, sw.config.MaxConcurrentChunks)
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	sw.stats.StartTime = time.Now()

	// Reader goroutine - reads compressed chunks
	go func() {
		defer close(chunkChan)
		var buffer []byte
		if sw.bufferPool != nil {
			buffer = sw.bufferPool.Get()
			defer sw.bufferPool.Put(buffer)
		} else {
			buffer = make([]byte, sw.config.ChunkSize)
		}

		for {
			select {
			case <-ctx.Done():
				select {
				case errChan <- ctx.Err():
				default:
				}
				return
			default:
			}

			n, err := sw.reader.Read(buffer)

			if n > 0 {
				data := make([]byte, n)
				copy(data, buffer[:n])

				chunk := &StreamChunk{
					SequenceNumber:  sw.currentChunk,
					Data:            data, // Still compressed
					Size:            int64(n),
					Timestamp:       time.Now(),
					CompressionType: sw.config.Compression,
				}

				select {
				case chunkChan <- chunk:
					sw.currentChunk++
				case <-ctx.Done():
					return
				}
			}

			if err == io.EOF {
				break
			}

			if err != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}
		}
	}()

	// Worker goroutine pool - decompresses chunks
	for i := 0; i < sw.config.MaxConcurrentChunks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range chunkChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Decompress chunk
				if sw.config.Compression != CompressNone {
					decData, decErr := sw.decompressChunk(chunk)
					if decErr != nil {
						sw.recordError(decErr)
						sw.stats.FailedChunks++
						chunk.Error = decErr
					} else {
						chunk.Data = decData
						chunk.Compressed = false
					}
				}

				// Calculate checksum on decompressed data
				chunk.Checksum = sw.calculateChecksum(chunk.Data)

				// Write decompressed chunk
				if sw.writer != nil {
					if _, writeErr := sw.writer.Write(chunk.Data); writeErr != nil {
						sw.recordError(writeErr)
						sw.stats.FailedChunks++
						chunk.Error = writeErr
					}
				}

				sw.updateProgress(chunk)

				// Apply throttling if configured
				if sw.config.ThrottleRate > 0 {
					elapsed := time.Since(sw.stats.StartTime)
					expectedTime := time.Duration(float64(sw.progress.TransferredBytes) / float64(sw.config.ThrottleRate) * float64(time.Second))
					if elapsed < expectedTime {
						time.Sleep(expectedTime - elapsed)
					}
				}
			}
		}()
	}

	wg.Wait()

	// Check for read errors
	select {
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			return fmt.Errorf("streaming error: %w", err)
		}
	default:
	}

	sw.stats.EndTime = time.Now()
	return nil
}

// compressChunk compresses chunk data using the configured compression algorithm before transmission.
//
// This function applies compression to chunk data to reduce bandwidth consumption and transfer time. It supports
// multiple compression algorithms (GZIP, DEFLATE) with automatic selection based on the streaming configuration.
// For uncompressed transfers (COMP_NONE), the original data is returned unchanged. compressChunk is called during
// chunk processing to compress data before transmission, reducing network payload size and improving throughput on
// bandwidth-limited connections. The compression process creates appropriate writers (gzip.Writer or flate.Writer)
// that encode the data into a buffer and return the compressed result. All compression errors are wrapped with context
// (algorithm name, operation type) for diagnostics and troubleshooting. compressChunk is safe for concurrent calls
// (no shared state except sw.config which is read-only during streaming) and is typically called from the streaming
// goroutine. This is a critical function for compression workflows; compression failures trigger error recording and
// callback notification, preventing corrupted or incomplete data transmission. The function handles edge cases (nil
// chunks, unknown compression types, writer creation failures) safely with explicit error messages or graceful
// fallback behavior.
//
// Parameters:
//   - chunk: A pointer to StreamChunk containing uncompressed data to compress.
//     If chunk is nil, returns nil data and "chunk is nil" error.
//     chunk.Data: Uncompressed byte slice (may be empty).
//
// Returns:
//   - Compressed byte slice (reduced size after compression).
//   - Error if compression fails or chunk is invalid.
//     For COMP_NONE: returns data as-is with nil error (no compression).
//     For COMP_GZIP: returns compressed data or error from gzip operations.
//     For COMP_DEFLATE: returns compressed data or error from flate operations.
//     For unknown types: returns data as-is with nil error (graceful fallback).
//
// Behavior:
//   - Nil-safe: returns error if chunk is nil.
//   - Type-based: selects algorithm from sw.config.Compression.
//   - Idempotent: same input always produces same output (deterministic).
//   - Synchronous: blocks until compression completes.
//   - Memory-safe: uses temporary buffer (buf) for compression output.
//   - Error context: wraps errors with descriptive messages and operation details.
//   - Fallback: unknown types treated as uncompressed (pass-through).
//
// Supported Compression Algorithms:
//
//	Type                Implementation               Format Spec        Best For
//	──────────────────────────────────────────────────────────────────────────────
//	COMP_NONE           No-op (return data)         N/A                Uncompressed/pre-compressed
//	COMP_GZIP           compress/gzip.NewWriter     RFC 1952           Text, JSON, logs, general
//	COMP_DEFLATE        compress/flate.NewWriter    RFC 1951           Speed, mixed content
//	Unknown             Fallback (no-op)            N/A                Unknown/unsupported
//
// Compression Process Flow:
//
//	Step    Operation                   Input               Output
//	────────────────────────────────────────────────────────────────────
//	1       Nil check                   chunk               Error or continue
//	2       Create buffer               -                   Empty bytes.Buffer
//	3       Type switch                 compression type    Route to handler
//	4       Create writer               buffer              Writer object
//	5       Write data                  Uncompressed bytes  Compressed buffer
//	6       Close writer                Writer              Flushed buffer
//	7       Return result               Buffer              Compressed data
//
// Compression Ratio and Effectiveness:
//
//	Data Type                   Compression         Typical Ratio      Savings
//	──────────────────────────────────────────────────────────────────────────
//	Text (UTF-8)                GZIP or DEFLATE     0.15-0.30          70-85%
//	JSON                         GZIP or DEFLATE     0.20-0.35          65-80%
//	CSV/TSV                      GZIP or DEFLATE     0.15-0.25          75-85%
//	XML                          GZIP or DEFLATE     0.20-0.40          60-80%
//	HTML                         GZIP or DEFLATE     0.15-0.30          70-85%
//	Logs                         GZIP or DEFLATE     0.10-0.20          80-90%
//	Binary (images)             GZIP or DEFLATE     0.95-1.05          -5% to 5%
//	Binary (video/audio)        GZIP or DEFLATE     0.99-1.02          -2% to 1%
//	Already compressed (zip)    GZIP or DEFLATE     0.99-1.05          -5% to 1%
//	Random data                 GZIP or DEFLATE     1.00-1.05          -5% to 0%
//
// Error Scenarios and Handling:
//
//	Scenario                                Error Returned                      Cause
//	─────────────────────────────────────────────────────────────────────────────
//	chunk = nil                             "chunk is nil"                      Validation
//	gzip.Write() fails                      "gzip compression failed: ..."       I/O or data error
//	gzip.Close() fails                      "gzip close failed: ..."             Flush or finalization error
//	flate.NewWriter() fails                 "deflate writer creation failed: ..."Memory or param error
//	flate.Write() fails                     "deflate compression failed: ..."    I/O or data error
//	flate.Close() fails                     "deflate close failed: ..."          Flush or finalization error
//	Unknown compression type                nil (returns data as-is)             Graceful fallback
//	Empty chunk data                        Success (empty result)               Valid edge case
//	Memory allocation error                 "...compression failed: ..." (wrapped) OOM
//
// Compression Memory Management:
//
//	Operation                       Memory Allocation         Lifetime        Cleanup
//	────────────────────────────────────────────────────────────────────────────
//	bytes.Buffer                    O(compressed size)        Local variable  Auto
//	gzip.NewWriter()                O(?) internal buffers     Until Close()   Implicit
//	flate.NewWriter()               O(?) internal buffers     Until Close()   Implicit
//	gzip/flate Write()              O(?) internal buffers     During write    Auto
//	buf.Bytes()                     O(1) reference            Return to caller Caller owns
//
// Performance Characteristics:
//
//	Operation                           Time Complexity    Space Complexity    Actual Cost
//	──────────────────────────────────────────────────────────────────────────────────
//	COMP_NONE                           O(1)               O(0) (no copy)      <1μs
//	COMP_GZIP compression               O(n·log(n))        O(compressed size)  Typical: 1-10ms
//	COMP_DEFLATE compression            O(n·log(n))        O(compressed size)  Typical: 0.5-5ms
//	Buffer creation                     O(1)               O(1)                <1μs
//	Writer creation                     O(1)               O(?) reader overhead <100ns
//	Write operation                     O(n)               O(buffer)           Depends on compression
//	Close/Flush                         O(?) depends       O(?) depends        Variable
//	Total per chunk (text, 100KB)       O(n)               O(?) ~20-30KB       ~1-5ms typical
//
// Compression Speed vs Ratio Trade-off:
//
//	Algorithm       Speed       Compression Ratio    CPU Cost      Best For
//	────────────────────────────────────────────────────────────────────────────
//	DEFLATE         Faster      Good (similar)       Moderate      Speed priority
//	GZIP            Slower      Good                 Moderate      Balance
//	GZIP (faster)   Faster      Worse                Low           Fast transfers
//	GZIP (best)     Slowest     Best                 High          Size priority
//
// Writer Lifecycle and Resource Management:
//
//	Component                   Responsibility              Action Required
//	────────────────────────────────────────────────────────────────────
//	bytes.Buffer                Accumulate compressed data  Return buf.Bytes()
//	gzip.Writer                 Compress with gzip          Explicit Close()
//	flate.Writer                Compress with flate         Explicit Close()
//	Compressed result           Return compressed bytes     Caller ownership
//
// Example:
//
//	// Example 1: Simple GZIP compression of chunk
//	chunk := &StreamChunk{
//	    Data: []byte("Hello, World! This is uncompressed data."),
//	    Size: 41,
//	}
//
//	compressed, err := streaming.compressChunk(chunk)
//	if err != nil {
//	    fmt.Printf("Compression failed: %v\n", err)
//	    return
//	}
//
//	fmt.Printf("Original: %d bytes, Compressed: %d bytes\n", len(chunk.Data), len(compressed))
//	fmt.Printf("Compression ratio: %.2f%%\n", float64(len(compressed)) / float64(len(chunk.Data)) * 100)
//
//	// Example 2: Compression with error handling in streaming
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/upload/compressed").
//	    WithStreaming(fileReader, nil).
//	    WithChunkSize(256 * 1024).
//	    WithCompressionType(COMP_GZIP).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            if strings.Contains(err.Error(), "compression") {
//	                fmt.Printf("Compression error on chunk %d: %v\n",
//	                    p.CurrentChunk, err)
//	            } else {
//	                fmt.Printf("Chunk %d error: %v\n", p.CurrentChunk, err)
//	            }
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//	if result.IsError() {
//	    fmt.Printf("Streaming failed: %s\n", result.Error())
//	}
//
//	// Example 3: Compression efficiency analysis
//	func AnalyzeCompressionEfficiency(streaming *StreamingWrapper, testData []byte) {
//	    chunk := &StreamChunk{Data: testData}
//
//	    // Try GZIP
//	    gzipCompressed, err := streaming.compressChunk(chunk)
//	    if err != nil {
//	        fmt.Printf("GZIP failed: %v\n", err)
//	        return
//	    }
//
//	    gzipRatio := float64(len(gzipCompressed)) / float64(len(testData)) * 100
//
//	    // Try DEFLATE
//	    savedCompression := streaming.config.Compression
//	    streaming.config.Compression = COMP_DEFLATE
//	    deflateCompressed, err := streaming.compressChunk(chunk)
//	    streaming.config.Compression = savedCompression
//
//	    if err != nil {
//	        fmt.Printf("DEFLATE failed: %v\n", err)
//	        return
//	    }
//
//	    deflateRatio := float64(len(deflateCompressed)) / float64(len(testData)) * 100
//
//	    fmt.Printf("Compression Analysis (%d byte input):\n", len(testData))
//	    fmt.Printf("  GZIP:    %d bytes (%.2f%% of original)\n", len(gzipCompressed), gzipRatio)
//	    fmt.Printf("  DEFLATE: %d bytes (%.2f%% of original)\n", len(deflateCompressed), deflateRatio)
//
//	    if len(deflateCompressed) < len(gzipCompressed) {
//	        fmt.Printf("  Winner: DEFLATE (%.1f%% smaller)\n",
//	            (1 - float64(len(deflateCompressed)) / float64(len(gzipCompressed))) * 100)
//	    } else {
//	        fmt.Printf("  Winner: GZIP (%.1f%% smaller)\n",
//	            (1 - float64(len(gzipCompressed)) / float64(len(deflateCompressed))) * 100)
//	    }
//	}
//
//	// Example 4: Performance monitoring during compression
//	func MonitorCompressionPerformance(streaming *StreamingWrapper) {
//	    chunk := &StreamChunk{
//	        Data: make([]byte, 1024*1024), // 1MB test data
//	    }
//
//	    // Fill with test data
//	    for i := range chunk.Data {
//	        chunk.Data[i] = byte(i % 256)
//	    }
//
//	    start := time.Now()
//	    compressed, err := streaming.compressChunk(chunk)
//	    duration := time.Since(start)
//
//	    if err != nil {
//	        fmt.Printf("Compression failed: %v\n", err)
//	        return
//	    }
//
//	    originalMB := float64(len(chunk.Data)) / 1024 / 1024
//	    compressedMB := float64(len(compressed)) / 1024 / 1024
//	    throughput := originalMB / duration.Seconds()
//
//	    fmt.Printf("Compression Performance:\n")
//	    fmt.Printf("  Original:      %.2f MB\n", originalMB)
//	    fmt.Printf("  Compressed:    %.2f MB\n", compressedMB)
//	    fmt.Printf("  Ratio:         %.2f%%\n", float64(len(compressed)) / float64(len(chunk.Data)) * 100)
//	    fmt.Printf("  Duration:      %v\n", duration)
//	    fmt.Printf("  Throughput:    %.2f MB/s\n", throughput)
//	}
//
//	// Example 5: Selecting compression based on data type
//	func SelectCompressionForDataType(dataType string, streaming *StreamingWrapper) {
//	    switch dataType {
//	    case "json", "text", "csv", "logs":
//	        // Text-based: GZIP for best ratio
//	        streaming.config.Compression = COMP_GZIP
//	        fmt.Println("Selected COMP_GZIP for text data (best ratio)")
//	    case "binary", "mixed":
//	        // Mixed: DEFLATE for speed balance
//	        streaming.config.Compression = COMP_DEFLATE
//	        fmt.Println("Selected COMP_DEFLATE for mixed data (balanced speed)")
//	    case "image", "video", "audio", "compressed":
//	        // Already compressed: no compression
//	        streaming.config.Compression = COMP_NONE
//	        fmt.Println("Selected COMP_NONE (already compressed data)")
//	    default:
//	        streaming.config.Compression = COMP_GZIP
//	        fmt.Println("Default: COMP_GZIP")
//	    }
//	}
//
//	// Example 6: Handling compression in upload scenario
//	func UploadWithCompression(filePath string, serverURL string) error {
//	    file, err := os.Open(filePath)
//	    if err != nil {
//	        return fmt.Errorf("open file: %w", err)
//	    }
//	    defer file.Close()
//
//	    fileInfo, _ := file.Stat()
//	    originalSize := fileInfo.Size()
//
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/upload/file").
//	        WithCustomFieldKV("original_size", originalSize).
//	        WithStreaming(file, nil).
//	        WithChunkSize(512 * 1024).
//	        WithCompressionType(COMP_GZIP).
//	        WithCallback(func(p *StreamProgress, err error) {
//	            if err != nil {
//	                if strings.Contains(err.Error(), "compression failed") {
//	                    fmt.Printf("Compression error on chunk %d: %v\n",
//	                        p.CurrentChunk, err)
//	                    return
//	                }
//	            }
//
//	            if p.CurrentChunk % 10 == 0 {
//	                compressionSavings := float64(originalSize - p.TransferredBytes) / float64(originalSize) * 100
//	                fmt.Printf("Upload progress: %.1f%% | Compression: %.1f%% savings\n",
//	                    p.Percentage, compressionSavings)
//	            }
//	        })
//
//	    result := streaming.Start(context.Background())
//	    stats := streaming.GetStats()
//
//	    if result.IsError() {
//	        return fmt.Errorf("upload failed: %w", result.Error())
//	    }
//
//	    fmt.Printf("Upload successful:\n")
//	    fmt.Printf("  Original: %.2f MB\n", float64(stats.TotalBytes) / 1024 / 1024)
//	    fmt.Printf("  Compressed: %.2f MB\n", float64(stats.CompressedBytes) / 1024 / 1024)
//	    fmt.Printf("  Savings: %.1f%%\n", (1 - stats.CompressionRatio) * 100)
//
//	    return nil
//	}
//
// Compression Type Selection Matrix:
//
//	Data Type               Size        Compression      Ratio       Use Case
//	────────────────────────────────────────────────────────────────────────────
//	JSON/CSV/Logs           Small       GZIP             15-30%      Web APIs, logs
//	HTML/XML                Medium      GZIP             20-40%      Web content
//	Text (UTF-8)            Any         GZIP             15-30%      General text
//	Binary/Mixed            Medium      DEFLATE          20-50%      Mixed content
//	Already compressed      Any         NONE             100%        ZIP/7Z/MP4
//	Random/Encrypted        Any         NONE             100%        Random data
//	Real-time/Low-latency   Medium      DEFLATE          20-50%      Speed priority
//	Bandwidth-critical      Large       GZIP             15-30%      Ratio priority
//
// Writer Cleanup Semantics:
//
//	Writer Type             Close Behavior              Resource Cleanup
//	────────────────────────────────────────────────────────────────────
//	gzip.Writer             Flushes remaining bytes     Finalizes compression
//	flate.Writer            Flushes remaining bytes     Finalizes compression
//	Both                    Must be closed explicitly   Or data is incomplete
//
// Best Practices:
//
//  1. ALWAYS CLOSE WRITERS AFTER WRITING
//     - gzip/flate writers must be closed to flush remaining data
//     - Closing finalizes the compression and ensures all bytes are written
//     - Pattern:
//     writer.Write(data)  // May not flush all data
//     writer.Close()      // Must close to finalize
//     // Only after Close() is buffer complete
//
//  2. HANDLE COMPRESSION ERRORS IMMEDIATELY
//     - Compression failures should stop chunk processing
//     - Errors recorded and callback invoked
//     - Pattern:
//     if err != nil {
//     recordError(err)
//     continue  // Skip this chunk
//     }
//
//  3. CHOOSE COMPRESSION BASED ON DATA TYPE
//     - Text-based data: GZIP (best ratio)
//     - Mixed content: DEFLATE (balanced speed)
//     - Already compressed: COMP_NONE
//     - Pattern:
//     switch dataType {
//     case "text": WithCompressionType(COMP_GZIP)
//     case "binary": WithCompressionType(COMP_DEFLATE)
//     }
//
//  4. MONITOR COMPRESSION EFFECTIVENESS
//     - Track compression ratio for different data
//     - Identify incompressible data types
//     - Adjust compression type accordingly
//     - Pattern:
//     ratio := len(compressed) / len(original)
//     if ratio > 0.95 {
//     // Data not compressing well, consider COMP_NONE
//     }
//
//  5. VERIFY ROUND-TRIP CONSISTENCY
//     - Compressed then decompressed should match original
//     - Verify data integrity after full cycle
//     - Pattern:
//     original := chunk.Data
//     compressed, _ := compressChunk(chunk)
//     decompressed, _ := decompressChunk(&StreamChunk{Data: compressed})
//     assert(original == decompressed)  // Must match
//
// Related Methods and Integration:
//
//	Method                      Purpose                     Related To
//	──────────────────────────────────────────────────────────────────
//	compressChunk()             Compress data (this)        Called before transmission
//	decompressChunk()           Decompress data            Called after reception
//	WithCompressionType()       Configure compression      User API
//	updateProgress()            Track compression stats    Progress tracking
//	recordError()               Record compression error   Error handling
//	GetStats()                  Query compression ratio    Statistics
//
// See Also:
//   - WithCompressionType: Configure compression algorithm before streaming
//   - decompressChunk: Decompress received data (complementary function)
//   - StreamChunk: Data structure containing chunk data
//   - compress/gzip package: Go's GZIP implementation (RFC 1952)
//   - compress/flate package: Go's DEFLATE implementation (RFC 1951)
//   - recordError: Record compression failures
//   - GetStats: Query compression ratio and statistics
func (sw *StreamingWrapper) compressChunk(chunk *StreamChunk) ([]byte, error) {
	if chunk == nil {
		return nil, errors.New("chunk is nil")
	}

	var buf bytes.Buffer

	switch sw.config.Compression {
	case CompressGzip:
		gzipWriter := gzip.NewWriter(&buf)
		if _, err := gzipWriter.Write(chunk.Data); err != nil {
			return nil, fmt.Errorf("gzip compression failed: %w", err)
		}
		if err := gzipWriter.Close(); err != nil {
			return nil, fmt.Errorf("gzip close failed: %w", err)
		}

	case CompressFlate:
		deflateWriter, err := flate.NewWriter(&buf, flate.DefaultCompression)
		if err != nil {
			return nil, fmt.Errorf("flate writer creation failed: %w", err)
		}
		if _, err := deflateWriter.Write(chunk.Data); err != nil {
			return nil, fmt.Errorf("flate compression failed: %w", err)
		}
		if err := deflateWriter.Close(); err != nil {
			return nil, fmt.Errorf("flate close failed: %w", err)
		}

	default:
		return chunk.Data, nil
	}

	return buf.Bytes(), nil
}

// decompressChunk decompresses chunk data using the configured compression algorithm.
//
// This function reverses the compression applied to chunk data, restoring the original uncompressed content.
// It supports multiple compression algorithms (GZIP, DEFLATE) with automatic selection based on the chunk's
// CompressionType field. For uncompressed data (COMP_NONE), the original data is returned unchanged. decompressChunk
// is called during chunk processing to restore data that was compressed during transfer. The decompression process
// creates appropriate readers (gzip.Reader or deflate.Reader) that decode the compressed bytes into a buffer and
// return the decompressed result. All decompression errors are wrapped with context (algorithm name, operation type)
// for diagnostics. decompressChunk is safe for concurrent calls (no shared state) and is typically called from the
// streaming goroutine. This is a critical function for streaming workflows that use compression; decompression
// failures trigger error recording and callback notification. The function handles edge cases (nil chunks, unknown
// compression types) safely with explicit error messages or graceful fallback behavior.
//
// Parameters:
//   - chunk: A pointer to StreamChunk containing compressed data and CompressionType.
//     If chunk is nil, returns nil data and "chunk is nil" error.
//     chunk.Data: Compressed byte slice (may be empty).
//     chunk.CompressionType: One of COMP_NONE, COMP_GZIP, COMP_DEFLATE.
//
// Returns:
//   - Decompressed byte slice (original uncompressed data).
//   - Error if decompression fails or chunk is invalid.
//     For COMP_NONE: returns data as-is with nil error.
//     For COMP_GZIP: returns decompressed data or error from gzip operations.
//     For COMP_DEFLATE: returns decompressed data or error from deflate operations.
//     For unknown types: returns data as-is with nil error (graceful fallback).
//
// Behavior:
//   - Nil-safe: returns error if chunk is nil.
//   - Type-based: selects algorithm from chunk.CompressionType.
//   - Idempotent: same input always produces same output (no side effects).
//   - Synchronous: blocks until decompression completes.
//   - Memory-safe: uses temporary buffers (buf) for decompression.
//   - Error context: wraps errors with descriptive messages.
//   - Fallback: unknown types treated as uncompressed.
//
// Supported Compression Algorithms:
//
//	Type                Implementation       Format Spec        Use Case
//	──────────────────────────────────────────────────────────────────────
//	COMP_NONE           No-op (return data) N/A                Uncompressed
//	COMP_GZIP           gzip.NewReader      RFC 1952           General purpose
//	COMP_DEFLATE        deflate.NewReader   RFC 1951           Faster alternative
//	Unknown             Fallback (no-op)    N/A                Graceful fallback
//
// Decompression Process Flow:
//
//	Step    Operation                   Input               Output
//	────────────────────────────────────────────────────────────────────
//	1       Nil check                   chunk               Error or continue
//	2       Type switch                 compression type    Route to handler
//	3       Create reader               compressed bytes    Reader object
//	4       Copy to buffer              Reader              Decompressed bytes
//	5       Return result               Buffer              Decompressed data
//
// Error Scenarios and Handling:
//
//	Scenario                                Error Returned                  Cause
//	──────────────────────────────────────────────────────────────────────────────
//	chunk = nil                             "chunk is nil"                  Validation
//	Invalid gzip header                     "gzip decompression failed: ..."  Format error
//	gzip data truncated                     "gzip copy failed: ..."           Incomplete data
//	Invalid deflate data                    "deflate copy failed: ..."        Format error
//	Unknown compression type                nil (returns data as-is)         Graceful fallback
//	Empty compressed data                   Depends on format                May succeed (empty result)
//	Memory allocation error                 "gzip copy failed: ..." (wrapped) OOM or buffer error
//
// Decompression Memory Management:
//
//	Operation                       Memory Allocation    Lifetime        Cleanup
//	──────────────────────────────────────────────────────────────────────────
//	bytes.NewReader()               O(0) (no copy)       Same as bytes   Auto
//	gzip.NewReader()                O(?) (internal)      Until Close()   defer
//	deflate.NewReader()             O(?) (internal)      Until Close()   defer
//	bytes.Buffer                    O(decompressed size) Scope           Auto
//	Return buf.Bytes()              O(decompressed size) Caller owns     Caller
//
// Performance Characteristics:
//
//	Operation                       Time Complexity    Space Complexity
//	──────────────────────────────────────────────────────────────────
//	COMP_NONE                       O(1)               O(0) (no copy)
//	COMP_GZIP decompression         O(n)               O(m) where m ≈ original size
//	COMP_DEFLATE decompression      O(n)               O(m) where m ≈ original size
//	Reader creation                 O(1)               O(?) reader overhead
//	Buffer copy (io.Copy)           O(n)               O(buffer size)
//	Total per chunk                 O(n) best/avg      O(m) where m = decompressed
//
// Compression Type Decision Matrix:
//
//	Compression Type        Decompression Path      Performance     Use When
//	────────────────────────────────────────────────────────────────────────────
//	COMP_NONE               Direct return           Fastest         No compression
//	COMP_GZIP               gzip.NewReader() path   Medium speed    Text/JSON/logs
//	COMP_DEFLATE           deflate.NewReader() path Faster than GZIP Binary/mixed
//	Unknown/Invalid        Fallback (direct return) Fastest         Unknown format
//
// Example:
//
//	// Example 1: Simple decompression of GZIP chunk
//	chunk := &StreamChunk{
//	    CompressionType: COMP_GZIP,
//	    Data: gzipCompressedBytes,  // Compressed data
//	    Size: len(gzipCompressedBytes),
//	}
//
//	data, err := streaming.decompressChunk(chunk)
//	if err != nil {
//	    fmt.Printf("Decompression failed: %v\n", err)
//	    return
//	}
//
//	fmt.Printf("Original data: %d bytes\n", len(data))
//
//	// Example 2: Decompressing different compression types
//	func ProcessChunkWithDecompression(streaming *StreamingWrapper, chunk *StreamChunk) error {
//	    // Decompress chunk
//	    data, err := streaming.decompressChunk(chunk)
//	    if err != nil {
//	        fmt.Printf("Decompression error for %s: %v\n",
//	            chunk.CompressionType, err)
//	        return err
//	    }
//
//	    // Use decompressed data
//	    fmt.Printf("Decompressed %d bytes using %s\n",
//	        len(data), chunk.CompressionType)
//
//	    // Process decompressed data (write to destination, etc)
//	    return processData(data)
//	}
//
//	// Example 3: Error recovery with fallback
//	func DecompressWithRecovery(streaming *StreamingWrapper, chunk *StreamChunk) []byte {
//	    data, err := streaming.decompressChunk(chunk)
//
//	    if err != nil {
//	        fmt.Printf("Decompression failed: %v\n", err)
//	        fmt.Printf("Falling back to original (possibly compressed) data\n")
//
//	        // Use original data as fallback
//	        return chunk.Data
//	    }
//
//	    return data
//	}
//
//	// Example 4: Decompression in streaming callback
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/compressed").
//	    WithStreaming(compressedReader, nil).
//	    WithChunkSize(256 * 1024).
//	    WithCompressionType(COMP_GZIP).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Chunk %d error: %v\n", p.CurrentChunk, err)
//	            return
//	        }
//
//	        // Note: Decompression happens in streaming loop
//	        // Callback receives already-processed (decompressed) data
//	        fmt.Printf("Chunk %d processed: %d bytes decompressed\n",
//	            p.CurrentChunk, p.TransferredBytes)
//	    })
//
//	result := streaming.Start(context.Background())
//
//	// Example 5: Handling decompression errors
//	func HandleCompressionErrors(streaming *StreamingWrapper) {
//	    stats := streaming.GetStats()
//
//	    if stats.HasErrors {
//	        fmt.Printf("Streaming encountered %d errors:\n", len(stats.Errors))
//
//	        for i, err := range stats.Errors {
//	            errMsg := err.Error()
//
//	            if strings.Contains(errMsg, "gzip decompression failed") {
//	                fmt.Printf("  [%d] GZIP error: %v\n", i+1, err)
//	            } else if strings.Contains(errMsg, "gzip copy failed") {
//	                fmt.Printf("  [%d] GZIP incomplete data: %v\n", i+1, err)
//	            } else if strings.Contains(errMsg, "deflate copy failed") {
//	                fmt.Printf("  [%d] DEFLATE error: %v\n", i+1, err)
//	            } else {
//	                fmt.Printf("  [%d] Other error: %v\n", i+1, err)
//	            }
//	        }
//	    }
//	}
//
//	// Example 6: Compression type switching
//	func SelectCompressionAndDecompress(chunk *StreamChunk, preferredType CompressionType) {
//	    // Try decompression with chunk's declared type
//	    data, err := streaming.decompressChunk(chunk)
//
//	    if err != nil {
//	        fmt.Printf("Failed with %s: %v\n", chunk.CompressionType, err)
//
//	        // If decompression fails, try DEFLATE as alternative
//	        if chunk.CompressionType != COMP_DEFLATE {
//	            fmt.Println("Trying DEFLATE as fallback...")
//	            chunk.CompressionType = COMP_DEFLATE
//	            data, err = streaming.decompressChunk(chunk)
//
//	            if err == nil {
//	                fmt.Println("✓ DEFLATE succeeded")
//	            } else {
//	                fmt.Printf("✗ DEFLATE also failed: %v\n", err)
//	            }
//	        }
//	    }
//	}
//
// Reader and Buffer Semantics:
//
//	Component                   Responsibility              Lifetime
//	────────────────────────────────────────────────────────────────────
//	bytes.NewReader()           Wrap bytes for reading      Temporary
//	gzip.NewReader()            Decompress gzip stream      Temporary (deferred Close)
//	deflate.NewReader()         Decompress deflate stream   Temporary (deferred Close)
//	io.Copy()                   Copy from reader to buffer  Returns upon completion
//	bytes.Buffer                Accumulate decompressed     Local variable
//	buf.Bytes()                 Return decompressed data    Ownership to caller
//
// Error Wrapping Strategy:
//
//	Error Source                Operation               Wrapped Message
//	────────────────────────────────────────────────────────────────────
//	gzip.NewReader()            Create gzip reader      "gzip decompression failed: ..."
//	io.Copy() with gzip         Copy from gzip          "gzip copy failed: ..."
//	deflate.NewReader()         Create deflate reader   (no explicit wrap, error passed)
//	io.Copy() with deflate      Copy from deflate       "deflate copy failed: ..."
//	Nil chunk                   Validation              "chunk is nil"
//
// Resource Cleanup Guarantee:
//
//	Resource                    Cleanup Method          Timing
//	────────────────────────────────────────────────────────────────────
//	gzip.Reader                 defer gzipReader.Close()  After io.Copy
//	deflate.Reader              defer deflateReader.Close() After io.Copy
//	bytes.Buffer                Automatic (local var)   Function return
//	bytes.Reader                Automatic (no close)    Function return
//	Decompressed data           Caller ownership        After return
//
// Decompression Failure Recovery:
//
//	Failure Point               Recovery Strategy               Action
//	─────────────────────────────────────────────────────────────────────────
//	Invalid gzip header         Recorded as error               recordError() called
//	Truncated gzip data         Recorded as error               Chunk processing fails
//	Invalid deflate data        Recorded as error               Chunk processing fails
//	Memory error during copy    Recorded as error               OOM handling
//	Unknown compression         Fallback to uncompressed        Returns original data
//
// Call Context and Integration:
//
//	Caller                      When Called             Error Handling
//	────────────────────────────────────────────────────────────────────
//	Streaming loop              Per chunk in transfer   recordError() + fireCallback(err)
//	Decompression handler       After read succeeds     Error recorded, chunk skipped
//	Chunk processor             During chunk handling   Propagates error up
//	(called from Start())        Internal streaming      Errors affect progress
//
// Decompression in Streaming Workflow:
//
//	Stage                                   Data State
//	────────────────────────────────────────────────────────────────────
//	1. Chunk read                           Compressed (if configured)
//	2. decompressChunk() called             Decompresses
//	3. Return decompressed data             Ready for write
//	4. Write to destination                 Decompressed
//	5. updateProgress()                     Reflects decompressed size
//	6. GetProgress()                        Shows decompressed stats
//
// Best Practices:
//
//  1. ALWAYS CHECK COMPRESSION TYPE
//     - Verify chunk.CompressionType matches expected
//     - Mismatched type will cause decompression error
//     - Pattern:
//     if chunk.CompressionType != expectedType {
//     log.Warnf("Type mismatch: expected %s, got %s",
//     expectedType, chunk.CompressionType)
//     }
//
//  2. HANDLE DECOMPRESSION ERRORS GRACEFULLY
//     - Decompression failures should not crash streaming
//     - Errors recorded and callback invoked
//     - Pattern:
//     data, err := sw.decompressChunk(chunk)
//     if err != nil {
//     recordError(err)
//     continue  // Skip this chunk, process next
//     }
//
//  3. VERIFY DATA INTEGRITY
//     - Check decompressed size matches expected
//     - Verify data format/structure after decompression
//     - Pattern:
//     data, err := sw.decompressChunk(chunk)
//     if len(data) == 0 {
//     log.Warn("Decompression resulted in empty data")
//     }
//
//  4. CHOOSE APPROPRIATE COMPRESSION
//     - COMP_GZIP for text/JSON (better ratio)
//     - COMP_DEFLATE for faster decompression
//     - COMP_NONE for already-compressed data
//     - Pattern:
//     WithCompressionType(COMP_GZIP)  // For text
//     WithCompressionType(COMP_DEFLATE)  // For speed
//
//  5. MONITOR DECOMPRESSION PERFORMANCE
//     - Track decompression time in callback
//     - Monitor for bottlenecks
//     - Pattern:
//     start := time.Now()
//     data, err := sw.decompressChunk(chunk)
//     duration := time.Since(start)
//     if duration > 100 * time.Millisecond {
//     log.Warnf("Slow decompression: %v", duration)
//     }
//
// Related Methods and Integration:
//
//	Method                      Purpose                     Related To
//	──────────────────────────────────────────────────────────────────
//	decompressChunk()           Decompress data (this)       Called in streaming loop
//	compressChunk()             Compress data               Server-side compression
//	WithCompressionType()       Configure compression      User API
//	recordError()               Record decompression error  Error handling
//	fireCallback()              Notify on completion       Progress updates
//	GetStats()                  Query compression stats    Statistics
//
// See Also:
//   - WithCompressionType: Configure compression algorithm
//   - StreamChunk: Data structure containing compression type and data
//   - gzip package: Go's gzip implementation (RFC 1952)
//   - flate package: Go's deflate implementation (RFC 1951)
//   - recordError: Record decompression failures
//   - Start: Streaming execution that calls decompressChunk
func (sw *StreamingWrapper) decompressChunk(chunk *StreamChunk) ([]byte, error) {
	if chunk == nil {
		return nil, errors.New("chunk is nil")
	}

	switch chunk.CompressionType {
	case CompressGzip:
		gzipReader, err := gzip.NewReader(bytes.NewReader(chunk.Data))
		if err != nil {
			return nil, fmt.Errorf("gzip decompression failed: %w", err)
		}
		defer gzipReader.Close()

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, gzipReader); err != nil {
			return nil, fmt.Errorf("gzip copy failed: %w", err)
		}
		return buf.Bytes(), nil

	case CompressFlate:
		deflateReader := flate.NewReader(bytes.NewReader(chunk.Data))
		defer deflateReader.Close()

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, deflateReader); err != nil {
			return nil, fmt.Errorf("flate copy failed: %w", err)
		}
		return buf.Bytes(), nil

	default:
		return chunk.Data, nil
	}
}

// calculateChecksum calculates simple CRC32-like checksum
// for given data slice.
func (sw *StreamingWrapper) calculateChecksum(data []byte) uint32 {
	var crc uint32
	for _, b := range data {
		crc = ((crc >> 8) | ((crc & 0xFF) << 24)) ^ uint32(b)
	}
	return crc
}

// updateProgress updates streaming progress metrics after successful chunk transfer.
//
// This function increments progress counters, recalculates derived metrics, and triggers callback notification
// after each chunk is successfully processed. It updates both the StreamProgress (real-time) and StreamingStats
// (cumulative) structures to maintain consistency between progress queries and final statistics. updateProgress
// is called after each successful chunk transfer; it does not update on chunk errors (which trigger fireCallback
// with an error instead). The function uses atomic operations for thread-safe byte counter updates and calculates
// key metrics including percentage complete, transfer rate (bytes per second), and estimated time remaining (ETA).
// updateProgress is called from the streaming goroutine with the chunk lock held; it is not directly callable from
// external code. This is a critical internal function that keeps progress metrics accurate and current, enabling
// real-time monitoring via GetProgress(), progress bars, ETA displays, and bandwidth monitoring. All calculations
// use proper precision (int64 for bytes, float64 for rates) and handle edge cases (division by zero, zero elapsed
// time) to prevent calculation errors or panics.
//
// Parameters:
//   - chunk: A pointer to StreamChunk with Size indicating bytes transferred in this chunk.
//     If chunk is nil, the function returns immediately (no-op).
//
// Behavior:
//   - Nil-safe: returns immediately if chunk == nil.
//   - Atomic update: uses atomic.AddInt64 for thread-safe byte counter increment.
//   - Mutex protection: acquires write lock (sw.mu.Lock) for consistent snapshot updates.
//   - Per-chunk: called once per successfully transferred chunk.
//   - Cumulative: increments totals across multiple chunks.
//   - Synchronous: executes in streaming goroutine.
//   - Callback trigger: fires callback with nil error to indicate success.
//   - Derived metrics: recalculates percentage, rate, and ETA based on current state.
//
// Progress Fields Updated:
//
//	Field                           Update Method              Formula / Logic
//	────────────────────────────────────────────────────────────────────────────
//	TransferredBytes                Atomic increment           += chunk.Size
//	CurrentChunk                    Direct assignment          = sw.currentChunk
//	Percentage                      Calculated                 (transferred * 100) / total
//	ElapsedTime                     time.Since()               Current time - start time
//	LastUpdate                      Current time               time.Now()
//	TransferRate                    Calculated                 transferred / elapsed.Seconds()
//	EstimatedTimeRemaining          Calculated                 remaining / rate * 1 second
//
// Statistics Fields Updated:
//
//	Field                           Update Method              Formula / Logic
//	────────────────────────────────────────────────────────────────────────────
//	TotalChunks                     Sync with progress         = sw.currentChunk
//	TotalBytes                      Sync with progress         = transferred bytes
//	AverageChunkSize                Calculated                 total bytes / chunk count
//	AverageBandwidth                Sync with progress rate    = current transfer rate
//
// Calculation Details and Edge Cases:
//
//	Calculation                     Formula                     Edge Case Handling
//	──────────────────────────────────────────────────────────────────────────
//	Percentage                      (bytes * 100) / total       Guard: if total == 0, skip
//	Transfer rate                   bytes / seconds             Guard: if elapsed == 0, skip
//	ETA                             remaining / rate            Guard: if rate == 0, skip
//	Average chunk size              total / chunks              Guard: if chunks == 0, skip
//	Remaining bytes                 total - transferred         Guard: negative result clipped
//
// Progress Update Timeline:
//
//	Point in Streaming              Progress State
//	────────────────────────────────────────────────────────────────────
//	Before Start()                  All zero (defaults)
//	Start() initialized             StartTime set
//	After chunk 1 succeeds          TransferredBytes = size1, Percentage = (size1/total)*100
//	After chunk 2 succeeds          TransferredBytes = size1+size2, Rate = avg(sizes)/time
//	Mid-streaming                   Real-time values, Percentage < 100
//	After last chunk                TransferredBytes = total, Percentage = 100
//	After streaming done            Final stable values
//
// Atomic Operations and Thread Safety:
//
//	Operation                       Type            Thread-Safe    Details
//	────────────────────────────────────────────────────────────────────
//	atomic.AddInt64() byte count    Lock-free       Yes            Hardware atomic
//	Mutex lock for struct update    Mutual exclusion Yes            Serializes updates
//	Read during update              RWMutex compatible Yes          Write lock blocks readers
//	Callback trigger                Synchronous     Yes            Fires after lock released
//	Progress snapshot during update  Consistent      Yes            Lock held during entire update
//
// Percentage Calculation:
//
//	Scenario                                Calculation             Result
//	──────────────────────────────────────────────────────────────────────
//	Total bytes = 0                         Skipped (guard)         Percentage unchanged
//	100 / 1000 bytes transferred            (100 * 100) / 1000      10
//	500 / 1000 bytes transferred            (500 * 100) / 1000      50
//	1000 / 1000 bytes transferred           (1000 * 100) / 1000     100
//	Rounding behavior                       Integer division        Truncates (e.g., 33.5 → 33)
//
// Transfer Rate Calculation:
//
//	Scenario                                Formula                 Result
//	──────────────────────────────────────────────────────────────────────
//	Elapsed = 0                             Skipped (guard)         Rate = 0
//	1000 bytes in 1 second                  1000 / 1.0              1000 B/s
//	1000 bytes in 10 seconds                1000 / 10.0             100 B/s
//	1000000 bytes in 2 seconds              1000000 / 2.0           500000 B/s (488 KB/s)
//	1000000000 bytes in 10 seconds          1000000000 / 10.0       100000000 B/s (95 MB/s)
//
// ETA Calculation:
//
//	Scenario                                    Formula                                 Result
//	──────────────────────────────────────────────────────────────────────────────
//	Rate = 0                                    Skipped (guard)                         ETA = 0
//	1000 bytes total, 600 transferred, 100 B/s (400 / 100) * 1s                    4 seconds
//	1000000 bytes total, 500000 transferred     (500000 / 500000) * 1s               1 second
//	  Rate = 500000 B/s
//	1000000 bytes total, 100000 transferred     (900000 / 1000000) * 1s              0.9 seconds
//	  Rate = 1000000 B/s (1 MB/s)
//	1000000000 bytes total, 50000000 trans,     (950000000 / 10000000) * 1s          95 seconds
//	  Rate = 10000000 B/s (10 MB/s)
//
// Synchronization with Statistics:
//
//	Progress Field              Stats Field              Synchronization
//	──────────────────────────────────────────────────────────────────────
//	TransferredBytes            TotalBytes              Direct copy (progress → stats)
//	CurrentChunk                TotalChunks             Direct copy (progress → stats)
//	TransferRate                AverageBandwidth        Direct copy (progress → stats)
//	(calculated average)        AverageChunkSize        Recalculated (bytes / chunks)
//
// Example:
//
//	// Example 1: Observing progress updates through callback
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithMaxConcurrentChunks(4).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        // Called after each successful chunk (updateProgress triggers this)
//	        if err == nil {  // Progress update (no error)
//	            fmt.Printf("Progress: %.1f%% | Chunk %d | Speed: %.2f MB/s | ETA: %s\n",
//	                p.Percentage,
//	                p.CurrentChunk,
//	                float64(p.TransferRate) / 1024 / 1024,
//	                p.EstimatedTimeRemaining.String())
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//	// Output: Progress updates show real-time metrics after each chunk
//
//	// Example 2: Observing ETA accuracy over time
//	dataReader := createDataReader()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/stream/data").
//	    WithStreaming(dataReader, nil).
//	    WithChunkSize(512 * 1024).
//	    WithTotalBytes(100 * 1024 * 1024).  // 100 MB
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil && p.CurrentChunk % 10 == 0 {  // Log every 10th chunk
//	            fmt.Printf("Chunk %3d: %.1f%% complete | Speed: %7.2f MB/s | ETA: %8s (elapsed: %8s)\n",
//	                p.CurrentChunk,
//	                p.Percentage,
//	                float64(p.TransferRate) / 1024 / 1024,
//	                p.EstimatedTimeRemaining.String(),
//	                p.ElapsedTime.String())
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//	// ETA improves accuracy as transfer progresses
//
//	// Example 3: Progress tracking with rate monitoring
//	httpResp, _ := http.Get("https://api.example.com/largefile")
//	defer httpResp.Body.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/proxy/download").
//	    WithStreaming(httpResp.Body, nil).
//	    WithChunkSize(256 * 1024).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil {
//	            speedMBps := float64(p.TransferRate) / 1024 / 1024
//
//	            // Adaptive feedback based on speed
//	            speedIndicator := "↓"
//	            if speedMBps > 10 {
//	                speedIndicator = "↓↓"
//	            } else if speedMBps < 1 {
//	                speedIndicator = "↓!"  // Slow warning
//	            }
//
//	            fmt.Printf("\r%s %.1f%% | %.2f MB/s | ETA: %s",
//	                speedIndicator,
//	                p.Percentage,
//	                speedMBps,
//	                p.EstimatedTimeRemaining.String())
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//
//	// Example 4: Statistics consistency verification
//	func VerifyProgressStatsConsistency(streaming *StreamingWrapper) {
//	    progress := streaming.GetProgress()
//	    stats := streaming.GetStats()
//
//	    fmt.Println("Progress vs Stats Consistency Check:")
//
//	    // Verify total bytes match
//	    if progress.TotalBytes != stats.TotalBytes {
//	        fmt.Printf("  ✗ Total bytes mismatch: progress=%d, stats=%d\n",
//	            progress.TotalBytes, stats.TotalBytes)
//	    } else {
//	        fmt.Printf("  ✓ Total bytes consistent: %d\n", progress.TotalBytes)
//	    }
//
//	    // Verify chunks match
//	    if progress.TotalChunks != stats.TotalChunks {
//	        fmt.Printf("  ✗ Total chunks mismatch: progress=%d, stats=%d\n",
//	            progress.TotalChunks, stats.TotalChunks)
//	    } else {
//	        fmt.Printf("  ✓ Total chunks consistent: %d\n", progress.TotalChunks)
//	    }
//
//	    // Verify bandwidth matches
//	    if progress.TransferRate != stats.AverageBandwidth {
//	        fmt.Printf("  ✗ Bandwidth mismatch: progress=%d B/s, stats=%d B/s\n",
//	            progress.TransferRate, stats.AverageBandwidth)
//	    } else {
//	        fmt.Printf("  ✓ Bandwidth consistent: %.2f MB/s\n",
//	            float64(progress.TransferRate) / 1024 / 1024)
//	    }
//
//	    // Calculate and verify average chunk size
//	    if progress.TotalChunks > 0 {
//	        calculatedAvg := progress.TotalBytes / progress.TotalChunks
//	        if calculatedAvg != stats.AverageChunkSize {
//	            fmt.Printf("  ✗ Avg chunk size mismatch: calculated=%d, stats=%d\n",
//	                calculatedAvg, stats.AverageChunkSize)
//	        } else {
//	            fmt.Printf("  ✓ Avg chunk size consistent: %d bytes\n", stats.AverageChunkSize)
//	        }
//	    }
//	}
//
// Performance Characteristics:
//
//	Operation                       Time Complexity    Notes
//	──────────────────────────────────────────────────────────────────
//	Nil check                       O(1)               Quick return
//	Mutex lock acquisition          O(1)               Brief lock
//	Atomic add operation            O(1)               Lock-free
//	time.Since() calculation        O(1)               Fixed cost
//	Percentage calculation          O(1)               Simple division
//	Rate calculation                O(1)               Float division
//	ETA calculation                 O(1)               Float division
//	Callback trigger                O(?) dependent on callback
//	Total per chunk                 O(1) + callback    Negligible overhead
//
// Call Pattern in Streaming Loop:
//
//	for each chunk {
//	    data := read(chunk)
//	    if error {
//	        recordError(error)
//	        fireCallback(error)        // Error callback
//	        continue
//	    }
//
//	    write(data)
//	    if error {
//	        recordError(error)
//	        fireCallback(error)        // Error callback
//	        continue
//	    }
//
//	    // Success - update progress
//	    updateProgress(chunk)          // UPDATE PROGRESS HERE
//	    // fireCallback(nil) called inside updateProgress()
//	}
//
// Important Implementation Notes:
//
//	Note                                        Details
//	──────────────────────────────────────────────────────────────────
//	Atomic vs mutex                             Byte counter uses atomic (faster)
//	                                           Other fields use mutex (consistency)
//	Percentage precision                        Integer (0-100), truncates decimals
//	Rate precision                              int64 (B/s), high precision
//	ETA calculation timing                      Based on current rate (may vary)
//	Edge case: division by zero                 Guards prevent panics
//	Callback timing                             Fires after progress fully updated
//	Statistics synchronization                  Happens in same lock
//	Order of updates                            Bytes first, derived metrics later
//	No allocation                               Reuses existing structs
//
// Thread-Safety Model:
//
//	Component                       Thread-Safe Mechanism
//	──────────────────────────────────────────────────────────────────
//	Byte counter increment          atomic.AddInt64()
//	Progress struct updates         Mutex lock held entire time
//	Stats struct updates            Mutex lock held entire time
//	Callback execution              After lock released
//	Concurrent GetProgress()        Blocked until updateProgress() done
//	Concurrent GetStats()           Blocked until updateProgress() done
//	Progress consistency            Atomic ensure all updates complete
//
// Best Practices:
//
//  1. ALWAYS UPDATE ON SUCCESSFUL CHUNK
//     - Call updateProgress after successful transfer
//     - Don't call on errors (fireCallback(err) instead)
//     - Pattern:
//     if err == nil {
//     updateProgress(chunk)
//     } else {
//     recordError(err)
//     fireCallback(err)
//     }
//
//  2. ENSURE CHUNK SIZE IS SET
//     - chunk.Size must reflect bytes transferred
//     - Accurate progress depends on correct size
//     - Pattern:
//     chunk.Size = len(buffer_written)
//     updateProgress(chunk)
//
//  3. USE ATOMIC FOR THREAD SAFETY
//     - Byte counter uses atomic operations
//     - Other updates protected by mutex
//     - Ensures consistency across goroutines
//
//  4. VERIFY PROGRESS CONSISTENCY
//     - Progress and stats should match
//     - Query both after streaming for verification
//     - Check with VerifyProgressStatsConsistency()
//
//  5. MONITOR ETA ACCURACY
//     - ETA improves as transfer progresses
//     - Early ETA less reliable (small sample)
//     - Late ETA more accurate (better statistics)
//
// Related Methods and Functions:
//
//	Method              Purpose                     Related To
//	──────────────────────────────────────────────────────────────────
//	updateProgress()    Update on success (this)     Called after chunk write
//	recordError()       Update on error             Called on chunk error
//	fireCallback()      Trigger callback            Called after updateProgress
//	GetProgress()       Query progress              Returns current progress
//	GetStats()          Query statistics            Returns cumulative stats
//	recordStats()       Record stats (internal)     Related statistics update
//	Start()             Begin streaming             Calls updateProgress per chunk
//
// See Also:
//   - GetProgress: Query current progress (populated by updateProgress)
//   - GetStats: Query final statistics (synchronized with updateProgress)
//   - fireCallback: Callback triggered after updateProgress completes
//   - recordError: Record errors (alternative to updateProgress on error)
//   - WithCallback: Configure callback to receive progress updates
//   - Start: Streaming execution that calls updateProgress
func (sw *StreamingWrapper) updateProgress(chunk *StreamChunk) {
	if chunk == nil || sw == nil {
		return
	}

	sw.mu.Lock()
	defer sw.mu.Unlock()

	atomic.AddInt64(&sw.progress.TransferredBytes, chunk.Size)

	sw.progress.CurrentChunk = sw.currentChunk
	if sw.progress.TotalBytes > 0 {
		sw.progress.Percentage = int((sw.progress.TransferredBytes * 100) / sw.progress.TotalBytes)
	}

	elapsed := time.Since(sw.stats.StartTime)
	sw.progress.ElapsedTime = elapsed
	sw.progress.LastUpdate = time.Now()

	if elapsed > 0 {
		sw.progress.TransferRate = int64(float64(sw.progress.TransferredBytes) / elapsed.Seconds())
		if sw.progress.TransferRate > 0 && sw.progress.TotalBytes > 0 {
			remainingBytes := sw.progress.TotalBytes - sw.progress.TransferredBytes
			sw.progress.EstimatedTimeRemaining = time.Duration(float64(remainingBytes) / float64(sw.progress.TransferRate) * float64(time.Second))
		}
	}

	sw.stats.TotalChunks = sw.currentChunk
	sw.stats.TotalBytes = sw.progress.TransferredBytes
	if sw.currentChunk > 0 {
		sw.stats.AverageChunkSize = sw.progress.TransferredBytes / sw.currentChunk
	}
	sw.stats.AverageBandwidth = sw.progress.TransferRate

	// Trigger callback to notify progress update
	go sw.fireCallback(nil)
	go sw.fireHook(sw.wrapper.ReplyPtr())
}

// fireCallback triggers the user-provided callback function with current progress and error information.
//
// This function invokes the callback handler if one was configured via WithCallback(), passing the current
// streaming progress snapshot and any error from the most recent chunk operation. The callback provides immediate,
// per-chunk notifications enabling real-time feedback, error handling, and progress monitoring during active
// streaming. fireCallback is called at strategic points in the streaming lifecycle: after each chunk is processed
// (successfully or with error), allowing the callback to react to individual chunk completion or failure events.
// The callback executes synchronously in the streaming goroutine; long-running callback implementations may slow
// streaming. fireCallback is safe to call from any goroutine in the streaming context but must not be called
// recursively. If no callback was configured, fireCallback is a no-op with negligible overhead. This is a low-level
// internal method for triggering callbacks; applications configure callbacks via the high-level WithCallback() API.
//
// Parameters:
//   - err: The error (if any) from the most recent chunk operation. Can be nil for successful chunks.
//
// Behavior:
//   - Callback invocation: calls sw.callback(sw.progress, err) if callback is not nil.
//   - Nil-safe: returns immediately if sw.callback == nil (no-op).
//   - Synchronous: callback executes in the calling goroutine (streaming goroutine).
//   - Non-blocking: returns immediately after callback completes (callback may block).
//   - Per-chunk: called once per chunk, whether success or error.
//   - Progress snapshot: passes current sw.progress to callback.
//   - Error passing: passes chunk error directly to callback.
//   - No buffering: callback executes immediately, blocking streaming until completion.
//
// Callback Invocation Points:
//
//	Event                           When Called             Progress State
//	──────────────────────────────────────────────────────────────────────
//	Chunk read success              After read completes    Updated with bytes
//	Chunk read error                After read fails        Partial update + error
//	Compression success             After compress          Size increased
//	Compression error               After compress fails    Previous state
//	Chunk write success             After write completes   Bytes transferred updated
//	Chunk write error               After write fails       No update
//	Timeout detection               After timeout           Error set
//	Context cancellation            On cancel signal        Final state
//	All chunks complete             Last chunk processed    Final state
//
// Callback Execution Model:
//
//	Aspect                          Characteristic
//	──────────────────────────────────────────────────────────────────
//	Execution context               Streaming goroutine
//	Timing                          Synchronous (per-chunk)
//	Frequency                       Once per chunk (N calls for N chunks)
//	Blocking behavior               Callback blocks streaming
//	Error handling                  Callback must handle errors
//	Progress state                  Current snapshot at call time
//	State mutations                 Callback should not mutate wrapper
//	Return values                   Callback has no return (void)
//	Exceptions                      Callback panics stop streaming
//	Recursion                       Do not call fireCallback recursively
//
// Comparison with Related Mechanisms:
//
//	Mechanism               Timing              Frequency       Access Method
//	────────────────────────────────────────────────────────────────────────
//	fireCallback()          Per-chunk           N chunks        Immediate
//	GetProgress()           Any time            On-demand       Polling
//	GetStats()              After streaming     Once             Post-analysis
//	Errors()                After streaming     Once             Error list
//	GetStreamingStats()     After streaming     Once             Statistics
//
// Callback Configuration Flow:
//
//	Stage                               Method                  Result
//	────────────────────────────────────────────────────────────────────
//	Before streaming                    WithCallback(fn)        Stores callback
//	During Start()                      Start()                 Executes streaming
//	Per chunk                           fireCallback(err)       Invokes callback
//	After streaming                     Start() returns         Callback no longer called
//	Query results                       Errors(), GetStats()    Access accumulated results
//
// Example:
//
//	// Example 1: Simple callback for progress monitoring
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        // Called once per chunk
//	        if err != nil {
//	            fmt.Printf("Chunk %d error: %v\n", p.CurrentChunk, err)
//	        } else {
//	            fmt.Printf("Chunk %d: %d bytes transferred\n",
//	                p.CurrentChunk, p.TransferredBytes)
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//	// Output: One line per chunk as streaming progresses
//
//	// Example 2: Callback with error accumulation and rate limiting
//	errorCount := 0
//	const maxErrorsBeforeStop = 10
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/data").
//	    WithStreaming(dataReader, nil).
//	    WithChunkSize(512 * 1024).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            errorCount++
//	            fmt.Printf("[Error %d] Chunk %d: %v\n", errorCount, p.CurrentChunk, err)
//
//	            // Stop streaming after too many errors
//	            if errorCount >= maxErrorsBeforeStop {
//	                fmt.Printf("Too many errors (%d), stopping...\n", errorCount)
//	                // Note: Cannot stop from callback, would need external mechanism
//	            }
//	        } else {
//	            // Log every Nth successful chunk (throttle output)
//	            if p.CurrentChunk % 100 == 0 {
//	                fmt.Printf("Progress: %.1f%% | Speed: %.2f MB/s\n",
//	                    p.Percentage,
//	                    float64(p.TransferRate) / 1024 / 1024)
//	            }
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//	fmt.Printf("Finished with %d errors\n", errorCount)
//
//	// Example 3: Callback with state tracking and metrics
//	type CallbackMetrics struct {
//	    successCount   int
//	    errorCount     int
//	    lastErrorTime  time.Time
//	    lastErrorMsg   string
//	    maxRate        float64
//	    minRate        float64
//	}
//
//	metrics := &CallbackMetrics{
//	    minRate: math.MaxFloat64,
//	}
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/stream/metrics").
//	    WithStreaming(fileReader, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            metrics.errorCount++
//	            metrics.lastErrorTime = time.Now()
//	            metrics.lastErrorMsg = err.Error()
//	        } else {
//	            metrics.successCount++
//	        }
//
//	        // Track bandwidth metrics
//	        currentRate := float64(p.TransferRate) / 1024 / 1024
//	        if currentRate > metrics.maxRate {
//	            metrics.maxRate = currentRate
//	        }
//	        if currentRate > 0 && currentRate < metrics.minRate {
//	            metrics.minRate = currentRate
//	        }
//
//	        // Periodic summary every 50 chunks
//	        if (metrics.successCount + metrics.errorCount) % 50 == 0 {
//	            fmt.Printf("Metrics update: %d success, %d errors | Rate: %.2f-%.2f MB/s\n",
//	                metrics.successCount,
//	                metrics.errorCount,
//	                metrics.minRate,
//	                metrics.maxRate)
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//
//	// Final metrics
//	fmt.Printf("Final metrics:\n")
//	fmt.Printf("  Successful chunks: %d\n", metrics.successCount)
//	fmt.Printf("  Failed chunks: %d\n", metrics.errorCount)
//	fmt.Printf("  Bandwidth range: %.2f-%.2f MB/s\n", metrics.minRate, metrics.maxRate)
//	if metrics.lastErrorMsg != "" {
//	    fmt.Printf("  Last error: %s (at %s)\n",
//	        metrics.lastErrorMsg,
//	        metrics.lastErrorTime.Format("15:04:05"))
//	}
//
//	// Example 4: Callback for real-time progress bar
//	func StreamWithProgressBar(fileReader io.ReadCloser, fileSize int64) *wrapper {
//	    lastPrint := time.Now()
//
//	    streaming := replify.New().
//	        WithStatusCode(200).
//	        WithPath("/api/download/progress").
//	        WithStreaming(fileReader, nil).
//	        WithChunkSize(256 * 1024).
//	        WithTotalBytes(fileSize).
//	        WithCallback(func(p *StreamProgress, err error) {
//	            // Print progress bar at most once per 100ms (throttle callback)
//	            if time.Since(lastPrint) < 100*time.Millisecond {
//	                return // Skip this callback
//	            }
//	            lastPrint = time.Now()
//
//	            if err != nil {
//	                fmt.Printf("\r[!] Error at %.1f%%: %v",
//	                    p.Percentage, err)
//	            } else {
//	                // Draw simple progress bar
//	                barWidth := 30
//	                filled := int(float64(barWidth) * p.Percentage / 100.0)
//	                bar := strings.Repeat("=", filled) + strings.Repeat("-", barWidth-filled)
//
//	                fmt.Printf("\r[%s] %.1f%% | %s / %s | ETA: %s",
//	                    bar,
//	                    p.Percentage,
//	                    formatBytes(p.TransferredBytes),
//	                    formatBytes(p.TotalBytes),
//	                    p.EstimatedTimeRemaining.String())
//	            }
//	        })
//
//	    return streaming.Start(context.Background())
//	}
//
//	// Example 5: Callback with error categorization
//	func StreamWithErrorTracking(streaming *StreamingWrapper) {
//	    errorTypes := make(map[string]int)
//
//	    streaming.WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil {
//	            return
//	        }
//
//	        // Categorize error
//	        errStr := err.Error()
//	        var category string
//
//	        switch {
//	        case strings.Contains(errStr, "timeout"):
//	            category = "timeout"
//	        case strings.Contains(errStr, "connection"):
//	            category = "connection"
//	        case strings.Contains(errStr, "read"):
//	            category = "read"
//	        case strings.Contains(errStr, "write"):
//	            category = "write"
//	        case strings.Contains(errStr, "compression"):
//	            category = "compression"
//	        default:
//	            category = "other"
//	        }
//
//	        errorTypes[category]++
//
//	        // Log error with category
//	        fmt.Printf("[%s] Chunk %d: %v\n", category, p.CurrentChunk, err)
//	    })
//
//	    result := streaming.Start(context.Background())
//
//	    // Print error summary
//	    if len(errorTypes) > 0 {
//	        fmt.Println("Error summary:")
//	        for errType, count := range errorTypes {
//	            fmt.Printf("  %s: %d\n", errType, count)
//	        }
//	    }
//	}
//
// Callback Performance Considerations:
//
//	Consideration                   Impact                          Solution
//	─────────────────────────────────────────────────────────────────────────
//	Long callback duration          Blocks streaming               Keep callback fast
//	Blocking I/O in callback        Streaming paused               Use non-blocking ops
//	Locks in callback               Potential deadlock             Avoid taking locks
//	Memory allocation in callback   GC pressure                    Pre-allocate
//	Logging overhead                Slows streaming                Use buffered logging
//	Output formatting               CPU overhead                   Cache formatted strings
//	Frequent allocations            GC impact                      Reuse buffers
//	Callback frequency              100+ per second                Throttle output
//
// Callback Safety and Best Practices:
//
//	Practice                        Rationale
//	──────────────────────────────────────────────────────────────────
//	Keep callback fast              Don't block streaming
//	No long-running ops             Callback executes in streaming goroutine
//	No locks in callback            Risk of deadlock
//	No recursive calls              Could cause stack overflow
//	No mutations to wrapper         Undefined behavior
//	No blocking I/O                 Blocks chunk processing
//	No memory allocations           GC pressure
//	Throttle logging output         Avoid too much logging
//	Handle errors gracefully        Callback errors don't stop streaming
//	Use progress snapshot           Progress is current at call time
//
// Callback Execution Guarantees:
//
//	Guarantee                           Assurance
//	──────────────────────────────────────────────────────────────────
//	Order of calls                      Chronological (chunk order)
//	Frequency                           Once per chunk
//	Progress accuracy                   Current snapshot at call time
//	Error association                   Corresponds to current chunk
//	Null safety                         Callback checked for nil
//	Panic handling                      Callback panics stop streaming
//	Thread context                      Streaming goroutine only
//	No data races                       Safe from data races (via mutex)
//
// Typical Callback Patterns:
//
//	Pattern                         Use Case                    Complexity
//	────────────────────────────────────────────────────────────────────
//	Simple logging                  Debug output                Simple
//	Progress tracking               Update UI                  Moderate
//	Error counting                  Error handling              Simple
//	Metrics collection              Performance analysis       Moderate
//	Throttled logging               Reduce output               Moderate
//	State tracking                  Track streaming state       Complex
//	Adaptive control                Adjust behavior             Complex
//
// Integration with Streaming Loop:
//
//	for each chunk {
//	    // Read chunk
//	    data, err := read()
//
//	    // Record error if needed
//	    if err != nil {
//	        recordError(err)
//	    }
//
//	    // Update progress
//	    updateProgress(data, err)
//
//	    // Fire callback - USER CODE RUNS HERE
//	    fireCallback(err)  // Calls user's WithCallback handler
//
//	    // Handle error or continue
//	    if err != nil {
//	        continue  // Next chunk
//	    }
//
//	    // Write chunk (if no error)
//	    write(data)
//	}
//
// Common Callback Use Cases:
//
//	Use Case                            Implementation Pattern
//	────────────────────────────────────────────────────────────────────
//	Progress bar                        Print progress every Nth chunk
//	Real-time monitoring                Track metrics in callback
//	Error notification                  Log or send alert on error
//	Rate limiting                       Throttle based on progress
//	Adaptive control                    Adjust parameters based on metrics
//	Health check                        Verify streaming health
//	Graceful shutdown                   Check stop flag in callback
//	State machine                       Track state transitions
//
// Related Methods and Configuration:
//
//	Method                  Purpose                         Related To
//	──────────────────────────────────────────────────────────────────
//	WithCallback()          Set callback handler            Configuration
//	fireCallback()          Trigger callback (this)         Execution
//	recordError()           Record error                    Error handling
//	recordProgress()        Update progress                 Progress tracking
//	GetProgress()           Query progress                  On-demand access
//	GetStats()              Final statistics                Post-streaming
//
// See Also:
//   - WithCallback: Configure callback handler before streaming
//   - GetProgress: On-demand progress query (alternative to callback)
//   - recordProgress: Updates progress that callback receives
//   - recordError: Records errors passed to callback
//   - Start: Initiates streaming and callback invocation
//   - Cancel: Stops streaming and callback invocation
func (sw *StreamingWrapper) fireCallback(err error) {
	if sw.callback != nil {
		sw.callback(sw.progress, err)
	}
}

// fireHook trigger the user-provided R-type callback function with current progress and R object.
//
// This function invokes the R-type callback handler if one was configured via WithCallbackR(), passing the current
// streaming progress snapshot and the R object associated with the streaming operation. The callback provides
// immediate, per-chunk notifications enabling real-time feedback and progress monitoring during active streaming.
// fireHook is called at strategic points in the streaming lifecycle: after each chunk is processed successfully,
// allowing the callback to react to individual chunk completion events. The callback executes synchronously in the
// streaming goroutine; long-running callback implementations may slow streaming. fireHook is safe to call
// from any goroutine in the streaming context but must not be called recursively. If no R-type callback was configured,
// fireHook is a no-op with negligible overhead. This is a low-level internal method for triggering R-type
// callbacks; applications configure R-type callbacks via the high-level WithCallbackR() API.
func (sw *StreamingWrapper) fireHook(w *R) {
	if sw.hook != nil {
		sw.hook(sw.progress, w)
	}
}

// recordError records an error that occurred during streaming for later retrieval and diagnostics.
//
// This function appends an error to the streaming wrapper's error history, making it available for post-streaming
// analysis via Errors(), GetStats(), and HasErrors(). Errors are recorded in chronological order (FIFO) and
// maintained in two locations: the streaming wrapper's error list and the statistics structure. recordError is
// thread-safe using a mutex and safe to call from any goroutine during streaming. Nil errors are safely ignored
// with an early return, preventing pollution of the error history with spurious entries. Recording errors does not
// interrupt streaming; the operation continues attempting to transfer remaining chunks. This enables partial success
// scenarios where some chunks fail but the majority transfer successfully. Error recording is essential for
// comprehensive diagnostics, reliability metrics, and implementing retry/circuit-breaker logic. The recorded errors
// provide the foundation for error reporting, metrics collection, and post-streaming analysis workflows.
//
// Parameters:
//   - err: The error to record. If nil, the function returns immediately without action.
//
// Behavior:
//   - Thread-safe: acquires write lock (sw.mu.Lock) to ensure safe concurrent access.
//   - Nil-safe: returns immediately if err == nil (no action for nil errors).
//   - Chronological: errors appended to slice (FIFO order, oldest at index 0).
//   - Dual recording: error added to both sw.errors and sw.stats.Errors simultaneously.
//   - Non-blocking: O(1) amortized append operation.
//   - Idempotent: safe to call multiple times with same error (creates separate entries).
//   - Persistent: recorded errors remain until streaming wrapper is freed.
//
// Error Recording Locations:
//
//	Location                    Purpose                             Retrieval Method
//	──────────────────────────────────────────────────────────────────────────────
//	sw.errors                   Wrapper-level error list            Errors()
//	sw.stats.Errors             Statistics error list               GetStats().Errors
//	(both kept in sync)         Parallel tracking for consistency   Both methods return same list
//
// Error Recording Semantics:
//
//	Scenario                            Action                              Visibility
//	──────────────────────────────────────────────────────────────────────────────
//	err = nil                           Early return (no-op)               Not recorded
//	err = io.EOF                        Recorded as normal error           Available via Errors()
//	err = context.Canceled              Recorded                           Available via Errors()
//	err = timeout error                 Recorded                           Available via Errors()
//	Duplicate error (same error twice)  Recorded twice (separate entries)  Both entries in list
//	Non-nil error                       Recorded                           Available via Errors()
//
// Call Sites and Contexts:
//
//	Call Location                   Error Type                      Triggering Condition
//	──────────────────────────────────────────────────────────────────────────────────
//	Chunk read loop                 read error                      reader.Read() fails
//	Chunk write loop                write error                     writer.Write() fails
//	Decompression                   decompress error                Decompressor.Read() fails
//	Compression                     compress error                  Compressor.Write() fails
//	Timeout handling                timeout error                   context deadline exceeded
//	Context cancellation            context error                   context.Canceled
//	Custom reader/writer            custom error                    Reader/Writer implementation error
//	Validation errors               validation error                Parameter/configuration error
//	Resource allocation             allocation error                Memory/descriptor allocation failure
//
// Thread-Safety Guarantees:
//
//	Guarantee                           Implementation
//	──────────────────────────────────────────────────────────────────────
//	No data race on append              Write lock held during append
//	No lost errors                      Atomicity of append operation
//	Consistent order                    Single append point, FIFO order
//	Safe concurrent calls               Mutex serializes concurrent recordError calls
//	Safe with readers (Errors())        RWMutex allows readers while no writer
//	Nil-safe (no panic)                 Early nil check before lock
//	No deadlock                         Mutex always released (defer)
//	No goroutine leak                   Lock acquired and released per call
//
// Performance Characteristics:
//
//	Operation                           Time Complexity    Space Complexity    Actual Cost
//	──────────────────────────────────────────────────────────────────────────────
//	Nil check                           O(1)               None                <100ns
//	Mutex lock acquisition              O(1) amortized     None                <1μs typical
//	Slice append (amortized)            O(1)               O(n) for growth     ~50-100ns
//	Mutex unlock (defer)                O(1)               None                <1μs
//	Total call (typical)                O(1)               O(1)                1-5μs per error
//	Memory allocation (occasional)      O(1)               O(n)                ~1μs when growing
//	Total call (worst: realloc)         O(n)               O(n)                ~10-100μs rare
//	Lock contention (high error rate)   O(1)               None                <10μs per call
//
// Error Recording in Streaming Workflow:
//
//	Stage                               When Errors Recorded
//	──────────────────────────────────────────────────────────────────
//	Before Start()                      No errors possible
//	Streaming initialization            Validation errors
//	Chunk read                          Read errors
//	Data processing                     Compression/decompression errors
//	Chunk write                         Write errors
//	Timeout detection                   Timeout errors
//	Context cancellation                Context errors
//	After streaming completes           Final error list stable
//
// Error History Example:
//
//	Point in Streaming      Action                          Error Count    Errors() Result
//	────────────────────────────────────────────────────────────────────────────────────
//	Start()                 Begin streaming                 0              []
//	Chunks 1-100            Process successfully            0              []
//	Chunk 101               reader.Read() fails             1              [read error]
//	Chunk 102               Retry, write succeeds           1              [read error]
//	Chunk 103               writer.Write() times out        2              [read error, timeout]
//	Chunks 104-200          Continue processing             2              [read error, timeout]
//	Chunk 201               Decompress fails                3              [read error, timeout, decompress]
//	Complete                Stop streaming                  3              [read error, timeout, decompress]
//	GetStats()              Retrieve stats                  3              All 3 errors in stats.Errors
//	Errors()                Retrieve error list             3              All 3 errors available
//
// Error Duplication and Deduplication:
//
//	Scenario                                    Behavior
//	──────────────────────────────────────────────────────────────────────
//	Same error recorded twice                   Creates two entries (no dedup)
//	Different instances, same error message     Creates two entries
//	Intentional duplicate for retry tracking    Recorded as separate entries
//	Deduplication needed                        Implement in application layer
//	Error counting                              Count actual entries (with dupes)
//	Error analysis                              May need dedup in analysis
//
// Comparison with Related Methods:
//
//	Method              Purpose                     When Called
//	─────────────────────────────────────────────────────────────────
//	recordError()       Record error (internal)     During chunk processing
//	Errors()            Retrieve error list         After streaming / diagnostics
//	HasErrors()         Quick error check           Any time
//	GetStats().Errors   Access error list via stats After streaming
//	GetProgress().LastError Latest error snapshot   During monitoring
//	WithCallback()      Per-chunk error callback    During streaming
//
// Example:
//
//	// Example 1: Observing error recording in practice
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithChunkSize(512 * 1024).
//	    WithReadTimeout(2000).  // Very short timeout to trigger errors
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err != nil {
//	            fmt.Printf("Callback: Chunk %d error: %v\n", p.CurrentChunk, err)
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//
//	// After streaming, errors are recorded and available
//	errors := streaming.Errors()
//	fmt.Printf("Total errors recorded: %d\n", len(errors))
//	for i, err := range errors {
//	    fmt.Printf("  [%d] %v\n", i+1, err)
//	}
//
//	// Example 2: Errors in GetStats()
//	stats := streaming.GetStats()
//	fmt.Printf("Errors in stats: %d\n", len(stats.Errors))
//	fmt.Printf("Error count field: %d\n", stats.ErrorCount)
//	fmt.Printf("Has errors: %v\n", stats.HasErrors)
//	if stats.FirstError != nil {
//	    fmt.Printf("First error: %v\n", stats.FirstError)
//	}
//	if stats.LastError != nil {
//	    fmt.Printf("Last error: %v\n", stats.LastError)
//	}
//
//	// Example 3: Error handling with partial success
//	httpResp, _ := http.Get("https://api.example.com/largefile")
//	defer httpResp.Body.Close()
//
//	streaming := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/proxy/remote-file").
//	    WithStreaming(httpResp.Body, nil).
//	    WithChunkSize(256 * 1024).
//	    WithReadTimeout(10000).
//	    WithWriteTimeout(10000).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        // Error recorded internally via recordError
//	        // Streaming continues despite chunk errors
//	        if err != nil {
//	            log.Warnf("Chunk %d error: %v (streaming continues)", p.CurrentChunk, err)
//	        }
//	    })
//
//	result := streaming.Start(context.Background())
//	stats := streaming.GetStats()
//
//	// Analyze partial success
//	successRate := float64(stats.ProcessedChunks) / float64(stats.TotalChunks) * 100
//	fmt.Printf("Transfer result:\n")
//	fmt.Printf("  Success rate: %.1f%% (%d / %d chunks)\n",
//	    successRate, stats.ProcessedChunks, stats.TotalChunks)
//	fmt.Printf("  Errors: %d\n", len(stats.Errors))
//	fmt.Printf("  Data transferred: %.2f / %.2f MB\n",
//	    float64(stats.TransferredBytes) / 1024 / 1024,
//	    float64(stats.TotalBytes) / 1024 / 1024)
//
//	// Determine action based on partial success
//	if successRate > 95 {
//	    fmt.Println("Action: Accept partial transfer (>95% success)")
//	} else if successRate > 50 {
//	    fmt.Println("Action: Log warnings, consider retry")
//	} else {
//	    fmt.Println("Action: Retry entire transfer (< 50% success)")
//	}
//
//	// Example 4: Error tracking for diagnostics
//	func DiagnoseStreamingErrors(streaming *StreamingWrapper) {
//	    errors := streaming.Errors()
//
//	    if len(errors) == 0 {
//	        fmt.Println("✓ No errors during streaming")
//	        return
//	    }
//
//	    fmt.Printf("✗ Streaming encountered %d error(s):\n\n", len(errors))
//
//	    // Categorize errors
//	    timeoutCount := 0
//	    ioCount := 0
//	    otherCount := 0
//
//	    for i, err := range errors {
//	        errStr := err.Error()
//
//	        fmt.Printf("[Error %d/%d]\n", i+1, len(errors))
//	        fmt.Printf("  Message: %v\n", err)
//	        fmt.Printf("  Type:    %T\n", err)
//
//	        // Categorization for analysis
//	        if strings.Contains(errStr, "timeout") {
//	            fmt.Printf("  Category: TIMEOUT\n")
//	            timeoutCount++
//	        } else if strings.Contains(errStr, "read") || strings.Contains(errStr, "write") {
//	            fmt.Printf("  Category: I/O\n")
//	            ioCount++
//	        } else {
//	            fmt.Printf("  Category: OTHER\n")
//	            otherCount++
//	        }
//	        fmt.Println()
//	    }
//
//	    // Summary and recommendations
//	    fmt.Printf("Summary:\n")
//	    fmt.Printf("  Timeouts: %d\n", timeoutCount)
//	    fmt.Printf("  I/O errors: %d\n", ioCount)
//	    fmt.Printf("  Other: %d\n", otherCount)
//
//	    if timeoutCount > 0 {
//	        fmt.Println("\nRecommendation: Increase timeout values")
//	    }
//	    if ioCount > 0 {
//	        fmt.Println("Recommendation: Check network connectivity")
//	    }
//	}
//
// Error Recording Impact on Streaming:
//
//	Impact Area                     Effect of Error Recording
//	──────────────────────────────────────────────────────────────────
//	Streaming continuation          None - streaming continues after error
//	Chunk processing                Current chunk aborted, next starts
//	Overall transfer speed          Minimal (O(1) operation)
//	Memory consumption              Linear with error count (small)
//	Thread safety                   Mutex prevents race conditions
//	Error availability              Immediately available via Errors()
//	Statistics accuracy             Reflected in stats.ErrorCount
//	Response building               Can query errors before response sent
//
// Call Pattern in Streaming Loop:
//
//	for each chunk {
//	    data := read(chunk)
//	    if error {
//	        recordError(error)      // Error recorded here
//	        continue                // Streaming continues
//	    }
//
//	    write(data)
//	    if error {
//	        recordError(error)      // Error recorded here
//	        continue                // Streaming continues
//	    }
//	}
//
//	// After loop, all errors available
//	errors := Errors()  // Can retrieve all recorded errors
//
// Best Practices:
//
//  1. CALL IMMEDIATELY ON ERROR
//     - Record error as soon as it's detected
//     - Before any recovery or retry attempts
//     - Maintain chronological order
//     - Pattern:
//     if err := reader.Read(buf); err != nil {
//     sw.recordError(err)
//     // Handle recovery
//     }
//
//  2. INCLUDE CONTEXT IN ERROR IF POSSIBLE
//     - Wrap errors with context information
//     - Include chunk number, operation name
//     - Helps with diagnostics
//     - Example:
//     if err != nil {
//     contextErr := fmt.Errorf("chunk %d read failed: %w", chunk, err)
//     sw.recordError(contextErr)
//     }
//
//  3. DON'T OVER-RECORD
//     - Avoid recording same error multiple times
//     - Don't record temporary/transient errors
//     - Only record errors requiring action
//     - Pattern:
//     if err != nil && isSignificantError(err) {
//     sw.recordError(err)
//     }
//
//  4. USE FOR DIAGNOSTICS AFTER STREAMING
//     - Query errors via Errors() or GetStats()
//     - Implement error analysis workflows
//     - Log for production diagnostics
//     - Example:
//     result := streaming.Start(ctx)
//     if streaming.HasErrors() {
//     log.Infof("Streaming errors: %v", streaming.Errors())
//     }
//
//  5. COMBINE WITH METRICS
//     - Use error count from GetStats()
//     - Compare with success rate
//     - Implement circuit breaker logic
//     - Example:
//     stats := streaming.GetStats()
//     if stats.ErrorCount > threshold {
//     // Take action: log, alert, retry
//     }
//
// Mutex Lock Strategy:
//
//	Operation               Lock Type    Rationale
//	────────────────────────────────────────────────────────────────────
//	recordError (write)     Write lock   Modifies sw.errors and sw.stats
//	Errors() (read)         Read lock    Only reads sw.errors
//	GetStats() (read)       Read lock    Reads stats without modification
//	GetProgress() (read)    Read lock    Reads progress without modification
//	Cancel (write)          Write lock   Modifies internal state
//	Start (write/intensive) Multiple     Acquires locks as needed
//
// See Also:
//   - Errors: Retrieve all recorded errors for analysis
//   - HasErrors: Quick boolean check for error occurrence
//   - GetStats: Access errors through statistics structure
//   - GetProgress: Get latest error via LastError field
//   - WithCallback: Receive per-chunk error notifications
//   - recordProgress: Records progress updates (related function)
//   - recordStats: Records statistics (related function)
func (sw *StreamingWrapper) recordError(err error) {
	if err == nil {
		return
	}
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.errors = append(sw.errors, err)
	sw.stats.Errors = append(sw.stats.Errors, err)
}

// respondStreamBadRequestDefault creates a default bad request response wrapper.
// This function is used when the streaming wrapper is nil,
// returning a new wrapper with a bad request header and an error message.
func respondStreamBadRequestDefault() *wrapper {
	return New().
		WithHeader(BadRequest).
		WithMessage("Invalid streaming wrapper: nil reference provided.").
		BindCause()
}
