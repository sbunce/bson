// Copyright 2013 Seth Bunce. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package bson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// Map, Slice, and BSON conform to this.
type Doc interface {
	Encode() (BSON, error) // Return raw BSON.
	MustEncode() BSON      // Return raw BSON. Panic if encode error.
}

// BSON doc type.
// This is the most commonly used doc type.
type Map map[string]interface{}

// BSON doc type.
// This is used when order must be preserved.
type Slice []Pair

// Element of Slice.
type Pair struct {
	Key string
	Val interface{}
}

// BSON doc type.
// Raw BSON document.
type BSON []byte

// Allows using a pre-encoded value.
func (this BSON) Encode() (BSON, error) {
	return this, nil
}

// Allows using a pre-encoded value.
func (this BSON) MustEncode() BSON {
	return this
}

// Return JSON.
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

// Decode BSON to Map.
func (this BSON) Map() (Map, error) {
	return ReadMap(bytes.NewBuffer(this))
}

// Decode BSON to Map, but don't decode nested docs.
func (this BSON) MapNoNest() (Map, error) {
	return ReadMapNoNest(bytes.NewBuffer(this))
}

// Decode BSON to Slice.
func (this BSON) Slice() (Slice, error) {
	return ReadSlice(bytes.NewBuffer(this))
}

// Decode BSON to Slice, but don't decode nested docs.
func (this BSON) SliceNoNest() (Slice, error) {
	return ReadSliceNoNest(bytes.NewBuffer(this))
}

// Return raw bson.
func (this Map) Encode() (BSON, error) {
	b, err := encodeMap("", this)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Return raw bson. Panic on error.
func (this Map) MustEncode() BSON {
	b, err := encodeMap("", this)
	if err != nil {
		panic(err)
	}
	return b
}

// Pretty print bson value.
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

// Pretty printer.
func (this Map) String() string {
	wr := bytes.NewBuffer(nil)
	fmt.Fprint(wr, "Map[")
	for k, v := range this {
		fmt.Fprintf(wr, "%v: %v", k, print(v))
	}
	fmt.Fprintf(wr, "]")
	return wr.String()
}

// Return raw bson.
func (this Slice) Encode() (BSON, error) {
	b, err := encodeSlice("", this)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Return raw bson. Panic on error.
func (this Slice) MustEncode() BSON {
	b, err := encodeSlice("", this)
	if err != nil {
		panic(err)
	}
	return b
}

// Pretty printer.
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
