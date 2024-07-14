package tests

import (
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMorphMany(t *testing.T) {
	goeloquent.RegistMorphMap(map[string]interface{}{
		"image": &Image{},
		"post":  &Post{},
		"video": &Video{},
		"user":  &User{},
	})
	//select * from `images` where `imageable_type` = ? and `imageable_id` is not null and `imageable_id` in (?,?,?) and (`driver` = ? or `size` >= ?)
	//[user 1 2 3 gcp 5120]
	var us []User
	_, e := DB.Model(&User{}).With(map[string]func(b *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"Images": func(b *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return b.Where(func(q *goeloquent.Builder) {
				q.Where("driver", "s3").OrWhere("size", ">=", 5120)
			})
		},
	}).Where("id", "<", 4).Get(&us)
	assert.Nil(t, e)

	for _, u := range us {
		assert.True(t, len(u.Images) > 0)
		for _, i := range u.Images {
			assert.True(t, i.Driver == "s3" || i.Size >= 5120)
			assert.Equal(t, u.ID, i.ImageableId)
			assert.Equal(t, "user", i.ImageableType)
		}
	}

	//select * from `images` where `imageable_id` = ? and `imageable_type` = ? and `size` > ?
	//[3 user 1024]
	var is []Image
	var u User
	_, e = DB.Model(&User{}).Find(&u, 3)
	assert.Nil(t, e)
	_, e = u.ImageRelation().Where("size", ">", 1024).Get(&is)
	assert.Nil(t, e)
	assert.True(t, len(is) > 0)
	for _, i := range is {
		assert.True(t, i.Size >= 5120)
		assert.Equal(t, u.ID, i.ImageableId)
		assert.Equal(t, "user", i.ImageableType)
	}
}
