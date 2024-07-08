package tests

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"github.com/glitterlip/goeloquent"
	"strings"
)

type User struct {
	*goeloquent.EloquentModel
	ID        int64        `goelo:"column:id;primaryKey"`
	Name      string       `goelo:"column:name"`
	Age       uint8        `goelo:"column:age"`
	Email     string       `goelo:"column:email"`
	Status    uint8        `goelo:"column:status"`
	Info      UserInfo     `goelo:"column:info"`
	CreatedAt sql.NullTime `goelo:"column:created_at;CREATED_AT"`
	UpdatedAt sql.NullTime `goelo:"column:updated_at;UPDATED_AT"`
	DeletedAt sql.NullTime `goelo:"column:deleted_at;DELETED_AT"`
	Phone     *Phone       `goelo:"HasOne:PhoneRelation"`
	Posts     []Post       `goelo:"HasMany:PostRelation"`
	Images    []*Image     `goelo:"MorphMany:ImageRelation"`
	Tags      UserTag      `goelo:"column:tags"`
}

type UserTag struct {
	Strs []string
}

func (c *UserTag) Scan(src any) error {
	str, ok := src.([]byte)
	if !ok {
		return nil
	} else {
		c.Strs = strings.Split(string(str), ",")
		return nil
	}
}
func (c UserTag) Value() (driver.Value, error) {
	return strings.Join(c.Strs, ","), nil
}

type UserInfo struct {
	Verified bool
	Age      int
	Address  string
	Links    []string
}

func (u UserInfo) Value() (driver.Value, error) {
	return json.Marshal(u)
}
func (u *UserInfo) Scan(src any) error {

	str, ok := src.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(str, u)
}
func (u *User) TableName() string {
	return "user_models"
}

func (u *User) PhoneRelation() *goeloquent.HasOneRelation {
	return u.HasOne(u, &Phone{}, "id", "user_id")
}

func (u *User) PostRelation() *goeloquent.HasManyRelation {
	return u.HasMany(u, &Post{}, "id", "user_id")
}

func (u *User) ImageRelation() *goeloquent.MorphManyRelation {
	return u.MorphMany(u, &Image{}, "id", "imageable_id", "imageable_type")
}
func (u *User) EloquentGetGuarded() map[string]struct{} {
	return map[string]struct{}{
		"id":     {},
		"status": {},
	}
}

func (u *User) EloquentGetDefaultAttributes() map[string]interface{} {
	return map[string]interface{}{
		"status": uint8(2),
	}
}
func (u *User) EloquentGetWithRelationCounts() map[string]goeloquent.RelationFunc {
	return map[string]goeloquent.RelationFunc{
		"Posts": func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return builder.Where("stauts", 1)
		},
	}
}

func (u *User) EloquentGetWithRelations() map[string]goeloquent.RelationFunc {
	return map[string]goeloquent.RelationFunc{
		"Phone": func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			builder.Where("country", "+2")
			return builder
		},
	}
}

type Phone struct {
	*goeloquent.EloquentModel
	ID      int64  `goelo:"column:id;primaryKey"`
	UserId  int64  `goelo:"column:user_id"`
	Country string `goelo:"column:country"`
	Tel     string `goelo:"column:tel"`
	Owner   User   `goelo:"BelongsTo:UserRelation"`
}

func (p *Phone) TableName() string {
	return "phones"
}

func (p *Phone) UserRelation() *goeloquent.BelongsToRelation {
	return p.BelongsTo(p, &User{}, "user_id", "id")
}

type Post struct {
	*goeloquent.EloquentModel
	ID     int64                  `goelo:"column:id;primaryKey"`
	UserId int64                  `goelo:"column:user_id"`
	Title  string                 `goelo:"column:title"`
	Tags   []string               `goelo:"column:tags"`
	Status uint8                  `goelo:"column:status"`
	User   User                   `goelo:"BelongsTo:UserRelation"`
	Meta   map[string]interface{} `goelo:"column:meta"`
	Images []Image                `goelo:"MorphMany:ImageRelation"`
}

func (p *Post) TableName() string {
	return "posts"
}

func (p *Post) UserRelation() *goeloquent.BelongsToRelation {
	return p.BelongsTo(p, &User{}, "user_id", "id")
}

func (p *Post) ImageRelation() *goeloquent.MorphManyRelation {
	return p.MorphMany(p, &Image{}, "id", "imageable_id", "imageable_type")
}

type Image struct {
	*goeloquent.EloquentModel
	ID            int64       `goelo:"column:id;primaryKey"`
	Path          string      `goelo:"column:path"`
	Size          int64       `goelo:"column:size"`
	Driver        string      `goelo:"column:driver"`
	ImageableId   int64       `goelo:"column:imageable_id"`
	ImageableType string      `goelo:"column:imageable_type"`
	Remark        string      `goelo:"column:remark"`
	Imageable     interface{} `goelo:"MorphTo:ImageableRelation"`
}

func (i *Image) TableName() string {
	return "images"
}

func (i *Image) ImageableRelation() *goeloquent.MorphToRelation {
	return i.MorphTo(i, "imageable_id", "imageable_type", "id")
}

type Video struct {
	*goeloquent.EloquentModel
	ID     int64  `goelo:"column:id;primaryKey"`
	UserID int64  `goelo:"column:user_id"`
	Size   int64  `goelo:"column:size"`
	Path   string `goelo:"column:path"`
	Title  string `goelo:"column:title"`
	Tags   []Tag  `goelo:"MorphToMany:TagsRelation"`
}

func (v *Video) TableName() string {
	return "videos"
}

func (v *Video) TagsRelation() *goeloquent.MorphToManyRelation {
	return v.MorphToMany(v, "tagable_id", "tagable_type", "tagables", "tag_id", "tag_id", "id", "id")
}

type Tag struct {
	*goeloquent.EloquentModel
	ID      int64   `goelo:"column:id;primaryKey"`
	Name    string  `goelo:"column:name"`
	Related int     `goelo:"column:related"`
	Posts   []Post  `goelo:"MorphByMany:PostsRelation"`
	Videos  []Video `goelo:"MorphByMany:VideosRelation"`
}

func (t *Tag) TableName() string {
	return "tags"
}

func (t *Tag) PostsRelation() *goeloquent.MorphByManyRelation {
	return t.MorphByMany(t, &Post{}, "tagables", "id", "id", "tag_id", "tagable_id", "tagable_type")
}

func (t *Tag) VideosRelation() *goeloquent.MorphByManyRelation {
	return t.MorphByMany(t, &Video{}, "tagables", "id", "id", "tag_id", "tagable_id", "tagable_type")
}

type Comment struct {
	*goeloquent.EloquentModel
	ID              int64        `goelo:"column:id;primaryKey"`
	Content         string       `goelo:"column:content"`
	Likes           int          `goelo:"column:likes"`
	CommentableId   int64        `goelo:"column:commentable_id"`
	CommentableType string       `goelo:"column:commentable_type"`
	Commentable     interface{}  `goelo:"MorphTo:CommentableRelation"`
	Parent          *Comment     `goelo:"BelongsTo:ParentRelation"`
	Children        []*Comment   `goelo:"HasMany:ChildrenRelation"`
	CreatedAt       sql.NullTime `goelo:"column:created_at"`
	UpdatedAt       sql.NullTime `goelo:"column:updated_at"`
	DeletedAt       sql.NullTime `goelo:"column:deleted_at"`
}

func (c *Comment) TableName() string {
	return "comments"
}

func (c *Comment) CommentableRelation() *goeloquent.MorphToRelation {
	return c.MorphTo(c, "commentable_id", "commentable_type", "id")
}

func (c *Comment) ParentRelation() *goeloquent.BelongsToRelation {
	return c.BelongsTo(c, &Comment{}, "parent_id", "id")
}

func (c *Comment) ChildrenRelation() *goeloquent.HasManyRelation {
	return c.HasMany(c, &Comment{}, "id", "parent_id")
}
