package tests

import "github.com/glitterlip/goeloquent"

type Image struct {
	*goeloquent.EloquentModel
	ID            int64       `goelo:"column:id;primaryKey"`
	Url           string      `goelo:"column:url"`
	ImageableType string      `goelo:"column:imageable_type"`
	ImageableId   int64       `goelo:"column:imageable_id"`
	Tags          []Tag       `goelo:"MorphToMany:TagsRelation"`
	Imageable     interface{} `goelo:"MorphTo:ImageableRelation"`
}

func (i *Image) ImageableRelation() *goeloquent.MorphToRelation {
	return i.MorphTo(i, "imageable_id", "imageable_type", "id")
}

func (i *Image) TagsRelation() *goeloquent.MorphToManyRelation {
	return i.MorphToMany(i, &Tag{}, "id", "id", "tagables", "tagable_id", "tagable_type", "tag_id")
}
