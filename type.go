package bson

// Wire types.
const (
	_FLOATING_POINT    = 0x01 // "\x01" e_name double           Floating point
	_STRING            = 0x02 // "\x02" e_name string           UTF-8 string
	_EMBEDDED_DOCUMENT = 0x03 // "\x03" e_name document         Embedded document
	_ARRAY             = 0x04 // "\x04" e_name document         Array
	_BINARY_DATA       = 0x05 // "\x05" e_name binary           Binary data
	_UNDEFINED         = 0x06 // "\x06" e_name                  Undefined - Deprecated
	_OBJECT_ID         = 0x07 // "\x07" e_name (byte*12)        ObjectId
	_BOOLEAN           = 0x08 // "\x08" e_name "\x00"           Boolean "false"
	                          // "\x08" e_name "\x01"           Boolean "true"
	_UTC_DATETIME      = 0x09 // "\x09" e_name int64            UTC datetime
	_NULL_VALUE        = 0x0A // "\x0A" e_name                  Null value
	_REGEXP            = 0x0B // "\x0B" e_name cstring cstring  Regular expression
	_DBPOINTER         = 0x0C // "\x0C" e_name string (byte*12) DBPointer - Deprecated
	_JAVASCRIPT        = 0x0D // "\x0D" e_name string           JavaScript code
	_SYMBOL            = 0x0E // "\x0E" e_name string           Symbol
	_JAVASCRIPT_SCOPE  = 0x0F // "\x0F" e_name code_w_s         JavaScript code w/ scope
	_32BIT_INTEGER     = 0x10 // "\x10" e_name int32            32-bit Integer
	_TIMESTAMP         = 0x11 // "\x11" e_name int64            Timestamp
	_64BIT_INTEGER     = 0x12 // "\x12" e_name int64            64-bit integer
	_MIN_KEY           = 0xFF // "\xFF" e_name                  Min key
	_MAX_KEY           = 0x7F // "\x7F" e_name                  Max key
)

// BSON type.
type Float float64

// BSON type.
type String string

// BSON type.
type Array []interface{}

// BSON type.
type Binary []byte

// BSON type. Value is ignored.
type Undefined struct{}

// BSON type. Must be 12 bytes. Typically generated with NewObjectId().
type ObjectId []byte

// BSON type.
type Bool bool

// BSON type. Milliseconds since unix epoch.
type UTCDateTime int64

// BSON type. Value is ignored.
type Null struct{}

// BSON type.
type Regexp struct {
	Pattern string
	Options string
}

// BSON type.
type DBPointer struct {
	Name     string
	ObjectId ObjectId
}

// BSON type.
type Javascript string

// BSON type.
type Symbol string

// BSON type. The Scope must be a BSON doc.
type JavascriptScope struct {
	Javascript string
	Scope      Map
}

// BSON type.
type Int32 int32

// BSON type.
type Timestamp int64

// BSON type.
type Int64 int64

// BSON type.
type MinKey struct{}

// BSON type.
type MaxKey struct{}
