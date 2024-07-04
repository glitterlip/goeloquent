package tests

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/glitterlip/goeloquent"
	"time"
)

type User struct {
	*goeloquent.EloquentModel
	Id          int64        `goelo:"column:id;primaryKey"`
	UserName    string       `goelo:"column:name"`
	Age         int          `goelo:"column:age"`
	Status      int          `goelo:"column:status"`
	ViewedPosts []Post       `goelo:"BelongsToMany:ViewedPostsRelation"`
	Friends     []User       `goelo:"BelongsToMany:FriendsRelation"`
	Address     Address      `goelo:"HasOne:AddressRelation"`
	Posts       []Post       `goelo:"HasMany:PostsRelation"`
	CreatedAt   time.Time    `goelo:"column:created_at;CREATED_AT"`
	UpdatedAt   sql.NullTime `goelo:"column:updated_at;UPDATED_AT"`
}

func (u *User) TableName() string {
	return "user"
}
func (u *User) EloquentSaving() error {
	if u.Age <= 0 {
		u.Age = 18
	}
	return nil
}
func (u *User) EloquentSaved() error {
	//save avatar
	var avatar Image
	goeloquent.InitModel(&avatar)
	avatar.Url = fmt.Sprintf("cdn.com/statis/users/%d/avatar.png", u.Id)
	avatar.ImageableType = "users"
	avatar.ImageableId = u.Id
	_, err := avatar.Save()
	return err
}
func (u *User) EloquentCreating() error {
	if u.Id != 0 {
		return errors.New("wrong id")
	}
	return nil
}
func (u *User) EloquentCreated() error {
	var post Post
	goeloquent.Init(&post)
	post.Title = "Hello I am here"
	post.AuthorId = u.Id
	_, err := post.Save()
	return err
}
func (u *User) EloquentDeleting() error {
	if u.Id == 1 {
		return errors.New("can't delete admin")
	}
	return nil
}
func (u *User) EloquentDeleted() error {
	_, err := DB.Model(&Post{}).Where("author_id", u.Id).Delete()
	return err
}

func (u *User) EloquentUpdated() error {
	var avatar Image
	DB.Model(&avatar).Where("imageable_id", u.Id).Where("imageable_type", "users").First(&avatar)
	avatar.Url = fmt.Sprintf("cdn.com/statis/users/%d/avatar-new.png", u.Id)
	_, err := avatar.Save()
	return err
}
func (u *User) EloquentUpdating() error {
	if u.IsDirty("Id") {
		return errors.New("id can not be changed")
	}
	return nil

}
func (u *User) EloquentRetrieving() error {
	return nil
}
func (u *User) EloquentRetrieved() error {
	return nil

}
func (u *User) ViewedPostsRelation() *goeloquent.BelongsToManyRelation {
	return u.BelongsToMany(u, &Post{}, "view_record", "users_id", "post_id", "id", "pid")
}
func (u *User) PostsRelation() *goeloquent.HasManyRelation {
	return u.HasMany(u, &Post{}, "id", "author_id")
}
func (u *User) FriendsRelation() *goeloquent.BelongsToManyRelation {
	rb := u.BelongsToMany(u, &User{}, "friends", "user_id", "friend_id", "id", "id")
	return rb
}
func (u *User) AddressRelation() *goeloquent.HasOneRelation {
	rb := u.HasOne(u, &Address{}, "id", "user_id")
	return rb
}
func (u *User) ImagesRelation() *goeloquent.MorphManyRelation {
	rb := u.MorphMany(u, &Image{}, "id", "imageable_id", "imageable_type")
	return rb
}
