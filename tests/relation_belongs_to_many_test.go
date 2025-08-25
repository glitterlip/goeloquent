package tests

import (
	"testing"

	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
)

type BelongsToManyUser struct {
	*goeloquent.EloquentModel
	Id      int64  `json:"id" goelo:"column:id;primaryKey;"`
	Name    string `json:"name" goelo:"column:name;"`
	Account string `json:"account" goelo:"column:account;"`
	Age     int    `json:"age" goelo:"column:age;"`
}

func (b *BelongsToManyUser) GetTableName(st *goeloquent.Statement) string {
	return "users"
}
func (b *BelongsToManyUser) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}

type BelongsToManyRoles struct {
	*goeloquent.EloquentModel
	Id    int64                `json:"id" goelo:"column:id;primaryKey;"`
	Name  string               `json:"name" goelo:"column:name;"`
	Users []*BelongsToManyUser `json:"users" goelo:"BelongsToMany:UsersRelation;"`
}

func (b *BelongsToManyRoles) GetTableName(st *goeloquent.Statement) string {
	return "roles"
}
func (b *BelongsToManyRoles) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}
func (b *BelongsToManyRoles) UsersRelation() *goeloquent.BelongsToManyRelation {
	return b.BelongsToMany(b, &BelongsToManyUser{}, "role_user", "role_id", "user_id", "id", "id")
}

func TestBelongsToManyWithDynamicDefaultUseParentModel(t *testing.T) {
}
func TestSaveMethodSetsForeignKeyOnModelBelongsToMany(t *testing.T) {
}
func TestEagerConstraintsAreProperlyAddedBelongsToMany(t *testing.T) {
	after := "drop table if exists users; drop table if exists roles; drop table if exists role_user;"
	before := after + `create table users (id integer primary key auto_increment,name varchar(255),account varchar(255),age integer);` +
		`create table roles (id integer primary key auto_increment,name varchar(255));` +
		`create table role_user (role_id integer,user_id integer,valid integer);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		_, err := conn.Model(&BelongsToManyRoles{}).Insert([]*BelongsToManyRoles{
			{Name: "customer"},
			{Name: "vip"},
			{Name: "suspended"},
			{Name: "expired"},
			{Name: "empty"},
		})
		assert.Nil(t, err)

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var roles []BelongsToManyRoles
		_, err = conn.Model(&BelongsToManyRoles{}).With("Users").Get(&roles)
		assert.ErrorIs(t, err, goeloquent.ErrorNotFound)
		assert.Equal(t, 2, len(sts))
		assert.Equal(t, "select * from `roles`", sts[0].RawSql)
		assert.Equal(t, "select `users`.*, `role_user`.`role_id` as `goelo_pivot_role_id`, `role_user`.`user_id` as `goelo_pivot_user_id` from `users` inner join `role_user` on `role_user`.`user_id` = `users`.`id` where `role_user`.`role_id` in (?, ?, ?, ?, ?)", sts[1].RawSql)
		assert.Equal(t, []interface{}{int64(1), int64(2), int64(3), int64(4), int64(5)}, sts[1].GetBindings())

	})
}
func TestRelationCountQueryCanBeBuiltBelongsToMany(t *testing.T) {
}
func TestCreateMethodProperlyCreatesNewModelBelongsToMany(t *testing.T) {
}
func TestRelationGetResultsBelongsToMany(t *testing.T) {
}
func TestRelationLoadsResultsBelongsToMany(t *testing.T) {
}
func TestSelfRelationCountBelongsToMany(t *testing.T) {

}
func TestSelectPivot(t *testing.T) {

}
func TestWherePivot(t *testing.T) {

}
