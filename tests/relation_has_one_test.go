package tests

import (
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHasOne(t *testing.T) {

	CreateUsers()
	//test get eager load
	//select * from `phones` where `phones`.`user_id` is not null and `phones`.`user_id` in (?,?,?,?,?)
	//[1 2 3 4 5]
	var us []User
	r, e := DB.Model(&us).With("Phone").Get(&us)
	assert.Nil(t, e)
	assert.Equal(t, r.Count, int64(5))
	for _, u := range us {
		assert.NotNil(t, u.Phone)
		assert.Equal(t, u.Phone.UserId, u.ID)
		assert.True(t, u.Phone.IsBooted)
	}

	//test get eager load with constraints  (espically with orwhere clause)
	//select * from `phones` where `phones`.`user_id` is not null and `phones`.`user_id` in (?,?,?,?,?) and (`country` = ? or `tel` = ?)
	//[1 2 3 4 5 +2 1563103]
	var us2 []User
	r, e = DB.Model(&us2).With(map[string]func(builder *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"Phone": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return q.Where(func(b *goeloquent.Builder) *goeloquent.Builder {
				return b.Where("country", "+2").OrWhere("tel", "1563103")
			})
		},
	}).Get(&us2)
	assert.Nil(t, e)
	c := 0
	for _, user := range us2 {
		if user.Phone != nil {
			c++
			assert.True(t, user.Phone.Country == "+2" || user.Phone.Tel == "1563103")
		}
	}
	assert.Equal(t, c, 2)
	//test find eager load with constraints
	//select * from `phones` where `phones`.`user_id` is not null and `phones`.`user_id` in (?) and (`country` = ? or `tel` = ?)
	//[1 +2 1563103]
	var u2 User
	r, e = DB.Model(&u2).With(map[string]func(b *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"Phone": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return q.Where(func(b *goeloquent.Builder) *goeloquent.Builder {
				return b.Where("country", "+2").OrWhere("tel", "1563103")
			})

		},
	}).Find(&u2, 1)
	assert.Nil(t, e)
	assert.NotNil(t, u2.Phone)
	assert.True(t, u2.Phone.Country == "+2" || u2.Phone.Tel == "1563103")
	assert.True(t, u2.Phone.IsBooted)

	//test relation get
	//select * from `phones` where `phones`.`user_id` = ? and `phones`.`user_id` is not null
	//[1]
	var u3 User
	r, e = DB.Model(&u3).Find(&u3, 1)
	assert.Nil(t, e)
	assert.Nil(t, u3.Phone)
	var phone Phone
	r, e = u3.PhoneRelation().Get(&phone)
	assert.Nil(t, e)
	assert.NotNil(t, phone)
	assert.Equal(t, phone.UserId, u3.ID)
	assert.True(t, phone.IsBooted)
	var phone2 Phone
	var u4 User
	//select * from `phones` where `phones`.`user_id` = ? and `phones`.`user_id` is not null and `country` = ?
	//[1 +2]
	r, e = DB.Model(&u4).Find(&u4, 1)
	r, e = u4.PhoneRelation().Where("country", "+2").Get(&phone2)

	var p Phone
	DB.Model(&p).First(&p)

	var u5 User
	r, e = DB.Model(&u5).WithCount("Phone").First(&u5)
	assert.Equal(t, u5.WithAggregates["PhoneCount"], float64(1))
	assert.Equal(t, r.Sql, "select *, (select Count(*) from `phones` where `user_models`.`id` = `phones`.`user_id`) as `goelo_orm_aggregate_PhoneCount` from `user_models` where `user_models`.`deleted_at` is null limit 1")
}
