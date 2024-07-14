package tests

import (
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMorphTo(t *testing.T) {
	goeloquent.RegistMorphMap(map[string]interface{}{
		"image": &Image{},
		"post":  &Post{},
		"video": &Video{},
		"user":  &User{},
	})
	//select * from `user_models` where `id` in (?,?,?) and `user_models`.`deleted_at` is null
	//[1 2 3]
	//select * from `posts` where `id` in (?,?)
	//[3 3]
	var is []Image
	r, e := DB.Model(&Image{}).With("Imageable").Where("driver", "s3").Where("size", ">", 1024).Get(&is)
	assert.Equal(t, r.Count, int64(5))
	assert.Nil(t, e)
	for _, image := range is {
		assert.Equal(t, image.Driver, "s3")
		assert.True(t, image.Size > 1024)
		assert.NotNil(t, image.Imageable)
		if user, ok := image.Imageable.(User); ok {
			assert.Equal(t, user.ID, image.ImageableId)
			assert.Equal(t, image.ImageableType, "user")
			assert.True(t, user.IsBooted)
		} else if post, ok := image.Imageable.(Post); ok {
			assert.Equal(t, post.ID, image.ImageableId)
			assert.Equal(t, image.ImageableType, "post")
			assert.True(t, post.IsBooted)
		}
	}

}
