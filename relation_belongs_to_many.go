package goeloquent

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

var (
	PivotAlias = "goelo_pivot_"
)

// BelongsToManyRelation for better understanding,rename the parameters. parent prefix represent for current model column, related represent for related model column,pivot prefix represent for pivot table
//for example we have user , role ,user_has_role table
//for role model parentkey => role table id column,relatedkey => user table id column,pivotrelatedkey => user_has_role table user_id column,pivotparentkey => user_has_role table role_id column
type BelongsToManyRelation struct {
	Relation
	PivotTable      string
	PivotSelfKey    string
	PivotRelatedKey string
	SelfKey         string
	RelatedKey      string
	PivotColumns    []string
	PivotWheres     []Where
	WithTimestamps  bool
	Builder         *Builder
}

func (m *EloquentModel) BelongsToMany(self interface{}, related interface{}, pivotTable, pivotSelfKey, pivotRelatedKey, selfKey, relatedKey string) *RelationBuilder {
	b := NewRelationBaseBuilder(related)
	relation := BelongsToManyRelation{
		Relation: Relation{
			Parent:  self,
			Related: related,
			Type:    RelationBelongsToMany,
		},
		PivotTable: pivotTable, PivotSelfKey: pivotSelfKey, PivotRelatedKey: pivotRelatedKey, SelfKey: selfKey, RelatedKey: relatedKey, Builder: b,
	}
	relatedModel := GetParsedModel(related)
	b.Join(relation.PivotTable, relation.PivotTable+"."+relation.PivotRelatedKey, "=", relatedModel.Table+"."+relation.RelatedKey)
	b.Select(relatedModel.Table + "." + "*")
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotSelfKey, PivotAlias, relation.PivotSelfKey))
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotRelatedKey, PivotAlias, relation.PivotRelatedKey))
	selfModel := GetParsedModel(self)
	selfDirect := reflect.Indirect(reflect.ValueOf(self))
	b.Where(relation.PivotSelfKey, selfDirect.Field(selfModel.FieldsByDbName[selfKey].Index).Interface())
	return &RelationBuilder{Builder: b, Relation: &relation}

}
func (relation *BelongsToManyRelation) SelectPivots(pivots ...string) {
	for _, pivot := range pivots {
		if strings.Contains(pivot, ".") {
			if strings.Contains(pivot, relation.PivotTable) {
				relation.Builder.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, pivot, PivotAlias, pivot))
			}
		} else {
			relation.Builder.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, pivot, PivotAlias, pivot))
		}
	}
}
func (relation *BelongsToManyRelation) WherePivots(pivotWheres ...Where) {
	for _, where := range pivotWheres {
		if strings.Contains(where.Column, ".") {
			if strings.Contains(where.Column, relation.PivotTable) {
				sa := strings.SplitN(where.Column, ".", 2)
				relation.Builder.Where(relation.PivotTable+"."+sa[1], where.Operator, where.Value, where.Boolean)
			}
		} else {
			relation.Builder.Where(relation.PivotTable+"."+where.Column, where.Operator, where.Value, where.Boolean)
		}
	}
}
func (relation *BelongsToManyRelation) AddEagerConstraints(parentModels interface{}) {
	parentParsedModel := GetParsedModel(relation.Parent)
	index := parentParsedModel.FieldsByDbName[relation.SelfKey].Index
	modelSlice := reflect.Indirect(reflect.ValueOf(parentModels))
	var parentKeys []interface{}
	if modelSlice.Type().Kind() == reflect.Slice {
		for i := 0; i < modelSlice.Len(); i++ {
			model := modelSlice.Index(i)
			modelKey := reflect.Indirect(model).Field(index).Interface()
			parentKeys = append(parentKeys, modelKey)
		}
	} else if ms, ok := parentModels.(*reflect.Value); ok {
		for i := 0; i < ms.Len(); i++ {
			modelKey := ms.Index(i).Field(index).Interface()
			parentKeys = append(parentKeys, modelKey)
		}
	} else {
		model := modelSlice
		modelKey := model.Field(index).Interface()
		parentKeys = append(parentKeys, modelKey)
	}
	relation.Builder.Reset(TYPE_WHERE)
	relation.Builder.WhereIn(relation.PivotTable+"."+relation.PivotSelfKey, parentKeys)
}
func MatchBelongsToMany(models interface{}, related interface{}, relation *BelongsToManyRelation) {
	relatedModels := related.(reflect.Value)
	parsedRelatedModel := GetParsedModel(relation.Related)
	self := GetParsedModel(relation.Parent)

	isPtr := self.FieldsByStructName[relation.Relation.Name].FieldType.Elem().Kind() == reflect.Ptr
	var relatedType reflect.Type
	if isPtr {
		relatedType = reflect.ValueOf(relation.Relation.Related).Type()
	} else {
		relatedType = reflect.ValueOf(relation.Relation.Related).Elem().Type()
	}
	slice := reflect.MakeSlice(reflect.SliceOf(relatedType), 0, 1)
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(slice))
	groupedResults := reflect.MakeMap(groupedResultsMapType)
	pivotSelfKey := PivotAlias + relation.PivotSelfKey
	if !relatedModels.IsValid() || relatedModels.IsNil() {
		return
	}
	for i := 0; i < relatedModels.Len(); i++ {
		relatedModel := relatedModels.Index(i)
		eloquentModelPtr := relatedModel.FieldByIndex(parsedRelatedModel.PivotFieldIndex[0:1])
		pivotMap := eloquentModelPtr.Elem().FieldByIndex(parsedRelatedModel.PivotFieldIndex[1:2]).Interface().(map[string]sql.NullString)
		groupKey := reflect.ValueOf(pivotMap[pivotSelfKey].String)
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
	modelKeyFieldIndex := self.FieldsByDbName[relation.SelfKey].Index

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
