package tests

import (
	"database/sql"
	"testing"

	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
)

type EloquentUser struct {
	*goeloquent.EloquentModel
	Id        int64        `json:"id" goelo:"column:id;primaryKey;"`
	Name      string       `json:"name" goelo:"column:name;"`
	Email     string       `json:"email" goelo:"column:email;"`
	CreatedAt sql.NullTime `json:"createdAt" goelo:"column:created_at;CREATED_AT"`
	UpdatedAt sql.NullTime `json:"updatedAt" goelo:"column:updated_at;UPDATED_AT"`
	DeletedAt sql.NullTime `json:"deletedAt" goelo:"column:deleted_at;DELETED_AT"`
}

func (e *EloquentUser) GetTableName(st *goeloquent.Statement) string {
	return "eloquent_users"
}
func (e *EloquentUser) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}
func TestFindMethod(t *testing.T) {
	after := "drop table if exists eloquent_users;"
	before := after + "create table eloquent_users (id int auto_increment primary key, name varchar(255), email varchar(255), created_at datetime, updated_at datetime, deleted_at datetime);"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		user := &EloquentUser{}
		st, err := conn.Model(user).Find(user, 1)
		assert.ErrorIs(t, err, goeloquent.ErrorNotFound)
		assert.Equal(t, st.RawSql, "select * from `eloquent_users` where `id` = ? limit 1")
		assert.Equal(t, st.GetBindings(), []interface{}{1})

		user.Init(user)
		st, err = user.Save(map[string]interface{}{
			"name":  "John Doe",
			"email": "john@gmail.com",
		})

		assert.Nil(t, err)
		assert.Equal(t, st.RowsAffected(), int64(1))
		var user1 EloquentUser

		st, err = conn.Model(user).Find(&user1, 1)
		assert.NoError(t, err)
		assert.Equal(t, st.RawSql, "select * from `eloquent_users` where `id` = ? limit 1")
		assert.Equal(t, st.GetBindings(), []interface{}{1})
		assert.Equal(t, user1.Id, int64(1))
		assert.Equal(t, user1.Name, "John Doe")
		assert.Equal(t, user1.Email, "john@gmail.com")
		assert.NotEmpty(t, user1.CreatedAt)
		assert.NotEmpty(t, user1.UpdatedAt)
		assert.Empty(t, user1.DeletedAt)

	})

}

func TestFirstMethod(t *testing.T) {
	after := "drop table if exists eloquent_users;"
	before := after + "create table eloquent_users (id int auto_increment primary key, name varchar(255), email varchar(255), created_at datetime, updated_at datetime, deleted_at datetime);"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		user := &EloquentUser{}
		st, err := conn.Model(user).Find(user, 1)
		assert.ErrorIs(t, err, goeloquent.ErrorNotFound)
		assert.Equal(t, st.RawSql, "select * from `eloquent_users` where `id` = ? limit 1")
		assert.Equal(t, st.GetBindings(), []interface{}{1})

		user.Init(user)
		st, err = user.Save(map[string]interface{}{
			"name":  "John Doe",
			"email": "john@gmail.com",
		})

		assert.Nil(t, err)
		assert.Equal(t, st.RowsAffected(), int64(1))
		var user1 EloquentUser

		st, err = conn.Model(user).Where("name", "John Doe").First(&user1, []interface{}{"id", "name", "email"})
		assert.NoError(t, err)
		assert.Equal(t, st.RawSql, "select `id`, `name`, `email` from `eloquent_users` where `name` = ? limit 1")
		assert.Equal(t, st.GetBindings(), []interface{}{"John Doe"})
		assert.Equal(t, user1.Id, int64(1))
		assert.Equal(t, user1.Name, "John Doe")
		assert.Equal(t, user1.Email, "john@gmail.com")
		assert.Empty(t, user1.CreatedAt)
		assert.Empty(t, user1.UpdatedAt)
		assert.Empty(t, user1.DeletedAt)

	})
}
func TestQualifyColumn(t *testing.T) {
	conn := GetConnection()
	eb := conn.Model(&EloquentUser{})
	assert.Equal(t, eb.QualifyColumn("name"), "eloquent_users.name")
	assert.Equal(t, eb.QualifyColumn("id"), "eloquent_users.id")
	assert.Equal(t, eb.QualifyColumn("email"), "eloquent_users.email")
	assert.Equal(t, eb.QualifyColumn("created_at"), "eloquent_users.created_at")
}
