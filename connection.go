package goeloquent

import (
	"database/sql"
	"errors"
	"time"
)

type Connection struct {
	DB             *sql.DB
	Config         *DBConfig
	ConnectionName string
}

func (c *Connection) Select(query string, bindings []interface{}, dest interface{}, mapping map[string]interface{}) (result Result, err error) {
	var stmt *sql.Stmt
	var rows *sql.Rows
	now := time.Now()
	stmt, err = c.DB.Prepare(query)
	result.Sql = query
	result.Bindings = bindings
	if err != nil {
		result.Error = err
		return
	}
	defer stmt.Close()
	rows, err = stmt.Query(bindings...)
	if err != nil {
		result.Error = err

		return
	}
	defer rows.Close()

	tr := ScanAll(rows, dest, mapping)
	result.Count = tr.Count
	result.Sql = query
	result.Bindings = bindings
	result.Time = time.Since(now)
	if result.Error != nil {
		err = errors.New(result.Error.Error())
	}
	DB.FireEvent(EventExecuted, result)

	return
}

func (c *Connection) BeginTransaction() (*Transaction, error) {
	begin, err := c.DB.Begin()
	if err != nil {
		return nil, errors.New(err.Error())
	}
	tx := &Transaction{
		Tx:             begin,
		ConnectionName: c.ConnectionName,
	}
	DB.FireEvent(EventTransactionBegin, tx)
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
			DB.FireEvent(EventTransactionRollback, err)
		} else {
			err = begin.Commit()
			DB.FireEvent(EventTransactionCommitted, err)
		}
	}()
	tx := &Transaction{
		Tx:             begin,
		ConnectionName: c.ConnectionName,
	}
	return closure(tx)
}

func (c *Connection) Insert(query string, bindings []interface{}) (result Result, err error) {
	return c.AffectingStatement(query, bindings)
}
func (c *Connection) Update(query string, bindings []interface{}) (result Result, err error) {
	return c.AffectingStatement(query, bindings)

}
func (c *Connection) Delete(query string, bindings []interface{}) (result Result, err error) {
	return c.AffectingStatement(query, bindings)
}

func (c *Connection) AffectingStatement(query string, bindings []interface{}) (result Result, err error) {
	result.Bindings = bindings
	result.Sql = query
	now := time.Now()
	stmt, errP := c.DB.Prepare(query)
	if errP != nil {
		err = errP
		result.Error = err
		return
	}
	defer stmt.Close()
	rawResult, err := stmt.Exec(bindings...)
	if err != nil {
		result.Error = err
		return
	}

	result = Result{
		Count:    0,
		Raw:      rawResult,
		Error:    nil,
		Sql:      query,
		Time:     time.Since(now),
		Bindings: bindings,
	}
	DB.FireEvent(EventExecuted, result)
	return
}

func (c *Connection) Table(tableName string) *Builder {
	builder := c.Query()
	builder.From(tableName)
	return builder
}

func (c *Connection) Statement(query string, bindings []interface{}) (Result, error) {
	return c.AffectingStatement(query, bindings)
}
func (c *Connection) GetDB() *sql.DB {
	return c.DB
}
func (c *Connection) Query() *Builder {
	return NewQueryBuilder(c)
}
func (c *Connection) GetConfig() *DBConfig {
	return c.Config
}
func (c *Connection) Model(model ...interface{}) *EloquentBuilder {
	return NewEloquentBuilder(model...)
}
