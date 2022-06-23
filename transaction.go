package goeloquent

import "database/sql"

type Transaction struct {
	Tx             *sql.Tx
	Config         *DBConfig
	ConnectionName string
}
type TxClosure func(tx *Transaction) (interface{}, error)

func (t *Transaction) Table(tableName string) *Builder {
	builder := NewTxBuilder(t)
	builder.Grammar = &MysqlGrammar{}
	builder.Grammar.SetTablePrefix(t.Config.Prefix)
	builder.Tx = t
	builder.Grammar.SetBuilder(builder)
	builder.From(tableName)
	return builder
}

func (t *Transaction) Select(query string, bindings []interface{}, dest interface{}, mapping map[string]interface{}) (result sql.Result, err error) {
	var stmt *sql.Stmt
	var rows *sql.Rows
	stmt, err = t.Tx.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	rows, err = t.Tx.Stmt(stmt).Query(bindings...)
	if err != nil {
		return
	}

	defer rows.Close()

	return ScanAll(rows, dest, mapping), nil
}
func (t *Transaction) AffectingStatement(query string, bindings []interface{}) (result sql.Result, err error) {

	stmt, errP := t.Tx.Prepare(query)
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
func (t *Transaction) Insert(query string, bindings []interface{}) (sql.Result, error) {
	return t.AffectingStatement(query, bindings)
}

func (t *Transaction) Update(query string, bindings []interface{}) (sql.Result, error) {
	return t.AffectingStatement(query, bindings)
}

func (t *Transaction) Delete(query string, bindings []interface{}) (sql.Result, error) {
	return t.AffectingStatement(query, bindings)
}

func (t *Transaction) Statement(query string, bindings []interface{}) (sql.Result, error) {
	return t.AffectingStatement(query, bindings)
}

func (t *Transaction) Commit() error {
	return t.Tx.Commit()
}

func (t *Transaction) RollBack() error {
	return t.Tx.Rollback()
}

func (t *Transaction) Query() *Builder {
	builder := NewTxBuilder(t)
	builder.Grammar = &MysqlGrammar{}
	builder.Grammar.SetBuilder(builder)
	builder.Grammar.SetTablePrefix(t.Config.Prefix)
	return builder
}

func (t *Transaction) Model(model interface{}) *Builder {
	parsed := GetParsedModel(model)
	var connectionName string
	if len(parsed.ConnectionName) > 0 {
		connectionName = parsed.ConnectionName
	} else {
		connectionName = Eloquent.getDefaultConnection()
	}
	builder := NewTxBuilder(t)
	builder.Tx = t
	builder.Grammar = &MysqlGrammar{}
	builder.Grammar.SetTablePrefix(Eloquent.Configs[connectionName].Prefix)
	builder.Grammar.SetBuilder(builder)
	builder.SetModel(model)
	return builder
}
