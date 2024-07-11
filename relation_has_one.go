package goeloquent

import (
	"fmt"
	"reflect"
)

type HasOneRelation struct {
	*Relation
	SelfColumn    string
	RelatedColumn string
}

func (r *HasOneRelation) AddEagerConstraints(models interface{}) {
	relatedParsedModel := GetParsedModel(r.RelatedModel)
	index := relatedParsedModel.FieldsByDbName[r.SelfColumn].Index
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
	r.Builder.WhereIn(relatedParsedModel.Table+"."+r.RelatedColumn, keys)
}
func (r *HasOneRelation) AddConstraints() {
	relatedParsedModel := GetParsedModel(r.RelatedModel)
	r.Builder.Where(relatedParsedModel.Table+"."+r.RelatedColumn, "=", r.GetSelfKey(r.SelfColumn))
	r.Builder.WhereNotNull(relatedParsedModel.Table + "." + r.RelatedColumn)
}

/*
MatchHasOne match the has one relation
selfModels: the models that has the relation,usually a pointer of  slice of models,or reflect.Value if it's a nested relation
related: the related models ,reflect.Value of slice
*/
func MatchHasOne(selfModels interface{}, relatedModelsValue reflect.Value, relation *HasOneRelation) {
	relatedModel := GetParsedModel(relation.RelatedModel)
	selfModel := GetParsedModel(relation.SelfModel)

	//make map[string]relatedModel , key is RelatedColumn,value is relatedModel pointer
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(relation.RelatedModel))
	groupedResults := reflect.MakeMap(groupedResultsMapType)

	//indicate if the relation field on self model is a pointer
	isPtr := selfModel.FieldsByStructName[relation.FieldName].FieldType.Kind() == reflect.Ptr

	if !relatedModelsValue.IsValid() || relatedModelsValue.IsNil() {
		return
	}

	//RelatedColumn in related model that is related to self model's self column
	relatedColumn := relatedModel.FieldsByDbName[relation.RelatedColumn]

	for i := 0; i < relatedModelsValue.Len(); i++ {
		related := relatedModelsValue.Index(i)
		groupKey := reflect.ValueOf(fmt.Sprint(related.FieldByIndex([]int{relatedColumn.Index})))
		groupedResults.SetMapIndex(groupKey, related.Addr())
	}

	targetSlice := reflect.ValueOf(selfModels).Elem()

	selfColumn := selfModel.FieldsByDbName[relation.SelfColumn]
	selfColumnIndex := selfColumn.Index
	selfRelationField := selfModel.FieldsByStructName[relation.FieldName]

	//self models is reflect.Value
	if rvP, ok := selfModels.(*reflect.Value); ok {
		for i := 0; i < rvP.Len(); i++ {
			model := rvP.Index(i)
			modelKey := model.Field(selfColumnIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				if isPtr {
					model.Field(selfRelationField.Index).Set(value)
				} else {
					model.Field(selfRelationField.Index).Set(value.Elem())
				}
			}

		}
	} else if targetSlice.Type().Kind() != reflect.Slice {
		model := targetSlice
		modelKey := model.Field(selfColumnIndex)
		modelKeyStr := fmt.Sprint(modelKey)
		value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
		if value.IsValid() {
			if isPtr {
				model.Field(selfRelationField.Index).Set(value)
			} else {
				model.Field(selfRelationField.Index).Set(value.Elem())
			}
		}

	} else {
		//selfModels is slice
		for i := 0; i < targetSlice.Len(); i++ {
			model := targetSlice.Index(i)
			modelKey := model.Field(selfColumnIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				if isPtr {
					model.Field(selfRelationField.Index).Set(value)
				} else {
					model.Field(selfRelationField.Index).Set(value.Elem())
				}
			}
		}
	}
}
