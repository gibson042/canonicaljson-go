// Copyright 2016 Richard Gibson. All rights reserved.
// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package canonicaljson

import (
	"encoding/json"
	"testing"
)

// Tests of simple examples.

type example struct {
	compact string
	indent  string
}

var examples = []example{
	{`1`, `1`},
	{`{}`, `{}`},
	{`[]`, `[]`},
	{`{"":2}`, "{\n\t\"\": 2\n}"},
	{`[3]`, "[\n\t3\n]"},
	{`[1,2,3]`, "[\n\t1,\n\t2,\n\t3\n]"},
	{`{"x":1}`, "{\n\t\"x\": 1\n}"},
	{ex1, ex1i},
}

var ex1 = `[true,false,null,"x",1,1.5E0,0,-500]`

var ex1i = `[
	true,
	false,
	null,
	"x",
	1,
	1.5E0,
	0,
	-500
]`

func TestMarshal(t *testing.T) {
	var obj interface{}
	for _, tt := range examples {
		if err := json.Unmarshal([]byte(tt.compact), &obj); err != nil {
			t.Errorf("json.Unmarshal(%#q): %v", tt.compact, err)
		} else {
			s, err := Marshal(obj)
			if err != nil {
				t.Errorf("Marshal(%#q): %v", tt.compact, err)
			} else if string(s) != tt.compact {
				t.Errorf("Marshal(%#q) = %#q, want original", tt.compact, string(s))
			}
		}

		if err := json.Unmarshal([]byte(tt.indent), &obj); err != nil {
			t.Errorf("json.Unmarshal(%#q): %v", tt.indent, err)
		} else {
			s, err := Marshal(obj)
			if err != nil {
				t.Errorf("Marshal(%#q): %v", tt.indent, err)
			} else if string(s) != tt.compact {
				t.Errorf("Marshal(%#q) = %#q, want %#q", tt.indent, string(s), tt.compact)
			}
		}
	}
}

func TestMarshalSeparators(t *testing.T) {
	// U+2028 and U+2029 should be unescaped inside strings.
	// They should not appear outside strings.
	tests := []struct {
		in, compact string
	}{
		{"{\"\u2028\": 1}", "{\"\u2028\":1}"},
		{"{\"\u2029\" :2}", "{\"\u2029\":2}"},
	}
	for _, tt := range tests {
		var obj interface{}
		if err := json.Unmarshal([]byte(tt.in), &obj); err != nil {
			t.Errorf("json.Unmarshal(%q): %v", tt.in, err)
		} else {
			s, err := Marshal(obj)
			if err != nil {
				t.Errorf("Marshal(%#q): %v", tt.in, err)
			} else if string(s) != tt.compact {
				t.Errorf("Marshal(%q) = %q, want %q", tt.in, s, tt.compact)
			}
		}
	}
}

func TestMarshalIndent(t *testing.T) {
	var obj interface{}
	for _, tt := range examples {
		if err := json.Unmarshal([]byte(tt.indent), &obj); err != nil {
			t.Errorf("json.Unmarshal(%#q): %v", tt.indent, err)
		} else {
			s, err := MarshalIndent(obj, "", "\t")
			if err != nil {
				t.Errorf("MarshalIndent(%#q): %v", tt.indent, err)
			} else if string(s) != tt.indent {
				t.Errorf("MarshalIndent(%#q) = %#q, want original", tt.indent, s)
			}
		}

		if err := json.Unmarshal([]byte(tt.compact), &obj); err != nil {
			t.Errorf("json.Unmarshal(%#q): %v", tt.compact, err)
		} else {
			s, err := MarshalIndent(obj, "", "\t")
			if err != nil {
				t.Errorf("MarshalIndent(%#q): %v", tt.compact, err)
			} else if string(s) != tt.indent {
				t.Errorf("MarshalIndent(%#q) = %#q, want %#q", tt.compact, s, tt.indent)
			}
		}
	}
}
