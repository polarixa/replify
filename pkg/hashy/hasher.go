package hashy

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/polarixa/replify/pkg/strutil"
)

// unwrapValue unwraps the value of a field.
//
// Parameters:
//   - value: The value of the field to unwrap.
//
// Returns:
//   - The unwrapped value.
//
// Example:
//
//	value := reflect.ValueOf(1)
//	unwrappedValue := h.unwrapValue(value)
//	fmt.Println(unwrappedValue) // 1
func (h *hasher) unwrapValue(value reflect.Value) reflect.Value {
	targetType := reflect.TypeFor[int]()

	for {
		if value.Kind() == reflect.Interface {
			value = value.Elem()
			continue
		}

		if value.Kind() == reflect.Ptr {
			if h.treatNilAsZero {
				targetType = value.Type().Elem()
			}
			value = reflect.Indirect(value)
			continue
		}

		break
	}

	// If it is nil, treat it like a zero value
	if !value.IsValid() {
		value = reflect.Zero(targetType)
	}

	return value
}

// normalizeValue normalizes the value of a field.
//
// Parameters:
//   - value: The value of the field to normalize.
//
// Returns:
//   - The normalized value.
//
// Example:
//
//	value := reflect.ValueOf(1)
//	normalizedValue := h.normalizeValue(value)
//	fmt.Println(normalizedValue) // 1
func (h *hasher) normalizeValue(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Int:
		return reflect.ValueOf(int64(value.Int()))
	case reflect.Uint:
		return reflect.ValueOf(uint64(value.Uint()))
	case reflect.Bool:
		var val int8
		if value.Bool() {
			val = 1
		}
		return reflect.ValueOf(val)
	}
	return value
}

// tryHashable tries to get the hash of a value.
//
// Parameters:
//   - value: The value to try to get the hash of.
//
// Returns:
//   - The hash of the value, a boolean indicating if the value is hashable, and an error if there is an error.
//
// Example:
//
//	value := reflect.ValueOf(1)
//	hash, ok, err := h.tryHashable(value)
//	fmt.Println(hash, ok, err) // 1 true nil
func (h *hasher) tryHashable(value reflect.Value) (uint64, bool, error) {
	// Try direct implementation
	if hashable, ok := value.Interface().(Hashable); ok {
		hash, err := hashable.Hash()
		return hash, true, err
	}

	// Try pointer implementation
	if value.CanAddr() {
		ptr := value.Addr()
		if hashable, ok := ptr.Interface().(Hashable); ok {
			hash, err := hashable.Hash()
			return hash, true, err
		}
	}

	return 0, false, nil // errors.New("pkg.hash: value is not hashable")
}

// hashZeroValue hashes a zero value.
//
// Returns:
//   - The hash of the zero value.
//
// Example:
//
//	hash, err := h.hashZeroValue()
//	fmt.Println(hash, err) // 0 nil
func (h *hasher) hashZeroValue() (uint64, error) {
	h.hash.Reset()
	if err := binary.Write(h.hash, binary.LittleEndian, int64(0)); err != nil {
		return 0, err
	}
	return h.hash.Sum64(), nil
}

// hashNumeric hashes a numeric value.
//
// Parameters:
//   - value: The value to hash.
//
// Returns:
//   - The hash of the value.
//
// Example:
//
//	value := reflect.ValueOf(1)
//	hash, err := h.hashNumeric(value)
//	fmt.Println(hash, err) // 1 nil
func (h *hasher) hashNumeric(value reflect.Value) (uint64, error) {
	h.hash.Reset()
	if err := binary.Write(h.hash, binary.LittleEndian, value.Interface()); err != nil {
		return 0, err
	}
	return h.hash.Sum64(), nil
}

// hashTime hashes a time value.
//
// Parameters:
//   - value: The value to hash.
//
// Returns:
//   - The hash of the value.
//
// Example:
//
//	value := reflect.ValueOf(time.Now())
//	hash, err := h.hashTime(value)
//	fmt.Println(hash, err) // time.Now().Unix() nil
func (h *hasher) hashTime(value reflect.Value) (uint64, error) {
	h.hash.Reset()
	iface := value.Interface()
	timeVal, ok := iface.(time.Time)
	if !ok {
		return 0, fmt.Errorf("hashy: expected time.Time, got %T", iface)
	}
	data, err := timeVal.MarshalBinary()
	if err != nil {
		return 0, err
	}
	if _, err := h.hash.Write(data); err != nil {
		return 0, err
	}
	return h.hash.Sum64(), nil
}

// hashString hashes a string value.
//
// Parameters:
//   - value: The value to hash.
//
// Returns:
//   - The hash of the value.
//
// Example:
//
//	value := reflect.ValueOf("hello")
//	hash, err := h.hashString(value)
//	fmt.Println(hash, err) // "hello" nil
func (h *hasher) hashString(value reflect.Value) (uint64, error) {
	h.hash.Reset()
	if _, err := io.WriteString(h.hash, value.String()); err != nil {
		return 0, err
	}
	return h.hash.Sum64(), nil
}

// hashUpdateOrdered hashes two values in order.
//
// Parameters:
//   - a: The first value to hash.
//   - b: The second value to hash.
//
// Returns:
//   - The hash of the two values.
//
// Example:
//
//	a := 1
//	b := 2
//	hash, err := h.hashUpdateOrdered(a, b)
//	fmt.Println(hash, err) // 3 nil
func (h *hasher) hashUpdateOrdered(a, b uint64) (uint64, error) {
	h.hash.Reset()
	if err := binary.Write(h.hash, binary.LittleEndian, a); err != nil {
		return 0, fmt.Errorf("pkg.hash: ordered write failed: %w", err)
	}
	if err := binary.Write(h.hash, binary.LittleEndian, b); err != nil {
		return 0, fmt.Errorf("pkg.hash: ordered write failed: %w", err)
	}
	return h.hash.Sum64(), nil
}

// hashFinishUnordered "hardens" the XOR result to prevent cancellation issues.
// After mixing a group of unique hashes with hashUpdateUnordered, it's necessary
// to call hashFinishUnordered. This prevents issues where XOR operations can
// cancel out when the same hash appears in different contexts.
//
// Parameters:
//   - a: The value to hash.
//
// Returns:
//   - The hash of the value.
//
// Example:
//
//	a := 1
//	hash, err := h.hashFinishUnordered(a)
//	fmt.Println(hash, err) // 1 nil
func (h *hasher) hashFinishUnordered(a uint64) (uint64, error) {
	h.hash.Reset()
	if err := binary.Write(h.hash, binary.LittleEndian, a); err != nil {
		return 0, fmt.Errorf("pkg.hash: unordered finish write failed: %w", err)
	}
	return h.hash.Sum64(), nil
}

// hashArray hashes an array value.
//
// Parameters:
//   - value: The value to hash.
//   - options: The options to use for the hash.
//
// Returns:
//   - The hash of the value.
//
// Example:
//
//	value := reflect.ValueOf([]int{1, 2, 3})
//	hash, err := h.hashArray(value, nil)
//	fmt.Println(hash, err) // 1 2 3 nil
func (h *hasher) hashArray(value reflect.Value, options *visitOptions) (uint64, error) {
	var accumulated uint64
	length := value.Len()

	for i := 0; i < length; i++ {
		current, err := h.hashValue(value.Index(i), nil)
		if err != nil {
			return 0, err
		}
		accumulated, err = h.hashUpdateOrdered(accumulated, current)
		if err != nil {
			return 0, err
		}
	}

	return accumulated, nil
}

// hashSlice hashes a slice value.
//
// Parameters:
//   - value: The value to hash.
//   - options: The options to use for the hash.
//
// Returns:
//   - The hash of the value.
//
// Example:
//
//	value := reflect.ValueOf([]int{1, 2, 3})
//	hash, err := h.hashSlice(value, nil)
//	fmt.Println(hash, err) // 1 2 3 nil
func (h *hasher) hashSlice(value reflect.Value, options *visitOptions) (uint64, error) {
	// Determine if this should be treated as a set (order-independent)
	isSet := h.slicesAsSets
	if options != nil && (options.flags&visitFlagSet) != 0 {
		isSet = true
	}

	var accumulated uint64
	length := value.Len()

	for i := 0; i < length; i++ {
		current, err := h.hashValue(value.Index(i), nil)
		if err != nil {
			return 0, err
		}

		if isSet {
			accumulated = hashUpdateUnordered(accumulated, current)
		} else {
			accumulated, err = h.hashUpdateOrdered(accumulated, current)
			if err != nil {
				return 0, err
			}
		}
	}

	if isSet {
		var err error
		accumulated, err = h.hashFinishUnordered(accumulated)
		if err != nil {
			return 0, err
		}
	}

	return accumulated, nil
}

// hashMap hashes a map value.
//
// Parameters:
//   - value: The value to hash.
//   - options: The options to use for the hash.
//
// Returns:
//   - The hash of the value.
//
// Example:
//
//	value := reflect.ValueOf(map[string]int{"a": 1, "b": 2})
//	hash, err := h.hashMap(value, nil)
//	fmt.Println(hash, err) // 1 2 nil
func (h *hasher) hashMap(value reflect.Value, options *visitOptions) (uint64, error) {
	var selector MapSelector
	if options != nil && options.structValue != nil {
		if impl, ok := options.structValue.(MapSelector); ok {
			selector = impl
		}
	}

	var accumulated uint64
	for _, key := range value.MapKeys() {
		mapValue := value.MapIndex(key)

		// Check if this map entry should be included
		if selector != nil {
			include, err := selector.SelectMapEntry()(options.fieldName, key.Interface(), mapValue.Interface())
			if err != nil {
				return 0, err
			}
			if !include {
				continue
			}
		}

		keyHash, err := h.hashValue(key, nil)
		if err != nil {
			return 0, err
		}

		valueHash, err := h.hashValue(mapValue, nil)
		if err != nil {
			return 0, err
		}

		fieldHash, err := h.hashUpdateOrdered(keyHash, valueHash)
		if err != nil {
			return 0, err
		}
		accumulated = hashUpdateUnordered(accumulated, fieldHash)
	}

	accumulated, err := h.hashFinishUnordered(accumulated)
	if err != nil {
		return 0, err
	}
	return accumulated, nil
}

// hashStruct hashes a struct value.
//
// Parameters:
//   - value: The value to hash.
//   - options: The options to use for the hash.
//
// Returns:
//   - The hash of the value.
//
// Example:
//
//	value := reflect.ValueOf(struct{ A int; B string }{"a", 1})
//	hash, err := h.hashStruct(value, nil)
//	fmt.Println(hash, err) // 1 2 nil
func (h *hasher) hashStruct(value reflect.Value, options *visitOptions) (uint64, error) {
	structType := value.Type()
	parent := value.Interface()

	// Check for FieldSelector interface
	var selector FieldSelector
	if impl, ok := parent.(FieldSelector); ok {
		selector = impl
	}

	// Check pointer for FieldIncluder interface
	if value.CanAddr() {
		ptr := value.Addr()
		if impl, ok := ptr.Interface().(FieldSelector); ok {
			selector = impl
		}
	}

	// Start with struct type name
	accumulated, err := h.hashValue(reflect.ValueOf(structType.Name()), nil)
	if err != nil {
		return 0, err
	}

	numFields := value.NumField()
	for i := 0; i < numFields; i++ {
		fieldValue := value.Field(i)
		fieldType := structType.Field(i)

		// Skip unexported fields (except for "_")
		if strutil.IsNotEmpty(fieldType.PkgPath) && fieldType.Name != "_" {
			continue
		}

		// Check tag for special directives
		tag := fieldType.Tag.Get(h.tagName)
		if tag == "ignore" || tag == "-" {
			continue
		}

		// Skip zero values if configured
		if h.ignoreZeroValue && fieldValue.IsZero() {
			continue
		}

		// Handle string tag
		if tag == "string" || h.useStringer {
			if stringer, ok := fieldValue.Interface().(fmt.Stringer); ok {
				fieldValue = reflect.ValueOf(stringer.String())
			} else if tag == "string" {
				return 0, &ErrNotStringer{Field: fieldType.Name}
			}
		}

		// Check FieldSelector interface
		if selector != nil {
			include, err := selector.SelectField()(fieldType.Name, fieldValue.Interface())
			if err != nil {
				return 0, err
			}
			if !include {
				continue
			}
		}

		// Determine visit flags, if the tag is "set", set the visitFlagSet flag
		// This is used to indicate that the slice should be treated as a set (order-independent)
		var flags visitFlag
		if tag == "set" {
			flags |= visitFlagSet
		}

		// Hash field name
		nameHash, err := h.hashValue(reflect.ValueOf(fieldType.Name), nil)
		if err != nil {
			return 0, err
		}

		// Hash field value
		valueHash, err := h.hashValue(fieldValue, &visitOptions{
			flags:       flags,
			structValue: parent,
			fieldName:   fieldType.Name,
		})
		if err != nil {
			return 0, err
		}

		fieldHash, err := h.hashUpdateOrdered(nameHash, valueHash)
		if err != nil {
			return 0, err
		}
		accumulated = hashUpdateUnordered(accumulated, fieldHash)
	}

	accumulated, err = h.hashFinishUnordered(accumulated)
	if err != nil {
		return 0, err
	}
	return accumulated, nil
}

// hashValue hashes a value.
//
// Parameters:
//   - value: The value to hash.
//   - options: The options to use for the hash.
//
// Returns:
//   - The hash of the value.
//
// Example:
//
//	value := reflect.ValueOf(1)
//	hash, err := h.hashValue(value, nil)
//	fmt.Println(hash, err) // 1 nil
func (h *hasher) hashValue(value reflect.Value, options *visitOptions) (uint64, error) {
	// Unwrap interfaces and pointers
	value = h.unwrapValue(value)

	if !value.IsValid() {
		return h.hashZeroValue()
	}

	// Check for custom Hashable implementation
	if hashVal, ok, err := h.tryHashable(value); ok || err != nil {
		return hashVal, err
	}

	// Normalize numeric types for binary writing
	value = h.normalizeValue(value)

	// Fast path for numeric types
	kind := value.Kind()
	if kind >= reflect.Int && kind <= reflect.Complex128 {
		return h.hashNumeric(value)
	}

	// Special handling for time.Time
	if value.Type() == timeType {
		return h.hashTime(value)
	}

	// Dispatch by kind
	switch kind {
	case reflect.Array:
		return h.hashArray(value, options)
	case reflect.Map:
		return h.hashMap(value, options)
	case reflect.Struct:
		return h.hashStruct(value, options)
	case reflect.Slice:
		return h.hashSlice(value, options)
	case reflect.String:
		return h.hashString(value)
	default:
		return 0, fmt.Errorf("pkg.hash: unsupported kind: %s", kind)
	}
}

// hashUpdateUnordered hashes two values in unordered.
//
// Parameters:
//   - a: The first value to hash.
//   - b: The second value to hash.
//
// Returns:
//   - The hash of the two values.
//
// Example:
//
//	a := 1
//	b := 2
//	hash, err := h.hashUpdateUnordered(a, b)
//	fmt.Println(hash, err) // 3 nil
func hashUpdateUnordered(a, b uint64) uint64 {
	return a ^ b
}
