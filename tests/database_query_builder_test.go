package tests

import (
	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"strings"
	"testing"
)

func GetBuilder(pretend ...bool) *goeloquent.QueryBuilder {
	if len(pretend) > 0 && pretend[0] == false {
		stmt := goeloquent.NewStatement(GetConnection())
		query := goeloquent.NewQueryBuilder(stmt)
		return query
	}
	stmt := goeloquent.NewStatement()
	query := goeloquent.NewQueryBuilder(stmt)
	return query.Pretend()
}
func GetConnection() goeloquent.Connection {
	dsn := os.Getenv("GOELOQUENT_TEST_DSN")
	if dsn == "" {
		panic("set an environment variable GOELOQUENT_TEST_DSN to run tests")
	}
	conn, err := goeloquent.Open("test", goeloquent.DBConfig{
		Driver: goeloquent.DriverMysql,
		DSN:    dsn,
	})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	return conn
}
func RunWithDB(before, after string, test func(conn goeloquent.Connection)) {
	conn := GetConnection()
	defer func() {
		if after != "" {
			conn.GetDB().Exec(strings.ReplaceAll(after, `"`, "`"))
		}
	}()
	if before != "" {
		for _, s := range strings.Split(before, ";") {
			if s == "" {
				continue
			}
			_, err := conn.GetDB().Exec(strings.ReplaceAll(s, `"`, "`"))

			if err != nil {
				panic("failed to run before statement: " + err.Error())
			}
		}
	}

	test(conn)

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
func TestOrWhereDaySqlServer(t *testing.T) {

}
func TestOrWhereMonthSqlServer(t *testing.T) {

}
func TestOrWhereYearPostgres(t *testing.T) {

}
func TestOrWhereYearSqlServer(t *testing.T) {

}
func TestWhereTimeOperatorOptionalPostgres(t *testing.T) {

}
func TestWhereTimeSqlServer(t *testing.T) {

}
func TestOrWhereTimePostgres(t *testing.T) {

}
func TestOrWhereTimeSqlServer(t *testing.T) {

}
func TestWhereDatePostgres(t *testing.T) {

}
func TestWhereDayPostgres(t *testing.T) {

}
func TestWhereMonthPostgres(t *testing.T) {

}
func TestWhereYearPostgres(t *testing.T) {

}
func TestWhereTimePostgres(t *testing.T) {

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

	b1 := GetBuilder()
	b1.Select("*").From("users").WhereTime("updated_at", ">", "13:00:00").WhereTime("created_at", "!=", "14:00:00", goeloquent.Or)
	assert.Equal(t, "select * from `users` where time(`updated_at`) > ? or time(`created_at`) != ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"13:00:00", "14:00:00"}, b1.GetBindings())
}
func TestWhereTimeOperatorOptionalMySql(t *testing.T) {
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

func TestWherePast(t *testing.T) {

}
func TestWherePastUsesArray(t *testing.T) {

}
func TestWhereTodayMySQL(t *testing.T) {

}
func TestPassingArrayToWhereTodayMySQL(t *testing.T) {

}
func TestWhereFuture(t *testing.T) {

}
func TestPassingArrayToWhereFuture(t *testing.T) {

}
func TestWhereLikePostgres(t *testing.T) {

}
func TestWhereLikeClausePostgres(t *testing.T) {

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

func TestWhereLikeClauseSqlite(t *testing.T) {

}
func TestWhereLikeClauseSqlServer(t *testing.T) {

}
func TestWhereDateSqlite(t *testing.T) {

}
func TestWhereDaySqlite(t *testing.T) {

}
func TestWhereMonthSqlite(t *testing.T) {

}
func TestWhereYearSqlite(t *testing.T) {

}
func TestWhereTimeSqlite(t *testing.T) {

}
func TestWhereTimeOperatorOptionalSqlite(t *testing.T) {

}
func TestWhereDateSqlServer(t *testing.T) {

}
func TestWhereDaySqlServer(t *testing.T) {

}
func TestWhereMonthSqlServer(t *testing.T) {

}
func TestWhereYearSqlServer(t *testing.T) {

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

func TestWhereIntegerInRaw(t *testing.T) {

}
func TestOrWhereIntegerInRaw(t *testing.T) {

}
func TestWhereIntegerNotInRaw(t *testing.T) {

}
func TestOrWhereIntegerNotInRaw(t *testing.T) {

}
func TestEmptyWhereIntegerInRaw(t *testing.T) {

}
func TestEmptyWhereIntegerNotInRaw(t *testing.T) {

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
func TestArrayWhereNulls(t *testing.T) {
	b := GetBuilder()
	b.WhereNull([]interface{}{"email", "name"}).From("users")
	assert.Equal(t, "select * from `users` where `email` is null and `name` is null", b.ToSql())

	b = GetBuilder()
	b.Where("name", "test").OrWhereNull([]interface{}{"email", "name"}).From("users")
	assert.Equal(t, "select * from `users` where `name` = ? or `email` is null or `name` is null", b.ToSql())
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
func TestArrayWhereNotNulls(t *testing.T) {
	b := GetBuilder()
	b.WhereNotNull([]interface{}{"email", "name"}).From("users")
	assert.Equal(t, "select * from `users` where `email` is not null and `name` is not null", b.ToSql())

	b = GetBuilder()
	b.Where("name", "test").OrWhereNotNull([]interface{}{"email", "name"}).From("users")
	assert.Equal(t, "select * from `users` where `name` = ? or `email` is not null or `name` is not null", b.ToSql())
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

func TestLatest(t *testing.T) {

}

func TestOldest(t *testing.T) {

}

func TestInRandowOrderMysql(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").InRandomOrder()
	assert.Equal(t, "select * from `users` order by RAND()", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

}

func TestInRandomOrderPostgres(t *testing.T) {

}

func TestInRandomOrderSqlServer(t *testing.T) {}
func TestOrderBysSqlServer(t *testing.T)      {}
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
	assert.Equal(t, b.Statement.Error.Error(), "invalid direction for order by: invalid")
}

func TestHavings(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Having("name", "=", "foo")
	assert.Equal(t, "select * from `users` having `name` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{"foo"}, b.GetRawBindings()["having"])

	b1 := GetBuilder()
	b1.Select().From("users").OrHaving("name", "<>", "bar").OrHaving("email", "baz")
	assert.Equal(t, "select * from `users` having `name` <> ? or `email` = ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"bar", "baz"}, b1.GetBindings())
	assert.ElementsMatch(t, []interface{}{"bar", "baz"}, b1.GetRawBindings()["having"])

	b2 := GetBuilder()
	b2.Select().From("users").GroupBy("email").Having("email", ">", 1)
	assert.Equal(t, "select * from `users` group by `email` having `email` > ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b2.GetBindings())

	b3 := GetBuilder()
	b3.Select("email as foo_email").From("users").Having("foo_email", ">", 1)
	assert.Equal(t, "select `email` as `foo_email` from `users` having `foo_email` > ?", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b3.GetBindings())

	b4 := GetBuilder()
	b4.Select([]interface{}{"category", goeloquent.Raw("count(*) as 'total'")}).
		From("item").Where("department", "=", "popular").
		GroupBy("category").Having("total", ">", goeloquent.Raw("3"))
	assert.Equal(t, "select `category`, count(*) as 'total' from `item` where `department` = ? group by `category` having `total` > 3", b4.ToSql())

	b5 := GetBuilder()
	b5.Select([]interface{}{"category", goeloquent.Raw("count(*) as 'total'")}).
		From("item").Where("department", "=", "popular").
		GroupBy("category").Having("total", ">", 3)
	assert.Equal(t, "select `category`, count(*) as 'total' from `item` where `department` = ? group by `category` having `total` > ?", b5.ToSql())
	assert.ElementsMatch(t, []interface{}{"popular", 3}, b5.GetBindings())

}

func TestNestedHavings(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Having("name", "=", "foo").OrHaving(func(builder *goeloquent.QueryBuilder) {
		builder.Having("email", "<>", "bar").OrHaving("age", ">", 30)
	})
	assert.Equal(t, "select * from `users` having `name` = ? or (`email` <> ? or `age` > ?)", b.ToSql())
	assert.Nil(t, b.Statement.Error)
}

func TestNestedHavingBindings(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Having("name", "=", "foo").OrHaving(func(builder *goeloquent.QueryBuilder) {
		builder.Having("email", "<>", "bar").OrHaving("age", ">", 30)
	}).HavingRaw("created_at > ?", []interface{}{"2023-01-01"})
	assert.Equal(t, "select * from `users` having `name` = ? or (`email` <> ? or `age` > ?) and created_at > ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "bar", 30, "2023-01-01"}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{"foo", "bar", 30, "2023-01-01"}, b.GetRawBindings()["having"])
}

func TestHavingBetweens(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").HavingBetween("age", []interface{}{18, 30})
	assert.Nil(t, b.Statement.Error)
	assert.Equal(t, "select * from `users` having `age` between ? and ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{18, 30}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{18, 30}, b.GetRawBindings()["having"])

}

func TestHavingNull(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").HavingNull("email")
	assert.Equal(t, "select * from `users` having `email` is null", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
	assert.Nil(t, b.Statement.Error)

	b = GetBuilder()
	b.Select().From("users").HavingNull("email").HavingNull("phone")
	assert.Equal(t, "select * from `users` having `email` is null and `phone` is null", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").OrHavingNull("email").OrHavingNotNull("phone")
	assert.Equal(t, "select * from `users` having `email` is null or `phone` is not null", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select("email as mail").From("users").HavingNull("mail")
	assert.Equal(t, "select `email` as `mail` from `users` having `mail` is null", b2.ToSql())

	b3 := GetBuilder()
	b3.Select([]interface{}{"email", goeloquent.Raw("count(*) as 'total'")}).From("users").Where("department", "=", "popular").GroupBy("email").HavingNull("total")
	assert.Equal(t, "select `email`, count(*) as 'total' from `users` where `department` = ? group by `email` having `total` is null", b3.ToSql())

}

func TestHavingNotNull(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").HavingNotNull("email")
	assert.Equal(t, "select * from `users` having `email` is not null", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
	assert.Nil(t, b.Statement.Error)

	b1 := GetBuilder()
	b1.Select().From("users").HavingNotNull("email").HavingNotNull("phone")
	assert.Equal(t, "select * from `users` having `email` is not null and `phone` is not null", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())

	b1 = GetBuilder()
	b1.Select().From("users").OrHavingNotNull("email").OrHavingNull("phone")
	assert.Equal(t, "select * from `users` having `email` is not null or `phone` is null", b1.ToSql())

	b1 = GetBuilder()
	b1.Select("*").From("users").GroupBy("email").HavingNotNull("email")
	assert.Equal(t, "select * from `users` group by `email` having `email` is not null", b1.ToSql())

	b2 := GetBuilder()
	b2.Select("email as mail").From("users").HavingNotNull("mail")
	assert.Equal(t, "select `email` as `mail` from `users` having `mail` is not null", b2.ToSql())

	b3 := GetBuilder()
	b3.Select([]interface{}{"email", goeloquent.Raw("count(*) as 'total'")}).From("users").Where("department", "=", "popular").GroupBy("email").HavingNotNull("total")
	assert.Equal(t, "select `email`, count(*) as 'total' from `users` where `department` = ? group by `email` having `total` is not null", b3.ToSql())

}

func TestHavingExpression(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Having(goeloquent.Raw("1 = 1"))
	assert.Equal(t, "select * from `users` having 1 = 1", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}

func TestHavingShortcut(t *testing.T) {

	b1 := GetBuilder()
	b1.Select().From("users").Having("name", "a").OrHaving("name", "bar")
	assert.Equal(t, "select * from `users` having `name` = ? or `name` = ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"a", "bar"}, b1.GetBindings())
}

func TestHavingFollowedBySelectGet(t *testing.T) {
	b := GetBuilder()
	b.From("item").Select([]interface{}{"category", goeloquent.Raw("count(*) as 'total'")}).Where("department", "=", "popular").GroupBy("category").
		Having("total", ">", 3)

	assert.Equal(t, "select `category`, count(*) as 'total' from `item` where `department` = ? group by `category` having `total` > ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"popular", 3}, b.GetBindings())

	b = GetBuilder()
	b.From("item").Select([]interface{}{"category", goeloquent.Raw("count(*) as 'total'")}).Where("department", "=", "popular").GroupBy("category").
		Having("total", ">", goeloquent.Raw("3"))

	assert.Equal(t, "select `category`, count(*) as 'total' from `item` where `department` = ? group by `category` having `total` > 3", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"popular"}, b.GetBindings())
}

func TestRawHavings(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").HavingRaw("name < foo")
	assert.Equal(t, "select * from `users` having name < foo", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{}, b.GetRawBindings()["having"])

	b = GetBuilder()
	b.Select().From("users").Having("name", "<>", 1).OrHavingRaw("user_foo < user_bar")
	assert.Equal(t, "select * from `users` having `name` <> ? or user_foo < user_bar", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").HavingBetween("age", []interface{}{18, 30}).OrHavingRaw("user_foo < user_bar")
	assert.Equal(t, "select * from `users` having `age` between ? and ? or user_foo < user_bar", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{18, 30}, b1.GetBindings())
}

func TestLimitsAndOffsets(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Limit(10).Offset(5)
	assert.Equal(t, "select * from `users` limit 10 offset 5", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Limit(0)
	assert.Equal(t, "select * from `users` limit 0", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").Skip(5).Take(6)
	assert.Equal(t, "select * from `users` limit 6 offset 5", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b2.GetBindings())

	b3 := GetBuilder()
	b3.Select().From("users").Skip(0).Take(0)
	assert.Equal(t, "select * from `users` limit 0 offset 0", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b3.GetBindings())

	b4 := GetBuilder()
	b4.Select().From("users").Limit(-10).Offset(-5).ReOrder()
	assert.Equal(t, "select * from `users` offset 0", b4.ToSql())

}

func TestForPage(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").ForPage(2, 15)
	assert.Equal(t, "select * from `users` limit 15 offset 15", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").ForPage(0, 15)
	assert.Equal(t, "select * from `users` limit 15 offset 0", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").ForPage(-2, 15)
	assert.Equal(t, "select * from `users` limit 15 offset 0", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b2.GetBindings())

	b3 := GetBuilder()
	b3.Select().From("users").ForPage(2, 0)
	assert.Equal(t, "select * from `users` limit 0 offset 0", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b3.GetBindings())

	b4 := GetBuilder()
	b4.Select().From("users").ForPage(0, 0)
	assert.Equal(t, "select * from `users` limit 0 offset 0", b4.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b4.GetBindings())

	b5 := GetBuilder()
	b5.Select().From("users").ForPage(-2, 0)
	assert.Equal(t, "select * from `users` limit 0 offset 0", b5.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b5.GetBindings())

}

func TestForPageBeforeId(t *testing.T) {

	b := GetBuilder()
	b.Select().From("users").ForPageBeforeId(15, 0)
	assert.Equal(t, "select * from `users` where `id` < ? order by `id` desc limit 15", b.ToSql())
	assert.ElementsMatch(t, []interface{}{int64(0)}, b.GetBindings())

}

func TestForPageAfterId(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").ForPageAfterId(15, 0)
	assert.Equal(t, "select * from `users` where `id` > ? order by `id` asc limit 15", b.ToSql())
	assert.ElementsMatch(t, []interface{}{int64(0)}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").ForPageAfterId(15, 10)
	assert.Equal(t, "select * from `users` where `id` > ? order by `id` asc limit 15", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{int64(10)}, b1.GetBindings())
}

func TestGetCountForPaginationWithBindings(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").SelectSub(func(builder *goeloquent.QueryBuilder) {
		builder.Select("name").From("users").Where("name", "John")
	}, "posts")
	b.Pretend()
	var count int
	c := goeloquent.Clone(b)
	c.Without([]goeloquent.Component{goeloquent.COMPONENT_COLUMN, goeloquent.COMPONENT_ORDER, goeloquent.COMPONENT_OFFSET, goeloquent.COMPONENT_LIMIT},
		[]goeloquent.Component{goeloquent.COMPONENT_SELECT, goeloquent.COMPONENT_ORDER}).
		Count(&count)

	assert.Equal(t, "select count(*) as aggregate from `users`", c.ToSql())
}

func TestGetCountForPaginationWithColumnAliases(t *testing.T) {
	b := GetBuilder()
	columns := []interface{}{"body as post_body", "teaser", "posts.created as published"}
	b.Select(columns).From("posts")
	b.Pretend()
	var count int
	c := goeloquent.Clone(b)
	c.Without([]goeloquent.Component{goeloquent.COMPONENT_COLUMN, goeloquent.COMPONENT_ORDER, goeloquent.COMPONENT_OFFSET, goeloquent.COMPONENT_LIMIT},
		[]goeloquent.Component{goeloquent.COMPONENT_SELECT, goeloquent.COMPONENT_ORDER}).
		Count(&count, goeloquent.WithoutSelectAliases(columns)...)

	assert.Equal(t, "select count(`body`, `teaser`, `posts`.`created`) as aggregate from `posts`", c.ToSql())
	assert.ElementsMatch(t, []interface{}{}, c.GetBindings())
}

func TestGetCountForPaginationWithUnion(t *testing.T) {
}
func TestGetCountForPaginationWithUnionOrders(t *testing.T) {
}
func TestGetCountForPaginationWithUnionLimitAndOffset(t *testing.T) {
}

func TestWhereShortcut(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("name", "a").OrWhere("name", "bar")
	assert.Equal(t, "select * from `users` where `name` = ? or `name` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"a", "bar"}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("name", "a").OrWhere("name", goeloquent.Raw("'bar'"))
	assert.Equal(t, "select * from `users` where `name` = ? or `name` = 'bar'", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"a"}, b1.GetBindings())
}

func TestOrWheresHaveConsistentResults(t *testing.T) {

}

func TestWhereWithArrayConditions(t *testing.T) {

	//mixed
	b1 := GetBuilder().Select().From("users").Where([][]interface{}{
		{"admin", "=", 1},
		{"id", "<", 10},
		{"source", "=", "301"},
		{"deleted", 0},
		{"role", "in", []interface{}{"admin", "manager", "owner"}},
		{"age", "between", []interface{}{18, 100}},
		{func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
			return builder.WhereYear("created_at", "<", 2010).WhereColumn("first_name", "last_name").OrWhereNull("created_at")
		}},
		{goeloquent.Raw("year(birthday) < 1998")},
		{"suspend", goeloquent.Raw("'nodoublequotes'")},
	})
	assert.Equal(t, b1.ToSql(),
		"select * from `users` where (`admin` = ? and `id` < ? and `source` = ? and `deleted` = ? and `role` in (?, ?, ?) and `age` between ? and ? "+
			"and (year(`created_at`) < ? and `first_name` = `last_name` or `created_at` is null) "+
			"and year(birthday) < 1998 and `suspend` = 'nodoublequotes')")
	assert.ElementsMatch(t, []interface{}{1, 10, "301", 0, "admin", "manager", "owner", 18, 100, 2010}, b1.GetBindings())

	b2 := GetBuilder()
	b2.From("users").Where([][]interface{}{{"foo", 1}, {"bar", 2}}, "or")
	assert.Equal(t, "select * from `users` where (`foo` = ? or `bar` = ?)", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2}, b2.GetBindings())

	b2 = GetBuilder()
	b2.From("users").Where(map[string]interface{}{"name": "foo", "age": 30, "email": goeloquent.Raw("'bar'")})
	assert.Equal(t, "select * from `users` where (`age` = ? and `email` = 'bar' and `name` = ?)", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", 30}, b2.GetBindings())

	b2 = GetBuilder()
	b2.From("users").Where(map[string]interface{}{"name": "foo", "age": 30, "email": goeloquent.Raw("'bar'")}, "or")
	assert.Equal(t, "select * from `users` where (`age` = ? or `email` = 'bar' or `name` = ?)", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", 30}, b2.GetBindings())

	b2 = GetBuilder()
	b2.From("users").Where([][]interface{}{{"foo", 1}, {"bar", "<", 2}}, "and")
	assert.Equal(t, "select * from `users` where (`foo` = ? and `bar` < ?)", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2}, b2.GetBindings())

	b2 = GetBuilder()
	b2.From("users").Where([][]interface{}{{"foo", 1}, {"bar", "<", 2}}, "or")
	assert.Equal(t, "select * from `users` where (`foo` = ? or `bar` < ?)", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2}, b2.GetBindings())

}
func TestNestedWheres(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("age", ">", 25).OrWhere(func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.Where("name", "foo").Where("email", "bar")
	}).Where("status", "active")
	assert.Equal(t, "select * from `users` where `age` > ? or (`name` = ? and `email` = ?) and `status` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{25, "foo", "bar", "active"}, b.GetBindings())
}
func TestNestedWhereBindings(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("age", ">", 25).OrWhere(func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.Where("name", "foo").Where("email", "bar")
	}).Where("status", "active")
	assert.Equal(t, "select * from `users` where `age` > ? or (`name` = ? and `email` = ?) and `status` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{25, "foo", "bar", "active"}, b.GetBindings())
}

func TestWhereNot(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereNot("name", "foo")
	assert.Equal(t, "select * from `users` where not `name` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo"}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").WhereNot(func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.Where("name", "foo").OrWhere("email", "bar")
	})
	assert.Equal(t, "select * from `users` where not (`name` = ? or `email` = ?)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "bar"}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").Where("age", 1).OrWhereNot(func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.Where("name", "foo").OrWhere("email", "bar")
	})
	assert.Equal(t, "select * from `users` where `age` = ? or not (`name` = ? or `email` = ?)", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo", "bar"}, b2.GetBindings())

}

func TestIncrementManyArgumentValidation1(t *testing.T) {
}
func TestIncrementManyArgumentValidation2(t *testing.T) {
}

func TestWhereNotWithArrayConditions(t *testing.T) {

}

func TestFullSubSelects(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("name", "a").OrWhere("id", "=", func(builder *goeloquent.QueryBuilder) {
		builder.SelectRaw("max(id)").From("users").Where("email", "=", "gmail")
	})
	assert.Equal(t, "select * from `users` where `name` = ? or `id` = (select max(id) from `users` where `email` = ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"a", "gmail"}, b.GetBindings())
}

func TestWhereExists(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereExists(func(builder *goeloquent.QueryBuilder) {
		builder.Select().From("posts").WhereColumn("posts.user_id", "users.id")
	})
	assert.Equal(t, "select * from `users` where exists (select * from `posts` where `posts`.`user_id` = `users`.`id`)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").WhereNotExists(func(builder *goeloquent.QueryBuilder) {
		builder.Select().From("posts").WhereColumn("posts.user_id", "users.id")
	})
	assert.Equal(t, "select * from `users` where not exists (select * from `posts` where `posts`.`user_id` = `users`.`id`)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").Where("id", 1).OrWhereExists(func(builder *goeloquent.QueryBuilder) {
		builder.Select().From("posts").WhereColumn("posts.user_id", "users.id")
	})
	assert.Equal(t, "select * from `users` where `id` = ? or exists (select * from `posts` where `posts`.`user_id` = `users`.`id`)", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b2.GetBindings())

	b3 := GetBuilder()
	b3.Select().From("users").Where("id", 1).OrWhereNotExists(func(builder *goeloquent.QueryBuilder) {
		builder.Select().From("posts").WhereColumn("posts.user_id", "users.id")
	})
	assert.Equal(t, "select * from `users` where `id` = ? or not exists (select * from `posts` where `posts`.`user_id` = `users`.`id`)", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b3.GetBindings())

}

func TestBasicJoins(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Join("posts", "users.id", "=", "posts.user_id")
	assert.Equal(t, "select * from `users` inner join `posts` on `users`.`id` = `posts`.`user_id`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Join("posts", "users.id", "=", "posts.user_id").LeftJoin("comments", "users.id", "=", "comments.user_id")
	assert.Equal(t, "select * from `users` inner join `posts` on `users`.`id` = `posts`.`user_id` left join `comments` on `users`.`id` = `comments`.`user_id`", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").LeftJoinWhere("posts", "users.id", "=", "bar").JoinWhere("posts", "users.id", "=", "foo")
	assert.Equal(t, "select * from `users` left join `posts` on `users`.`id` = ? inner join `posts` on `users`.`id` = ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{"bar", "foo"}, b2.GetBindings())
}

func TestCrossJoins(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").CrossJoin("posts")
	assert.Equal(t, "select * from `users` cross join `posts`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Join("posts", "posts.user_id", "=", "users.id", goeloquent.JoinTypeCross)
	assert.Equal(t, "select * from `users` cross join `posts` on `posts`.`user_id` = `users`.`id`", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").CrossJoin("posts", "users.id", "=", "posts.user_id")
	assert.Equal(t, "select * from `users` cross join `posts` on `users`.`id` = `posts`.`user_id`", b2.ToSql())

}

func TestCrossJoinSubs(t *testing.T) {
	b := GetBuilder()
	b.SelectRaw("(sale / overall.sales) * 100 AS percent_of_total").From("sales").
		CrossJoinSub(GetBuilder().SelectRaw("SUM(sale) AS sales").From("sales"), "overall")

	assert.Equal(t, "select (sale / overall.sales) * 100 AS percent_of_total from `sales` cross join (select SUM(sale) AS sales from `sales`) as `overall`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}

func TestComplexJoin(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) *goeloquent.JoinBuilder {
		return builder.On("users.id", "contacts.user_id").OrOn("users.name", "=", "contacts.name")
	})
	assert.Equal(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` or `users`.`name` = `contacts`.`name`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.Where("users.id", "=", "foo").OrWhere("users.name", "=", "bar")
	})
	assert.Equal(t, "select * from `users` inner join `contacts` on `users`.`id` = ? or `users`.`name` = ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "bar"}, b1.GetBindings())

}

func TestJoinWhereNull(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").WhereNull("contacts.deleted_at")
	})
	assert.Equal(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` and `contacts`.`deleted_at` is null", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").OrWhereNull("contacts.deleted_at")
	})
	assert.Equal(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` or `contacts`.`deleted_at` is null", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())
}

func TestJoinWhereNotNull(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").WhereNotNull("contacts.deleted_at")
	})
	assert.Equal(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` and `contacts`.`deleted_at` is not null", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").OrWhereNotNull("contacts.deleted_at")
	})
	assert.Equal(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` or `contacts`.`deleted_at` is not null", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())
}

func TestJoinWhereIn(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").WhereIn("contacts.name", []interface{}{1, 2, 3})
	})
	assert.Equal(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` and `contacts`.`name` in (?, ?, ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2, 3}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").OrWhereIn("contacts.name", []interface{}{1, 2, 3})
	})
	assert.Equal(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` or `contacts`.`name` in (?, ?, ?)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2, 3}, b1.GetBindings())
}

func TestJoinWhereInSubQuery(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) {
		tb := GetBuilder().Select("name").From("contacts").Where("active", 1)
		builder.On("users.id", "=", "contacts.user_id").WhereIn("contacts.name", tb)
	})
	assert.Equal(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` and `contacts`.`name` in (select `name` from `contacts` where `active` = ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) {
		tb := GetBuilder().Select("name").From("contacts").Where("active", 1)
		builder.On("users.id", "=", "contacts.user_id").OrWhereIn("contacts.name", tb)
	})
	assert.Equal(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` or `contacts`.`name` in (select `name` from `contacts` where `active` = ?)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())
}

func TestJoinWhereNotIn(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").WhereNotIn("contacts.name", []interface{}{1, 2, 3})
	})
	assert.Equal(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` and `contacts`.`name` not in (?, ?, ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2, 3}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").OrWhereNotIn("contacts.name", []interface{}{1, 2, 3})
	})
	assert.Equal(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` or `contacts`.`name` not in (?, ?, ?)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2, 3}, b1.GetBindings())
}

func TestJoinsWithNestedConditions(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").LeftJoin("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").Where(func(subbuilder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
			return subbuilder.Where("contacts.name", "foo").OrWhere("contacts.email", "bar")
		})
	})
	assert.Equal(t, "select * from `users` left join `contacts` on `users`.`id` = `contacts`.`user_id` and (`contacts`.`name` = ? or `contacts`.`email` = ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "bar"}, b.GetBindings())
}

func TestJoinWithAdvancedConditions(t *testing.T) {

	b := GetBuilder()
	b.Select().From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "contacts.user_id").Where(func(subbuilder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
			return subbuilder.Where("contacts.name", "foo").OrWhere(func(subbuilder2 *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
				return subbuilder2.Where("contacts.email", "bar").WhereNotNull("contacts.phone")
			})
		})
	})
	assert.Equal(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` and (`contacts`.`name` = ? or (`contacts`.`email` = ? and `contacts`.`phone` is not null))", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "bar"}, b.GetBindings())
}

func TestJoinsWithSubqueryCondition(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").LeftJoin("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").WhereIn("contacts_type_id", func(subbuilder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
			return subbuilder.Select("id").From("contacts_type").Where("active", 1).WhereNull("invalid")
		})
	})
	assert.Equal(t, "select * from `users` left join `contacts` on `users`.`id` = `contacts`.`user_id` and `contacts_type_id` in (select `id` from `contacts_type` where `active` = ? and `invalid` is null)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").LeftJoin("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").WhereExists(func(subbuilder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
			return subbuilder.Select().From("contacts_type").SelectRaw("1").WhereRaw("contacts_type.id = contacts.contacts_type_id").
				Where("active", 1).WhereNull("invalid")
		})
	})
	assert.Equal(t, "select * from `users` left join `contacts` on `users`.`id` = `contacts`.`user_id` and exists (select 1 from `contacts_type` where contacts_type.id = contacts.contacts_type_id and `active` = ? and `invalid` is null)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())
}

func TestJoinsWithAdvancedSubqueryCondition(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").LeftJoin("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").WhereExists(func(subbuilder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
			return subbuilder.Select().From("contacts_type").SelectRaw("1").WhereRaw("contacts_type.id = contacts.contacts_type_id").
				Where("active", 1).WhereNull("invalid").WhereIn("level_id", func(subbuilder2 *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
				return subbuilder2.Select("id").From("levels").Where("active", 1)
			})
		})
	})
	assert.Equal(t, "select * from `users` left join `contacts` on `users`.`id` = `contacts`.`user_id` and exists (select 1 from `contacts_type` where contacts_type.id = contacts.contacts_type_id and `active` = ? and `invalid` is null and `level_id` in (select `id` from `levels` where `active` = ?))", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 1}, b.GetBindings())
}

func TestJoinsWithNestedJoins(t *testing.T) {
	b := GetBuilder()
	b.From("users").Select("users.id", "contacts.id", "contact_types.id").LeftJoin("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").
			Join("contact_types", "contacts.contact_type_id", "=", "contact_types.id")
	})
	assert.Equal(t, "select `users`.`id`, `contacts`.`id`, `contact_types`.`id` from `users` left join (`contacts` inner join `contact_types` on `contacts`.`contact_type_id` = `contact_types`.`id`) on `users`.`id` = `contacts`.`user_id`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}

func TestJoinsWithMultipleNestedJoins(t *testing.T) {
	b := GetBuilder()
	b.Select("users.id", "contacts.user_id", "contact_types.id", "planets.id", "countries.id").From("users").LeftJoin("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.user_id").
			Join("contact_types", "contacts.contact_type_id", "=", "contact_types.id").
			LeftJoin("countries", func(subbuilder *goeloquent.JoinBuilder) {
				subbuilder.On("contacts.country", "=", "countries.country").
					Join("planets", func(sub1 *goeloquent.JoinBuilder) {
						sub1.On("countries.planet_id", "=", "planets.id").
							Where("planet.is_settleted", 1).
							Where("planet.population", ">", 1000000)
					})
			})
	})
	assert.Equal(t, "select `users`.`id`, `contacts`.`user_id`, `contact_types`.`id`, `planets`.`id`, `countries`.`id` from `users` left join "+
		"(`contacts` inner join `contact_types` on `contacts`.`contact_type_id` = `contact_types`.`id` "+
		"left join (`countries` inner join `planets` on `countries`.`planet_id` = `planets`.`id` and `planet`.`is_settleted` = ? and `planet`.`population` > ?) on `contacts`.`country` = `countries`.`country`) on `users`.`id` = `contacts`.`user_id`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 1000000}, b.GetBindings())
}

func TestJoinsWithNestedJoinWithAdvancedSubqueryCondition(t *testing.T) {
	b := GetBuilder()
	b.Select("users.id", "contacts.id", "contact_types.id").From("users").LeftJoin("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.id").Join("contact_types", "contacts.contact_type_id", "=", "contact_types.id").WhereExists(func(subbuilder *goeloquent.QueryBuilder) {
			subbuilder.Select().From("countries").WhereColumn("contacts.country", "countries.country").
				Join("planets", func(sub1 *goeloquent.JoinBuilder) {
					sub1.On("countries.planet_id", "=", "planets.id").Where("planets.is_settleted", 1)
				}).Where("planet.population", ">=", 1000000)
		})
	})
	assert.Equal(t, "select `users`.`id`, `contacts`.`id`, `contact_types`.`id` from `users` left join (`contacts` inner join `contact_types` on `contacts`.`contact_type_id` = `contact_types`.`id`) on `users`.`id` = `contacts`.`id` and exists (select * from `countries` inner join `planets` on `countries`.`planet_id` = `planets`.`id` and `planets`.`is_settleted` = ? where `contacts`.`country` = `countries`.`country` and `planet`.`population` >= ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 1000000}, b.GetBindings())
}

func TestJoinWithNestedOnCondition(t *testing.T) {
	b := GetBuilder()
	b.Select("users.id").From("users").Join("contacts", func(builder *goeloquent.JoinBuilder) {
		builder.On("users.id", "=", "contacts.id").AddNestedWhereQuery(GetBuilder().Where("contacts.id", 1))
	})
	assert.Equal(t, "select `users`.`id` from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` and (`contacts`.`id` = ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
}

func TestJoinSub(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").JoinSub("select * from contacts", "sub", "users.id", "=", "sub.id")
	assert.Equal(t, "select * from `users` inner join (select * from contacts) as `sub` on `users`.`id` = `sub`.`id`", b.ToSql())

	b1 := GetBuilder()
	b1.Select().From("users").JoinSub(func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.From("contacts")
	}, "sub", "users.id", "=", "sub.id")
	assert.Equal(t, "select * from `users` inner join (select * from `contacts`) as `sub` on `users`.`id` = `sub`.`id`", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())

	sub1 := GetBuilder().From("contacts").Where("active", 1)
	sub2 := GetBuilder().From("contacts").Where("name", "foo")

	b2 := GetBuilder()
	b2.From("users").JoinSub(sub1, "sub1", "users.id", "=", 2, goeloquent.JoinTypeInner, true).
		JoinSub(sub2, "sub2", "users.id", "=", "sub2.user_id")

	assert.Equal(t, "select * from `users` inner join (select * from `contacts` where `active` = ?) as `sub1` on `users`.`id` = ? "+
		"inner join (select * from `contacts` where `name` = ?) as `sub2` on `users`.`id` = `sub2`.`user_id`", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2, "foo"}, b2.GetBindings())
}

func TestJoinSubWithPrefix(t *testing.T) {
	b := GetBuilder()
	b.Grammar.SetTablePrefix("prefix_")
	b.Select().From("users").JoinSub("select * from contacts", "sub", "users.id", "=", "sub.id")
	assert.Equal(t, "select * from `prefix_users` inner join (select * from contacts) as `prefix_sub` on `prefix_users`.`id` = `prefix_sub`.`id`", b.ToSql())
}

func TestLeftJoinSub(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").LeftJoinSub("select * from contacts", "sub", "users.id", "=", "sub.id")
	assert.Equal(t, "select * from `users` left join (select * from contacts) as `sub` on `users`.`id` = `sub`.`id`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}
func TestRightJoinSub(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").RightJoinSub("select * from contacts", "sub", "users.id", "=", "sub.id")
	assert.Equal(t, "select * from `users` right join (select * from contacts) as `sub` on `users`.`id` = `sub`.`id`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}

func TestJoinLateral(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").JoinLateral("select * from `contacts` where `contracts`.`user_id` = `users`.`id`", "sub", goeloquent.JoinTypeInner)
	assert.Equal(t, "select * from `users` inner join lateral (select * from `contacts` where `contracts`.`user_id` = `users`.`id`) as `sub` on true", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").JoinLateral(func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.Select().From("contacts").WhereColumn("contracts.user_id", "users.id")
	}, "sub", goeloquent.JoinTypeInner)
	assert.Equal(t, "select * from `users` inner join lateral (select * from `contacts` where `contracts`.`user_id` = `users`.`id`) as `sub` on true", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())

	sub1 := GetBuilder().From("contacts").WhereColumn("contracts.user_id", "users.id").Where("active", 1)
	sub2 := GetBuilder().From("contacts").WhereColumn("contracts.user_id", "users.id").Where("name", "foo")

	b2 := GetBuilder()
	b2.From("users").JoinLateral(sub1, "sub1", goeloquent.JoinTypeInner).JoinLateral(sub2, "sub2", goeloquent.JoinTypeInner)
	assert.Equal(t, "select * from `users` inner join lateral (select * from `contacts` where `contracts`.`user_id` = `users`.`id` and `active` = ?) as `sub1` on true "+
		"inner join lateral (select * from `contacts` where `contracts`.`user_id` = `users`.`id` and `name` = ?) as `sub2` on true", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b2.GetBindings())
}

func TestJoinLateralMariaDB(t *testing.T) {

}

func TestJoinLateralSQLlite(t *testing.T) {

}

func TestJoinLateralPostgres(t *testing.T) {

}

func TestJoinLateralSqlServer(t *testing.T) {

}

func TestJoinLateralWithPrefix(t *testing.T) {
	b := GetBuilder()
	b.Grammar.SetTablePrefix("prefix_")
	b.Select().From("users").JoinLateral("select * from `contacts` where `contacts`.`user_id` = `users`.`id`", "sub", goeloquent.JoinTypeInner)
	assert.Equal(t, "select * from `prefix_users` inner join lateral (select * from `contacts` where `contacts`.`user_id` = `users`.`id`) as `prefix_sub` on true", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}

func TestLeftJoinLateral(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").LeftJoinLateral("select * from `contacts` where `contacts`.`user_id` = `users`.`id`", "sub")
	assert.Equal(t, "select * from `users` left join lateral (select * from `contacts` where `contacts`.`user_id` = `users`.`id`) as `sub` on true", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}

func TestLeftJoinLateralSqlServer(t *testing.T) {
}

func TestRawExpressionsInSelect(t *testing.T) {
	b := GetBuilder()
	b.Select(goeloquent.Raw("substr(foo,6)")).From("users")
	assert.Equal(t, "select substr(foo,6) from `users`", b.ToSql())
}

func TestFindReturnsFirstResultByID(t *testing.T) {
	var dest map[string]interface{}
	b := GetBuilder()
	b.Select().From("users").Find(&dest, 1)
	assert.Equal(t, "select * from `users` where `id` = ? limit 1", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())

}

func TestFindOrReturnsFirstResultById(t *testing.T) {

}

func TestFirstMethodReturnsFirstResult(t *testing.T) {
	var dest map[string]interface{}
	b := GetBuilder()
	b.Select().From("users").Where("name", "foo").First(&dest)
	assert.Equal(t, "select * from `users` where `name` = ? limit 1", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo"}, b.GetBindings())
}

func TestFirstOrFailMethodReturnsFirstResult(t *testing.T) {

}

func TestFirstOrFailMethodThrowsRecordNotFoundException(t *testing.T) {

}

func TestPluckMethodGetsCollectionOfColumnValues(t *testing.T) {
	var dest []string
	b := GetBuilder()
	b.Select().From("users").Pluck(&dest, "name")
	assert.Equal(t, "select `name` from `users`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}

func TestPluckAvoidsDuplicateColumnSelection(t *testing.T) {

}

func TestImplode(t *testing.T) {

}
func TestValueMethodReturnsSingleColumn(t *testing.T) {
	var dest string
	b := GetBuilder()
	b.Select().From("users").Where("id", 1).Value(&dest, "name")
	assert.Equal(t, "select `name` from `users` where `id` = ? limit 1", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
}

func TestRawValueMethodReturnsSingleColumn(t *testing.T) {
	var dest string
	b := GetBuilder()
	b.Select().From("users").Where("id", 1).RawValue(&dest, "UPPER('foo')")
	assert.Equal(t, "select UPPER('foo') from `users` where `id` = ? limit 1", b.ToSql())
}

func TestAggregateFunctions(t *testing.T) {
	var count int64
	b := GetBuilder()
	b.Select().From("users").Where("active", 1).Count(&count)
	assert.Equal(t, "select count(*) as aggregate from `users` where `active` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
	assert.Equal(t, int64(0), count)

	b1 := GetBuilder()
	var exist bool
	b1.Select().From("users").Where("active", 1).Exists(&exist)
	assert.Equal(t, "select exists(select * from `users` where `active` = ?) as `exists`", b1.ToSql())

	b2 := GetBuilder()
	var notExist bool
	b2.Select().From("users").Where("active", 1).DoesntExist(&notExist)
	assert.Equal(t, "select exists(select * from `users` where `active` = ?) as `exists`", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b2.GetBindings())

	b3 := GetBuilder()
	b3.Select().From("users").Where("active", 1).Max(&count, "age")
	assert.Equal(t, "select max(`age`) as aggregate from `users` where `active` = ?", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b3.GetBindings())

	b4 := GetBuilder()
	b4.Select().From("users").Where("active", 1).Min(&count, "age")
	assert.Equal(t, "select min(`age`) as aggregate from `users` where `active` = ?", b4.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b4.GetBindings())

	b5 := GetBuilder()
	b5.Select().From("users").Where("active", 1).Sum(&count, "balance")
	assert.Equal(t, "select sum(`balance`) as aggregate from `users` where `active` = ?", b5.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b5.GetBindings())
	assert.Equal(t, int64(0), count)

	b6 := GetBuilder()
	b6.Select().From("users").Where("active", 1).Avg(&count, "age")
	assert.Equal(t, "select avg(`age`) as aggregate from `users` where `active` = ?", b6.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b6.GetBindings())
	assert.Equal(t, int64(0), count)

	b7 := GetBuilder()
	b7.Select().From("users").Where("active", 1).Aggregate(&count, "COUNT", "age")
	assert.Equal(t, "select COUNT(`age`) as aggregate from `users` where `active` = ?", b7.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b7.GetBindings())
	assert.Equal(t, int64(0), count)
}

func TestSqlServerExists(t *testing.T)                        {}
func TestExistsOr(t *testing.T)                               {}
func TestDoesntExistsOr(t *testing.T)                         {}
func TestAggregateResetFollowedByGet(t *testing.T)            {}
func TestAggregateResetFollowedBySelectGet(t *testing.T)      {}
func TestAggregateResetFollowedByGetWithColumns(t *testing.T) {}
func TestAggregateWithSubSelect(t *testing.T) {
	var count int64
	b := GetBuilder()
	b.Select().From("users").SelectSub(func(q *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return q.From("posts").Select("foo", "bar").Where("title", "baz")
	}, "posts").Count(&count)
	assert.Equal(t, "select count(*) as aggregate from `users`", b.ToSql())
}

func TestSubqueriesBindings(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("email", "=", func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.Select(goeloquent.Raw("max(id)")).From("users").Where("email", "=", "gmail").
			OrderByRaw("email like ?", []interface{}{"%.com"}).
			GroupBy("id").Having("id", "=", 4)
	}).OrWhere("id", "=", "foo").GroupBy("id").Having("id", "=", 5)
	assert.Equal(t, "select * from `users` where `email` = (select max(id) from `users` where `email` = ? group by `id` having `id` = ? order by email like ?) or `id` = ? group by `id` having `id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"gmail", 4, "%.com", "foo", 5}, b.GetBindings())
}

func TestInsertMethod(t *testing.T) {
	b := GetBuilder()
	b.From("users").Insert(map[string]interface{}{
		"name":  "John",
		"email": "foo",
	})
	assert.Equal(t, "insert into `users` (`email`, `name`) values (?, ?)", b.RawSql)
	assert.ElementsMatch(t, []interface{}{"foo", "John"}, b.GetBindings())
}

func TestInsertUsingMethod(t *testing.T) {
	b := GetBuilder()
	b.From("users").InsertUsing([]interface{}{"name", "email"}, func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.Select("name", "email").From("contacts").Where("active", 1)
	})
	assert.Equal(t, "insert into `users` (`name`, `email`) select `name`, `email` from `contacts` where `active` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
}

func TestInsertUsingWithEmptyColumns(t *testing.T) {
	b := GetBuilder()
	b.From("users").InsertUsing([]interface{}{}, func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.From("contacts").Where("active", 1)
	})
	assert.Equal(t, "insert into `users` select * from `contacts` where `active` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())

}

func TestInsertUsingInvalidSubquery(t *testing.T) {
}
func TestInsertOrIgnoreMethod(t *testing.T) {

}
func TestMysqlInsertOrIgnoreMethod(t *testing.T) {
	b := GetBuilder()
	b.From("users").InsertOrIgnore(map[string]interface{}{
		"name":  "John",
		"email": "foo",
	})
	assert.Equal(t, "insert ignore into `users` (`email`, `name`) values (?, ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "John"}, b.GetBindings())
}

func TestPostgresInsertOrIgnoreMethod(t *testing.T)       {}
func TestSQLiteInsertOrIgnoreMethod(t *testing.T)         {}
func TestSqlServerInsertOrIgnoreMethod(t *testing.T)      {}
func TestInsertOrIgnoreUsingMethod(t *testing.T)          {}
func TestSqlServerInsertOrIgnoreUsingMethod(t *testing.T) {}
func TestMySqlInsertOrIgnoreUsingMethod(t *testing.T) {
	b := GetBuilder()
	b.From("users").InsertOrIgnoreUsing([]interface{}{"name", "email"}, func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.Select("name", "email").From("contacts").Where("active", 1)
	})

	assert.Equal(t, "insert ignore into `users` (`name`, `email`) select `name`, `email` from `contacts` where `active` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
}
func TestMySqlInsertOrIgnoreUsingInvalidSubquery(t *testing.T) {
	b := GetBuilder()
	_, err := b.From("users").InsertOrIgnoreUsing([]interface{}{}, nil)
	assert.Equal(t, goeloquent.ErrorSubQueryInvalid.Error(), err.Error())

}
func TestPostgresInsertOrIgnoreUsingMethod(t *testing.T)           {}
func TestPostgresInsertOrIgnoreUsingWithEmptyColumns(t *testing.T) {}
func TestPostgresInsertOrIgnoreUsingInvalidSubquery(t *testing.T)  {}
func TestSQLiteInsertOrIgnoreUsingMethod(t *testing.T)             {}
func TestSQLiteInsertOrIgnoreUsingWithEmptyColumns(t *testing.T)   {}
func TestSQLiteInsertOrIgnoreUsingInvalidSubquery(t *testing.T)    {}
func TestInsertGetIdMethod(t *testing.T) {
	b := GetBuilder()
	_, e := b.From("users").InsertGetId(map[string]interface{}{
		"name":  "John",
		"email": "foo",
	})
	assert.Nil(t, e)
	assert.Equal(t, "insert into `users` (`email`, `name`) values (?, ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "John"}, b.GetBindings())
}
func TestInsertGetIdMethodRemovesExpressions(t *testing.T) {
	b := GetBuilder()
	_, e := b.From("users").InsertGetId(map[string]interface{}{
		"name":  goeloquent.Raw("UPPER('John')"),
		"email": "foo",
	})
	assert.Nil(t, e)
	assert.Equal(t, "insert into `users` (`email`, `name`) values (?, UPPER('John'))", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo"}, b.GetBindings())
}
func TestInsertGetIdWithEmptyValues(t *testing.T) {}
func TestInsertMethodRespectsRawBindings(t *testing.T) {
	b := GetBuilder()
	b.From("users").Insert(map[string]interface{}{
		"name":  goeloquent.Raw("UPPER('John')"),
		"email": "foo",
	})
	assert.Equal(t, "insert into `users` (`email`, `name`) values (?, UPPER('John'))", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo"}, b.GetBindings())
}
func TestMultipleInsertsWithExpressionValues(t *testing.T) {
	b := GetBuilder()
	b.From("users").Insert([]map[string]interface{}{
		{"name": goeloquent.Raw("UPPER('John')"), "email": goeloquent.Raw("UPPER('email1')")},
		{"name": goeloquent.Raw("UPPER('Jane')"), "email": goeloquent.Raw("UPPER('email2')")},
	})
	assert.Equal(t, "insert into `users` (`email`, `name`) values (UPPER('email1'), UPPER('John')), (UPPER('email2'), UPPER('Jane'))", b.ToSql())

	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}

func TestUpdateMethod(t *testing.T) {
	b := GetBuilder()
	b.From("users").Where("id", 1).Update(map[string]interface{}{
		"name":  "John",
		"email": "foo",
	})
	assert.Equal(t, "update `users` set `email` = ?, `name` = ? where `id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "John", 1}, b.GetBindings())

	b1 := GetBuilder()
	b1.From("users").Where("id", 1).OrderByDesc("age").Limit(5).Update(map[string]interface{}{
		"name":  "John",
		"email": "foo",
	})
	assert.Equal(t, "update `users` set `email` = ?, `name` = ? where `id` = ? order by `age` desc limit 5", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "John", 1}, b1.GetBindings())
}

func TestUpsertMethod(t *testing.T) {
	b := GetBuilder()
	b.From("users").Upsert([]map[string]interface{}{
		{"name": "John", "email": "foo"},
		{"name": "Jane", "email": "bar"},
	}, []string{"email"}, []string{"name"})
	assert.Equal(t, "insert into `users` (`email`, `name`) values (?, ?), (?, ?) on duplicate key update `name` = values(`name`)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "John", "bar", "Jane"}, b.GetBindings())

}

func TestUpsertMethodWithUpdateColumns(t *testing.T) {
	b := GetBuilder()
	b.From("users").Upsert([]map[string]interface{}{
		{"name": "John", "email": "foo"},
		{"name": "Jane", "email": "bar"},
	}, []string{"email"}, []string{"name"})
	assert.Equal(t, "insert into `users` (`email`, `name`) values (?, ?), (?, ?) on duplicate key update `name` = values(`name`)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "John", "bar", "Jane"}, b.GetBindings())
}

func TestUpdateMethodWithJoins(t *testing.T) {}

func TestUpdateMethodWithJoinsOnSqlServer(t *testing.T) {}

func TestUpdateMethodWithJoinsOnMySql(t *testing.T) {
	b := GetBuilder()
	b.From("users").Join("contacts", "users.id", "=", "contacts.user_id").Where("users.id", 1).Update(map[string]interface{}{
		"name":  "John",
		"email": "foo",
	})
	assert.Equal(t, "update `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` set `email` = ?, `name` = ? where `users`.`id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "John", 1}, b.GetBindings())

	b1 := GetBuilder()
	b1.From("users").Join("contacts", func(b *goeloquent.JoinBuilder) *goeloquent.JoinBuilder {
		b.On("users.id", "=", "contacts.user_id").Where("users.active", 1)
		return b
	}).Update(map[string]interface{}{
		"name":  "John",
		"email": "foo",
	})
	assert.Equal(t, "update `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` and `users`.`active` = ? set `email` = ?, `name` = ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo", "John"}, b1.GetBindings())
}
func TestUpdateMethodWithJoinsOnSQLite(t *testing.T)              {}
func TestUpdateMethodWithJoinsAndAliasesOnSqlServer(t *testing.T) {}
func TestUpdateMethodWithoutJoinsOnPostgres(t *testing.T)         {}
func TestUpdateMethodWithJoinsOnPostgres(t *testing.T)            {}
func TestUpdateFromMethodWithJoinsOnPostgres(t *testing.T)        {}
func TestUpdateMethodRespectsRaw(t *testing.T) {

	b := GetBuilder()
	b.From("users").Where("id", 1).Update(map[string]interface{}{
		"name":  goeloquent.Raw("UPPER('John')"),
		"email": "foo",
	})
	assert.Equal(t, "update `users` set `email` = ?, `name` = UPPER('John') where `id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", 1}, b.GetBindings())
}
func TestUpdateMethodWorksWithQueryAsValue(t *testing.T) {
	b := GetBuilder()
	b.From("users").Where("id", 1).Update(map[string]interface{}{
		"name": "John",
		"email": func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
			return builder.SelectRaw("sum(age)").From("contacts").WhereColumn("contacts.user_id", "users.id").Where("active", 1)
		},
	})
	assert.Equal(t, "update `users` set `email` = (select sum(age) from `contacts` where `contacts`.`user_id` = `users`.`id` and `active` = ?), `name` = ? where `id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "John", 1}, b.GetBindings())
}

func TestUpdateOrInsertMethod(t *testing.T) {
	before := "drop table if exists users;"
	after := "create table users (id int auto_increment primary key, name varchar(255), email varchar(255), active int)"

	RunWithDB(before+after, after, func(conn goeloquent.Connection) {
		b := conn.Query()
		b.From("users").UpdateOrInsert(map[string]interface{}{
			"email": "Jim",
		}, map[string]interface{}{
			"name": "John",
		})
		assert.Equal(t, "insert into `users` (`email`, `name`) values (?, ?)", b.ToSql())
		assert.ElementsMatch(t, []interface{}{"Jim", "John"}, b.GetBindings())

		b = conn.Query()
		b.From("users").UpdateOrInsert(map[string]interface{}{
			"email": "Jim",
		}, map[string]interface{}{
			"name": "Jim",
		})
		assert.Equal(t, "update `users` set `name` = ? where (`email` = ?) limit 1", b.ToSql())
		assert.ElementsMatch(t, []interface{}{"Jim", "Jim"}, b.GetBindings())

	})

}

func TestUpdateOrInsertMethodWorksWithEmptyUpdateValues(t *testing.T) {

}

func TestDeleteMethod(t *testing.T) {
	b := GetBuilder()
	b.From("users").Where("id", 1).Delete()
	assert.Equal(t, "delete from `users` where `id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())

	b1 := GetBuilder()
	b1.From("users").Delete([]interface{}{1})
	assert.Equal(t, "delete from `users` where `id` in (?)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())

	b2 := GetBuilder()
	b2.From("users").SelectRaw("?", []interface{}{"ignore"}).Delete([]interface{}{1})
	assert.Equal(t, "delete from `users` where `id` in (?)", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, goeloquent.PrepareBindsForDelete(b2.Bindings))

	b3 := GetBuilder()
	b3.From("users").Where("email", "foo").OrderBy("id").Limit(1).Delete()
	assert.Equal(t, "delete from `users` where `email` = ? order by `id` asc limit 1", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo"}, b3.GetBindings())

}

func TestDeleteWithJoinMethod(t *testing.T) {
	b := GetBuilder()
	b.From("users").Join("contacts", "users.id", "=", "contacts.user_id").Where("users.id", 1).OrderBy("users.id").Limit(1).Delete()
	assert.Equal(t, "delete `users` from `users` inner join `contacts` on `users`.`id` = `contacts`.`user_id` where `users`.`id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())

	b1 := GetBuilder()
	b1.From("users as A").Join("contacts as b", func(builder *goeloquent.JoinBuilder) {
		builder.On("a.id", "=", "b.user_id").Where("b.active", 1)
	}).OrderBy("id", "desc").Limit(1).Delete()
	assert.Equal(t, "delete `A` from `users` as `A` inner join `contacts` as `b` on `a`.`id` = `b`.`user_id` and `b`.`active` = ? ", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())
}

func TestTruncateMethod(t *testing.T) {
	b := GetBuilder()
	b.From("users").Truncate()
	assert.Equal(t, "truncate table `users`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}

func TestTruncateMethodWithPrefix(t *testing.T) {
	b := GetBuilder()
	b.Grammar.SetTablePrefix("prefix_")
	b.From("users").Truncate()
	assert.Equal(t, "truncate table `prefix_users`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}
func TestTruncateMethodWithPrefixAndSchema(t *testing.T) {
	b := GetBuilder()
	b.Grammar.SetTablePrefix("prefix_")
	b.From("myschema.users").Truncate()
	assert.Equal(t, "truncate table `myschema`.`prefix_users`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}
func TestPreserveAddsClosureToArray(t *testing.T) {
	b := GetBuilder()
	b.BeforeQuery(func(statement *goeloquent.Statement) {

	})
	assert.Equal(t, 1, len(b.BeforeQueryCallBacks))
	assert.IsType(t, goeloquent.StatementFunc(nil), b.BeforeQueryCallBacks[0])

	b1 := GetBuilder()
	b1.AfterQuery(func(statement *goeloquent.Statement) {

	})
	assert.Equal(t, 1, len(b1.AfterQueryCallBacks))
	assert.IsType(t, goeloquent.StatementFunc(nil), b1.AfterQueryCallBacks[0])
}

func TestApplyPreserveCleansArray(t *testing.T) {
	b := GetBuilder()
	b.BeforeQuery(func(statement *goeloquent.Statement) {

	})
	assert.Equal(t, 1, len(b.BeforeQueryCallBacks))
	b.ApplyBeforeQueryCallbacks()
	assert.Equal(t, 0, len(b.BeforeQueryCallBacks))

	b1 := GetBuilder()
	b1.AfterQuery(func(statement *goeloquent.Statement) {

	})
	assert.Equal(t, 1, len(b1.AfterQueryCallBacks))
	b1.ApplyAfterQueryCallbacks()
	assert.Equal(t, 0, len(b1.AfterQueryCallBacks))
}

func TestPreservedAreAppliedByToSql(t *testing.T) {
	b := GetBuilder()
	b.BeforeQuery(func(statement *goeloquent.Statement) {
		statement.Where("foo", "bar")
	})
	sql := b.Select().From("users").ToSql()
	assert.Equal(t, "select * from `users` where `foo` = ?", sql)
	assert.ElementsMatch(t, []interface{}{"bar"}, b.GetBindings())
}

func TestPreservedAreAppliedByInsert(t *testing.T) {
	b := GetBuilder()
	b.BeforeQuery(func(statement *goeloquent.Statement) {
		statement.From("users")
	})
	b.Insert(map[string]interface{}{
		"name":  "John",
		"email": "foo",
	})
	assert.Equal(t, "insert into `users` (`email`, `name`) values (?, ?)", b.RawSql)
	assert.ElementsMatch(t, []interface{}{"foo", "John"}, b.GetBindings())
}
func TestPreservedAreAppliedByInsertGetId(t *testing.T) {
	b := GetBuilder()
	b.BeforeQuery(func(statement *goeloquent.Statement) {
		statement.From("users")
	})
	_, err := b.InsertGetId(map[string]interface{}{
		"name":  "John",
		"email": "foo",
	})
	assert.Nil(t, err)
	assert.Equal(t, "insert into `users` (`email`, `name`) values (?, ?)", b.RawSql)
	assert.ElementsMatch(t, []interface{}{"foo", "John"}, b.GetBindings())
}

func TestPreservedAreAppliedByInsertUsing(t *testing.T) {
	b := GetBuilder()
	b.BeforeQuery(func(statement *goeloquent.Statement) {
		statement.From("users")
	})
	b.InsertUsing([]interface{}{"name", "email"}, func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.Select("name", "email").From("contacts").Where("active", 1)
	})
	assert.Equal(t, "insert into `users` (`name`, `email`) select `name`, `email` from `contacts` where `active` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
}
func TestPreservedAreAppliedByUpsert(t *testing.T) {
	b := GetBuilder()
	b.BeforeQuery(func(statement *goeloquent.Statement) {
		statement.From("users")
	})
	b.Upsert([]map[string]interface{}{
		{"name": "John", "email": "foo"},
		{"name": "Jane", "email": "bar"},
	}, []string{"email"}, []string{"name"})
	assert.Equal(t, "insert into `users` (`email`, `name`) values (?, ?), (?, ?) on duplicate key update `name` = values(`name`)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "John", "bar", "Jane"}, b.GetBindings())
}
func TestPreservedAreAppliedByUpdate(t *testing.T) {
	b := GetBuilder()
	b.BeforeQuery(func(statement *goeloquent.Statement) {
		statement.Where("id", 1)
	}).From("users")
	b.Update(map[string]interface{}{
		"name":  "John",
		"email": "foo",
	})
	assert.Equal(t, "update `users` set `email` = ?, `name` = ? where `id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"foo", "John", 1}, b.GetBindings())
}
func TestPreservedAreAppliedByDelete(t *testing.T) {
	b := GetBuilder()
	b.BeforeQuery(func(statement *goeloquent.Statement) {
		statement.Where("id", 1)
	}).From("users")
	b.Delete()
	assert.Equal(t, "delete from `users` where `id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
}
func TestPreservedAreAppliedByTruncate(t *testing.T) {
	b := GetBuilder()
	b.BeforeQuery(func(statement *goeloquent.Statement) {
		statement.From("users")
	})
	b.Truncate()
	assert.Equal(t, "truncate table `users`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}

func TestPreservedAreAppliedByExists(t *testing.T) {
	b := GetBuilder()
	b.BeforeQuery(func(statement *goeloquent.Statement) {
		statement.From("users")
	})
	var exists bool
	b.Exists(&exists)
	assert.Equal(t, "select exists(select * from `users`) as `exists`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}

func TestPostgresInsertGetId(t *testing.T) {}

func TestMySqlWrapping(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users")
	assert.Equal(t, "select * from `users`", b.ToSql())
}
func TestMySqlUpdateWrappingJson(t *testing.T) {
	after := "drop table if exists `users`;"
	before := after + "create table `users` (id int auto_increment primary key, name json, email varchar(255), active int);"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("users").Insert(map[string]interface{}{
			"name":   goeloquent.Raw("json_object('first_name', 'john', 'last_name','doe')"),
			"active": 1,
		})
		q, err := conn.Query().From("users").Where("active", 1).
			Update(map[string]interface{}{
				"name->first_name": "jane",
				"name->last_name":  "doe1",
			})
		assert.Nil(t, err)
		assert.Equal(t, "update `users` set `name` = json_set(`name`, '$.\"first_name\"', ?), `name` = json_set(`name`, '$.\"last_name\"', ?) where `active` = ?", q.ToSql())

		c, err := q.Result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, c, int64(1))
	})

}
func TestMySqlUpdateWrappingNestedJson(t *testing.T) {
	after := "drop table if exists `users`;"
	before := after + "create table `users` (id int auto_increment primary key, name json, email varchar(255), active int);"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("users").Insert(map[string]interface{}{
			"name":   goeloquent.Raw("json_object('meta', json_object('first_name', 'john', 'last_name', 'foo'))"),
			"active": 1,
		})
		q, err := conn.Query().From("users").Where("active", 1).
			Update(map[string]interface{}{
				"name->meta->last_name":  "baz",
				"name->meta->first_name": "foo",
			})
		assert.Nil(t, err)
		assert.Equal(t, "update `users` set `name` = json_set(`name`, '$.\"meta\".\"first_name\"', ?), `name` = json_set(`name`, '$.\"meta\".\"last_name\"', ?) where `active` = ?", q.ToSql())

		assert.ElementsMatch(t, []interface{}{"foo", "baz", 1}, q.GetBindings())
		c, err := q.Result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, c, int64(1))
	})
}
func TestMySqlUpdateWrappingJsonArray(t *testing.T) {

}

func TestMySqlUpdateWrappingJsonPathArrayIndex(t *testing.T) {

	after := "drop table if exists `users`;"
	before := "drop table if exists `users`;create table `users` (id int auto_increment primary key, options json,meta json, email varchar(255), active int);"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Table("users").Insert(map[string]interface{}{
			"active":  1,
			"options": goeloquent.Raw("json_array(json_object('2fa', true), json_object('2fa', true))"),
			"meta":    goeloquent.Raw("json_object('tags', json_array(json_array('foo', 'bar'), json_array('baz', 'qux')) )"),
		})
		q, err := conn.Query().From("users").Where("active", 1).
			Update(map[string]interface{}{
				"options->[1]->2fa": false,
				"meta->tags[0][2]":  "prime",
			})
		assert.Nil(t, err)

		assert.Equal(t, "update `users` set `meta` = json_set(`meta`, '$.\"tags\"[0][2]', ?), `options` = json_set(`options`, '$[1].\"2fa\"', false) where `active` = ?", q.RawSql)
		assert.ElementsMatch(t, []interface{}{"prime", 1}, q.GetBindings())
		c, err := q.Result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, c, int64(1))
	})
}

func TestMySqlUpdateWithJsonPreparesBindingsCorrectly(t *testing.T) {

	b := GetBuilder()
	b.From("users").Where("id", 1).Update(map[string]interface{}{
		"options->enabled": false,
		"updated_at":       "2025-01-01 00:00:00",
	})
	assert.Equal(t, "update `users` set `options` = json_set(`options`, '$.\"enabled\"', false), `updated_at` = ? where `id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"2025-01-01 00:00:00", 1}, b.GetBindings())

	b1 := GetBuilder()
	b1.From("users").Where("id", 0).Update(map[string]interface{}{
		"options->size": 43,
		"updated_at":    "2025-01-01 00:00:00",
	})
	assert.Equal(t, "update `users` set `options` = json_set(`options`, '$.\"size\"', ?), `updated_at` = ? where `id` = ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{43, "2025-01-01 00:00:00", 0}, b1.GetBindings())

	b2 := GetBuilder()
	b2.From("users").Where("id", 0).Update(map[string]interface{}{
		"options->size": goeloquent.Raw("43"),
	})
	assert.Equal(t, "update `users` set `options` = json_set(`options`, '$.\"size\"', 43) where `id` = ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{0}, b2.GetBindings())
}

func TestPostgresUpdateWrappingJson(t *testing.T) {
}

func TestPostgresUpdateWrappingJsonArray(t *testing.T) {
}
func TestPostgresUpdateWrappingJsonPathArrayIndex(t *testing.T) {
}
func TestSQLiteUpdateWrappingJsonArray(t *testing.T) {
}
func TestSQLiteUpdateWrappingNestedJsonArray(t *testing.T) {
}
func TestSQLiteUpdateWrappingJsonPathArrayIndex(t *testing.T) {
}
func TestMySqlWrappingJsonWithString(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("name->first_name", "John")
	assert.Equal(t, "select * from `users` where json_unquote(json_extract(`name`, '$.\"first_name\"')) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"John"}, b.GetBindings())
}
func TestMySqlWrappingJsonWithInteger(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("info->age", 30)
	assert.Equal(t, "select * from `users` where json_unquote(json_extract(`info`, '$.\"age\"')) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{30}, b.GetBindings())
}
func TestMySqlWrappingJsonWithDouble(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("info->height", 1.75)
	assert.Equal(t, "select * from `users` where json_unquote(json_extract(`info`, '$.\"height\"')) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1.75}, b.GetBindings())
}
func TestMySqlWrappingJsonWithBoolean(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("info->active", true)
	assert.Equal(t, "select * from `users` where json_extract(`info`, '$.\"active\"') = true", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}
func TestMySqlWrappingJsonWithBooleanAndIntegerThatLooksLikeOne(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("info->active", true).Where("item->active", false).Where("items->bool", "=", 0).Where("items->not", "=", 1)
	assert.Equal(t, "select * from `users` where json_extract(`info`, '$.\"active\"') = true and json_extract(`item`, '$.\"active\"') = false and json_unquote(json_extract(`items`, '$.\"bool\"')) = ? and json_unquote(json_extract(`items`, '$.\"not\"')) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{0, 1}, b.GetBindings())
}
func TestJsonPathEscaping(t *testing.T) {

}
func TestMySqlWrappingJson(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereRaw(`info->'$."details"' = 1`)
	assert.Equal(t, "select * from `users` where info->'$.\"details\"' = 1", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("users.items->price", "=", 1).OrderBy("items->price")
	assert.Equal(t, "select * from `users` where json_unquote(json_extract(`users`.`items`, '$.\"price\"')) = ? order by json_unquote(json_extract(`items`, '$.\"price\"')) asc", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").Where("items->price->in_user", "=", 1)
	assert.Equal(t, "select * from `users` where json_unquote(json_extract(`items`, '$.\"price\".\"in_user\"')) = ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b2.GetBindings())

	b3 := GetBuilder()
	b3.Select().From("users").Where("items->price->in_user", "=", 1).Where("items->age", 2)
	assert.Equal(t, "select * from `users` where json_unquote(json_extract(`items`, '$.\"price\".\"in_user\"')) = ? and json_unquote(json_extract(`items`, '$.\"age\"')) = ?", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2}, b3.GetBindings())
}
func TestPostgresWrappingJson(t *testing.T) {

}
func TestSqlServerWrappingJson(t *testing.T) {
}
func TestSqliteWrappingJson(t *testing.T) {
}
func TestSQLiteOrderBy(t *testing.T) {
}
func TestSqlServerLimitsAndOffsets(t *testing.T) {
}
func TestMySqlSoundsLikeOperator(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("name", "sounds like", "John")
	assert.Equal(t, "select * from `users` where `name` sounds like ?", b.ToSql())
}
func TestBitwiseOperators(t *testing.T) {
}
func TestMergeWheresCanMergeWheresAndBindings(t *testing.T) {

	b := GetBuilder()
	b.Select().From("users").Where("name", "=", "John")
	b.MergeWheres([]goeloquent.Where{
		{
			Type:     goeloquent.WhereTypeBasic,
			Column:   "email",
			Operator: "=",
			Value:    "test",
			Boolean:  "and",
		},
	}, []interface{}{"test"})
	assert.Equal(t, "select * from `users` where `name` = ? and `email` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"John", "test"}, b.GetBindings())
}
func TestPrepareValueAndOperator(t *testing.T) {
	//test PrepareParams instead
	b := GetBuilder()
	b.Where("name", "John")
	where := b.Wheres[0]
	assert.Equal(t, "name", where.Column)
	assert.Equal(t, "=", where.Operator)
	assert.Equal(t, "John", where.Value)
	assert.Equal(t, "and", where.Boolean)
	assert.Equal(t, goeloquent.WhereTypeBasic, where.Type)

	b.Where("email", "like", goeloquent.Raw("%@gmail.com"))
	where = b.Wheres[1]
	assert.Equal(t, "email", where.Column)
	assert.Equal(t, "like", where.Operator)
	assert.Equal(t, goeloquent.Raw("%@gmail.com"), where.Value)
	assert.Equal(t, "and", where.Boolean)
	assert.Equal(t, goeloquent.WhereTypeBasic, where.Type)

	b.OrWhere("role", "admin")
	where = b.Wheres[2]
	assert.Equal(t, "role", where.Column)
	assert.Equal(t, "=", where.Operator)
	assert.Equal(t, "admin", where.Value)
	assert.Equal(t, "or", where.Boolean)
	assert.Equal(t, goeloquent.WhereTypeBasic, where.Type)

	b.OrWhereNotIn("status", []interface{}{"banned", "inactive"})
	where = b.Wheres[3]
	assert.Equal(t, "status", where.Column)
	assert.Equal(t, []interface{}{"banned", "inactive"}, where.Values)
	assert.Equal(t, "or", where.Boolean)
	assert.Equal(t, goeloquent.WhereTypeNotIn, where.Type)

	b1 := GetBuilder()
	b1.WhereNull("deleted_at", goeloquent.Not)
	where1 := b1.Wheres[0]
	assert.Equal(t, "deleted_at", where1.Column)
	assert.Equal(t, goeloquent.WhereTypeNotNull, where1.Type)
	assert.Equal(t, "and", where1.Boolean)

	b2 := GetBuilder()
	b2.Table("users").WhereNull("deleted_at").Where("role", "in", []interface{}{"admin", "user"}, goeloquent.Or, goeloquent.Not).Where("tag", "in", []interface{}{"tag1", "tag2"})

	where2 := b2.Wheres[1]
	assert.Equal(t, "role", where2.Column)
	assert.Equal(t, []interface{}{"admin", "user"}, where2.Values)
	assert.Equal(t, "or", where2.Boolean)
	assert.Equal(t, goeloquent.WhereTypeNotIn, where2.Type)
	assert.Equal(t, "select * from `users` where `deleted_at` is null or `role` not in (?, ?) and `tag` in (?, ?)", b2.ToSql())

}
func TestPrepareValueAndOperatorExpectException(t *testing.T) {
}
func TestProvidingNullWithOperatorsBuildsCorrectly(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("name", nil)
	assert.Equal(t, "select * from `users` where `name` is null", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("bot", nil).Where("name", "!=", nil).Where("email", "<=>", nil, goeloquent.Or).Where("age", "<>", nil)
	assert.Equal(t, "select * from `users` where `bot` is null and `name` is not null and `email` is null and `age` is not null", b1.ToSql())
}
func TestDynamicWhere(t *testing.T) {
}
func TestDynamicWhereIsNotGreedy(t *testing.T) {
}
func TestCallTriggersDynamicWhere(t *testing.T) {
}
func TestBuilderThrowsExpectedExceptionWithUndefinedMethod(t *testing.T) {
}
func TestMySqlLock(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("foo", "bar").Lock()
	assert.Equal(t, "select * from `users` where `foo` = ? for update", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"bar"}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("foo", "bar").Lock(false)
	assert.Equal(t, "select * from `users` where `foo` = ? lock in share mode", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"bar"}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").Where("foo", "bar").Lock("lock in share mode")
	assert.Equal(t, "select * from `users` where `foo` = ? lock in share mode", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{"bar"}, b2.GetBindings())
}
func TestPostgresLock(t *testing.T) {
}
func TestSqlServerLock(t *testing.T) {
}
func TestSelectWithLockUsesWritePdo(t *testing.T) {
}
func TestBindingOrder(t *testing.T) {

	b := GetBuilder()
	b.Select().From("users").Join("second", func(b *goeloquent.JoinBuilder) *goeloquent.JoinBuilder {
		b.Where("deleted", 0)
		return b
	}).Where("active", 1).GroupBy("city").Having("ppl", ">", 7).OrderByRaw("match ('foo') against(?)", []interface{}{"bar"})
	assert.Equal(t, "select * from `users` inner join `second` on `deleted` = ? where `active` = ? group by `city` having `ppl` > ? order by match ('foo') against(?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{0, 1, 7, "bar"}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").OrderByRaw("match ('foo') against(?)", []interface{}{"bar"}).Having("ppl", ">", 7).GroupBy("city").Where("active", 1).Join("second", func(b *goeloquent.JoinBuilder) *goeloquent.JoinBuilder {
		b.Where("deleted", 0)
		return b
	})

	assert.Equal(t, "select * from `users` inner join `second` on `deleted` = ? where `active` = ? group by `city` having `ppl` > ? order by match ('foo') against(?)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{0, 1, 7, "bar"}, b1.GetBindings())
}
func TestAddBindingWithArrayMergesBindings(t *testing.T) {
	b := GetBuilder()
	b.AddBinding([]interface{}{"foo", "bar"}, goeloquent.COMPONENT_WHERE)
	b.AddBinding([]interface{}{"baz"}, goeloquent.COMPONENT_WHERE)
	assert.Equal(t, []interface{}{"foo", "bar", "baz"}, b.GetBindings())

}
func TestAddBindingWithArrayMergesBindingsInCorrectOrder(t *testing.T) {
	b := GetBuilder()
	b.AddBinding([]interface{}{"bar", "baz"}, goeloquent.COMPONENT_HAVING)
	b.AddBinding([]interface{}{"foo"}, goeloquent.COMPONENT_WHERE)
	assert.Equal(t, []interface{}{"foo", "bar", "baz"}, b.GetBindings())
}
func TestAddBindingWithEnum(t *testing.T) {

}
func TestMergeBuilders(t *testing.T) {
}
func TestMergeBuildersBindingOrder(t *testing.T) {
}
func TestSubSelect(t *testing.T) {
	b := GetBuilder()
	b.From("one").Select("foo", "bar").Where("k", "v").SelectSub(func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.From("two").Select("baz").Where("k2", "v2")
	}, "sub")
	assert.Equal(t, "select `foo`, `bar`, (select `baz` from `two` where `k2` = ?) as `sub` from `one` where `k` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"v2", "v"}, b.GetBindings())

}
func TestSubSelectResetBindings(t *testing.T) {
}
func TestSqlServerWhereDate(t *testing.T) {
}
func TestUppercaseLeadingBooleansAreRemoved(t *testing.T) {

	b := GetBuilder()
	b.Select().From("users").Where("active", "=", 1, "AND")
	assert.Equal(t, "select * from `users` where `active` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
}
func TestLowercaseLeadingBooleansAreRemoved(t *testing.T) {

	b := GetBuilder()
	b.Select().From("users").Where("active", "=", 1, "and")
	assert.Equal(t, "select * from `users` where `active` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
}
func TestCaseInsensitiveLeadingBooleansAreRemoved(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("active", "=", 1, "And")
	assert.Equal(t, "select * from `users` where `active` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
}
func TestTableValuedFunctionAsTableInSqlServer(t *testing.T) {

}
func TestChunkWithLastChunkComplete(t *testing.T) {

	before := "drop table if exists `users`;" + "create table `users` (id int auto_increment primary key, name varchar(255), status varchar(255));"
	after := "drop table if exists `users`;"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		b := conn.Query()
		var users []map[string]interface{}
		for i := 0; i < 10; i++ {
			var status string
			if i%2 == 0 {
				status = "active"
			} else {
				status = "inactive"
			}
			users = append(users, map[string]interface{}{
				"name":   "User" + strconv.Itoa(i),
				"status": status,
			})
		}
		r, err := b.Table("users").Insert(&users)
		assert.Nil(t, err)
		assert.Equal(t, int64(10), r.RowsAffected())

		var count int64
		b = conn.Query()
		_, err = b.Table("users").OrderBy("id").Where("status", "active").Count(&count)
		assert.Nil(t, err)
		assert.Equal(t, int64(5), count)
		var dest []map[string]interface{}
		count = 0
		b = conn.Query()
		stop, e := b.Where("status", "active").
			Where("id", ">", 1).
			OrderBy("id").
			From("users").
			Chunk(&dest, 2, func(dest interface{}) (bool, error) {
				items := *(dest.(*[]map[string]interface{}))
				count += int64(len(items))
				for _, user := range items {
					assert.Equal(t, "active", user["status"])
				}
				return false, nil
			})
		assert.Equal(t, int64(4), count)
		assert.False(t, stop)
		assert.Nil(t, e)

	})
}
func TestChunkWithLastChunkPartial(t *testing.T) {
	before := "drop table if exists `users`;" + "create table `users` (id int auto_increment primary key, name varchar(255), status varchar(255));"
	after := "drop table if exists `users`;"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		b := conn.Query()
		var users []map[string]interface{}
		for i := 0; i < 10; i++ {
			var status string
			if i%2 == 0 {
				status = "active"
			} else {
				status = "inactive"
			}
			users = append(users, map[string]interface{}{
				"name":   "User" + strconv.Itoa(i),
				"status": status,
			})
		}
		r, err := b.Table("users").Insert(&users)
		assert.Nil(t, err)
		assert.Equal(t, int64(10), r.RowsAffected())

		var count int64
		b = conn.Query()
		_, err = b.Table("users").OrderBy("id").Where("status", "active").Count(&count)
		assert.Nil(t, err)
		assert.Equal(t, int64(5), count)
		var dest []map[string]interface{}
		count = 0
		b = conn.Query()
		stop, e := b.Where("status", "active").
			OrderBy("id").
			From("users").
			Chunk(&dest, 2, func(dest interface{}) (bool, error) {
				items := *(dest.(*[]map[string]interface{}))
				count += int64(len(items))
				for _, user := range items {
					assert.Equal(t, "active", user["status"])
				}
				return false, nil
			})
		assert.Equal(t, int64(5), count)
		assert.False(t, stop)
		assert.Nil(t, e)

	})
}
func TestChunkCanBeStoppedByReturningFalse(t *testing.T) {
	before := "drop table if exists `users`;" + "create table `users` (id int auto_increment primary key, name varchar(255), status varchar(255));"
	after := "drop table if exists `users`;"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		b := conn.Query()
		var users []map[string]interface{}
		for i := 0; i < 10; i++ {
			var status string
			if i%2 == 0 {
				status = "active"
			} else {
				status = "inactive"
			}
			users = append(users, map[string]interface{}{
				"name":   "User" + strconv.Itoa(i),
				"status": status,
			})
		}
		r, err := b.Table("users").Insert(&users)
		assert.Nil(t, err)
		assert.Equal(t, int64(10), r.RowsAffected())

		var count int64
		b = conn.Query()
		_, err = b.Table("users").OrderBy("id").Count(&count)
		assert.Nil(t, err)
		assert.Equal(t, int64(10), count)
		var dest []map[string]interface{}
		count = 0
		b = conn.Query()
		stop, e := b.OrderBy("id").From("users").
			Chunk(&dest, 2, func(dest interface{}) (bool, error) {
				items := *(dest.(*[]map[string]interface{}))
				count += int64(len(items))

				if count > 3 {
					return true, nil
				}
				return false, nil
			})
		assert.Equal(t, int64(4), count)
		assert.True(t, stop)
		assert.Nil(t, e)

	})
}
func TestChunkWithCountZero(t *testing.T) {
	before := "drop table if exists `users`;" + "create table `users` (id int auto_increment primary key, name varchar(255), status varchar(255));"
	after := "drop table if exists `users`;"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		b := conn.Query()
		var users []map[string]interface{}
		for i := 0; i < 10; i++ {
			var status string
			if i%2 == 0 {
				status = "active"
			} else {
				status = "inactive"
			}
			users = append(users, map[string]interface{}{
				"name":   "User" + strconv.Itoa(i),
				"status": status,
			})
		}
		r, err := b.Table("users").Insert(&users)
		assert.Nil(t, err)
		assert.Equal(t, int64(10), r.RowsAffected())

		var count int64
		b = conn.Query()
		_, err = b.Table("users").OrderBy("id").Count(&count)
		assert.Nil(t, err)
		assert.Equal(t, int64(10), count)
		var dest []map[string]interface{}
		var called bool
		b = conn.Query()
		stopped, err := b.OrderBy("id").From("users").Where("status", nil).
			ChunkById(&dest, 2, "id", func(dest interface{}) (bool, error) {
				called = true
				return false, nil
			})
		assert.False(t, called)
		assert.Nil(t, err)
		assert.False(t, stopped)

	})
}
func TestChunkByIdOnArrays(t *testing.T) {

}
func TestChunkPaginatesUsingIdWithLastChunkComplete(t *testing.T) {
}
func TestChunkPaginatesUsingIdWithLastChunkPartial(t *testing.T) {
}
func TestChunkPaginatesUsingIdWithCountZero(t *testing.T) {
}
func TestChunkPaginatesUsingIdWithAlias(t *testing.T) {
}
func TestChunkPaginatesUsingIdDesc(t *testing.T) {
}
func TestPaginate(t *testing.T) {
	var perPage = 2
	var page = 3
	var columns = []string{"id", "name"}
	before := "drop table if exists `users`;" + "create table `users` (id int auto_increment primary key, name varchar(255), status varchar(255));"
	after := "drop table if exists `users`;"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		b := conn.Query()
		var users []map[string]interface{}
		for i := 0; i < 10; i++ {
			var status string
			if i%2 == 0 {
				status = "active"
			} else {
				status = "inactive"
			}
			users = append(users, map[string]interface{}{
				"name":   "User" + strconv.Itoa(i),
				"status": status,
			})
		}
		r, err := b.Table("users").Insert(&users)
		assert.Nil(t, err)
		assert.Equal(t, int64(10), r.RowsAffected())
		b = conn.Query()

		var res []map[string]interface{}
		var sqls []string
		var bindings []interface{}
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, params ...interface{}) bool {
			st := params[0].(*goeloquent.Statement)
			sqls = append(sqls, st.RawSql)
			bindings = append(bindings, st.GetBindings()...)
			return true
		})

		p, _, err := b.From("users").Where("id", ">", 1).Paginate(&res, int64(perPage), int64(page), columns)
		assert.Nil(t, err)
		assert.Equal(t, int64(9), p.Total)
		items := p.Items.(*[]map[string]interface{})
		assert.Equal(t, 2, len(*items))
		assert.ElementsMatch(t, []string{
			"select count(*) as aggregate from `users` where `id` > ?",
			"select `id`, `name` from `users` where `id` > ? limit 2 offset 4",
		}, sqls)
		assert.ElementsMatch(t, []interface{}{1, 1}, bindings)

	})
}
func TestPaginateWithDefaultArguments(t *testing.T) {

}
func TestPaginateWhenNoResults(t *testing.T) {
}
func TestPaginateWithSpecificColumns(t *testing.T) {
}
func TestPaginateWithTotalOverride(t *testing.T) {
}
func TestCursorPaginate(t *testing.T) {
}
func TestCursorPaginateMultipleOrderColumns(t *testing.T) {
}
func TestCursorPaginateWithDefaultArguments(t *testing.T) {
}
func TestCursorPaginateWhenNoResults(t *testing.T) {
}
func TestCursorPaginateWithSpecificColumns(t *testing.T) {
}
func TestCursorPaginateWithMixedOrders(t *testing.T) {
}
func TestCursorPaginateWithDynamicColumnInSelectRaw(t *testing.T) {
}
func TestCursorPaginateWithDynamicColumnWithCastInSelectRaw(t *testing.T) {
}
func TestCursorPaginateWithDynamicColumnInSelectSub(t *testing.T) {
}
func TestCursorPaginateWithUnionWheres(t *testing.T) {
}
func TestCursorPaginateWithMultipleUnionsAndMultipleWheres(t *testing.T) {
}
func TestCursorPaginateWithUnionMultipleWheresMultipleOrders(t *testing.T) {
}
func TestCursorPaginateWithUnionWheresWithRawOrderExpression(t *testing.T) {
}
func TestCursorPaginateWithUnionWheresReverseOrder(t *testing.T) {
}
func TestCursorPaginateWithUnionWheresMultipleOrders(t *testing.T) {
}
func TestCursorPaginateWithUnionWheresAndAliassedOrderColumns(t *testing.T) {
}
func TestWhereExpression(t *testing.T) {
}
func TestWhereRowValues(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereRowValues([]interface{}{"age", "year"}, "<", []interface{}{2, 3})
	assert.Equal(t, "select * from `users` where (`age`, `year`) < (?, ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{2, 3}, b.GetBindings())

	b = GetBuilder()
	b.From("users").Where("status", 1).OrWhereRowValues([]interface{}{"active", "inactive"}, ">", []interface{}{1, 0})
	assert.Equal(t, "select * from `users` where `status` = ? or (`active`, `inactive`) > (?, ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 1, 0}, b.GetBindings())

	b = GetBuilder()
	b.From("users").WhereRowValues([]interface{}{"age", "year"}, ">", []interface{}{2, goeloquent.Raw("3")})
	assert.Equal(t, "select * from `users` where (`age`, `year`) > (?, 3)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{2}, b.GetBindings())
}
func TestWhereRowValuesArityMismatch(t *testing.T) {
	b := GetBuilder()
	b.From("users").Where("status", 1).OrWhereRowValues([]interface{}{"active", "inactive"}, ">", []interface{}{1})
	assert.Equal(t, "", b.ToSql())
	assert.Equal(t, goeloquent.ErrorWhereRowValuesMismatch.Error(), b.GetError().Error())
}
func TestWhereJsonContainsMySql(t *testing.T) {
	b := GetBuilder()
	b.From("users").WhereJsonContains("options", []interface{}{"en"})

	assert.Equal(t, "select * from `users` where json_contains(`options`, ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"[\"en\"]"}, b.GetBindings())

	b = GetBuilder()
	b.From("users").WhereJsonContains("users.options->languages", []interface{}{"en"})
	assert.Equal(t, "select * from `users` where json_contains(`users`.`options`, ?, '$.\"languages\"')", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"[\"en\"]"}, b.GetBindings())

	b = GetBuilder()
	b.From("users").Where("id", 1).OrWhereJsonContains("options->languages", goeloquent.Raw("'[\"en\"]'"))
	assert.Equal(t, "select * from `users` where `id` = ? or json_contains(`options`, '[\"en\"]', '$.\"languages\"')", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())

}
func TestWhereJsonOverlapsMySql(t *testing.T) {
	b := GetBuilder()
	b.From("users").WhereJsonOverlaps("options", []interface{}{"en", "fr"})

	assert.Equal(t, "select * from `users` where json_overlaps(`options`, ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"[\"en\",\"fr\"]"}, b.GetBindings())

	b = GetBuilder()
	b.From("users").WhereJsonOverlaps("users.options->languages", []interface{}{"en", "fr"})
	assert.Equal(t, "select * from `users` where json_overlaps(`users`.`options`, ?, '$.\"languages\"')", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"[\"en\",\"fr\"]"}, b.GetBindings())

	b = GetBuilder()
	b.From("users").Where("id", 1).OrWhereJsonOverlaps("options->languages", goeloquent.Raw("'[\"en\",\"fr\"]'"))
	assert.Equal(t, "select * from `users` where `id` = ? or json_overlaps(`options`, '[\"en\",\"fr\"]', '$.\"languages\"')", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())

}
func TestWhereJsonContainsPostgres(t *testing.T) {
}
func TestWhereJsonContainsSqlite(t *testing.T) {
}
func TestWhereJsonContainsSqlServer(t *testing.T) {
}
func TestWhereJsonDoesntContainMySql(t *testing.T) {
	b := GetBuilder()
	b.From("users").WhereJsonDoesntContain("options", []interface{}{"en"})

	assert.Equal(t, "select * from `users` where not json_contains(`options`, ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"[\"en\"]"}, b.GetBindings())

	b = GetBuilder()
	b.From("users").WhereJsonDoesntContain("users.options->languages", []interface{}{"en"})
	assert.Equal(t, "select * from `users` where not json_contains(`users`.`options`, ?, '$.\"languages\"')", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"[\"en\"]"}, b.GetBindings())

	b = GetBuilder()
	b.From("users").Where("id", 1).OrWhereJsonDoesntContain("options->languages", goeloquent.Raw("'[\"en\"]'"))
	assert.Equal(t, "select * from `users` where `id` = ? or not json_contains(`options`, '[\"en\"]', '$.\"languages\"')", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
}
func TestWhereJsonDoesntOverlapMySql(t *testing.T) {
	b := GetBuilder()
	b.From("users").WhereJsonDoesntOverlap("options", []interface{}{"en", "fr"})
	assert.Equal(t, "select * from `users` where not json_overlaps(`options`, ?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"[\"en\",\"fr\"]"}, b.GetBindings())

	b = GetBuilder()
	b.From("users").WhereJsonDoesntOverlap("users.options->languages", []interface{}{"en", "fr"})
	assert.Equal(t, "select * from `users` where not json_overlaps(`users`.`options`, ?, '$.\"languages\"')", b.ToSql())
	assert.ElementsMatch(t, []interface{}{"[\"en\",\"fr\"]"}, b.GetBindings())

	b = GetBuilder()
	b.From("users").Where("id", 1).OrWhereJsonDoesntOverlap("options->languages", goeloquent.Raw("'[\"en\",\"fr\"]'"))
	assert.Equal(t, "select * from `users` where `id` = ? or not json_overlaps(`options`, '[\"en\",\"fr\"]', '$.\"languages\"')", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
}
func TestWhereJsonDoesntContainPostgres(t *testing.T) {
}
func TestWhereJsonDoesntContainSqlite(t *testing.T) {
}
func TestWhereJsonDoesntContainSqlServer(t *testing.T) {
}
func TestWhereJsonContainsKeyMySql(t *testing.T) {

	var res []map[string]interface{}
	b := GetBuilder()
	_, err := b.From("users").WhereJsonContainsKey("users.options->languages").Get(&res)
	assert.Nil(t, err)
	assert.Equal(t, "select * from `users` where ifnull(json_contains_path(`users`.`options`, 'one', '$.\"languages\"'), 0)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	_, err = b1.From("users").WhereJsonContainsKey("options->languages->primary").Get(&res)
	assert.Nil(t, err)
	assert.Equal(t, "select * from `users` where ifnull(json_contains_path(`options`, 'one', '$.\"languages\".\"primary\"'), 0)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())

	b2 := GetBuilder()
	_, err = b2.From("users").Where("id", 1).OrWhereJsonContainsKey("options->languages", goeloquent.Raw("'en'")).Get(&res)
	assert.Nil(t, err)
	assert.Equal(t, "select * from `users` where `id` = ? or ifnull(json_contains_path(`options`, 'one', '$.\"languages\"'), 0)", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b2.GetBindings())

	b3 := GetBuilder()
	_, err = b3.From("users").WhereJsonContainsKey("options->languages->primary").OrWhereJsonContainsKey("options->languages[0][1]").Get(&res)
	assert.Nil(t, err)
	assert.Equal(t, "select * from `users` where ifnull(json_contains_path(`options`, 'one', '$.\"languages\".\"primary\"'), 0) or ifnull(json_contains_path(`options`, 'one', '$.\"languages\"[0][1]'), 0)", b3.ToSql())
}
func TestWhereJsonContainsKeyPostgres(t *testing.T) {
}
func TestWhereJsonContainsKeySqlite(t *testing.T) {
}
func TestWhereJsonContainsKeySqlServer(t *testing.T) {
}
func TestWhereJsonDoesntContainKeyMySql(t *testing.T) {
	b := GetBuilder()
	b.From("users").WhereJsonDoesntContainKey("options->languages")

	assert.Equal(t, "select * from `users` where not ifnull(json_contains_path(`options`, 'one', '$.\"languages\"'), 0)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.From("users").WhereJsonDoesntContainKey("options->languages->primary")
	assert.Equal(t, "select * from `users` where not ifnull(json_contains_path(`options`, 'one', '$.\"languages\".\"primary\"'), 0)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())

	b2 := GetBuilder()
	b2.From("users").Where("id", 1).OrWhereJsonDoesntContainKey("options->languages", goeloquent.Raw("'en'"))
	assert.Equal(t, "select * from `users` where `id` = ? or not ifnull(json_contains_path(`options`, 'one', '$.\"languages\"'), 0)", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b2.GetBindings())
}
func TestWhereJsonDoesntContainKeyPostgres(t *testing.T) {
}
func TestWhereJsonDoesntContainKeySqlite(t *testing.T) {
}
func TestWhereJsonDoesntContainKeySqlServer(t *testing.T) {
}
func TestWhereJsonLengthMySql(t *testing.T) {
	b := GetBuilder()
	b.From("users").WhereJsonLength("options", 0)

	assert.Equal(t, "select * from `users` where json_length(`options`) = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{0}, b.GetBindings())

	b1 := GetBuilder()
	b1.From("users").WhereJsonLength("users.options->languages", ">", 3)
	assert.Equal(t, "select * from `users` where json_length(`users`.`options`, '$.\"languages\"') > ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{3}, b1.GetBindings())

	b2 := GetBuilder()
	b2.From("users").Where("id", 1).OrWhereJsonLength("options->languages", ">", goeloquent.Raw("4"))
	assert.Equal(t, "select * from `users` where `id` = ? or json_length(`options`, '$.\"languages\"') > 4", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b2.GetBindings())
}
func TestWhereJsonLengthPostgres(t *testing.T) {
}
func TestWhereJsonLengthSqlite(t *testing.T) {
}
func TestWhereJsonLengthSqlServer(t *testing.T) {
}
func TestFrom(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users")
	assert.Equal(t, "select * from `users`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users as u")
	assert.Equal(t, "select * from `users` as `u`", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select().From(goeloquent.Raw("users as u"))
	assert.Equal(t, "select * from users as u", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b2.GetBindings())

	b3 := GetBuilder()
	b3.Select().From("users", "u")
	assert.Equal(t, "select * from `users` as `u`", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b3.GetBindings())
}
func TestFromSub(t *testing.T) {
	b := GetBuilder()
	b.FromSub(func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.Select(goeloquent.Raw("max(last_seen_at) as last")).From("user_sessions").Where("active", 1)
	}, "sessions").Where("bar", ">", 1)
	assert.Equal(t, "select * from (select max(last_seen_at) as last from `user_sessions` where `active` = ?) as `sessions` where `bar` > ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 1}, b.GetBindings())

}
func TestFromSubWithPrefix(t *testing.T) {
	c := GetConnection()
	c.SetTablePrefix("goelo_")
	b := c.Query()

	b.FromSub(func(builder *goeloquent.QueryBuilder) *goeloquent.QueryBuilder {
		return builder.Select(goeloquent.Raw("max(last_seen_at) as last_seen")).From("user_sessions").Where("active", 1)
	}, "sessions").Where("bar", ">", 1)
	assert.Equal(t, "select * from (select max(last_seen_at) as last_seen from `goelo_user_sessions` where `active` = ?) as `sessions` where `bar` > ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 1}, b.GetBindings())
}
func TestFromSubWithoutBindings(t *testing.T) {

	b := GetBuilder()
	b.FromSub([]interface{}{"invalid"}, "sessions")
	assert.Equal(t, goeloquent.ErrorSubQueryInvalid.Error(), b.GetError().Error())
}
func TestFromRaw(t *testing.T) {
	b := GetBuilder()
	b.FromRaw("(select max(last_seen_at) as last from `user_sessions`) as `sessions`")
	assert.Equal(t, "select * from (select max(last_seen_at) as last from `user_sessions`) as `sessions`", b.ToSql())

}
func TestFromRawOnSqlServer(t *testing.T) {
}
func TestFromRawWithWhereOnTheMainQuery(t *testing.T) {
	b := GetBuilder()
	b.FromRaw("(select max(last_seen_at) as last from `user_sessions`) as `sessions`").Where("last_seen_at", ">", 1520652582)
	assert.Equal(t, "select * from (select max(last_seen_at) as last from `user_sessions`) as `sessions` where `last_seen_at` > ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1520652582}, b.GetBindings())
}
func TestFromQuestionMarkOperatorOnPostgres(t *testing.T) {
}
func TestUseIndexMySql(t *testing.T) {

	b := GetBuilder()
	b.Select().From("users").UseIndex("index_name")
	assert.Equal(t, "select * from `users` use index (index_name)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())

}
func TestForceIndexMySql(t *testing.T) {
	b1 := GetBuilder()
	b1.Select().From("users").ForceIndex("index_name")
	assert.Equal(t, "select * from `users` force index (index_name)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b1.GetBindings())
}
func TestIgnoreIndexMySql(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").IgnoreIndex("index_name")
	assert.Equal(t, "select * from `users` ignore index (index_name)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}
func TestUseIndexSqlite(t *testing.T) {
}
func TestForceIndexSqlite(t *testing.T) {
}
func TestIgnoreIndexSqlite(t *testing.T) {
}
func TestUseIndexSqlServer(t *testing.T) {
}
func TestForceIndexSqlServer(t *testing.T) {
}
func TestIgnoreIndexSqlServer(t *testing.T) {
}
func TestClone(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users")
	clone := goeloquent.Clone(b).Where("id", 1)
	assert.Equal(t, "select * from `users` where `id` = ?", clone.ToSql())
	assert.Equal(t, "select * from `users`", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, clone.GetBindings())
	assert.ElementsMatch(t, []interface{}{}, b.GetBindings())
}
func TestCloneWithout(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("id", 1).OrderBy("name")
	clone := b.CloneWithout([]goeloquent.Component{goeloquent.COMPONENT_ORDER}, []goeloquent.Component{})
	assert.Equal(t, "select * from `users` where `id` = ?", clone.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, clone.GetBindings())
	assert.Equal(t, "select * from `users` where `id` = ? order by `name` asc", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
}
func TestCloneWithoutBindings(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("id", 1).OrderBy("name")
	clone := b.CloneWithout([]goeloquent.Component{goeloquent.COMPONENT_WHERE}, []goeloquent.Component{goeloquent.COMPONENT_WHERE})
	assert.Equal(t, "select * from `users` order by `name` asc", clone.ToSql())
	assert.ElementsMatch(t, []interface{}{}, clone.GetBindings())
	assert.Equal(t, "select * from `users` where `id` = ? order by `name` asc", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
}
func TestToRawSql(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("id", 1).OrderBy("name")
	assert.Equal(t, "select * from `users` where `id` = 1 order by `name` asc", b.ToRawSql())
}
