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
	r.Builder.WhereIn(r.RelatedColumn, parentModelRelatedKeys)
}
func (r *BelongsToRelation) AddConstraints() {
	r.Builder.Where(r.RelatedColumn, "=", r.GetSelfKey(r.SelfColumn))
}

func MatchBelongsTo(models interface{}, related interface{}, relation *BelongsToRelation) {
	relatedModels := related.(reflect.Value)
	parsedRelatedModel := GetParsedModel(relation.RelatedModel)
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(relation.Relation.RelatedModel))
	groupedResults := reflect.MakeMap(groupedResultsMapType)
	parent := GetParsedModel(relation.SelfModel)
	relationFieldIsPtr := parent.FieldsByStructName[relation.Relation.FieldName].FieldType.Kind() == reflect.Ptr
	if !relatedModels.IsValid() || relatedModels.IsNil() {
		return
	}

	//create a map[relatedKey]*relatedModel
	for i := 0; i < relatedModels.Len(); i++ {
		result := relatedModels.Index(i)
		relatedKeyIndex := parsedRelatedModel.FieldsByDbName[relation.RelatedColumn].Index
		groupKeyStr := fmt.Sprint(result.FieldByIndex([]int{relatedKeyIndex}))
		groupKey := reflect.ValueOf(groupKeyStr)
		groupedResults.SetMapIndex(groupKey, result.Addr())
	}

	targetSlice := reflect.Indirect(reflect.ValueOf(models))
	modelRelationFieldIndex := parent.FieldsByStructName[relation.Relation.FieldName].Index
	modelKeyFieldIndex := parent.FieldsByDbName[relation.SelfColumn].Index
	if rvP, ok := models.(*reflect.Value); ok {
		for i := 0; i < rvP.Len(); i++ {
			model := rvP.Index(i)
			modelKey := model.Field(modelKeyFieldIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				if relationFieldIsPtr {
					model.Field(modelRelationFieldIndex).Set(value)
				} else {
					model.Field(modelRelationFieldIndex).Set(value.Elem())
				}
			}
		}
	} else if targetSlice.Type().Kind() != reflect.Slice {
		model := targetSlice
		modelKey := model.Field(modelKeyFieldIndex)
		modelKeyStr := fmt.Sprint(modelKey)
		value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
		if value.IsValid() {
			if !model.Field(modelRelationFieldIndex).CanSet() {
				panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.FieldName].Name))
			}
			if relationFieldIsPtr {
				model.Field(modelRelationFieldIndex).Set(value)
			} else {
				model.Field(modelRelationFieldIndex).Set(value.Elem())
			}
		}

	} else {
		//iterate parentmodels find its match relation and set its relation field
		for i := 0; i < targetSlice.Len(); i++ {
			model := targetSlice.Index(i)
			modelKey := model.Field(modelKeyFieldIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				if !model.Field(modelRelationFieldIndex).CanSet() {
					panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.FieldName].Name))
				}
				if relationFieldIsPtr {
					model.Field(modelRelationFieldIndex).Set(value)
				} else {
					model.Field(modelRelationFieldIndex).Set(value.Elem())
				}
			}
		}
	}
}
