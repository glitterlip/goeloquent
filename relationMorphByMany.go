package goeloquent

import (
	"database/sql"
	"fmt"
	"reflect"
)

// MorphByManyRelation  for better understanding,rename the parameters. parent prefix represent for current model column, related represent for related model column,pivot prefix represent for pivot table
//for example we have post,tag and tagable table
//for tag model parentkey => user table id column,relatedkey => phone table id column,relatedparentkey => phone table user_id column
type MorphByManyRelation struct {
	Relation
	PivotTable      string
	PivotParentKey  string
	PivotRelatedKey string
	ParentKey       string
	RelatedKey      string
	Builder         *Builder
	PivotTypeColumn       string
}

func (m *EloquentModel) MorphByMany(self , related interface{}, pivotTable, pivotParentKey, pivotRelatedKey, parentKey, relatedKey, pivotTypeColumn string) *RelationBuilder {
	b := NewRelationBaseBuilder(related)
	relation := MorphByManyRelation{
		Relation: Relation{
			Parent:  self,
			Related: related,
			Type:    RelationMorphByMany,
		},
		PivotTable: pivotTable, PivotParentKey: pivotParentKey, PivotRelatedKey: pivotRelatedKey, ParentKey: parentKey, RelatedKey: relatedKey, Builder: b, PivotTypeColumn: pivotTypeColumn,
	}
	selfModel := GetParsedModel(self)
	relatedModel := GetParsedModel(related)
	b.Join(relation.PivotTable, relation.PivotTable+"."+relation.PivotRelatedKey, "=", relatedModel.Table+"."+relation.RelatedKey)
	b.Select(relatedModel.Table + "." + "*")
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotRelatedKey, PivotAlias, relation.PivotRelatedKey))
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotParentKey, PivotAlias, relation.PivotParentKey))
	selfDirect := reflect.Indirect(reflect.ValueOf(self))
	b.Where(relation.PivotParentKey, selfDirect.Field(selfModel.FieldsByDbName[parentKey].Index))
	modelMorphName := Eloquent.MorphModelMap[relatedModel.Name]
	b.Where(pivotTypeColumn, modelMorphName)
	return &RelationBuilder{Builder: b, Relation: &relation}

}

func (r *MorphByManyRelation) AddEagerConstraints(models interface{}) {
	parentParsedModel := GetParsedModel(r.Parent)
	index := parentParsedModel.FieldsByDbName[r.ParentKey].Index
	modelSlice := reflect.Indirect(reflect.ValueOf(models))
	var keys []interface{}
	if modelSlice.Type().Kind() == reflect.Slice {
		for i := 0; i < modelSlice.Len(); i++ {
			model := modelSlice.Index(i)
			modelKey := reflect.Indirect(model).Field(index).Interface()
			keys = append(keys, modelKey)
		}
	} else {
		model := modelSlice
		modelKey := model.Field(index).Interface()
		keys = append(keys, modelKey)
	}
	r.Builder.Wheres = nil
	r.Builder.WhereIn(r.PivotTable+"."+r.PivotParentKey, keys)
	r.Builder.Where(r.PivotTypeColumn, Eloquent.MorphModelMap[parentParsedModel.Name])
}
func MatchMorphByMany(models interface{}, related interface{}, relation *MorphByManyRelation) {
	relatedValue := related.(reflect.Value)
	relatedResults := relatedValue
	relatedModel := GetParsedModel(relation.Related)
	parent := GetParsedModel(relation.Parent)
	isPtr := parent.FieldsByStructName[relation.Relation.Name].FieldType.Elem().Kind() == reflect.Ptr
	var relatedType reflect.Type

	if isPtr {
		relatedType = reflect.ValueOf(relation.Relation.Related).Type()
	} else {
		relatedType = reflect.ValueOf(relation.Relation.Related).Elem().Type()
	}
	slice := reflect.MakeSlice(reflect.SliceOf(relatedType), 0, 1)
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(slice))
	groupedResults := reflect.MakeMap(groupedResultsMapType)
	pivotKey := PivotAlias + relation.PivotRelatedKey
	if !relatedResults.IsValid() || relatedResults.IsNil() {
		return
	}
	for i := 0; i < relatedResults.Len(); i++ {
		result := relatedResults.Index(i)
		pivotMap := result.FieldByIndex(relatedModel.PivotFieldIndex).Interface().(map[string]sql.NullString)
		groupKey := reflect.ValueOf(pivotMap[pivotKey].String)
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

	modelRelationFiledIndex := parent.FieldsByStructName[relation.Relation.Name].Index
	modelKeyFiledIndex := parent.FieldsByDbName[relation.ParentKey].Index

	if targetSlice.Type().Kind() != reflect.Slice {
		model := targetSlice
		modelKey := model.Field(modelKeyFiledIndex)
		modelKeyStr := fmt.Sprint(modelKey)
		value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
		if value.IsValid() {
			value = value.Interface().(reflect.Value)
			if !model.Field(modelRelationFiledIndex).CanSet() {
				panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.Name].Name))
			}
			model.Field(modelRelationFiledIndex).Set(value)
		}
	} else {
		for i := 0; i < targetSlice.Len(); i++ {
			model := targetSlice.Index(i)
			if model.Kind() == reflect.Ptr {
				model = reflect.Indirect(model)
			}
			modelKey := model.Field(modelKeyFiledIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				if !model.Field(modelRelationFiledIndex).CanSet() {
					panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.Name].Name))
				}
				model.Field(modelRelationFiledIndex).Set(value)
			}

		}
	}
}
