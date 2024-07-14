package tests

import (
	_ "fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMorphToMany(t *testing.T) {
	//select `tags`.*, `tagables`.`tag_id` as `goelo_orm_pivot_tag_id`, `tagables`.`tagable_id` as `goelo_orm_pivot_tagable_id`, `tagables`.`tag_url` as `goelo_pivot_tag_url`, `tagables`.`show_in_list` as `goelo_pivot_show_in_list` from `tags` inner join `tagables` on `tagables`.`tag_id` = `tags`.`id` where `tagable_type` = ? and `tagables`.`tag_id` in (?,?) and (`tags`.`related` > ? or `tags`.`id` < ?) and `show_in_list` = ?
	//[post 1 2 0 3 1]
	var ps []Post
	_, e := DB.Model(&Post{}).With(map[string]func(b *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"TagModels": func(b *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return b.Where(func(q *goeloquent.Builder) {
				q.Where("tags.related", ">", 0).OrWhere("tags.id", "<", 3)
			}).WherePivot("show_in_list", 1).WithPivot("tagables.tag_url", "tagables.show_in_list")
		},
	}).Where("id", "<", 3).Get(&ps)
	assert.Nil(t, e)
	for _, p := range ps {
		assert.True(t, len(p.TagModels) > 0)
		for _, tag := range p.TagModels {
			assert.True(t, tag.Pivot["show_in_list"].(int64) == 1)
			assert.True(t, len(string(tag.Pivot["tag_url"].([]byte))) > 1)
			assert.True(t, tag.Related > 0 || tag.ID < 3)
		}
	}

	//select `tags`.*, `tagables`.`tag_id` as `goelo_orm_pivot_tag_id`, `tagables`.`tagable_id` as `goelo_orm_pivot_tagable_id`, `tagables`.`tag_url` as `goelo_pivot_tag_url`, `tagables`.`show_in_list` as `goelo_pivot_show_in_list` from `tags` inner join `tagables` on `tagables`.`tag_id` = `tags`.`id` where `tagable_id` = ? and `tagable_type` = ? and (`related` > ? or `tags`.`id` < ?) and `show_in_list` = ?
	//[1 post 0 3 1]
	var p Post
	_, e = DB.Model(&Post{}).Where("id", 1).First(&p)
	assert.Nil(t, e)
	assert.Nil(t, p.TagModels)
	var ts []Tag
	_, e = p.TagsRelation().Where(func(q *goeloquent.Builder) {
		q.Where("related", ">", 0).OrWhere("tags.id", "<", 3)
	}).WherePivot("show_in_list", 1).
		WithPivot("tagables.tag_url", "tagables.show_in_list").
		Get(&ts)
	assert.Nil(t, e)
	assert.Greater(t, len(ts), 0)
	for _, tag := range ts {
		assert.True(t, tag.Pivot["show_in_list"].(int64) == 1)
		assert.True(t, len(string(tag.Pivot["tag_url"].([]byte))) > 1)
		assert.True(t, tag.Related > 0 || tag.ID < 3)
	}

}
