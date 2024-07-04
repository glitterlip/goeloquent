package goeloquent

import (
	"reflect"
	"regexp"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
func InterfaceToSlice(param interface{}) []interface{} {
	if p, ok := param.([]interface{}); ok {
		return p
	}
	tv := reflect.Indirect(reflect.ValueOf(param))
	var res []interface{}
	if tv.Type().Kind() == reflect.Slice {
		for i := 0; i < tv.Len(); i++ {
			res = append(res, tv.Index(i).Interface())
		}
	} else {
		panic("not slice")
	}
	return res
}

// StrPtr
// return a pointer of string helper function for create mapping
func StrPtr() *string {
	return new(string)
}

// UintPtr
// return a pointer of uint64 helper function for create mapping
func UintPtr() *uint64 {
	return new(uint64)
}
