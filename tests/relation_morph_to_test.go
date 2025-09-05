package tests

import (
	"testing"

	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
)

type MorphToComment struct {
	*goeloquent.EloquentModel
	Id              int64       `json:"id" goelo:"column:id;primaryKey;"`
	Content         string      `json:"content" goelo:"column:content;"`
	CommentableId   int64       `json:"commentable_id" goelo:"column:commentable_id;"`
	CommentableType string      `json:"commentable_type" goelo:"column:commentable_type;"`
	Commentable     interface{} `json:"commentable" goelo:"MorphTo:CommentRelation"`
	Post            MorphToPost `json:"post" goelo:"MorphTo:PostRelation"`
}

func (m *MorphToComment) GetTableName(st *goeloquent.Statement) string {
	return "comments"
}
func (m *MorphToComment) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}
func (m *MorphToComment) CommentRelation() *goeloquent.MorphToRelation {
	return m.MorphTo(m, "commentable_id", "commentable_type")
}
func (m *MorphToComment) PostRelation() *goeloquent.MorphToRelation {
	config, _ := goeloquent.GetParsedModel(&MorphToPost{})
	r := m.MorphTo(m, "commentable_id", "commentable_type", map[string]*goeloquent.MorphToConfig{
		"Post": {
			ModelConfig: config,
			RelatedKey:  "id",
			MorphType:   "posts",
		},
	})
	return r
}

type MorphToPost struct {
	*goeloquent.EloquentModel
	Id      int64  `json:"id" goelo:"column:id;primaryKey;"`
	Title   string `json:"title" goelo:"column:title;"`
	Content string `json:"content" goelo:"column:content;"`
}

func (m *MorphToPost) GetTableName(st *goeloquent.Statement) string {
	return "posts"
}
func (m *MorphToPost) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}

type MorphToVideo struct {
	*goeloquent.EloquentModel
	Id    int64  `json:"id" goelo:"column:id;primaryKey;"`
	Title string `json:"title" goelo:"column:title;"`
	Url   string `json:"url" goelo:"column:url;"`
}

func (m *MorphToVideo) GetTableName(st *goeloquent.Statement) string {
	return "videos"
}
func (m *MorphToVideo) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}
func TestMorphToWithDynamicDefaultUseParentModel(t *testing.T) {
}
func TestSaveMethodSetsForeignKeyOnModelMorphTo(t *testing.T) {
}
func TestEagerConstraintsAreProperlyAddedMorphTo(t *testing.T) {
	after := "drop table if exists comments;drop table if exists posts;drop table if exists videos;"
	before := after + "create table posts (id integer primary key auto_increment,title text,content text);create table videos (id integer primary key auto_increment,title text,url text);create table comments (id integer primary key auto_increment,content text,commentable_id integer,commentable_type text);"
	goeloquent.DB.SetMorphMaps(map[string]interface{}{
		"posts":  &MorphToPost{},
		"videos": &MorphToVideo{},
	})
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		_, err := conn.Table("comments").Insert([]map[string]interface{}{
			{"content": "this is comment 1", "commentable_id": 1, "commentable_type": "posts"},
			{"content": "this is comment 2", "commentable_id": 3, "commentable_type": "videos"},
			{"content": "this is comment 3", "commentable_id": 2, "commentable_type": "posts"},
			{"content": "this is comment 4", "commentable_id": 4, "commentable_type": "videos"},
		})
		assert.Nil(t, err)

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})

		var comments []MorphToComment
		_, err = conn.Model(&comments).With("Commentable").Get(&comments)
		assert.Equal(t, len(comments), 4)
		assert.Equal(t, len(sts), 3)
		assert.Equal(t, sts[0].RawSql, "select * from `comments`")
		assert.Equal(t, sts[1].RawSql, "select * from `posts` where `id` in (?, ?)")
		assert.Equal(t, sts[2].RawSql, "select * from `videos` where `id` in (?, ?)")
		assert.Empty(t, sts[0].GetBindings())
		assert.Equal(t, sts[1].GetBindings(), []interface{}{"1", "2"})
		assert.Equal(t, sts[2].GetBindings(), []interface{}{"3", "4"})

	})

}
func TestModelsAreProperlyMatchedToParentsMorphTo(t *testing.T) {
	after := "drop table if exists comments;drop table if exists posts;drop table if exists videos;"
	before := after + "create table posts (id integer primary key auto_increment,title text,content text);create table videos (id integer primary key auto_increment,title text,url text);create table comments (id integer primary key auto_increment,content text,commentable_id integer,commentable_type text);"
	goeloquent.DB.SetMorphMaps(map[string]interface{}{
		"posts":  &MorphToPost{},
		"videos": &MorphToVideo{},
	})
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		_, err := conn.Table("comments").Insert([]map[string]interface{}{
			{"content": "this is comment 1", "commentable_id": 1, "commentable_type": "posts"},
			{"content": "this is comment 2", "commentable_id": 3, "commentable_type": "videos"},
			{"content": "this is comment 3", "commentable_id": 2, "commentable_type": "posts"},
			{"content": "this is comment 4", "commentable_id": 4, "commentable_type": "videos"},
			{"content": "this is comment 4", "commentable_id": 5, "commentable_type": "videos"},
		})
		assert.Nil(t, err)
		_, err = conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "this is post 1", "content": "post 1 content"},
			{"title": "this is post 2", "content": "post 2 content"},
		})
		assert.Nil(t, err)
		_, err = conn.Table("videos").Insert([]map[string]interface{}{
			{"title": "this is video 1", "url": "http://video3"},
			{"title": "this is video 2", "url": "http://video4"},
			{"title": "this is video 3", "url": "http://video4"},
			{"title": "this is video 4", "url": "http://video4"},
		})
		assert.Nil(t, err)

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})

		var comments []MorphToComment
		_, err = conn.Model(&comments).With("Commentable").Get(&comments)
		assert.Equal(t, len(comments), 5)
		assert.Equal(t, len(sts), 3)
		assert.Equal(t, sts[0].RawSql, "select * from `comments`")
		assert.Equal(t, sts[1].RawSql, "select * from `posts` where `id` in (?, ?)")
		assert.Equal(t, sts[2].RawSql, "select * from `videos` where `id` in (?, ?, ?)")
		assert.Empty(t, sts[0].GetBindings())
		assert.ElementsMatch(t, sts[1].GetBindings(), []interface{}{"1", "2"})
		assert.ElementsMatch(t, sts[2].GetBindings(), []interface{}{"3", "4", "5"})

		for _, comment := range comments {
			if comment.CommentableId <= 2 {
				assert.NotNil(t, comment.Commentable)
				post, ok := comment.Commentable.(MorphToPost)
				assert.Equal(t, "posts", comment.CommentableType)
				assert.True(t, ok)
				assert.Equal(t, post.Id, comment.CommentableId)
			} else if comment.CommentableId == 5 {
				assert.Nil(t, comment.Commentable)
			} else {
				assert.Equal(t, "videos", comment.CommentableType)
				assert.NotNil(t, comment.Commentable)
				video := comment.Commentable.(MorphToVideo)
				assert.Equal(t, video.Id, comment.CommentableId)
			}
		}

	})
}
func TestRelationCountQueryCanBeBuiltMorphTo(t *testing.T) {
	after := "drop table if exists comments;drop table if exists posts;drop table if exists videos;"
	before := after + "create table posts (id integer primary key auto_increment,title text,content text);create table videos (id integer primary key auto_increment,title text,url text);create table comments (id integer primary key auto_increment,content text,commentable_id integer,commentable_type text);"
	goeloquent.DB.SetMorphMaps(map[string]interface{}{
		"posts":  &MorphToPost{},
		"videos": &MorphToVideo{},
	})
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		_, err := conn.Table("comments").Insert([]map[string]interface{}{
			{"content": "valid", "commentable_id": 1, "commentable_type": "posts"},
			{"content": "valid", "commentable_id": 3, "commentable_type": "videos"},
			{"content": "invalid", "commentable_id": 2, "commentable_type": "posts"},
			{"content": "invalid", "commentable_id": 4, "commentable_type": "videos"},
			{"content": "valid", "commentable_id": 5, "commentable_type": "videos"},
		})
		assert.Nil(t, err)
		_, err = conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "this is post 1", "content": "post 1 content"},
			{"title": "this is post 2", "content": "post 2 content"},
		})
		assert.Nil(t, err)
		_, err = conn.Table("videos").Insert([]map[string]interface{}{
			{"title": "this is video 1", "url": "http://video3"},
			{"title": "this is video 2", "url": "http://video4"},
			{"title": "this is video 3", "url": "http://video4"},
			{"title": "this is video 4", "url": "http://video4"},
		})
		assert.Nil(t, err)

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})

		var comments []MorphToComment
		_, err = conn.Model(&comments).With("Commentable").Get(&comments)
		assert.Equal(t, len(comments), 5)
		assert.Equal(t, len(sts), 3)
		assert.Equal(t, sts[0].RawSql, "select * from `comments`")
		assert.Equal(t, sts[1].RawSql, "select * from `posts` where `id` in (?, ?)")
		assert.Equal(t, sts[2].RawSql, "select * from `videos` where `id` in (?, ?, ?)")
		assert.Empty(t, sts[0].GetBindings())
		assert.ElementsMatch(t, sts[1].GetBindings(), []interface{}{"1", "2"})
		assert.ElementsMatch(t, sts[2].GetBindings(), []interface{}{"3", "4", "5"})

		for _, comment := range comments {
			if comment.CommentableId <= 2 {
				assert.NotNil(t, comment.Commentable)
				post, ok := comment.Commentable.(MorphToPost)
				assert.Equal(t, "posts", comment.CommentableType)
				assert.True(t, ok)
				assert.Equal(t, post.Id, comment.CommentableId)
			} else if comment.CommentableId == 5 {
				assert.Nil(t, comment.Commentable)
			} else {
				assert.Equal(t, "videos", comment.CommentableType)
				assert.NotNil(t, comment.Commentable)
				video := comment.Commentable.(MorphToVideo)
				assert.Equal(t, video.Id, comment.CommentableId)
			}
		}
		sts = []*goeloquent.Statement{}
		var comments1 []MorphToComment
		_, err = conn.Model(&comments1).Has("Post").Get(&comments1)
		assert.Nil(t, err)
		assert.Equal(t, len(comments1), 2)
		assert.Equal(t, len(sts), 1)
		assert.Equal(t, sts[0].RawSql, "select * from `comments` where exists (select * from `posts` where `comments`.`commentable_id` = `posts`.`id` and `comments`.`commentable_type` = ?)")
		assert.Equal(t, sts[0].GetBindings(), []interface{}{"posts"})

		sts = []*goeloquent.Statement{}
		var comments2 []MorphToComment
		_, err = conn.Model(&comments2).With("Post").WhereHas("Post", func(st *goeloquent.Statement) {
			st.Where("title", "this is post 2")
		}).Get(&comments2)
		assert.Nil(t, err)
		assert.Equal(t, len(comments2), 1)
		assert.Equal(t, len(sts), 2)
		assert.Equal(t, sts[0].RawSql, "select * from `comments` where exists (select * from `posts` where `comments`.`commentable_id` = `posts`.`id` and `comments`.`commentable_type` = ? and `title` = ?)")
		assert.Equal(t, sts[0].GetBindings(), []interface{}{"posts", "this is post 2"})
		assert.Equal(t, comments2[0].CommentableId, comments2[0].Post.Id)
		assert.Equal(t, comments2[0].Post.Title, "this is post 2")
	})
}
func TestCreateMethodProperlyCreatesNewModelMorphTo(t *testing.T) {
}
func TestRelationGetResultsMorphTo(t *testing.T) {
	after := "drop table if exists comments;drop table if exists posts;drop table if exists videos;"
	before := after + "create table posts (id integer primary key auto_increment,title text,content text);create table videos (id integer primary key auto_increment,title text,url text);create table comments (id integer primary key auto_increment,content text,commentable_id integer,commentable_type text);"
	goeloquent.DB.SetMorphMaps(map[string]interface{}{
		"posts":  &MorphToPost{},
		"videos": &MorphToVideo{},
	})
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		_, err := conn.Table("comments").Insert([]map[string]interface{}{
			{"content": "valid", "commentable_id": 1, "commentable_type": "posts"},
			{"content": "valid", "commentable_id": 3, "commentable_type": "videos"},
			{"content": "invalid", "commentable_id": 2, "commentable_type": "posts"},
			{"content": "invalid", "commentable_id": 4, "commentable_type": "videos"},
			{"content": "valid", "commentable_id": 5, "commentable_type": "videos"},
		})
		assert.Nil(t, err)
		_, err = conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "this is post 1", "content": "post 1 content"},
			{"title": "this is post 2", "content": "post 2 content"},
		})
		assert.Nil(t, err)
		_, err = conn.Table("videos").Insert([]map[string]interface{}{
			{"title": "this is video 1", "url": "http://video3"},
			{"title": "this is video 2", "url": "http://video4"},
			{"title": "this is video 3", "url": "http://video4"},
			{"title": "this is video 4", "url": "http://video4"},
		})
		assert.Nil(t, err)

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})

		var comment MorphToComment
		conn.Model(&comment).Find(&comment, 3)
		var post MorphToPost
		comment.CommentRelation().Get(&post)
		assert.Equal(t, int64(2), post.Id)
		assert.Equal(t, comment.CommentableId, post.Id)
		assert.Equal(t, comment.CommentableType, "posts")

		assert.Equal(t, len(sts), 2)
		assert.Equal(t, sts[0].RawSql, "select * from `comments` where `comments`.`id` = ? limit 1")
		assert.Equal(t, sts[1].RawSql, "select * from `posts` where `posts`.`id` = ?")
		assert.Equal(t, sts[0].GetBindings(), []interface{}{3})
		assert.Equal(t, sts[1].GetBindings(), []interface{}{int64(2)})

		var comment1 MorphToComment
		conn.Model(&comment1).Find(&comment1, 5)
		var video MorphToVideo
		_, err = comment1.CommentRelation().Get(&video)
		assert.ErrorIs(t, goeloquent.ErrorNotFound, err)
		assert.Nil(t, comment1.Commentable)
	})
}
func TestRelationLoadsResultsMorphTo(t *testing.T) {
}
func TestSelfRelationCountMorphTo(t *testing.T) {

}
