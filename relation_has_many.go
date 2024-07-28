package goeloquent

import (
	"fmt"
	"reflect"
)

type HasManyRelation struct {
	*Relation
	RelatedColumn string
	SelfColumn    string
}

func (r *HasManyRelation) AddEagerConstraints(selfModels interface{}) {
	relatedParsedModel := GetParsedModel(r.Relation.RelatedModel)
	relatedSelfModel := GetParsedModel(r.Relation.SelfModel)
	selfColumnField := relatedSelfModel.FieldsByDbName[r.SelfColumn]
	selfModelSlice := reflect.Indirect(reflect.ValueOf(selfModels))
	var keys []interface{}
	//extract keys from self models
	if selfModelSlice.Type().Kind() == reflect.Slice {
		for i := 0; i < selfModelSlice.Len(); i++ {
			selfModel := selfModelSlice.Index(i)
			selfModelKey := reflect.Indirect(selfModel).Field(selfColumnField.Index).Interface()
			keys = append(keys, selfModelKey)
		}
	} else if ms, ok := selfModels.(*reflect.Value); ok {
		for i := 0; i < ms.Len(); i++ {
			selfModelKey := ms.Index(i).Field(selfColumnField.Index).Interface()
			keys = append(keys, selfModelKey)
		}
	} else {
		model := selfModelSlice
		modelKey := model.Field(selfColumnField.Index).Interface()
		keys = append(keys, modelKey)
	}
	//remove first where clause to simulate the Relation::noConstraints function in laravel
	r.Wheres = r.Wheres[1:]
	r.Bindings[TYPE_WHERE] = r.Bindings[TYPE_WHERE][1:]
	r.Builder.WhereIn(relatedParsedModel.Table+"."+r.RelatedColumn, keys)
}

func (r *HasManyRelation) AddConstraints() {
	relatedParsedModel := GetParsedModel(r.Relation.RelatedModel)
	r.Builder.Where(relatedParsedModel.Table+"."+r.RelatedColumn, "=", r.GetSelfKey(r.SelfColumn))
	r.Builder.WhereNotNull(relatedParsedModel.Table + "." + r.RelatedColumn)
}
func MatchHasMany(models interface{}, related interface{}, relation *HasManyRelation) {
	relatedModelsValue := related.(reflect.Value)
	relatedModels := relatedModelsValue
	relatedModel := GetParsedModel(relation.RelatedModel)
	self := GetParsedModel(relation.SelfModel)
	isPtr := self.FieldsByStructName[relation.Relation.FieldName].FieldType.Elem().Kind() == reflect.Ptr

	var relatedType reflect.Type
	if isPtr {
		relatedType = reflect.ValueOf(relation.Relation.RelatedModel).Type()
	} else {
		relatedType = reflect.ValueOf(relation.Relation.RelatedModel).Elem().Type()
	}
	slice := reflect.MakeSlice(reflect.SliceOf(relatedType), 0, 1)
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(slice))
	groupedResults := reflect.MakeMap(groupedResultsMapType)

	if !relatedModels.IsValid() || relatedModels.IsNil() {
		return
	}

	for i := 0; i < relatedModels.Len(); i++ {
		result := relatedModels.Index(i)
		foreignKeyIndex := relatedModel.FieldsByDbName[relation.RelatedColumn].Index
		groupKey := reflect.ValueOf(fmt.Sprint(result.FieldByIndex([]int{foreignKeyIndex})))
		existed := groupedResults.MapIndex(groupKey)
		if !existed.IsValid() {
			existed = reflect.New(slice.Type()).Elem()
		} else {
			existed = existed.Interface().(reflect.Value)
		}
		ptr := reflect.New(slice.Type())
		if isPtr {
			v := reflect.Append(existed, result.Addr())
			ptr.Elem().Set(v)
		} else {
			v := reflect.Append(existed, result)
			ptr.Elem().Set(v)
		}
		groupedResults.SetMapIndex(groupKey, reflect.ValueOf(ptr.Elem()))
	}

	targetSlice := reflect.Indirect(reflect.ValueOf(models))

	modelRelationFieldIndex := self.FieldsByStructName[relation.Relation.FieldName].Index
	modelKeyFieldIndex := self.FieldsByDbName[relation.SelfColumn].Index

	if rvP, ok := models.(*reflect.Value); ok {
		for i := 0; i < rvP.Len(); i++ {
			model := rvP.Index(i)
			modelKey := model.Field(modelKeyFieldIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				model.Field(modelRelationFieldIndex).Set(value)
			}

		}
	} else if targetSlice.Type().Kind() != reflect.Slice {
		model := targetSlice
		modelKey := model.Field(modelKeyFieldIndex)
		modelKeyStr := fmt.Sprint(modelKey)
		value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
		if value.IsValid() {
			value = value.Interface().(reflect.Value)

			model.Field(modelRelationFieldIndex).Set(value)
		}
	} else {
		for i := 0; i < targetSlice.Len(); i++ {
			model := targetSlice.Index(i)
			modelKey := model.Field(modelKeyFieldIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			modelSlice := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			value := modelSlice
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				model.Field(modelRelationFieldIndex).Set(value)
			}
		}
	}

}
func (r *HasManyRelation) GetRelationExistenceQuery(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {

	if relatedQuery.FromTable.(string) == selfQuery.FromTable.(string) {
		return r.GetRelationExistenceQueryForSelfRelation(relatedQuery, selfQuery, alias, columns)
	}
	relatedParsed := GetParsedModel(r.Relation.RelatedModel)
	selfParsed := GetParsedModel(r.Relation.SelfModel)
	return relatedQuery.Select(Raw(columns)).WhereColumn(selfParsed.Table+"."+r.SelfColumn, "=", relatedParsed.Table+"."+r.RelatedColumn)

}

func (r *HasManyRelation) GetRelationExistenceQueryForSelfRelation(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {
	relatedParsed := GetParsedModel(r.Relation.RelatedModel)
	selfParsed := GetParsedModel(r.Relation.SelfModel)
	relatedQuery.From(relatedQuery.FromTable.(string) + " as " + alias)
	relatedQuery.Select(Raw(columns)).WhereColumn(r.SelfColumn, "=", r.RelatedColumn)
	return relatedQuery.Select(Raw(columns)).WhereColumn(selfParsed.Table+"."+r.SelfColumn, "=", relatedParsed.Table+"."+r.RelatedColumn)
}
func (r *HasManyRelation) GetSelf() *Model {
	return GetParsedModel(r.SelfModel)
}
func (r *HasManyRelation) GetRelated() *Model {
	return GetParsedModel(r.RelatedModel)
}
