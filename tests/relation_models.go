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
	Viewers  []UserT   `goelo:"BelongsToMany:ViewersRelation"`
	Images   []Image   `goelo:"MorphMany:ImagesRelation"`
	Image    Image     `goelo:"MorphOne:ImageRelation"`
	Tags     []Tag     `goelo:"MorphToMany:TagsRelation"`
}

func (p *Post) AuthorRelation() *goeloquent.RelationBuilder {
	return p.BelongsTo(p, &UserT{}, "author_id", "id")
}
func (p *Post) CommentsRelation() *goeloquent.RelationBuilder {
	return p.HasMany(p, &Comment{}, "post_id", "id")
}

func (p *Post) ViewersRelation() *goeloquent.RelationBuilder {
	return p.BelongsToMany(p, &UserT{}, "view_records", "user_id", "post_id", "id", "id")
}
func (p *Post) TagsRelation() *goeloquent.RelationBuilder {
	rb := p.MorphToMany(p, &Tag{}, "tagable", "tag_id", "tagable_id", "pid", "tid", "tagable_type")
	rb.EnableLogQuery()
	return rb
}
func (p *Post) ImagesRelation() *goeloquent.RelationBuilder {
	rb := p.MorphMany(p, &Image{}, "imageable_type", "imageable_id", "pid")
	rb.EnableLogQuery()
	return rb
}
func (p *Post) ImageRelation() *goeloquent.RelationBuilder {
	rb := p.MorphOne(p, &Image{}, "imageable_type", "imageable_id", "pid")
	rb.EnableLogQuery()
	return rb
}

const (
	FriendStatusWaiting = 1
	FriendStatusNormal  = 2
	FriendStatusDeleted = 3
	FriendStatusBlocked = 4
)

type UserT struct {
	*goeloquent.EloquentModel
	Id          int64   `goelo:"column:id;primaryKey"`
	UserName    string  `goelo:"column:name"`
	Age         int     `goelo:"column:age"`
	ViewedPosts []Post  `goelo:"BelongsToMany:ViewedPostsRelation"`
	Friends     []UserT `goelo:"BelongsToMany:FriendsRelation"`
	Address     Address `goelo:"HasOne:AddressRelation"`
	Posts       []Post  `goelo:"HasMany:PostsRelation"`

	CreatedAt time.Time    `goelo:"column:created_at;CREATED_AT"`
	UpdatedAt sql.NullTime `goelo:"column:updated_at;UPDATED_AT"`
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
func (u *UserT) PostsRelation() *goeloquent.RelationBuilder {
	return u.HasMany(u, &Post{}, "id", "author_id")
}
func (u *UserT) FriendsRelation() *goeloquent.RelationBuilder {
	rb := u.BelongsToMany(u, &UserT{}, "friends", "user_id", "friend_id", "id", "id")
	rb.Builder.EnableLogQuery()
	return rb
}
func (u *UserT) AddressRelation() *goeloquent.RelationBuilder {
	rb := u.HasOne(u, &Address{}, "id", "user_id")
	rb.EnableLogQuery()
	return rb
}
func (u *UserT) ImagesRelation() *goeloquent.RelationBuilder {
	rb := u.MorphMany(u, &Image{}, "imageable_type", "imageable_id", "id")
	rb.EnableLogQuery()
	return rb
}

type Friends struct {
	ID         int64     `goelo:"column:id;primaryKey"`
	UserId     int64     `goelo:"column:user_id"`
	FriendId   int64     `goelo:"column:friend_id"`
	Status     int       `goelo:"column:status"`
	Time       time.Time `goelo:"column:time"`
	Additional string    `goelo:"column:additional"`
}
type Address struct {
	*goeloquent.EloquentModel
	ID      int64  `goelo:"column:id;primaryKey"`
	UserT   *UserT `goelo:"BelongsTo:UserRelation"`
	UserId  int64  `goelo:"column:user_id"`
	Country string `goelo:"column:country"`
	State   string `goelo:"column:state"`
	City    string `goelo:"column:city"`
	Detail  string `goelo:"column:detail"`
}

func (a *Address) UserRelation() *goeloquent.RelationBuilder {
	rb := a.BelongsTo(a, &Address{}, "id", "post_id")
	rb.EnableLogQuery()
	return rb
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
	ID            int64       `goelo:"column:id;primaryKey"`
	Url           string      `goelo:"column:url"`
	ImageableType string      `goelo:"column:imageable_type"`
	ImageableId   int64       `goelo:"column:imageable_id"`
	Tags          []Tag       `goelo:"MorphToMany:TagsRelation"`
	Imageable     interface{} `goelo:"MorphTo:ImageableRelation"`
}

func (i *Image) ImageableRelation() *goeloquent.RelationBuilder {
	return i.MorphTo(i, "imageable_id", "id", "imageable_type")
}

func (i *Image) TagsRelation() *goeloquent.RelationBuilder {
	return i.MorphToMany(i, &Tag{}, "tagables", "tag_id", "tagable_id", "id", "id", "tagable_type")
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
DROP TABLE IF EXISTS "comment","image","post","tag","tagable","user","users","view_record","friends","address";
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
  "id" int(10) unsigned zerofill NOT NULL AUTO_INCREMENT,
  "tag_id" int(11) NOT NULL,
  "tagable_id" int(11) DEFAULT NULL,
  "status" int(10) unsigned NOT NULL DEFAULT '0',
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
CREATE TABLE "address" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "user_id" int(10) unsigned NOT NULL,
  "country" tinyint(4) NOT NULL,
  "state" varchar(255) NOT NULL,
  "city" varchar(255) NOT NULL,
  "detail" varchar(255) DEFAULT NULL,
  PRIMARY KEY ("id")
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`
	drop = `
DROP TABLE IF EXISTS "comment","image","post","tag","tagable","user","users","view_record","friends","address";
`
	return
}
