package tests

import (
	_ "fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/assert"
	"testing"
)

func TestHasMany(t *testing.T) {
	CreateUsers()
	//test get eager load
	//select * from `posts` where `posts`.`user_id` is not null and `posts`.`user_id` in (?,?,?,?,?)
	//[1 2 3 4 5]
	var us []User
	r, e := DB.Model(&us).With("Posts").Get(&us)
	assert.Nil(t, e)
	assert.Equal(t, r.Count, int64(5))
	for _, u := range us {
		for _, p := range u.Posts {
			assert.Equal(t, p.UserId, u.ID)
			assert.True(t, len(p.Tags.Strs) > 0)
			assert.True(t, p.IsBooted)
		}
	}

	//test get eager load with constraints  (espically with orwhere clause)
	//select * from `posts` where `posts`.`user_id` is not null and `posts`.`user_id` in (?,?,?,?,?) and (`status` = ? or `status` = ?)
	//[1 2 3 4 5 0 1]
	var us2 []User
	r, e = DB.Model(&us2).With(map[string]func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"Posts": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			q.Where(func(cq *goeloquent.Builder) *goeloquent.Builder {
				return cq.Where("status", 0).OrWhere("status", 1)
			})
			return q
		}}).Get(&us2)
	assert.Nil(t, e)
	assert.Equal(t, r.Count, int64(5))
	for _, u := range us2 {
		for _, p := range u.Posts {
			assert.Equal(t, p.UserId, u.ID)
			assert.True(t, len(p.Tags.Strs) > 0)
			assert.True(t, p.IsBooted)
			assert.True(t, p.Status == 0 || p.Status == 1)
		}
	}

	//test find eager load with constraints
	//select * from `posts` where `posts`.`user_id` is not null and `posts`.`user_id` in (?)
	//[1]
	var u User
	r, e = DB.Model(&u).With("Posts").Find(&u, 1)
	assert.Nil(t, e)
	assert.True(t, len(u.Posts) == 2)
	for _, p := range u.Posts {
		assert.Equal(t, p.UserId, u.ID)
		assert.True(t, len(p.Tags.Strs) > 0)
		assert.True(t, p.IsBooted)
	}

	//select * from `posts` where `posts`.`user_id` is not null and `posts`.`user_id` in (?) and (`status` = ? or `status` = ?)
	//[1 0 1]
	var u3 User
	r, e = DB.Model(&u3).With(map[string]func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"Posts": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			q.Where(func(cq *goeloquent.Builder) *goeloquent.Builder {
				return cq.Where("status", 0).OrWhere("status", 1)
			})
			return q
		}}).Find(&u3, 1)
	assert.Nil(t, e)
	assert.True(t, len(u3.Posts) == 1)
	for _, p := range u3.Posts {
		assert.Equal(t, p.UserId, u3.ID)
		assert.True(t, len(p.Tags.Strs) > 0)
		assert.True(t, p.IsBooted)
		assert.True(t, p.Status == 0 || p.Status == 1)
	}
	//test model load
	//select * from `posts` where `posts`.`user_id` = ? and `posts`.`user_id` is not null
	//[1]
	var u4 User
	r, e = DB.Model(&u4).Find(&u4, 1)
	assert.Nil(t, e)
	assert.Equal(t, u4.ID, int64(1))
	var posts []Post
	r, e = u4.PostRelation().Get(&posts)
	assert.Nil(t, e)
	assert.True(t, len(posts) == 2)
	for _, p := range posts {
		assert.Equal(t, p.UserId, u4.ID)
		assert.True(t, len(p.Tags.Strs) > 0)
		assert.True(t, p.IsBooted)

	}

	//test relation get
	//select * from `posts` where `posts`.`user_id` = ? and `posts`.`user_id` is not null and (`status` = ? or `status` = ?)
	//[1 0 1]
	var u5 User
	r, e = DB.Model(&u5).Find(&u5, 1)
	assert.Nil(t, e)
	assert.Equal(t, u5.ID, int64(1))

	var posts2 []Post
	r, e = u5.PostRelation().Where(func(cq *goeloquent.Builder) *goeloquent.Builder {
		return cq.Where("status", 0).OrWhere("status", 1)
	}).Get(&posts2)
	assert.Nil(t, e)
	assert.True(t, len(posts2) == 1)
	for _, p := range posts2 {
		assert.Equal(t, p.UserId, u5.ID)
		assert.True(t, len(p.Tags.Strs) > 0)
		assert.True(t, p.IsBooted)
		assert.True(t, p.Status == 0 || p.Status == 1)
	}

	var us1 []User
	r, e = DB.Model(&us).WithCount("Posts").Get(&us1)

	assert.Equal(t, r.Count, int64(5))
	assert.Nil(t, e)
	assert.Equal(t, r.Sql, "select *, (select Count(*) from `posts` where `user_models`.`id` = `posts`.`user_id`) as `goelo_orm_aggregate_PostsCount` from `user_models` where `user_models`.`deleted_at` is null")
	for _, user := range us1 {
		if user.ID != 3 {
			assert.True(t, user.WithAggregates["PostsCount"] > 0)
		} else {
			assert.True(t, user.WithAggregates["PostsCount"] == 0)
		}

	}

	var u2 User
	r, e = DB.Model(&u2).WithCount([]string{"PostCount", "PostCountWithTrashed"}).First(&u2)
	assert.Nil(t, e)
	assert.Equal(t, r.Sql, "select *, (select Count(*) from `posts` where `user_models`.`id` = `posts`.`user_id` and `status` > ?) as `goelo_orm_aggregate_PostCount`, (select Count(*) from `posts` where `user_models`.`id` = `posts`.`user_id` and `status` > ?) as `goelo_orm_aggregate_PostCountWithTrashed` from `user_models` where `user_models`.`deleted_at` is null limit 1")
	assert.Equal(t, u2.WithAggregates["PostCount"], float64(1))
	assert.Equal(t, u2.WithAggregates["PostCountWithTrashed"], float64(1))
	assert.Equal(t, u2.PostCount, float64(1))
	assert.Equal(t, u2.PostCountWithTrashed, float64(1))

}
