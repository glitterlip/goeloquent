package goeloquent

import (
	"errors"
	"fmt"
	"github.com/glitterlip/goeloquent/connectors"
	"github.com/glitterlip/goeloquent/eloquent"
	"github.com/glitterlip/goeloquent/query"
	"reflect"
)

type DatabaseManager struct {
	Connections map[string]*Connection
	Factory     connectors.ConnectionFactory
	Configs     map[string]*connectors.DBConfig
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

/*
GetConfig get a db config by name
*/
func (dm *DatabaseManager) GetConfig(name string) *connectors.DBConfig {
	return dm.Configs[name]
}

/*
MakeConnection make a connection by name
*/
func (dm *DatabaseManager) MakeConnection(connectionName string) *Connection {
	config, ok := dm.Configs[connectionName]
	if !ok {
		panic(errors.New(fmt.Sprintf("Database connection %s not configured.", connectionName)))
	}

	conn := dm.Factory.Make(config)
	conn.ConnectionName = connectionName
	dm.Connections[connectionName] = conn
	return conn
}

/*
Table get a query builder and set table name
*/
func (dm *DatabaseManager) Table(params ...string) *query.Builder {
	builder := dm.Query()
	builder.Table(params...)
	return builder
}

/*
Model get a eloquent builder and set model

1. Mode(&User{})

2. Mode("User")
*/
func (dm *DatabaseManager) Model(model interface{}, connectionName ...string) *eloquent.Builder {
	parsed := eloquent.GetParsedModel(model)

	var conn string
	if len(connectionName) > 0 {
		conn = connectionName[0]
	} else if len(parsed.ConnectionName) > 0 {
		conn = parsed.ConnectionName
	} else {
		conn = DefaultConnectionName
	}
	c := dm.Connection(conn)
	builder := query.NewBuilder(c)

	return &eloquent.Builder{
		Builder:   builder,
		BaseModel: parsed,
	}

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
func (dm *DatabaseManager) Query() *query.Builder {
	defaultConn := dm.getDefaultConnection()
	c := dm.Connection(defaultConn)
	return query.NewBuilder(c)
}
func (dm *DatabaseManager) Transaction(closure TxClosure) (interface{}, error) {
	ic := dm.Connections[DefaultConnectionName]
	return (*ic).Transaction(closure)
}
func (dm *DatabaseManager) BeginTransaction() (*Transaction, error) {
	ic := dm.Connections[DefaultConnectionName]
	return (*ic).BeginTransaction()
}

func (dm *DatabaseManager) Create(model interface{}) (res Result, err error) {
	return dm.Save(model)
}

func (dm *DatabaseManager) Save(modelP interface{}) (res Result, err error) {
	//TODO: use connection as proxy
	//TODO: batch create/update/delete
	parsed := eloquent.GetParsedModel(modelP)
	model := reflect.Indirect(reflect.ValueOf(modelP))
	if parsed.IsEloquent {
		ininted := !model.Field(parsed.EloquentModelFieldIndex).IsZero()
		if ininted {
			t := model.Field(parsed.EloquentModelFieldIndex).Interface().(*eloquent.EloquentModel)
			return t.Save()
		} else {
			return eloquent.InitModel(modelP).Create()
		}
	} else {
		//TODO: save plain struct
	}
	return
}
func (dm *DatabaseManager) Boot(modelP interface{}) *eloquent.EloquentModel {
	parsed := eloquent.GetParsedModel(modelP)
	model := reflect.Indirect(reflect.ValueOf(modelP))
	if parsed.IsEloquent {
		ininted := !model.Field(parsed.EloquentModelFieldIndex).IsZero()
		if !ininted {
			return eloquent.InitModel(modelP)
		}
		return model.Field(parsed.EloquentModelFieldIndex).Interface().(*eloquent.EloquentModel)
	}
	return nil
}
