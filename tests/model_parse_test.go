package tests

import (
	"database/sql"
	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type TestPlainModel struct {
	Id        int64     `goelo:"column:id;"`
	Name      string    `goelo:"column:name;"`
	Email     string    `goelo:"column:email;"`
	CreatedAt time.Time `goelo:"column:created_at;"`
	UpdatedAt time.Time
}

func TestParsePlainStruct(t *testing.T) {
	model := &TestPlainModel{}
	parsed, err := goeloquent.ParseModel(model)
	assert.Nil(t, err)
	assert.False(t, parsed.IsEloquent)
	assert.Equal(t, parsed.Name, "github.com/glitterlip/goeloquent/v2/tests/TestPlainModel")
	assert.Equal(t, parsed.TableName, "test_plain_model")
	assert.Equal(t, len(parsed.FieldsByColumnName), 5)
	assert.Equal(t, len(parsed.FieldsByStructName), 5)
	assert.Nil(t, parsed.PrimaryKeyField)
	assert.False(t, parsed.PrimaryKeyAutoIncrementing)
	assert.Equal(t, parsed.CreatedAt, "")
	assert.Equal(t, parsed.UpdatedAt, "")
	assert.Equal(t, parsed.DeletedAt, "")
	for _, s := range []string{"Id", "Name", "Email", "CreatedAt", "UpdatedAt"} {
		_, ok := parsed.FieldsByStructName[s]
		assert.True(t, ok, "field %s not found in FieldsByStructName", s)
		_, ok = parsed.FieldsByColumnName[goeloquent.ToSnakeCase(s)]
		assert.True(t, ok, "field %s not found in FieldsByColumnName", s)
	}

}

type TestEloqumentModel struct {
	*goeloquent.EloquentModel
	Id         int64        `json:"id" goelo:"column:id;primaryKey;"`
	Name       string       `json:"name" goelo:"column:name;"`
	Email      string       `json:"email" goelo:"column:email;"`
	Roles      UserRoles    `json:"roles" goelo:"column:roles;"`
	CreatedAt  sql.NullTime `json:"createdAt" goelo:"column:created_at;CREATED_AT"`
	UpdatedAt  sql.NullTime `json:"updatedAt" goelo:"column:updated_at;UPDATED_AT"`
	DeletedAt  sql.NullTime `json:"deletedAt" goelo:"column:deleted_at;DELETED_AT"`
	PostsCount int64        `json:"postsCount" goelo:"withAggregate:Posts;"`
}

func (t *TestEloqumentModel) EloquentGetGuarded() map[string]struct{} {
	return map[string]struct{}{
		"id":     {},
		"status": {},
	}
}
func (t *TestEloqumentModel) EloquentGetWithRelations() map[string]goeloquent.RelationFunc {
	return map[string]goeloquent.RelationFunc{
		"Phone": func(builder *goeloquent.Statement) {
			builder.Where("countrycode", "+2")
		},
	}
}

func (t *TestEloqumentModel) EloquentGetWithRelationAggregates() map[string]goeloquent.RelationAggregate {
	return map[string]goeloquent.RelationAggregate{
		"PostCount": {
			FuncName:          "Count",
			Column:            "*",
			RelationFieldName: "Posts",
			Constraint: func(builder *goeloquent.Statement) *goeloquent.Statement {
				builder.Where("status", ">", 1)
				return builder
			},
		},
	}
}
func (t *TestEloqumentModel) EloquentGetDefaultAttributes() map[string]interface{} {
	return map[string]interface{}{
		"status": "active",
		"name":   "default name",
	}
}
func (t *TestEloqumentModel) GetTableName(st *goeloquent.Statement) string {
	if st.Context.Value("table") != nil {
		return "test"
	}
	return "test_eloqument_models"
}
func TestParseEloquentModel(t *testing.T) {
	meta, err := goeloquent.ParseModel(&TestEloqumentModel{})
	assert.Nil(t, err)
	assert.True(t, meta.IsEloquent)
	assert.Equal(t, meta.Name, "github.com/glitterlip/goeloquent/v2/tests/TestEloqumentModel")
	assert.Equal(t, meta.TableName, "test_eloqument_model")
	assert.Equal(t, len(meta.FieldsByColumnName), 7)
	assert.Equal(t, len(meta.FieldsByStructName), 8)
	assert.NotNil(t, meta.PrimaryKeyField)
	assert.Equal(t, meta.PrimaryKeyField.Name, "Id")
	assert.Equal(t, meta.PrimaryKeyField.ColumnName, "id")
	assert.True(t, meta.PrimaryKeyAutoIncrementing)
	assert.True(t, meta.SoftDelete)
	assert.Equal(t, meta.EloquentModelFieldIndex, 0)

	assert.Equal(t, meta.Guards, map[string]struct{}{
		"id":     {},
		"status": {},
	})
	assert.Equal(t, meta.DefaultAttributes, map[string]interface{}{
		"status": "active",
		"name":   "default name",
	})
	st := goeloquent.NewStatement()
	goeloquent.NewQueryBuilder(st)
	st.Table(meta.TableName)
	meta.EagerRelations["Phone"](st)
	assert.Equal(t, st.ToSql(), "select * from `test_eloqument_model` where `countrycode` = ?")
	assert.Equal(t, st.GetBindings(), []interface{}{"+2"})
	assert.Equal(t, meta.EagerRelationAggregates["PostCount"].Column, "*")
	assert.Equal(t, meta.EagerRelationAggregates["PostCount"].FuncName, "Count")
	st = goeloquent.NewStatement()
	goeloquent.NewQueryBuilder(st)
	st.Table(meta.TableName)
	meta.EagerRelationAggregates["PostCount"].Constraint(st)
	assert.Equal(t, st.ToSql(), "select * from `test_eloqument_model` where `status` > ?")
	assert.Equal(t, st.GetBindings(), []interface{}{1})
	model := &TestEloqumentModel{
		EloquentModel: &goeloquent.EloquentModel{},
	}
	assert.Equal(t, model.GetTableName(st), "test_eloqument_models")

}
