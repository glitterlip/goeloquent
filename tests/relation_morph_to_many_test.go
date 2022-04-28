package tests

import (
	_ "fmt"
	goeloquent "github.com/glitterlip/go-eloquent"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestMorphToMany(t *testing.T) {
	c, d := CreateRelationTables()
	RunWithDB(c, d, func() {
		goeloquent.RegistMorphMap(map[string]interface{}{
			"posts": &Post{},
			"users": &UserT{},
		})
		//test saving,saved
		var u1 UserT
		u1.UserName = "u1"
		DB.Create(&u1)
		var t1, t2, t3, t4 Tag
		var p1, p2, p3 Post
		var pp1, pp2 Post
		var ps []Post
		p1.Title = "Intro"
		p2.Title = "Intro2"
		p3.Title = "Intro3"
		p1.AuthorId = u1.Id
		p2.AuthorId = u1.Id
		p3.AuthorId = u1.Id
		DB.Save(&p1)
		DB.Save(&p2)
		DB.Save(&p3)

		t1.Name = "golang"
		DB.Save(&t1)
		t2.Name = "js"
		DB.Save(&t2)
		t3.Name = "java"
		DB.Save(&t3)
		t4.Name = "laravel"
		DB.Save(&t4)

		var ts []map[string]interface{}
		var tps = []Post{p1, p2}
		for i := 0; i < 2; i++ {
			for j := 0; j < 4; j++ {
				tagMap := make(map[string]interface{})
				tagMap["tag_id"] = j + 1
				tagMap["tagable_id"] = tps[i].ID
				tagMap["tagable_type"] = "posts"
				tagMap["status"] = j % 2
				ts = append(ts, tagMap)
			}
		}

		_, err := DB.Table("tagable").Insert(ts)
		assert.Nil(t, err)
		//test find
		DB.Model(&p1).With("Tags").WithPivot("tagable_id", "tagable_type", "tag_id", "status").WherePivot("status", 1).Find(&pp1, p1.ID)
		assert.Greater(t, len(pp1.Tags), 0)
		for _, tag := range pp1.Tags {
			assert.Equal(t, tag.Pivot["status"].String, "1")
			assert.Equal(t, tag.Pivot["tagable_id"].String, strconv.Itoa(int(pp1.ID)))
			assert.Equal(t, tag.Pivot["tag_id"].String, strconv.Itoa(int(tag.ID)))
		}
		//test get
		result, err := DB.Model(&Post{}).With("Tags").WithPivot("status").WhereIn("pid", []int64{p1.ID, p2.ID}).Get(&ps)
		assert.Nil(t, err)
		count, _ := result.RowsAffected()
		assert.Equal(t, int64(2), count)
		for _, p := range ps {
			for _, tag := range p.Tags {
				assert.Contains(t, []string{"0", "1"}, tag.Pivot["status"].String)
				assert.Equal(t, tag.Pivot["tagable_id"].String, strconv.Itoa(int(p.ID)))
				assert.Equal(t, tag.Pivot["tag_id"].String, strconv.Itoa(int(tag.ID)))
			}
		}
		//test not find
		c, err := DB.Model(&p1).With("Tags").Find(&pp2, p3.ID)
		count, _ = c.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, 0, len(pp2.Tags))
		//test lazyload
		var lazy Post
		var lazyTags []Tag
		DB.Model(&p1).Find(&lazy, p1.ID)
		lazy.TagsRelation().WithPivot("status").WherePivot("status", 1).Get(&lazyTags)
		assert.Greater(t, len(lazyTags), 0)
		for _, tag := range lazyTags {
			assert.Equal(t, tag.Pivot["status"].String, "1")
			assert.Equal(t, tag.Pivot["tagable_id"].String, strconv.Itoa(int(lazy.ID)))
			assert.Equal(t, tag.Pivot["tag_id"].String, strconv.Itoa(int(tag.ID)))
		}
	})
	//TODO: test create update

}
