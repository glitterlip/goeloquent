package goeloquent

import (
	"database/sql"
	"errors"
	"sync"
)

const DefaultConnectionName = "default"

type DatabaseManager struct {
	Connections  sync.Map
	Listeners    map[EventName][]func(EventName, ...interface{}) bool
	ParsedModels sync.Map //store parsed models map[modelName]*ModelConfig
	MorphMaps    sync.Map //convert string from database column to model for relation
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
	return nil, ErrorConnectionNotSet
}

/*
Listen registers a listener for the specified event.

If the listener is nil, it removes the listener for that event.

listeners will stop if error happens or return a false
*/
func (m *DatabaseManager) Listen(name EventName, listener func(EventName, ...interface{}) bool) {
	if listener == nil {
		delete(m.Listeners, name)
		return
	}
	if _, ok := m.Listeners[name]; !ok {
		m.Listeners[name] = []func(EventName, ...interface{}) bool{}
	}
	m.Listeners[name] = append(m.Listeners[name], listener)
}
func (m *DatabaseManager) Fire(name EventName, args ...interface{}) (stop bool) {

	defer func() {
		if r := recover(); r != nil {
			stop = false
		}
	}()
	stop = true
	if listeners, ok := m.Listeners[name]; ok {
		for _, listener := range listeners {
			if !listener(name, args...) {
				stop = false
				break
			}
		}
	}
	return stop
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
	return nil, ErrorConnectionNotSet
}
func (m *DatabaseManager) Query() *Statement {
	conn, err := m.DefaultConnection()
	stmt := NewStatement(conn)
	NewQueryBuilder(stmt)
	if err != nil {
		stmt.AddError(err)
	}
	return stmt
}
func (m *DatabaseManager) Table(name ...string) *QueryBuilder {

	qb := m.Query()
	qb.Table(name[0], name[1:]...)

	return qb.QueryBuilder
}
func (m *DatabaseManager) Model(model ...interface{}) *EloquentBuilder {

	eb := NewEloquentBuilder(NewQueryBuilder(NewStatement()))
	eb.SetModel(model...)

	return eb.Eloquent
}

func (m *DatabaseManager) SetMorphMaps(maps map[string]interface{}) {
	for k, v := range maps {
		if config, ok := v.(*ModelConfig); ok {
			m.MorphMaps.Store(k, config)
		} else {
			config, err := ParseModel(v)
			if err == nil {
				m.MorphMaps.Store(k, config)
			}
		}
	}
}
