package goeloquent

import (
	"fmt"
	"reflect"
)

type MorphOneRelation struct {
	*Relation
	RelatedModelIdColumn        string
	RelatedModelTypeColumn      string
	SelfColumn                  string
	RelatedModelTypeColumnValue string
}

func (r *MorphOneRelation) AddEagerConstraints(models interface{}) {
	selfParsedModel := GetParsedModel(r.SelfModel)
	index := selfParsedModel.FieldsByDbName[r.SelfColumn].Index
	modelSlice := reflect.Indirect(reflect.ValueOf(models))
	var keys []interface{}
	if modelSlice.Type().Kind() == reflect.Slice {
		for i := 0; i < modelSlice.Len(); i++ {
			model := modelSlice.Index(i)
			modelKey := reflect.Indirect(model).Field(index).Interface()
			keys = append(keys, modelKey)
		}
	} else if ms, ok := models.(*reflect.Value); ok {
		for i := 0; i < ms.Len(); i++ {
			modelKey := ms.Index(i).Field(index).Interface()
			keys = append(keys, modelKey)
		}
	} else {
		model := modelSlice
		modelKey := model.Field(index).Interface()
		keys = append(keys, modelKey)
	}
	r.Builder.WhereNotNull(r.RelatedModelIdColumn)
	r.Builder.Where(r.RelatedModelTypeColumn, r.RelatedModelTypeColumnValue)
	r.Builder.WhereIn(r.RelatedModelIdColumn, keys)
}
func (r *MorphOneRelation) AddConstraints() {
	selfParsedModel := GetParsedModel(r.SelfModel)
	selfDirect := reflect.Indirect(reflect.ValueOf(r.SelfModel))
	r.Builder.Where(r.RelatedModelTypeColumn, r.RelatedModelTypeColumnValue)
	r.Builder.Where(r.RelatedModelIdColumn, "=", selfDirect.Field(selfParsedModel.FieldsByDbName[r.RelatedModelIdColumn].Index).Interface())
}
func MatchMorphOne(models interface{}, related interface{}, relation *MorphOneRelation) {
	relatedModelsValue := related.(reflect.Value)
	relatedResults := relatedModelsValue
	relatedModel := GetParsedModel(relation.RelatedModel)
	relatedType := reflect.ValueOf(relation.Relation.RelatedModel).Elem().Type()
	slice := reflect.MakeSlice(reflect.SliceOf(relatedType), 0, 1)
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(slice))
	groupedResults := reflect.MakeMap(groupedResultsMapType)
	selfParsedModel := GetParsedModel(relation.SelfModel)
	isPtr := selfParsedModel.FieldsByStructName[relation.Relation.FieldName].FieldType.Kind() == reflect.Ptr
	if !relatedResults.IsValid() || relatedResults.IsNil() {
		return
	}
	if isPtr {
		for i := 0; i < relatedResults.Len(); i++ {
			result := relatedResults.Index(i)
			ownerKeyIndex := relatedModel.FieldsByDbName[relation.RelatedModelIdColumn].Index
			groupKeyS := fmt.Sprint(result.FieldByIndex([]int{ownerKeyIndex}))
			groupKey := reflect.ValueOf(groupKeyS)
			existed := groupedResults.MapIndex(groupKey)
			if !existed.IsValid() {
				existed = reflect.New(slice.Type())
			} else {
				existed = existed.Interface().(reflect.Value)
			}
			ptr := reflect.New(slice.Type())
			ptr.Elem().Set(reflect.Append(existed.Elem(), result))
			groupedResults.SetMapIndex(groupKey, reflect.ValueOf(ptr))
		}
	} else {
		for i := 0; i < relatedResults.Len(); i++ {
			result := relatedResults.Index(i)
			ownerKeyIndex := relatedModel.FieldsByDbName[relation.RelatedModelIdColumn].Index
			groupKeyS := fmt.Sprint(result.FieldByIndex([]int{ownerKeyIndex}))
			groupKey := reflect.ValueOf(groupKeyS)
			existed := groupedResults.MapIndex(groupKey)
			if !existed.IsValid() {
				existed = reflect.New(slice.Type()).Elem()
			} else {
				existed = existed.Interface().(reflect.Value)
			}
			modelSlice := reflect.Append(existed, result)
			groupedResults.SetMapIndex(groupKey, reflect.ValueOf(modelSlice))
		}
	}
	targetSlice := reflect.Indirect(reflect.ValueOf(models))

	modelRelationFieldIndex := selfParsedModel.FieldsByStructName[relation.Relation.FieldName].Index
	modelKeyFieldIndex := selfParsedModel.FieldsByDbName[relation.SelfColumn].Index
	if rvP, ok := models.(*reflect.Value); ok {
		for i := 0; i < rvP.Len(); i++ {
			model := rvP.Index(i)
			modelKey := model.Field(modelKeyFieldIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				if isPtr {
					model.Field(modelRelationFieldIndex).Set(value.Elem().Index(0).Addr())
				} else {
					model.Field(modelRelationFieldIndex).Set(value.Index(0))

				}
			}
		}
	} else if targetSlice.Type().Kind() != reflect.Slice {
		model := targetSlice
		modelKey := model.Field(modelKeyFieldIndex)
		modelKeyStr := fmt.Sprint(modelKey)
		modelSlice := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
		value := modelSlice
		if value.IsValid() {
			value = value.Interface().(reflect.Value)
			if !model.Field(modelRelationFieldIndex).CanSet() {
				panic(fmt.Sprintf("model: %s field: %s cant be set", selfParsedModel.Name, selfParsedModel.FieldsByStructName[relation.Relation.FieldName].Name))
			}
			if isPtr {
				model.Field(modelRelationFieldIndex).Set(value.Elem().Index(0).Addr())
			} else {
				model.Field(modelRelationFieldIndex).Set(value.Index(0))
			}
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
				if !model.Field(modelRelationFieldIndex).CanSet() {
					panic(fmt.Sprintf("model: %s field: %s cant be set", selfParsedModel.Name, selfParsedModel.FieldsByStructName[relation.Relation.FieldName].Name))
				}
				if isPtr {
					model.Field(modelRelationFieldIndex).Set(value.Elem().Index(0).Addr())
				} else {
					model.Field(modelRelationFieldIndex).Set(value.Index(0))
				}
			}
		}
	}
}
