package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	//test not init
	var user1, user2 User
	assert.Nil(t, user1.EloquentModel)
	assert.PanicsWithValuef(t, "call Init(&model) first,or set modelPointer by call Save(&model)", func() {
		user1.Save()
	}, "should panic")
	assert.True(t, !reflect.ValueOf(user1.EloquentModel).Elem().IsValid())

	//test inited
	goeloquent.Init(&user2)
	assert.True(t, user2.IsBooted)
	assert.True(t, reflect.ValueOf(user2.EloquentModel).Elem().IsValid())
}
func TestCreateMethod(t *testing.T) {

	info := UserInfo{
		Verified: true,
		Age:      18,
		Address:  "NY",
		Links:    []string{"https://x.com", "https://twitch.com"},
	}

	_, e := DB.Raw().Exec("truncate table user_models")
	assert.Nil(t, e)
	var u User
	u.Name = "Alice"
	u.Email = "Alice@gmail.com"
	u.Age = 18
	u.Status = 1
	u.Tags.Strs = []string{"tag1", "tag2"}
	u.Info = info

	r, e := u.Save(&u)
	assert.True(t, u.IsBooted)
	assert.True(t, u.WasRecentlyCreated)
	assert.True(t, u.Exists)
	assert.Nil(t, e)

	a, err := r.RowsAffected()
	assert.Equal(t, int64(1), u.ID)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), a)

	//test create with map EloquentGetGuarded
	user := new(User)
	r1, err1 := user.Fill(map[string]interface{}{
		"name": "Bob",
		"age":  uint8(20),
		"id":   4,
	}, false, user).Save()

	assert.Nil(t, err1)
	c, _ := r1.RowsAffected()
	assert.Equal(t, c, int64(1))
	assert.Subset(t, r1.Bindings, []interface{}{"Bob", uint8(20)})
	assert.Equal(t, len(r1.Bindings), 3) // created_at

	assert.Equal(t, int64(2), user.ID)
	assert.Equal(t, "Bob", user.Name)
	assert.Equal(t, uint8(20), user.Age)

	//test create with only/except

	var ut User
	DB.Init(&ut)
	ut.Name = "ut"
	ut.Email = "ut@gmail.com"
	ut.Age = 28
	ut.Status = 2
	ut.Tags.Strs = []string{"tag3", "tag3"}
	ut.Info = info
	r2, err2 := ut.Only("name", "email").Save(&ut)
	assert.Nil(t, err2)
	assert.Equal(t, int64(3), ut.ID)
	assert.Equal(t, len(r2.Bindings), 3)
	assert.Subset(t, r2.Bindings, []interface{}{"ut", "ut@gmail.com"})

	var ut2 User
	DB.Init(&ut2)
	ut2.Name = "ut"
	ut2.Email = "ut@gmail.com"
	ut2.Age = 28
	ut2.Status = 2
	ut2.Tags.Strs = []string{"tag3", "tag3"}
	ut2.Info = info
	r3, err3 := ut2.Except("name").Save()
	assert.Nil(t, err3)
	assert.Equal(t, int64(4), ut2.ID)
	istr, _ := json.Marshal(info)
	assert.Subset(t, r3.Bindings, []interface{}{"ut@gmail.com", uint8(28), uint8(2), "tag3,tag3", istr})
	assert.Equal(t, len(r3.Bindings), 6)
}
func TestFindMethod(t *testing.T) {
	//testFindMethod
	DB.Raw().Exec("truncate table user_models")

	var us []map[string]interface{}
	for i := 1; i < 10; i++ {
		us = append(us, map[string]interface{}{
			"name":   fmt.Sprintf("user%d", i),
			"age":    i,
			"email":  fmt.Sprintf("user%d@gmail.com", i),
			"status": 0,
		})
	}
	r, e := DB.Table("user_models").Insert(us)
	c, _ := r.RowsAffected()
	assert.Equal(t, int64(9), c)
	assert.Nil(t, e)

	var user User
	b := DB.Model(&user).WithTrashed()
	_, err := b.Find(&user, 4)
	assert.Nil(t, err)

	assert.Equal(t, b.GetRawBindings()[goeloquent.TYPE_WHERE], []interface{}{4})

	assert.Equal(t, int64(4), user.ID)
	//testFindManyMethod
	var users []User
	b1 := DB.Model(&user)
	b1.Find(&users, []interface{}{5, 7})
	assert.Equal(t, b1.GetRawBindings()[goeloquent.TYPE_WHERE], []interface{}{5, 7})
	assert.Equal(t, len(users), 2)
	assert.ElementsMatch(t, []interface{}{users[0].ID, users[1].ID}, []interface{}{int64(5), int64(7)})

}

func TestFirst(t *testing.T) {
	//testFirstMethod
	TestCreateMethod(t)
	info := UserInfo{
		Verified: true,
		Age:      18,
		Address:  "NY",
		Links:    []string{"https://x.com", "https://twitch.com"},
	}

	var created User
	r, e := DB.Model().First(&created)
	assert.Nil(t, e)
	assert.Equal(t, int64(1), r.Count)
	assert.Equal(t, "select * from `user_models` limit 1", r.Sql)

	assert.Equal(t, int64(1), created.ID)
	assert.Equal(t, "Alice", created.Name)
	assert.Equal(t, uint8(18), created.Age)
	assert.Equal(t, uint8(1), created.Status)
	assert.Equal(t, created.Tags.Strs, []string{"tag1", "tag2"})
	assert.Equal(t, created.Info, info)

}

func TestQualifyColumn(t *testing.T) {
	//testQualifyColumn
	//testQualifyColumns
}

func TestAttrs(t *testing.T) {
	DB.Raw().Exec("truncate table user_models")

	var u1 User
	DB.Init(&u1)

	u1.Name = "u1"
	assert.Equal(t, u1.GetDirty(), map[string]interface{}{
		"Name":   "u1",
		"Status": uint8(2),
	})
	assert.Empty(t, u1.GetChanges())
	u1.Age = 11
	assert.True(t, u1.IsDirty("Age"))
	assert.Empty(t, u1.GetChanges())

	save, err := DB.Save(&u1)
	assert.Nil(t, err)
	count, _ := save.RowsAffected()
	assert.Equal(t, int64(1), count)
	assert.Nil(t, err)
	assert.Equal(t, u1.ID, int64(1))
	assert.Empty(t, u1.GetDirty())
	assert.Equal(t, map[string]interface{}{
		"Age": uint8(11),
		"CreatedAt": sql.NullTime{
			Time:  u1.CreatedAt.Time,
			Valid: true,
		},
		"ID":     int64(1),
		"Name":   "u1",
		"Status": uint8(2),
	}, u1.GetChanges())

}

type UserEvent struct {
	*goeloquent.EloquentModel
	ID        int64        `goelo:"column:id;primaryKey"`
	Name      string       `goelo:"column:name"`
	Age       uint8        `goelo:"column:age"`
	Email     string       `goelo:"column:email"`
	Status    uint8        `goelo:"column:status"`
	Info      UserInfo     `goelo:"column:info"`
	CreatedAt sql.NullTime `goelo:"column:created_at;CREATED_AT"`
	UpdatedAt sql.NullTime `goelo:"column:updated_at;UPDATED_AT"`
}

func (u *UserEvent) TableName() string {
	return "user_models"
}
func (u *UserEvent) EloquentSaving() error {

	if !strings.Contains(u.Name, "saving") {
		return errors.New("cant save")
	}
	return nil
}
func (u *UserEvent) EloquentSaved() error {

	b, _ := json.Marshal(u.Info)
	DB.Table("logs").Insert(map[string]interface{}{
		"logable_id":   u.ID,
		"logable_type": "user",
		"operator_id":  u.Context.Value("user_id"),
		"remark":       "user saved",
		"meta":         string(b),
	})
	return nil
}
func (u *UserEvent) EloquentCreating() error {

	if !strings.Contains(u.Name, "creating") {
		return errors.New("cant create")
	}
	return nil
}
func (u *UserEvent) EloquentCreated() error {
	b, _ := json.Marshal(u.Info)
	DB.Table("logs").Insert(map[string]interface{}{
		"logable_id":   u.ID,
		"logable_type": "user",
		"operator_id":  u.Context.Value("user_id"),
		"remark":       "user created",
		"meta":         string(b),
	})
	return nil

}
func (u *UserEvent) EloquentUpdating() error {
	if !strings.Contains(u.Name, "updating") {
		return errors.New("cant update")
	}
	return nil
}
func (u *UserEvent) EloquentUpdated() error {
	b, _ := json.Marshal(u.Info)
	DB.Table("logs").Insert(map[string]interface{}{
		"logable_id":   u.ID,
		"logable_type": "user",
		"operator_id":  u.Context.Value("user_id"),
		"remark":       "user updated",
		"meta":         string(b),
	})
	return nil
}
func (u *UserEvent) EloquentDeleting() error {
	if !strings.Contains(u.Name, "deleting") {
		return errors.New("cant delete")
	}
	return nil
}
func (u *UserEvent) EloquentDeleted() error {
	DB.Table("logs").Insert(map[string]interface{}{
		"logable_id":   u.ID,
		"logable_type": "user",
		"operator_id":  u.Context.Value("user_id"),
		"remark":       "user deleted",
		"meta":         "",
	})
	return nil
}
func (u *UserEvent) EloquentRetrieving() error {
	if id, ok := u.Context.Value("user_id").(int); ok && id != 2 {
		return errors.New("cant retrieve")
	}
	return nil
}
func (u *UserEvent) EloquentRetrieved() error {
	DB.Table("logs").Insert(map[string]interface{}{
		"logable_id":   u.ID,
		"logable_type": "user",
		"operator_id":  u.ID,
		"remark":       "user retrieved",
		"meta":         "",
	})
	return nil
}
func TestEvents(t *testing.T) {
	DB.Raw().Exec("truncate table user_models")
	DB.Raw().Exec("truncate table logs")
	//test saving,saved

	var uc UserEvent
	var c, l int

	DB.Init(&uc)
	ctx := context.WithValue(context.Background(), "user_id", 1)
	uc.WithContext(ctx)
	uc.Name = "test"
	uc.Email = "test"
	uc.Status = 1
	uc.Age = 1
	_, e := uc.Save()
	//test saving
	assert.Error(t, e, "cant save")
	DB.Table("user_models").Count(&c)
	assert.Equal(t, c, 0)

	uc.Name = "saving"
	_, e = uc.Save()
	assert.Error(t, e, "cant create")
	//test creating
	DB.Table("user_models").Count(&c)
	assert.Equal(t, c, 0)

	uc.Name = "saving,creating"
	_, e = uc.Save()
	assert.Nil(t, e)
	assert.Equal(t, int64(1), uc.ID)
	DB.Table("user_models").Count(&c)
	assert.Equal(t, c, 1)

	//test created
	DB.Table("logs").Where("logable_id", uc.ID).
		Where("remark", "user created").
		Where("operator_id", 1).
		Count(&l)
	assert.Equal(t, l, 1)
	l = 0
	//test saved
	DB.Table("logs").Where("logable_id", uc.ID).
		Where("remark", "user saved").
		Where("operator_id", 1).
		Count(&l)
	assert.Equal(t, l, 1)

	//test updating
	uc.Name = uc.Name + ",update"
	_, e = uc.Save()
	assert.Error(t, e, "cant update")

	uc.Name = uc.Name + ",updating"
	_, e = uc.Save()
	assert.Nil(t, e)
	//test updated
	DB.Table("logs").Where("logable_id", uc.ID).
		Where("remark", "user updated").
		Where("operator_id", 1).
		Count(&l)
	assert.Equal(t, l, 1)

	//test retrieve
	var uf UserEvent

	DB.Init(&uf)
	uf.WithContext(context.WithValue(context.Background(), "user_id", 1))
	_, e = DB.Model(&uf).Find(&uf, 1)
	assert.Error(t, e, "cant retrieve")

	uf.WithContext(context.WithValue(context.Background(), "user_id", 2))
	_, e = DB.Model(&uf).Find(&uf, 1)
	assert.Nil(t, e)

	//test retrieved
	DB.Table("logs").Where("logable_id", uf.ID).
		Where("remark", "user retrieved").
		Where("operator_id", 1).
		Count(&l)
	assert.Equal(t, l, 1)
	//test deleting

	uc.Name = uc.Name + ",delete"
	_, e = uc.Delete()
	assert.Error(t, e, "cant delete")

	uc.Name = uc.Name + ",deleting"
	_, e = uc.Delete()
	assert.Nil(t, e)
	//test deleted
	DB.Table("logs").Where("logable_id", uc.ID).
		Where("remark", "user deleted").
		Count(&l)
	assert.Equal(t, l, 1)

	var um UserEvent
	DB.Init(&um)
	um.Mute(goeloquent.EventALL)
	um.Name = "mute"
	_, e = um.Save()
	assert.Nil(t, e)
	assert.Equal(t, int64(2), um.ID)
	um.Name = "mute,update"
	_, e = um.Save()
	assert.Nil(t, e)
}
func TestTimeStamps(t *testing.T) {

	CreateUsers()
	//test timestamp is appended
	var user User
	DB.Init(&user)
	user.Name = "Alice"
	r, e := user.Save()
	assert.Nil(t, e)
	c, _ := r.RowsAffected()
	assert.Equal(t, int64(1), c)
	assert.True(t, strings.Contains(r.Sql, "insert into `user_models` ("))
	assert.Equal(t, len(r.Bindings), 3)
	assert.Contains(t, r.Bindings, "Alice", 1)
	user.Email = "asd"
	user.Status = 3
	r1, e1 := user.Save()
	assert.Nil(t, e1)
	c1, _ := r1.RowsAffected()
	assert.Equal(t, int64(1), c1)
	//assert.Equal(t, r1.Sql, "update `user_models` set `created_at` = ? , `id` = ? , `name` = ? , `status` = ? , `updated_at` = ? , `email` = ? where `id` = ?")
	assert.True(t, strings.Contains(r1.Sql, "update `user_models` set "))
	assert.Subset(t, r1.Bindings, []interface{}{int64(6), "Alice", uint8(3), "asd"})
	assert.Equal(t, len(r1.Bindings), 7)
	user.Age = 22
	r2, e2 := user.Only("age").Save()
	assert.Nil(t, e2)
	c2, _ := r2.RowsAffected()
	assert.Equal(t, int64(1), c2)
	assert.Contains(t, r2.Sql, "update `user_models` set ")
	assert.Subset(t, r2.Bindings, []interface{}{uint8(22), int64(6)})
	assert.Contains(t, r2.Sql, "update `user_models` set `", "`updated_at` = ?")
	//test timestamo is ignored when set it manually
	var u1, u2 User
	DB.Init(&u1)
	u1.Name = "Alice"
	u1.Age = 22
	u1.Email = "asd"
	u1.CreatedAt = sql.NullTime{
		Time:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local),
		Valid: true,
	}
	_, e = u1.Save()
	assert.Nil(t, e)
	_, e = DB.Model(&u2).Find(&u2, u1.ID)
	assert.Nil(t, e)
	assert.Equal(t, u1.CreatedAt.Time.Unix(), u2.CreatedAt.Time.Unix())
}
func TestSoftDeletes(t *testing.T) {

	var user User
	DB.Init(&user)
	r, e := DB.Model(&user).Find(&user, 1)
	assert.Nil(t, e)
	assert.Equal(t, r.Sql, "select * from `user_models` where `id` = ? and `user_models`.`deleted_at` is null")

	r1, e1 := DB.Model(&user).WithTrashed().Find(&user, 1)
	assert.Nil(t, e1)
	assert.Equal(t, r1.Sql, "select * from `user_models` where `id` = ?")

	r2, e2 := DB.Model(&user).OnlyTrashed().Find(&user, 1)
	assert.Nil(t, e2)
	assert.Equal(t, r2.Sql, "select * from `user_models` where `user_models`.`deleted_at` is not null and `id` = ?")

}

type UserDynamic struct {
	*goeloquent.EloquentModel
	ID    int    `goelo:"column:id;primaryKey"`
	Name  string `goelo:"column:name"`
	Age   int    `goelo:"column:age"`
	Email string `goelo:"column:email"`
}

func (u *UserDynamic) ResolveTableName(b *goeloquent.EloquentBuilder) string {
	idV := b.Context.Value("id")
	if id, ok := idV.(int); ok {
		return fmt.Sprintf("users_%d", id%10)
	} else {
		return "users_2024"
	}
}
func (u *UserDynamic) ResolveConnectionName(b *goeloquent.EloquentBuilder) string {
	idV := b.Context.Value("id")
	if _, ok := idV.(int); ok {
		return "default"
	} else {
		return "users_2024"
	}
}
func TestDynamicResolver(t *testing.T) {

	//test model dynamic table resolver
	//test model dynamic connection resolver

	assert.PanicsWithValuef(t, "Database connection users_2024 not configured.", func() {
		var u2 UserDynamic
		r, _ := DB.Model(&u2).Where("name", "a").First(&u2)
		assert.Equal(t, r.Sql, "select * from `users_2024` where `name` = ? limit 1")

	}, "Database connection users_2024 not configured.")
	var u2 UserDynamic
	r, _ := DB.Model(&u2).WithContext(context.WithValue(context.Background(), "id", 1)).Where("name", "a").First(&u2)
	assert.Equal(t, r.Sql, "select * from `users_1` where `name` = ? limit 1")

}

// TODO
func TestPivotMapping(t *testing.T) {

}
