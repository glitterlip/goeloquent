package goeloquent

import (
	"fmt"
	"reflect"
)

type Relations string

func GetMorphMap(name string) string {
	v, ok := RegisteredMorphModelsMap.Load(name)
	if !ok {
		panic(fmt.Sprintf("no registered model found for %s", name))
	}
	return v.(string)
}

const (
	RelationHasOne    Relations = "HasOne"
	RelationBelongsTo Relations = "BelongsTo"

	RelationHasMany Relations = "HasMany"

	RelationHasOneThrough  Relations = "HasOneThrough"
	RelationHasManyThrough Relations = "HasManyThrough"

	RelationBelongsToMany Relations = "BelongsToMany"

	RelationMorphOne  Relations = "MorphOne"
	RelationMorphMany Relations = "MorphMany"
	RelationMorphTo   Relations = "MorphTo"

	RelationMorphToMany Relations = "MorphToMany"
	RelationMorphByMany Relations = "MorphByMany"
	PivotAlias                    = "goelo_pivot_"
	OrmPivotAlias                 = "goelo_orm_pivot_"
)

type Relation struct {
	SelfModel        interface{} // parent/self model pointer,usually a pointer to a struct , &User{}
	RelatedModel     interface{} // related model,usually a pointer to a struct , &User{}
	RelationTypeName Relations   // relation type name
	FieldName        string      // field name in self model corresponding to this relation
	*EloquentBuilder             // RelationBuilder

}

func (r *Relation) GetEloquentBuilder() *EloquentBuilder {
	return r.EloquentBuilder
}

type RelationI interface {
	AddEagerConstraints(models interface{})
	AddConstraints()
	GetEloquentBuilder() *EloquentBuilder

	//Match(models []interface{}, results []interface{}, relation string) //Match the eagerly loaded results to their parents.
	//GetResults() //Get the results of the relationship.
	//GetEager() //Get the relationship for eager loading.
	//Get(dest interface{}, columns ...interface{}) //Get the results of the relationship.

	//GetRelationExistenceQuery(Builder $query, Builder $parentQuery, columns = ['*']) //Get the relationship for exists query.
	//GetRelationExistenceCountQuery(Builder $query, Builder $parentQuery) //Get the relationship count of an eager load query.

	//GetKeys(models []interface{},key string) //Get all the primary keys for an array of models.
	//GetQuery() *Builder //Get the underlying query for the relation.
	//GetQualifiedParentKeyName()string //
}

func NewRelationBaseBuilder(related interface{}) *EloquentBuilder {

	if related == nil {
		return ToEloquentBuilder(NewQueryBuilder())
	}
	return NewEloquentBuilder(related)
}
func (r *Relation) GetSelfKey(key string) interface{} {
	feild := GetParsedModel(r.SelfModel).FieldsByDbName[key]
	return reflect.ValueOf(r.SelfModel).Elem().FieldByName(feild.Name).Interface()
}

/*
HasMany Define a one-to-many relationship.
*/
func (m *EloquentModel) HasMany(selfModelPointer, relatedModelPointer interface{}, selfColumn, relatedColumn string) *HasManyRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := HasManyRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationHasMany,
			EloquentBuilder:  b,
		},
		SelfColumn:    selfColumn,
		RelatedColumn: relatedColumn,
	}

	relation.AddConstraints()
	return &relation
}
func (m *EloquentModel) BelongsTo(selfModelPointer, relatedModelPointer interface{}, selfColumn, relatedColumn string) *BelongsToRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := BelongsToRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationBelongsTo,
			EloquentBuilder:  b,
		},
		SelfColumn:    selfColumn,
		RelatedColumn: relatedColumn,
	}
	relation.AddConstraints()

	return &relation

}
func (m *EloquentModel) BelongsToMany(selfModelPointer, relatedModelPointer interface{}, pivotTable, pivotSelfColumn, pivotRelatedColumn, selfModelColumn, relatedModelColumn string) *BelongsToManyRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := BelongsToManyRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationBelongsToMany,
			EloquentBuilder:  b,
		},
		PivotTable:         pivotTable,
		PivotSelfColumn:    pivotSelfColumn,
		PivotRelatedColumn: pivotRelatedColumn,
		SelfColumn:         selfModelColumn,
		RelatedColumn:      relatedModelColumn,
	}
	relatedModel := GetParsedModel(relatedModelPointer)
	b.Join(relation.PivotTable, relation.PivotTable+"."+relation.PivotRelatedColumn, "=", relatedModel.Table+"."+relation.RelatedColumn)
	b.Select(relatedModel.Table + "." + "*")
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotSelfColumn, OrmPivotAlias, relation.PivotSelfColumn))
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotRelatedColumn, OrmPivotAlias, relation.PivotRelatedColumn))
	relation.AddConstraints()
	return &relation

}
func (m *EloquentModel) HasOne(selfModelPointer, relatedModelPointer interface{}, selfColumn, relatedColumn string) *HasOneRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := HasOneRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationHasOne,
			EloquentBuilder:  b,
		},
		RelatedColumn: relatedColumn,
		SelfColumn:    selfColumn,
	}
	relation.AddConstraints()
	return &relation
}

/*
MorphTo Create a new morph to relationship instance.
*/
func (m *EloquentModel) MorphTo(selfModelPointer interface{}, selfRelatedIdColumn, selfRelatedTypeColumn, relatedModelIdColumn string) *MorphToRelation {

	builder := NewRelationBaseBuilder(nil)
	relation := MorphToRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     nil,
			RelationTypeName: RelationMorphTo,
			EloquentBuilder:  builder,
		},
		SelfRelatedIdColumn:   selfRelatedIdColumn,
		RelatedModelIdColumn:  relatedModelIdColumn,
		SelfRelatedTypeColumn: selfRelatedTypeColumn,
	}

	relation.AddConstraints()

	return &relation

}

/*
MorphOne Define a polymorphic one-to-one relationship.
*/
func (m *EloquentModel) MorphOne(selfModelPointer, relatedModelPointer interface{}, selfColumn, relatedModelIdColumn, relatedModelTypeColumn string, morphType ...string) *MorphOneRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := MorphOneRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationMorphOne,
			EloquentBuilder:  b,
		},
		RelatedModelIdColumn:   relatedModelIdColumn,
		SelfColumn:             selfColumn,
		RelatedModelTypeColumn: relatedModelTypeColumn,
	}
	selfModel := GetParsedModel(selfModelPointer)
	if len(morphType) > 0 {
		relation.RelatedModelTypeColumnValue = morphType[0]
	} else {
		relation.RelatedModelTypeColumnValue = GetMorphMap(selfModel.Name)
	}
	relation.AddConstraints()

	return &relation

}

/*
MorphMany Define a polymorphic one-to-many relationship.
*/
func (m *EloquentModel) MorphMany(selfModelPointer, relatedModelPointer interface{}, selfColumn, relatedModelIdColumn, relatedModelTypeColumn string, relatedModelTypeColumnValue ...string) *MorphManyRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := MorphManyRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationMorphOne,
			EloquentBuilder:  b,
		},
		RelatedModelIdColumn:   relatedModelIdColumn,
		SelfColumn:             selfColumn,
		RelatedModelTypeColumn: relatedModelTypeColumn,
	}
	selfModel := GetParsedModel(selfModelPointer)

	if len(relatedModelTypeColumnValue) > 0 {
		relation.RelatedModelTypeColumnValue = relatedModelTypeColumnValue[0]
	} else {
		relation.RelatedModelTypeColumnValue = GetMorphMap(selfModel.Name)
	}

	b.Where(relatedModelTypeColumn, "=", relation.RelatedModelTypeColumnValue)

	relation.AddConstraints()

	return &relation

}

/*
MorphToMany Define a polymorphic many-to-many relationship.
*/
func (m *EloquentModel) MorphToMany(selfModelPointer, relatedModelPointer interface{}, selfIdColumn, relatedIdColumn, pivotTable, pivotSelfIdColumn, pivotSelfTypeColumn, pivotRelatedIdColumn string, morphType ...string) *MorphToManyRelation {

	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := MorphToManyRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationMorphToMany,
			EloquentBuilder:  b,
		},
		PivotTable:           pivotTable,
		PivotSelfIdColumn:    pivotSelfIdColumn,
		PivotSelfTypeColumn:  pivotSelfTypeColumn,
		SelfIdColumn:         selfIdColumn,
		RelatedIdColumn:      relatedIdColumn,
		PivotRelatedIdColumn: pivotRelatedIdColumn,
	}
	relatedModel := GetParsedModel(relatedModelPointer)
	selfModel := GetParsedModel(selfModelPointer)
	b.Join(relation.PivotTable, relation.PivotTable+"."+relation.PivotRelatedIdColumn, "=", relatedModel.Table+"."+relation.RelatedIdColumn)
	b.Select(relatedModel.Table + "." + "*")
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotRelatedIdColumn, OrmPivotAlias, relation.PivotRelatedIdColumn))
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotSelfIdColumn, OrmPivotAlias, relation.PivotSelfIdColumn))
	var selfModelTypeColumnValue string

	if len(morphType) > 0 {
		selfModelTypeColumnValue = morphType[0]
	} else {
		selfModelTypeColumnValue = GetMorphMap(selfModel.Name)
	}
	relation.SelfModelTypeColumnValue = selfModelTypeColumnValue
	relation.AddConstraints()
	return &relation

}

/*
MorphByMany Define a polymorphic many-to-many relationship.
*/
func (m *EloquentModel) MorphByMany(selfModelPointer, relatedModelPointer interface{}, pivotTable, selfColumn, relatedIdColumn, pivotSelfColumn, pivotRelatedIdColumn, pivotRelatedTypeColumn string) *MorphByManyRelation {
	b := NewRelationBaseBuilder(relatedModelPointer)
	relation := MorphByManyRelation{
		Relation: &Relation{
			SelfModel:        selfModelPointer,
			RelatedModel:     relatedModelPointer,
			RelationTypeName: RelationMorphByMany,
			EloquentBuilder:  b,
		},
		PivotTable:             pivotTable,
		PivotSelfColumn:        pivotSelfColumn,
		PivotRelatedIdColumn:   pivotRelatedIdColumn,
		SelfColumn:             selfColumn,
		RelatedIdColumn:        relatedIdColumn,
		PivotRelatedTypeColumn: pivotRelatedTypeColumn,
	}
	selfModel := GetParsedModel(selfModelPointer)
	relatedModel := GetParsedModel(relatedModelPointer)
	modelMorphName := GetMorphMap(selfModel.Name)

	b.Join(relation.PivotTable, relation.PivotTable+"."+relation.PivotRelatedIdColumn, "=", relatedModel.Table+"."+relation.RelatedIdColumn)
	b.Select(relatedModel.Table + "." + "*")
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.PivotRelatedIdColumn, OrmPivotAlias, relation.PivotRelatedIdColumn))
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.SelfColumn, OrmPivotAlias, relation.PivotSelfColumn))
	b.Select(fmt.Sprintf("%s.%s as %s%s", relation.PivotTable, relation.RelationTypeName, OrmPivotAlias, relation.RelationTypeName))

	relation.RelatedModelTypeColumnValue = modelMorphName
	relation.AddConstraints()
	return &relation

}
