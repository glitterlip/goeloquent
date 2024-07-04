package tests

import "github.com/glitterlip/goeloquent"

type Tag struct {
	*goeloquent.EloquentModel
	ID     int64   `goelo:"column:tid;primaryKey"`
	Name   string  `goelo:"column:name"`
	Posts  []Post  `goelo:"MorphByMany:PostsRelation"`
	Images []Image `goelo:"MorphByMany:ImagesRelation"`
}

func (t *Tag) ImagesRelation() *goeloquent.MorphByManyRelation {
	return t.MorphByMany(t, &Image{}, "tagables", "tid", "tagable_id", "tag_id", "id", "tagable_type")
}
func (t *Tag) PostsRelation() *goeloquent.MorphByManyRelation {
	return t.MorphByMany(t, &Post{}, "tagables", "tid", "tagable_id", "tag_id", "id", "tagable_type")
}
