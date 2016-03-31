// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package canonicaljson

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"
)

// Test values for the stream test.
// One of each JSON kind.
var streamTest = []interface{}{
	0.1,
	"hello",
	nil,
	true,
	false,
	[]interface{}{"a", "b", "c"},
	map[string]interface{}{"K": "Kelvin", "ß": "long s"},
	3.14, // another value to make sure something can follow map
}

var streamEncoded = `1.0E-1
"hello"
null
true
false
["a","b","c"]
{"ß":"long s","K":"Kelvin"}
3.14E0
`

func TestEncoder(t *testing.T) {
	for i := 0; i <= len(streamTest); i++ {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		for j, v := range streamTest[0:i] {
			if err := enc.Encode(v); err != nil {
				t.Fatalf("encode #%d: %v", j, err)
			}
		}
		if have, want := buf.String(), nlines(streamEncoded, i); have != want {
			t.Errorf("encoding %d items: mismatch", i)
			diff(t, []byte(have), []byte(want))
			break
		}
	}
}

func nlines(s string, n int) string {
	if n <= 0 {
		return ""
	}
	for i, c := range s {
		if c == '\n' {
			if n--; n == 0 {
				return s[0 : i+1]
			}
		}
	}
	return s
}

func TestRawMessage(t *testing.T) {
	// TODO(rsc): Should not need the * in *RawMessage
	var data struct {
		X  float64
		Id *RawMessage
		Y  float32
	}
	const raw = `["\u0056",null]`
	const msg = `{"Id":["V",null],"X":1.0E-1,"Y":2.0E-1}`
	data.X = 0.1
	data.Y = 0.2
	id := RawMessage(raw)
	data.Id = &id
	b, err := Marshal(&data)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(b) != msg {
		t.Fatalf("Marshal: have %#q want %#q", b, msg)
	}
}

func TestNullRawMessage(t *testing.T) {
	// TODO(rsc): Should not need the * in *RawMessage
	var data struct {
		X  float64
		Id *RawMessage
		Y  float32
	}
	data.Id = new(RawMessage)
	const msg = `{"Id":null,"X":1.0E-1,"Y":2.0E-1}`
	err := json.Unmarshal([]byte(msg), &data)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if data.Id != nil {
		t.Fatalf("Raw mismatch: have non-nil, want nil")
	}
	b, err := Marshal(&data)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(b) != msg {
		t.Fatalf("Marshal: have %#q want %#q", b, msg)
	}
}

var blockingTests = []string{
	`{"x": 1}`,
	`[1, 2, 3]`,
}

func BenchmarkEncoderEncode(b *testing.B) {
	b.ReportAllocs()
	type T struct {
		X, Y string
	}
	v := &T{"foo", "bar"}
	for i := 0; i < b.N; i++ {
		if err := NewEncoder(ioutil.Discard).Encode(v); err != nil {
			b.Fatal(err)
		}
	}
}

type tokenStreamCase struct {
	json      string
	expTokens []interface{}
}

type decodeThis struct {
	v interface{}
}

var tokenStreamCases []tokenStreamCase = []tokenStreamCase{
	// streaming token cases
	{json: `10`, expTokens: []interface{}{float64(10)}},
	{json: ` [10] `, expTokens: []interface{}{
		Delim('['), float64(10), Delim(']')}},
	{json: ` [false,10,"b"] `, expTokens: []interface{}{
		Delim('['), false, float64(10), "b", Delim(']')}},
	{json: `{ "a": 1 }`, expTokens: []interface{}{
		Delim('{'), "a", float64(1), Delim('}')}},
	{json: `{"a": 1, "b":"3"}`, expTokens: []interface{}{
		Delim('{'), "a", float64(1), "b", "3", Delim('}')}},
	{json: ` [{"a": 1},{"a": 2}] `, expTokens: []interface{}{
		Delim('['),
		Delim('{'), "a", float64(1), Delim('}'),
		Delim('{'), "a", float64(2), Delim('}'),
		Delim(']')}},
	{json: `{"obj": {"a": 1}}`, expTokens: []interface{}{
		Delim('{'), "obj", Delim('{'), "a", float64(1), Delim('}'),
		Delim('}')}},
	{json: `{"obj": [{"a": 1}]}`, expTokens: []interface{}{
		Delim('{'), "obj", Delim('['),
		Delim('{'), "a", float64(1), Delim('}'),
		Delim(']'), Delim('}')}},

	// streaming tokens with intermittent Decode()
	{json: `{ "a": 1 }`, expTokens: []interface{}{
		Delim('{'), "a",
		decodeThis{float64(1)},
		Delim('}')}},
	{json: ` [ { "a" : 1 } ] `, expTokens: []interface{}{
		Delim('['),
		decodeThis{map[string]interface{}{"a": float64(1)}},
		Delim(']')}},
	{json: ` [{"a": 1},{"a": 2}] `, expTokens: []interface{}{
		Delim('['),
		decodeThis{map[string]interface{}{"a": float64(1)}},
		decodeThis{map[string]interface{}{"a": float64(2)}},
		Delim(']')}},
	{json: `{ "obj" : [ { "a" : 1 } ] }`, expTokens: []interface{}{
		Delim('{'), "obj", Delim('['),
		decodeThis{map[string]interface{}{"a": float64(1)}},
		Delim(']'), Delim('}')}},

	{json: `{"obj": {"a": 1}}`, expTokens: []interface{}{
		Delim('{'), "obj",
		decodeThis{map[string]interface{}{"a": float64(1)}},
		Delim('}')}},
	{json: `{"obj": [{"a": 1}]}`, expTokens: []interface{}{
		Delim('{'), "obj",
		decodeThis{[]interface{}{
			map[string]interface{}{"a": float64(1)},
		}},
		Delim('}')}},
	{json: ` [{"a": 1} {"a": 2}] `, expTokens: []interface{}{
		Delim('['),
		decodeThis{map[string]interface{}{"a": float64(1)}},
		decodeThis{&SyntaxError{"expected comma after array element", 0}},
	}},
	{json: `{ "a" 1 }`, expTokens: []interface{}{
		Delim('{'), "a",
		decodeThis{&SyntaxError{"expected colon after object key", 0}},
	}},
}

func diff(t *testing.T, a, b []byte) {
	for i := 0; ; i++ {
		if i >= len(a) || i >= len(b) || a[i] != b[i] {
			j := i - 10
			if j < 0 {
				j = 0
			}
			t.Errorf("diverge at %d: «%s» vs «%s»", i, trim(a[j:]), trim(b[j:]))
			return
		}
	}
}

func trim(b []byte) []byte {
	if len(b) > 20 {
		return b[0:20]
	}
	return b
}
