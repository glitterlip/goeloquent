package goeloquent

import "database/sql"

type IConnection interface {
	Table(tableName string) *Builder
	Model(modelPointer interface{}) *Builder
	Select(query string, bindings []interface{}, dest interface{}) (sql.Result, error)
	Insert(query string, bindings []interface{}) (sql.Result, error)
	Update(query string, bindings []interface{}) (sql.Result, error)
	Delete(query string, bindings []interface{}) (sql.Result, error)
	Statement(query string, bindings []interface{}) (sql.Result, error)
	BeginTransaction()
	Commit() error
	RollBack()
	GetDB() *sql.DB
	//GetGrammar() IGrammar
	//SetGrammar(IGrammar)
	//Run()
}
