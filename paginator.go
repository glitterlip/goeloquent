package goeloquent

import "math"

type Paginator struct {
	Items       interface{}
	Total       int64
	PerPage     int64
	CurrentPage int64
}

func (p Paginator) LastPage() int64 {
	return int64(int(math.Ceil(float64(p.Total / p.PerPage))))
}
