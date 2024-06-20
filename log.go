package goeloquent

import (
	"time"
)

type Log struct {
	SQL      string
	Bindings []interface{}
	Result   Result
	Time     time.Duration
}
