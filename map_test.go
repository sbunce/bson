// Copyright 2013 Seth Bunce. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package bson

import (
	"reflect"
	"testing"
)

// Convert Map -> bson -> Map then compare Maps.
var mapTest = []Map{
	Map{"Float": Float(123.123)},
	Map{"String": String("123")},
	Map{"embed": Map{"foo": String("bar")}},
	Map{"Array": Array{String("foo"), String("bar")}},
	Map{"Binary": Binary{0x00, 0x01}},
	Map{"Undefined": Undefined{}},
	Map{"ObjectId": ObjectId{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00}},
	Map{"Bool": Bool(true), "false": Bool(false)},
	Map{"UTCDateTime": UTCDateTime(123)},
	Map{"Null": Null{}},
	Map{"Regexp": Regexp{"foo", "bar"}},
	Map{"DBPointer": DBPointer{"foo", ObjectId{0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}},
	Map{"Javascript": Javascript("foo")},
	Map{"Symbol": Symbol("foo")},
	Map{"JavascriptScope": JavascriptScope{"foo", Map{"bar": String("baz")}}},
	Map{"Int32": Int32(123)},
	Map{"Timestamp": Timestamp(123)},
	Map{"Int64": Int64(123)},
	Map{"MinKey": MinKey{}},
	Map{"MaxKey": MaxKey{}},
}

func TestMap(t *testing.T) {
	for _, d0 := range mapTest {
		bs, err := d0.Encode()
		if err != nil {
			t.Fatal(err, d0)
		}
		d1, err := bs.Map()
		if err != nil {
			t.Fatal(err, d0, d1)
		}
		if !reflect.DeepEqual(d0, d1) {
			t.Fatal(d0, d1)
		}
	}
}

func TestMapNoNest(t *testing.T) {
	nest := Slice{{"abc", Int64(123)}}
	src := Map{
		"foo":  String("bar"),
		"nest": nest,
	}
	exp := Map{
		"foo":  String("bar"),
		"nest": nest.MustEncode(),
	}
	bs, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	dst, err := bs.MapNoNest()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dst, exp) {
		t.Fatal(dst)
	}
}
