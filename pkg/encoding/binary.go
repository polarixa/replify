package encoding

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// IsBinary is the recursive worker behind IsBinaryBody. It resolves
// unambiguous types directly, falls back to content sniffing for strings,
// byte slices, and readers, and unwraps one level of pointer indirection
// for any other concrete type.
//
// The function returns true if the value is determined to be binary, and false otherwise.
//
// Parameters:
//   - value: The value to be checked for binary content. It can be of any type.
//
// Returns:
//   - A boolean indicating whether the provided value is binary (true) or not (false).
func IsBinary(value any) bool {
	switch v := value.(type) {
	case nil:
		return false
	// unambiguously textual, regardless of content
	case json.RawMessage:
		return false
	case *json.RawMessage:
		return false
	case error:
		return false
	case fmt.Stringer:
		return sniffString(v.String())
	case *fmt.Stringer:
		if v == nil {
			return false
		}
		return sniffString((*v).String())
	// strings.Reader only ever wraps text
	case *strings.Reader:
		return false
	// unambiguously binary-shaped containers
	case []byte:
		return sniffBytes(v)
	case *[]byte:
		if v == nil {
			return false
		}
		return sniffBytes(*v)
	// content-sniffed strings
	case string:
		return sniffString(v)
	case *string:
		if v == nil {
			return false
		}
		return sniffString(*v)
	// seekable readers (files, sysx.Resource content, bytes.Reader, ...)
	case io.ReadSeeker:
		return sniffSeekableReader(v)
	// non-seekable readers can't be sampled without consuming
	// the body meant for the response, so treat conservatively as binary.
	case io.Reader:
		return true
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return false
			}
			return IsBinary(rv.Elem().Interface())
		}
		return false
	}
}
