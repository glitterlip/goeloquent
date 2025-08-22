package tests

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
)

type HasOneUser struct {
	*goeloquent.EloquentModel
	Id             int64           `json:"id" goelo:"column:id;primaryKey;"`
	Name           string          `json:"name" goelo:"column:name;"`
	Email          string          `json:"email" goelo:"column:email;"`
	HasOneUserInfo *HasOneUserInfo `json:"HasOneUserInfo" goelo:"HasOne:UserInfoRelation;"`
	CreatedAt      sql.NullTime    `json:"createdAt" goelo:"column:created_at;CREATED_AT"`
	UpdatedAt      sql.NullTime    `json:"updatedAt" goelo:"column:updated_at;UPDATED_AT"`
	DeletedAt      sql.NullTime    `json:"deletedAt" goelo:"column:deleted_at;"`
}

func (h *HasOneUser) GetTableName(st *goeloquent.Statement) string {
	return "models"
}

func (h *HasOneUser) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}
func (h *HasOneUser) UserInfoRelation() *goeloquent.HasOneRelation {
	r := h.HasOne(h, &HasOneUserInfo{}, "id", "user_id")
	r.WithDefault(func(parentModel interface{}) interface{} {
		user := parentModel.(*HasOneUser)
		return &HasOneUserInfo{
			Meta: MetaInfo{
				Address:  fmt.Sprintf("default address for user %d", user.Id),
				Age:      18,
				Verified: false,
				Tags:     []string{"user"},
			},
		}
	})
	return r
}

type HasOneUserInfo struct {
	*goeloquent.EloquentModel
	Id     int64    `json:"id" goelo:"column:id;primaryKey;autoIncrement:false"`
	UserId int64    `json:"userId" goelo:"column:user_id;"`
	Meta   MetaInfo `json:"meta" goelo:"column:meta;cast:json;"`
}

func (u *HasOneUserInfo) GetTableName(st *goeloquent.Statement) string {
	return "user_info"
}

type MetaInfo struct {
	Address  string   `json:"address"`
	Age      int      `json:"age"`
	Verified bool     `json:"verified"`
	Tags     []string `json:"tags"`
}

func (u *HasOneUserInfo) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}
func TestHasOneWithDynamicDefaultUseParentModel(t *testing.T) {

}
func TestSaveMethodSetsForeignKeyOnModel(t *testing.T) {

}
func TestEagerConstraintsAreProperlyAdded(t *testing.T) {
	after := "drop table if exists models; drop table if exists user_info;"
	before := after + "create table models (id int auto_increment  primary key, name varchar(255), email varchar(255), roles varchar(255), created_at datetime, updated_at datetime, deleted_at datetime); " +
		"create table user_info (id int auto_increment primary key , user_id int, meta json)"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		user1 := &HasOneUser{
			Name:  "John Doe",
			Email: "test@gmail.com",
		}
		user2 := &HasOneUser{
			Name:  "Jane Doe1",
			Email: "test2@gmail.com",
		}
		st, err := user1.Init(user1).Save()
		user2.Init(user2).Save()
		assert.Nil(t, err)
		assert.Equal(t, "insert into `models` (`created_at`, `deleted_at`, `email`, `name`, `updated_at`) values (?, ?, ?, ?, ?)", st.RawSql)
		assert.Equal(t, user1.Id, int64(1))

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})

		var users []*HasOneUser
		_, err = conn.Model(&users).With("HasOneUserInfo").Get(&users)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, 2, len(sts))
		assert.Equal(t, "select * from `user_info` where `user_info`.`user_id` in (?, ?)", sts[1].RawSql)
		assert.Equal(t, []interface{}{int64(1), int64(2)}, sts[1].GetBindings())
	})

}
func TestModelsAreProperlyMatchedToParents(t *testing.T) {
	after := "drop table if exists models; drop table if exists user_info;"
	before := after + "create table models (id int auto_increment  primary key, name varchar(255), email varchar(255), roles varchar(255), created_at datetime, updated_at datetime, deleted_at datetime); " +
		"create table user_info (id int auto_increment primary key , user_id int, meta json)"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("models").Insert([]map[string]interface{}{
			{"name": "John Doe", "email": "qqq"},
			{"name": "Jane Doe1", "email": "www"},
		})
		conn.Table("user_info").Insert([]map[string]interface{}{
			{"user_id": 3, "meta": `{"address":"address1","age":20,"verified":true,"tags":["tag1","tag2"]}`},
			{"user_id": 1, "meta": `{"address":"address2","age":30,"verified":false,"tags":["tag3","tag4"]}`},
			{"user_id": 2, "meta": `{"address":"address2","age":25,"verified":false,"tags":["tag3","tag4"]}`},
		})
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var users []*HasOneUser
		_, err := conn.Model(&users).With("HasOneUserInfo").Get(&users)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, 2, len(sts))
		assert.Equal(t, "select * from `user_info` where `user_info`.`user_id` in (?, ?)", sts[1].RawSql)
		assert.Equal(t, []interface{}{int64(1), int64(2)}, sts[1].GetBindings())
		for _, user := range users {
			assert.NotZero(t, user.HasOneUserInfo)
			assert.Equal(t, user.Id, user.HasOneUserInfo.UserId)
		}

	})

}
func TestRelationCountQueryCanBeBuilt(t *testing.T) {
	after := "drop table if exists models; drop table if exists user_info;"
	before := after + "create table models (id int auto_increment  primary key, name varchar(255), email varchar(255), roles varchar(255), created_at datetime, updated_at datetime, deleted_at datetime); " +
		"create table user_info (id int auto_increment primary key , user_id int, meta json)"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("models").Insert([]map[string]interface{}{
			{"name": "John Doe", "email": "qqq"},
			{"name": "Jane Doe1", "email": "www"},
		})
		conn.Table("user_info").Insert([]map[string]interface{}{
			{"user_id": 3, "meta": `{"address":"address1","age":20,"verified":true,"tags":["tag1","tag2"]}`},
			{"user_id": 1, "meta": `{"address":"address2","age":30,"verified":false,"tags":["tag3","tag4"]}`},
			{"user_id": 2, "meta": `{"address":"address2","age":25,"verified":false,"tags":["tag3","tag4"]}`},
		})
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var users []*HasOneUser
		r, err := conn.Model(&users).Has("HasOneUserInfo").With("HasOneUserInfo").Get(&users)
		assert.Nil(t, err)
		assert.Equal(t, r.RawSql, "select * from `models` where exists (select * from `user_info` where `models`.`id` = `user_info`.`user_id`)")
		assert.Equal(t, 2, len(users))
		for _, user := range users {
			user.HasOneUserInfo.UserId = user.Id
		}
		var users1 []*HasOneUser

		r, err = conn.Model(&users).WhereHas("HasOneUserInfo", func(st *goeloquent.Statement) {
			st.Where("meta->age", ">", 25)
		}).With("HasOneUserInfo").Get(&users1)

		assert.Nil(t, err)
		assert.Equal(t, r.RawSql, "select * from `models` where exists (select * from `user_info` where `models`.`id` = `user_info`.`user_id` and json_unquote(json_extract(`meta`, '$.\"age\"')) > ?)")
		assert.Equal(t, len(users1), 1)
		assert.Equal(t, users1[0].HasOneUserInfo.UserId, users1[0].Id)
		assert.Equal(t, users1[0].HasOneUserInfo.Meta.Age, 30)
	})
}
func TestRelationGetResults(t *testing.T) {
	after := "drop table if exists models; drop table if exists user_info;"
	before := after + "create table models (id int auto_increment  primary key, name varchar(255), email varchar(255), roles varchar(255), created_at datetime, updated_at datetime, deleted_at datetime); " +
		"create table user_info (id int auto_increment primary key , user_id int, meta json)"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("models").Insert([]map[string]interface{}{
			{"name": "John Doe", "email": "qqq"},
			{"name": "Jane Doe1", "email": "www"},
		})
		conn.Table("user_info").Insert([]map[string]interface{}{
			{"user_id": 3, "meta": `{"address":"address1","age":20,"verified":true,"tags":["tag1","tag2"]}`},
			{"user_id": 1, "meta": `{"address":"address2","age":30,"verified":false,"tags":["tag3","tag4"]}`},
			{"user_id": 2, "meta": `{"address":"address2","age":25,"verified":false,"tags":["tag3","tag4"]}`},
		})
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})

		var u HasOneUser
		r, err := conn.Model(&u).First(&u)
		assert.Equal(t, u.Id, int64(1))
		var info HasOneUserInfo
		r, err = u.UserInfoRelation().First(&info)
		assert.Nil(t, err)
		assert.Equal(t, info.UserId, u.Id)
		assert.Equal(t, info.Meta.Age, 30)
		assert.Equal(t, "select * from `user_info` where `user_info`.`user_id` = ? and `user_info`.`user_id` is not null limit 1", r.RawSql)
		assert.Equal(t, r.GetBindings(), []interface{}{int64(1)})
	})
}
