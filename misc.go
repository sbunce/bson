// Copyright 2013 Seth Bunce. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package bson

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"time"
)

// Last ID used.
// Use requestID function to get next.
var lastCount int32

// Concatenate name to path.
func catpath(path, name string) string {
	if path == "" {
		return name
	}
	return strings.Join([]string{path, name}, ".")
}

// Indirect all interfaces/pointers.
func indirect(v reflect.Value) reflect.Value {
loop:
	for {
		switch v.Kind() {
		case reflect.Interface, reflect.Ptr:
			v = v.Elem()
		default:
			break loop
		}
	}
	return v
}

// Indirect and allocate if needed.
// If nil interface then allocate Map.
func indirectAlloc(v reflect.Value) reflect.Value {
loop:
	for {
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				// If nil interface default to Map.
				v.Set(reflect.MakeMap(reflect.TypeOf(Map{})))
			}
			v = v.Elem()
		case reflect.Ptr:
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		case reflect.Map:
			if v.IsNil() {
				v.Set(reflect.MakeMap(v.Type()))
			}
			break loop
		case reflect.Slice:
			if v.IsNil() {
				v.Set(reflect.MakeSlice(v.Type(), v.Len(), 0))
			}
			break loop
		default:
			break loop
		}
	}
	return v
}

// Create unique incrementing ObjectId.
// Same format used by MongoDB.
// 
//   +---+---+---+---+---+---+---+---+---+---+---+---+
//   |       A       |     B     |   C   |     D     |
//   +---+---+---+---+---+---+---+---+---+---+---+---+
//     0   1   2   3   4   5   6   7   8   9  10  11
//   A = unix time (big endian), B = machine ID (first 3 bytes of md5 host name),
//   C = PID, D = incrementing counter (big endian)
func NewObjectId() (ObjectId, error) {
	// oid := make(ObjectId, 0)
	buf := bytes.NewBuffer(make([]byte, 0, 12))
	if err := binary.Write(buf, binary.BigEndian, int32(time.Now().Unix())); err != nil {
		return nil, err
	}
	name, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	hash := md5.New()
	if _, err := hash.Write([]byte(name)); err != nil {
		return nil, err
	}
	if _, err := buf.Write(hash.Sum(nil)[:3]); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, int16(os.Getpid())); err != nil {
		return nil, err
	}
	// Wrap at 2^24 because we only use 3 bytes.
	cnt := atomic.AddInt32(&lastCount, 1) % 16777215
	cntbuf := make([]byte, 4)
	binary.BigEndian.PutUint32(cntbuf, uint32(cnt))
	if _, err := buf.Write(cntbuf[1:]); err != nil {
		return nil, err
	}
	return ObjectId(buf.Bytes()), nil
}
