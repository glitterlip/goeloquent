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
func TestModelsAreProperlyMatchedToParentsBelongsToMany(t *testing.T) {
	after := "drop table if exists users; drop table if exists roles; drop table if exists role_user;"
	before := after + `create table users (id integer primary key auto_increment,name varchar(255),account varchar(255),age integer);` +
		`create table roles (id integer primary key auto_increment,name varchar(255));` +
		`create table role_user (role_id integer not null,user_id integer not null,valid integer not null);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		_, err := conn.Model(&BelongsToManyUser{}).Insert([]*BelongsToManyUser{
			{Name: "User1", Account: "user1", Age: 20},
			{Name: "User2", Account: "user2", Age: 25},
			{Name: "User3", Account: "user3", Age: 30},
			{Name: "User4", Account: "user4", Age: 35},
			{Name: "User5", Account: "user5", Age: 40},
			{Name: "User6", Account: "user6", Age: 45},
		})
		assert.Nil(t, err)
		_, err = conn.Model(&BelongsToManyRoles{}).Insert([]*BelongsToManyRoles{
			{Name: "customer"},
			{Name: "vip"},
			{Name: "suspended"},
			{Name: "expired"},
			{Name: "empty"},
		})
		assert.Nil(t, err)

		_, err = conn.Model(&goeloquent.EloquentModel{}).Table("role_user").Insert([]map[string]interface{}{
			{"role_id": 1, "user_id": 2, "valid": 1},
			{"role_id": 1, "user_id": 1, "valid": 1},
			{"role_id": 2, "user_id": 3, "valid": 1},
			{"role_id": 2, "user_id": 4, "valid": 1},
			{"role_id": 3, "user_id": 5, "valid": 0},
			{"role_id": 4, "user_id": 5, "valid": 1},
			{"role_id": 5, "user_id": 7, "valid": 1},
		})

		assert.Nil(t, err)
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var roles []BelongsToManyRoles
		_, err = conn.Model(&BelongsToManyRoles{}).With("Users").Get(&roles)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(sts))
		assert.Equal(t, "select * from `roles`", sts[0].RawSql)
		assert.Equal(t, "select `users`.*, `role_user`.`role_id` as `goelo_pivot_role_id`, `role_user`.`user_id` as `goelo_pivot_user_id` from `users` inner join `role_user` on `role_user`.`user_id` = `users`.`id` where `role_user`.`role_id` in (?, ?, ?, ?, ?)", sts[1].RawSql)
		assert.Equal(t, []interface{}{int64(1), int64(2), int64(3), int64(4), int64(5)}, sts[1].GetBindings())
		assert.Equal(t, 5, len(roles))
		for _, role := range roles {
			if role.Id == 5 {
				assert.Empty(t, role.Users)
			} else {
				assert.NotEmpty(t, role.Users)
				for _, user := range role.Users {
					assert.Equal(t, int32(role.Id), user.Pivot["role_id"])
				}
			}
		}

	})
}
func TestRelationCountQueryCanBeBuiltBelongsToMany(t *testing.T) {
}
func TestCreateMethodProperlyCreatesNewModelBelongsToMany(t *testing.T) {
}
func TestRelationGetResultsBelongsToMany(t *testing.T) {
	after := "drop table if exists users; drop table if exists roles; drop table if exists role_user;"
	before := after + `create table users (id integer primary key auto_increment,name varchar(255),account varchar(255),age integer);` +
		`create table roles (id integer primary key auto_increment,name varchar(255));` +
		`create table role_user (role_id integer not null,user_id integer not null,valid integer not null);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		_, err := conn.Model(&BelongsToManyUser{}).Insert([]*BelongsToManyUser{
			{Name: "User1", Account: "user1", Age: 20},
			{Name: "User2", Account: "user2", Age: 25},
			{Name: "User3", Account: "user3", Age: 30},
			{Name: "User4", Account: "user4", Age: 35},
			{Name: "User5", Account: "user5", Age: 40},
			{Name: "User6", Account: "user6", Age: 45},
		})
		assert.Nil(t, err)
		_, err = conn.Model(&BelongsToManyRoles{}).Insert([]*BelongsToManyRoles{
			{Name: "customer"},
			{Name: "vip"},
			{Name: "suspended"},
			{Name: "expired"},
			{Name: "empty"},
		})
		assert.Nil(t, err)

		_, err = conn.Model(&goeloquent.EloquentModel{}).Table("role_user").Insert([]map[string]interface{}{
			{"role_id": 1, "user_id": 2, "valid": 1},
			{"role_id": 1, "user_id": 1, "valid": 1},
			{"role_id": 2, "user_id": 3, "valid": 1},
			{"role_id": 2, "user_id": 4, "valid": 1},
			{"role_id": 3, "user_id": 5, "valid": 0},
			{"role_id": 4, "user_id": 5, "valid": 1},
			{"role_id": 5, "user_id": 7, "valid": 1},
		})

		assert.Nil(t, err)
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var roles []BelongsToManyRoles
		_, err = conn.Model(&BelongsToManyRoles{}).With("Users").Get(&roles)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(sts))
		assert.Equal(t, "select * from `roles`", sts[0].RawSql)
		assert.Equal(t, "select `users`.*, `role_user`.`role_id` as `goelo_pivot_role_id`, `role_user`.`user_id` as `goelo_pivot_user_id` from `users` inner join `role_user` on `role_user`.`user_id` = `users`.`id` where `role_user`.`role_id` in (?, ?, ?, ?, ?)", sts[1].RawSql)
		assert.Equal(t, []interface{}{int64(1), int64(2), int64(3), int64(4), int64(5)}, sts[1].GetBindings())
		assert.Equal(t, 5, len(roles))
		for _, role := range roles {
			if role.Id == 5 {
				assert.Empty(t, role.Users)
			} else {
				assert.NotEmpty(t, role.Users)
				for _, user := range role.Users {
					assert.Equal(t, int32(role.Id), user.Pivot["role_id"])
				}
			}
		}

		var role BelongsToManyRoles
		sts = []*goeloquent.Statement{}
		_, err = conn.Model(&BelongsToManyRoles{}).Where("id", 2).First(&role)
		assert.Nil(t, err)
		var users []*BelongsToManyUser
		role.UsersRelation().Where("age", 35).Get(&users)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(sts))
		assert.Equal(t, "select * from `users` inner join `role_user` on `role_user`.`user_id` = `users`.`id` where `role_user`.`role_id` = ? and `age` = ?", sts[1].RawSql)
		assert.Equal(t, []interface{}{int64(2), 35}, sts[1].GetBindings())
		assert.Equal(t, 1, len(users))
		assert.Equal(t, 35, users[0].Age)

	})
}
func TestRelationLoadsResultsBelongsToMany(t *testing.T) {
}
func TestSelfRelationCountBelongsToMany(t *testing.T) {
	after := "drop table if exists users; drop table if exists roles; drop table if exists role_user;"
	before := after + `create table users (id integer primary key auto_increment,name varchar(255),account varchar(255),age integer);` +
		`create table roles (id integer primary key auto_increment,name varchar(255));` +
		`create table role_user (role_id integer not null,user_id integer not null,valid integer not null);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		_, err := conn.Model(&BelongsToManyUser{}).Insert([]*BelongsToManyUser{
			{Name: "User1", Account: "user1", Age: 20},
			{Name: "User2", Account: "user2", Age: 25},
			{Name: "User3", Account: "user3", Age: 30},
			{Name: "User4", Account: "user4", Age: 35},
			{Name: "User5", Account: "user5", Age: 40},
			{Name: "User6", Account: "user6", Age: 45},
		})
		assert.Nil(t, err)
		_, err = conn.Model(&BelongsToManyRoles{}).Insert([]*BelongsToManyRoles{
			{Name: "customer"},
			{Name: "vip"},
			{Name: "suspended"},
			{Name: "expired"},
			{Name: "empty"},
		})
		assert.Nil(t, err)

		_, err = conn.Model(&goeloquent.EloquentModel{}).Table("role_user").Insert([]map[string]interface{}{
			{"role_id": 1, "user_id": 2, "valid": 1},
			{"role_id": 1, "user_id": 1, "valid": 1},
			{"role_id": 2, "user_id": 3, "valid": 1},
			{"role_id": 2, "user_id": 4, "valid": 1},
			{"role_id": 3, "user_id": 5, "valid": 0},
			{"role_id": 4, "user_id": 5, "valid": 1},
			{"role_id": 5, "user_id": 7, "valid": 1},
		})

		assert.Nil(t, err)
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var roles []BelongsToManyRoles
		_, err = conn.Model(&BelongsToManyRoles{}).With("Users").Has("Users", ">=", 2).Get(&roles)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(sts))
		assert.Equal(t, "select * from `roles` where (select count(*) from `users` inner join `role_user` on `role_user`.`user_id` = `users`.`id` where `roles`.`id` = `role_user`.`role_id`) >= 2", sts[0].RawSql)
		assert.Empty(t, sts[0].GetBindings())
		assert.Equal(t, 2, len(roles))
		for _, role := range roles {
			assert.NotEmpty(t, role.Users)
			for _, user := range role.Users {
				assert.Equal(t, int32(role.Id), user.Pivot["role_id"])
			}
		}
		var roles1 []BelongsToManyRoles
		sts = []*goeloquent.Statement{}
		_, err = conn.Model(&BelongsToManyRoles{}).With("Users").WhereHas("Users", func(q *goeloquent.Statement) {
			q.Where("age", ">", 30)
		}).Get(&roles1)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(sts))
		assert.Equal(t, "select * from `roles` where exists (select * from `users` inner join `role_user` on `role_user`.`user_id` = `users`.`id` where `roles`.`id` = `role_user`.`role_id` and `age` > ?)", sts[0].RawSql)
		assert.Equal(t, []interface{}{30}, sts[0].GetBindings())
		assert.Equal(t, 3, len(roles1))
		for _, role := range roles1 {
			assert.NotEmpty(t, role.Users)
			for _, user := range role.Users {
				assert.Equal(t, int32(role.Id), user.Pivot["role_id"])
				assert.GreaterOrEqual(t, user.Age, 30)
			}
		}

	})
}
func TestSelectPivot(t *testing.T) {

}
func TestWherePivot(t *testing.T) {

}
