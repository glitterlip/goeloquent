package goeloquent

type Paginator struct {
	Items       interface{}
	Total       int64
	PerPage     int64
	CurrentPage int64
}
