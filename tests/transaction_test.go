package tests

//func TestTransaction(t *testing.T) {
//
//	createUsers, dropUsers := UserTableSql()
//	now := time.Now()
//	u1 := map[string]interface{}{
//		"name":       "TestTransaction",
//		"age":        18,
//		"created_at": now,
//	}
//	DB.Raw("default").Exec(strings.ReplaceAll(dropUsers, `"`, "`"))
//	DB.Raw("default").Exec(strings.ReplaceAll(createUsers, `"`, "`"))
//
//	_, err := DB.Transaction(func(tx *goeloquent.Transaction) (interface{}, error) {
//		tx.Statement("set session transaction isolation level read uncommitted; ", nil)
//		_, errInsert := tx.Table("users").Insert(u1)
//		assert.Nil(t, errInsert)
//		var c int
//		tx.Table("users").Where("name", "TestTransaction").Count(&c)
//		assert.Equal(t, 1, c)
//		panic(errors.New("test roll back"))
//		return true, nil
//	})
//	assert.Equal(t, err.Error(), "test roll back")
//	var count int
//	_, err1 := DB.Table("users").Where("name", "TestTransaction").Count(&count)
//	assert.Nil(t, err1)
//	assert.Equal(t, count, 0)
//
//	_, err2 := DB.Transaction(func(tx *goeloquent.Transaction) (interface{}, error) {
//		tx.Statement("set session transaction isolation level read uncommitted; ", nil)
//		_, errInsert := tx.Table("users").Insert(u1)
//		assert.Nil(t, errInsert)
//		var id int
//		tx.Table("users").Where("id", 10/id).Update(map[string]interface{}{
//			"name": "new name",
//		})
//		return true, nil
//	})
//	assert.Equal(t, err2.Error(), "runtime error: integer divide by zero")
//	_, err = DB.Table("users").Where("name", "TestTransaction").OrWhere("name", "new name").Count(&count)
//	assert.Nil(t, err)
//	assert.Equal(t, count, 0)
//
//	_, err3 := DB.Transaction(func(tx *goeloquent.Transaction) (interface{}, error) {
//		tx.Statement("set session transaction isolation level read uncommitted; ", nil)
//		_, errInsert := tx.Table("users").Insert(u1)
//		assert.Nil(t, errInsert)
//		panic("string")
//		return true, nil
//	})
//	assert.Equal(t, err3.Error(), "error occurred during transaction")
//	_, err = DB.Table("users").Where("name", "TestTransaction").Count(&count)
//	assert.Nil(t, err)
//	assert.Equal(t, count, 0)
//
//	res, err4 := DB.Transaction(func(tx *goeloquent.Transaction) (interface{}, error) {
//		tx.Statement("set session transaction isolation level read uncommitted; ", nil)
//		_, errInsert := tx.Table("users").Insert(u1)
//		assert.Nil(t, errInsert)
//		return true, nil
//	})
//	assert.True(t, res.(bool))
//	_, err = DB.Table("users").Where("name", "TestTransaction").Count(&count)
//	assert.Nil(t, err4)
//	assert.Equal(t, count, 1)
//
//	DB.Raw("default").Exec(strings.ReplaceAll(dropUsers, `"`, "`"))
//}
