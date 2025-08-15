package goeloquent

import "errors"

var (
	ErrorUpsupportedDriver      = errors.New("unsupported driver")
	ErrorNoValueAndOperators    = errors.New("no value and operators provided for query builder")
	ErrorInvalidOrder           = errors.New("invalid direction for order by")
	ErrorNotFound               = errors.New("record not found in database")
	ErrorSubQueryInvalid        = errors.New("sub query must be a string,expression,querybuilder or func that returns a querybuilder")
	ErrorChunkWithoutOrder      = errors.New("chunk requires at least one orderby clause")
	ErrorNotPtr                 = errors.New("destination must be a pointer")
	ErrorWhereRowValuesMismatch = errors.New("the number of columns must match the number of values")
	ErrorModelNotInitialized    = errors.New("model not initialized, call Init(model) first")
	ErrorModelPointerError      = errors.New("model pointer is nil or not a pointer to a struct")
)
