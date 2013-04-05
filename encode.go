package bson

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Encode bson doc. This func includes path for error reporting.
func encodeMap(path string, m Map) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	// This will be replaced by the size of the doc later.
	if err := binary.Write(buf, binary.LittleEndian, uint32(0)); err != nil {
		return nil, err
	}
	// Encode.
	for name, v := range m {
		if err := encodeVal(buf, catpath(path, name), name, v); err != nil {
			return nil, err
		}
	}
	// End of BSON null byte.
	if err := buf.WriteByte(0x00); err != nil {
		return nil, err
	}
	// Write size of document at start of BSON.
	binary.LittleEndian.PutUint32(buf.Bytes(), uint32(buf.Len()))
	return buf.Bytes(), nil
}

// Encode bson doc. This func includes path for error reporting.
func encodeSlice(path string, s Slice) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	// This will be replaced by the size of the doc later.
	if err := binary.Write(buf, binary.LittleEndian, uint32(0)); err != nil {
		return nil, err
	}
	// Encode.
	for _, pair := range s {
		if err := encodeVal(buf, catpath(path, pair.Key), pair.Key, pair.Val); err != nil {
			return nil, err
		}
	}
	// End of BSON null byte.
	if err := buf.WriteByte(0x00); err != nil {
		return nil, err
	}
	// Write size of document at start of BSON.
	binary.LittleEndian.PutUint32(buf.Bytes(), uint32(buf.Len()))
	return buf.Bytes(), nil
}

// Return raw bson.
func EncodeStruct(src interface{}) (BSON, error) {
	return encodeStruct("", src)
}

// Return raw bson. Panic on error.
func MustEncodeStruct(src interface{}) BSON {
	b, err := encodeStruct("", src)
	if err != nil {
		panic(err)
	}
	return b
}

// Encode bson doc type.
func encodeStruct(path string, src interface{}) ([]byte, error) {
	rv := indirect(reflect.ValueOf(src))
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%v, expected struct.", path)
	}
	buf := bytes.NewBuffer(make([]byte, 0))
	// This will be replaced by the size of the doc later.
	if err := binary.Write(buf, binary.LittleEndian, uint32(0)); err != nil {
		return nil, err
	}
	// Encode.
	for i := 0; i < rv.NumField(); i++ {
		sv := rv.Type().Field(i)
		if sv.PkgPath != "" {
			// Unexported field.
			continue
		}
		name := sv.Name
		fv := rv.Field(i)
		fv = indirect(rv.Field(i))
		if tag := sv.Tag.Get("bson"); tag != "" {
			tok := strings.Split(tag, ",")
			if tok[0] == "-" {
				// Ignore field.
				continue
			}
			if tok[0] != "" {
				// Renamed field.
				name = tok[0]
			}
			if len(tok) == 2 && tok[1] == "omitempty" && isEmptyValue(fv) {
				// Empty field, omitempty true.
				continue
			}
		}
		if err := encodeVal(buf, catpath(path, name), name, fv.Interface()); err != nil {
			return nil, err
		}
	}
	// End of BSON null byte.
	if err := buf.WriteByte(0x00); err != nil {
		return nil, err
	}
	// Write size of document at start of BSON.
	binary.LittleEndian.PutUint32(buf.Bytes(), uint32(buf.Len()))
	return buf.Bytes(), nil
}

func encodeVal(buf *bytes.Buffer, path, name string, src interface{}) error {
	if src == nil {
		return encodeNull(buf, name)
	}
	rvsrc := reflect.ValueOf(src)
	if rvsrc.Kind() == reflect.Ptr && rvsrc.IsNil() {
		return encodeNull(buf, name)
	}
	src = indirect(rvsrc).Interface()

	// Try non-reflect first.
	switch srct := src.(type) {
	case Float:
		return encodeFloat(buf, name, srct)
	case String:
		return encodeString(buf, name, srct)
	case Map:
		return encodeEmbeddedDocument(buf, path, name, srct)
	case Slice:
		return encodeEmbeddedDocument(buf, path, name, srct)
	case BSON:
		_, err := buf.Write(srct)
		return err
	case Array:
		return encodeArray(buf, path, name, srct)
	case Binary:
		return encodeBinary(buf, name, srct)
	case Undefined:
		return encodeUndefined(buf, name)
	case ObjectId:
		return encodeObjectId(buf, path, name, srct)
	case Bool:
		return encodeBool(buf, name, srct)
	case UTCDateTime:
		return encodeUTCDateTime(buf, name, srct)
	case Null:
		return encodeNull(buf, name)
	case Regexp:
		return encodeRegexp(buf, name, srct)
	case DBPointer:
		return encodeDBPointer(buf, path, name, srct)
	case Javascript:
		return encodeJavascript(buf, name, srct)
	case Symbol:
		return encodeSymbol(buf, name, srct)
	case JavascriptScope:
		return encodeJavascriptScope(buf, path, name, srct)
	case Int32:
		return encodeInt32(buf, name, srct)
	case Timestamp:
		return encodeTimestamp(buf, name, srct)
	case Int64:
		return encodeInt64(buf, name, srct)
	case MinKey:
		return encodeMinKey(buf, name)
	case MaxKey:
		return encodeMaxKey(buf, name)
	case bool:
		return encodeBool(buf, name, Bool(srct))
	case int8:
		return encodeInt32(buf, name, Int32(srct))
	case int16:
		return encodeInt32(buf, name, Int32(srct))
	case int32:
		return encodeInt32(buf, name, Int32(srct))
	case int:
		return encodeInt64(buf, name, Int64(srct))
	case int64:
		return encodeInt64(buf, name, Int64(srct))
	case float64:
		return encodeFloat(buf, name, Float(srct))
	case string:
		return encodeString(buf, name, String(srct))
	case time.Time:
		return encodeUTCDateTime(buf, name, UTCDateTime(srct.UnixNano()/1000/1000))
	case []byte:
		return encodeBinary(buf, name, srct)
	default:
		// Fall back to reflect.
		switch rvsrc.Kind() {
		case reflect.Bool:
			return encodeBool(buf, name, Bool(rvsrc.Bool()))
		case reflect.Int8, reflect.Int16, reflect.Int32:
			return encodeInt32(buf, name, Int32(rvsrc.Int()))
		case reflect.Int, reflect.Int64:
			return encodeInt64(buf, name, Int64(rvsrc.Int()))
		case reflect.Float64:
			return encodeFloat(buf, name, Float(rvsrc.Float()))
		case reflect.Slice:
			a := make(Array, rvsrc.Len())
			for i := 0; i < rvsrc.Len(); i++ {
				a[i] = rvsrc.Index(i).Interface()
			}
			return encodeArray(buf, path, name, a)
		case reflect.String:
			return encodeString(buf, name, String(rvsrc.String()))
		}
	}
	return fmt.Errorf("%v, cannot encode %T.\n", path, src)
}

func encodeArray(buf *bytes.Buffer, path, name string, val Array) error {
	// Array is encoded as a document with incrementing numeric keys.
	if len(val) == 0 {
		return nil
	}
	// type
	if err := buf.WriteByte(_ARRAY); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// Create array doc.
	tmp := bytes.NewBuffer(make([]byte, 0))
	// This will be replaced by the size of the doc later.
	if err := binary.Write(tmp, binary.LittleEndian, uint32(0)); err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		name := strconv.Itoa(i)
		var newpath string
		if path == "" {
			newpath = name
		} else {
			newpath = strings.Join([]string{path, name}, ".")
		}
		if err := encodeVal(tmp, newpath, name, val[i]); err != nil {
			return err
		}
	}
	// End of BSON null byte.
	if err := tmp.WriteByte(0x00); err != nil {
		return err
	}
	// Write size of document at start of BSON.
	binary.LittleEndian.PutUint32(tmp.Bytes(), uint32(tmp.Len()))
	buf.Write(tmp.Bytes())
	return nil
}

func encodeBinary(buf *bytes.Buffer, name string, val Binary) error {
	// type
	if err := buf.WriteByte(_BINARY_DATA); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// value
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(val))); err != nil {
		return err
	}
	// Always use binary/generic subtype.
	if err := buf.WriteByte(0x00); err != nil {
		return err
	}
	if _, err := buf.Write(val); err != nil {
		return err
	}
	return nil
}

func encodeBool(buf *bytes.Buffer, name string, val Bool) error {
	// type
	if err := buf.WriteByte(_BOOLEAN); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// value
	if val {
		if err := buf.WriteByte(0x01); err != nil {
			return err
		}
	} else {
		if err := buf.WriteByte(0x00); err != nil {
			return err
		}
	}
	return nil
}

func encodeDBPointer(buf *bytes.Buffer, path, name string, val DBPointer) error {
	if len(val.ObjectId) != 12 {
		return fmt.Errorf("%v, DBPointer must be 12 bytes.", path)
	}
	// type
	if err := buf.WriteByte(_DBPOINTER); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// Name
	if err := writeString(buf, val.Name); err != nil {
		return err
	}
	// ObjectId
	if _, err := buf.Write(val.ObjectId); err != nil {
		return err
	}
	return nil
}

func encodeEmbeddedDocument(buf *bytes.Buffer, path, name string, val Doc) error {
	// type
	if err := buf.WriteByte(_EMBEDDED_DOCUMENT); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// value
	if a, ok := val.(Map); ok {
		b, err := encodeMap(catpath(path, name), a)
		if err != nil {
			return err
		}
		if _, err := buf.Write(b); err != nil {
			return err
		}
	} else if a, ok := val.(Slice); ok {
		b, err := encodeSlice(catpath(path, name), a)
		if err != nil {
			return err
		}
		if _, err := buf.Write(b); err != nil {
			return err
		}
	} else {
		panic("Programmer mistake, faile to handle Doc type.")
	}

	return nil
}

func encodeFloat(buf *bytes.Buffer, name string, val Float) error {
	// type
	if err := buf.WriteByte(_FLOATING_POINT); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// value
	u := math.Float64bits(float64(val))
	b := make([]byte, 8)
	b[7] = byte(u >> 56)
	b[6] = byte(u >> 48)
	b[5] = byte(u >> 40)
	b[4] = byte(u >> 32)
	b[3] = byte(u >> 24)
	b[2] = byte(u >> 16)
	b[1] = byte(u >> 8)
	b[0] = byte(u)
	if _, err := buf.Write(b); err != nil {
		return err
	}
	return nil
}

func encodeInt32(buf *bytes.Buffer, name string, val Int32) error {
	// type
	if err := buf.WriteByte(_32BIT_INTEGER); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// value
	return binary.Write(buf, binary.LittleEndian, val)
}

func encodeInt64(buf *bytes.Buffer, name string, val Int64) error {
	// type
	buf.WriteByte(_64BIT_INTEGER)
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// value
	binary.Write(buf, binary.LittleEndian, val)
	return nil
}

func encodeJavascript(buf *bytes.Buffer, name string, val Javascript) error {
	// type
	if err := buf.WriteByte(_JAVASCRIPT); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// value
	return writeString(buf, string(val))
}

func encodeJavascriptScope(buf *bytes.Buffer, path, name string, val JavascriptScope) error {
	// type
	if err := buf.WriteByte(_JAVASCRIPT_SCOPE); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// Start code_w_s.
	tmp := bytes.NewBuffer(make([]byte, 0))

	// This will be replaced by the size of code_w_s.
	if err := binary.Write(tmp, binary.LittleEndian, uint32(0)); err != nil {
		return err
	}
	// Write Javascript.
	if err := writeString(tmp, val.Javascript); err != nil {
		return err
	}
	// Write scope.
	b, err := encodeMap(catpath(path, name), val.Scope)
	if err != nil {
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		return err
	}
	// Write size of document at start of code_w_s.
	binary.LittleEndian.PutUint32(tmp.Bytes(), uint32(tmp.Len()))
	buf.Write(tmp.Bytes())
	return nil
}

func encodeMaxKey(buf *bytes.Buffer, name string) error {
	// type
	if err := buf.WriteByte(_MAX_KEY); err != nil {
		return err
	}
	// name
	return writeCstring(buf, name)
}

func encodeMinKey(buf *bytes.Buffer, name string) error {
	// type
	if err := buf.WriteByte(_MIN_KEY); err != nil {
		return err
	}
	// name
	return writeCstring(buf, name)
}

func encodeNull(buf *bytes.Buffer, name string) error {
	// type
	if err := buf.WriteByte(_NULL_VALUE); err != nil {
		return err
	}
	// name
	return writeCstring(buf, name)
}

func encodeObjectId(buf *bytes.Buffer, path, name string, val ObjectId) error {
	if len(val) != 12 {
		return fmt.Errorf("%v, ObjectId must be 12 bytes.", path)
	}
	// type
	if err := buf.WriteByte(_OBJECT_ID); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// value
	if _, err := buf.Write(val); err != nil {
		return err
	}
	return nil
}

func encodeRegexp(buf *bytes.Buffer, name string, val Regexp) error {
	// type
	if err := buf.WriteByte(_REGEXP); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// regex
	if err := writeCstring(buf, val.Pattern); err != nil {
		return err
	}
	// options
	return writeCstring(buf, val.Options)
}

func encodeString(buf *bytes.Buffer, name string, val String) error {
	// type
	if err := buf.WriteByte(_STRING); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// value
	return writeString(buf, string(val))
}

func encodeSymbol(buf *bytes.Buffer, name string, val Symbol) error {
	// type
	if err := buf.WriteByte(_SYMBOL); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// value
	return writeString(buf, string(val))
}

func encodeTimestamp(buf *bytes.Buffer, name string, val Timestamp) error {
	// type
	if err := buf.WriteByte(_TIMESTAMP); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// value
	return binary.Write(buf, binary.LittleEndian, uint64(val))
}

func encodeUndefined(buf *bytes.Buffer, name string) error {
	// type
	if err := buf.WriteByte(_UNDEFINED); err != nil {
		return err
	}
	// name
	return writeCstring(buf, name)
}

func encodeUTCDateTime(buf *bytes.Buffer, name string, val UTCDateTime) error {
	// type
	if err := buf.WriteByte(_UTC_DATETIME); err != nil {
		return err
	}
	// name
	if err := writeCstring(buf, name); err != nil {
		return err
	}
	// value
	return binary.Write(buf, binary.LittleEndian, val)
}

// Copied verbatim from the json package in the standard library.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// Write BSON cstring.
func writeCstring(buf *bytes.Buffer, s string) error {
	if _, err := buf.WriteString(s); err != nil {
		return err
	}
	return buf.WriteByte(0x00)
}

// Write BSON string.
func writeString(buf *bytes.Buffer, s string) error {
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(s)+1)); err != nil {
		return err
	}
	if _, err := buf.WriteString(s); err != nil {
		return err
	}
	return buf.WriteByte(0x00)
}
