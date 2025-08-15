package goeloquent

type JoinBuilder struct {
	Table   interface{}
	Lateral bool
	Type    JoinType
	RawSql  string
	*QueryBuilder
}

func NewJoin(parent *QueryBuilder, joinType JoinType, table interface{}) *JoinBuilder {
	stmt := NewStatement()
	b := NewQueryBuilder(stmt)
	b.Grammar = parent.Grammar
	b.From(table)
	b.IsJoin = true
	return &JoinBuilder{
		Type:         joinType,
		Table:        table,
		QueryBuilder: b,
	}
}

/*
Join Add a join clause to the query.

 1. Join("contacts", "users.id", "contacts.user_id")

    inner join contacts on users.id = contacts.user_id
*/
func (q *QueryBuilder) Join(table interface{}, params ...interface{}) *QueryBuilder {
	q.Components[COMPONENT_JOIN] = struct{}{}
	first, operator, second, joinType, isWhere := PrepareJoinParams(params...)
	if function, ok := first.(func(*JoinBuilder)); ok {
		clause := NewJoin(q, joinType, table)
		function(clause)
		q.Joins = append(q.Joins, clause)
		q.AddBinding(clause.GetBindings(), COMPONENT_JOIN)
		return q
	} else if function, ok := first.(func(*JoinBuilder) *JoinBuilder); ok {
		clause := NewJoin(q, joinType, table)
		clause = function(clause)
		q.Joins = append(q.Joins, clause)
		q.AddBinding(clause.GetBindings(), COMPONENT_JOIN)
		return q
	}

	join := NewJoin(q, joinType, table)

	if isWhere {
		join.Where(first, operator, second)
	} else {
		join.On(first, operator, second)
	}

	q.Joins = append(q.Joins, join)
	q.AddBinding(join.GetBindings(), COMPONENT_JOIN)

	return q
}

/*
JoinWhere Add a "join where" clause to the query.
*/
func (q *QueryBuilder) JoinWhere(table interface{}, params ...interface{}) *QueryBuilder {
	first, operator, second, joinType, isWhere := PrepareJoinParams(append(params, true)...)
	return q.Join(table, first, operator, second, joinType, isWhere)
}

/*
JoinSub Add a subquery join clause to the query.
*/
func (q *QueryBuilder) JoinSub(query interface{}, as, first, operator, second interface{}, params ...interface{}) *QueryBuilder {
	sql, bindings := q.CreateSub(query)
	var joinType JoinType
	var isWhere bool
	switch len(params) {
	case 0:
		joinType = JoinTypeInner
	case 1:
		joinType = params[0].(JoinType)
	case 2:
		joinType = params[0].(JoinType)
		isWhere = params[1].(bool)
	}
	if len(bindings) > 0 {
		q.AddBinding(bindings, COMPONENT_JOIN)
	}
	return q.Join(Raw("("+sql+") as "+q.Grammar.WrapTable(as)), first, operator, second, joinType, isWhere)
}

/*
JoinLateral Add a lateral join clause to the query.
*/
func (q *QueryBuilder) JoinLateral(query interface{}, as string, joinType JoinType) *QueryBuilder {
	sql, bindings := q.CreateSub(query)
	if len(bindings) > 0 {
		q.AddBinding(bindings, COMPONENT_JOIN)
	}
	join := NewJoin(q, joinType, Raw("("+sql+") as "+q.Grammar.WrapTable(as)))
	join.Lateral = true
	q.Joins = append(q.Joins, join)
	q.Components[COMPONENT_JOIN] = struct{}{}
	return q

}

/*
LeftJoinLateral Add a lateral left join to the query.
*/
func (q *QueryBuilder) LeftJoinLateral(query interface{}, as string) *QueryBuilder {
	return q.JoinLateral(query, as, JoinTypeLeft)
}

/*
LeftJoin Add a left join to the query.
LeftJoin(table,firstColumn,operator,secondColumn)
*/
func (q *QueryBuilder) LeftJoin(table interface{}, params ...interface{}) *QueryBuilder {
	first, operator, second, joinType, isWhere := PrepareJoinParams(append(params, JoinTypeLeft, false)...)
	return q.Join(table, first, operator, second, joinType, isWhere)
}

/*
LeftJoinWhere Add a "left join where" clause to the query.
*/
func (q *QueryBuilder) LeftJoinWhere(table interface{}, params ...interface{}) *QueryBuilder {
	first, operator, second, joinType, isWhere := PrepareJoinParams(append(params, JoinTypeLeft, true)...)
	return q.Join(table, first, operator, second, joinType, isWhere)
}

/*
LeftJoinSub Add a subquery left join clause to the query.
*/
func (q *QueryBuilder) LeftJoinSub(query, as, first, operator, second interface{}) *QueryBuilder {
	return q.JoinSub(query, as, first, operator, second, JoinTypeLeft)
}

/*
RightJoinLateral Add a lateral left join to the query.
*/
func (q *QueryBuilder) RightJoinLateral(query interface{}, as string) *QueryBuilder {
	return q.JoinLateral(query, as, JoinTypeRight)
}

/*
RightJoin Add a right join to the query.
*/
func (q *QueryBuilder) RightJoin(table interface{}, params ...interface{}) *QueryBuilder {
	first, operator, second, joinType, isWhere := PrepareJoinParams(append(params, JoinTypeRight, false)...)
	return q.Join(table, first, operator, second, joinType, isWhere)
}

/*
RightJoinWhere Add a "right join where" clause to the query.
*/
func (q *QueryBuilder) RightJoinWhere(table interface{}, params ...interface{}) *QueryBuilder {
	first, operator, second, joinType, isWhere := PrepareJoinParams(append(params, JoinTypeRight, true)...)
	return q.Join(table, first, operator, second, joinType, isWhere)
}

/*
RightJoinSub Add a subquery right join clause to the query.
*/
func (q *QueryBuilder) RightJoinSub(query interface{}, as, first, operator, second interface{}) *QueryBuilder {
	return q.JoinSub(query, as, first, operator, second, JoinTypeRight)
}

/*
CrossJoin Add a cross join to the query.
*/
func (q *QueryBuilder) CrossJoin(table interface{}, params ...interface{}) *QueryBuilder {
	if len(params) > 0 {
		first, operator, second, joinType, isWhere := PrepareJoinParams(append(params, JoinTypeCross, false)...)
		return q.Join(table, first, operator, second, joinType, isWhere)
	}
	q.Joins = append(q.Joins, NewJoin(q, JoinTypeCross, table))
	q.Components[COMPONENT_JOIN] = struct{}{}
	return q
}

/*
CrossJoinSub Add a subquery cross join clause to the query.
*/
func (q *QueryBuilder) CrossJoinSub(query *QueryBuilder, as string) *QueryBuilder {
	q.Components[COMPONENT_JOIN] = struct{}{}
	sql, bindings := q.CreateSub(query)
	if len(bindings) > 0 {
		q.AddBinding(bindings, COMPONENT_JOIN)
	}
	q.Joins = append(q.Joins, NewJoin(q, JoinTypeCross, Raw("("+sql+") as "+q.Grammar.WrapTable(as))))

	return q
}

/*
On Add an "on" clause to the query.
*/
func (j *JoinBuilder) On(first interface{}, params ...interface{}) *JoinBuilder {
	switch first.(type) {
	case func(*JoinBuilder):
		boolean := "and"
		if len(params) > 0 {
			boolean = params[0].(string)
		}
		j.WhereNested(first, boolean)
		return j
	case func(*JoinBuilder) *JoinBuilder:
		boolean := "and"
		if len(params) > 0 {
			boolean = params[0].(string)
		}
		j.WhereNested(first, boolean)
		return j
	}

	j.WhereColumn(first, params...)
	return j
}

func (j *JoinBuilder) OrOn(first interface{}, params ...interface{}) *JoinBuilder {
	switch first.(type) {
	case func(*JoinBuilder):
	case func(*JoinBuilder) *JoinBuilder:
		boolean := "or"
		if len(params) > 0 {
			boolean = params[0].(string)
		}
		j.WhereNested(first, boolean)
		return j
	}

	j.WhereColumn(first, append(params, "or")...)
	return j
}

func PrepareJoinParams(params ...interface{}) (interface{}, string, interface{}, JoinType, bool) {
	var first, second interface{}
	var operator string
	joinType := JoinTypeInner
	var isWhere bool
	var count int
	var hasIsWhere, hasJoinType bool
	for _, param := range params {
		if boolean, ok := param.(bool); ok {
			hasIsWhere = true
			isWhere = boolean
			count++
		} else if str, ok := param.(JoinType); ok {
			hasJoinType = true
			joinType = str
			count++
		}

	}
	switch len(params) - count {
	case 0:
		first = ""
		operator = "="
		second = ""
		if !hasIsWhere {
			isWhere = false
		}
		if !hasJoinType {
			joinType = JoinTypeInner
		}
	case 1:
		first = params[0]
		operator = "="
		second = ""
		if !hasIsWhere {
			isWhere = false
		}
		if !hasJoinType {
			joinType = JoinTypeInner
		}
	case 2:
		first = params[0]
		operator = "="
		second = params[1].(string)
		if !hasIsWhere {
			isWhere = false
		}
		if !hasJoinType {
			joinType = JoinTypeInner
		}
	case 3:
		first = params[0]
		operator = params[1].(string)
		second = params[2]
		if !hasIsWhere {
			isWhere = false
		}
		if !hasJoinType {
			joinType = JoinTypeInner
		}
	}

	return first, operator, second, joinType, isWhere
}
