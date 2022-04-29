package tests

import (
	_ "fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestBelongsToMany(t *testing.T) {
	c, d := CreateRelationTables()
	RunWithDB(c, d, func() {
		//test saving,saved
		var u1, u2, u3, u4, u5 UserT
		var us []UserT
		var uu1, uu2 UserT
		u1.UserName = "u1"
		u2.UserName = "u2"
		u3.UserName = "u3"
		u4.UserName = "u4"
		u5.UserName = "u5"
		DB.Save(&u1)
		DB.Save(&u2)
		DB.Save(&u3)
		DB.Save(&u4)
		DB.Save(&u5)
		DB.Table("friends").Insert([]map[string]interface{}{
			{
				"user_id":    u1.Id,
				"friend_id":  u2.Id,
				"status":     FriendStatusWaiting,
				"time":       time.Now(),
				"additional": "hi I am John",
			},
			{
				"user_id":    u1.Id,
				"friend_id":  u3.Id,
				"status":     FriendStatusNormal,
				"time":       time.Now(),
				"additional": "Jack",
			},
			{
				"user_id":    u1.Id,
				"friend_id":  u4.Id,
				"status":     FriendStatusNormal,
				"time":       time.Now(),
				"additional": "Sam",
			},
			{
				"user_id":    u1.Id,
				"friend_id":  u5.Id,
				"status":     FriendStatusBlocked,
				"time":       time.Now(),
				"additional": "moron",
			},
			{
				"user_id":    u2.Id,
				"friend_id":  u3.Id,
				"status":     FriendStatusNormal,
				"time":       time.Now(),
				"additional": "Sara",
			},
			{
				"user_id":    u2.Id,
				"friend_id":  u4.Id,
				"status":     FriendStatusNormal,
				"time":       time.Now(),
				"additional": "Anna",
			},
			{
				"user_id":    u2.Id,
				"friend_id":  u5.Id,
				"status":     FriendStatusWaiting,
				"time":       time.Now(),
				"additional": "hi I am Alice",
			},
			{
				"user_id":    u2.Id,
				"friend_id":  u1.Id,
				"status":     FriendStatusDeleted,
				"time":       time.Now(),
				"additional": "",
			},
		})

		//test find
		DB.Model(&u1).With("Friends").WherePivot("status", FriendStatusNormal).WithPivot("status", "user_id", "friend_id", "additional").Find(&uu1, u1.Id)
		assert.Equal(t, 2, len(uu1.Friends))
		for _, friend := range uu1.Friends {
			assert.Equal(t, friend.Pivot["status"].String, strconv.Itoa(FriendStatusNormal))
			assert.Equal(t, friend.Pivot["user_id"].String, strconv.Itoa(int(u1.Id)))
		}
		//test get
		result, err := DB.Model(&u1).With("Friends").WhereIn("id", []int64{u1.Id, u2.Id}).Get(&us)
		assert.Nil(t, err)
		count, _ := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), count)
		for _, u := range us {
			for _, friend := range u.Friends {
				assert.Equal(t, friend.Pivot["user_id"].String, strconv.Itoa(int(u.Id)))
			}
		}
		//test not find
		c, err := DB.Model(&u1).With("Friends").Find(&uu2, u3.Id)
		count, _ = c.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, 0, len(uu2.Friends))
		//test lazyload
		var lazy UserT
		var lazyUser []UserT
		DB.Model(&u1).Find(&lazy, u1.Id)
		lazy.FriendsRelation().WherePivot("status", FriendStatusNormal).WithPivot("status", "user_id", "additional").Get(&lazyUser)
		assert.Equal(t, 2, len(lazyUser))
		for _, friend := range lazy.Friends {
			assert.Equal(t, friend.Pivot["status"].String, strconv.Itoa(FriendStatusNormal))
			assert.Equal(t, friend.Pivot["user_id"].String, strconv.Itoa(int(lazy.Id)))
		}
	})
	//TODO: test create update
}
