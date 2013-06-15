// Copyright 2013 Seth Bunce. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package bson

import (
	"bytes"
	"testing"
)

func TestNewObjectId(t *testing.T) {
	oid0, err := NewObjectId()
	if err != nil {
		t.Fatal(err)
	}
	if len(oid0) != 12 {
		t.Fatal(len(oid0))
	}
	oid1, err := NewObjectId()
	if err != nil {
		t.Fatal(err)
	}
	// ObjectIds should be increasing.
	if bytes.Compare(oid0, oid1) >= 0 {
		t.Fatal()
	}
}

func TestReadOne(t *testing.T) {
	foo := Map{"abc": "cba"}
	bar := Map{"123": "321"}
	foob := foo.MustEncode()
	barb := bar.MustEncode()
	rd := bytes.NewBuffer(append(foob, barb...))
	buf, err := ReadOne(rd)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, foob) != 0 {
		t.Fatal()
	}
	buf, err = ReadOne(rd)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(buf, barb) != 0 {
		t.Fatal()
	}
}
