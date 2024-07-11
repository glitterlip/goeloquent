package tests

import (
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHasOne(t *testing.T) {

	CreateRelationTables()
	CreateUsers()
	//test get eager load
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
	var u User
	r, e = DB.Model(&u).With(map[string]func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder{
		"Phone": func(q *goeloquent.EloquentBuilder) *goeloquent.EloquentBuilder {
			return q.Where(func(b *goeloquent.Builder) *goeloquent.Builder {
				return b.Where("country", "+2").OrWhere("tel", "1563103")
			})

		},
	}).Find(&u, 1)
	assert.Nil(t, e)
	assert.NotNil(t, u.Phone)
	assert.True(t, u.Phone.Country == "+2" || u.Phone.Tel == "1563103")

	//test model load
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
}
