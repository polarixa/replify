package replify

import (
	"encoding/base64"

	"github.com/polarixa/replify/pkg/slogger"
)

// EncodeBodyBase64 encodes the body of the [wrapper] instance in Base64 and updates the body data.
//
// This function converts the body data of the [wrapper] instance to a Base64-encoded string and stores it back in the wrapper.
// It returns the updated [wrapper] instance to allow method chaining.
//
// Returns:
//   - The updated [wrapper] instance with the body encoded in Base64.
func (w *wrapper) EncodeBodyBase64() *wrapper {
	if !w.IsBodyPresent() {
		return w
	}
	// Suppress the current body before encoding it in Base64.
	// This ensures that the original body is preserved and can be restored if needed.
	w.WithBodySuppressed(w.Body())
	w.data = base64.StdEncoding.EncodeToString(w.BodyBytes())
	w.bodyBase64 = true
	return w
}

// DecodeBodyBase64 decodes the Base64-encoded body of the [wrapper] instance and updates the body data.
//
// This function checks if the body is marked as Base64-encoded before attempting to decode it.
// If the body is not marked as Base64-encoded, it logs a warning and skips decoding.
// If the decoding fails, it logs an error and returns the wrapper instance without modifying the body.
// It returns the updated [wrapper] instance to allow method chaining.
//
// Returns:
//   - The updated [wrapper] instance with the body decoded from Base64, or the original instance if decoding fails or is not applicable.
func (w *wrapper) DecodeBodyBase64() *wrapper {
	if !w.IsBodyPresent() {
		return w
	}
	// Check if the body is marked as Base64-encoded before attempting to decode it.
	// If the body is not marked as Base64-encoded, log a warning and skip decoding.
	if !w.bodyBase64 {
		slogger.Warnf("body is not Base64-encoded")
		return w
	}
	// Decode the Base64-encoded body and handle any errors that occur during decoding.
	decoded, err := base64.StdEncoding.DecodeString(w.BodyString())
	if err != nil {
		w.WithErrorAck(err)
		return w
	}
	w.data = string(decoded)
	w.bodyBase64 = false
	return w
}
