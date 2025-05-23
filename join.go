package goeloquent

type JoinBuilder struct {
	Table   string
	Lateral bool
	Type    string
	*QueryBuilder
}

func NewJoinBuilder() *JoinBuilder {
	return &JoinBuilder{}
}
