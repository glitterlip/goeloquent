package tests

import (
	_ "fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHasMany(t *testing.T) {
	c, d := CreateRelationTables()
	RunWithDB(c, d, func() {
		//test saving,saved
		var u1, u2, u3 UserT
		var uu1, uu2 UserT
		var us []UserT
		u1.UserName = "u1"
		u2.UserName = "u2"
		u3.UserName = "u3"
		DB.Create(&u1)
		DB.Create(&u2)
		DB.Create(&u3)
		var p1, p2, p3 Post
		p1.Title = "Intro"
		p2.Title = "Intro2"
		p3.Title = "Intro3"
		p1.AuthorId = u2.Id
		p2.AuthorId = u1.Id
		p3.AuthorId = u2.Id
		DB.Save(&p1)
		DB.Save(&p2)
		DB.Save(&p3)

		//test find
		DB.Model(&u1).With("Posts").Find(&uu1, u1.Id)
		assert.Equal(t, 2, len(uu1.Posts))
		for _, post := range uu1.Posts {
			assert.Equal(t, uu1.Id, post.AuthorId)
		}

		//test get
		result, err := DB.Model(&UserT{}).With("Posts").Get(&us)
		assert.Nil(t, err)
		count, _ := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(3), count)
		for _, user := range us {
			if user.Id == u2.Id {
				assert.Equal(t, 3, len(user.Posts))
			}
			for _, post := range user.Posts {
				assert.Equal(t, post.AuthorId, user.Id)
			}
		}
		//test not find
		c, err := DB.Model(&p1).With("Posts").Find(&uu2, u3.Id)
		count, _ = c.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, 0, len(uu2.Posts))
		//test lazyload
		var lazy UserT
		var lazyPosts []Post
		DB.Model(&u1).Find(&lazy, u2.Id)
		lazy.PostsRelation().Get(&lazyPosts)
		for _, lazyPost := range lazyPosts {
			assert.Equal(t, lazyPost.AuthorId, lazy.Id)
		}
	})
	//TODO: test create update
}
