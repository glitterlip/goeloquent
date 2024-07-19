package tests

import (
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Tag1 struct {
	*goeloquent.EloquentModel
	ID   int64  `goelo:"column:tid;primaryKey"`
	Name string `goelo:"column:name"`
}

func (t *Tag1) EloquentAddGlobalScopes() map[string]goeloquent.ScopeFunc {
	return map[string]goeloquent.ScopeFunc{
		"active": func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return builder.Where("active", 1)
		},
	}
}

type Tag2 struct {
	*goeloquent.EloquentModel
	ID   int64  `goelo:"column:tid;primaryKey"`
	Name string `goelo:"column:name"`
}

func (t *Tag2) EloquentAddGlobalScopes() map[string]goeloquent.ScopeFunc {
	return map[string]goeloquent.ScopeFunc{
		"active": func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return builder.Where("active", 1)
		},
		"defaultOrder": func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return builder.OrderBy("id")
		},
	}
}

type Tag3 struct {
	*goeloquent.EloquentModel
	ID   int64  `goelo:"column:tid;primaryKey"`
	Name string `goelo:"column:name"`
}

func (t *Tag3) EloquentAddGlobalScopes() map[string]goeloquent.ScopeFunc {
	return map[string]goeloquent.ScopeFunc{
		"email": func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return builder.Where("email", "email1").OrWhere("email", "email2")
		},
		"select": func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return builder.Select("email", "password")
		},
		"defaultOrder": func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return builder.OrderBy("id")
		},
		"active": func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return builder.Where("active", 1)
		},
	}
}

func TestGlobalScopeIsApplied(t *testing.T) {
	var tags []Tag1
	b := DB.Model(&Tag1{})
	_, err := b.Pretend().Get(&tags)
	assert.Nil(t, err)
	assert.Equal(t, "select * from `tag1` where `active` = ?", b.PreparedSql)
	assert.Equal(t, []interface{}{1}, b.GetBindings())
}
func TestGlobalScopeCanBeRemoved(t *testing.T) {
	var tags []Tag1
	b := DB.Model(&Tag1{})
	_, err := b.Pretend().WithOutGlobalScopes("active").Get(&tags)
	assert.Nil(t, err)
	assert.Equal(t, "select * from `tag1`", b.PreparedSql)
	assert.Empty(t, b.GetBindings())
}
func TestAllGlobalScopesCanBeRemoved(t *testing.T) {
	var tags []Tag2
	b := DB.Model(&Tag2{})
	_, err := b.Pretend().WithOutGlobalScopes().Get(&tags)
	assert.Nil(t, err)
	assert.Equal(t, "select * from `tag2`", b.PreparedSql)
	assert.Empty(t, b.GetBindings())
}
func TestGlobalScopesWithOrWhereConditionsAreNested(t *testing.T) {
	var tags []Tag3
	b := DB.Model(&Tag3{})
	r, err := b.Pretend().Get(&tags)
	assert.Nil(t, err)
	assert.ElementsMatch(t, r.Bindings, []interface{}{"email1", "email2", 1})

}
