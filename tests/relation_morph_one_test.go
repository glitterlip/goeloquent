package tests

import (
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMorphOne(t *testing.T) {

	//select * from `images` where `imageable_type` = ? and `imageable_id` is not null and `imageable_id` in (?,?,?,?,?) and `driver` = ?
	//[user 1 2 3 4 5 s3]
	var us []User
	_, e := DB.Model(&us).With(map[string]func(b *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"Avatar": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return q.Where("driver", "s3")
		},
	}).Get(&us)
	assert.Nil(t, e)
	var c int
	for _, u := range us {
		if u.Avatar.ID > 0 {
			c++
			assert.True(t, u.IsBooted)
			assert.True(t, u.Avatar.IsBooted)
			assert.True(t, u.Avatar.ImageableId == u.ID)
			assert.True(t, u.Avatar.ImageableType == "user")
			assert.True(t, u.Avatar.Driver == "s3")
		}

	}
	assert.True(t, c > 0)

	//select * from `images` where `imageable_type` = ? and `imageable_id` is not null and `imageable_id` in (?) and `driver` = ?
	//[user 1 s3]
	var u User
	_, e = DB.Model(&u).With(map[string]func(b *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"Avatar": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return q.Where("driver", "s3")
		},
	}).Find(&u, 1)
	assert.Nil(t, e)
	assert.True(t, u.IsBooted)
	assert.True(t, u.Avatar.IsBooted)
	assert.True(t, u.Avatar.ImageableId == u.ID)
	assert.True(t, u.Avatar.ImageableType == "user")
	assert.True(t, u.Avatar.Driver == "s3")

	var u3 User
	_, e = DB.Model(&u3).Find(&u3, 3)
	assert.Nil(t, e)
	assert.True(t, u3.IsBooted)
	assert.True(t, u3.Avatar.ID == 0)
	var a Image
	_, e = u3.AvatarRelation().Where("driver", "s3").Get(&a)
	assert.Nil(t, e)
	assert.True(t, a.IsBooted)
	assert.True(t, a.ImageableId == u3.ID)
	assert.True(t, a.ImageableType == "user")
	assert.True(t, a.Driver == "s3")

	var u4 User
	r, e := DB.Model(&u4).WithCount("Avatar").Find(&u4, 3)
	assert.True(t, u4.WithAggregates["AvatarCount"] == 2)
	assert.Equal(t, r.Sql, "select *, (select Count(*) from `images` where `user_models`.`id` = `images`.`imageable_id` and `images`.`imageable_type` = ?) as `goelo_orm_aggregate_AvatarCount` from `user_models` where `id` = ? and `user_models`.`deleted_at` is null")

}
