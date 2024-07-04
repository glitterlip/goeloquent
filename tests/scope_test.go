package tests

import (
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func ScopeFunc(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
	builder.OrderBy("id", "desc")
	return builder
}
func AgeScope(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
	builder.Where("age", ">", 18)
	return builder
}

type Tag1 struct {
	*goeloquent.EloquentModel
	ID   int64  `goelo:"column:tid;primaryKey"`
	Name string `goelo:"column:name"`
}

func (t *Tag1) AddGlobalScopes() map[string]goeloquent.ScopeFunc {
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

func (t *Tag2) AddGlobalScopes() map[string]goeloquent.ScopeFunc {
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

func (t *Tag3) AddGlobalScopes() map[string]goeloquent.ScopeFunc {
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
	_, err := b.Pretend().Get(&tags)
	assert.Nil(t, err)
	sql := b.ToSql()
	var in bool
	if sql == "select `email`, `password` from `tag3` where `active` = ? and (`email` = ? or `email` = ?) order by `id` asc" {
		assert.Equal(t, []interface{}{1, "email1", "email2"}, b.GetBindings())
		in = true
	} else if sql == "select `email`, `password` from `tag3` where (`email` = ? or `email` = ?) and `active` = ? order by `id` asc" {
		assert.Equal(t, []interface{}{"email1", "email2", 1}, b.GetBindings())
		in = true
	}
	assert.True(t, in)

}
