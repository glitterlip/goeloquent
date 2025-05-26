package tests

import (
	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GetBuilder() *goeloquent.QueryBuilder {
	stmt := goeloquent.NewStatement()
	query := goeloquent.NewQueryBuilder(stmt)
	return query
}
func TestBasicSelect(t *testing.T) {
	query := GetBuilder()
	query.From("users")
	assert.Equal(t, query.ToSql(), "select * from `users`")

	query = GetBuilder()
	query.Select("*").From("users")

	assert.Nil(t, query.Statement.Error)
	assert.Equal(t, query.ToSql(), "select * from `users`")

	query = GetBuilder()
	query.Select("id", "name").From("users")
	assert.Nil(t, query.Statement.Error)
	assert.Equal(t, query.ToSql(), "select `id`, `name` from `users`")

	query = GetBuilder()
	query.Select([]interface{}{"id", "name"}).From("users")
	assert.Nil(t, query.Statement.Error)
	assert.Equal(t, query.ToSql(), "select `id`, `name` from `users`")

}
func TestBasicSelectWithGetColumns(t *testing.T) {

	query := GetBuilder()
	users := map[string]interface{}{}
	query.From("users").Get(&users)
	assert.Nil(t, query.Statement.Error)
	assert.Equal(t, query.ToSql(), "select * from `users`")

	query = GetBuilder()
	query.From("users").Get(&users, "id", "name")
	assert.Nil(t, query.Statement.Error)
	assert.Equal(t, query.ToSql(), "select `id`, `name` from `users`")

	query = GetBuilder()
	query.From("users").Get(&users, []interface{}{"id", "name"})
	assert.Nil(t, query.Statement.Error)
	assert.Equal(t, query.ToSql(), "select `id`, `name` from `users`")

}
