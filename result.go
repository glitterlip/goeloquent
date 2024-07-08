package goeloquent

import (
	"database/sql"
	"time"
)

type Result struct {
	Count    int64 //count of rows query returned
	Raw      sql.Result
	Error    error
	Sql      string
	Bindings []interface{}
	Time     time.Duration
}

func (s *Result) LastInsertId() (int64, error) {
	return s.Raw.LastInsertId()
}

func (s *Result) RowsAffected() (int64, error) {
	return s.Raw.RowsAffected()
}
