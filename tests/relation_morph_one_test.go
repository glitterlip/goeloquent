package tests

import (
	"fmt"
	_ "fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMorphOne(t *testing.T) {
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
		i1.Url = fmt.Sprintf("cdn.com/statis/posts/%d/1.png", p2.ID)
		i1.ImageableType = "posts"
		i1.ImageableId = p2.ID
		i2.Url = fmt.Sprintf("cdn.com/statis/posts/%d/1.png", p1.ID)
		i2.ImageableType = "posts"
		i2.ImageableId = p1.ID
		DB.Save(&i1)
		DB.Save(&i2)
		//test find
		b := DB.Model(&p1).With("Image")
		b.Find(&pp1, p2.ID)
		assert.Equal(t, p2.ID, pp1.Image.ImageableId)
		assert.Equal(t, "posts", pp1.Image.ImageableType)
		//test get
		result, err := DB.Model(&Post{}).With("Image").WhereIn("pid", []int64{p1.ID, p2.ID}).Get(&ps)
		assert.Nil(t, err)
		count, _ := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), count)
		assert.Equal(t, 2, len(ps))
		for _, p := range ps {
			assert.Equal(t, p.ID, p.Image.ImageableId)
		}
		//test not find
		c, err := DB.Model(&p1).With("Image").Find(&pp3, p3.ID)
		count, _ = c.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, int64(0), pp3.Image.ImageableId)
		assert.Equal(t, "", pp3.Image.ImageableType)
		//test lazyload
		var lazy Post
		var lazyImage Image
		DB.Model(&p1).Find(&lazy, p2.ID)
		lazy.ImageRelation().Get(&lazyImage)
		assert.Equal(t, lazyImage.ImageableId, p2.ID)
	})
	//TODO: test create update
}
