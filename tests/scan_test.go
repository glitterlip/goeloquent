package tests

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

type UserScan struct {
	*goeloquent.EloquentModel
	Id    int64       `goelo:"column:id"`
	Name  string      `goelo:"column:name"`
	Age   int         `goelo:"column:age"`
	Info  JsonStruct  `goelo:"column:info"`
	Roles CommaString `goelo:"column:roles"`
}

func (u *UserScan) TableName() string {
	return "users"
}

type CommaString struct {
	Strs []string
}

func (c *CommaString) Scan(src any) error {
	str, ok := src.([]byte)
	if !ok {
		fmt.Printf("%#v", src)
		return nil
	} else {
		c.Strs = strings.Split(string(str), ",")
		return nil
	}
}
func (c CommaString) Value() (driver.Value, error) {
	return strings.Join(c.Strs, ","), nil
}

type JsonStruct struct {
	Name    string
	Deleted bool
	Age     int
	Roles   []string
}

func (j JsonStruct) Value() (driver.Value, error) {
	return json.Marshal(j)
}
func (j *JsonStruct) Scan(src any) error {
	str, ok := src.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(str, j)

}
func TestBasicScan(t *testing.T) {
	createUsers, dropUsers := UserTableSql()
	RunWithDB(createUsers, dropUsers, func() {
		DB.Raw("default").Exec("insert into `users` (`name`,`age`) values ('Alice',33)")
		DB.Insert("insert into `users` (`name`,`age`,`status`)  values (?,?,?),(?,?,?)", []interface{}{"John", 12, 1, "Joe", 22, 1})
		//test scan map
		var row = make(map[string]interface{})
		r, err := DB.Select("select * from users where name = ? ", []interface{}{"Alice"}, &row)
		assert.Nil(t, err)
		assert.Equal(t, "Alice", string(row["name"].([]uint8)))

		//test scan slice of map
		var rows []map[string]interface{}
		r, err = DB.Select("select * from users ", nil, &rows)
		assert.Nil(t, err)
		assert.Equal(t, int64(3), r.Count)
		//test scan plain struct
		type User struct {
			Id   int64  `goelo:"column:id"`
			Name string `goelo:"column:name"`
			Age  int    `goelo:"column:age"`
		}
		type MappingUser struct {
			Id       int64  `goelo:"column:id"`
			UserName string `goelo:"column:name"`
			Age      int    `goelo:"column:age"`
		}
		var user User
		var mapUser MappingUser
		res, err := DB.Select("select * from users order by id asc limit ?", []interface{}{1}, &user)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), res.Count)
		assert.Equal(t, int64(1), user.Id)
		assert.Equal(t, 33, user.Age)
		assert.Equal(t, "Alice", user.Name)

		res, err = DB.Select("select * from users order by id asc limit ? offset 1", []interface{}{1}, &mapUser)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), res.Count)
		assert.Equal(t, int64(2), mapUser.Id)
		assert.Equal(t, 12, mapUser.Age)
		assert.Equal(t, "John", mapUser.UserName)

		//select slice of struct
		var us []UserScan
		res, err = DB.Select("select * from users order by id asc limit ?", []interface{}{2}, &us)
		assert.Nil(t, err)
		assert.Equal(t, int64(2), res.Count)
		assert.Equal(t, int64(1), us[0].Id)
		assert.Equal(t, int64(2), us[1].Id)
		assert.Equal(t, 33, us[0].Age)
		assert.Equal(t, 12, us[1].Age)
		assert.Equal(t, "Alice", us[0].Name)
		assert.Equal(t, "John", us[1].Name)
	})
}

func TestScanInterface(t *testing.T) {
	createUsers, dropUsers := UserTableSql()
	RunWithDB(createUsers, dropUsers, func() {

		var i = JsonStruct{
			Name:    "Test",
			Deleted: true,
			Age:     10,
			Roles:   []string{"premium", "user"},
		}
		user := &UserScan{
			Name:  "Test",
			Age:   10,
			Info:  i,
			Roles: CommaString{Strs: []string{"premium", "user"}},
		}

		_, ok := reflect.ValueOf(i).Interface().(driver.Valuer)
		assert.True(t, ok)

		DB.Init(user)
		r, err := user.Save()
		assert.Nil(t, err)
		affected, err := r.RowsAffected()
		assert.Equal(t, int64(1), affected)

		var created UserScan
		r, err = DB.Model().First(&created)
		assert.Nil(t, err)
		assert.Equal(t, created.Info, i)
		assert.ElementsMatch(t, created.Roles.Strs, []string{"premium", "user"})

		var us []UserScan
		r, err = DB.Model().Get(&us)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), r.Count)
		assert.Equal(t, us[0].Info, i)
		assert.ElementsMatch(t, us[0].Roles.Strs, []string{"premium", "user"})
	})
}
