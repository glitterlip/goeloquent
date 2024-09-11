package tests

import (
	_ "fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestMorphedByMany(t *testing.T) {

	goeloquent.RegistMorphMap(map[string]interface{}{
		"image": &Image{},
		"post":  &Post{},
		"video": &Video{},
		"user":  &User{},
		"tag":   &Tag{},
	})
	var ts []Tag
	r, e := DB.Model(&ts).With(map[string]func(*goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"Posts": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return q.WherePivot("show_in_list", 1).WithPivot("tagables.show_in_list", "tagables.tag_url")
		},
	}).Get(&ts)

	assert.Nil(t, e)
	assert.Equal(t, r.Count, int64(8))

	for _, tag := range ts {
		for _, p := range tag.Posts {
			assert.True(t, strconv.Itoa(int(tag.ID)) == p.Pivot[goeloquent.OrmPivotAlias+"tag_id"].(string))
			assert.True(t, p.Pivot["show_in_list"].(int64) == 1)
			assert.True(t, p.Pivot[goeloquent.OrmPivotAlias+"tagable_id"].(string) == strconv.Itoa(int(p.ID)))
		}
	}

	var tag1 Tag
	r, e = DB.Model(&tag1).With(map[string]func(*goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"Posts": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return q.WherePivot("show_in_list", 1).WithPivot("tagables.show_in_list", "tagables.tag_url")
		},
	}).Find(&tag1, 1)

	assert.Nil(t, e)
	assert.True(t, len(tag1.Posts) == 1)

}
