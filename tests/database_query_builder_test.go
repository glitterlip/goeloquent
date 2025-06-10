package tests

import (
	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GetBuilder() *goeloquent.QueryBuilder {
	stmt := goeloquent.NewStatement()
	query := goeloquent.NewQueryBuilder(stmt)
	return query.Pretend()
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

func TestMySqlWrappingProtectsQuotationMarks(t *testing.T) {

}
func TestOrWhereDayPostgres(t *testing.T) {

}
func testOrWhereDaySqlServer(t *testing.T) {

}
func testOrWhereMonthSqlServer(t *testing.T) {

}
func testOrWhereYearPostgres(t *testing.T) {

}
func testOrWhereYearSqlServer(t *testing.T) {

}
func testWhereTimeOperatorOptionalPostgres(t *testing.T) {

}
func testWhereTimeSqlServer(t *testing.T) {

}
func testOrWhereTimePostgres(t *testing.T) {

}
func testOrWhereTimeSqlServer(t *testing.T) {

}
func testWhereDatePostgres(t *testing.T) {

}
func testWhereDayPostgres(t *testing.T) {

}
func testWhereMonthPostgres(t *testing.T) {

}
func testWhereYearPostgres(t *testing.T) {

}
func testWhereTimePostgres(t *testing.T) {

}
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
	assert.Equal(t, "select * from `users` where year(`banned_at`) = ? or year(`created_at`) = ?", b.ToSql())
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

func testWherePast(t testing.T) {

}
func testWherePastUsesArray(t *testing.T) {

}
func testWhereTodayMySQL(t *testing.T) {

}
func testPassingArrayToWhereTodayMySQL(t *testing.T) {

}
func testWhereFuture(t *testing.T) {

}
func testPassingArrayToWhereFuture(t *testing.T) {

}
func testWhereLikePostgres(t *testing.T) {

}
func testWhereLikeClausePostgres(t *testing.T) {

}

func TestWhereLikeClauseMysql(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereLike("name", "1").WhereNotLike("name", "Jim")
	assert.Equal(t, "select * from `users` where `name` like ? and `name` not like ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"1", "Jim"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{"1", "Jim"}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").WhereLike("name", "john").CaseSensitive()
	assert.Equal(t, "select * from `users` where `name` like binary ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"john"}, b1.GetBindings())
	assert.ElementsMatch(t, []interface{}{"john"}, b1.GetRawBindings()["where"])

	b2 := GetBuilder()
	b2.Select("*").From("users").WhereNotLike("name", "john").OrWhereLike("name", "john")
	assert.Equal(t, "select * from `users` where `name` not like ? or `name` like ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{"john", "john"}, b2.GetBindings())
	assert.ElementsMatch(t, []interface{}{"john", "john"}, b2.GetRawBindings()["where"])

	b3 := GetBuilder()
	b3.Select("*").From("users").WhereNotLike("name", "john").CaseSensitive().OrWhereNotLike("name", "john")
	assert.Equal(t, "select * from `users` where `name` not like binary ? or `name` not like ?", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{"john", "john"}, b3.GetBindings())

}

func testWhereLikeClauseSqlite(t *testing.T) {

}
func testWhereLikeClauseSqlServer(t *testing.T) {

}
func testWhereDateSqlite(t *testing.T) {

}
func testWhereDaySqlite(t *testing.T) {

}
func testWhereMonthSqlite(t *testing.T) {

}
func testWhereYearSqlite(t *testing.T) {

}
func testWhereTimeSqlite(t *testing.T) {

}
func testWhereTimeOperatorOptionalSqlite(t *testing.T) {

}
func testWhereDateSqlServer(t *testing.T) {

}
func testWhereDaySqlServer(t *testing.T) {

}
func testWhereMonthSqlServer(t *testing.T) {

}
func testWhereYearSqlServer(t *testing.T) {

}

func TestWhereBetweens(t *testing.T) {

	b := GetBuilder()
	b.Select("*").From("users").WhereBetween("id", []interface{}{1, 10})
	assert.Equal(t, "select * from `users` where `id` between ? and ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 10}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, 10}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").WhereNotBetween("id", []interface{}{1, 10, 100})
	assert.Equal(t, "select * from `users` where `id` not between ? and ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 10}, b1.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, 10}, b1.GetRawBindings()["where"])

	b2 := GetBuilder()
	b2.Select("*").From("users").WhereBetween("id", []interface{}{goeloquent.Raw("1"), goeloquent.Raw("10")})
	assert.Equal(t, "select * from `users` where `id` between 1 and 10", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b2.GetBindings())
}

func TestOrWhereBetween(t *testing.T) {

	b := GetBuilder()
	b.Select("*").From("users").Where("id", 1).OrWhereBetween("id", []interface{}{1, 10})
	assert.Equal(t, "select * from `users` where `id` = ? or `id` between ? and ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 1, 10}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, 1, 10}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").Where("id", 1).OrWhereNotBetween("id", []interface{}{1, 10, 100})
	assert.Nil(t, b1.Statement.Error)
	assert.Equal(t, "select * from `users` where `id` = ? or `id` not between ? and ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 1, 10}, b1.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, 1, 10}, b1.GetRawBindings()["where"])

	b2 := GetBuilder()
	b2.Select("*").From("users").Where("id", 1).OrWhereBetween("id", []interface{}{goeloquent.Raw("1"), goeloquent.Raw("10")})
	assert.Equal(t, "select * from `users` where `id` = ? or `id` between 1 and 10", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b2.GetBindings())
}

func TestOrWhereNotBetween(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").Where("id", 1).OrWhereNotBetween("id", []interface{}{1, 10})
	assert.Equal(t, "select * from `users` where `id` = ? or `id` not between ? and ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 1, 10}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, 1, 10}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").Where("id", 1).OrWhereNotBetween("id", []interface{}{goeloquent.Raw("1"), goeloquent.Raw("10")})
	assert.Equal(t, "select * from `users` where `id` = ? or `id` not between 1 and 10", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())
}

func TestWhereBetweenColumns(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereBetweenColumns("id", []interface{}{"users.created_at", "users.updated_at"})
	assert.Equal(t, "select * from `users` where `id` between `users`.`created_at` and `users`.`updated_at`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"users.created_at", "users.updated_at"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{"users.created_at", "users.updated_at"}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").WhereNotBetweenColumns("id", []interface{}{"users.created_at", "users.updated_at"})
	assert.Equal(t, "select * from `users` where `id` not between `users`.`created_at` and `users`.`updated_at`", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"users.created_at", "users.updated_at"}, b1.GetBindings())
	assert.ElementsMatch(t, []interface{}{"users.created_at", "users.updated_at"}, b1.GetRawBindings()["where"])

	b2 := GetBuilder()
	b2.Select("*").From("users").WhereBetweenColumns("id", []interface{}{goeloquent.Raw("users.created_at"), goeloquent.Raw("users.updated_at")})
	assert.Equal(t, "select * from `users` where `id` between users.created_at and users.updated_at", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b2.GetBindings())
}

func TestOrWhereBetweenColumns(t *testing.T) {

	b := GetBuilder()
	b.Select("*").From("users").Where("id", 1).OrWhereBetweenColumns("id", []interface{}{"users.created_at", "users.updated_at"})
	assert.Equal(t, "select * from `users` where `id` = ? or `id` between `users`.`created_at` and `users`.`updated_at`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "users.created_at", "users.updated_at"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, "users.created_at", "users.updated_at"}, b.GetRawBindings()["where"])

	b2 := GetBuilder()
	b2.Select("*").From("users").Where("id", 1).OrWhereBetweenColumns("id", []interface{}{goeloquent.Raw("users.created_at"), goeloquent.Raw("users.updated_at")})
	assert.Equal(t, "select * from `users` where `id` = ? or `id` between users.created_at and users.updated_at", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b2.GetBindings())
}

func TestOrWhereNotBetweenColumns(t *testing.T) {
	b1 := GetBuilder()
	b1.Select("*").From("users").Where("id", 1).OrWhereNotBetweenColumns("id", []interface{}{"users.created_at", "users.updated_at"})
	assert.Equal(t, "select * from `users` where `id` = ? or `id` not between `users`.`created_at` and `users`.`updated_at`", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "users.created_at", "users.updated_at"}, b1.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, "users.created_at", "users.updated_at"}, b1.GetRawBindings()["where"])

	b2 := GetBuilder()
	b2.Select("*").From("users").Where("id", 1).OrWhereNotBetweenColumns("id", []interface{}{goeloquent.Raw("users.created_at"), goeloquent.Raw("users.updated_at")})
	assert.Equal(t, "select * from `users` where `id` = ? or `id` not between users.created_at and users.updated_at", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b2.GetBindings())
}

func TestBasicOrWheres(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").Where("id", 1).OrWhere("email", "=", "foo").OrWhere("name", "bar")
	assert.Equal(t, "select * from `users` where `id` = ? or `email` = ? or `name` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo", "bar"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, "foo", "bar"}, b.GetRawBindings()["where"])
}

func TestBasicOrWhereNot(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").Where("id", 1).OrWhereNot("email", "<>", "foo").OrWhereNot("name", "bar")
	assert.Equal(t, "select * from `users` where `id` = ? or not `email` <> ? or not `name` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo", "bar"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, "foo", "bar"}, b.GetRawBindings()["where"])
}

func TestRawWheres(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereRaw("id = ?", []interface{}{1})
	assert.Equal(t, "select * from `users` where id = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").WhereRaw(goeloquent.Raw("id = 1 or email = `ad`"), []interface{}{})
	assert.Equal(t, "select * from `users` where id = 1 or email = `ad`", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())
}

func TestRawOrWheres(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").Where("id", 1).OrWhereRaw("email = ?", []interface{}{"foo"}).OrWhereRaw("name = ?", []interface{}{"bar"})
	assert.Equal(t, "select * from `users` where `id` = ? or email = ? or name = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo", "bar"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, "foo", "bar"}, b.GetRawBindings()["where"])

}

func TestBasicWhereIns(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereIn("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `id` in (?, ?, ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2, 3}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("name", "Joe").OrWhereIn("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `name` = ? or `id` in (?, ?, ?)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"Joe", 1, 2, 3}, b1.GetBindings())

}

func TestBasicWhereNotIns(t *testing.T) {

	b := GetBuilder()
	b.Select().From("users").WhereNotIn("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `id` not in (?, ?, ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2, 3}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("name", "Joe").OrWhereNotIn("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `name` = ? or `id` not in (?, ?, ?)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"Joe", 1, 2, 3}, b1.GetBindings())
}

func TestRawWhereIns(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereIn("id", []interface{}{goeloquent.Raw("1")})
	assert.Equal(t, "select * from `users` where `id` in (1)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("name", "Amber").OrWhereIn("id", []interface{}{goeloquent.Raw("1")})
	assert.Equal(t, "select * from `users` where `name` = ? or `id` in (1)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"Amber"}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").WhereIn("id", []interface{}{goeloquent.Raw("select id from users where email like '%@gmail.com'")})
	assert.Equal(t, "select * from `users` where `id` in (select id from users where email like '%@gmail.com')", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b2.GetBindings())
}

func TestEmptyWhereIns(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereIn("id", []interface{}{})
	assert.Equal(t, "select * from `users` where 0 = 1", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("name", "a").OrWhereIn("id", []interface{}{})
	assert.Equal(t, "select * from `users` where `name` = ? or 0 = 1", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"a"}, b1.GetBindings())
}

// testWhereIntegerInRaw
func TestEmptyWhereNotIns(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereNotIn("id", []interface{}{})
	assert.Equal(t, "select * from `users` where 1 = 1", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("name", "a").OrWhereNotIn("id", []interface{}{})
	assert.Equal(t, "select * from `users` where `name` = ? or 1 = 1", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"a"}, b1.GetBindings())
}

func testWhereIntegerInRaw(t testing.T) {

}
func testOrWhereIntegerInRaw(t testing.T) {

}
func testWhereIntegerNotInRaw(t testing.T) {

}
func testOrWhereIntegerNotInRaw(t testing.T) {

}
func testEmptyWhereIntegerInRaw(t *testing.T) {

}
func testEmptyWhereIntegerNotInRaw(t *testing.T) {

}
func TestBasicWhereColumn(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereColumn("first_name", "last_name").OrWhereColumn("first_name", "middle_name")
	assert.Equal(t, "select * from `users` where `first_name` = `last_name` or `first_name` = `middle_name`", b.ToSql())
	assert.Equal(t, 0, len(b.GetBindings()))

	b1 := GetBuilder()
	b1.Select().From("users").Where("id", 1).WhereColumn("updated_at", ">", "created_at")
	assert.Equal(t, "select * from `users` where `id` = ? and `updated_at` > `created_at`", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())
}

func TestArrayWhereColumn(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereColumn([][]interface{}{{"username", "nickname"}})
	assert.Equal(t, "select * from `users` where (`username` = `nickname`)", b.ToSql())
	assert.Equal(t, 0, len(b.GetBindings()))

	b1 := GetBuilder()
	b1.Select().From("users").Where("id", 1).OrWhereColumn([][]interface{}{{"first_name", "last_name"}, {"created_at", ">", "updated_at"}})
	assert.Equal(t, "select * from `users` where `id` = ? or (`first_name` = `last_name` or `created_at` > `updated_at`)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())

}

func TestWhereFulltextMySql(t *testing.T) {
	b := GetBuilder()
	b.Select("*").From("users").WhereFullText("name", "foo bar")
	assert.Equal(t, "select * from `users` where match (`name`) against (? in natural language mode)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo bar"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{"foo bar"}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select("*").From("users").WhereFullText("name", "foo bar").FulltextMode("boolean")
	assert.Equal(t, "select * from `users` where match (`name`) against (? in boolean mode)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo bar"}, b1.GetBindings())
	b1.GetRawBindings()["where"] = []interface{}{"foo bar"}

	b2 := GetBuilder()
	b2.Select("*").From("users").WhereFullText("name", "+Hello -World").Expand().FulltextMode("boolean")
	assert.Equal(t, "select * from `users` where match (`name`) against (? in boolean mode with query expansion)", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{"+Hello -World"}, b2.GetBindings())
	b2.GetRawBindings()["where"] = []interface{}{"+Hello -World"}
}

func TestWhereFulltextPostgres(t *testing.T) {

}

func TestWhereAll(t *testing.T) {

}

func TestOrWhereAll(t *testing.T) {

}

func TestWhereAny(t *testing.T) {

}

func TestOrWhereAny(t *testing.T) {

}

func TestWhereNone(t *testing.T) {

}

func TestOrWhereNone(t *testing.T) {

}

func TestUnions(t *testing.T) {

}

func TestUnionAlls(t *testing.T) {

}

func TestMultipleUnions(t *testing.T) {

}

func TestMultipleUnionAlls(t *testing.T) {

}
func TestUnionOrderBys(t *testing.T) {

}

func TestUnionLimitsAndOffsets(t *testing.T) {

}

func TestUnionWithJoin(t *testing.T) {

}

func TestMySqlUnionOrderBys(t *testing.T) {

}
func TestMySqlUnionLimitsAndOffsets(t *testing.T) {

}
func TestUnionAggregate(t *testing.T) {

}
func TestHavingAggregate(t *testing.T) {

	b := GetBuilder()
	b.From("posts").SelectSub(func(builder *goeloquent.QueryBuilder) {
		builder.From("videos").Select("count(*)").WhereColumn("posts.id", "=", "videos.post_id")
	}, "videos_count").Having("videos_count", ">", 10)
	var count int
	_, err := b.Pretend().Count(&count)
	assert.Nil(t, err)
	assert.Equal(t, "select count(*) as aggregate from (select (select `count(*)` from `videos` where `posts`.`id` = `videos`.`post_id`) as `videos_count` from `posts` having `videos_count` > ?) as `temp_table`", b.ToSql())

}

func TestSubSelectWhereIns(t *testing.T) {
	b := GetBuilder()
	b.From("users").WhereIn("id", func(builder *goeloquent.QueryBuilder) {
		builder.Select("id").From("users").Where("age", ">", 25).Take(3)
	})
	assert.Equal(t, "select * from `users` where `id` in (select `id` from `users` where `age` > ? limit 3)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{25}, b.GetBindings())

	b1 := GetBuilder()
	b1.From("users").WhereNotIn("id", func(builder *goeloquent.QueryBuilder) {
		builder.Select("id").From("users").Where("age", ">", 25).Take(5)
	})
	assert.Equal(t, "select * from `users` where `id` not in (select `id` from `users` where `age` > ? limit 5)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{25}, b1.GetBindings())

}

func TestBasicWhereNulls(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereNull("email")
	assert.Equal(t, "select * from `users` where `email` is null", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
	assert.Nil(t, b.Statement.Error)

	b1 := GetBuilder()
	b1.Select().From("users").Where("id", 1).OrWhereNull("email")
	assert.Equal(t, "select * from `users` where `id` = ? or `email` is null", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())
}

func TestBasicWhereNullExpressionsMysql(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereNull(goeloquent.Raw("email"))
	assert.Equal(t, "select * from `users` where email is null", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
	assert.Nil(t, b.Statement.Error)

	b1 := GetBuilder()
	b1.Select().From("users").Where("id", 1).OrWhereNull(goeloquent.Raw("email"))
	assert.Equal(t, "select * from `users` where `id` = ? or email is null", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())
}
func TestJsonWhereNullMysql(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereNull("items->id")
	assert.Equal(t, "select * from `users` where (json_extract(`items`, '$.\"id\"') is null OR json_type(json_extract(`items`, '$.\"id\"')) = 'NULL')", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}
func TestJsonWhereNotNullMysql(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereNotNull("items->id")
	assert.Equal(t, "select * from `users` where (json_extract(`items`, '$.\"id\"') is not null AND json_type(json_extract(`items`, '$.\"id\"')) != 'NULL')", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}
func TestJsonWhereNullExpressionMysql(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereNull(goeloquent.Raw("items->id"))
	assert.Equal(t, "select * from `users` where (json_extract(`items`, '$.\"id\"') is null OR json_type(json_extract(`items`, '$.\"id\"')) = 'NULL')", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}

func TestJsonWhereNotNullExpressionMysql(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereNotNull(goeloquent.Raw("items->id"))
	assert.Equal(t, "select * from `users` where (json_extract(`items`, '$.\"id\"') is not null AND json_type(json_extract(`items`, '$.\"id\"')) != 'NULL')", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}
func testArrayWhereNulls(t *testing.T) {
}
func TestBasicWhereNotNulls(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereNotNull("email")
	assert.Equal(t, "select * from `users` where `email` is not null", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
	assert.Nil(t, b.Statement.Error)

	b1 := GetBuilder()
	b1.Select().From("users").Where("id", 1).OrWhereNotNull("email")
	assert.Equal(t, "select * from `users` where `id` = ? or `email` is not null", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())
}
func testArrayWhereNotNulls(t *testing.T) {
}

func TestGroupBys(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").GroupBy("name")
	assert.Equal(t, "select * from `users` group by `name`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").GroupBy("name", "email")
	assert.Equal(t, "select * from `users` group by `name`, `email`", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").GroupBy([]interface{}{"name", "email"})
	assert.Equal(t, "select * from `users` group by `name`, `email`", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b2.GetBindings())

	b3 := GetBuilder()
	b3.Select().From("users").GroupBy([]string{"name", "email"})
	assert.Equal(t, "select * from `users` group by `name`, `email`", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b3.GetBindings())

	b4 := GetBuilder()
	b4.Select().From("users").GroupBy(goeloquent.Raw("DATE(created_at)"))
	assert.Equal(t, "select * from `users` group by DATE(created_at)", b4.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b4.GetBindings())

	b5 := GetBuilder()
	b5.Select().From("users").GroupByRaw("DATE(created_at), ? DESC", []interface{}{"name"})
	assert.Equal(t, "select * from `users` group by DATE(created_at), ? DESC", b5.ToSql())
	assert.ElementsMatch(t, []interface{}{"name"}, b5.GetBindings())

	b6 := GetBuilder()
	b6.Select().From("users").HavingRaw("?", []interface{}{"havingRawBinding"}).GroupByRaw("?", []interface{}{"groupByRawBinding"}).
		WhereRaw("?", []interface{}{"whereRawBinding"}).OrderByRaw("?", []interface{}{"orderByRawBinding"})
	assert.Equal(t, "select * from `users` where ? group by ? having ? order by ?", b6.ToSql())
	assert.ElementsMatch(t, []interface{}{"whereRawBinding", "groupByRawBinding", "havingRawBinding", "orderByRawBinding"}, b6.GetBindings())

}

func TestOrderBys(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").OrderBy("name").OrderBy("age", "desc")

	assert.Equal(t, "select * from `users` order by `name` asc, `age` desc", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b.Without([]goeloquent.Component{goeloquent.COMPONENT_ORDER}, []goeloquent.Component{goeloquent.COMPONENT_ORDER})
	assert.Equal(t, "select * from `users`", b.ToSql())

	b1 := GetBuilder()
	b1.Select().From("users").OrderBy("name", "asc").OrderByRaw("'age' ? desc", []interface{}{"age1"})
	assert.Equal(t, "select * from `users` order by `name` asc, 'age' ? desc", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"age1"}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").OrderByDesc("age")
	assert.Equal(t, "select * from `users` order by `age` desc", b2.ToSql())

}

func testLatest(t *testing.T) {

}

func testOldest(t *testing.T) {

}

func TestInRandowOrderMysql(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").InRandomOrder()
	assert.Equal(t, "select * from `users` order by RAND()", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

}

func testInRandomOrderPostgres(t *testing.T) {

}

func testInRandomOrderSqlServer(t *testing.T) {}
func testOrderBysSqlServer(t testing.T)       {}
func TestRecorder(t *testing.T) {
	// Test the recorder functionality
	b := GetBuilder()
	b.Select().From("users").Where("id", 1).OrderBy("name")
	assert.Equal(t, "select * from `users` where `id` = ? order by `name` asc", b.ToSql())
	b.ReOrder()
	assert.Equal(t, "select * from `users` where `id` = ?", b.ToSql())

	b1 := GetBuilder()
	b1.Select().From("users").Where("id", 1).OrderBy("name")
	assert.Equal(t, "select * from `users` where `id` = ? order by `name` asc", b1.ToSql())
	b1.ReOrder("age", "desc")
	assert.Equal(t, "select * from `users` where `id` = ? order by `age` desc", b1.ToSql())

	b2 := GetBuilder()
	b2.Select().From("users").Where("id", 1).OrderByRaw("'name' asc ? desc", []interface{}{"name1"})
	assert.Equal(t, "select * from `users` where `id` = ? order by 'name' asc ? desc", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "name1"}, b2.GetBindings())
}

func TestOrderBySubQueries(t *testing.T) {
	b := GetBuilder()
	sub := func(builder *goeloquent.QueryBuilder) {
		builder.Select("created_at").From("logins").WhereColumn("user_id", "users.id").Limit(1)
	}
	b.Select().From("users").OrderBy(sub)
	assert.Equal(t, "select * from `users` order by (select `created_at` from `logins` where `user_id` = `users`.`id` limit 1) asc", b.ToSql())

	b.Without([]goeloquent.Component{goeloquent.COMPONENT_ORDER}, []goeloquent.Component{goeloquent.COMPONENT_ORDER})
	b.OrderBy(sub, "desc")
	assert.Equal(t, "select * from `users` order by (select `created_at` from `logins` where `user_id` = `users`.`id` limit 1) desc", b.ToSql())

	b1 := GetBuilder()
	b1.Select().From("users").OrderByDesc(sub)
	assert.Equal(t, "select * from `users` order by (select `created_at` from `logins` where `user_id` = `users`.`id` limit 1) desc", b1.ToSql())
}

func TestOrderByInvalidDirectionParam(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").OrderBy("name", "invalid")
	assert.Error(t, b.Statement.Error, "invalid direction for order by: invalid")
}
