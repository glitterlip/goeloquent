package tests

import (
	"fmt"
	_ "fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMorphTo(t *testing.T) {
	c, d := CreateRelationTables()
	RunWithDB(c, d, func() {
		goeloquent.RegistMorphMap(map[string]interface{}{
			"posts": &Post{},
			"users": &UserT{},
		})
		//test saving,saved
		var u1, u2 UserT
		u1.UserName = "u1"
		u2.UserName = "u2"
		DB.Create(&u1)
		DB.Create(&u2)

		var i1, ii1, ii2 Image
		var images []Image
		i1.Url = fmt.Sprintf("cdn.com/statis/posts/%d/1.png", 1)
		i1.ImageableType = "users"
		i1.ImageableId = 10
		DB.Save(&i1)
		//test find
		b := DB.Model(&i1).With("Imageable")
		b.First(&ii1)
		assert.IsType(t, UserT{}, ii1.Imageable)
		//test get
		result, err := DB.Model(&i1).With("Imageable").Where("imageable_type", "users").Limit(2).Get(&images)
		assert.Nil(t, err)
		count, _ := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), count)
		for _, image := range images {
			switch imageable := image.Imageable.(type) {
			case UserT:
				imageable.Id = image.ImageableId
				image.ImageableType = "users"
			default:
				panic("type error")
			}
		}
		//test not find
		c, err := DB.Model(&ii2).With("Imageable").Find(&ii2, i1.ID)
		count, _ = c.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), count)
		assert.Nil(t, ii2.Imageable)
		//test lazyload
		var lazy Image
		var imageableLazy = make(map[string]interface{})
		DB.Model(&lazy).Find(&lazy, ii1.ID)
		lazy.ImageableRelation().Find(&imageableLazy, lazy.ID)
		assert.Equal(t, imageableLazy["id"], lazy.ImageableId)

	})
	//TODO: test create update
}
