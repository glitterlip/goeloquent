package tests

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadingWithColumns(t *testing.T) {
	c, d := CreateRelationTables()
	RunWithDB(c, d, func() {
		//test saving,saved
		var u1, u2, u3, u4 User
		var uu1 User
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
		q := DB.Model(&u1)
		q.Select("id").With("Address:city,state").Find(&uu1, u1.Id)
		assert.Equal(t, []interface{}{"city", "state"}, q.EagerLoad["Address"](uu1.AddressRelation().EloquentBuilder).Columns)
	})
}
