// Copyright 2013 Seth Bunce. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package bson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// Doc is a BSON document. Map, Slice, and BSON conform to this.
type Doc interface {
	// Encode returns raw BSON.
	Encode() (BSON, error)

	// MustEncode returns raw BSON and panics upon error.
	MustEncode() BSON
}

// Map is a BSON document type. This should be used when the order of encoded
// elements does not matter.
type Map map[string]interface{}

// Slice is a BSON document type. This should be used when the order of encoded
// elements matters.
type Slice []Pair

// Pair is a element of a Slice. Typically the field names are not specifie
// when using this type.
type Pair struct {
	Key string
	Val interface{}
}

// BSON is a raw BSON document. This is provided so that it's not necessary to
// decode BSON to include it in another document.
type BSON []byte

// Encode returns the raw BSON document.
func (this BSON) Encode() (BSON, error) {
	return this, nil
}

// MustEncode returns the raw BSON document. Never panics.
func (this BSON) MustEncode() BSON {
	return this
}

// JSON transcodes the BSON document to JSON.
func (this BSON) JSON() (string, error) {
	m, err := this.Map()
	if err != nil {
		return "", err
	}
	j, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(j), nil
}

// Map decodes the BSON to a Map. Order of encded elements is not preserved.
func (this BSON) Map() (Map, error) {
	return ReadMap(bytes.NewBuffer(this))
}

// MapNoNest decodes the BSON to a Map, but leaves nested documents encoded as
// BSON. This is useful when it's not necessary to decode the whole document.
func (this BSON) MapNoNest() (Map, error) {
	return ReadMapNoNest(bytes.NewBuffer(this))
}

// Slice decodes the BSON to a Slice. Order of encoded elements is preserved.
func (this BSON) Slice() (Slice, error) {
	return ReadSlice(bytes.NewBuffer(this))
}

// Decode BSON to Slice, but don't decode nested docs. This is useful when it's
// not necessary to decode the whole document.
func (this BSON) SliceNoNest() (Slice, error) {
	return ReadSliceNoNest(bytes.NewBuffer(this))
}

// Encode Map to BSON.
func (this Map) Encode() (BSON, error) {
	b, err := encodeMap("", this)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// MustEncode panics if Map cannot be encoded to BSON.
func (this Map) MustEncode() BSON {
	b, err := encodeMap("", this)
	if err != nil {
		panic(err)
	}
	return b
}

// print pretty-prints BSON value.
func print(v interface{}) string {
	switch vt := v.(type) {
	case Map:
		return vt.String()
	case Slice:
		return vt.String()
	case BSON:
		return fmt.Sprintf("BSON(%v)", vt)
	case Float:
		return fmt.Sprintf("Float(%v)", vt)
	case String:
		return fmt.Sprintf("String(%v)", vt)
	case Array:
		wr := bytes.NewBuffer(nil)
		fmt.Fprint(wr, "Array([")
		for i, vtv := range vt {
			fmt.Fprint(wr, print(vtv))
			if i != len(vt)-1 {
				fmt.Fprint(wr, " ")
			}
		}
		fmt.Fprintf(wr, "])")
		return wr.String()
	case Binary:
		return fmt.Sprintf("Binary(%v)", vt)
	case Undefined:
		return "Undefined()"
	case ObjectId:
		return fmt.Sprintf("ObjectId(%v)", vt)
	case Bool:
		return fmt.Sprintf("Bool(%v)", vt)
	case UTCDateTime:
		return fmt.Sprintf("UTCDateTime(%v)", time.Unix(0, int64(vt)*1000*1000))
	case Null:
		return "Null()"
	case Regexp:
		return fmt.Sprintf("Regexp(Pattern(%v) Options(%v))", vt.Pattern, vt.Options)
	case DBPointer:
		return fmt.Sprintf("DBPointer(Name(%v) ObjectId(%v))", vt.Name, vt.ObjectId)
	case Javascript:
		return fmt.Sprintf("Javascript(%v)", vt)
	case Symbol:
		return fmt.Sprintf("Symbol(%v)", vt)
	case JavascriptScope:
		return fmt.Sprintf("JavascriptScope(Javascript(%v) Scope(%v))", vt.Javascript, vt.Scope)
	case Int32:
		return fmt.Sprintf("Int32(%v)", vt)
	case Timestamp:
		return fmt.Sprintf("Timestamp(%v)", vt)
	case Int64:
		return fmt.Sprintf("Int64(%v)", vt)
	case MinKey:
		return "MinKey()"
	case MaxKey:
		return "MaxKey()"
	}
	return fmt.Sprint(v)
}

// String pretty prints the Map with BSON types.
func (this Map) String() string {
	wr := bytes.NewBuffer(nil)
	fmt.Fprint(wr, "Map[")
	for k, v := range this {
		fmt.Fprintf(wr, "%v: %v", k, print(v))
	}
	fmt.Fprintf(wr, "]")
	return wr.String()
}

// Encode Slice to BSON.
func (this Slice) Encode() (BSON, error) {
	b, err := encodeSlice("", this)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// MustEncode panics if Slice cannot be encoded to BSON.
func (this Slice) MustEncode() BSON {
	b, err := encodeSlice("", this)
	if err != nil {
		panic(err)
	}
	return b
}

// String pretty prints the Slice with BSON types.
func (this Slice) String() string {
	wr := bytes.NewBuffer(nil)
	fmt.Fprint(wr, "Slice[")
	for i, v := range this {
		fmt.Fprintf(wr, "%v: %v", v.Key, print(v.Val))
		if i != len(this)-1 {
			fmt.Fprint(wr, " ")
		}
	}
	fmt.Fprintf(wr, "]")
	return wr.String()
}
