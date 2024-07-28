package goeloquent

import (
	"fmt"
	"reflect"
)

type BelongsToRelation struct {
	*Relation
	SelfColumn    string
	RelatedColumn string
}

func (r *BelongsToRelation) AddEagerConstraints(parentModels interface{}) {
	selfParsedModel := GetParsedModel(r.Relation.SelfModel)
	selfRelatedKeyIndex := selfParsedModel.FieldsByDbName[r.SelfColumn].Index
	parentModelSlice := reflect.Indirect(reflect.ValueOf(parentModels))
	var parentModelRelatedKeys []interface{}
	if parentModelSlice.Type().Kind() == reflect.Slice {
		for i := 0; i < parentModelSlice.Len(); i++ {
			parentModel := parentModelSlice.Index(i)
			parentModelRelatedKey := reflect.Indirect(parentModel).Field(selfRelatedKeyIndex).Interface()
			parentModelRelatedKeys = append(parentModelRelatedKeys, parentModelRelatedKey)
		}
	} else if ms, ok := parentModels.(*reflect.Value); ok {
		for i := 0; i < ms.Len(); i++ {
			modelKey := ms.Index(i).Field(selfRelatedKeyIndex).Interface()
			parentModelRelatedKeys = append(parentModelRelatedKeys, modelKey)
		}
	} else {
		model := parentModelSlice
		modelKey := model.Field(selfRelatedKeyIndex).Interface()
		parentModelRelatedKeys = append(parentModelRelatedKeys, modelKey)
	}
	relatedParsedModel := GetParsedModel(r.RelatedModel)
	//remove first where clause to simulate the Relation::noConstraints function in laravel
	r.Wheres = r.Wheres[1:]
	r.Bindings[TYPE_WHERE] = r.Bindings[TYPE_WHERE][1:]
	r.Builder.WhereIn(relatedParsedModel.Table+"."+r.RelatedColumn, parentModelRelatedKeys)
}
func (r *BelongsToRelation) AddConstraints() {
	relatedParsedModel := GetParsedModel(r.RelatedModel)
	r.Builder.Where(relatedParsedModel.Table+"."+r.RelatedColumn, "=", r.GetSelfKey(r.SelfColumn))
	r.Builder.WhereNotNull(relatedParsedModel.Table + "." + r.RelatedColumn)
}

func MatchBelongsTo(selfModels interface{}, relatedModelsValue reflect.Value, relation *BelongsToRelation) {
	relatedModel := GetParsedModel(relation.RelatedModel)
	selfModel := GetParsedModel(relation.SelfModel)

	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(relation.Relation.RelatedModel))
	groupedResults := reflect.MakeMap(groupedResultsMapType)

	//indicate if the relation field on self model is a pointer
	isPtr := selfModel.FieldsByStructName[relation.Relation.FieldName].FieldType.Kind() == reflect.Ptr
	if !relatedModelsValue.IsValid() || relatedModelsValue.IsNil() {
		return
	}

	//create a map[relatedKey]*relatedModel
	for i := 0; i < relatedModelsValue.Len(); i++ {
		related := relatedModelsValue.Index(i)
		relatedKey := relatedModel.FieldsByDbName[relation.RelatedColumn]
		groupKeyStr := fmt.Sprint(related.FieldByIndex([]int{relatedKey.Index}))
		groupKey := reflect.ValueOf(groupKeyStr)
		groupedResults.SetMapIndex(groupKey, related.Addr())
	}

	targetSlice := reflect.Indirect(reflect.ValueOf(selfModels))

	selfColumn := selfModel.FieldsByDbName[relation.SelfColumn]
	selfColumnIndex := selfColumn.Index
	selfRelationField := selfModel.FieldsByStructName[relation.Relation.FieldName]

	if rvP, ok := selfModels.(*reflect.Value); ok {
		for i := 0; i < rvP.Len(); i++ {
			model := rvP.Index(i)
			modelKey := model.Field(selfColumnIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				if isPtr {
					model.Field(selfRelationField.Index).Set(value)
				} else {
					model.Field(selfRelationField.Index).Set(value.Elem())
				}
			}
		}
	} else if targetSlice.Type().Kind() != reflect.Slice {
		model := targetSlice
		modelKey := model.Field(selfColumnIndex)
		modelKeyStr := fmt.Sprint(modelKey)
		value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
		if value.IsValid() {
			if !model.Field(selfRelationField.Index).CanSet() {
				panic(fmt.Sprintf("model: %s field: %s cant be set", selfModel.Name, selfModel.FieldsByDbName[relation.SelfColumn].Name))
			}
			if isPtr {
				model.Field(selfRelationField.Index).Set(value)
			} else {
				model.Field(selfRelationField.Index).Set(value.Elem())
			}
		}

	} else {
		//iterate parentmodels find its match relation and set its relation field
		for i := 0; i < targetSlice.Len(); i++ {
			model := targetSlice.Index(i)
			modelKey := model.Field(selfColumnIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				if !model.Field(selfColumnIndex).CanSet() {
					panic(fmt.Sprintf("model: %s field: %s cant be set", selfModel.Name, selfModel.FieldsByStructName[relation.Relation.FieldName].Name))
				}
				if isPtr {
					model.Field(selfRelationField.Index).Set(value)
				} else {
					model.Field(selfRelationField.Index).Set(value.Elem())
				}
			}
		}
	}
}

func (r *BelongsToRelation) GetRelationExistenceQuery(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {

	if selfQuery.FromTable == relatedQuery.FromTable {
		return r.GetRelationExistenceQueryForSelfRelation(relatedQuery, selfQuery, alias, columns)
	}
	return relatedQuery.Select(Raw(columns)).WhereColumn(fmt.Sprintf("%s.%s", GetParsedModel(r.RelatedModel).Table, r.RelatedColumn), "=", fmt.Sprintf("%s.%s", GetParsedModel(r.SelfModel).Table, r.SelfColumn))

}

func (r *BelongsToRelation) GetRelationExistenceQueryForSelfRelation(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {
	tableAlias := relatedQuery.FromTable.(string) + " as " + alias
	relatedQuery.Select(Raw(columns)).From(tableAlias)

	return relatedQuery.WhereColumn(fmt.Sprintf("%s.%s", GetParsedModel(r.RelatedModel).Table, r.RelatedColumn), "=", fmt.Sprintf("%s.%s", GetParsedModel(r.SelfModel).Table, r.SelfColumn))
}

func (r *BelongsToRelation) GetSelf() *Model {
	return GetParsedModel(r.SelfModel)
}
func (r *BelongsToRelation) GetRelated() *Model {
	return GetParsedModel(r.RelatedModel)
}
