package goeloquent

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

/*
ToSnakeCase converts a camelCase or PascalCase string to snake_case.
*/
func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
func IsQueryable(value interface{}) bool {

	switch value.(type) {
	case QueryBuilder, *QueryBuilder, EloquentBuilder, *EloquentBuilder, func(*QueryBuilder) *QueryBuilder, func(*EloquentBuilder) *EloquentBuilder, func(*QueryBuilder), func(*EloquentBuilder):
		return true
	}
	return false
}

func IsJsonSelector(value string) bool {

	return strings.Contains(value, "->")
}

/*
Equals compares two values of the same kind and returns true if they are equal.
*/
func Equals(a, b interface{}, kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool:
		return a.(bool) == b.(bool)
	case reflect.Int:
		return a.(int) == b.(int) // int is platform dependent, use int64 for comparison
	case reflect.Int8:
		return a.(int8) == b.(int8)
	case reflect.Int16:
		return a.(int16) == b.(int16) // int16 is also platform dependent, use int64 for comparison
	case reflect.Int32:
		return a.(int32) == b.(int32)
	case reflect.Int64:
		return a.(int64) == b.(int64)
	case reflect.Uint:
		return a.(uint) == b.(uint)
	case reflect.Uint8:
		return a.(uint8) == b.(uint8)
	case reflect.Uint16:
		return a.(uint16) == b.(uint16)
	case reflect.Uint32:
		return a.(uint32) == b.(uint32)
	case reflect.Uint64:
		return a.(uint64) == b.(uint64)
	case reflect.Uintptr:
		return a.(uintptr) == b.(uintptr)
	case reflect.Float32:
		return a.(float32) == b.(float32)
	case reflect.Float64:
		return a.(float64) == b.(float64)
	case reflect.String:
		return a.(string) == b.(string)
	case reflect.Complex64:
		return a.(complex64) == b.(complex64)
	case reflect.Complex128:
		return a.(complex128) == b.(complex128)
	case reflect.Struct, reflect.Map, reflect.Slice:
		return reflect.DeepEqual(reflect.ValueOf(a).Interface(), reflect.ValueOf(b).Interface())
	case reflect.Array:
		if reflect.ValueOf(a).Len() != reflect.ValueOf(b).Len() {
			return false
		}
		return a == b
	default:
		return false
	}
}

func CloneQueryBuilder(query *QueryBuilder) *QueryBuilder {
	b := &QueryBuilder{
		Grammar:              query.Grammar,
		Components:           make(map[Component]struct{}),
		Bindings:             make(map[Component][]interface{}),
		Aggregates:           query.Aggregates,
		Columns:              append([]interface{}{}, query.Columns...),
		IsDistinct:           query.IsDistinct,
		IsExist:              query.IsExist,
		DistinctColumns:      append([]string{}, query.DistinctColumns...),
		FromTable:            query.FromTable,
		IndexHint:            query.IndexHint,
		IsJoin:               query.IsJoin,
		Joins:                append([]*JoinBuilder{}, query.Joins...),
		Wheres:               append([]Where{}, query.Wheres...),
		Groups:               append([]interface{}{}, query.Groups...),
		Havings:              append([]Having{}, query.Havings...),
		Orders:               append([]Order{}, query.Orders...),
		LimitNum:             query.LimitNum,
		Grouplimit:           query.Grouplimit,
		OffsetNum:            query.OffsetNum,
		Locks:                query.Locks,
		BeforeQueryCallBacks: append([]StatementFunc{}, query.BeforeQueryCallBacks...),
		AfterQueryCallBacks:  append([]StatementFunc{}, query.AfterQueryCallBacks...),
		TablePrefix:          query.TablePrefix,
		Pretending:           query.Pretending,
		Parent:               query.Parent,
		//RawSql:               query.RawSql,
		RawBindings: append([]interface{}{}, query.RawBindings...),
	}

	for component, _ := range query.Components {
		b.Components[component] = struct{}{}
	}
	for component, bindings := range query.Bindings {
		b.Bindings[component] = append(b.Bindings[component], bindings...)
	}
	b.Statement = CloneStatement(query.Statement)
	b.Statement.QueryBuilder = b
	b.Statement.Context = query.Statement.Context
	return b
}
func CloneEloquentBuilder(eloquentBuilder *EloquentBuilder) *EloquentBuilder {
	base := CloneQueryBuilder(eloquentBuilder.QueryBuilder)
	eb := &EloquentBuilder{
		Base:          base,
		ModelConfig:   eloquentBuilder.ModelConfig,
		Scopes:        map[string]ScopeFunc{},
		RemovedScopes: map[string]struct{}{},
		EagerLoad:     map[string]RelationFunc{},
		Pivots:        append([]string{}, eloquentBuilder.Pivots...),
		PivotWheres:   append([]Where{}, eloquentBuilder.Wheres...),
	}
	for s, scopeFunc := range eloquentBuilder.Scopes {
		eb.Scopes[s] = scopeFunc
	}
	stmt := CloneStatement(eloquentBuilder.Statement)
	eb.Statement = stmt
	eb.Statement.QueryBuilder = base
	eb.Statement.Eloquent = eb
	eb.Statement.Context = base.Statement.Context

	return eb
}
func CloneStatement(stmt *Statement) *Statement {
	newStmt := &Statement{
		Context:    stmt.Context,
		Connection: stmt.Connection,
		Tx:         stmt.Tx,
		Error:      stmt.Error,
		Eloquent:   stmt.Eloquent,
		Pretending: stmt.Pretending,
		Dest:       stmt.Dest,
		DestValue:  stmt.DestValue,
	}
	return newStmt
}

/*
GroupItemsByKey groups items by key. return a map[stringKey][]*itemType
*/
func GroupItemsByKey(items interface{}, key string, config *ModelConfig, isSingle bool) map[string]interface{} {
	grouped := make(map[string]interface{})
	itemsValue := reflect.Indirect(reflect.ValueOf(items))
	if itemsValue.Kind() != reflect.Slice && itemsValue.Kind() != reflect.Array {
		panic("GroupItemsByKey: items must be a slice or array")
	}

	length := itemsValue.Len()

	isPtr := itemsValue.Type().Elem().Kind() == reflect.Ptr
	field := config.LookupField(key)
	name := field.Name
	for i := 0; i < length; i++ {
		itemV := itemsValue.Index(i)
		if !isPtr {
			itemV = itemV.Addr()
		}
		itemKey := itemV.Elem().FieldByName(name).Interface()
		itemKeyStr := fmt.Sprint(itemKey)
		if !isSingle {
			if _, ok := grouped[fmt.Sprint(itemKey)]; !ok {
				slice := reflect.MakeSlice(reflect.SliceOf(reflect.PointerTo(config.ModelType)), 0, length)
				slice = reflect.Append(slice, itemV)
				grouped[itemKeyStr] = slice.Interface()
			} else {
				slice := reflect.ValueOf(grouped[itemKeyStr])
				slice = reflect.Append(slice, itemV)
				grouped[itemKeyStr] = slice.Interface()
			}
		} else {
			grouped[itemKeyStr] = itemV.Interface()
		}

	}
	return grouped
}
