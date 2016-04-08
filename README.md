# canonicaljson-go
Go library for producing JSON in canonical format as specified by [http://gibson042.github.io/canonicaljson-spec/](http://gibson042.github.io/canonicaljson-spec/).
The provided interface matches that of standard package "[encoding/json](https://golang.org/pkg/encoding/json/)" wherever they overlap:
* [func `Marshal`](https://golang.org/pkg/encoding/json/#Marshal)
* [func `MarshalIndent`](https://golang.org/pkg/encoding/json/#MarshalIndent)
* [type `Encoder`](https://golang.org/pkg/encoding/json/#Encoder)

Types from the standard package are also accepted wherever relevant.

Test this package by invoking `test.sh`.

```
PACKAGE DOCUMENTATION

package canonicaljson
    import "."

    Package canonicaljson implements canonical serialization of Go objects
    to JSON as specified in "JSON Canonical Form" Internet Draft
    https://tools.ietf.org/html/draft-staykov-hu-json-canonical-form-00 and
    updated to include non-exponential integer representation and
    shortest-representation normalization of all strings (both keys and
    values). The provided interface should match that of standard package
    "encoding/json" (from which it is derived) wherever they overlap (and in
    fact, this package is essentially a 2016-03-09 fork from
    golang/go@9d77ad8d34ce56e182adc30cd21af50a4b00932c:src/encoding/json ).
    Notable differences:

	- Object keys are sorted lexicographically by codepoint
	- Non-integer JSON numbers are represented in capital-E exponential
	  notation with significand in (-10, 10) and no insignificant signs
	  or zeroes beyond those required to force a decimal point.
	- JSON strings are represented in UTF-8 with minimal byte length,
	  using escapes only when necessary for validity and Unicode
	  escapes (lowercase hex) only when there is no shorter option.

FUNCTIONS

func Marshal(v interface{}) ([]byte, error)
    Marshal returns the canonical JSON encoding of v.

    Marshal traverses the value v recursively. If an encountered value
    implements the json.Marshaler interface and is not a nil pointer,
    Marshal calls its MarshalJSON method to produce JSON. If no MarshalJSON
    method is present but the value implements encoding.TextMarshaler
    instead, Marshal calls its MarshalText method. The nil pointer exception
    is not strictly necessary but mimics a similar, necessary exception in
    the behavior of json.UnmarshalJSON.

    Otherwise, Marshal uses the following type-dependent default encodings:

    Boolean values encode as JSON booleans.

    Floating point, integer, and Number values encode as JSON numbers.
    Non-fractional values become sequences of digits without leading spaces;
    fractional values are represented in capital-E exponential notation with
    the shortest possible significand of magnitude less than 10 that
    includes at least one digit both before and after the decimal point, and
    the shortest possible non-empty exponent.

    String values encode as JSON strings coerced to valid UTF-8, replacing
    invalid bytes with U+FFFD REPLACEMENT CHARACTER. U+2028 LINE SEPARATOR
    and U+2029 PARAGRAPH SEPARATOR (valid in JSON strings but invalid in
    _JavaScript_ strings) are *not* escaped. Control characters U+0000
    through U+001F are replaced with their shortest escape sequence, 4
    lowercase hex characters except for the following:

	- \b U+0008 BACKSPACE
	- \t U+0009 CHARACTER TABULATION ("tab")
	- \n U+000A LINE FEED ("newline")
	- \f U+000C FORM FEED
	- \r U+000D CARRIAGE RETURN

    Array and slice values encode as JSON arrays, except that []byte encodes
    as a base64-encoded string, and a nil slice encodes as the null JSON
    object.

    Struct values encode as JSON objects. Each exported struct field becomes
    a member of the object unless

	- the field's tag is "-", or
	- the field is empty and its tag specifies the "omitempty" option.

    The empty values are false, 0, any nil pointer or interface value, and
    any array, slice, map, or string of length zero. The object's default
    key string is the struct field name but can be specified in the struct
    field's tag value. The "json" key in the struct field's tag value is the
    key name, followed by an optional comma and options. Examples:

	// Field is ignored by this package.
	Field int `json:"-"`

	// Field appears in JSON as key "myName".
	Field int `json:"myName"`

	// Field appears in JSON as key "myName" and
	// the field is omitted from the object if its value is empty,
	// as defined above.
	Field int `json:"myName,omitempty"`

	// Field appears in JSON as key "Field" (the default), but
	// the field is skipped if empty.
	// Note the leading comma.
	Field int `json:",omitempty"`

    The "string" option signals that a field is stored as JSON inside a
    JSON-encoded string. It applies only to fields of string, floating
    point, integer, or boolean types. This extra level of encoding is
    sometimes used when communicating with JavaScript programs:

	Int64String int64 `json:",string"`

    The key name will be used if it's a non-empty string consisting of only
    Unicode letters, digits, dollar signs, percent signs, hyphens,
    underscores and slashes.

    Anonymous struct fields are usually marshaled as if their inner exported
    fields were fields in the outer struct, subject to the usual Go
    visibility rules amended as described in the next paragraph. An
    anonymous struct field with a name given in its JSON tag is treated as
    having that name, rather than being anonymous. An anonymous struct field
    of interface type is treated the same as having that type as its name,
    rather than being anonymous.

    The Go visibility rules for struct fields are amended for JSON when
    deciding which field to marshal. If there are multiple fields at the
    same level, and that level is the least nested (and would therefore be
    the nesting level selected by the usual Go rules), the following extra
    rules apply:

    1) Of those fields, if any are JSON-tagged, only tagged fields are
    considered, even if there are multiple untagged fields that would
    otherwise conflict. 2) If there is exactly one field (tagged or not
    according to the first rule), that is selected. 3) Otherwise there are
    multiple fields, and all are ignored; no error occurs.

    Handling of anonymous struct fields is new in Go 1.1. Prior to Go 1.1,
    anonymous struct fields were ignored. To force ignoring of an anonymous
    struct field in both current and earlier versions, give the field a JSON
    tag of "-".

    Map values encode as JSON objects. The map's key type must be string;
    the map keys are used as JSON object keys, subject to the UTF-8 coercion
    described for string values above.

    Pointer values encode as the value pointed to. A nil pointer encodes as
    the null JSON object.

    Interface values encode as the value contained in the interface. A nil
    interface value encodes as the null JSON object.

    Channel, complex, and function values cannot be encoded in JSON.
    Attempting to encode such a value causes Marshal to return an
    UnsupportedTypeError.

    JSON cannot represent cyclic data structures and Marshal does not handle
    them. Passing cyclic structures to Marshal will result in an infinite
    recursion.

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)
    MarshalIndent is like Marshal, but adds whitespace for more readable
    output.

TYPES

type Delim rune
    A Delim is a JSON array or object delimiter, one of [ ] { or }.

func (d Delim) String() string

type Encoder struct {
    // contains filtered or unexported fields
}
    An Encoder writes JSON objects to an output stream.

func NewEncoder(w io.Writer) *Encoder
    NewEncoder returns a new encoder that writes to w.

func (enc *Encoder) Encode(v interface{}) error
    Encode writes the JSON encoding of v to the stream, followed by a
    newline character.

    See the documentation for Marshal for details about the conversion of Go
    values to JSON.

type Marshaler interface {
    MarshalJSON() ([]byte, error)
}
    Marshaler is the interface implemented by objects that can marshal
    themselves into valid JSON.

type MarshalerError struct {
    Type reflect.Type
    Err  error
}

func (e *MarshalerError) Error() string

type Number string
    A Number represents a JSON number literal.

func (n Number) Float64() (float64, error)
    Float64 returns the number as a float64.

func (n Number) Int64() (int64, error)
    Int64 returns the number as an int64.

func (n Number) String() string
    String returns the literal text of the number.

type RawMessage []byte
    RawMessage is a raw encoded JSON object. It implements json.Marshaler
    and can be used to precompute a JSON encoding.

func (m *RawMessage) MarshalJSON() ([]byte, error)
    MarshalJSON returns *m as the JSON encoding of m.

type SyntaxError struct {
    Offset int64 // error occurred after reading Offset bytes
    // contains filtered or unexported fields
}
    A SyntaxError is a description of a JSON syntax error.

func (e *SyntaxError) Error() string

type Token interface{}
    A Token holds a value of one of these types:

	Delim, for the four JSON delimiters [ ] { }
	bool, for JSON booleans
	float64, for JSON numbers
	Number, for JSON numbers
	string, for JSON string literals
	nil, for JSON null

type UnsupportedTypeError struct {
    Type reflect.Type
}
    An UnsupportedTypeError is returned by Marshal when attempting to encode
    an unsupported value type.

func (e *UnsupportedTypeError) Error() string

type UnsupportedValueError struct {
    Value reflect.Value
    Str   string
}

func (e *UnsupportedValueError) Error() string

SUBDIRECTORIES

	canonicaljson-spec
	cli

```
