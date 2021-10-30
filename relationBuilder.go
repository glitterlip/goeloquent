package goeloquent

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type RelationBuilder struct {
	*Builder
	Relation RelationI
}

func (r *RelationBuilder) Get(dest interface{}) (result sql.Result, err error) {
	//TODO:applyScopes
	if _, ok := r.Relation.(*MorphToRelation); ok {
		if reflect.Indirect(reflect.ValueOf(dest)).Type().Name() == "EloquentModel" {
			e := dest.(*EloquentModel)
			p := reflect.New(r.Model.ModelType)
			e.ModelPointer = p
			dest = p.Interface()
		}
	}
	result, err = r.Builder.Get(dest)
	if len(r.Builder.EagerLoad) > 0 {
		//value := reflect.ValueOf(dest)
		//get pointer value
		//realDest := reflect.Indirect(value)
		//models := realDest.Interface().([]interface{})
		r.EagerLoadRelations(dest)
	}

	return
}
func (r *RelationBuilder) EagerLoadRelations(models interface{}) {
	dest := r.Builder.Dest
	//value := reflect.ValueOf(dest)
	//get pointer value
	//realDest := reflect.Indirect(value)

	var model *Model
	if t, ok := models.(*reflect.Value); ok {
		sliceEleType := t.Type().Elem()
		model = GetParsedModel(sliceEleType.PkgPath() + "." + sliceEleType.Name())
	} else {
		model = GetParsedModel(models)
	}

	//models = realDest.Interface()
	for relationName, fn := range r.Builder.EagerLoad {
		if !strings.Contains(relationName, ".") {
			r.EagerLoadRelation(models, model, relationName, fn)
		}
	}
}
func (r *RelationBuilder) LoadPivotColumns(pivots ...string) {
	switch r.Relation.(type) {
	case *BelongsToManyRelation:
		relation := r.Relation.(*BelongsToManyRelation)
		relation.SelectPivots(pivots...)
	}
}
func (r *RelationBuilder) LoadPivotWheres(pivotWheres ...Where) {
	switch r.Relation.(type) {
	case *BelongsToManyRelation:
		relation := r.Relation.(*BelongsToManyRelation)
		relation.WherePivots(pivotWheres...)
	}
}
func (r *RelationBuilder) EagerLoadRelation(models interface{}, model *Model, relationName string, constraints func(*RelationBuilder) *RelationBuilder) {
	if relationMethod, ok := model.Relations[relationName]; ok {
		var params []reflect.Value
		builderI := relationMethod.Call(params)[0].Interface()
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
			//dynamic constraints
			builder = constraints(builder)
			relationResults := builder.GetEager(builder.Relation)
			r.Match(models, relationResults, builder.Relation, relationName)
		}
	} else {
		panic(fmt.Sprintf(" relation : %s for model: %s didn't return a relationbuilder ", relationName, model.Name))
	}
}
func (r *RelationBuilder) GetEager(relation interface{}) interface{} {
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
			modelPointer := GetMorphDBMap(key)
			models := reflect.MakeSlice(reflect.SliceOf(modelPointer.Type()), 0, 10)
			_, err := Eloquent.Model(modelPointer.Type()).WhereIn(morphto.RelatedKey, keys).Get(&models)
			if err != nil {
				panic(err.Error())
			}
			relationResults[key] = models
		}
		return relationResults
	}
	return nil
}
func (r *RelationBuilder) Match(models interface{}, relationResults interface{}, relationI interface{}, relationName string) {

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
