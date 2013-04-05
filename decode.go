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

// Maximum document size. Set at max of MongoDB.
const maxDocLen = 16 * 1024 * 1024

// Read one raw bson doc and return it. Do not parse.
func ReadOne(rd io.Reader) (BSON, error) {
	docLen, err := readInt32(rd)
	if err != nil {
		return nil, err
	}
	if docLen > maxDocLen {
		return nil, errors.New("Doc exceeded maximum size.")
	}
	buf := make([]byte, int(docLen))
	binary.LittleEndian.PutUint32(buf, uint32(docLen))
	if _, err := io.ReadFull(rd, buf[4:]); err != nil {
		return nil, err
	}
	return buf, nil
}

// Read one map.
func ReadMap(rd io.Reader) (m Map, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()
	return decodeMap(rd, "", true)
}

// Read one map, don't decode nested docs.
func ReadMapNoNest(rd io.Reader) (m Map, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()
	return decodeMap(rd, "", false)
}

// Read one Slice.
func ReadSlice(rd io.Reader) (s Slice, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()
	return decodeSlice(rd, "", true)
}

// Read one Slice, don't decode nested docs.
func ReadSliceNoNest(rd io.Reader) (s Slice, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()
	return decodeSlice(rd, "", false)
}

// Decode to Map.
// Path is used to keep track of errors.
// If nest is true then documents are recursively decoded.
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

// Decode to Slice.
// Path is used to keep track of errors.
// If nest is true then documents are recursively decoded.
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

func decodeMaxKey(rd *bufio.Reader) (string, MaxKey, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", MaxKey{}, err
	}
	return name, MaxKey{}, nil
}

func decodeMinKey(rd *bufio.Reader) (string, MinKey, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", MinKey{}, err
	}
	return name, MinKey{}, nil
}

func decodeNull(rd *bufio.Reader) (string, Null, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", Null{}, err
	}
	return name, Null{}, nil
}

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

func decodeUndefined(rd *bufio.Reader) (string, Undefined, error) {
	// name
	name, err := readCstring(rd)
	if err != nil {
		return "", Undefined{}, err
	}
	return name, Undefined{}, nil
}

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

// Read BSON cstring.
func readCstring(rd *bufio.Reader) (string, error) {
	s, err := rd.ReadString(0x00)
	if err != nil {
		return "", err
	}
	return s[:len(s)-1], nil
}

// Read BSON int32.
func readInt32(rd io.Reader) (int32, error) {
	var i int32
	if err := binary.Read(rd, binary.LittleEndian, &i); err != nil {
		return 0, err
	}
	return i, nil
}

// Read BSON int64.
func readInt64(rd io.Reader) (int64, error) {
	var i int64
	if err := binary.Read(rd, binary.LittleEndian, &i); err != nil {
		return 0, err
	}
	return i, nil
}

// Read BSON string.
func readString(rd *bufio.Reader) (string, error) {
	var sLen int32
	if err := binary.Read(rd, binary.LittleEndian, &sLen); err != nil {
		return "", err
	}
	if sLen == 0 {
		return "", nil
	}
	if sLen > maxDocLen {
		// This is a sanity check to make sure we got a reasonable len.
		return "", errors.New(fmt.Sprint("String too large, ", sLen,
			" bytes."))
	}
	b := make([]byte, sLen)
	if _, err := io.ReadFull(rd, b); err != nil {
		return "", err
	}
	return string(b[:len(b)-1]), nil
}
