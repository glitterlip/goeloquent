package goeloquent

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

type DatabaseManager struct {
	Connections map[string]*Connection
	Factory     ConnectionFactory
	Configs     map[string]*DBConfig
}

func (dm *DatabaseManager) Connection(connectionName string) *Connection {
	connection, ok := dm.Connections[connectionName]
	if !ok {
		dm.Connections[connectionName] = dm.MakeConnection(connectionName)
		connection, _ = dm.Connections[connectionName]
	}
	return connection
}
func (dm *DatabaseManager) Conn(connectionName string) Connection {
	return *(dm.Connection(connectionName))
}
func (dm *DatabaseManager) getDefaultConnection() (defaultConnectionName string) {
	defaultConnectionName = "default"
	return
}
func (dm *DatabaseManager) GetConfig(name string) *DBConfig {
	return dm.Configs[name]
}
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
func (dm *DatabaseManager) Table(params ...string) *Builder {
	defaultConn := dm.getDefaultConnection()
	c := dm.Connection(defaultConn)
	builder := NewBuilder(c)
	builder.Grammar = &MysqlGrammar{}
	builder.Grammar.SetTablePrefix(dm.Configs[defaultConn].Prefix)
	builder.Grammar.SetBuilder(builder)
	builder.Table(params...)
	return builder

}
func (dm *DatabaseManager) Model(model ...interface{}) *Builder {
	if len(model) > 0 {
		parsed := GetParsedModel(model[0])
		var connectionName string
		if len(parsed.ConnectionName) > 0 {
			connectionName = parsed.ConnectionName
		} else {
			connectionName = dm.getDefaultConnection()
		}
		c := dm.Connection(connectionName)
		builder := NewBuilder(c)
		builder.Grammar = &MysqlGrammar{}
		builder.Grammar.SetTablePrefix(dm.Configs[connectionName].Prefix)
		builder.Grammar.SetBuilder(builder)
		builder.SetModel(model[0])
		return builder
	} else {
		c := dm.Connection("default")
		builder := NewBuilder(c)
		builder.Grammar = &MysqlGrammar{}
		builder.Grammar.SetBuilder(builder)
		return builder
	}

}
func (dm *DatabaseManager) Select(query string, bindings []interface{}, dest interface{}) (sql.Result, error) {
	ic := dm.Connections["default"]
	return (*ic).Select(query, bindings, dest)
}
func (dm *DatabaseManager) Insert(query string, bindings []interface{}) (sql.Result, error) {
	ic := dm.Connections["default"]
	return (*ic).Insert(query, bindings)
}
func (dm *DatabaseManager) Update(query string, bindings []interface{}) (sql.Result, error) {
	ic := dm.Connections["default"]
	return (*ic).Update(query, bindings)
}
func (dm *DatabaseManager) Delete(query string, bindings []interface{}) (sql.Result, error) {
	ic := dm.Connections["default"]
	return (*ic).Delete(query, bindings)
}
func (dm *DatabaseManager) Statement(query string, bindings []interface{}) (sql.Result, error) {
	ic := dm.Connections["default"]
	return (*ic).Delete(query, bindings)
}
func (dm *DatabaseManager) Query() *Builder {
	defaultConn := dm.getDefaultConnection()
	c := dm.Connection(defaultConn)
	builder := NewBuilder(c)
	builder.Grammar = &MysqlGrammar{}
	builder.Grammar.SetTablePrefix(dm.Configs[defaultConn].Prefix)
	builder.Grammar.SetBuilder(builder)
	return builder
}
func (dm *DatabaseManager) Transaction(closure TxClosure) (interface{}, error) {
	ic := dm.Connections["default"]
	return (*ic).Transaction(closure)
}
func (dm *DatabaseManager) BeginTransaction() (*Transaction, error) {
	ic := dm.Connections["default"]
	return (*ic).BeginTransaction()
}

func (dm *DatabaseManager) Create(model interface{}) (res sql.Result, err error) {
	return dm.Save(model)
}

/*Save
save/update single model
*/
func (dm *DatabaseManager) Save(modelP interface{}) (res sql.Result, err error) {
	//TODO: use connection as proxy
	//TODO: batch create/update/delete
	parsed := GetParsedModel(modelP)
	model := reflect.Indirect(reflect.ValueOf(modelP))
	if parsed.IsEloquent {
		ininted := !model.Field(parsed.PivotFieldIndex[0]).IsZero()
		if ininted {
			t := model.Field(parsed.PivotFieldIndex[0]).Interface().(*EloquentModel)
			return t.Save()
		} else {
			return InitModel(modelP).Create()
		}
	} else {

	}
	return
}
func (dm *DatabaseManager) Boot(modelP interface{}) *EloquentModel {
	parsed := GetParsedModel(modelP)
	model := reflect.Indirect(reflect.ValueOf(modelP))
	if parsed.IsEloquent {
		ininted := !model.Field(parsed.PivotFieldIndex[0]).IsZero()
		if !ininted {
			return InitModel(modelP)
		}
		return model.Field(parsed.PivotFieldIndex[0]).Interface().(*EloquentModel)
	}
	return nil
}
