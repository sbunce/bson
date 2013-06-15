bson
====

Golang BSON encoder/decoder that allows being very strict about type conversion.

Coercion
========
  Coercion is supported. To avoid all coercions exact bson types can be used.
  
  Encoding Coercions (types not listed are unsupported):
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

  *Binary is encoded with subtype 0x00. Subtypes are ignored while decoding.
  *If using exact BSON types then encoding/decoding is perfectly symmetric.

Documents
=========
  The following go types serve as bson documents.
 
  Map
    Most commonly used type. Does not preserve order.
  Slice
    Used to preserve order, otherwise Map should be used.
  struct
    Only struct -> bson is supported.

  Struct Tags:
    Field int `bson:"-"`                // Ignored.
    Field int `bson:"myName"`           // Encoded with key "myName".
    Field int `bson:"myName,omitempty"` // Key "myName". Ignore if empty value.
    Field int `bson:",omitempty"`       // Ignore if zero (note the ',').

  Empty Value:
    This is exactly the same as the json package in the standard library.
    false, 0, any nil pointer or interface value, any array, slice, map or string
    of length zero.

Reach
=====
	When unmarshaling BSON there is significant boiler plate when trying to get
	a object deeply nested in a doc. For this reason "Reach" funcs were created.
	With a reach func you can specify how to reach in to a document and pick out
	one object in a very concise way.

	doc := Map{
		"foo": Map{
			"bar": Bool(true),
		},
	}
	var dst bool
	ok, err = doc.Reach(&dst, "foo", "bar")
	if err != nil {
		return err
	}
	if ok {
		fmt.Println("bar found")
	} else {
		fmt.Println("bar not found")
	}
