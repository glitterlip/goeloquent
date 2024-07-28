package goeloquent

import (
	"fmt"
	"reflect"
)

type MorphToRelation struct {
	*Relation
	SelfRelatedIdColumn   string
	SelfRelatedTypeColumn string
	RelatedModelIdColumn  string
	Groups                map[string][]interface{} //filed name => []interface{}
}
type MorphToResult struct {
	MorphToResultP reflect.Value
}

func (r *MorphToRelation) AddEagerConstraints(models interface{}) {

	modelSlice := reflect.Indirect(reflect.ValueOf(models))
	groups := make(map[string][]interface{})
	if modelSlice.Type().Kind() == reflect.Slice {
		parsedSelfModel := GetParsedModel(modelSlice.Type().Elem())
		typeColumnIndex := parsedSelfModel.FieldsByDbName[r.SelfRelatedTypeColumn].Index
		idColumnIndex := parsedSelfModel.FieldsByDbName[r.SelfRelatedIdColumn].Index
		for i := 0; i < modelSlice.Len(); i++ {
			model := modelSlice.Index(i)
			t := reflect.Indirect(model).Field(typeColumnIndex).Interface().(string)
			id := reflect.Indirect(model).Field(idColumnIndex).Interface()
			if keys, ok := groups[t]; ok {
				keys = append(keys, id)
				groups[t] = keys
			} else {
				groups[t] = []interface{}{id}
			}
		}
	} else if ms, ok := models.(*reflect.Value); ok {
		parsed := GetParsedModel(ms.Type().Elem())
		typeColumnIndex := parsed.FieldsByDbName[r.SelfRelatedTypeColumn].Index
		idColumnIndex := parsed.FieldsByDbName[r.SelfRelatedIdColumn].Index
		for i := 0; i < ms.Len(); i++ {
			model := ms.Index(i)
			t := reflect.Indirect(model).Field(typeColumnIndex).Interface().(string)
			id := reflect.Indirect(model).Field(idColumnIndex).Interface()
			if keys, ok := groups[t]; ok {
				keys = append(keys, id)
				groups[t] = keys
			} else {
				groups[t] = []interface{}{id}
			}
		}
	} else {
		model := modelSlice
		parsed := GetParsedModel(modelSlice.Type())
		typeColumnIndex := parsed.FieldsByDbName[r.SelfRelatedTypeColumn].Index
		idColumnIndex := parsed.FieldsByDbName[r.SelfRelatedIdColumn].Index
		t := model.Field(typeColumnIndex).Interface().(string)
		id := model.Field(idColumnIndex).Interface()
		groups[t] = []interface{}{id}
	}
	r.Groups = groups

}
func (r *MorphToRelation) AddConstraints() {
	r.Builder.Where(r.RelatedModelIdColumn, "=", r.GetSelfKey(r.SelfRelatedIdColumn))
	if key, ok := r.GetSelfKey(r.SelfRelatedTypeColumn).(string); ok && len(key) > 0 {
		modelPointer := GetMorphDBMap(r.GetSelfKey(r.SelfRelatedTypeColumn).(string))
		r.Relation.RelatedModel = modelPointer
		r.EloquentBuilder.SetModel(modelPointer)
	}

}
func MatchMorphTo(selfModels interface{}, releated reflect.Value, relation *MorphToRelation) {
	releatedResults := releated.Interface().(map[string]reflect.Value)
	//releatedResults := releated
	//map[string][]reflect.Value{}
	morphMapResults := make(map[string]map[string]reflect.Value) //map[relatedTypeString]map[relatedIdString]reflect.Value
	parsedSelfModel := GetParsedModel(relation.SelfModel)
	isPtr := parsedSelfModel.FieldsByStructName[relation.Relation.FieldName].FieldType.Kind() == reflect.Ptr

	for morphType, relatedSliceValue := range releatedResults {
		groupedResults := make(map[string]reflect.Value)
		parsedMorphModel := GetParsedModel(GetMorphDBMap(morphType).Type())

		for i := 0; i < relatedSliceValue.Len(); i++ {
			morphModelValue := relatedSliceValue.Index(i)
			ownerKeyIndex := parsedMorphModel.FieldsByDbName[relation.RelatedModelIdColumn].Index
			idIndex := morphModelValue.Field(ownerKeyIndex)
			idStr := fmt.Sprint(idIndex.Interface())

			groupedResults[idStr] = morphModelValue
		}
		morphMapResults[morphType] = groupedResults
	}

	targetSlice := reflect.Indirect(reflect.ValueOf(selfModels))
	rv, ok := selfModels.(*reflect.Value)

	modelRelationFieldIndex := parsedSelfModel.FieldsByStructName[relation.Relation.FieldName].Index
	modelMorphIdFieldIndex := parsedSelfModel.FieldsByDbName[relation.SelfRelatedIdColumn].Index
	modelMorphTypeFieldIndex := parsedSelfModel.FieldsByDbName[relation.SelfRelatedTypeColumn].Index

	if targetSlice.Type().Kind() != reflect.Slice && !ok {
		model := targetSlice
		modelMorphType := model.Field(modelMorphTypeFieldIndex)
		modelMorphId := model.Field(modelMorphIdFieldIndex)
		modelKeyStr := fmt.Sprint(modelMorphId)
		groupKeyStr := fmt.Sprint(modelMorphType)
		groupedResults, ok := morphMapResults[groupKeyStr]
		if ok {
			morphedModel, ok := groupedResults[modelKeyStr]
			if ok {
				if !model.Field(modelRelationFieldIndex).CanSet() {
					panic(fmt.Sprintf("model: %s field: %s cant be set", parsedSelfModel.Name, parsedSelfModel.FieldsByStructName[relation.Relation.FieldName].Name))
				}
				model.Field(modelRelationFieldIndex).Set(morphedModel)
			}
		}

	} else {
		if ok {
			targetSlice = *rv
		}
		for i := 0; i < targetSlice.Len(); i++ {
			model := targetSlice.Index(i)
			modelMorphType := model.Field(modelMorphTypeFieldIndex)
			modelMorphId := model.Field(modelMorphIdFieldIndex)
			modelKeyStr := fmt.Sprint(modelMorphId)
			groupKeyStr := fmt.Sprint(modelMorphType)
			morphedResults, ok := morphMapResults[groupKeyStr]
			if ok {
				morphedModel, ok := morphedResults[modelKeyStr]
				if ok {
					if !model.Field(modelRelationFieldIndex).CanSet() {
						panic(fmt.Sprintf("model: %s field: %s cant be set", parsedSelfModel.Name, parsedSelfModel.FieldsByStructName[relation.Relation.FieldName].Name))
					}
					if isPtr {
						model.Field(modelRelationFieldIndex).Set(morphedModel.Addr())
					} else {
						model.Field(modelRelationFieldIndex).Set(morphedModel)
					}
				}
			}
		}
	}
}
func (r *MorphToRelation) GetRelationExistenceQuery(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {

	if selfQuery.FromTable == relatedQuery.FromTable {
		return r.GetRelationExistenceQueryForSelfRelation(relatedQuery, selfQuery, alias, columns)
	}
	return relatedQuery.Select(Raw(columns)).WhereColumn(GetParsedModel(r.RelatedModel).Table, "=", r.SelfRelatedIdColumn)

}

func (r *MorphToRelation) GetRelationExistenceQueryForSelfRelation(relatedQuery *EloquentBuilder, selfQuery *EloquentBuilder, alias string, columns string) *EloquentBuilder {
	tableAlias := relatedQuery.FromTable.(string) + " as " + alias
	relatedQuery.Select(Raw(columns)).From(tableAlias)

	return relatedQuery.WhereColumn(tableAlias+"."+r.RelatedModelIdColumn, "=", r.SelfRelatedIdColumn)
}
func (r *MorphToRelation) GetSelf() *Model {
	return GetParsedModel(r.SelfModel)
}
func (r *MorphToRelation) GetRelated() *Model {
	return GetParsedModel(r.RelatedModel)
}
