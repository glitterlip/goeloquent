package goeloquent

import (
	"fmt"
	"reflect"
)

// BelongsToRelation for better understanding,rename the parameters. parent prefix represent for current model, related represent for related model
//for example we have user and phone table
//for phone  parentkey is phone table id column,relatedkey is user table id column,parentRelatedKey is phone table user_id column
//the goal is to find phones' user
type BelongsToRelation struct {
	Relation
	ParentRelatedKey string
	RelatedKey       string
	Parent           interface{}
	Builder          *Builder
}

func (m *EloquentModel) BelongsTo(parent interface{}, related interface{}, parentRelatedKey string, relatedKey string) *RelationBuilder {
	b := NewRelationBaseBuilder(related)
	relation := BelongsToRelation{
		Relation: Relation{
			Parent:  parent,
			Related: related,
			Type:    RelationBelongsTo,
		},
		ParentRelatedKey: parentRelatedKey, RelatedKey: relatedKey, Parent: parent, Builder: b,
	}
	parentModel := GetParsedModel(parent)
	parentDirect := reflect.Indirect(reflect.ValueOf(parent))
	//select * from users(related table) where users.id(relatedkey) = phone.user_id(parentRelatedKey)
	b.Where(relation.RelatedKey, parentDirect.Field(parentModel.FieldsByDbName[parentRelatedKey].Index).Interface())

	return &RelationBuilder{Builder: b, Relation: &relation}

}
func (r *BelongsToRelation) AddEagerConstraints(parentModels interface{}) {
	parentParsedModel := GetParsedModel(r.Parent)
	parentRelatedKeyIndex := parentParsedModel.FieldsByDbName[r.ParentRelatedKey].Index
	parentModelSlice := reflect.Indirect(reflect.ValueOf(parentModels))
	var parentModelRelatedKeys []interface{}
	if parentModelSlice.Type().Kind() == reflect.Slice {
		for i := 0; i < parentModelSlice.Len(); i++ {
			parentModel := parentModelSlice.Index(i)
			parentModelRelatedKey := reflect.Indirect(parentModel).Field(parentRelatedKeyIndex).Interface()
			parentModelRelatedKeys = append(parentModelRelatedKeys, parentModelRelatedKey)
		}
	} else if ms, ok := parentModels.(*reflect.Value); ok {
		for i := 0; i < ms.Len(); i++ {
			modelKey := ms.Index(i).Field(parentRelatedKeyIndex).Interface()
			parentModelRelatedKeys = append(parentModelRelatedKeys, modelKey)
		}
	} else {
		model := parentModelSlice
		modelKey := model.Field(parentRelatedKeyIndex).Interface()
		parentModelRelatedKeys = append(parentModelRelatedKeys, modelKey)
	}
	r.Builder.Wheres = nil
	//select * from users(related table) where users.id(relatedkey) in phones user_ids(parentModelRelatedKeys)
	r.Builder.WhereIn(r.RelatedKey, parentModelRelatedKeys)
}

//parentModel.ParentRelatedKey = relatedModel.RelatedKey
func MatchBelongsTo(models interface{}, related interface{}, relation *BelongsToRelation) {
	relatedModels := related.(reflect.Value)
	parsedRelatedModel := GetParsedModel(relation.Related)
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(relation.Relation.Related))
	groupedResults := reflect.MakeMap(groupedResultsMapType)
	parent := GetParsedModel(relation.Parent)
	relationFieldIsPtr := parent.FieldsByStructName[relation.Relation.Name].FieldType.Kind() == reflect.Ptr
	if !relatedModels.IsValid() || relatedModels.IsNil() {
		return
	}

	//create a map[relatedKey]*relatedModel
	for i := 0; i < relatedModels.Len(); i++ {
		result := relatedModels.Index(i)
		relatedKeyIndex := parsedRelatedModel.FieldsByDbName[relation.RelatedKey].Index
		groupKeyStr := fmt.Sprint(result.FieldByIndex([]int{relatedKeyIndex}))
		groupKey := reflect.ValueOf(groupKeyStr)
		groupedResults.SetMapIndex(groupKey, result.Addr())
	}

	targetSlice := reflect.Indirect(reflect.ValueOf(models))
	modelRelationFiledIndex := parent.FieldsByStructName[relation.Relation.Name].Index
	modelKeyFiledIndex := parent.FieldsByDbName[relation.ParentRelatedKey].Index
	if rvP, ok := models.(*reflect.Value); ok {
		for i := 0; i < rvP.Len(); i++ {
			model := rvP.Index(i)
			modelKey := model.Field(modelKeyFiledIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				if relationFieldIsPtr {
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
			if !model.Field(modelRelationFiledIndex).CanSet() {
				panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.Name].Name))
			}
			if relationFieldIsPtr {
				model.Field(modelRelationFiledIndex).Set(value)
			} else {
				model.Field(modelRelationFiledIndex).Set(value.Elem())
			}
		}

	} else {
		//iterate parentmodels find its match relation and set its relation field
		for i := 0; i < targetSlice.Len(); i++ {
			model := targetSlice.Index(i)
			modelKey := model.Field(modelKeyFiledIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				if !model.Field(modelRelationFiledIndex).CanSet() {
					panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.Name].Name))
				}
				if relationFieldIsPtr {
					model.Field(modelRelationFiledIndex).Set(value)
				} else {
					model.Field(modelRelationFiledIndex).Set(value.Elem())
				}
			}
		}
	}
}
