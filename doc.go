/*
Package bson implements reading and writing bson.

	The BSON encoder/decoder allows the programmer to be strict about coercion.
	If exact BSON types are used (which is encouraged) then the encoder/decoder
	is 100% symmetric.

	Documents:
	The following types serve as BSON documents.

	Map:    Does not preserve order. Most commonly used document type.
	Slice:  Preserves order. If order is not required use Map.
	BSON:   Raw BSON. Used to support preencoded BSON for efficiency.
	struct: Only struct encoding is supported.

	Supported struct tags:
	Field int `bson:"-"`                // Ignored.
	Field int `bson:"myName"`           // Encoded with key "myName".
	Field int `bson:"myName,omitempty"` // Key "myName". Ignore if empty value.
	Field int `bson:",omitempty"`       // Ignore if zero (note the ',').

	Empty values:
	Empty value is defined as false, 0, nil, empty slice, empty map, or empty
	string.

	Coercion:
	Coercion is used when exact BSON types are not used. The following coercions
	are supported. Types not listed are unsupported and will generate errors
	during encoding.

	nil       -> Null
	bool      -> Bool
	int       -> Int64
	int8      -> Int32
	int16     -> Int32
	int32     -> Int32
	int64     -> Int64
	float64   -> Float
	string    -> String
	time.Time -> UTCDateTime
	[]byte    -> Binary

	*Binary is encoded with subtype 0x00.
	*Binary subtypes are ignored while decoding.

	Reaching Into Documents:
	There is significant boiler plate associated with unmarshaling BSON. For this
	reason a "reach" funcs are prodied to traverse documents and pick out specific
	values. See example code.

	Example:
	doc := Map{"foo": Map{"bar": String("baz")}}
	var dst string
	ok, err := doc.Reach(&dst, "foo", "bar")
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("foo.bar not found.")
	}
	fmt.Print(dst)
	// Output: baz
*/
package bson

/*
	BSON Specification

	Basic Types:
	The following basic types are used as terminals in the rest of the grammar.
	Each type must be serialized in little-endian format.

	byte    1 byte  (8-bits)
	int32   4 bytes (32-bit signed integer)
	int64   8 bytes (64-bit signed integer)
	double  8 bytes (64-bit IEEE 754 floating point)

	Non-terminals:
	The following specifies the rest of the BSON grammar. Note that quoted
	strings represent terminals, and should be interpreted with C semantics (e.g.
	"\x01" represents the byte 0000 0001). Also note that we use the * operator
	as shorthand for repetition (e.g. ("\x01"*2) is "\x01\x01"). When used as a
	unary operator, * means that the repetition can occur 0 or more times.

	document ::= int32 e_list "\x00"            BSON Document
	e_list   ::= element e_list                 Sequence of elements
	           | ""	
	element  ::= "\x01" e_name double           Floating point
	           | "\x02" e_name string           UTF-8 string
	           | "\x03" e_name document         Embedded document
	           | "\x04" e_name document         Array
	           | "\x05" e_name binary           Binary data
	           | "\x06" e_name                  Undefined — Deprecated
	           | "\x07" e_name (byte*12)        ObjectId
	           | "\x08" e_name "\x00"           Boolean "false"
	           | "\x08" e_name "\x01"           Boolean "true"
	           | "\x09" e_name int64            UTC datetime
	           | "\x0A" e_name                  Null value
	           | "\x0B" e_name cstring cstring  Regular expression
	           | "\x0C" e_name string (byte*12) DBPointer — Deprecated
	           | "\x0D" e_name string           JavaScript code
	           | "\x0E" e_name string           Symbol
	           | "\x0F" e_name code_w_s         JavaScript code w/ scope
	           | "\x10" e_name int32            32-bit Integer
	           | "\x11" e_name int64            Timestamp
	           | "\x12" e_name int64            64-bit integer
	           | "\xFF" e_name                  Min key
	           | "\x7F" e_name                  Max key
	e_name	 ::= cstring                        Key name
	string	 ::= int32 (byte*) "\x00"           String
	cstring	 ::= (byte*) "\x00"                 CString
	binary	 ::= int32 subtype (byte*)          Binary
	subtype	 ::= "\x00"                         Binary / Generic
	           | "\x01"                         Function
	           | "\x02"                         Binary (Old)
	           | "\x03"                         UUID
	           | "\x05"                         MD5
	           | "\x80"                         User defined
	code_w_s ::= int32 string document          Code w/ scope

	Examples:
	{"hello": "world"}
	"\x16\x00\x00\x00\x02hello\x00\x06\x00\x00\x00world\x00\x00"

	{"BSON": ["awesome", 5.05, 1986]}
	"1\x00\x00\x00\x04BSON\x00&\x00\x00\x00\x020\x00\x08\x00\x00\x00awesome\x00
	\x011\x00333333\x14@\x102\x00\xc2\x07\x00\x00\x00\x00"
*/
