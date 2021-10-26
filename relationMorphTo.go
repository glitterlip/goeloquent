package goeloquent

import (
	"fmt"
	"reflect"
)

type MorphToRelation struct {
	Relation
	ParentRelatedKey  string
	RelatedKey        string
	Builder           *Builder
	ParentRelatedType string
	Groups            map[string][]interface{}
}
type MorphToResult struct {
	MorphToResultP reflect.Value
}

func (m *MorphToResult) GetMorph() {

}
func (m *EloquentModel) MorphTo(self interface{}, parentRelatedKey, relatedKey, parentRelatedType string) *RelationBuilder {

	builder := NewRelationBaseBuilder(nil)
	relation := MorphToRelation{
		Relation: Relation{
			Parent:  self,
			Related: nil,
			Type:    RelationMorphTo,
		},
		ParentRelatedKey: parentRelatedKey, RelatedKey: relatedKey, Builder: builder, ParentRelatedType: parentRelatedType,
	}
	parent := GetParsedModel(self)
	relatedTypeField := parent.FieldsByDbName[parentRelatedType]
	t := reflect.Indirect(reflect.ValueOf(self)).Field(relatedTypeField.Index).Interface().(string)
	if t != "" {
		m := Eloquent.MorphDBMap[t]
		builder.SetModel(m.Type())
		builder.Where(relatedKey, reflect.Indirect(reflect.ValueOf(self)).Field(parent.FieldsByDbName[parentRelatedKey].Index).Interface())
	}

	return &RelationBuilder{Builder: builder, Relation: &relation}

}

func (r *MorphToRelation) AddEagerConstraints(models interface{}) {
	r.Builder.Wheres = nil
	modelSlice := reflect.Indirect(reflect.ValueOf(models))
	groups := make(map[string][]interface{})
	if modelSlice.Type().Kind() == reflect.Slice {
		parsed := GetParsedModel(modelSlice.Type().Elem())
		typeColumnIndex := parsed.FieldsByDbName[r.ParentRelatedType].Index
		idColumnIndex := parsed.FieldsByDbName[r.ParentRelatedKey].Index
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
	} else {
		model := modelSlice
		parsed := GetParsedModel(modelSlice.Type())
		typeColumnIndex := parsed.FieldsByDbName[r.ParentRelatedType].Index
		idColumnIndex := parsed.FieldsByDbName[r.ParentRelatedKey].Index
		t := model.Field(typeColumnIndex).Interface().(string)
		id := model.Field(idColumnIndex).Interface()
		groups[t] = []interface{}{id}
	}
	r.Groups = groups
}
func MatchMorphTo(models interface{}, releated interface{}, relation *MorphToRelation) {
	releatedValue := releated.(map[string]reflect.Value)
	releatedResults := releatedValue
	//groupedResults := make(map[string]map[string]interface{})
	morphMapResults := make(map[string]map[string]reflect.Value)
	parent := GetParsedModel(relation.Parent)
	isPtr := parent.FieldsByStructName[relation.Relation.Name].FieldType.Kind() == reflect.Ptr
	for morphType, relatedSliceValue := range releatedResults {
		groupedResults := make(map[string]reflect.Value)
		for i := 0; i < relatedSliceValue.Len(); i++ {
			morphModelValue := relatedSliceValue.Index(i)
			parsedMorphModel := GetParsedModel(Eloquent.MorphDBMap[morphType].Type())
			ownerKeyIndex := parsedMorphModel.FieldsByDbName[relation.RelatedKey].Index
			idIndex := morphModelValue.Field(ownerKeyIndex)
			idStr := fmt.Sprint(idIndex.Interface())
			if isPtr {
				groupedResults[idStr] = morphModelValue.Addr()
			} else {
				groupedResults[idStr] = morphModelValue

			}
		}
		morphMapResults[morphType] = groupedResults
	}

	targetSlice := reflect.Indirect(reflect.ValueOf(models))

	modelRelationFiledIndex := parent.FieldsByStructName[relation.Relation.Name].Index
	modelMorphIdFiledIndex := parent.FieldsByDbName[relation.ParentRelatedKey].Index
	modelMorphTypeFiledIndex := parent.FieldsByDbName[relation.ParentRelatedType].Index

	if targetSlice.Type().Kind() != reflect.Slice {
		model := targetSlice
		modelMorphType := model.Field(modelMorphTypeFiledIndex)
		modelMorphId := model.Field(modelMorphIdFiledIndex)
		modelKeyStr := fmt.Sprint(modelMorphId)
		groupKeyStr := fmt.Sprint(modelMorphType)
		groupedResults, ok := morphMapResults[groupKeyStr]
		if ok {
			morphedModel, ok := groupedResults[modelKeyStr]
			if ok {
				if !model.Field(modelRelationFiledIndex).CanSet() {
					panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.Name].Name))
				}
				if isPtr {
					model.Field(modelRelationFiledIndex).Set(morphedModel.Addr())
				} else {
					model.Field(modelRelationFiledIndex).Set(morphedModel)
				}
			}
		}

	} else {
		for i := 0; i < targetSlice.Len(); i++ {
			model := targetSlice.Index(i)
			modelMorphType := model.Field(modelMorphTypeFiledIndex)
			modelMorphId := model.Field(modelMorphIdFiledIndex)
			modelKeyStr := fmt.Sprint(modelMorphId)
			groupKeyStr := fmt.Sprint(modelMorphType)
			morphedResults, ok := morphMapResults[groupKeyStr]
			if ok {
				morphedModel, ok := morphedResults[modelKeyStr]
				if ok {
					if !model.Field(modelRelationFiledIndex).CanSet() {
						panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.Name].Name))
					}
					if isPtr {
						model.Field(modelRelationFiledIndex).Set(morphedModel.Addr())
					} else {
						model.Field(modelRelationFiledIndex).Set(morphedModel)
					}
				}
			}
		}
	}
}
