package goeloquent

type Expression string

func Raw(expr string) Expression {
	return Expression(expr)
}
