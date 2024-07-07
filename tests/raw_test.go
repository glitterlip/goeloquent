package tests

import (
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRawMethods(t *testing.T) {
	createUsers, dropUsers := UserTableSql()
	RunWithDB(createUsers, dropUsers, func() {

		//test raw
		exec, err := DB.Raw("default").Exec("insert into `users` (`name`,`age`) values ('Alice',33)")
		assert.Nil(t, err)
		count, _ := exec.RowsAffected()
		assert.Equal(t, int64(1), count)

		//test insert
		insert, err := DB.Insert("insert into `users` (`name`,`age`,`status`)  values (?,?,?),(?,?,?)", []interface{}{"John", 12, 1, "Joe", 22, 1})
		assert.Nil(t, err)
		count, _ = insert.RowsAffected()
		assert.Equal(t, int64(2), count)

		//test select map
		var row = make(map[string]interface{})
		r, err := DB.Select("select * from users where name = ? ", []interface{}{"Alice"}, &row)
		assert.Nil(t, err)
		assert.Equal(t, "Alice", string(row["name"].([]uint8)))

		//test select slice of map
		var rows []map[string]interface{}
		r, err = DB.Select("select * from users ", nil, &rows)
		assert.Nil(t, err)
		assert.Equal(t, int64(3), r.Count)
		goeloquent.ParsedModelsMap.Delete("github.com/glitterlip/goeloquent/tests.User")
		//test select struct
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
		var us []User
		res, err = DB.Select("select * from users order by id asc limit ?", []interface{}{2}, &us)
		assert.Nil(t, err)
		assert.Equal(t, int64(2), res.Count)
		assert.Equal(t, int64(1), us[0].Id)
		assert.Equal(t, int64(2), us[1].Id)
		assert.Equal(t, 33, us[0].Age)
		assert.Equal(t, 12, us[1].Age)
		assert.Equal(t, "Alice", us[0].Name)
		assert.Equal(t, "John", us[1].Name)

		//test update
		u, err := DB.Update("update users set age = 18 where age < 18 ", []interface{}{})
		assert.Nil(t, err)
		updated, _ := u.RowsAffected()
		assert.Equal(t, int64(1), updated)

		//test delete
		d, err := DB.Delete("delete from users where age = 18", []interface{}{})
		assert.Nil(t, err)
		deleted, _ := d.RowsAffected()
		assert.Equal(t, int64(1), deleted)

		var c int64
		userC, err := DB.Select("select count(1) from users", []interface{}{}, &c)

		assert.Nil(t, err)
		assert.Equal(t, int64(1), userC.Count)
		assert.Equal(t, int64(2), c)

	})

}
