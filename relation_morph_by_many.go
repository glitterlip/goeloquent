package goeloquent

import (
	"fmt"
	"reflect"
)

type MorphByManyRelation struct {
	*Relation
	PivotTable                  string
	PivotRelatedIdColumn        string
	PivotRelatedTypeColumn      string
	PivotSelfColumn             string
	SelfIdColumn                string
	RelatedIdColumn             string
	RelatedModelTypeColumnValue string
}

func (r *MorphByManyRelation) AddEagerConstraints(models interface{}) {
	parentParsedModel := GetParsedModel(r.SelfModel)
	index := parentParsedModel.FieldsByDbName[r.SelfIdColumn].Index
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
	r.Wheres = r.Wheres[1:]
	r.Bindings[TYPE_WHERE] = r.Bindings[TYPE_WHERE][1:]
	r.Builder.WhereIn(r.PivotTable+"."+r.PivotSelfColumn, keys)
}

func (r *MorphByManyRelation) AddConstraints() {
	r.Builder.Where(r.PivotRelatedIdColumn, "=", r.GetSelfKey(r.SelfIdColumn))
	r.Builder.Where(r.PivotTable+"."+r.PivotRelatedTypeColumn, r.RelatedModelTypeColumnValue)
}

func MatchMorphByMany(models interface{}, related interface{}, relation *MorphByManyRelation) {
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
		pivotMap := result.FieldByIndex([]int{relatedModel.EloquentModelFieldIndex, EloquentModelPivotFieldIndex}).Interface().(map[string]interface{})
		groupKey := reflect.ValueOf(pivotMap[OrmPivotAlias+relation.PivotSelfColumn].(string))
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
	modelKeyFieldIndex := parent.FieldsByDbName[relation.SelfIdColumn].Index

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
func (r *MorphByManyRelation) GetRelationExistenceQuery(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {

	if selfQuery.FromTable == relatedQuery.FromTable {
		return r.GetRelationExistenceQueryForSelfJoin(relatedQuery, selfQuery, alias, columns)
	}
	relatedQuery.Join(r.PivotTable, r.PivotTable+"."+r.PivotRelatedIdColumn, "=", GetParsedModel(r.RelatedModel).Table+"."+r.RelatedIdColumn)

	return relatedQuery.Select(Raw(columns)).WhereColumn(GetParsedModel(r.RelatedModel).Table, "=", r.SelfIdColumn).Where(r.PivotTable+"."+r.PivotRelatedTypeColumn, "=", r.RelatedModelTypeColumnValue)

}

func (r *MorphByManyRelation) GetRelationExistenceQueryForSelfJoin(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {
	relatedQuery.Select(Raw(columns))
	tableAlias := relatedQuery.FromTable.(string) + " as " + OrmAggregateAlias
	relatedQuery.From(tableAlias)
	relatedQuery.Join(tableAlias, tableAlias+"."+r.PivotRelatedIdColumn, "=", GetParsedModel(r.RelatedModel).Table+"."+r.RelatedIdColumn).Where(
		tableAlias+"."+r.PivotRelatedTypeColumn, "=", r.RelatedModelTypeColumnValue)
	return relatedQuery.WhereColumn(tableAlias+"."+r.RelatedIdColumn, "=", r.SelfIdColumn)
}
func (r *MorphByManyRelation) GetSelf() *Model {
	return GetParsedModel(r.SelfModel)
}
func (r *MorphByManyRelation) GetRelated() *Model {
	return GetParsedModel(r.RelatedModel)
}
