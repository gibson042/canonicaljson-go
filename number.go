// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Represents JSON data structure using native Go types: booleans, floats,
// strings, arrays, and maps.

package canonicaljson

import (
	"reflect"
	"strconv"
)

// A Number represents a JSON number literal.
type Number string

// String returns the literal text of the number.
func (n Number) String() string { return string(n) }

// Float64 returns the number as a float64.
func (n Number) Float64() (float64, error) {
	return strconv.ParseFloat(string(n), 64)
}

// Int64 returns the number as an int64.
func (n Number) Int64() (int64, error) {
	return strconv.ParseInt(string(n), 10, 64)
}

// isValidNumber reports whether s is a valid JSON number literal.
func isValidNumber(s string) bool {
	// This function implements the JSON numbers grammar.
	// See https://tools.ietf.org/html/rfc7159#section-6
	// and http://json.org/number.gif

	if s == "" {
		return false
	}

	// Optional -
	if s[0] == '-' {
		s = s[1:]
		if s == "" {
			return false
		}
	}

	// Digits
	switch {
	default:
		return false

	case s[0] == '0':
		s = s[1:]

	case '1' <= s[0] && s[0] <= '9':
		s = s[1:]
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// . followed by 1 or more digits.
	if len(s) >= 2 && s[0] == '.' && '0' <= s[1] && s[1] <= '9' {
		s = s[2:]
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// e or E followed by an optional - or + and
	// 1 or more digits.
	if len(s) >= 2 && (s[0] == 'e' || s[0] == 'E') {
		s = s[1:]
		if s[0] == '+' || s[0] == '-' {
			s = s[1:]
			if s == "" {
				return false
			}
		}
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// Make sure we are at the end.
	return s == ""
}

var numberType = reflect.TypeOf(Number(""))
