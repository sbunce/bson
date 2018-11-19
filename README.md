BSON encoder/decoder.
=
This BSON encoder/decoder allows the programmer to be strict about coercion. If exact BSON types are used (which is encouraged) then the encoder/decoder is 100% symmetric.


Documents
---------
The following types serve as BSON documents.

#### Map
Map is the most commonly used document type. However, it does not support ordering. For this reason Slice is provided.

#### Slice
Slice is used when ordering in the encoded document must be preserved. If order is not required a Map should be used.

#### BSON
BSON is a raw BSON type. This is a supported document type so we can use preencoded documents for encoding efficiency. It also allows us to partially decode a document for decoding effiency.

#### Structs
Currently structs are only supported by the encoder.

    Field int `bson:"-"`                // Ignored.
    Field int `bson:"myName"`           // Encoded with key "myName".
    Field int `bson:"myName,omitempty"` // Key "myName". Ignore if empty value.
    Field int `bson:",omitempty"`       // Ignore if zero (note the ',').

    *Empty value is defined as false, 0, nil, empty slice, empty map, or empty string.

Coercion
--------
Coercion is used when exact BSON types are not used. The following coercions are supported. Types not listed are unsupported and will generate errors during encoding.

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

Reach
-----
There is significant boiler plate associated with unmarshaling BSON. For this reason "reach" funcs are provided to traverse documents and pick out specific values.
 
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
