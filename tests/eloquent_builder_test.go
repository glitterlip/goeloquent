package tests

import (
	"database/sql"
	"fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type User struct {
	*goeloquent.EloquentModel
	Id        int64          `goelo:"column:id;primaryKey"`
	UserName  sql.NullString `goelo:"column:name"`
	Age       int            `goelo:"column:age"`
	Status    int            `goelo:"column:status"`
	CreatedAt time.Time      `goelo:"column:created_at,timestatmp:create"`
	UpdatedAt sql.NullTime   `goelo:"column:updated_at,timestatmp:update"`
}

func (user *User) TableName() string {
	return "users"
}

type UserTable struct {
	*goeloquent.EloquentModel
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
	*goeloquent.EloquentModel
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
		*goeloquent.EloquentModel
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

func TestQualifyColumn(t *testing.T) {
	//testQualifyColumn
	//testQualifyColumns
}

func TestValueMethod(t *testing.T) {
	//testValueMethodWithModelFound
	//testValueMethodWithModelNotFound
	createUsers, dropUsers := UserTableSql()
	now := time.Now()
	u1 := map[string]interface{}{
		"name":       "eloquent",
		"age":        18,
		"created_at": now,
	}
	RunWithDB(createUsers, dropUsers, func() {
		var age int
		var name, name1, notFound string
		DB.Table("users").Insert(u1)
		b := DB.Table("users").Where("age", 18)
		b1 := DB.Table("users").Where("age", 18)
		b2 := DB.Model(&User{}).Where("age", 18)
		b3 := DB.Model(&User{}).Where("age", 8)
		b.Value(&age, "age")
		b1.Value(&name, "name")
		b2.Value(&name1, "name")
		b3.Value(&notFound, "name")
		assert.Equal(t, "select `age` from `users` where `age` = ? limit 1", b.ToSql())
		assert.Equal(t, "select `name` from `users` where `age` = ? limit 1", b1.ToSql())
		assert.Equal(t, "select `name` from `users` where `age` = ? limit 1", b2.ToSql())
		assert.Equal(t, "select `name` from `users` where `age` = ? limit 1", b3.ToSql())
		assert.Equal(t, 18, age)
		assert.Equal(t, "eloquent", name)
		assert.Equal(t, "eloquent", name1)
		assert.Equal(t, "", notFound)
	})

}

func TestSaveMethod(t *testing.T) {
	c, d := CreateRelationTables()
	RunWithDB(c, d, func() {
		var u1, u2, u3 UserT
		var us []UserT
		u1.UserName = "u1"
		u2.UserName = "u2"
		u3.UserName = "u3"
		u1.Age = 10
		u2.Age = 20
		u3.Age = 30
		_, err := DB.Save(&u1)
		assert.Nil(t, err)
		assert.Greater(t, u1.Id, int64(0))
		u1.Age = 11
		save, err := DB.Save(&u1)
		count, _ := save.RowsAffected()
		assert.Equal(t, int64(1), count)
		assert.Nil(t, err)
		assert.Greater(t, u1.Id, int64(0))
		us = append(us, u2, u3)
		insert, err := DB.Model().Insert(&us)
		assert.Nil(t, err)
		count, _ = insert.RowsAffected()
		assert.Equal(t, int64(2), count)
	})
}
func TestAttrs(t *testing.T) {
	c, d := CreateRelationTables()
	RunWithDB(c, d, func() {
		var u1 UserT
		u1.UserName = "u1"
		_, err := DB.Save(&u1)
		assert.Empty(t, u1.GetDirty())
		assert.Empty(t, u1.GetChanges())
		assert.Nil(t, err)
		assert.Greater(t, u1.Id, int64(0))
		u1.Age = 11
		assert.True(t, u1.IsDirty("Age"))
		assert.Empty(t, u1.GetChanges())
		assert.Equal(t, map[string]interface{}{
			"Age": 11,
		}, u1.GetDirty())
		save, err := DB.Save(&u1)
		count, _ := save.RowsAffected()
		assert.Equal(t, int64(1), count)
		assert.Nil(t, err)
		assert.Greater(t, u1.Id, int64(0))
		assert.Empty(t, u1.GetDirty())
		assert.Equal(t, map[string]interface{}{
			"Age": 11,
		}, u1.GetChanges())

	})
}

func TestEvents(t *testing.T) {
	c, d := CreateRelationTables()
	RunWithDB(c, d, func() {
		//test saving,saved
		var u1, u2 UserT
		u1.UserName = "u1"
		_, err := DB.Save(&u1)
		assert.Nil(t, err)
		assert.Greater(t, u1.Id, int64(0))
		assert.Equal(t, 18, u1.Age)
		u2.UserName = "Bob"
		u2.Age = 200
		res, err := DB.Save(&u2)
		assert.Equal(t, err.Error(), "too old")
		assert.Nil(t, res)
		var avatar Image
		DB.Model().Where("imageable_id", u1.Id).Where("imageable_type", "users").First(&avatar)
		assert.Equal(t, avatar.ImageableId, u1.Id)
		assert.Equal(t, avatar.ImageableType, "users")

	})
	RunWithDB(c, d, func() {
		//test creating,created
		var u3, u4 UserT
		var post Post
		u3.Id = 10
		r, err := DB.Save(&u3)
		assert.Nil(t, r)
		assert.Equal(t, "wrong id", err.Error())
		u4.UserName = "Frank"
		DB.Save(&u4)
		_, err = DB.Model().Where("author_id", u4.Id).First(&post)
		assert.Nil(t, err)
		assert.Equal(t, post.AuthorId, u4.Id)
	})
	RunWithDB(c, d, func() {
		//test updating,updated
		var u6, u7 UserT
		u6.UserName = "Alice"
		DB.Save(&u6)
		DB.Model().Find(&u7, u6.Id)
		u7.Id = u7.Id + 10
		save, err := u7.Save()
		assert.Nil(t, save)
		assert.Equal(t, err.Error(), "id can not be changed")
		u6.UserName = "Alexa"
		u6.Save()
		newUrl := fmt.Sprintf("cdn.com/statis/users/%d/avatar-new.png", u6.Id)
		var avatar Image
		DB.Model().Where("imageable_id", u6.Id).Where("imageable_type", "users").First(&avatar)
		assert.Equal(t, newUrl, avatar.Url)
	})
	RunWithDB(c, d, func() {
		//test deleting deleted
		var u1, u2 UserT
		var post Post
		u1.UserName = "Alice"
		u2.UserName = "Joe"
		DB.Save(&u1)
		DB.Save(&u2)
		DB.Model().Where("author_id", u1.Id).First(&post)
		assert.Equal(t, post.AuthorId, u1.Id)
		_, err := u1.Delete()
		assert.Equal(t, "can't delete admin", err.Error())
		_, err = u2.Delete()
		n := 1
		DB.Table("post").Where("author_id", u2.Id).Count(&n)
		assert.Equal(t, 0, n)

	})
	//test mute events
	RunWithDB(c, d, func() {
		//test deleting deleted
		var u1, u2 UserT
		var post Post
		var image Image
		u1.UserName = "Alice"
		u2.UserName = "Bob"
		DB.Boot(&u1)
		DB.Boot(&u2)
		u1.Mute(goeloquent.EventALL)
		u2.Mute(goeloquent.EventSaving)
		DB.Save(&u1)
		DB.Save(&u2)
		var count int64
		imageR, err := DB.Table("image").Where("imageable_id", u1.Id).Where("imageable_type", "users").First(&image)
		DB.Table("post").Where("author_id", u2.Id).First(&post)
		assert.Nil(t, err)
		count, _ = imageR.RowsAffected()
		assert.Equal(t, int64(0), count)
		assert.Equal(t, u2.Id, post.AuthorId)

	})
}
func TestTimeStamps(t *testing.T) {

}
