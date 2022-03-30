package goeloquent

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"strings"
	"time"
)

var Operators = []string{
	"=", "<", ">", "<=", ">=", "<>", "!=", "<=>",
	"like", "like binary", "not like", "ilike",
	"&", "|", "^", "<<", ">>",
	"rlike", "regexp", "not regexp",
	"~", "~*", "!~", "!~*", "similar to",
	"not similar to", "not ilike", "~~*", "!~~*"}
var SelectComponents = []string{
	TYPE_AGGREGRATE,
	TYPE_COLUMN,
	TYPE_FROM,
	TYPE_JOIN,
	TYPE_WHERE,
	TYPE_GROUP_BY,
	TYPE_HAVING,
	TYPE_ORDER,
	TYPE_LIMIT,
	TYPE_OFFSET,
	TYPE_UNION,
	TYPE_LOCK,
}
var Bindings = map[string]struct{}{
	TYPE_SELECT:      {},
	TYPE_FROM:        {},
	TYPE_JOIN:        {},
	TYPE_UPDATE:      {},
	TYPE_WHERE:       {},
	TYPE_GROUP_BY:    {},
	TYPE_HAVING:      {},
	TYPE_ORDER:       {},
	TYPE_UNION:       {},
	TYPE_UNION_ORDER: {},
	TYPE_INSERT:      {},
}

type Builder struct {
	Connection *Connection
	Tx         *Transaction
	Grammar    IGrammar
	//Processor   processors.IProcessor
	PreSql    strings.Builder
	Bindings  map[string][]interface{} //available options (select,from,join,where,groupBy,having,order,union,unionOrder)
	FromTable interface{}
	//TablePrefix     string
	TableAlias      string
	Wheres          []Where
	Aggregates      []Aggregate
	Columns         []interface{} // The columns that should be returned.
	IsDistinct      bool          // Indicates if the query returns distinct results.
	DistinctColumns []string      // distinct columns.
	Joins           []*Builder
	Groups          []interface{}
	Havings         []Having
	Orders          []Order
	LimitNum        int
	OffsetNum       int
	//Unions           []Where
	//UnionLimit       int
	//UnionOffset      int
	//UnionOrders      int
	Components       map[string]struct{} //SelectComponents
	LockMode         interface{}
	LoggingQueries   bool
	Pretending       bool
	PreparedSql      string
	Model            *Model
	Dest             interface{}
	DestReflectValue reflect.Value
	EagerLoad        map[string]func(*RelationBuilder) *RelationBuilder //TODO map[string]callback to dynamicly add constraints
	Pivots           []string
	PivotWheres      []Where
	OnlyColumns      map[string]interface{}
	ExceptColumns    map[string]interface{}
	JoinBuilder      bool
	JoinType         string
	JoinTable        interface{}
	UseWrite         bool //TODO:
	QueryCallBacks   []func(builder *Builder)
}

type Log struct {
	SQL      string
	Bindings []interface{}
	Result   sql.Result
	Time     time.Duration
}

func (b *Builder) WithOutGlobalScope(...interface{}) *Builder {

	return b
}
func (b *Builder) WithOutGlobalScopes(...interface{}) *Builder {

	return b
}

const (
	CONDITION_TYPE_BASIC          = "basic"
	CONDITION_TYPE_COLUMN         = "column"
	CONDITION_TYPE_RAW            = "raw"
	CONDITION_TYPE_IN             = "in"
	CONDITION_TYPE_NOT_IN         = "not in"
	CONDITION_TYPE_NULL           = "null"
	CONDITION_TYPE_BETWEEN        = "between"
	CONDITION_TYPE_BETWEEN_COLUMN = "between column"
	CONDITION_TYPE_NOT_BETWEEN    = "not between"
	CONDITION_TYPE_DATE           = "date"
	CONDITION_TYPE_TIME           = "time"
	CONDITION_TYPE_DATETIME       = "datetime"
	CONDITION_TYPE_DAY            = "day"
	CONDITION_TYPE_MONTH          = "month"
	CONDITION_TYPE_YEAR           = "year"
	CONDITION_TYPE_CLOSURE        = "closure" //todo
	CONDITION_TYPE_NESTED         = "nested"
	CONDITION_TYPE_SUB            = "subquery" //todo
	CONDITION_TYPE_EXIST          = "exist"
	CONDITION_TYPE_NOT_EXIST      = "not exist"
	CONDITION_TYPE_ROW_VALUES     = "rowValues" //todo
	BOOLEAN_AND                   = "and"
	BOOLEAN_OR                    = "or"
	CONDITION_JOIN_NOT            = "not" //todo
	JOIN_TYPE_LEFT                = "left"
	JOIN_TYPE_RIGHT               = "right"
	JOIN_TYPE_INNER               = "inner"
	JOIN_TYPE_CROSS               = "cross"
	ORDER_ASC                     = "asc"
	ORDER_DESC                    = "desc"
	TYPE_SELECT                   = "select"
	TYPE_FROM                     = "from"
	TYPE_JOIN                     = "join"
	TYPE_WHERE                    = "where"
	TYPE_GROUP_BY                 = "groupBy"
	TYPE_HAVING                   = "having"
	TYPE_ORDER                    = "order"
	TYPE_UNION                    = "union"
	TYPE_UNION_ORDER              = "unionOrder"
	TYPE_COLUMN                   = "column"
	TYPE_AGGREGRATE               = "aggregrate"
	TYPE_OFFSET                   = "offset"
	TYPE_LIMIT                    = "limit"
	TYPE_LOCK                     = "lock"
	TYPE_INSERT                   = "insert"
	TYPE_UPDATE                   = "update"
)

type Aggregate struct {
	AggregateName   string
	AggregateColumn string
}

type Order struct {
	OrderType string
	Direction string
	Column    string
	RawSql    interface{}
}
type Having struct {
	HavingType     string
	HavingColumn   string
	HavingOperator string
	HavingValue    interface{}
	HavingBoolean  string
	RawSql         interface{}
	Not            bool
}
type Where struct {
	Type         string
	Column       string
	Operator     string
	FirstColumn  string
	SecondColumn string
	RawSql       interface{}
	Value        interface{}
	Values       []interface{}
	Boolean      string
	Not          bool //not in,not between,not null
	Query        *Builder
}

func NewBuilder(c *Connection) *Builder {
	b := Builder{
		Connection: c,
		Components: make(map[string]struct{}),
		EagerLoad:  make(map[string]func(builder *RelationBuilder) *RelationBuilder),
		//Processor:  processors.MysqlProcessor{},
		Bindings:       make(map[string][]interface{}),
		LoggingQueries: c.Config.EnableLog,
	}
	return &b
}
func NewTxBuilder(tx *Transaction) *Builder {
	b := Builder{
		Components:     make(map[string]struct{}),
		EagerLoad:      make(map[string]func(builder *RelationBuilder) *RelationBuilder),
		Tx:             tx,
		Bindings:       make(map[string][]interface{}),
		LoggingQueries: tx.Config.EnableLog,
	}
	return &b
}
func CloneBuilder(b *Builder) *Builder {
	cb := Builder{
		Connection:     b.Connection,
		Components:     make(map[string]struct{}),
		EagerLoad:      make(map[string]func(builder *RelationBuilder) *RelationBuilder),
		Grammar:        &MysqlGrammar{},
		Tx:             b.Tx,
		Bindings:       make(map[string][]interface{}),
		LoggingQueries: b.LoggingQueries,
	}
	cb.Grammar.SetTablePrefix(b.Grammar.GetTablePrefix())
	cb.Grammar.SetBuilder(&cb)
	return &cb
}

/*
Clone Clone the query.
*/
func Clone(original *Builder) *Builder {
	newBuilder := Builder{
		Connection:       original.Connection,
		Tx:               original.Tx,
		PreSql:           strings.Builder{},
		Bindings:         make(map[string][]interface{}, len(original.Bindings)),
		FromTable:        original.FromTable,
		TableAlias:       original.TableAlias,
		Wheres:           make([]Where, len(original.Wheres)),
		Aggregates:       make([]Aggregate, len(original.Aggregates)),
		Columns:          make([]interface{}, len(original.Columns)),
		IsDistinct:       original.IsDistinct,
		DistinctColumns:  make([]string, len(original.DistinctColumns)),
		Joins:            make([]*Builder, len(original.Joins)),
		Groups:           make([]interface{}, len(original.Groups)),
		Havings:          make([]Having, len(original.Havings)),
		Orders:           make([]Order, len(original.Orders)),
		LimitNum:         original.LimitNum,
		OffsetNum:        original.OffsetNum,
		Components:       make(map[string]struct{}, len(original.Components)),
		LockMode:         original.LockMode,
		LoggingQueries:   original.LoggingQueries,
		Pretending:       original.Pretending,
		PreparedSql:      "",
		Model:            original.Model,
		Dest:             nil,
		DestReflectValue: original.DestReflectValue,
		EagerLoad:        original.EagerLoad,
		Pivots:           make([]string, len(original.Pivots)),
		PivotWheres:      make([]Where, len(original.Wheres)),
		OnlyColumns:      make(map[string]interface{}, len(original.OnlyColumns)),
		ExceptColumns:    make(map[string]interface{}, len(original.ExceptColumns)),
		JoinBuilder:      original.JoinBuilder,
		JoinType:         original.JoinType,
		JoinTable:        original.JoinTable,
	}
	for key, _ := range original.Bindings {
		copy(newBuilder.Bindings[key], newBuilder.Bindings[key])
	}
	copy(newBuilder.Wheres, original.Wheres)
	copy(newBuilder.Aggregates, original.Aggregates)
	copy(newBuilder.Columns, original.Columns)
	copy(newBuilder.DistinctColumns, original.DistinctColumns)
	copy(newBuilder.Groups, original.Groups)
	copy(newBuilder.Havings, original.Havings)
	copy(newBuilder.Orders, original.Orders)
	copy(newBuilder.Pivots, original.Pivots)
	copy(newBuilder.PivotWheres, original.PivotWheres)
	for key, _ := range original.Components {
		newBuilder.Components[key] = original.Components[key]
	}
	for key, _ := range original.OnlyColumns {
		newBuilder.OnlyColumns[key] = original.OnlyColumns[key]
	}
	for key, _ := range original.ExceptColumns {
		newBuilder.ExceptColumns[key] = original.ExceptColumns[key]
	}
	for _, join := range original.Joins {
		newBuilder.Joins = append(newBuilder.Joins, join.Clone()) //TODO: add tests
	}
	newBuilder.Grammar = &MysqlGrammar{
		Prefix:  original.Grammar.GetTablePrefix(),
		Builder: &newBuilder,
	}
	return &newBuilder
}
func (b *Builder) Clone() *Builder {
	return Clone(b)
}

/*CloneWithout
CloneWithoutClone the query without the given properties.
*/
func CloneWithout(original *Builder, without ...string) *Builder {
	b := Clone(original)
	b.Reset(without...)
	return b
}
func (b *Builder) CloneWithout(without ...string) *Builder {
	return CloneWithout(b, without...)
}

/*
CloneWithoutBindings Clone the query without the given bindings.
*/
func CloneWithoutBindings(original *Builder, bindings ...string) *Builder {
	b := Clone(original)
	for _, binding := range bindings {
		b.Bindings[binding] = nil
	}
	return b
}
func (b *Builder) CloneWithoutBindings(bindings ...string) *Builder {
	return CloneWithoutBindings(b, bindings...)
}

// Select set columns to be selected
// 1. Select("column1","column2","column3")
// 2. Select()
func (b *Builder) Select(columns ...interface{}) *Builder {
	b.Columns = []interface{}{}
	b.Components[TYPE_COLUMN] = struct{}{}
	if len(columns) == 0 {
		b.Columns = append(b.Columns, "*")
	}

	for i := 0; i < len(columns); i++ {
		if c, ok := columns[i].(string); ok {
			b.Columns = append(b.Columns, c)
		} else if mf, ok := columns[i].(map[string]func(builder *Builder)); ok {
			for as, q := range mf {
				b.SelectSub(q, as)
			}
		} else if mb, ok := columns[i].(map[string]*Builder); ok {
			for as, q := range mb {
				b.SelectSub(q, as)
			}
		} else if raw, ok := columns[i].(Expression); ok {
			b.AddSelect(raw)
		}
	}
	return b
}

//SelectSub Add a subselect expression to the query.
func (b *Builder) SelectSub(query interface{}, as string) *Builder {
	qStr, bindings := b.CreateSub(query)
	queryStr := fmt.Sprintf("( %s ) as %s", qStr, b.Grammar.Wrap(as))

	return b.SelectRaw(queryStr, bindings)

}

// AddSelect Add a new select column to the query
// 1. slice of string
// 2. map[string]{"alias"}
func (b *Builder) AddSelect(columns ...interface{}) *Builder {
	b.Components[TYPE_COLUMN] = struct{}{}

	for i := 0; i < len(columns); i++ {
		if str, ok := columns[i].(string); ok {
			b.Columns = append(b.Columns, str)
		} else if m, ok := columns[i].(map[string]interface{}); ok {
			for as, q := range m {
				b.SelectSub(q, as)
			}
		} else if e, ok := columns[i].(Expression); ok {
			b.Columns = append(b.Columns, e)
		}
	}
	return b
}

// SelectRaw Add a new "raw" select expression to the query.
func (b *Builder) SelectRaw(expression string, bindings ...[]interface{}) *Builder {
	b.AddSelect(Expression(expression))
	if len(bindings) > 0 {
		b.AddBinding(bindings[0], TYPE_SELECT)
	}
	return b
}

//CreateSub Creates a subquery and parse it.
func (b *Builder) CreateSub(query interface{}) (string, []interface{}) {
	var builder *Builder
	if bT, ok := query.(*Builder); ok {
		builder = bT
	} else if function, ok := query.(func(builder *Builder)); ok {
		builder = CloneBuilder(b)
		function(builder)
	} else if str, ok := query.(string); ok {
		return b.ParseSub(str)
	} else {
		panic("can not create sub")
	}
	return b.ParseSub(builder)
}

/*
ParseSub Parse the subquery into SQL and bindings.
*/
func (b *Builder) ParseSub(query interface{}) (string, []interface{}) {
	if s, ok := query.(string); ok {
		return s, []interface{}{}
	} else if builder, ok := query.(*Builder); ok {
		return builder.ToSql(), builder.GetBindings()
	}
	panic("A subquery must be a query builder instance, a Closure, or a string.")
}

/*
ToSql Get the SQL representation of the query.
*/
func (b *Builder) ToSql() string {

	if len(b.PreparedSql) > 0 {
		return b.PreparedSql
	}
	return b.Grammar.CompileSelect()
}

/*
Distinct Force the query to only return distinct results.
*/
func (b *Builder) Distinct(distinct ...string) *Builder {
	b.IsDistinct = true
	if len(distinct) > 0 {
		b.DistinctColumns = append(b.DistinctColumns, distinct...)
	}
	return b
}

/*
IsQueryable Determine if the value is a query builder instance or a Closure.
*/
func IsQueryable(value interface{}) bool {
	switch value.(type) {
	case Builder, *Builder, RelationBuilder, *RelationBuilder:
		return true
	case BelongsToRelation, BelongsToManyRelation, HasManyRelation, HasManyThrough, HasOneRelation, HasOneThrough, MorphManyRelation, MorphByManyRelation, MorphOneRelation, MorphToManyRelation, MorphToRelation:
		return true
	case func(builder *Builder):
		return true
	default:
		return false
	}
}

/*
Table Begin a fluent query against a database table.
*/
func (b *Builder) Table(params ...string) *Builder {
	if len(params) == 1 {
		return b.From(params[0])
	} else {
		return b.From(params[0], params[1])
	}
}

/*
From Set the table which the query is targeting.
*/
func (b *Builder) From(table interface{}, params ...string) *Builder {
	if IsQueryable(table) {
		return b.FromSub(table, params[0])
	}
	b.Components[TYPE_FROM] = struct{}{}
	if len(params) == 1 {
		b.TableAlias = params[0]
		b.FromTable = fmt.Sprintf("%s as %s", table, params[0])
	} else {
		b.FromTable = table.(string)
	}
	return b
}

/*
FromSub Makes "from" fetch from a subquery.
*/
func (b *Builder) FromSub(table interface{}, as string) *Builder {
	qStr, bindings := b.CreateSub(table)
	queryStr := fmt.Sprintf("(%s) as %s", qStr, b.Grammar.WrapTable(as))

	return b.FromRaw(queryStr, bindings)
}

/*
FromRaw Add a raw from clause to the query.
*/
func (b *Builder) FromRaw(raw interface{}, bindings ...[]interface{}) *Builder {
	var expression Expression
	if str, ok := raw.(string); ok {
		expression = Expression(str)
	} else {
		expression = raw.(Expression)
	}
	b.FromTable = expression
	b.Components[TYPE_FROM] = struct{}{}
	if len(bindings) > 0 {
		b.AddBinding(bindings[0], TYPE_FROM)
	}
	return b
}

/*
Join Add a join clause to the query.
*/
func (b *Builder) Join(table string, first interface{}, params ...interface{}) *Builder {
	var operator, second, joinType string
	var isWhere = false
	length := len(params)
	switch length {
	case 0:
		if function, ok := first.(func(builder *Builder)); ok {
			clause := NewJoin(b, JOIN_TYPE_INNER, table)
			function(clause)
			b.Joins = append(b.Joins, clause)
			b.Components[TYPE_JOIN] = struct{}{}
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
			function(clause)
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
			function(clause)
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
RightJoinWhere Add a "right join where" clause to the query.
*/
func (b *Builder) RightJoinWhere(table, firstColumn, joinOperator, secondColumn string) *Builder {
	return b.joinWhere(table, firstColumn, joinOperator, secondColumn, JOIN_TYPE_RIGHT)
}
func NewJoin(builder *Builder, joinType string, table interface{}) *Builder {
	cb := CloneBuilder(builder)
	cb.JoinBuilder = true
	cb.JoinType = joinType
	cb.JoinTable = table

	return cb
}
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
			function(clause)
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
	b.AddBinding(bindings, TYPE_JOIN)
	clause := NewJoin(b, JOIN_TYPE_CROSS, Raw(expr))
	b.Joins = append(b.Joins, clause)
	b.Components[TYPE_JOIN] = struct{}{}

	return b
}

/*
join Add a join clause to the query.
*/
func (b *Builder) join(table, first, operator, second, joinType, isWhere interface{}) *Builder {
	//$table, $first, $operator = null, $second = null, $type = 'inner', $where = false
	b.Components[TYPE_JOIN] = struct{}{}

	if function, ok := first.(func(builder *Builder)); ok {
		clause := NewJoin(b, JOIN_TYPE_INNER, table)
		clause.Grammar.SetTablePrefix(b.Grammar.GetTablePrefix())
		function(clause)
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
joinWhere Add a "join where" clause to the query.
*/
func (b *Builder) joinWhere(table, firstColumn, joinOperator, secondColumn, joinType string) *Builder {
	return b.join(table, firstColumn, joinOperator, secondColumn, joinType, true)
}

/*
JoinWhere Add a "join where" clause to the query.
*/
func (b *Builder) JoinWhere(table, firstColumn, joinOperator, secondColumn string) *Builder {
	return b.joinWhere(table, firstColumn, joinOperator, secondColumn, JOIN_TYPE_INNER)
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

	return b.join(Raw(expr), first, operator, second, joinType, false)
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

	return b.join(Raw(expr), first, operator, second, joinType, false)
}

//AddBinding Add a binding to the query.
func (b *Builder) AddBinding(value []interface{}, bindingType string) *Builder {
	if _, ok := Bindings[bindingType]; !ok {
		log.Panicf("invalid binding type:%s\n", bindingType)
	}
	var tv []interface{}
	for _, v := range value {
		if _, ok := v.(Expression); !ok {
			tv = append(tv, v)
		}
	}
	b.Bindings[bindingType] = append(b.Bindings[bindingType], tv...)
	return b
}

//GetBindings Get the current query value bindings in a flattened slice.
func (b *Builder) GetBindings() (res []interface{}) {
	for key, _ := range Bindings {
		if bindings, ok := b.Bindings[key]; ok {
			res = append(res, bindings...)
		}
	}
	return
}

// GetRawBindings Get the raw map of array of bindings.
func (b *Builder) GetRawBindings() map[string][]interface{} {
	return b.Bindings
}

/*
MergeBindings Merge an array of bindings into our bindings.
*/
//func (b *Builder) MergeBindings(builder *Builder) *Builder {
//	res := make(map[string][]interface{})
//
//	for i, i2 := range collection {
//
//	}
//	return b
//}

/*
Where Add a basic where clause to the query.
column,operator,value,
*/
func (b *Builder) Where(params ...interface{}) *Builder {

	//map of where conditions
	if maps, ok := params[0].([][]interface{}); ok {
		for _, conditions := range maps {
			b.Where(conditions...)
		}
		return b
	}

	paramsLength := len(params)
	var operator string
	var value interface{}
	var boolean = BOOLEAN_AND
	if clousure, ok := params[0].(func(builder *Builder)); ok {
		var boolean string
		if paramsLength > 1 {
			boolean = params[1].(string)
		} else {
			boolean = BOOLEAN_AND
		}
		//clousure
		cb := CloneBuilder(b)
		clousure(cb)
		return b.AddNestedWhereQuery(cb, boolean)
	} else if where, ok := params[0].(Where); ok {
		b.Wheres = append(b.Wheres, where)
		b.Components[TYPE_WHERE] = struct{}{}
		return b
	} else if wheres, ok := params[0].([]Where); ok {
		b.Wheres = append(b.Wheres, wheres...)
		b.Components[TYPE_WHERE] = struct{}{}
		return b
	} else if e, ok := params[0].(Expression); ok {
		if paramsLength > 1 {
			boolean = params[1].(string)
		} else {
			boolean = BOOLEAN_AND
		}
		b.Wheres = append(b.Wheres, Where{
			Type:    CONDITION_TYPE_RAW,
			RawSql:  e,
			Boolean: boolean,
		})
		b.Components[TYPE_WHERE] = struct{}{}
		return b
	} else if m, ok := params[0].(map[string]interface{}); ok {
		boolean = BOOLEAN_AND
		if paramsLength > 1 {
			boolean = params[1].(string)
		}
		cb := CloneBuilder(b)
		for k, v := range m {
			cb.Where(k, v)
		}
		return b.AddNestedWhereQuery(cb, boolean)
	}
	switch paramsLength {
	case 2:
		//assume operator is = and omitted
		operator = "="
		value = params[1]
	case 3:
		//correspond to column,operator,value
		operator = params[1].(string)
		value = params[2]
	case 4:
		//correspond to column,operator,value,boolean jointer
		operator = params[1].(string)
		value = params[2]
		boolean = params[3].(string)
	}
	column := params[0].(string)
	//operator might be in/not in/between/not between,in there cases we need take value as slice
	if strings.Contains("in,not in,between,not between", operator) {
		switch operator {
		case CONDITION_TYPE_IN:
			b.WhereIn(column, value, boolean)
			return b
		case CONDITION_TYPE_NOT_IN:
			b.WhereNotIn(column, value, boolean)
			return b
		case CONDITION_TYPE_BETWEEN:
			b.WhereBetween(column, value, boolean)
			return b
		case CONDITION_TYPE_NOT_BETWEEN:
			b.WhereNotBetween(column, value, boolean)
			return b
		}
	}
	if f, ok := value.(func(builder *Builder)); ok {
		return b.WhereSub(column, operator, f, boolean)
	}
	b.Wheres = append(b.Wheres, Where{
		Type:     CONDITION_TYPE_BASIC,
		Column:   column,
		Operator: operator,
		Value:    value,
		Boolean:  boolean,
	})
	b.AddBinding([]interface{}{value}, TYPE_WHERE)
	b.Components[TYPE_WHERE] = struct{}{}
	return b
}
func (b *Builder) WherePivot(params ...interface{}) *Builder {

	column := params[0].(string)
	paramsLength := len(params)
	var operator string
	var value interface{}
	var boolean = BOOLEAN_AND
	switch paramsLength {
	case 2:
		operator = "="
		value = params[1]
	case 3:
		operator = params[1].(string)
		value = params[2]
	case 4:
		operator = params[1].(string)
		value = params[2]
		boolean = params[3].(string)
	}

	b.PivotWheres = append(b.PivotWheres, Where{
		Type:     CONDITION_TYPE_BASIC,
		Column:   column,
		Operator: operator,
		Value:    value,
		Boolean:  boolean,
	})
	return b
}

/*
OrWhere Add an "or where" clause to the query.
*/
func (b *Builder) OrWhere(params ...interface{}) *Builder {
	paramsLength := len(params)
	if clousure, ok := params[0].(func(builder *Builder)); ok {
		return b.Where(clousure, BOOLEAN_OR)
	}
	if paramsLength == 2 {
		params = []interface{}{params[0], "=", params[1], BOOLEAN_OR}
	} else {
		params = append(params, BOOLEAN_OR)
	}
	return b.Where(params...)
}

/*
WhereColumn Add a "where" clause comparing two columns to the query.
*/
func (b *Builder) WhereColumn(first interface{}, second ...string) *Builder {
	length := len(second)
	var firstColumn = first
	var secondColumn, operator, boolean string
	if arr, ok := first.([][]interface{}); ok {
		return b.WhereNested(func(builder *Builder) {
			for _, term := range arr {
				var strs []string
				for i := 1; i < len(term); i++ {
					strs = append(strs, term[i].(string))
				}
				builder.WhereColumn(term[0], strs...)
			}
		})

	}
	switch length {
	case 1:
		secondColumn = second[0]
		operator = "="
		boolean = BOOLEAN_AND
	case 2:
		operator = second[0]
		secondColumn = second[1]
		boolean = BOOLEAN_AND
	case 3:
		operator = second[0]
		secondColumn = second[1]
		boolean = second[2]
	default:
		panic("wrong arguements in where column")
	}
	b.Wheres = append(b.Wheres, Where{
		Type:         CONDITION_TYPE_COLUMN,
		FirstColumn:  firstColumn.(string),
		Operator:     operator,
		SecondColumn: secondColumn,
		Boolean:      boolean,
	})
	b.Components[TYPE_WHERE] = struct{}{}

	return b
}

/*
OrWhereColumn Add an "or where" clause comparing two columns to the query.
*/
func (b *Builder) OrWhereColumn(first string, second ...string) *Builder {
	var ts = make([]string, 3, 3)
	switch len(second) {
	case 1:
		ts = []string{"=", second[0], BOOLEAN_OR}
	case 2:
		ts = []string{second[0], second[1], BOOLEAN_OR}
	}
	return b.WhereColumn(first, ts...)
}

/*
WhereRaw Add a raw where clause to the query.
*/
func (b *Builder) WhereRaw(rawSql string, params ...interface{}) *Builder {
	paramsLength := len(params)
	var boolean string
	var bindings []interface{}
	switch paramsLength {
	case 1:
		bindings = params[0].([]interface{})
		b.AddBinding(bindings, TYPE_WHERE)
		boolean = BOOLEAN_AND
	case 2:
		bindings = params[0].([]interface{})
		b.AddBinding(bindings, TYPE_WHERE)
		boolean = params[1].(string)
	}
	b.Wheres = append(b.Wheres, Where{
		Type:    CONDITION_TYPE_RAW,
		RawSql:  Raw(rawSql),
		Boolean: boolean,
	})
	b.Components[TYPE_WHERE] = struct{}{}

	return b

}

/*
OrWhereRaw Add a raw or where clause to the query.
*/
func (b *Builder) OrWhereRaw(rawSql string, bindings ...[]interface{}) *Builder {
	switch len(bindings) {
	case 0:
		return b.WhereRaw(rawSql, []interface{}{}, BOOLEAN_OR)
	case 1:
		return b.WhereRaw(rawSql, bindings[0], BOOLEAN_OR)
	}
	panic(errors.New("arguements mismatch"))
}

/*
WhereIn Add a "where in" clause to the query.
column values boolean not
*/
func (b *Builder) WhereIn(params ...interface{}) *Builder {
	paramsLength := len(params)
	var boolean string
	not := false
	if paramsLength > 2 {
		boolean = params[2].(string)
	} else {
		boolean = BOOLEAN_AND
	}
	if paramsLength > 3 {
		not = params[3].(bool)
	}

	var values []interface{}
	if IsQueryable(params[1]) {
		queryStr, bindings := b.CreateSub(params[1])
		values = append(values, Raw(queryStr))
		b.AddBinding(bindings, TYPE_WHERE)
	} else {
		values = InterfaceToSlice(params[1])
	}
	b.Wheres = append(b.Wheres, Where{
		Type:    CONDITION_TYPE_IN,
		Column:  params[0].(string),
		Values:  values,
		Boolean: boolean,
		Not:     not,
	})
	b.Components[TYPE_WHERE] = struct{}{}
	b.AddBinding(values, TYPE_WHERE)
	return b
}

/*
OrWhereIn Add an "or where in" clause to the query.
column values
*/
func (b *Builder) OrWhereIn(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_OR, false)
	return b.WhereIn(params...)
}

//column values [ boolean ]
func (b *Builder) WhereNotIn(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_AND, true)
	return b.WhereIn(params...)
}

/*
OrWhereNotIn Add an "or where not in" clause to the query.
column values
*/
func (b *Builder) OrWhereNotIn(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_OR, true)
	return b.WhereIn(params...)
}

/*
WhereNull Add a "where null" clause to the query.
   params takes in below order:
   1. column string
   2. boolean string in [2]string{"and","or"}
   3. type string "not"
*/
func (b *Builder) WhereNull(column interface{}, params ...interface{}) *Builder {
	paramsLength := len(params)
	var boolean = BOOLEAN_AND
	var not = false
	switch paramsLength {
	case 1:
		boolean = params[0].(string)
	case 2:
		boolean = params[0].(string)
		not = params[1].(bool)
	}
	b.Components[TYPE_WHERE] = struct{}{}
	if single, ok := column.(string); ok {
		b.Wheres = append(b.Wheres, Where{
			Type:    CONDITION_TYPE_NULL,
			Column:  single,
			Boolean: boolean,
			Not:     not,
		})
	} else if slice, ok := column.([]interface{}); ok {
		for _, i := range slice {
			b.Wheres = append(b.Wheres, Where{
				Type:    CONDITION_TYPE_NULL,
				Column:  i.(string),
				Boolean: boolean,
				Not:     not,
			})
		}
	} else if slice, ok := column.([]string); ok {
		for _, i := range slice {
			b.Wheres = append(b.Wheres, Where{
				Type:    CONDITION_TYPE_NULL,
				Column:  i,
				Boolean: boolean,
				Not:     not,
			})
		}
	}

	return b
}

/*
WhereNotNull Add a "where not null" clause to the query.
*/
func (b *Builder) WhereNotNull(column interface{}, params ...interface{}) *Builder {
	paramsLength := len(params)
	if paramsLength == 0 {
		params = append(params, BOOLEAN_AND, true)
	} else if paramsLength == 1 {
		params = append(params, true)
	}
	return b.WhereNull(column, params...)
}

/*
OrWhereNull Add an "or where null" clause to the query.
column not
*/
func (b *Builder) OrWhereNull(column interface{}, params ...interface{}) *Builder {
	paramsLength := len(params)
	if paramsLength == 0 {
		params = append(params, BOOLEAN_OR, false)
	} else if paramsLength == 1 {
		params = []interface{}{BOOLEAN_OR, params[0]}
	}
	return b.WhereNull(column, params...)
}

/*
OrWhereNotNull Add an "or where not null" clause to the query.
*/
func (b *Builder) OrWhereNotNull(column interface{}) *Builder {
	params := []interface{}{BOOLEAN_OR, true}
	return b.WhereNull(column, params...)
}

/*
WhereBetween Add a where between statement to the query.

   params takes in below order:
   1. WhereBetween(column string,values []interface{"min","max"})
   2. WhereBetween(column string,values []interface{"min","max"},"and/or")
   3. WhereBetween(column string,values []interface{"min","max","and/or",true/false})

*/
func (b *Builder) WhereBetween(params ...interface{}) *Builder {
	paramsLength := len(params)
	var boolean = BOOLEAN_AND
	not := false
	if paramsLength > 2 {
		boolean = params[2].(string)
	}
	var betweenType = CONDITION_TYPE_BETWEEN
	if paramsLength > 3 {
		not = params[3].(bool)
	}
	b.Components[TYPE_WHERE] = struct{}{}
	tvalues := params[1].([]interface{})[0:2]
	for _, tvalue := range tvalues {
		if _, ok := tvalue.(Expression); !ok {
			b.AddBinding([]interface{}{tvalue}, TYPE_WHERE)
		}
	}

	b.Wheres = append(b.Wheres, Where{
		Type:    betweenType,
		Column:  params[0].(string),
		Boolean: boolean,
		Values:  params[1].([]interface{})[0:2],
		Not:     not,
	})
	return b
}
func (b *Builder) WhereNotBetween(params ...interface{}) *Builder {
	if len(params) == 2 {
		params = append(params, BOOLEAN_AND, true)
	}
	return b.WhereBetween(params...)
}

/*
OrWhereBetween Add an or where between statement to the query.
*/
func (b *Builder) OrWhereBetween(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_OR)
	return b.WhereBetween(params...)
}

/*
OrWhereNotBetween Add an or where not between statement to the query.
*/
func (b *Builder) OrWhereNotBetween(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_OR, true)

	return b.WhereBetween(params...)
}

/*
WhereBetweenColumns Add a where between statement using columns to the query.
*/
func (b *Builder) WhereBetweenColumns(column string, values []interface{}, params ...interface{}) *Builder {
	paramsLength := len(params)
	var boolean = BOOLEAN_AND
	var betweenType = CONDITION_TYPE_BETWEEN_COLUMN
	not := false
	if paramsLength > 0 {
		boolean = params[0].(string)
	}
	if paramsLength > 1 {
		not = params[1].(bool)
	}

	b.Components[TYPE_WHERE] = struct{}{}

	b.Wheres = append(b.Wheres, Where{
		Type:    betweenType,
		Column:  column,
		Boolean: boolean,
		Values:  values,
		Not:     not,
	})
	return b
}

//AddTimeBasedWhere Add a time based (year, month, day, time) statement to the query.
//params order : timefuncionname column operator value boolean
//minimum : timefuncionname column value
func (b *Builder) AddTimeBasedWhere(params ...interface{}) *Builder {
	paramsLength := len(params)
	var timeType = params[0]
	var boolean = BOOLEAN_AND
	var operator string
	var value interface{}
	var tvalue interface{}
	//timefunction column value
	if paramsLength == 3 {
		operator = "="
		tvalue = params[2]
	} else if paramsLength > 3 {
		//timefunction column operator value
		operator = params[2].(string)
		tvalue = params[3]
		//timefunction column operator value boolean
		if paramsLength > 4 && params[4].(string) != boolean {
			boolean = BOOLEAN_OR
		}
	} else {
		tvalue = params[3]
	}
	switch tvalue.(type) {
	case string:
		value = tvalue.(string)
	case int:
		value = tvalue.(int)
	case time.Time:
		switch timeType.(string) {
		case CONDITION_TYPE_DATE:
			value = tvalue.(time.Time).Format("2006-01-02")
		case CONDITION_TYPE_MONTH:
			value = tvalue.(time.Time).Format("01")
		case CONDITION_TYPE_YEAR:
			value = tvalue.(time.Time).Format("2006")
		case CONDITION_TYPE_TIME:
			value = tvalue.(time.Time).Format("15:04:05")
		case CONDITION_TYPE_DAY:
			value = tvalue.(time.Time).Format("02")
		}
	case Expression:
		value = tvalue.(Expression)
	}
	b.Wheres = append(b.Wheres, Where{
		Type:     timeType.(string),
		Column:   params[1].(string),
		Boolean:  boolean,
		Value:    value,
		Operator: operator,
	})
	b.AddBinding([]interface{}{value}, TYPE_WHERE)
	b.Components[TYPE_WHERE] = struct{}{}
	return b
}

//column operator value boolean
func (b *Builder) WhereDate(params ...interface{}) *Builder {
	p := append([]interface{}{CONDITION_TYPE_DATE}, params...)
	return b.AddTimeBasedWhere(p...)
}
func (b *Builder) WhereTime(params ...interface{}) *Builder {
	p := append([]interface{}{CONDITION_TYPE_TIME}, params...)
	return b.AddTimeBasedWhere(p...)
}
func (b *Builder) WhereDay(params ...interface{}) *Builder {
	p := append([]interface{}{CONDITION_TYPE_DAY}, params...)
	return b.AddTimeBasedWhere(p...)
}
func (b *Builder) WhereMonth(params ...interface{}) *Builder {
	p := append([]interface{}{CONDITION_TYPE_MONTH}, params...)
	return b.AddTimeBasedWhere(p...)
}
func (b *Builder) WhereYear(params ...interface{}) *Builder {
	p := append([]interface{}{CONDITION_TYPE_YEAR}, params...)
	return b.AddTimeBasedWhere(p...)
}

/*
WhereNested Add a nested where statement to the query.
*/
func (b *Builder) WhereNested(params ...interface{}) *Builder {
	paramsLength := len(params)
	if paramsLength == 1 {
		params = append(params, BOOLEAN_AND)
	}
	cb := CloneBuilder(b)
	switch params[0].(type) {
	case Where:
		cb.Wheres = append(cb.Wheres, params[0].(Where))
	case [][]interface{}:
		tp := params[0].([][]interface{})
		for i := 0; i < len(tp); i++ {
			cb.Where(tp[i]...)
		}
	case []interface{}:
		cb.Where(params[0].([]interface{}))
	case func(builder *Builder):
		var boolean string
		if paramsLength > 1 {
			boolean = params[1].(string)
		} else {
			boolean = BOOLEAN_AND
		}
		closure := params[0].(func(builder *Builder))
		closure(cb)
		return b.AddNestedWhereQuery(cb, boolean)
	}
	b.Wheres = append(b.Wheres, Where{
		Type:    CONDITION_TYPE_NESTED,
		Boolean: params[1].(string),
		Value:   cb,
	})
	b.Components[TYPE_WHERE] = struct{}{}
	return b
}
func (b *Builder) WhereSub(column string, operator string, value func(builder *Builder), boolean string) *Builder {
	cb := CloneBuilder(b)
	value(cb)
	b.Wheres = append(b.Wheres, Where{
		Type:     CONDITION_TYPE_SUB,
		Operator: operator,
		Value:    cb,
		Column:   column,
		Boolean:  boolean,
	})
	b.Components[TYPE_WHERE] = struct{}{}
	b.AddBinding(cb.GetBindings(), TYPE_WHERE)
	return b
}

//WhereExists Add an exists clause to the query.
// 1. WhereExists(cb,"and",false)
// 2. WhereExists(cb,"and")
// 3. WhereExists(cb)
func (b *Builder) WhereExists(cb func(builder *Builder), params ...interface{}) *Builder {
	newBuilder := CloneBuilder(b)
	cb(newBuilder)
	boolean := BOOLEAN_AND
	not := false
	switch len(params) {
	case 1:
		boolean = params[0].(string)
	case 2:
		boolean = params[0].(string)
		not = params[1].(bool)
	}

	return b.AddWhereExistsQuery(newBuilder, boolean, not)
}

/*
OrWhereExists Add an exists clause to the query.
*/
func (b *Builder) OrWhereExists(cb func(builder *Builder), params ...interface{}) *Builder {
	not := false
	if len(params) > 0 {
		not = params[0].(bool)
	}
	return b.WhereExists(cb, BOOLEAN_OR, not)
}

/*
WhereNotExists Add a where not exists clause to the query.
*/
func (b *Builder) WhereNotExists(cb func(builder *Builder), params ...interface{}) *Builder {
	boolean := BOOLEAN_AND
	if len(params) > 0 {
		boolean = params[0].(string)
	}
	return b.WhereExists(cb, boolean, true)
}

/*
OrWhereNotExists Add a where not exists clause to the query.
*/
func (b *Builder) OrWhereNotExists(cb func(builder *Builder), params ...interface{}) *Builder {
	return b.OrWhereExists(cb, true)
}

// AddWhereExistsQuery  Add an exists clause to the query.
func (b *Builder) AddWhereExistsQuery(builder *Builder, boolean string, not bool) *Builder {
	var n bool
	if not {
		n = true
	} else {
		n = false
	}
	b.Wheres = append(b.Wheres, Where{
		Type:    CONDITION_TYPE_EXIST,
		Query:   builder,
		Boolean: boolean,
		Not:     n,
	})
	b.Components[TYPE_WHERE] = struct{}{}
	b.AddBinding(builder.GetBindings(), TYPE_WHERE)
	return b
}

/*
GroupBy Add a "group by" clause to the query.
column operator value boolean
*/
func (b *Builder) GroupBy(columns ...interface{}) *Builder {
	for _, column := range columns {
		b.Groups = append(b.Groups, column)
	}
	b.Components[TYPE_GROUP_BY] = struct{}{}
	return b
}

/*
GroupByRaw Add a raw groupBy clause to the query.
*/
func (b *Builder) GroupByRaw(sql string, bindings ...[]interface{}) *Builder {
	b.Groups = append(b.Groups, Expression(sql))
	if len(bindings) > 0 {
		b.AddBinding(bindings[0], TYPE_GROUP_BY)
	}
	b.Components[TYPE_GROUP_BY] = struct{}{}

	return b
}

/*
Having Add a "having" clause to the query.
column operator value boolean
*/
func (b *Builder) Having(params ...interface{}) *Builder {
	havingBoolean := BOOLEAN_AND
	havingOperator := "="
	var havingValue interface{}
	var havingColumn string
	length := len(params)
	switch length {
	case 2:
		havingColumn = params[0].(string)
		havingValue = params[1]
	case 3:
		havingColumn = params[0].(string)
		havingOperator = params[1].(string)
		havingValue = params[2]
	case 4:
		havingColumn = params[0].(string)
		havingOperator = params[1].(string)
		havingValue = params[2]
		havingBoolean = params[3].(string)
	}

	having := Having{
		HavingType:     CONDITION_TYPE_BASIC,
		HavingColumn:   havingColumn,
		HavingOperator: havingOperator,
		HavingValue:    havingValue,
		HavingBoolean:  havingBoolean,
	}
	b.AddBinding([]interface{}{havingValue}, TYPE_HAVING)
	b.Components[TYPE_HAVING] = struct{}{}
	b.Havings = append(b.Havings, having)
	return b
}

/*
HavingRaw Add a raw having clause to the query.
*/
func (b *Builder) HavingRaw(params ...interface{}) *Builder {
	length := len(params)
	havingBoolean := BOOLEAN_AND
	var expression Expression
	switch length {
	case 1:
		if expr, ok := params[0].(Expression); ok {
			expression = expr
		} else {
			expression = Expression(params[0].(string))
		}
	case 2:
		if expr, ok := params[0].(Expression); ok {
			expression = expr
		} else {
			expression = Expression(params[0].(string))
		}
		b.AddBinding(params[1].([]interface{}), TYPE_HAVING)
	case 3:
		if expr, ok := params[0].(Expression); ok {
			expression = expr
		} else {
			expression = Expression(params[0].(string))
		}
		b.AddBinding(params[1].([]interface{}), TYPE_HAVING)
		havingBoolean = params[2].(string)
	}
	having := Having{
		HavingType:    CONDITION_TYPE_RAW,
		HavingValue:   expression,
		HavingBoolean: havingBoolean,
		RawSql:        expression,
	}
	b.Components[TYPE_HAVING] = struct{}{}
	b.Havings = append(b.Havings, having)
	return b
}

/*
OrHavingRaw Add a raw having clause to the query.
*/
func (b *Builder) OrHavingRaw(params ...interface{}) *Builder {
	bindings := []interface{}{}
	if len(params) == 2 {
		bindings = params[1].([]interface{})
	}
	return b.HavingRaw(params[0], bindings, BOOLEAN_OR)
}

/*
OrHaving Add an "or having" clause to the query.
*/
func (b *Builder) OrHaving(params ...interface{}) *Builder {
	return b.Having(params[0], "=", params[1], BOOLEAN_OR)
}

/*
HavingBetween Add a "having between " clause to the query.
*/
func (b *Builder) HavingBetween(column string, params ...interface{}) *Builder {
	var values []interface{}
	boolean := BOOLEAN_AND
	not := false
	length := len(params)
	switch length {
	case 1:
		values = params[0].([]interface{})[0:2]
	case 2:
		values = params[0].([]interface{})[0:2]
		boolean = params[1].(string)
	case 3:
		values = params[0].([]interface{})[0:2]
		boolean = params[1].(string)
		not = params[2].(bool)
	}
	having := Having{
		HavingType:    CONDITION_TYPE_BETWEEN,
		HavingColumn:  column,
		HavingValue:   values,
		HavingBoolean: boolean,
		Not:           not,
	}
	b.Components[TYPE_HAVING] = struct{}{}
	b.AddBinding(values, TYPE_HAVING)

	b.Havings = append(b.Havings, having)
	return b
}

/*
OrderBy Add an "order by" clause to the query.
*/
func (b *Builder) OrderBy(params ...interface{}) *Builder {
	var order = ORDER_ASC
	if r, ok := params[0].(Expression); ok {
		b.Orders = append(b.Orders, Order{
			RawSql:    r,
			OrderType: CONDITION_TYPE_RAW,
		})
		b.Components[TYPE_ORDER] = struct{}{}

		return b
	}
	if len(params) > 1 {
		order = params[1].(string)
	}
	if order != ORDER_ASC && order != ORDER_DESC {
		panic(errors.New("wrong order direction: " + order))
	}
	b.Orders = append(b.Orders, Order{
		Direction: order,
		Column:    params[0].(string),
	})
	b.Components[TYPE_ORDER] = struct{}{}

	return b
}

/*
OrderByRaw Add a raw "order by" clause to the query.
*/
func (b *Builder) OrderByRaw(sql string, bindings []interface{}) *Builder {
	b.Orders = append(b.Orders, Order{
		OrderType: CONDITION_TYPE_RAW,
		RawSql:    Raw(sql),
	})
	b.Components[TYPE_ORDER] = struct{}{}
	b.AddBinding(bindings, TYPE_ORDER)
	return b
}
func (b *Builder) OrderByDesc(column string) *Builder {
	return b.OrderBy(column, ORDER_DESC)
}

/*
ReOrder Remove all existing orders and optionally add a new order.
*/
func (b *Builder) ReOrder(params ...string) *Builder {
	b.Orders = nil
	b.Bindings["order"] = nil
	delete(b.Components, TYPE_ORDER)
	length := len(params)
	if length == 1 {
		b.OrderBy(InterfaceToSlice(params[0]))
	} else if length == 2 {
		b.OrderBy(InterfaceToSlice(params[0:2])...)
	}
	return b
}

/*
Limit Set the "limit" value of the query.
*/
func (b *Builder) Limit(n int) *Builder {
	b.Components[TYPE_LIMIT] = struct{}{}
	b.LimitNum = int(math.Max(0, float64(n)))
	return b
}

/*
Offset Set the "offset" value of the query.
*/
func (b *Builder) Offset(n int) *Builder {
	b.OffsetNum = int(math.Max(0, float64(n)))
	b.Components[TYPE_OFFSET] = struct{}{}
	return b
}

/*
Union Add a union statement to the query.
*/
//func (b *Builder) Union(n int) *Builder {
//	b.OffsetNum = n
//	return b
//}

/*
Lock Lock the selected rows in the table for updating.
*/
func (b *Builder) Lock(lock ...interface{}) *Builder {
	if len(lock) == 0 {
		b.LockMode = true
	} else {
		b.LockMode = lock[0]
	}
	b.Components[TYPE_LOCK] = struct{}{}
	return b
}

func (b *Builder) WhereKey(keys interface{}) *Builder {
	pt := reflect.TypeOf(keys)
	var primaryKeyColumn string
	if b.Model != nil {
		primaryKeyColumn = b.Model.PrimaryKey.ColumnName
	} else {
		primaryKeyColumn = "id"
	}
	if pt.Kind() == reflect.Slice {
		b.WhereIn(primaryKeyColumn, keys)
	} else {
		b.Where(primaryKeyColumn, keys)
	}
	return b
}
func (b *Builder) WhereMap(params map[string]interface{}) *Builder {
	for key, param := range params {
		b.Where(key, "=", param)
	}
	return b
}
func (b *Builder) WhereModel(model interface{}) *Builder {
	v := reflect.Indirect(reflect.ValueOf(model))
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.IsValid() && !f.IsZero() && b.Model.Fields[i].ColumnName != "" {
			b.Where(b.Model.Fields[i].ColumnName, "=", f.Interface())
		}
	}
	return b
}

//AddNestedWhereQuery Add another query builder as a nested where to the query builder.
func (b *Builder) AddNestedWhereQuery(builder *Builder, boolean string) *Builder {

	if len(builder.Wheres) > 0 {
		b.Wheres = append(b.Wheres, Where{
			Type:    CONDITION_TYPE_NESTED,
			Value:   builder,
			Boolean: boolean,
		})
		b.AddBinding(builder.GetRawBindings()[TYPE_WHERE], TYPE_WHERE)
		b.Components[TYPE_WHERE] = struct{}{}
	}
	return b
}

func (b *Builder) Commit() error {
	return b.Tx.Commit()
}
func (b *Builder) Rollback() error {
	return b.Tx.RollBack()
}

/*
Find Execute a query for a single record by ID.
*/
func (b *Builder) Find(dest interface{}, params interface{}) (result sql.Result, err error) {
	b.WhereKey(params)
	d := reflect.Indirect(reflect.ValueOf(dest))
	if d.Type().Kind() == reflect.Slice {
		return b.Get(dest)
	} else {
		return b.First(dest)
	}
}

/*

 */
func (b *Builder) First(dest interface{}) (result sql.Result, err error) {
	b.Limit(1)
	return b.Get(dest)
}

//parameter model should either be a model pointer or a reflect.Type
func (b *Builder) SetModel(model interface{}) *Builder {
	if model != nil {
		b.Model = GetParsedModel(model)
		b.From(b.Model.Table)
	}
	if b.Connection == nil {
		var typ reflect.Type
		if t, ok := model.(reflect.Type); ok {
			typ = t
		} else {
			typ = reflect.TypeOf(model)
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
		}
		if c, ok := reflect.New(typ).Elem().Interface().(ConnectionName); ok {
			b.Connection = Eloquent.Connection(c.ConnectionName())
		} else {
			b.Connection = Eloquent.Connection("default")

		}
	}
	return b

}

/*
RunSelect Run the query as a "select" statement against the connection.
*/
func (b *Builder) RunSelect() (result sql.Result, err error) {
	if b.Tx != nil {
		result, err = b.Tx.Select(b.Grammar.CompileSelect(), b.GetBindings(), b.Dest)
	} else {
		result, err = b.Connection.Select(b.Grammar.CompileSelect(), b.GetBindings(), b.Dest)
	}
	if err != nil {
		return
	}
	d := reflect.TypeOf(b.Dest).Elem()
	if d.Kind() == reflect.Slice {
		d = d.Elem()
	}
	if b.Model != nil && b.Model.IsEloquent && d.Kind() == reflect.Struct {
		c, _ := result.RowsAffected()
		BatchSync(b.Dest, c > 0)
	}
	return
}
func (b *Builder) WithPivot(columns ...string) *Builder {
	b.Pivots = append(b.Pivots, columns...)
	return b
}

/*
Exists Determine if any rows exist for the current query.
*/
func (b *Builder) Exists() (exists bool) {
	b.Connection.Select(b.Grammar.CompileExists(), b.GetBindings(), &exists)
	return exists
}

/*
DoesntExist Determine if no rows exist for the current query.
*/
func (b *Builder) DoesntExist() bool {
	return !b.Exists()
}

/*
Aggregate Execute an aggregate function on the database.

*/
func (b *Builder) Aggregate(dest interface{}, fn string, column ...string) (result sql.Result, err error) {
	var start = time.Now()
	b.Dest = dest
	if column == nil {
		column = append(column, "*")
	}
	b.Aggregates = append(b.Aggregates, Aggregate{
		AggregateName:   fn,
		AggregateColumn: column[0],
	})
	b.Components[TYPE_AGGREGRATE] = struct{}{}
	result, err = b.RunSelect()
	b.logQuery(b.PreparedSql, b.GetBindings(), time.Since(start), result)
	return
}

/*
Count Retrieve the "count" result of the query.
*/
func (b *Builder) Count(dest interface{}, column ...string) (result sql.Result, err error) {
	return b.Aggregate(dest, "count", column...)
}

/*
Min Retrieve the minimum value of a given column.
*/
func (b *Builder) Min(dest interface{}, column ...string) (result sql.Result, err error) {

	return b.Aggregate(dest, "min", column...)

}

/*
Max Retrieve the maximum value of a given column.
*/
func (b *Builder) Max(dest interface{}, column ...string) (result sql.Result, err error) {

	return b.Aggregate(dest, "max", column...)

}

/*
Avg Alias for the "avg" method.
*/
func (b *Builder) Avg(dest interface{}, column ...string) (result sql.Result, err error) {

	return b.Aggregate(dest, "avg", column...)

}
func (b *Builder) Sum(dest interface{}, column ...string) (result sql.Result, err error) {

	return b.Aggregate(dest, "sum", column...)

}
func (b *Builder) ForPage(page, perPage int64) *Builder {

	b.Offset(int((page - 1) * perPage)).Limit(int(perPage))
	return b
}
func (b *Builder) Only(columns ...string) *Builder {
	b.OnlyColumns = make(map[string]interface{}, len(columns))
	for i := 0; i < len(columns); i++ {
		b.OnlyColumns[columns[i]] = nil
	}
	return b
}
func (b *Builder) Except(columns ...string) *Builder {
	b.ExceptColumns = make(map[string]interface{}, len(columns))
	for i := 0; i < len(columns); i++ {
		b.ExceptColumns[columns[i]] = nil
	}
	return b
}
func (b *Builder) FileterColumn(column string) bool {
	if b.OnlyColumns != nil {
		if _, ok := b.OnlyColumns[column]; !ok {
			return false
		}
	}
	if b.ExceptColumns != nil {
		if _, ok := b.ExceptColumns[column]; ok {
			return false
		}
	}
	return true
}

func PrepareInsertValues(values interface{}) []map[string]interface{} {
	rv := reflect.ValueOf(values)
	var items []map[string]interface{}
	if rv.Kind() == reflect.Ptr {
		rv = reflect.Indirect(rv)
	}
	if rv.Kind() == reflect.Map {
		items = append(items, rv.Interface().(map[string]interface{}))
	} else if rv.Kind() == reflect.Slice {
		eleType := rv.Type().Elem()
		if eleType.Kind() == reflect.Ptr {
			switch eleType.Elem().Kind() {
			case reflect.Struct:
				for i := 0; i < rv.Len(); i++ {
					items = append(items, ExtractStruct(rv.Index(i).Elem().Interface()))
				}
			case reflect.Map:
				for i := 0; i < rv.Len(); i++ {
					items = append(items, rv.Index(i).Elem().Interface().(map[string]interface{}))
				}
			}
		} else if eleType.Kind() == reflect.Map {
			if tv, ok := values.(*[]map[string]interface{}); ok {
				items = *tv
			} else {
				items = values.([]map[string]interface{})
			}
		} else if eleType.Kind() == reflect.Struct {
			for i := 0; i < rv.Len(); i++ {
				items = append(items, ExtractStruct(rv.Index(i).Interface()))
			}
		}
	} else if rv.Kind() == reflect.Struct {
		items = append(items, ExtractStruct(rv.Interface()))
	}
	return items
}

/*
Insert new records into the database.
*/
func (b *Builder) Insert(values interface{}) (result sql.Result, err error) {
	var start = time.Now()
	items := PrepareInsertValues(values)
	b.Grammar.CompileInsert(items)

	if b.Tx != nil {
		result, err = b.Tx.Insert(b.PreparedSql, b.GetBindings())
	} else {
		result, err = b.Connection.Insert(b.PreparedSql, b.GetBindings())
	}

	b.logQuery(b.PreparedSql, b.GetBindings(), time.Since(start), result)
	return result, err
}

/*
InsertGetId Insert a new record and get the value of the primary key.
*/
func (b *Builder) InsertGetId(values interface{}) (int64, error) {

	insert, err := b.Insert(values)
	if err != nil {
		return 0, err
	}
	id, _ := insert.LastInsertId()
	return id, nil
}

/*
InsertOrIgnore Insert a new record and get the value of the primary key.
*/
func (b *Builder) InsertOrIgnore(values interface{}) (result sql.Result, err error) {

	var start = time.Now()
	items := PrepareInsertValues(values)
	b.Grammar.CompileInsertOrIgnore(items)

	if b.Tx != nil {
		result, err = b.Tx.Insert(b.PreparedSql, b.GetBindings())
	} else {
		result, err = b.Connection.Insert(b.PreparedSql, b.GetBindings())
	}

	b.logQuery(b.PreparedSql, b.GetBindings(), time.Since(start), result)
	return result, err
}
func (b *Builder) Update(v map[string]interface{}) (result sql.Result, err error) {
	var start = time.Now()
	b.Grammar.CompileUpdate(v)
	if b.Tx != nil {
		result, err = b.Tx.Update(b.PreparedSql, b.GetBindings())
	} else {
		result, err = b.Connection.Update(b.PreparedSql, b.GetBindings())
	}
	b.logQuery(b.PreparedSql, b.GetBindings(), time.Since(start), result)
	return result, err
}
func (b *Builder) UpdateOrInsert(n int) interface{} {

	return 1
}

/*
Decrement Decrement a column's value by a given amount.
*/
func (b *Builder) Decrement(column string, amount int, extra ...map[string]interface{}) (result sql.Result, err error) {

	var update map[string]interface{}
	wrapped := b.Grammar.Wrap(column)

	if len(extra) == 0 {
		update = make(map[string]interface{})
	} else {
		update = extra[0]
	}
	update[column] = Expression(fmt.Sprintf("%s - %d", wrapped, amount))

	return b.Update(update)
}

/*
Increment Increment a column's value by a given amount.
*/
func (b *Builder) Increment(column string, amount int, extra ...map[string]interface{}) (result sql.Result, err error) {

	var update map[string]interface{}
	wrapped := b.Grammar.Wrap(column)

	if len(extra) == 0 {
		update = make(map[string]interface{})
	} else {
		update = extra[0]
	}
	update[column] = Expression(fmt.Sprintf("%s + %d", wrapped, amount))

	return b.Update(update)
}
func (b *Builder) InRandomOrder() {

}

/*
Delete Delete records from the database.
//TODO: Delete(1) Delete(1,2,3) Delete([]interface{}{1,2,3})
*/
func (b *Builder) Delete() (result sql.Result, err error) {
	var start = time.Now()
	b.Grammar.CompileDelete()
	if b.Tx != nil {
		result, err = b.Tx.Delete(b.PreparedSql, b.GetBindings())
	} else {
		result, err = b.Connection.Delete(b.PreparedSql, b.GetBindings())
	}
	b.logQuery(b.PreparedSql, b.GetBindings(), time.Since(start), result)
	return result, err
}
func (b *Builder) Raw() *sql.DB {
	return b.Connection.GetDB()
}

func (b *Builder) With(relations ...interface{}) *Builder {
	var name string
	var res = make(map[string]func(*RelationBuilder) *RelationBuilder)
	for _, relation := range relations {
		switch relation.(type) {
		case string:
			name = relation.(string)
			res = b.addNestedWiths(name, res)
			res[name] = func(builder *RelationBuilder) *RelationBuilder {
				return builder
			}
		case []string:
			for _, r := range relation.([]string) {
				name = r
				res = b.addNestedWiths(name, res)
				res[name] = func(builder *RelationBuilder) *RelationBuilder {
					return builder
				}
			}
		case map[string]func(builder *RelationBuilder) *RelationBuilder:
			for relationName, fn := range relation.(map[string]func(builder *RelationBuilder) *RelationBuilder) {
				name = relationName
				res = b.addNestedWiths(name, res)
				res[relationName] = fn
			}
		}
	}

	for s, f := range res {
		b.EagerLoad[s] = f
	}
	return b
}
func (b *Builder) addNestedWiths(name string, results map[string]func(builder *RelationBuilder) *RelationBuilder) map[string]func(builder *RelationBuilder) *RelationBuilder {
	var progress []string
	for _, segment := range strings.Split(name, ".") {
		progress = append(progress, segment)
		var ts string
		for j := 0; j < len(progress); j++ {
			ts = strings.Join(progress, ".")
		}
		if _, ok := results[ts]; !ok {
			results[ts] = func(builder *RelationBuilder) *RelationBuilder {
				return builder
			}
		}
	}
	return results
}
func (b *Builder) logQuery(query string, bindings []interface{}, elapsed time.Duration, result sql.Result) {
	if b.LoggingQueries && Eloquent.LogFunc != nil {
		Eloquent.LogFunc(Log{
			SQL:      query,
			Bindings: bindings,
			Result:   result,
			Time:     elapsed,
		})
	}

}

/*
Get Execute the query as a "select" statement.
*/
func (b *Builder) Get(dest interface{}, columns ...interface{}) (result sql.Result, err error) {
	if b.FromTable == nil {
		b.Model = GetParsedModel(dest)
		b.From(b.Model.Table)
		b.Connection = Eloquent.Connection(b.Model.ConnectionName)
		b.Grammar.SetTablePrefix(Eloquent.Configs[b.Model.ConnectionName].Prefix)

	}
	if len(columns) > 0 {
		b.Select(columns...)
	}
	var start = time.Now()
	b.Dest = dest
	b.DestReflectValue = reflect.ValueOf(dest)
	if _, ok := b.Components[TYPE_COLUMN]; !ok || len(b.Columns) == 0 {
		b.Components[TYPE_COLUMN] = struct{}{}
		b.Columns = append(b.Columns, "*")
	}

	result, err = b.RunSelect()
	b.logQuery(b.PreparedSql, b.GetBindings(), time.Since(start), result)
	if len(b.EagerLoad) > 0 && result.(ScanResult).Count > 0 {
		rb := RelationBuilder{
			Builder: b,
		}
		rb.EagerLoadRelations(dest)
	}
	return
}

/*
Pluck Get a collection instance containing the values of a given column.
*/
func (b *Builder) Pluck(dest interface{}, params string) (sql.Result, error) {
	return b.Get(dest, params)
}

/*
When Apply the callback if the given "value" is truthy.

	1. When(true,func(builder *Builder))
	2. When(true,func(builder *Builder),func(builder *Builder)) //with default callback

*/
func (b *Builder) When(boolean bool, cb ...func(builder *Builder)) *Builder {
	if boolean {
		cb[0](b)
	} else if len(cb) == 2 {
		//if false and we have default callback
		cb[1](b)
	}
	return b
}
func (b *Builder) Value(dest interface{}, column string) (sql.Result, error) {
	return b.Get(dest, column)
}

/*Reset
reset bindings and components
*/
func (b *Builder) Reset(targets ...string) *Builder {
	for _, componentName := range targets {
		switch componentName {
		case TYPE_ORDER:
			delete(b.Components, TYPE_ORDER)
			delete(b.Bindings, TYPE_ORDER)
			b.Orders = nil
		case TYPE_LIMIT:
			delete(b.Bindings, TYPE_LIMIT)
			delete(b.Components, TYPE_LIMIT)
			b.LimitNum = 0
		case TYPE_OFFSET:
			delete(b.Bindings, TYPE_OFFSET)
			delete(b.Components, TYPE_OFFSET)
			b.OffsetNum = 0
		case TYPE_WHERE:
			delete(b.Bindings, TYPE_WHERE)
			delete(b.Components, TYPE_WHERE)
			b.Wheres = nil
		}
	}
	return b
}
func (b *Builder) EnableLogQuery() *Builder {
	b.LoggingQueries = true
	return b
}

/*BeforeQuery
Register a closure to be invoked before the query is executed.
*/
func (b *Builder) BeforeQuery(func(builder *Builder)) *Builder {

	return b
}

/*ApplyQueryCallbacks
Invoke the "before query" modification callbacks.

*/
func (b *Builder) ApplyQueryCallbacks() {
	for _, callBack := range b.QueryCallBacks {
		callBack(b)
	}
	b.QueryCallBacks = nil
}

/*
Paginate Paginate the given query into a simple paginator.
*/
func (b *Builder) Paginate(p *Paginator, columns ...interface{}) (*Paginator, error) {
	if len(b.Groups) > 0 || len(b.Havings) > 0 {
		panic("having/group pagination not supported")
	}
	cb := CloneWithout(b, TYPE_COLUMN, TYPE_ORDER, TYPE_OFFSET, TYPE_LIMIT)
	cb.Bindings = b.GetRawBindings()
	_, err := cb.Count(&p.Total)
	if err != nil {
		return nil, err
	}
	_, err = b.ForPage(p.CurrentPage, p.PerPage).Get(p.Items, columns...)
	if err != nil {
		return nil, err
	}
	return p, nil
}

/*
SetAggregate Set the aggregate property without running the query.
*/
func (b *Builder) SetAggregate(function string, column ...string) *Builder {
	if len(column) == 0 {
		column = append(column, "*")
	}
	b.Aggregates = append(b.Aggregates, Aggregate{
		AggregateName:   function,
		AggregateColumn: column[0],
	})
	b.Components[TYPE_AGGREGRATE] = struct{}{}
	if len(b.Groups) == 0 {
		b.Reset(TYPE_ORDER)
	}
	return b
}

/*
GetCountForPagination Get the count of the total records for the paginator.
*/
func (b *Builder) GetCountForPagination() (total int) {

	return
}
