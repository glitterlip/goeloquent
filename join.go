package goeloquent

import (
	"errors"
	"fmt"
)

type JoinType string

type JoinBuilder struct {
	JoinType string // inner,left,right,full,cross
	Table    interface{}
	*Builder //parent builder
}

func NewJoin(parent *Builder, joinType string, table interface{}) *JoinBuilder {
	b := NewQueryBuilder(parent.Connection).From(table)
	b.IsJoin = true
	return &JoinBuilder{
		JoinType: joinType,
		Table:    table,
		Builder:  b,
	}
}

/*
Join Add a join clause to the query.

 1. Join("contacts", "users.id", "contacts.user_id")

    inner join contacts on users.id = contacts.user_id
*/
func (b *Builder) Join(table string, first interface{}, params ...interface{}) *Builder {
	var operator, second, joinType string
	var isWhere = false
	b.Components[TYPE_JOIN] = struct{}{}

	switch len(params) {
	case 0:
		if function, ok := first.(func(builder *Builder)); ok {
			clause := NewJoin(b, JOIN_TYPE_INNER, table)
			function(clause.Builder)
			b.Joins = append(b.Joins, clause)
			b.AddBinding(clause.GetBindings(), TYPE_JOIN)
			return b
		} else {
			panic(errors.New("arguements num mismatch"))
		}
	case 1:
		operator = "="
		second = params[0].(string)
		joinType = JOIN_TYPE_INNER
	case 2:
		operator = params[0].(string)
		second = params[1].(string)
		joinType = JOIN_TYPE_INNER
	case 3:
		operator = params[0].(string)
		second = params[1].(string)
		joinType = params[2].(string)
	case 4:
		operator = params[0].(string)
		second = params[1].(string)
		joinType = params[2].(string)
		isWhere = params[3].(bool)
	}
	return b.join(table, first, operator, second, joinType, isWhere)
}

/*
join Add a join clause to the query.
*/
func (b *Builder) join(table interface{}, first, operator, second, joinType, isWhere interface{}) *Builder {
	b.Components[TYPE_JOIN] = struct{}{}

	if function, ok := first.(func(builder *Builder)); ok {
		clause := NewJoin(b, joinType.(string), table)
		function(clause.Builder)
		b.Joins = append(b.Joins, clause)
		b.AddBinding(clause.GetBindings(), TYPE_JOIN)
		return b
	}

	clause := NewJoin(b, joinType.(string), table)

	if isWhere.(bool) {
		clause.Where(first, operator, second)
	} else {
		clause.On(first, operator, second, BOOLEAN_AND)
	}
	b.AddBinding(clause.GetBindings(), TYPE_JOIN)

	clause.Grammar.SetTablePrefix(b.Grammar.GetTablePrefix())

	b.Joins = append(b.Joins, clause)

	return b
}

/*
JoinWhere Add a "join where" clause to the query.

	JoinWhere("contacts", "users.id", "1") // inner join contacts on users.id = 1
	JoinWhere("contacts", "users.id", "=", "123") // inner join contacts on users.id = 123
	JoinWhere("contacts", "users.id", "=", "234", "or") // inner join contacts on users.id = 234 or
*/
func (b *Builder) JoinWhere(table, firstColumn, joinOperator, secondColumn string) *Builder {
	return b.joinWhere(table, firstColumn, joinOperator, secondColumn, JOIN_TYPE_INNER)
}
func (b *Builder) joinWhere(table, firstColumn, joinOperator, secondColumn, joinType string) *Builder {
	return b.join(table, firstColumn, joinOperator, secondColumn, joinType, true)
}

/*
JoinSub Add a subquery join clause to the query.
*/
func (b *Builder) JoinSub(query interface{}, as string, first interface{}, params ...interface{}) *Builder {
	queryStr, bindings := b.CreateSub(query)
	expr := fmt.Sprintf("(%s) as %s", queryStr, b.Grammar.WrapTable(as))
	var operator string
	joinType := JOIN_TYPE_INNER
	var isWhere = false
	var second interface{}
	switch len(params) {
	case 1:
		operator = "="
		second = params[0]
	case 2:
		operator = params[0].(string)
		second = params[1]
	case 3:
		operator = params[0].(string)
		second = params[1]
		joinType = params[2].(string)
	case 4:
		operator = params[0].(string)
		second = params[1]
		joinType = params[2].(string)
		isWhere = params[3].(bool)
	}
	b.AddBinding(bindings, TYPE_JOIN)

	return b.join(Raw(expr), first, operator, second, joinType, isWhere)
}

/*
JoinLateral Add a "lateral join" clause to the query. //TODO
$builder = $this->getMySqlBuilder();

	$builder->getConnection()->shouldReceive('getDatabaseName');
	$builder->from('users')->joinLateral('select * from `contacts` where `contracts`.`user_id` = `users`.`id`', 'sub');
	$this->assertSame('select * from `users` inner join lateral (select * from `contacts` where `contracts`.`user_id` = `users`.`id`) as `sub` on true', $builder->toSql());

	$builder = $this->getMySqlBuilder();
	$builder->getConnection()->shouldReceive('getDatabaseName');
	$builder->from('users')->joinLateral(function ($q) {
	    $q->from('contacts')->whereColumn('contracts.user_id', 'users.id');
	}, 'sub');
	$this->assertSame('select * from `users` inner join lateral (select * from `contacts` where `contracts`.`user_id` = `users`.`id`) as `sub` on true', $builder->toSql());

	$builder = $this->getMySqlBuilder();
	$builder->getConnection()->shouldReceive('getDatabaseName');
	$sub = $this->getMySqlBuilder();
	$sub->getConnection()->shouldReceive('getDatabaseName');
	$eloquentBuilder = new EloquentBuilder($sub->from('contacts')->whereColumn('contracts.user_id', 'users.id'));
	$builder->from('users')->joinLateral($eloquentBuilder, 'sub');
	$this->assertSame('select * from `users` inner join lateral (select * from `contacts` where `contracts`.`user_id` = `users`.`id`) as `sub` on true', $builder->toSql());

	$sub1 = $this->getMySqlBuilder();
	$sub1->getConnection()->shouldReceive('getDatabaseName');
	$sub1 = $sub1->from('contacts')->whereColumn('contracts.user_id', 'users.id')->where('name', 'foo');

	$sub2 = $this->getMySqlBuilder();
	$sub2->getConnection()->shouldReceive('getDatabaseName');
	$sub2 = $sub2->from('contacts')->whereColumn('contracts.user_id', 'users.id')->where('name', 'bar');

	$builder = $this->getMySqlBuilder();
	$builder->getConnection()->shouldReceive('getDatabaseName');
	$builder->from('users')->joinLateral($sub1, 'sub1')->joinLateral($sub2, 'sub2');

	$expected = 'select * from `users` ';
	$expected .= 'inner join lateral (select * from `contacts` where `contracts`.`user_id` = `users`.`id` and `name` = ?) as `sub1` on true ';
	$expected .= 'inner join lateral (select * from `contacts` where `contracts`.`user_id` = `users`.`id` and `name` = ?) as `sub2` on true';

	$this->assertEquals($expected, $builder->toSql());
	$this->assertEquals(['foo', 'bar'], $builder->getRawBindings()['join']);

	$this->expectException(InvalidArgumentException::class);
	$builder = $this->getMySqlBuilder();
	$builder->from('users')->joinLateral(['foo'], 'sub');
*/
func (b *Builder) JoinLateral(query string, as string, joinType string) *Builder {

	return b
}

/*
LeftJoinLateral Add a "left lateral join" clause to the query. //TODO
*/
func (b *Builder) LeftJoinLateral(query string, as string, joinType string) *Builder {

	return b

}

/*
LeftJoin Add a left join to the query.
*/
func (b *Builder) LeftJoin(table string, firstColumn interface{}, params ...interface{}) *Builder {
	var operator, second string
	joinType := JOIN_TYPE_LEFT
	length := len(params)
	switch length {
	case 0:
		if function, ok := firstColumn.(func(builder *Builder)); ok {
			clause := NewJoin(b, joinType, table)
			function(clause.Builder)
			b.Components[TYPE_JOIN] = struct{}{}
			b.Joins = append(b.Joins, clause)
			b.AddBinding(clause.GetBindings(), TYPE_JOIN)
			return b
		} else {
			panic(errors.New("arguements num mismatch"))
		}
	case 1:
		operator = "="
		second = params[0].(string)
	case 2:
		operator = params[0].(string)
		second = params[1].(string)
	}
	return b.join(table, firstColumn, operator, second, joinType, false)
}

/*
LeftJoinWhere Add a "join where" clause to the query.
*/
func (b *Builder) LeftJoinWhere(table, firstColumn, joinOperator, secondColumn string) *Builder {
	return b.joinWhere(table, firstColumn, joinOperator, secondColumn, JOIN_TYPE_LEFT)
}

/*
LeftJoinSub Add a subquery left join to the query.
*/
func (b *Builder) LeftJoinSub(query interface{}, as string, first interface{}, params ...interface{}) *Builder {
	queryStr, bindings := b.CreateSub(query)
	expr := fmt.Sprintf("(%s) as %s", queryStr, b.Grammar.WrapTable(as))
	var operator string
	joinType := JOIN_TYPE_LEFT
	var second interface{}
	switch len(params) {
	case 1:
		operator = "="
		second = params[0]
	case 2:
		operator = params[0].(string)
		second = params[1]

	}
	b.AddBinding(bindings, TYPE_JOIN)

	return b.join(expr, first, operator, second, joinType, false)
}

/*
RightJoin Add a right join to the query.
*/
func (b *Builder) RightJoin(table string, firstColumn interface{}, params ...interface{}) *Builder {

	var operator, second string
	joinType := JOIN_TYPE_RIGHT
	length := len(params)
	switch length {
	case 0:
		if function, ok := firstColumn.(func(builder *Builder)); ok {
			clause := NewJoin(b, joinType, table)
			function(clause.Builder)
			b.Components[TYPE_JOIN] = struct{}{}
			b.Joins = append(b.Joins, clause)
			b.AddBinding(clause.GetBindings(), TYPE_JOIN)
			return b
		} else {
			panic(errors.New("arguements num mismatch"))
		}
	case 1:
		operator = "="
		second = params[0].(string)
	case 2:
		operator = params[0].(string)
		second = params[1].(string)

	}
	return b.join(table, firstColumn, operator, second, joinType, false)
}

/*
RightJoinWhere Add a "right join where" clause to the query.
*/
func (b *Builder) RightJoinWhere(table, firstColumn, joinOperator, secondColumn string) *Builder {
	return b.joinWhere(table, firstColumn, joinOperator, secondColumn, JOIN_TYPE_RIGHT)
}

/*
RightJoinSub Add a subquery right join to the query.
*/
func (b *Builder) RightJoinSub(query interface{}, as string, first interface{}, params ...interface{}) *Builder {
	queryStr, bindings := b.CreateSub(query)
	expr := fmt.Sprintf("(%s) as %s", queryStr, b.Grammar.WrapTable(as))
	var operator string
	joinType := JOIN_TYPE_RIGHT
	var second interface{}
	switch len(params) {
	case 1:
		operator = "="
		second = params[0]
	case 2:
		operator = params[0].(string)
		second = params[1]

	}
	b.AddBinding(bindings, TYPE_JOIN)

	return b.join(expr, first, operator, second, joinType, false)
}

/*
CrossJoin Add a "cross join" clause to the query.
*/
func (b *Builder) CrossJoin(table string, params ...interface{}) *Builder {
	var operator, first, second string
	joinType := JOIN_TYPE_CROSS
	length := len(params)
	switch length {
	case 0:
		clause := NewJoin(b, joinType, table)
		b.Joins = append(b.Joins, clause)
		b.Components[TYPE_JOIN] = struct{}{}

		return b
	case 1:
		if function, ok := params[0].(func(builder *Builder)); ok {
			clause := NewJoin(b, joinType, table)
			function(clause.Builder)
			b.Joins = append(b.Joins, clause)
			b.AddBinding(clause.GetBindings(), TYPE_JOIN)
			b.Components[TYPE_JOIN] = struct{}{}

			return b
		} else {
			panic(errors.New("cross join arguements mismatch"))
		}
	case 2:
		first = params[0].(string)
		operator = "="
		second = params[1].(string)
	case 3:
		first = params[0].(string)
		operator = params[1].(string)
		second = params[2].(string)

	}
	return b.join(table, first, operator, second, joinType, false)
}

/*
CrossJoinSub Add a subquery cross join to the query.
*/
func (b *Builder) CrossJoinSub(query interface{}, as string) *Builder {
	queryStr, bindings := b.CreateSub(query)
	expr := fmt.Sprintf("(%s) as %s", queryStr, b.Grammar.WrapTable(as))
	b.Joins = append(b.Joins, NewJoin(b, JOIN_TYPE_CROSS, expr))
	b.AddBinding(bindings, TYPE_JOIN)

	return b
}

//	func NewJoin(builder *Builder, joinType string, table interface{}) *Builder {
//		cb := CloneBuilderWithTable(builder)
//		cb.JoinBuilder = true
//		cb.JoinType = joinType
//		cb.JoinTable = table
//
//		return cb
//	}
func (b *Builder) On(first interface{}, params ...interface{}) *Builder {
	var second string
	boolean := BOOLEAN_AND
	operator := "="
	switch len(params) {
	case 0:
		if function, ok := first.(func(builder *Builder)); ok {
			b.WhereNested(function, boolean)
			return b
		}
		panic(errors.New("arguements mismatch"))
	case 1:
		second = params[0].(string)
	case 2:
		operator = params[0].(string)
		second = params[1].(string)
	case 3:
		operator = params[0].(string)
		second = params[1].(string)
		boolean = params[2].(string)
	}

	b.WhereColumn(first.(string), operator, second, boolean)
	return b

}
func (b *Builder) OrOn(first interface{}, params ...interface{}) *Builder {
	var second string
	boolean := BOOLEAN_OR
	operator := "="
	switch len(params) {
	case 0:
		if function, ok := first.(func(builder *Builder)); ok {
			b.WhereNested(function, boolean)
			return b
		}
		panic(errors.New("arguements mismatch"))
	case 1:
		second = params[0].(string)
		return b.On(first, operator, second, boolean)
	case 2:
		operator = params[0].(string)
		second = params[1].(string)
		return b.On(first, operator, second, boolean)
	}
	panic(errors.New("arguements mismatch"))
}

// On is used to add a join clause to the query.
// On("users.id", "=", "contacts.user_id")
// On("users.id", "=", "contacts.user_id","or")
//
//	On(func(builder *Builder) {
//	    builder.Where("users.id", "=", "contacts.user_id")
//	})
func (j *JoinBuilder) On(params ...interface{}) *Builder {
	if f, ok := params[0].(func(builder *Builder)); ok {
		return j.Builder.WhereNested(f, params[1])
	}

	return j.WhereColumn(params[0], params[1].(string), params[2].(string), params[3].(string))
}

// OrOn is used to add an "or on" clause to the query.
// OrOn("users.id", "=", "contacts.user_id")
func (j *JoinBuilder) OrOn(params ...interface{}) *Builder {
	return j.On(params[0], params[1].(string), params[2].(string), "or")
}

func (j *JoinBuilder) NewQuery() *Builder {
	return j.Clone()
}

func (j *JoinBuilder) ForSubQuery() *Builder {
	return j.NewQuery()
}
