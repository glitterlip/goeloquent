package tests

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
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

type DefaultModel struct {
	*goeloquent.EloquentModel
	Id    int64  `json:"id" goelo:"column:id;primaryKey;"`
	Name  string `json:"name" goelo:"column:name;"`
	Email string `json:"email" goelo:"column:email;"`
	Role  string `json:"role" goelo:"column:role;"`
}

func (d *DefaultModel) GetDefaults() map[string]interface{} {
	return map[string]interface{}{
		"role": "user",
	}
}

func TestDefaultAttributes(t *testing.T) {
	model := &DefaultModel{}
	parsed, err := goeloquent.ParseModel(model)
	assert.Nil(t, err)

	assert.Equal(t, map[string]interface{}{
		"role": "user",
		"Role": "user",
	}, parsed.DefaultAttributes)

	model.Init(model).Fill(map[string]interface{}{
		"id":   int64(1),
		"name": "test",
	})

	assert.Equal(t, "user", model.Role)
}

type OriginalModel struct {
	*goeloquent.EloquentModel
	Id    int64  `json:"id" goelo:"column:id;primaryKey;"`
	Name  string `json:"name" goelo:"column:name;"`
	Email string `json:"email" goelo:"column:email;"`
}

func TestGetOriginal(t *testing.T) {
	model := &OriginalModel{
		Name:  "original",
		Email: "original@gmail.com",
	}
	_, err := goeloquent.ParseModel(model)
	model.Init(model)

	model.Name = "changed"
	model.Email = "changed@gmail.com"
	assert.Nil(t, err)
	assert.Equal(t, "original", model.GetOriginal("name"))
	assert.Equal(t, "original@gmail.com", model.GetOriginal("email"))

}

type DirtyModel struct {
	*goeloquent.EloquentModel
	Id    int64        `json:"id" goelo:"column:id;primaryKey;"`
	Name  string       `json:"name" goelo:"column:name;"`
	Email string       `json:"email" goelo:"column:email;"`
	Time  sql.NullTime `json:"time" goelo:"column:time;"`
	Bool  bool         `json:"bool" goelo:"column:bool;"`
}

func (d *DirtyModel) GetTableName(st *goeloquent.Statement) string {
	return "dirty_model"
}
func (d *DirtyModel) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}
func TestIsDirty(t *testing.T) {
	now := time.Now()
	model := &DirtyModel{
		Name:  "original",
		Email: "original@gmail.com",
		Time:  sql.NullTime{Valid: true, Time: now.Add(time.Hour)},
		Bool:  true,
	}
	model.Init(model).Fill(map[string]interface{}{
		"Name":  "new",
		"Email": "new@github.com",
	})
	model.Time = sql.NullTime{Valid: true, Time: now.Add(time.Hour * 2)}
	model.Bool = false

	assert.True(t, model.IsDirty("Name"))
	assert.True(t, model.IsDirty("Email"))
	assert.True(t, model.IsDirty("Time"))
	assert.True(t, model.IsDirty("Bool"))

}

func TestGetDirtyAttributes(t *testing.T) {
	now := time.Now()
	model := &DirtyModel{
		Name:  "original",
		Email: "original@gmail.com",
		Time:  sql.NullTime{Valid: true, Time: now.Add(time.Hour)},
		Bool:  true,
	}
	model.Init(model).Fill(map[string]interface{}{
		"name":  "new",
		"email": "new@github.com",
	})
	assert.Equal(t, map[string]interface{}{
		"name":  "new",
		"email": "new@github.com",
	}, model.GetDirty())
	model.Time = sql.NullTime{Valid: true, Time: now.Add(time.Hour * 2)}
	model.Bool = false
	assert.Equal(t, map[string]interface{}{
		"name":  "new",
		"email": "new@github.com",
		"bool":  false,
		"time":  sql.NullTime{Valid: true, Time: now.Add(time.Hour * 2)},
	}, model.GetDirty())
}
func TestGetChangedAttributes(t *testing.T) {
	after := "drop table if exists `dirty_model`"
	before := "create table `dirty_model` (`id` bigint unsigned not null auto_increment primary key, `name` varchar(255) not null, `email` varchar(255) not null, `time` datetime null, `bool` boolean not null default false) charset=utf8mb4 collate=utf8mb4_general_ci;"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		now := time.Now()
		model := &DirtyModel{
			Name:  "original",
			Email: "original@gmail.com",
			Time:  sql.NullTime{Valid: true, Time: now.Add(time.Hour)},
			Bool:  true,
		}
		model.Init(model).Fill(map[string]interface{}{
			"Name":  "new",
			"Email": "new@github.com",
		})
		model.Time = sql.NullTime{Valid: true, Time: now.Add(time.Hour * 2)}
		model.Bool = false
		model.Save()

		assert.Nil(t, model.Error)
		assert.Equal(t, model.RawSql, "insert into `dirty_model` (`bool`, `email`, `name`, `time`) values (?, ?, ?, ?)")
		assert.Equal(t, map[string]interface{}{
			"name":  "new",
			"email": "new@github.com",
			"time":  sql.NullTime{Valid: true, Time: now.Add(time.Hour * 2)},
			"bool":  false,
		}, model.GetChanges())
	})

}