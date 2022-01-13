package goeloquent

import (
	"fmt"
	"reflect"
)

// HasOneRelation HasOne for better understanding,rename the parameters. parent prefix represent for current model column, related represent for related model column,pivot prefix represent for pivot table
//for example we have user and phone table
//for user model parentkey => user table id column,relatedkey => phone table id column,relatedparentkey => phone table user_id column
type HasOneRelation struct {
	Relation
	RelatedParentKey string
	ParentKey        string
	Builder          *Builder
}

func (m *EloquentModel) HasOne(self interface{}, related interface{}, relatedParentKey, parentKey string) *RelationBuilder {
	b := NewRelationBaseBuilder(related)
	if m.Tx != nil {
		b.Tx = m.Tx
	}
	relation := HasOneRelation{
		Relation{
			Parent:  self,
			Related: related,
			Type:    RelationHasOne,
		},
		relatedParentKey, parentKey, b,
	}
	selfModel := GetParsedModel(self)
	selfDirect := reflect.Indirect(reflect.ValueOf(self))
	b.Where(relatedParentKey, "=", selfDirect.Field(selfModel.FieldsByDbName[parentKey].Index).Interface())
	b.WhereNotNull(relatedParentKey)

	return &RelationBuilder{Builder: b, Relation: &relation}
}
func (r *HasOneRelation) AddEagerConstraints(models interface{}) {
	relatedParsedModel := GetParsedModel(r.Related)
	index := relatedParsedModel.FieldsByDbName[r.ParentKey].Index
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
	r.Builder.Wheres = nil
	r.Builder.WhereNotNull(r.RelatedParentKey)
	r.Builder.WhereIn(r.RelatedParentKey, keys)
}
func MatchHasOne(models interface{}, related interface{}, relation *HasOneRelation) {
	relatedModelsValue := related.(reflect.Value)
	relatedModels := relatedModelsValue
	relatedModel := GetParsedModel(relation.Related)
	parsedModel := GetParsedModel(relation.Parent)
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(relation.Relation.Related))
	groupedResults := reflect.MakeMap(groupedResultsMapType)
	isPtr := parsedModel.FieldsByStructName[relation.Relation.Name].FieldType.Kind() == reflect.Ptr
	if !relatedModels.IsValid() || relatedModels.IsNil() {
		return
	}
	for i := 0; i < relatedModels.Len(); i++ {
		result := relatedModels.Index(i)
		foreignKeyIndex := relatedModel.FieldsByDbName[relation.RelatedParentKey].Index
		groupKey := reflect.ValueOf(fmt.Sprint(result.FieldByIndex([]int{foreignKeyIndex})))
		groupedResults.SetMapIndex(groupKey, result.Addr())
	}

	targetSlice := reflect.Indirect(reflect.ValueOf(models))
	modelRelationFiledIndex := parsedModel.FieldsByStructName[relation.Relation.Name].Index
	modelKeyFiledIndex := parsedModel.FieldsByDbName[relation.ParentKey].Index
	if rvP, ok := models.(*reflect.Value); ok {
		for i := 0; i < rvP.Len(); i++ {
			model := rvP.Index(i)
			modelKey := model.Field(modelKeyFiledIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				if isPtr {
					model.Field(modelRelationFiledIndex).Set(value)
				} else {
					model.Field(modelRelationFiledIndex).Set(value.Elem())
				}
			}

		}
	} else if targetSlice.Type().Kind() != reflect.Slice {
		model := targetSlice
		modelKey := model.Field(modelKeyFiledIndex)
		modelKeyStr := fmt.Sprint(modelKey)
		value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
		if value.IsValid() {
			if isPtr {
				model.Field(modelRelationFiledIndex).Set(value)
			} else {
				model.Field(modelRelationFiledIndex).Set(value.Elem())
			}
		}

	} else {
		for i := 0; i < targetSlice.Len(); i++ {
			model := targetSlice.Index(i)
			modelKey := model.Field(modelKeyFiledIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				if isPtr {
					model.Field(modelRelationFiledIndex).Set(value)
				} else {
					model.Field(modelRelationFiledIndex).Set(value.Elem())
				}
			}
		}
	}
}
