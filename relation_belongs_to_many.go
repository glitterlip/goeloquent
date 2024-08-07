package goeloquent

import (
	"fmt"
	"reflect"
)

type BelongsToManyRelation struct {
	*Relation
	PivotTable         string
	PivotSelfColumn    string
	PivotRelatedColumn string
	SelfColumn         string
	RelatedColumn      string
	PivotColumns       []string //extra columns to select in pivot table
	PivotWheres        []Where  //extra where conditions in pivot table
	WithTimestamps     bool
}

func (r *BelongsToManyRelation) AddConstraints() {
	r.Builder.Where(r.PivotTable+"."+r.PivotSelfColumn, "=", r.GetSelfKey(r.SelfColumn))
}

func (r *BelongsToManyRelation) AddEagerConstraints(selfModels interface{}) {
	selfParsedModel := GetParsedModel(r.SelfModel)
	index := selfParsedModel.FieldsByDbName[r.SelfColumn].Index
	modelSlice := reflect.Indirect(reflect.ValueOf(selfModels))
	var selfKeys []interface{}
	if modelSlice.Type().Kind() == reflect.Slice {
		for i := 0; i < modelSlice.Len(); i++ {
			model := modelSlice.Index(i)
			modelKey := reflect.Indirect(model).Field(index).Interface()
			selfKeys = append(selfKeys, modelKey)
		}
	} else if ms, ok := selfModels.(*reflect.Value); ok {
		for i := 0; i < ms.Len(); i++ {
			modelKey := ms.Index(i).Field(index).Interface()
			selfKeys = append(selfKeys, modelKey)
		}
	} else {
		model := modelSlice
		modelKey := model.Field(index).Interface()
		selfKeys = append(selfKeys, modelKey)
	}
	//remove first where clause to simulate the Relation::noConstraints function in laravel
	r.Wheres = r.Wheres[1:]
	r.Bindings[TYPE_WHERE] = r.Bindings[TYPE_WHERE][1:]
	r.Builder.WhereIn(r.PivotTable+"."+r.PivotSelfColumn, selfKeys)
}
func MatchBelongsToMany(selfModels interface{}, relatedModels reflect.Value, relation *BelongsToManyRelation) {
	parsedRelatedModel := GetParsedModel(relation.RelatedModel)
	selfParsedModel := GetParsedModel(relation.SelfModel)

	isPtr := selfParsedModel.FieldsByStructName[relation.Relation.FieldName].FieldType.Elem().Kind() == reflect.Ptr
	var relatedType reflect.Type
	if isPtr {
		relatedType = reflect.ValueOf(relation.Relation.RelatedModel).Type()
	} else {
		relatedType = reflect.ValueOf(relation.Relation.RelatedModel).Elem().Type()
	}
	slice := reflect.MakeSlice(reflect.SliceOf(relatedType), 0, 1)
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(slice))
	groupedResults := reflect.MakeMap(groupedResultsMapType)
	pivotSelfKey := OrmPivotAlias + relation.PivotSelfColumn
	if !relatedModels.IsValid() || relatedModels.IsNil() {
		return
	}
	for i := 0; i < relatedModels.Len(); i++ {
		relatedModel := relatedModels.Index(i)
		eloquentModelPtr := relatedModel.Field(parsedRelatedModel.EloquentModelFieldIndex)

		pivotMap := eloquentModelPtr.Elem().Field(EloquentModelPivotFieldIndex).Interface().(map[string]interface{})
		groupKey := reflect.ValueOf(pivotMap[pivotSelfKey].(string))
		existed := groupedResults.MapIndex(groupKey)
		if !existed.IsValid() {
			existed = reflect.New(slice.Type()).Elem()
		} else {
			existed = existed.Interface().(reflect.Value)
		}
		ptr := reflect.New(slice.Type())
		if isPtr {
			v := reflect.Append(existed, relatedModel.Addr())
			ptr.Elem().Set(v)
		} else {
			v := reflect.Append(existed, relatedModel)
			ptr.Elem().Set(v)
		}
		groupedResults.SetMapIndex(groupKey, reflect.ValueOf(ptr.Elem()))
	}

	targetSlice := reflect.Indirect(reflect.ValueOf(selfModels))

	modelRelationFieldIndex := selfParsedModel.FieldsByStructName[relation.Relation.FieldName].Index
	modelKeyFieldIndex := selfParsedModel.FieldsByDbName[relation.SelfColumn].Index

	if rvP, ok := selfModels.(*reflect.Value); ok {
		for i := 0; i < rvP.Len(); i++ {
			e := rvP.Index(i)
			modelKey := e.Field(modelKeyFieldIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				e.Field(modelRelationFieldIndex).Set(value)
			}

		}
	} else if targetSlice.Type().Kind() != reflect.Slice {
		model := targetSlice
		modelKey := model.Field(modelKeyFieldIndex)
		modelKeyStr := fmt.Sprint(modelKey.Interface())
		value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
		if value.IsValid() {
			value = value.Interface().(reflect.Value)
			if !model.Field(modelRelationFieldIndex).CanSet() {
				panic(fmt.Sprintf("model: %s field: %s cant be set", selfParsedModel.Name, selfParsedModel.FieldsByStructName[relation.Relation.FieldName].Name))
			}
			model.Field(modelRelationFieldIndex).Set(value)
		}
	} else {
		for i := 0; i < targetSlice.Len(); i++ {
			model := targetSlice.Index(i)
			modelKey := model.Field(modelKeyFieldIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				if !model.Field(modelRelationFieldIndex).CanSet() {
					panic(fmt.Sprintf("model: %s field: %s cant be set", selfParsedModel.Name, selfParsedModel.FieldsByStructName[relation.Relation.FieldName].Name))
				}
				model.Field(modelRelationFieldIndex).Set(value)
			}

		}
	}
}
func (r *BelongsToManyRelation) GetRelationExistenceQuery(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {

	if selfQuery.FromTable == relatedQuery.FromTable {
		return r.GetRelationExistenceQueryForSelfJoin(relatedQuery, selfQuery, alias, columns)
	}

	return relatedQuery.Select(Raw(columns)).WhereColumn(GetParsedModel(r.SelfModel).Table+"."+r.SelfColumn, "=", r.PivotTable+"."+r.PivotSelfColumn)

}

func (r *BelongsToManyRelation) GetRelationExistenceQueryForSelfJoin(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {
	relatedQuery.Select(Raw(columns))
	tableAlias := relatedQuery.FromTable.(string) + " as " + OrmAggregateAlias
	relatedQuery.From(tableAlias)
	return relatedQuery.WhereColumn(GetParsedModel(r.SelfModel).Table+"."+r.SelfColumn, "=", r.PivotTable+"."+r.PivotSelfColumn)
}
func (r *BelongsToManyRelation) GetSelf() *Model {
	return GetParsedModel(r.SelfModel)
}
func (r *BelongsToManyRelation) GetRelated() *Model {
	return GetParsedModel(r.RelatedModel)
}
