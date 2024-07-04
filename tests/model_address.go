package tests

import "github.com/glitterlip/goeloquent"

type Address struct {
	*goeloquent.EloquentModel
	ID      int64  `goelo:"column:id;primaryKey"`
	User    *User  `goelo:"BelongsTo:UserRelation"`
	UserId  int64  `goelo:"column:user_id"`
	Country string `goelo:"column:country"`
	State   string `goelo:"column:state"`
	City    string `goelo:"column:city"`
	Detail  string `goelo:"column:detail"`
}

func (a *Address) UserRelation() *goeloquent.BelongsToRelation {
	rb := a.BelongsTo(a, &User{}, "id", "user_id")
	return rb
}
