package tests

import (
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

		var countUsers int64

		_, err = DB.Select("select count(*) from users", []interface{}{}, &countUsers)
		assert.Nil(t, err)
		assert.Equal(t, int64(3), countUsers)

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
