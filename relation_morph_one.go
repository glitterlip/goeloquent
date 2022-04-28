package goeloquent

import (
	"fmt"
	"reflect"
)

type MorphOneRelation struct {
	Relation
	RelatedIdColumn   string
	ParentKey         string
	Builder           *Builder
	RelatedTypeColumn string
	MorphType         string
}

func (m *EloquentModel) MorphOne(self interface{}, related interface{}, relatedTypeColumn, relatedIdColumn, selfKey string) *RelationBuilder {
	b := NewRelationBaseBuilder(related)
	relation := MorphOneRelation{
		Relation: Relation{
			Parent:  self,
			Related: related,
			Type:    RelationMorphOne,
		},
		RelatedIdColumn: relatedIdColumn, ParentKey: selfKey, Builder: b, RelatedTypeColumn: relatedTypeColumn,
	}
	selfModel := GetParsedModel(self)
	selfDirect := reflect.Indirect(reflect.ValueOf(self))
	relation.MorphType = GetMorphMap(selfModel.Name)
	b.Where(relatedIdColumn, "=", selfDirect.Field(selfModel.FieldsByDbName[selfKey].Index).Interface())
	b.WhereNotNull(relatedIdColumn)
	b.Where(relatedTypeColumn, "=", relation.MorphType)

	return &RelationBuilder{Builder: b, Relation: &relation}

}
func (r *MorphOneRelation) AddEagerConstraints(models interface{}) {
	parentParsedModel := GetParsedModel(r.Parent)
	index := parentParsedModel.FieldsByDbName[r.ParentKey].Index
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
	r.Builder.Reset(TYPE_WHERE)
	r.Builder.WhereNotNull(r.RelatedIdColumn)
	r.Builder.Where(r.RelatedTypeColumn, r.MorphType)
	r.Builder.WhereIn(r.RelatedIdColumn, keys)
}
func MatchMorphOne(models interface{}, related interface{}, relation *MorphOneRelation) {
	relatedModelsValue := related.(reflect.Value)
	relatedResults := relatedModelsValue
	relatedModel := GetParsedModel(relation.Related)
	relatedType := reflect.ValueOf(relation.Relation.Related).Elem().Type()
	slice := reflect.MakeSlice(reflect.SliceOf(relatedType), 0, 1)
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(slice))
	groupedResults := reflect.MakeMap(groupedResultsMapType)
	parent := GetParsedModel(relation.Parent)
	isPtr := parent.FieldsByStructName[relation.Relation.Name].FieldType.Kind() == reflect.Ptr
	if !relatedResults.IsValid() || relatedResults.IsNil() {
		return
	}
	if isPtr {
		for i := 0; i < relatedResults.Len(); i++ {
			result := relatedResults.Index(i)
			ownerKeyIndex := relatedModel.FieldsByDbName[relation.RelatedIdColumn].Index
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
			ownerKeyIndex := relatedModel.FieldsByDbName[relation.RelatedIdColumn].Index
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

	modelRelationFieldIndex := parent.FieldsByStructName[relation.Relation.Name].Index
	modelKeyFieldIndex := parent.FieldsByDbName[relation.ParentKey].Index
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
				panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.Name].Name))
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
					panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.Name].Name))
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
