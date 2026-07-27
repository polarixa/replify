package replify

import (
	"fmt"
	"time"

	"github.com/polarixa/replify/pkg/coll"
	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/fj"
	"github.com/polarixa/replify/pkg/randn"
	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strchain"
	"github.com/polarixa/replify/pkg/strutil"
)

// WithApiVersion sets the API version for the [meta] instance.
//
// This function updates the `apiVersion` field of the [meta] instance with the specified value
// and returns the updated [meta] instance for method chaining.
//
// Parameters:
//   - `v`: A string representing the API version to set.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) WithApiVersion(v string) *meta {
	m.apiVersion = v
	return m
}

// WithApiVersionf sets the API version for the [meta] instance using a formatted string.
//
// This function constructs a formatted string for the API version using the provided `format` string
// and arguments (`args`). It then assigns the formatted value to the `apiVersion` field of the [meta] instance.
// The method supports method chaining by returning a pointer to the modified [meta] instance.
//
// Parameters:
//   - format: A format string to construct the API version.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) WithApiVersionf(format string, args ...any) *meta {
	return m.WithApiVersion(fmt.Sprintf(format, args...))
}

// WithRequestID sets the request ID for the [meta] instance.
//
// This function updates the `requestID` field of the [meta] instance with the specified value
// and returns the updated [meta] instance for method chaining.
//
// Parameters:
//   - `v`: A string representing the request ID to set.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) WithRequestID(v string) *meta {
	m.requestID = v
	return m
}

// WithRequestIDf sets the request ID for the [meta] instance using a formatted string.
//
// This function constructs a formatted string for the request ID using the provided `format` string
// and arguments (`args`). It then assigns the formatted value to the `requestID` field of the [meta] instance.
// The method supports method chaining by returning a pointer to the modified [meta] instance.
//
// Parameters:
//   - format: A format string to construct the request ID.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) WithRequestIDf(format string, args ...any) *meta {
	return m.WithRequestID(fmt.Sprintf(format, args...))
}

// WithLocale sets the locale for the [meta] instance.
//
// This function updates the `locale` field of the [meta] instance with the specified value
// and returns the updated [meta] instance for method chaining.
//
// Parameters:
//   - `v`: A string representing the locale to set.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) WithLocale(v string) *meta {
	m.locale = v
	return m
}

// WithLocalef sets the locale for the [meta] instance using a formatted string.
//
// This function constructs a formatted string for the locale using the provided `format` string
// and arguments (`args`). It then assigns the formatted value to the `locale` field of the [meta] instance.
//
// The method supports method chaining by returning a pointer to the modified [meta] instance.
//
// Parameters:
//   - format: A format string to construct the locale.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) WithLocalef(format string, args ...any) *meta {
	return m.WithLocale(fmt.Sprintf(format, args...))
}

// WithLocaleValue sets the locale for the [meta] instance using a `Locale` type.
//
// This function updates the `locale` field of the [meta] instance with the string representation
// of the provided `Locale` value. It returns the updated [meta] instance for method chaining.
//
// Parameters:
//   - `locale`: A `Locale` type representing the locale to set.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) WithLocaleValue(locale Locale) *meta {
	return m.WithLocale(string(locale))
}

// WithRequestedTime sets the requested time for the [meta] instance.
//
// This function updates the `requestedTime` field of the [meta] instance with the specified value
// and returns the updated [meta] instance for method chaining.
//
// Parameters:
//   - `v`: A `time.Time` object representing the requested time to set.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) WithRequestedTime(v time.Time) *meta {
	m.requestedTime = v
	return m
}

// WithCustomFields sets the custom fields for the [meta] instance.
//
// This function updates the `customFields` map of the [meta] instance with the provided values
// and returns the updated [meta] instance for method chaining.
//
// Parameters:
//   - `values`: A map of string keys to interface{} values representing custom fields.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) WithCustomFields(values map[string]any) *meta {
	m.customFields = values
	return m
}

// WithCustomFieldKV sets a specific custom field key-value pair for the [meta] instance.
//
// This function adds or updates a custom field in the `customFields` map of the [meta] instance.
// If the `customFields` map is empty, it is initialized first.
//
// Parameters:
//   - `key`: A string representing the custom field key.
//   - `value`: An interface{} representing the value to associate with the custom field key.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) WithCustomFieldKV(key string, value any) *meta {
	if !m.IsCustomPresent() {
		m.customFields = make(map[string]interface{})
	}
	m.customFields[key] = value
	return m
}

// WithCustomFieldKVf sets a specific custom field key-value pair for the [meta] instance
// using a formatted value.
//
// This function creates a formatted string value using the provided `format` string and
// `args`. It then calls `WithCustomFieldKV` to add or update the custom field with the
// specified key and the formatted value. The modified [meta] instance is returned for
// method chaining.
//
// Parameters:
//   - key: A string representing the key for the custom field.
//   - format: A format string to construct the value.
//   - args: A variadic list of arguments to be interpolated into the format string.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) WithCustomFieldKVf(key string, format string, args ...any) *meta {
	return m.WithCustomFieldKV(key, fmt.Sprintf(format, args...))
}

// WithDeltaValue sets the delta value for the [meta] instance.
// Represents the magnitude of change introduced by payload normalization or transformation.
//
// This function updates the `deltaValue` field of the [meta] instance with the specified float64 value
// and returns the updated [meta] instance for method chaining.
//
// Parameters:
//   - `value`: A float64 representing the delta value to set.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) WithDeltaValue(value float64) *meta {
	m.deltaValue = value
	return m
}

// WithDeltaCnt sets the delta count for the [meta] instance.
// Represents the number of changes introduced by payload normalization or transformation.
//
// This function updates the `deltaCnt` field of the [meta] instance with the specified int value
// and returns the updated [meta] instance for method chaining.
//
// Parameters:
//   - `value`: An int representing the delta count to set.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) WithDeltaCnt(value int) *meta {
	m.deltaCnt = value
	return m
}

// Available checks whether the [meta] instance is available (non-nil).
//
// This function ensures that the [meta] instance is valid and not nil
// to safely access its fields and methods.
//
// Returns:
//   - `true` if the [meta] instance is not nil.
//   - `false` if the [meta] instance is nil.
func (m *meta) Available() bool {
	return m != nil
}

// IsApiVersionPresent checks whether the API version is present in the [meta] instance.
//
// This function verifies that the [meta] instance is available and that
// the `apiVersion` field contains a non-empty value.
//
// Returns:
//   - `true` if `apiVersion` is non-empty.
//   - `false` if [meta] is unavailable or `apiVersion` is empty.
func (m *meta) IsApiVersionPresent() bool {
	return m.Available() && strutil.IsNotEmpty(m.apiVersion)
}

// IsRequestIDPresent checks whether the request ID is present in the [meta] instance.
//
// This function verifies that the [meta] instance is available and that
// the `requestID` field contains a non-empty value.
//
// Returns:
//   - `true` if `requestID` is non-empty.
//   - `false` if [meta] is unavailable or `requestID` is empty.
func (m *meta) IsRequestIDPresent() bool {
	return m.Available() && strutil.IsNotEmpty(m.requestID)
}

// IsLocalePresent checks whether the locale information is present in the [meta] instance.
//
// This function verifies that the [meta] instance is available and that
// the `locale` field contains a non-empty value.
//
// Returns:
//   - `true` if `locale` is non-empty.
//   - `false` if [meta] is unavailable or `locale` is empty.
func (m *meta) IsLocalePresent() bool {
	return m.Available() && strutil.IsNotEmpty(m.locale)
}

// IsRequestedTimePresent checks whether the requested time is present in the [meta] instance.
//
// This function verifies that the [meta] instance is available and that
// the `requestedTime` field is not the zero value of `time.Time`.
//
// Returns:
//   - `true` if `requestedTime` is not the zero value of `time.Time`.
//   - `false` if [meta] is unavailable or `requestedTime` is uninitialized.
func (m *meta) IsRequestedTimePresent() bool {
	return m.Available() && m.requestedTime != time.Time{} && !m.requestedTime.IsZero()
}

// IsCustomPresent checks whether custom fields are present in the [meta] instance.
//
// This function verifies that the [meta] instance is available and that
// the `customFields` field is non-nil and contains at least one entry.
//
// Returns:
//   - `true` if `customFields` is non-nil and has a non-zero length.
//   - `false` if [meta] is unavailable, `customFields` is nil, or it is empty.
func (m *meta) IsCustomPresent() bool {
	return m.Available() && m.customFields != nil && len(m.customFields) > 0
}

// IsDeltaValuePresent checks whether the delta value is present in the [meta] instance.
//
// This function verifies that the [meta] instance is available and that
// the `deltaValue` field is greater than zero.
//
// Returns:
//   - `true` if `deltaValue` is greater than zero.
//   - `false` if [meta] is unavailable or `deltaValue` is zero or negative.
func (m *meta) IsDeltaValuePresent() bool {
	return m.Available() && m.deltaValue > 0
}

// IsDeltaCntPresent checks whether the delta count is present in the [meta] instance.
//
// This function verifies that the [meta] instance is available and that
// the `deltaCnt` field is greater than zero.
//
// Returns:
//   - `true` if `deltaCnt` is greater than zero.
//   - `false` if [meta] is unavailable or `deltaCnt` is zero or negative.
func (m *meta) IsDeltaCntPresent() bool {
	return m.Available() && m.deltaCnt > 0
}

// IsCustomKeyPresent checks whether a specific key is present in the custom fields of the [meta] instance.
//
// This function first verifies that the `customFields` field is available and contains data using
// `IsCustomPresent`. If so, it checks if the specified key exists within the `customFields` map.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//
// Returns:
//   - `true` if the `customFields` map is available and contains the specified key.
//   - `false` if `customFields` is nil, empty, or does not contain the specified key.
func (m *meta) IsCustomKeyPresent(key string) bool {
	return m.IsCustomPresent() && coll.ContainsKeyComp(m.customFields, key)
}

// JSONCustomFields returns the custom fields of the [meta] instance as a JSON string.
//
// This function first checks if the [meta] instance is available. If not, it returns an empty string.
// Otherwise, it uses `jsonpass` to convert the `customFields` map into a JSON formatted string.
//
// Returns:
//   - A JSON formatted string representation of the `customFields` map if [meta] is available.
//   - An empty string if [meta] is unavailable.
func (m *meta) JSONCustomFields() string {
	if !m.IsCustomPresent() {
		return ""
	}
	return jsonpass(m.customFields)
}

// OnCustom retrieves the value associated with a specific key in the custom fields of the [meta] instance.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key.
// If the [meta] instance is unavailable or the key is not present, it returns `nil`.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map if the key is present.
//   - `nil` if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) OnCustom(key string) any {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return nil
	}
	return m.customFields[key]
}

// CustomBool retrieves the value associated with a specific key in the custom fields of the [meta] instance as a boolean.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a boolean using `conv.BoolOrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: A boolean representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a boolean if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomBool(key string, defaultValue bool) bool {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.BoolOrDefault(m.customFields[key], defaultValue)
}

// JSONCustomBool retrieves the value associated with a specific key in the custom fields of the [meta] instance as a boolean.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a boolean using `conv.BoolOrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: A boolean representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a boolean if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomBool(path string, defaultValue bool) bool {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Bool()
	}
	return defaultValue
}

// CustomDuration retrieves the value associated with a specific key in the custom fields of the [meta] instance as a time.Duration.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a time.Duration using `conv.DurationOrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: A time.Duration representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a time.Duration if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomDuration(key string, defaultValue time.Duration) time.Duration {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.DurationOrDefault(m.customFields[key], defaultValue)
}

// JSONCustomDuration retrieves the value associated with a specific key in the custom fields of the [meta] instance as a time.Duration.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a time.Duration using `conv.DurationOrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: A time.Duration representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a time.Duration if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomDuration(path string, defaultValue time.Duration) time.Duration {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Duration()
	}
	return defaultValue
}

// CustomString retrieves the value associated with a specific key in the custom fields of the [meta] instance as a string.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a string using `conv.StringOrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: A string representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a string if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomString(key string, defaultValue string) string {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.StringOrDefault(m.customFields[key], defaultValue)
}

// JSONCustomString retrieves the value associated with a specific key in the custom fields of the [meta] instance as a string.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a string using `conv.StringOrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: A string representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a string if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomString(path string, defaultValue string) string {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.String()
	}
	return defaultValue
}

// CustomTime retrieves the value associated with a specific key in the custom fields of the [meta] instance as a time.Time.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a time.Time using `conv.TimeOrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: A time.Time representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a time.Time if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomTime(key string, defaultValue time.Time) time.Time {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.TimeOrDefault(m.customFields[key], defaultValue)
}

// JSONCustomTime retrieves the value associated with a specific key in the custom fields of the [meta] instance as a time.Time.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a time.Time using `conv.TimeOrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: A time.Time representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a time.Time if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomTime(path string, defaultValue time.Time) time.Time {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Time()
	}
	return defaultValue
}

// CustomInt retrieves the value associated with a specific key in the custom fields of the [meta] instance as an integer.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to an integer using `conv.IntOrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns 0.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: An integer representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as an integer if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomInt(key string, defaultValue int) int {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.IntOrDefault(m.customFields[key], defaultValue)
}

// JSONCustomInt retrieves the value associated with a specific key in the custom fields of the [meta] instance as an integer.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to an integer using `conv.IntOrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: An integer representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as an integer if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomInt(path string, defaultValue int) int {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Int()
	}
	return defaultValue
}

// CustomInt8 retrieves the value associated with a specific key in the custom fields of the [meta] instance as an int8.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to an int8 using `conv.Int8OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: An int8 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as an int8 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomInt8(key string, defaultValue int8) int8 {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.Int8OrDefault(m.customFields[key], defaultValue)
}

// JSONCustomInt8 retrieves the value associated with a specific key in the custom fields of the [meta] instance as an int8.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to an int8 using `conv.Int8OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: An int8 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as an int8 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomInt8(path string, defaultValue int8) int8 {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Int8()
	}
	return defaultValue
}

// CustomInt16 retrieves the value associated with a specific key in the custom fields of the [meta] instance as an int16.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to an int16 using `conv.Int16OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: An int16 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as an int16 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomInt16(key string, defaultValue int16) int16 {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.Int16OrDefault(m.customFields[key], defaultValue)
}

// JSONCustomInt16 retrieves the value associated with a specific key in the custom fields of the [meta] instance as an int16.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to an int16 using `conv.Int16OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: An int16 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as an int16 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomInt16(path string, defaultValue int16) int16 {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Int16()
	}
	return defaultValue
}

// CustomInt32 retrieves the value associated with a specific key in the custom fields of the [meta] instance as an int32.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to an int32 using `conv.Int32OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: An int32 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as an int32 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomInt32(key string, defaultValue int32) int32 {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.Int32OrDefault(m.customFields[key], defaultValue)
}

// JSONCustomInt32 retrieves the value associated with a specific key in the custom fields of the [meta] instance as an int32.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to an int32 using `conv.Int32OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: An int32 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as an int32 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomInt32(path string, defaultValue int32) int32 {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Int32()
	}
	return defaultValue
}

// CustomInt64 retrieves the value associated with a specific key in the custom fields of the [meta] instance as an int64.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to an int64 using `conv.Int64OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: An int64 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as an int64 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomInt64(key string, defaultValue int64) int64 {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.Int64OrDefault(m.customFields[key], defaultValue)
}

// JSONCustomInt64 retrieves the value associated with a specific key in the custom fields of the [meta] instance as an int64.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to an int64 using `conv.Int64OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: An int64 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as an int64 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomInt64(path string, defaultValue int64) int64 {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Int64()
	}
	return defaultValue
}

// CustomUint retrieves the value associated with a specific key in the custom fields of the [meta] instance as a uint.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a uint using `conv.UintOrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: A uint representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a uint if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomUint(key string, defaultValue uint) uint {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.UintOrDefault(m.customFields[key], defaultValue)
}

// JSONCustomUint retrieves the value associated with a specific key in the custom fields of the [meta] instance as a uint.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a uint using `conv.UintOrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: A uint representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a uint if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomUint(path string, defaultValue uint) uint {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Uint()
	}
	return defaultValue
}

// CustomUint8 retrieves the value associated with a specific key in the custom fields of the [meta] instance as a uint8.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a uint8 using `conv.Uint8OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: A uint8 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a uint8 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomUint8(key string, defaultValue uint8) uint8 {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.Uint8OrDefault(m.customFields[key], defaultValue)
}

// JSONCustomUint8 retrieves the value associated with a specific key in the custom fields of the [meta] instance as a uint8.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a uint8 using `conv.Uint8OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: A uint8 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a uint8 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomUint8(path string, defaultValue uint8) uint8 {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Uint8()
	}
	return defaultValue
}

// CustomUint16 retrieves the value associated with a specific key in the custom fields of the [meta] instance as a uint16.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a uint16 using `conv.Uint16OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: A uint16 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a uint16 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomUint16(key string, defaultValue uint16) uint16 {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.Uint16OrDefault(m.customFields[key], defaultValue)
}

// JSONCustomUint16 retrieves the value associated with a specific key in the custom fields of the [meta] instance as a uint16.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a uint16 using `conv.Uint16OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: A uint16 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a uint16 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomUint16(path string, defaultValue uint16) uint16 {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Uint16()
	}
	return defaultValue
}

// CustomUint32 retrieves the value associated with a specific key in the custom fields of the [meta] instance as a uint32.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a uint32 using `conv.Uint32OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: A uint32 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a uint32 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomUint32(key string, defaultValue uint32) uint32 {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.Uint32OrDefault(m.customFields[key], defaultValue)
}

// JSONCustomUint32 retrieves the value associated with a specific key in the custom fields of the [meta] instance as a uint32.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a uint32 using `conv.Uint32OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: A uint32 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a uint32 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomUint32(path string, defaultValue uint32) uint32 {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Uint32()
	}
	return defaultValue
}

// CustomUint64 retrieves the value associated with a specific key in the custom fields of the [meta] instance as a uint64.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a uint64 using `conv.Uint64OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: A uint64 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a uint64 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomUint64(key string, defaultValue uint64) uint64 {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.Uint64OrDefault(m.customFields[key], defaultValue)
}

// JSONCustomUint64 retrieves the value associated with a specific key in the custom fields of the [meta] instance as a uint64.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a uint64 using `conv.Uint64OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: A uint64 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a uint64 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomUint64(path string, defaultValue uint64) uint64 {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Uint64()
	}
	return defaultValue
}

// CustomFloat32 retrieves the value associated with a specific key in the custom fields of the [meta] instance as a float32.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a float32 using `conv.Float32OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: A float32 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a float32 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomFloat32(key string, defaultValue float32) float32 {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.Float32OrDefault(m.customFields[key], defaultValue)
}

// JSONCustomFloat32 retrieves the value associated with a specific key in the custom fields of the [meta] instance as a float32.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a float32 using `conv.Float32OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: A float32 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a float32 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomFloat32(path string, defaultValue float32) float32 {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Float32()
	}
	return defaultValue
}

// CustomFloat64 retrieves the value associated with a specific key in the custom fields of the [meta] instance as a float64.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a float64 using `conv.Float64OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `key`: A string representing the key to search for in the `customFields` map.
//   - `defaultValue`: A float64 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a float64 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) CustomFloat64(key string, defaultValue float64) float64 {
	if !m.Available() || !m.IsCustomKeyPresent(key) {
		return defaultValue
	}
	return conv.Float64OrDefault(m.customFields[key], defaultValue)
}

// JSONCustomFloat64 retrieves the value associated with a specific key in the custom fields of the [meta] instance as a float64.
//
// This function checks whether the [meta] instance is available and whether the specified key exists
// in the `customFields` map. If both conditions are met, it returns the corresponding value for the key
// converted to a float64 using `conv.Float64OrDefault`. If the [meta] instance is unavailable or the key is not present,
// it returns the default value.
//
// Parameters:
//   - `path`: A string representing the JSON path to search for in the `customFields` map.
//   - `defaultValue`: A float64 representing the default value to return if the key is not present.
//
// Returns:
//   - The value associated with the specified key in the `customFields` map as a float64 if the key is present.
//   - The default value if the [meta] instance is unavailable or the key does not exist in the `customFields` map.
func (m *meta) JSONCustomFloat64(path string, defaultValue float64) float64 {
	if !m.IsCustomPresent() {
		return defaultValue
	}
	ctx := fj.Get(m.JSONCustomFields(), path)
	if ctx.Exists() {
		return ctx.Float64()
	}
	return defaultValue
}

// ApiVersion retrieves the API version from the [meta] instance.
//
// This function checks if the [meta] instance is available (non-nil) before attempting
// to retrieve the `apiVersion`. If the [meta] instance is unavailable, it returns an empty string.
//
// Returns:
//   - The API version as a string if available.
//   - An empty string if the [meta] instance is unavailable.
func (m *meta) ApiVersion() string {
	if !m.Available() {
		return ""
	}
	return m.apiVersion
}

// RequestID retrieves the request ID from the [meta] instance.
//
// This function checks if the [meta] instance is available (non-nil) before retrieving
// the `requestID`. If the [meta] instance is unavailable, it returns an empty string.
//
// Returns:
//   - The request ID as a string if available.
//   - An empty string if the [meta] instance is unavailable.
func (m *meta) RequestID() string {
	if !m.Available() {
		return ""
	}
	m.autoRequestID()
	return m.requestID
}

// Locale retrieves the locale from the [meta] instance.
//
// This function checks if the [meta] instance is available (non-nil) before retrieving
// the `locale`. If the [meta] instance is unavailable, it returns an empty string.
//
// Returns:
//   - The locale as a string if available.
//   - An empty string if the [meta] instance is unavailable.
func (m *meta) Locale() string {
	if !m.Available() {
		return ""
	}
	return m.locale
}

// RequestedTime retrieves the requested time from the [meta] instance.
//
// This function checks if the [meta] instance is available (non-nil) before retrieving
// the `requestedTime`. If the [meta] instance is unavailable, it returns the zero value
// of `time.Time` (January 1, year 1, 00:00:00 UTC).
//
// Returns:
//   - The requested time as a `time.Time` object if available.
//   - The zero value of `time.Time` if the [meta] instance is unavailable.
func (m *meta) RequestedTime() time.Time {
	if !m.Available() {
		return time.Time{}
	}
	return m.requestedTime
}

// CustomFields retrieves the custom fields from the [meta] instance.
//
// This function checks if the [meta] instance is available (non-nil) before retrieving
// the `customFields`. If the [meta] instance is unavailable, it returns `nil`.
//
// Returns:
//   - A map of custom fields if available.
//   - `nil` if the [meta] instance is unavailable.
func (m *meta) CustomFields() map[string]any {
	if !m.Available() {
		return nil
	}
	return m.customFields
}

// IncreaseDeltaCnt increments the delta count for the [meta] instance by 1.
// Represents an additional change introduced by payload normalization or transformation.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) IncreaseDeltaCnt() *meta {
	m.deltaCnt++
	return m
}

// DecreaseDeltaCnt decrements the delta count for the [meta] instance by 1.
// Represents the removal of a change introduced by payload normalization or transformation.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) DecreaseDeltaCnt() *meta {
	m.deltaCnt--
	return m
}

// Respond generates a map representation of the [meta] instance.
//
// This method collects various fields of the [meta] instance (e.g., `apiVersion`, `requestID`, etc.)
// and organizes them into a key-value map. Only fields that are available and valid
// (e.g., non-empty or initialized) are included in the resulting map.
//
// Fields included in the response:
//   - `api_version`: The API version, if present.
//   - `request_id`: The unique request identifier, if present.
//   - `locale`: The locale information, if present.
//   - `requested_time`: The requested time, if it is initialized.
//   - `custom_fields`: A map of custom fields, if present.
//
// Returns:
//   - A `map[string]interface{}` containing the structured metadata.
func (m *meta) Respond() map[string]any {
	mk := make(map[string]any)
	if !m.Available() {
		return mk
	}
	if m.IsApiVersionPresent() {
		mk["api_version"] = m.apiVersion
	}
	if m.IsRequestIDPresent() {
		mk["request_id"] = m.requestID
	}
	if m.IsLocalePresent() {
		mk["locale"] = m.locale
	}
	if m.IsRequestedTimePresent() {
		mk["requested_time"] = m.requestedTime.Format("2006-01-02 15:04:05.999999")
	}
	if m.IsCustomPresent() {
		mk["custom_fields"] = m.customFields
	}
	if m.IsDeltaCntPresent() {
		mk["delta_cnt"] = m.deltaCnt
	}
	if m.IsDeltaValuePresent() {
		mk["delta_value"] = m.deltaValue
	}
	return mk
}

// JSON serializes the [meta] instance into a compact JSON string.
//
// This function uses the `encoding.JSON` utility to create a compact JSON representation
// of the [meta] instance. The resulting string is formatted without additional whitespace,
// suitable for efficient storage or transmission of metadata.
//
// Returns:
//   - A compact JSON string representation of the [meta] instance.
func (m *meta) JSON() string {
	return jsonpass(m.Respond())
}

// JSONPretty serializes the [meta] instance into a prettified JSON string.
//
// This function calls the `encoding.JSONPretty` utility to produce a formatted, human-readable
// JSON string representation of the [meta] instance. The output is useful for debugging
// or inspecting metadata in a more structured format.
//
// Returns:
//   - A prettified JSON string representation of the [meta] instance.
func (m *meta) JSONPretty() string {
	return jsonpretty(m.Respond())
}

// RandRequestID generates and sets a random request ID for the [meta] instance.
//
// This function utilizes the `CryptoID` function to generate a unique request ID
// and assigns it to the `requestID` field of the [meta] instance. The modified
// [meta] instance is returned to allow for method chaining.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) RandRequestID() *meta {
	return m.forceRequestID()
}

// RandDeltaValue sets a random delta value for the [meta] instance.
//
// This function assigns the current Unix nanosecond timestamp (converted to float64)
// to the `deltaValue` field of the [meta] instance. The modified [meta] instance
// is returned to allow for method chaining.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) RandDeltaValue() *meta {
	return m.WithDeltaValue(randn.RandFt64())
}

// Equal checks if the current [meta] instance is equal to another [meta] instance.
//
// This function compares the fields of the current [meta] instance with those of another [meta] instance.
// It returns `true` if all fields are equal, and `false` otherwise. The comparison includes
// checking for nil values, field-by-field equality, and deep comparison of the `customFields` map.
//
// Parameters:
//   - `other`: A pointer to another [meta] instance to compare against.
//
// Returns:
//   - A boolean value indicating whether the two [meta] instances are equal.
func (m *meta) Equal(other *meta) bool {
	if m == other {
		return true
	}
	if other == nil {
		return false
	}
	if m.apiVersion != other.apiVersion {
		return false
	}
	if m.requestID != other.requestID {
		return false
	}
	if m.locale != other.locale {
		return false
	}
	if !m.requestedTime.Equal(other.requestedTime) {
		return false
	}
	if len(m.customFields) != len(other.customFields) {
		return false
	}
	for k, v := range m.customFields {
		if ov, ok := other.customFields[k]; !ok || ov != v {
			return false
		}
	}
	if m.deltaValue != other.deltaValue {
		return false
	}
	if m.deltaCnt != other.deltaCnt {
		return false
	}
	return true
}

// String provides a string representation of the [meta] instance.
//
// This function constructs a string that includes the values of the [meta] instance's fields,
// such as `apiVersion`, `requestID`, `locale`, `requestedTime`, `customFields`, `deltaCnt`, and `deltaValue`.
// The output is formatted in a readable manner, making it useful for logging or debugging purposes.
// If the [meta] instance is nil, it returns an empty string.
//
// Returns:
//   - A string representation of the [meta] instance, including its fields, or an empty string if the instance is nil.
func (m *meta) String() string {
	sw := strchain.New()
	if m == nil {
		return sw.String()
	}
	if m.IsApiVersionPresent() {
		sw.AppendF("api_version=%s", m.apiVersion)
	}
	if m.IsRequestIDPresent() {
		sw.Space()
		sw.AppendF("request_id=%s", m.requestID)
	}
	if m.IsLocalePresent() {
		sw.Space()
		sw.AppendF("locale=%s", m.locale)
	}
	if m.IsRequestedTimePresent() {
		sw.Space()
		sw.AppendF("requested_time=%s", m.requestedTime.Format("2006-01-02 15:04:05.999999"))
	}
	if m.IsCustomPresent() {
		sw.Space()
		sw.AppendF("custom_fields=%v", conv.StringOrEmpty(m.customFields))
	}
	if m.IsDeltaCntPresent() {
		sw.Space()
		sw.AppendF("delta_cnt=%d", m.deltaCnt)
	}
	if m.IsDeltaValuePresent() {
		sw.Space()
		sw.AppendF("delta_value=%f", m.deltaValue)
	}
	return sw.String()
}

// Logging dispatches a structured log entry for this response using [slogger] with the log message set to a default or custom message.
// The log level is automatically selected based on the HTTP status code range:
//
//   - 1xx → Debug  (informational)
//   - 2xx → Info   (success)
//   - 3xx → Warn   (redirection)
//   - 4xx → Error  (client error)
//   - 5xx → Error  (server error; [slogger.Logger.Fatal] is intentionally avoided because it calls os.Exit(1))
//   - other → Trace (no status code set)
//
// The log field key is "META" and its value is the structured map returned
// by [meta.Respond], serialized as JSON by the active formatter.
//
// # Thread-safety
//
// Logging is safe for concurrent use. The supplied logger is never mutated: a
// goroutine-local child is derived via [slogger.Logger.With] on every call so
// that the caller-skip adjustment and caller-enable flag stay local to the
// current goroutine. Concurrent callers sharing the same *[slogger.Logger]
// will not race. The meta field (locale) is read exactly once per call to give a consistent snapshot; meta fields are expected to be immutable after construction via [Meta] / [With*] options.
//
// # Caller reporting
//
// Caller information is always enabled for this call. callerSkip is set to 3
// to skip Logging, the slogger trampoline (Trace/Debug/Info/Warn/Error), and the caller of Logging, so the reported file and line resolve to the actual call site of Logging.
//
// Parameters:
//   - `logger`: optional *[slogger.Logger] to use. When omitted or nil, the
//     package-level global logger ([slogger.GlobalLogger]) is used.
//
// Returns:
//
// the receiver *meta unchanged, enabling method chaining.
//
// Example:
//
//	replify.Meta().
//	    WithLocale("en-US").
//	    WithCustomFieldKV("request_id", reqID).
//	    Logging()
func (m *meta) Logging(logger ...*slogger.Logger) *meta {
	if m == nil {
		return m
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	msg := strutil.DefaultIfEmpty(m.locale, "replify::meta::logging")

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	logAtLevel(child, slogger.InfoLevel, msg, slogger.JSON("META", m.Respond()))
	return m
}

// Slogging dispatches a structured log entry for this response using [slogger] with the log message set to the meta's string representation.
// The log level is automatically selected based on the HTTP status code range:
//
//   - 1xx → Debug  (informational)
//   - 2xx → Info   (success)
//   - 3xx → Warn   (redirection)
//   - 4xx → Error  (client error)
//   - 5xx → Error  (server error; [slogger.Logger.Fatal] is intentionally avoided because it calls os.Exit(1))
//   - other → Trace (no status code set)
//
// The log field key is "META" and its value is the structured map returned
// by [meta.Respond], serialized as Text by the active formatter.
//
// # Thread-safety
//
// Slogging is safe for concurrent use. The supplied logger is never mutated: a
// goroutine-local child is derived via [slogger.Logger.With] on every call so
// that the caller-skip adjustment and caller-enable flag stay local to the
// current goroutine. Concurrent callers sharing the same *[slogger.Logger]
// will not race. The meta field (locale) is read exactly once per call to give a consistent snapshot; meta fields are expected to be immutable after construction via [Meta] / [With*] options.
//
// # Caller reporting
//
// Caller information is always enabled for this call. callerSkip is set to 3
// to skip Slogging, the slogger trampoline (Trace/Debug/Info/Warn/Error), and the caller of Slogging, so the reported file and line resolve to the actual call site of Slogging.
//
// Parameters:
//   - `logger`: optional *[slogger.Logger] to use. When omitted or nil, the
//     package-level global logger ([slogger.GlobalLogger]) is used.
//
// Returns:
//
// the receiver *meta unchanged, enabling method chaining.
//
// Example:
//
//	replify.Meta().
//	    WithLocale("en-US").
//	    WithCustomFieldKV("request_id", reqID).
//	    Slogging()
func (m *meta) Slogging(logger ...*slogger.Logger) *meta {
	if m == nil {
		return m
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	slogAtLevel(child, slogger.InfoLevel, m.String())
	return m
}

// Reply creates a new [M] instance that wraps the current [meta] instance.
//
// This function constructs a new [M] instance, which is a wrapper around the current [meta] instance.
// The returned [M] instance can be used to facilitate structured responses or further processing.
//
// Returns:
//   - A new [M] instance containing the current [meta] instance.
func (m *meta) Reply() M {
	return M{meta: m}
}

// ReplyPtr creates a new [M] instance that wraps the current [meta] instance and returns a pointer to it.
//
// This function constructs a new [M] instance, which is a wrapper around the current [meta] instance.
// The returned pointer to the [M] instance can be used to facilitate structured responses or further processing.
//
// Returns:
//   - A pointer to a new [M] instance containing the current [meta] instance.
func (m *meta) ReplyPtr() *M {
	return &M{meta: m}
}

// autoRequestID generates and sets a random request ID for the [meta] instance if it is not already present.
//
// This function checks if the `requestID` field of the [meta] instance is already set. If it is not present,
// it generates a new random request ID using the `CryptoID` function and assigns it to the `requestID` field.
// The modified [meta] instance is returned to allow for method chaining.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) autoRequestID() *meta {
	if !m.IsRequestIDPresent() {
		m.forceRequestID()
	}
	return m
}

// forceRequestID generates and sets a random request ID for the [meta] instance, regardless of its current state.
//
// This function unconditionally generates a new random request ID using the `CryptoID` function and assigns it
// to the `requestID` field of the [meta] instance. The modified [meta] instance is returned to allow for method chaining.
//
// Returns:
//   - A pointer to the modified [meta] instance, enabling method chaining.
func (m *meta) forceRequestID() *meta {
	m.WithRequestID(randn.NewXID().String())
	return m
}
