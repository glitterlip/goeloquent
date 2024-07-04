package tests

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHasOne(t *testing.T) {
	c, d := CreateRelationTables()
	RunWithDB(c, d, func() {
		//test saving,saved
		var u1, u2, u3, u4 User
		var uu1, uu3 User
		var us []User
		u1.UserName = "u1"
		u2.UserName = "u2"
		u3.UserName = "u3"
		u4.UserName = "u4"
		DB.Create(&u1)
		DB.Create(&u2)
		DB.Create(&u3)
		DB.Create(&u4)
		addresses := []map[string]interface{}{
			{
				"user_id": u2.Id,
				"country": 1,
				"state":   "California",
				"city":    "Sacramento",
				"detail":  "Golden State",
			},
			{
				"user_id": u1.Id,
				"country": 1,
				"state":   "Florida",
				"city":    "Tallahassee",
				"detail":  "Sunshine State",
			},
			{
				"user_id": u4.Id,
				"country": 1,
				"state":   "Delaware",
				"city":    "Dover",
				"detail":  "First State",
			},
		}
		DB.Table("address").Insert(&addresses)
		//test find
		DB.Model(&u1).With("Address").Find(&uu1, u1.Id)
		assert.Equal(t, uu1.Address.UserId, u1.Id)
		//test get
		result, err := DB.Model(&User{}).With("Address").WhereIn("id", []int64{u2.Id, u4.Id}).Get(&us)
		assert.Nil(t, err)
		count, _ := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), count)
		assert.Equal(t, 2, len(us))
		for _, u := range us {
			assert.Equal(t, u.Address.UserId, u.Id)
		}
		//test not find
		c, err := DB.Model(&u1).With("Address").Find(&uu3, u3.Id)
		count, _ = c.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, int64(0), uu3.Address.UserId)
		//test lazyload
		var lazy User
		var lazyAddress Address
		DB.Model(&u1).Find(&lazy, u2.Id)
		lazy.AddressRelation().Get(&lazyAddress)
		assert.Equal(t, lazy.Id, lazyAddress.UserId)
	})
}
