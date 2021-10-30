package goeloquent

import (
	"fmt"
	"reflect"
	"sync"
)

var Eloquent *DB

type DB struct {
	DatabaseManager
	RegisteredModelsMap      sync.Map
	RegisteredMorphModelsMap sync.Map
	RegisteredDBMap          sync.Map
	ParsedModelsMap          sync.Map
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
	for DbName, pointer := range morphMap {
		if reflect.TypeOf(pointer).Kind() != reflect.Ptr {
			panic("morph map value must be model pointer")
		}
		Eloquent.RegisteredDBMap.Store(DbName, reflect.Indirect(reflect.ValueOf(pointer)))
		Eloquent.RegisteredMorphModelsMap.Store(reflect.Indirect(reflect.ValueOf(pointer)).Type().Name(), DbName)
	}
}
func GetMorphMap(name string) string {
	v, ok := Eloquent.RegisteredMorphModelsMap.Load(name)
	if !ok {
		panic(fmt.Sprintf("no registered model found for %s", name))
	}
	return v.(string)
}
func GetMorphDBMap(name string) reflect.Value {
	v, ok := Eloquent.RegisteredDBMap.Load(name)
	if !ok {
		panic(fmt.Sprintf("no registered model found for %s", name))
	}
	return v.(reflect.Value)
}
func RegisterModels(models []interface{}) {
	for _, m := range models {
		t := reflect.ValueOf(m).Elem().Type()
		Eloquent.RegisteredModelsMap.Store(t.PkgPath()+"."+t.Name(), reflect.Indirect(reflect.ValueOf(m)))
	}
}
func GetRegisteredModel(name string) reflect.Value {
	v, ok := Eloquent.RegisteredModelsMap.Load(name)
	if !ok {
		panic(fmt.Sprintf("no registered model found for %s", name))
	}
	return v.(reflect.Value)
}
func GetParsed(name string) (interface{}, bool) {
	return Eloquent.ParsedModelsMap.Load(name)
}
