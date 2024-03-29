package goeloquent

import (
	"fmt"
	"reflect"
)

// HasManyRelation for better understanding,rename the parameters. parent prefix represent for current model column, related represent for related model column,pivot prefix represent for pivot table
// for example we have user,address table
// for user model hasmanyaddress relation parentkey => user table id column,relatedkey => address table id column,relatedparentkey => address table user_id column
type HasManyRelation struct {
	Relation
	ReleatedKey string
	SelfKey     string
	Builder     *Builder
}

/*
HasMany is a relation that can be used to retrieve the related model of a one-to-many relation.
For example
address table
id  user_id  address
2   4        USA
3   4        EU
4   4        JP

	type User struct{
		...
		Addresses     []Address `goelo:"HasMany:AddressesRelation"`
		...
	}

	func (u *User) AddressesRelation() *goeloquent.RelationBuilder {
		return u.HasMany(u, &Address{}, "id", "user_id")
	}

DB.Model(&User{}).With("Addresses").Get(&users)
*/
func (m *EloquentModel) HasMany(self interface{}, related interface{}, selfKey, relatedKey string) *RelationBuilder {
	b := NewRelationBaseBuilder(related)
	relation := HasManyRelation{
		Relation{
			Parent:  self,
			Related: related,
			Type:    RelationHasMany,
		}, relatedKey, selfKey, b,
	}
	selfModel := GetParsedModel(self)
	selfDirect := reflect.Indirect(reflect.ValueOf(self))
	b.Where(relatedKey, "=", selfDirect.Field(selfModel.FieldsByDbName[selfKey].Index).Interface())
	b.WhereNotNull(relatedKey)

	return &RelationBuilder{Builder: b, Relation: &relation}
}
func (r *HasManyRelation) AddEagerConstraints(models interface{}) {
	relatedParsedModel := GetParsedModel(r.Parent)
	index := relatedParsedModel.FieldsByDbName[r.SelfKey].Index
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
	r.Builder.WhereNotNull(r.ReleatedKey)
	r.Builder.WhereIn(r.ReleatedKey, keys)
}
func MatchHasMany(models interface{}, related interface{}, relation *HasManyRelation) {
	relatedModelsValue := related.(reflect.Value)
	relatedModels := relatedModelsValue
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

	if !relatedModels.IsValid() || relatedModels.IsNil() {
		return
	}

	for i := 0; i < relatedModels.Len(); i++ {
		result := relatedModels.Index(i)
		foreignKeyIndex := relatedModel.FieldsByDbName[relation.ReleatedKey].Index
		groupKey := reflect.ValueOf(fmt.Sprint(result.FieldByIndex([]int{foreignKeyIndex})))
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
	modelKeyFieldIndex := parent.FieldsByDbName[relation.SelfKey].Index

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

			model.Field(modelRelationFieldIndex).Set(value)
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
				model.Field(modelRelationFieldIndex).Set(value)
			}
		}
	}

}
