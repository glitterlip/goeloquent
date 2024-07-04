package tests

import "github.com/glitterlip/goeloquent"

type Post struct {
	*goeloquent.EloquentModel
	ID       int64     `goelo:"column:pid;primaryKey"`
	Title    string    `goelo:"column:title"`
	Author   User      `goelo:"BelongsTo:AuthorRelation"`
	AuthorId int64     `goelo:"column:author_id"`
	Comments []Comment `goelo:"HasMany:CommentsRelation"`
	Viewers  []User    `goelo:"BelongsToMany:ViewersRelation"`
	Images   []Image   `goelo:"MorphMany:ImagesRelation"`
	Image    Image     `goelo:"MorphOne:ImageRelation"`
	Tags     []Tag     `goelo:"MorphToMany:TagsRelation"`
}

func (p *Post) AuthorRelation() *goeloquent.BelongsToRelation {
	return p.BelongsTo(p, &User{}, "pid", "author_id")
}
func (p *Post) CommentsRelation() *goeloquent.HasManyRelation {
	return p.HasMany(p, &Comment{}, "pid", "post_id")
}

func (p *Post) ViewersRelation() *goeloquent.BelongsToManyRelation {
	return p.BelongsToMany(p, &User{}, "view_records", "post_id", "user_id", "id", "id")
}
func (p *Post) TagsRelation() *goeloquent.MorphToManyRelation {
	return p.MorphToMany(p, &Tag{}, "pid", "tid", "tagables", "post_id", "tagable_type", "tag_id", "post")
}
func (p *Post) ImagesRelation() *goeloquent.MorphManyRelation {
	return p.MorphMany(p, &Image{}, "pid", "imageable_id", "imageable_type")
}
func (p *Post) ImageRelation() *goeloquent.MorphOneRelation {
	rb := p.MorphOne(p, &Image{}, "pid", "imageable_type", "imageable_id")
	return rb
}
