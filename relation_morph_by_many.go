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
	SelfColumn                  string
	RelatedIdColumn             string
	RelatedModelTypeColumnValue string
}

func (r *MorphByManyRelation) AddEagerConstraints(models interface{}) {
	parentParsedModel := GetParsedModel(r.SelfModel)
	index := parentParsedModel.FieldsByDbName[r.SelfColumn].Index
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
	r.Builder.WhereIn(r.PivotTable+"."+r.PivotSelfColumn, keys)
	relatedParsedModel := GetParsedModel(r.RelatedModel)
	r.Builder.Where(r.PivotRelatedTypeColumn, GetMorphMap(relatedParsedModel.Name))
}

func (r *MorphByManyRelation) AddConstraints() {
	selfModel := GetParsedModel(r.SelfModel)
	selfDirect := reflect.Indirect(reflect.ValueOf(r.Relation.SelfModel))
	r.Builder.Where(r.PivotSelfColumn, selfDirect.Field(selfModel.FieldsByDbName[r.SelfColumn].Index).Interface())
	r.Builder.Where(r.PivotRelatedTypeColumn, r.RelatedModelTypeColumnValue)
}

func MatchMorphByMany(models interface{}, related interface{}, relation *MorphByManyRelation) {
	relatedValue := related.(reflect.Value)
	relatedResults := relatedValue
	relatedModel := GetParsedModel(relation.RelatedModel)
	parent := GetParsedModel(relation.SelfModel)
	isPtr := parent.FieldsByStructName[relation.Relation.FieldName].FieldType.Elem().Kind() == reflect.Ptr
	var relatedType reflect.Type

	if isPtr {
		relatedType = reflect.ValueOf(relation.Relation.RelatedModel).Type()
	} else {
		relatedType = reflect.ValueOf(relation.Relation.RelatedModel).Elem().Type()
	}
	slice := reflect.MakeSlice(reflect.SliceOf(relatedType), 0, 1)
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(slice))
	groupedResults := reflect.MakeMap(groupedResultsMapType)
	pivotKey := OrmPivotAlias + relation.PivotSelfColumn
	if !relatedResults.IsValid() || relatedResults.IsNil() {
		return
	}
	for i := 0; i < relatedResults.Len(); i++ {
		result := relatedResults.Index(i)
		pivotMap := result.FieldByIndex([]int{relatedModel.PivotFieldIndex}).Interface().(map[string]interface{})
		groupKey := reflect.ValueOf(pivotMap[pivotKey].(string))
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
			if !model.Field(modelRelationFieldIndex).CanSet() {
				panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.FieldName].Name))
			}
			model.Field(modelRelationFieldIndex).Set(value)
		}
	} else {
		for i := 0; i < targetSlice.Len(); i++ {
			model := targetSlice.Index(i)
			if model.Kind() == reflect.Ptr {
				model = reflect.Indirect(model)
			}
			modelKey := model.Field(modelKeyFieldIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				if !model.Field(modelRelationFieldIndex).CanSet() {
					panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.FieldName].Name))
				}
				model.Field(modelRelationFieldIndex).Set(value)
			}

		}
	}
}
func (r *MorphByManyRelation) GetRelationExistenceQuery(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {

	if selfQuery.FromTable == relatedQuery.FromTable {
		return r.GetRelationExistenceQueryForSelfJoin(relatedQuery, selfQuery, alias, columns)
	}
	relatedQuery.Join(r.PivotTable, r.PivotTable+"."+r.PivotRelatedIdColumn, "=", GetParsedModel(r.RelatedModel).Table+"."+r.RelatedIdColumn)

	return relatedQuery.Select(Raw(columns)).WhereColumn(GetParsedModel(r.RelatedModel).Table, "=", r.SelfColumn).Where(r.PivotTable+"."+r.PivotRelatedTypeColumn, "=", r.RelatedModelTypeColumnValue)

}

func (r *MorphByManyRelation) GetRelationExistenceQueryForSelfJoin(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {
	relatedQuery.Select(Raw(columns))
	tableAlias := relatedQuery.FromTable.(string) + " as " + OrmAggregateAlias
	relatedQuery.From(tableAlias)
	relatedQuery.Join(tableAlias, tableAlias+"."+r.PivotRelatedIdColumn, "=", GetParsedModel(r.RelatedModel).Table+"."+r.RelatedIdColumn).Where(
		tableAlias+"."+r.PivotRelatedTypeColumn, "=", r.RelatedModelTypeColumnValue)
	return relatedQuery.WhereColumn(tableAlias+"."+r.RelatedIdColumn, "=", r.SelfColumn)
}
func (r *MorphByManyRelation) GetSelf() *Model {
	return GetParsedModel(r.SelfModel)
}
func (r *MorphByManyRelation) GetRelated() *Model {
	return GetParsedModel(r.RelatedModel)
}
