package goeloquent

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

type Statement struct {
	Context    context.Context
	Connection Connection
	Tx         *sql.Tx
	Error      error
	Result     sql.Result
	Rows       int64
	*QueryBuilder
	Eloquent   *EloquentBuilder
	Pretending bool
	Dest       interface{}
	DestValue  reflect.Value
}
type StatementFunc func(*Statement)
type StatementChainFunc func(*Statement) *Statement
type ScopeFunc = func(*Statement)
type RelationFunc = func(*Statement)
type RelationAggregate struct {
	FuncName          string
	Column            string
	RelationFieldName string
	Constraint        StatementChainFunc
}
