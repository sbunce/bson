package bson

import (
	"reflect"
	"testing"
	"time"
)

func TestEncodeCoercion(t *testing.T) {
	now := time.Now()
	src := Map{
		"null":    nil,
		"bool":    true,
		"int" :    int(123),
		"int8":    int8(123),
		"int16":   int16(123),
		"int32":   int32(123),
		"int64":   int64(123),
		"float64": float64(123.123),
		"string":  "foo",
		"gotime":  now,
	}
	exp := Map{
		"null":    Null{},
		"bool":    Bool(true),
		"int" :    Int64(123),
		"int8":    Int32(123),
		"int16":   Int32(123),
		"int32":   Int32(123),
		"int64":   Int64(123),
		"float64": Float(123.123),
		"string":  String("foo"),
		"gotime":  UTCDateTime(now.UnixNano()/1000/1000),
	}
	bs, err := src.Encode()
	if err != nil {
		t.Fatal(err, src)
	}
	dst, err := bs.Map()
	if err != nil {
		t.Fatal(err, dst, exp)
	}
	if !reflect.DeepEqual(dst, exp) {
		t.Fatal(dst, exp)
	}
}

func TestReachCoerce(t *testing.T) {
	src := Map{
		"foo": Map{
			"Float":       Float(123.123),
			"String":      String("foo"),
			"Binary":      Binary{0x00, 0x01},
			"ObjectId":    ObjectId{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00},
			"Bool":        Bool(true),
			"UTCDateTime": UTCDateTime(123),
			"Javascript":  Javascript("foo"),
			"Symbol":      Symbol("foo"),
			"Int32":       Int32(123),
			"Timestamp":   Timestamp(123),
			"Int64":       Int64(123),
		},
	}

	// Float
	var floatTest float64
	ok, err := src.Reach(&floatTest, "foo", "Float")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'Float'.")
	}
	if floatTest != 123.123 {
		t.Fatal(floatTest)
	}

	// String
	var stringTest string
	ok, err = src.Reach(&stringTest, "foo", "String")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'String'.")
	}
	if stringTest != "foo" {
		t.Fatal(stringTest)
	}

	// Binary
	var binaryTest []byte
	ok, err = src.Reach(&binaryTest, "foo", "Binary")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'Binary'.")
	}
	if !reflect.DeepEqual(binaryTest, []byte{0x00, 0x01}) {
		t.Fatal(binaryTest)
	}

	// ObjectId
	var objectIdTest []byte
	ok, err = src.Reach(&objectIdTest, "foo", "ObjectId")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'ObjectId.")
	}
	if !reflect.DeepEqual(objectIdTest, []byte{0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) {
		t.Fatal(objectIdTest)
	}

	// Bool
	var boolTest bool
	ok, err = src.Reach(&boolTest, "foo", "Bool")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'Bool'.")
	}
	if boolTest != true {
		t.Fatal(boolTest)
	}

	// UTCDateTime
	var uTest0 int64
	ok, err = src.Reach(&uTest0, "foo", "UTCDateTime")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'UTCDateTime'.")
	}
	if uTest0 != 123 {
		t.Fatal(uTest0)
	}

	// UTCDateTime
	var uTest1 time.Time
	ok, err = src.Reach(&uTest1, "foo", "UTCDateTime")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'UTCDateTime'.")
	}
	if uTest1.UnixNano() != 123*1e3 {
		t.Fatal(uTest1.UnixNano())
	}

	// Javascript
	var jsTest string
	ok, err = src.Reach(&jsTest, "foo", "Javascript")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'Javascript'.")
	}
	if jsTest != "foo" {
		t.Fatal(jsTest)
	}

	// Symbol
	var symbolTest string
	ok, err = src.Reach(&symbolTest, "foo", "Symbol")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'Symbol'.")
	}
	if symbolTest != "foo" {
		t.Fatal(symbolTest)
	}

	// Int32
	var int32Test0 int32
	ok, err = src.Reach(&int32Test0, "foo", "Int32")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'Int32'.")
	}
	if int32Test0 != 123 {
		t.Fatal(int32Test0)
	}

	// Int32
	var int32Test1 int64
	ok, err = src.Reach(&int32Test1, "foo", "Int32")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'Int32'.")
	}
	if int32Test1 != 123 {
		t.Fatal(int32Test1)
	}

	// Timestamp
	var tsTest0 int64
	ok, err = src.Reach(&tsTest0, "foo", "UTCDateTime")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'UTCDateTime.")
	}
	if tsTest0 != 123 {
		t.Fatal(tsTest0)
	}

	// Timestamp
	var tsTest1 time.Time
	ok, err = src.Reach(&tsTest1, "foo", "UTCDateTime")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'UTCDateTime'.")
	}
	if tsTest1.UnixNano() != 123*1e3 {
		t.Fatal(tsTest1.UnixNano())
	}

	// Int64
	var int64Test int32
	ok, err = src.Reach(&int64Test, "foo", "Int32")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Expected to find 'Int32'.")
	}
	if int64Test != 123 {
		t.Fatal(int64Test)
	}
}
