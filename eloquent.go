package goeloquent

import "reflect"

var Eloquent *DB

type DB struct {
	DatabaseManager
	MorphDBMap    map[string]reflect.Value //dbvalue=>model
	MorphModelMap map[string]string        //modelname=>dbvalue

}

func Open(config map[string]DBConfig) *DB {
	db := DB{
		DatabaseManager: DatabaseManager{
			Configs:     config,
			Connections: make(map[string]*IConnection),
		},
	}
	db.Connection("default")
	Eloquent = &db
	return Eloquent
}
func RegistMorphMap(morphMap map[string]interface{}) {
	Eloquent.MorphDBMap = make(map[string]reflect.Value)
	Eloquent.MorphModelMap = make(map[string]string)
	for DbName, pointer := range morphMap {
		if reflect.TypeOf(pointer).Kind() != reflect.Ptr {
			panic("morph map value must be model pointer")
		}
		Eloquent.MorphDBMap[DbName] = reflect.Indirect(reflect.ValueOf(pointer))
		Eloquent.MorphModelMap[reflect.Indirect(reflect.ValueOf(pointer)).Type().Name()] = DbName
	}
}
