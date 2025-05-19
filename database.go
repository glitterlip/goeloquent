package goeloquent

import (
	"database/sql"
	"errors"
	"sync"
)

const DefaultConnectionName = "default"

type DatabaseManager struct {
	Connections sync.Map
	Models      sync.Map
	MorphMap    sync.Map
	Listeners   map[string][]func(string, ...interface{}) bool
}

func (m *DatabaseManager) Conn(name string) (Connection, error) {
	if conn, ok := m.Connections.Load(name); ok {
		return conn.(*MysqlConnection), nil
	}
	return nil, errors.New("connection [" + name + "] not found")
}
func (m *DatabaseManager) DefaultConnection() (Connection, error) {
	if conn, ok := m.Connections.Load(DefaultConnectionName); ok {
		return conn.(*MysqlConnection), nil
	}
	return nil, errors.New("default connection not found")
}
func (m *DatabaseManager) Listen(name string, listener func(string, ...interface{}) bool) {
	if _, ok := m.Listeners[name]; !ok {
		m.Listeners[name] = []func(string, ...interface{}) bool{}
	}
	m.Listeners[name] = append(m.Listeners[name], listener)
}
func (m *DatabaseManager) DB(name ...string) (*sql.DB, error) {
	if len(name) > 0 {
		if conn, ok := m.Connections.Load(name[0]); ok {
			return conn.(*MysqlConnection).GetDB(), nil
		}
		return nil, errors.New("connection [" + name[0] + "] not found")
	}
	if conn, ok := m.Connections.Load(DefaultConnectionName); ok {
		return conn.(*MysqlConnection).GetDB(), nil
	}
	return nil, errors.New("default connection not found")
}
func (m *DatabaseManager) Query() *Statement {

	stmt := NewStatement()
	NewQueryBuilder(stmt)

	return stmt
}
func (m *DatabaseManager) Table(name ...string) *Statement {

	qb := m.Query()
	qb.Table(name...)

	return qb.Statement
}
func (m *DatabaseManager) Model(model ...interface{}) *Statement {

	eb := NewEloquentBuilder(NewQueryBuilder(NewStatement()))
	eb.SetModel(model...)

	return eb.QueryBuilder.Statement
}
