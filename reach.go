package bson

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

// Reach in to document to get a value.
// If dst is nil or a pointer to nil then a new object will be allocated.
//
// Returns true if object found, false if object not present.
// Return error if there is a coercion problem.
//
// Supported Coercions:
//   Float       -> float64
//   String      -> string
//   Binary      -> []byte
//   ObjectID    -> []byte
//   Bool        -> bool
//   UTCDateTime -> int64, time.Time
//   Javascript  -> string
//   Symbol      -> string
//   Int32       -> int32, int64
//   Timestamp   -> int64, time.Time
//   Int64       -> int64
//
// To disable coercion use only bson types.
func (this Map) Reach(dst interface{}, dot ...string) (bool, error) {
	if dst == nil {
		return false, errors.New("dst must not be nil.")
	}
	src := reach(this, dot...)
	if src == nil {
		return false, nil
	}
	return assign(dst, src)
}

// Same as map reach.
func (this Slice) Reach(dst interface{}, dot ...string) (bool, error) {
	if dst == nil {
		return false, errors.New("dst must not be nil.")
	}
	src := reach(this, dot...)
	if src == nil {
		return false, nil
	}
	return assign(dst, src)
}

func reach(cur interface{}, dot ...string) interface{} {
	path := ""
	for _, name := range dot {
		path = catpath(path, name)
		switch curt := cur.(type) {
		case Float, String, Array, Binary, Undefined, ObjectId, Bool, UTCDateTime,
			Null, Javascript, Symbol, Int32, Timestamp, Int64, MinKey, MaxKey:
			return nil
		case Map:
			a, ok := curt[name]
			if !ok {
				return nil
			}
			cur = a
		case Slice:
			ok := false
			for _, v := range curt {
				if v.Key == name {
					ok = true
					cur = v
					break
				}
			}
			if !ok {
				return nil
			}
		case Regexp:
			if name == "Pattern" {
				cur = curt.Pattern
			} else if name == "Options" {
				cur = curt.Options
			} else {
				return nil
			}
		case DBPointer:
			if name == "Name" {
				cur = curt.Name
			} else if name == "ObjectId" {
				cur = curt.ObjectId
			} else {
				return nil
			}
		case JavascriptScope:
			if name == "Javascript" {
				cur = curt.Javascript
			} else if name == "Scope" {
				cur = curt.Scope
			} else {
				return nil
			}
		default:
			return nil
		}
	}
	return cur
}

func assignError(dst reflect.Value, src interface{}) error {
	return fmt.Errorf("cannot coerce %T to %T.", src, dst.Interface())
}

// Assign and coerce if needed.
func assign(dst, src interface{}) (bool, error) {
	dstrv := indirectAlloc(reflect.ValueOf(dst))
	switch srct := src.(type) {
	case Float:
		if dstrv.Kind() != reflect.Float64 {
			return false, assignError(dstrv, src)
		}
		dstrv.SetFloat(float64(srct))
	case String:
		if dstrv.Kind() != reflect.String {
			return false, assignError(dstrv, src)
		}
		dstrv.SetString(string(srct))
	case Map:
		switch dstrv.Interface().(type) {
		case Map:
			dstrv.Set(reflect.ValueOf(srct))
		default:
			return false, assignError(dstrv, src)
		}
	case Slice:
		switch dstrv.Interface().(type) {
		case Slice:
			dstrv.Set(reflect.ValueOf(srct))
		default:
			return false, assignError(dstrv, src)
		}
	case Array:
		switch dstrv.Interface().(type) {
		case Array:
			dstrv.Set(reflect.ValueOf(srct))
		default:
			switch dstrv.Interface().(type) {
			case Array:
				dstrv.Set(reflect.ValueOf(srct))
			default:
				return false, assignError(dstrv, src)
			}
		}
	case Binary:
		if dstrv.Kind() != reflect.Slice && dstrv.Elem().Kind() != reflect.Uint8 {
			return false, assignError(dstrv, src)
		}
		dstrv.SetBytes([]byte(srct))
	case Undefined:
		// Nothing to do.
	case ObjectId:
		if dstrv.Kind() != reflect.Slice && dstrv.Elem().Kind() != reflect.Uint8 {
			return false, assignError(dstrv, src)
		}
		dstrv.SetBytes([]byte(srct))
	case Bool:
		if dstrv.Kind() != reflect.Bool {
			return false, assignError(dstrv, src)
		}
		dstrv.SetBool(bool(srct))
	case UTCDateTime:
		switch dstrv.Interface().(type) {
		case time.Time:
			// BSON time is milliseconds since unix epoch.
			// Go time is nanoseconds since unix epoch.
			dstrv.Set(reflect.ValueOf(time.Unix(0, int64(srct)*1e3)))
		default:
			if dstrv.Kind() != reflect.Int64 {
				return false, assignError(dstrv, src)
			}
			dstrv.SetInt(int64(srct))
		}
	case Null:
		// Nothing to do.
	case Regexp:
		switch dstrv.Interface().(type) {
		case Regexp:
			dstrv.Set(reflect.ValueOf(srct))
		default:
			return false, assignError(dstrv, src)
		}
	case DBPointer:
			switch dstrv.Interface().(type) {
			case DBPointer:
				dstrv.Set(reflect.ValueOf(srct))
			default:
				return false, assignError(dstrv, src)
			}
	case Javascript:
		if dstrv.Kind() != reflect.String {
			return false, assignError(dstrv, src)
		}
		dstrv.SetString(string(srct))
	case Symbol:
		if dstrv.Kind() != reflect.String {
			return false, assignError(dstrv, src)
		}
		dstrv.SetString(string(srct))
	case JavascriptScope:
		switch dstrv.Interface().(type) {
		case JavascriptScope:
			dstrv.Set(reflect.ValueOf(srct))
		default:
			return false, assignError(dstrv, src)
		}
	case Int32:
		if dstrv.Kind() != reflect.Int32 && dstrv.Kind() != reflect.Int64 {
			return false, assignError(dstrv, src)
		}
		dstrv.SetInt(int64(srct))
	case Timestamp:
		switch dstrv.Interface().(type) {
		case time.Time:
			// BSON time is milliseconds since unix epoch.
			// Go time is nanoseconds since unix epoch.
			dstrv.Set(reflect.ValueOf(time.Unix(0, int64(srct)*1e3)))
		default:
			if dstrv.Kind() != reflect.Int64 {
				return false, assignError(dstrv, src)
			}
			dstrv.SetInt(int64(srct))
		}
	case Int64:
		if dstrv.Kind() != reflect.Int64 {
			return false, assignError(dstrv, src)
		}
		dstrv.SetInt(int64(srct))
	case MinKey:
		// Nothing to do.
	case MaxKey:
		// Nothing to do.
	}
	return true, nil
}
