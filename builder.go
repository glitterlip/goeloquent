package goeloquent

import (
	"database/sql"
	"fmt"
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
	"aggregate",
	"columns",
	"from",
	"joins",
	"wheres",
	"groups",
	"havings",
	"orders",
	"limit",
	"offset",
	"unions",
	"lock",
}

type Builder struct {
	Connection IConnection
	Grammar    IGrammar
	//Processor   processors.IProcessor
	Sql              string
	PreSql           strings.Builder
	Bindings         []interface{}
	FromTable        string
	TablePrefix      string
	TableAlias       string
	Wheres           []Where
	Aggregates       []Aggregate
	Columns          []string // The columns that should be returned.
	IsDistinct       bool     // Indicates if the query returns distinct results.
	Joins            []Join
	Groups           []string
	Havings          []Having
	Orders           []Order
	LimitNum         int
	OffsetNum        int
	Unions           []Where
	UnionLimit       int
	UnionOffset      int
	UnionOrders      int
	Components       map[string]interface{}
	Lock             string
	LoggingQueries   bool
	QueryLog         []map[string]interface{}
	Pretending       bool
	PreparedSql      string
	Model            *Model
	Dest             interface{}
	DestReflectValue reflect.Value
	Values           []map[string]interface{}
	EagerLoad        map[string]func(*RelationBuilder) *RelationBuilder //TODO map[string]callback to dynamicly add constraints
	Pivots           []string
	PivotWheres      []Where
}

//type EloquentBuilder struct {
//	Builder
//	Model  *Model
//	Booted bool
//	//GlobalScopes
//	//RemovedScope
//}

func (b *Builder) WithOutGlobalScope(...interface{}) *Builder {

	return b
}
func (b *Builder) WithOutGlobalScopes(...interface{}) *Builder {

	return b
}

const (
	CONDITION_TYPE_BASIC      = "basic"
	CONDITION_TYPE_COLUMN     = "column"
	CONDITION_TYPE_RAW        = "raw"
	CONDITION_TYPE_IN         = "in"
	CONDITION_TYPE_NULL       = "null"
	CONDITION_TYPE_BETWEEN    = "between"
	CONDITION_TYPE_DATE       = "date"
	CONDITION_TYPE_TIME       = "time"
	CONDITION_TYPE_DATETIME   = "datetime"
	CONDITION_TYPE_DAY        = "day"
	CONDITION_TYPE_MONTH      = "month"
	CONDITION_TYPE_YEAR       = "year"
	CONDITION_TYPE_CLOSURE    = "closure"   //todo
	CONDITION_TYPE_NESTED     = "nested"    //todo
	CONDITION_TYPE_SUB        = "subquery"  //todo
	CONDITION_TYPE_EXIST      = "exist"     //todo
	CONDITION_TYPE_ROW_VALUES = "rowValues" //todo
	BOOLEAN_AND               = "and"
	BOOLEAN_OR                = "or"
	CONDITION_JOIN_NOT        = "not" //todo
	JOIN_TYPE_LEFT            = "left"
	JOIN_TYPE_RIGHT           = "right"
	JOIN_TYPE_INNER           = "inner"
	JOIN_TYPE_CROSS           = "cross"
	ORDER_ASC                 = "asc"
	ORDER_DESC                = "desc"
)

type Aggregate struct {
	AggregateName   string
	AggregateColumn string
}
type Join struct {
	JoinType       string
	JoinTable      string
	FirstColumn    string
	ColumnOperator string
	SecondColumn   string
	JoinOperator   string
}
type Order struct {
	Direction string
	Column    string
}
type Having struct {
	HavingType     string
	HavingColumn   string
	HavingOperator string
	HavingValue    interface{}
	HavingBoolean  string
	RawSql         string
}
type Where struct {
	Type         string
	Column       string
	Operator     string
	FirstColumn  string
	SecondColumn string
	RawSql       string
	Value        interface{}
	Values       []interface{}
	Boolean      string
	Not          bool //not in,not between,not null
}

func NewBuilder(c IConnection) *Builder {
	b := Builder{
		Connection: c,
		Components: make(map[string]interface{}),
		EagerLoad:  make(map[string]func(builder *RelationBuilder) *RelationBuilder),
		//Processor:  processors.MysqlProcessor{},
	}
	return &b
}

func (b *Builder) Select(columns ...string) *Builder {
	if columns == nil {
		columns = []string{"*"}
	}
	b.Components["columns"] = nil
	b.Columns = append(b.Columns, columns...)
	return b
}
func (b *Builder) MapInto(dest interface{}) *Builder {
	b.Dest = dest
	return b
}
func (b *Builder) Distinct() *Builder {
	b.IsDistinct = true
	return b
}

func (b *Builder) From(params ...string) *Builder {
	b.Components["from"] = nil
	if len(params) == 2 {
		b.TableAlias = params[1]
		b.FromTable = fmt.Sprintf("%s as %s", params[0], params[1])
	} else {
		b.FromTable = params[0]
	}
	return b
}

func (b *Builder) Join(table, firstColumn, joinOperator, secondColumn string) *Builder {
	return b.join(table, firstColumn, joinOperator, secondColumn, JOIN_TYPE_INNER)
}

func (b *Builder) RightJoin(table, firstColumn, joinOperator, secondColumn string) *Builder {
	return b.join(table, firstColumn, joinOperator, secondColumn, JOIN_TYPE_RIGHT)
}

func (b *Builder) LeftJoin(table, firstColumn, joinOperator, secondColumn string) *Builder {
	return b.join(table, firstColumn, joinOperator, secondColumn, JOIN_TYPE_LEFT)
}

func (b *Builder) CrossJoin(table, firstColumn, joinOperator, secondColumn string) *Builder {
	return b.join(table, firstColumn, joinOperator, secondColumn, JOIN_TYPE_CROSS)
}

func (b *Builder) join(table, firstColumn, joinOperator, secondColumn, joinType string) *Builder {
	b.Components["joins"] = nil
	b.Joins = append(b.Joins, Join{
		JoinType:       joinType,
		JoinTable:      table,
		FirstColumn:    firstColumn,
		ColumnOperator: joinOperator,
		SecondColumn:   secondColumn,
		JoinOperator:   "on",
	})

	return b
}

//column,operator,value,
func (b *Builder) Where(params ...interface{}) *Builder {

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

	b.Wheres = append(b.Wheres, Where{
		Type:     CONDITION_TYPE_BASIC,
		Column:   column,
		Operator: operator,
		Value:    value,
		Boolean:  boolean,
	})
	b.Components["wheres"] = nil
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
func (b *Builder) OrWhere(params ...interface{}) *Builder {
	paramsLength := len(params)
	if paramsLength == 2 {
		params = []interface{}{params[0], "=", params[1], BOOLEAN_OR}
	} else {
		params = append(params, BOOLEAN_OR)
	}
	return b.Where(params)
}
func (b *Builder) WhereColumn(first string, operator string, second ...string) *Builder {
	length := len(second)
	var boolean = BOOLEAN_AND
	var firstColumn = first
	var secondColumn string
	secondColumn = second[0]
	if length == 2 {
		boolean = second[1]
	}
	b.Wheres = append(b.Wheres, Where{
		Type:         CONDITION_TYPE_COLUMN,
		FirstColumn:  firstColumn,
		Operator:     operator,
		SecondColumn: secondColumn,
		Boolean:      boolean,
	})
	b.Components["wheres"] = nil

	return b
}
func (b *Builder) OrWhereColumn(first, operator, second string) *Builder {
	return b.WhereColumn(first, operator, second, BOOLEAN_OR)
}
func (b *Builder) WhereRaw(params ...string) *Builder {
	paramsLength := len(params)
	var boolean string
	value := params[0]
	if paramsLength == 1 {
		boolean = BOOLEAN_AND
	} else {
		boolean = params[1]
	}
	b.Wheres = append(b.Wheres, Where{
		Type:    CONDITION_TYPE_RAW,
		RawSql:  value,
		Boolean: boolean,
	})
	b.Components["wheres"] = nil

	return b

}
func (b *Builder) OrWhereRaw(rawSql string) *Builder {
	return b.WhereRaw(rawSql, BOOLEAN_OR)
}

//column values boolean not
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
	pv := reflect.ValueOf(params[1])
	if pv.Type().Kind() == reflect.Slice {
		for i := 0; i < pv.Len(); i++ {
			values = append(values, pv.Index(i).Interface())
		}
	} else {
		values = append(values, pv.Interface())
	}

	b.Wheres = append(b.Wheres, Where{
		Type:    CONDITION_TYPE_IN,
		Column:  params[0].(string),
		Values:  values,
		Boolean: boolean,
		Not:     not,
	})
	b.Components["wheres"] = nil
	return b
}

//column values
func (b *Builder) OrWhereIn(params ...interface{}) *Builder {
	params[2] = BOOLEAN_OR
	return b.WhereIn(params)
}

//column values [ boolean ]
func (b *Builder) WhereNotIn(params ...interface{}) *Builder {
	if len(params) == 2 {
		params[2] = BOOLEAN_AND
		params[3] = true
	} else {
		params[3] = true
	}
	return b.WhereIn(params)
}

//column values
func (b *Builder) OrWhereNotIn(params ...interface{}) *Builder {
	params[2] = BOOLEAN_OR
	params[3] = false
	return b.WhereIn(params)
}

/*
   params takes in below order:
   1. column string
   2. boolean string in [2]string{"and","or"}
   3. type string "not"
*/
func (b *Builder) WhereNull(column string, params ...interface{}) *Builder {
	paramsLength := len(params)
	var boolean = BOOLEAN_AND
	var not = false
	if paramsLength > 0 {
		if params[0] != BOOLEAN_AND {
			boolean = BOOLEAN_OR
		}
	}
	if paramsLength > 1 {
		not = params[1].(bool)
	}
	b.Components["wheres"] = nil
	b.Wheres = append(b.Wheres, Where{
		Type:    CONDITION_TYPE_NULL,
		Column:  column,
		Boolean: boolean,
		Not:     not,
	})
	return b
}
func (b *Builder) WhereNotNull(column string, params ...interface{}) *Builder {
	paramsLength := len(params)
	if paramsLength == 0 {
		params = append(params, BOOLEAN_AND, true)
	} else if paramsLength == 1 {
		params = append(params, true)
	}
	return b.WhereNull(column, params...)
}

func (b *Builder) OrWhereNull(column string, params ...interface{}) *Builder {
	paramsLength := len(params)
	if paramsLength == 0 {
		params = append(params, BOOLEAN_OR, false)
	} else if paramsLength == 1 {
		params = append(params, false)
	}
	return b.WhereNull(column, params...)
}
func (b *Builder) OrWhereNotNull(column string) *Builder {
	params := []interface{}{BOOLEAN_OR, true}
	return b.WhereNull(column, params)
}

/*
   params takes in below order:
   1. column string
   2. values []string{"min","max"}
   3. boolean string in [2]string{"and","or"}
   4. not in [true,false]
*/
func (b *Builder) WhereBetween(params ...interface{}) *Builder {
	paramsLength := len(params)
	var boolean = BOOLEAN_AND
	not := false
	if paramsLength > 2 {
		boolean = params[3].(string)
	}
	var betweenType = CONDITION_TYPE_BETWEEN
	if paramsLength > 3 {
		not = params[3].(bool)
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
		params[2] = BOOLEAN_AND
	}
	params[3] = true
	return b.WhereBetween(params)
}
func (b *Builder) OrWhereBetween(params ...interface{}) *Builder {
	params[2] = BOOLEAN_OR
	return b.WhereBetween(params)
}
func (b *Builder) OrWhereNotBetween(params ...interface{}) *Builder {
	params[2] = BOOLEAN_OR
	params[3] = true
	return b.WhereBetween(params)
}

//timefuncion column operator value boolean

func (b *Builder) AddTimeBasedWhere(params ...interface{}) *Builder {
	paramsLength := len(params)
	var timeType = params[0]
	var boolean = BOOLEAN_AND
	var operator string
	var value string
	//timefunction column value
	if paramsLength == 3 {
		operator = "="
		value = params[2].(string)
	} else if paramsLength > 3 {
		//timefunction column operator value
		operator = params[2].(string)
		value = params[3].(string)
		//timefunction column operator value boolean
		if paramsLength > 4 && params[4].(string) != boolean {
			boolean = BOOLEAN_OR
		}
	}
	b.Wheres = append(b.Wheres, Where{
		Type:     timeType.(string),
		Column:   params[0].(string),
		Boolean:  boolean,
		Value:    value,
		Operator: operator,
	})
	b.Components["wheres"] = nil
	return b
}

//column operator value boolean
func (b *Builder) WhereDate(params ...interface{}) *Builder {
	return b.AddTimeBasedWhere(CONDITION_TYPE_DATE, params)
}
func (b *Builder) WhereTime(params ...interface{}) *Builder {

	return b.AddTimeBasedWhere(CONDITION_TYPE_TIME, params)

}
func (b *Builder) WhereDay(params ...interface{}) *Builder {

	return b.AddTimeBasedWhere(CONDITION_TYPE_DAY, params)
}
func (b *Builder) WhereMonth(params ...interface{}) *Builder {
	return b.AddTimeBasedWhere(CONDITION_TYPE_MONTH, params)
}
func (b *Builder) WhereYear(params ...interface{}) *Builder {
	return b.AddTimeBasedWhere(CONDITION_TYPE_YEAR, params)
}
func (b *Builder) WhereNested(params ...interface{}) *Builder {
	return b.Where()
}
func (b *Builder) WhereExists(params ...interface{}) *Builder {
	return b.Where()
}

//column operator value boolean
func (b *Builder) GroupBy(column string) *Builder {
	b.Groups = append(b.Groups, column)
	b.Components["groups"] = nil
	return b
}
func (b *Builder) Having(params ...interface{}) *Builder {
	havingBoolean := BOOLEAN_AND
	if len(params) > 3 {
		if params[4] != BOOLEAN_AND {
			havingBoolean = BOOLEAN_OR
		}
	}
	having := Having{
		HavingType:     CONDITION_TYPE_BASIC,
		HavingColumn:   params[0].(string),
		HavingOperator: params[1].(string),
		HavingValue:    params[2],
		HavingBoolean:  havingBoolean,
	}
	b.Components["havings"] = nil
	b.Havings = append(b.Havings, having)
	return b
}
func (b *Builder) OrHaving(params ...interface{}) *Builder {
	params[4] = BOOLEAN_OR
	return b.Having(params)
}
func (b *Builder) HavingBetween(params ...interface{}) *Builder {
	havingBoolean := BOOLEAN_AND
	if len(params) > 3 {
		if params[4] != BOOLEAN_AND {
			havingBoolean = BOOLEAN_OR
		}
	}
	having := Having{
		HavingType:     CONDITION_TYPE_BETWEEN,
		HavingColumn:   params[0].(string),
		HavingOperator: params[1].(string),
		HavingValue:    params[2],
		HavingBoolean:  havingBoolean,
	}
	b.Components["havings"] = nil
	b.Havings = append(b.Havings, having)
	return b
}
func (b *Builder) OrderBy(params ...string) *Builder {
	var order = ORDER_ASC
	if len(params) > 1 {
		order = params[1]
	}
	b.Orders = append(b.Orders, Order{
		Direction: order,
		Column:    params[0],
	})
	b.Components["orders"] = nil

	return b
}
func (b *Builder) Limit(n int) *Builder {
	b.Components["limit"] = nil
	b.LimitNum = n
	return b
}
func (b *Builder) Offset(n int) *Builder {
	b.OffsetNum = n
	b.Components["offset"] = nil
	return b
}

func (b *Builder) Union(n int) *Builder {
	b.OffsetNum = n
	return b
}
func (b *Builder) LockForUpdate() *Builder {
	b.Lock = " for update "
	return b
}
func (b *Builder) SharedLock() *Builder {
	b.Lock = " lock in share mode "
	return b
}
func (b *Builder) WhereKey(keys interface{}) *Builder {

	b.WhereIn(b.Model.PrimaryKey.ColumnName, keys)
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
func (b *Builder) BeginTransaction() *Builder {
	b.Connection.BeginTransaction()
	return b
}
func (b *Builder) Commit() error {
	return b.Connection.Commit()
}
func (b *Builder) Rollback() {
	b.Connection.RollBack()
	return
}
func (b *Builder) Find(dest interface{}, params interface{}) (result sql.Result, err error) {
	b.WhereKey(params)
	d := reflect.Indirect(reflect.ValueOf(dest))
	if d.Type().Kind() == reflect.Slice {
		return b.Get(dest)
	} else {
		return b.First(dest)
	}
}
func (b *Builder) First(dest interface{}) (result sql.Result, err error) {
	b.Limit(1)
	return b.Get(dest)
}
func (b *Builder) SetModel(model interface{}) *Builder {
	if model != nil {
		b.Model = GetParsedModel(model)
		b.From(b.Model.Table)
	}
	if b.Connection == nil {
		if c, ok := reflect.New(model.(reflect.Type)).Elem().Interface().(ConnectionName); ok {
			b.Connection = *Eloquent.Connection(c.ConnectionName())
		} else {
			b.Connection = *Eloquent.Connection("default")

		}
	}
	return b

}
func (b *Builder) RunSelect() (result sql.Result, err error) {
	result, err = b.Connection.Select(b.Grammar.CompileSelect(), b.Bindings, b.Dest)
	if b.Model.IsEloquent {
		BatchSync(b.Dest, true)
	}
	return
}
func (b *Builder) WithPivot(columns ...string) *Builder {
	b.Pivots = append(b.Pivots, columns...)
	return b
}
func (b *Builder) Paginate(n int) *Builder {
	b.OffsetNum = n
	return b
}
func (b *Builder) SimplePaginate(n int) *Builder {
	b.OffsetNum = n
	return b
}
func (b *Builder) Exists(n int) *Builder {
	b.OffsetNum = n
	return b
}
func (b *Builder) Pluck(n int) *Builder {
	b.OffsetNum = n
	return b
}
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
	b.Components["aggregate"] = nil
	result, err = b.RunSelect()
	b.logQuery(b.PreparedSql, b.Bindings, time.Since(start))
	return
}
func (b *Builder) Count(dest interface{}, column ...string) (result sql.Result, err error) {
	return b.Aggregate(dest, "count", column...)
}
func (b *Builder) Min(dest interface{}, column ...string) (result sql.Result, err error) {

	return b.Aggregate(dest, "min", column...)

}
func (b *Builder) Max(dest interface{}, column ...string) (result sql.Result, err error) {

	return b.Aggregate(dest, "max", column...)

}
func (b *Builder) Avg(dest interface{}, column ...string) (result sql.Result, err error) {

	return b.Aggregate(dest, "avg", column...)

}
func (b *Builder) Sum(dest interface{}, column ...string) (result sql.Result, err error) {

	return b.Aggregate(dest, "sum", column...)

}

//todo: support struct
func (b *Builder) Insert(values interface{}) (sql.Result, error) {
	v := reflect.Indirect(reflect.ValueOf(values))
	var items []map[string]interface{}
	if v.Kind() == reflect.Slice {
		items = v.Interface().([]map[string]interface{})
	} else {
		s := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(values)), 1, 1)
		s.Index(0).Set(v)
		items = s.Interface().([]map[string]interface{})
	}
	b.Grammar.CompileInsert(items)
	return b.Connection.Insert(b.PreparedSql, b.Bindings)
}
func (b *Builder) InsertGetId(n int) interface{} {

	return 1
}
func (b *Builder) Update(v map[string]interface{}) (sql.Result, error) {

	b.Grammar.CompileUpdate(v)
	return b.Connection.Update(b.PreparedSql, b.Bindings)
}
func (b *Builder) UpdateOrInsert(n int) interface{} {

	return 1
}
func (b *Builder) Increment(n int) interface{} {

	return 1
}
func (b *Builder) Decrement(n int) interface{} {

	return 1
}
func (b *Builder) Delete() (sql.Result, error) {

	b.Grammar.CompileDelete()
	return b.Connection.Delete(b.PreparedSql, b.Bindings)
}
func (b *Builder) Raw(n int) interface{} {

	return 1
}
func (b *Builder) Dd(n int) interface{} {

	return 1
}
func (b *Builder) Dump(n int) interface{} {

	return 1
}
func (b *Builder) ToSql(n int) interface{} {

	return 1
}
func (b *Builder) Pretend() {

	b.Pretending = true
}

func (b *Builder) Load(...interface{}) *Builder {
	return b
}

func (b *Builder) With(relations ...interface{}) *Builder {
	for _, relation := range relations {
		switch relation.(type) {
		case string:
			b.EagerLoad[relation.(string)] = func(builder *RelationBuilder) *RelationBuilder {
				return builder
			}
		case []string:
			for _, r := range relation.([]string) {
				b.EagerLoad[r] = func(builder *RelationBuilder) *RelationBuilder {
					return builder
				}
			}
		case map[string]func(builder *Builder) *Builder:
			for relation, fn := range relation.(map[string]func(builder *RelationBuilder) *RelationBuilder) {
				b.EagerLoad[relation] = fn
			}
		}
	}

	return b
}
func (b *Builder) logQuery(query string, bindings []interface{}, elapsed time.Duration) {
	//if b.LoggingQueries {
	log := map[string]interface{}{
		"query":    query,
		"bindings": bindings,
		"time":     elapsed.String(),
	}
	fmt.Println(log)
	//b.QueryLog = append(b.QueryLog)
	//}
}
func (b *Builder) Get(dest interface{}) (result sql.Result, err error) {
	var start = time.Now()
	b.Dest = dest
	b.DestReflectValue = reflect.ValueOf(dest)
	if _, ok := b.Components["columns"]; !ok {
		b.Components["columns"] = nil
		b.Columns = append(b.Columns, "*")
	}

	result, err = b.RunSelect()
	b.logQuery(b.PreparedSql, b.Bindings, time.Since(start))
	if len(b.EagerLoad) > 0 {
		rb := RelationBuilder{
			Builder: b,
		}
		rb.EagerLoadRelations(dest)
	}
	return
}
