package tests

import (
	"testing"

	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
)

type MorphOneUser struct {
	*goeloquent.EloquentModel
	Id      int64         `json:"id" goelo:"column:id;primaryKey;"`
	Name    string        `json:"name" goelo:"column:name;"`
	Account string        `json:"account" goelo:"column:account;"`
	Image   MorphOneImage `json:"image" goelo:"MorphOne:ImageRelation;"`
}

func (b *MorphOneUser) GetTableName(st *goeloquent.Statement) string {
	return "users"
}
func (b *MorphOneUser) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}
func (b *MorphOneUser) ImageRelation() *goeloquent.MorphOneRelation {
	return b.MorphOne(b, &MorphOneImage{}, "id", "imageable_id", "imageable_type")
}

type MorphOnePost struct {
	*goeloquent.EloquentModel
	Id    int64  `json:"id" goelo:"column:id;primaryKey;"`
	Title string `json:"title" goelo:"column:title;"`
	Body  string `json:"body" goelo:"column:body;"`
}

type MorphOneImage struct {
	*goeloquent.EloquentModel
	Id            int64  `json:"id" goelo:"column:id;primaryKey;"`
	ImageableId   int64  `json:"imageable_id" goelo:"column:imageable_id;"`
	ImageableType string `json:"imageable_type" goelo:"column:imageable_type;"`
	Url           string `json:"url" goelo:"column:url;"`
}

func (b *MorphOneImage) GetTableName(st *goeloquent.Statement) string {
	return "images"
}
func (b *MorphOneImage) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}

func TestMorphOneWithDynamicDefaultUseParentModel(t *testing.T) {
}
func TestSaveMethodSetsForeignKeyOnModelMorphOne(t *testing.T) {
}
func TestEagerConstraintsAreProperlyAddedMorphOne(t *testing.T) {
	after := "drop table if exists users; drop table if exists images; drop table if exists posts;"
	before := after + `create table users (id integer primary key auto_increment,name varchar(255),account varchar(255));` +
		`create table images (id integer primary key auto_increment,imageable_id integer not null ,imageable_type varchar(255) not null,url varchar(255) not null );` +
		`create table posts (id integer primary key auto_increment,title varchar(255) not null ,body varchar(255) not null);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		_, err := conn.Model(&MorphOneUser{}).Insert([]*MorphOneUser{
			{Name: "customer", Account: "customer"},
			{Name: "vip", Account: "vip"},
			{Name: "suspended", Account: "suspended"},
			{Name: "expired", Account: "expired"},
			{Name: "empty", Account: "empty"},
		})
		assert.Nil(t, err)

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var users []MorphOneUser
		_, err = conn.Model(&MorphOneUser{}).With("Image").Get(&users)
		assert.ErrorIs(t, goeloquent.ErrorNotFound, err)
		assert.Equal(t, len(sts), 2)
		assert.Equal(t, sts[1].RawSql, "select * from `images` where `images`.`imageable_type` = ? and `images`.`imageable_id` is not null and `images`.`imageable_id` in (?, ?, ?, ?, ?)")
		assert.Equal(t, []interface{}{"MorphOneUser", int64(1), int64(2), int64(3), int64(4), int64(5)}, sts[1].GetBindings())

	})
}
func TestModelsAreProperlyMatchedToParentsMorphOne(t *testing.T) {
	after := "drop table if exists users; drop table if exists images; drop table if exists posts;"
	before := after + `create table users (id integer primary key auto_increment,name varchar(255),account varchar(255));` +
		`create table images (id integer primary key auto_increment,imageable_id integer not null ,imageable_type varchar(255) not null,url varchar(255) not null );` +
		`create table posts (id integer primary key auto_increment,title varchar(255) not null ,body varchar(255) not null);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		_, err := conn.Model(&MorphOneUser{}).Insert([]*MorphOneUser{
			{Name: "customer", Account: "customer"},
			{Name: "vip", Account: "vip"},
			{Name: "suspended", Account: "suspended"},
			{Name: "expired", Account: "expired"},
			{Name: "empty", Account: "empty"},
		})
		assert.Nil(t, err)
		_, err = conn.Table("images").Insert([]map[string]interface{}{
			{"imageable_id": 6, "imageable_type": "MorphOnePost", "url": "image1.jpg"},
			{"imageable_id": 5, "imageable_type": "MorphOnePost", "url": "image2.jpg"},
			{"imageable_id": 1, "imageable_type": "MorphOneUser", "url": "image3.jpg"},
			{"imageable_id": 2, "imageable_type": "MorphOneUser", "url": "image4.jpg"},
			{"imageable_id": 3, "imageable_type": "MorphOneUser", "url": "image5.jpg"},
		})

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var users []MorphOneUser
		_, err = conn.Model(&MorphOneUser{}).With("Image").Get(&users)
		assert.Nil(t, err)
		assert.Equal(t, len(sts), 2)
		assert.Equal(t, sts[1].RawSql, "select * from `images` where `images`.`imageable_type` = ? and `images`.`imageable_id` is not null and `images`.`imageable_id` in (?, ?, ?, ?, ?)")
		assert.Equal(t, []interface{}{"MorphOneUser", int64(1), int64(2), int64(3), int64(4), int64(5)}, sts[1].GetBindings())
		assert.Equal(t, 5, len(users))
		assert.Equal(t, users[0].Image.ImageableId, users[0].Id)
		assert.Equal(t, users[1].Image.ImageableId, users[1].Id)
		assert.Equal(t, users[2].Image.ImageableId, users[2].Id)
		assert.Empty(t, users[3].Image)
		assert.Empty(t, users[4].Image)

	})
}
func TestRelationCountQueryCanBeBuiltMorphOne(t *testing.T) {
	after := "drop table if exists users; drop table if exists images; drop table if exists posts;"
	before := after + `create table users (id integer primary key auto_increment,name varchar(255),account varchar(255));` +
		`create table images (id integer primary key auto_increment,imageable_id integer not null ,imageable_type varchar(255) not null,url varchar(255) not null );` +
		`create table posts (id integer primary key auto_increment,title varchar(255) not null ,body varchar(255) not null);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		_, err := conn.Model(&MorphOneUser{}).Insert([]*MorphOneUser{
			{Name: "customer", Account: "customer"},
			{Name: "vip", Account: "vip"},
			{Name: "suspended", Account: "suspended"},
			{Name: "expired", Account: "expired"},
			{Name: "empty", Account: "empty"},
		})
		assert.Nil(t, err)
		_, err = conn.Table("images").Insert([]map[string]interface{}{
			{"imageable_id": 6, "imageable_type": "MorphOnePost", "url": "image1.jpg"},
			{"imageable_id": 5, "imageable_type": "MorphOnePost", "url": "image2.jpg"},
			{"imageable_id": 1, "imageable_type": "MorphOneUser", "url": "image3.jpg"},
			{"imageable_id": 2, "imageable_type": "MorphOneUser", "url": "image4.jpg"},
			{"imageable_id": 3, "imageable_type": "MorphOneUser", "url": "image5.jpg"},
		})
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var users []MorphOneUser
		_, err = conn.Model(&MorphOneUser{}).Has("Image").Get(&users)
		assert.Equal(t, len(sts), 1)
		assert.Equal(t, sts[0].RawSql, "select * from `users` where exists (select * from `images` where `users`.`id` = `images`.`imageable_id` and `images`.`imageable_type` = ?)")
		assert.Equal(t, []interface{}{"MorphOneUser"}, sts[0].GetBindings())

		sts = []*goeloquent.Statement{}
		var users1 []MorphOneUser
		_, err = conn.Model(&MorphOneUser{}).WhereHas("Image", func(q *goeloquent.Statement) {
			q.Where("url", "like", "%3.jpg%")
		}).Get(&users1)
		assert.Equal(t, len(sts), 1)
		assert.Equal(t, "select * from `users` where exists (select * from `images` where `users`.`id` = `images`.`imageable_id` and `images`.`imageable_type` = ? and `url` like ?)", sts[0].RawSql)
		assert.Equal(t, []interface{}{"MorphOneUser", "%3.jpg%"}, sts[0].GetBindings())

	})
}
func TestCreateMethodProperlyCreatesNewModelMorphOne(t *testing.T) {
}
func TestRelationGetResultsMorphOne(t *testing.T) {
	after := "drop table if exists users; drop table if exists images; drop table if exists posts;"
	before := after + `create table users (id integer primary key auto_increment,name varchar(255),account varchar(255));` +
		`create table images (id integer primary key auto_increment,imageable_id integer not null ,imageable_type varchar(255) not null,url varchar(255) not null );` +
		`create table posts (id integer primary key auto_increment,title varchar(255) not null ,body varchar(255) not null);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		_, err := conn.Model(&MorphOneUser{}).Insert([]*MorphOneUser{
			{Name: "customer", Account: "customer"},
			{Name: "vip", Account: "vip"},
			{Name: "suspended", Account: "suspended"},
			{Name: "expired", Account: "expired"},
			{Name: "empty", Account: "empty"},
		})
		assert.Nil(t, err)
		_, err = conn.Table("images").Insert([]map[string]interface{}{
			{"imageable_id": 6, "imageable_type": "MorphOnePost", "url": "image1.jpg"},
			{"imageable_id": 5, "imageable_type": "MorphOnePost", "url": "image2.jpg"},
			{"imageable_id": 1, "imageable_type": "MorphOneUser", "url": "image3.jpg"},
			{"imageable_id": 2, "imageable_type": "MorphOneUser", "url": "image4.jpg"},
			{"imageable_id": 3, "imageable_type": "MorphOneUser", "url": "image5.jpg"},
		})

		var user MorphOneUser
		_, err = conn.Model(&MorphOneUser{}).First(&user)
		assert.Nil(t, err)
		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})
		var image MorphOneImage
		user.ImageRelation().Get(&image)
		assert.Equal(t, len(sts), 1)
		assert.Equal(t, sts[0].RawSql, "select * from `images` where `images`.`imageable_id` = ? and `images`.`imageable_type` = ? and `images`.`imageable_id` is not null")
		assert.Equal(t, []interface{}{int64(1), "MorphOneUser"}, sts[0].GetBindings())
		assert.Equal(t, image.Id, int64(3))

	})
}
func TestRelationLoadsResultsMorphOne(t *testing.T) {
}
func TestSelfRelationCountMorphOne(t *testing.T) {

}
