package goeloquent

import (
	"database/sql"
)

type Transaction struct {
	*sql.Tx
	ConnectionName string
	*Connection
}

type TxClosure func(tx *Transaction) (Result, error)

func (t *Transaction) Select(query string, bindings []interface{}, dest interface{}, mapping map[string]interface{}) (result Result, err error) {
	var stmt *sql.Stmt
	var rows *sql.Rows
	stmt, err = t.Tx.Prepare(query)
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

func (t *Transaction) Insert(query string, bindings []interface{}) (result Result, err error) {
	return t.AffectingStatement(query, bindings)
}
func (t *Transaction) Update(query string, bindings []interface{}) (result Result, err error) {
	return t.AffectingStatement(query, bindings)

}
func (t *Transaction) Delete(query string, bindings []interface{}) (result Result, err error) {
	return t.AffectingStatement(query, bindings)
}

func (t *Transaction) AffectingStatement(query string, bindings []interface{}) (result Result, err error) {
	stmt, errP := t.Tx.Prepare(query)
	if errP != nil {
		err = errP
		return
	}
	defer stmt.Close()
	rawResult, err := stmt.Exec(bindings...)
	if err != nil {
		return
	}

	return Result{
		Count: 0,
		Raw:   rawResult,
		Error: nil,
	}, nil
}

func (t *Transaction) Table(tableName string) *Builder {
	builder := t.Query()
	builder.From(tableName)
	return builder
}

func (t *Transaction) Statement(query string, bindings []interface{}) (Result, error) {
	return t.AffectingStatement(query, bindings)
}

func (t *Transaction) Query() *Builder {
	return NewTxBuilder(t)
}
func (t *Transaction) GetConfig() *DBConfig {
	return t.Config
}
func NewTxBuilder(tx *Transaction) *Builder {
	b := Builder{
		Components: make(map[string]struct{}),
		Tx:         tx,
		Bindings:   make(map[string][]interface{}),
	}
	b.Grammar = &MysqlGrammar{}
	b.Grammar.SetTablePrefix(tx.Config.Prefix)
	b.Grammar.SetBuilder(&b)
	return &b
}

func (t *Transaction) Commit() error {

	err := t.Tx.Commit()
	DB.FireEvent(EventTransactionCommitted, err)
	return err
}
func (t *Transaction) Rollback() error {
	err := t.Tx.Rollback()
	DB.FireEvent(EventTransactionRollback, err)
	return err
}
func (t *Transaction) Model(model interface{}) *EloquentBuilder {
	return NewEloquentBuilder(model)
}
