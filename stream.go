// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package canonicaljson

import (
	"io"
)

func nonSpace(b []byte) bool {
	for _, c := range b {
		if !isSpace(c) {
			return true
		}
	}
	return false
}

// An Encoder writes JSON objects to an output stream.
type Encoder struct {
	w   io.Writer
	err error
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the JSON encoding of v to the stream,
// followed by a newline character.
//
// See the documentation for Marshal for details about the
// conversion of Go values to JSON.
func (enc *Encoder) Encode(v interface{}) error {
	if enc.err != nil {
		return enc.err
	}
	e := newEncodeState()
	err := e.marshal(v)
	if err != nil {
		return err
	}

	// Terminate each value with a newline.
	// This makes the output look a little nicer
	// when debugging, and some kind of space
	// is required if the encoded value was a number,
	// so that the reader knows there aren't more
	// digits coming.
	e.WriteByte('\n')

	if _, err = enc.w.Write(e.Bytes()); err != nil {
		enc.err = err
	}
	encodeStatePool.Put(e)
	return err
}

// RawMessage is a raw encoded JSON object.
// It implements Marshaler and can
// be used to precompute a JSON encoding.
type RawMessage []byte

// MarshalJSON returns *m as the JSON encoding of m.
func (m *RawMessage) MarshalJSON() ([]byte, error) {
	return *m, nil
}

var _ Marshaler = (*RawMessage)(nil)

// A Token holds a value of one of these types:
//
//	Delim, for the four JSON delimiters [ ] { }
//	bool, for JSON booleans
//	float64, for JSON numbers
//	Number, for JSON numbers
//	string, for JSON string literals
//	nil, for JSON null
//
type Token interface{}

const (
	tokenTopValue = iota
	tokenArrayStart
	tokenArrayValue
	tokenArrayComma
	tokenObjectStart
	tokenObjectKey
	tokenObjectColon
	tokenObjectValue
	tokenObjectComma
)

// A Delim is a JSON array or object delimiter, one of [ ] { or }.
type Delim rune

func (d Delim) String() string {
	return string(d)
}

func clearOffset(err error) {
	if s, ok := err.(*SyntaxError); ok {
		s.Offset = 0
	}
}

/*
TODO

// EncodeToken writes the given JSON token to the stream.
// It returns an error if the delimiters [ ] { } are not properly used.
//
// EncodeToken does not call Flush, because usually it is part of
// a larger operation such as Encode, and those will call Flush when finished.
// Callers that create an Encoder and then invoke EncodeToken directly,
// without using Encode, need to call Flush when finished to ensure that
// the JSON is written to the underlying writer.
func (e *Encoder) EncodeToken(t Token) error  {
	...
}

*/
