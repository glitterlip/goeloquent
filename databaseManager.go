package goeloquent

import (
	"database/sql"
	"fmt"
)

type DatabaseManager struct {
	Connections map[string]*IConnection
	Factory     ConnectionFactory
	Configs     map[string]DBConfig
}

func (dm *DatabaseManager) Connection(connectionName string) *IConnection {
	connection, ok := dm.Connections[connectionName]
	if !ok {
		dm.Connections[connectionName] = dm.MakeConnection(connectionName)
		connection, _ = dm.Connections[connectionName]
	}
	return connection
}
func (dm *DatabaseManager) Conn(connectionName string) IConnection {
	return *(dm.Connection(connectionName))
}
func (dm *DatabaseManager) getDefaultConnection() (defaultConnectionName string) {
	defaultConnectionName = "default"
	return
}

func (dm *DatabaseManager) MakeConnection(connectionName string) *IConnection {
	config, ok := dm.Configs[connectionName]
	if !ok {
		panic(fmt.Sprintf("Database connection %s not configured.", connectionName))
	}

	conn := dm.Factory.Make(config)
	dm.Connections[connectionName] = &conn
	return &conn
}
func (dm *DatabaseManager) Table(params ...string) *Builder {
	defaultConn := dm.getDefaultConnection()
	c := dm.Connection(defaultConn)
	builder := NewBuilder(*c)
	builder.Grammar = &MysqlGrammar{}
	builder.Grammar.SetTablePrefix(dm.Configs[defaultConn].Prefix)
	builder.Grammar.SetBuilder(builder)
	builder.From(params...)
	return builder

}
func (dm *DatabaseManager) Model(model interface{}) *Builder {
	parsed := GetParsedModel(model)
	var connectionName string
	if len(parsed.ConnectionName) > 0 {
		connectionName = parsed.ConnectionName
	} else {
		connectionName = dm.getDefaultConnection()
	}
	c := dm.Connection(connectionName)
	builder := NewBuilder(*c)
	builder.Grammar = &MysqlGrammar{}
	builder.Grammar.SetTablePrefix(dm.Configs[connectionName].Prefix)
	builder.Grammar.SetBuilder(builder)
	builder.SetModel(model)
	return builder
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
