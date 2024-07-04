package goeloquent

import "math"

type Paginator struct {
	Items       interface{} `json:"items"`
	Total       int64       `json:"total"`
	PerPage     int64       `json:"per_page"`
	CurrentPage int64       `json:"current_page"`
}

func (p Paginator) LastPage() int64 {
	return int64(int(math.Ceil(float64(p.Total / p.PerPage))))
}
func (p Paginator) GetItems() interface{} {
	return p.Items
}
func (p Paginator) PageSize() int64 {
	return p.PerPage
}

func (p Paginator) Page() int64 {
	return p.CurrentPage
}

func (p Paginator) Count() int64 {
	return p.Total
}
