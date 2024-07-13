package tests

import (
	_ "fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBelongsTo(t *testing.T) {
	CreateUsers()
	//test get eager load
	//select * from `user_models` where `user_models`.`id` is not null and `user_models`.`id` in (?,?,?,?,?,?,?,?) and `deleted_at` is null
	//[5 5 4 2 1 1 2 4]
	var ps []Post
	r, e := DB.Model(&ps).With("User").Get(&ps)
	assert.Nil(t, e)
	assert.Equal(t, r.Count, int64(8))
	for _, p := range ps {
		assert.True(t, p.User.ID == p.UserId)
		assert.True(t, p.User.IsBooted)
	}
	//test get eager load with constraints
	//select * from `user_models` where `user_models`.`id` is not null and `user_models`.`id` in (?,?,?,?,?,?,?,?) and (`status` = ? or `email` = ?) and `deleted_at` is null
	//[5 5 4 2 1 1 2 4 1 1]
	var ps2 []Post
	r, e = DB.Model(&ps2).With(map[string]func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"User": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			q.Where(func(cq *goeloquent.Builder) *goeloquent.Builder {
				return cq.Where("status", 1).OrWhere("email", 1)
			})
			return q
		}}).Get(&ps2)
	assert.Nil(t, e)
	assert.Equal(t, r.Count, int64(8))
	var uc int
	for _, p := range ps2 {
		if p.User.ID > 0 {
			uc++
			assert.True(t, p.User.ID == p.UserId)
			assert.True(t, p.User.IsBooted)
			assert.True(t, p.User.Status == 1 || p.User.Email == "123")
		}
	}
	assert.Equal(t, uc, 4)
	//test find eager load with constraints
	//select * from `user_models` where `user_models`.`id` is not null and `user_models`.`id` in (?) and (`status` = ? or `email` = ?) and `deleted_at` is null
	//[5 1 123]
	var ps3 Post
	r, e = DB.Model(&ps3).With(map[string]func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"User": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			q.Where(func(cq *goeloquent.Builder) *goeloquent.Builder {
				return cq.Where("status", 1).OrWhere("email", "123")
			})
			return q
		}}).Find(&ps3, 1)
	assert.Nil(t, e)
	assert.True(t, ps3.User.ID == ps3.UserId)
	assert.True(t, ps3.User.IsBooted)
	assert.True(t, ps3.User.Status == 1 || ps3.User.Email == "123")

	//test relation get

	var ps4 Post
	r, e = DB.Model(&ps4).Find(&ps4, 6)
	assert.Nil(t, e)
	assert.Equal(t, ps4.UserId, int64(1))
	var owner User
	//select * from `user_models` where `user_models`.`id` = ? and `user_models`.`id` is not null and `deleted_at` is null
	//[1]
	r, e = ps4.UserRelation().Get(&owner)
	assert.Nil(t, e)
	assert.NotNil(t, owner)
	assert.True(t, owner.ID == ps4.UserId)
	assert.True(t, owner.IsBooted)
	var ps5 Post
	r, e = DB.Model(&ps5).Find(&ps5, 5)
	assert.Nil(t, e)
	var owner2 User
	//select * from `user_models` where `user_models`.`id` = ? and `user_models`.`id` is not null and (`status` = ? or `email` = ?) and `deleted_at` is null
	//[1 1 qwe]
	r, e = ps5.UserRelation().Where(func(q *goeloquent.Builder) {
		q.Where("status", 1).OrWhere("email", "qwe")
	}).Get(&owner2)
	assert.Nil(t, e)
	assert.Equal(t, owner2.ID, ps5.UserId)
	assert.True(t, owner2.Status == 1 || owner2.Email == "qwe")

}
