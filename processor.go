package goeloquent

import "reflect"

type IProcessor interface {
	processSelect()
	processInsertGetId()
	processColumnListing()
}

func SetAttr(obj interface{}, key string, value interface{}) {
	ve := reflect.ValueOf(obj).Elem()
	vf := ve.FieldByName(key)
	vf.Set(reflect.ValueOf(value))
}
