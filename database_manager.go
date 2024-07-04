package goeloquent

import (
	"fmt"
	"reflect"
	"sync"
)

type DatabaseManager struct {
	Connections map[string]*Connection
	Factory     ConnectionFactory
	Configs     map[string]*DBConfig
	Listeners   map[string][]interface{}
}

var ParsedModelsMap sync.Map          //pkg+modelname:*eloquent.Model , get parsed model config
var RegisteredModelsMap sync.Map      //name:reflect.Value
var RegisteredMorphModelsMap sync.Map //model pointer:alias , when save relation to db, convert models.User => users
var RegisteredDBMap sync.Map          //alias:model pointer , when query relation from db, convert users => models.User
func init() {
	ParsedModelsMap = sync.Map{}
	RegisteredModelsMap = sync.Map{}
	RegisteredMorphModelsMap = sync.Map{}
	RegisteredDBMap = sync.Map{}
}

/*
Connection get a connection pointer by name
*/
func (dm *DatabaseManager) Connection(connectionName string) *Connection {
	connection, ok := dm.Connections[connectionName]
	if !ok {
		dm.Connections[connectionName] = dm.MakeConnection(connectionName)
		connection, _ = dm.Connections[connectionName]
	}
	return connection
}

/*
Conn get a connection by name
*/
func (dm *DatabaseManager) Conn(connectionName string) Connection {
	return *(dm.Connection(connectionName))
}

func (dm *DatabaseManager) getDefaultConnection() (defaultConnectionName string) {
	defaultConnectionName = DefaultConnectionName
	return
}
func (dm *DatabaseManager) Listen(eventName string, listener interface{}) {
	if _, ok := dm.Listeners[eventName]; !ok {
		dm.Listeners[eventName] = []interface{}{}
	}
	dm.Listeners[eventName] = append(dm.Listeners[eventName], listener)
}

/*
GetConfig get a db config by name
*/
func (dm *DatabaseManager) GetConfig(name string) *DBConfig {
	return dm.Configs[name]
}

/*
MakeConnection make a connection by name
*/
func (dm *DatabaseManager) MakeConnection(connectionName string) *Connection {
	config, ok := dm.Configs[connectionName]
	if !ok {
		panic(fmt.Sprintf("Database connection %s not configured.", connectionName))
	}

	conn := dm.Factory.Make(config)
	conn.ConnectionName = connectionName
	dm.Connections[connectionName] = conn
	return conn
}

/*
Table get a query builder and set table name
*/
func (dm *DatabaseManager) Table(params ...string) *Builder {
	builder := dm.Query()
	builder.Table(params...)
	return builder
}

/*
Model get a eloquent builder and set model

1. Mode(&User{})

2. Mode("User")
*/
func (dm *DatabaseManager) Model(model ...interface{}) *EloquentBuilder {

	c := dm.Connection(DefaultConnectionName)
	if len(model) == 0 {
		return c.Model()
	}
	return c.Model(model)

}
func (dm *DatabaseManager) Select(query string, bindings []interface{}, dest interface{}) (Result, error) {
	ic := dm.Connections[DefaultConnectionName]
	return (*ic).Select(query, bindings, dest, nil)
}
func (dm *DatabaseManager) Insert(query string, bindings []interface{}) (Result, error) {
	ic := dm.Connections[DefaultConnectionName]
	return (*ic).Insert(query, bindings)
}
func (dm *DatabaseManager) Update(query string, bindings []interface{}) (Result, error) {
	ic := dm.Connections[DefaultConnectionName]
	return (*ic).Update(query, bindings)
}
func (dm *DatabaseManager) Delete(query string, bindings []interface{}) (Result, error) {
	ic := dm.Connections[DefaultConnectionName]
	return (*ic).Delete(query, bindings)
}
func (dm *DatabaseManager) Statement(query string, bindings []interface{}) (Result, error) {
	ic := dm.Connections[DefaultConnectionName]
	return (*ic).Delete(query, bindings)
}
func (dm *DatabaseManager) Query() *Builder {
	defaultConn := dm.getDefaultConnection()
	c := dm.Connection(defaultConn)
	return NewQueryBuilder(c)
}

//func (dm *DatabaseManager) Transaction(closure TxClosure) (interface{}, error) {
//	ic := dm.Connections[DefaultConnectionName]
//	return (*ic).Transaction(closure)
//}
//func (dm *DatabaseManager) BeginTransaction() (*interfaces.ITransaction, error) {
//	ic := dm.Connections[DefaultConnectionName]
//	return (*ic).BeginTransaction()
//}

func (dm *DatabaseManager) Create(model interface{}) (res Result, err error) {
	return dm.Save(model)
}

func (dm *DatabaseManager) Save(modelP interface{}) (res Result, err error) {
	//TODO: use connection as proxy
	//TODO: batch create/update/delete
	parsed := GetParsedModel(modelP)
	model := reflect.Indirect(reflect.ValueOf(modelP))
	if parsed.IsEloquent {
		ininted := !model.Field(parsed.EloquentModelFieldIndex).IsZero()
		if ininted {
			t := model.Field(parsed.EloquentModelFieldIndex).Interface().(*EloquentModel)
			return t.Save()
		} else {
			return InitModel(modelP).Create()
		}
	} else {
		//TODO: save plain struct
	}
	return
}
func (dm *DatabaseManager) Boot(modelP interface{}) *EloquentModel {
	parsed := GetParsedModel(modelP)
	model := reflect.Indirect(reflect.ValueOf(modelP))
	if parsed.IsEloquent {
		ininted := !model.Field(parsed.EloquentModelFieldIndex).IsZero()
		if !ininted {
			return InitModel(modelP)
		}
		return model.Field(parsed.EloquentModelFieldIndex).Interface().(*EloquentModel)
	}
	return nil
}
