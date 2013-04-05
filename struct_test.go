// Copyright 2013 Seth Bunce. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package bson

import (
	"reflect"
	"testing"
)

type structTest struct {
	src interface{} // Encode this.
	dst interface{} // Decode to this.
	exp interface{} // Expect dst will equal this.
}

// Test struct tags.
type tags struct {
	Ignore     string `bson:"-"`
	Rename     string `bson:"rename_ok"`
	OmitRename string `bson:"omitrename_ok,omitempty"`
	Omit       string `bson:",omitempty"`
}

// Test that unexported field is ignored.
type unexport struct {
	foo string
}

var structTests = []structTest{
	// Struct tags. Encode with omit field empty.
	structTest{
		src: tags{
			Ignore:     "foo",
			Rename:     "bar",
			OmitRename: "",
			Omit:       "",       
		},
		exp: Map{
			"rename_ok": String("bar"),
		},
	},
	// Struct tags. Encode with omit fields not empty.
	structTest{
		src: tags{
			Ignore:     "foo",
			Rename:     "bar",
			OmitRename: "123",
			Omit:       "321",       
		},
		exp: Map{
			"rename_ok":     String("bar"),
			"omitrename_ok": String("123"),
			"Omit":          String("321"),
		},
	},
	// Unexported field.
	structTest{
		src: unexport{
			foo: "bar",
		},
		exp: Map{},
	},
}

func TestStruct(t *testing.T) {
	for _, st := range structTests {
		bs, err := EncodeStruct(st.src)
		if err != nil {
			t.Fatal(err, st.src)
		}
		st.dst, err = bs.Map()
		if err != nil {
			t.Fatal(err, st.dst, st.exp)
		}
		if !reflect.DeepEqual(st.dst, st.exp) {
			t.Fatal(indirect(reflect.ValueOf(st.dst)).Interface(), st.exp)
		}
	}
}
