package goeloquent

import (
	"reflect"
)

const (
	RelationHasOne    = "HasOne"
	RelationBelongsTo = "BelongsTo"

	RelationHasMany = "HasMany"

	RelationHasOneThrough  = "HasOneThrough"
	RelationHasManyThrough = "HasManyThrough"

	RelationBelongsToMany = "BelongsToMany"

	RelationMorphOne  = "MorphOne"
	RelationMorphMany = "MorphMany"
	RelationMorphTo   = "MorphTo"

	RelationMorphToMany = "MorphToMany"
	RelationMorphByMany = "MorphByMany"
)

type Relation struct {
	Parent  interface{}
	Related interface{}
	Type    string
	Builder *RelationBuilder
	Name    string
	//MorphMap [] ?
}
type RelationI interface {
	AddEagerConstraints(interface{})
}

type HasOneThrough struct {
	Relation
	ThroughParent  interface{}
	FarParent      interface{}
	FirstKey       string
	SecondKey      string
	LocalKey       string
	SecondLocalKey string
}
type HasManyThrough struct {
	Relation
	ThroughParent  interface{}
	FarParent      interface{}
	FirstKey       string
	SecondKey      string
	LocalKey       string
	SecondLocalKey string
}

func NewRelationBaseBuilder(related interface{}) *Builder {
	if related == nil {
		baseBuilder := Builder{
			Components: make(map[string]struct{}),
			Grammar:    &MysqlGrammar{},
			EagerLoad:  make(map[string]func(builder *RelationBuilder) *RelationBuilder),
			Bindings:   make(map[string][]interface{}),
		}
		p := &baseBuilder
		baseBuilder.Grammar.SetBuilder(p)
		return &baseBuilder
	}
	relatedP := reflect.ValueOf(related)
	relatedModel := reflect.Indirect(relatedP)
	connectionName := "default"
	if c, ok := relatedModel.Interface().(ConnectionName); ok {
		connectionName = c.ConnectionName()
	}
	connection := Eloquent.Connection(connectionName)
	baseBuilder := Builder{
		Connection: connection,
		Components: make(map[string]struct{}),
		Grammar:    &MysqlGrammar{},
		EagerLoad:  make(map[string]func(builder *RelationBuilder) *RelationBuilder),
		Bindings:   make(map[string][]interface{}),
	}
	p := &baseBuilder
	baseBuilder.SetModel(related)
	baseBuilder.Grammar.SetBuilder(p)
	return &baseBuilder
}
