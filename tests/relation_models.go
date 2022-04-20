package tests

import (
	"database/sql"
	_ "database/sql"
	"errors"
	"fmt"
	goeloquent "github.com/glitterlip/go-eloquent"
	"time"
	_ "time"
)

type Post struct {
	*goeloquent.EloquentModel
	ID       int64     `goelo:"column:pid;primaryKey"`
	Title    string    `goelo:"column:title"`
	Author   UserT     `goelo:"BelongsTo:AuthorRelation"`
	AuthorId int64     `goelo:"column:author_id"`
	Comments []Comment `goelo:"HasMany:CommentsRelation"`
	Address  Address   `goelo:"HasOne:AddressRelation"`
	Viewers  []UserT   `goelo:"BelongsToMany:ViewersRelation"`
	Images   []Image   `goelo:"MorphMany:ImagesRelation"`
	Image    Image     `goelo:"MorphOne:ImageRelation"`
	Tags     []Tag     `goelo:"MorphToMany:TagsRelation"`
}

func (p *Post) AuthorRelation() *goeloquent.RelationBuilder {
	return p.BelongsTo(p, &UserT{}, "user_id", "id")
}
func (p *Post) CommentsRelation() *goeloquent.RelationBuilder {
	return p.HasMany(p, &Comment{}, "post_id", "id")
}
func (p *Post) AddressRelation() *goeloquent.RelationBuilder {
	return p.HasOne(p, &Address{}, "post_id", "id")
}
func (p *Post) ViewersRelation() *goeloquent.RelationBuilder {
	return p.BelongsToMany(p, &UserT{}, "view_record", "user_id", "post_id", "id", "id")
}
func (p *Post) TagsRelation() *goeloquent.RelationBuilder {
	return p.MorphToMany(p, &Tag{}, "tagables", "tag_id", "tagable_id", "id", "id", "tagable_type")
}
func (p *Post) ImagesRelation() *goeloquent.RelationBuilder {
	return p.MorphMany(p, &Image{}, "imageable_type", "imageable_id", "id")
}
func (p *Post) ImageRelation() *goeloquent.RelationBuilder {
	return p.MorphOne(p, &Image{}, "imageable_type", "imageable_id", "id")
}

type UserT struct {
	*goeloquent.EloquentModel
	Id          int64        `goelo:"column:id;primaryKey"`
	UserName    string       `goelo:"column:name"`
	Age         int          `goelo:"column:age"`
	ViewedPosts []Post       `goelo:"BelongsToMany:ViewedPostsRelation"`
	Friends     []UserT      `goelo:"BelongsToMany:FriendsRelation"`
	CreatedAt   time.Time    `goelo:"column:created_at;CREATED_AT"`
	UpdatedAt   sql.NullTime `goelo:"column:updated_at;UPDATED_AT"`
}

func (u *UserT) TableName() string {
	return "user"
}
func (u *UserT) Saving(builder *goeloquent.Builder) (err error) {
	if u.Age <= 0 {
		u.Age = 18
	}
	if u.Age > 100 {
		return errors.New("too old")
	}
	return
}
func (u *UserT) Saved(builder *goeloquent.Builder) (err error) {
	//save avatar
	var avatar Image
	avatar.Url = fmt.Sprintf("cdn.com/statis/users/%d/avatar.png", u.Id)
	avatar.ImageableType = "users"
	avatar.ImageableId = u.Id
	DB.Save(&avatar)
	return
}
func (u *UserT) Creating(builder *goeloquent.Builder) (err error) {
	if u.Id > 0 {
		return errors.New("wrong id")
	}
	return nil
}
func (u *UserT) Created(builder *goeloquent.Builder) (err error) {
	var post Post
	post.Title = "Hello I am here"
	post.AuthorId = u.Id
	_, err = DB.Save(&post)
	if err != nil {
		return err
	}
	return nil
}
func (u *UserT) Deleting(builder *goeloquent.Builder) (err error) {
	if u.Id == 1 {
		return errors.New("can't delete admin")
	}
	return nil
}
func (u *UserT) Deleted(builder *goeloquent.Builder) (err error) {
	DB.Model(&Post{}).Where("author_id", u.Id).Delete()
	return nil
}

func (u *UserT) Updated(builder *goeloquent.Builder) (err error) {
	var avatar Image
	DB.Model().Where("imageable_id", u.Id).Where("imageable_type", "users").First(&avatar)
	avatar.Url = fmt.Sprintf("cdn.com/statis/users/%d/avatar-new.png", u.Id)
	avatar.Save()
	return nil
}
func (u *UserT) Updating(builder *goeloquent.Builder) (err error) {
	if u.IsDirty("Id") {
		return errors.New("id can not be changed")
	}
	return nil
}
func (u *UserT) ViewedPostsRelation() *goeloquent.RelationBuilder {
	return u.BelongsToMany(u, &Post{}, "view_record", "post_id", "users_id", "id", "id")
}
func (u *UserT) FriendsRelation() *goeloquent.RelationBuilder {
	return u.BelongsToMany(u, &UserT{}, "friends", "user_id", "friend_id", "id", "id")
}

type Address struct {
	*goeloquent.EloquentModel
	ID     int64  `goelo:"column:id;primaryKey"`
	UserTT *UserT `goelo:"BelongsTo:UserTTRelation"`
}

func (a *Address) UserTRelation() *goeloquent.RelationBuilder {
	return a.BelongsTo(a, &Address{}, "id", "post_id")
}

type Comment struct {
	*goeloquent.EloquentModel
	ID   int64 `goelo:"column:cid;primaryKey"`
	Post Post  `goelo:"BelongsTo:PostRelation"`
}

func (c *Comment) PostRelation() *goeloquent.RelationBuilder {
	return c.BelongsTo(c, &Post{}, "id", "post_id")
}

type Image struct {
	*goeloquent.EloquentModel
	ID            int64  `goelo:"column:id;primaryKey"`
	Url           string `goelo:"column:url"`
	ImageableType string `goelo:"column:imageable_type"`
	ImageableId   int64  `goelo:"column:imageable_id"`
	Tags          []Tag  `goelo:"MorphToMany:TagsRelation"`
}

func (i *Image) TagsRelation() *goeloquent.RelationBuilder {
	return i.MorphToMany(i, &Tag{}, "tagables", "tag_id", "tagable_id", "id", "id", "tagable_type")
}
func (i *Image) Imageable() *goeloquent.RelationBuilder {
	return i.MorphTo(i, "imageable_id", "id", "imageable_type")
}

type Tag struct {
	*goeloquent.EloquentModel
	ID    int64  `goelo:"column:tid;primaryKey"`
	Name  string `goelo:"column:name"`
	Post  Post   `goelo:"MorphByMany:PostsRelation"`
	Image Image  `goelo:"MorphByMany:ImagesRelation"`
}

func (t *Tag) ImagesRelation() *goeloquent.RelationBuilder {
	return t.MorphByMany(t, &Post{}, "tagables", "id", "tagable_id", "id", "id", "tagable_type")
}
func (t *Tag) PostsRelation() *goeloquent.RelationBuilder {
	return t.MorphByMany(t, &Post{}, "tagables", "id", "tagable_id", "id", "id", "tagable_type")
}
func CreateRelationTables() (create, drop string) {
	create = `
DROP TABLE IF EXISTS "comment","image","post","tag","tagable","user","users","view_record","friends";
CREATE TABLE "comment" (
  "cid" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "post_id" int(11) NOT NULL,
  "content" varchar(255) DEFAULT NULL,
  "user_id" int(11) NOT NULL,
  "upvotes" int(11) NOT NULL DEFAULT '0',
  "downvotes" int(11) NOT NULL DEFAULT '0',
  "comment_id" int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY ("cid") USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "image" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "url" varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '',
  "imageable_id" int(11) DEFAULT NULL,
  "imageable_type" varchar(255) CHARACTER SET utf8 DEFAULT NULL,
  PRIMARY KEY ("id")
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "post" (
  "pid" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "title" varchar(255) NOT NULL DEFAULT '',
  "author_id" int(10) unsigned NOT NULL DEFAULT '0',
  "created_at" datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" datetime DEFAULT NULL,
  "deleted_at" datetime DEFAULT NULL,
  PRIMARY KEY ("pid") USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "tag" (
  "tid" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "name" varchar(255) DEFAULT '',
  PRIMARY KEY ("tid") USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "tagable" (
  "id" int(10) unsigned zerofill NOT NULL,
  "tag_id" int(11) NOT NULL,
  "tagable_id" int(11) DEFAULT NULL,
  "tagable_type" varchar(255) DEFAULT NULL,
  PRIMARY KEY ("id")
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "user" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "age" int(10) unsigned NOT NULL DEFAULT '0',
  "name" varchar(255) NOT NULL DEFAULT '',
  "created_at" datetime DEFAULT CURRENT_TIMESTAMP,
  "updated_at" datetime DEFAULT NULL,
  PRIMARY KEY ("id") USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "view_record" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "user_id" int(11) DEFAULT NULL,
  "post_id" int(11) DEFAULT NULL,
  "view_time" datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "friends" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "user_id" int(11) NOT NULL,
  "friend_id" int(11) NOT NULL,
  "status" tinyint(4) NOT NULL,
  "time" datetime DEFAULT NULL,
  "additional" varchar(255) DEFAULT '',
  PRIMARY KEY ("id")
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`
	drop = `
DROP TABLE IF EXISTS "comment","image","post","tag","tagable","user","users","view_record","friends";
`
	return
}
