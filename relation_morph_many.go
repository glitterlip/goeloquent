package goeloquent

import (
	"fmt"
	"reflect"
)

type MorphManyRelation struct {
	*Relation
	RelatedModelIdColumn        string
	RelatedModelTypeColumn      string
	SelfColumn                  string
	RelatedModelTypeColumnValue string
}

func (r *MorphManyRelation) AddEagerConstraints(models interface{}) {
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
	//remove first where clause to simulate the Relation::noConstraints function in laravel
	r.Wheres = r.Wheres[1:]
	r.Bindings[TYPE_WHERE] = r.Bindings[TYPE_WHERE][1:]
	r.Builder.WhereNotNull(r.RelatedModelIdColumn)
	r.Builder.WhereIn(r.RelatedModelIdColumn, keys)
}
func (r *MorphManyRelation) AddConstraints() {
	r.Builder.Where(r.RelatedModelIdColumn, "=", r.GetSelfKey(r.SelfColumn))
	r.Builder.Where(r.RelatedModelTypeColumn, "=", r.RelatedModelTypeColumnValue)
}
func MatchMorphMany(models interface{}, related interface{}, relation *MorphManyRelation) {
	relatedModelsValue := related.(reflect.Value)
	relatedModels := relatedModelsValue
	relatedModel := GetParsedModel(relation.RelatedModel)
	relatedType := reflect.ValueOf(relation.Relation.RelatedModel).Elem().Type()

	parent := GetParsedModel(relation.SelfModel)
	relationFieldIsPtr := parent.FieldsByStructName[relation.Relation.FieldName].FieldType.Kind() == reflect.Ptr
	var sliceEleIsptr bool
	if relationFieldIsPtr {
		sliceEleIsptr = parent.FieldsByStructName[relation.Relation.FieldName].FieldType.Elem().Elem().Kind() == reflect.Ptr
	} else {
		sliceEleIsptr = parent.FieldsByStructName[relation.Relation.FieldName].FieldType.Elem().Kind() == reflect.Ptr
	}
	var slice reflect.Value
	if sliceEleIsptr {
		slice = reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(relatedType)), 0, 1)
	} else {
		slice = reflect.MakeSlice(reflect.SliceOf(relatedType), 0, 1)
	}
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(slice))
	groupedResults := reflect.MakeMap(groupedResultsMapType)
	if !relatedModels.IsValid() || relatedModels.IsNil() {
		return
	}
	for i := 0; i < relatedModels.Len(); i++ {
		result := relatedModels.Index(i)
		ownerKeyIndex := relatedModel.FieldsByDbName[relation.RelatedModelIdColumn].Index
		groupKey := reflect.ValueOf(fmt.Sprint(result.FieldByIndex([]int{ownerKeyIndex})))
		existed := groupedResults.MapIndex(groupKey)
		if !existed.IsValid() {
			existed = reflect.New(slice.Type())
		} else {
			//after initilized and store in map and then get from the map,slice is a reflect.value(struct),we need to convert it
			existed = existed.Interface().(reflect.Value)
		}
		ptr := reflect.New(slice.Type())
		if sliceEleIsptr {
			v := reflect.Append(existed.Elem(), result.Addr())
			ptr.Elem().Set(v)
		} else {
			ptr.Elem().Set(reflect.Append(existed.Elem(), result))

		}
		groupedResults.SetMapIndex(groupKey, reflect.ValueOf(ptr))
	}

	targetSlice := reflect.Indirect(reflect.ValueOf(models))

	modelRelationFieldIndex := parent.FieldsByStructName[relation.Relation.FieldName].Index
	modelKeyFieldIndex := parent.FieldsByDbName[relation.SelfColumn].Index

	if rvP, ok := models.(*reflect.Value); ok {
		for i := 0; i < rvP.Len(); i++ {
			e := rvP.Index(i)
			modelKey := e.Field(modelKeyFieldIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				if relationFieldIsPtr {
					e.Field(modelRelationFieldIndex).Set(value)
				} else {
					e.Field(modelRelationFieldIndex).Set(value.Elem())
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
				panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.FieldName].Name))
			}
			if relationFieldIsPtr {
				model.Field(modelRelationFieldIndex).Set(value)
			} else {
				model.Field(modelRelationFieldIndex).Set(value.Elem())
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
func (r *MorphManyRelation) GetRelationExistenceQuery(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {

	if relatedQuery.FromTable.(string) == selfQuery.FromTable.(string) {
		return r.GetRelationExistenceQueryForSelfRelation(relatedQuery, selfQuery, alias, columns)
	}
	relatedParsed := GetParsedModel(r.Relation.RelatedModel)
	selfParsed := GetParsedModel(r.Relation.SelfModel)
	return relatedQuery.Select(Raw(columns)).WhereColumn(selfParsed.Table+"."+r.SelfColumn, "=", relatedParsed.Table+"."+r.RelatedModelIdColumn).Where(relatedParsed.Table+"."+r.RelatedModelTypeColumn, "=", r.RelatedModelTypeColumnValue)

}

func (r *MorphManyRelation) GetRelationExistenceQueryForSelfRelation(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {
	relatedParsed := GetParsedModel(r.Relation.RelatedModel)
	selfParsed := GetParsedModel(r.Relation.SelfModel)
	relatedQuery.From(relatedQuery.FromTable.(string) + " as " + alias)
	relatedQuery.Select(Raw(columns)).WhereColumn(r.SelfColumn, "=", r.RelatedModelIdColumn).Where(r.RelatedModelTypeColumn, "=", r.RelatedModelTypeColumnValue)
	return relatedQuery.Select(Raw(columns)).WhereColumn(selfParsed.Table+"."+r.SelfColumn, "=", relatedParsed.Table+"."+r.RelatedModelIdColumn).Where(relatedParsed.Table+"."+r.RelatedModelTypeColumn, "=", r.RelatedModelTypeColumnValue)
}
func (r *MorphManyRelation) GetSelf() *Model {
	return GetParsedModel(r.SelfModel)
}
func (r *MorphManyRelation) GetRelated() *Model {
	return GetParsedModel(r.RelatedModel)
}
