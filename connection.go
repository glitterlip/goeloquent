package goeloquent

import (
	"database/sql"
	"errors"
	_ "reflect"
	_ "time"
)

type Connection struct {
	DB             *sql.DB
	Config         *DBConfig
	ConnectionName string
}

type IConnection interface {
	Insert(query string, bindings []interface{}) (result sql.Result, err error)
	Select(query string, bindings []interface{}, dest interface{}, mapping map[string]interface{}) (result sql.Result, err error)
	Update(query string, bindings []interface{}) (result sql.Result, err error)
	Delete(query string, bindings []interface{}) (result sql.Result, err error)
	AffectingStatement(query string, bindings []interface{}) (result sql.Result, err error)
	Statement(query string, bindings []interface{}) (sql.Result, error)
	Table(tableName string) *Builder
}

type Preparer interface {
	Prepare(query string) (*sql.Stmt, error)
}
type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}
type ITransaction interface {
	BeginTransaction() (*Transaction, error)
	Transaction(closure TxClosure) (interface{}, error)
}

func (c *Connection) Select(query string, bindings []interface{}, dest interface{}, mapping map[string]interface{}) (result sql.Result, err error) {
	var stmt *sql.Stmt
	var rows *sql.Rows
	stmt, err = c.DB.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	rows, err = stmt.Query(bindings...)
	if err != nil {
		return
	}
	defer rows.Close()

	return ScanAll(rows, dest, mapping), nil
}

func (c *Connection) BeginTransaction() (*Transaction, error) {
	begin, err := c.DB.Begin()
	if err != nil {
		return nil, errors.New(err.Error())
	}
	tx := &Transaction{
		Tx:             begin,
		Config:         c.Config,
		ConnectionName: c.ConnectionName,
	}
	return tx, nil
}
func (c *Connection) Transaction(closure TxClosure) (res interface{}, err error) {
	begin, err := c.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if result := recover(); result != nil {
			if e, ok := result.(error); ok {
				err = e
			} else {
				err = errors.New("error occurred during transaction")
			}
			_ = begin.Rollback()
		} else {
			err = begin.Commit()
		}
	}()
	tx := &Transaction{
		Tx:             begin,
		Config:         c.Config,
		ConnectionName: c.ConnectionName,
	}
	return closure(tx)
}

func (c *Connection) Insert(query string, bindings []interface{}) (result sql.Result, err error) {
	return c.AffectingStatement(query, bindings)
}
func (c *Connection) Update(query string, bindings []interface{}) (result sql.Result, err error) {
	return c.AffectingStatement(query, bindings)

}
func (c *Connection) Delete(query string, bindings []interface{}) (result sql.Result, err error) {
	return c.AffectingStatement(query, bindings)
}

func (c *Connection) AffectingStatement(query string, bindings []interface{}) (result sql.Result, err error) {
	stmt, errP := c.DB.Prepare(query)
	if errP != nil {
		err = errP
		return
	}
	defer stmt.Close()
	result, err = stmt.Exec(bindings...)
	if err != nil {
		return
	}

	return
}

func (c *Connection) Table(tableName string) *Builder {
	builder := NewBuilder(c)
	builder.Grammar = &MysqlGrammar{}
	builder.Grammar.SetTablePrefix(c.Config.Prefix)
	builder.Grammar.SetBuilder(builder)
	builder.From(tableName)
	return builder
}

func (c *Connection) Model(modelPointer interface{}) *Builder {
	builder := NewBuilder(c)
	builder.Grammar = &MysqlGrammar{}
	builder.Grammar.SetTablePrefix(c.Config.Prefix)
	builder.Grammar.SetBuilder(builder)
	builder.SetModel(modelPointer)
	return builder
}
func (c *Connection) Statement(query string, bindings []interface{}) (sql.Result, error) {
	return c.AffectingStatement(query, bindings)
}
func (c *Connection) GetDB() *sql.DB {
	return c.DB
}
func (c *Connection) Query() *Builder {
	builder := NewBuilder(c)
	builder.Grammar = &MysqlGrammar{}
	builder.Grammar.SetBuilder(builder)
	builder.Grammar.SetTablePrefix(c.Config.Prefix)
	return builder
}
func (c *Connection) GetConfig() *DBConfig {
	return c.Config
}
