package conv

import (
	"reflect"
	"strings"

	"github.com/polarixa/replify/pkg/common"
	"github.com/polarixa/replify/pkg/strutil"
)

// ///////////////////////////
// Section:  Map conversion
// ///////////////////////////

// ToMap converts the given value to a map[string]any.
//
// Supported conversions:
//   - map[string]any:  returned as-is
//   - map[K]V: keys converted to string, values to any
//   - struct: field names as keys, field values as values
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - A map[string]any.
//   - An error if conversion fails.
func (c *Converter) ToMap(from any) (map[string]any, error) {
	if from == nil {
		if c.nilAsZero {
			return nil, nil
		}
		return nil, newConvError(from, "map[string]any")
	}

	// Fast path for map[string]any
	if v, ok := from.(map[string]any); ok {
		return v, nil
	}

	// Fast path for map[string]interface{}
	if v, ok := from.(map[string]interface{}); ok {
		return v, nil
	}

	value := reflect.ValueOf(from)
	kind := value.Kind()

	// Handle pointer
	if kind == reflect.Ptr {
		if value.IsNil() {
			return nil, nil
		}
		value = value.Elem()
		kind = value.Kind()
	}

	switch kind {
	case reflect.Map:
		return c.mapToStringMap(value)
	case reflect.Struct:
		return c.structToMap(value)
	default:
		return nil, newConvErrorf("cannot convert %T to map[string]any", from)
	}
}

// ///////////////////////////
// Section: ToStringMap conversion
// ///////////////////////////

// ToStringMap converts the given value to a map[string]string.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - A map[string]string.
//   - An error if conversion fails.
func (c *Converter) ToStringMap(from any) (map[string]string, error) {
	if from == nil {
		if c.nilAsZero {
			return nil, nil
		}
		return nil, newConvError(from, "map[string]string")
	}

	// Fast path for map[string]string
	if v, ok := from.(map[string]string); ok {
		return v, nil
	}

	// Convert to map[string]any first
	anyMap, err := c.ToMap(from)
	if err != nil {
		return nil, err
	}

	// Convert all values to strings
	result := make(map[string]string, len(anyMap))
	for k, v := range anyMap {
		str, err := c.String(v)
		if err != nil {
			return nil, newConvErrorf("cannot convert value for key %q to string: %v", k, err)
		}
		result[k] = str
	}

	return result, nil
}

// ToIntMap converts the given value to a map[string]int.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - A map[string]int.
//   - An error if conversion fails.
func (c *Converter) ToIntMap(from any) (map[string]int, error) {
	if from == nil {
		if c.nilAsZero {
			return nil, nil
		}
		return nil, newConvError(from, "map[string]int")
	}

	// Fast path for map[string]int
	if v, ok := from.(map[string]int); ok {
		return v, nil
	}

	// Convert to map[string]any first
	anyMap, err := c.ToMap(from)
	if err != nil {
		return nil, err
	}

	// Convert all values to int
	result := make(map[string]int, len(anyMap))
	for k, v := range anyMap {
		intVal, err := c.Int(v)
		if err != nil {
			return nil, newConvErrorf("cannot convert value for key %q to int: %v", k, err)
		}
		result[k] = intVal
	}

	return result, nil
}

// ToFloat64Map converts the given value to a map[string]float64.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - A map[string]float64.
//   - An error if conversion fails.
func (c *Converter) ToFloat64Map(from any) (map[string]float64, error) {
	if from == nil {
		if c.nilAsZero {
			return nil, nil
		}
		return nil, newConvError(from, "map[string]float64")
	}

	// Fast path for map[string]float64
	if v, ok := from.(map[string]float64); ok {
		return v, nil
	}

	// Convert to map[string]any first
	anyMap, err := c.ToMap(from)
	if err != nil {
		return nil, err
	}

	// Convert all values to float64
	result := make(map[string]float64, len(anyMap))
	for k, v := range anyMap {
		floatVal, err := c.Float64(v)
		if err != nil {
			return nil, newConvErrorf("cannot convert value for key %q to float64: %v", k, err)
		}
		result[k] = floatVal
	}

	return result, nil
}

// ToBoolMap converts the given value to a map[string]bool.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - A map[string]bool.
//   - An error if conversion fails.
func (c *Converter) ToBoolMap(from any) (map[string]bool, error) {
	if from == nil {
		if c.nilAsZero {
			return nil, nil
		}
		return nil, newConvError(from, "map[string]bool")
	}

	// Fast path for map[string]bool
	if v, ok := from.(map[string]bool); ok {
		return v, nil
	}

	// Convert to map[string]any first
	anyMap, err := c.ToMap(from)
	if err != nil {
		return nil, err
	}

	// Convert all values to bool
	result := make(map[string]bool, len(anyMap))
	for k, v := range anyMap {
		boolVal, err := c.Bool(v)
		if err != nil {
			return nil, newConvErrorf("cannot convert value for key %q to bool: %v", k, err)
		}
		result[k] = boolVal
	}

	return result, nil
}

// ///////////////////////////
// Section: Package-level Map functions
// ///////////////////////////

// MapTo converts the given value to map[string]any.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - A map[string]any.
//   - An error if conversion fails.
func MapTo(from any) (map[string]any, error) {
	return defaultConverter.ToMap(from)
}

// StringMap converts the given value to map[string]string.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - A map[string]string.
//   - An error if conversion fails.
func StringMap(from any) (map[string]string, error) {
	return defaultConverter.ToStringMap(from)
}

// IntMap converts the given value to map[string]int.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - A map[string]int.
//   - An error if conversion fails.
func IntMap(from any) (map[string]int, error) {
	return defaultConverter.ToIntMap(from)
}

// Float64Map converts the given value to map[string]float64.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - A map[string]float64.
//   - An error if conversion fails.
func Float64Map(from any) (map[string]float64, error) {
	return defaultConverter.ToFloat64Map(from)
}

// BoolMap converts the given value to map[string]bool.
//
// Parameters:
//   - from: The value to convert.
//
// Returns:
//   - A map[string]bool.
//   - An error if conversion fails.
func BoolMap(from any) (map[string]bool, error) {
	return defaultConverter.ToBoolMap(from)
}

// mapToStringMap converts a reflect.Value representing a map to a map with string keys.
//
// Parameters:
//   - `value`: A reflect.Value representing the map to be converted.
//
// Returns:
//   - A map with string keys and values of type `any`.
//   - An error if any key cannot be converted to a string.
func (c *Converter) mapToStringMap(value reflect.Value) (map[string]any, error) {
	result := make(map[string]any, value.Len())
	iter := value.MapRange()

	for iter.Next() {
		keyStr, err := c.String(iter.Key().Interface())
		if err != nil {
			return nil, newConvErrorf("cannot convert map key to string: %v", err)
		}
		result[keyStr] = iter.Value().Interface()
	}

	return result, nil
}

// getFieldName retrieves the field name from struct tags or defaults to the field's actual name.
//
// Parameters:
//   - `field`: The reflect.StructField representing the struct field.
//
// Returns:
//   - A string representing the field name, derived from the "json" or "yaml" tags if present, or the field's actual name otherwise.
//
// Example:
//
//	fieldName := converter.getFieldName(myStructField)
func (c *Converter) getFieldName(field reflect.StructField) string {
	// Check json tag first
	if jsonTag := field.Tag.Get("json"); strutil.IsNotEmpty(jsonTag) {
		parts := strings.Split(jsonTag, ",")
		if strutil.IsNotEmpty(parts[0]) {
			return parts[0]
		}
	}

	// Check yaml tag
	if yamlTag := field.Tag.Get("yaml"); strutil.IsNotEmpty(yamlTag) {
		parts := strings.Split(yamlTag, ",")
		if strutil.IsNotEmpty(parts[0]) {
			return parts[0]
		}
	}

	// Use field name
	return field.Name
}

// shouldOmitEmpty checks if a struct field should be omitted based on the "omitempty" tag and its value.
//
// Parameters:
//   - `field`: The reflect.StructField representing the struct field.
//   - `value`: The reflect.Value of the field to be checked.
//
// Returns:
//   - A boolean indicating whether the field should be omitted (true) or not (false).
//
// Example:
//
//	fieldShouldOmit := converter.shouldOmitEmpty(myStructField, myFieldValue)
func (c *Converter) shouldOmitEmpty(field reflect.StructField, value reflect.Value) bool {
	jsonTag := field.Tag.Get("json")
	if !strings.Contains(jsonTag, "omitempty") {
		return false
	}

	return common.IsEmptyValue(value)
}

// structToMap converts a reflect.Value representing a struct to a map with string keys.
//
// Parameters:
//   - `value`: A reflect.Value representing the struct to be converted.
//
// Returns:
//   - A map with string keys and values of type `any`.
//   - An error if any issues occur during conversion.
func (c *Converter) structToMap(value reflect.Value) (map[string]any, error) {
	valueType := value.Type()
	result := make(map[string]any, value.NumField())

	for i := 0; i < value.NumField(); i++ {
		field := valueType.Field(i)
		fieldValue := value.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name from json tag or use field name
		fieldName := c.getFieldName(field)

		// Skip fields with json: "-"
		if fieldName == "-" {
			continue
		}

		// Handle omitempty
		if c.shouldOmitEmpty(field, fieldValue) {
			continue
		}

		result[fieldName] = fieldValue.Interface()
	}

	return result, nil
}
