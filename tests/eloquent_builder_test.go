package tests

import (
	"database/sql"
	goeloquent "github.com/glitterlip/go-eloquent"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type User struct {
	goeloquent.EloquentModel
	Id        int64          `goelo:"column:id;primaryKey"`
	UserName  sql.NullString `goelo:"column:name"`
	Age       int            `goelo:"column:age"`
	CreatedAt time.Time      `goelo:"column:created_at,timestatmp:create"`
	UpdatedAt sql.NullTime   `goelo:"column:updated_at,timestatmp:update"`
}

func (user *User) TableName() string {
	return "users"
}

type UserTable struct {
	goeloquent.EloquentModel
	Id        int64          `goelo:"column:User_Sid;primaryKey"`
	UserName  sql.NullString `goelo:"column:name"`
	Age       int            `goelo:"column:age"`
	CreatedAt time.Time      `goelo:"column:create_time;CREATED_AT"`
	UpdatedAt sql.NullTime   `goelo:"column:update_Time;UPDATED_AT"`
}

func (user *UserTable) TableName() string {
	return "user_Table"
}

type UserConnection struct {
	goeloquent.EloquentModel
	Id        int64          `goelo:"column:id;primaryKey"`
	UserName  sql.NullString `goelo:"column:name"`
	Age       int            `goelo:"column:age"`
	CreatedAt time.Time      `goelo:"column:created_at"`
	UpdatedAt sql.NullTime   `goelo:"column:updated_at"`
}

func (user *UserConnection) ConnectionName() string {
	return "chat"
}
func TestParseModel(t *testing.T) {
	type U struct {
		goeloquent.EloquentModel
	}
	var user User
	var userT UserTable
	var userC UserConnection
	var userS []UserConnection
	var userPS []*UserTable
	b := DB.Model(&userT)
	b1 := DB.Model(&userC)
	b2 := DB.Model(&userS)
	b3 := DB.Model(&userPS)
	b4 := DB.Model()
	b5 := DB.Model()
	b6 := DB.Model(&user)
	b7 := DB.Model(&U{})
	b4.First(&userC)
	b5.Get(&userPS)
	//parse table
	assert.Equal(t, "user_Table", b.FromTable)
	assert.Equal(t, "user_Table", b.FromTable)
	assert.Equal(t, "user_connection", b1.FromTable)
	assert.Equal(t, "user_Table", b3.FromTable)
	assert.Equal(t, "user_Table", b5.FromTable)
	assert.Equal(t, "users", b6.FromTable)
	assert.Equal(t, "u", b7.FromTable)
	//pase connection
	assert.Equal(t, "default", b.Connection.ConnectionName)
	assert.Equal(t, "chat", b1.Connection.ConnectionName)
	assert.Equal(t, "chat", b2.Connection.ConnectionName)
	assert.Equal(t, "chat", b4.Connection.ConnectionName)
	//parse field
	assert.Equal(t, "User_Sid", b.Model.PrimaryKey.ColumnName)
	assert.Equal(t, "create_time", b.Model.CreatedAt)
	assert.Equal(t, "update_Time", b.Model.UpdatedAt)

}
func TestFindMethod(t *testing.T) {
	//testFindMethod
	createUsers, dropUsers := UserTableSql()
	now := time.Now()
	u1 := map[string]interface{}{
		"name":       "go-eloquent",
		"age":        18,
		"created_at": now,
	}
	RunWithDB(createUsers, dropUsers, func() {
		var user User
		DB.Table("users").Insert(u1)
		b := DB.Model()
		b.Find(&user, 14)
		assert.Equal(t, "select * from `users` where `id` = ? limit 1", b.ToSql())
		assert.Equal(t, b.GetBindings(), []interface{}{14})

		//testFindManyMethod
		var users []User
		b1 := DB.Model()
		b1.Find(&users, []interface{}{1, 14})
		assert.Equal(t, "select * from `users` where `id` in (?,?)", b1.ToSql())
		assert.Equal(t, []interface{}{1, 14}, b1.GetBindings())
	})

}

func TestFindOrNew(t *testing.T) {
	//testFindOrNew
	//testFindOrFail
}
func TestFirst(t *testing.T) {
	//testFirstMethod
	createUsers, dropUsers := UserTableSql()
	now := time.Now()
	u1 := map[string]interface{}{
		"name":       "go-eloquent",
		"age":        18,
		"created_at": now,
	}
	RunWithDB(createUsers, dropUsers, func() {
		DB.Table("users").Insert(u1)
		var user User
		b := DB.Model()
		b1 := DB.Model()
		b.First(&user)
		b1.First(&user, "id", "name")
		assert.Equal(t, "select * from `users` limit 1", b.ToSql())
		assert.Equal(t, "select `id`, `name` from `users` limit 1", b1.ToSql())
	})
}