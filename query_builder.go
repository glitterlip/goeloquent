package goeloquent

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Component string
type HavingType string
type BooleanType string
type NotType bool
type JoinType string

const (
	COMPONENT_SELECT      Component = "select"
	COMPONENT_FROM        Component = "from"
	COMPONENT_INDEX_HINT  Component = "indexHint"
	COMPONENT_JOIN        Component = "join"
	COMPONENT_UPDATE      Component = "update"
	COMPONENT_WHERE       Component = "where"
	COMPONENT_GROUP_BY    Component = "groupBy"
	COMPONENT_HAVING      Component = "having"
	COMPONENT_ORDER       Component = "order"
	COMPONENT_UNION_ORDER Component = "unionOrder"
	COMPONENT_INSERT      Component = "insert"
	COMPONENT_COLUMN      Component = "column"
	COMPONENT_AGGREGRATE  Component = "aggregrate"
	COMPONENT_OFFSET      Component = "offset"
	COMPONENT_LIMIT       Component = "limit"
	COMPONENT_GROUP_LIMIT Component = "groupLimit"
	COMPONENT_LOCK        Component = "lock"

	WhereTypeBasic           WhereType = "basic"
	WhereTypeExpression      WhereType = "expression"
	WhereTypeBitwise         WhereType = "bitwise"
	WhereTypeJsonBoolean     WhereType = "jsonBoolean"
	WhereTypeColumn          WhereType = "column"
	WhereTypeRaw             WhereType = "raw"
	WhereTypeLike            WhereType = "like"
	WhereTypeNotIn           WhereType = "notIn"
	WhereTypeIn              WhereType = "in"
	WhereTypeNotInRaw        WhereType = "notInRaw"
	WhereTypeInRaw           WhereType = "inRaw"
	WhereTypeNotNull         WhereType = "notNull"
	WhereTypeNull            WhereType = "null"
	WhereTypeBetween         WhereType = "between"
	WhereTypeBetweenColumn   WhereType = "betweenColumn"
	WhereTypeNested          WhereType = "nested"
	WhereTypeSub             WhereType = "sub"
	WhereTypeNotExists       WhereType = "notExists"
	WhereTypeExists          WhereType = "exists"
	WhereTypeRowValues       WhereType = "rowValues"
	WhereTypeJsonContains    WhereType = "jsonContains"
	WhereTypeJsonOverlaps    WhereType = "jsonOverlaps"
	WhereTypeJsonContainsKey WhereType = "jsonContainsKey"
	WhereTypeJsonLength      WhereType = "jsonLength"
	WhereTypeFulltext        WhereType = "fulltext"
	WhereTypeDate            WhereType = "date"
	WhereTypeTime            WhereType = "time"
	WhereTypeDay             WhereType = "day"
	WhereTypeMonth           WhereType = "month"
	WhereTypeYear            WhereType = "year"

	HavingTypeRaw        HavingType = "raw"
	HavingTypeBasic      HavingType = "basic"
	HavingTypeBetween    HavingType = "between"
	HavingTypeNull       HavingType = "null"
	HavingTypeNotNull    HavingType = "notNull"
	HavingTypeBitwise    HavingType = "bitwise"
	HavingTypeExpression HavingType = "expression"
	HavingTypeNested     HavingType = "nested"

	JoinTypeInner JoinType = "inner"
	JoinTypeLeft  JoinType = "left"
	JoinTypeRight JoinType = "right"
	JoinTypeCross JoinType = "cross"
	JoinTypeFull  JoinType = "full"

	And BooleanType = "and"
	Or  BooleanType = "or"

	Not NotType = true

	DEFAULT_USE = 1 //use "=" as default operator
	DEFAULT_NOT = 0
)

var (
	SelectComponents = []Component{
		COMPONENT_AGGREGRATE,
		COMPONENT_COLUMN,
		COMPONENT_FROM,
		COMPONENT_INDEX_HINT,
		COMPONENT_JOIN,
		COMPONENT_WHERE,
		COMPONENT_GROUP_BY,
		COMPONENT_HAVING,
		COMPONENT_ORDER,
		COMPONENT_LIMIT,
		COMPONENT_OFFSET,
		COMPONENT_LOCK,
	}
	Bindings = map[Component]struct{}{
		COMPONENT_SELECT:      {},
		COMPONENT_FROM:        {},
		COMPONENT_JOIN:        {},
		COMPONENT_UPDATE:      {},
		COMPONENT_WHERE:       {},
		COMPONENT_GROUP_BY:    {},
		COMPONENT_HAVING:      {},
		COMPONENT_ORDER:       {},
		COMPONENT_UNION_ORDER: {},
		COMPONENT_INSERT:      {},
	}
	BindingKeysInOrder = []Component{COMPONENT_SELECT, COMPONENT_FROM, COMPONENT_JOIN, COMPONENT_UPDATE, COMPONENT_WHERE, COMPONENT_GROUP_BY, COMPONENT_HAVING, COMPONENT_ORDER, COMPONENT_UNION_ORDER, COMPONENT_INSERT}
	BitwiseOperators   = map[string]struct{}{
		"&":  {},
		"|":  {},
		"^":  {},
		"<<": {},
		">>": {},
		"&~": {},
	}
	NullableOperators = map[string]struct{}{
		"=":   {},
		"<=>": {},
		"<>":  {},
		"!=":  {},
	}
	Operators = map[string]struct{}{
		"=":              {},
		"<":              {},
		">":              {},
		"<=":             {},
		">=":             {},
		"<>":             {},
		"!=":             {},
		"<=>":            {},
		"like":           {},
		"like binary":    {},
		"not like":       {},
		"ilike":          {},
		"&":              {},
		"|":              {},
		"^":              {},
		"<<":             {},
		">>":             {},
		"&~":             {},
		"is":             {},
		"is not":         {},
		"rlike":          {},
		"not rlike":      {},
		"regexp":         {},
		"not regexp":     {},
		"~":              {},
		"~*":             {},
		"!~":             {},
		"!~*":            {},
		"similar to":     {},
		"not similar to": {},
		"not ilike":      {},
		"~~*":            {},
		"!~~*":           {},
	}
	ShortCutOperators = map[string]struct{}{
		"in":          {},
		"not in":      {},
		"between":     {},
		"not between": {},
	}
)

type QueryBuilder struct {
	Statement       *Statement
	Grammar         Grammar
	Components      map[Component]struct{}
	Bindings        map[Component][]interface{}
	Aggregates      Aggregate
	Columns         []interface{}
	IsDistinct      interface{}
	IsExist         bool
	DistinctColumns []string
	FromTable       interface{}
	IndexHint       IndexHint
	IsJoin          bool
	Joins           []*JoinBuilder
	Wheres          []Where
	Groups          []interface{}
	Havings         []Having
	Orders          []Order
	LimitNum        int64
	Grouplimit      GroupLimit
	OffsetNum       int64
	//Unions unsupported,use raw sql
	Locks                string
	BeforeQueryCallBacks []StatementFunc
	AfterQueryCallBacks  []StatementFunc
	TablePrefix          string
	Pretending           bool
	Parent               *QueryBuilder
	RawSql               string        //compiled sql
	RawBindings          []interface{} //compiled sql bindings
	Mapping              map[string]string
}
type Aggregate struct {
	AggregateName    string        //aggregate function
	AggregateColumns []interface{} //columns string or expression
}
type GroupLimit struct {
	Value  int64
	Column string
}
type Order struct {
	OrderType string
	Direction string
	Column    string //string or expression
	RawSql    string
}
type Having struct {
	Type           HavingType
	HavingColumn   string
	HavingOperator string
	HavingValue    []interface{}
	HavingBoolean  string
	RawSql         string
	Not            bool
	Query          *QueryBuilder
}
type WhereType string
type Where struct {
	Type          WhereType
	Boolean       string
	Column        interface{}
	Columns       []interface{} //rowValues
	Operator      string
	First         string //wherecolumn first column
	Second        string //wherecolumn second column
	RawSql        interface{}
	Value         interface{}
	Values        []interface{} // wherein values
	Not           bool          //not in,not between,not null
	Mode          string        //fulltext mode
	Expanded      bool          //fulltext expansion
	CaseSensitive bool          //like case-sensitive
	Query         *QueryBuilder //nested where
}
type IndexHint struct {
	Type  string
	Index string
}

func NewQueryBuilder(stmt ...*Statement) *QueryBuilder {
	qb := &QueryBuilder{
		Grammar:    NewMysqlGrammar(),
		Components: make(map[Component]struct{}),
		Bindings:   make(map[Component][]interface{}),
	}
	if len(stmt) > 0 {
		qb.Statement = stmt[0]
		stmt[0].QueryBuilder = qb
		if stmt[0].Connection != nil {
			prefix := stmt[0].Connection.GetTablePrefix()
			if prefix != "" {
				qb.TablePrefix = prefix
			}
		}
	}
	return qb
}

/*
Without unset components and bindings
*/
func (q *QueryBuilder) Without(components []Component, bindings []Component) *QueryBuilder {
	if q.RawSql != "" {
		//unset compiled sql and bingdings
		q.RawSql = ""
		q.RawBindings = []interface{}{}
	}
	for _, component := range components {
		delete(q.Components, component)
		switch component {
		case COMPONENT_SELECT:
			q.Columns = nil
		case COMPONENT_FROM:
			q.FromTable = nil
		case COMPONENT_JOIN:
			q.IsJoin = false
			q.Joins = nil
		case COMPONENT_WHERE:
			q.Wheres = nil
		case COMPONENT_GROUP_BY:
			q.Groups = nil
		case COMPONENT_HAVING:
			q.Havings = nil
		case COMPONENT_ORDER:
			q.Orders = nil
		case COMPONENT_LIMIT:
			q.LimitNum = 0
		case COMPONENT_OFFSET:
			q.OffsetNum = 0
		case COMPONENT_GROUP_LIMIT:
			q.Grouplimit = GroupLimit{}
		case COMPONENT_LOCK:
			q.Locks = ""
		case COMPONENT_AGGREGRATE:
			q.Aggregates = Aggregate{}
		case COMPONENT_INDEX_HINT:
			q.IndexHint = IndexHint{}
		case COMPONENT_INSERT:
			q.Statement.Dest = nil
		case COMPONENT_COLUMN:
			q.Columns = nil
		}
	}
	for _, binding := range bindings {
		delete(q.Bindings, binding)
	}
	return q
}

/*
Select set the columns to be selected

 1. Select("name")

    select `name` from table

 2. Select("name","email")

    select `name`,`email` from table

 3. Select([]interface{}{"name","age"})

    select `name`,`age` from table

 4. Select([]string{"name","age"})

    select `name`,`age` from table

 5. Select("users.*",Raw("COUNT(posts.id) as post_count"))

    select `users`.*,COUNT(posts.id) as post_count from table

 6. Select("users.*", func(qb *goeloquent.QueryBuilder) {
    qb.Select("COUNT(posts.id)").From("posts").WhereColumn("posts.user_id", "users.id")
    },"post_count")

    select `users`.*, (select COUNT(posts.id) from posts where posts.user_id = users.id) as post_count from table

 5. Select("users.*", func(eb *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
    return eb.Select("COUNT(posts.id)").From("posts").WhereColumn("posts.user_id", "users.id")
    },"post_count")

    select `users`.*, (select COUNT(posts.id) from posts where posts.user_id = users.id) as post_count from table
*/
func (q *QueryBuilder) Select(columns ...interface{}) *QueryBuilder {
	q.Components[COMPONENT_COLUMN] = struct{}{}
	if q.Columns == nil {
		q.Columns = []interface{}{}
	}

	for i := 0; i < len(columns); i++ {
		if IsQueryable(columns[i]) {
			if len(columns) != 2 {
				q.Statement.AddError(errors.New("first parameter must be subquery,second parameter must be alias string when use subquery"))
				return q
			}
			if _, ok := columns[1].(string); !ok {
				q.Statement.AddError(errors.New("first parameter must be subquery,second parameter must be alias string when use subquery"))
				return q
			}
			return q.SelectSub(columns[0], columns[1].(string))
		}
		switch columns[i].(type) {
		case string, Expression:
			q.Columns = append(q.Columns, columns[i])
		case []interface{}:
			cols := columns[i].([]interface{})
			q.Columns = append(q.Columns, cols...)
		case []string:
			cols := columns[i].([]string)
			for _, col := range cols {
				q.Columns = append(q.Columns, col)
			}
		default:
			q.Statement.AddError(errors.New("unsupported type for select"))
		}
	}
	return q
}

/*
SelectSub Add a subselect expression to the query.

 1. SelectSub("select max(id) from users","max_id")

    select *, (select max(id) from users) as max_id from table

 2. SelectSub(func(qb *goeloquent.QueryBuilder) {
    qb.Select("max(id)").From("users").Where("email", "like", "gmail.com")
    },"max_id")

    select *, (select max(id) from users where email like 'gmail.com') as max_id from table

 3. SelectSub(func(eb *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
    return eb.Select("max(id)").From("users").Where("email", "like", "gmail.com")
    },"max_id")

    select *, (select max(id) from users where email like 'gmail.com') as max_id from table
*/
func (q *QueryBuilder) SelectSub(query interface{}, as string) *QueryBuilder {
	qStr, bindings := q.CreateSub(query)
	queryStr := fmt.Sprintf("(%s) as %s", qStr, q.Grammar.Wrap(as))

	return q.SelectRaw(queryStr, bindings)
}

/*
SelectRaw Add a new "raw" select expression to the query.
*/
func (q *QueryBuilder) SelectRaw(expression string, bindings ...[]interface{}) *QueryBuilder {
	q.AddSelect(Expression(expression))
	if len(bindings) > 0 {
		q.AddBinding(bindings[0], COMPONENT_SELECT)
	}
	return q
}

/*
FromSub Makes "from" fetch from a subquery.

 1. FromSub(goeloquent.Raw("select category, avg(price) as avg_price from products group by category"), "category_averages")

    from(select category, avg(price) as avg_price from products group by category) as category_averages

 2. FromSub(func(qb *goeloquent.QueryBuilder) {
    qb.Select("category").SelectRaw("avg(price) as avg_price").From("products").GroupBy("category")
    }, "category_averages")

    from(select category, avg(price) as avg_price from products group by category) as category_averages

 3. FromSub(func(eb *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
    return eb.Select("category").SelectRaw("avg(price) as avg_price").From("products").GroupBy("category")
    }, "category_averages")

    from(select category, avg(price) as avg_price from products group by category) as category_averages
*/
func (q *QueryBuilder) FromSub(table interface{}, as string) *QueryBuilder {
	qStr, bindings := q.CreateSub(table)
	queryStr := fmt.Sprintf("(%s) as %s", qStr, q.Grammar.WrapTable(as))

	return q.FromRaw(queryStr, bindings)
}

/*
FromRaw Add a raw from clause to the query.

 1. FromRaw(goeloquent.Raw(`(select max(last_seen_at) as last_seen_at from "user_sessions") as "sessions""`))

    select * from (select max(last_seen_at) as last_seen_at from "user_sessions") as "sessions"

 2. FromRaw("users as u")

    select * from users as u
*/
func (q *QueryBuilder) FromRaw(raw string, bindings ...[]interface{}) *QueryBuilder {
	expression := Raw(raw)
	q.FromTable = expression
	q.Components[COMPONENT_FROM] = struct{}{}
	if len(bindings) > 0 {
		q.AddBinding(bindings[0], COMPONENT_FROM)
	}
	return q
}

/*
CreateSub Creates a subquery and parse it.

 1. CreateSub(goeloquent.Raw("select max(id) from users"))

    (select max(id) from users)

 2. CreateSub(func(qb *goeloquent.QueryBuilder) {
    qb.Select("max(id)").From("users").Where("email", "like", "gmail.com")
    })

    (select max(id) from users where email like 'gmail.com')

 3. CreateSub(func(eb *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
    return eb.Select("max(id)").From("users").Where("email", "like", "gmail.com")
    })

    (select max(id) from users where email like 'gmail.com')
*/
func (q *QueryBuilder) CreateSub(query interface{}) (string, []interface{}) {
	var builder *QueryBuilder
	switch query.(type) {
	case *QueryBuilder:
		builder = query.(*QueryBuilder)
		return q.ParseSub(builder)
	case *EloquentBuilder:
		builder = query.(*EloquentBuilder).QueryBuilder
		return q.ParseSub(builder)
	case func(*QueryBuilder):
		builder = q.ForSubQuery()
		query.(func(*QueryBuilder))(builder)
		return q.ParseSub(builder)
	case func(*QueryBuilder) *QueryBuilder:
		builder = q.ForSubQuery()
		builder = query.(func(*QueryBuilder) *QueryBuilder)(builder)
		return q.ParseSub(builder)
	case func(*EloquentBuilder):
		eb := NewEloquentBuilder(q.ForSubQuery())
		query.(func(*EloquentBuilder))(eb)
		return q.ParseSub(eb.QueryBuilder)
	case func(*EloquentBuilder) *EloquentBuilder:
		eb := NewEloquentBuilder(q.ForSubQuery())
		eb = query.(func(*EloquentBuilder) *EloquentBuilder)(eb)
		return q.ParseSub(eb.QueryBuilder)
	case string, Expression:
		return q.ParseSub(query)
	}
	q.Statement.AddError(ErrorSubQueryInvalid)
	return "", nil
}

/*
ParseSub Parse the subquery into SQL and bindings.
*/
func (q *QueryBuilder) ParseSub(query interface{}) (string, []interface{}) {
	switch query.(type) {
	case string:
		return query.(string), []interface{}{}
	case Expression:
		return string(query.(Expression)), []interface{}{}
	case *QueryBuilder:
		return query.(*QueryBuilder).ToSql(), query.(*QueryBuilder).GetBindings()
	case *EloquentBuilder:
		return query.(*EloquentBuilder).QueryBuilder.ToSql(), query.(*EloquentBuilder).QueryBuilder.GetBindings()
	}

	q.Statement.AddError(ErrorSubQueryInvalid)
	return "", nil
}


}

/*
GetRawBindings Get the raw map of array of bindings.
*/
func (q *QueryBuilder) GetRawBindings() map[Component][]interface{} {

	return q.Bindings
}

/*
AddBinding Add a binding to the query.
*/
func (q *QueryBuilder) AddBinding(bindings []interface{}, component Component) *QueryBuilder {

	for _, binding := range bindings {
		if _, ok := binding.(Expression); !ok {
			q.Bindings[component] = append(q.Bindings[component], binding)
		}
	}
	return q
}

/*
SetBindings Set the bindings on the query builder.
*/
func (q *QueryBuilder) SetBindings(bindings []interface{}, component Component) *QueryBuilder {

	q.Bindings[component] = bindings
	return q
}
func (q *QueryBuilder) Table(name ...string) *QueryBuilder {
	switch len(name) {
	case 0:
		q.Statement.Error = errors.New("table name is required")
	case 1:

	}

	return q
}

/*
GroupLimit Add a "group limit" clause to the query.
*/
func (q *QueryBuilder) GroupLimit(value int, column string) *QueryBuilder {
	if value >= 0 {
		q.Grouplimit.Value = value
		q.Grouplimit.Column = column
		q.Components[COMPONENT_GROUP_LIMIT] = struct{}{}
	}
	return q
}

/*
Lock Lock the selected rows in the table.

 1. Lock()

    lock for update

 2. Lock(false)

    lock in share mode

 3. Lock("lock in share mode")

    lock in share mode
*/
func (q *QueryBuilder) Lock(locks ...interface{}) *QueryBuilder {
	q.Components[COMPONENT_LOCK] = struct{}{}

	if len(locks) > 0 {
		switch locks[0].(type) {
		case string:
			q.Locks = locks[0].(string)
		case bool:
			if locks[0].(bool) {
				q.Locks = "for update"
			} else {
				q.Locks = "lock in share mode"
			}
		}
	} else {
		q.Locks = "for update"
	}
	return q
}
