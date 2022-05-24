package tests

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRawMethods(t *testing.T) {
	var row = make(map[string]interface{})
	createUsers, dropUsers := UserTableSql()
	RunWithDB(createUsers, dropUsers, func() {
		//test raw
		exec, err := DB.Raw("default").Exec("insert into `users` (`name`,`age`) values ('Alice',33)")
		assert.Nil(t, err)
		c, _ := exec.RowsAffected()
		assert.Equal(t, int64(1), c)

		//test insert
		insert, err := DB.Insert("insert into `users` (`name`,`age`,`status`)  values (?,?,?),(?,?,?)", []interface{}{"John", 12, 1, "Joe", 22, 1})
		assert.Nil(t, err)
		c, _ = insert.RowsAffected()
		assert.Equal(t, int64(2), c)

		//test select
		r, err := DB.Select("select * from users limit ? ", []interface{}{1}, &row)
		c, err = r.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, "Alice", string(row["name"].([]uint8)))
		assert.Equal(t, int64(1), c)

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

		c = 0
		userC, err := DB.Select("select count(1) from users", []interface{}{}, &c)

		assert.Nil(t, err)
		count, err := userC.RowsAffected()
		assert.Equal(t, int64(1), count)
		assert.Equal(t, int64(2), c)

	})

}
