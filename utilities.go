package replify

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/encoding"
	"github.com/polarixa/replify/pkg/slogger"
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

// safeBody checks if the provided value is a valid JSON string or byte slice and returns a safe representation.
//
// This function takes an input value and determines if it is a valid JSON string or byte slice.
// If the value is a valid JSON string, it returns a `json.RawMessage` containing the JSON data.
// If the value is a valid JSON byte slice, it also returns a `json.RawMessage` containing the JSON data.
// For any other type of value, it returns the original value as is.
//
// Parameters:
//   - value: The input value to be checked and processed.
//
// Returns:
//   - A `json.RawMessage` if the input is a valid JSON string or byte slice.
//   - The original value for any other type of input.
func safeBody(value any) any {
	var result any
	switch v := value.(type) {
	case string:
		if encoding.IsValidJSONString(v) {
			result = json.RawMessage(encoding.CompactJSON([]byte(v)))
		} else {
			result = v
		}
	case []byte:
		if encoding.IsValidJSON(v) {
			result = json.RawMessage(encoding.CompactJSON(v))
		} else {
			result = string(v)
		}
	case *[]byte:
		if v != nil && encoding.IsValidJSON(*v) {
			result = json.RawMessage(encoding.CompactJSON(*v))
		} else if v != nil {
			result = string(*v)
		} else {
			result = nil
		}
	default:
		result = value
	}

	return result
}

// dumpJSON creates a seekable in-process [sysx.Resource] backed by a
// temporary file from an already-serialized JSON payload. It is the shared
// core used by both [wrapper.Dump] and [wrapper.DumpTo], so the
// [sysx.NewResource] call site exists in exactly one place.
func dumpJSON(payload []byte) (*sysx.Resource, error) {
	return sysx.NewResource().
		WithName("replify-dump-*.json").
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
		WithName("replify-dump-*.md").
		WithContentType(sysx.MimeText).
		FromTempFile(func(w io.Writer) error {
			_, err := w.Write(payload)
			return err
		})
}

// dumpBodyStream serializes body into a seekable [sysx.Resource] backed by a
// spill buffer and writes output to a plain-text (.txt) file because the body
// can carry any Go value — not only valid JSON.
//
// Serialization strategy (type-switch):
//   - string       → written as raw UTF-8, no quoting.
//   - []byte       → written as raw bytes, no encoding.
//   - json.RawMessage → written as-is (already serialized JSON).
//   - anything else → JSON-encoded via [json.Encoder]; output is valid JSON
//     followed by a single newline, which is idiomatic for line-delimited logs.
//
// # Memory / large-body handling
//
// The body is streamed through an [io.Pipe] into [sysx.Resource.FromReader]
// which manages a spill buffer: payloads up to [sysx.DefaultSpillThreshold]
// (8 MiB) are held in memory; beyond that, the overflow is transparently
// spilled to a self-removing temp file. The full body is never allocated as
// a single []byte.
//
// # Thread-safety
//
// dumpBodyStream is safe for concurrent invocation:
//   - Each call creates an independent [io.Pipe] pair, goroutine, and
//     [sysx.Resource] instance. No package-level or wrapper state is mutated
//     after the function returns.
//   - The producer goroutine only reads the body value that was passed in.
//     Callers must not mutate the body (or any data it references) while the
//     returned Resource is open; this is the same contract as [json.Marshal].
//   - Error propagation between the goroutine and the consumer is done
//     entirely through [io.PipeWriter.CloseWithError], which is the
//     synchronization primitive [io.Pipe] is designed for. A dedicated
//     channel carries the goroutine's final error so the consumer can
//     distinguish a write failure from an EOF.
//
// The caller owns the returned Resource and must call Close exactly once to
// release the underlying buffer or temp file.
func dumpBodyStream(body any) (*sysx.Resource, error) {
	pr, pw := io.Pipe()
	// errCh carries the producer's terminal error (nil on success) so that
	// FromReader can propagate it to the caller without a data race.
	errCh := make(chan error, 1)
	go func() {
		var err error
		v, err := conv.String(body)
		if err != nil {
			errCh <- err
			pw.CloseWithError(err)
			return
		}
		_, err = io.WriteString(pw, v)
		if err != nil {
			errCh <- err
			pw.CloseWithError(err)
			return
		}
		errCh <- nil
		pw.Close()
	}()

	res, err := sysx.NewResource().
		WithName("replify-dump-body-*.txt").
		WithContentType(sysx.MimeText).
		FromReader(pr)
	if err != nil {
		// Unblock the goroutine so it does not leak.
		pr.CloseWithError(err)
		<-errCh
		return nil, err
	}

	// Drain the channel so the goroutine is fully released before we return.
	if gErr := <-errCh; gErr != nil {
		res.Close()
		return nil, gErr
	}
	return res, nil
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
