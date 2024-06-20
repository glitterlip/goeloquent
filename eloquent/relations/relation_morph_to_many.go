package relations

import (
	"fmt"
	"github.com/glitterlip/goeloquent/eloquent"
	"reflect"
)

type MorphToManyRelation struct {
	eloquent.Relation
	PivotTable                  string
	PivotRelatedIdColumn        string
	PivotRelatedTypeColumn      string
	PivotSelfColumn             string
	SelfColumn                  string
	RelatedIdColumn             string
	RelatedModelTypeColumnValue string
}

func (r *MorphToManyRelation) AddEagerConstraints(models interface{}) {
	parentParsedModel := eloquent.GetParsedModel(r.SelfModel)
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
	r.Builder.Where(r.PivotRelatedTypeColumn, r.RelatedModelTypeColumnValue)

}
func (r *MorphToManyRelation) AddConstraints() {
	selfModel := eloquent.GetParsedModel(r.SelfModel)
	selfDirect := reflect.Indirect(reflect.ValueOf(r.SelfModel))
	r.Builder.Where(r.PivotSelfColumn, selfDirect.Field(selfModel.FieldsByDbName[r.SelfColumn].Index).Interface())
	r.Builder.Where(r.PivotRelatedTypeColumn, r.RelatedModelTypeColumnValue)
}
func MatchMorphToMany(models interface{}, related interface{}, relation *MorphToManyRelation) {
	relatedValue := related.(reflect.Value)
	relatedResults := relatedValue
	relatedModel := eloquent.GetParsedModel(relation.RelatedModel)
	parent := eloquent.GetParsedModel(relation.SelfModel)
	isPtr := parent.FieldsByStructName[relation.Relation.Name].FieldType.Elem().Kind() == reflect.Ptr
	var relatedType reflect.Type

	if isPtr {
		relatedType = reflect.ValueOf(relation.Relation.RelatedModel).Type()
	} else {
		relatedType = reflect.ValueOf(relation.Relation.RelatedModel).Elem().Type()
	}
	slice := reflect.MakeSlice(reflect.SliceOf(relatedType), 0, 1)
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(slice))
	groupedResults := reflect.MakeMap(groupedResultsMapType)
	pivotKey := eloquent.OrmPivotAlias + relation.PivotSelfColumn
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

	modelRelationFieldIndex := parent.FieldsByStructName[relation.Relation.Name].Index
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
				panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.Name].Name))
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
					panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.Name].Name))
				}
				model.Field(modelRelationFieldIndex).Set(value)
			}

		}
	}
}
