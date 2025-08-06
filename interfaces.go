package goeloquent

import (
	"database/sql"
)

type Connection interface {
	GetTablePrefix() string
	SetTablePrefix(prefix string) Connection
	GetDB() *sql.DB
	Query() *Statement
	Table(table interface{}, alias ...string) *Statement
	Model(model ...interface{}) *Statement
	Raw(value string) Expression
	SelectOne(*Statement) *Statement
	Select(*Statement) *Statement
	Insert(*Statement) *Statement
	Update(*Statement) *Statement
	Delete(*Statement) *Statement
	Statement(*Statement) *Statement
	AffectingStatement(*Statement) *Statement
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
	WrapAliasedValue(value string) string
	WrapAliasedTable(table string, prefix ...string) string
	WrapSegments(segments []string) string
	WrapValue(value string) string
	WrapJsonSelector(value string) string
	Columnize(value []interface{}) string
	Parameter(value interface{}) string
	Parameterize(value []interface{}) string
	QuoteString(value interface{}) string
	CompileRandom(seed ...interface{}) string
	CompileSelect(*QueryBuilder) string
	CompileInsert(*QueryBuilder, []map[string]interface{}) (string, []interface{})
	CompileUpdate(*QueryBuilder, map[string]interface{}) (string, []interface{})
	CompileDelete(*QueryBuilder) string
	GetError() error
	SetTablePrefix(prefix string)
	GetOperators() map[string]struct{}
}
