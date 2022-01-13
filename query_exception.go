package goeloquent

import (
	"fmt"
	"strings"
)

type QueryException struct {
	Err      error
	Sql      string
	Bindings []interface{}
}

func (e *QueryException) Error() string {
	var errString strings.Builder
	errString.Grow(1024)
	errString.WriteString("\x1b[91mOops🔥\x1b[39m \n\t")
	errString.WriteString(e.Err.Error())
	errString.WriteString(";\n\tError SQL: ")
	errString.WriteString(e.Sql)
	errString.WriteString("\n\tBindings:[")
	for k, v := range e.Bindings {
		switch s := v.(type) {
		case string:
			if k == 0 {
				errString.WriteString(s)
			} else {
				errString.WriteString(", ")
				errString.WriteString(s)
			}
		case int:
			if k == 0 {
				errString.WriteString(fmt.Sprintf("%v", s))
			} else {
				errString.WriteString(", ")
				errString.WriteString(fmt.Sprintf("%v", s))
			}
		default:
			if k == 0 {
				errString.WriteString("???")
			} else {
				errString.WriteString(", ???")
			}
		}
	}
	errString.WriteString("]")
	return errString.String()
}
