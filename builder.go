package goeloquent

import (
	"context"
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
var BindingKeysInOrder = []string{TYPE_SELECT, TYPE_FROM, TYPE_JOIN, TYPE_UPDATE, TYPE_WHERE, TYPE_GROUP_BY, TYPE_HAVING, TYPE_ORDER, TYPE_UNION, TYPE_UNION_ORDER, TYPE_INSERT}

type Paginatot Paginator
type BuilderChainFunc func(builder *Builder) *Builder
type BuilderFunc func(builder *Builder)

// Builder Base query builder
type Builder struct {
	Connection *Connection   //database connection
	Tx         *Transaction  //database transaction , if it is not nil ,use this to execute sql
	Grammar    *MysqlGrammar //convert builder to sql
	//Processor   processors.IProcessor
	PreSql    strings.Builder          //sql string builder
	Bindings  map[string][]interface{} //available options (select,from,join,where,groupBy,having,order,union,unionOrder)
	FromTable interface{}
	//TablePrefix     string
	TableAlias      string
	Wheres          []Where
	Aggregates      []Aggregate
	Columns         []interface{} // The columns that should be returned.
	IsDistinct      bool          // Indicates if the query returns distinct results.
	DistinctColumns []string      // distinct columns.
	Joins           []*JoinBuilder
	IsJoin          bool
	Groups          []interface{}
	Havings         []Having
	Orders          []Order
	LimitNum        int
	OffsetNum       int
	IndexHint       []string //TODO: hint/force/ignore
	//Unions           []Where
	//UnionLimit       int
	//UnionOffset      int
	//UnionOrders      int
	Components  map[string]struct{} //SelectComponents
	LockMode    interface{}
	Pretending  bool
	PreparedSql string //compiled sql string
	//Model                *eloquent.Model
	Dest                 interface{} // scan dest
	OnlyColumns          map[string]interface{}
	ExceptColumns        map[string]interface{}
	UseWrite             bool //TODO
	BeforeQueryCallBacks []func(builder *Builder)
	AfterQueryCallBacks  []func(builder *Builder)
	DataMapping          map[string]interface{} //column type when use map as scan dest
	Context              context.Context
	Debug                bool //debug mode
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
	CONDITION_TYPE_SUB            = "subquery"
	CONDITION_TYPE_EXIST          = "exist"
	CONDITION_TYPE_NOT_EXIST      = "not exist"
	CONDITION_TYPE_ROW_VALUES     = "rowValues"
	CONDITION_TYPE_JSON_CONTAINS  = "jsonContains"
	BOOLEAN_AND                   = "and"
	BOOLEAN_OR                    = "or"
	BOOLEAN_NOT                   = "not"
	CONDITION_JOIN_NOT            = "not"
	JOIN_TYPE_LEFT                = "left"
	JOIN_TYPE_RIGHT               = "right"
	JOIN_TYPE_INNER               = "inner"
	JOIN_TYPE_CROSS               = "cross"
	JOIN_TYPE_LATERAL             = "lateral"
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
	AggregateName   string //sql aggregate function name
	AggregateColumn string //column name
}

type Order struct {
	OrderType string
	Direction string
	Column    interface{} //string or expression
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
	HavingQuery    Builder
}
type Where struct {
	Type         string
	Column       string
	Columns      []string
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

func NewQueryBuilder(c ...*Connection) *Builder {
	b := Builder{
		Components: make(map[string]struct{}),
		Bindings:   make(map[string][]interface{}),
		Context:    context.Background(),
		Grammar:    &MysqlGrammar{},
	}
	b.Grammar.SetBuilder(&b)
	if len(c) > 0 && c[0] != nil {
		b.Connection = c[0]
		b.Grammar.SetTablePrefix(c[0].Config.Prefix)
	}
	return &b
}

/*
Clone Clone the query.
*/
func Clone(original *Builder) *Builder {
	newBuilder := Builder{
		Connection:      original.Connection,
		Tx:              original.Tx,
		PreSql:          strings.Builder{},
		Bindings:        make(map[string][]interface{}, len(original.Bindings)),
		FromTable:       original.FromTable,
		TableAlias:      original.TableAlias,
		Wheres:          make([]Where, len(original.Wheres)),
		Aggregates:      make([]Aggregate, len(original.Aggregates)),
		Columns:         make([]interface{}, len(original.Columns)),
		IsDistinct:      original.IsDistinct,
		DistinctColumns: make([]string, len(original.DistinctColumns)),
		Joins:           make([]*JoinBuilder, len(original.Joins)),
		Groups:          make([]interface{}, len(original.Groups)),
		Havings:         make([]Having, len(original.Havings)),
		Orders:          make([]Order, len(original.Orders)),
		LimitNum:        original.LimitNum,
		OffsetNum:       original.OffsetNum,
		Components:      make(map[string]struct{}, len(original.Components)),
		LockMode:        original.LockMode,
		Pretending:      original.Pretending,
		PreparedSql:     "",
		Dest:            nil,
		OnlyColumns:     make(map[string]interface{}, len(original.OnlyColumns)),
		ExceptColumns:   make(map[string]interface{}, len(original.ExceptColumns)),
		Context:         context.WithValue(original.Context, "parent", original),
		DataMapping:     make(map[string]interface{}),
	}
	for key, _ := range original.Bindings {
		newBuilder.Bindings[key] = make([]interface{}, len(original.Bindings[key]))
		copy(newBuilder.Bindings[key], original.Bindings[key])
	}
	copy(newBuilder.Wheres, original.Wheres)
	copy(newBuilder.Aggregates, original.Aggregates)
	copy(newBuilder.Columns, original.Columns)
	copy(newBuilder.DistinctColumns, original.DistinctColumns)
	copy(newBuilder.Groups, original.Groups)
	copy(newBuilder.Havings, original.Havings)
	copy(newBuilder.Orders, original.Orders)
	for key, _ := range original.Components {
		newBuilder.Components[key] = original.Components[key]
	}
	for key, _ := range original.OnlyColumns {
		newBuilder.OnlyColumns[key] = original.OnlyColumns[key]
	}
	for key, _ := range original.ExceptColumns {
		newBuilder.ExceptColumns[key] = original.ExceptColumns[key]
	}
	//for _, join := range original.Joins {
	//	newBuilder.Joins = append(newBuilder.Joins, join.Clone()) //TODO: add tests
	//}
	for key, _ := range original.DataMapping {
		newBuilder.DataMapping[key] = original.DataMapping[key]
	}
	newBuilder.Grammar = &MysqlGrammar{
		Prefix:  original.Grammar.GetTablePrefix(),
		Builder: &newBuilder,
	}
	return &newBuilder
}
func (b *Builder) WithContext(ctx context.Context) *Builder {
	b.Context = ctx
	return b
}
func (b *Builder) Clone() *Builder {
	return Clone(b)
}

/*
CloneWithout Clone the query without the given components.
*/
func CloneWithout(original *Builder, without ...string) *Builder {
	b := Clone(original)
	b.Reset(without...)
	return b
}

/*
CloneWithout Clone the query without the given components.
*/
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

// Select set the columns to be selected
func (b *Builder) Select(columns ...interface{}) *Builder {
	b.Components[TYPE_COLUMN] = struct{}{}
	if b.Columns == nil {
		b.Columns = []interface{}{}
	}

	for i := 0; i < len(columns); i++ {
		switch columnType := columns[i].(type) {
		case func(builder *Builder):
		case func(builder *Builder) *Builder:
			return b.SelectSub(columns[0], columns[1].(string))
		case string:
			b.Columns = append(b.Columns, columnType)
		case map[string]interface{}:
			for as, q := range columnType {
				switch q.(type) {
				case func(builder *Builder):
					b.SelectSub(q, as)
				case *Builder:
					b.SelectSub(q, as)
				case Expression:
					b.AddSelect(q)
				case string:
					b.Columns = append(b.Columns, q)
				default:
					panic(errors.New("unsupported type for select"))
				}
			}
		case Expression:
			b.AddSelect(columnType)
		case []string:
			cols := columns[i].([]string)
			for _, col := range cols {
				b.Columns = append(b.Columns, col)
			}
		default:
			panic(errors.New("unsupported type for select"))
		}
	}
	return b
}

// SelectSub Add a subselect expression to the query.
func (b *Builder) SelectSub(query interface{}, as string) *Builder {
	qStr, bindings := b.CreateSub(query)
	queryStr := fmt.Sprintf("(%s) as %s", qStr, b.Grammar.Wrap(as))

	return b.SelectRaw(queryStr, bindings)

}

// SelectRaw Add a new "raw" select expression to the query.
func (b *Builder) SelectRaw(expression string, bindings ...[]interface{}) *Builder {
	b.AddSelect(Expression(expression))
	if len(bindings) > 0 {
		b.AddBinding(bindings[0], TYPE_SELECT)
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

 1. FromRaw(goeloquent.Raw(`(select max(last_seen_at) as last_seen_at from "user_sessions") as "sessions""`))

    select * from (select max(last_seen_at) as last_seen_at from "user_sessions") as "sessions"

 2. FromRaw("users as u")

    select * from users as u
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
CreateSub Creates a subquery and parse it.
*/
func (b *Builder) CreateSub(query interface{}) (string, []interface{}) {
	var builder *Builder
	if bT, ok := query.(*Builder); ok {
		builder = bT
	} else if function, ok := query.(func(builder *Builder)); ok {
		builder = b.ForSubQuery()
		function(builder)
	} else if function, ok := query.(func(builder *Builder) *Builder); ok {
		builder = b.ForSubQuery()
		function(builder)
	} else if str, ok := query.(string); ok {
		return b.ParseSub(str)
	} else if str, ok := query.(Expression); ok {
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
	} else if expression, ok := query.(Expression); ok {
		return string(expression), []interface{}{}
	}
	panic("A subquery must be a query builder instance, a Closure, or a string.")
}

/*
PrependDatabaseNameIfCrossDatabaseQuery Prepend the database name if the query is a cross database query.
TODO
*/
func (b *Builder) PrependDatabaseNameIfCrossDatabaseQuery(table string) string {
	return ""
}

/*
AddSelect Add a new select column to the query

 1. []string{"id","name"}

    select id,name

 2. map[string]interface{} {"id":"uid","name":"username"}

    select id as uid,name as username
*/
func (b *Builder) AddSelect(columns ...interface{}) *Builder {
	b.Components[TYPE_COLUMN] = struct{}{}

	for i := 0; i < len(columns); i++ {
		switch columnType := columns[i].(type) {
		case string:
			b.Columns = append(b.Columns, columnType)
		case map[string]interface{}:
			for as, q := range columnType {
				b.SelectSub(q, as)
			}
		case Expression:
			b.Columns = append(b.Columns, columnType)
		}
	}
	return b
}

/*
ToSql Get the SQL representation of the query.
*/
func (b *Builder) ToSql() string {
	b.ApplyBeforeQueryCallbacks()
	if len(b.PreparedSql) > 0 {
		b.PreparedSql = ""
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
	case Builder, *Builder, EloquentBuilder:
		return true
	case BelongsToRelation,
		BelongsToManyRelation,
		HasManyRelation,
		HasManyThrough,
		HasOneThrough,
		HasOneRelation,
		MorphManyRelation,
		MorphByManyRelation,
		MorphOneRelation,
		MorphToManyRelation,
		MorphToRelation:
		return true
	case func(builder *Builder):
		return true
	case func(builder *Builder) *Builder:
		return true
	default:
		return false
	}
}

/*
Table Begin a fluent query against a database table.

1. Table("users") => select * from users

2. Table("users", "u") => select * from users as u
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

 1. From("users") => select * from users

 2. From("users as u") => select * from users as u
*/
func (b *Builder) From(table interface{}, params ...string) *Builder {
	if IsQueryable(table) {
		return b.FromSub(table, params[0])
	}
	b.Components[TYPE_FROM] = struct{}{}
	if len(params) == 1 {
		b.TableAlias = params[0]
		b.FromTable = fmt.Sprintf("%s as %s", table, params[0])
	} else if e, ok := table.(Expression); ok {
		b.FromTable = e
	} else {
		b.FromTable = table.(string)
	}
	return b
}

/*
AddBinding Add a binding to the query.
*/
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

/*
GetBindings Get the current query value bindings in a flattened slice.
*/
func (b *Builder) GetBindings() (res []interface{}) {
	for _, key := range BindingKeysInOrder {
		if bindings, ok := b.Bindings[key]; ok {
			res = append(res, bindings...)
		}
	}
	return
}

/*
GetRawBindings Get the raw map of array of bindings.
*/
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

b:= DB.Query()

 1. b.Select().From("users").Where("id", 1)

    select * from users where id = 1

 2. b.Select().From("users").Where("id", ">", 1)

    select * from users where id > 1

 3. b.Select().From("users").Where(query.Where{
    Type:     query.CONDITION_TYPE_BASIC,
    Column:   "alias",
    Operator: "=",
    Value:    "boss",
    })

    select * from users where alias = 'boss'

 4. b.Select().From("users").Where([]query.Where{
    {
    Type:     query.CONDITION_TYPE_BASIC,
    Column:   "alias",
    Operator: "=",
    Value:    "boss",
    },
    {
    Type:     query.CONDITION_TYPE_BASIC,
    Column:   "age",
    Operator: ">",
    Value:    18,
    Boolean:  query.BOOLEAN_OR,
    },
    })

    select * from users where alias = 'boss' or age > 18

 5. b.Select().From("users").Where([][]interface{}{
    {"alias", "=", "boss"},
    {"age", ">", 18,query.BOOLEAN_OR},
    })

    select * from users where alias = 'boss' or age > 18

 6. b.Select().From("users").Where("role","in",[]string{"admin","boss"})

    select * from users where role in ('admin','boss')

 7. b.Select().From("users").Where("age","between",[]interface{18,60,100})

    select * from users where age between 18 and 60

 8. b.Select().From("users").Where("name","Jack").Where(func(builder *query.Builder){
    builder.Where("age",">",18)
    builder.OrWhere("age","<",60)
    })

    select * from users where  name = 'Jack' and (age > 18 or age < 60)
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
	switch condition := params[0].(type) {
	case func(builder *Builder):
		var boolean string
		if paramsLength > 1 {
			boolean = params[1].(string)
		} else {
			boolean = BOOLEAN_AND
		}
		cb := NewQueryBuilder()
		if b.FromTable != nil {
			cb.From(b.FromTable)
		}
		condition(cb)
		return b.AddNestedWhereQuery(cb, boolean)
	case func(builder *Builder) *Builder:
		var boolean string
		if paramsLength > 1 {
			boolean = params[1].(string)
		} else {
			boolean = BOOLEAN_AND
		}
		cb := condition(NewQueryBuilder())
		return b.AddNestedWhereQuery(cb, boolean)
	case Where:
		b.Wheres = append(b.Wheres, condition)
		b.Components[TYPE_WHERE] = struct{}{}
		return b

	case []Where:
		b.Wheres = append(b.Wheres, condition...)
		b.Components[TYPE_WHERE] = struct{}{}
		return b
	case Expression:
		if paramsLength > 1 {
			boolean = params[1].(string)
		} else {
			boolean = BOOLEAN_AND
		}
		b.Wheres = append(b.Wheres, Where{
			Type:    CONDITION_TYPE_RAW,
			RawSql:  condition,
			Boolean: boolean,
		})
		b.Components[TYPE_WHERE] = struct{}{}
		return b
	case map[string]interface{}:
		boolean = BOOLEAN_AND
		if paramsLength > 1 {
			boolean = params[1].(string)
		}
		cb := NewQueryBuilder()
		for k, v := range condition {
			cb.Where(k, v)
		}
		return b.AddNestedWhereQuery(cb, boolean)
	}
	switch paramsLength {
	case 1:
		panic("where clause must have at least 2 arguments or type of func(builder *Builder)/func(builder *Builder) *Builder")
	case 2:
		//assume operator is "=" and omitted
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
	switch f := value.(type) {

	case func(builder *Builder):
	case func(builder *Builder) *Builder:
		return b.WhereSub(column, operator, f, boolean)
	}
	if f, ok := value.(func(builder *Builder)); ok {
		return b.WhereSub(column, operator, f, boolean)
	}
	if reflect.TypeOf(value).Kind() == reflect.Slice {
		value = reflect.ValueOf(value).Index(0).Elem().Interface()
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
WhereNot Add a "where not" clause to the query.
*/
func (b *Builder) WhereNot(column string, params ...interface{}) *Builder {
	switch len(params) {
	case 1:
		return b.Where(column, "=", params[0], "and not")
	case 2:
		return b.Where(column, params[0].(string), params[1], "and not")
	default:
		panic("wrong arguements in where not,maybe try WhereRaw")
	}
}

/*
OrWhereNot Add a "or where not" clause to the query.

 1. Where("age",1).OrWhereNot("id",1)

    select * from users where age = 1 or not id = 1

 2. Where("age",1).OrWhereNot("id","<>",1)

    select * from users where age = 1 or not id <> 1
*/
func (b *Builder) OrWhereNot(params ...interface{}) *Builder {
	switch len(params) {
	case 1:
		return b.Where(params[0], BOOLEAN_OR)
	case 2:
		return b.Where(params[0], "=", params[1], BOOLEAN_OR+" "+BOOLEAN_NOT)
	case 3:
		return b.Where(params[0], params[1].(string), params[2], BOOLEAN_OR+" "+BOOLEAN_NOT)
	default:
		panic("wrong arguements in or where not,maybe try WhereRaw")
	}
}

/*
WhereColumn Add a "where" clause comparing two columns to the query.

1. WhereColumn("first_name","=","last_name")

	select * from users where first_name = last_name

2. WhereColumn("updated_at",">","created_at")

		select * from users where updated_at > created_at

	 3. WhereColumn([][]interface{}{
	    {"first_name","=","last_name"},
	    {"updated_at",">","created_at"},
	    })

	    select * from users where (first_name = last_name or updated_at > created_at)

4. //TODO joinlateral
select * from `users` inner join lateral (select * from `contacts` where `contracts`.`user_id` = `users`.`id`) as `sub` on true
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
	var ts = make([]string, 3)
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

1. WhereRaw("id = ? or email = ?",[]interface{}{1,"gmail"})

	select * from users where id = 1 or email = 'gmail'

2. Where("age",1).WhereRaw("id = ? or email = ?",[]interface{}{1,"gmail"},"or")

	select * from users where age = 1 or (id = 1 or email = 'gmail')
*/
func (b *Builder) WhereRaw(rawSql string, params ...interface{}) *Builder {
	var boolean string
	var bindings []interface{}
	switch len(params) {
	case 0:
		boolean = BOOLEAN_AND
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

1. WhereIn("id",[]interface{}{1,2,3})

	select * from users where id in (1,2,3)

2. WhereIn("id",NewQueryBuilder().Select("id").From("users").Where("age",">",18))

		select * from users where id in (select id from users where age > 18)

	 3. WhereIn("id",func(builder *query.Builder){
	    builder.Select("id").From("users").Where("age",">",18)
	    })

	    select * from users where id in (select id from users where age > 18)
*/
func (b *Builder) WhereIn(params ...interface{}) *Builder {
	var boolean string
	not := false
	switch len(params) {
	case 0, 1:
		panic("wrong arguements in where in")
	case 2:
		boolean = BOOLEAN_AND

	case 3:
		boolean = params[2].(string)

	case 4:
		boolean = params[2].(string)
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
*/
func (b *Builder) OrWhereIn(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_OR, false)
	return b.WhereIn(params...)
}

/*
WhereNotIn Add a "where not in" clause to the query.
*/
func (b *Builder) WhereNotIn(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_AND, true)
	return b.WhereIn(params...)
}

/*
OrWhereNotIn Add an "or where not in" clause to the query.
*/
func (b *Builder) OrWhereNotIn(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_OR, true)
	return b.WhereIn(params...)
}

/*
WhereNull Add a "where null" clause to the query.
1. WhereNull("name")

	select * from users where name is null

2. WhereNull([]string{"name","age"})

	select * from users where name is null and age is null

3. Where("age",18).WhereNull("name","or")

	select * from users where age = 18 or name is null

4. Where("age",18).WhereNull("name","or",true)

	select * from users where age = 18 or name is not null
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
	switch columnTemp := column.(type) {
	case string:
		b.Wheres = append(b.Wheres, Where{
			Type:    CONDITION_TYPE_NULL,
			Column:  columnTemp,
			Boolean: boolean,
			Not:     not,
		})
	case []interface{}:
		for _, i := range columnTemp {
			b.Wheres = append(b.Wheres, Where{
				Type:    CONDITION_TYPE_NULL,
				Column:  i.(string),
				Boolean: boolean,
				Not:     not,
			})
		}
	case []string:
		for _, i := range columnTemp {
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
WhereNotNull Add a "where not null" clause to the query.
*/
func (b *Builder) WhereNotNull(column interface{}, params ...interface{}) *Builder {
	params = append(params, BOOLEAN_AND, true)
	return b.WhereNull(column, params...)

}

/*
WhereBetween Add a where between statement to the query.

1. WhereBetween("age",[]interface{18,60})

	select * from users where age between 18 and 60

2. Where("name","Jim").WhereBetween("age",[]interface{18,60},"or")

	select * from users where name = 'Jim' or age between 18 and 60

3. Where("name","Jim").WhereBetween("age",[]interface{18,goeloquent.Raw("30")},BOOLEAN_OR,true)

	select * from users where name = 'Jim' or age not between 18 and 30
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

/*
WhereBetweenColumns Add a where between statement using columns to the query.

1. WhereBetweenColumns("id",[]interface{}{"created_at","updated_at"})

	select * from users where id between created_at and updated_at
*/
func (b *Builder) WhereBetweenColumns(column string, values []interface{}, params ...interface{}) *Builder {
	paramsLength := len(params)
	var boolean = BOOLEAN_AND
	var betweenType = CONDITION_TYPE_BETWEEN_COLUMN
	not := false
	switch paramsLength {
	case 1:
		boolean = params[0].(string)
	case 2:
		boolean = params[0].(string)
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

/*
OrWhereBetween Add an or where between statement to the query.
*/
func (b *Builder) OrWhereBetween(params ...interface{}) *Builder {
	paramsLength := len(params)
	var boolean = BOOLEAN_OR
	not := false

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

/*
OrWhereBetweenColumns Add an or where between statement using columns to the query.

1. OrWhereBetweenColumns("id",[]interface{}{"created_at","updated_at"})

	select * from users where id between created_at and updated_at

2. Where("id",1).OrWhereBetweenColumns("id",[]interface{}{"created_at","updated_at"},BOOLEAN_OR)

	select * from users where id = 1 or id between created_at and updated_at

3. Where("id",1).OrWhereBetweenColumns("id",[]interface{}{"created_at","updated_at"},BOOLEAN_OR,true)

	select * from users where id = 1 or id not between created_at and updated_at
*/
func (b *Builder) OrWhereBetweenColumns(column string, values []interface{}, params ...interface{}) *Builder {
	paramsLength := len(params)
	var boolean = BOOLEAN_OR
	not := false
	switch paramsLength {
	case 0:
	case 1:
		boolean = params[0].(string)
	case 2:
		boolean = params[0].(string)
		not = params[1].(bool)
	default:
		panic("wrong arguements in or where between columns")
	}

	b.Components[TYPE_WHERE] = struct{}{}

	b.Wheres = append(b.Wheres, Where{
		Type:    CONDITION_TYPE_BETWEEN_COLUMN,
		Column:  column,
		Boolean: boolean,
		Values:  values,
		Not:     not,
	})
	return b
}

/*
WhereNotBetween Add a where not between statement to the query.
1. Select().From("users").WhereNotBetween("age", []interface{}{18, 30, 300})

	select * from users where age not between 18 and 30
*/
func (b *Builder) WhereNotBetween(column string, params ...interface{}) *Builder {
	paramsLength := len(params)
	var boolean = BOOLEAN_AND
	not := true
	if paramsLength > 1 {
		boolean = params[1].(string)

	}

	var betweenType = CONDITION_TYPE_BETWEEN

	b.Components[TYPE_WHERE] = struct{}{}
	tvalues := params[0].([]interface{})[0:2]
	for _, tvalue := range tvalues {
		if _, ok := tvalue.(Expression); !ok {
			b.AddBinding([]interface{}{tvalue}, TYPE_WHERE)
		}
	}

	b.Wheres = append(b.Wheres, Where{
		Type:    betweenType,
		Column:  column,
		Boolean: boolean,
		Values:  params[0].([]interface{})[0:2],
		Not:     not,
	})
	return b
}

/*
WhereNotBetweenColumns Add a where not between statement using columns to the query.

1. Where("id",1).WhereNotBetweenColumns("id",[]interface{}{"created_at","updated_at"})

	select * from users where id = 1 and id not between created_at and updated_at

2. Where("id",1).WhereNotBetweenColumns("id",[]interface{}{"created_at","updated_at"},BOOLEAN_OR)

	select * from users where id = 1 or id not between created_at and updated_at
*/
func (b *Builder) WhereNotBetweenColumns(column string, values []interface{}, params ...interface{}) *Builder {

	return b.WhereBetweenColumns(column, values, params...)
}

/*
OrWhereNotBetween Add an or where not between statement to the query.
*/
func (b *Builder) OrWhereNotBetween(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_OR, true)

	return b.WhereBetween(params...)
}

/*
OrWhereNotBetweenColumns Add an or where not between statement using columns to the query.

Where("id",1).OrWhereNotBetweenColumns("id",[]interface{}{"created_at","updated_at"})
*/
func (b *Builder) OrWhereNotBetweenColumns(column string, values []interface{}) *Builder {

	return b.WhereBetweenColumns(column, values, BOOLEAN_OR, true)
}

/*
OrWhereNotNull Add an "or where not null" clause to the query.
*/
func (b *Builder) OrWhereNotNull(column interface{}) *Builder {
	params := []interface{}{BOOLEAN_OR, true}
	return b.WhereNull(column, params...)
}

/*
WhereDate Add a "where date" statement to the query.

 1. Select().From("users").WhereDate("created_at",time.Now())

    select * from users where date(created_at) = '2022-01-01'

 2. Select().From("users").WhereDate("created_at",">",time.Now())

    select * from users where date(created_at) > '2022-01-01'

 3. Select().From("users").Where("name","Jackie").WhereDate("created_at",">",time.Now(),"or")

    select * from users where name = 'Jackie' or date(created_at) > '2022-01-01'
*/
func (b *Builder) WhereDate(params ...interface{}) *Builder {
	p := append([]interface{}{CONDITION_TYPE_DATE}, params...)
	return b.AddTimeBasedWhere(p...)
}

//TODO: OrWhereDate

/*
WhereTime Add a "where time" statement to the query.

 1. Select().From("users").WhereTime("created_at",time.Now())

    select * from users where time(created_at) = '12:00:00'

 2. Select().From("users").WhereTime("created_at",">",time.Now())

    select * from users where time(created_at) > '12:00:00'

 3. Select().From("users").Where("name","Jackie").WhereTime("created_at",">",time.Now(),"or")

    select * from users where name = 'Jackie' or time(created_at) > '12:00:00'
*/
func (b *Builder) WhereTime(params ...interface{}) *Builder {
	p := append([]interface{}{CONDITION_TYPE_TIME}, params...)
	return b.AddTimeBasedWhere(p...)
}

//TODO: OrWhereTime

/*
WhereDay Add a "where day" statement to the query.

 1. Select().From("users").WhereDay("created_at",1)

    select * from users where day(created_at) = 1

 2. Select().From("users").WhereDay("created_at",">",1)

    select * from users where day(created_at) > 1

 3. Select().From("users").Where("name","Jackie").WhereDay("created_at",">",1,"or")

    select * from users where name = 'Jackie' or day(created_at) > 1
*/
func (b *Builder) WhereDay(params ...interface{}) *Builder {
	p := append([]interface{}{CONDITION_TYPE_DAY}, params...)
	return b.AddTimeBasedWhere(p...)
}

//TODO: OrWhereDay

/*
WhereMonth Add a "where month" statement to the query.

1. Select().From("users").WhereMonth("created_at", 1)

	select * from users where month(created_at) = 1

2. Select().From("users").WhereMonth("created",">",1)

	select * from users where month(created) > 1

3. Select().From("users").Where("name","Jackie").WhereMonth("created",">",1,"or")

	select * from users where name = 'Jackie' or month(created) > 1
*/
func (b *Builder) WhereMonth(params ...interface{}) *Builder {
	p := append([]interface{}{CONDITION_TYPE_MONTH}, params...)
	return b.AddTimeBasedWhere(p...)
}

//TODO: OrWhereMonth

/*
WhereYear Add a "where year" statement to the query.
1. Select().From("users").WhereYear("created_at", 2022)

	select * from users where year(created_at) = 2022

2. Select().From("users").WhereYear("created",">",2022)

	select * from users where year(created) > 2022

3. Select().From("users").Where("name","Jackie").WhereYear("created",">",2022,"or")

	select * from users where name = 'Jackie' or year(created) > 2022
*/
func (b *Builder) WhereYear(params ...interface{}) *Builder {
	p := append([]interface{}{CONDITION_TYPE_YEAR}, params...)
	return b.AddTimeBasedWhere(p...)
}

//TODO: OrWhereYear

/*
AddTimeBasedWhere Add a time based (year, month, day, time) statement to the query.
*/
func (b *Builder) AddTimeBasedWhere(params ...interface{}) *Builder {
	var timeType = params[0]
	var boolean = BOOLEAN_AND
	var operator string
	var value interface{}
	var tvalue interface{}
	switch len(params) {

	case 0, 1, 2:
		panic("wrong arguements in time based where")
	case 3:
		operator = "="
		tvalue = params[2]
	case 4:
		operator = params[2].(string)
		tvalue = params[3]
	case 5:
		operator = params[2].(string)
		tvalue = params[3]
		boolean = params[4].(string)
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

/*
WhereNested Add a nested where statement to the query.

 1. WhereNested([][]interface{}{{"age",">",18},{ "age","<",60,"or"}})

    select * from users where (age > 18 or age < 60)

 2. WhereNested(func(builder *query.Builder){
    builder.Where("age",">",18)
    builder.OrWhere("age","<",60)
    })

    select * from users where (age > 18 or age < 60)
*/
func (b *Builder) WhereNested(params ...interface{}) *Builder {
	paramsLength := len(params)
	if paramsLength == 1 {
		params = append(params, BOOLEAN_AND)
	}
	cb := b.ForNestedWhere()
	switch converted := params[0].(type) {
	case Where:
		cb.Wheres = append(cb.Wheres, params[0].(Where))
	case []Where:
		cb.Wheres = append(cb.Wheres, params[0].([]Where)...)
	case [][]interface{}:
		tp := params[0].([][]interface{})
		for i := 0; i < len(tp); i++ {
			cb.Where(tp[i]...)
		}
	case []interface{}:
		cb.Where(params[0].([]interface{}))
	case func(builder *Builder) *Builder:
		var boolean string
		if paramsLength > 1 {
			boolean = params[1].(string)
		} else {
			boolean = BOOLEAN_AND
		}
		cb = converted(cb)
		return b.AddNestedWhereQuery(cb, boolean)
	case func(builder *Builder):
		var boolean string
		if paramsLength > 1 {
			boolean = params[1].(string)
		} else {
			boolean = BOOLEAN_AND
		}
		converted(cb)
		return b.AddNestedWhereQuery(cb, boolean)
	}
	b.Wheres = append(b.Wheres, Where{
		Type:    CONDITION_TYPE_NESTED,
		Boolean: params[1].(string),
		Value:   cb,
	})
	b.Components[TYPE_WHERE] = struct{}{}
	b.AddBinding(cb.GetBindings(), TYPE_WHERE)
	return b
}

func (b *Builder) ForNestedWhere() *Builder {
	return NewQueryBuilder(b.Connection).From(b.FromTable)
}

/*
ForSubQuery Create a new query instance for a sub-query.
*/
func (b *Builder) ForSubQuery() *Builder {
	cb := NewQueryBuilder(b.Connection)
	cb.Grammar.Prefix = b.Grammar.Prefix
	return cb
}

/*
WhereSub Add a full sub-select to the query.
*/
func (b *Builder) WhereSub(column string, operator string, value interface{}, boolean string) *Builder {
	cb := b.ForSubQuery()
	switch f := value.(type) {
	case func(builder *Builder):
		f(cb)
	case func(builder *Builder) *Builder:
		cb = f(cb)
	default:
		panic("sub query must be [func(builder *Builder)] or [func(builder *Builder) *Builder]")
	}
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

/*
WhereExists Add an exists clause to the query.

 1. WhereExists(func(builder *Builder){
    builder.Select().From("users").Where("age",">",18)
    })

    select * from users where exists (select * from users where age > 18)

 2. Where("name","Pam").WhereExists(func(builder *Builder){
    builder.Select().From("users").Where("age",">",18)
    },"or")

    select * from users where name = 'Pam' or exists (select * from users where age > 18)

 3. Where("name","Pam").WhereExists(func(builder *Builder){
    builder.Select().From("users").Where("age",">",18)
    },"or",true)

    select * from users where name = 'Pam' or not exists (select * from users where age > 18)
*/
func (b *Builder) WhereExists(cb func(builder *Builder), params ...interface{}) *Builder {
	newBuilder := b.ForSubQuery()
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

/*
AddWhereExistsQuery  Add an exists clause to the query.
*/
func (b *Builder) AddWhereExistsQuery(builder *Builder, boolean string, not bool) *Builder {

	b.Wheres = append(b.Wheres, Where{
		Type:    CONDITION_TYPE_EXIST,
		Query:   builder,
		Boolean: boolean,
		Not:     not,
	})
	b.Components[TYPE_WHERE] = struct{}{}
	b.AddBinding(builder.GetBindings(), TYPE_WHERE)
	return b
}

/*
	TODO

whereJsonContains Add a "where json contains" clause to the query.
*/
func (b *Builder) WhereJsonContains(column string, value interface{}, params ...interface{}) *Builder {
	var boolean = BOOLEAN_AND
	var not = false
	switch len(params) {
	case 0:
		boolean = BOOLEAN_AND
	case 1:
		boolean = params[0].(string)
		not = false
	case 2:
		boolean = params[0].(string)
		not = params[1].(bool)
	}
	b.Wheres = append(b.Wheres, Where{
		Type:    CONDITION_TYPE_JSON_CONTAINS,
		Column:  column,
		Value:   value,
		Boolean: boolean,
		Not:     not,
	})
	b.Components[TYPE_WHERE] = struct{}{}
	return b

}

/*
	TODO

WhereJsonOverlaps Add a "where json overlaps" clause to the query.
*/
func (b *Builder) WhereJsonOverlaps(column string, value interface{}, params ...interface{}) *Builder {
	return b
}

/*
TODO
WhereJsonContainsKey Add a "where json contains key" clause to the query.
*/
func (b *Builder) WhereJsonContainsKey(column string, value interface{}, params ...interface{}) *Builder {
	return b
}

/*
TODO
WhereJsonLength Add a "where json length" clause to the query.
*/
func (b *Builder) WhereJsonLength(column string, operator string, value interface{}, params ...interface{}) *Builder {
	return b
}

/*
GroupBy Add a "group by" clause to the query.

 1. GroupBy("name","email")

    select * from users group by name,email
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

 1. Select().From("users").GroupByRaw("DATE(created_at), ? DESC", []interface{}{"foo"})

    select * from `users` group by DATE(created_at), foo DESC
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

 1. GetBuilder().From("users").Select().Having("age", ">", 1)

    select * from `users` having `age` > 1

 2. GetBuilder().From("users").Select().GroupBy("age").Having("age", ">", 1)

    select * from `users` group by `age` having `age` > 1

 3. GetBuilder().Select().From("users").Having(func(builder *goeloquent.Builder) *goeloquent.Builder {
    return builder.Having("email", "foo").OrHaving("email", "bar")
    }).OrHaving("email", "baz")

    select * from `users` having (`email` = 'foo' or `email` = 'bar') or `email` = 'baz'
*/
func (b *Builder) Having(params ...interface{}) *Builder {
	havingBoolean := BOOLEAN_AND
	havingOperator := "="
	var havingValue interface{}
	var havingColumn string
	length := len(params)
	switch length {
	case 1:
		if IsQueryable(params[0]) {
			return b.HavingNested(params...)
		}
	case 2:
		if IsQueryable(params[0]) {
			return b.HavingNested(params...)
		}
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

 1. GetBuilder().Select().From("users").HavingRaw("user_foo < user_bar")

    select * from `users` having user_foo < user_bar
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
OrHaving Add an "or having" clause to the query.
*/
func (b *Builder) OrHaving(params ...interface{}) *Builder {
	switch len(params) {
	case 2:
		return b.Having(params[0], "=", params[1], BOOLEAN_OR)
	case 3:
		return b.Having(params[0], params[1].(string), params[2], BOOLEAN_OR)
	default:
		panic("wrong arguements in or having")
	}
}

/*
HavingNested Add a nested having statement to the query.
*/
func (b *Builder) HavingNested(params ...interface{}) *Builder {
	cb := params[0].(func(builder *Builder) *Builder)
	q := b.ForNestedWhere()
	cb(q)
	and := BOOLEAN_AND
	if len(params) > 1 {
		and = params[1].(string)
	}
	return b.AddNestedHavingQuery(*q, and)
}
func (b *Builder) AddNestedHavingQuery(builder Builder, boolean string) *Builder {
	b.Havings = append(b.Havings, Having{
		HavingType:    CONDITION_TYPE_NESTED,
		HavingBoolean: boolean,
		HavingQuery:   builder,
	})
	b.Components[TYPE_HAVING] = struct{}{}
	b.AddBinding(builder.GetRawBindings()[TYPE_HAVING], TYPE_HAVING)
	return b
}

/*
HavingNull Add a "having null" clause to the query.

 1. HavingNull("name")

    select * from users having name is null

 2. HavingNull("name","or")

    select * from users having name is null or

 3. HavingNull("name","or",true)

    select * from users having name is not null or
*/
func (b *Builder) HavingNull(column string, params ...interface{}) *Builder {
	var boolean string
	not := false
	length := len(params)
	switch length {
	case 0:
	case 1:
		boolean = params[0].(string)
	case 2:
		boolean = params[0].(string)
		not = params[1].(bool)
	}
	having := Having{
		HavingType:    CONDITION_TYPE_NULL,
		HavingColumn:  column,
		HavingBoolean: boolean,
		Not:           not,
	}
	b.Components[TYPE_HAVING] = struct{}{}
	b.Havings = append(b.Havings, having)
	return b

}

/*
OrHavingNull Add an "or having null" clause to the query.

1. OrHavingNull("name")

	select * from users having name is null

2. OrHavingNull("name",true)

	select * from users having name is not null
*/
func (b *Builder) OrHavingNull(column string, params ...interface{}) *Builder {
	switch len(params) {
	case 0:
		return b.HavingNull(column, BOOLEAN_OR, false)
	case 1:
		return b.HavingNull(column, BOOLEAN_OR, params[0].(bool))
	default:
		return b.HavingNull(column, BOOLEAN_OR, false)
	}
}

/*
HavingNotNull Add a "having not null" clause to the query.

1. HavingNotNull("name")

	select * from users having name is not null

2. HavingNotNull("name","or")

	select * from users having name is not null or
*/
func (b *Builder) HavingNotNull(column string, params ...interface{}) *Builder {

	switch len(params) {
	case 0:
		return b.HavingNull(column, BOOLEAN_AND, true)
	case 1:
		return b.HavingNull(column, params[0].(string), true)
	default:
		return b.HavingNull(column, BOOLEAN_AND, true)

	}
}

/*
OrHavingNotNull Add an "or having not null" clause to the query.

1. OrHavingNotNull("name")

	select * from users having name is not null
*/
func (b *Builder) OrHavingNotNull(column string, params ...interface{}) *Builder {
	params = append(params, BOOLEAN_OR, true)
	return b.HavingNull(column, params...)
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

 1. OrderBy("name")

    select * from users order by name asc

 2. OrderBy("name","desc")

    select * from users order by name desc
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
	column := params[0]
	if IsQueryable(params[0]) {
		str, bindings := b.CreateSub(params[0])
		b.AddBinding(bindings, TYPE_ORDER)
		column = Raw("(" + str + ")")
	}
	if len(params) > 1 {
		order = params[1].(string)
	}
	if order != ORDER_ASC && order != ORDER_DESC {
		panic(errors.New("wrong order direction: " + order))
	}
	b.Orders = append(b.Orders, Order{
		Direction: order,
		Column:    column,
	})
	b.Components[TYPE_ORDER] = struct{}{}

	return b
}

/*
OrderByDesc Add a descending "order by" clause to the query.

 1. OrderByDesc("name")

    select * from users order by name desc
*/
func (b *Builder) OrderByDesc(column string) *Builder {
	return b.OrderBy(column, ORDER_DESC)
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

func (b *Builder) WhereMap(params map[string]interface{}) *Builder {
	for key, param := range params {
		b.Where(key, "=", param)
	}
	return b
}

/*
AddNestedWhereQuery Add another query builder as a nested where to the query builder.
*/
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

/*
Find Execute a query for a single record by ID.
*/
func (b *Builder) Find(dest interface{}, params ...interface{}) (result Result, err error) {
	if reflect.Indirect(reflect.ValueOf(dest)).Kind() != reflect.Slice {
		b.Limit(1)
	}
	b.ApplyBeforeQueryCallbacks()
	if len(params) > 1 {
		b.WhereIn("id", params)
	} else {
		b.Where("id", params)
	}
	return b.Get(dest)
}

/*
First Execute the query and get the first result.
*/
func (b *Builder) First(dest interface{}, columns ...interface{}) (result Result, err error) {
	b.Limit(1)
	return b.Get(dest, columns...)
}

/*
RunSelect run the query as a "select" statement against the connection.
*/
func (b *Builder) RunSelect() (result Result, err error) {
	result, err = b.Run(b.ToSql(), b.GetBindings(), func() (result Result, err error) {
		if b.Pretending {
			return Result{
				Sql:      b.PreparedSql,
				Bindings: b.GetBindings(),
				Count:    0,
				Error:    nil,
				Time:     0,
				Raw:      nil,
			}, nil
		}
		if b.Tx != nil {
			result, err = b.Tx.Select(b.PreparedSql, b.GetBindings(), b.Dest, b.DataMapping)
		} else {
			result, err = b.GetConnection().Select(b.PreparedSql, b.GetBindings(), b.Dest, b.DataMapping)
		}
		return
	})

	return
}

func (b *Builder) GetConnection() *Connection {
	if b.Connection != nil {
		return b.Connection
	}
	return DB.Connection(DefaultConnectionName)
}
func (b *Builder) Run(query string, bindings []interface{}, callback func() (result Result, err error)) (result Result, err error) {
	defer func() {
		catchedErr := recover()
		if catchedErr != nil {
			switch catchedErr.(type) {
			case string:
				err = errors.New(catchedErr.(string))
			case error:
				err = catchedErr.(error)
			default:
				err = errors.New("unknown panic")
			}
		}
	}()
	start := time.Now()
	result, err = callback()
	result.Bindings = bindings
	result.Time = time.Since(start)
	result.Sql = query
	return

}

/*
Exists Determine if any rows exist for the current query.
*/
func (b *Builder) Exists() (exists bool, err error) {

	b.ApplyBeforeQueryCallbacks()
	var count int
	_, err = b.Run(b.Grammar.CompileExists(), b.GetBindings(), func() (result Result, err error) {
		result, err = b.GetConnection().Select(b.PreparedSql, b.GetBindings(), &count, nil)
		return
	})
	if err != nil {
		return false, err
	}
	b.ApplyAfterQueryCallbacks()
	if count > 0 {
		return true, nil
	}
	return false, nil
}

/*
DoesntExist Determine if no rows exist for the current query.
*/
func (b *Builder) DoesntExist() (notExists bool, err error) {
	e, err := b.Exists()
	if err != nil {
		return false, err
	}
	return !e, nil
}

/*
Aggregate Execute an aggregate function on the database.
*/
func (b *Builder) Aggregate(dest interface{}, fn string, column ...string) (result Result, err error) {
	b.Dest = dest
	if column == nil {
		column = append(column, "*")
	}
	b.Aggregates = append(b.Aggregates, Aggregate{
		AggregateName:   fn,
		AggregateColumn: column[0],
	})
	b.Components[TYPE_AGGREGRATE] = struct{}{}
	return b.RunSelect()
}

/*
Count Retrieve the "count" result of the query.
*/
func (b *Builder) Count(dest interface{}, column ...string) (result Result, err error) {
	return b.Aggregate(dest, "count", column...)
}

/*
Min Retrieve the minimum value of a given column.
*/
func (b *Builder) Min(dest interface{}, column ...string) (result Result, err error) {

	return b.Aggregate(dest, "min", column...)

}

/*
Max Retrieve the maximum value of a given column.
*/
func (b *Builder) Max(dest interface{}, column ...string) (result Result, err error) {

	return b.Aggregate(dest, "max", column...)

}

/*
Avg Alias for the "avg" method.
*/
func (b *Builder) Avg(dest interface{}, column ...string) (result Result, err error) {

	return b.Aggregate(dest, "avg", column...)

}
func (b *Builder) Sum(dest interface{}, column ...string) (result Result, err error) {

	return b.Aggregate(dest, "sum", column...)

}

/*
ForPage Set the limit and offset for a given page.
*/
func (b *Builder) ForPage(page, perPage int64) *Builder {

	return b.Offset(int((page - 1) * perPage)).Limit(int(perPage))
}

/*
ForPageBeforeId Constrain the query to the previous "page" of results before a given ID.
*/
func (b *Builder) ForPageBeforeId(perpage, id int, column string) *Builder {

	return b.Where(column, "<", id).OrderBy(column, "desc").Limit(perpage)
}
func (b *Builder) ForPageAfterId(perpage, id int, column string) *Builder {

	return b.Where(column, ">", id).OrderBy(column, "asc").Limit(perpage)
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
values can be []map[string]interface{},map[string]interface{},struct,pointer of struct,pointer of slice of struct
*/
func (b *Builder) Insert(values interface{}) (result Result, err error) {
	items := PrepareInsertValues(values)
	b.Prepare(values)
	b.ApplyBeforeQueryCallbacks()
	result, err = b.Run(b.Grammar.CompileInsert(items), b.GetBindings(), func() (result Result, err error) {
		if b.Pretending {
			return Result{
				Sql:      b.PreparedSql,
				Bindings: b.GetBindings(),
				Count:    0,
				Error:    nil,
				Time:     0,
				Raw:      nil,
			}, nil
		}
		if b.Tx != nil {
			result, err = b.Tx.Insert(b.PreparedSql, b.GetBindings())
		} else {
			result, err = b.GetConnection().Insert(b.PreparedSql, b.GetBindings())
		}
		return
	})
	b.ApplyAfterQueryCallbacks()
	return
}

/*
InsertGetId Insert a new record and get the value of the primary key.
*/
func (b *Builder) InsertGetId(values interface{}) (int64, error) {
	b.ApplyBeforeQueryCallbacks()
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
func (b *Builder) InsertOrIgnore(values interface{}) (result Result, err error) {
	items := PrepareInsertValues(values)

	b.ApplyBeforeQueryCallbacks()
	result, err = b.Run(b.Grammar.CompileInsertOrIgnore(items), b.GetBindings(), func() (result Result, err error) {
		if b.Pretending {
			return Result{
				Sql:      b.PreparedSql,
				Bindings: b.GetBindings(),
				Count:    0,
				Error:    nil,
				Time:     0,
				Raw:      nil,
			}, nil
		}
		if b.Tx != nil {
			result, err = b.Tx.Insert(b.PreparedSql, b.GetBindings())
		} else {
			result, err = b.GetConnection().Insert(b.PreparedSql, b.GetBindings())
		}
		return
	})
	b.ApplyAfterQueryCallbacks()
	return
}

/*
Update Update records in the database.

 1. Table("users").Where("id",1).Update(map[string]interface{}{"name":"Jackie","age":18})

    update users set name = 'Jackie', age = 18 where id = 1
*/
func (b *Builder) Update(v map[string]interface{}) (result Result, err error) {
	b.ApplyBeforeQueryCallbacks()
	result, err = b.Run(b.Grammar.CompileUpdate(v), b.GetBindings(), func() (result Result, err error) {
		if b.Pretending {
			return Result{
				Sql:      b.PreparedSql,
				Bindings: b.GetBindings(),
				Count:    0,
				Error:    nil,
				Time:     0,
				Raw:      nil,
			}, nil
		}
		if b.Tx != nil {
			result, err = b.Tx.Update(b.PreparedSql, b.GetBindings())
		} else {
			result, err = b.GetConnection().Update(b.PreparedSql, b.GetBindings())
		}
		return
	})
	b.ApplyAfterQueryCallbacks()
	return
}

/*
UpdateOrInsert insert or update a record matching the attributes, and fill it with values.
*/
func (b *Builder) UpdateOrInsert(conditions map[string]interface{}, values map[string]interface{}) (updated bool, err error) {
	exist, err := b.Where(conditions).Exists()
	if err != nil {
		return
	}
	if !exist {
		for k, v := range values {
			conditions[k] = v
		}
		b.Reset(TYPE_WHERE)
		_, err = b.Insert(conditions)
		if err != nil {
			return
		}
		return false, nil
	} else {
		_, err = b.Limit(1).Update(values)
		if err != nil {
			return
		}
		return true, nil
	}
}

/*
Model Convert a base builder to an Eloquent builder.
*/
func (b *Builder) Model(model ...interface{}) *EloquentBuilder {

	eb := ToEloquentBuilder(b)
	if len(model) > 0 {
		eb.SetModel(model[0])
	}
	return eb
}

/*
Upsert Insert new records or update the existing one
*/
//func (b *Builder) Upsert(values interface{}, uniqueColumns []string, updateColumns interface{}) (result Result, err error) {
//	items := PrepareInsertValues(values)
//	if b.FromTable == nil {
//		b.Model = eloquent.GetParsedModel(values)
//		b.From(b.Model.Table)
//		b.Connection = goeloquent.DB.Connection(b.Model.ConnectionName)
//		b.Grammar.SetTablePrefix(goeloquent.DB.Configs[b.Model.ConnectionName].Prefix)
//	}
//	b.ApplyAfterQueryCallbacks()
//	b.PreparedSql = b.Grammar.CompileUpsert(items, uniqueColumns, updateColumns)
//	if b.Pretending {
//		return
//	}
//	if b.Tx != nil {
//		return b.Tx.AffectingStatement(b.PreparedSql, b.GetBindings())
//	} else {
//		return b.Connection.AffectingStatement(b.PreparedSql, b.GetBindings())
//	}
//}

/*
Decrement Decrement a column's value by a given amount.
*/
func (b *Builder) Decrement(column string, amount int, extra ...map[string]interface{}) (result Result, err error) {

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
func (b *Builder) Increment(column string, amount int, extra ...map[string]interface{}) (result Result, err error) {

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

/*
InRandomOrder Put the query's results in random order.
*/
func (b *Builder) InRandomOrder(seed ...int) *Builder {
	return b.OrderByRaw(b.Grammar.CompileRandom(seed...), []interface{}{})
}

/*
Delete Delete records from the database.
*/
func (b *Builder) Delete(id ...interface{}) (result Result, err error) {
	if len(id) > 0 {
		b.Where("id", id[0])
	}
	b.ApplyBeforeQueryCallbacks()
	result, err = b.Run(b.Grammar.CompileDelete(), b.GetBindings(), func() (result Result, err error) {
		if b.Pretending {
			return Result{
				Sql:      b.PreparedSql,
				Bindings: b.GetBindings(),
				Count:    0,
				Error:    nil,
				Time:     0,
				Raw:      nil,
			}, nil
		}
		if b.Tx != nil {
			result, err = b.Tx.Delete(b.PreparedSql, b.GetBindings())
		} else {
			result, err = b.GetConnection().Delete(b.PreparedSql, b.GetBindings())
		}
		return
	})

	b.ApplyAfterQueryCallbacks()
	return
}

func (b *Builder) NewBuilder() *Builder {
	if b.Tx != nil {
		return NewTxBuilder(b.Tx)
	}
	return NewQueryBuilder(b.Connection)
}
func (b *Builder) Raw() *sql.DB {
	return b.Connection.GetDB()
}

/*
Get Execute the query as a "select" statement.

 1. select *

    users := make([]User,10)

    Model(&User{}).Get(&users)

    select * from users

 2. select id,user_name

    var usersMap []map[string]interface{}

    DB.Connection("default").Table("users").Get(&usersMap, "id", "user_name")

    select id,user_name from users
*/
func (b *Builder) Get(dest interface{}, columns ...interface{}) (result Result, err error) {

	if len(columns) > 0 {
		b.Select(columns...)
	}

	b.Dest = dest
	b.ApplyBeforeQueryCallbacks()
	result, err = b.RunSelect()
	b.ApplyAfterQueryCallbacks()

	return
}

/*
Pluck Get a collection instance containing the values of a given column.
*/
func (b *Builder) Pluck(dest interface{}, params string) (Result, error) {
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
func (b *Builder) Value(dest interface{}, column string) (Result, error) {
	return b.First(dest, column)
}
func (b *Builder) Tap(cb func(builder *Builder) *Builder) *Builder {
	return cb(b)
}

/*
Reset
reset bindings and components
*/
func (b *Builder) Reset(targets ...string) *Builder {
	for _, componentName := range targets {
		switch componentName {
		case TYPE_COLUMN:
			delete(b.Components, TYPE_COLUMN)
			delete(b.Bindings, TYPE_COLUMN)
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
		case TYPE_SELECT:
			delete(b.Bindings, TYPE_SELECT)
			delete(b.Components, TYPE_WHERE)
			b.Columns = nil
		default:
			panic("unknown component name: " + componentName)
		}
	}
	return b
}

/*
BeforeQuery Register a closure to be invoked before the query is executed.
*/
func (b *Builder) BeforeQuery(callback func(builder *Builder)) *Builder {

	b.BeforeQueryCallBacks = append(b.BeforeQueryCallBacks, callback)
	return b
}

/*
AfterQuery Register a closure to be invoked after the query is executed.
*/
func (b *Builder) AfterQuery(callback func(builder *Builder)) *Builder {

	b.AfterQueryCallBacks = append(b.AfterQueryCallBacks, callback)

	return b
}

/*
ApplyAfterQueryCallbacks Apply the after query callbacks to the builder.
*/
func (b *Builder) ApplyAfterQueryCallbacks() {
	for _, callBack := range b.AfterQueryCallBacks {
		callBack(b)
	}
	b.AfterQueryCallBacks = nil
}

/*
ApplyBeforeQueryCallbacks Apply the before query callbacks to the builder.
*/
func (b *Builder) ApplyBeforeQueryCallbacks() {
	for _, callBack := range b.BeforeQueryCallBacks {
		callBack(b)
	}
	b.BeforeQueryCallBacks = nil
}

func (b *Builder) addNewWheresWithInGroup(count int) *Builder {
	wheres := b.Wheres
	b.Wheres = nil
	b.groupWhereSliceForScope(wheres[0:count])
	b.groupWhereSliceForScope(wheres[count:])
	return b
}
func (b *Builder) groupWhereSliceForScope(wheres []Where) *Builder {

	index := 0
	for i, where := range wheres {
		if where.Boolean == BOOLEAN_OR {
			index = i
			break
		}
	}
	if index > 0 {
		b.WhereNested(wheres, wheres[0].Boolean)
	} else {
		b.Where(wheres)
	}
	return b
}

/*
Paginate Paginate the given query into a simple paginator.
items should be a pointer of slice
perPage is the page size
currentPage starts from 1
columns is the database table fileds to be selected,default is *

 1. Paginate(&users,10,1)

    select * from users limit 10 offset 0

 2. Paginate(&users,10,2,[]interface{})
*/
func (b *Builder) Paginate(items interface{}, perPage, currentPage int64, columns ...interface{}) (*Paginator, error) {
	if len(b.Groups) > 0 || len(b.Havings) > 0 {
		panic("having/group pagination not supported")
	}
	p := &Paginator{
		Items:       items,
		Total:       0,
		PerPage:     perPage,
		CurrentPage: currentPage,
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
func (b *Builder) PaginateUsingPaginator(p *Paginator, columns ...interface{}) (*Paginator, error) {
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
func (b *Builder) GetCountForPagination() (int64, error) {
	var c int64
	_, err := b.CloneWithout(TYPE_COLUMN, TYPE_ORDER, TYPE_OFFSET, TYPE_LIMIT).
		CloneWithoutBindings(TYPE_SELECT, TYPE_ORDER).Count(&c)
	return c, err
}

func (b *Builder) Chunk(dest interface{}, chunkSize int64, callback func(dest interface{}) error) (err error) {
	if len(b.Orders) == 0 {
		panic(errors.New("must specify an orderby clause when using Chunk method"))
	}
	var page int64 = 1
	var count int64 = 0
	tempDest := reflect.New(reflect.Indirect(reflect.ValueOf(dest)).Type()).Interface()
	get, err := b.ForPage(1, chunkSize).Get(tempDest)
	if err != nil {
		return
	}
	count = get.Count
	for count > 0 {
		err = callback(tempDest)
		if err != nil {
			return
		}
		if count != chunkSize {
			break
		} else {
			page++
			tempDest = reflect.New(reflect.Indirect(reflect.ValueOf(dest)).Type()).Interface()
			get, err = b.ForPage(page, chunkSize).Get(tempDest)
			if err != nil {
				return err
			}
			count = get.Count
		}
	}
	return nil
}
func (b *Builder) ChunkById(dest interface{}, chunkSize int64, callback func(dest interface{}) error, extra ...string) (err error) {
	eleType := reflect.TypeOf(dest).Elem().Elem()
	if eleType.Kind() == reflect.Ptr {
		eleType = eleType.Elem()
	}

	isMap := eleType.Kind() == reflect.Map
	var column string
	var item *Model

	if isMap {
		if extra == nil || len(extra) == 0 {
			if len(b.Orders) == 0 {
				panic(errors.New("must specify an orderby clause when using ChunkById method"))
			} else {
				column = b.Orders[0].Column.(string)
			}
		} else {
			column = extra[0]
		}
	} else {
		item = GetParsedModel(dest)
		column = item.PrimaryKey.ColumnName
	}

	page := int64(1)
	count := int64(0)
	tempDest := reflect.New(reflect.Indirect(reflect.ValueOf(dest)).Type())
	nb := Clone(b)
	get, err := nb.ForPage(1, chunkSize).OrderBy(column).Get(tempDest.Interface())
	if err != nil {
		return
	}
	count = get.Count
	for count > 0 {
		err = callback(tempDest.Interface())
		if err != nil {
			return
		}
		if count != chunkSize {
			break
		} else {
			page++
			nb = Clone(b)
			lastEle := tempDest.Elem().Index(tempDest.Elem().Len() - 1)
			var lastId interface{}
			if isMap {
				lastId = lastEle.MapIndex(reflect.ValueOf(column)).Interface()
			} else {
				lastId = lastEle.Field(item.PrimaryKey.Index).Interface()
			}
			tempDest = reflect.New(reflect.Indirect(reflect.ValueOf(dest)).Type())
			get, err = nb.Limit(int(chunkSize)).Where(column, ">", lastId).OrderBy(column).Get(tempDest.Interface())
			if err != nil {
				return err
			}
			count = get.Count
		}
	}
	return nil
}
func (b *Builder) Pretend() *Builder {
	b.Pretending = true
	return b
}

/*
Implode concatenate values of a given column as a string.
*/
func (b *Builder) Implode(column string, glue ...string) (string, error) {
	var dest []string
	_, err := b.Pluck(&dest, column)
	if err != nil {
		return "", nil
	}
	var sep string
	if len(glue) == 0 {
		sep = ""
	} else {
		sep = glue[0]
	}

	return strings.Join(dest, sep), nil
}

/*
WhereRowValues Adds a where condition using row values.
*/
func (b *Builder) WhereRowValues(columns []string, operator string, values []interface{}, params ...string) *Builder {
	if len(columns) != len(values) {
		panic(errors.New("argements number mismatch"))
	}
	var boolean = BOOLEAN_AND
	if len(params) == 1 {
		boolean = params[0]
	}
	b.Components[TYPE_WHERE] = struct{}{}
	b.Wheres = append(b.Wheres, Where{
		Type:     CONDITION_TYPE_ROW_VALUES,
		Operator: operator,
		Columns:  columns,
		Values:   values,
		Boolean:  boolean,
	})
	return b
}

// /*
// WhereStruct used for bind request data to struct
// */
//
//	func (b *Builder) WhereStruct(structPointer interface{}) *Builder {
//		b.Where(goeloquent.ExtractStruct(structPointer))
//		return b
//	}
func (b *Builder) Mapping(mapping map[string]interface{}) *Builder {
	m := make(map[string]interface{})
	for k, v := range mapping {
		m[PivotAlias+k] = v
	}
	b.DataMapping = m
	return b
}
func ExtractStruct(target interface{}) map[string]interface{} {
	tv := reflect.Indirect(reflect.ValueOf(target))
	tt := tv.Type()
	result := make(map[string]interface{}, tv.NumField())
	if tt.Kind() == reflect.Struct {
		m := GetParsedModel(tt)
		for column, f := range m.FieldsByDbName {
			keyIndex := f.Index
			if !tv.Field(keyIndex).IsZero() {
				result[column] = tv.Field(keyIndex).Interface()
			}
		}
	} else {
		for i := 0; i < tv.NumField(); i++ {
			key := ToSnakeCase(tt.Field(i).Name)
			result[key] = tv.Field(i).Interface()
		}
	}
	return result
}
func (b *Builder) QualifyColumn(column interface{}) string {
	if e, ok := column.(Expression); ok {
		return string(e)
	}
	c := column.(string)
	if strings.Contains(c, ".") {
		return c
	}
	return b.FromTable.(string) + "." + c
}
func (b *Builder) Prepare(dest interface{}) {
	if b.FromTable == nil {
		m := GetParsedModel(dest)
		b.From(m.Table)
	}
}
