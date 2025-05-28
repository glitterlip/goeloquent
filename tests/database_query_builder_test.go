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

//	func TestWheresWithArrayValue(t *testing.T) {
//		b := GetBuilder()
//		b.Select().From("users").Where("id", []interface{}{1, 2, 3})
//		assert.Equal(t, "select * from `users` where `id` = ?", b.ToSql())
//		assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
//
//		b1 := GetBuilder()
//		b1.Select().From("users").Where("id", "=", []interface{}{1, 2, 3}, goeloquent.And)
//		assert.Equal(t, "select * from `users` where `id` = ?", b1.ToSql())
//		assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())
//
//		b2 := GetBuilder()
//		b2.Select().From("users").Where("id", "!=", []interface{}{1, 2, 3}, goeloquent.Or)
//		assert.Equal(t, "select * from `users` where `id` != ?", b2.ToSql())
//		assert.ElementsMatch(t, []interface{}{1}, b2.GetBindings())
//
//		b3 := GetBuilder()
//		b3.Select().From("users").Where("id", "<>", []interface{}{1, 2, 3})
//		assert.Equal(t, "select * from `users` where `id` <> ?", b3.ToSql())
//		assert.ElementsMatch(t, []interface{}{1}, b3.GetBindings())
//	}
func TestDateBasedWheresAcceptsTwoArguments(t *testing.T) {

	b := GetBuilder()
	b.Select("*").From("users").WhereDate("created_at", "2023-01-01")
	assert.Equal(t, "select * from `users` where date(`created_at`) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"2023-01-01"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{"2023-01-01"}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").WhereDay("created_at", "2023-01-01")
	assert.Equal(t, "select * from `users` where day(`created_at`) = ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"2023-01-01"}, b1.GetBindings())
	assert.ElementsMatch(t, []interface{}{"2023-01-01"}, b1.GetRawBindings()["where"])

	b2 := GetBuilder()
	b2.Select("*").From("users").WhereMonth("created_at", 10)
	assert.Equal(t, "select * from `users` where month(`created_at`) = ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{10}, b2.GetBindings())
	assert.ElementsMatch(t, []interface{}{10}, b2.GetRawBindings()["where"])

	b3 := GetBuilder()
	b3.Select("*").From("users").WhereYear("created_at", 2022)
	assert.Equal(t, "select * from `users` where year(`created_at`) = ?", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{2022}, b3.GetBindings())
	assert.ElementsMatch(t, []interface{}{2022}, b3.GetRawBindings()["where"])

}

func TestDateBasedOrWheresAcceptsTwoArguments(t *testing.T) {

	b1 := GetBuilder()
	b1.Select().From("users").Where("id", 1).OrWhereDate("created_at", 1)
	assert.Equal(t, "select * from `users` where `id` = ? or date(`created_at`) = ?", b1.ToSql())
	b2 := GetBuilder()
	b2.Select().From("users").Where("id", 1).OrWhereDay("created_at", 1)
	assert.Equal(t, "select * from `users` where `id` = ? or day(`created_at`) = ?", b2.ToSql())
	b3 := GetBuilder()
	b3.Select().From("users").Where("id", 1).OrWhereMonth("created_at", 1)
	assert.Equal(t, "select * from `users` where `id` = ? or month(`created_at`) = ?", b3.ToSql())
	b4 := GetBuilder()
	b4.Select().From("users").Where("id", 1).OrWhereYear("created_at", 1)
	assert.Equal(t, "select * from `users` where `id` = ? or year(`created_at`) = ?", b4.ToSql())
}
func TestDateBasedWheresExpressionIsNotBound(t *testing.T) {
	b1 := GetBuilder()
	b1.Select().From("users").WhereDate("created_at", goeloquent.Raw("NOW()")).Where("age", ">", 18)
	assert.Equal(t, "select * from `users` where date(`created_at`) = NOW() and `age` > ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{18}, b1.GetBindings())
	assert.ElementsMatch(t, []interface{}{18}, b1.GetRawBindings()["where"])

	b2 := GetBuilder()
	b2.Select().From("users").WhereMonth("created_at", goeloquent.Raw("NOW()"))
	assert.Equal(t, "select * from `users` where month(`created_at`) = NOW()", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b2.GetBindings())

	b3 := GetBuilder()
	b3.Select().From("users").WhereYear("created_at", goeloquent.Raw("NOW()"))
	assert.Equal(t, "select * from `users` where year(`created_at`) = NOW()", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b2.GetRawBindings()["where"])

	b11 := GetBuilder()
	b11.Select().From("users").WhereDay("created_at", goeloquent.Raw("NOW()"))
	assert.Equal(t, "select * from `users` where day(`created_at`) = NOW()", b11.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b11.GetBindings())
}
func TestWhereDateMySql(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereDate("created_at", "=", "2023-01-01", goeloquent.Or)
	assert.Equal(t, "select * from `users` where date(`created_at`) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"2023-01-01"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{"2023-01-01"}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").WhereDay("updated_at", ">", "2025-01-01").WhereDate("created_at", "!=", "202-02-01", goeloquent.Or)
	assert.Equal(t, "select * from `users` where day(`updated_at`) > ? or date(`created_at`) != ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"2025-01-01", "202-02-01"}, b1.GetBindings())

}

func TestWhereDayMysql(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereDay("created_at", "=", "2023-01-01", goeloquent.Or)
	assert.Equal(t, "select * from `users` where day(`created_at`) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"2023-01-01"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{"2023-01-01"}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").WhereDay("updated_at", ">", "2025-01-01").WhereDay("created_at", "!=", "202-02-01", goeloquent.Or)
	assert.Equal(t, "select * from `users` where day(`updated_at`) > ? or day(`created_at`) != ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"2025-01-01", "202-02-01"}, b1.GetBindings())
}

func TestOrWhereDayMysql(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereDay("banned_at", 1).OrWhereDay("created_at", "2023-01-01")
	assert.Equal(t, "select * from `users` where day(`banned_at`) = ? or day(`created_at`) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "2023-01-01"}, b.GetBindings())
}

func TestWhereMonthMysql(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereMonth("created_at", 1, goeloquent.Or)
	assert.Equal(t, "select * from `users` where month(`created_at`) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").WhereMonth("updated_at", ">", 2).WhereMonth("created_at", "!=", 3, goeloquent.Or)
	assert.Equal(t, "select * from `users` where month(`updated_at`) > ? or month(`created_at`) != ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{2, 3}, b1.GetBindings())
}

func TestOrWhereMonthMysql(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereMonth("banned_at", 1).OrWhereMonth("created_at", 2)
	assert.Equal(t, "select * from `users` where month(`banned_at`) = ? or month(`created_at`) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2}, b.GetBindings())
}

func TestWhereYearMysql(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereYear("created_at", 2023, goeloquent.Or)
	assert.Equal(t, "select * from `users` where year(`created_at`) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{2023}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{2023}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").WhereYear("updated_at", ">", 2022).WhereYear("created_at", "!=", 2021, goeloquent.Or)
	assert.Equal(t, "select * from `users` where year(`updated_at`) > ? or year(`created_at`) != ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{2022, 2021}, b1.GetBindings())
}

func TestOrWhereYearMysql(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereYear("banned_at", 2023).OrWhereYear("created_at", 2024)
	assert.Equal(t, "select * furom `users` where year(`banned_at`) = ? or year(`created_at`) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{2023, 2024}, b.GetBindings())
}
func TestWhereTimeMysql(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereTime("created_at", "12:00", goeloquent.Or)
	assert.Equal(t, "select * from `users` where time(`created_at`) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"12:00"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{"12:00"}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").WhereTime("updated_at", ">", "13:00:00").WhereTime("created_at", "!=", "14:00:00", goeloquent.Or)
	assert.Equal(t, "select * from `users` where time(`updated_at`) > ? or time(`created_at`) != ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"13:00:00", "14:00:00"}, b1.GetBindings())
}

func TestOrWhereTimeMysql(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereTime("banned_at", "12:00").OrWhereTime("created_at", "<=", "13:00")
	assert.Equal(t, "select * from `users` where time(`banned_at`) = ? or time(`created_at`) <= ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"12:00", "13:00"}, b.GetBindings())
}

// testWherePast
// testWherePastUsesArray
// testWhereTodayMySQL
// testPassingArrayToWhereTodayMySQL
// testWhereFuture
// testPassingArrayToWhereFuture
