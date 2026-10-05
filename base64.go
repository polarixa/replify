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

// EncodeBodyBase64URL encodes the body of the [wrapper] instance using URL-safe Base64 encoding.
//
// This function converts the body data of the [wrapper] instance to a URL-safe Base64-encoded string
// and stores it back in the wrapper. It is particularly useful for safely transmitting binary data in URLs or query parameters.
// The original body is suppressed and preserved internally to allow for future restoration.
//
// Returns:
//   - The updated [wrapper] instance with the body encoded in URL-safe Base64.
func (w *wrapper) EncodeBodyBase64URL() *wrapper {
	if !w.IsBodyPresent() {
		return w
	}
	// Suppress the current body before encoding it in URL-safe Base64.
	w.WithBodySuppressed(w.Body())
	w.data = base64.URLEncoding.EncodeToString(w.BodyBytes())
	w.bodyBase64 = true
	return w
}

// DecodeBodyBase64URL decodes the URL-safe Base64-encoded body of the [wrapper] instance and updates the body data.
//
// This function checks if the body is marked as Base64-encoded before attempting to decode it.
// It first attempts standard URL-safe decoding, and falls back to Raw URL-safe decoding (unpadded) if standard decoding fails.
// If both decoding attempts fail, it logs an error acknowledgment and returns the wrapper instance without modifying the body.
//
// Returns:
//   - The updated [wrapper] instance with the body decoded from URL-safe Base64.
func (w *wrapper) DecodeBodyBase64URL() *wrapper {
	if !w.IsBodyPresent() {
		return w
	}
	if !w.bodyBase64 {
		slogger.Warnf("body is not Base64-encoded")
		return w
	}

	// Attempt standard URL-safe decoding first.
	decoded, err := base64.URLEncoding.DecodeString(w.BodyString())
	if err != nil {
		// Fallback to Raw URL-safe decoding for unpadded strings.
		decoded, err = base64.RawURLEncoding.DecodeString(w.BodyString())
		if err != nil {
			w.WithErrorAck(err)
			return w
		}
	}
	w.data = string(decoded)
	w.bodyBase64 = false
	return w
}

// EncodeBodyBase64Raw encodes the body of the [wrapper] instance using Raw (unpadded) standard Base64 encoding.
//
// This function converts the body data of the [wrapper] instance to an unpadded Base64 string
// and stores it back in the wrapper. Unpadded Base64 is often used in cryptographic contexts or JWTs.
//
// Returns:
//   - The updated [wrapper] instance with the body encoded in unpadded Base64.
func (w *wrapper) EncodeBodyBase64Raw() *wrapper {
	if !w.IsBodyPresent() {
		return w
	}
	w.WithBodySuppressed(w.Body())
	w.data = base64.RawStdEncoding.EncodeToString(w.BodyBytes())
	w.bodyBase64 = true
	return w
}

// DecodeBodyBase64Raw decodes the Raw (unpadded) Base64-encoded body of the [wrapper] instance.
//
// Returns:
//   - The updated [wrapper] instance with the body decoded from unpadded Base64.
func (w *wrapper) DecodeBodyBase64Raw() *wrapper {
	if !w.IsBodyPresent() {
		return w
	}
	if !w.bodyBase64 {
		slogger.Warnf("body is not Base64-encoded")
		return w
	}
	decoded, err := base64.RawStdEncoding.DecodeString(w.BodyString())
	if err != nil {
		w.WithErrorAck(err)
		return w
	}
	w.data = string(decoded)
	w.bodyBase64 = false
	return w
}
