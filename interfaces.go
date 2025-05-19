package goeloquent

import "database/sql"

type Connection interface {
	GetDB() *sql.DB
	Table(table interface{}, alias ...string) *Statement
	Query() *Statement
	Raw(value string) Expression
	SelectOne(query string, bindings interface{}) (*Statement, error)
	Select(query string, bindings interface{}) (*Statement, error)
	Insert(query string, bindings interface{}) (*Statement, error)
	Update(query string, bindings interface{}) (*Statement, error)
	Delete(query string, bindings interface{}) (*Statement, error)
	Statement(query string, bindings interface{}) (*Statement, error)
	AffectingStatement(query string, bindings interface{}) (*Statement, error)
	Transaction(tx func(*Statement) error, options ...*sql.TxOptions) error
	BeginTransaction(options ...*sql.TxOptions) *Statement
	Commit() *Statement
	Rollback() *Statement
	//RollbackTo(name string) //todo
	//SavePoint(name string) //todo
}
type Grammar interface {
	WrapTable(table interface{}, prefix ...string) string
	Wrap(value interface{}) string
	WrapAliasedValue(value interface{}) string
	WrapAliasedTable(table string, prefix ...string) string
	WrapSegments(segments []string) string
	WrapValue(value string) string
	WrapJsonSelector(value string) string
	Columnize(value []interface{}) string
	Parameter(value interface{}) string
	Parameterize(value []interface{}) string
	QuoteString(value interface{}) string
}
