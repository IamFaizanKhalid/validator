// Package validator implements value validations
//
// Copyright 2014 Roberto Teixeira <robteix@robteix.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	"reflect"
	"regexp"
	"strconv"
	"unicode/utf8"
)

// nonzero tests whether a variable value non-zero
// as defined by the golang spec.
func nonzero(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	var valid bool
	switch st.Kind() {
	case reflect.String:
		valid = utf8.RuneCountInString(st.String()) != 0
	case reflect.Ptr, reflect.Interface:
		valid = !st.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array:
		valid = st.Len() != 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valid = st.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		valid = st.Uint() != 0
	case reflect.Float32, reflect.Float64:
		valid = st.Float() != 0
	case reflect.Bool:
		valid = st.Bool()
	case reflect.Invalid:
		valid = false // always invalid
	case reflect.Struct:
		valid = true // always valid since only nil pointers are empty
	default:
		return ErrUnsupported
	}

	if !valid {
		return ErrZeroValue
	}
	return nil
}

// length tests whether a variable's length is equal to a given
// value. For strings it tests the number of characters whereas
// for maps and slices it tests the number of items.
func length(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	var valid bool
	if st.Kind() == reflect.Ptr {
		if st.IsNil() {
			return nil
		}
		st = st.Elem()
	}
	switch st.Kind() {
	case reflect.String:
		p, err := asInt(param)
		if err != nil {
			return ErrBadParameter
		}
		valid = int64(utf8.RuneCountInString(st.String())) == p
	case reflect.Slice, reflect.Map, reflect.Array:
		p, err := asInt(param)
		if err != nil {
			return ErrBadParameter
		}
		valid = int64(st.Len()) == p
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p, err := asInt(param)
		if err != nil {
			return ErrBadParameter
		}
		valid = st.Int() == p
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p, err := asUint(param)
		if err != nil {
			return ErrBadParameter
		}
		valid = st.Uint() == p
	case reflect.Float32, reflect.Float64:
		p, err := asFloat(param)
		if err != nil {
			return ErrBadParameter
		}
		valid = st.Float() == p
	default:
		return ErrUnsupported
	}
	if !valid {
		return ErrLen
	}
	return nil
}

// min tests whether a variable value is larger or equal to a given
// number. For number types, it's a simple lesser-than test; for
// strings it tests the number of characters whereas for maps
// and slices it tests the number of items.
func min(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	invalid := false
	if st.Kind() == reflect.Ptr {
		if st.IsNil() {
			return nil
		}
		st = st.Elem()
	}
	switch st.Kind() {
	case reflect.String:
		p, err := asInt(param)
		if err != nil {
			return ErrBadParameter
		}
		invalid = int64(utf8.RuneCountInString(st.String())) < p
	case reflect.Slice, reflect.Map, reflect.Array:
		p, err := asInt(param)
		if err != nil {
			return ErrBadParameter
		}
		invalid = int64(st.Len()) < p
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p, err := asInt(param)
		if err != nil {
			return ErrBadParameter
		}
		invalid = st.Int() < p
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p, err := asUint(param)
		if err != nil {
			return ErrBadParameter
		}
		invalid = st.Uint() < p
	case reflect.Float32, reflect.Float64:
		p, err := asFloat(param)
		if err != nil {
			return ErrBadParameter
		}
		invalid = st.Float() < p
	default:
		return ErrUnsupported
	}
	if invalid {
		return ErrMin
	}
	return nil
}

// max tests whether a variable value is lesser than a given
// value. For numbers, it's a simple lesser-than test; for
// strings it tests the number of characters whereas for maps
// and slices it tests the number of items.
func max(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	var invalid bool
	if st.Kind() == reflect.Ptr {
		if st.IsNil() {
			return nil
		}
		st = st.Elem()
	}
	switch st.Kind() {
	case reflect.String:
		p, err := asInt(param)
		if err != nil {
			return ErrBadParameter
		}
		invalid = int64(utf8.RuneCountInString(st.String())) > p
	case reflect.Slice, reflect.Map, reflect.Array:
		p, err := asInt(param)
		if err != nil {
			return ErrBadParameter
		}
		invalid = int64(st.Len()) > p
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p, err := asInt(param)
		if err != nil {
			return ErrBadParameter
		}
		invalid = st.Int() > p
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p, err := asUint(param)
		if err != nil {
			return ErrBadParameter
		}
		invalid = st.Uint() > p
	case reflect.Float32, reflect.Float64:
		p, err := asFloat(param)
		if err != nil {
			return ErrBadParameter
		}
		invalid = st.Float() > p
	default:
		return ErrUnsupported
	}
	if invalid {
		return ErrMax
	}
	return nil
}

// regex is the builtin validation function that checks
// whether the string variable matches a regular expression
func regex(v interface{}, param string) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.String {
		return ErrUnsupported
	}
	s := rv.String()
	re, err := regexp.Compile(param)
	if err != nil {
		return ErrBadParameter
	}

	if !re.MatchString(s) {
		return ErrRegexp
	}
	return nil
}

// asInt returns the parameter as a int64
// or panics if it can't convert
func asInt(param string) (int64, error) {
	i, err := strconv.ParseInt(param, 0, 64)
	if err != nil {
		return 0, ErrBadParameter
	}
	return i, nil
}

// asUint returns the parameter as a uint64
// or panics if it can't convert
func asUint(param string) (uint64, error) {
	i, err := strconv.ParseUint(param, 0, 64)
	if err != nil {
		return 0, ErrBadParameter
	}
	return i, nil
}

// asFloat returns the parameter as a float64
// or panics if it can't convert
func asFloat(param string) (float64, error) {
	i, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return 0.0, ErrBadParameter
	}
	return i, nil
}

// nonnil validates that the given pointer is not nil
func nonnil(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	// if we got a non-pointer then we most likely got
	// the value for a pointer field, either way, its not
	// nil
	switch st.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Func:
		if st.IsNil() {
			return ErrZeroValue
		}
	case reflect.Invalid:
		// the only way its invalid is if its an interface that's nil
		return ErrZeroValue
	}
	return nil
}

// unique validates uniqueness of elements of an array
func unique(v interface{}, _ string) error {
	arr := reflect.ValueOf(v)
	if arr.Kind() == reflect.Ptr {
		if arr.IsNil() {
			return nil
		}
		arr = arr.Elem()
	}
	switch arr.Kind() {
	case reflect.Slice, reflect.Array:
		m := make(map[interface{}]bool)
		for i := 0; i < arr.Len(); i++ {
			val := arr.Index(i).Interface()
			_, exists := m[val]
			if exists {
				return ErrRepeating
			}

			m[val] = true
		}
	default:
		return ErrUnsupported
	}

	return nil
}
