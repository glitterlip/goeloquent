package goeloquent

import (
	"database/sql"
	"fmt"
	"github.com/glitterlip/goeloquent/connectors"
	"github.com/glitterlip/goeloquent/eloquent"
	"github.com/glitterlip/goeloquent/query"
	"github.com/glitterlip/goeloquent/query/grammar"
	"reflect"
	"sync"
)

const (
	DefaultConnectionName = "default"
)

var Eloquent *DB

type DB struct {
	DatabaseManager
	RegisteredModelsMap      sync.Map //name:reflect.Value
	RegisteredMorphModelsMap sync.Map //model pointer:alias , when save relation to db, convert models.User => users
	RegisteredDBMap          sync.Map //alias:model pointer , when query relation from db, convert users => models.User
	ParsedModelsMap          sync.Map //pkg+modelname:*eloquent.Model , get parsed model config
	LogFunc                  func(log Log)
}
type EloquentModel = eloquent.EloquentModel
type Builder = query.Builder
type MysqlGrammar = grammar.MysqlGrammar
type MacroFunc = func(builder *query.Builder, params ...interface{}) *query.Builder

func CloneBuilderWithTable(b *Builder) *Builder {
	return query.Clone(b)
}
func NewBuilder(c *Connection) *Builder {
	return query.NewBuilder(c)
}
func Clone(original *Builder) *Builder {
	return query.Clone(original)
}
func CloneWithout(original *Builder, without ...string) *Builder {
	return query.CloneWithout(original, without...)
}
func (d *DB) SetLogger(f func(log Log)) *DB {
	d.LogFunc = f
	return d
}
func Open(config map[string]connectors.DBConfig) *DB {
	var configP = make(map[string]*connectors.DBConfig)
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
func (d *DB) AddConfig(name string, config *connectors.DBConfig) *DB {
	Eloquent.Configs[name] = config
	return d
}
func (d *DB) GetConfigs() map[string]*connectors.DBConfig {
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
