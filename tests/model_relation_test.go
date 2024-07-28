package tests

import (
	"fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAggregate(t *testing.T) {
	CreateUsers()
	var u User
	r, e := DB.Model(&u).WithMax("Posts", "status").Find(&u, 1)
	fmt.Println(r.Sql)
	assert.Nil(t, e)
	assert.Equal(t, u.WithAggregates["PostsMaxStatus"], float64(2))
	assert.Equal(t, r.Sql, "select *, (select Max(posts.status) from `posts` where `user_models`.`id` = `posts`.`user_id`) as `goelo_orm_aggregate_PostsMaxStatus` from `user_models` where `id` = ? and `user_models`.`deleted_at` is null")

	var role Role
	r, e = DB.Model(&role).WithAvg("Users", "status").First(&role)
	assert.Nil(t, e)
	assert.Equal(t, r.Sql, "select *, (select Avg(user_models.status) from `user_models` inner join `role_users` on `role_users`.`user_id` = `user_models`.`id` where `roles`.`id` = `role_users`.`role_id`) as `goelo_orm_aggregate_UsersAvgStatus` from `roles` limit 1")
	assert.Equal(t, role.WithAggregates["UsersAvgStatus"], float64(1.3333))

	var role1 Role
	r, e = DB.Model(&role).WithSum("Users", "age").First(&role1)
	assert.Nil(t, e)
	assert.Equal(t, r.Sql, "select *, (select Sum(user_models.age) from `user_models` inner join `role_users` on `role_users`.`user_id` = `user_models`.`id` where `roles`.`id` = `role_users`.`role_id`) as `goelo_orm_aggregate_UsersSumAge` from `roles` limit 1")
	assert.Equal(t, role1.WithAggregates["UsersSumAge"], float64(36))

	var u1 User
	r, e = DB.Model(&u1).WithExists("Posts", func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
		return builder.Where("status", 3)
	}).Find(&u1, 1)
	assert.Nil(t, e)
	assert.Equal(t, r.Sql, "select *, Exists(select * from `posts` where `user_models`.`id` = `posts`.`user_id` and `status` = ?) as goelo_orm_aggregate_PostsExists from `user_models` where `id` = ? and `user_models`.`deleted_at` is null")
	assert.Equal(t, u1.WithAggregates["PostsExists"], float64(0))

	var u2 User
	r, e = DB.Model(&u1).WithExists("Posts", func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
		return builder.Where("status", 1)
	}).Find(&u2, 1)
	assert.Nil(t, e)
	assert.Equal(t, u2.WithAggregates["PostsExists"], float64(1))
}
