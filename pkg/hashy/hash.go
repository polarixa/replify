package hashy

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"hash/crc32"
	"hash/crc64"
	"hash/fnv"
	"reflect"
	"strconv"

	"github.com/polarixa/replify/pkg/strutil"
)

// NewHash creates a new hash.Hash instance for the given algorithm.
//
// Parameters:
//   - algo: The hash algorithm to use.
//
// Returns:
//   - hash.Hash: The hash.Hash instance.
func NewHash(algo HashAlgorithm) hash.Hash {
	switch algo {
	case H_CRC32:
		return crc32.NewIEEE()
	case H_CRC64:
		return crc64.New(crc64.MakeTable(crc64.ISO))
	case H_MD5:
		return md5.New()
	case H_SHA1:
		return sha1.New()
	case H_SHA224:
		return sha256.New224()
	case H_SHA256:
		return sha256.New()
	case H_SHA384:
		return sha512.New384()
	case H_SHA512:
		return sha512.New()
	case H_SHA512_224:
		return sha512.New512_224()
	case H_SHA512_256:
		return sha512.New512_256()
	default:
		return fnv.New64a()
	}
}

// NewHash64 creates a new hash.Hash64 instance for the given algorithm.
// Algorithms that do not natively implement hash.Hash64 (e.g. MD5, SHA-*)
// fall back to FNV-1a rather than panicking.
//
// Parameters:
//   - algo: The hash algorithm to use.
//
// Returns:
//   - hash.Hash64: The hash.Hash64 instance.
func NewHash64(algo HashAlgorithm) hash.Hash64 {
	h := NewHash(algo)
	if h == nil {
		return fnv.New64a()
	}
	if h64, ok := h.(hash.Hash64); ok {
		return h64
	}
	// Algorithms such as MD5 and SHA-* implement hash.Hash but not hash.Hash64.
	// Fall back to FNV-1a to avoid a runtime panic.
	return fnv.New64a()
}

// HashValue generates a 64-bit hash value for a single value with options.
// This is the primary hashing function.
//
// Parameters:
//   - value: Any Go value (struct, slice, map, primitive, etc.)
//   - options: Optional configuration (nil uses defaults)
//
// Returns:
//   - uint64: The computed hash value (never zero for valid inputs)
//   - error: Non-nil if hashing fails
//
// Concurrency: HashValue is safe for concurrent use provided that either
// (a) options is nil, or (b) the options were built with WithHasherFunc so
// that a fresh hash.Hash64 is created for every call. Reusing options that
// carry a single hash.Hash64 instance (via WithHasher) from multiple
// goroutines simultaneously causes a data race.
//
// Example:
//
//	value := 1
//	hash, err := HashValue(value, nil)
//	fmt.Println(hash, err) // 1 nil
func HashValue(value any, options *hashOptions) (uint64, error) {
	if options == nil {
		options = DefaultOptions()
	}

	if err := options.validate(); err != nil {
		return 0, err
	}

	// Prefer HasherFunc (factory) over a shared instance for concurrency safety.
	var h hash.Hash64
	if options.HasherFunc != nil {
		h = options.HasherFunc()
	} else {
		h = options.Hasher
	}

	hasher := &hasher{
		hash:            h,
		tagName:         options.TagName,
		treatNilAsZero:  options.ZeroNil,
		ignoreZeroValue: options.IgnoreZeroValue,
		slicesAsSets:    options.SlicesAsSets,
		useStringer:     options.UseStringer,
	}

	hasher.hash.Reset()
	return hasher.hashValue(reflect.ValueOf(value), nil)
}

// Hash generates a 64-bit hash value for the given data.
// It accepts variadic arguments - if only one argument is provided, it hashes
// that value. If multiple arguments are provided, it hashes them as a tuple.
// The last argument can optionally be *Options.
//
// Examples:
//
//	hash, err := Hash(myStruct)                    // Single value
//	hash, err := Hash(val1, val2, val3)            // Multiple values
//	hash, err := Hash(myStruct, opts)              // With options
//	hash, err := Hash(val1, val2, val3, opts)      // Multiple values with options
//
// The hash is deterministic: identical values always produce identical hashes.
//
// Returns:
//   - uint64: The computed hash value (never zero for valid inputs)
//   - error: Non-nil if hashing fails
func Hash(data ...any) (uint64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("pkg.hash: no data provided")
	}

	var opts *hashOptions
	var values []any

	if len(data) > 0 {
		if o, ok := data[len(data)-1].(*hashOptions); ok {
			opts = o
			values = data[:len(data)-1]
		} else {
			values = data
		}
	}

	if len(values) == 0 {
		return 0, fmt.Errorf("pkg.hash: no data provided")
	}

	// If single value, hash it directly
	if len(values) == 1 {
		return HashValue(values[0], opts)
	}

	return HashValue(values, opts)
}

// Hash10 generates a decimal hash string for the given data.
// It accepts variadic arguments - if only one argument is provided, it hashes
// that value. If multiple arguments are provided, it hashes them as a tuple.
// The last argument can optionally be *Options.
//
// Returns:
//   - string: The computed hash string (never empty for valid inputs)
//   - error: Non-nil if hashing fails
//
// Example:
//
//	hash, err := Hash10(myStruct)                    // Single value
//	hash, err := Hash10(val1, val2, val3)            // Multiple values
//	hash, err := Hash10(myStruct, opts)              // With options
//	hash, err := Hash10(val1, val2, val3, opts)      // Multiple values with options
//
// The hash is deterministic: identical values always produce identical hashes.
func Hash10(data ...any) (string, error) {
	hash, err := Hash(data...)
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(hash, 10), nil
}

// Hash16 generates a hexadecimal hash string for the given data.
// It accepts variadic arguments - if only one argument is provided, it hashes
// that value. If multiple arguments are provided, it hashes them as a tuple.
// The last argument can optionally be *Options.
//
// Returns:
//   - string: The computed hash string (never empty for valid inputs)
//   - error: Non-nil if hashing fails
//
// Example:
//
//	hash, err := Hash16(myStruct)                    // Single value
//	hash, err := Hash16(val1, val2, val3)            // Multiple values
//	hash, err := Hash16(myStruct, opts)              // With options
//	hash, err := Hash16(val1, val2, val3, opts)      // Multiple values with options
//
// The hash is deterministic: identical values always produce identical hashes.
func Hash16(data ...any) (string, error) {
	hash, err := Hash(data...)
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(hash, 16), nil
}

// Hash16Padded generates a hexadecimal hash string for the given data.
// It accepts variadic arguments - if only one argument is provided, it hashes
// that value. If multiple arguments are provided, it hashes them as a tuple.
// The last argument can optionally be *Options.
//
// Returns:
//   - string: The computed hash string (never empty for valid inputs)
//   - error: Non-nil if hashing fails
//
// Example:
//
//	hash, err := Hash16Padded(myStruct)                    // Single value
//	hash, err := Hash16Padded(val1, val2, val3)            // Multiple values
//	hash, err := Hash16Padded(myStruct, opts)              // With options
//	hash, err := Hash16Padded(val1, val2, val3, opts)      // Multiple values with options
//
// The hash is deterministic: identical values always produce identical hashes.
func Hash16Padded(data ...any) (string, error) {
	hash, err := Hash(data...)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%016x", hash), nil
}

// Hash32 generates a base32 encoded hash string for the given data.
// It accepts variadic arguments - if only one argument is provided, it hashes
// that value. If multiple arguments are provided, it hashes them as a tuple.
// The last argument can optionally be *Options.
//
// Returns:
//   - string: The computed hash string (never empty for valid inputs)
//   - error: Non-nil if hashing fails
//
// Example:
//
//	hash, err := Hash32(myStruct)                    // Single value
//	hash, err := Hash32(val1, val2, val3)            // Multiple values
//	hash, err := Hash32(myStruct, opts)              // With options
//	hash, err := Hash32(val1, val2, val3, opts)      // Multiple values with options
//
// The hash is deterministic: identical values always produce identical hashes.§
func Hash32(data ...any) (string, error) {
	hash, err := Hash(data...)
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(hash, 32), nil
}

// Hash64 generates a base64 encoded hash string for the given data.
// It accepts variadic arguments - if only one argument is provided, it hashes
// that value. If multiple arguments are provided, it hashes them as a tuple.
// The last argument can optionally be *Options.
//
// Returns:
//   - string: The computed hash string (never empty for valid inputs)
//   - error: Non-nil if hashing fails
//
// Example:
//
//	hash, err := Hash64(myStruct)                    // Single value
//	hash, err := Hash64(val1, val2, val3)            // Multiple values
//	hash, err := Hash64(myStruct, opts)              // With options
//	hash, err := Hash64(val1, val2, val3, opts)      // Multiple values with options
//
// The hash is deterministic: identical values always produce identical hashes.
func Hash64(data ...any) (string, error) {
	hash, err := Hash(data...)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString([]byte(fmt.Appendf(nil, "%v", hash))), nil
}

// Hash256 generates a 256-bit hash string for the given data.
// It accepts variadic arguments - if only one argument is provided, it hashes
// that value. If multiple arguments are provided, it hashes them as a tuple.
// The last argument can optionally be *Options.
//
// Returns:
//   - string: The computed hash string (never empty for valid inputs)
//   - error: Non-nil if hashing fails
//
// Example:
//
//	hash, err := Hash256(myStruct)                    // Single value
//	hash, err := Hash256(val1, val2, val3)            // Multiple values
//	hash, err := Hash256(myStruct, opts)              // With options
//	hash, err := Hash256(val1, val2, val3, opts)      // Multiple values with options
//
// The hash is deterministic: identical values always produce identical hashes.
func Hash256(data ...any) (string, error) {
	hash, err := Hash(data...)
	if err != nil {
		return "", err
	}
	return strutil.Hash256(fmt.Sprintf("%v", hash)), nil
}

// HashHex16 generates a hexadecimal hash string for the given data.
// It accepts variadic arguments - if only one argument is provided, it hashes
// that value. If multiple arguments are provided, it hashes them as a tuple.
// The last argument can optionally be *Options.
//
// Returns:
//   - string: The computed hash string (never empty for valid inputs)
//   - error: Non-nil if hashing fails
//
// Example:
//
//	hash, err := HashHex16(myStruct)                    // Single value
//	hash, err := HashHex16(val1, val2, val3)            // Multiple values
//	hash, err := HashHex16(myStruct, opts)              // With options
//	hash, err := HashHex16(val1, val2, val3, opts)      // Multiple values with options
//
// The hash is deterministic: identical values always produce identical hashes.
func HashHex16(data ...any) (string, error) {
	hash, err := Hash(data...)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash), nil
}
