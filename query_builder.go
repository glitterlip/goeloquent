package goeloquent

import (
	"errors"
	"fmt"
)

type Component string
type HavingType string
type QuerybuilderFunc func(*QueryBuilder)
type QuerybuilderChainFunc func(*QueryBuilder) *QueryBuilder

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
	COMPONENT_LOCK        Component = "lock"

	WhereTypeBasic           WhereType  = "basic"
	WhereTypeExpression      WhereType  = "expression"
	WhereTypeBitwise         WhereType  = "bitwise"
	WhereTypeJsonBoolean     WhereType  = "jsonBoolean"
	WhereTypeColumn          WhereType  = "column"
	WhereTypeRaw             WhereType  = "raw"
	WhereTypeLike            WhereType  = "like"
	WhereTypeNotIn           WhereType  = "notIn"
	WhereTypeIn              WhereType  = "in"
	WhereTypeNotInRaw        WhereType  = "notInRaw"
	WhereTypeInRaw           WhereType  = "inRaw"
	WhereTypeNotNull         WhereType  = "notNull"
	WhereTypeNull            WhereType  = "null"
	WhereTypeBetween         WhereType  = "between"
	WhereTypeBetweenColumn   WhereType  = "betweenColumn"
	WhereTypeNested          WhereType  = "nested"
	WhereTypeSub             WhereType  = "sub"
	WhereTypeNotExists       WhereType  = "notExists"
	WhereTypeExists          WhereType  = "exists"
	WhereTypeRowValues       WhereType  = "rowValues"
	WhereTypeJsonContains    WhereType  = "jsonContains"
	WhereTypeJsonOverlaps    WhereType  = "jsonOverlaps"
	WhereTypeJsonContainsKey WhereType  = "jsonContainsKey"
	WhereTypeJsonLength      WhereType  = "jsonLength"
	WhereTypeFulltext        WhereType  = "fulltext"
	WhereTypeDate            WhereType  = "date"
	WhereTypeTime            WhereType  = "time"
	WhereTypeDay             WhereType  = "day"
	WhereTypeMonth           WhereType  = "month"
	WhereTypeYear            WhereType  = "year"
	HavingTypeRaw            HavingType = "raw"
	HavingTypeBasic          HavingType = "basic"
	HavingTypeBetween        HavingType = "between"
	HavingTypeNull           HavingType = "null"
	HavingTypeNotNull        HavingType = "notNull"
	HavingTypeBitwise        HavingType = "bitwise"
	HavingTypeExpression     HavingType = "expression"
	HavingTypeNested         HavingType = "nested"
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
	BindingKeysInOrder = []Component{COMPONENT_SELECT, COMPONENT_FROM, COMPONENT_JOIN, COMPONENT_UPDATE, COMPONENT_WHERE, COMPONENT_GROUP_BY, COMPONENT_HAVING, COMPONENT_ORDER, COMPONENT_UNION, COMPONENT_UNION_ORDER, COMPONENT_INSERT}
	BitwiseOperators   = []string{"&", "|", "^", "<<", ">>", "&~"}
)

type QueryBuilder struct {
	Statement  *Statement
	Grammar    Grammar
	Components map[Component]struct{}
	Bindings   map[Component][]interface{}
	Aggregate  Aggregate
	Columns    []interface{}
	Distinct   interface{}
	From       interface{}
	IndexHint  IndexHint
	IsJoin     bool
	Joins      []*JoinBuilder
	Wheres     []Where
	Groups     []interface{}
	Havings    []Having
	Orders     []Order
	Limit      int
	GroupLimit int
	Offset     int
	//Unions unsupported,use raw sql
	//Lock //todo
	BeforeQueryCallBacks []StatementFunc
	AfterQueryCallBacks  []StatementFunc
	TablePrefix          string
}
type Aggregate struct {
	AggregateName    string        //aggregate function
	AggregateColumns []interface{} //columns string or expression
}
type Order struct {
	OrderType string
	Direction string
	Column    interface{} //string or expression
	RawSql    interface{}
}
type Having struct {
	Type           HavingType
	HavingColumn   string
	HavingOperator string
	HavingValue    interface{}
	HavingBoolean  string
	RawSql         interface{}
	Not            bool
	Query          *QueryBuilder
}
type WhereType string
type Where struct {
	Type          WhereType
	Boolean       string
	Column        string
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
	CaseSensitive bool          //like case sensitive
	Query         *QueryBuilder
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
		Grammar: NewMysqlGrammar(),
	}
	if len(stmt) > 0 {
		qb.Statement = stmt[0]
		stmt[0].QueryBuilder = qb
	}
	return qb
}

// Select set the columns to be selected
func (q *QueryBuilder) Select(columns ...interface{}) *QueryBuilder {
	q.Components[COMPONENT_COLUMN] = struct{}{}
	if q.Columns == nil {
		q.Columns = []interface{}{}
	}

	for i := 0; i < len(columns); i++ {
		if IsQueryable(columns[i]) {
			return q.SelectSub(columns[0], columns[1].(string))
		}
		switch columnType := columns[i].(type) {
		case QuerybuilderFunc:
			q.SelectSub(columns[0], columns[1].(string))
			return q
		case QuerybuilderChainFunc:
		case string:
			q.Columns = append(q.Columns, columns[i])
		case map[string]interface{}:
			for as, c := range columnType {
				switch c.(type) {
				case func(builder *QueryBuilder):
					q.SelectSub(q, as)
				case *QueryBuilder:
					q.SelectSub(q, as)
				case Expression:
					q.AddSelect(q)
				case string:
					q.Columns = append(q.Columns, q)
				default:
					panic(errors.New("unsupported type for select"))
				}
			}
		case Expression:
			q.AddSelect(columnType)
		case []string:
			cols := columns[i].([]string)
			for _, col := range cols {
				q.Columns = append(q.Columns, col)
			}
		default:
			panic(errors.New("unsupported type for select"))
		}
	}
	return q
}

// SelectSub Add a subselect expression to the query.
func (q *QueryBuilder) SelectSub(query interface{}, as string) *QueryBuilder {
	qStr, bindings := q.CreateSub(query)
	queryStr := fmt.Sprintf("(%s) as %s", qStr, q.Grammar.Wrap(as))

	return q.SelectRaw(queryStr, bindings)

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
