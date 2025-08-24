package tests

import (
	"testing"

	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
)

type BelongsToUser struct {
	*goeloquent.EloquentModel
	Id       int64  `json:"id" goelo:"column:id;primaryKey;"`
	Name     string `json:"name" goelo:"column:name;"`
	Account  string `json:"account" goelo:"column:account;"`
	Verified int    `json:"verified" goelo:"column:verified;"`
}

func (b *BelongsToUser) GetTableName(st *goeloquent.Statement) string {
	return "users"
}
func (b *BelongsToUser) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}

type BelongsToUserProfile struct {
	*goeloquent.EloquentModel
	Id      int64          `json:"id" goelo:"column:id;primaryKey;"`
	UserId  int64          `json:"userId" goelo:"column:user_id;"`
	Country string         `json:"country" goelo:"column:country;"`
	User    *BelongsToUser `json:"user" goelo:"BelongsTo:UserRelation;"`
	Address string         `json:"address" goelo:"column:address;"`
}

func (b *BelongsToUserProfile) GetTableName(st *goeloquent.Statement) string {
	return "user_profiles"
}
func (b *BelongsToUserProfile) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}

func (b *BelongsToUserProfile) UserRelation() *goeloquent.BelongsToRelation {
	return b.BelongsToRelation(b, &BelongsToUser{}, "user_id", "id")
}

func TestBelongsToWithDynamicDefaultUseParentModel(t *testing.T) {
}
func TestSaveMethodSetsForeignKeyOnModelBelongsTo(t *testing.T) {
}
func TestEagerConstraintsAreProperlyAddedBelongsTo(t *testing.T) {
	after := "drop table if exists users; drop table if exists user_profiles;"
	before := after + `create table users (id integer primary key auto_increment,name varchar(255),account varchar(255),verified integer);create table user_profiles (id integer primary key auto_increment,user_id integer,country varchar(255),address varchar(255));`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		_, err := conn.Model(&BelongsToUserProfile{}).Insert([]*BelongsToUserProfile{
			{UserId: 2, Country: "USA", Address: "123 Main St"},
			{UserId: 1, Country: "Canada", Address: "456 Maple Ave"},
			{UserId: 4, Country: "Unknown", Address: "789 Elm St"},
		})
		assert.Nil(t, err)
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var profiles []BelongsToUserProfile
		conn.Model(&BelongsToUserProfile{}).With("User").Get(&profiles)
		assert.Equal(t, 2, len(sts))
		assert.Equal(t, "select * from `user_profiles`", sts[0].RawSql)
		assert.Equal(t, "select * from `users` where `users`.`id` in (?, ?, ?)", sts[1].RawSql)
		assert.Equal(t, []interface{}{int64(2), int64(1), int64(4)}, sts[1].GetBindings())

	})

}
func TestModelsAreProperlyMatchedToParentsBelongsTo(t *testing.T) {
	after := "drop table if exists users; drop table if exists user_profiles;"
	before := after + `create table users (id integer primary key auto_increment,name varchar(255),account varchar(255),verified integer);create table user_profiles (id integer primary key auto_increment,user_id integer,country varchar(255),address varchar(255));`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		user1 := &BelongsToUser{
			Name:     "John Doe",
			Account:  "johndoe",
			Verified: 1,
		}
		user2 := &BelongsToUser{
			Name:     "Jane Doe",
			Account:  "janedoe",
			Verified: 1,
		}
		_, err := conn.Model(user2).Insert([]*BelongsToUser{
			user1, user2,
		})
		assert.Nil(t, err)
		_, err = conn.Model(&BelongsToUserProfile{}).Insert([]*BelongsToUserProfile{
			{UserId: 2, Country: "USA", Address: "123 Main St"},
			{UserId: 1, Country: "Canada", Address: "456 Maple Ave"},
			{UserId: 4, Country: "Unknown", Address: "789 Elm St"},
		})
		assert.Nil(t, err)
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var profiles []BelongsToUserProfile
		_, err = conn.Model(&BelongsToUserProfile{}).With("User").Get(&profiles)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(profiles))
		assert.Equal(t, "Jane Doe", profiles[0].User.Name)
		assert.Equal(t, "John Doe", profiles[1].User.Name)
		assert.Empty(t, profiles[2].User)

	})
}
func TestRelationCountQueryCanBeBuiltBelongsTo(t *testing.T) {
	after := "drop table if exists users; drop table if exists user_profiles;"
	before := after + `create table users (id integer primary key auto_increment,name varchar(255),account varchar(255),verified integer);create table user_profiles (id integer primary key auto_increment,user_id integer,country varchar(255),address varchar(255));`

	RunWithDB(before, after, func(conn goeloquent.Connection) {
		user1 := &BelongsToUser{
			Name:     "John Doe",
			Account:  "johndoe",
			Verified: 1,
		}
		user2 := &BelongsToUser{
			Name:     "Jane Doe",
			Account:  "janedoe",
			Verified: 0,
		}
		_, err := conn.Model(user2).Insert([]*BelongsToUser{
			user1, user2,
		})
		assert.Nil(t, err)
		_, err = conn.Model(&BelongsToUserProfile{}).Insert([]*BelongsToUserProfile{
			{UserId: 2, Country: "USA", Address: "123 Main St"},
			{UserId: 1, Country: "Canada", Address: "456 Maple Ave"},
			{UserId: 4, Country: "Unknown", Address: "789 Elm St"},
		})
		assert.Nil(t, err)
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var profiles []BelongsToUserProfile
		_, err = conn.Model(&BelongsToUserProfile{}).With("User").Has("User").Get(&profiles)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(profiles))
		assert.Equal(t, 2, len(sts))
		assert.Equal(t, "select * from `user_profiles` where exists (select * from `users` where `user_profiles`.`user_id` = `users`.`id`)", sts[0].RawSql)
		assert.Equal(t, "select * from `users` where `users`.`id` in (?, ?)", sts[1].RawSql)
		assert.Equal(t, []interface{}{int64(2), int64(1)}, sts[1].GetBindings())
		sts = []*goeloquent.Statement{}
		var profiles2 []BelongsToUserProfile
		_, err = conn.Model(&BelongsToUserProfile{}).With("User").WhereHas("User", func(q *goeloquent.Statement) {
			q.Where("verified", 1)
		}).Get(&profiles2)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(profiles2))
		assert.Equal(t, "John Doe", profiles2[0].User.Name)
		assert.Equal(t, 2, len(sts))
		assert.Equal(t, "select * from `user_profiles` where exists (select * from `users` where `user_profiles`.`user_id` = `users`.`id` and `verified` = ?)", sts[0].RawSql)
		assert.Equal(t, []interface{}{1}, sts[0].GetBindings())
		assert.Equal(t, "select * from `users` where `users`.`id` in (?)", sts[1].RawSql)
		assert.Equal(t, []interface{}{int64(1)}, sts[1].GetBindings())

	})
}
func TestCreateMethodProperlyCreatesNewModelBelongsTo(t *testing.T) {
}
func TestRelationGetResultsBelongsTo(t *testing.T) {
	after := "drop table if exists users; drop table if exists user_profiles;"
	before := after + `create table users (id integer primary key auto_increment,name varchar(255),account varchar(255),verified integer);create table user_profiles (id integer primary key auto_increment,user_id integer,country varchar(255),address varchar(255));`

	RunWithDB(before, after, func(conn goeloquent.Connection) {
		user1 := &BelongsToUser{
			Name:     "John Doe",
			Account:  "johndoe",
			Verified: 1,
		}
		user2 := &BelongsToUser{
			Name:     "Jane Doe",
			Account:  "janedoe",
			Verified: 0,
		}
		_, err := conn.Model(user2).Insert([]*BelongsToUser{
			user1, user2,
		})
		assert.Nil(t, err)
		_, err = conn.Model(&BelongsToUserProfile{}).Insert([]*BelongsToUserProfile{
			{UserId: 2, Country: "USA", Address: "123 Main St"},
			{UserId: 1, Country: "Canada", Address: "456 Maple Ave"},
			{UserId: 4, Country: "Unknown", Address: "789 Elm St"},
		})
		assert.Nil(t, err)
		var profile BelongsToUserProfile
		_, err = conn.Model(&BelongsToUserProfile{}).First(&profile)
		assert.Nil(t, err)
		var user BelongsToUser
		st, err := profile.UserRelation().Get(&user)
		assert.Nil(t, err)
		assert.Equal(t, "select * from `users` where `users`.`id` = ? and `users`.`id` is not null", st.RawSql)
		assert.Equal(t, []interface{}{int64(2)}, st.GetBindings())
		assert.Equal(t, "Jane Doe", user.Name)

	})
}
func TestRelationLoadsResultsBelongsTo(t *testing.T) {
}
