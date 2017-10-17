// Copyright 2016 Richard Gibson. All rights reserved.
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package canonicaljson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"
	"unicode"
)

type Optionals struct {
	Sr string `json:"sr"`
	So string `json:"so,omitempty"`
	Sw string `json:"-"`

	Ir int `json:"omitempty"` // actually named omitempty, not an option
	Io int `json:"io,omitempty"`

	Slr []string `json:"slr,random"`
	Slo []string `json:"slo,omitempty"`

	Mr map[string]interface{} `json:"mr"`
	Mo map[string]interface{} `json:",omitempty"`

	Fr float64 `json:"fr"`
	Fo float64 `json:"fo,omitempty"`

	Br bool `json:"br"`
	Bo bool `json:"bo,omitempty"`

	Ur uint `json:"ur"`
	Uo uint `json:"uo,omitempty"`

	Str struct{} `json:"str"`
	Sto struct{} `json:"sto,omitempty"`
}

var optionalsExpected = `{
 "br": false,
 "fr": 0,
 "mr": {},
 "omitempty": 0,
 "slr": null,
 "sr": "",
 "sto": {},
 "str": {},
 "ur": 0
}`

func TestOmitEmpty(t *testing.T) {
	var o Optionals
	o.Sw = "something"
	o.Mr = map[string]interface{}{}
	o.Mo = map[string]interface{}{}

	got, err := MarshalIndent(&o, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	if got := string(got); got != optionalsExpected {
		t.Errorf(" got: %s\nwant: %s\n", got, optionalsExpected)
	}
}

type StringTag struct {
	BoolStr bool   `json:",string"`
	IntStr  int64  `json:",string"`
	StrStr  string `json:",string"`
}

var stringTagExpected = `{
 "BoolStr": "true",
 "IntStr": "42",
 "StrStr": "\"xzbit\""
}`

func TestStringTag(t *testing.T) {
	var s StringTag
	s.BoolStr = true
	s.IntStr = 42
	s.StrStr = "xzbit"
	got, err := MarshalIndent(&s, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	if got := string(got); got != stringTagExpected {
		t.Fatalf(" got: %s\nwant: %s\n", got, stringTagExpected)
	}

	// Verify that it round-trips with the standard package.
	var s2 StringTag
	err = json.NewDecoder(bytes.NewReader(got)).Decode(&s2)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !reflect.DeepEqual(s, s2) {
		t.Fatalf("decode didn't match.\nsource: %#v\nEncoded as:\n%s\ndecode: %#v", s, string(got), s2)
	}
}

// byte slices are special even if they're renamed types.
type renamedByte byte
type renamedByteSlice []byte
type renamedRenamedByteSlice []renamedByte

func TestEncodeRenamedByteSlice(t *testing.T) {
	s := renamedByteSlice("abc")
	result, err := Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	expect := `"YWJj"`
	if string(result) != expect {
		t.Errorf(" got %s want %s", result, expect)
	}
	r := renamedRenamedByteSlice("abc")
	result, err = Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != expect {
		t.Errorf(" got %s want %s", result, expect)
	}
}

var unsupportedValues = []interface{}{
	math.NaN(),
	math.Inf(-1),
	math.Inf(1),
	// Ill-formed UTF-8
	"\x80",
	"\xcf",
	"\xed",
	"\xed\xa0",
	"\xed\xa0\x7e",
	"\xed\xa0\xc0",
	"\xf8\x82\x83\x84",
	"\xed\xa0\x80\xed\xbf\xbf",
	"\xed\xa0\x80\xed",
	"hello\xffworld",
	"\xff",
	"\xff\xff",
	"a\xffb",
	"\xe6\x97\xa5\xe6\x9c\xac\xff\xaa\x9e",
}

func TestUnsupportedValues(t *testing.T) {
	for i, v := range unsupportedValues {
		label := fmt.Sprintf("#%d [%#v]", i, v)
		if _, err := Marshal(v); err != nil {
			if _, ok := err.(*UnsupportedValueError); !ok {
				t.Errorf("%s: got %T, want UnsupportedValueError", label, err)
			}
		} else {
			t.Errorf("%s: expected error", label)
		}
	}
}

// Ref has Marshaler and Unmarshaler methods with pointer receiver.
type Ref int

func (*Ref) MarshalJSON() ([]byte, error) {
	return []byte(`"ref"`), nil
}

func (r *Ref) UnmarshalJSON([]byte) error {
	*r = 12
	return nil
}

// Val has Marshaler methods with value receiver.
type Val int

func (Val) MarshalJSON() ([]byte, error) {
	return []byte(`"val"`), nil
}

// RefText has Marshaler and Unmarshaler methods with pointer receiver.
type RefText int

func (*RefText) MarshalText() ([]byte, error) {
	return []byte(`"ref"`), nil
}

func (r *RefText) UnmarshalText([]byte) error {
	*r = 13
	return nil
}

// ValText has Marshaler methods with value receiver.
type ValText int

func (ValText) MarshalText() ([]byte, error) {
	return []byte(`"val"`), nil
}

func TestRefValMarshal(t *testing.T) {
	var s = struct {
		R0 Ref
		R1 *Ref
		R2 RefText
		R3 *RefText
		V0 Val
		V1 *Val
		V2 ValText
		V3 *ValText
	}{
		R0: 12,
		R1: new(Ref),
		R2: 14,
		R3: new(RefText),
		V0: 13,
		V1: new(Val),
		V2: 15,
		V3: new(ValText),
	}
	const want = `{"R0":"ref","R1":"ref","R2":"\"ref\"","R3":"\"ref\"","V0":"val","V1":"val","V2":"\"val\"","V3":"\"val\""}`
	b, err := Marshal(&s)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := string(b); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// C implements Marshaler and returns unescaped JSON.
type C int

func (C) MarshalJSON() ([]byte, error) {
	return []byte(`"<&>"`), nil
}

// CText implements Marshaler and returns unescaped text.
type CText int

func (CText) MarshalText() ([]byte, error) {
	return []byte(`"<&>"`), nil
}

func TestMarshalerEscaping(t *testing.T) {
	var c C
	want := `"<&>"`
	b, err := Marshal(c)
	if err != nil {
		t.Fatalf("Marshal(c): %v", err)
	}
	if got := string(b); got != want {
		t.Errorf("Marshal(c) = %#q, want %#q", got, want)
	}

	var ct CText
	want = `"\"<&>\""`
	b, err = Marshal(ct)
	if err != nil {
		t.Fatalf("Marshal(ct): %v", err)
	}
	if got := string(b); got != want {
		t.Errorf("Marshal(ct) = %#q, want %#q", got, want)
	}
}

type IntType int

type MyStruct struct {
	IntType
}

func TestAnonymousNonstruct(t *testing.T) {
	var i IntType = 11
	a := MyStruct{i}
	const want = `{"IntType":11}`

	b, err := Marshal(a)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := string(b); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

type BugA struct {
	S string
}

type BugB struct {
	BugA
	S string
}

type BugC struct {
	S string
}

// Legal Go: We never use the repeated embedded field (S).
type BugX struct {
	A int
	BugA
	BugB
}

// Issue 5245.
func TestEmbeddedBug(t *testing.T) {
	v := BugB{
		BugA{"A"},
		"B",
	}
	b, err := Marshal(v)
	if err != nil {
		t.Fatal("Marshal:", err)
	}
	want := `{"S":"B"}`
	got := string(b)
	if got != want {
		t.Fatalf("Marshal: got %s want %s", got, want)
	}
	// Now check that the duplicate field, S, does not appear.
	x := BugX{
		A: 23,
	}
	b, err = Marshal(x)
	if err != nil {
		t.Fatal("Marshal:", err)
	}
	want = `{"A":23}`
	got = string(b)
	if got != want {
		t.Fatalf("Marshal: got %s want %s", got, want)
	}
}

type BugD struct { // Same as BugA after tagging.
	XXX string `json:"S"`
}

// BugD's tagged S field should dominate BugA's.
type BugY struct {
	BugA
	BugD
}

// Test that a field with a tag dominates untagged fields.
func TestTaggedFieldDominates(t *testing.T) {
	v := BugY{
		BugA{"BugA"},
		BugD{"BugD"},
	}
	b, err := Marshal(v)
	if err != nil {
		t.Fatal("Marshal:", err)
	}
	want := `{"S":"BugD"}`
	got := string(b)
	if got != want {
		t.Fatalf("Marshal: got %s want %s", got, want)
	}
}

// There are no tags here, so S should not appear.
type BugZ struct {
	BugA
	BugC
	BugY // Contains a tagged S field through BugD; should not dominate.
}

func TestDuplicatedFieldDisappears(t *testing.T) {
	v := BugZ{
		BugA{"BugA"},
		BugC{"BugC"},
		BugY{
			BugA{"nested BugA"},
			BugD{"nested BugD"},
		},
	}
	b, err := Marshal(v)
	if err != nil {
		t.Fatal("Marshal:", err)
	}
	want := `{}`
	got := string(b)
	if got != want {
		t.Fatalf("Marshal: got %s want %s", got, want)
	}
}

func TestStringBytes(t *testing.T) {
	// Test that encodeState.stringBytes and encodeState.string use the same encoding.
	es := &encodeState{}
	var r []rune
	for i := '\u0000'; i <= unicode.MaxRune; i++ {
		r = append(r, i)
	}
	// Include some valid non-UTF-8 sequences.
	s := string(r) + "\xed\xbf\xbf\xed\xa0\x80"
	es.string(s)

	esBytes := &encodeState{}
	esBytes.stringBytes([]byte(s))

	enc := es.Buffer.String()
	encBytes := esBytes.Buffer.String()
	if enc != encBytes {
		i := 0
		for i < len(enc) && i < len(encBytes) && enc[i] == encBytes[i] {
			i++
		}
		enc = enc[i:]
		encBytes = encBytes[i:]
		i = 0
		for i < len(enc) && i < len(encBytes) && enc[len(enc)-i-1] == encBytes[len(encBytes)-i-1] {
			i++
		}
		enc = enc[:len(enc)-i]
		encBytes = encBytes[:len(encBytes)-i]

		if len(enc) > 20 {
			enc = enc[:20] + "..."
		}
		if len(encBytes) > 20 {
			encBytes = encBytes[:20] + "..."
		}

		t.Errorf("encodings differ at %#q vs %#q", enc, encBytes)
	}
}

func TestIssue6458(t *testing.T) {
	type Foo struct {
		M RawMessage
	}
	x := Foo{RawMessage(`"foo"`)}

	b, err := Marshal(&x)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"M":"foo"}`; string(b) != want {
		t.Errorf("Marshal(&x) = %#q; want %#q", b, want)
	}

	b, err = Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	if want := `{"M":"ImZvbyI="}`; string(b) != want {
		t.Errorf("Marshal(x) = %#q; want %#q", b, want)
	}
}

func TestIssue10281(t *testing.T) {
	type Foo struct {
		N json.Number
	}
	x := Foo{json.Number(`invalid`)}

	b, err := Marshal(&x)
	if err == nil {
		t.Errorf("Marshal(&x) = %#q; want error", b)
	}
}

// golang.org/issue/8582
func TestEncodePointerString(t *testing.T) {
	type stringPointer struct {
		N *int64 `json:"n,string"`
	}
	var n int64 = 42
	b, err := Marshal(stringPointer{N: &n})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got, want := string(b), `{"n":"42"}`; got != want {
		t.Errorf("Marshal = %s, want %s", got, want)
	}
	var back stringPointer
	err = json.Unmarshal(b, &back)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if back.N == nil {
		t.Fatalf("Unmarshalled nil N field")
	}
	if *back.N != 42 {
		t.Fatalf("*N = %d; want 42", *back.N)
	}
}

var encodeStringTests = []struct {
	in  string
	out string
}{
	{"", `""`},
	{"\x00", `"\u0000"`},
	{"\x01", `"\u0001"`},
	{"\x02", `"\u0002"`},
	{"\x03", `"\u0003"`},
	{"\x04", `"\u0004"`},
	{"\x05", `"\u0005"`},
	{"\x06", `"\u0006"`},
	{"\x07", `"\u0007"`},
	{"\x08", `"\b"`},
	{"\x09", `"\t"`},
	{"\x0a", `"\n"`},
	{"\x0b", `"\u000B"`},
	{"\x0c", `"\f"`},
	{"\x0d", `"\r"`},
	{"\x0e", `"\u000E"`},
	{"\x0f", `"\u000F"`},
	{"\x10", `"\u0010"`},
	{"\x11", `"\u0011"`},
	{"\x12", `"\u0012"`},
	{"\x13", `"\u0013"`},
	{"\x14", `"\u0014"`},
	{"\x15", `"\u0015"`},
	{"\x16", `"\u0016"`},
	{"\x17", `"\u0017"`},
	{"\x18", `"\u0018"`},
	{"\x19", `"\u0019"`},
	{"\x1a", `"\u001A"`},
	{"\x1b", `"\u001B"`},
	{"\x1c", `"\u001C"`},
	{"\x1d", `"\u001D"`},
	{"\x1e", `"\u001E"`},
	{"\x1f", `"\u001F"`},
	{"\x7f", "\"\x7f\""},
	{"\xe6\x97\xa5\xe6\x9c\xac", `"日本"`},
	{"\xed\xa0\x80", `"\uD800"`},
	{"\xed\xa0\x80a", `"\uD800a"`},
	{"\xed\xbf\xbf", `"\uDFFF"`},
	{"\xed\xbf\xbfa", `"\uDFFFa"`},
	{"\xed\xbf\xbf\xed\xa0\x80", `"\uDFFF\uD800"`},
	{"\xed\xbf\xbf\xed\xa0\x80a", `"\uDFFF\uD800a"`},
}

func TestEncodeString(t *testing.T) {
	for _, tt := range encodeStringTests {
		b, err := Marshal(tt.in)
		if err != nil {
			t.Errorf("Marshal(%q): %v", tt.in, err)
			continue
		}
		out := string(b)
		if out != tt.out {
			t.Errorf("Marshal(%q) = %#q, want %#q", tt.in, out, tt.out)
		}
	}
}

var named = map[string]interface{}{
	"\x08": nil,
	"\x09": nil,
	"\x0a": nil,
	"\x0c": nil,
	"\x0d": nil,
	"\x00": nil,
	"\x01": nil,
	"\x02": nil,
	"\x03": nil,
	"\x04": nil,
	"\x05": nil,
	"\x06": nil,
	"\x07": nil,
	"\x0b": nil,
	"\x0e": nil,
	"\x0f": nil,
	"\x10": nil,
	"\x11": nil,
	"\x12": nil,
	"\x13": nil,
	"\x14": nil,
	"\x15": nil,
	"\x16": nil,
	"\x17": nil,
	"\x18": nil,
	"\x19": nil,
	"\x1a": nil,
	"\x1b": nil,
	"\x1c": nil,
	"\x1d": nil,
	"\x1e": nil,
	"\x1f": nil,
	"\x7f": nil,
}

var namedExpected = `{
 "\u0000": null,
 "\u0001": null,
 "\u0002": null,
 "\u0003": null,
 "\u0004": null,
 "\u0005": null,
 "\u0006": null,
 "\u0007": null,
 "\b": null,
 "\t": null,
 "\n": null,
 "\u000B": null,
 "\f": null,
 "\r": null,
 "\u000E": null,
 "\u000F": null,
 "\u0010": null,
 "\u0011": null,
 "\u0012": null,
 "\u0013": null,
 "\u0014": null,
 "\u0015": null,
 "\u0016": null,
 "\u0017": null,
 "\u0018": null,
 "\u0019": null,
 "\u001A": null,
 "\u001B": null,
 "\u001C": null,
 "\u001D": null,
 "\u001E": null,
 "\u001F": null,
 "` + "\x7f" + `": null
}`

func TestEncodeKey(t *testing.T) {
	result, err := MarshalIndent(named, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	if result := string(result); result != namedExpected {
		t.Errorf(" got: %s\nwant: %s\n", result, namedExpected)
	}
}

var floats = map[string][]string{
	"2.5E-3": []string{
		"0.025e-1", "0.0250e-1", "0.02500e-1",
		"0.25e-2", "0.250e-2", "0.2500e-2",
	},
	"2.5E-2": []string{
		"0.025e0", "0.0250e0", "0.02500e0",
		"0.025", "0.0250", "0.02500",
		"0.25e-1", "0.250e-1", "0.2500e-1",
		"2.5e-2", "2.50e-2", "2.500e-2",
	},
	"2.5E-1": []string{
		"0.025e1", "0.0250e1", "0.02500e1",
		"0.25e0", "0.250e0", "0.2500e0",
		"0.25", "0.250", "0.2500",
		"2.5e-1", "2.50e-1", "2.500e-1",
		"25e-2", "25.0e-2", "25.00e-2",
	},
	"2.5E0": []string{
		"0.025e2", "0.0250e2", "0.02500e2",
		"0.25e1", "0.250e1", "0.2500e1",
		"2.5e0", "2.50e0", "2.500e0",
		"2.5", "2.50", "2.500",
		"25e-1", "25.0e-1", "25.00e-1",
		"250e-2", "250.0e-2", "250.00e-2",
	},
	"25": []string{
		"0.25e2", "0.250e2", "0.2500e2",
		"2.5e1", "2.50e1", "2.500e1",
		"25e0", "25.0e0", "25.00e0",
		"25", "25.0", "25.00",
		"250e-1", "250.0e-1", "250.00e-1",
	},
	"250": []string{
		"2.5e2", "2.50e2", "2.500e2",
		"25e1", "25.0e1", "25.00e1",
		"250e0", "250.0e0", "250.00e0",
		"250", "250.0", "250.00",
	},
	"2500": []string{
		"25e2", "25.0e2", "25.00e2",
		"250e1", "250.0e1", "250.00e1",
	},
}

func TestFloat(t *testing.T) {
	for expected, inputs := range floats {
		for _, input := range inputs {
			inputFloat, _ := strconv.ParseFloat(input, 64)
			result, err := Marshal(inputFloat)
			if err != nil {
				t.Fatal(err)
			}
			if string(result) != expected {
				t.Errorf(" float(%s)\n got: %s\n want: %s\n", input, result, expected)
			}

			inputNumber := Number(input)
			result, err = Marshal(inputNumber)
			if err != nil {
				t.Fatal(err)
			}
			if string(result) != expected {
				t.Errorf(" Number(%s)\n got: %s\n want: %s\n", input, result, expected)
			}

			inputJsonNumber := json.Number(input)
			result, err = Marshal(inputJsonNumber)
			if err != nil {
				t.Fatal(err)
			}
			if string(result) != expected {
				t.Errorf(" json.Number(%s)\n got: %s\n want: %s\n", input, result, expected)
			}
		}
	}
}
