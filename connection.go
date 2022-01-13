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

func (c *Connection) Select(query string, bindings []interface{}, dest interface{}) (result sql.Result, err error) {
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

	return ScanAll(rows, dest), nil
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
func (c *Connection) Transaction(closure TxClosure) (interface{}, error) {
	begin, err := c.DB.Begin()
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		if err := recover(); err != nil {
			_ = begin.Rollback()
			panic(err)
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
