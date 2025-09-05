package tests

import (
	"testing"

	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
)

type Tag struct {
	*goeloquent.EloquentModel
	Id      int64       `json:"id" goelo:"column:id;primaryKey;autoIncrement"`
	Name    string      `json:"name" goelo:"column:name"`
	Valid   int8        `json:"valid" goelo:"column:valid"`
	Tagable interface{} `json:"tagable" goelo:"MorphedByMany:TagablesRelation"`
}

func (t *Tag) GetTableName(st *goeloquent.Statement) string {
	return "tags"
}
func (t *Tag) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}
func (t *Tag) TagablesRelation() *goeloquent.MorphedByManyRelation {
	return t.MorphedByMany(t, "tagables", "tag_id", "tagable_id", "tagable_type", "id", "id")
}

type Post struct {
	*goeloquent.EloquentModel
	Id      int64  `json:"id" goelo:"column:id;primaryKey;autoIncrement"`
	Title   string `json:"title" goelo:"column:title"`
	Content string `json:"content" goelo:"column:content"`
	Tags    []Tag  `json:"tags" goelo:"MorphToMany:TagsRelation"`
}

func (p *Post) GetTableName(st *goeloquent.Statement) string {
	return "posts"
}
func (p *Post) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}
func (p *Post) TagsRelation() *goeloquent.MorphToManyRelation {
	return p.MorphToMany(p, &Tag{}, "tagables", "tagable_id", "tagable_type", "tag_id", "id", "id", "posts")
}

type Video struct {
	*goeloquent.EloquentModel
	Id    int64  `json:"id" goelo:"column:id;primaryKey;autoIncrement"`
	Title string `json:"title" goelo:"column:title"`
	Url   string `json:"url" goelo:"column:url"`

	Tags []Tag `json:"tags" goelo:"MorphToMany:TagsRelation"`
}

func (v *Video) GetTableName(st *goeloquent.Statement) string {
	return "videos"
}
func (v *Video) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}
func (v *Video) TagsRelation() *goeloquent.MorphToManyRelation {
	return v.MorphToMany(v, &Tag{}, "tagables", "tagable_id", "tagable_type", "tag_id", "id", "id", "videos")
}

func TestMorphToManyWithDynamicDefaultUseParentModel(t *testing.T) {
}
func TestSaveMethodSetsForeignKeyOnModelMorphToMany(t *testing.T) {
}
func TestEagerConstraintsAreProperlyAddedMorphToMany(t *testing.T) {
	goeloquent.DB.SetMorphMaps(map[string]interface{}{
		"posts":  &Post{},
		"videos": &Video{},
	})
	after := "drop table if exists posts; drop table if exists videos; drop table if exists tags; drop table if exists tagables;"
	before := after + "create table posts (id integer primary key auto_increment, title varchar(255), content varchar(255));" +
		"create table videos (id integer primary key auto_increment, title varchar(255), url varchar(255));" +
		"create table tags (id integer primary key auto_increment, name varchar(255), valid integer);" +
		"create table tagables (tag_id integer, tagable_id integer, tagable_type varchar(255));"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "post1", "content": "post1 content"},
			{"title": "post2", "content": "post2 content"},
		})

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var posts []Post
		conn.Model(&Post{}).With("Tags").Get(&posts)
		assert.Equal(t, len(sts), 2)
		assert.Equal(t, sts[0].RawSql, "select * from `posts`")
		assert.Equal(t, sts[1].RawSql, "select `tags`.*, `tagables`.`tagable_id` as `goelo_pivot_tagable_id`, `tagables`.`tagable_type` as `goelo_pivot_tagable_type`, `tagables`.`tag_id` as `goelo_pivot_tag_id` from `tags` inner join `tagables` on `tagables`.`tag_id` = `tags`.`id` where `tagables`.`tagable_type` = ? and `tagables`.`tagable_id` in (?, ?)")
		assert.Equal(t, sts[1].GetBindings(), []interface{}{"posts", int64(1), int64(2)})
	})
}
func TestModelsAreProperlyMatchedToParentsMorphToMany(t *testing.T) {
	goeloquent.DB.SetMorphMaps(map[string]interface{}{
		"posts":  &Post{},
		"videos": &Video{},
	})
	after := "drop table if exists posts; drop table if exists videos; drop table if exists tags; drop table if exists tagables;"
	before := after + "create table posts (id integer primary key auto_increment, title varchar(255), content varchar(255));" +
		"create table videos (id integer primary key auto_increment, title varchar(255), url varchar(255));" +
		"create table tags (id integer primary key auto_increment, name varchar(255), valid integer);" +
		"create table tagables (tag_id integer, tagable_id integer not null , tagable_type varchar(255));"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "post1", "content": "post1 content"},
			{"title": "post2", "content": "post2 content"},
			{"title": "post3", "content": "post3 content"},
		})

		conn.Table("tagables").Insert([]map[string]interface{}{
			{"tag_id": 1, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 1, "tagable_id": 3, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 3, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 3, "tagable_type": "posts"},
			{"tag_id": 1, "tagable_id": 2, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 2, "tagable_type": "posts"},
		})
		conn.Table("tags").Insert([]map[string]interface{}{
			{"name": "tag1", "valid": 1},
			{"name": "tag2", "valid": 1},
			{"name": "tag3", "valid": 1},
		})

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var posts []Post
		conn.Model(&Post{}).With("Tags").Get(&posts)
		assert.Equal(t, len(sts), 2)
		assert.Equal(t, sts[0].RawSql, "select * from `posts`")
		assert.Equal(t, sts[1].RawSql, "select `tags`.*, `tagables`.`tagable_id` as `goelo_pivot_tagable_id`, `tagables`.`tagable_type` as `goelo_pivot_tagable_type`, `tagables`.`tag_id` as `goelo_pivot_tag_id` from `tags` inner join `tagables` on `tagables`.`tag_id` = `tags`.`id` where `tagables`.`tagable_type` = ? and `tagables`.`tagable_id` in (?, ?, ?)")
		assert.Equal(t, sts[1].GetBindings(), []interface{}{"posts", int64(1), int64(2), int64(3)})
		for _, post := range posts {
			assert.Equal(t, int64(len(post.Tags)), post.Id)
			for _, tag := range post.Tags {
				post.Id = int64(tag.Pivot["tagable_id"].(int32))
				tag.Pivot["tag_id"] = tag.Id
			}

		}
	})
}
func TestRelationCountQueryCanBeBuiltMorphToMany(t *testing.T) {
	goeloquent.DB.SetMorphMaps(map[string]interface{}{
		"posts":  &Post{},
		"videos": &Video{},
	})
	after := "drop table if exists posts; drop table if exists videos; drop table if exists tags; drop table if exists tagables;"
	before := after + "create table posts (id integer primary key auto_increment, title varchar(255), content varchar(255));" +
		"create table videos (id integer primary key auto_increment, title varchar(255), url varchar(255));" +
		"create table tags (id integer primary key auto_increment, name varchar(255), valid integer);" +
		"create table tagables (tag_id integer, tagable_id integer not null , tagable_type varchar(255));"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "post1", "content": "post1 content"},
			{"title": "post2", "content": "post2 content"},
			{"title": "post3", "content": "post3 content"},
		})

		conn.Table("tagables").Insert([]map[string]interface{}{
			{"tag_id": 1, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 1, "tagable_id": 3, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 3, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 3, "tagable_type": "posts"},
			{"tag_id": 1, "tagable_id": 2, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 2, "tagable_type": "posts"},
		})
		conn.Table("tags").Insert([]map[string]interface{}{
			{"name": "tag1", "valid": 1},
			{"name": "tag2", "valid": 1},
			{"name": "tag3", "valid": 1},
		})

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var posts []Post
		conn.Model(&Post{}).Has("Tags", ">", 2).Get(&posts)
		assert.Equal(t, len(sts), 1)
		assert.Equal(t, sts[0].RawSql, "select * from `posts` where (select count(*) from `tags` inner join `tagables` on `tagables`.`tag_id` = `tags`.`id` where `posts`.`id` = `tagables`.`tagable_id` and `tagables`.`tagable_type` = ?) > 2")
		assert.Equal(t, len(posts), 1)
		assert.Equal(t, posts[0].Id, int64(3))

		sts = []*goeloquent.Statement{}
		var posts1 []Post
		conn.Model(&Post{}).WhereHas("Tags", func(st *goeloquent.Statement) {
			st.Where("name", "tag2")
		}).Get(&posts1)

		assert.Equal(t, len(sts), 1)
		assert.Equal(t, sts[0].RawSql, "select * from `posts` where exists (select * from `tags` inner join `tagables` on `tagables`.`tag_id` = `tags`.`id` where `posts`.`id` = `tagables`.`tagable_id` and `name` = ? and `tagables`.`tagable_type` = ?)")
		assert.Equal(t, sts[0].GetBindings(), []interface{}{"tag2", "posts"})
		assert.Equal(t, len(posts1), 2)
		assert.ElementsMatch(t, []interface{}{posts1[0].Id, posts1[1].Id}, []interface{}{int64(2), int64(3)})
	})
}
func TestCreateMethodProperlyCreatesNewModelMorphToMany(t *testing.T) {
}
func TestRelationGetResultsMorphToMany(t *testing.T) {
	goeloquent.DB.SetMorphMaps(map[string]interface{}{
		"posts":  &Post{},
		"videos": &Video{},
	})
	after := "drop table if exists posts; drop table if exists videos; drop table if exists tags; drop table if exists tagables;"
	before := after + "create table posts (id integer primary key auto_increment, title varchar(255), content varchar(255));" +
		"create table videos (id integer primary key auto_increment, title varchar(255), url varchar(255));" +
		"create table tags (id integer primary key auto_increment, name varchar(255), valid integer);" +
		"create table tagables (tag_id integer, tagable_id integer not null , tagable_type varchar(255));"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "post1", "content": "post1 content"},
			{"title": "post2", "content": "post2 content"},
			{"title": "post3", "content": "post3 content"},
		})

		conn.Table("tagables").Insert([]map[string]interface{}{
			{"tag_id": 1, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 1, "tagable_id": 3, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 3, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 3, "tagable_type": "posts"},
			{"tag_id": 1, "tagable_id": 2, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 2, "tagable_type": "posts"},
		})
		conn.Table("tags").Insert([]map[string]interface{}{
			{"name": "tag1", "valid": 1},
			{"name": "tag2", "valid": 1},
			{"name": "tag3", "valid": 1},
		})

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var post Post
		conn.Model(&Post{}).Find(&post, 3)
		var tags []Tag
		post.TagsRelation().Get(&tags)
		assert.Equal(t, len(sts), 2)
		assert.Equal(t, sts[0].RawSql, "select * from `posts` where `posts`.`id` = ? limit 1")
		assert.Equal(t, sts[1].RawSql, "select * from `tags` inner join `tagables` on `tagables`.`tag_id` = `tags`.`id` where `tagables`.`tagable_id` = ? and `tagables`.`tagable_type` = ?")
		assert.Equal(t, sts[1].GetBindings(), []interface{}{int64(3), "posts"})
		assert.Equal(t, len(tags), 3)
		for _, tag := range tags {
			assert.Equal(t, int64(tag.Pivot["tagable_id"].(int32)), int64(3))
			tag.Pivot["tag_id"] = tag.Id
		}

	})
}
func TestRelationLoadsResultsMorphToMany(t *testing.T) {
}
func TestSelfRelationCountMorphToMany(t *testing.T) {

}
