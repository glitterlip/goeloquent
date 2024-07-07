package goeloquent

import (
	"database/sql"
	"fmt"
	"reflect"
)

const (
	DefaultConnectionName = "default"
)

var DB *DatabaseManager

type MacroFunc = func(builder *Builder, params ...interface{}) *Builder

func CloneBuilderWithTable(b *Builder) *Builder {
	return Clone(b)
}

func Open(config map[string]DBConfig) *DatabaseManager {
	var configP = make(map[string]*DBConfig)
	for name := range config {
		c := config[name]
		configP[name] = &c
	}
	db := DatabaseManager{
		Configs:     configP,
		Connections: make(map[string]*Connection),
	}
	db.Connection("default")
	DB = &db
	db.FireEvent(EventOpened, config)
	return DB
}
func (d DatabaseManager) AddConfig(name string, config *DBConfig) DatabaseManager {
	DB.Configs[name] = config
	return d
}
func (d DatabaseManager) GetConfigs() map[string]*DBConfig {
	return DB.Configs
}
func RegistMorphMap(morphMap map[string]interface{}) {
	for DbName, pointer := range morphMap {
		if reflect.TypeOf(pointer).Kind() != reflect.Ptr {
			panic("morph map value must be model pointer")
		}
		RegisteredDBMap.Store(DbName, reflect.Indirect(reflect.ValueOf(pointer)))
		RegisteredMorphModelsMap.Store(reflect.Indirect(reflect.ValueOf(pointer)).Type().Name(), DbName)
	}
}

func GetMorphDBMap(name string) reflect.Value {
	v, ok := RegisteredDBMap.Load(name)
	if !ok {
		panic(fmt.Sprintf("no registered model found for %s", name))
	}
	return v.(reflect.Value)
}
func RegisterModels(models []interface{}) {
	for _, m := range models {
		t := reflect.ValueOf(m).Elem().Type()
		i := reflect.Indirect(reflect.ValueOf(m))
		RegisteredModelsMap.Store(t.PkgPath()+"."+t.Name(), i)
		RegisteredModelsMap.Store(t.Name(), i)
	}
}

func (DatabaseManager) Raw(connectionName ...string) *sql.DB {

	if len(connectionName) > 0 {
		c := DB.Connection(connectionName[0])
		return (*c).GetDB()
	} else {
		c := DB.Connection("default")
		return (*c).GetDB()
	}
}
