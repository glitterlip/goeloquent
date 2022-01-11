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
func ExtractStruct(target interface{}) map[string]interface{} {
	tv := reflect.Indirect(reflect.ValueOf(target))
	tt := tv.Type()
	result := make(map[string]interface{}, tv.NumField())
	for i := 0; i < tv.NumField(); i++ {
		//when using simple struct ,provide a convenient way to set table
		if tt.Field(i).Name == "Table" && strings.Contains(tt.Field(i).Tag.Get("goelo"), "TableName:") {
			continue
		}
		key := ToSnakeCase(tt.Field(i).Name)
		result[key] = tv.Field(i).Interface()
	}
	return result
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
