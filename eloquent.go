package goeloquent

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type ScopeFunc = func(builder *EloquentBuilder) *EloquentBuilder
type RelationFunc = func(builder *EloquentBuilder) *EloquentBuilder
type EloquentBuilderFunc = func(builder *EloquentBuilder)
type EloquentBuilderChainFunc = func(builder *EloquentBuilder) *EloquentBuilder

var DefaultEloquentBuilderFunc = func(builder *EloquentBuilder) {

}
var DefaultEloquentBuilderChainFunc = func(builder *EloquentBuilder) *EloquentBuilder {
	return builder
}
var DefaultConstraint = func(builder *EloquentBuilder) *EloquentBuilder {
	return builder
}

type EloquentBuilder struct {
	*Builder
	BaseModel     *Model
	EagerLoad     map[string]func(builder *EloquentBuilder) *EloquentBuilder
	RemovedScopes map[string]struct{}
	Pivots        []string
	PivotWheres   []Where

	BeforeQueryCallBacks []func(*EloquentBuilder)
	AfterQueryCallBacks  []func(*EloquentBuilder)
}

func NewEloquentBuilder(model ...interface{}) (b *EloquentBuilder) {
	b = &EloquentBuilder{
		Builder:       NewQueryBuilder(),
		EagerLoad:     make(map[string]func(builder *EloquentBuilder) *EloquentBuilder),
		RemovedScopes: map[string]struct{}{},
	}
	if len(model) > 0 {
		b.SetModel(model[0])
	}
	return b
}
func ToEloquentBuilder(builder *Builder) *EloquentBuilder {
	return &EloquentBuilder{
		Builder: builder,
	}
}

/*
BeforeQuery Register a closure to be invoked before the query is executed.
*/
func (b *EloquentBuilder) BeforeQuery(callback func(builder *EloquentBuilder)) *EloquentBuilder {

	b.BeforeQueryCallBacks = append(b.BeforeQueryCallBacks, callback)
	return b
}

/*
AfterQuery Register a closure to be invoked after the query is executed.
*/
func (b *EloquentBuilder) AfterQuery(callback func(builder *EloquentBuilder)) *EloquentBuilder {

	b.AfterQueryCallBacks = append(b.AfterQueryCallBacks, callback)

	return b
}

/*
With Set the relationships that should be eager loaded.

 1. .Model(&User{}).With("info").Get(&users)

 2. .Model(&User{}).With("info.address","info.socials").Get(&users)

 3. .Model(&User{}).With("info:id,name,email,dob", "posts").Get(&users)

 4. .Model(&User{}).With([]string{"info", "posts"}).Get(&users)

 5. .Model(&User{}).With(map[string]func(builder *EloquentBuilder) *EloquentBuilder{
    "info": func(builder *EloquentBuilder) *EloquentBuilder {
    return builder.Select("id", "name")
    },
    "posts": func(builder *EloquentBuilder) *EloquentBuilder {
    return builder.Select("id", "title").Where("status",1)
    }}).Get(&users)

 6. .Model(&User{}).With([]string{"info", "posts"}).Get(&users)
*/
func (b *EloquentBuilder) With(relations ...interface{}) *EloquentBuilder {
	var name, relationName string
	var res = make(map[string]func(*EloquentBuilder) *EloquentBuilder)
	for _, relation := range relations {
		switch relation.(type) {
		case string:
			name = relation.(string)
			if pos := strings.Index(name, ":"); pos != -1 {
				relationName = name[0:pos]
			} else {
				relationName = name
			}
			res = AddNestedWiths(relationName, res)
		case []string:
			for _, r := range relation.([]string) {
				b.With(r)
			}
		case map[string]func(builder *EloquentBuilder) *EloquentBuilder:
			for tempRelationName, fn := range relation.(map[string]func(builder *EloquentBuilder) *EloquentBuilder) {
				res[tempRelationName] = fn
				res = AddNestedWiths(tempRelationName, res)
			}
		}
	}

	for s, f := range res {
		b.EagerLoad[s] = f
	}
	return b
}

/*
AddNestedWiths Parse the nested relationships in a relation.
*/
func AddNestedWiths(name string, results map[string]func(builder *EloquentBuilder) *EloquentBuilder) map[string]func(builder *EloquentBuilder) *EloquentBuilder {

	var progress []string
	for _, segment := range strings.Split(name, ".") {
		progress = append(progress, segment)
		ts := strings.Join(progress, ".")
		//set an empty closure if the relation is not set yet
		if _, ok := results[ts]; !ok {
			results[ts] = DefaultConstraint
		}
	}
	return results
}

// TODO whereKey
// TODO whereKeyNot

/*
Prepare check table,connection,model config before executing the query
*/
func (b *EloquentBuilder) Prepare(dest interface{}) {
	if b.BaseModel == nil {
		b.BaseModel = GetParsedModel(dest)
	}
	if b.FromTable == nil {
		if b.BaseModel.TableResolver != nil {
			b.Table(b.BaseModel.TableResolver(b))
		} else {
			b.Table(b.BaseModel.Table)
		}
	}
	if b.Connection == nil {
		if b.BaseModel.ConnectionResolver != nil {
			b.Connection = DB.Connection(b.BaseModel.ConnectionResolver(b))
		} else {
			b.Connection = DB.Connection(b.BaseModel.ConnectionName)
		}
	}
}
func (b *EloquentBuilder) Get(dest interface{}, columns ...interface{}) (result Result, err error) {

	b.ApplyGlobalScopes()
	//table resolver connection resolver
	b.Prepare(dest)
	d := reflect.TypeOf(dest).Elem()
	if d.Kind() == reflect.Slice {
		d = d.Elem()
		if d.Kind() == reflect.Ptr {
			d = d.Elem()
		}
	} else if b.BaseModel.IsEloquent {
		m := reflect.ValueOf(dest).MethodByName(EventRetrieving)
		if m.IsValid() {
			v := m.Call([]reflect.Value{})
			if e, ok := v[0].Interface().(error); ok && e != nil {
				return Result{}, e
			}
		}
	}
	if len(b.EagerLoad) == 0 {
		if len(b.Pivots) > 0 {
			WithPivots(b, b.Pivots)
		}
		if len(b.PivotWheres) > 0 {
			WherePivots(b, b.PivotWheres)
		}
	}
	result, err = b.Builder.Get(dest, columns...)

	if err == nil && b.BaseModel.IsEloquent && d.Kind() == reflect.Struct {
		BatchSync(b.Dest, result.Count > 0)
		m := reflect.ValueOf(dest).MethodByName(EventRetrieved)
		if m.IsValid() {
			m.Call([]reflect.Value{})
		}

	}
	if len(b.EagerLoad) > 0 && result.Count > 0 {
		b.EagerLoadRelations(dest)
	}
	return
}

/*
Find Find a model by its primary key.
*/
func (b *EloquentBuilder) Find(dest interface{}, id interface{}) (result Result, err error) {
	b.WhereKey(id)
	return b.Get(dest)
}

/*
First Execute the query and get the first result.
*/
func (b *EloquentBuilder) First(dest interface{}, columns ...interface{}) (result Result, err error) {
	b.Limit(1)
	return b.Get(dest, columns...)
}

/*
WhereKey Add a where clause on the primary key to the query.
*/
func (b *EloquentBuilder) WhereKey(keys interface{}) *EloquentBuilder {
	if b.BaseModel == nil {
		panic("set model first")
	}
	if reflect.ValueOf(keys).Kind() == reflect.Slice {
		b.WhereIn(b.BaseModel.PrimaryKey.ColumnName, keys)

	} else {
		b.Where(b.BaseModel.PrimaryKey.ColumnName, keys)
	}
	return b
}

/*
FirstOrNew get the first record matching the attributes or instantiate it.
*/
func (b *EloquentBuilder) FirstOrNew(target interface{}, conditions ...map[string]interface{}) (found bool, err error) {
	res, err := b.SetModel(target).Where(conditions[0]).First(target)
	c, err := res.RowsAffected()
	found = c == 1
	if !found {
		err = Fill(target, conditions...)
		if err != nil {
			return
		}
		return false, nil
	}
	return true, nil
}

/*
FirstOrCreate get the first record matching the attributes or create it.
*/
func (b *EloquentBuilder) FirstOrCreate(target interface{}, conditions ...map[string]interface{}) (found bool, err error) {
	res, err := b.SetModel(target).Where(conditions[0]).First(target)
	c, err := res.RowsAffected()
	found = c == 1
	if !found {
		err = Fill(target, conditions...)
		if err != nil {
			return
		}
		_, err = DB.Save(target)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func (b *EloquentBuilder) EagerLoadRelations(models interface{}) {

	var model *Model
	//with reflect.Value of makeslice
	if t, ok := models.(*reflect.Value); ok {
		sliceEleType := t.Type().Elem()
		model = GetParsedModel(sliceEleType.PkgPath() + "." + sliceEleType.Name())
	} else {
		model = GetParsedModel(models)
	}

	//models = realDest.Interface()
	for relationName, fn := range b.EagerLoad {
		if !strings.Contains(relationName, ".") {
			b.EagerLoadRelation(models, model, relationName, fn)
		}
	}
}

func (b *EloquentBuilder) EagerLoadRelation(models interface{}, model *Model, relationName string, constraints func(builder *EloquentBuilder) *EloquentBuilder) {
	if pos := strings.Index(relationName, ":"); pos != -1 {
		relationName = relationName[0:pos]
	}
	if relationMethod, ok := model.Relations[relationName]; ok {
		var params []reflect.Value
		var builder *EloquentBuilder
		var relation RelationI
		relationI := relationMethod.Call(params)[0].Interface()
		switch relationTemp := relationI.(type) {
		case *BelongsToRelation:
			relationTemp.FieldName = relationName
			relation = relationTemp
			builder = relationTemp.EloquentBuilder
		case *BelongsToManyRelation:
			relationTemp.FieldName = relationName
			relation = relationTemp
			builder = relationTemp.EloquentBuilder
		case *HasManyRelation:
			relationTemp.FieldName = relationName
			relation = relationTemp
			builder = relationTemp.EloquentBuilder
		case *HasOneRelation:
			relationTemp.FieldName = relationName
			relation = relationTemp
			builder = relationTemp.EloquentBuilder
		case *HasOneThrough:
			//relationTemp.FieldName = relationName
			//relation = relationTemp
			//builder = relationTemp.EloquentBuilder
		case *HasManyThrough:
			//relationTemp.FieldName = relationName
			//relation = relationTemp
			//builder = relationTemp.EloquentBuilder
		case *MorphByManyRelation:
			relationTemp.FieldName = relationName
			relation = relationTemp
			builder = relationTemp.EloquentBuilder
		case *MorphManyRelation:
			relationTemp.FieldName = relationName
			relation = relationTemp
			builder = relationTemp.EloquentBuilder
		case *MorphOneRelation:
			relationTemp.FieldName = relationName
			relation = relationTemp
			builder = relationTemp.EloquentBuilder
		case *MorphToRelation:
			relationTemp.FieldName = relationName
			relation = relationTemp
			builder = relationTemp.EloquentBuilder
		case *MorphToManyRelation:
			relationTemp.FieldName = relationName
			relation = relationTemp
			builder = relationTemp.EloquentBuilder
		default:
			panic(fmt.Sprintf(" relation : %s for model: %s didn't return a relationbuilder ", relationName, model.Name))
		}
		//load nested relations
		wanted := relationName + "."
		nestedRelation := make(map[string]func(eloquentBuilder *EloquentBuilder) *EloquentBuilder)
		for name, f := range b.EagerLoad {
			if strings.Index(name, wanted) == 0 {
				nestedRelation[strings.Replace(name, wanted, "", -1)] = f
			}
		}
		if len(nestedRelation) > 0 {
			builder.With(nestedRelation)
		}
		//make sure this line runs first so we clear previous wheres and then add dynamic constraints
		relation.AddEagerConstraints(models)
		builder.LoadPivotColumns(relation)
		builder.LoadPivotWheres(relation)
		//dynamic constraints
		builder = constraints(builder)

		relationResults := builder.GetEager(relation)
		builder.Match(models, relationResults, relation, relationName)
	} else {
		panic(fmt.Sprintf(" relation : %s for model: %s didn't return a relationbuilder ", relationName, model.Name))
	}
}

func (b *EloquentBuilder) Model(model interface{}) *EloquentBuilder {
	return b.SetModel(model)
}

/*
SetModel set the model for the eloquent builder

	Model should be a struct or a pointer to a struct with an embedded *eloquent.EloquentModel/*go-eloquent.EloquentModel interface


	e.g. SetModel(&User{})
*/
func (b *EloquentBuilder) SetModel(model interface{}) *EloquentBuilder {
	if model != nil {
		if m, ok := model.(*Model); ok {
			b.BaseModel = m
		} else {
			b.BaseModel = GetParsedModel(model)
		}
		if b.BaseModel.TableResolver == nil {
			b.From(b.BaseModel.Table)
		}
	}
	if b.Connection == nil && b.BaseModel.ConnectionResolver == nil {
		b.Connection = DB.Connection(b.BaseModel.ConnectionName)
	}
	return b

}
func (b *EloquentBuilder) WherePivot(params ...interface{}) *EloquentBuilder {

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
	b.AddBinding([]interface{}{value}, TYPE_WHERE)
	b.Components[TYPE_WHERE] = struct{}{}
	return b
}

/*
WithPivot Set the columns on the pivot table to retrieve.

1. WithPivot("user_roles.status","user_roles.expired_at")
*/
func (b *EloquentBuilder) WithPivot(columns ...string) *EloquentBuilder {
	b.Pivots = append(b.Pivots, columns...)
	return b
}

func (b *EloquentBuilder) ApplyGlobalScopes() {
	if b.BaseModel != nil && len(b.BaseModel.GlobalScopes) > 0 {
		for name, scopeFunc := range b.BaseModel.GlobalScopes {
			if _, removed := b.RemovedScopes[name]; !removed {
				b.callScope(scopeFunc)
			}
		}
	}
}
func (b *EloquentBuilder) callScope(scope ScopeFunc) *EloquentBuilder {
	scope(b)

	/*
		TODO global scope .Where("status",1) with runtime where age > 18 or role = "admin" should be compiled to
		"where status = 1 and (age > 18 or role = 'admin')"
	*/
	return b
}
func (b *EloquentBuilder) WithOutGlobalScopes(names ...string) *EloquentBuilder {
	if len(names) == 0 {
		//remove all
		for name, _ := range b.BaseModel.GlobalScopes {
			b.RemovedScopes[name] = struct{}{}
		}
	} else {
		for _, name := range names {
			b.RemovedScopes[name] = struct{}{}
		}
	}

	return b
}

/*
UpdateOrCreate Create or update a record matching the attributes, and fill it with values.
Deprecated:
*/
func (b *EloquentBuilder) UpdateOrCreate(target interface{}, conditions, values map[string]interface{}) (updated bool, err error) {
	rows, err := b.SetModel(target).Where(conditions).First(target)
	if err != nil {
		return
	}
	c := rows.Count
	if c == 1 {
		err = Fill(target, values)
	} else {
		err = Fill(target, conditions, values)
	}
	if err != nil {
		return
	}
	_, err = DB.Save(target)
	if err != nil {
		return
	}
	return c == 1, nil

}
func (b *EloquentBuilder) LoadPivotColumns(relation RelationI) {
	switch relation.(type) {
	case *BelongsToManyRelation:
	case *MorphToManyRelation:
	case *MorphByManyRelation:
		WithPivots(b, b.Pivots)
	}
}
func (b *EloquentBuilder) LoadPivotWheres(relation RelationI) {
	switch r := relation.(type) {
	case *BelongsToManyRelation:
	case *MorphToManyRelation:
	case *MorphByManyRelation:
		WherePivots(r.EloquentBuilder, b.PivotWheres)
	}
}
func (b *EloquentBuilder) GetEager(relationT RelationI) reflect.Value {
	switch relation := relationT.(type) {
	case *BelongsToRelation,
		*BelongsToManyRelation,
		*HasOneRelation,
		*HasManyRelation,
		*MorphManyRelation,
		*MorphOneRelation,
		*MorphToManyRelation,
		*MorphByManyRelation:
		relationResults := reflect.MakeSlice(reflect.SliceOf(b.BaseModel.ModelType), 0, 10)
		_, err := b.Get(&relationResults)
		if err != nil {
			panic(err.Error() + "\n " + b.PreparedSql)
		}
		return relationResults
	case *MorphToRelation:
		morphto := relation
		relationResults := map[string]reflect.Value{}
		for key, keys := range morphto.Groups {
			modelPointer := GetMorphDBMap(key)
			models := reflect.MakeSlice(reflect.SliceOf(modelPointer.Type()), 0, 10)
			nb := DB.Model(modelPointer.Type())
			nb.Connection = b.Connection
			nb.Tx = b.Tx
			_, err := nb.WhereIn(morphto.RelatedModelIdColumn, keys).Get(&models)
			if err != nil {
				panic(err.Error())
			}
			relationResults[key] = models
		}
		return reflect.ValueOf(relationResults)
	default:
		panic(fmt.Sprintf("relation type %T not supported", relationT))
	}
}
func (b *EloquentBuilder) Match(models interface{}, relationResults reflect.Value, relationI RelationI, relationName string) {

	//results := relationResults.(*reflect.Value)
	switch relationI.(type) {
	case *HasOneRelation:
		relation := relationI.(*HasOneRelation)
		relation.RelationTypeName = Relations(relationName)
		MatchHasOne(models, relationResults, relation)
	case *HasManyRelation:
		relation := relationI.(*HasManyRelation)
		relation.RelationTypeName = Relations(relationName)
		MatchHasMany(models, relationResults, relation)
	case *BelongsToRelation:
		relation := relationI.(*BelongsToRelation)
		relation.RelationTypeName = Relations(relationName)
		MatchBelongsTo(models, relationResults, relation)
	case *BelongsToManyRelation:
		relation := relationI.(*BelongsToManyRelation)
		relation.RelationTypeName = Relations(relationName)
		MatchBelongsToMany(models, relationResults, relation)
	//case *HasManyThrough:
	//case *HasOneThrough:
	case *MorphOneRelation:
		relation := relationI.(*MorphOneRelation)
		relation.RelationTypeName = Relations(relationName)
		MatchMorphOne(models, relationResults, relation)
	case *MorphManyRelation:
		relation := relationI.(*MorphManyRelation)
		relation.RelationTypeName = Relations(relationName)
		MatchMorphMany(models, relationResults, relation)
	case *MorphToRelation:
		relation := relationI.(*MorphToRelation)
		relation.RelationTypeName = Relations(relationName)
		MatchMorphTo(models, relationResults, relation)
	case *MorphToManyRelation:
		relation := relationI.(*MorphToManyRelation)
		relation.RelationTypeName = Relations(relationName)
		MatchMorphToMany(models, relationResults, relation)
	case *MorphByManyRelation:
		relation := relationI.(*MorphByManyRelation)
		relation.RelationTypeName = Relations(relationName)
		MatchMorphByMany(models, relationResults, relation)
	}
}
func (b *EloquentBuilder) WithMax(relation string, column string, constraints ...EloquentBuilderChainFunc) *EloquentBuilder {
	if len(constraints) > 0 {
		return b.WithAggregate(map[string]EloquentBuilderChainFunc{relation: constraints[0]}, column, "Max")
	}
	return b.WithAggregate(map[string]EloquentBuilderChainFunc{relation: DefaultConstraint}, column, "Max")
}
func (b *EloquentBuilder) WithMin(relation string, column string, constraints ...EloquentBuilderChainFunc) *EloquentBuilder {
	if len(constraints) > 0 {
		return b.WithAggregate(map[string]EloquentBuilderChainFunc{relation: constraints[0]}, column, "Min")
	}
	return b.WithAggregate(map[string]EloquentBuilderChainFunc{relation: DefaultConstraint}, column, "Min")
}

func (b *EloquentBuilder) WithSum(relation string, column string, constraints ...EloquentBuilderChainFunc) *EloquentBuilder {
	if len(constraints) > 0 {
		return b.WithAggregate(map[string]EloquentBuilderChainFunc{relation: constraints[0]}, column, "Sum")
	}
	return b.WithAggregate(map[string]EloquentBuilderChainFunc{relation: DefaultConstraint}, column, "Sum")
}
func (b *EloquentBuilder) WithAvg(relation string, column string, constraints ...EloquentBuilderChainFunc) *EloquentBuilder {
	if len(constraints) > 0 {
		return b.WithAggregate(map[string]EloquentBuilderChainFunc{relation: constraints[0]}, column, "Avg")
	}
	return b.WithAggregate(map[string]EloquentBuilderChainFunc{relation: DefaultConstraint}, column, "Avg")
}
func (b *EloquentBuilder) WithExists(relation string, constraints ...EloquentBuilderChainFunc) *EloquentBuilder {
	if len(constraints) > 0 {
		return b.WithAggregate(map[string]EloquentBuilderChainFunc{relation: constraints[0]}, "*", "Exists")
	}
	return b.WithAggregate(map[string]EloquentBuilderChainFunc{relation: DefaultConstraint}, "*", "Exists")
}

/*
WithAggregate Add a relationship count / aggregate function to the query.

nested relations are not supported yet
*/
func (b *EloquentBuilder) WithAggregate(relations map[string]EloquentBuilderChainFunc, column string, functionName string) *EloquentBuilder {
	if len(b.Columns) == 0 {
		b.Select("*")
	}

	var aliasCount = 1

	for relationName, constraint := range relations {

		aliasCount++
		var name, alias string
		tempStrs := strings.Split(relationName, " as ")
		if len(tempStrs) == 2 {
			name = tempStrs[0]
			alias = OrmAggregateAlias + tempStrs[1]
		} else {
			name = tempStrs[0]
			str := fmt.Sprintf("%s%s%s%s", OrmAggregateAlias, name, functionName, strings.Title(column))
			re := regexp.MustCompile(`[^[:alnum:][:space:]_]`)
			alias = re.ReplaceAllString(str, "")

			if ag, ok := b.BaseModel.EagerRelationAggregates[name]; ok {
				column = ag.Column
				relationName = ag.RelationFieldName
				constraint = ag.Constraint
				alias = fmt.Sprintf("%s%s", OrmAggregateAlias, name)
				name = relationName
			}
		}
		if _, ok := b.BaseModel.Relations[name]; !ok {
			if _, ok := b.BaseModel.EagerRelationAggregates[name]; !ok {
				panic(fmt.Sprintf("Relation method [%s] not found in model:[%s]", name, b.BaseModel.Name))
			}
		}

		relation := b.BaseModel.Relations[name].Call([]reflect.Value{})[0].Interface().(RelationI)
		var aggregateColumn = column

		var expression = ""
		if column != "*" {
			aggregateColumn = fmt.Sprintf("%s.%s", relation.GetRelated().Table, column)
		}

		if functionName == "Exists" {
			expression = aggregateColumn
		} else {
			expression = fmt.Sprintf("%s(%s)", functionName, aggregateColumn)
		}

		rq := relation.GetEloquentBuilder()
		//relation.RemoveDefaultConstraints()
		rq.Reset(TYPE_WHERE, TYPE_SELECT)

		q := relation.GetRelationExistenceQuery(rq, b, fmt.Sprintf("%s%d", OrmAggregateAlias, aliasCount), expression)

		constraint(q)
		q.Orders = []Order{}
		q.Bindings[TYPE_ORDER] = []interface{}{}
		if functionName == "Exists" {
			b.SelectRaw(fmt.Sprintf("Exists(%s) as %s", q.ToSql(), alias), q.GetBindings())
		} else {
			b.SelectSub(q.Builder, alias)
		}

	}

	return b
}

/*
WithCount Add a relationship count / aggregate function to the query.

 1. WithCount("Posts")
 2. WithCount([]string{"Posts", "Videos"})
 3. WithCount(map[string]func(builder *EloquentBuilder) *EloquentBuilder{
    "Posts as valid": func(builder *EloquentBuilder) *EloquentBuilder {
    return builder.Where("status", 1)
    },
    "Posts as deleted": func(builder *EloquentBuilder) *EloquentBuilder {
    return builder.Where("status", 0)
    },

})
*/
func (b *EloquentBuilder) WithCount(relations interface{}) *EloquentBuilder {
	converted := map[string]EloquentBuilderChainFunc{}
	switch r := relations.(type) {
	case string:
		converted[r] = DefaultConstraint
	case []string:
		for _, relation := range r {
			converted[relation] = DefaultConstraint
		}
	case map[string]EloquentBuilderChainFunc:
		converted = r
	default:
		panic(errors.New("invalid argument for WithCount, available types are string, []string, map[string]func(builder *EloquentBuilder) *EloquentBuilder"))
	}

	return b.WithAggregate(converted, "*", "Count")
}
func WithPivots(builder *EloquentBuilder, columns []string) {
	for _, pivot := range columns {
		if strings.Contains(pivot, ".") {
			ss := strings.SplitN(pivot, ".", 2)
			column := ss[1]
			//user_roles.status => goelo_pivot_status
			builder.Select(fmt.Sprintf("%s as %s%s", pivot, PivotAlias, column))
		}
	}
}
func WherePivots(builder *EloquentBuilder, wheres []Where) {
	for _, where := range wheres {
		builder.Where(where)
	}
}

/**
proxy for builder
*/

// Select set the columns to be selected
func (b *EloquentBuilder) Select(columns ...interface{}) *EloquentBuilder {
	b.Builder.Select(columns...)
	return b
}

// SelectSub Add a subselect expression to the query.
func (b *EloquentBuilder) SelectSub(query interface{}, as string) *EloquentBuilder {
	b.Builder.SelectSub(query, as)
	return b

}

// SelectRaw Add a new "raw" select expression to the query.
func (b *EloquentBuilder) SelectRaw(expression string, bindings ...[]interface{}) *EloquentBuilder {
	b.Builder.SelectRaw(expression, bindings...)
	return b
}

func (b *EloquentBuilder) FromSub(table interface{}, as string) *EloquentBuilder {

	b.Builder.FromSub(table, as)
	return b
}
func (b *EloquentBuilder) FromRaw(raw interface{}, bindings ...[]interface{}) *EloquentBuilder {
	b.Builder.FromRaw(raw, bindings...)
	return b
}

/*
PrependDatabaseNameIfCrossDatabaseQuery Prepend the database name if the query is a cross database query.
TODO
*/
func (b *EloquentBuilder) PrependDatabaseNameIfCrossDatabaseQuery(table string) string {
	return ""
}

func (b *EloquentBuilder) AddSelect(columns ...interface{}) *EloquentBuilder {
	b.Builder.AddSelect(columns...)
	return b
}

func (b *EloquentBuilder) Distinct(distinct ...string) *EloquentBuilder {
	b.Builder.Distinct(distinct...)
	return b
}

func (b *EloquentBuilder) Table(params ...string) *EloquentBuilder {
	b.Builder.Table(params...)
	return b
}

func (b *EloquentBuilder) From(table interface{}, params ...string) *EloquentBuilder {
	b.Builder.From(table, params...)
	return b
}

func (b *EloquentBuilder) Where(params ...interface{}) *EloquentBuilder {
	b.Builder.Where(params...)
	return b
}

func (b *EloquentBuilder) OrWhere(params ...interface{}) *EloquentBuilder {
	b.Builder.OrWhere(params...)
	return b
}

func (b *EloquentBuilder) WhereColumn(first interface{}, second ...string) *EloquentBuilder {
	b.Builder.WhereColumn(first, second...)

	return b
}

func (b *EloquentBuilder) OrWhereColumn(first string, second ...string) *EloquentBuilder {
	b.Builder.OrWhereColumn(first, second...)
	return b
}
func (b *EloquentBuilder) WhereRaw(rawSql string, params ...interface{}) *EloquentBuilder {
	b.Builder.WhereRaw(rawSql, params...)
	return b

}

func (b *EloquentBuilder) OrWhereRaw(rawSql string, bindings ...[]interface{}) *EloquentBuilder {
	b.Builder.OrWhereRaw(rawSql, bindings...)
	return b
}

func (b *EloquentBuilder) WhereIn(params ...interface{}) *EloquentBuilder {
	b.Builder.WhereIn(params...)
	return b
}

func (b *EloquentBuilder) OrWhereIn(params ...interface{}) *EloquentBuilder {
	b.Builder.OrWhereIn(params...)
	return b
}

func (b *EloquentBuilder) WhereNotIn(params ...interface{}) *EloquentBuilder {
	b.Builder.WhereNotIn(params...)
	return b
}

func (b *EloquentBuilder) OrWhereNotIn(params ...interface{}) *EloquentBuilder {
	b.Builder.OrWhereNotIn(params...)
	return b
}

func (b *EloquentBuilder) WhereNull(column interface{}, params ...interface{}) *EloquentBuilder {
	b.Builder.WhereNull(column, params...)
	return b
}

func (b *EloquentBuilder) OrWhereNull(column interface{}, params ...interface{}) *EloquentBuilder {
	b.Builder.OrWhereNull(column, params...)
	return b
}

func (b *EloquentBuilder) WhereNotNull(column interface{}, params ...interface{}) *EloquentBuilder {
	b.Builder.WhereNotNull(column, params...)
	return b

}

func (b *EloquentBuilder) WhereBetween(params ...interface{}) *EloquentBuilder {
	b.Builder.WhereBetween(params...)
	return b
}

func (b *EloquentBuilder) WhereBetweenColumns(column string, values []interface{}, params ...interface{}) *EloquentBuilder {
	b.Builder.WhereBetweenColumns(column, values, params...)
	return b
}

func (b *EloquentBuilder) OrWhereBetween(params ...interface{}) *EloquentBuilder {
	b.Builder.OrWhereBetween(params...)
	return b
}

func (b *EloquentBuilder) OrWhereBetweenColumns(column string, values []interface{}, params ...interface{}) *EloquentBuilder {
	b.Builder.OrWhereBetweenColumns(column, values, params...)
	return b
}

func (b *EloquentBuilder) WhereNotBetween(column string, params ...interface{}) *EloquentBuilder {
	b.Builder.WhereNotBetween(column, params...)
	return b
}

func (b *EloquentBuilder) WhereNotBetweenColumns(column string, values []interface{}, params ...interface{}) *EloquentBuilder {
	b.Builder.WhereNotBetweenColumns(column, values, params...)
	return b
}

func (b *EloquentBuilder) OrWhereNotBetween(params ...interface{}) *EloquentBuilder {
	b.Builder.OrWhereNotBetween(params...)
	return b
}

func (b *EloquentBuilder) OrWhereNotBetweenColumns(column string, values []interface{}) *EloquentBuilder {
	b.Builder.OrWhereNotBetweenColumns(column, values)
	return b
}

func (b *EloquentBuilder) OrWhereNotNull(column interface{}) *EloquentBuilder {
	b.Builder.OrWhereNotNull(column)
	return b
}

func (b *EloquentBuilder) WhereDate(params ...interface{}) *EloquentBuilder {
	b.Builder.WhereDate(params...)
	return b
}

func (b *EloquentBuilder) WhereTime(params ...interface{}) *EloquentBuilder {
	b.Builder.WhereTime(params...)
	return b
}

func (b *EloquentBuilder) WhereDay(params ...interface{}) *EloquentBuilder {
	b.Builder.WhereDay(params...)
	return b
}

func (b *EloquentBuilder) WhereMonth(params ...interface{}) *EloquentBuilder {
	b.Builder.WhereMonth(params...)
	return b
}

func (b *EloquentBuilder) WhereYear(params ...interface{}) *EloquentBuilder {
	b.Builder.WhereYear(params...)
	return b
}

func (b *EloquentBuilder) WhereNested(params ...interface{}) *EloquentBuilder {
	b.Builder.WhereNested(params...)
	return b
}

func (b *EloquentBuilder) WhereSub(column string, operator string, value func(builder *Builder), boolean string) *EloquentBuilder {
	b.Builder.WhereSub(column, operator, value, boolean)
	return b
}

func (b *EloquentBuilder) WhereExists(cb func(builder *Builder), params ...interface{}) *EloquentBuilder {

	b.Builder.WhereExists(cb, params...)
	return b
}

func (b *EloquentBuilder) OrWhereExists(cb func(builder *Builder), params ...interface{}) *EloquentBuilder {

	b.Builder.OrWhereExists(cb, params...)
	return b
}

func (b *EloquentBuilder) WhereNotExists(cb func(builder *Builder), params ...interface{}) *EloquentBuilder {

	b.Builder.WhereNotExists(cb, params...)
	return b
}

func (b *EloquentBuilder) OrWhereNotExists(cb func(builder *Builder), params ...interface{}) *EloquentBuilder {
	b.Builder.OrWhereNotExists(cb, params...)
	return b
}

func (b *EloquentBuilder) WhereJsonContains(column string, value interface{}, params ...interface{}) *EloquentBuilder {
	b.Builder.WhereJsonContains(column, value, params...)
	return b

}

func (b *EloquentBuilder) WhereJsonOverlaps(column string, value interface{}, params ...interface{}) *EloquentBuilder {
	return b
}

func (b *EloquentBuilder) WhereJsonContainsKey(column string, value interface{}, params ...interface{}) *EloquentBuilder {
	return b
}

func (b *EloquentBuilder) WhereJsonLength(column string, operator string, value interface{}, params ...interface{}) *EloquentBuilder {
	return b
}

func (b *EloquentBuilder) GroupBy(columns ...interface{}) *EloquentBuilder {
	b.Builder.GroupBy(columns...)
	return b
}

func (b *EloquentBuilder) GroupByRaw(sql string, bindings ...[]interface{}) *EloquentBuilder {
	b.Builder.GroupByRaw(sql, bindings...)
	return b
}

func (b *EloquentBuilder) Having(params ...interface{}) *EloquentBuilder {
	b.Builder.Having(params...)
	return b
}

func (b *EloquentBuilder) HavingRaw(params ...interface{}) *EloquentBuilder {
	b.Builder.HavingRaw(params...)
	return b
}

func (b *EloquentBuilder) OrHaving(params ...interface{}) *EloquentBuilder {
	b.Builder.OrHaving(params...)
	return b
}

func (b *EloquentBuilder) OrHavingRaw(params ...interface{}) *EloquentBuilder {
	b.Builder.OrHavingRaw(params...)
	return b
}

func (b *EloquentBuilder) HavingBetween(column string, params ...interface{}) *EloquentBuilder {
	b.Builder.HavingBetween(column, params...)
	return b
}

func (b *EloquentBuilder) OrderBy(params ...interface{}) *EloquentBuilder {
	b.Builder.OrderBy(params...)

	return b
}

func (b *EloquentBuilder) OrderByDesc(column string) *EloquentBuilder {
	b.Builder.OrderByDesc(column)
	return b
}

func (b *EloquentBuilder) OrderByRaw(sql string, bindings []interface{}) *EloquentBuilder {
	b.Builder.OrderByRaw(sql, bindings)
	return b
}

func (b *EloquentBuilder) ReOrder(params ...string) *EloquentBuilder {
	b.Builder.ReOrder(params...)
	return b
}

func (b *EloquentBuilder) Limit(n int) *EloquentBuilder {
	b.Builder.Limit(n)
	return b
}

func (b *EloquentBuilder) Offset(n int) *EloquentBuilder {
	b.Builder.Offset(n)
	return b
}

func (b *EloquentBuilder) Lock(lock ...interface{}) *EloquentBuilder {
	b.Builder.Lock(lock...)
	return b
}

func (b *EloquentBuilder) WhereMap(params map[string]interface{}) *EloquentBuilder {
	b.Builder.WhereMap(params)
	return b
}

func (b *EloquentBuilder) ForPageBeforeId(perpage, id int, column string) *EloquentBuilder {

	b.Builder.ForPageBeforeId(perpage, id, column)
	return b
}
func (b *EloquentBuilder) ForPageAfterId(perpage, id int, column string) *EloquentBuilder {

	b.Builder.ForPageAfterId(perpage, id, column)
	return b
}
func (b *EloquentBuilder) Only(columns ...string) *EloquentBuilder {
	b.Builder.Only(columns...)
	return b
}
func (b *EloquentBuilder) Except(columns ...string) *EloquentBuilder {
	b.Builder.Except(columns...)
	return b
}

func (b *EloquentBuilder) InRandomOrder(seed ...int) *EloquentBuilder {
	b.Builder.InRandomOrder(seed...)
	return b
}

func (b *EloquentBuilder) WhereRowValues(columns []string, operator string, values []interface{}, params ...string) *EloquentBuilder {
	b.Builder.WhereRowValues(columns, operator, values, params...)
	return b
}
func (b *EloquentBuilder) Pretend() *EloquentBuilder {
	b.Builder.Pretend()
	return b
}

func (b *EloquentBuilder) Tap(callback func(builder *EloquentBuilder) *EloquentBuilder) *EloquentBuilder {

	return callback(b)
}
func (b *EloquentBuilder) When(boolean bool, cb ...func(builder *EloquentBuilder)) *EloquentBuilder {
	if boolean {
		cb[0](b)
	} else if len(cb) == 2 {
		cb[1](b)
	}
	return b
}
func (b *EloquentBuilder) WithTrashed() *EloquentBuilder {
	return b.WithOutGlobalScopes(GlobalScopeWithoutTrashed)
}
func (b *EloquentBuilder) OnlyTrashed() *EloquentBuilder {
	return b.WithOutGlobalScopes(GlobalScopeWithoutTrashed).WhereNotNull(b.BaseModel.Table + "." + b.BaseModel.DeletedAt)
}
func (b *EloquentBuilder) WithContext(ctx context.Context) *EloquentBuilder {
	b.Builder.WithContext(ctx)
	return b
}
func (b *EloquentBuilder) Paginate(items interface{}, perPage, currentPage int64, columns ...interface{}) (*Paginator, error) {
	b.ApplyGlobalScopes()
	b.Prepare(items)
	return b.Builder.Paginate(items, perPage, currentPage, columns...)
}
func (b *EloquentBuilder) ForPage(page, perPage int64) *EloquentBuilder {

	b.Offset(int((page - 1) * perPage)).Limit(int(perPage))
	return b
}
func (b *EloquentBuilder) Chunk(dest interface{}, chunkSize int64, callback func(dest interface{}) error) (err error) {

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
