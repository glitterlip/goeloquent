package goeloquent

import (
	"database/sql"
	_ "reflect"
	_ "time"
)

type MysqlConnection struct {
	Connection     *sql.DB
	Config         DBConfig
	ConnectionName string
	Transactions   int
	Tx             *sql.Tx
}

func (c *MysqlConnection) Select(query string, bindings []interface{}, dest interface{}) (result sql.Result, err error) {

	var stmt *sql.Stmt
	var rows *sql.Rows
	if c.Tx != nil {
		stmt, err = c.Tx.Prepare(query)
		defer stmt.Close()
		if err != nil {
			return
		}
		rows, err = c.Tx.Stmt(stmt).Query(bindings...)
		if err != nil {
			return
		}
	} else {
		stmt, err = c.Connection.Prepare(query)
		defer stmt.Close()
		if err != nil {
			return
		}
		rows, err = stmt.Query(bindings...)
		if err != nil {
			return
		}
	}
	defer rows.Close()

	return ScanAll(rows, dest), nil
}

func (c *MysqlConnection) BeginTransaction() {
	if c.Tx != nil {
		panic("recreating transaction, nested transactions is not supported")
	} else {
		begin, err := c.Connection.Begin()
		if err != nil {
			panic(err.Error())
		}
		c.Transactions += 1
		c.Tx = begin
	}
}

func (c *MysqlConnection) Commit() error {
	var err error
	if c.Transactions > 0 {
		err = c.Tx.Commit()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *MysqlConnection) RollBack() {
	_ = c.Tx.Rollback()
}

func (c *MysqlConnection) Insert(query string, bindings []interface{}) (result sql.Result, err error) {
	return c.AffectingStatement(query, bindings)
}
func (c *MysqlConnection) Update(query string, bindings []interface{}) (result sql.Result, err error) {
	return c.AffectingStatement(query, bindings)

}
func (c *MysqlConnection) Delete(query string, bindings []interface{}) (result sql.Result, err error) {
	return c.AffectingStatement(query, bindings)

}

func (c *MysqlConnection) AffectingStatement(query string, bindings []interface{}) (result sql.Result, err error) {

	if c.Tx != nil {
		stmt, errT := c.Tx.Prepare(query)
		if errT != nil {
			err = errT
			return
		}
		defer stmt.Close()
		result, err = c.Tx.Stmt(stmt).Exec(bindings...)
		if err != nil {
			return
		}
	} else {
		stmt, errP := c.Connection.Prepare(query)
		if errP != nil {
			err = errP
			return
		}
		defer stmt.Close()
		result, err = stmt.Exec(bindings...)
		if err != nil {
			return
		}
	}

	return
}
