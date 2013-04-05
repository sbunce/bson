// Copyright 2013 Seth Bunce. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package bson

import (
	"reflect"
	"testing"
)

// Convert Slice -> bson -> Slice then compare Slices.
var sliceTest = []Slice{
	Slice{{"Float", Float(123.123)}},
	Slice{{"String", String("123")}},
	Slice{{"embed", Slice{{"foo", String("bar")}}}},
	Slice{{"Array", Array{String("foo"), String("bar")}}},
	Slice{{"Binary", Binary{0x00, 0x01}}},
	Slice{{"Undefined", Undefined{}}},
	Slice{{"ObjectId", ObjectId{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00}}},
	Slice{{"Bool", Bool(true)}, {"false", Bool(false)}},
	Slice{{"UTCDateTime", UTCDateTime(123)}},
	Slice{{"Null", Null{}}},
	Slice{{"Regexp", Regexp{"foo", "bar"}}},
	Slice{{"DBPointer", DBPointer{"foo", ObjectId{0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}}},
	Slice{{"Javascript", Javascript("foo")}},
	Slice{{"Symbol", Symbol("foo")}},
	Slice{{"JavascriptScope", JavascriptScope{"foo", Map{"bar": String("baz")}}}},
	Slice{{"Int32", Int32(123)}},
	Slice{{"Timestamp", Timestamp(123)}},
	Slice{{"Int64", Int64(123)}},
	Slice{{"MinKey", MinKey{}}},
	Slice{{"MaxKey", MaxKey{}}},
}

func TestSlice(t *testing.T) {
	for _, d0 := range sliceTest {
		bs, err := d0.Encode()
		if err != nil {
			t.Fatal(err, d0)
		}
		d1, err := bs.Slice()
		if err != nil {
			t.Fatal(err, d0, d1)
		}
		if !reflect.DeepEqual(d0, d1) {
			t.Fatal(d0, d1)
		}
	}
}

func TestSliceNoNest(t *testing.T) {
	nest := Slice{{"abc", Int64(123)}}
	src := Slice{
		{"foo",  String("bar")},
		{"nest", nest},
	}
	exp := Slice{
		{"foo",  String("bar")},
		{"nest", nest.MustEncode()},
	}
	bs, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	dst, err := bs.SliceNoNest()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dst, exp) {
		t.Fatal(dst)
	}
}
