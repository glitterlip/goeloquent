package goeloquent

import (
	"fmt"
	"reflect"
)

type MorphManyRelation struct {
	Relation
	RelatedIdColumn  string
	ParentKey        string
	Builder          *Builder
	RelatedTpeColumn string
	MorphType        string
}

/*
MorphMany is a relation that can be used to retrieve the parent model of a polymorphic relation.
For example
images table
id  imageable_id  imageable_type path
2   1             User    	/path/to/image
3   2             User   	/path/to/image
4   1             Post   	/path/to/image
5   1             Post   	/path/to/image

	type Post struct{
		...
		Images     []Image `goelo:"MorphMany:ImagesRelation"`
		...
	}

	func (p *Post) ImagesRelation() *goeloquent.RelationBuilder {
		return p.MorphMany(p, &Image{}, "imageable_id", "id", "imageable_type")
	}

DB.Model(&Post{}).Where("id",1).With("Images").First(&post)
*/
func (m *EloquentModel) MorphMany(self interface{}, related interface{}, relatedTypeColumn, relatedIdColumn, selfKey string) *RelationBuilder {
	b := NewRelationBaseBuilder(related)
	relation := MorphManyRelation{
		Relation: Relation{
			Parent:  self,
			Related: related,
			Type:    RelationMorphOne,
		},
		RelatedIdColumn: relatedIdColumn, ParentKey: selfKey, Builder: b, RelatedTpeColumn: relatedTypeColumn,
	}
	selfModel := GetParsedModel(self)
	selfDirect := reflect.Indirect(reflect.ValueOf(self))
	relation.MorphType = GetMorphMap(selfModel.Name)

	b.Where(relatedIdColumn, "=", selfDirect.Field(selfModel.FieldsByDbName[selfKey].Index).Interface())
	b.WhereNotNull(relatedIdColumn)
	b.Where(relatedTypeColumn, "=", relation.MorphType)

	return &RelationBuilder{Builder: b, Relation: &relation}

}
func (r *MorphManyRelation) AddEagerConstraints(models interface{}) {
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
	r.Builder.Reset(TYPE_WHERE)
	r.Builder.WhereNotNull(r.RelatedIdColumn)
	r.Builder.Where(r.RelatedTpeColumn, "=", r.MorphType)
	r.Builder.WhereIn(r.RelatedIdColumn, keys)
}
func MatchMorphMany(models interface{}, related interface{}, relation *MorphManyRelation) {
	relatedModelsValue := related.(reflect.Value)
	relatedModels := relatedModelsValue
	relatedModel := GetParsedModel(relation.Related)
	relatedType := reflect.ValueOf(relation.Relation.Related).Elem().Type()

	parent := GetParsedModel(relation.Parent)
	relationFieldIsPtr := parent.FieldsByStructName[relation.Relation.Name].FieldType.Kind() == reflect.Ptr
	var sliceEleIsptr bool
	if relationFieldIsPtr {
		sliceEleIsptr = parent.FieldsByStructName[relation.Relation.Name].FieldType.Elem().Elem().Kind() == reflect.Ptr
	} else {
		sliceEleIsptr = parent.FieldsByStructName[relation.Relation.Name].FieldType.Elem().Kind() == reflect.Ptr
	}
	var slice reflect.Value
	if sliceEleIsptr {
		slice = reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(relatedType)), 0, 1)
	} else {
		slice = reflect.MakeSlice(reflect.SliceOf(relatedType), 0, 1)
	}
	groupedResultsMapType := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(slice))
	groupedResults := reflect.MakeMap(groupedResultsMapType)
	if !relatedModels.IsValid() || relatedModels.IsNil() {
		return
	}
	for i := 0; i < relatedModels.Len(); i++ {
		result := relatedModels.Index(i)
		ownerKeyIndex := relatedModel.FieldsByDbName[relation.RelatedIdColumn].Index
		groupKey := reflect.ValueOf(fmt.Sprint(result.FieldByIndex([]int{ownerKeyIndex})))
		existed := groupedResults.MapIndex(groupKey)
		if !existed.IsValid() {
			existed = reflect.New(slice.Type())
		} else {
			//after initilized and store in map and then get from the map,slice is a reflect.value(struct),we need to convert it
			existed = existed.Interface().(reflect.Value)
		}
		ptr := reflect.New(slice.Type())
		if sliceEleIsptr {
			v := reflect.Append(existed.Elem(), result.Addr())
			ptr.Elem().Set(v)
		} else {
			ptr.Elem().Set(reflect.Append(existed.Elem(), result))

		}
		groupedResults.SetMapIndex(groupKey, reflect.ValueOf(ptr))
	}

	targetSlice := reflect.Indirect(reflect.ValueOf(models))

	modelRelationFieldIndex := parent.FieldsByStructName[relation.Relation.Name].Index
	modelKeyFieldIndex := parent.FieldsByDbName[relation.ParentKey].Index

	if rvP, ok := models.(*reflect.Value); ok {
		for i := 0; i < rvP.Len(); i++ {
			e := rvP.Index(i)
			modelKey := e.Field(modelKeyFieldIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			value := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				if relationFieldIsPtr {
					e.Field(modelRelationFieldIndex).Set(value)
				} else {
					e.Field(modelRelationFieldIndex).Set(value.Elem())
				}
			}

		}
	} else if targetSlice.Type().Kind() != reflect.Slice {
		model := targetSlice
		modelKey := model.Field(modelKeyFieldIndex)
		modelKeyStr := fmt.Sprint(modelKey)
		modelSlice := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
		value := modelSlice
		if value.IsValid() {
			value = value.Interface().(reflect.Value)
			if !model.Field(modelRelationFieldIndex).CanSet() {
				panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.Name].Name))
			}
			if relationFieldIsPtr {
				model.Field(modelRelationFieldIndex).Set(value)
			} else {
				model.Field(modelRelationFieldIndex).Set(value.Elem())
			}
		}

	} else {
		for i := 0; i < targetSlice.Len(); i++ {
			model := targetSlice.Index(i)
			modelKey := model.Field(modelKeyFieldIndex)
			modelKeyStr := fmt.Sprint(modelKey)
			modelSlice := groupedResults.MapIndex(reflect.ValueOf(modelKeyStr))
			value := modelSlice
			if value.IsValid() {
				value = value.Interface().(reflect.Value)
				if !model.Field(modelRelationFieldIndex).CanSet() {
					panic(fmt.Sprintf("model: %s field: %s cant be set", parent.Name, parent.FieldsByStructName[relation.Relation.Name].Name))
				}
				if relationFieldIsPtr {
					model.Field(modelRelationFieldIndex).Set(value)
				} else {
					model.Field(modelRelationFieldIndex).Set(value.Elem())
				}
			}
		}
	}
}
