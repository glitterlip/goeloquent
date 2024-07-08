package tests

import (
	"database/sql"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type DefaultModel struct {
	*goeloquent.EloquentModel
	ID   int64  `goelo:"column:id;primaryKey"`
	Name string `goelo:"column:name_alias"`
}
type TableName struct {
	*goeloquent.EloquentModel
	ID            int64         `goelo:"column:id;primaryKey"`
	Title         string        `goelo:"column:title"`
	Status        int           `goelo:"column:status"`
	Permissions   []string      `goelo:"column:permissions"`
	CreateTime    sql.NullTime  `goelo:"column:createdAt"`
	Update        sql.NullTime  `goelo:"column:updated"`
	Deleted       sql.NullTime  `goelo:"DELETED_AT"`
	Exclude       string        `test:"column:exclude;"`
	HasOneDefault *DefaultModel `goelo:"HasOne:RelationHasOne"`
	Children      []TableName   `goelo:"HasMany:RelationHasMany"`
	ValidChildren []TableName   `goelo:"HasMany:RelationHasMany"`
	Parent        *TableName    `goelo:"BelongsTo:RelationBelongsTo"`
}

func (t *TableName) RelationBelongsTo() *goeloquent.BelongsToRelation {
	return t.BelongsTo(t, &TableName{}, "pid", "id")
}

func (t *TableName) RelationHasOne() *goeloquent.HasOneRelation {
	return t.HasOne(t, &DefaultModel{}, "id", "id")
}
func (t *TableName) RelationHasMany() *goeloquent.HasManyRelation {
	return t.HasMany(t, &TableName{}, "id", "pid")
}
func (t *TableName) RelationValidHasMany() *goeloquent.HasManyRelation {
	r := t.HasMany(t, &TableName{}, "id", "pid")
	r.Where("status", 1)
	return r
}
func (t *TableName) EloquentGetWithRelations() map[string]goeloquent.RelationFunc {
	return map[string]goeloquent.RelationFunc{
		"Parent": func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return builder.Where("status", ">=", 1)
		},
	}
}
func (t *TableName) EloquentGetWithRelationCounts() map[string]goeloquent.RelationFunc {
	return map[string]goeloquent.RelationFunc{
		"ValidChildren": func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return builder
		},
	}
}
func (t *TableName) TableName() string {
	return "t_name"
}
func (t *TableName) ConnectionName() string {
	return "test"
}
func (t *TableName) EloquentAddGlobalScopes() map[string]goeloquent.ScopeFunc {
	return map[string]goeloquent.ScopeFunc{
		"Valid": func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return builder.Where("status", 1)
		},
	}
}
func (t *TableName) EloquentGetDefaultAttributes() map[string]interface{} {
	return map[string]interface{}{
		"status":      1,
		"permissions": []string{"create", "update"},
	}
}
func (t *TableName) EloquentGetGuarded() map[string]struct{} {
	return map[string]struct{}{
		"id":          {},
		"status":      {},
		"permissions": {},
	}
}
func (t *TableName) EloquentGetFillable() map[string]struct{} {
	return map[string]struct{}{
		"title": {},
	}
}
func TestParseModel(t *testing.T) {

	parsed := goeloquent.GetParsedModel(&DefaultModel{})
	assert.Equal(t, "id", parsed.FieldsByStructName["ID"].ColumnName)
	assert.Equal(t, "id", parsed.PrimaryKey.ColumnName)
	assert.Equal(t, "name_alias", parsed.FieldsByStructName["Name"].ColumnName)
	assert.Equal(t, 2, parsed.FieldsByStructName["Name"].Index)
	assert.Equal(t, false, parsed.SoftDelete)

	parsed1 := goeloquent.GetParsedModel(&TableName{})

	assert.Equal(t, "t_name", parsed1.Table)
	assert.Equal(t, "test", parsed1.ConnectionName)
	assert.Equal(t, true, parsed1.SoftDelete)
	assert.Equal(t, map[string]interface{}{
		"status":      1,
		"permissions": []string{"create", "update"},
	}, parsed1.DefaultAttributes)
	_, has := parsed1.FieldsByStructName["Exclude"]
	assert.True(t, !has)
	assert.True(t, reflect.DeepEqual(
		map[string]struct{}{
			"id":          {},
			"status":      {},
			"permissions": {},
		}, parsed1.Guards,
	))

	assert.True(t, reflect.DeepEqual(parsed1.Fillables, map[string]struct{}{
		"title": {},
	}))

}
func TestParseRelation(t *testing.T) {
	parsed1 := goeloquent.GetParsedModel(&TableName{})

	assert.Contains(t, parsed1.EagerRelations, "Parent")
	assert.Contains(t, parsed1.EagerRelationCounts, "ValidChildren")

}
