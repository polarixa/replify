package hashy

import (
	"fmt"
	"hash"

	"github.com/polarixa/replify/pkg/strutil"
)

// WithHasher sets the hash function to use.
//
// Parameters:
//   - h: The hash function to use.
//
// Returns:
//   - A pointer to the `OptionsBuilder` struct.
//
// Deprecated: prefer WithHasherFunc for concurrent-safe use. A single
// hash.Hash64 instance is stateful; passing the same *hashOptions to
// Hash/HashValue from multiple goroutines concurrently causes a data race.
//
// Example:
//
//	builder := NewOptions().WithHasher(fnv.New64a())
//	opts := builder.Build()
func (b *OptionsBuilder) WithHasher(h hash.Hash64) *OptionsBuilder {
	b.opts.Hasher = h
	return b
}

// WithHasherFunc sets a factory function that creates a fresh hash.Hash64
// for every hashing operation. Using a factory function is safe for
// concurrent use by multiple goroutines.
//
// Parameters:
//   - fn: A function that returns a new hash.Hash64 instance each time it is called.
//
// Returns:
//   - A pointer to the `OptionsBuilder` struct.
//
// Example:
//
//	builder := NewOptions().WithHasherFunc(fnv.New64a)
//	opts := builder.Build()
func (b *OptionsBuilder) WithHasherFunc(fn func() hash.Hash64) *OptionsBuilder {
	b.opts.HasherFunc = fn
	return b
}

// WithTagName sets the struct tag to look at when hashing the structure.
//
// Parameters:
//   - name: The name of the struct tag to look at.
//
// Returns:
//   - A pointer to the `OptionsBuilder` struct.
//
// Example:
//
//	builder := NewOptions().WithTagName("json")
//	opts := builder.Build()
func (b *OptionsBuilder) WithTagName(name string) *OptionsBuilder {
	b.opts.TagName = name
	return b
}

// WithZeroNil sets whether nil pointer should be treated equal to a zero value of pointed type.
//
// Parameters:
//   - zeroNil: A boolean indicating whether nil pointer should be treated equal to a zero value of pointed type.
//
// Returns:
//   - A pointer to the `OptionsBuilder` struct.
//
// Example:
//
//	builder := NewOptions().WithZeroNil(true)
//	opts := builder.Build()
func (b *OptionsBuilder) WithZeroNil(zeroNil bool) *OptionsBuilder {
	b.opts.ZeroNil = zeroNil
	return b
}

// WithIgnoreZeroValue sets whether zero value fields should be ignored for hash calculation.
//
// Parameters:
//   - ignore: A boolean indicating whether zero value fields should be ignored for hash calculation.
//
// Returns:
//   - A pointer to the `OptionsBuilder` struct.
//
// Example:
//
//	builder := NewOptions().WithIgnoreZeroValue(true)
//	opts := builder.Build()
func (b *OptionsBuilder) WithIgnoreZeroValue(ignore bool) *OptionsBuilder {
	b.opts.IgnoreZeroValue = ignore
	return b
}

// WithSlicesAsSets sets whether slices should be treated as sets.
//
// Parameters:
//   - asSets: A boolean indicating whether slices should be treated as sets.
//
// Returns:
//   - A pointer to the `OptionsBuilder` struct.
//
// Example:
//
//	builder := NewOptions().WithSlicesAsSets(true)
//	opts := builder.Build()
func (b *OptionsBuilder) WithSlicesAsSets(asSets bool) *OptionsBuilder {
	b.opts.SlicesAsSets = asSets
	return b
}

// WithUseStringer sets whether fmt.Stringer should be used always.
//
// Parameters:
//   - useStringer: A boolean indicating whether fmt.Stringer should be used always.
//
// Returns:
//   - A pointer to the `OptionsBuilder` struct.
//
// Example:
//
//	builder := NewOptions().WithUseStringer(true)
//	opts := builder.Build()
func (b *OptionsBuilder) WithUseStringer(useStringer bool) *OptionsBuilder {
	b.opts.UseStringer = useStringer
	return b
}

// Build builds the options.
//
// Returns:
//   - A pointer to the `Options` struct.
//
// Example:
//
//	builder := NewOptions().Build()
//	opts := builder.Build()
func (b *OptionsBuilder) Build() *hashOptions {
	return &b.opts
}

// validate checks if options are valid.
//
// Parameters:
//   - o: A pointer to the `Options` struct to validate.
//
// Returns:
//   - An error if the options are invalid, otherwise nil.
func (o *hashOptions) validate() error {
	if o.HasherFunc == nil && o.Hasher == nil {
		return fmt.Errorf("pkg.hash.options: hasher cannot be nil")
	}
	if strutil.IsEmpty(o.TagName) {
		return fmt.Errorf("pkg.hash.options: tag name cannot be empty")
	}
	return nil
}
