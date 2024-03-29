package goeloquent

import (
	"database/sql"
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
	RegisteredMacros         map[string]MacroFunc
	ParsedModelsMap          sync.Map
	LogFunc                  func(log Log)
}
type MacroFunc = func(builder *Builder, params ...interface{}) *Builder

func (d *DB) SetLogger(f func(log Log)) *DB {
	d.LogFunc = f
	return d
}
func Open(config map[string]DBConfig) *DB {
	var configP = make(map[string]*DBConfig)
	for name := range config {
		c := config[name]
		configP[name] = &c
	}
	db := DB{
		DatabaseManager: DatabaseManager{
			Configs:     configP,
			Connections: make(map[string]*Connection),
		},
	}
	db.Connection("default")
	Eloquent = &db
	return Eloquent
}
func (d *DB) AddConfig(name string, config *DBConfig) *DB {
	Eloquent.Configs[name] = config
	return d
}
func (d *DB) GetConfigs() map[string]*DBConfig {
	return Eloquent.Configs
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
		i := reflect.Indirect(reflect.ValueOf(m))
		Eloquent.RegisteredModelsMap.Store(t.PkgPath()+"."+t.Name(), i)
		Eloquent.RegisteredModelsMap.Store(t.Name(), i)
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
func (*DB) Raw(connectionName ...string) *sql.DB {

	if len(connectionName) > 0 {
		c := Eloquent.Connection(connectionName[0])
		return (*c).GetDB()
	} else {
		c := Eloquent.Connection("default")
		return (*c).GetDB()
	}
}
func (d *DB) RegistMacroMap(macroMap map[string]MacroFunc) {
	d.RegisteredMacros = macroMap
}
