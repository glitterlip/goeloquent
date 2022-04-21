package tests

import (
	_ "fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBelongsTo(t *testing.T) {
	c, d := CreateRelationTables()
	RunWithDB(c, d, func() {
		//test saving,saved
		var u1, u2, u3 UserT
		u1.UserName = "u1"
		u1.UserName = "u2"
		u1.UserName = "u3"
		DB.Create(&u1)
		DB.Create(&u2)
		DB.Create(&u3)
		var p1, p2, p3, p4 Post
		var pp1, pp4 Post
		var ps []Post
		p1.Title = "Intro"
		p2.Title = "Intro2"
		p3.Title = "Intro3"
		p4.Title = "Intro4"
		p1.AuthorId = u2.Id
		p2.AuthorId = u3.Id
		p3.AuthorId = u1.Id
		DB.Save(&p1)
		DB.Save(&p2)
		DB.Save(&p3)
		DB.Save(&p4)
		//test find
		DB.Model(&p1).With("Author").Find(&pp1, p1.ID)
		assert.Equal(t, pp1.AuthorId, u2.Id)
		assert.Equal(t, pp1.Author.Id, pp1.AuthorId)
		//test get
		result, err := DB.Model(&Post{}).With("Author").WhereIn("pid", []int64{p2.ID, p3.ID}).Get(&ps)
		assert.Nil(t, err)
		count, _ := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), count)
		assert.Equal(t, 2, len(ps))
		for _, p := range ps {
			assert.Equal(t, p.AuthorId, p.Author.Id)
		}
		//test not find
		c, err := DB.Model(&p1).With("Author").Find(&pp4, p4.ID)
		count, _ = c.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, int64(0), pp4.Author.Id)
		assert.Equal(t, int64(0), pp4.AuthorId)
		//test lazyload
		var lazy Post
		var lazyUser UserT
		DB.Model(&p1).Find(&lazy, p2.ID)
		lazy.AuthorRelation().Get(&lazyUser)
		assert.Equal(t, lazy.AuthorId, lazyUser.Id)
	})
	//TODO: test create update
}
