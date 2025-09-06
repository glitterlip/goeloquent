package tests

import (
	"testing"

	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
)

func TestMorphedByManyWithDynamicDefaultUseParentModel(t *testing.T) {
}
func TestSaveMethodSetsForeignKeyOnModelMorphedByMany(t *testing.T) {
}
func TestEagerConstraintsAreProperlyAddedMorphedByMany(t *testing.T) {
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

		conn.Table("videos").Insert([]map[string]interface{}{
			{"title": "video1", "url": "http://example.com/video1"},
			{"title": "video2", "url": "http://example.com/video2"},
			{"title": "video3", "url": "http://example.com/video3"},
		})

		conn.Table("tagables").Insert([]map[string]interface{}{
			{"tag_id": 1, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 1, "tagable_id": 1, "tagable_type": "videos"},
			{"tag_id": 2, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 1, "tagable_type": "videos"},
			{"tag_id": 2, "tagable_id": 2, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 2, "tagable_type": "videos"},
			{"tag_id": 3, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 1, "tagable_type": "videos"},
			{"tag_id": 3, "tagable_id": 2, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 2, "tagable_type": "videos"},
			{"tag_id": 3, "tagable_id": 3, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 3, "tagable_type": "videos"},
			{"tag_id": 1, "tagable_id": 5, "tagable_type": "videos"},
			{"tag_id": 1, "tagable_id": 5, "tagable_type": "posts"},
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
		var tags []Tag
		conn.Model(&Tag{}).With([]string{"Videos", "Posts"}).Get(&tags)
		assert.Equal(t, 3, len(tags))
		assert.Equal(t, len(sts), 3)
		sqls := []string{"select * from `tags`",
			"select `videos`.*, `tagables`.`tag_id` as `goelo_pivot_tag_id`, `tagables`.`tagable_type` as `goelo_pivot_tagable_type`, `tagables`.`tagable_id` as `goelo_pivot_tagable_id` from `videos` inner join `tagables` on `tagables`.`tagable_id` = `videos`.`id` where `tagables`.`tagable_type` = 'videos' and `tagables`.`tag_id` in (1, 2, 3)",
			"select `posts`.*, `tagables`.`tag_id` as `goelo_pivot_tag_id`, `tagables`.`tagable_type` as `goelo_pivot_tagable_type`, `tagables`.`tagable_id` as `goelo_pivot_tagable_id` from `posts` inner join `tagables` on `tagables`.`tagable_id` = `posts`.`id` where `tagables`.`tagable_type` = 'posts' and `tagables`.`tag_id` in (1, 2, 3)",
		}
		assert.Contains(t, sqls, sts[0].ToRawSql())
		assert.Contains(t, sqls, sts[1].ToRawSql())
		assert.Contains(t, sqls, sts[2].ToRawSql())

	})
}
func TestModelsAreProperlyMatchedToParentsMorphedByMany(t *testing.T) {
	goeloquent.DB.SetMorphMaps(map[string]interface{}{
		"posts":  &Post{},
		"videos": &Video{},
	})
	after := "drop table if exists posts; drop table if exists videos; drop table if exists tags; drop table if exists tagables;"
	before := after + "create table posts (id integer primary key auto_increment, title varchar(255), content varchar(255));" +
		"create table videos (id integer primary key auto_increment, title varchar(255), url varchar(255));" +
		"create table tags (id integer primary key auto_increment, name varchar(255), valid integer);" +
		"create table tagables (id integer primary key auto_increment,tag_id integer not null, tagable_id integer not null , tagable_type varchar(255) not null );"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "post1", "content": "post1 content"},
			{"title": "post2", "content": "post2 content"},
			{"title": "post3", "content": "post3 content"},
		})

		conn.Table("videos").Insert([]map[string]interface{}{
			{"title": "video1", "url": "http://example.com/video1"},
			{"title": "video2", "url": "http://example.com/video2"},
			{"title": "video3", "url": "http://example.com/video3"},
		})

		conn.Table("tagables").Insert([]map[string]interface{}{
			{"tag_id": 1, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 1, "tagable_id": 1, "tagable_type": "videos"},
			{"tag_id": 2, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 1, "tagable_type": "videos"},
			{"tag_id": 2, "tagable_id": 2, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 2, "tagable_type": "videos"},
			{"tag_id": 3, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 1, "tagable_type": "videos"},
			{"tag_id": 3, "tagable_id": 2, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 2, "tagable_type": "videos"},
			{"tag_id": 3, "tagable_id": 3, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 3, "tagable_type": "videos"},
			{"tag_id": 1, "tagable_id": 5, "tagable_type": "videos"},
			{"tag_id": 1, "tagable_id": 5, "tagable_type": "posts"},
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
		var tags []Tag
		conn.Model(&Tag{}).With([]string{"Videos", "Posts"}).Get(&tags)
		assert.Equal(t, 3, len(tags))
		assert.Equal(t, len(sts), 3)
		for _, tag := range tags {
			assert.Equal(t, len(tag.Posts), int(tag.Id))
			assert.Equal(t, len(tag.Videos), int(tag.Id))
			for _, post := range tag.Posts {
				assert.Equal(t, post.Pivot["tag_id"], int32(tag.Id))
				assert.Equal(t, post.Pivot["tagable_type"], "posts")
				assert.Equal(t, post.Pivot["tagable_id"], int32(post.Id))
			}
			for _, video := range tag.Videos {
				assert.Equal(t, video.Pivot["tag_id"], int32(tag.Id))
				assert.Equal(t, video.Pivot["tagable_type"], "videos")
				assert.Equal(t, video.Pivot["tagable_id"], int32(video.Id))
			}
		}

	})
}
func TestRelationCountQueryCanBeBuiltMorphedByMany(t *testing.T) {
	goeloquent.DB.SetMorphMaps(map[string]interface{}{
		"posts":  &Post{},
		"videos": &Video{},
	})
	after := "drop table if exists posts; drop table if exists videos; drop table if exists tags; drop table if exists tagables;"
	before := after + "create table posts (id integer primary key auto_increment, title varchar(255), content varchar(255));" +
		"create table videos (id integer primary key auto_increment, title varchar(255), url varchar(255));" +
		"create table tags (id integer primary key auto_increment, name varchar(255), valid integer);" +
		"create table tagables (id integer primary key auto_increment,tag_id integer not null, tagable_id integer not null , tagable_type varchar(255) not null );"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "post1", "content": "post1 content"},
			{"title": "post2", "content": "post2 content"},
			{"title": "post3", "content": "post3 content"},
		})

		conn.Table("videos").Insert([]map[string]interface{}{
			{"title": "video1", "url": "http://example.com/video1"},
			{"title": "video2", "url": "http://example.com/video2"},
			{"title": "video3", "url": "http://example.com/video3"},
		})

		conn.Table("tagables").Insert([]map[string]interface{}{
			{"tag_id": 1, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 1, "tagable_type": "videos"},
			{"tag_id": 2, "tagable_id": 2, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 2, "tagable_type": "videos"},
			{"tag_id": 3, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 1, "tagable_type": "videos"},
			{"tag_id": 3, "tagable_id": 2, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 2, "tagable_type": "videos"},
			{"tag_id": 3, "tagable_id": 3, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 3, "tagable_type": "videos"},
			{"tag_id": 1, "tagable_id": 5, "tagable_type": "videos"},
			{"tag_id": 1, "tagable_id": 5, "tagable_type": "posts"},
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
		var tags []Tag
		conn.Model(&Tag{}).Has("Videos").Get(&tags)
		assert.Equal(t, len(tags), 2)
		assert.Equal(t, "select * from `tags` where exists (select * from `videos` inner join `tagables` on `tagables`.`tagable_id` = `videos`.`id` where `tags`.`id` = `tagables`.`tag_id` and `tagables`.`tagable_type` = 'videos')", sts[0].ToRawSql())

		tags = []Tag{}
		sts = []*goeloquent.Statement{}
		conn.Model(&Tag{}).Has("Posts", ">", 2).Get(&tags)
		assert.Equal(t, len(tags), 1)
		assert.Equal(t, "select * from `tags` where (select count(*) from `posts` inner join `tagables` on `tagables`.`tagable_id` = `posts`.`id` where `tags`.`id` = `tagables`.`tag_id` and `tagables`.`tagable_type` = 'posts') > 2", sts[0].ToRawSql())

		tags = []Tag{}
		sts = []*goeloquent.Statement{}
		conn.Model(&Tag{}).Has("Posts", ">=", 2).Get(&tags)
		assert.Equal(t, len(tags), 2)
		assert.Equal(t, "select * from `tags` where (select count(*) from `posts` inner join `tagables` on `tagables`.`tagable_id` = `posts`.`id` where `tags`.`id` = `tagables`.`tag_id` and `tagables`.`tagable_type` = 'posts') >= 2", sts[0].ToRawSql())

		tags = []Tag{}
		sts = []*goeloquent.Statement{}
		conn.Model(&Tag{}).Has("Posts", "<", 3).Get(&tags)
		assert.Equal(t, len(tags), 2)
		assert.Equal(t, "select * from `tags` where (select count(*) from `posts` inner join `tagables` on `tagables`.`tagable_id` = `posts`.`id` where `tags`.`id` = `tagables`.`tag_id` and `tagables`.`tagable_type` = 'posts') < 3", sts[0].ToRawSql())

		tags = []Tag{}
		sts = []*goeloquent.Statement{}
		_, err := conn.Model(&Tag{}).WhereHas("Videos", func(st *goeloquent.Statement) {
			st.Where("videos.id", ">", 1)
		}, ">=", 2).Get(&tags)
		assert.Nil(t, err)
		assert.Equal(t, len(tags), 1)
		assert.Equal(t, sts[0].ToRawSql(), "select * from `tags` where (select count(*) from `videos` inner join `tagables` on `tagables`.`tagable_id` = `videos`.`id` where `tags`.`id` = `tagables`.`tag_id` and `videos`.`id` > 1 and `tagables`.`tagable_type` = 'videos') >= 2")

	})
}
func TestCreateMethodProperlyCreatesNewModelMorphedByMany(t *testing.T) {
}
func TestRelationGetResultsMorphedByMany(t *testing.T) {
	goeloquent.DB.SetMorphMaps(map[string]interface{}{
		"posts":  &Post{},
		"videos": &Video{},
	})
	after := "drop table if exists posts; drop table if exists videos; drop table if exists tags; drop table if exists tagables;"
	before := after + "create table posts (id integer primary key auto_increment, title varchar(255), content varchar(255));" +
		"create table videos (id integer primary key auto_increment, title varchar(255), url varchar(255));" +
		"create table tags (id integer primary key auto_increment, name varchar(255), valid integer);" +
		"create table tagables (id integer primary key auto_increment,tag_id integer not null, tagable_id integer not null , tagable_type varchar(255) not null );"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "post1", "content": "post1 content"},
			{"title": "post2", "content": "post2 content"},
			{"title": "post3", "content": "post3 content"},
		})

		conn.Table("videos").Insert([]map[string]interface{}{
			{"title": "video1", "url": "http://example.com/video1"},
			{"title": "video2", "url": "http://example.com/video2"},
			{"title": "video3", "url": "http://example.com/video3"},
		})

		conn.Table("tagables").Insert([]map[string]interface{}{
			{"tag_id": 1, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 1, "tagable_type": "videos"},
			{"tag_id": 2, "tagable_id": 2, "tagable_type": "posts"},
			{"tag_id": 2, "tagable_id": 2, "tagable_type": "videos"},
			{"tag_id": 3, "tagable_id": 1, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 1, "tagable_type": "videos"},
			{"tag_id": 3, "tagable_id": 2, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 2, "tagable_type": "videos"},
			{"tag_id": 3, "tagable_id": 3, "tagable_type": "posts"},
			{"tag_id": 3, "tagable_id": 3, "tagable_type": "videos"},
			{"tag_id": 1, "tagable_id": 5, "tagable_type": "videos"},
			{"tag_id": 1, "tagable_id": 5, "tagable_type": "posts"},
		})
		conn.Table("tags").Insert([]map[string]interface{}{
			{"name": "tag1", "valid": 1},
			{"name": "tag2", "valid": 1},
			{"name": "tag3", "valid": 1},
		})

		var tag Tag
		conn.Model(&Tag{}).Where("id", 3).Get(&tag)
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var posts []Post
		tag.PostsRelation().Get(&posts)
		assert.Equal(t, 3, len(posts))
		assert.Equal(t, "select * from `posts` inner join `tagables` on `tagables`.`tagable_id` = `posts`.`id` where `tagables`.`tag_id` = 3 and `tagables`.`tagable_type` = 'posts'", sts[0].ToRawSql())
		for _, post := range posts {
			assert.Equal(t, post.Pivot["tag_id"], int32(3))
			assert.Equal(t, post.Pivot["tagable_type"], "posts")
		}
	})
}
func TestRelationLoadsResultsMorphedByMany(t *testing.T) {
}
func TestSelfRelationCountMorphedByMany(t *testing.T) {

}
