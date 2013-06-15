// Copyright 2013 Seth Bunce. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package bson

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
)

// maxDocLen is max supported size (bytes) of a document.
const maxDocLen = 64 * 1024 * 1024

// ReadOne BSON document.
func ReadOne(rd io.Reader) (BSON, error) {
	// Read length of document.
	docLen, err := readInt32(rd)
	if err != nil {
		return nil, err
	}

	// Sanity check length.
	if docLen > maxDocLen {
		return nil, errors.New("Doc exceeded maximum size.")
	}

	// Read the document.
	buf := make([]byte, int(docLen))
	binary.LittleEndian.PutUint32(buf, uint32(docLen))
	if _, err := io.ReadFull(rd, buf[4:]); err != nil {
		return nil, err
	}

	return buf, nil
}

// ReadMap reads one Map.
func ReadMap(rd io.Reader) (m Map, err error) {
	// Just in case of programming mistake. Not intentionally used.
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()

	return decodeMap(rd, "", true)
}

// ReadMapNoNest reads one Map, but doesn't decode nested documents.
func ReadMapNoNest(rd io.Reader) (m Map, err error) {
	// Just in case of programming mistake. Not intentionally used.
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()

	return decodeMap(rd, "", false)
}

// ReadSlice reads one Slice, but doesn't decode nested documents.
func ReadSlice(rd io.Reader) (s Slice, err error) {
	// Just in case of programming mistake. Not intentionally used.
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()

	return decodeSlice(rd, "", true)
}

// ReadSliceNoNest reads one Slice, but doesn't decode nested documents.
func ReadSliceNoNest(rd io.Reader) (s Slice, err error) {
	// Just in case of programming mistake. Not intentionally used.
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()

	return decodeSlice(rd, "", false)
}

// decodeMap decodes to a Map. The path is used to keep track of where we've
// recursed to in the document. If nest is true then nested documents are
// decoded.
func decodeMap(rdTmp io.Reader, path string, nest bool) (Map, error) {
	// Read doc length.
	docLen, err := readInt32(rdTmp)
	if err != nil {
		return nil, err
	}
	if docLen > maxDocLen {
		return nil, errors.New("Doc exceeded maximum size.")
	}
	rd := bufio.NewReader(io.LimitReader(rdTmp, int64(docLen-4)))

	// Read doc.
	dst := Map{}
	for {
		eType, err := rd.ReadByte()
		if err != nil {
			return nil, err
		}
		switch eType {
		case 0x00:
			return dst, nil
		case _FLOATING_POINT:
			name, val, err := decodeFloat(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _STRING:
			name, val, err := decodeString(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _EMBEDDED_DOCUMENT:
			// While decoding Map default to Map.
			// name
			name, err := readCstring(rd)
			if err != nil {
				return nil, err
			}
			if !nest {
				bs, err := ReadOne(rd)
				if err != nil {
					return nil, err
				}
				dst[name] = bs
			} else {
				// value
				val, err := decodeMap(rd, catpath(path, name), true)
				if err != nil {
					return nil, err
				}
				dst[name] = val
			}
		case _ARRAY:
			name, val, err := decodeArray(rd, path)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _BINARY_DATA:
			name, val, err := decodeBinary(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _UNDEFINED:
			name, val, err := decodeUndefined(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _OBJECT_ID:
			name, val, err := decodeObjectId(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _BOOLEAN:
			name, val, err := decodeBool(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _UTC_DATETIME:
			name, val, err := decodeUTCDateTime(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _NULL_VALUE:
			name, val, err := decodeNull(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _REGEXP:
			name, val, err := decodeRegexp(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _DBPOINTER:
			name, val, err := decodeDBPointer(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _JAVASCRIPT:
			name, val, err := decodeJavascript(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _SYMBOL:
			name, val, err := decodeSymbol(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _JAVASCRIPT_SCOPE:
			name, val, err := decodeJavascriptScope(rd, path)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _32BIT_INTEGER:
			name, val, err := decodeInt32(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _TIMESTAMP:
			name, val, err := decodeTimestamp(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _64BIT_INTEGER:
			name, val, err := decodeInt64(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _MIN_KEY:
			name, val, err := decodeMinKey(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		case _MAX_KEY:
			name, val, err := decodeMaxKey(rd)
			if err != nil {
				return nil, err
			}
			dst[name] = val
		default:
			return nil, fmt.Errorf("Unsupported type '%X'.", eType)
		}
	}
	return nil, nil
}

// decodeSlice decodes to a Slice. The path is used to keep track of where we've
// recursed to in the document. If nest is true then nested documents are
// decoded.
func decodeSlice(rdTmp io.Reader, path string, nest bool) (Slice, error) {
	// Read doc length.
	docLen, err := readInt32(rdTmp)
	if err != nil {
		return nil, err
	}
	if docLen > maxDocLen {
		return nil, errors.New("Doc exceeded maximum size.")
	}
	rd := bufio.NewReader(io.LimitReader(rdTmp, int64(docLen-4)))

	// Read doc.
	dst := Slice{}
	for {
		eType, err := rd.ReadByte()
		if err != nil {
			return nil, err
		}
		switch eType {
		case 0x00:
			return dst, nil
		case _FLOATING_POINT:
			name, val, err := decodeFloat(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _STRING:
			name, val, err := decodeString(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _EMBEDDED_DOCUMENT:
			// While decoding Slice default to Slice.
			// name
			name, err := readCstring(rd)
			if err != nil {
				return nil, err
			}
			if !nest {
				bs, err := ReadOne(rd)
				if err != nil {
					return nil, err
				}
				dst = append(dst, Pair{Key: name, Val: bs})
			} else {
				// value
				val, err := decodeSlice(rd, catpath(path, name), true)
				if err != nil {
					return nil, err
				}
				dst = append(dst, Pair{Key: name, Val: val})
			}
		case _ARRAY:
			name, val, err := decodeArray(rd, path)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _BINARY_DATA:
			name, val, err := decodeBinary(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _UNDEFINED:
			name, val, err := decodeUndefined(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _OBJECT_ID:
			name, val, err := decodeObjectId(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _BOOLEAN:
			name, val, err := decodeBool(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _UTC_DATETIME:
			name, val, err := decodeUTCDateTime(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _NULL_VALUE:
			name, val, err := decodeNull(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _REGEXP:
			name, val, err := decodeRegexp(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _DBPOINTER:
			name, val, err := decodeDBPointer(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _JAVASCRIPT:
			name, val, err := decodeJavascript(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _SYMBOL:
			name, val, err := decodeSymbol(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _JAVASCRIPT_SCOPE:
			name, val, err := decodeJavascriptScope(rd, path)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _32BIT_INTEGER:
			name, val, err := decodeInt32(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _TIMESTAMP:
			name, val, err := decodeTimestamp(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _64BIT_INTEGER:
			name, val, err := decodeInt64(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _MIN_KEY:
			name, val, err := decodeMinKey(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		case _MAX_KEY:
			name, val, err := decodeMaxKey(rd)
			if err != nil {
				return nil, err
			}
			dst = append(dst, Pair{Key: name, Val: val})
		default:
			return nil, fmt.Errorf("Unsupported type '%X'.", eType)
		}
	}
	return nil, nil
}

// decodeArray decodes a BSON Array element.
func decodeArray(rd *bufio.Reader, path string) (string, Array, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", nil, err
	}

	// value
	doc, err := decodeMap(rd, path, true)
	if err != nil {
		return "", nil, err
	}

	// BSON index names may not be ordered. Sort.
	ns := make([]string, 0, len(doc))
	for name, _ := range doc {
		ns = append(ns, name)
	}
	sort.Strings(ns)

	// Build slice.
	slice := reflect.MakeSlice(reflect.TypeOf(Array{}), 0, len(ns))
	for _, name := range ns {
		slice = reflect.Append(slice, reflect.ValueOf(doc[name]))
	}
	return name, slice.Interface().(Array), nil
}

// decodeBinary decodes BSON Binary element.
func decodeBinary(rd *bufio.Reader) (string, Binary, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", nil, err
	}

	// value
	dataLen, err := readInt32(rd)
	if err != nil {
		return "", nil, err
	}

	// discard subtype
	_, err = rd.ReadByte()
	if err != nil {
		return "", nil, err
	}
	b := make([]byte, dataLen)
	_, err = io.ReadFull(rd, b)
	if err != nil {
		return "", nil, err
	}
	return name, Binary(b), nil
}

// decodeBool decodes BSON Bool element.
func decodeBool(rd *bufio.Reader) (string, Bool, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", false, err
	}

	// value
	b, err := rd.ReadByte()
	if err != nil {
		return "", false, err
	}
	return name, Bool(b == 0x01), nil
}

// decodeDBPointer decodes BSON DBPointer.
func decodeDBPointer(rd *bufio.Reader) (string, DBPointer, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", DBPointer{}, err
	}

	// value
	Name, err := readString(rd)
	if err != nil {
		return "", DBPointer{}, err
	}
	b := make([]byte, 12)
	_, err = io.ReadFull(rd, b)
	if err != nil {
		return "", DBPointer{}, err
	}
	return name, DBPointer{Name: Name, ObjectId: ObjectId(b)}, nil
}

// decodeFloat decodes BSON Float element.
func decodeFloat(rd *bufio.Reader) (string, Float, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", Float(0), err
	}

	// value
	b := make([]byte, 8)
	_, err = io.ReadFull(rd, b)
	if err != nil {
		return "", Float(0), err
	}
	var u uint64
	u += uint64(b[7]) << 56
	u += uint64(b[6]) << 48
	u += uint64(b[5]) << 40
	u += uint64(b[4]) << 32
	u += uint64(b[3]) << 24
	u += uint64(b[2]) << 16
	u += uint64(b[1]) << 8
	u += uint64(b[0])
	return name, Float(math.Float64frombits(u)), nil
}

// decodeInt32 decodes BSON Int32 element.
func decodeInt32(rd *bufio.Reader) (string, Int32, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", 0, err
	}

	// value
	i32, err := readInt32(rd)
	if err != nil {
		return "", 0, err
	}
	return name, Int32(i32), nil
}

// decodeInt64 decodes BSON Int64 element.
func decodeInt64(rd *bufio.Reader) (string, Int64, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", 0, err
	}

	// value
	i64, err := readInt64(rd)
	if err != nil {
		return "", 0, err
	}
	return name, Int64(i64), nil
}

// decodeJavascript decodes BSON Javascript element.
func decodeJavascript(rd *bufio.Reader) (string, Javascript, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", "", err
	}

	// value
	s, err := readString(rd)
	if err != nil {
		return "", "", err
	}
	return name, Javascript(s), nil
}

// decodeJavascriptScope decodes BSON JavascriptScope element.
func decodeJavascriptScope(rd *bufio.Reader, path string) (string, JavascriptScope, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", JavascriptScope{}, err
	}

	// value
	_, err = readInt32(rd)
	if err != nil {
		return "", JavascriptScope{}, err
	}
	js, err := readString(rd)
	if err != nil {
		return "", JavascriptScope{}, err
	}
	m, err := decodeMap(rd, "", true)
	if err != nil {
		return "", JavascriptScope{}, err
	}
	return name, JavascriptScope{Javascript: js, Scope: m}, nil
}

// decodeMaxKey decodes BSON MaxKey element.
func decodeMaxKey(rd *bufio.Reader) (string, MaxKey, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", MaxKey{}, err
	}
	return name, MaxKey{}, nil
}

// decodeMinKey decodes BSON JavascriptMinKey element.
func decodeMinKey(rd *bufio.Reader) (string, MinKey, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", MinKey{}, err
	}
	return name, MinKey{}, nil
}

// decodeNull decodes BSON Null element.
func decodeNull(rd *bufio.Reader) (string, Null, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", Null{}, err
	}
	return name, Null{}, nil
}

// decodeObjectId decodes BSON ObjectId element.
func decodeObjectId(rd *bufio.Reader) (string, ObjectId, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", nil, err
	}

	// value
	b := make([]byte, 12)
	_, err = io.ReadFull(rd, b)
	if err != nil {
		return "", nil, err
	}
	return name, ObjectId(b), nil
}

// decodeRegexp decodes BSON Regexp element.
func decodeRegexp(rd *bufio.Reader) (string, Regexp, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", Regexp{}, err
	}

	// pattern
	pattern, err := readCstring(rd)
	if err != nil {
		return "", Regexp{}, err
	}

	// options
	options, err := readCstring(rd)
	if err != nil {
		return "", Regexp{}, err
	}
	return name, Regexp{Pattern: pattern, Options: options}, nil
}

// decodeString decodes BSON String element.
func decodeString(rd *bufio.Reader) (string, String, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", "", err
	}

	// value
	s, err := readString(rd)
	if err != nil {
		return "", "", err
	}
	return name, String(s), nil
}

// decodeSymbol decodes BSON Symbol element.
func decodeSymbol(rd *bufio.Reader) (string, Symbol, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", "", err
	}

	// value
	s, err := readString(rd)
	if err != nil {
		return "", "", err
	}
	return name, Symbol(s), nil
}

// decodeTimestamp decodes BSON Timestamp element.
func decodeTimestamp(rd *bufio.Reader) (string, Timestamp, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", 0, err
	}

	// value
	i64, err := readInt64(rd)
	if err != nil {
		return "", 0, err
	}
	return name, Timestamp(i64), nil
}

// decodeUndefined decodes BSON Undefined element.
func decodeUndefined(rd *bufio.Reader) (string, Undefined, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", Undefined{}, err
	}
	return name, Undefined{}, nil
}

// decodeUTCDateTime decodes BSON UTCDateTime element.
func decodeUTCDateTime(rd *bufio.Reader) (string, UTCDateTime, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", 0, err
	}

	// value
	i64, err := readInt64(rd)
	if err != nil {
		return "", 0, err
	}
	return name, UTCDateTime(i64), nil
}

// readCString reads one BSON C string. This is not a BSON element.
func readCstring(rd *bufio.Reader) (string, error) {
	s, err := rd.ReadString(0x00)
	if err != nil {
		return "", err
	}
	return s[:len(s)-1], nil
}

// readBSONInt32 reads one int32. This is not a BSON element.
func readInt32(rd io.Reader) (int32, error) {
	var i int32
	if err := binary.Read(rd, binary.LittleEndian, &i); err != nil {
		return 0, err
	}
	return i, nil
}

// readInt64 reads one int64. This is not a BSON element.
func readInt64(rd io.Reader) (int64, error) {
	var i int64
	if err := binary.Read(rd, binary.LittleEndian, &i); err != nil {
		return 0, err
	}
	return i, nil
}

// readString reads one string. This is not a BSON element.
func readString(rd *bufio.Reader) (string, error) {
	// Read string length.
	var sLen int32
	if err := binary.Read(rd, binary.LittleEndian, &sLen); err != nil {
		return "", err
	}
	if sLen == 0 {
		return "", nil
	}

	// Read string.
	b := make([]byte, sLen)
	if _, err := io.ReadFull(rd, b); err != nil {
		return "", err
	}
	return string(b[:len(b)-1]), nil
}
