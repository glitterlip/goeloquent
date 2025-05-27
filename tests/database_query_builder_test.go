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

func TestBasicTableWrappingProtectsQuotationMarks(t *testing.T) {
	query := GetBuilder()
	query.From("`users`")
	assert.Nil(t, query.Statement.Error)
	assert.Equal(t, query.ToSql(), "select * from ```users```")

	query = GetBuilder()
	query.From("some`table")
	assert.Nil(t, query.Statement.Error)
	assert.Equal(t, query.ToSql(), "select * from `some``table`")
}
func TestAliasWrappingAsWholeConstant(t *testing.T) {
	query := GetBuilder()
	query.Select("x.y as foo.bar").From("baz")
	assert.Nil(t, query.Statement.Error)
	assert.Equal(t, query.ToSql(), "select `x`.`y` as `foo.bar` from `baz`")
}

func TestAliasWrappingWithSpacesInDatabaseName(t *testing.T) {
	b1 := GetBuilder()
	b1.Select("w x.y.z as foo.bar").From("baz")
	assert.Equal(t, "select `w x`.`y`.`z` as `foo.bar` from `baz`", b1.ToSql())
}

func TestAddingSelects(t *testing.T) {
	b := GetBuilder()
	b.Select("foo").AddSelect("bar").AddSelect("baz", "boom").AddSelect("bar").From("users")
	assert.Nil(t, b.Statement.Error)
	assert.Equal(t, "select `foo`, `bar`, `baz`, `boom`, `bar` from `users`", b.ToSql())
}

func TestBasicSelectWithPrefix(t *testing.T) {
	b := GetBuilder()
	b.Grammar.SetTablePrefix("prefix_")
	b.From("users").Select()
	assert.Equal(t, "select * from `prefix_users`", b.ToSql())
}

func TestBasicSelectDistinct(t *testing.T) {
	b := GetBuilder()
	b.Distinct().Select("foo", "bar").From("users")
	assert.Equal(t, "select distinct `foo`, `bar` from `users`", b.ToSql())
}
func TestBasicSelectDistinctOnColumns(t *testing.T) {
	b := GetBuilder()
	b.Distinct("foo").Select("foo", "bar").From("users")
	assert.Equal(t, "select distinct `foo`, `bar` from `users`", b.ToSql())
}
func TestBasicAlias(t *testing.T) {
	b := GetBuilder()
	b.Select("foo as bar").From("users")
	assert.Equal(t, "select `foo` as `bar` from `users`", b.ToSql())

}
func TestAliasWithPrefix(t *testing.T) {

	b1 := GetBuilder()
	b1.Grammar.SetTablePrefix("prefix_")
	b1.Select("foo as bar", "baz").From("users as people")
	assert.Equal(t, "select `foo` as `bar`, `baz` from `prefix_users` as `prefix_people`", b1.ToSql())
}

func TestJoinAliasesWithPrefix(t *testing.T) {
	b := GetBuilder()
	b.Grammar.SetTablePrefix("prefix_")
	b.Select("*").From("services").Join("translations AS t", "t.item_id", "=", "services.id")
	assert.Equal(t, "select * from `prefix_services` inner join `prefix_translations` as `prefix_t` on `prefix_t`.`item_id` = `prefix_services`.`id`", b.ToSql())

}

func TestBasicTableWrapping(t *testing.T) {
	b2 := GetBuilder()
	b2.Select().From("public.users")
	assert.Equal(t, "select * from `public`.`users`", b2.ToSql())
}

func TestWhenCallback(t *testing.T) {
	b := GetBuilder()
	cb := func(builder *goeloquent.QueryBuilder) {
		builder.Where("id", "=", 1)
	}
	b.Select("*").From("users").When(true, cb).Where("email", "foo")
	assert.Equal(t, "select * from `users` where `id` = ? and `email` = ?", b.ToSql())

	b1 := GetBuilder()
	b1.Select("*").From("users").When(false, cb).Where("email", "foo")
	assert.Equal(t, "select * from `users` where `email` = ?", b1.ToSql())
}
func TestWhenCallbackWithReturn(t *testing.T) {
	cb := func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		builder.Where("id", "=", 1)
		return builder
	}
	b := GetBuilder()
	b.Select("*").From("users").When(true, cb).Where("email", "foo")
	assert.Equal(t, "select * from `users` where `id` = ? and `email` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").When(false, cb).Where("email", "foo")
	assert.Equal(t, "select * from `users` where `email` = ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo"}, b1.GetBindings())
	assert.ElementsMatch(t, []interface{}{"foo"}, b1.GetRawBindings()["where"])

}
func TestWhenCallbackWithDefault(t *testing.T) {
	cb := func(builder *goeloquent.QueryBuilder) {
		builder.Where("id", "=", 1)
	}
	b2 := GetBuilder()
	b3 := GetBuilder()
	defaultCb := func(builder *goeloquent.QueryBuilder) {
		builder.Where("id", "=", 2)
	}
	b2.Select("*").From("users").When(false, cb, defaultCb).Where("email", "foo")
	assert.Equal(t, "select * from `users` where `id` = ? and `email` = ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{2, "foo"}, b2.GetBindings())
	assert.ElementsMatch(t, []interface{}{2, "foo"}, b2.GetRawBindings()["where"])

	b3.Select("*").From("users").When(true, cb, defaultCb).Where("email", "foo")
	assert.Equal(t, "select * from `users` where `id` = ? and `email` = ?", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b3.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b3.GetRawBindings()["where"])
}
func TestTapCallback(t *testing.T) {
	b := GetBuilder()
	cb := func(builder *goeloquent.QueryBuilder) {
		builder.Where("id", "=", 1)
	}

	b.Select("*").From("users").Tap(cb).Where("email", "foo")
	assert.Equal(t, "select * from `users` where `id` = ? and `email` = ?", b.ToSql())

}

func TestPipeCallback(t *testing.T) {
	b := GetBuilder()
	cb := func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		builder.Where("id", "=", 1)
		return builder
	}

	b.Select("*").From("users").Pipe(cb).Where("email", "foo")
	assert.Equal(t, "select * from `users` where `id` = ? and `email` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b.GetRawBindings()["where"])
}

func TestBasicWheres(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").Where("id", "=", 1)
	assert.Equal(t, "select * from `users` where `id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").Where("id", 1).Where("email", "=", "foo")
	assert.Equal(t, "select * from `users` where `id` = ? and `email` = ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b1.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b1.GetRawBindings()["where"])

	b2 := GetBuilder()
	b2.Select("*").From("users").Where("id", 1).Where("age", ">", 4, "or")
	assert.Equal(t, "select * from `users` where `id` = ? or `age` > ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 4}, b2.GetBindings())
}

func TestBasicWhereNot(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereNot("id", "=", 1)
	assert.Equal(t, "select * from `users` where not `id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").WhereNot("id", 1).WhereNot("email", "=", "foo", "or")
	assert.Equal(t, "select * from `users` where not `id` = ? or not `email` = ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b1.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b1.GetRawBindings()["where"])
}
