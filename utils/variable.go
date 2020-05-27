/*
 * @Author: ZhenpengDeng(monitor1379)
 * @Date: 2020-05-27 18:12:50
 * @Last Modified by: ZhenpengDeng(monitor1379)
 * @Last Modified time: 2020-05-27 23:53:14
 */
package utils

import (
	"bytes"
	"encoding/gob"
	"reflect"
)

func GetGobSizeOfVariable(v interface{}) (int64, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(v)
	if err != nil {
		return 0, err
	}

	return int64(buf.Len()), nil
}

func GetVariableMemoryUsage(v interface{}) int64 {
	// n, err := GetGobSizeOfVariable(v)
	// if err != nil {
	// 	return 0
	// }
	n := Sizeof(v)
	return n
}

func Sizeof(v interface{}) int64 {
	vv := reflect.ValueOf(v)
	return sizeof(vv)
}

func sizeof(v reflect.Value) int64 {
	switch v.Kind() {
	case reflect.Invalid:
		return 0
	case reflect.Ptr:
		return sizeof(v.Elem())
	case reflect.Bool, reflect.Int8, reflect.Uint8:
		return 1
	case reflect.Int16, reflect.Uint16:
		return 2
	case reflect.Int, reflect.Uint, reflect.Int32, reflect.Uint32, reflect.Float32:
		return 4
	case reflect.Int64, reflect.Uint64, reflect.Float64, reflect.Complex64, reflect.Uintptr:
		return 8
	case reflect.Complex128:
		return 16
	case reflect.String:
		return int64(v.Len())
	case reflect.Slice, reflect.Array:
		if v.Len() > 0 {
			return int64(v.Len()) * sizeof(v.Index(0))
		}
		return 0
	case reflect.Map:
		var s int64
		for _, key := range v.MapKeys() {
			s += sizeof(key) + sizeof(v.MapIndex(key))
		}
		return s

	case reflect.Struct:
		var s int64
		for i := 0; i < v.NumField(); i++ {
			s += sizeof(v.Field(i))
		}
		return s

	default:
		panic("unsupport variable type for calculating memory usage")
	}
	return 0
}
