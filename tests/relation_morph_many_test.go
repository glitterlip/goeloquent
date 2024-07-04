package tests

import (
	"fmt"
	_ "fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMorphMany(t *testing.T) {
	c, d := CreateRelationTables()
	RunWithDB(c, d, func() {
		goeloquent.RegistMorphMap(map[string]interface{}{
			"posts": &Post{},
			"users": &User{},
		})
		//test saving,saved
		var u1, u2 User
		u1.UserName = "u1"
		u2.UserName = "u2"
		DB.Create(&u1)
		DB.Create(&u2)
		var p1, p2, p3, p4 Post
		var pp1, pp3 Post
		var ps []Post
		p1.Title = "Intro"
		p2.Title = "Intro2"
		p3.Title = "Intro3"
		p4.Title = "Intro4"
		p1.AuthorId = u2.Id
		p2.AuthorId = u1.Id
		p3.AuthorId = u1.Id
		p4.AuthorId = u1.Id
		DB.Save(&p1)
		DB.Save(&p2)
		DB.Save(&p3)
		DB.Save(&p4)
		var i1, i2 Image
		i1.Url = fmt.Sprintf("cdn.com/statis/posts/%d/1.png", p1.ID)
		i1.ImageableType = "posts"
		i1.ImageableId = p1.ID
		i2.Url = fmt.Sprintf("cdn.com/statis/posts/%d/1.png", p1.ID)
		i2.ImageableType = "posts"
		i2.ImageableId = p1.ID
		DB.Save(&i1)
		DB.Save(&i2)
		//test find
		b := DB.Model(&p1).With("Images")
		b.Find(&pp1, p1.ID)
		assert.Greater(t, len(pp1.Images), 0)

		for _, image := range pp1.Images {
			assert.Equal(t, image.ImageableId, p1.ID)
			assert.Equal(t, "posts", image.ImageableType)
		}
		//test get
		result, err := DB.Model(&Post{}).With("Images").WhereIn("pid", []int64{p1.ID, p2.ID}).Get(&ps)
		assert.Nil(t, err)
		count, _ := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), count)
		assert.Equal(t, 2, len(ps))
		var imageCount int
		for _, p := range ps {
			imageCount += len(p.Images)
			for _, image := range p.Images {
				assert.Equal(t, image.ImageableId, p.ID)
				assert.Equal(t, "posts", image.ImageableType)
			}
		}
		assert.Greater(t, imageCount, 0)
		//test not find
		c, err := DB.Model(&p1).With("Images").Find(&pp3, p3.ID)
		count, _ = c.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, 0, len(pp3.Images))
		//test lazyload
		var lazy Post
		var lazyImages []Image
		DB.Model(&p1).Find(&lazy, p1.ID)
		lazy.ImagesRelation().Get(&lazyImages)
		for _, image := range lazyImages {
			assert.Equal(t, image.ImageableId, lazy.ID)
			assert.Equal(t, "posts", image.ImageableType)
		}
	})
	//TODO: test create update
}
