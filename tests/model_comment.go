package tests

import "github.com/glitterlip/goeloquent"

type Comment struct {
	*goeloquent.EloquentModel
	ID   int64 `goelo:"column:cid;primaryKey"`
	Post Post  `goelo:"BelongsTo:PostRelation"`
}

func (c *Comment) PostRelation() *goeloquent.BelongsToRelation {
	return c.BelongsTo(c, &Post{}, "id", "post_id")
}
