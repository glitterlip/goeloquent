package tests

import (
	_ "fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestMorphedByMany(t *testing.T) {
	c, d := CreateRelationTables()
	RunWithDB(c, d, func() {
		goeloquent.RegistMorphMap(map[string]interface{}{
			"post":   &Post{},
			"users":  &User{},
			"images": &Image{},
			"tag":    &Tag{},
		})
		//test saving,saved
		var u1, u2, u3 User
		var p1, p2 Post
		var ts []Tag
		var t1, t2, t3 Tag
		var tt1, tt3 Tag
		var i1, i2 Image
		u1.UserName = "u1"
		DB.Create(&u1)
		u1.ImagesRelation().Get(&i1)
		p1.Title = "Intro"
		p2.Title = "Intro2"
		p1.AuthorId = u2.Id
		p2.AuthorId = u3.Id
		DB.Save(&p1)
		DB.Save(&p2)
		i2.ImageableId = p1.ID
		i2.ImageableType = "post"
		DB.Save(&i2)

		t1.Name = "php"
		t2.Name = "java"
		t3.Name = "golang"
		DB.Save(&t1)
		DB.Save(&t2)
		DB.Save(&t3)

		DB.Table("tagables").Insert([]map[string]interface{}{
			{
				"tag_id":       t1.ID,
				"tagable_id":   p1.ID,
				"tagable_type": "post",
				"status":       0,
			},
			{
				"tag_id":       t1.ID,
				"tagable_id":   p2.ID,
				"tagable_type": "post",
				"status":       1,
			},
			{
				"tag_id":       t2.ID,
				"tagable_id":   p1.ID,
				"tagable_type": "post",
				"status":       0,
			},
			{
				"tag_id":       t2.ID,
				"tagable_id":   p2.ID,
				"tagable_type": "post",
				"status":       1,
			},
			{
				"tag_id":       t1.ID,
				"tagable_id":   i1.ID,
				"tagable_type": "images",
				"status":       0,
			},
			{
				"tag_id":       t1.ID,
				"tagable_id":   i2.ID,
				"tagable_type": "images",
				"status":       1,
			},

			{
				"tag_id":       t2.ID,
				"tagable_id":   i1.ID,
				"tagable_type": "images",
				"status":       0,
			},
			{
				"tag_id":       t2.ID,
				"tagable_id":   i2.ID,
				"tagable_type": "images",
				"status":       1,
			},
		})
		//test find
		DB.Model(&t1).With("Posts").Find(&tt1, t1.ID)
		for _, post := range tt1.Posts {
			assert.Equal(t, post.Pivot["goelo_orm_pivot_tagable_id"].(string), strconv.Itoa(int(post.ID)))
			assert.Equal(t, post.Pivot["goelo_orm_pivot_tagable_type"].(string), "post")
		}
		//test get
		result, err := DB.Model(&Tag{}).With("Images").WhereIn("tid", []int64{t1.ID, t2.ID}).WherePivot("status", 1).WithPivot("status", "tagable_id").Get(&ts)
		assert.Nil(t, err)
		count, _ := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), count)
		assert.Equal(t, 2, len(ts))
		for _, tag := range ts {
			assert.True(t, len(tag.Images) > 0)
			for _, image := range tag.Images {
				assert.Equal(t, image.Pivot["tagable_id"].(int64), image.ID)
			}
		}
		//test not find
		c, err := DB.Model(&t1).With("Images").Find(&tt3, t3.ID)
		count, _ = c.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, 0, len(tt3.Images))
		//test lazyload
		var lazy Tag
		var lazyImages []Image
		DB.Model(&t1).Find(&lazy, t2.ID)
		lazy.ImagesRelation().Get(&lazyImages)
		assert.Equal(t, 2, len(lazyImages))
		for _, image := range lazyImages {
			assert.Equal(t, image.Pivot["goelo_orm_pivot_tag_id"].(string), strconv.Itoa(int(t2.ID)))
			assert.Equal(t, image.Pivot["goelo_orm_pivot_tagable_id"].(string), strconv.Itoa(int(image.ID)))
			assert.Equal(t, image.Pivot["goelo_orm_pivot_tagable_type"].(string), "images")
		}
	})
	//TODO: test create update
}
