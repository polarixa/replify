package encoding_test

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/polarixa/replify/pkg/encoding"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test helpers & fixture types
// ─────────────────────────────────────────────────────────────────────────────

// jsonMarshalerImpl is a value-receiver json.Marshaler.
type jsonMarshalerImpl struct{ V string }

func (j jsonMarshalerImpl) MarshalJSON() ([]byte, error) {
	return []byte(`"custom:` + j.V + `"`), nil
}

// jsonMarshalerPtrImpl is a pointer-receiver json.Marshaler.
type jsonMarshalerPtrImpl struct{ V string }

func (j *jsonMarshalerPtrImpl) MarshalJSON() ([]byte, error) {
	return []byte(`"ptr:` + j.V + `"`), nil
}

// textMarshalerImpl is a value-receiver encoding.TextMarshaler.
type textMarshalerImpl struct{ V string }

func (t textMarshalerImpl) MarshalText() ([]byte, error) {
	return []byte("text:" + t.V), nil
}

// textMarshalerPtrImpl is a pointer-receiver encoding.TextMarshaler.
type textMarshalerPtrImpl struct{ V string }

func (t *textMarshalerPtrImpl) MarshalText() ([]byte, error) {
	return []byte("ptext:" + t.V), nil
}

// panicMarshalerImpl is a json.Marshaler that panics unconditionally.
type panicMarshalerImpl struct{}

func (p panicMarshalerImpl) MarshalJSON() ([]byte, error) {
	panic("deliberate panic in marshaler")
}

// ─── Struct fixtures ──────────────────────────────────────────────────────────

type flatStruct struct {
	Name string
	Age  int
}

type taggedStruct struct {
	ID    int    `json:"id"`
	Email string `json:"email,omitempty"`
	Pass  string `json:"-"`
}

type omitStruct struct {
	A string `json:"a,omitempty"`
	B int    `json:"b,omitempty"`
	C bool   `json:"c,omitempty"`
	D []int  `json:"d,omitempty"`
}

type innerStruct struct {
	X int
	Y int
}

type embeddedStruct struct {
	innerStruct
	Z int
}

type embeddedNamedStruct struct {
	innerStruct `json:"inner"`
	Z           int
}

type embeddedPtrStruct struct {
	*innerStruct
	Z int
}

type embeddedNilPtrStruct struct {
	*innerStruct
	Z int
}

type complexStruct struct {
	C complex128
	R float64
}

type nestedStruct struct {
	Child flatStruct
	Score float64
}

// ─── Map key fixtures ─────────────────────────────────────────────────────────

type textKeyType struct{ V string }

func (t textKeyType) MarshalText() ([]byte, error) { return []byte(t.V), nil }

// ─────────────────────────────────────────────────────────────────────────────
// JSON() – public API
// ─────────────────────────────────────────────────────────────────────────────

func TestJSON_Nil(t *testing.T) {
	if got := encoding.JSON(nil); got != "" {
		t.Fatalf("JSON(nil) = %q; want %q", got, "")
	}
}

func TestJSON_NilPointer(t *testing.T) {
	var p *int
	if got := encoding.JSON(p); got != "null" {
		t.Fatalf("JSON(nil *int) = %q; want \"null\"", got)
	}
}

func TestJSON_NilSlice(t *testing.T) {
	var s []int
	if got := encoding.JSON(s); got != "null" {
		t.Fatalf("JSON(nil []int) = %q; want \"null\"", got)
	}
}

func TestJSON_NilMap(t *testing.T) {
	var m map[string]int
	if got := encoding.JSON(m); got != "null" {
		t.Fatalf("JSON(nil map) = %q; want \"null\"", got)
	}
}

func TestJSON_NilInterface(t *testing.T) {
	var i any
	if got := encoding.JSON(i); got != "" {
		t.Fatalf("JSON(nil any) = %q; want %q", got, "")
	}
}

func TestJSON_String(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"hello", `"hello"`},
		{"", `""`},
		{"with\nnewline", `"with\nnewline"`},
		{"tab\there", `"tab\there"`},
		{"unicode: \u4e2d", `"unicode: 中"`},
	}
	for _, tc := range cases {
		if got := encoding.JSON(tc.in); got != tc.want {
			t.Errorf("JSON(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}

func TestJSON_StringPointer(t *testing.T) {
	s := "hello"
	if got := encoding.JSON(&s); got != `"hello"` {
		t.Fatalf("JSON(*string) = %q; want %q", got, `"hello"`)
	}
}

func TestJSON_NilStringPointer(t *testing.T) {
	var p *string
	if got := encoding.JSON(p); got != "null" {
		t.Fatalf("JSON(nil *string) = %q; want \"null\"", got)
	}
}

func TestJSON_Bool(t *testing.T) {
	if got := encoding.JSON(true); got != "true" {
		t.Fatalf("JSON(true) = %q; want \"true\"", got)
	}
	if got := encoding.JSON(false); got != "false" {
		t.Fatalf("JSON(false) = %q; want \"false\"", got)
	}
}

func TestJSON_Integers(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{int(0), "0"},
		{int(-1), "-1"},
		{int8(127), "127"},
		{int16(-32768), "-32768"},
		{int32(2147483647), "2147483647"},
		{int64(-9223372036854775808), "-9223372036854775808"},
		{uint(0), "0"},
		{uint8(255), "255"},
		{uint16(65535), "65535"},
		{uint32(4294967295), "4294967295"},
		{uint64(18446744073709551615), "18446744073709551615"},
	}
	for _, tc := range cases {
		if got := encoding.JSON(tc.in); got != tc.want {
			t.Errorf("JSON(%v) = %q; want %q", tc.in, got, tc.want)
		}
	}
}

func TestJSON_Uintptr(t *testing.T) {
	got := encoding.JSON(uintptr(0xff))
	if !strings.HasPrefix(got, `"0x`) {
		t.Fatalf("JSON(uintptr) = %q; want quoted hex string", got)
	}
}

func TestJSON_Floats(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{float32(3.14), "3.14"},
		{float64(3.14), "3.14"},
		{float64(0), "0"},
		{float64(-1.5), "-1.5"},
		{float64(1e100), "1e+100"},
	}
	for _, tc := range cases {
		if got := encoding.JSON(tc.in); got != tc.want {
			t.Errorf("JSON(%v) = %q; want %q", tc.in, got, tc.want)
		}
	}
}

func TestJSON_NonFiniteFloat_Error(t *testing.T) {
	if got := encoding.JSON(math.NaN()); got != "null" {
		t.Fatalf("JSON(NaN) = %q; want null", got)
	}
	if got := encoding.JSON(math.Inf(1)); got != "null" {
		t.Fatalf("JSON(+Inf) = %q; want null", got)
	}
	if got := encoding.JSON(math.Inf(-1)); got != "null" {
		t.Fatalf("JSON(-Inf) = %q; want null", got)
	}
}

func TestJSON_Complex64(t *testing.T) {
	got := encoding.JSON(complex64(1 + 2i))
	want := `{"real":1,"imag":2}`
	if got != want {
		t.Fatalf("JSON(complex64) = %q; want %q", got, want)
	}
}

func TestJSON_Complex128(t *testing.T) {
	got := encoding.JSON(complex(3.5, -1.25))
	want := `{"real":3.5,"imag":-1.25}`
	if got != want {
		t.Fatalf("JSON(complex128) = %q; want %q", got, want)
	}
}

func TestJSON_Complex128_Zero(t *testing.T) {
	got := encoding.JSON(complex128(0))
	want := `{"real":0,"imag":0}`
	if got != want {
		t.Fatalf("JSON(complex128(0)) = %q; want %q", got, want)
	}
}

func TestJSON_RawMessage(t *testing.T) {
	raw := json.RawMessage(`{"a":1}`)
	if got := encoding.JSON(raw); got != `{"a":1}` {
		t.Fatalf("JSON(RawMessage) = %q", got)
	}
}

func TestJSON_RawMessagePtr(t *testing.T) {
	raw := json.RawMessage(`[1,2]`)
	if got := encoding.JSON(&raw); got != `[1,2]` {
		t.Fatalf("JSON(*RawMessage) = %q", got)
	}
}

func TestJSON_NilRawMessagePtr(t *testing.T) {
	var raw *json.RawMessage
	if got := encoding.JSON(raw); got != "null" {
		t.Fatalf("JSON(nil *RawMessage) = %q; want \"null\"", got)
	}
}

func TestJSON_InvalidRawMessage(t *testing.T) {
	raw := json.RawMessage(`{bad json}`)
	if got := encoding.JSON(raw); got != "" {
		t.Fatalf("JSON(invalid RawMessage) = %q; want empty string", got)
	}
}

func TestJSON_FlatStruct(t *testing.T) {
	v := flatStruct{Name: "Alice", Age: 30}
	got := encoding.JSON(v)
	want := `{"Name":"Alice","Age":30}`
	if got != want {
		t.Fatalf("JSON(flatStruct) = %q; want %q", got, want)
	}
}

func TestJSON_TaggedStruct_Full(t *testing.T) {
	v := taggedStruct{ID: 1, Email: "a@b.com", Pass: "secret"}
	got := encoding.JSON(v)
	if strings.Contains(got, "secret") {
		t.Fatal("tagged '-' field must not appear in output")
	}
	if !strings.Contains(got, `"email":"a@b.com"`) {
		t.Fatalf("email field missing in %q", got)
	}
	if !strings.Contains(got, `"id":1`) {
		t.Fatalf("id field missing in %q", got)
	}
}

func TestJSON_TaggedStruct_Omitempty(t *testing.T) {
	v := taggedStruct{ID: 2, Email: "", Pass: "x"}
	got := encoding.JSON(v)
	if strings.Contains(got, "email") {
		t.Fatalf("empty omitempty field must not appear in %q", got)
	}
}

func TestJSON_OmitemptyAllTypes(t *testing.T) {
	v := omitStruct{}
	got := encoding.JSON(v)
	if got != "{}" {
		t.Fatalf("all-zero omitempty struct = %q; want \"{}\"", got)
	}

	v2 := omitStruct{A: "x", B: 1, C: true, D: []int{1}}
	got2 := encoding.JSON(v2)
	for _, key := range []string{`"a"`, `"b"`, `"c"`, `"d"`} {
		if !strings.Contains(got2, key) {
			t.Errorf("non-zero omitempty field %s missing in %q", key, got2)
		}
	}
}

func TestJSON_EmbeddedStruct_Promoted(t *testing.T) {
	v := embeddedStruct{innerStruct: innerStruct{X: 1, Y: 2}, Z: 3}
	got := encoding.JSON(v)
	if !strings.Contains(got, `"X":1`) || !strings.Contains(got, `"Y":2`) || !strings.Contains(got, `"Z":3`) {
		t.Fatalf("embedded fields not promoted in %q", got)
	}
	if strings.Contains(got, "innerStruct") {
		t.Fatalf("embedded type name must not appear in %q", got)
	}
}

func TestJSON_EmbeddedStruct_Named(t *testing.T) {
	v := embeddedNamedStruct{innerStruct: innerStruct{X: 9, Y: 8}, Z: 7}
	got := encoding.JSON(v)
	if !strings.Contains(got, `"inner"`) {
		t.Fatalf("explicitly named embedded field missing in %q", got)
	}
	if !strings.Contains(got, `"Z":7`) {
		t.Fatalf("Z field missing in %q", got)
	}
}

func TestJSON_EmbeddedStruct_PtrNonNil(t *testing.T) {
	v := embeddedPtrStruct{innerStruct: &innerStruct{X: 5, Y: 6}, Z: 7}
	got := encoding.JSON(v)
	if !strings.Contains(got, `"X":5`) || !strings.Contains(got, `"Y":6`) {
		t.Fatalf("embedded ptr struct fields not promoted in %q", got)
	}
}

func TestJSON_EmbeddedStruct_PtrNil(t *testing.T) {
	v := embeddedNilPtrStruct{innerStruct: nil, Z: 3}
	got := encoding.JSON(v)
	if !strings.Contains(got, `"Z":3`) {
		t.Fatalf("Z field missing in %q", got)
	}
	if strings.Contains(got, "X") || strings.Contains(got, "Y") {
		t.Fatalf("nil embedded ptr fields must not appear in %q", got)
	}
}

func TestJSON_NestedStruct(t *testing.T) {
	v := nestedStruct{Child: flatStruct{Name: "Bob", Age: 5}, Score: 9.9}
	got := encoding.JSON(v)
	if !strings.Contains(got, `"Child"`) || !strings.Contains(got, `"Name":"Bob"`) {
		t.Fatalf("nested struct not encoded correctly: %q", got)
	}
}

func TestJSON_ComplexStruct(t *testing.T) {
	v := complexStruct{C: complex(1, 2), R: 3.14}
	got := encoding.JSON(v)
	if !strings.Contains(got, `"real":1`) || !strings.Contains(got, `"imag":2`) {
		t.Fatalf("complex field in struct not encoded correctly: %q", got)
	}
	if !strings.Contains(got, `"R":3.14`) {
		t.Fatalf("R field missing in %q", got)
	}
}

func TestJSON_Slice(t *testing.T) {
	got := encoding.JSON([]int{1, 2, 3})
	if got != "[1,2,3]" {
		t.Fatalf("JSON([]int) = %q; want \"[1,2,3]\"", got)
	}
}

func TestJSON_EmptySlice(t *testing.T) {
	got := encoding.JSON([]int{})
	if got != "[]" {
		t.Fatalf("JSON([]int{}) = %q; want \"[]\"", got)
	}
}

func TestJSON_SliceOfStrings(t *testing.T) {
	got := encoding.JSON([]string{"a", "b"})
	if got != `["a","b"]` {
		t.Fatalf("JSON([]string) = %q", got)
	}
}

func TestJSON_Array(t *testing.T) {
	got := encoding.JSON([3]int{4, 5, 6})
	if got != "[4,5,6]" {
		t.Fatalf("JSON([3]int) = %q; want \"[4,5,6]\"", got)
	}
}

func TestJSON_Map_StringKey(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1}
	got := encoding.JSON(m)
	if got != `{"a":1,"b":2}` {
		t.Fatalf("JSON(map[string]int) = %q; want sorted keys", got)
	}
}

func TestJSON_Map_IntKey(t *testing.T) {
	m := map[int]string{2: "two", 1: "one"}
	got := encoding.JSON(m)
	if got != `{"1":"one","2":"two"}` {
		t.Fatalf("JSON(map[int]string) = %q", got)
	}
}

func TestJSON_Map_BoolKey(t *testing.T) {
	m := map[bool]int{true: 1, false: 0}
	got := encoding.JSON(m)
	if got != `{"false":0,"true":1}` {
		t.Fatalf("JSON(map[bool]int) = %q; want sorted bool keys", got)
	}
}

func TestJSON_Map_Float32Key(t *testing.T) {
	m := map[float32]string{1.5: "x"}
	got := encoding.JSON(m)
	if !strings.Contains(got, `"1.5"`) {
		t.Fatalf("JSON(map[float32]string) = %q; want float32 key", got)
	}
}

func TestJSON_Map_Float64Key(t *testing.T) {
	m := map[float64]string{3.14: "pi"}
	got := encoding.JSON(m)
	if !strings.Contains(got, `"3.14"`) {
		t.Fatalf("JSON(map[float64]string) = %q; want float64 key", got)
	}
}

func TestJSON_Map_TextMarshalerKey(t *testing.T) {
	m := map[textKeyType]int{{V: "hello"}: 42}
	got := encoding.JSON(m)
	if got != `{"hello":42}` {
		t.Fatalf("JSON(map[textKeyType]int) = %q; want {\"hello\":42}", got)
	}
}

func TestJSON_JSONMarshalerValue(t *testing.T) {
	v := jsonMarshalerImpl{V: "test"}
	got := encoding.JSON(v)
	if got != `"custom:test"` {
		t.Fatalf("JSON(jsonMarshalerImpl) = %q; want %q", got, `"custom:test"`)
	}
}

func TestJSON_JSONMarshalerPtr(t *testing.T) {
	v := &jsonMarshalerPtrImpl{V: "ptr"}
	got := encoding.JSON(v)
	if got != `"ptr:ptr"` {
		t.Fatalf("JSON(*jsonMarshalerPtrImpl) = %q; want %q", got, `"ptr:ptr"`)
	}
}

func TestJSON_TextMarshalerValue(t *testing.T) {
	v := textMarshalerImpl{V: "hello"}
	got := encoding.JSON(v)
	if got != `"text:hello"` {
		t.Fatalf("JSON(textMarshalerImpl) = %q; want %q", got, `"text:hello"`)
	}
}

func TestJSON_TextMarshalerPtr(t *testing.T) {
	v := &textMarshalerPtrImpl{V: "world"}
	got := encoding.JSON(v)
	if got != `"ptext:world"` {
		t.Fatalf("JSON(*textMarshalerPtrImpl) = %q; want %q", got, `"ptext:world"`)
	}
}

func TestJSON_PanicMarshaler(t *testing.T) {
	got := encoding.JSON(panicMarshalerImpl{})
	if got != "" {
		t.Fatalf("JSON(panicMarshalerImpl) = %q; want empty string on recovered panic", got)
	}
}

func TestJSON_Pretty_Map(t *testing.T) {
	m := map[string]int{"a": 1}
	got := encoding.JSON(m, true)
	if !strings.Contains(got, "\n") || !strings.Contains(got, "    ") {
		t.Fatalf("JSON(map, pretty=true) = %q; want indented output", got)
	}
}

func TestJSON_Pretty_Struct(t *testing.T) {
	v := flatStruct{Name: "X", Age: 1}
	got := encoding.JSON(v, true)
	if !strings.Contains(got, "\n") {
		t.Fatalf("JSON(struct, pretty=true) = %q; want indented output", got)
	}
}

func TestJSON_Pretty_RawMessage(t *testing.T) {
	raw := json.RawMessage(`{"x":1}`)
	got := encoding.JSON(raw, true)
	if !strings.Contains(got, "\n") {
		t.Fatalf("JSON(RawMessage, pretty=true) = %q; want indented output", got)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// JSONE() – public API with error returns
// ─────────────────────────────────────────────────────────────────────────────

func TestJSONE_Nil_ReturnsError(t *testing.T) {
	_, err := encoding.JSONE(nil)
	if !errors.Is(err, encoding.ErrNilInterface) {
		t.Fatalf("JSONE(nil) error = %v; want ErrNilInterface", err)
	}
}

func TestJSONE_NilInterface_ReturnsError(t *testing.T) {
	var i any
	_, err := encoding.JSONE(i)
	if !errors.Is(err, encoding.ErrNilInterface) {
		t.Fatalf("JSONE(nil any) error = %v; want ErrNilInterface", err)
	}
}

func TestJSONE_NilPointer_ReturnsNull(t *testing.T) {
	var p *int
	s, err := encoding.JSONE(p)
	if err != nil {
		t.Fatalf("JSONE(nil *int) unexpected error: %v", err)
	}
	if s != "null" {
		t.Fatalf("JSONE(nil *int) = %q; want \"null\"", s)
	}
}

func TestJSONE_NilMap_ReturnsNull(t *testing.T) {
	var m map[string]int
	s, err := encoding.JSONE(m)
	if err != nil {
		t.Fatalf("JSONE(nil map) unexpected error: %v", err)
	}
	if s != "null" {
		t.Fatalf("JSONE(nil map) = %q; want \"null\"", s)
	}
}

func TestJSONE_ValidValue(t *testing.T) {
	s, err := encoding.JSONE(42)
	if err != nil {
		t.Fatalf("JSONE(42) unexpected error: %v", err)
	}
	if s != "42" {
		t.Fatalf("JSONE(42) = %q; want \"42\"", s)
	}
}

func TestJSONE_NonFiniteFloat(t *testing.T) {
	_, err := encoding.JSONE(math.NaN())
	if !errors.Is(err, encoding.ErrNonFiniteFloat) {
		t.Fatalf("JSONE(NaN) error = %v; want ErrNonFiniteFloat", err)
	}
	_, err = encoding.JSONE(math.Inf(1))
	if !errors.Is(err, encoding.ErrNonFiniteFloat) {
		t.Fatalf("JSONE(+Inf) error = %v; want ErrNonFiniteFloat", err)
	}
	_, err = encoding.JSONE(math.Inf(-1))
	if !errors.Is(err, encoding.ErrNonFiniteFloat) {
		t.Fatalf("JSONE(-Inf) error = %v; want ErrNonFiniteFloat", err)
	}
}

func TestJSONE_InvalidRawMessage(t *testing.T) {
	_, err := encoding.JSONE(json.RawMessage(`not-json`))
	if !errors.Is(err, encoding.ErrInvalidRawMessage) {
		t.Fatalf("JSONE(invalid RawMessage) error = %v; want ErrInvalidRawMessage", err)
	}
}

func TestJSONE_PanicMarshaler(t *testing.T) {
	_, err := encoding.JSONE(panicMarshalerImpl{})
	if !errors.Is(err, encoding.ErrMarshalPanicRecovered) {
		t.Fatalf("JSONE(panicMarshalerImpl) error = %v; want ErrMarshalPanicRecovered", err)
	}
}

func TestJSONE_Pretty(t *testing.T) {
	s, err := encoding.JSONE(map[string]int{"k": 1}, true)
	if err != nil {
		t.Fatalf("JSONE(map, pretty) unexpected error: %v", err)
	}
	if !strings.Contains(s, "\n") {
		t.Fatalf("JSONE(map, pretty) = %q; want indented output", s)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Round-trip: JSON output is valid JSON
// ─────────────────────────────────────────────────────────────────────────────

func TestJSON_OutputIsValidJSON(t *testing.T) {
	values := []any{
		42,
		"hello",
		true,
		false,
		3.14,
		[]int{1, 2, 3},
		map[string]int{"a": 1},
		flatStruct{Name: "Z", Age: 9},
		taggedStruct{ID: 5, Email: "e@e.com", Pass: "p"},
		embeddedStruct{innerStruct: innerStruct{X: 1, Y: 2}, Z: 3},
		nestedStruct{Child: flatStruct{Name: "N", Age: 1}, Score: 0.5},
		complexStruct{C: complex(1, 2), R: 1.0},
		jsonMarshalerImpl{V: "v"},
		textMarshalerImpl{V: "t"},
		(*int)(nil),
		[]int(nil),
	}
	for _, v := range values {
		s := encoding.JSON(v)
		if s == "" {
			continue
		}
		if !json.Valid([]byte(s)) {
			t.Errorf("JSON(%T) produced invalid JSON: %s", v, s)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Determinism: same input → same output across multiple calls
// ─────────────────────────────────────────────────────────────────────────────

func TestJSON_Deterministic(t *testing.T) {
	m := map[string]int{"c": 3, "a": 1, "b": 2}
	first := encoding.JSON(m)
	for i := 0; i < 20; i++ {
		if got := encoding.JSON(m); got != first {
			t.Fatalf("JSON(map) is non-deterministic: %q vs %q", first, got)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Reflect-level helpers — exercised indirectly via the public API.
// Direct calls to unexported helpers (mapKeyString, scalarJSONToken, etc.) are
// not available from an external test package; the tests below cover the same
// behaviour through JSON / JSONE.
// ─────────────────────────────────────────────────────────────────────────────

// TestJSON_Map_UnsupportedKey verifies that a map with an unsupported key type
// produces an empty string (error path) rather than panicking.
func TestJSON_Map_UnsupportedKey(t *testing.T) {
	// Construct a map[struct]int via reflection so the compiler does not reject it.
	type badKey struct{ v int } // no TextMarshaler
	mt := reflect.MapOf(reflect.TypeOf(badKey{}), reflect.TypeOf(0))
	mv := reflect.MakeMap(mt)
	mv.SetMapIndex(reflect.ValueOf(badKey{v: 1}), reflect.ValueOf(42))

	got := encoding.JSON(mv.Interface())
	if got != "" {
		t.Fatalf("JSON(map[badKey]int) = %q; want empty string for unsupported key type", got)
	}
}

// TestJSON_ScalarTypes bundles quick checks for every scalar kind so that
// coverage of scalarJSONToken is maintained without access to the unexported
// function.
func TestJSON_ScalarTypes(t *testing.T) {
	cases := []struct {
		label string
		in    any
		want  string
	}{
		{"int8", int8(-1), "-1"},
		{"int16", int16(32767), "32767"},
		{"int32", int32(-100), "-100"},
		{"uint8", uint8(200), "200"},
		{"uint16", uint16(60000), "60000"},
		{"uint32", uint32(4000000000), "4000000000"},
		{"float32 zero", float32(0), "0"},
		{"float64 large", float64(1e308), "1e+308"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
	}
	for _, tc := range cases {
		if got := encoding.JSON(tc.in); got != tc.want {
			t.Errorf("[%s] JSON(%v) = %q; want %q", tc.label, tc.in, got, tc.want)
		}
	}
}

// TestJSON_DeepNesting verifies that deeply nested structs and slices are
// handled without stack overflow.
func TestJSON_DeepNesting(t *testing.T) {
	type Node struct {
		V    int
		Next *Node
	}
	head := &Node{V: 1, Next: &Node{V: 2, Next: &Node{V: 3}}}
	got := encoding.JSON(head)
	if !json.Valid([]byte(got)) {
		t.Fatalf("JSON(deep nested) produced invalid JSON: %s", got)
	}
	if !strings.Contains(got, `"V":3`) {
		t.Fatalf("innermost node missing in %q", got)
	}
}

// TestJSON_SliceOfStructs verifies slices whose elements are structs with tags.
func TestJSON_SliceOfStructs(t *testing.T) {
	items := []taggedStruct{
		{ID: 1, Email: "a@a.com", Pass: "p1"},
		{ID: 2, Email: "", Pass: "p2"},
	}
	got := encoding.JSON(items)
	if !json.Valid([]byte(got)) {
		t.Fatalf("JSON([]taggedStruct) invalid JSON: %s", got)
	}
	if strings.Contains(got, "p1") || strings.Contains(got, "p2") {
		t.Fatalf("'-' tagged field leaked into output: %s", got)
	}
}

// TestJSON_MapOfStructValues verifies that struct values inside a map are
// encoded correctly.
func TestJSON_MapOfStructValues(t *testing.T) {
	m := map[string]flatStruct{
		"b": {Name: "B", Age: 2},
		"a": {Name: "A", Age: 1},
	}
	got := encoding.JSON(m)
	if !json.Valid([]byte(got)) {
		t.Fatalf("JSON(map[string]flatStruct) invalid JSON: %s", got)
	}
	// Keys must be sorted.
	aIdx := strings.Index(got, `"a"`)
	bIdx := strings.Index(got, `"b"`)
	if aIdx > bIdx {
		t.Fatalf("map keys not sorted in %q", got)
	}
}

// TestJSONE_Pretty_Complex verifies pretty-printing for a complex-containing struct.
func TestJSONE_Pretty_Complex(t *testing.T) {
	v := complexStruct{C: complex(2, -1), R: 0.5}
	s, err := encoding.JSONE(v, true)
	if err != nil {
		t.Fatalf("JSONE(complexStruct, pretty) unexpected error: %v", err)
	}
	if !json.Valid([]byte(s)) {
		t.Fatalf("JSONE(complexStruct, pretty) invalid JSON: %s", s)
	}
	if !strings.Contains(s, "\n") {
		t.Fatalf("JSONE(complexStruct, pretty) not indented: %s", s)
	}
}

// TestJSON_EmptyMap verifies that an empty (non-nil) map encodes to "{}".
func TestJSON_EmptyMap(t *testing.T) {
	got := encoding.JSON(map[string]int{})
	if got != "{}" {
		t.Fatalf("JSON(empty map) = %q; want \"{}\"", got)
	}
}

// TestJSON_PointerToStruct verifies that a non-nil pointer to a struct is
// transparently dereferenced.
func TestJSON_PointerToStruct(t *testing.T) {
	v := &flatStruct{Name: "ptr", Age: 7}
	got := encoding.JSON(v)
	want := `{"Name":"ptr","Age":7}`
	if got != want {
		t.Fatalf("JSON(*flatStruct) = %q; want %q", got, want)
	}
}

// TestJSON_DoublePointer verifies double-pointer dereferencing.
func TestJSON_DoublePointer(t *testing.T) {
	n := 42
	p := &n
	got := encoding.JSON(&p)
	if got != "42" {
		t.Fatalf("JSON(**int) = %q; want \"42\"", got)
	}
}

// TestJSON_SliceOfInterfaces verifies that a []any slice encodes each element
// with the correct dynamic type.
func TestJSON_SliceOfInterfaces(t *testing.T) {
	s := []any{1, "two", true, nil}
	got := encoding.JSON(s)
	// nil inside a slice should encode as JSON null.
	if !strings.Contains(got, "null") {
		t.Fatalf("nil element not encoded as null in %q", got)
	}
	if !json.Valid([]byte(got)) {
		t.Fatalf("JSON([]any) invalid JSON: %s", got)
	}
}
