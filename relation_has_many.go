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

func (r *HasManyRelation) AddEagerConstraints(models interface{}) {
	relatedParsedModel := GetParsedModel(r.Relation.RelatedModel)
	ReleatedColumnField := relatedParsedModel.FieldsByDbName[r.RelatedColumn]
	modelSlice := reflect.Indirect(reflect.ValueOf(models))
	var keys []interface{}
	if modelSlice.Type().Kind() == reflect.Slice {
		for i := 0; i < modelSlice.Len(); i++ {
			model := modelSlice.Index(i)
			modelKey := reflect.Indirect(model).Field(ReleatedColumnField.Index).Interface()
			keys = append(keys, modelKey)
		}
	} else if ms, ok := models.(*reflect.Value); ok {
		for i := 0; i < ms.Len(); i++ {
			modelKey := ms.Index(i).Field(ReleatedColumnField.Index).Interface()
			keys = append(keys, modelKey)
		}
	} else {
		model := modelSlice
		modelKey := model.Field(ReleatedColumnField.Index).Interface()
		keys = append(keys, modelKey)
	}
	r.Builder.WhereNotNull(r.RelatedColumn)
	r.Builder.WhereIn(r.RelatedColumn, keys)
}

func (r *HasManyRelation) AddConstraints() {
	r.Builder.Where(r.RelatedColumn, "=", r.GetSelfKey(r.SelfColumn))
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
