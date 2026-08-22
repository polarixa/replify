package replify

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/encoding"
	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strutil"
	"github.com/polarixa/replify/pkg/sysx"
)

// calculateSize calculates the size of the marshaled data.
// It uses encoding.Marshal to marshal the data and returns the length of the resulting byte slice.
// If an error occurs during marshaling, it returns 0.
func calculateSize(data any) int {
	_bytes, err := encoding.MarshalJSON(data)
	if err != nil {
		return 0
	}
	return len(_bytes)
}

// compress compresses the given data using gzip and encodes it in base64.
// It first marshals the data using encoding.Marshal, then compresses the resulting byte slice
// using gzip. The compressed data is then encoded in base64 and returned as a string.
// If any error occurs during marshaling or compression, it returns an empty string.
func compress(data any) string {
	_bytes, err := encoding.MarshalJSON(data)
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err = gz.Write(_bytes)
	if err != nil {
		return ""
	}
	err = gz.Close()
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// decompress decompresses the given data using gzip and decodes it from base64.
// It first decodes the base64 encoded data using base64.StdEncoding.DecodeString,
// then decompresses the resulting byte slice using gzip. The decompressed data is
// then unmarshaled using encoding.Unmarshal and returned as an interface{}.
// If any error occurs during decoding or decompression, it returns nil.
func decompress(data string) any {
	_bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil
	}
	gz, err := gzip.NewReader(bytes.NewReader(_bytes))
	if err != nil {
		return nil
	}
	defer gz.Close()
	var buf bytes.Buffer
	_, err = buf.ReadFrom(gz)
	if err != nil {
		return nil
	}
	var result any
	if err := encoding.UnmarshalJSON(buf.Bytes(), &result); err != nil {
		return nil
	}
	return result
}

// chunk takes a response represented as a map and returns a slice of byte slices,
// where each byte slice is a chunk of the JSON representation of the response.
// This is useful for streaming large responses in smaller segments.
// If the JSON encoding fails, it returns nil.
func chunk(data map[string]any) [][]byte {
	_bytes, err := encoding.MarshalJSON(data)
	if err != nil {
		return nil
	}
	var chunks [][]byte
	for i := 0; i < len(_bytes); i += defaultChunkSize {
		end := i + defaultChunkSize
		if end > len(_bytes) {
			end = len(_bytes)
		}
		// Create a copy of the chunk to avoid referencing the underlying array.
		// This is important to ensure that each chunk is independent and can be
		// processed separately without affecting the others.
		chunk := make([]byte, end-i)
		copy(chunk, _bytes[i:end])
		chunks = append(chunks, chunk)
	}
	return chunks
}

// jsonpass converts a Go value to its JSON string representation or returns the value directly if it is already a string.
//
// This function checks if the input data is a string; if so, it returns it directly.
// Otherwise, it marshals the input value `data` into a JSON string using the
// MarshalToString function. If an error occurs during marshalling, it returns an empty string.
//
// Parameters:
//   - `data`: The Go value to be converted to JSON, or a string to be returned directly.
//
// Returns:
//   - A string containing the JSON representation of the input value, or an empty string if an error occurs.
//
// Example:
//
//	jsonStr := jsonpass(myStruct)
func jsonpass(data any) string {
	return encoding.JSON(data)
}

// jsonpretty converts a Go value to its pretty-printed JSON string representation or returns the value directly if it is already a string.
//
// This function checks if the input data is a string; if so, it returns it directly.
// Otherwise, it marshals the input value `data` into a formatted JSON string using
// the MarshalIndent function. If an error occurs during marshalling, it returns an empty string.
//
// Parameters:
//   - `data`: The Go value to be converted to pretty-printed JSON, or a string to be returned directly.
//
// Returns:
//   - A string containing the pretty-printed JSON representation of the input value, or an empty string if an error occurs.
//
// Example:
//
//	jsonPrettyStr := jsonpretty(myStruct)
func jsonpretty(data any) string {
	return encoding.JSONPretty(data)
}

// httpStatusLevel maps an HTTP status code to its corresponding [slogger.Level].
//
//   - 1xx → Debug  (informational)
//   - 2xx → Info   (success)
//   - 3xx → Warn   (redirection)
//   - 4xx → Error  (client error)
//   - 5xx → Error  (server error; Fatal is avoided — it calls os.Exit(1))
//   - other → Trace
func httpStatusLevel(code int) slogger.Level {
	switch {
	case code >= 400:
		return slogger.ErrorLevel
	case code >= 300:
		return slogger.WarnLevel
	case code >= 200:
		return slogger.InfoLevel
	case code >= 100:
		return slogger.DebugLevel
	default:
		return slogger.TraceLevel
	}
}

// logAtLevel dispatches a single log entry to l at the given level.
// It uses the appropriate method of the slogger.Logger based on the provided slogger.Level.
//
// Parameters:
//   - `l`: The slogger.Logger instance to which the log entry will be dispatched.
//   - `lvl`: The slogger.Level indicating the severity of the log entry (e.g., ErrorLevel, WarnLevel, InfoLevel, DebugLevel, TraceLevel).
//   - `msg`: The message string to be logged.
//   - `f`: A slogger.Field containing additional structured data to be included in the log entry.
//
// The function uses a switch statement to determine which logging method to call on the logger based on the provided level.
// If the level does not match any of the defined levels (ErrorLevel, WarnLevel, InfoLevel, DebugLevel), it defaults to using Trace.
func logAtLevel(l *slogger.Logger, lvl slogger.Level, msg string, f slogger.Field) {
	switch lvl {
	case slogger.ErrorLevel:
		l.Error(msg, f)
	case slogger.WarnLevel:
		l.Warn(msg, f)
	case slogger.InfoLevel:
		l.Info(msg, f)
	case slogger.DebugLevel:
		l.Debug(msg, f)
	default:
		l.Trace(msg, f)
	}
}

// slogAtLevel dispatches a single log entry to l at the given level without any structured fields.
// It uses the appropriate method of the slogger.Logger based on the provided slogger.Level.
//
// Parameters:
//   - `l`: The slogger.Logger instance to which the log entry will be dispatched.
//   - `lvl`: The slogger.Level indicating the severity of the log entry (e.g., ErrorLevel, WarnLevel, InfoLevel, DebugLevel).
//   - `msg`: The message string to be logged.
//
// The function uses a switch statement to determine which logging method to call on the logger based on the provided level.
// If the level does not match any of the defined levels (ErrorLevel, WarnLevel, InfoLevel, DebugLevel), it defaults to using Trace.
func slogAtLevel(l *slogger.Logger, lvl slogger.Level, msg string) {
	switch lvl {
	case slogger.ErrorLevel:
		l.Error(msg)
	case slogger.WarnLevel:
		l.Warn(msg)
	case slogger.InfoLevel:
		l.Info(msg)
	case slogger.DebugLevel:
		l.Debug(msg)
	default:
		l.Trace(msg)
	}
}

// dumpJSON creates a seekable in-process [sysx.Resource] backed by a
// temporary file from an already-serialized JSON payload. It is the shared
// core used by both [wrapper.Dump] and [wrapper.DumpTo], so the
// [sysx.NewResource] call site exists in exactly one place.
func dumpJSON(payload []byte) (*sysx.Resource, error) {
	return sysx.NewResource().
		WithName("w_snapshot.json").
		WithTempPattern("w_snapshot-*.json").
		WithContentType(sysx.MimeJSON).
		FromTempFile(func(w io.Writer) error {
			_, err := w.Write(payload)
			return err
		})
}

// dumpMarkdown creates a seekable in-process [sysx.Resource] backed by a
// temporary file from an already-serialized Markdown payload.
func dumpMarkdown(payload []byte) (*sysx.Resource, error) {
	return sysx.NewResource().
		WithName("w_snapshot.md").
		WithTempPattern("w_snapshot-*.md").
		WithContentType(sysx.MimeText).
		FromTempFile(func(w io.Writer) error {
			_, err := w.Write(payload)
			return err
		})
}

// dumpAny creates a seekable in-process [sysx.Resource] backed by a
// temporary file from a generic Go value. The value is first converted to a
// string using [conv.String].
func dumpAny(payload any) (*sysx.Resource, error) {
	var body string
	body, err := conv.String(safeCastValue(payload))
	if err != nil {
		return nil, err
	}
	return sysx.NewResource().
		WithName("w_snapshot_body.txt").
		WithTempPattern("w_snapshot_body-*.txt").
		WithContentType(sysx.MimeText).
		FromTempFile(func(w io.Writer) error {
			_, err := w.Write([]byte(body))
			return err
		})
}

// escapeMarkdownPipe escapes characters that would break Markdown
// table rendering (pipe characters and newlines).
func escapeMarkdownPipe(s string) string {
	if s == "" {
		return s
	}
	s = strings.ReplaceAll(s, "|", `\|`)
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// parseStackFrame parses a fully-qualified Go function name into a display name,
// the bare function/method name used as a call label, and whether it is a runtime frame.
//
// It handles package names, receiver types, and method names, including special cases for runtime functions.
//
// Examples:
//
//	"github.com/polarixa/replify/pkg/strutil.(*StringWeaver).Append" → displayName: "strutil.StringWeaver.Append", callLabel: "Append", isRuntime: false
//	"runtime.main" → displayName: "runtime", callLabel: "main", isRuntime: true
//	"net/http.(*Server).Serve" → displayName: "http.Server.Serve", callLabel: "Serve", isRuntime: false
func parseStackFrame(fullName string) (displayName, callLabel string, isRuntime bool) {
	i := strings.LastIndex(fullName, "/")
	short := fullName[i+1:]

	// Special case for runtime package functions
	if after, ok := strings.CutPrefix(short, "runtime."); ok {
		return "runtime", after, true
	}

	before, after, ok := strings.Cut(short, ".")
	if !ok {
		return short, short, false
	}
	pkg := before
	rest := after

	// Receiver method: "(*Type).Method" or "(Type).Method"
	if strings.HasPrefix(rest, "(") {
		closeIdx := strings.Index(rest, ")")
		if closeIdx > 0 {
			typePart := strings.TrimPrefix(rest[1:closeIdx], "*")
			if closeIdx+2 < len(rest) {
				method := rest[closeIdx+2:]
				return typePart + "." + method, method, false
			}
			return typePart, typePart, false
		}
	}

	return pkg + "." + rest, rest, false
}

// parseStackFrameParticipants parses the stack frame lines and extracts the unique participants for the sequence diagram.
// It returns a slice of [sequenceParticipant] structs representing the participants in the call chain.
//
// The function skips frames that are part of the runtime epilogue, as they do not provide meaningful diagnostic information.
// It also ensures that each participant is unique by using a map to track seen display names.
func parseStackFrameParticipants(lines []string) []sequenceParticipant {
	var participants []sequenceParticipant
	seen := make(map[string]bool)

	if len(lines) == 0 {
		return participants
	}

	// Skip frames that are part of the runtime epilogue,
	// as they do not provide meaningful diagnostic information.
	skipFrames := []string{
		"runtime.goexit",
		"runtime.mstart",
	}

	for _, line := range lines {
		spaceIdx := strings.LastIndex(line, " ")
		funcFull := strings.TrimSpace(line)
		location := ""
		if spaceIdx > 0 {
			funcFull = strings.TrimSpace(line[:spaceIdx])
			raw := strings.TrimSpace(line[spaceIdx+1:])
			// Shorten the absolute path to just "file.go:line"
			if slashIdx := strings.LastIndex(raw, "/"); slashIdx >= 0 {
				location = raw[slashIdx+1:]
			} else {
				location = raw
			}
		}
		skip := false

		// Skip frames that are part of the runtime epilogue,
		// as they do not provide meaningful diagnostic information.
		for _, frame := range skipFrames {
			if strings.Contains(funcFull, frame) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Parse the function name to extract the display name and call label for the sequence diagram.
		displayName, callLabel, isRuntime := parseStackFrame(funcFull)
		if isRuntime {
			displayName = "runtime"
		}
		if seen[displayName] {
			continue
		}
		seen[displayName] = true

		participants = append(participants, sequenceParticipant{
			displayName: displayName,
			callLabel:   callLabel,
			isRuntime:   isRuntime,
			location:    location,
		})
	}

	return participants
}

// castString attempts to convert a string value into a JSON-compatible representation.
// If the string is a valid JSON string, it returns a json.RawMessage containing the compacted JSON.
// If the string is not valid JSON, it returns the original string value.
//
// Parameters:
//   - value: The input string to be cast.
//
// Returns:
//   - as: The resulting value, which may be a json.RawMessage or the original string.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castString(value *string) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "string")
	if value == nil || strutil.IsEmpty(*value) {
		return value, w.OK()
	}

	// try to sanitize the string value for JSON parsing
	sanitizeValue, err := encoding.NormalizeJSON(*value)
	if err != nil {
		if strutil.IsEmpty(sanitizeValue) {
			return value, w.
				WithHeader(BadRequest).
				WithErrorAck(err).
				WithMessage("failed to sanitize string value for JSON parsing")
		}
	}

	// if the sanitized value is a valid JSON string, return it as json.RawMessage
	if encoding.IsValidJSONString(sanitizeValue) {
		as = json.RawMessage(encoding.CompactJSON([]byte(sanitizeValue)))
		return as, w.OK()
	}

	as = sanitizeValue
	return as, w.OK()
}

// castBytes attempts to convert a byte slice into a JSON-compatible representation.
// If the byte slice is a valid JSON byte slice, it returns a json.RawMessage containing the compacted JSON.
// If the byte slice is not valid JSON, it returns the original byte slice.
//
// Parameters:
//   - value: A pointer to the input byte slice to be cast.
//
// Returns:
//   - as: The resulting value, which may be a json.RawMessage or the original byte slice.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castBytes(value *[]byte) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "[]byte")
	if value == nil || len(*value) == 0 {
		return value, w.OK()
	}

	sanitizeValue := value

	// if the sanitized value is a valid JSON byte slice, return it as json.RawMessage
	if encoding.IsValidJSON(*sanitizeValue) {
		as = json.RawMessage(encoding.CompactJSON(*sanitizeValue))
		return as, w.OK()
	}
	as = string(*sanitizeValue)
	return as, w.OK()
}

// castBool attempts to convert a boolean pointer into a JSON-compatible representation.
// If the boolean pointer is nil, it returns nil. Otherwise, it returns the dereferenced boolean value.
//
// Parameters:
//   - value: A pointer to the input boolean value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a boolean or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castBool(value *bool) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "bool")
	if value == nil {
		return value, w.OK()
	}
	as = *value
	return as, w.OK()
}

// castInt attempts to convert an integer pointer into a JSON-compatible representation.
// If the integer pointer is nil, it returns nil. Otherwise, it returns the dereferenced integer value.
//
// Parameters:
//   - value: A pointer to the input integer value to be cast.
//
// Returns:
//   - as: The resulting value, which may be an integer or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castInt(value *int) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "int")
	if value == nil {
		return value, w.OK()
	}
	as = *value
	return as, w.OK()
}

// castInt8 attempts to convert an int8 pointer into a JSON-compatible representation.
// If the int8 pointer is nil, it returns nil. Otherwise, it returns the dereferenced int8 value.
//
// Parameters:
//   - value: A pointer to the input int8 value to be cast.
//
// Returns:
//   - as: The resulting value, which may be an int8 or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castInt8(value *int8) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "int8")
	if value == nil {
		return value, w.OK()
	}
	as = *value
	return as, w.OK()
}

// castInt16 attempts to convert an int16 pointer into a JSON-compatible representation.
// If the int16 pointer is nil, it returns nil. Otherwise, it returns the dereferenced int16 value.
//
// Parameters:
//   - value: A pointer to the input int16 value to be cast.
//
// Returns:
//   - as: The resulting value, which may be an int16 or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castInt16(value *int16) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "int16")
	if value == nil {
		return value, w.OK()
	}
	as = *value
	return as, w.OK()
}

// castInt32 attempts to convert an int32 pointer into a JSON-compatible representation.
// If the int32 pointer is nil, it returns nil. Otherwise, it returns the dereferenced int32 value.
//
// Parameters:
//   - value: A pointer to the input int32 value to be cast.
//
// Returns:
//   - as: The resulting value, which may be an int32 or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castInt32(value *int32) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "int32")
	if value == nil {
		return value, w.OK()
	}
	as = *value
	return as, w.OK()
}

// castInt64 attempts to convert an int64 pointer into a JSON-compatible representation.
// If the int64 pointer is nil, it returns nil. Otherwise, it returns the dereferenced int64 value.
//
// Parameters:
//   - value: A pointer to the input int64 value to be cast.
//
// Returns:
//   - as: The resulting value, which may be an int64 or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castInt64(value *int64) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "int64")
	if value == nil {
		return value, w.OK()
	}
	as = *value
	return as, w.OK()
}

// castUint attempts to convert a uint pointer into a JSON-compatible representation.
// If the uint pointer is nil, it returns nil. Otherwise, it returns the dereferenced uint value.
//
// Parameters:
//   - value: A pointer to the input uint value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a uint or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castUint(value *uint) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "uint")
	if value == nil {
		return value, w.OK()
	}
	as = *value
	return as, w.OK()
}

// castUint8 attempts to convert a uint8 pointer into a JSON-compatible representation.
// If the uint8 pointer is nil, it returns nil. Otherwise, it returns the dereferenced uint8 value.
//
// Parameters:
//   - value: A pointer to the input uint8 value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a uint8 or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castUint8(value *uint8) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "uint8")
	if value == nil {
		return value, w.OK()
	}
	as = *value
	return as, w.OK()
}

// castUint16 attempts to convert a uint16 pointer into a JSON-compatible representation.
// If the uint16 pointer is nil, it returns nil. Otherwise, it returns the dereferenced uint16 value.
//
// Parameters:
//   - value: A pointer to the input uint16 value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a uint16 or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castUint16(value *uint16) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "uint16")
	if value == nil {
		return value, w.OK()
	}
	as = *value
	return as, w.OK()
}

// castUint32 attempts to convert a uint32 pointer into a JSON-compatible representation.
// If the uint32 pointer is nil, it returns nil. Otherwise, it returns the dereferenced uint32 value.
//
// Parameters:
//   - value: A pointer to the input uint32 value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a uint32 or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castUint32(value *uint32) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "uint32")
	if value == nil {
		return value, w.OK()
	}
	as = *value
	return as, w.OK()
}

// castUint64 attempts to convert a uint64 pointer into a JSON-compatible representation.
// If the uint64 pointer is nil, it returns nil. Otherwise, it returns the dereferenced uint64 value.
//
// Parameters:
//   - value: A pointer to the input uint64 value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a uint64 or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castUint64(value *uint64) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "uint64")
	if value == nil {
		return value, w.OK()
	}
	as = *value
	return as, w.OK()
}

// castFloat32 attempts to convert a float32 pointer into a JSON-compatible representation.
// If the float32 pointer is nil, it returns nil. Otherwise, it returns the dereferenced float32 value.
//
// Parameters:
//   - value: A pointer to the input float32 value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a float32 or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castFloat32(value *float32) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "float32")
	if value == nil {
		return value, w.OK()
	}
	as = *value
	return as, w.OK()
}

// castFloat64 attempts to convert a float64 pointer into a JSON-compatible representation.
// If the float64 pointer is nil, it returns nil. Otherwise, it returns the dereferenced float64 value.
//
// Parameters:
//   - value: A pointer to the input float64 value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a float64 or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castFloat64(value *float64) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "float64")
	if value == nil {
		return value, w.OK()
	}
	as = *value
	return as, w.OK()
}

// castComplex64 attempts to convert a complex64 pointer into a JSON-compatible representation.
// If the complex64 pointer is nil, it returns nil. Otherwise, it converts the complex64 value to a string representation.
//
// Parameters:
//   - value: A pointer to the input complex64 value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a string representation of the complex64 or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castComplex64(value *complex64) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "complex64")
	if value == nil {
		return value, w.OK()
	}
	val, err := conv.String(*value)
	if err != nil {
		return nil, w.
			WithHeader(BadRequest).
			WithErrorAck(err).
			WithMessage("failed to convert complex64 to string")
	}
	as = val
	return as, w.OK()
}

// castComplex128 attempts to convert a complex128 pointer into a JSON-compatible representation.
// If the complex128 pointer is nil, it returns nil. Otherwise, it converts the complex128 value to a string representation.
//
// Parameters:
//   - value: A pointer to the input complex128 value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a string representation of the complex128 or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castComplex128(value *complex128) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "complex128")
	if value == nil {
		return value, w.OK()
	}
	val, err := conv.String(*value)
	if err != nil {
		return nil, w.
			WithHeader(BadRequest).
			WithErrorAck(err).
			WithMessage("failed to convert complex128 to string")
	}
	as = val
	return as, w.OK()
}

// castTime attempts to convert a time.Time pointer into a JSON-compatible representation.
// If the time.Time pointer is nil, it returns nil. Otherwise, it formats the time value as an RFC3339Nano string.
//
// Parameters:
//   - value: A pointer to the input time.Time value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a formatted string representation of the time.Time or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castTime(value *time.Time) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "time.Time")
	if value == nil {
		return value, w.OK()
	}
	as = value.Format(time.RFC3339Nano)
	return as, w.OK()
}

// castDuration attempts to convert a time.Duration pointer into a JSON-compatible representation.
// If the time.Duration pointer is nil, it returns nil. Otherwise, it converts the duration to its string representation.
//
// Parameters:
//   - value: A pointer to the input time.Duration value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a string representation of the time.Duration or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castDuration(value *time.Duration) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "time.Duration")
	if value == nil {
		return value, w.OK()
	}
	as = value.String()
	return as, w.OK()
}

// castError attempts to convert an error value into a JSON-compatible representation.
// If the error value is nil, it returns nil. Otherwise, it returns the error message as a string.
//
// Parameters:
//   - value: The input error value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a string representation of the error or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castError(value error) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "error")
	if value == nil {
		return value, w.OK()
	}
	as = value.Error()
	return as, w.OK()
}

// castFmtStringer attempts to convert a fmt.Stringer pointer into a JSON-compatible representation.
// If the fmt.Stringer pointer is nil, it returns nil. Otherwise, it calls the String() method and returns the result.
//
// Parameters:
//   - value: A pointer to the input fmt.Stringer value to be cast.
//
// Returns:
//   - as: The resulting value, which may be a string representation of the fmt.Stringer or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castFmtStringer(value *fmt.Stringer) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "fmt.Stringer")
	if value == nil {
		return value, w.OK()
	}
	as = (*value).String()
	return as, w.OK()
}

// castRunes attempts to convert a slice of runes into a JSON-compatible representation.
// If the slice of runes is nil or empty, it returns nil. Otherwise, it converts the slice of runes to a string.
//
// Parameters:
//   - value: A pointer to the input slice of runes to be cast.
//
// Returns:
//   - as: The resulting value, which may be a string representation of the slice of runes or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castRunes(value *[]rune) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "[]rune")
	if value == nil || len(*value) == 0 {
		return value, w.OK()
	}

	as = string(*value)
	return as, w.OK()
}

// castJSONRawMessage attempts to convert a json.RawMessage pointer into a JSON-compatible representation.
// If the json.RawMessage pointer is nil or empty, it returns nil. Otherwise, it converts the json.RawMessage to a string.
//
// Parameters:
//   - value: A pointer to the input json.RawMessage to be cast.
//
// Returns:
//   - as: The resulting value, which may be a string representation of the json.RawMessage or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castJSONRawMessage(value *json.RawMessage) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "json.RawMessage")
	if value == nil || len(*value) == 0 {
		return value, w.OK()
	}

	as = string(*value)
	return as, w.OK()
}

// castJSONMarshaler attempts to convert a json.Marshaler pointer into a JSON-compatible representation.
// If the json.Marshaler pointer is nil, it returns nil. Otherwise, it calls the MarshalJSON() method and returns the result as a string.
//
// Parameters:
//   - value: A pointer to the input json.Marshaler to be cast.
//
// Returns:
//   - as: The resulting value, which may be a string representation of the json.Marshaler or nil.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castJSONMarshaler(value *json.Marshaler) (as any, w *wrapper) {
	w = New().Processing().WithDebuggingKV("type", "json.Marshaler")
	if value == nil {
		return value, w.OK()
	}

	bytes, err := (*value).MarshalJSON()
	if err != nil {
		return nil, w.
			WithHeader(BadRequest).
			WithErrorAck(err).
			WithMessagef("MarshalJSON() failed for value of type %T", *value)
	}

	as = string(bytes)
	return as, w.OK()
}

// castValue attempts to convert a generic Go value into a JSON-compatible representation.
// It handles various types, including strings, byte slices, runes, booleans, integers, floats, complex numbers, time values, errors, and JSON-related types.
// If the value is not one of the recognized types, it attempts to marshal it into JSON.
//
// Parameters:
//   - value: The input value of any type to be cast.
//
// Returns:
//   - as: The resulting value, which may be a JSON-compatible representation or the original value.
//   - w: A pointer to a wrapper instance indicating the status of the operation.
func castValue(value any) (as any, w *wrapper) {
	w = New().Processing()
	switch v := value.(type) {
	case string:
		return castString(&v)
	case *string:
		return castString(v)
	case []byte:
		return castBytes(&v)
	case *[]byte:
		return castBytes(v)
	case []rune:
		return castRunes(&v)
	case *[]rune:
		return castRunes(v)
	case bool:
		return castBool(&v)
	case *bool:
		return castBool(v)
	case int:
		return castInt(&v)
	case *int:
		return castInt(v)
	case int8:
		return castInt8(&v)
	case *int8:
		return castInt8(v)
	case int16:
		return castInt16(&v)
	case *int16:
		return castInt16(v)
	case int32:
		return castInt32(&v)
	case *int32:
		return castInt32(v)
	case int64:
		return castInt64(&v)
	case *int64:
		return castInt64(v)
	case uint:
		return castUint(&v)
	case *uint:
		return castUint(v)
	case uint8:
		return castUint8(&v)
	case *uint8:
		return castUint8(v)
	case uint16:
		return castUint16(&v)
	case *uint16:
		return castUint16(v)
	case uint32:
		return castUint32(&v)
	case *uint32:
		return castUint32(v)
	case uint64:
		return castUint64(&v)
	case *uint64:
		return castUint64(v)
	case float32:
		return castFloat32(&v)
	case *float32:
		return castFloat32(v)
	case float64:
		return castFloat64(&v)
	case *float64:
		return castFloat64(v)
	case complex64:
		return castComplex64(&v)
	case *complex64:
		return castComplex64(v)
	case complex128:
		return castComplex128(&v)
	case *complex128:
		return castComplex128(v)
	case time.Time:
		return castTime(&v)
	case *time.Time:
		return castTime(v)
	case time.Duration:
		return castDuration(&v)
	case *time.Duration:
		return castDuration(v)
	case error:
		return castError(v)
	case fmt.Stringer:
		return castFmtStringer(&v)
	case *fmt.Stringer:
		return castFmtStringer(v)
	case json.RawMessage:
		return castJSONRawMessage(&v)
	case *json.RawMessage:
		return castJSONRawMessage(v)
	case json.Marshaler:
		return castJSONMarshaler(&v)
	case *json.Marshaler:
		return castJSONMarshaler(v)
	default:
		jsonVal, err := encoding.JSONE(v)
		if err != nil {
			return nil, w.
				WithHeader(BadRequest).
				WithErrorAck(err).
				WithMessagef("cannot marshal value of type %T to JSON", v)
		}
		as = jsonVal
		return as, w.OK()
	}
}

// safeCastValue attempts to convert a generic Go value into a JSON-compatible representation, similar to CastValue.
// It returns only the resulting value, ignoring any wrapper information about the operation's status.
//
// Parameters:
//   - value: The input value of any type to be cast.
//
// Returns:
//   - as: The resulting value, which may be a JSON-compatible representation or the original value.
func safeCastValue(value any) (as any) {
	as, _ = castValue(value)
	return as
}
