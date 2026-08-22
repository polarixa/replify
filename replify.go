package replify

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"reflect"
	"time"

	"github.com/polarixa/replify/pkg/coll"
	"github.com/polarixa/replify/pkg/common"
	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/encoding"
	"github.com/polarixa/replify/pkg/fj"
	"github.com/polarixa/replify/pkg/hashy"
	"github.com/polarixa/replify/pkg/strchain"
	"github.com/polarixa/replify/pkg/strutil"
	"github.com/polarixa/replify/pkg/translit"
)

// Available checks whether the [wrapper] instance is non-nil.
//
// This function ensures that the [wrapper] object exists and is not nil.
// It serves as a safety check to avoid null pointer dereferences when accessing the instance's fields or methods.
//
// Returns:
//   - A boolean value indicating whether the [wrapper] instance is non-nil:
//   - `true` if the [wrapper] instance is non-nil.
//   - `false` if the [wrapper] instance is nil.
func (w *wrapper) Available() bool {
	return w != nil
}

// Error retrieves the error associated with the [wrapper] instance.
//
// This function returns the `errors` field of the [wrapper], which contains
// any errors encountered during the operation of the [wrapper].
//
// Returns:
//   - An error object, or `nil` if no errors are present.
func (w *wrapper) Error() string {
	if !w.Available() {
		return ""
	}
	cause := w.Cause()
	if cause == nil {
		return ""
	}
	return cause.Error()
}

// Cause traverses the error chain and returns the underlying cause of the error
// associated with the [wrapper] instance.
//
// This function checks if the error stored in the [wrapper] is itself another
// [wrapper] instance. If so, it recursively calls `Cause` on the inner error
// to find the ultimate cause. Otherwise, it returns the current error.
//
// Returns:
//   - The underlying cause of the error, which can be another error or the original error.
func (w *wrapper) Cause() error {
	// prevent nil pointer dereference
	// then just leave it as empty string
	base := errors.New("")
	if !w.Available() {
		return base
	}
	w.autoAdjust() // ensure auto-adjustment of the error chain before retrieving the cause
	if w.errors == nil {
		return base
	}

	// Traverse through wrapped errors.
	// We will use Unwrap() method to unwrap errors instead of checking for *wrapper explicitly.
	// This way, we can traverse to the innermost cause regardless of error type.
	visited := make(map[error]bool)
	cause := w.errors
	for cause != nil && !visited[cause] {
		visited[cause] = true
		if err, ok := cause.(interface{ Cause() error }); ok {
			next := err.Cause()
			if next == cause { // Prevent self-reference
				break
			}
			cause = next
		} else {
			break
		}
	}
	return cause
}

// StatusCode retrieves the HTTP status code associated with the [wrapper] instance.
//
// This function returns the `statusCode` field of the [wrapper], which represents
// the HTTP status code for the response, indicating the outcome of the request.
//
// Returns:
//   - An integer representing the HTTP status code.
func (w *wrapper) StatusCode() int {
	if !w.Available() {
		return 0
	}
	if w.IsHeaderPresent() && w.statusCode <= 0 {
		w.statusCode = w.header.Code()
	}
	return w.statusCode
}

// StatusText returns a human-readable string representation of the HTTP status.
//
// This function combines the status code with its associated status text, which
// is retrieved using the `http.StatusText` function from the `net/http` package.
// The returned string follows the format "statusCode (statusText)".
//
// For example, if the status code is 200, the function will return "200 (OK)".
// If the status code is 404, it will return "404 (Not Found)".
//
// Returns:
//   - A string formatted as "statusCode (statusText)", where `statusCode` is the
//     numeric HTTP status code and `statusText` is the corresponding textual description.
func (w *wrapper) StatusText() string {
	return fmt.Sprintf("%d (%s)", w.StatusCode(), http.StatusText(w.StatusCode()))
}

// Message retrieves the message associated with the [wrapper] instance.
//
// This function returns the `message` field of the [wrapper], which typically
// provides additional context or a description of the operation's outcome.
//
// Returns:
//   - A string representing the message.
func (w *wrapper) Message() string {
	if !w.Available() {
		return ""
	}
	return w.message
}

// Total retrieves the total number of items associated with the [wrapper] instance.
//
// This function returns the `total` field of the [wrapper], which indicates
// the total number of items available, often used in paginated responses.
//
// Returns:
//   - An integer representing the total number of items.
func (w *wrapper) Total() int {
	if !w.Available() {
		return 0
	}
	return w.total
}

// Body retrieves the body data associated with the [wrapper] instance.
//
// This function returns the `data` field of the [wrapper], which contains
// the primary data payload of the response.
//
// Returns:
//   - The body data (of any type), or `nil` if no body data is present.
func (w *wrapper) Body() any {
	if !w.Available() {
		return nil
	}
	return w.data
}

// BodyString retrieves the body data as a string from the [wrapper] instance.
//
// This function checks if the [wrapper] instance is available and then attempts to convert the body data to a string.
// If the conversion is successful, it returns the string representation of the body data. If the conversion fails or if the [wrapper] instance is not available, it returns an empty string.
// This is useful for cases where the body data is expected to be a string or when a string representation of the body is needed for logging or debugging purposes.
//
// Returns:
//   - A string representation of the body data, or an empty string if the [wrapper] instance is not available or if the body data cannot be converted to a string.
func (w *wrapper) BodyString() string {
	if !w.Available() {
		return ""
	}
	return conv.StringOrDefault(safeCastValueSupervised(w.data), "<empty>")
}

// CompressSafe compresses the body data if it exceeds a specified threshold.
//
// This function checks if the [wrapper] instance is available and if the body data
// exceeds the specified threshold for compression. If the body data is larger than
// the threshold, it compresses the data using gzip and updates the body with the
// compressed data. It also adds debugging information about the compression process,
// including the original and compressed sizes.
// If the threshold is not specified or is less than or equal to zero, it defaults to 1024 bytes (1KB).
// It also removes any empty debugging fields to clean up the response.
// Parameters:
//   - `threshold`: An integer representing the size threshold for compression.
//     If the body data size exceeds this threshold, it will be compressed.
//
// Returns:
//   - A pointer to the [wrapper] instance, allowing for method chaining.
//
// If the [wrapper] is not available, it returns the original instance without modifications.
func (w *wrapper) CompressSafe(threshold int) *wrapper {
	if !w.Available() {
		return w
	}
	if threshold <= 0 {
		threshold = 1024 // 1KB threshold for compression
	}

	var originalSize int
	if s, ok := w.data.(string); ok {
		originalSize = len(s)
	} else if w.data != nil {
		originalSize = calculateSize(w.data)
	}

	// If the body data size is less than or equal to the threshold, return the original instance.
	// Otherwise, compress the body data and update the instance with the compressed data.
	if originalSize <= threshold {
		return w
	}

	// Compress the body data and update the instance with the compressed data.
	// If the compression fails, return the original instance without modifications.
	compressed := compress(w.data)
	if strutil.IsEmpty(compressed) {
		return w // compression failed, leave body unchanged
	}

	// Update the instance with the compressed data and debugging information.
	w.
		WithBody(compressed).
		WithDebuggingKV("compression", "gzip").
		WithDebuggingKV("original_size", originalSize).
		WithDebuggingKV("compressed_size", len(compressed))
	return w
}

// DecompressSafe decompresses the body data if it is compressed.
//
// This function checks if the [wrapper] instance is available and if the body data
// is compressed. If the body data is compressed, it decompresses the data using gzip
// and updates the instance with the decompressed data. It also adds debugging information
// about the decompression process, including the original and decompressed sizes.
// If the body data is not compressed, it returns the original instance without modifications.
//
// Returns:
//   - A pointer to the [wrapper] instance, allowing for method chaining.
//
// If the [wrapper] is not available, it returns the original instance without modifications.
func (w *wrapper) DecompressSafe() *wrapper {
	if !w.Available() {
		return w
	}
	if s, ok := w.data.(string); ok {
		originalSize := len(s)
		w.data = decompress(s)
		decompressed, _ := w.data.(string)
		// Update the instance with the decompressed data and debugging information.
		w.
			WithBody(w.data).
			WithDebuggingKV("decompression", "gzip").
			WithDebuggingKV("original_size", originalSize).
			WithDebuggingKV("decompressed_size", len(decompressed))
	}

	return w
}

// Stream retrieves a channel that streams the body data of the [wrapper] instance.
//
// This function checks if the body data is present and, if so, streams the data
// in chunks. It creates a buffered channel to hold the streamed data, allowing
// for asynchronous processing of the response body.
// If the body is not present, it returns an empty channel.
// The streaming is done in a separate goroutine to avoid blocking the main execution flow.
// The body data is chunked into smaller parts using the `Chunk` function, which
// splits the response data into manageable segments for efficient streaming.
//
// Returns:
//   - A channel of byte slices that streams the body data.
//   - An empty channel if the body data is not present.
//
// This is useful for handling large responses in a memory-efficient manner,
// allowing the consumer to process each chunk as it becomes available.
// Note: The channel is closed automatically when the streaming is complete.
// If the body is not present, it returns an empty channel.
func (w *wrapper) Stream() <-chan []byte {
	ch := make(chan []byte, 1)
	if !w.IsBodyPresent() {
		return ch
	}
	go func() {
		defer close(ch)
		// Chunk the response data into smaller parts.
		// This is useful for streaming large responses in smaller segments.
		// We will use the Chunk function to split the response data into manageable chunks.
		chunks := chunk(w.Respond())
		for _, chunk := range chunks {
			ch <- chunk
		}
	}()
	return ch
}

// Debugging retrieves the debugging information from the [wrapper] instance.
//
// This function checks if the [wrapper] instance is available (non-nil) before returning
// the value of the `debug` field. If the [wrapper] is not available, it returns an
// empty map to ensure safe usage.
//
// Returns:
//   - A `map[string]interface{}` containing the debugging information.
//   - An empty map if the [wrapper] instance is not available.
func (w *wrapper) Debugging() map[string]any {
	if !w.Available() {
		return nil
	}
	return w.debug
}

// JSONDebugging retrieves the debugging information from the [wrapper] instance as a JSON string.
//
// This function checks if the [wrapper] instance is available (non-nil) before returning
// the value of the `debug` field as a JSON string. If the [wrapper] is not available, it returns an
// empty string to ensure safe usage.
//
// Returns:
//   - A `string` containing the debugging information as a JSON string.
//   - An empty string if the [wrapper] instance is not available.
func (w *wrapper) JSONDebugging() string {
	if !w.Available() {
		return ""
	}
	return jsonpass(w.debug)
}

// OnDebugging retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `nil` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//
// Returns:
//   - The value associated with the specified debugging key if it exists.
//   - `nil` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) OnDebugging(key string) any {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return nil
	}
	return w.debug[key]
}

// DebuggingBool retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A boolean value to return if the key is not available.
//
// Returns:
//   - The boolean value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingBool(key string, defaultValue bool) bool {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.BoolOrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingBool retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A boolean value to return if the key is not available.
//
// Returns:
//   - The boolean value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingBool(path string, defaultValue bool) bool {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Bool()
	}
	return defaultValue
}

// DebuggingString retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A string value to return if the key is not available.
//
// Returns:
//   - The string value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingString(key string, defaultValue string) string {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.StringOrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingString retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A string value to return if the key is not available.
//
// Returns:
//   - The string value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingString(path string, defaultValue string) string {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.String()
	}
	return defaultValue
}

// DebuggingTime retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A time.Time value to return if the key is not available.
//
// Returns:
//   - The time.Time value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingTime(key string, defaultValue time.Time) time.Time {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.TimeOrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingTime retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A time.Time value to return if the key is not available.
//
// Returns:
//   - The time.Time value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingTime(path string, defaultValue time.Time) time.Time {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Time()
	}
	return defaultValue
}

// DebuggingInt retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: An integer value to return if the key is not available.
//
// Returns:
//   - The integer value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingInt(key string, defaultValue int) int {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.IntOrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingInt retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: An integer value to return if the key is not available.
//
// Returns:
//   - The integer value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingInt(path string, defaultValue int) int {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Int()
	}
	return defaultValue
}

// DebuggingInt8 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: An int8 value to return if the key is not available.
//
// Returns:
//   - The int8 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingInt8(key string, defaultValue int8) int8 {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.Int8OrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingInt8 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: An int8 value to return if the key is not available.
//
// Returns:
//   - The int8 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingInt8(path string, defaultValue int8) int8 {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Int8()
	}
	return defaultValue
}

// DebuggingInt16 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: An int16 value to return if the key is not available.
//
// Returns:
//   - The int16 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingInt16(key string, defaultValue int16) int16 {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.Int16OrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingInt16 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: An int16 value to return if the key is not available.
//
// Returns:
//   - The int16 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingInt16(path string, defaultValue int16) int16 {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Int16()
	}
	return defaultValue
}

// DebuggingInt32 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: An int32 value to return if the key is not available.
//
// Returns:
//   - The int32 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingInt32(key string, defaultValue int32) int32 {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.Int32OrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingInt32 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: An int32 value to return if the key is not available.
//
// Returns:
//   - The int32 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingInt32(path string, defaultValue int32) int32 {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Int32()
	}
	return defaultValue
}

// DebuggingInt64 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: An int64 value to return if the key is not available.
//
// Returns:
//   - The int64 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingInt64(key string, defaultValue int64) int64 {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.Int64OrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingInt64 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: An int64 value to return if the key is not available.
//
// Returns:
//   - The int64 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingInt64(path string, defaultValue int64) int64 {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Int64()
	}
	return defaultValue
}

// DebuggingUint retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A uint value to return if the key is not available.
//
// Returns:
//   - The uint value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingUint(key string, defaultValue uint) uint {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.UintOrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingUint retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A uint value to return if the key is not available.
//
// Returns:
//   - The uint value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingUint(path string, defaultValue uint) uint {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Uint()
	}
	return defaultValue
}

// DebuggingUint8 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A uint8 value to return if the key is not available.
//
// Returns:
//   - The uint8 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingUint8(key string, defaultValue uint8) uint8 {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.Uint8OrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingUint8 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A uint8 value to return if the key is not available.
//
// Returns:
//   - The uint8 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingUint8(path string, defaultValue uint8) uint8 {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Uint8()
	}
	return defaultValue
}

// DebuggingUint16 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A uint16 value to return if the key is not available.
//
// Returns:
//   - The uint16 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingUint16(key string, defaultValue uint16) uint16 {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.Uint16OrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingUint16 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A uint16 value to return if the key is not available.
//
// Returns:
//   - The uint16 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingUint16(path string, defaultValue uint16) uint16 {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Uint16()
	}
	return defaultValue
}

// DebuggingUint32 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A uint32 value to return if the key is not available.
//
// Returns:
//   - The uint32 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingUint32(key string, defaultValue uint32) uint32 {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.Uint32OrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingUint32 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A uint32 value to return if the key is not available.
//
// Returns:
//   - The uint32 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingUint32(path string, defaultValue uint32) uint32 {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Uint32()
	}
	return defaultValue
}

// DebuggingUint64 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A uint64 value to return if the key is not available.
//
// Returns:
//   - The uint64 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingUint64(key string, defaultValue uint64) uint64 {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.Uint64OrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingUint64 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A uint64 value to return if the key is not available.
//
// Returns:
//   - The uint64 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingUint64(path string, defaultValue uint64) uint64 {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Uint64()
	}
	return defaultValue
}

// DebuggingFloat32 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A float32 value to return if the key is not available.
//
// Returns:
//   - The float32 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingFloat32(key string, defaultValue float32) float32 {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.Float32OrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingFloat32 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A float32 value to return if the key is not available.
//
// Returns:
//   - The float32 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingFloat32(path string, defaultValue float32) float32 {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Float32()
	}
	return defaultValue
}

// DebuggingFloat64 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A float64 value to return if the key is not available.
//
// Returns:
//   - The float64 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingFloat64(key string, defaultValue float64) float64 {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.Float64OrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingFloat64 retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A float64 value to return if the key is not available.
//
// Returns:
//   - The float64 value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingFloat64(path string, defaultValue float64) float64 {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Float64()
	}
	return defaultValue
}

// DebuggingDuration retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `key`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A time.Duration value to return if the key is not available.
//
// Returns:
//   - The time.Duration value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) DebuggingDuration(key string, defaultValue time.Duration) time.Duration {
	if !w.Available() || !w.IsDebuggingKeyPresent(key) {
		return defaultValue
	}
	return conv.DurationOrDefault(w.debug[key], defaultValue)
}

// JSONDebuggingDuration retrieves the value of a specific debugging key from the [wrapper] instance.
//
// This function checks if the [wrapper] is available (non-nil) and if the specified debugging key
// is present in the `debug` map. If both conditions are met, it returns the value associated with
// the specified key. Otherwise, it returns `defaultValue` to indicate the key is not available.
//
// Parameters:
//   - `path`: A string representing the debugging key to retrieve.
//   - `defaultValue`: A time.Duration value to return if the key is not available.
//
// Returns:
//   - The time.Duration value associated with the specified debugging key if it exists.
//   - `defaultValue` if the [wrapper] is unavailable or the key is not present in the `debug` map.
func (w *wrapper) JSONDebuggingDuration(path string, defaultValue time.Duration) time.Duration {
	if !w.Available() {
		return defaultValue
	}
	ctx := fj.Get(w.JSONDebugging(), path)
	if ctx.Exists() {
		return ctx.Duration()
	}
	return defaultValue
}

// Pagination retrieves the [pagination] instance associated with the [wrapper].
//
// This function returns the [pagination] field of the [wrapper], allowing access to
// pagination details such as the current page, total pages, and total items. If no
// pagination information is available, it returns `nil`.
//
// Returns:
//   - A pointer to the [pagination] instance if available.
//   - `nil` if the [pagination] field is not set.
func (w *wrapper) Pagination() *pagination {
	return w.pagination
}

// Cursor retrieves the [cursor] instance associated with the [wrapper].
//
// This function returns the [cursor] field of the [wrapper], which may contain information
// related to cursor-based pagination or navigation. If no cursor information is available,
// it returns `nil`.
//
// Returns:
//   - A pointer to the [cursor] instance if available.
//   - `nil` if the [cursor] field is not set.
func (w *wrapper) Cursor() *cursor {
	return w.cursor
}

// Meta retrieves the [meta] information from the [wrapper] instance.
//
// This function returns the [meta] field, which contains metadata related to the response or data
// in the [wrapper] instance. If no [meta] information is set, it returns `nil`.
//
// Returns:
//   - A pointer to the [meta] instance associated with the [wrapper].
//   - `nil` if no [meta] information is available.
func (w *wrapper) Meta() *meta {
	return w.meta
}

// Header retrieves the [header] associated with the [wrapper] instance.
//
// This function returns the [header] field from the [wrapper] instance, which contains
// information about the HTTP response or any other relevant metadata. If the [wrapper]
// instance is correctly initialized, it will return the [header]; otherwise, it may
// return `nil` if the [header] has not been set.
//
// Returns:
//   - A pointer to the [header] instance associated with the [wrapper].
//   - `nil` if the [header] is not set or the [wrapper] is uninitialized.
func (w *wrapper) Header() *header {
	return w.header
}

// IsDebuggingPresent checks whether debugging information is present in the [wrapper] instance.
//
// This function verifies if the `debug` field of the [wrapper] is not nil and contains at least one entry.
// It returns `true` if debugging information is available; otherwise, it returns `false`.
//
// Returns:
//   - A boolean value indicating whether debugging information is present:
//   - `true` if `debug` is not nil and contains data.
//   - `false` if `debug` is nil or empty.
func (w *wrapper) IsDebuggingPresent() bool {
	return w.Available() && w.debug != nil && len(w.debug) > 0
}

// IsDebuggingKeyPresent checks whether a specific key exists in the `debug` information.
//
// This function first checks if debugging information is present using `IsDebuggingPresent()`.
// Then it uses `coll.MapContainsKey` to verify if the given key is present within the `debug` map.
//
// Parameters:
//   - `key`: The key to search for within the `debug` field.
//
// Returns:
//   - A boolean value indicating whether the specified key is present in the `debug` map:
//   - `true` if the `debug` field is present and contains the specified key.
//   - `false` if `debug` is nil or does not contain the key.
func (w *wrapper) IsDebuggingKeyPresent(key string) bool {
	return w.IsDebuggingPresent() && coll.ContainsKeyComp(w.debug, key)
}

// IsBodyPresent checks whether the body data is present in the [wrapper] instance.
//
// This function checks if the `data` field of the [wrapper] is not nil, indicating that the body contains data.
//
// Returns:
//   - A boolean value indicating whether the body data is present:
//   - `true` if `data` is not nil.
//   - `false` if `data` is nil.
func (w *wrapper) IsBodyPresent() bool {
	if w == nil {
		return false
	}
	value := reflect.ValueOf(w.data)
	return !common.IsEmptyValue(value)
}

// IsJSONBody checks whether the body data is a valid JSON string.
//
// This function first checks if the [wrapper] is available and if the body data is present using `IsBodyPresent()`.
// Then it uses the `JSON()` function to retrieve the body data as a JSON string and checks if it is valid using `fj.IsValidJSON()`.
//
// Returns:
//   - A boolean value indicating whether the body data is a valid JSON string:
//   - `true` if the [wrapper] is available, the body data is present, and the body data is a valid JSON string.
//   - `false` if the [wrapper] is not available, the body data is not present, or the body data is not a valid JSON string.
func (w *wrapper) IsJSONBody() bool {
	if !w.Available() || !w.IsBodyPresent() {
		return false
	}
	json := encoding.JSON(w.data)
	return fj.IsValidJSONString(json) && encoding.IsValidJSONString(json)
}

// IsHeaderPresent checks whether header information is present in the [wrapper] instance.
//
// This function checks if the [header] field of the [wrapper] is not nil, indicating that header information is included.
//
// Returns:
//   - A boolean value indicating whether header information is present:
//   - `true` if [header] is not nil.
//   - `false` if [header] is nil.
func (w *wrapper) IsHeaderPresent() bool {
	return w.Available() && w.header != nil
}

// IsMetaPresent checks whether metadata information is present in the [wrapper] instance.
//
// This function checks if the [meta] field of the [wrapper] is not nil, indicating that metadata is available.
//
// Returns:
//   - A boolean value indicating whether metadata is present:
//   - `true` if [meta] is not nil.
//   - `false` if [meta] is nil.
func (w *wrapper) IsMetaPresent() bool {
	return w.Available() && w.meta != nil
}

// IsPagingPresent checks whether pagination information is present in the [wrapper] instance.
//
// This function checks if the [pagination] field of the [wrapper] is not nil, indicating that pagination details are included.
//
// Returns:
//   - A boolean value indicating whether pagination information is present:
//   - `true` if [pagination] is not nil.
//   - `false` if [pagination] is nil.
func (w *wrapper) IsPagingPresent() bool {
	return w.Available() && w.pagination != nil
}

// IsCursorPresent checks whether a cursor is present in the [wrapper] instance.
//
// This function checks if the `cursor` field of the [wrapper] is not nil, indicating that a cursor is available for pagination or data retrieval.
//
// Returns:
//   - A boolean value indicating whether a cursor is present:
//   - `true` if [cursor] is not nil.
//   - `false` if [cursor] is nil.
func (w *wrapper) IsCursorPresent() bool {
	return w.Available() && w.cursor != nil
}

// IsErrorPresent checks whether an error is present in the [wrapper] instance.
//
// This function checks if the `errors` field of the [wrapper] is not nil, indicating that an error has occurred.
//
// Returns:
//   - A boolean value indicating whether an error is present:
//   - `true` if `errors` is not nil.
//   - `false` if `errors` is nil.
func (w *wrapper) IsErrorPresent() bool {
	return w.Available() && w.errors != nil
}

// IsTotalPresent checks whether the total number of items is present in the [wrapper] instance.
//
// This function checks if the `total` field of the [wrapper] is greater than or equal to 0,
// indicating that a valid total number of items has been set.
//
// Returns:
//   - A boolean value indicating whether the total is present:
//   - `true` if `total` is greater than or equal to 0.
//   - `false` if `total` is negative (indicating no total value).
func (w *wrapper) IsTotalPresent() bool {
	return w.Available() && w.total >= 0
}

// IsStatusCodePresent checks whether a valid status code is present in the [wrapper] instance.
//
// This function checks if the `statusCode` field of the [wrapper] is greater than 0,
// indicating that a valid HTTP status code has been set.
//
// Returns:
//   - A boolean value indicating whether the status code is present:
//   - `true` if `statusCode` is greater than 0.
//   - `false` if `statusCode` is less than or equal to 0.
func (w *wrapper) IsStatusCodePresent() bool {
	return w.Available() && w.statusCode > 0
}

// IsError checks whether there is an error present in the [wrapper] instance.
//
// This function returns `true` if the [wrapper] contains an error, which can be any of the following:
//   - An error present in the `errors` field.
//   - A client error (4xx status code) or a server error (5xx status code).
//
// Returns:
//   - A boolean value indicating whether there is an error:
//   - `true` if there is an error present, either in the `errors` field or as an HTTP client/server error.
//   - `false` if no error is found.
func (w *wrapper) IsError() bool {
	return w.IsErrorPresent() || w.IsClientError() || w.IsServerError()
}

// IsSuccess checks whether the HTTP status code indicates a successful response.
//
// This function checks if the `statusCode` is between 200 and 299, inclusive, which indicates a successful HTTP response.
//
// Returns:
//   - A boolean value indicating whether the HTTP response was successful:
//   - `true` if the status code is between 200 and 299 (inclusive).
//   - `false` if the status code is outside of this range.
func (w *wrapper) IsSuccess() bool {
	return w.Available() && (200 <= w.statusCode) && (w.statusCode <= 299)
}

// IsInformational checks whether the HTTP status code indicates an informational response.
//
// This function checks if the `statusCode` is between 100 and 199, inclusive, which indicates an informational HTTP response.
//
// Returns:
//   - A boolean value indicating whether the HTTP response is informational:
//   - `true` if the status code is between 100 and 199 (inclusive).
//   - `false` if the status code is outside of this range.
func (w *wrapper) IsInformational() bool {
	return w.Available() && (100 <= w.statusCode) && (w.statusCode <= 199)
}

// IsRedirection checks whether the HTTP status code indicates a redirection response.
//
// This function checks if the `statusCode` is between 300 and 399, inclusive, which indicates a redirection HTTP response.
//
// Returns:
//   - A boolean value indicating whether the HTTP response is a redirection:
//   - `true` if the status code is between 300 and 399 (inclusive).
//   - `false` if the status code is outside of this range.
func (w *wrapper) IsRedirection() bool {
	return w.Available() && (300 <= w.statusCode) && (w.statusCode <= 399)
}

// IsClientError checks whether the HTTP status code indicates a client error.
//
// This function checks if the `statusCode` is between 400 and 499, inclusive, which indicates a client error HTTP response.
//
// Returns:
//   - A boolean value indicating whether the HTTP response is a client error:
//   - `true` if the status code is between 400 and 499 (inclusive).
//   - `false` if the status code is outside of this range.
func (w *wrapper) IsClientError() bool {
	return w.Available() && (400 <= w.statusCode) && (w.statusCode <= 499)
}

// IsServerError checks whether the HTTP status code indicates a server error.
//
// This function checks if the `statusCode` is between 500 and 599, inclusive, which indicates a server error HTTP response.
//
// Returns:
//   - A boolean value indicating whether the HTTP response is a server error:
//   - `true` if the status code is between 500 and 599 (inclusive).
//   - `false` if the status code is outside of this range.
func (w *wrapper) IsServerError() bool {
	return w.Available() && (500 <= w.statusCode) && (w.statusCode <= 599)
}

// IsLastPage checks whether the current page is the last page of results.
//
// This function verifies that pagination information is present and then checks if the current page is the last page.
// It combines the checks of `IsPagingPresent()` and `IsLast()` to ensure that the pagination structure exists
// and that it represents the last page.
//
// Returns:
//   - A boolean value indicating whether the current page is the last page:
//   - `true` if pagination is present and the current page is the last one.
//   - `false` if pagination is not present or the current page is not the last.
func (w *wrapper) IsLastPage() bool {
	return w.Available() && w.IsPagingPresent() && w.pagination.IsLast()
}

// EqualHeader compares the header information of the [wrapper] instance with another [header] instance.
//
// This function checks if the [wrapper] is available and if the provided [header] instance is not nil.
// It then compares the `code` and `text` fields of the [wrapper]'s header with those of the provided [header].
// The comparison of the `text` field is case-insensitive.
//
// Parameters:
//   - `h`: A pointer to a [header] instance to compare with the [wrapper]'s header.
//
// Returns:
//   - A boolean value indicating whether the headers are equal:
//   - `true` if both headers have the same code and text (case-insensitive).
//   - `false` if the [wrapper] is not available, the provided header is nil, or the headers do not match.
func (w *wrapper) EqualHeader(h *header) bool {
	if !w.Available() || h == nil {
		return false
	}
	if w.header == nil {
		return false
	}
	return w.header.Equal(h)
}

// EqualPages compares the pagination information of the [wrapper] instance with another [pagination] instance.
//
// This function checks if the [wrapper] is available and if the provided [pagination] instance is not nil.
// It then compares the pagination details of the [wrapper] with those of the provided [pagination] instance.
//
// Parameters:
//   - `p`: A pointer to a [pagination] instance to compare with the [wrapper]'s pagination.
//
// Returns:
//   - A boolean value indicating whether the pagination information is equal:
//   - `true` if both pagination instances have the same pagination details.
//   - `false` if the [wrapper] is not available, the provided pagination is nil, or the pagination details do not match.
func (w *wrapper) EqualPages(p *pagination) bool {
	if !w.Available() || p == nil {
		return false
	}
	if w.pagination == nil {
		return false
	}
	return w.pagination.Equal(p)
}

// EqualMeta compares the metadata information of the [wrapper] instance with another [meta] instance.
//
// This function checks if the [wrapper] is available and if the provided [meta] instance is not nil.
// It then compares the metadata details of the [wrapper] with those of the provided [meta] instance.
//
// Parameters:
//   - `m`: A pointer to a [meta] instance to compare with the [wrapper]'s metadata.
//
// Returns:
//   - A boolean value indicating whether the metadata information is equal:
//   - `true` if both meta instances have the same metadata details.
//   - `false` if the [wrapper] is not available, the provided meta is nil, or the metadata details do not match.
func (w *wrapper) EqualMeta(m *meta) bool {
	if !w.Available() || m == nil {
		return false
	}
	if w.meta == nil {
		return false
	}
	return w.meta.Equal(m)
}

// EqualCursor compares the cursor information of the [wrapper] instance with another [cursor] instance.
//
// This function checks if the [wrapper] is available and if the provided [cursor] instance is not nil.
// It then compares the cursor details of the [wrapper] with those of the provided [cursor] instance.
//
// Parameters:
//   - `c`: A pointer to a [cursor] instance to compare with the [wrapper]'s cursor.
//
// Returns:
//   - A boolean value indicating whether the cursor information is equal:
//   - `true` if both cursor instances have the same cursor details.
//   - `false` if the [wrapper] is not available, the provided cursor is nil, or the cursor details do not match.
func (w *wrapper) EqualCursor(c *cursor) bool {
	if !w.Available() || c == nil {
		return false
	}
	if w.cursor == nil {
		return false
	}
	return w.cursor.Equal(c)
}

// Clone creates a deep copy of the [wrapper] instance.
//
// This function creates a new [wrapper] instance with the same fields as the original instance.
// It creates a new [header], [meta], and [pagination] instances and copies the values from the original instance.
// It also creates a new `debug` map and copies the values from the original instance.
//
// Returns:
//   - A pointer to the cloned [wrapper] instance.
//   - `nil` if the [wrapper] instance is not available.
func (w *wrapper) Clone() *wrapper {
	if !w.Available() {
		return New()
	}

	clone := &wrapper{
		statusCode: w.statusCode,
		total:      w.total,
		message:    w.message,
		data:       w.data,
		path:       w.path,
		errors:     w.errors,
	}

	// Clone header
	if w.header != nil {
		clone.header = Header().
			WithCode(w.header.code).
			WithText(w.header.text).
			WithType(w.header.typez).
			WithDescription(w.header.description)
	}

	// Clone meta
	if w.meta != nil {
		clone.meta = Meta().
			WithApiVersion(w.meta.apiVersion).
			WithRequestID(w.meta.requestID).
			WithLocale(w.meta.locale).
			WithRequestedTime(w.meta.requestedTime)

		if w.meta.customFields != nil {
			customFieldsCopy := make(map[string]any)
			maps.Copy(customFieldsCopy, w.meta.customFields)
			clone.meta.WithCustomFields(customFieldsCopy)
		}
	}

	// Clone pagination
	if w.pagination != nil {
		clone.pagination = Pages().
			WithPage(w.pagination.page).
			WithPerPage(w.pagination.perPage).
			WithTotalPages(w.pagination.totalPages).
			WithTotalItems(w.pagination.totalItems).
			WithIsLast(w.pagination.isLast)
	}

	// Clone debug
	if w.debug != nil {
		clone.debug = make(map[string]any)
		maps.Copy(clone.debug, w.debug)
	}

	return clone
}

// Reset resets the [wrapper] instance to its initial state.
//
// This function sets the [wrapper] instance to its initial state by resetting
// the `statusCode`, `total`, `message`, `path`, `cacheHash`, `data`, `debug`,
// [header], `errors`, [pagination], and `cachedWrap` fields to their default values.
// It also resets the [meta] instance to its initial state.
//
// Returns:
//   - A pointer to the reset [wrapper] instance.
//   - `nil` if the [wrapper] instance is not available.
func (w *wrapper) Reset() *wrapper {
	if !w.Available() {
		return New()
	}

	// Reset status code and total
	w.total = 0
	w.statusCode = 0

	// Reset message, path, and cache hash
	w.path = ""
	w.message = ""
	w.cacheHash = ""

	// Reset data, debug, header, errors, pagination, and cached wrap
	w.data = nil
	w.debug = nil
	w.header = nil
	w.errors = nil
	w.pagination = nil
	w.cachedWrap = nil

	// Reset meta
	w.meta = defaultMetaValues()

	return w
}

// DeltaValue retrieves the delta value from the [meta] instance.
//
// This function checks if the [meta] instance is present and returns the `deltaValue` field.
// If the [meta] instance is not present, it returns a default value of `0`.
//
// Returns:
//   - A float64 representing the delta value.
func (w *wrapper) DeltaValue() float64 {
	if !w.IsMetaPresent() {
		return 0
	}
	return w.meta.deltaValue
}

// DeltaCnt retrieves the delta count from the [meta] instance.
//
// This function checks if the [meta] instance is present and returns the `deltaCnt` field.
// If the [meta] instance is not present, it returns a default value of `0`.
//
// Returns:
//   - An integer representing the delta count.
func (w *wrapper) DeltaCnt() int {
	if !w.IsMetaPresent() {
		return 0
	}
	return w.meta.deltaCnt
}

// WithStatusCode sets the HTTP status code for the [wrapper] instance.
// Ensure that code is between 100 and 599, defaults to 500 if invalid value.
//
// This function updates the `statusCode` field of the [wrapper] and
// returns the modified [wrapper] instance to allow method chaining.
//
// Parameters:
//   - `code`: An integer representing the HTTP status code to set.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
//
// Deprecated: This method is deprecated. Use [WithHeader] instead to set the status code and text together.
func (w *wrapper) WithStatusCode(code int) *wrapper {
	if code < 100 || code > 599 {
		code = http.StatusInternalServerError
	}
	w.statusCode = code
	w.header = Header().WithCode(code).WithText(http.StatusText(code))
	return w
}

// WithTotal sets the total number of items for the [wrapper] instance.
//
// This function updates the `total` field of the [wrapper] and
// returns the modified [wrapper] instance to allow method chaining.
//
// Parameters:
//   - `total`: An integer representing the total number of items to set.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithTotal(total int) *wrapper {
	w.total = total
	return w
}

// WithMessage sets a message for the [wrapper] instance.
//
// This function updates the `message` field of the [wrapper] with the provided string
// and returns the modified [wrapper] instance to allow method chaining.
//
// Parameters:
//   - `message`: A string message to be set in the [wrapper].
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithMessage(message string) *wrapper {
	w.message = message
	return w
}

// WithMessagef sets a formatted message for the [wrapper] instance.
//
// This function constructs a formatted string using the provided format string and arguments,
// assigns it to the `message` field of the [wrapper], and returns the modified instance.
//
// Parameters:
//   - message: A format string for constructing the message.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the modified [wrapper] instance, enabling method chaining.
func (w *wrapper) WithMessagef(message string, args ...any) *wrapper {
	w.message = fmt.Sprintf(message, args...)
	return w
}

// WithBody sets the body data for the [wrapper] instance.
//
// This function updates the `data` field of the [wrapper] with the provided value
// and returns the modified [wrapper] instance to allow method chaining.
//
// Parameters:
//   - `v`: The value to be set as the body data, which can be any type.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
//
// Example:
//
//	w, err := replify.New().WithBody(myStruct)
//
// Notes:
//   - This function does not validate or normalize the input value.
//   - It simply assigns the value to the `data` field of the [wrapper].
//   - The value will be marshalled to JSON when the [wrapper] is converted to a string.
//   - Consider using [WithBodySuppressed] instead if you need to normalize the input value.
func (w *wrapper) WithBody(v any) *wrapper {
	w.data = v
	return w
}

// WithBodySuppressed normalizes the input value and sets it as the body data for
// the [wrapper] instance.
//
// The method accepts any Go value and handles it according to its dynamic type:
//
//   - string        - the string is passed through encoding.NormalizeJSON, which
//     strips common JSON corruption artifacts (BOM, null bytes, escaped structural
//     quotes, trailing commas) before setting the result as the body.
//   - []byte        - treated as a raw string; the same NormalizeJSON pipeline is
//     applied after converting to string.
//   - json.RawMessage - validated directly; if invalid, an error is returned.
//   - any other type - marshaled to JSON via encoding.JSONToken and set as the body,
//     which is by definition already valid JSON.
//   - nil           - returns an error; nil cannot be normalized.
//
// If normalization succeeds, the cleaned value is stored as the body and the method
// returns the updated wrapper and nil.  If it fails, the body is left unchanged and
// a descriptive error is returned.
//
// Parameters:
//   - v: The value to normalize and set as the body.
//
// Returns:
//   - A pointer to the modified [wrapper] instance and nil on success.
//   - The unchanged [wrapper] instance and an error if normalization fails.
//
// Example:
//
//	// From a raw-string with escaped structural quotes:
//	w, err := replify.New().WithBodySuppressed(`{\"key\": "value"}`)
//
//	// From a struct:
//	w, err := replify.New().WithBodySuppressed(myStruct)
func (w *wrapper) WithBodySuppressed(v any) (*wrapper, error) {
	if v == nil {
		e2w := New().
			BadRequest().
			WithMessage("request body suppressed value must not be nil").
			WithError("WithBodySuppressed: value must not be nil; provide a non-nil value to normalize")
		e2w.Slogging()
		return w, e2w.Cause()
	}
	as, e2w := castValue(v)
	if e2w.IsError() {
		e2w.Slogging()
		return w, e2w.Cause()
	}
	w.data = as
	return w, nil
}

// WithPath sets the request path for the [wrapper] instance.
//
// This function updates the `path` field of the [wrapper] with the provided string
// and returns the modified [wrapper] instance to allow method chaining.
//
// Parameters:
//   - `v`: A string representing the request path.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithPath(v string) *wrapper {
	w.path = v
	return w
}

// WithPathf sets a formatted request path for the [wrapper] instance.
//
// This function constructs a formatted string using the provided format string `v` and arguments `args`,
// assigns the resulting string to the `path` field of the [wrapper], and returns the modified instance.
//
// Parameters:
//   - v: A format string for constructing the request path.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the modified [wrapper] instance, enabling method chaining.
func (w *wrapper) WithPathf(v string, args ...any) *wrapper {
	w.path = fmt.Sprintf(v, args...)
	return w
}

// WithHeader sets the header for the [wrapper] instance.
//
// This function updates the [header] field of the [wrapper] with the provided [header]
// instance and returns the modified [wrapper] instance to allow method chaining.
//
// Parameters:
//   - `v`: A pointer to a [header] struct that will be set in the [wrapper].
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithHeader(v *header) *wrapper {
	if v == nil {
		return w
	}
	w.header = v
	w.statusCode = v.Code()
	return w
}

// WithMeta sets the metadata for the [wrapper] instance.
//
// This function updates the [meta] field of the [wrapper] with the provided [meta]
// instance and returns the modified [wrapper] instance to allow method chaining.
//
// Parameters:
//   - `v`: A pointer to a [meta] struct that will be set in the [wrapper].
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithMeta(v *meta) *wrapper {
	w.meta = v
	return w
}

// WithPagination sets the pagination information for the [wrapper] instance.
//
// This function updates the [pagination] field of the [wrapper] with the provided [pagination]
// instance and returns the modified [wrapper] instance to allow method chaining.
//
// Parameters:
//   - `v`: A pointer to a [pagination] struct that will be set in the [wrapper].
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithPagination(v *pagination) *wrapper {
	w.pagination = v
	return w
}

// WithCursor sets the cursor information for the [wrapper] instance.
//
// This function updates the [cursor] field of the [wrapper] with the provided [cursor]
// instance and returns the modified [wrapper] instance to allow method chaining.
//
// Parameters:
//   - `v`: A pointer to a [cursor] struct that will be set in the [wrapper].
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithCursor(v *cursor) *wrapper {
	w.cursor = v
	return w
}

// WithDebugging sets the debugging information for the [wrapper] instance.
//
// This function updates the `debug` field of the [wrapper] with the provided map of debugging data
// and returns the modified [wrapper] instance to allow method chaining.
//
// Parameters:
//   - `v`: A map containing debugging information to be set in the [wrapper].
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithDebugging(v map[string]any) *wrapper {
	w.debug = v
	return w
}

// WithError sets an error for the [wrapper] instance.
//
// This function updates the `errors` field of the [wrapper] with the provided error
// and returns the modified [wrapper] instance to allow method chaining.
//
// Parameters:
//   - `err`: An error object to be set in the [wrapper].
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
// func (w *wrapper) WithError(err error) *wrapper {
// 	w.errors = err
// 	return w
// }

// WithError sets an error for the [wrapper] instance using a plain error message.
//
// This function creates an error object from the provided message, assigns it to
// the `errors` field of the [wrapper], and returns the modified instance.
//
// Parameters:
//   - message: A string containing the error message to be wrapped as an error object.
//
// Returns:
//   - A pointer to the modified [wrapper] instance to support method chaining.
func (w *wrapper) WithError(message string) *wrapper {
	w.errors = NewError(message)
	return w
}

// WithErrorf sets a formatted error for the [wrapper] instance.
//
// This function uses a formatted string and arguments to construct an error object,
// assigns it to the `errors` field of the [wrapper], and returns the modified instance.
//
// Parameters:
//   - format: A format string for constructing the error message.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the modified [wrapper] instance to support method chaining.
func (w *wrapper) WithErrorf(format string, args ...any) *wrapper {
	w.errors = NewErrorf(format, args...)
	return w
}

// WithErrorAck sets an error with a stack trace for the [wrapper] instance.
//
// This function wraps the provided error with stack trace information, assigns it
// to the `errors` field of the [wrapper], and returns the modified instance.
//
// Parameters:
//   - err: The error object to be wrapped with stack trace information.
//
// Returns:
//   - A pointer to the modified [wrapper] instance to support method chaining.
func (w *wrapper) WithErrorAck(err error) *wrapper {
	w.errors = NewErrorAck(err)
	return w
}

// AppendErrorAck wraps an existing error with an additional message and sets it for the [wrapper] instance.
//
// This function adds context to the provided error by wrapping it with an additional message.
// The resulting error is assigned to the `errors` field of the [wrapper].
//
// Parameters:
//   - err: The original error to be wrapped.
//   - message: A string message to add context to the error.
//
// Returns:
//   - A pointer to the modified [wrapper] instance to support method chaining.
func (w *wrapper) AppendErrorAck(err error, message string) *wrapper {
	w.errors = AppendErrorAck(err, message)
	return w
}

// WithErrorAckf wraps an existing error with a formatted message and sets it for the [wrapper] instance.
//
// This function adds context to the provided error by wrapping it with a formatted message.
// The resulting error is assigned to the `errors` field of the [wrapper].
//
// Parameters:
//   - err: The original error to be wrapped.
//   - format: A format string for constructing the contextual error message.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the modified [wrapper] instance to support method chaining.
func (w *wrapper) WithErrorAckf(err error, format string, args ...any) *wrapper {
	w.errors = NewErrorAckf(err, format, args...)
	return w
}

// AppendError adds a plain contextual message to an existing error and sets it for the [wrapper] instance.
//
// This function wraps the provided error with an additional plain message and assigns it
// to the `errors` field of the [wrapper].
//
// Parameters:
//   - err: The original error to be wrapped.
//   - message: A plain string message to add context to the error.
//
// Returns:
//   - A pointer to the modified [wrapper] instance to support method chaining.
func (w *wrapper) AppendError(err error, message string) *wrapper {
	w.errors = AppendError(err, message)
	return w
}

// AppendErrorf adds a formatted contextual message to an existing error and sets it for the [wrapper] instance.
//
// This function wraps the provided error with an additional formatted message and assigns it
// to the `errors` field of the [wrapper].
//
// Parameters:
//   - err: The original error to be wrapped.
//   - format: A format string for constructing the contextual error message.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the modified [wrapper] instance to support method chaining.
func (w *wrapper) AppendErrorf(err error, format string, args ...any) *wrapper {
	w.errors = AppendErrorf(err, format, args...)
	return w
}

// AppendErrors appends multiple errors to the [wrapper] instance.
//
// This function iterates through the provided errors slice, skips nil entries,
// and folds each error into the wrapper's internal error chain.
//
// Rules:
//   - If `w.errors` is nil, it will be set to the first non-nil error.
//   - For subsequent errors, it wraps the existing chain using AppendErrorAck
//     with the error's message.
//
// Parameters:
//   - errs: A slice of errors to be appended.
//
// Returns:
//   - A pointer to the modified [wrapper] instance to support method chaining.
func (w *wrapper) AppendErrors(errs []error) *wrapper {
	if !w.Available() || len(errs) == 0 {
		return w
	}
	for _, err := range errs {
		if err == nil {
			continue
		}
		if w.errors == nil {
			w.errors = err
			continue
		}
		w.errors = AppendErrorAck(w.errors, err.Error())
	}
	return w
}

// BindCause sets the error for the [wrapper] instance using its current message.
//
// This function creates an error object from the `message` field of the [wrapper],
// assigns it to the `errors` field, and returns the modified instance.
// If an error is already present, the message is appended as an additional cause.
// It allows for method chaining.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) BindCause() *wrapper {
	if !w.Available() {
		return w
	}
	if strutil.IsNotEmpty(w.message) {
		if w.errors == nil {
			w.errors = NewError(w.message)
		} else {
			w.errors = AppendErrorAck(w.errors, w.message)
		}
	}
	return w
}

// StackTrace returns the [StackTrace] embedded in the error held by this [wrapper].
//
// The method is strictly error-focused: it only returns a non-nil [StackTrace]
// when an error is actually present ([IsError] returns true) and the stored
// error was created with a stack-capturing constructor ([NewError], [NewErrorf],
// [NewErrorAck], [NewErrorAckf], [AppendErrorAck], [AppendErrorAckf]).
//
// Returns nil in all other cases:
//   - The [wrapper] is nil or unavailable.
//   - No error is set (w.errors == nil / [IsError] returns false).
//   - The stored error carries no embedded stack (e.g. bare [AppendError] /
//     [AppendErrorf]).
//
// Returns:
//   - The [StackTrace] from the stored error, or nil when no stack-aware error
//     is present.
func (w *wrapper) StackTrace() StackTrace {
	if !w.Available() || !w.IsError() {
		return nil
	}
	w.autoAdjust() // ensure the error chain is fully initialized before checking for stack traces
	type stackTracer interface {
		StackTrace() StackTrace
	}
	if st, ok := w.errors.(stackTracer); ok {
		return st.StackTrace()
	}
	return nil
}

// StackTraceString returns a formatted, human-readable representation of the
// [StackTrace] associated with this [wrapper].
//
// Returns an empty string when [StackTrace] returns nil (i.e. no error is set
// or the error carries no embedded stack trace).
//
// Returns:
//   - A multi-line string with one formatted frame per line, or an empty string.
func (w *wrapper) StackTraceString() string {
	trace := w.StackTrace()
	if len(trace) == 0 {
		return ""
	}
	return fmt.Sprintf("%+v", trace)
}

// WithSkipBody controls whether the body payload is omitted from output.
//
// When skip is true the body is excluded from [String], [build], and
// [Slogging] — useful when the payload is too large or sensitive to
// include in logs. The body data itself is still stored on the wrapper
// and can be retrieved via [Body] or [BodyString] at any time.
//
// The default value is false (body is always rendered).
//
// Parameters:
//   - skip: Pass true to suppress the body; false to restore default behavior.
//
// Returns:
//   - A pointer to the [wrapper] instance, enabling method chaining.
func (w *wrapper) WithSkipBody(skip bool) *wrapper {
	if !w.Available() {
		return w
	}
	w.skipBody = skip
	return w
}

// IsBodySkipped reports whether the body payload is currently suppressed
// from output by [String], [build], and [Slogging].
//
// Returns:
//   - true if [WithSkipBody](true) has been called on this instance.
//   - false otherwise (default).
func (w *wrapper) IsBodySkipped() bool {
	return w.Available() && w.skipBody
}

// WithStackTrace captures the call stack at the point this method is invoked
// and stores the formatted frames in the [wrapper]'s debug map under the
// key "stack_trace". Each frame is serialized using [Frame.MarshalText].
//
// This records the code location where a response was built—independently of
// where any error was created—making it complementary to [InjectStackTrace].
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithStackTrace() *wrapper {
	if !w.Available() {
		return w
	}
	trace := Callers().StackTrace()
	frames := make([]string, 0, len(trace))
	for _, f := range trace {
		if text, err := f.MarshalText(); err == nil && len(text) > 0 {
			frames = append(frames, string(text))
		}
	}
	return w.WithDebuggingKV("stack_trace", frames)
}

// InjectStackTrace extracts the [StackTrace] from the [wrapper]'s stored error
// and writes the formatted frames into the debug map under the key
// "error_stack_trace". Each frame is serialized using [Frame.MarshalText].
//
// This is a no-op when no error is set ([IsError] returns false). It is
// particularly useful for surfacing the exact origin of an error in
// diagnostic or development responses without altering the error itself.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) InjectStackTrace() *wrapper {
	if !w.Available() || !w.IsError() {
		return w
	}
	trace := w.StackTrace()
	frames := make([]string, 0, len(trace))
	for _, f := range trace {
		if text, err := f.MarshalText(); err == nil && len(text) > 0 {
			frames = append(frames, string(text))
		}
	}
	return w.WithDebuggingKV("error_stack_trace", frames)
}

// DisableInjectStackTrace removes the "error_stack_trace" entry from the debug map of the [wrapper] instance.
//
// This function checks if debugging information is present in the [wrapper]. If it is,
// it deletes the "error_stack_trace" key from the `debug` map. The modified [wrapper]
// instance is then returned to allow method chaining.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) DisableInjectStackTrace() *wrapper {
	if !w.Available() {
		return w
	}
	if w.IsDebuggingPresent() {
		delete(w.debug, "error_stack_trace")
	}
	return w
}

// WithDebuggingKV adds a key-value pair to the debugging information in the [wrapper] instance.
//
// This function checks if debugging information is already present. If it is not, it initializes
// an empty map. Then it adds the given key-value pair to the `debug` map and returns the modified
// [wrapper] instance to allow method chaining.
//
// Parameters:
//   - `key`: The key for the debugging information to be added.
//   - `value`: The value associated with the key to be added to the `debug` map.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithDebuggingKV(key string, value any) *wrapper {
	if !w.IsDebuggingPresent() {
		w.debug = make(map[string]any)
	}
	w.debug[key] = value
	return w
}

// WithDebuggingKVf adds a formatted key-value pair to the debugging information in the [wrapper] instance.
//
// This function creates a formatted string value using the provided `format` string and `args`,
// then delegates to `WithDebuggingKV` to add the resulting key-value pair to the `debug` map.
// It returns the modified [wrapper] instance for method chaining.
//
// Parameters:
//   - key: A string representing the key for the debugging information.
//   - format: A format string for constructing the value.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the modified [wrapper] instance, enabling method chaining.
func (w *wrapper) WithDebuggingKVf(key string, format string, args ...any) *wrapper {
	return w.WithDebuggingKV(key, fmt.Sprintf(format, args...))
}

// WithApiVersion sets the API version in the [meta] field of the [wrapper] instance.
//
// This function checks if the [meta] information is present in the [wrapper]. If it is not,
// a new [meta] instance is created. Then, it calls the `WithApiVersion` method on the [meta]
// instance to set the API version.
//
// Parameters:
//   - `v`: A string representing the API version to set.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithApiVersion(v string) *wrapper {
	if !w.IsMetaPresent() {
		w.meta = Meta()
	}
	w.meta.WithApiVersion(v)
	return w
}

// WithApiVersionf sets the API version in the [meta] field of the [wrapper] instance using a formatted string.
//
// This function ensures that the [meta] field in the [wrapper] is initialized. If the [meta]
// field is not present, a new [meta] instance is created using the `NewMeta` function.
// Once the [meta] instance is ready, it updates the API version using the `WithApiVersionf` method
// on the [meta] instance. The API version is constructed by interpolating the provided `format`
// string with the variadic arguments (`args`).
//
// Parameters:
//   - format: A format string used to construct the API version.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the modified [wrapper] instance, enabling method chaining.
func (w *wrapper) WithApiVersionf(format string, args ...any) *wrapper {
	if !w.IsMetaPresent() {
		w.meta = Meta()
	}
	w.meta.WithApiVersionf(format, args...)
	return w
}

// WithRequestID sets the request ID in the [meta] field of the [wrapper] instance.
//
// This function ensures that if [meta] information is not already set in the [wrapper], a new
// [meta] instance is created. Then, it calls the `WithRequestID` method on the [meta] instance
// to set the request ID.
//
// Parameters:
//   - `v`: A string representing the request ID to set.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithRequestID(v string) *wrapper {
	if !w.IsMetaPresent() {
		w.meta = Meta()
	}
	w.meta.WithRequestID(v)
	return w
}

// WithRequestIDf sets the request ID in the [meta] field of the [wrapper] instance using a formatted string.
//
// This function ensures that the [meta] field in the [wrapper] is initialized. If the [meta] field
// is not already present, a new [meta] instance is created using the `NewMeta` function.
// Once the [meta] instance is ready, it updates the request ID by calling the `WithRequestIDf`
// method on the [meta] instance. The request ID is constructed using the provided `format` string
// and the variadic `args`.
//
// Parameters:
//   - format: A format string used to construct the request ID.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the modified [wrapper] instance, allowing for method chaining.
func (w *wrapper) WithRequestIDf(format string, args ...any) *wrapper {
	if !w.IsMetaPresent() {
		w.meta = Meta()
	}
	w.meta.WithRequestIDf(format, args...)
	return w
}

// RandRequestID generates and sets a random request ID in the [meta] field of the [wrapper] instance.
//
// This function checks if the [meta] field is present in the [wrapper]. If it is not,
// a new [meta] instance is created. Then, it calls the `RandRequestID` method on the [meta]
// instance to generate and set a random request ID.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) RandRequestID() *wrapper {
	if !w.IsMetaPresent() {
		w.meta = Meta()
	}
	w.meta.RandRequestID()
	return w
}

// RandDeltaValue generates and sets a random delta value in the [meta] field of the [wrapper] instance.
//
// This function checks if the [meta] field is present in the [wrapper]. If it is not,
// a new [meta] instance is created. Then, it calls the `RandDeltaValue` method on the [meta]
// instance to generate and set a random delta value.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) RandDeltaValue() *wrapper {
	if !w.IsMetaPresent() {
		w.meta = Meta()
	}
	w.meta.RandDeltaValue()
	w.meta.IncreaseDeltaCnt()
	return w
}

// IncreaseDeltaCnt increments the delta count in the [meta] field of the [wrapper] instance.
//
// This function ensures the [meta] field is present, creating a new instance if needed, and
// increments the delta count in the [meta] using the `IncreaseDeltaCnt` method.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) IncreaseDeltaCnt() *wrapper {
	if !w.IsMetaPresent() {
		w.meta = Meta()
	}
	w.meta.IncreaseDeltaCnt()
	return w
}

// DecreaseDeltaCnt decrements the delta count in the [meta] field of the [wrapper] instance.
//
// This function ensures the [meta] field is present, creating a new instance if needed, and
// decrements the delta count in the [meta] using the `DecreaseDeltaCnt` method.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) DecreaseDeltaCnt() *wrapper {
	if !w.IsMetaPresent() {
		w.meta = Meta()
	}
	w.meta.DecreaseDeltaCnt()
	return w
}

// WithLocale sets the locale in the [meta] field of the [wrapper] instance.
//
// This function ensures the [meta] field is present, creating a new instance if needed, and
// sets the locale in the [meta] using the `WithLocale` method.
//
// Parameters:
//   - `v`: A string representing the locale to set.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithLocale(v string) *wrapper {
	if !w.IsMetaPresent() {
		w.meta = Meta()
	}
	w.meta.WithLocale(v)
	return w
}

// WithRequestedTime sets the requested time in the [meta] field of the [wrapper] instance.
//
// This function ensures that the [meta] field exists, and if not, creates a new one. It then
// sets the requested time in the [meta] using the `WithRequestedTime` method.
//
// Parameters:
//   - `v`: A `time.Time` value representing the requested time.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithRequestedTime(v time.Time) *wrapper {
	if !w.IsMetaPresent() {
		w.meta = Meta()
	}
	w.meta.WithRequestedTime(v)
	return w
}

// WithCustomFields sets the custom fields in the [meta] field of the [wrapper] instance.
//
// This function checks if the [meta] field is present. If not, it creates a new [meta] instance
// and sets the provided custom fields using the `WithCustomFields` method.
//
// Parameters:
//   - `values`: A map representing the custom fields to set in the [meta].
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithCustomFields(values map[string]any) *wrapper {
	if !w.IsMetaPresent() {
		w.meta = Meta()
	}
	w.meta.WithCustomFields(values)
	return w
}

// WithCustomFieldKV sets a specific custom field key-value pair in the [meta] field of the [wrapper] instance.
//
// This function ensures that if the [meta] field is not already set, a new [meta] instance is created.
// It then adds the provided key-value pair to the custom fields of [meta] using the `WithCustomFieldKV` method.
//
// Parameters:
//   - `key`: A string representing the custom field key to set.
//   - `value`: The value associated with the custom field key.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) WithCustomFieldKV(key string, value any) *wrapper {
	if !w.IsMetaPresent() {
		w.meta = Meta()
	}
	w.meta.WithCustomFieldKV(key, value)
	return w
}

// WithCustomFieldKVf sets a specific custom field key-value pair in the [meta] field of the [wrapper] instance
// using a formatted value.
//
// This function constructs a formatted string value using the provided `format` string and arguments (`args`).
// It then calls the `WithCustomFieldKV` method to add or update the custom field with the specified key and
// the formatted value. If the [meta] field of the [wrapper] instance is not initialized, it is created
// before setting the custom field.
//
// Parameters:
//   - key: A string representing the key for the custom field.
//   - format: A format string to construct the value.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the modified [wrapper] instance, enabling method chaining.
func (w *wrapper) WithCustomFieldKVf(key string, format string, args ...any) *wrapper {
	if !w.IsMetaPresent() {
		w.meta = Meta()
	}
	w.meta.WithCustomFieldKVf(key, format, args...)
	return w
}

// WithPage sets the current page number in the wrapper's pagination.
//
// If the pagination object is not already initialized, it creates a new one
// using the `NewPagination` function. The specified page number is then
// applied to the pagination instance.
//
// Parameters:
//   - v: The page number to set.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) WithPage(v int) *wrapper {
	if !w.IsPagingPresent() {
		w.pagination = Pages()
	}
	w.pagination.WithPage(v)
	return w
}

// WithPerPage sets the number of items per page in the wrapper's pagination.
//
// If the pagination object is not already initialized, it creates a new one
// using the `NewPagination` function. The specified items-per-page value
// is then applied to the pagination instance.
//
// Parameters:
//   - v: The number of items per page to set.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) WithPerPage(v int) *wrapper {
	if !w.IsPagingPresent() {
		w.pagination = Pages()
	}
	w.pagination.WithPerPage(v)
	return w
}

// WithTotalPages sets the total number of pages in the wrapper's pagination.
//
// If the pagination object is not already initialized, it creates a new one
// using the `NewPagination` function. The specified total pages value
// is then applied to the pagination instance.
//
// Parameters:
//   - v: The total number of pages to set.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) WithTotalPages(v int) *wrapper {
	if !w.IsPagingPresent() {
		w.pagination = Pages()
	}
	w.pagination.WithTotalPages(v)
	return w
}

// WithTotalItems sets the total number of items in the wrapper's pagination.
//
// If the pagination object is not already initialized, it creates a new one
// using the `NewPagination` function. The specified total items value
// is then applied to the pagination instance.
//
// Parameters:
//   - v: The total number of items to set.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) WithTotalItems(v int) *wrapper {
	if !w.IsPagingPresent() {
		w.pagination = Pages()
	}
	w.pagination.WithTotalItems(v)
	return w
}

// WithIsLast sets whether the current page is the last one in the wrapper's pagination.
//
// If the pagination object is not already initialized, it creates a new one
// using the `NewPagination` function. The specified boolean value is then
// applied to indicate whether the current page is the last.
//
// Parameters:
//   - v: A boolean indicating whether the current page is the last.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) WithIsLast(v bool) *wrapper {
	if !w.IsPagingPresent() {
		w.pagination = Pages()
	}
	w.pagination.WithIsLast(v)
	return w
}

// WithCursorNext sets the next cursor value in the [wrapper]'s [cursor].
//
// If the cursor object is not already initialized, it creates a new one
// using the `[Cursors]` function. The specified next cursor value is then
// applied to the cursor instance.
//
// Parameters:
//   - v: The next cursor value to set.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) WithCursorNext(v string) *wrapper {
	if !w.IsCursorPresent() {
		w.cursor = Cursors()
	}
	w.cursor.WithNext(v)
	return w
}

// WithCursorPrev sets the previous cursor value in the [wrapper]'s [cursor].
//
// If the cursor object is not already initialized, it creates a new one
// using the `[Cursors]` function. The specified previous cursor value is then
// applied to the cursor instance.
//
// Parameters:
//   - v: The previous cursor value to set.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) WithCursorPrev(v string) *wrapper {
	if !w.IsCursorPresent() {
		w.cursor = Cursors()
	}
	w.cursor.WithPrevious(v)
	return w
}

// WithCursorNextF sets the next cursor value in the [wrapper]'s [cursor] using a formatted string.
//
// If the cursor object is not already initialized, it creates a new one
// using the `[Cursors]` function. The specified next cursor value is then
// applied to the cursor instance using a formatted string.
//
// Parameters:
//   - format: A format string for constructing the next cursor value.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) WithCursorNextF(format string, args ...any) *wrapper {
	if !w.IsCursorPresent() {
		w.cursor = Cursors()
	}
	w.cursor.WithNextF(format, args...)
	return w
}

// WithCursorPrevF sets the previous cursor value in the [wrapper]'s [cursor] using a formatted string.
//
// If the cursor object is not already initialized, it creates a new one
// using the `[Cursors]` function. The specified previous cursor value is then
// applied to the cursor instance using a formatted string.
//
// Parameters:
//   - format: A format string for constructing the previous cursor value.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) WithCursorPrevF(format string, args ...any) *wrapper {
	if !w.IsCursorPresent() {
		w.cursor = Cursors()
	}
	w.cursor.WithPreviousF(format, args...)
	return w
}

// WithCursorHasNext sets whether there is a next cursor available in the [wrapper]'s [cursor].
//
// If the cursor object is not already initialized, it creates a new one
// using the `[Cursors]` function. The specified boolean value is then
// applied to indicate whether there is a next cursor available.
//
// Parameters:
//   - v: A boolean indicating whether there is a next cursor.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) WithCursorHasNext(v bool) *wrapper {
	if !w.IsCursorPresent() {
		w.cursor = Cursors()
	}
	w.cursor.WithHasNext(v)
	return w
}

// WithCursorHasPrev sets whether there is a previous cursor available in the [wrapper]'s [cursor].
//
// If the cursor object is not already initialized, it creates a new one
// using the `[Cursors]` function. The specified boolean value is then
// applied to indicate whether there is a previous cursor available.
//
// Parameters:
//   - v: A boolean indicating whether there is a previous cursor.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) WithCursorHasPrev(v bool) *wrapper {
	if !w.IsCursorPresent() {
		w.cursor = Cursors()
	}
	w.cursor.WithHasPrevious(v)
	return w
}

// WithCursorLimit sets the limit for the number of items in the [wrapper]'s [cursor].
//
// If the cursor object is not already initialized, it creates a new one
// using the `[Cursors]` function. The specified limit value is then
// applied to the cursor instance.
//
// Parameters:
//   - v: An integer representing the limit for the number of items.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) WithCursorLimit(v int) *wrapper {
	if !w.IsCursorPresent() {
		w.cursor = Cursors()
	}
	w.cursor.WithLimit(v)
	return w
}

// MustHash256 generates a hash string for the [wrapper] instance.
//
// This method concatenates the values of the `statusCode`, `message`, `data`, and [meta] fields
// into a single string and then computes a hash of that string using the `strutil.MustHash256` function.
// The resulting hash string can be used for various purposes, such as caching or integrity checks.
func (w *wrapper) MustHash256() (string, *wrapper) {
	if !w.Available() {
		return "", w
	}
	h, err := hashy.Hash256(w.StatusCode(), w.message, w.data, w.meta.Respond())
	if err != nil {
		return "", New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("Failed to generate hash")
	}
	return h, New().
		WithHeader(OK).
		WithMessage("Successfully generated hash")
}

// Hash256 generates a hash string for the [wrapper] instance.
//
// This method generates a hash string for the [wrapper] instance using the `Hash256` method.
// If the [wrapper] instance is not available or the hash generation fails, it returns an empty string.
//
// Returns:
//   - A string representing the hash value.
//   - An empty string if the [wrapper] instance is not available or the hash generation fails.
func (w *wrapper) Hash256() string {
	hash, _w := w.MustHash256()
	if _w.IsError() {
		return ""
	}
	return hash
}

// MustHash generates a hash value for the [wrapper] instance.
//
// This method generates a hash value for the [wrapper] instance using the `MustHash` method.
// If the [wrapper] instance is not available or the hash generation fails, it returns an error.
//
// Returns:
//   - A uint64 representing the hash value.
//   - An error if the [wrapper] instance is not available or the hash generation fails.
func (w *wrapper) MustHash() (uint64, *wrapper) {
	if !w.Available() {
		return 0, w
	}
	h, err := hashy.Hash(w.StatusCode(), w.message, w.data, w.meta.Respond())
	if err != nil {
		return 0, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("Failed to generate hash")
	}
	return h, New().
		WithHeader(OK).
		WithMessage("Successfully generated hash")
}

// This method generates a hash value for the [wrapper] instance using the `Hash` method.
// If the [wrapper] instance is not available or the hash generation fails, it returns an empty string.
//
// Returns:
//   - A string representing the hash value.
//   - An empty string if the [wrapper] instance is not available or the hash generation fails.
func (w *wrapper) Hash() uint64 {
	hash, _w := w.MustHash()
	if _w.IsError() {
		return 0
	}
	return hash
}

// WithStreaming enables streaming mode for the wrapper and returns a streaming wrapper for enhanced data transfer capabilities.
//
// This function is the primary entry point for activating streaming functionality on an existing wrapper instance.
// It creates a new StreamingWrapper that preserves the metadata and context of the original wrapper while adding
// streaming-specific features such as chunk-based transfer, compression, progress tracking, and bandwidth throttling.
// The returned StreamingWrapper allows for method chaining to configure streaming parameters before initiating transfer.
//
// Parameters:
//   - reader: An io.Reader implementation providing the source data stream (e.g., *os.File, *http.Response.Body, *bytes.Buffer).
//     Cannot be nil; streaming will fail if no valid reader is provided.
//   - config: A *StreamConfig containing streaming configuration options (chunk size, compression, strategy, concurrency).
//     If nil, a default configuration is automatically created with sensible defaults:
//   - ChunkSize: 65536 bytes (64KB)
//   - Strategy: STRATEGY_BUFFERED (balanced throughput and memory)
//   - Compression: COMP_NONE
//   - MaxConcurrentChunks: 4
//
// Returns:
//   - A pointer to a new StreamingWrapper instance that wraps the original wrapper.
//   - The StreamingWrapper preserves all metadata from the original wrapper.
//   - If the receiver wrapper is nil, creates a new default wrapper before enabling streaming.
//   - The returned StreamingWrapper can be chained with configuration methods before calling Start().
//
// Example:
//
//	file, _ := os.Open("large_file.bin")
//	defer file.Close()
//
//	// Simple streaming with defaults
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/file").
//	    WithStreaming(file, nil).
//	    WithChunkSize(1024 * 1024).
//	    WithCompressionType(COMP_GZIP).
//	    WithCallback(func(p *StreamProgress, err error) {
//	        if err == nil {
//	            fmt.Printf("Transferred: %.2f MB / %.2f MB\n",
//	                float64(p.TransferredBytes) / 1024 / 1024,
//	                float64(p.TotalBytes) / 1024 / 1024)
//	        }
//	    }).
//	    Start(context.Background()).
//	    WithMessage("File transfer completed")
//
// See Also:
//   - AsStreaming: Simplified version with default configuration
//   - Start: Initiates the streaming operation
//   - WithChunkSize: Configures chunk size
//   - WithCompressionType: Enables data compression
func (w *wrapper) WithStreaming(reader io.Reader, config *StreamConfig) *StreamingWrapper {
	if w == nil {
		return NewStreaming(reader, config)
	}
	if config == nil {
		config = NewStreamConfig()
	}

	sw := NewStreaming(reader, config)
	// Copy existing wrapper metadata
	sw.wrapper = w
	sw.wrapper.WithMessage("Streaming mode enabled")

	return sw
}

// AsStreaming converts a regular wrapper instance into a streaming-enabled response with default configuration.
//
// This function provides a simplified, one-line alternative to WithStreaming for common streaming scenarios.
// It automatically creates a new wrapper if the receiver is nil and applies default streaming configuration,
// eliminating the need for manual configuration object creation. This is ideal for quick implementations where
// standard settings (64KB chunks, buffered strategy, no compression) are acceptable.
//
// Parameters:
//   - reader: An io.Reader implementation providing the source data stream (e.g., *os.File, *http.Response.Body, *bytes.Buffer).
//     Cannot be nil; streaming will fail if no valid reader is provided.
//
// Returns:
//   - A pointer to a new StreamingWrapper instance configured with default settings:
//   - ChunkSize: 65536 bytes (64KB)
//   - Strategy: STRATEGY_BUFFERED
//   - Compression: COMP_NONE
//   - MaxConcurrentChunks: 4
//   - UseBufferPool: true
//   - ReadTimeout: 30 seconds
//   - WriteTimeout: 30 seconds
//   - If the receiver wrapper is nil, automatically creates a new wrapper before enabling streaming.
//   - Returns a StreamingWrapper ready for optional configuration before calling Start().
//
// Example:
//
//	// Minimal streaming setup with defaults - best for simple file downloads
//	file, _ := os.Open("document.pdf")
//	defer file.Close()
//
//	result := replify.New().
//	    WithStatusCode(200).
//	    WithPath("/api/download/document").
//	    AsStreaming(file).
//	    WithTotalBytes(fileSize).
//	    Start(context.Background())
//
//	// Or without creating a new wrapper first
//	result := (*wrapper)(nil).
//	    AsStreaming(file).
//	    Start(context.Background())
//
// Comparison:
//
//	// Using AsStreaming (simple, defaults only)
//	streaming := response.AsStreaming(reader)
//
//	// Using WithStreaming (more control)
//	streaming := response.WithStreaming(reader, &StreamConfig{
//	    ChunkSize:           512 * 1024,
//	    Compression:         COMP_GZIP,
//	    MaxConcurrentChunks: 8,
//	})
//
// See Also:
//   - WithStreaming: For custom streaming configuration
//   - NewStreamConfig: To create custom configuration objects
//   - Start: Initiates the streaming operation
//   - WithCallback: Adds progress tracking after AsStreaming
func (w *wrapper) AsStreaming(reader io.Reader) *StreamingWrapper {
	if w == nil {
		w = New()
	}
	return w.WithStreaming(reader, NewStreamConfig())
}

// hashFor computes a fast, allocation-free cache key over every field that
// build() serializes. It must be called without any mutex held; wrapper fields
// are expected to be stable (immutable after Wrap-time construction via With*
// options), so concurrent reads are safe.
//
// Unlike Hash256(), this helper avoids the *wrapper allocation that
// MustHash256() introduces on every call, and it covers ALL nine fields that
// build() writes to the response map—the public Hash256() only covers four.
func (w *wrapper) hashFor() string {
	h, err := hashy.Hash256(
		w.StatusCode(),
		w.message,
		w.data,
		w.header.Respond(),
		w.meta.Respond(),
		w.pagination.Respond(),
		w.debug,
		w.total,
		w.path,
	)
	if err != nil {
		return ""
	}
	return h
}

// Respond generates a map representation of the [wrapper] instance.
//
// This method collects various fields of the [wrapper] (e.g., `data`, [header], [meta], etc.)
// and organizes them into a key-value map. Only non-nil or meaningful fields are added
// to the resulting map to ensure a clean and concise response structure.
//
// Fields included in the response:
//   - `data`: The primary data payload, if present.
//   - `headers`: The structured header details, if present.
//   - [meta]: Metadata about the response, if present.
//   - [pagination]: Pagination details, if applicable.
//   - `debug`: Debugging information, if provided.
//   - `total`: Total number of items, if set to a valid non-negative value.
//   - `status_code`: The HTTP status code, if greater than 0.
//   - `message`: A descriptive message, if not empty.
//   - `path`: The request path, if not empty.
//
// # Caching
//
// The result is cached and reused as long as the wrapper's state does not
// change. Cache validity is checked with a hash that covers all nine output
// fields. The hash is computed before any mutex is acquired, so concurrent
// readers only contend on the brief cache-read/write critical sections—not on
// the (potentially expensive) hash or build steps.
//
// # Thread-safety
//
// Safe for concurrent use. Multiple goroutines may call Respond on the same
// wrapper simultaneously; at most one will execute build() for any given state.
//
// Returns:
//   - A `map[string]interface{}` containing the structured response data.
func (w *wrapper) Respond() map[string]any {
	if !w.Available() {
		return nil
	}

	// Compute the hash BEFORE taking any lock.
	// Fields are immutable after construction, so this concurrent read is safe.
	// Keeping the potentially-expensive hash computation outside the lock
	// prevents readers from blocking each other.
	hash := w.hashFor()

	// Fast path: check cache under read lock.
	w.cacheMutex.RLock()
	if w.cacheHash == hash && w.cachedWrap != nil {
		cached := w.cachedWrap
		w.cacheMutex.RUnlock()
		return cached
	}
	w.cacheMutex.RUnlock()

	// Slow path: acquire write lock, double-check, then rebuild.
	// The hash was computed from immutable fields, so it is stable across the
	// gap between the two lock acquisitions; no recomputation is needed.
	w.cacheMutex.Lock()
	defer w.cacheMutex.Unlock()

	if w.cacheHash == hash && w.cachedWrap != nil {
		return w.cachedWrap
	}

	response := w.build()
	w.cachedWrap = response
	w.cacheHash = hash
	return response
}

// R represents a wrapper around the main [wrapper] struct. It is used as a high-level
// abstraction to provide a simplified interface for handling API responses.
// The `R` type allows for easier manipulation of the wrapped data, metadata, and other
// response components, while maintaining the flexibility of the underlying [wrapper] structure.
//
// Example usage:
//
//	var response replify.R = replify.New().Reply()
//	fmt.Println(response.JSON())  // Prints the wrapped response details, including data, headers, and metadata.
func (w *wrapper) Reply() R {
	w.autoAdjust()
	return R{wrapper: w}
}

// ReplyPtr returns a pointer to a new R instance that wraps the current [wrapper].
//
// This method creates a new `R` struct, initializing it with the current [wrapper] instance,
// and returns a pointer to this new `R` instance. This allows for easier manipulation
// of the wrapped data and metadata through the `R` abstraction.
//
// Returns:
//   - A pointer to an `R` struct that wraps the current [wrapper] instance.
//
// Example usage:
//
//	var responsePtr *replify.R = replify.New().ReplyPtr()
//	fmt.Println(responsePtr.JSON())  // Prints the wrapped response details, including data, headers, and metadata.
func (w *wrapper) ReplyPtr() *R {
	w.autoAdjust()
	return &R{wrapper: w}
}

// JSON serializes the [wrapper] instance into a compact JSON string.
//
// This function uses the `encoding.JSON` utility to generate a JSON representation
// of the [wrapper] instance. The output is a compact JSON string with no additional
// whitespace or formatting.
//
// Returns:
//   - A compact JSON string representation of the [wrapper] instance.
func (w *wrapper) JSON() string {
	return jsonpass(w.Respond())
}

// JSONPretty serializes the [wrapper] instance into a prettified JSON string.
//
// This function uses the `encoding.JSONPretty` utility to generate a JSON representation
// of the [wrapper] instance. The output is a human-readable JSON string with
// proper indentation and formatting for better readability.
//
// Returns:
//   - A prettified JSON string representation of the [wrapper] instance.
func (w *wrapper) JSONPretty() string {
	return jsonpretty(w.Respond())
}

// JSONBytes serializes the [wrapper] instance into a JSON byte slice.
//
// This function first checks if the [wrapper] is available and if the body data is a valid JSON string using `IsJSONBody()`.
// If both conditions are met, it returns the JSON byte slice. Otherwise, it returns an empty byte slice.
//
// Returns:
//   - A byte slice containing the JSON representation of the [wrapper] instance.
//   - An empty byte slice if the [wrapper] is not available or the body data is not a valid JSON string.
func (w *wrapper) JSONBytes() []byte {
	return []byte(w.JSON())
}

// String returns a string representation of the [wrapper] instance.
//
// This method constructs a human-readable string representation of the [wrapper] instance by concatenating
// the values of various fields (e.g., `status_code`, `message`, `data`, `headers`, etc.) into a single string.
// Each field is included in the output only if it is present and meaningful (e.g., non-empty strings, non-nil values).
// The resulting string provides a concise summary of the [wrapper]'s state, which can be useful for logging or debugging purposes.
//
// Returns:
//   - A string representation of the [wrapper] instance, including relevant fields such as status code, message, data, headers, metadata, pagination, debugging information, total count, and path.
func (w *wrapper) String() string {
	w.autoAdjust()
	sw := strchain.New()
	if w.IsStatusCodePresent() {
		sw.AppendF("status_code=%q", w.StatusText()).Space()
	}
	if strutil.IsNotEmpty(w.path) {
		sw.AppendF("path=%q", w.path).Space()
	}
	if strutil.IsNotEmpty(w.message) {
		sw.AppendF("message=%q", w.message).Space()
	}
	if w.IsError() {
		sw.AppendF("error=%q", w.Error()).Space()
	}
	if w.IsBodyPresent() && !w.skipBody {
		sw.AppendF("data=%+v", w.BodyString()).Space()
	}
	if w.IsTotalPresent() {
		sw.AppendF("total=%d", w.total).Space()
	}
	if w.IsPagingPresent() {
		sw.AppendF("pagination=%q", w.pagination.String()).Space()
	}
	if w.IsMetaPresent() {
		sw.AppendF("meta=%q", w.meta.String()).Space()
	}
	if w.IsHeaderPresent() {
		sw.AppendF("header=%q", w.header.String()).Space()
	}
	if w.IsDebuggingPresent() {
		sw.AppendF("debug=%q", conv.StringOrEmpty(w.debug)).Space()
	}
	if w.IsCursorPresent() {
		sw.AppendF("cursor=%q", w.cursor.String()).Space()
	}
	return sw.String()
}

// UnidecodeBody returns a transliterated version of the [wrapper]'s body content.
//
// This method uses the [translit.Unidecode] function to convert the body content into a
// plain ASCII representation, which can be useful for logging, debugging, or displaying
// content in environments that do not support Unicode.
//
// If the body is not present, it returns an empty string. If the body is JSON, it
// first serializes the JSON content before applying transliteration.
//
// Returns:
//   - A string containing the transliterated body content, or an empty string if the body is not present.
func (w *wrapper) UnidecodeBody() string {
	if !w.IsBodyPresent() {
		return ""
	}
	if w.IsJSONBody() {
		return translit.Unidecode(w.JSON())
	}
	return translit.Unidecode(w.BodyString())
}

// autoAdjust automatically synchronizes the [wrapper]'s error field with its message
// when the HTTP status code indicates a client (4xx) or server (5xx) error and no
// explicit error has been set yet.
//
// This method is called internally by [Reply] and [ReplyPtr] to ensure that error
// responses always carry a cause derived from the message, mirroring the behavior
// of [BindCause] but triggered automatically rather than explicitly by the caller.
// It is a no-op when:
//   - the [wrapper] is nil.
//   - the status code is not an error code (i.e. not 4xx or 5xx).
//   - an error is already present (w.errors != nil).
//   - the message field is empty.
func (w *wrapper) autoAdjust() {
	if !w.Available() {
		return
	}

	// If the status code indicates a client or server error and no explicit error is set,
	// create an error from the message or a default message if the message is empty.
	if (w.IsClientError() || w.IsServerError()) && !w.IsErrorPresent() {
		if strutil.IsNotEmpty(w.message) {
			w.errors = NewError(w.message)
		} else {
			w.errors = NewErrorf("HTTP %d error with no message", w.StatusCode())
		}
	}

	// If an error is present but the status code indicates success, override the status code to 500 Internal Server Error
	// to maintain consistency between the presence of an error and the status code.
	if w.IsErrorPresent() && w.IsSuccess() {
		base := InternalServerError
		w.
			WithDebuggingKVf("status_code_override", "status code indicates %s but error is present; auto-corrected to %s",
				w.StatusText(), base.StatusText()).
			WithHeader(base)
	}
}

// build generates a map representation of the [wrapper] instance.
// This method collects various fields of the [wrapper] (e.g., `data`, [header], [meta], etc.)
// and organizes them into a key-value map. It ensures that only non-empty or meaningful fields
// are included in the resulting map, providing a clean and structured response.
// The following fields are included in the response:
//   - `data`: The primary data payload, if present.
//   - `headers`: The structured header details, if present.
//   - [meta]: Metadata about the response, if present.
//   - [pagination]: Pagination details, if applicable.
//   - `debug`: Debugging information, if provided.
//   - `total`: Total number of items, if set to a valid non-negative value.
//   - `status_code`: The HTTP status code, if greater than 0.
//   - `message`: A descriptive message, if not empty.
//   - `path`: The request path, if not empty.
//
// Returns:
//   - A `map[string]interface{}` containing the structured response data.
func (w *wrapper) build() map[string]any {
	m := make(map[string]any)
	if w.IsStatusCodePresent() {
		m["status_code"] = w.statusCode
	}
	if strutil.IsNotEmpty(w.path) {
		m["path"] = w.path
	}
	if strutil.IsNotEmpty(w.message) {
		m["message"] = w.message
	}
	if w.IsBodyPresent() && !w.skipBody {
		m["data"] = safeCastValueSupervised(w.data)
	}
	if w.IsTotalPresent() {
		m["total"] = w.total
	}
	if w.IsPagingPresent() {
		m["pagination"] = w.pagination.Respond()
	}
	if w.IsMetaPresent() {
		m["meta"] = w.meta.Respond()
	}
	if w.IsHeaderPresent() {
		m["header"] = w.header.Respond()
	}
	if w.IsDebuggingPresent() {
		m["debug"] = w.debug
	}
	if w.IsCursorPresent() {
		m["cursor"] = w.cursor.Respond()
	}
	return m
}

// Value returns the integer value of the StatusCode.
//
// This method allows for easy retrieval of the underlying integer value of a StatusCode instance,
// which can be useful for comparisons, logging, or when interfacing with APIs that require numeric status codes.
//
// Returns:
//   - An integer representing the value of the StatusCode.
func (s StatusCode) Value() int {
	return int(s)
}

// StatusText returns a string representation of the StatusCode, combining its integer value with the standard HTTP status text.
//
// This method formats the StatusCode as a string in the format "code (text)", where "code" is the integer value of the status code
// and "text" is the standard HTTP status text corresponding to that code (e.g., "200 (OK)", "404 (Not Found)").
//
// Returns:
//   - A string representing the StatusCode in a human-readable format.
func (s StatusCode) StatusText() string {
	return fmt.Sprintf("%d (%s)", s.Value(), http.StatusText(s.Value()))
}
