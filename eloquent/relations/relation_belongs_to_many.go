package relations

import (
	"fmt"
	"github.com/glitterlip/goeloquent/eloquent"
	"github.com/glitterlip/goeloquent/query"
	"reflect"
)

type BelongsToManyRelation struct {
	eloquent.Relation
	PivotTable         string
	PivotSelfColumn    string
	PivotRelatedColumn string
	SelfColumn         string
	RelatedColumn      string
	PivotColumns       []string      //extra columns to select in pivot table
	PivotWheres        []query.Where //extra where conditions in pivot table
	WithTimestamps     bool
}

func (r *BelongsToManyRelation) AddConstraints() {
	r.Builder.Where(r.RelatedColumn, "=", r.GetSelfKey(r.SelfColumn))
}

func (r *BelongsToManyRelation) AddEagerConstraints(parentModels interface{}) {
	selfParsedModel := eloquent.GetParsedModel(r.SelfModel)
	index := selfParsedModel.FieldsByDbName[r.SelfColumn].Index
	modelSlice := reflect.Indirect(reflect.ValueOf(parentModels))
	var selfKeys []interface{}
	if modelSlice.Type().Kind() == reflect.Slice {
		for i := 0; i < modelSlice.Len(); i++ {
			model := modelSlice.Index(i)
			modelKey := reflect.Indirect(model).Field(index).Interface()
			selfKeys = append(selfKeys, modelKey)
		}
	} else if ms, ok := parentModels.(*reflect.Value); ok {
		for i := 0; i < ms.Len(); i++ {
			modelKey := ms.Index(i).Field(index).Interface()
			selfKeys = append(selfKeys, modelKey)
		}
	} else {
		model := modelSlice
		modelKey := model.Field(index).Interface()
		selfKeys = append(selfKeys, modelKey)
	}
	r.Builder.WhereIn(r.PivotTable+"."+r.PivotSelfColumn, selfKeys)
}
func MatchBelongsToMany(models interface{}, relatedModelsR interface{}, relation *BelongsToManyRelation) {
	relatedModels := relatedModelsR.(reflect.Value)
	parsedRelatedModel := eloquent.GetParsedModel(relation.RelatedModel)
	self := eloquent.GetParsedModel(relation.SelfModel)

	isPtr := self.FieldsByStructName[relation.Relation.Name].FieldType.Elem().Kind() == reflect.Ptr
	var relatedType reflect.Type
	if isPtr {
		relatedType = reflect.ValueOf(relation.Relation.RelatedModel).Type()
	} else {
		relatedType = reflect.ValueOf(relation.Relation.RelatedModel).Elem().Type()
	}
	slice := reflect.MakeSlice(reflect.SliceOf(relatedType), 0, 1)
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(slice))
	groupedResults := reflect.MakeMap(groupedResultsMapType)
	pivotSelfKey := eloquent.OrmPivotAlias + relation.PivotSelfColumn
	if !relatedModels.IsValid() || relatedModels.IsNil() {
		return
	}
	for i := 0; i < relatedModels.Len(); i++ {
		relatedModel := relatedModels.Index(i)
		eloquentModelPtr := relatedModel.FieldByIndex([]int{parsedRelatedModel.EloquentModelFieldIndex})
		pivotMap := eloquentModelPtr.Elem().FieldByIndex([]int{parsedRelatedModel.PivotFieldIndex}).Interface().(map[string]interface{})
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

	targetSlice := reflect.Indirect(reflect.ValueOf(models))

	modelRelationFieldIndex := self.FieldsByStructName[relation.Relation.Name].Index
	modelKeyFieldIndex := self.FieldsByDbName[relation.SelfColumn].Index

	if rvP, ok := models.(*reflect.Value); ok {
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
				panic(fmt.Sprintf("model: %s field: %s cant be set", self.Name, self.FieldsByStructName[relation.Relation.Name].Name))
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
					panic(fmt.Sprintf("model: %s field: %s cant be set", self.Name, self.FieldsByStructName[relation.Relation.Name].Name))
				}
				model.Field(modelRelationFieldIndex).Set(value)
			}

		}
	}
}
