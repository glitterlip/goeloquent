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
	OnlyColumns      map[string]interface{}
	ExceptColumns    map[string]interface{}
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
	CONDITION_TYPE_BASIC       = "basic"
	CONDITION_TYPE_COLUMN      = "column"
	CONDITION_TYPE_RAW         = "raw"
	CONDITION_TYPE_IN          = "in"
	CONDITION_TYPE_NOT_IN      = "not in"
	CONDITION_TYPE_NULL        = "null"
	CONDITION_TYPE_BETWEEN     = "between"
	CONDITION_TYPE_NOT_BETWEEN = "not between"
	CONDITION_TYPE_DATE        = "date"
	CONDITION_TYPE_TIME        = "time"
	CONDITION_TYPE_DATETIME    = "datetime"
	CONDITION_TYPE_DAY         = "day"
	CONDITION_TYPE_MONTH       = "month"
	CONDITION_TYPE_YEAR        = "year"
	CONDITION_TYPE_CLOSURE     = "closure"   //todo
	CONDITION_TYPE_NESTED      = "nested"    //todo
	CONDITION_TYPE_SUB         = "subquery"  //todo
	CONDITION_TYPE_EXIST       = "exist"     //todo
	CONDITION_TYPE_ROW_VALUES  = "rowValues" //todo
	BOOLEAN_AND                = "and"
	BOOLEAN_OR                 = "or"
	CONDITION_JOIN_NOT         = "not" //todo
	JOIN_TYPE_LEFT             = "left"
	JOIN_TYPE_RIGHT            = "right"
	JOIN_TYPE_INNER            = "inner"
	JOIN_TYPE_CROSS            = "cross"
	ORDER_ASC                  = "asc"
	ORDER_DESC                 = "desc"
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
func CloneBuilder(b *Builder) *Builder {
	cb := Builder{
		Connection: b.Connection,
		Components: make(map[string]interface{}),
		EagerLoad:  make(map[string]func(builder *RelationBuilder) *RelationBuilder),
		Grammar:    &MysqlGrammar{},
	}
	cb.Grammar.SetTablePrefix(b.TablePrefix)
	cb.Grammar.SetBuilder(&cb)
	cb.From(b.FromTable)
	return &cb
}
func (b *Builder) Select(columns ...string) *Builder {
	if columns == nil {
		columns = []string{"*"}
	}
	b.Components["columns"] = nil
	b.Columns = append(b.Columns, columns...)
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

	//map of where conditions
	if maps, ok := params[0].([][]interface{}); ok {
		for _, conditions := range maps {
			b.Where(conditions)
		}
		return b
	}
	//convert item of map of where conditions
	if tp, ok := params[0].([]interface{}); ok {
		params = tp
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
		return b.addNestedWhereQuery(cb, boolean)
	} else if where, ok := params[0].(Where); ok {
		b.Wheres = append(b.Wheres, where)
		b.Components["wheres"] = nil
		return b
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
	if clousure, ok := params[0].(func(builder *Builder)); ok {
		return b.Where(clousure, BOOLEAN_OR)
	}
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

	b.Wheres = append(b.Wheres, Where{
		Type:    CONDITION_TYPE_IN,
		Column:  params[0].(string),
		Values:  params[1].([]interface{}),
		Boolean: boolean,
		Not:     not,
	})
	b.Components["wheres"] = nil
	return b
}

//column values
func (b *Builder) OrWhereIn(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_OR, false)
	return b.WhereIn(params...)
}

//column values [ boolean ]
func (b *Builder) WhereNotIn(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_AND, true)
	return b.WhereIn(params...)
}

//column values
func (b *Builder) OrWhereNotIn(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_OR, true)
	return b.WhereIn(params...)
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

//column,boolean
func (b *Builder) WhereNotNull(column string, params ...interface{}) *Builder {
	paramsLength := len(params)
	if paramsLength == 0 {
		params = append(params, BOOLEAN_AND, true)
	} else if paramsLength == 1 {
		params = append(params, true)
	}
	return b.WhereNull(column, params...)
}

//column not
func (b *Builder) OrWhereNull(column string, params ...interface{}) *Builder {
	paramsLength := len(params)
	if paramsLength == 0 {
		params = append(params, BOOLEAN_OR, false)
	} else if paramsLength == 1 {
		params = []interface{}{BOOLEAN_OR, params[0]}
	}
	return b.WhereNull(column, params...)
}
func (b *Builder) OrWhereNotNull(column string) *Builder {
	params := []interface{}{BOOLEAN_OR, true}
	return b.WhereNull(column, params...)
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
		boolean = params[2].(string)
	}
	var betweenType = CONDITION_TYPE_BETWEEN
	if paramsLength > 3 {
		not = params[3].(bool)
	}
	b.Components["wheres"] = nil
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
func (b *Builder) OrWhereBetween(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_OR)
	return b.WhereBetween(params...)
}
func (b *Builder) OrWhereNotBetween(params ...interface{}) *Builder {
	params = append(params, BOOLEAN_OR, true)

	return b.WhereBetween(params...)
}

//timefuncion column operator value boolean
//minum timefuncion column value

func (b *Builder) AddTimeBasedWhere(params ...interface{}) *Builder {
	paramsLength := len(params)
	var timeType = params[0]
	var boolean = BOOLEAN_AND
	var operator string
	var value string
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
	}
	b.Wheres = append(b.Wheres, Where{
		Type:     timeType.(string),
		Column:   params[1].(string),
		Boolean:  boolean,
		Value:    value,
		Operator: operator,
	})
	b.Components["wheres"] = nil
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
func (b *Builder) WhereNested(params ...interface{}) *Builder {
	if len(params) == 1 {
		params = append(params, BOOLEAN_AND)
	}
	cb := CloneBuilder(b)
	switch params[0].(type) {
	case Where:
		cb.Wheres = append(cb.Wheres, params[0].(Where))
	case [][]interface{}:
		tp := params[0].([][]interface{})
		for i := 0; i < len(tp); i++ {
			cb.Where(tp[i])
		}
	case []interface{}:
		cb.Where(params[0].([]interface{}))
	}
	b.Wheres = append(b.Wheres, Where{
		Type:    CONDITION_TYPE_NESTED,
		Boolean: params[1].(string),
		Value:   cb,
	})
	b.Components["wheres"] = nil
	return b
}
func (b *Builder) WhereSub(column string, operator string, value func(builder *Builder), boolean string) *Builder {
	b.Wheres = append(b.Wheres, Where{
		Type:     CONDITION_TYPE_SUB,
		Operator: operator,
		Value:    value,
		Column:   column,
		Boolean:  boolean,
	})
	b.Components["wheres"] = nil
	return b
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
	pt := reflect.TypeOf(keys)
	if pt.Kind() == reflect.Slice {
		b.WhereIn(b.Model.PrimaryKey.ColumnName, keys)
	} else {
		b.WhereIn(b.Model.PrimaryKey.ColumnName, keys)
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
func (b *Builder) addNestedWhereQuery(cloneBuilder *Builder, boolean string) *Builder {

	if len(cloneBuilder.Wheres) > 0 {
		b.Wheres = append(b.Wheres, Where{
			Type:    CONDITION_TYPE_NESTED,
			Value:   cloneBuilder,
			Boolean: boolean,
		})
		b.Components["wheres"] = nil
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

// SetModel
//parameter model should either be a model pointer or a reflect.Type
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
	b.logQuery(b.PreparedSql, b.Bindings, time.Since(start), result)
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

func (b *Builder) Insert(values interface{}) (sql.Result, error) {
	var start = time.Now()
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
	b.Grammar.CompileInsert(items)
	result, err := b.Connection.Insert(b.PreparedSql, b.Bindings)
	if len(items) == 1 && rv.Kind() == reflect.Struct {
		//set id for simple struct
		m, _ := GetParsed(rv.Type().PkgPath() + "." + rv.Type().Name())
		if m != nil {
			mp := m.(*Model)
			if !mp.IsEloquent {
				if mp.PrimaryKey.Name != "" && mp.PrimaryKey.FieldType.Kind() == reflect.Int64 {
					id, _ := result.LastInsertId()
					rv.Field(mp.PrimaryKey.Index).Set(reflect.ValueOf(id))
				}

			}
		}
	}
	b.logQuery(b.PreparedSql, b.Bindings, time.Since(start), result)
	return result, err
}
func (b *Builder) InsertGetId(n int) interface{} {

	return 1
}
func (b *Builder) Update(v map[string]interface{}) (sql.Result, error) {
	var start = time.Now()
	b.Grammar.CompileUpdate(v)
	result, err := b.Connection.Update(b.PreparedSql, b.Bindings)
	b.logQuery(b.PreparedSql, b.Bindings, time.Since(start), result)
	return result, err
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
	var start = time.Now()
	b.Grammar.CompileDelete()
	result, err := b.Connection.Delete(b.PreparedSql, b.Bindings)
	b.logQuery(b.PreparedSql, b.Bindings, time.Since(start), result)
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
	if Eloquent.LogFunc != nil {
		Eloquent.LogFunc(Log{
			SQL:      query,
			Bindings: bindings,
			Result:   result,
			Time:     elapsed,
		})
	}

}
func (b *Builder) Get(dest interface{}, columns ...string) (result sql.Result, err error) {
	if len(columns) > 0 {
		b.Select(columns...)
	}
	var start = time.Now()
	b.Dest = dest
	b.DestReflectValue = reflect.ValueOf(dest)
	if _, ok := b.Components["columns"]; !ok {
		b.Components["columns"] = nil
		b.Columns = append(b.Columns, "*")
	}

	result, err = b.RunSelect()
	b.logQuery(b.PreparedSql, b.Bindings, time.Since(start), result)
	if len(b.EagerLoad) > 0 {
		rb := RelationBuilder{
			Builder: b,
		}
		rb.EagerLoadRelations(dest)
	}
	return
}
func (b *Builder) Pluck(dest interface{}, params string) (sql.Result, error) {
	return b.Get(dest, params)
}

func (b *Builder) Value(dest interface{}, column string) (sql.Result, error) {
	return b.Get(dest, column)
}
