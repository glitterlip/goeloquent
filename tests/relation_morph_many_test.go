package tests

import (
	"database/sql"
	"testing"

	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
)

type MorphManyPost struct {
	*goeloquent.EloquentModel
	Id        int64               `json:"id" goelo:"column:id;primaryKey;"`
	Title     string              `json:"title" goelo:"column:title;"`
	Body      string              `json:"body" goelo:"column:body;"`
	CreatedAt sql.NullTime        `json:"created_at" goelo:"column:created_at;CREATED_AT"`
	UpdatedAt sql.NullTime        `json:"updated_at" goelo:"column:updated_at;UPDATED_AT"`
	Comments  []*MorphManyComment `json:"comments" goelo:"MorphMany:CommentsRelation;"`

	ValidComments []MorphManyComment `json:"valid_comments" goelo:"MorphMany:ValidCommentsRelation"`
}

func (m *MorphManyPost) GetTableName(st *goeloquent.Statement) string {
	return "posts"
}
func (m *MorphManyPost) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}

func (m *MorphManyPost) CommentsRelation() *goeloquent.MorphManyRelation {
	return m.MorphMany(m, &MorphManyComment{}, "id", "commentable_id", "commentable_type")
}
func (m *MorphManyPost) ValidCommentsRelation() *goeloquent.MorphManyRelation {
	r := m.MorphMany(m, &MorphManyComment{}, "id", "commentable_id", "commentable_type")
	r.Where("content", "valid")
	return r
}

type MorphManyComment struct {
	*goeloquent.EloquentModel
	Id              int64        `json:"id" goelo:"column:id;primaryKey;"`
	Content         string       `json:"content" goelo:"column:content;"`
	CommentableId   int64        `json:"commentable_id" goelo:"column:commentable_id;"`
	CommentableType string       `json:"commentable_type" goelo:"column:commentable_type;"`
	CreatedAt       sql.NullTime `json:"created_at" goelo:"column:created_at;"`
	UpdatedAt       sql.NullTime `json:"updated_at" goelo:"column:updated_at;"`
}

func (m *MorphManyComment) GetTableName(st *goeloquent.Statement) string {
	return "comments"
}
func (m *MorphManyComment) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}

func TestMorphManyWithDynamicDefaultUseParentModel(t *testing.T) {
}
func TestSaveMethodSetsForeignKeyOnModelMorphMany(t *testing.T) {
}
func TestEagerConstraintsAreProperlyAddedMorphMany(t *testing.T) {
	after := "drop table if exists posts;drop table if exists comments;"
	before := after + `create table posts (id integer primary key auto_increment,title varchar(255) not null ,body varchar(255) not null);create table comments (id integer primary key auto_increment,content varchar(255) not null,commentable_id integer not null ,commentable_type varchar(255) not null,created_at datetime,updated_at datetime);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "post1", "body": "post body 1"},
			{"title": "post2", "body": "post body 2"},
			{"title": "post3", "body": "post body 3"},
		})

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var posts []MorphManyPost
		conn.Model(&MorphManyPost{}).With("Comments").Get(&posts)
		assert.Len(t, posts, 3)
		assert.Len(t, sts, 2)
		assert.Equal(t, "select * from `posts`", sts[0].RawSql)
		assert.Empty(t, sts[0].GetBindings())
		assert.Equal(t, "select * from `comments` where `comments`.`commentable_type` = ? and `comments`.`commentable_id` is not null and `comments`.`commentable_id` in (?, ?, ?)", sts[1].RawSql)
		assert.Equal(t, []interface{}{"MorphManyPost", int64(1), int64(2), int64(3)}, sts[1].GetBindings())

	})
}
func TestModelsAreProperlyMatchedToParentsMorphMany(t *testing.T) {
	after := "drop table if exists posts;drop table if exists comments;"
	before := after + `create table posts (id integer primary key auto_increment,title varchar(255) not null ,body varchar(255) not null);create table comments (id integer primary key auto_increment,content varchar(255) not null,commentable_id integer not null ,commentable_type varchar(255) not null,created_at datetime,updated_at datetime);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "post1", "body": "post body 1"},
			{"title": "post2", "body": "post body 2"},
			{"title": "post3", "body": "post body 3"},
			{"title": "post4", "body": "post body 4"},
		})
		conn.Table("comments").Insert([]map[string]interface{}{
			{"content": "comment4", "commentable_id": 4, "commentable_type": "MorphManyPost"},
			{"content": "comment4", "commentable_id": 4, "commentable_type": "MorphManyPost"},
			{"content": "comment4", "commentable_id": 4, "commentable_type": "MorphManyPost"},
			{"content": "comment3", "commentable_id": 3, "commentable_type": "MorphManyPost"},
			{"content": "comment3", "commentable_id": 3, "commentable_type": "MorphManyPost"},
			{"content": "comment2", "commentable_id": 2, "commentable_type": "MorphManyPost"},
		})
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var posts []MorphManyPost
		conn.Model(&MorphManyPost{}).With("Comments").Get(&posts)
		for _, post := range posts {
			if post.Id == 1 {
				assert.Empty(t, post.Comments)
			} else {
				assert.Equal(t, len(post.Comments), int(post.Id)-1)
				for _, comment := range post.Comments {
					assert.Equal(t, comment.CommentableId, post.Id)
					assert.Equal(t, comment.CommentableType, "MorphManyPost")
				}
			}
		}

	})
}
func TestRelationCountQueryCanBeBuiltMorphMany(t *testing.T) {
	after := "drop table if exists posts;drop table if exists comments;"
	before := after + `create table posts (id integer primary key auto_increment,title varchar(255) not null ,body varchar(255) not null);create table comments (id integer primary key auto_increment,content varchar(255) not null,commentable_id integer not null ,commentable_type varchar(255) not null,created_at datetime,updated_at datetime);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "post1", "body": "post body 1"},
			{"title": "post2", "body": "post body 2"},
			{"title": "post3", "body": "post body 3"},
			{"title": "post4", "body": "post body 4"},
		})
		conn.Table("comments").Insert([]map[string]interface{}{
			{"content": "valid", "commentable_id": 4, "commentable_type": "MorphManyPost"},
			{"content": "invalid", "commentable_id": 4, "commentable_type": "MorphManyPost"},
			{"content": "valid", "commentable_id": 4, "commentable_type": "MorphManyPost"},
			{"content": "invalid", "commentable_id": 3, "commentable_type": "MorphManyPost"},
			{"content": "valid", "commentable_id": 3, "commentable_type": "MorphManyPost"},
			{"content": "invalid", "commentable_id": 2, "commentable_type": "MorphManyPost"},
		})
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var posts []MorphManyPost
		conn.Model(&MorphManyPost{}).Has("Comments").Get(&posts)
		assert.Equal(t, len(sts), 1)
		assert.Equal(t, sts[0].RawSql, "select * from `posts` where exists (select * from `comments` where `posts`.`id` = `comments`.`commentable_id` and `comments`.`commentable_type` = ?)")
		assert.Len(t, posts, 3)
		assert.Len(t, sts, 1)
		assert.Equal(t, []interface{}{"MorphManyPost"}, sts[0].GetBindings())

		sts = []*goeloquent.Statement{}
		var posts2 []MorphManyPost
		conn.Model(&MorphManyPost{}).Has("Comments", ">=", 3).Get(&posts2)
		assert.Equal(t, len(posts2), 1)
		assert.Equal(t, len(sts), 1)
		assert.Equal(t, sts[0].RawSql, "select * from `posts` where (select count(*) from `comments` where `posts`.`id` = `comments`.`commentable_id` and `comments`.`commentable_type` = ?) >= 3")
		assert.Equal(t, []interface{}{"MorphManyPost"}, sts[0].GetBindings())
		assert.Equal(t, posts2[0].Id, int64(4))

		sts = []*goeloquent.Statement{}
		var posts3 []MorphManyPost
		_, err := conn.Model(&MorphManyPost{}).WhereHas("Comments", func(q *goeloquent.Statement) {
			q.Where("content", "valid")
		}, "=", 2).Get(&posts3)
		assert.Nil(t, err)
		assert.Equal(t, len(sts), 1)
		assert.Equal(t, sts[0].RawSql, "select * from `posts` where (select count(*) from `comments` where `posts`.`id` = `comments`.`commentable_id` and `comments`.`commentable_type` = ? and `content` = ?) = 2")
		assert.Equal(t, []interface{}{"MorphManyPost", "valid"}, sts[0].GetBindings())
		assert.Equal(t, len(posts3), 1)
		assert.Equal(t, posts3[0].Id, int64(4))
	})
}
func TestCreateMethodProperlyCreatesNewModelMorphMany(t *testing.T) {
}
func TestRelationGetResultsMorphMany(t *testing.T) {
	after := "drop table if exists posts;drop table if exists comments;"
	before := after + `create table posts (id integer primary key auto_increment,title varchar(255) not null ,body varchar(255) not null);create table comments (id integer primary key auto_increment,content varchar(255) not null,commentable_id integer not null ,commentable_type varchar(255) not null,created_at datetime,updated_at datetime);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("posts").Insert([]map[string]interface{}{
			{"title": "post1", "body": "post body 1"},
			{"title": "post2", "body": "post body 2"},
			{"title": "post3", "body": "post body 3"},
			{"title": "post4", "body": "post body 4"},
		})
		conn.Table("comments").Insert([]map[string]interface{}{
			{"content": "valid", "commentable_id": 4, "commentable_type": "MorphManyPost"},
			{"content": "invalid", "commentable_id": 4, "commentable_type": "MorphManyPost"},
			{"content": "valid", "commentable_id": 4, "commentable_type": "MorphManyPost"},
			{"content": "invalid", "commentable_id": 3, "commentable_type": "MorphManyPost"},
			{"content": "valid", "commentable_id": 3, "commentable_type": "MorphManyPost"},
			{"content": "invalid", "commentable_id": 2, "commentable_type": "MorphManyPost"},
		})
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})

		var post MorphManyPost
		_, err := conn.Model(&MorphManyPost{}).Find(&post, 4)
		assert.Nil(t, err)
		var comments []*MorphManyComment
		post.CommentsRelation().Get(&comments)
		assert.Len(t, sts, 2)
		assert.Equal(t, sts[0].RawSql, "select * from `posts` where `posts`.`id` = ? limit 1")
		assert.Equal(t, []interface{}{4}, sts[0].GetBindings())
		assert.Equal(t, sts[1].RawSql, "select * from `comments` where `comments`.`commentable_id` = ? and `comments`.`commentable_type` = ? and `comments`.`commentable_id` is not null")
		assert.Equal(t, []interface{}{int64(4), "MorphManyPost"}, sts[1].GetBindings())
		assert.Len(t, comments, 3)
		for _, comment := range comments {
			assert.Equal(t, comment.CommentableId, int64(4))
			assert.Equal(t, comment.CommentableType, "MorphManyPost")
		}

		sts = []*goeloquent.Statement{}
		var validComments []*MorphManyComment
		post.ValidCommentsRelation().Get(&validComments)
		assert.Len(t, sts, 1)
		assert.Equal(t, sts[0].RawSql, "select * from `comments` where `comments`.`commentable_id` = ? and `comments`.`commentable_type` = ? and `comments`.`commentable_id` is not null and `content` = ?")
		assert.Equal(t, []interface{}{int64(4), "MorphManyPost", "valid"}, sts[0].GetBindings())
		assert.Len(t, validComments, 2)
		for _, comment := range validComments {
			assert.Equal(t, comment.CommentableId, int64(4))
			assert.Equal(t, comment.CommentableType, "MorphManyPost")
			assert.Equal(t, comment.Content, "valid")
		}

	})
}
func TestRelationLoadsResultsMorphMany(t *testing.T) {
}
func TestSelfRelationCountMorphMany(t *testing.T) {

}
