package tests

import (
	"fmt"
	_ "fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/assert"
	_ "strconv"
	"strings"
	"testing"
	_ "time"
)

func TestBelongsToMany(t *testing.T) {

	var rs []Role
	//select `user_models`.*, `role_users`.`role_id` as `goelo_orm_pivot_role_id`, `role_users`.`user_id` as `goelo_orm_pivot_user_id`, `role_users`.`meta` as `goelo_pivot_meta`, `role_users`.`status` as `goelo_pivot_status`, `role_users`.`role_id` as `goelo_pivot_role_id`, `role_users`.`user_id` as `goelo_pivot_user_id` from `user_models` inner join `role_users` on `role_users`.`user_id` = `user_models`.`id` where `role_users`.`role_id` in (?,?,?) and (`user_models`.`status` = ? or `user_models`.`email` is not null) and `user_models`.`deleted_at` is null and `role_users`.`status` = ?
	//[1 2 3 1 1]
	_, e := DB.Model(&rs).With(map[string]func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"Users": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return q.Where(func(b *goeloquent.Builder) *goeloquent.Builder {
				return b.Where("user_models.status", 1).OrWhereNotNull("user_models.email")
			}).WherePivot("role_users.status", 1).WithPivot("role_users.meta", "role_users.status", "role_users.role_id", "role_users.user_id")
		},
	}).Get(&rs)
	assert.Nil(t, e)
	for _, r := range rs {
		for _, u := range r.Users {
			assert.True(t, u.IsBooted)
			assert.True(t, u.Pivot["meta"] != "")
			assert.True(t, u.Pivot["status"].(int64) == 1)
			assert.True(t, u.ID > 0)
			assert.True(t, u.Status == 1 || u.Email != "")
			assert.True(t, u.Pivot["role_id"].(int64) == r.ID)
			assert.True(t, u.Pivot["user_id"].(int64) == u.ID)
		}
	}

	//select `user_models`.*, `role_users`.`role_id` as `goelo_orm_pivot_role_id`, `role_users`.`user_id` as `goelo_orm_pivot_user_id`, `role_users`.`meta` as `goelo_pivot_meta`, `role_users`.`status` as `goelo_pivot_status`, `role_users`.`role_id` as `goelo_pivot_role_id`, `role_users`.`user_id` as `goelo_pivot_user_id` from `user_models` inner join `role_users` on `role_users`.`user_id` = `user_models`.`id` where `role_users`.`role_id` in (?) and (`user_models`.`status` = ? or `user_models`.`email` is not null) and `user_models`.`deleted_at` is null and `role_users`.`status` = ?
	//[1 1 1]
	var r Role
	_, e = DB.Model(&r).With(map[string]func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"Users": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return q.Where(func(b *goeloquent.Builder) *goeloquent.Builder {
				return b.Where("user_models.status", 1).OrWhereNotNull("user_models.email")
			}).WherePivot("role_users.status", 1).WithPivot("role_users.meta", "role_users.status", "role_users.role_id", "role_users.user_id")
		},
	}).Find(&r, 1)
	assert.Nil(t, e)
	assert.True(t, len(r.Users) > 0)
	for _, u := range r.Users {
		assert.True(t, u.IsBooted)
		assert.True(t, u.Pivot["meta"] != "")
		assert.True(t, u.Pivot["status"].(int64) == 1)
		assert.True(t, u.ID > 0)
		assert.True(t, u.Status == 1 || u.Email != "")
		assert.True(t, u.Pivot["role_id"].(int64) == r.ID)
		assert.True(t, u.Pivot["user_id"].(int64) == u.ID)
	}

	//select `user_models`.*, `role_users`.`role_id` as `goelo_orm_pivot_role_id`, `role_users`.`user_id` as `goelo_orm_pivot_user_id`, `role_users`.`meta` as `goelo_pivot_meta`, `role_users`.`status` as `goelo_pivot_status`, `role_users`.`role_id` as `goelo_pivot_role_id`, `role_users`.`user_id` as `goelo_pivot_user_id` from `user_models` inner join `role_users` on `role_users`.`user_id` = `user_models`.`id` where `role_users`.`role_id` = ? and (`user_models`.`status` = ? or `user_models`.`email` is not null) and `user_models`.`deleted_at` is null and `role_users`.`status` = ?
	//[1 1 1]
	var r2 Role
	_, e = DB.Model(&r2).Find(&r2, 1)
	assert.Nil(t, e)
	assert.True(t, len(r2.Users) == 0)
	var us []User
	_, e = r2.UsersRelation().Where(func(b *goeloquent.Builder) *goeloquent.Builder {
		return b.Where("user_models.status", 1).OrWhereNotNull("user_models.email")
	}).WherePivot("role_users.status", 1).
		WithPivot("role_users.meta", "role_users.status", "role_users.role_id", "role_users.user_id").
		Get(&us)
	assert.Nil(t, e)
	assert.True(t, len(us) > 0)
	for _, u := range us {
		assert.True(t, u.IsBooted)
		assert.True(t, u.Pivot["meta"] != "")
		assert.True(t, u.Pivot["status"].(int64) == 1)
		assert.True(t, u.ID > 0)
		assert.True(t, u.Status == 1 || u.Email != "")
		assert.True(t, u.Pivot["role_id"].(int64) == r2.ID)
		assert.True(t, u.Pivot["user_id"].(int64) == u.ID)
	}

	var rs1 []Role

	res, e := DB.Model(&Role{}).WithCount(map[string]func(b *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"Users as s1": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return q.Where("user_models.status", 1)
		},
		"Users as s2": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return q.Where("user_models.status", 2)
		},
	}).Get(&rs1)
	assert.Nil(t, e)
	assert.True(t, res.Count > 0)
	assert.True(t, strings.Contains(res.Sql, "(select Count(*) from `user_models` inner join `role_users` on `role_users`.`user_id` = `user_models`.`id` where `roles`.`id` = `role_users`.`role_id` and `user_models`.`status` = ?) as `goelo_orm_aggregate_s1`"))
	assert.True(t, strings.Contains(res.Sql, "(select Count(*) from `user_models` inner join `role_users` on `role_users`.`user_id` = `user_models`.`id` where `roles`.`id` = `role_users`.`role_id` and `user_models`.`status` = ?) as `goelo_orm_aggregate_s2`"))
	for _, rss := range rs1 {
		fmt.Println(rss.WithAggregates)
		assert.True(t, rss.WithAggregates["s1"] > 0)
		assert.True(t, rss.WithAggregates["s2"] > 0)
	}

}
