package tests

import (
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadingWithColumns(t *testing.T) {

}

func TestNestedRelation(t *testing.T) {
	goeloquent.RegistMorphMap(map[string]interface{}{
		"post": &Post{},
	})
	goeloquent.RegisterModels([]interface{}{&User{}, &Post{}})
	var us []User

	_, e := DB.Model(&us).With("Posts.TagModels").Find(&us, []int{4, 5})
	assert.Nil(t, e)
	assert.Equal(t, 2, len(us))
	for _, u := range us {
		assert.Greater(t, len(u.Posts), 0)
		for _, p := range u.Posts {
			assert.Equal(t, p.UserId, u.ID)
			assert.Greater(t, len(p.TagModels), 0)
		}
	}
}
func TestRelationPagination(t *testing.T) {
	goeloquent.RegistMorphMap(map[string]interface{}{
		"post": &Post{},
	})
	goeloquent.RegisterModels([]interface{}{&User{}, &Post{}})
	var us []User

	_, e := DB.Model(&us).With("Posts").Paginate(&us, 2, 1)
	assert.Nil(t, e)
	assert.Equal(t, 2, len(us))
	for _, u := range us {
		assert.Greater(t, len(u.Posts), 0)
	}
}
