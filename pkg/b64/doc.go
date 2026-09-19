// Package base64validator provides an object-oriented, extensible way to
// validate whether a string is valid Base64-encoded data.
//
// The original function under review only checked the standard Base64
// alphabet (base64.StdEncoding) and had a few subtle gaps:
//
//   - It couldn't validate URL-safe Base64 (base64.URLEncoding) or
//     unpadded variants (RawStdEncoding / RawURLEncoding) without
//     duplicating the whole function.
//   - There was no way to plug in a custom/alternate encoding for testing
//     or for services that use a different Base64 dialect.
//   - The "encode-then-compare" round-trip check is correct (it catches
//     non-canonical padding, wrong alphabets, and stray whitespace), but
//     it was buried in a free function with no seams for reuse or mocking.
//
// This package keeps that same correct round-trip strategy but wraps it in
// a small, composable, interface-driven design so callers can validate
// against one specific encoding or check a string against several
// encodings at once.
package b64
