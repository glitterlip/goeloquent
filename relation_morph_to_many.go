package goeloquent

import (
	"fmt"
	"reflect"
)

type MorphToManyRelation struct {
	*Relation
	PivotTable               string
	PivotSelfIdColumn        string
	PivotSelfTypeColumn      string
	PivotRelatedIdColumn     string
	SelfIdColumn             string
	RelatedIdColumn          string
	SelfModelTypeColumnValue string
}

func (r *MorphToManyRelation) AddEagerConstraints(models interface{}) {
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
	//remove first where clause to simulate the Relation::noConstraints function in laravel
	r.Wheres = r.Wheres[1:]
	r.Bindings[TYPE_WHERE] = r.Bindings[TYPE_WHERE][1:]
	r.Builder.WhereIn(r.PivotTable+"."+r.PivotSelfIdColumn, keys)

}
func (r *MorphToManyRelation) AddConstraints() {
	selfModel := GetParsedModel(r.SelfModel)
	selfDirect := reflect.Indirect(reflect.ValueOf(r.SelfModel))
	r.Builder.Where(r.PivotSelfIdColumn, selfDirect.Field(selfModel.FieldsByDbName[r.SelfIdColumn].Index).Interface())
	r.Builder.Where(r.PivotSelfTypeColumn, r.SelfModelTypeColumnValue)
}
func MatchMorphToMany(selfModels interface{}, related reflect.Value, relation *MorphToManyRelation) {
	relatedResults := related
	relatedParsedModel := GetParsedModel(relation.RelatedModel)
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
	pivotKey := OrmPivotAlias + relation.PivotSelfIdColumn
	if !relatedResults.IsValid() || relatedResults.IsNil() {
		return
	}
	for i := 0; i < relatedResults.Len(); i++ {
		result := relatedResults.Index(i)
		pivotMap := result.FieldByIndex([]int{relatedParsedModel.EloquentModelFieldIndex, relatedParsedModel.PivotFieldIndex}).Interface().(map[string]interface{})
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

	targetSlice := reflect.Indirect(reflect.ValueOf(selfModels))

	modelRelationFieldIndex := selfParsedModel.FieldsByStructName[relation.Relation.FieldName].Index
	modelKeyFieldIndex := selfParsedModel.FieldsByDbName[relation.SelfIdColumn].Index

	if rvP, ok := selfModels.(*reflect.Value); ok {
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
				panic(fmt.Sprintf("model: %s field: %s cant be set", selfParsedModel.Name, selfParsedModel.FieldsByStructName[relation.Relation.FieldName].Name))
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
					panic(fmt.Sprintf("model: %s field: %s cant be set", selfParsedModel.Name, selfParsedModel.FieldsByStructName[relation.Relation.FieldName].Name))
				}
				model.Field(modelRelationFieldIndex).Set(value)
			}

		}
	}
}
func (r *MorphToManyRelation) GetRelationExistenceQuery(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {

	if selfQuery.FromTable == relatedQuery.FromTable {
		return r.GetRelationExistenceQueryForSelfJoin(relatedQuery, selfQuery, alias, columns)
	}
	relatedParsed := GetParsedModel(r.Relation.RelatedModel)
	selfParsed := GetParsedModel(r.Relation.SelfModel)
	return relatedQuery.Select(Raw(columns)).
		WhereColumn(relatedParsed.Table+"."+r.RelatedIdColumn, "=", selfParsed.Table+"."+r.SelfIdColumn).
		Where(r.PivotTable+"."+r.PivotSelfTypeColumn, "=", r.SelfModelTypeColumnValue)

}

func (r *MorphToManyRelation) GetRelationExistenceQueryForSelfJoin(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {
	relatedParsed := GetParsedModel(r.Relation.RelatedModel)
	selfParsed := GetParsedModel(r.Relation.SelfModel)
	relatedQuery.Select(Raw(columns))
	tableAlias := relatedQuery.FromTable.(string) + " as " + OrmAggregateAlias
	relatedQuery.From(tableAlias)
	return relatedQuery.WhereColumn(relatedParsed.Table+"."+r.RelatedIdColumn, "=", selfParsed.Table+"."+r.SelfIdColumn)
}
func (r *MorphToManyRelation) GetSelf() *Model {
	return GetParsedModel(r.SelfModel)
}
func (r *MorphToManyRelation) GetRelated() *Model {
	return GetParsedModel(r.RelatedModel)
}
