package eloquent

import (
	"fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/glitterlip/goeloquent/query"
	"reflect"
	"strings"
)

type RelationBuilder struct {
	*Builder
	Relation RelationI
}

func (r *Builder) LoadPivotColumns(pivots ...string) {
	switch relation := r.Relation.(type) {
	case *BelongsToManyRelation:
		WithPivots(relation.Builder, relation.PivotTable, pivots)
	case *MorphToManyRelation:
		WithPivots(relation.Builder, relation.PivotTable, pivots)
	case *MorphByManyRelation:
		WithPivots(relation.Builder, relation.PivotTable, pivots)
	}
}
func (r *Builder) LoadPivotWheres(pivotWheres ...Where) {
	switch relation := r.Relation.(type) {
	case *BelongsToManyRelation:
		WherePivots(relation.Builder, relation.PivotTable, pivotWheres)
	case *MorphToManyRelation:
		WherePivots(relation.Builder, relation.PivotTable, pivotWheres)
	case *MorphByManyRelation:
		WherePivots(relation.Builder, relation.PivotTable, pivotWheres)
	}
}
func (r *Builder) EagerLoadRelation(models interface{}, model *Model, relationName string, constraints func(*RelationBuilder) *RelationBuilder) {
	if pos := strings.Index(relationName, ":"); pos != -1 {
		relationName = relationName[0:pos]
	}
	if relationMethod, ok := model.Relations[relationName]; ok {
		var params []reflect.Value
		builderI := relationMethod.Call(params)[0].Interface()
		//FIXME: builder config lost cause log missing
		if builder, ok := builderI.(*RelationBuilder); ok {
			//load nested relations
			wanted := relationName + "."
			nestedRelation := make(map[string]func(*RelationBuilder) *RelationBuilder)
			for name, f := range r.Builder.EagerLoad {
				if strings.Index(name, wanted) == 0 {
					nestedRelation[strings.Replace(name, wanted, "", -1)] = f
				}
			}
			if len(nestedRelation) > 0 {
				builder.With(nestedRelation)
			}
			//make sure this line runs first so we clear previous wheres and then add dynamic constraints
			builder.Relation.AddEagerConstraints(models)
			builder.LoadPivotColumns(r.Builder.Pivots...)
			builder.LoadPivotWheres(r.Builder.PivotWheres...)
			builder.DataMapping = r.Builder.DataMapping
			//dynamic constraints
			builder = constraints(builder)
			//copy conn/tx to new builder
			builder.Connection = r.Connection
			builder.Tx = r.Tx
			relationResults := builder.GetEager(builder.Relation)
			r.Match(models, relationResults, builder.Relation, relationName)
		}
	} else {
		panic(fmt.Sprintf(" relation : %s for model: %s didn't return a relationbuilder ", relationName, model.Name))
	}
}
func (r *Builder) GetEager(relation interface{}) interface{} {
	switch relation.(type) {
	case *BelongsToRelation,
		*BelongsToManyRelation,
		*HasOneRelation,
		*HasManyRelation,
		*MorphManyRelation,
		*MorphOneRelation,
		*MorphToManyRelation,
		*MorphByManyRelation:
		relationResults := reflect.MakeSlice(reflect.SliceOf(r.Builder.Model.ModelType), 0, 10)
		_, err := r.Get(&relationResults)
		if err != nil {
			panic(err.Error())
		}
		return relationResults
	case *MorphToRelation:
		morphto := relation.(*MorphToRelation)
		relationResults := make(map[string]reflect.Value)
		for key, keys := range morphto.Groups {
			modelPointer := goeloquent.GetMorphDBMap(key)
			models := reflect.MakeSlice(reflect.SliceOf(modelPointer.Type()), 0, 10)
			nb := Eloquent.Model(modelPointer.Type())
			nb.Connection = r.Connection
			nb.Tx = r.Tx
			_, err := nb.WhereIn(morphto.RelatedKey, keys).Get(&models)
			if err != nil {
				panic(err.Error())
			}
			relationResults[key] = models
		}
		return relationResults
	}
	return nil
}
func (r *Builder) Match(models interface{}, relationResults interface{}, relationI interface{}, relationName string) {

	//results := relationResults.(*reflect.Value)
	switch relationI.(type) {
	case *HasOneRelation:
		relation := relationI.(*HasOneRelation)
		relation.Name = relationName
		MatchHasOne(models, relationResults, relation)
	case *HasManyRelation:
		relation := relationI.(*HasManyRelation)
		relation.Name = relationName
		MatchHasMany(models, relationResults, relation)
	case *BelongsToRelation:
		relation := relationI.(*BelongsToRelation)
		relation.Name = relationName
		MatchBelongsTo(models, relationResults, relation)
	case *BelongsToManyRelation:
		relation := relationI.(*BelongsToManyRelation)
		relation.Name = relationName
		MatchBelongsToMany(models, relationResults, relation)
	case *HasManyThrough:
	case *HasOneThrough:
	case *MorphOneRelation:
		relation := relationI.(*MorphOneRelation)
		relation.Name = relationName
		MatchMorphOne(models, relationResults, relation)
	case *MorphManyRelation:
		relation := relationI.(*MorphManyRelation)
		relation.Name = relationName
		MatchMorphMany(models, relationResults, relation)
	case *MorphToRelation:
		relation := relationI.(*MorphToRelation)
		relation.Name = relationName
		MatchMorphTo(models, relationResults, relation)
	case *MorphToManyRelation:
		relation := relationI.(*MorphToManyRelation)
		relation.Name = relationName
		MatchMorphToMany(models, relationResults, relation)
	case *MorphByManyRelation:
		relation := relationI.(*MorphByManyRelation)
		relation.Name = relationName
		MatchMorphByMany(models, relationResults, relation)
	}
}
func (r *Builder) WithAggregate(relation []interface{}, column string, function string) *Builder {

	return nil
}
func (r *Builder) WithCount(relation []interface{}) *Builder {
	return r.WithAggregate(relation, "*", "count")
}
func WithPivots(builder *Builder, table string, columns []string) {
	for _, pivot := range columns {
		if strings.Contains(pivot, ".") {
			if strings.Contains(pivot, table) {
				builder.Select(fmt.Sprintf("%s.%s as %s%s", table, pivot, PivotAlias, pivot))
			}
		} else {
			builder.Select(fmt.Sprintf("%s.%s as %s%s", table, pivot, PivotAlias, pivot))
		}
	}
}
func WherePivots(builder *Builder, table string, wheres []query.Where) {
	for _, where := range wheres {
		if strings.Contains(where.Column, ".") {
			if strings.Contains(where.Column, table) {
				sa := strings.SplitN(where.Column, ".", 2)
				builder.Where(table+"."+sa[1], where.Operator, where.Value, where.Boolean)
			}
		} else {
			builder.Where(table+"."+where.Column, where.Operator, where.Value, where.Boolean)
		}
	}
}
