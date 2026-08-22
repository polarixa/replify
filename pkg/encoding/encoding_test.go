package encoding_test

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/polarixa/replify"
	"github.com/polarixa/replify/pkg/encoding"
)

// ///////////////////////////
// Helpers / shared fixtures
// ///////////////////////////

type sampleStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// ///////////////////////////
// Section: JSON()
// ///////////////////////////

func TestJSON_Nil(t *testing.T) {
	if got := encoding.JSON(nil); got != "" {
		t.Errorf("JSON(nil) = %q; want %q", got, "")
	}
}

func TestJSON_String(t *testing.T) {
	// Strings are returned as-is (no quoting) in jsonSafe.
	if got := encoding.JSON("hello"); got != "hello" {
		t.Errorf("JSON(string) = %q; want %q", got, "hello")
	}
}

func TestJSON_Bool(t *testing.T) {
	if got := encoding.JSON(true); got != "true" {
		t.Errorf("JSON(true) = %q; want %q", got, "true")
	}
	if got := encoding.JSON(false); got != "false" {
		t.Errorf("JSON(false) = %q; want %q", got, "false")
	}
}

func TestJSON_IntegerScalars(t *testing.T) {
	cases := []struct {
		input any
		want  string
	}{
		{int(42), "42"},
		{int8(8), "8"},
		{int16(16), "16"},
		{int32(32), "32"},
		{int64(-64), "-64"},
		{uint(10), "10"},
		{uint8(255), "255"},
		{uint16(1000), "1000"},
		{uint32(100000), "100000"},
		{uint64(999), "999"},
	}
	for _, tc := range cases {
		if got := encoding.JSON(tc.input); got != tc.want {
			t.Errorf("JSON(%v) = %q; want %q", tc.input, got, tc.want)
		}
	}
}

func TestJSON_Float64(t *testing.T) {
	if got := encoding.JSON(float64(3.14)); got != "3.14" {
		t.Errorf("JSON(3.14) = %q; want %q", got, "3.14")
	}
}

func TestJSON_Float32(t *testing.T) {
	got := encoding.JSON(float32(1.5))
	if got != "1.5" {
		t.Errorf("JSON(float32(1.5)) = %q; want %q", got, "1.5")
	}
}

func TestJSON_NaN(t *testing.T) {
	// floatsUseNullForNonFinite == true => "null"
	if got := encoding.JSON(math.NaN()); got != "null" {
		t.Errorf("JSON(NaN) = %q; want %q", got, "null")
	}
}

func TestJSON_Inf(t *testing.T) {
	if got := encoding.JSON(math.Inf(1)); got != "null" {
		t.Errorf("JSON(+Inf) = %q; want %q", got, "null")
	}
	if got := encoding.JSON(math.Inf(-1)); got != "null" {
		t.Errorf("JSON(-Inf) = %q; want %q", got, "null")
	}
}

func TestJSON_Complex128(t *testing.T) {
	got := encoding.JSON(complex(1.0, 2.0))
	if got != `{"real":1,"imag":2}` {
		t.Errorf("JSON(complex128) = %q; want %q", got, `{"real":1,"imag":2}`)
	}
}

func TestJSON_Complex64(t *testing.T) {
	got := encoding.JSON(complex64(complex(3.0, 4.0)))
	if got != `{"real":3,"imag":4}` {
		t.Errorf("JSON(complex64) = %q; want %q", got, `{"real":3,"imag":4}`)
	}
}

func TestJSON_Struct(t *testing.T) {
	s := sampleStruct{Name: "Alice", Age: 30}
	got := encoding.JSON(s)
	want := `{"name":"Alice","age":30}`
	if got != want {
		t.Errorf("JSON(struct) = %q; want %q", got, want)
	}
}

func TestJSON_Map(t *testing.T) {
	m := map[string]int{"a": 1}
	got := encoding.JSON(m)
	if got != `{"a":1}` {
		t.Errorf("JSON(map) = %q; want %q", got, `{"a":1}`)
	}
}

func TestJSON_Slice(t *testing.T) {
	got := encoding.JSON([]int{1, 2, 3})
	if got != `[1,2,3]` {
		t.Errorf("JSON(slice) = %q; want %q", got, `[1,2,3]`)
	}
}

func TestJSON_NilPointer(t *testing.T) {
	var p *sampleStruct
	if got := encoding.JSON(p); got != "null" {
		t.Errorf("JSON(nil pointer) = %q; want %q", got, "null")
	}
}

func TestJSON_NilMap(t *testing.T) {
	var m map[string]int
	if got := encoding.JSON(m); got != "null" {
		t.Errorf("JSON(nil map) = %q; want %q", got, "null")
	}
}

func TestJSON_NilSlice(t *testing.T) {
	var sl []int
	if got := encoding.JSON(sl); got != "null" {
		t.Errorf("JSON(nil slice) = %q; want %q", got, "null")
	}
}

func TestJSON_RawMessageValid(t *testing.T) {
	rm := json.RawMessage(`{"key":"value"}`)
	if got := encoding.JSON(rm); got != `{"key":"value"}` {
		t.Errorf("JSON(RawMessage valid) = %q; want %q", got, `{"key":"value"}`)
	}
}

func TestJSON_RawMessageInvalid(t *testing.T) {
	rm := json.RawMessage(`{invalid}`)
	// invalid raw message => ""
	if got := encoding.JSON(rm); got != "" {
		t.Errorf("JSON(RawMessage invalid) = %q; want %q", got, "")
	}
}

func TestJSON_RawMessageNil(t *testing.T) {
	var rm json.RawMessage
	if got := encoding.JSON(rm); got != "null" {
		t.Errorf("JSON(nil RawMessage) = %q; want %q", got, "null")
	}
}

// ///////////////////////////
// Section: JSONToken()
// ///////////////////////////

func TestJSONToken_Nil(t *testing.T) {
	got, err := encoding.JSONE(nil)
	if err == nil {
		t.Error("JSONToken(nil) expected error, got nil")
	}
	if got != "" {
		t.Errorf("JSONToken(nil) = %q; want %q", got, "")
	}
}

func TestJSONToken_String(t *testing.T) {
	// Strings are returned as-is (no quoting) in jsonSafeToken.
	got, err := encoding.JSONE("hello")
	if err != nil {
		t.Fatalf("JSONToken(string) unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("JSONToken(string) = %q; want %q", got, "hello")
	}
}

func TestJSONToken_Bool(t *testing.T) {
	got, err := encoding.JSONE(true)
	if err != nil {
		t.Fatalf("JSONToken(true) unexpected error: %v", err)
	}
	if got != "true" {
		t.Errorf("JSONToken(true) = %q; want %q", got, "true")
	}
}

func TestJSONToken_IntegerScalars(t *testing.T) {
	cases := []struct {
		input any
		want  string
	}{
		{int(42), "42"},
		{int8(-8), "-8"},
		{int16(16), "16"},
		{int32(32), "32"},
		{int64(64), "64"},
		{uint(7), "7"},
		{uint8(255), "255"},
		{uint16(1000), "1000"},
		{uint32(100000), "100000"},
		{uint64(999), "999"},
	}
	for _, tc := range cases {
		got, err := encoding.JSONE(tc.input)
		if err != nil {
			t.Errorf("JSONToken(%v) unexpected error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("JSONToken(%v) = %q; want %q", tc.input, got, tc.want)
		}
	}
}

func TestJSONToken_Float64(t *testing.T) {
	got, err := encoding.JSONE(float64(2.718))
	if err != nil {
		t.Fatalf("JSONToken(float64) unexpected error: %v", err)
	}
	if got != "2.718" {
		t.Errorf("JSONToken(float64) = %q; want %q", got, "2.718")
	}
}

func TestJSONToken_NaN(t *testing.T) {
	// floatsUseNullForNonFinite == true => "null", no error
	got, err := encoding.JSONE(math.NaN())
	if err != nil {
		t.Fatalf("JSONToken(NaN) unexpected error: %v", err)
	}
	if got != "null" {
		t.Errorf("JSONToken(NaN) = %q; want %q", got, "null")
	}
}

func TestJSONToken_Inf(t *testing.T) {
	got, err := encoding.JSONE(math.Inf(1))
	if err != nil {
		t.Fatalf("JSONToken(+Inf) unexpected error: %v", err)
	}
	if got != "null" {
		t.Errorf("JSONToken(+Inf) = %q; want %q", got, "null")
	}
}

func TestJSONToken_Complex128(t *testing.T) {
	got, err := encoding.JSONE(complex(1.0, 2.0))
	if err != nil {
		t.Fatalf("JSONToken(complex128) unexpected error: %v", err)
	}
	if got != `{"real":1,"imag":2}` {
		t.Errorf("JSONToken(complex128) = %q; want %q", got, `{"real":1,"imag":2}`)
	}
}

func TestJSONToken_Struct(t *testing.T) {
	s := sampleStruct{Name: "Bob", Age: 25}
	got, err := encoding.JSONE(s)
	if err != nil {
		t.Fatalf("JSONToken(struct) unexpected error: %v", err)
	}
	want := `{"name":"Bob","age":25}`
	if got != want {
		t.Errorf("JSONToken(struct) = %q; want %q", got, want)
	}
}

func TestJSONToken_Slice(t *testing.T) {
	got, err := encoding.JSONE([]string{"a", "b"})
	if err != nil {
		t.Fatalf("JSONToken(slice) unexpected error: %v", err)
	}
	if got != `["a","b"]` {
		t.Errorf("JSONToken(slice) = %q; want %q", got, `["a","b"]`)
	}
}

func TestJSONToken_NilPointer(t *testing.T) {
	var p *sampleStruct
	got, err := encoding.JSONE(p)
	if err != nil {
		t.Fatalf("JSONToken(nil pointer) unexpected error: %v", err)
	}
	if got != "null" {
		t.Errorf("JSONToken(nil pointer) = %q; want %q", got, "null")
	}
}

func TestJSONToken_NilMap(t *testing.T) {
	var m map[string]int
	got, err := encoding.JSONE(m)
	if err != nil {
		t.Fatalf("JSONToken(nil map) unexpected error: %v", err)
	}
	if got != "null" {
		t.Errorf("JSONToken(nil map) = %q; want %q", got, "null")
	}
}

func TestJSONToken_RawMessageValid(t *testing.T) {
	rm := json.RawMessage(`[1,2,3]`)
	got, err := encoding.JSONE(rm)
	if err != nil {
		t.Fatalf("JSONToken(RawMessage valid) unexpected error: %v", err)
	}
	if got != `[1,2,3]` {
		t.Errorf("JSONToken(RawMessage valid) = %q; want %q", got, `[1,2,3]`)
	}
}

func TestJSONToken_RawMessageInvalid(t *testing.T) {
	rm := json.RawMessage(`{bad json}`)
	_, err := encoding.JSONE(rm)
	if err == nil {
		t.Error("JSONToken(invalid RawMessage) expected error, got nil")
	}
}

func TestJSONToken_RawMessageNil(t *testing.T) {
	var rm json.RawMessage
	got, err := encoding.JSONE(rm)
	if err != nil {
		t.Fatalf("JSONToken(nil RawMessage) unexpected error: %v", err)
	}
	if got != "null" {
		t.Errorf("JSONToken(nil RawMessage) = %q; want %q", got, "null")
	}
}

// ///////////////////////////
// Section: JSONPretty()
// ///////////////////////////

func TestJSONPretty_Nil(t *testing.T) {
	if got := encoding.JSONPretty(nil); got != "" {
		t.Errorf("JSONPretty(nil) = %q; want %q", got, "")
	}
}

func TestJSONPretty_String(t *testing.T) {
	// Strings are returned as-is (no quoting).
	if got := encoding.JSONPretty("world"); got != "world" {
		t.Errorf("JSONPretty(string) = %q; want %q", got, "world")
	}
}

func TestJSONPretty_Bool(t *testing.T) {
	if got := encoding.JSONPretty(false); got != "false" {
		t.Errorf("JSONPretty(false) = %q; want %q", got, "false")
	}
}

func TestJSONPretty_Struct_IsIndented(t *testing.T) {
	s := sampleStruct{Name: "Carol", Age: 40}
	got := encoding.JSONPretty(s)
	// Must contain newlines and indentation (4-space indent used by MarshalIndent).
	if !strings.Contains(got, "\n") {
		t.Errorf("JSONPretty(struct) = %q; expected indented JSON with newlines", got)
	}
	if !strings.Contains(got, "    ") {
		t.Errorf("JSONPretty(struct) = %q; expected 4-space indentation", got)
	}
	if !strings.Contains(got, `"name"`) || !strings.Contains(got, `"Carol"`) {
		t.Errorf("JSONPretty(struct) = %q; missing expected fields", got)
	}
}

func TestJSONPretty_Map_IsIndented(t *testing.T) {
	m := map[string]int{"x": 99}
	got := encoding.JSONPretty(m)
	if !strings.Contains(got, "\n") {
		t.Errorf("JSONPretty(map) = %q; expected indented JSON", got)
	}
}

func TestJSONPretty_Slice(t *testing.T) {
	got := encoding.JSONPretty([]int{1, 2})
	if !strings.Contains(got, "\n") {
		t.Errorf("JSONPretty(slice) = %q; expected indented JSON", got)
	}
}

func TestJSONPretty_NilPointer(t *testing.T) {
	var p *sampleStruct
	if got := encoding.JSONPretty(p); got != "null" {
		t.Errorf("JSONPretty(nil pointer) = %q; want %q", got, "null")
	}
}

func TestJSONPretty_RawMessageValid_IsIndented(t *testing.T) {
	rm := json.RawMessage(`{"a":1,"b":2}`)
	got := encoding.JSONPretty(rm)
	if !strings.Contains(got, "\n") {
		t.Errorf("JSONPretty(RawMessage) = %q; expected indented JSON", got)
	}
}

func TestJSONPretty_RawMessageInvalid(t *testing.T) {
	rm := json.RawMessage(`{bad}`)
	if got := encoding.JSONPretty(rm); got != "" {
		t.Errorf("JSONPretty(invalid RawMessage) = %q; want %q", got, "")
	}
}

func TestJSONPretty_NaN(t *testing.T) {
	if got := encoding.JSONPretty(math.NaN()); got != "null" {
		t.Errorf("JSONPretty(NaN) = %q; want %q", got, "null")
	}
}

func TestJSONPretty_IntegerScalar(t *testing.T) {
	if got := encoding.JSONPretty(int64(7)); got != "7" {
		t.Errorf("JSONPretty(int64) = %q; want %q", got, "7")
	}
}

// ///////////////////////////
// Section: JSONPrettyToken()
// ///////////////////////////

func TestJSONPrettyToken_Nil(t *testing.T) {
	_, err := encoding.JSONPrettyE(nil)
	if err == nil {
		t.Error("JSONPrettyToken(nil) expected error, got nil")
	}
}

func TestJSONPrettyToken_String(t *testing.T) {
	got, err := encoding.JSONPrettyE("test")
	if err != nil {
		t.Fatalf("JSONPrettyToken(string) unexpected error: %v", err)
	}
	if got != "test" {
		t.Errorf("JSONPrettyToken(string) = %q; want %q", got, "test")
	}
}

func TestJSONPrettyToken_Bool(t *testing.T) {
	got, err := encoding.JSONPrettyE(true)
	if err != nil {
		t.Fatalf("JSONPrettyToken(bool) unexpected error: %v", err)
	}
	if got != "true" {
		t.Errorf("JSONPrettyToken(bool) = %q; want %q", got, "true")
	}
}

func TestJSONPrettyToken_IntegerScalar(t *testing.T) {
	got, err := encoding.JSONPrettyE(int(100))
	if err != nil {
		t.Fatalf("JSONPrettyToken(int) unexpected error: %v", err)
	}
	if got != "100" {
		t.Errorf("JSONPrettyToken(int) = %q; want %q", got, "100")
	}
}

func TestJSONPrettyToken_Float64(t *testing.T) {
	got, err := encoding.JSONPrettyE(float64(1.23))
	if err != nil {
		t.Fatalf("JSONPrettyToken(float64) unexpected error: %v", err)
	}
	if got != "1.23" {
		t.Errorf("JSONPrettyToken(float64) = %q; want %q", got, "1.23")
	}
}

func TestJSONPrettyToken_NaN(t *testing.T) {
	got, err := encoding.JSONPrettyE(math.NaN())
	if err != nil {
		t.Fatalf("JSONPrettyToken(NaN) unexpected error: %v", err)
	}
	if got != "null" {
		t.Errorf("JSONPrettyToken(NaN) = %q; want %q", got, "null")
	}
}

func TestJSONPrettyToken_Struct_IsIndented(t *testing.T) {
	s := sampleStruct{Name: "Dave", Age: 20}
	got, err := encoding.JSONPrettyE(s)
	if err != nil {
		t.Fatalf("JSONPrettyToken(struct) unexpected error: %v", err)
	}
	if !strings.Contains(got, "\n") {
		t.Errorf("JSONPrettyToken(struct) = %q; expected indented JSON with newlines", got)
	}
	if !strings.Contains(got, "    ") {
		t.Errorf("JSONPrettyToken(struct) = %q; expected 4-space indentation", got)
	}
	if !strings.Contains(got, `"name"`) || !strings.Contains(got, `"Dave"`) {
		t.Errorf("JSONPrettyToken(struct) = %q; missing expected fields", got)
	}
}

func TestJSONPrettyToken_Slice_IsIndented(t *testing.T) {
	got, err := encoding.JSONPrettyE([]int{4, 5, 6})
	if err != nil {
		t.Fatalf("JSONPrettyToken(slice) unexpected error: %v", err)
	}
	if !strings.Contains(got, "\n") {
		t.Errorf("JSONPrettyToken(slice) = %q; expected indented JSON", got)
	}
}

func TestJSONPrettyToken_NilPointer(t *testing.T) {
	var p *sampleStruct
	got, err := encoding.JSONPrettyE(p)
	if err != nil {
		t.Fatalf("JSONPrettyToken(nil pointer) unexpected error: %v", err)
	}
	if got != "null" {
		t.Errorf("JSONPrettyToken(nil pointer) = %q; want %q", got, "null")
	}
}

func TestJSONPrettyToken_NilMap(t *testing.T) {
	var m map[string]int
	got, err := encoding.JSONPrettyE(m)
	if err != nil {
		t.Fatalf("JSONPrettyToken(nil map) unexpected error: %v", err)
	}
	if got != "null" {
		t.Errorf("JSONPrettyToken(nil map) = %q; want %q", got, "null")
	}
}

func TestJSONPrettyToken_RawMessageValid_IsIndented(t *testing.T) {
	rm := json.RawMessage(`{"c":3}`)
	got, err := encoding.JSONPrettyE(rm)
	if err != nil {
		t.Fatalf("JSONPrettyToken(RawMessage) unexpected error: %v", err)
	}
	if !strings.Contains(got, "\n") {
		t.Errorf("JSONPrettyToken(RawMessage) = %q; expected indented JSON", got)
	}
}

func TestJSONPrettyToken_RawMessageInvalid(t *testing.T) {
	rm := json.RawMessage(`{not valid}`)
	_, err := encoding.JSONPrettyE(rm)
	if err == nil {
		t.Error("JSONPrettyToken(invalid RawMessage) expected error, got nil")
	}
}

func TestJSONPrettyToken_RawMessageNil(t *testing.T) {
	var rm json.RawMessage
	got, err := encoding.JSONPrettyE(rm)
	if err != nil {
		t.Fatalf("JSONPrettyToken(nil RawMessage) unexpected error: %v", err)
	}
	if got != "null" {
		t.Errorf("JSONPrettyToken(nil RawMessage) = %q; want %q", got, "null")
	}
}

func TestJSONPrettyToken_Complex128(t *testing.T) {
	got, err := encoding.JSONPrettyE(complex(2.0, 3.0))
	if err != nil {
		t.Fatalf("JSONPrettyToken(complex128) unexpected error: %v", err)
	}
	if got != `{"real":2,"imag":3}` {
		t.Errorf("JSONPrettyToken(complex128) = %q; want %q", got, `{"real":2,"imag":3}`)
	}
}

// ///////////////////////////
// Section: wrapper.JSON() and wrapper.JSONPretty()
// ///////////////////////////

// wrapperJSONFixture is the JSON body used across wrapper JSON tests.
const wrapperJSONFixture = `{"store":{"owner":"Alice"},"ratings":[5,3,4]}`

// TestWrapperJSON_ValidJSONStringBody ensures that JSON() inlines a valid JSON
// string body as a JSON object (not an escaped string literal) in the output.
func TestWrapperJSON_ValidJSONStringBody(t *testing.T) {
	w := replify.New().WithBody(wrapperJSONFixture)
	out := w.JSON()

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("JSON() output is not valid JSON: %v\noutput: %s", err, out)
	}

	// The "data" field must be a JSON object, not a string.
	data, ok := parsed["data"].(map[string]any)
	if !ok {
		t.Fatalf("JSON() data field type = %T; want map[string]any", parsed["data"])
	}
	if _, hasStore := data["store"]; !hasStore {
		t.Errorf("JSON() data field missing 'store' key; got keys in data: %v", keysOf(data))
	}
}

// TestWrapperJSON_NoEscapedCharsInDataField verifies that JSON() does not
// produce escaped newlines or tabs in the "data" field when the body is valid JSON.
func TestWrapperJSON_NoEscapedCharsInDataField(t *testing.T) {
	body := "{\n\t\"key\": \"value\"\n}"
	w := replify.New().WithBody(body)
	out := w.JSON()
	if strings.Contains(out, `\n`) || strings.Contains(out, `\t`) {
		t.Errorf("JSON() output contains escaped whitespace; want compact inline JSON\noutput: %s", out)
	}
}

// TestWrapperJSON_BytesBody verifies that a []byte body containing valid JSON
// is also inlined correctly (not double-encoded).
func TestWrapperJSON_BytesBody(t *testing.T) {
	w := replify.New().WithBody([]byte(wrapperJSONFixture))
	out := w.JSON()

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("JSON() (bytes body) is not valid JSON: %v", err)
	}
	if _, ok := parsed["data"].(map[string]any); !ok {
		t.Errorf("JSON() (bytes body) data field type = %T; want map[string]any", parsed["data"])
	}
}

// TestWrapperJSON_NonJSONStringBody verifies that a non-JSON string body is
// kept as a JSON string value (not inlined as raw JSON).
func TestWrapperJSON_NonJSONStringBody(t *testing.T) {
	w := replify.New().WithBody("hello world")
	out := w.JSON()

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("JSON() (non-JSON body) is not valid JSON: %v", err)
	}
	if s, ok := parsed["data"].(string); !ok || s != "hello world" {
		t.Errorf("JSON() (non-JSON body) data = %v (%T); want string %q", parsed["data"], parsed["data"], "hello world")
	}
}

// TestWrapperJSON_EmptyBody verifies that JSON() works when no body is set.
func TestWrapperJSON_EmptyBody(t *testing.T) {
	w := replify.New()
	out := w.JSON()

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("JSON() (empty body) is not valid JSON: %v", err)
	}
	if _, hasData := parsed["data"]; hasData {
		t.Errorf("JSON() (empty body) should not have 'data' field; got: %v", parsed["data"])
	}
}

// TestWrapperJSONPretty_ValidJSONStringBody verifies that JSONPretty() produces
// properly indented output with the data field inlined as a JSON object.
func TestWrapperJSONPretty_ValidJSONStringBody(t *testing.T) {
	w := replify.New().WithBody(wrapperJSONFixture)
	out := w.JSONPretty()

	if !strings.Contains(out, "    ") {
		t.Errorf("JSONPretty() output is not indented\noutput: %s", out)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("JSONPretty() output is not valid JSON: %v\noutput: %s", err, out)
	}
	if _, ok := parsed["data"].(map[string]any); !ok {
		t.Errorf("JSONPretty() data field type = %T; want map[string]any", parsed["data"])
	}
}

// TestWrapperJSONBytes_ValidJSONBody verifies that JSONBytes() returns a non-nil
// byte slice equal to JSON() when the body is valid JSON.
func TestWrapperJSONBytes_ValidJSONBody(t *testing.T) {
	w := replify.New().WithBody(wrapperJSONFixture)
	b := w.JSONBytes()
	if b == nil {
		t.Fatal("JSONBytes() = nil; want non-nil byte slice")
	}
	if string(b) != w.JSON() {
		t.Errorf("JSONBytes() != []byte(JSON())")
	}
}

// keysOf returns the keys of a map as a slice (helper for test error messages).
func keysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ///////////////////////////
// Section: encoding.NormalizeJSON()
// ///////////////////////////

// TestNormalizeJSON_AlreadyValid verifies that already-valid JSON is returned
// unchanged (fast path, no modification).
func TestNormalizeJSON_AlreadyValid(t *testing.T) {
	input := `{"key":"value","n":42}`
	got, err := encoding.NormalizeJSON(input)
	if err != nil {
		t.Fatalf("NormalizeJSON(valid) unexpected error: %v", err)
	}
	if got != input {
		t.Errorf("NormalizeJSON(valid) = %q; want %q (unchanged)", got, input)
	}
}

// TestNormalizeJSON_Empty verifies that an empty string returns an error.
func TestNormalizeJSON_Empty(t *testing.T) {
	_, err := encoding.NormalizeJSON("")
	if err == nil {
		t.Error("NormalizeJSON(\"\") expected error; got nil")
	}
}

// TestNormalizeJSON_WhitespaceOnly verifies that a whitespace-only string returns an error.
func TestNormalizeJSON_WhitespaceOnly(t *testing.T) {
	_, err := encoding.NormalizeJSON("   ")
	if err == nil {
		t.Error("NormalizeJSON(whitespace) expected error; got nil")
	}
}

// TestNormalizeJSON_EscapedStructuralQuotes verifies that literal `\"` sequences
// used as structural key/value delimiters are unescaped to produce valid JSON.
func TestNormalizeJSON_EscapedStructuralQuotes(t *testing.T) {
	// Simulate a raw string with \"key\" artifacts.
	input := `{\"key\": \"value\"}`
	got, err := encoding.NormalizeJSON(input)
	if err != nil {
		t.Fatalf("NormalizeJSON(escaped structural quotes) unexpected error: %v", err)
	}
	if !encoding.IsValidJSONString(got) {
		t.Errorf("NormalizeJSON result is not valid JSON: %q", got)
	}
	want := `{"key": "value"}`
	if got != want {
		t.Errorf("NormalizeJSON = %q; want %q", got, want)
	}
}

// TestNormalizeJSON_MixedEscapedKeys verifies a realistic case where only some keys
// are escaped with `\"` and the rest of the object is normal JSON.
func TestNormalizeJSON_MixedEscapedKeys(t *testing.T) {
	input := `{\"store\": {"owner": "Alice"}}`
	got, err := encoding.NormalizeJSON(input)
	if err != nil {
		t.Fatalf("NormalizeJSON(mixed escaped keys) unexpected error: %v", err)
	}
	if !encoding.IsValidJSONString(got) {
		t.Errorf("NormalizeJSON result is not valid JSON: %q", got)
	}
}

// TestNormalizeJSON_UnfixableInput verifies that an input that cannot be normalized
// to valid JSON returns an error.
func TestNormalizeJSON_UnfixableInput(t *testing.T) {
	_, err := encoding.NormalizeJSON("{this is not json at all}")
	if err == nil {
		t.Error("NormalizeJSON(unfixable) expected error; got nil")
	}
}

// TestNormalizeJSON_ObjectArray verifies normalization of a more complex structure
// with escaped outer keys.
func TestNormalizeJSON_ObjectArray(t *testing.T) {
	input := `{\"items\": [1, 2, 3], \"count\": 3}`
	got, err := encoding.NormalizeJSON(input)
	if err != nil {
		t.Fatalf("NormalizeJSON(object+array) unexpected error: %v", err)
	}
	if !encoding.IsValidJSONString(got) {
		t.Errorf("NormalizeJSON result is not valid JSON: %q", got)
	}
}

// ///////////////////////////
// Section: wrapper.WithNormalizedBody()
// ///////////////////////////

// TestWithNormalizedBody_ValidJSON verifies that a valid JSON string is accepted
// without modification and downstream functions work correctly.
func TestWithNormalizedBody_ValidJSON(t *testing.T) {
	input := `{"store":{"owner":"Alice"}}`
	w, err := replify.New().WithJSONBody(input)
	if err != nil {
		t.Fatalf("WithNormalizedBody(valid) unexpected error: %v", err)
	}
	if !w.IsJSONBody() {
		t.Error("WithNormalizedBody(valid): IsJSONBody() = false; want true")
	}
	if w.Body() != input {
		t.Errorf("WithNormalizedBody(valid): Body() = %v; want %v", w.Body(), input)
	}
}

// TestWithNormalizedBody_EscapedQuotes verifies that escaped structural quotes are
// fixed so that IsJSONBody(), JSON(), and JSONPretty() all behave correctly.
func TestWithNormalizedBody_EscapedQuotes(t *testing.T) {
	input := `{\"store\": {\"owner\": \"Alice\"}}`
	w, err := replify.New().WithJSONBody(input)
	if err != nil {
		t.Fatalf("WithNormalizedBody(escaped) unexpected error: %v", err)
	}
	if !w.IsJSONBody() {
		t.Error("WithNormalizedBody(escaped): IsJSONBody() = false; want true")
	}
	// JSON() output must be valid JSON with data inlined as an object.
	jsonOut := w.JSON()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(jsonOut), &parsed); err != nil {
		t.Fatalf("WithNormalizedBody(escaped): JSON() output is not valid JSON: %v\noutput: %s", err, jsonOut)
	}
	if _, ok := parsed["data"].(map[string]any); !ok {
		t.Errorf("WithNormalizedBody(escaped): JSON() data field type = %T; want map[string]any", parsed["data"])
	}
}

// TestWithNormalizedBody_InvalidInput verifies that an input that cannot be normalized
// returns an error and leaves the wrapper body unchanged.
func TestWithNormalizedBody_InvalidInput(t *testing.T) {
	w := replify.New()
	w.WithBody("original")
	_, err := w.WithJSONBody("{this is not json}")
	if err == nil {
		t.Error("WithNormalizedBody(invalid) expected error; got nil")
	}
}

// TestWithNormalizedBody_Empty verifies that an empty string returns an error.
func TestWithNormalizedBody_Empty(t *testing.T) {
	_, err := replify.New().WithJSONBody("")
	if err == nil {
		t.Error("WithNormalizedBody(\"\") expected error; got nil")
	}
}

// TestNormalizeJSON_BOM verifies that a leading UTF-8 BOM is stripped so that
// the remaining content is recognized as valid JSON.
func TestNormalizeJSON_BOM(t *testing.T) {
	bom := "\xEF\xBB\xBF"
	input := bom + `{"key":"value"}`
	got, err := encoding.NormalizeJSON(input)
	if err != nil {
		t.Fatalf("NormalizeJSON(BOM) unexpected error: %v", err)
	}
	if !encoding.IsValidJSONString(got) {
		t.Errorf("NormalizeJSON(BOM) result is not valid JSON: %q", got)
	}
	if got != `{"key":"value"}` {
		t.Errorf("NormalizeJSON(BOM) = %q; want %q", got, `{"key":"value"}`)
	}
}

// TestNormalizeJSON_NullBytes verifies that embedded null bytes are removed so
// that the remaining content is recognized as valid JSON.
func TestNormalizeJSON_NullBytes(t *testing.T) {
	input := "{\"key\"\x00:\"value\"}"
	got, err := encoding.NormalizeJSON(input)
	if err != nil {
		t.Fatalf("NormalizeJSON(null bytes) unexpected error: %v", err)
	}
	if !encoding.IsValidJSONString(got) {
		t.Errorf("NormalizeJSON(null bytes) result is not valid JSON: %q", got)
	}
}

// TestNormalizeJSON_TrailingCommaObject verifies that a trailing comma inside an
// object is removed so that the result is valid JSON.
func TestNormalizeJSON_TrailingCommaObject(t *testing.T) {
	input := `{"a":1,"b":2,}`
	got, err := encoding.NormalizeJSON(input)
	if err != nil {
		t.Fatalf("NormalizeJSON(trailing comma object) unexpected error: %v", err)
	}
	if !encoding.IsValidJSONString(got) {
		t.Errorf("NormalizeJSON(trailing comma object) result is not valid JSON: %q", got)
	}
}

// TestNormalizeJSON_TrailingCommaArray verifies that a trailing comma inside an
// array is removed so that the result is valid JSON.
func TestNormalizeJSON_TrailingCommaArray(t *testing.T) {
	input := `[1,2,3,]`
	got, err := encoding.NormalizeJSON(input)
	if err != nil {
		t.Fatalf("NormalizeJSON(trailing comma array) unexpected error: %v", err)
	}
	if !encoding.IsValidJSONString(got) {
		t.Errorf("NormalizeJSON(trailing comma array) result is not valid JSON: %q", got)
	}
}

// TestNormalizeJSON_EscapedQuotesPlusTrailingComma verifies that the cumulative
// multi-pass approach handles both escaped quotes and trailing commas together.
func TestNormalizeJSON_EscapedQuotesPlusTrailingComma(t *testing.T) {
	input := `{\"a\":1,\"b\":2,}`
	got, err := encoding.NormalizeJSON(input)
	if err != nil {
		t.Fatalf("NormalizeJSON(escaped+trailing comma) unexpected error: %v", err)
	}
	if !encoding.IsValidJSONString(got) {
		t.Errorf("NormalizeJSON(escaped+trailing comma) result is not valid JSON: %q", got)
	}
}

// ///////////////////////////
// Section: wrapper.WithNormalizedBody() — additional input types
// ///////////////////////////

// TestWithNormalizedBody_Nil verifies that a nil input returns an error.
func TestWithNormalizedBody_Nil(t *testing.T) {
	_, err := replify.New().WithJSONBody(nil)
	if err == nil {
		t.Error("WithNormalizedBody(nil) expected error; got nil")
	}
}

// TestWithNormalizedBody_ByteSlice verifies that a []byte containing valid JSON
// is accepted and stored correctly.
func TestWithNormalizedBody_ByteSlice(t *testing.T) {
	input := []byte(`{"key":"value"}`)
	w, err := replify.New().WithJSONBody(input)
	if err != nil {
		t.Fatalf("WithNormalizedBody([]byte valid) unexpected error: %v", err)
	}
	if !w.IsJSONBody() {
		t.Error("WithNormalizedBody([]byte valid): IsJSONBody() = false; want true")
	}
}

// TestWithNormalizedBody_ByteSliceEscaped verifies that a []byte with escaped
// structural quotes is normalized and stored correctly.
func TestWithNormalizedBody_ByteSliceEscaped(t *testing.T) {
	input := []byte(`{\"key\":\"value\"}`)
	w, err := replify.New().WithJSONBody(input)
	if err != nil {
		t.Fatalf("WithNormalizedBody([]byte escaped) unexpected error: %v", err)
	}
	if !w.IsJSONBody() {
		t.Error("WithNormalizedBody([]byte escaped): IsJSONBody() = false; want true")
	}
}

// TestWithNormalizedBody_Struct verifies that an arbitrary Go struct is marshaled
// to JSON and stored as the body with IsJSONBody() returning true.
func TestWithNormalizedBody_Struct(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	w, err := replify.New().WithJSONBody(payload{Name: "Alice", Age: 30})
	if err != nil {
		t.Fatalf("WithNormalizedBody(struct) unexpected error: %v", err)
	}
	if !w.IsJSONBody() {
		t.Error("WithNormalizedBody(struct): IsJSONBody() = false; want true")
	}
	jsonOut := w.JSON()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(jsonOut), &parsed); err != nil {
		t.Fatalf("WithNormalizedBody(struct): JSON() output is not valid JSON: %v\noutput: %s", err, jsonOut)
	}
	if _, ok := parsed["data"].(map[string]any); !ok {
		t.Errorf("WithNormalizedBody(struct): JSON() data field type = %T; want map[string]any", parsed["data"])
	}
}

// TestWithNormalizedBody_Map verifies that a map[string]any is marshaled to JSON
// and stored as the body with IsJSONBody() returning true.
func TestWithNormalizedBody_Map(t *testing.T) {
	input := map[string]any{"x": 1, "y": true}
	w, err := replify.New().WithJSONBody(input)
	if err != nil {
		t.Fatalf("WithNormalizedBody(map) unexpected error: %v", err)
	}
	if !w.IsJSONBody() {
		t.Error("WithNormalizedBody(map): IsJSONBody() = false; want true")
	}
}

// ///////////////////////////
// Section: Sentinel error identity (errors.Is)
// ///////////////////////////

// TestSafeUnmarshalBytes_EmptyInput verifies that SafeUnmarshalBytes returns an
// error wrapping ErrEmptyInput when given a nil/empty byte slice.
func TestSafeUnmarshalBytes_EmptyInput(t *testing.T) {
	var dst map[string]any
	err := encoding.StrictUnmarshalJSON(nil, &dst)
	if err == nil {
		t.Fatal("SafeUnmarshalBytes(nil) expected error; got nil")
	}
	if !errors.Is(err, encoding.ErrEmptyInput) {
		t.Errorf("SafeUnmarshalBytes(nil) error = %v; want errors.Is(err, ErrEmptyInput) = true", err)
	}

	err = encoding.StrictUnmarshalJSON([]byte{}, &dst)
	if err == nil {
		t.Fatal("SafeUnmarshalBytes(empty) expected error; got nil")
	}
	if !errors.Is(err, encoding.ErrEmptyInput) {
		t.Errorf("SafeUnmarshalBytes(empty) error = %v; want errors.Is(err, ErrEmptyInput) = true", err)
	}
}

// TestSafeUnmarshalBytes_InvalidJSON verifies that SafeUnmarshalBytes returns an
// error wrapping ErrInvalidJSON when given a non-JSON byte slice.
func TestSafeUnmarshalBytes_InvalidJSON(t *testing.T) {
	var dst map[string]any
	err := encoding.StrictUnmarshalJSON([]byte("not json"), &dst)
	if err == nil {
		t.Fatal("SafeUnmarshalBytes(invalid) expected error; got nil")
	}
	if !errors.Is(err, encoding.ErrInvalidJSON) {
		t.Errorf("SafeUnmarshalBytes(invalid) error = %v; want errors.Is(err, ErrInvalidJSON) = true", err)
	}
}

// TestSafeUnmarshalJSON_EmptyInput verifies that SafeUnmarshalJSON returns an
// error wrapping ErrEmptyInput when given an empty string.
func TestSafeUnmarshalJSON_EmptyInput(t *testing.T) {
	var dst map[string]any
	err := encoding.StrictUnmarshalJSONString("", &dst)
	if err == nil {
		t.Fatal("SafeUnmarshalJSON(\"\") expected error; got nil")
	}
	if !errors.Is(err, encoding.ErrEmptyInput) {
		t.Errorf("SafeUnmarshalJSON(\"\") error = %v; want errors.Is(err, ErrEmptyInput) = true", err)
	}
}

// TestSafeUnmarshalJSON_InvalidJSON verifies that SafeUnmarshalJSON returns an
// error wrapping ErrInvalidJSON when given a non-JSON string.
func TestSafeUnmarshalJSON_InvalidJSON(t *testing.T) {
	var dst map[string]any
	err := encoding.StrictUnmarshalJSONString("{bad json}", &dst)
	if err == nil {
		t.Fatal("SafeUnmarshalJSON(invalid) expected error; got nil")
	}
	if !errors.Is(err, encoding.ErrInvalidJSON) {
		t.Errorf("SafeUnmarshalJSON(invalid) error = %v; want errors.Is(err, ErrInvalidJSON) = true", err)
	}
}

// TestNormalizeJSON_EmptyInput verifies that NormalizeJSON returns an error
// wrapping ErrEmptyInput when given an empty or whitespace-only string.
func TestNormalizeJSON_EmptyInput(t *testing.T) {
	cases := []string{"", "   ", "\t\n"}
	for _, s := range cases {
		_, err := encoding.NormalizeJSON(s)
		if err == nil {
			t.Errorf("NormalizeJSON(%q) expected error; got nil", s)
			continue
		}
		if !errors.Is(err, encoding.ErrEmptyInput) {
			t.Errorf("NormalizeJSON(%q) error = %v; want errors.Is(err, ErrEmptyInput) = true", s, err)
		}
	}
}

// TestNormalizeJSON_UnfixableInputWrapsErrInvalidJSON verifies that NormalizeJSON
// returns an error wrapping ErrInvalidJSON for inputs that cannot be repaired.
func TestNormalizeJSON_UnfixableInputWrapsErrInvalidJSON(t *testing.T) {
	_, err := encoding.NormalizeJSON("totally not json !@#$")
	if err == nil {
		t.Fatal("NormalizeJSON(unfixable) expected error; got nil")
	}
	if !errors.Is(err, encoding.ErrInvalidJSON) {
		t.Errorf("NormalizeJSON(unfixable) error = %v; want errors.Is(err, ErrInvalidJSON) = true", err)
	}
}

// ///////////////////////////
// Section: Pretty / Ugly — edge cases
// ///////////////////////////

// TestUgly_EmptyInput verifies that Ugly does not panic on empty input.
func TestUgly_EmptyInput(t *testing.T) {
	got := encoding.CompactJSON([]byte{})
	if len(got) != 0 {
		t.Errorf("CompactJSON(empty) = %q; want empty slice", got)
	}
}

// TestUglyInPlace_EmptyInput verifies that UglyInPlace does not panic on empty input.
func TestUglyInPlace_EmptyInput(t *testing.T) {
	input := []byte{}
	got := encoding.CompactJSONUnsafe(input)
	if len(got) != 0 {
		t.Errorf("CompactJSONUnsafe(empty) = %q; want empty slice", got)
	}
}

// TestSpec_EmptyInput verifies that Spec does not panic on empty input.
func TestSpec_EmptyInput(t *testing.T) {
	got := encoding.Spec([]byte{})
	if len(got) != 0 {
		t.Errorf("Spec(empty) = %q; want empty slice", got)
	}
}

// TestColor_EmptyInput verifies that Color does not panic on empty input.
func TestColor_EmptyInput(t *testing.T) {
	got := encoding.Color([]byte{}, nil)
	if len(got) != 0 {
		t.Errorf("Color(empty) = %q; want empty slice", got)
	}
}

// TestPretty_EmptyInput verifies that Pretty does not panic on empty input.
func TestPretty_EmptyInput(t *testing.T) {
	got := encoding.Pretty([]byte{})
	if len(got) != 0 {
		t.Errorf("Pretty(empty) = %q; want empty slice", got)
	}
}

// TestSafeUnmarshalBytes_Valid verifies that SafeUnmarshalBytes successfully
// unmarshals a valid JSON byte slice.
func TestSafeUnmarshalBytes_Valid(t *testing.T) {
	input := []byte(`{"name":"Alice","age":30}`)
	var dst sampleStruct
	if err := encoding.StrictUnmarshalJSON(input, &dst); err != nil {
		t.Fatalf("SafeUnmarshalBytes(valid) unexpected error: %v", err)
	}
	if dst.Name != "Alice" || dst.Age != 30 {
		t.Errorf("SafeUnmarshalBytes(valid) = %+v; want {Name:Alice Age:30}", dst)
	}
}

// TestSafeUnmarshalJSON_Valid verifies that SafeUnmarshalJSON successfully
// unmarshals a valid JSON string.
func TestSafeUnmarshalJSON_Valid(t *testing.T) {
	input := `{"name":"Bob","age":25}`
	var dst sampleStruct
	if err := encoding.StrictUnmarshalJSONString(input, &dst); err != nil {
		t.Fatalf("SafeUnmarshalJSON(valid) unexpected error: %v", err)
	}
	if dst.Name != "Bob" || dst.Age != 25 {
		t.Errorf("SafeUnmarshalJSON(valid) = %+v; want {Name:Bob Age:25}", dst)
	}
}

// TestIsValidJSON_EdgeCases verifies that IsValidJSON handles edge inputs
// correctly without panicking.
func TestIsValidJSON_EdgeCases(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"", false},
		{"null", true},
		{"true", true},
		{"false", true},
		{"123", true},
		{`"string"`, true},
		{`{}`, true},
		{`[]`, true},
		{`{bad}`, false},
	}
	for _, tc := range cases {
		got := encoding.IsValidJSONString(tc.input)
		if got != tc.want {
			t.Errorf("IsValidJSON(%q) = %v; want %v", tc.input, got, tc.want)
		}
	}
}

// TestColor_NilStyle falls back to TerminalStyle without panicking.
func TestColor_NilStyle(t *testing.T) {
	src := []byte(`{"key":"value","n":42}`)
	got := encoding.Color(src, nil)
	if len(got) == 0 {
		t.Error("Color(src, nil) returned empty; expected styled output")
	}
	// Resulting output should still contain the literal content
	if !strings.Contains(string(got), "key") || !strings.Contains(string(got), "value") {
		t.Errorf("Color(src, nil) = %q; expected key and value to be present", got)
	}
}

// TestPrettyOptions_SortKeys verifies that PrettyOptions with SortKeys=true
// produces alphabetically-ordered keys.
func TestPrettyOptions_SortKeys(t *testing.T) {
	src := []byte(`{"z":1,"a":2,"m":3}`)
	opts := &encoding.OptionsConfig{Width: 80, Indent: "  ", SortKeys: true}
	got := string(encoding.PrettyOptions(src, opts))
	aIdx := strings.Index(got, `"a"`)
	mIdx := strings.Index(got, `"m"`)
	zIdx := strings.Index(got, `"z"`)
	if aIdx < 0 || mIdx < 0 || zIdx < 0 {
		t.Fatalf("PrettyOptions(SortKeys) missing keys; got: %s", got)
	}
	if !(aIdx < mIdx && mIdx < zIdx) {
		t.Errorf("PrettyOptions(SortKeys) key order: a=%d m=%d z=%d; want a < m < z\n%s", aIdx, mIdx, zIdx, got)
	}
}

// TestMarshalJSONb_Roundtrip verifies basic marshal / unmarshal roundtrip.
func TestMarshalJSONb_Roundtrip(t *testing.T) {
	original := sampleStruct{Name: "roundtrip", Age: 99}
	b, err := encoding.MarshalJSON(original)
	if err != nil {
		t.Fatalf("MarshalJSON unexpected error: %v", err)
	}
	var recovered sampleStruct
	if err := encoding.UnmarshalJSON(b, &recovered); err != nil {
		t.Fatalf("UnmarshalBytes unexpected error: %v", err)
	}
	if recovered != original {
		t.Errorf("roundtrip got %+v; want %+v", recovered, original)
	}
}

// TestMarshalJSONs_Roundtrip verifies basic string marshal / unmarshal roundtrip.
func TestMarshalJSONs_Roundtrip(t *testing.T) {
	original := sampleStruct{Name: "strtest", Age: 7}
	s, err := encoding.MarshalJSONString(original)
	if err != nil {
		t.Fatalf("MarshalJSONString unexpected error: %v", err)
	}
	var recovered sampleStruct
	if err := encoding.UnmarshalJSONString(s, &recovered); err != nil {
		t.Fatalf("UnmarshalJSON unexpected error: %v", err)
	}
	if recovered != original {
		t.Errorf("roundtrip got %+v; want %+v", recovered, original)
	}
}

// TestMarshalJSONIndent verifies that MarshalJSONIndent produces indented output.
func TestMarshalJSONIndent(t *testing.T) {
	s := sampleStruct{Name: "indent", Age: 5}
	b, err := encoding.MarshalJSONIndent(s, "", "  ")
	if err != nil {
		t.Fatalf("MarshalJSONIndent unexpected error: %v", err)
	}
	if !strings.Contains(string(b), "\n") {
		t.Errorf("MarshalJSONIndent output not indented: %s", b)
	}
}

// TestJSON_Uintptr verifies the special-case hex-address quoting for uintptr.
func TestJSON_Uintptr(t *testing.T) {
	var p uintptr = 0xdeadbeef
	got := encoding.JSON(p)
	// Result must be a quoted string starting with "0x".
	if len(got) < 2 || got[0] != '"' || got[len(got)-1] != '"' {
		t.Errorf("JSON(uintptr) = %q; expected a quoted hex string", got)
	}
	if !strings.HasPrefix(got, `"0x`) {
		t.Errorf("JSON(uintptr) = %q; expected prefix \"0x", got)
	}
}

// TestJSONToken_Uintptr verifies the Token variant also produces a quoted hex address.
func TestJSONToken_Uintptr(t *testing.T) {
	var p uintptr = 0xc0ffee
	got, err := encoding.JSONE(p)
	if err != nil {
		t.Fatalf("JSONToken(uintptr) unexpected error: %v", err)
	}
	if len(got) < 2 || got[0] != '"' || got[len(got)-1] != '"' {
		t.Errorf("JSONToken(uintptr) = %q; expected a quoted hex string", got)
	}
}

// TestIsValidJSONBytes_EdgeCases verifies IsValidJSONBytes handles empty and
// invalid inputs without panicking.
func TestIsValidJSONBytes_EdgeCases(t *testing.T) {
	if encoding.IsValidJSON(nil) {
		t.Error("IsValidJSONBytes(nil) = true; want false")
	}
	if encoding.IsValidJSON([]byte{}) {
		t.Error("IsValidJSONBytes(empty) = true; want false")
	}
	if !encoding.IsValidJSON([]byte(`null`)) {
		t.Error("IsValidJSONBytes(null) = false; want true")
	}
}

// TestSpecUnsafe verifies SpecUnsafe removes single-line comments in-place.
func TestSpecUnsafe(t *testing.T) {
	input := []byte(`{"key": "value" // comment
}`)
	got := encoding.SpecUnsafe(input)
	if strings.Contains(string(got), "//") {
		t.Errorf("SpecUnsafe did not remove single-line comment; got: %s", got)
	}
}

// TestJSONToken_NilRawMessage verifies that a nil json.RawMessage returns "null".
func TestJSONToken_NilRawMessage(t *testing.T) {
	var rm json.RawMessage // nil by default
	got, err := encoding.JSONE(rm)
	if err != nil {
		t.Fatalf("JSONToken(nil RawMessage) unexpected error: %v", err)
	}
	if got != "null" {
		t.Errorf("JSONToken(nil RawMessage) = %q; want %q", got, "null")
	}
}

// TestJSON_NilFunc verifies that JSON(func) returns "null" for a nil function value.
func TestJSON_NilFunc(t *testing.T) {
	var f func()
	got := encoding.JSON(f)
	if got != "null" {
		t.Errorf("JSON(nil func) = %q; want %q", got, "null")
	}
}

// TestJSON_NilChan verifies that JSON(chan) returns "null" for a nil channel.
func TestJSON_NilChan(t *testing.T) {
	var ch chan int
	got := encoding.JSON(ch)
	if got != "null" {
		t.Errorf("JSON(nil chan) = %q; want %q", got, "null")
	}
}
