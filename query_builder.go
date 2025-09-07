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

// todo: lock
type Lock struct {
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
 2. Select("name","email")
 3. Select([]interface{}{"name","age"})
 4. Select(Raw("raw sql"))
*/
func (q *QueryBuilder) Select(columns ...interface{}) *QueryBuilder {
	q.Components[COMPONENT_COLUMN] = struct{}{}
	if q.Columns == nil {
		q.Columns = []interface{}{}
	}

	for i := 0; i < len(columns); i++ {
		if IsQueryable(columns[i]) {
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
