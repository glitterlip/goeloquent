package tests

import (
	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

type FillableModel struct {
	*goeloquent.EloquentModel
	Id    int64  `json:"id" goelo:"column:id;primaryKey;"`
	Name  string `json:"name" goelo:"column:name;"`
	Email string `json:"email" goelo:"column:email;"`
}

func (f *FillableModel) GetFillable() map[string]struct{} {
	return map[string]struct{}{
		"name": {},
	}
}
func TestFillables(t *testing.T) {

	model := &FillableModel{}
	parsed, err := goeloquent.ParseModel(model)
	assert.Nil(t, err)

	assert.Equal(t, map[string]struct{}{
		"Name": {},
		"name": {},
	}, parsed.Fillables)

	model.Init(model).Fill(map[string]interface{}{
		"name":  "test",
		"email": "john@github.com",
		"Email": "john@github.com",
	})

	assert.Equal(t, "test", model.Name)
	assert.Equal(t, "", model.Email)
}

type GuardedModel struct {
	*goeloquent.EloquentModel
	Id    int64  `json:"id" goelo:"column:id;primaryKey;"`
	Name  string `json:"name" goelo:"column:name;"`
	Email string `json:"email" goelo:"column:email;"`
}

func (g *GuardedModel) GetGuarded() map[string]struct{} {
	return map[string]struct{}{
		"name": {},
	}
}
func TestGuards(t *testing.T) {
	model := &GuardedModel{}
	parsed, err := goeloquent.ParseModel(model)
	assert.Nil(t, err)
	assert.Equal(t, map[string]struct{}{
		"Name": {},
		"name": {},
	}, parsed.Guards)

	model.Init(model).Fill(map[string]interface{}{
		"name":  "test",
		"Name":  "test",
		"email": "john@github.com",
	})
	assert.Equal(t, "", model.Name)
	assert.Equal(t, "john@github.com", model.Email)
}

type ConflictModel struct {
	*goeloquent.EloquentModel
	Id    int64  `json:"id" goelo:"column:id;primaryKey;"`
	Name  string `json:"name" goelo:"column:name;"`
	Email string `json:"email" goelo:"column:email;"`
}

func (c *ConflictModel) GetGuarded() map[string]struct{} {
	return map[string]struct{}{
		"name": {},
	}
}
func (c *ConflictModel) GetFillable() map[string]struct{} {
	return map[string]struct{}{
		"email": {},
	}
}
func TestConflicts(t *testing.T) {
	_, err := goeloquent.ParseModel(&ConflictModel{})
	assert.Equal(t, err.Error(), "Parse model failed:github.com/glitterlip/goeloquent/v2/tests/ConflictModel can not use guarded with fillable")
}