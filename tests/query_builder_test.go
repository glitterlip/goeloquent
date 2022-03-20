package tests

import (
	goeloquent "github.com/glitterlip/go-eloquent"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

//test DatabaseQueryBuilderTest

func GetBuilder() *goeloquent.Builder {
	c := DB().Connection("default")
	builder := goeloquent.NewBuilder(c)
	builder.Grammar = &goeloquent.MysqlGrammar{}
	builder.Grammar.SetTablePrefix(c.Config.Prefix)
	builder.Grammar.SetBuilder(builder)
	return builder
}
func Connect() {
	defaultConfig := map[string]goeloquent.DBConfig{
		"default": getDefaultConfig(),
	}
	db = goeloquent.Open(defaultConfig)
	chatConfig := getChatConfig()
	db.AddConfig("chat", &chatConfig)
}
func TestBasicSelect(t *testing.T) {
	//testBasicSelect
	b := GetBuilder()
	b.Select("*").From("users")
	assert.Equal(t, "select * from `users`", b.ToSql())
}

func TestBasicSelectWithColumns(t *testing.T) {
	//testBasicSelectWithGetColumns
	b := GetBuilder()
	var dest []map[string]interface{}
	b.From("users").Get(&dest)
	assert.Equal(t, "select * from `users`", b.PreparedSql)

	b1 := goeloquent.CloneBuilder(b)
	b1.From("users").Select("id", "name").Get(&dest)
	assert.Equal(t, "select `id`, `name` from `users`", b1.PreparedSql)

	b2 := goeloquent.CloneBuilder(b)
	b2.From("users", "u").Get(&dest, "id", "name")
	assert.Equal(t, "select `id`, `name` from `users` as `u`", b2.PreparedSql)
}

//TODO:testBasicSelectUseWritePdo
func TestWrap(t *testing.T) {
	//testAliasWrappingAsWholeConstant
	b := GetBuilder()
	b.Select("x.y as foo.bar").From("baz")
	assert.Equal(t, "select `x`.`y` as `foo.bar` from `baz`", b.ToSql())

	//testAliasWrappingWithSpacesInDatabaseName
	b1 := goeloquent.CloneBuilder(b)
	b1.Select("w x.y.z as foo.bar").From("baz")
	assert.Equal(t, "select `w x`.`y`.`z` as `foo.bar` from `baz`", b1.ToSql())

	//testBasicTableWrapping
	b2 := goeloquent.CloneBuilder(b)
	b2.Select().From("public.users")
	assert.Equal(t, "select * from `public`.`users`", b2.ToSql())

	//testMySqlWrappingProtectsQuotationMarks
	b3 := GetBuilder()
	b3.Select("*").From("groups`users")
	assert.Equal(t, "select * from `groups``users`", b3.ToSql())

	//TODO:testBasicTableWrappingProtectsQuotationMarks
}

func TestAddSelect(t *testing.T) {
	//testAddingSelects
	b := GetBuilder()
	b.Select("foo").AddSelect("bar").AddSelect("baz", "boom").From("users")
	assert.Equal(t, "select `foo`, `bar`, `baz`, `boom` from `users`", b.ToSql())
}

func TestTablePrefix(t *testing.T) {
	//testBasicSelectWithPrefix
	b := GetBuilder()
	b.Grammar.SetTablePrefix("prefix_")
	b.From("users").Select()
	assert.Equal(t, "select * from `prefix_users`", b.ToSql())
}

func TestDistinct(t *testing.T) {
	//testBasicSelectDistinct
	b := GetBuilder()
	b.Distinct().Select("foo", "bar").From("users")
	assert.Equal(t, "select distinct `foo`, `bar` from `users`", b.ToSql())
	//TODO:testBasicSelectDistinctOnColumns
}
func TestAlias(t *testing.T) {
	//testBasicAlias
	b := GetBuilder()
	b.Select("foo as bar").From("users")
	assert.Equal(t, "select `foo` as `bar` from `users`", b.ToSql())

	//testAliasWithPrefix
	b1 := goeloquent.CloneBuilder(b)
	b1.Grammar.SetTablePrefix("prefix_")
	b1.Select("foo as bar", "baz").From("users as u")
	assert.Equal(t, "select `foo` as `bar`, `baz` from `prefix_users` as `prefix_u`", b1.ToSql())
	//TODO:testJoinAliasesWithPrefix

}

func TestWhenCallback(t *testing.T) {
	//testWhenCallback
	b := GetBuilder()
	cb := func(builder *goeloquent.Builder) {
		builder.Where("age", ">", 10)
	}
	b.Select("*").From("users").When(true, cb).Where("name", "John")
	assert.Equal(t, "select * from `users` where `age` > ? and `name` = ?", b.ToSql())

	b1 := goeloquent.CloneBuilder(b)
	b1.Select("*").From("users").When(false, cb).Where("name", "John")
	assert.Equal(t, "select * from `users` where `name` = ?", b1.ToSql())

	//testWhenCallbackWithDefault
	b2 := goeloquent.CloneBuilder(b)
	b3 := goeloquent.CloneBuilder(b)
	defaultCb := func(builder *goeloquent.Builder) {
		builder.Where("age", "<", 18)
	}
	b2.Select("*").From("users").When(false, cb, defaultCb).Where("name", "John")
	assert.Equal(t, "select * from `users` where `age` < ? and `name` = ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{18, "John"}, b2.GetBindings())
	assert.ElementsMatch(t, []interface{}{18, "John"}, b2.GetRawBindings()["where"])

	b3.Select("*").From("users").When(true, cb, defaultCb).Where("name", "John")
	assert.Equal(t, "select * from `users` where `age` > ? and `name` = ?", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{10, "John"}, b3.GetBindings())
	assert.ElementsMatch(t, []interface{}{10, "John"}, b3.GetRawBindings()["where"])
}
func TestBasicWheres(t *testing.T) {
	//testBasicWheres
	b := GetBuilder()
	b.Select().From("users").Where("id", 1)
	assert.Equal(t, "select * from `users` where `id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("id", "=", 1)
	assert.Equal(t, "select * from `users` where `id` = ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())

	//testRawWheres
	b2 := GetBuilder()
	b2.Select().From("users").WhereRaw("id = ? or email = ?", []interface{}{1, "foo"})
	assert.Equal(t, "select * from `users` where id = ? or email = ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b2.GetBindings())

	//testWhereShortcut
	b3 := GetBuilder().Select().From("users").Where("id", 1).OrWhere("name", "Jack")
	ShouldEqual(t, "select * from `users` where `id` = ? or `name` = ?", b3)
	ElementsShouldMatch(t, []interface{}{1, "Jack"}, b3.GetBindings())

	//testWhereWithArrayConditions testWheresWithArrayValue
	b4 := GetBuilder().Select().From("users").Where([][]interface{}{
		{"admin", "=", 1, goeloquent.BOOLEAN_AND},
		{"id", "<", 10, goeloquent.BOOLEAN_OR},                                           //or
		{"source", "=", "301"},                                                           //ommit boolean default and
		{"deleted", 0},                                                                   //ommit operator default "="
		{"role", "in", []interface{}{"admin", "manager", "owner"}},                       // where in
		{"age", "between", []interface{}{18, 60, 100}},                                   // where between
		{map[string]interface{}{"name": "Bob", "location": "NY"}, goeloquent.BOOLEAN_OR}, // where nested
		{func(builder *goeloquent.Builder) { //builder
			builder.WhereYear("created_at", "<", 2010).WhereColumn("first_name", "last_name").OrWhereNull("created_at")
		}},
		{goeloquent.Raw("year(birthday) < 1998")},       //expression
		{"suspend", goeloquent.Raw("'nodoublequotes'")}, //raw value
		{goeloquent.Where{
			Type:     goeloquent.CONDITION_TYPE_BASIC,
			Column:   "alias",
			Operator: "=",
			Value:    "boss",
			Boolean:  goeloquent.BOOLEAN_OR,
		}},
	})
	ShouldEqual(t, "select * from `users` where `admin` = ? or `id` < ? and `source` = ? and `deleted` = ? and `role` in (?,?,?) and `age` between ? and ? or (`name` = ? and `location` = ?) and (year(`created_at`) < ? and `first_name` = `last_name` or `created_at` is null) and year(birthday) < 1998 and `suspend` = 'nodoublequotes' or `alias` = ?", b4)
	ElementsShouldMatch(t, []interface{}{1, 10, "301", 0, "admin", "manager", "owner", 18, 60, "Bob", "NY", 2010}, b4.GetBindings())

	//testNestedWheres
	b5 := GetBuilder().Select().From("users").Where("email", "foo").OrWhere(func(builder *goeloquent.Builder) {
		builder.Where("name", "bar").Where("age", "=", 25)
	})
	ShouldEqual(t, "select * from `users` where `email` = ? or (`name` = ? and `age` = ?)", b5)
	ElementsShouldMatch(t, []interface{}{"foo", "bar", 25}, b5.GetBindings())
}

func TestDateBasedWheres(t *testing.T) {
	//testDateBasedWheresAcceptsTwoArguments testWhereDateMySql testWhereDayMySql testOrWhereDayMySql testWhereMonthMySql testOrWhereMonthMySql testWhereYearMySql testOrWhereYearMySql testWhereTimeMySql
	b := GetBuilder()
	b.Select().From("users").WhereDate("created_at", "2022-01-11")
	assert.Equal(t, "select * from `users` where date(`created_at`) = ?", b.ToSql())
	assert.Equal(t, []interface{}{"2022-01-11"}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").WhereDay("created_at", "11")
	assert.Equal(t, "select * from `users` where day(`created_at`) = ?", b1.ToSql())
	assert.Equal(t, []interface{}{"11"}, b1.GetRawBindings()["where"])

	b2 := GetBuilder()
	b2.Select().From("users").WhereMonth("created_at", 5)
	assert.Equal(t, "select * from `users` where month(`created_at`) = ?", b2.ToSql())
	assert.Equal(t, []interface{}{5}, b2.GetBindings())

	b3 := GetBuilder()
	b3.Select().From("users").WhereYear("created_at", 2022).WhereMonth("created_at", 7).WhereDate("created_at", "=", "21", goeloquent.BOOLEAN_OR)
	assert.Equal(t, "select * from `users` where year(`created_at`) = ? and month(`created_at`) = ? or date(`created_at`) = ?", b3.ToSql())
	assert.Equal(t, []interface{}{2022, 7, "21"}, b3.GetBindings())

	b4 := GetBuilder()
	b4.Select().From("users").WhereTime("created_at", "22:00").WhereMonth("created_at", 7).WhereDate("created_at", "=", "21", goeloquent.BOOLEAN_OR)
	assert.Equal(t, "select * from `users` where time(`created_at`) = ? and month(`created_at`) = ? or date(`created_at`) = ?", b4.ToSql())
	assert.Equal(t, []interface{}{"22:00", 7, "21"}, b4.GetBindings())

}

func TestOrWhereDate(t *testing.T) {
	//testDateBasedOrWheresAcceptsTwoArguments

	b4 := GetBuilder()
	b4.Select().From("users").Where("id", 1).WhereDate("created_at", "=", 1, goeloquent.BOOLEAN_OR)
	assert.Equal(t, "select * from `users` where `id` = ? or date(`created_at`) = ?", b4.ToSql())
	b5 := GetBuilder()
	b5.Select().From("users").Where("id", 1).WhereDay("created_at", "=", 1, goeloquent.BOOLEAN_OR)
	assert.Equal(t, "select * from `users` where `id` = ? or day(`created_at`) = ?", b5.ToSql())
	b6 := GetBuilder()
	b6.Select().From("users").Where("id", 1).WhereMonth("created_at", "=", 1, goeloquent.BOOLEAN_OR)
	assert.Equal(t, "select * from `users` where `id` = ? or month(`created_at`) = ?", b6.ToSql())
	b7 := GetBuilder()
	b7.Select().From("users").Where("id", 1).WhereYear("created_at", "=", 1, goeloquent.BOOLEAN_OR)
	assert.Equal(t, "select * from `users` where `id` = ? or year(`created_at`) = ?", b7.ToSql())
}
func TestWhereDateExpression(t *testing.T) {
	//testDateBasedWheresExpressionIsNotBound
	b8 := GetBuilder()
	b8.Select().From("users").WhereDate("created_at", goeloquent.Raw("NOW()")).Where("age", ">", 18)
	assert.Equal(t, "select * from `users` where date(`created_at`) = NOW() and `age` > ?", b8.ToSql())
	assert.ElementsMatch(t, []interface{}{18}, b8.GetBindings())
	assert.ElementsMatch(t, []interface{}{18}, b8.GetRawBindings()["where"])

	b9 := GetBuilder()
	b9.Select().From("users").WhereMonth("created_at", goeloquent.Raw("NOW()"))
	assert.Equal(t, "select * from `users` where month(`created_at`) = NOW()", b9.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b9.GetBindings())

	b10 := GetBuilder()
	b10.Select().From("users").WhereYear("created_at", goeloquent.Raw("NOW()"))
	assert.Equal(t, "select * from `users` where year(`created_at`) = NOW()", b10.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b9.GetRawBindings()["where"])

	b11 := GetBuilder()
	b11.Select().From("users").WhereDay("created_at", goeloquent.Raw("NOW()"))
	assert.Equal(t, "select * from `users` where day(`created_at`) = NOW()", b11.ToSql())
	assert.ElementsMatch(t, []interface{}{}, b11.GetBindings())
}

func TestWhereBetween(t *testing.T) {
	//testWhereBetweens
	b := GetBuilder()
	b.Select().From("users").WhereBetween("age", []interface{}{18, 30})
	assert.Equal(t, "select * from `users` where `age` between ? and ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{18, 30}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{18, 30}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select().From("users").WhereNotBetween("age", []interface{}{18, 30, 300})
	assert.Equal(t, "select * from `users` where `age` not between ? and ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{18, 30}, b1.GetBindings())
	assert.ElementsMatch(t, []interface{}{18, 30}, b1.GetRawBindings()["where"])

	b2 := GetBuilder()
	b2.Select().From("users").WhereBetween("age", []interface{}{18, 30}, goeloquent.BOOLEAN_AND, true)
	assert.Equal(t, "select * from `users` where `age` not between ? and ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{18, 30}, b2.GetBindings())
	assert.ElementsMatch(t, []interface{}{18, 30}, b2.GetRawBindings()["where"])

	b3 := GetBuilder()
	b3.Select().From("users").WhereBetween("age", []interface{}{18, 30}).OrWhere("admin", 1)
	assert.Equal(t, "select * from `users` where `age` between ? and ? or `admin` = ?", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{18, 30, 1}, b3.GetBindings())
	assert.ElementsMatch(t, []interface{}{18, 30, 1}, b3.GetRawBindings()["where"])

	b4 := GetBuilder()
	b4.Select().From("users").WhereBetween("age", []interface{}{10, goeloquent.Raw("30")}).Where("admin", 1)
	assert.Equal(t, "select * from `users` where `age` between ? and 30 and `admin` = ?", b4.ToSql())
	assert.ElementsMatch(t, []interface{}{10, 1}, b4.GetBindings())
	assert.ElementsMatch(t, []interface{}{10, 1}, b4.GetRawBindings()["where"])
}

func TestOrWhere(t *testing.T) {
	//testBasicOrWheres
	builder := GetBuilder()
	builder.Select().From("users").Where("id", 1).OrWhere("email", "foo")
	assert.Equal(t, "select * from `users` where `id` = ? or `email` = ?", builder.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, builder.GetBindings())

	//testRawOrWheres
	b2 := GetBuilder()
	b2.Select().From("users").Where("id", 1).OrWhereRaw("email = ?", []interface{}{"test@gmail.com"})
	assert.Equal(t, "select * from `users` where `id` = ? or email = ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "test@gmail.com"}, b2.GetBindings())

}

func TestWhereIn(t *testing.T) {
	//testBasicWhereIns
	b := GetBuilder()
	b.Select().From("users").WhereIn("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `id` in (?,?,?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2, 3}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("name", "Joe").OrWhereIn("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `name` = ? or `id` in (?,?,?)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"Joe", 1, 2, 3}, b1.GetBindings())

	//testBasicWhereNotIns
	b2 := GetBuilder()
	b2.Select().From("users").WhereNotIn("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `id` not in (?,?,?)", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2, 3}, b2.GetBindings())

	b3 := GetBuilder()
	b3.Select().From("users").Where("name", "Jim").OrWhereNotIn("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `name` = ? or `id` not in (?,?,?)", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{"Jim", 1, 2, 3}, b3.GetBindings())

	//testRawWhereIns
	b4 := GetBuilder()
	b4.Select().From("users").Where("name", "Jim").OrWhereNotIn("id", []interface{}{1, goeloquent.Raw("test"), 3})
	assert.Equal(t, "select * from `users` where `name` = ? or `id` not in (?,test,?)", b4.ToSql())
	assert.ElementsMatch(t, []interface{}{"Jim", 1, 3}, b4.GetBindings())
}

func TestWhereColumn(t *testing.T) {
	//testBasicWhereColumn
	b := GetBuilder()
	b.Select().From("users").WhereColumn("first_name", "last_name").OrWhereColumn("first_name", "middle_name")
	assert.Equal(t, "select * from `users` where `first_name` = `last_name` or `first_name` = `middle_name`", b.ToSql())
	assert.Equal(t, 0, len(b.GetBindings()))

	b1 := GetBuilder()
	b1.Select().From("users").Where("id", 1).WhereColumn("updated_at", ">", "created_at")
	assert.Equal(t, "select * from `users` where `id` = ? and `updated_at` > `created_at`", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())
}
func TestHavingAggregate(t *testing.T) {
	//testHavingAggregate
	//TODO:fixme
	//expected := "select count(*) as aggregate from (select (select `count(*)` from `videos` where `posts`.`id` = `videos`.`post_id`) as `videos_count` from `posts` having `videos_count` > ?) as `temp_table`"
	//b := GetBuilder()
	//var c int
	//subFunc := func(builder *goeloquent.Builder) {
	//	builder.From("videos").Select("count(*)").WhereColumn("posts.id", "=", "videos.post_id")
	//}
	//b.From("posts").SelectSub(subFunc, "video_count").Having("videos_count", ">", 1).Count(&c)
	//assert.Equal(t, expected, b.PreparedSql)
}
func TestSubSelectWhereIns(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereIn("id", func(builder *goeloquent.Builder) {
		builder.Select("id").From("users").Where("age", ">", 25).Limit(3)
	})
	expected := "select * from `users` where `id` in (select `id` from `users` where `age` > ? limit 3)"
	assert.Equal(t, expected, b.ToSql())
	assert.ElementsMatch(t, []interface{}{25}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").WhereNotIn("id", func(builder *goeloquent.Builder) {
		builder.Select("id").From("users").Where("age", ">", 25).Limit(3)
	})
	expected1 := "select * from `users` where `id` not in (select `id` from `users` where `age` > ? limit 3)"
	assert.Equal(t, expected1, b1.ToSql())
	assert.ElementsMatch(t, []interface{}{25}, b1.GetBindings())

}

func TestWhereNulls(t *testing.T) {
	//testBasicWhereNulls
	b := GetBuilder()
	b.Select().From("users").WhereNull("id")
	assert.Equal(t, "select * from `users` where `id` is null", b.ToSql())
	assert.Empty(t, b.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").Where("id", 0).OrWhereNull("id")
	assert.Equal(t, "select * from `users` where `id` = ? or `id` is null", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{0}, b2.GetBindings())

	//testArrayWhereNulls
	b3 := GetBuilder()
	b3.Select().From("users").WhereNull([]interface{}{"id", "deleted_at"})
	assert.Equal(t, "select * from `users` where `id` is null and `deleted_at` is null", b3.ToSql())

	b4 := GetBuilder()
	b4.Select().From("users").Where("id", "<", 0).OrWhereNull([]interface{}{"id", "email"})
	assert.Equal(t, "select * from `users` where `id` < ? or `id` is null or `email` is null", b4.ToSql())
	assert.ElementsMatch(t, []interface{}{0}, b4.GetBindings())

	//testBasicWhereNotNulls
	b5 := GetBuilder()
	b5.Select().From("users").WhereNotNull("id")
	assert.Equal(t, "select * from `users` where `id` is not null", b5.ToSql())
	assert.Empty(t, b5.GetBindings())

	b6 := GetBuilder()
	b6.Select().From("users").Where("id", 0).OrWhereNotNull("id")
	assert.Equal(t, "select * from `users` where `id` = ? or `id` is not null", b6.ToSql())
	assert.ElementsMatch(t, []interface{}{0}, b6.GetBindings())

	//testArrayWhereNotNulls
	b7 := GetBuilder()
	b7.Select().From("users").WhereNotNull([]interface{}{"id", "deleted_at"})
	assert.Equal(t, "select * from `users` where `id` is not null and `deleted_at` is not null", b7.ToSql())

	b8 := GetBuilder()
	b8.Select().From("users").Where("id", "<", 0).OrWhereNotNull([]interface{}{"id", "email"})
	assert.Equal(t, "select * from `users` where `id` < ? or `id` is not null or `email` is not null", b8.ToSql())
	assert.ElementsMatch(t, []interface{}{0}, b8.GetBindings())
}

func TestGroupBys(t *testing.T) {
	//testGroupBys
	b := GetBuilder()
	b.Select().From("users").GroupBy("email")
	assert.Equal(t, "select * from `users` group by `email`", b.ToSql())

	b1 := GetBuilder()
	b1.Select().From("users").GroupBy("email", "id")
	ShouldEqual(t, "select * from `users` group by `email`, `id`", b1)

	ShouldEqual(t, "select * from `users` group by DATE(created_at)",
		GetBuilder().Select().From("users").GroupBy(goeloquent.Raw("DATE(created_at)")))

	sql := GetBuilder().Select().From("users").GroupByRaw("DATE(created_at), ? DESC", []interface{}{"foo"})
	ShouldEqual(t, "select * from `users` group by DATE(created_at), ? DESC", sql)
	ElementsShouldMatch(t, []interface{}{"foo"}, sql.GetBindings())

	ElementsShouldMatch(t, []interface{}{"whereRawBinding", "groupByRawBinding", "havingRawBinding"},
		GetBuilder().HavingRaw("?", []interface{}{"havingRawBinding"}).GroupByRaw("?", []interface{}{"groupByRawBinding"}).WhereRaw("?", []interface{}{"whereRawBinding"}).GetBindings())
}

func TestOrderBys(t *testing.T) {
	//testOrderBys
	ShouldEqual(t, "select * from `users` order by `email` asc, `age` desc",
		GetBuilder().From("users").Select().OrderBy("email").OrderBy("age", "desc"))

	sql := GetBuilder().Select().From("users").OrderBy("email").OrderByRaw("`age` ? desc", []interface{}{"foo"})
	ShouldEqual(t, "select * from `users` order by `email` asc, `age` ? desc", sql)
	ElementsShouldMatch(t, []interface{}{"foo"}, sql.GetBindings())

	sql = GetBuilder().Select().From("users").OrderByDesc("email")
	ShouldEqual(t, "select * from `users` order by `email` desc", sql)

	//testReorder
	sql = GetBuilder().From("users").Select().OrderBy("email")
	ShouldEqual(t, "select * from `users` order by `email` asc", sql)
	sql.ReOrder()
	ShouldEqual(t, "select * from `users`", sql)

	sql = GetBuilder().From("users").Select().OrderBy("email")
	ShouldEqual(t, "select * from `users` order by `email` asc", sql)
	sql.ReOrder("name", "desc")
	ShouldEqual(t, "select * from `users` order by `name` desc", sql)

	sql = GetBuilder().Select().From("users").OrderByRaw("?", []interface{}{true})
	ElementsShouldMatch(t, []interface{}{true}, sql.GetBindings())

	sql.ReOrder()
	ElementsShouldMatch(t, []interface{}{}, sql.GetBindings())

	//testOrderBySubQueries //todo
}
func TestHavings(t *testing.T) {
	//testHavings
	ShouldEqual(t, "select * from `users` having `email` > ?", GetBuilder().From("users").Select().Having("email", ">", 1))
	sql := GetBuilder().From("users").Select().OrHaving("email", "=", "a@gmail.com").OrHaving("email", "=", "b@gmail.com")

	ShouldEqual(t, "select * from `users` having `email` = ? or `email` = ?", sql)

	sql = GetBuilder().From("users").Select().GroupBy("email").Having("email", ">", 1)
	ShouldEqual(t, "select * from `users` group by `email` having `email` > ?", sql)
	ElementsShouldMatch(t, []interface{}{1}, sql.GetBindings())

	ShouldEqual(t, "select `email` as `foo_email` from `users` having `foo_email` > ?", GetBuilder().From("users").Select("email as foo_email").Having("foo_email", ">", 1))

	ShouldEqual(t, "select `category`, count(*) as `total` from `item` where `department` = ? group by `category` having `total` > 3",
		GetBuilder().Select("category", goeloquent.Raw("count(*) as `total`")).From("item").Where("department", "popular").GroupBy("category").Having("total", ">", goeloquent.Raw("3")))

	//testHavingBetweens
	b := GetBuilder().Select().From("users").HavingBetween("id", []interface{}{1, 3, 5})
	ShouldEqual(t, "select * from `users` having `id` between ? and ?", b)
	ElementsShouldMatch(t, []interface{}{1, 3}, b.GetBindings())

	//testHavingShortcut
	ShouldEqual(t, "select * from `users` having `email` = ? or `email` = ?", GetBuilder().Select().From("users").Having("email", 1).OrHaving("email", 2))

	//testRawHavings
	ShouldEqual(t, "select * from `users` having user_foo < user_bar", GetBuilder().Select().From("users").HavingRaw("user_foo < user_bar"))
	ShouldEqual(t, "select * from `users` having `baz` = ? or user_foo < user_bar", GetBuilder().Select().From("users").Having("baz", "=", 1).OrHavingRaw("user_foo < user_bar"))
	ShouldEqual(t, "select * from `users` having `last_login_date` between ? and ? or user_foo < user_bar", GetBuilder().Select().From("users").HavingBetween("last_login_date", []interface{}{"2018-11-11", "2022-02-02"}).OrHavingRaw("user_foo < user_bar"))
}
func TestLimiAndOffsets(t *testing.T) {
	//testLimitsAndOffsets
	ShouldEqual(t, "select * from `users` limit 10 offset 5", GetBuilder().Select().From("users").Offset(5).Limit(10))

	ShouldEqual(t, "select * from `users` limit 0", GetBuilder().Select().From("users").Limit(0))

}

//TODO:testHavingFollowedBySelectGet
func TestForPage(t *testing.T) {
	//testForPage
	ShouldEqual(t, "select * from `users` limit 15 offset 15", GetBuilder().Select().From("users").ForPage(2, 15))
	ShouldEqual(t, "select * from `users` limit 15 offset 0", GetBuilder().Select().From("users").ForPage(0, 15))
	ShouldEqual(t, "select * from `users` limit 15 offset 0", GetBuilder().Select().From("users").ForPage(-2, 15))
	ShouldEqual(t, "select * from `users` limit 0 offset 0", GetBuilder().Select().From("users").ForPage(2, 0))
	ShouldEqual(t, "select * from `users` limit 0 offset 0", GetBuilder().Select().From("users").ForPage(0, 0))
	ShouldEqual(t, "select * from `users` limit 0 offset 0", GetBuilder().Select().From("users").ForPage(-2, 0))

	//TODO:testGetCountForPaginationWithBindings testGetCountForPaginationWithColumnAliases
}

func TestSubSelects(t *testing.T) {
	//TODO:testFullSubSelects
}
func TestWhereExists(t *testing.T) {
	//testWhereExists
	b := GetBuilder().Select().From("orders").WhereExists(func(builder *goeloquent.Builder) {
		builder.Select("*").From("products").Where("products.id", goeloquent.Raw("`orders`.`id`"))
	})
	ShouldEqual(t, "select * from `orders` where exists (select * from `products` where `products`.`id` = `orders`.`id`)", b)

	b1 := GetBuilder().Select().From("orders").WhereNotExists(func(builder *goeloquent.Builder) {
		builder.Select().From("products").Where("products.id", "=", goeloquent.Raw("`orders`.`id`"))
	})
	ShouldEqual(t, "select * from `orders` where not exists (select * from `products` where `products`.`id` = `orders`.`id`)", b1)

	b2 := GetBuilder().Select().From("orders").Where("id", 1).OrWhereExists(func(builder *goeloquent.Builder) {
		builder.Select().From("products").Where("products.id", goeloquent.Raw("`orders`.`id`"))
	})
	ShouldEqual(t, "select * from `orders` where `id` = ? or exists (select * from `products` where `products`.`id` = `orders`.`id`)", b2)

	b3 := GetBuilder().Select().From("orders").Where("id", 1).OrWhereNotExists(func(builder *goeloquent.Builder) {
		builder.Select().From("products").Where("products.id", goeloquent.Raw("`orders`.`id`"))
	})
	ShouldEqual(t, "select * from `orders` where `id` = ? or not exists (select * from `products` where `products`.`id` = `orders`.`id`)", b3)
}
func TestJoins(t *testing.T) {
	//testBasicJoins
	b := GetBuilder().Select().From("users").Join("contacts", "users.id", "=", "contacts.id")
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id`", b)

	b1 := GetBuilder().Select().From("users").Join("contacts", "users.id", "=", "contacts.id").LeftJoin("photos", "users.id", "=", "photos.id")
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` left join `photos` on `users`.`id` = `photos`.`id`", b1)

	b2 := GetBuilder().Select().From("users").LeftJoinWhere("photos", "users.id", "=", "bar").JoinWhere("photos", "users.id", "=", "foo")
	ShouldEqual(t, "select * from `users` left join `photos` on `users`.`id` = ? inner join `photos` on `users`.`id` = ?", b2)
	ElementsShouldMatch(t, []interface{}{"bar", "foo"}, b2.GetBindings())
	//testComplexJoin
	b3 := GetBuilder().Select().From("users").Join("contacts", func(builder *goeloquent.Builder) {
		builder.On("users.id", "=", "contacts.id").OrOn("users.name", "=", "contacts.name")
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` or `users`.`name` = `contacts`.`name`", b3)

	b4 := GetBuilder().Select().From("users").Join("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.Where("users.id", "=", "foo").OrWhere("users.name", "=", "bar")
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = ? or `users`.`name` = ?", b4)
	ElementsShouldMatch(t, []interface{}{"foo", "bar"}, b4.GetBindings())

	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = ? or `users`.`name` = ?", b4)
	ElementsShouldMatch(t, []interface{}{"foo", "bar"}, b4.GetBindings())
	//testJoinWhereNull
	b5 := GetBuilder().Select().From("users").Join("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").WhereNull("contacts.deleted_at")
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` and `contacts`.`deleted_at` is null", b5)
	b6 := GetBuilder().Select().From("users").Join("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").OrWhereNull("contacts.deleted_at")
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` or `contacts`.`deleted_at` is null", b6)
	//testJoinWhereNotNull
	b7 := GetBuilder().Select().From("users").Join("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").WhereNotNull("contacts.deleted_at")
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` and `contacts`.`deleted_at` is not null", b7)

	b8 := GetBuilder().Select().From("users").Join("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").OrWhereNotNull("contacts.deleted_at")
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` or `contacts`.`deleted_at` is not null", b8)

}
func TestJoinIn(t *testing.T) {
	//testJoinWhereIn
	b9 := GetBuilder().Select().From("users").Join("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").WhereIn("contacts.name", []interface{}{48, "baz", nil})
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` and `contacts`.`name` in (?,?,?)", b9)
	ElementsShouldMatch(t, []interface{}{48, "baz", nil}, b9.GetBindings())

	b10 := GetBuilder().Select().From("users").Join("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").OrWhereIn("contacts.name", []interface{}{48, "baz", nil})
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` or `contacts`.`name` in (?,?,?)", b10)
	ElementsShouldMatch(t, []interface{}{48, "baz", nil}, b10.GetBindings())

	//testJoinWhereInSubquery
	b11 := GetBuilder().Select().From("users").Join("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").WhereIn("contacts.name", GetBuilder().Select("name").From("contacts").Where("name", "baz"))
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` and `contacts`.`name` in (select `name` from `contacts` where `name` = ?)", b11)
	ElementsShouldMatch(t, []interface{}{"baz"}, b11.GetBindings())

	b12 := GetBuilder().Select().From("users").Join("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").OrWhereIn("contacts.name", GetBuilder().Select("name").From("contacts").Where("name", "baz"))
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` or `contacts`.`name` in (select `name` from `contacts` where `name` = ?)", b12)
	ElementsShouldMatch(t, []interface{}{"baz"}, b12.GetBindings())

	//testJoinWhereNotIn
	b13 := GetBuilder().Select().From("users").Join("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").WhereNotIn("contacts.name", []interface{}{48, "baz", nil})
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` and `contacts`.`name` not in (?,?,?)", b13)
	ElementsShouldMatch(t, []interface{}{48, "baz", nil}, b13.GetBindings())

	b14 := GetBuilder().Select().From("users").Join("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").OrWhereNotIn("contacts.name", []interface{}{48, "baz", nil})
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` or `contacts`.`name` not in (?,?,?)", b14)
	ElementsShouldMatch(t, []interface{}{48, "baz", nil}, b14.GetBindings())
}
func TestJoinNested(t *testing.T) {
	//testJoinsWithNestedConditions
	b15 := GetBuilder().Select().From("users").LeftJoin("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").Where(func(builder *goeloquent.Builder) {
			builder.Where("contacts.country", "=", "US").OrWhere("contacts.is_partner", "=", 1)
		})
	})
	ShouldEqual(t, "select * from `users` left join `contacts` on `users`.`id` = `contacts`.`id` and (`contacts`.`country` = ? or `contacts`.`is_partner` = ?)", b15)
	ElementsShouldMatch(t, []interface{}{"US", 1}, b15.GetBindings())
	b16 := GetBuilder().Select().From("users").LeftJoin("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").Where("contacts.is_active", "=", 1).OrOn(func(builder *goeloquent.Builder) {
			builder.OrWhere(func(inner *goeloquent.Builder) {
				inner.Where("contacts.country", "=", "UK").OrOn("contacts.type", "=", "users.type")
			}).Where(func(inner2 *goeloquent.Builder) {
				inner2.Where("contacts.country", "=", "US").OrWhereNull("contacts.is_partner")
			})
		})
	})

	ShouldEqual(t, "select * from `users` left join `contacts` on `users`.`id` = `contacts`.`id` and `contacts`.`is_active` = ? or ((`contacts`.`country` = ? or `contacts`.`type` = `users`.`type`) and (`contacts`.`country` = ? or `contacts`.`is_partner` is null))", b16)
	ElementsShouldMatch(t, []interface{}{1, "UK", "US"}, b16.GetBindings())
	//testJoinsWithAdvancedConditions
	b17 := GetBuilder().Select().From("users").LeftJoin("contacts", func(builder *goeloquent.Builder) {
		builder.On("users.id", "contacts.id").Where(func(builder2 *goeloquent.Builder) {
			builder2.Where("role", "admin").OrWhereNull("contacts.disabled").OrWhereRaw("year(contacts.created_at) = 2016")
		})
	})
	ShouldEqual(t, "select * from `users` left join `contacts` on `users`.`id` = `contacts`.`id` and (`role` = ? or `contacts`.`disabled` is null or year(contacts.created_at) = 2016)", b17)
	ElementsShouldMatch(t, []interface{}{"admin"}, b17.GetBindings())
}
func TestJoinSub(t *testing.T) {
	//testJoinsWithSubqueryCondition
	b := GetBuilder().Select().From("users").LeftJoin("contacts", func(builder *goeloquent.Builder) {
		builder.On("users.id", "contacts.id").WhereIn("contact_type_id", func(builder2 *goeloquent.Builder) {
			builder2.Select("id").From("contact_types").Where("category_id", "1").WhereNull("deleted_at")
		})
	})
	ShouldEqual(t, "select * from `users` left join `contacts` on `users`.`id` = `contacts`.`id` and `contact_type_id` in (select `id` from `contact_types` where `category_id` = ? and `deleted_at` is null)", b)
	ElementsShouldMatch(t, []interface{}{"1"}, b.GetBindings())

	b1 := GetBuilder().Select().From("users").LeftJoin("contacts", func(builder *goeloquent.Builder) {
		builder.On("users.id", "contacts.id").WhereExists(func(builder2 *goeloquent.Builder) {
			builder2.SelectRaw("1").From("contact_types").WhereRaw("contact_types.id = contacts.contact_type_id").Where("category_id", "1").WhereNull("deleted_at")
		})
	})
	ShouldEqual(t, "select * from `users` left join `contacts` on `users`.`id` = `contacts`.`id` and exists (select 1 from `contact_types` where contact_types.id = contacts.contact_type_id and `category_id` = ? and `deleted_at` is null)", b1)
	ElementsShouldMatch(t, []interface{}{"1"}, b1.GetBindings())
}
func TestJoinsWithAdvancedSubquery(t *testing.T) {
	//testJoinsWithAdvancedSubqueryCondition
	b := GetBuilder().Select().From("users").LeftJoin("contacts", func(builder *goeloquent.Builder) {
		builder.On("users.id", "contacts.id").WhereExists(func(builder2 *goeloquent.Builder) {
			builder2.SelectRaw("1").From("contact_types").
				WhereRaw("contact_types.id = contacts.contact_type_id").
				Where("category_id", "1").WhereNull("deleted_at").WhereIn("level_id", func(builder3 *goeloquent.Builder) {
				builder3.Select("id").From("levels").Where("is_active", true)
			})
		})
	})
	ShouldEqual(t, "select * from `users` left join `contacts` on `users`.`id` = `contacts`.`id` and exists (select 1 from `contact_types` where contact_types.id = contacts.contact_type_id and `category_id` = ? and `deleted_at` is null and `level_id` in (select `id` from `levels` where `is_active` = ?))", b)
	ElementsShouldMatch(t, []interface{}{"1", true}, b.GetBindings())
	//testJoinsWithNestedJoins
	b1 := GetBuilder().Select("users.id", "contacts.id", "contact_types.id").From("users").LeftJoin("contacts", func(builder2 *goeloquent.Builder) {
		builder2.On("users.id", "contacts.id").Join("contact_types", "contacts.contact_type_id", "=", "contact_types.id")
	})
	ShouldEqual(t, "select `users`.`id`, `contacts`.`id`, `contact_types`.`id` from `users` left join (`contacts` inner join `contact_types` on `contacts`.`contact_type_id` = `contact_types`.`id`) on `users`.`id` = `contacts`.`id`", b1)

	//testJoinsWithMultipleNestedJoins
	b2 := GetBuilder().Select("users.id", "contacts.id", "contact_types.id", "countrys.id", "planets.id").From("users").LeftJoin("contacts", func(builder2 *goeloquent.Builder) {
		builder2.On("users.id", "contacts.id").Join("contact_types", "contacts.contact_type_id", "=", "contact_types.id").
			LeftJoin("countrys", func(builde3 *goeloquent.Builder) {
				builde3.On("contacts.country", "=", "countrys.country").Join("planets", func(builder4 *goeloquent.Builder) {
					builder4.On("countrys.planet_id", "=", "planet.id").Where("planet.is_settled", "=", 1).Where("planet.population", ">=", 10000)
				})
			})
	})
	ShouldEqual(t, "select `users`.`id`, `contacts`.`id`, `contact_types`.`id`, `countrys`.`id`, `planets`.`id` from `users` left join (`contacts` inner join `contact_types` on `contacts`.`contact_type_id` = `contact_types`.`id` left join (`countrys` inner join `planets` on `countrys`.`planet_id` = `planet`.`id` and `planet`.`is_settled` = ? and `planet`.`population` >= ?) on `contacts`.`country` = `countrys`.`country`) on `users`.`id` = `contacts`.`id`", b2)
	ElementsShouldMatch(t, []interface{}{1, 10000}, b2.GetBindings())
	//testJoinsWithNestedJoinWithAdvancedSubqueryCondition
	b3 := GetBuilder().Select("users.id", "contacts.id", "contact_types.id").From("users").LeftJoin("contacts", func(builder *goeloquent.Builder) {
		builder.On("users.id", "contacts.id").Join("contact_types", "contacts.contact_type_id", "=", "contact_types.id").
			WhereExists(func(builder1 *goeloquent.Builder) {
				builder1.Select().From("countrys").WhereColumn("contacts.country", "=", "countrys.country").Join("planets", func(builder2 *goeloquent.Builder) {
					builder2.On("countrys.planet_id", "=", "planet.id").Where("planet.is_settled", "=", 1)
				}).Where("planet.population", ">=", 10000)

			})
	})
	ShouldEqual(t, "select `users`.`id`, `contacts`.`id`, `contact_types`.`id` from `users` left join (`contacts` inner join `contact_types` on `contacts`.`contact_type_id` = `contact_types`.`id`) on `users`.`id` = `contacts`.`id` and exists (select * from `countrys` inner join `planets` on `countrys`.`planet_id` = `planet`.`id` and `planet`.`is_settled` = ? where `contacts`.`country` = `countrys`.`country` and `planet`.`population` >= ?)", b3)
	ElementsShouldMatch(t, []interface{}{1, 10000}, b3.GetBindings())
	//testJoinSub
	b4 := GetBuilder().Select().From("users").JoinSub("select * from `contacts`", "sub", "users.id", "=", "sub.id")
	ShouldEqual(t, "select * from `users` inner join (select * from `contacts`) as `sub` on `users`.`id` = `sub`.`id`", b4)

	b5 := GetBuilder().Select().From("users").JoinSub(GetBuilder().Select().From("contacts"), "sub", "users.id", "=", "sub.id")
	ShouldEqual(t, "select * from `users` inner join (select * from `contacts`) as `sub` on `users`.`id` = `sub`.`id`", b5)
	c1 := GetBuilder().From("contacts").Select().Where("name", "foo")
	c2 := GetBuilder().From("contacts").Select().Where("name", "bar")
	b6 := GetBuilder().Select().From("users").JoinSub(c1, "sub1", "users.id", "=", 1, "inner", true).JoinSub(c2, "sub2", "users.id", "=", "sub2.user_id")
	expected := "select * from `users` inner join (select * from `contacts` where `name` = ?) as `sub1` on `users`.`id` = ? inner join (select * from `contacts` where `name` = ?) as `sub2` on `users`.`id` = `sub2`.`user_id`"
	ShouldEqual(t, expected, b6)
	ElementsShouldMatch(t, []interface{}{"foo", 1, "bar"}, b6.GetBindings())
	//testJoinSubWithPrefix
	b7 := GetBuilder()
	b7.Grammar.SetTablePrefix("prefix_")
	b7.Select().From("users").JoinSub("select * from `contacts`", "sub", "users.id", "=", "sub.id")
	ShouldEqual(t, "select * from `prefix_users` inner join (select * from `contacts`) as `prefix_sub` on `prefix_users`.`id` = `prefix_sub`.`id`", b7)

	//testLeftJoinSub
	b8 := GetBuilder().Select().From("users").LeftJoinSub(GetBuilder().Select().From("contacts"), "sub", "users.id", "=", "sub.id")
	ShouldEqual(t, "select * from `users` left join (select * from `contacts`) as `sub` on `users`.`id` = `sub`.`id`", b8)

	//testRightJoinSub
	b9 := GetBuilder().Select().From("users").RightJoinSub(GetBuilder().Select().From("contacts"), "sub", "users.id", "=", "sub.id")
	ShouldEqual(t, "select * from `users` right join (select * from `contacts`) as `sub` on `users`.`id` = `sub`.`id`", b9)
	//testRawExpressionsInSelect
	b10 := GetBuilder().Select(goeloquent.Raw("substr(foo,6)")).From("users")
	ShouldEqual(t, "select substr(foo,6) from `users`", b10)

	//testSubqueriesBindings
	f := GetBuilder()
	f.Select().From("users").Where("email", "=", func(builder *goeloquent.Builder) {
		builder.Select(goeloquent.Raw("max(id)")).From("users").Where("email", "=", "bar").
			OrderByRaw("email like ?", []interface{}{"%.com"}).GroupBy("id").Having("id", "=", 4)
	}).OrWhere("id", "=", "foo").GroupBy("id").Having("id", "=", 5)
	ElementsShouldMatch(t, []interface{}{"bar", 4, "%.com", "foo", 5}, f.GetBindings())

}
func TestFroms(t *testing.T) {
	//testFromSub
	b := GetBuilder().Select().FromSub(func(builder *goeloquent.Builder) {
		builder.Select(goeloquent.Raw("max(last_seen_at) as last_seen_at")).From("user_sessions").Where("foo", "=", 1)
	}, "sessions").Where("bar", "<", 10)
	ShouldEqual(t, "select * from (select max(last_seen_at) as last_seen_at from `user_sessions` where `foo` = ?) as `sessions` where `bar` < ?", b)
	ElementsShouldMatch(t, []interface{}{1, 10}, b.GetBindings())
	//testFromSubWithPrefix
	b1 := GetBuilder()
	b1.Grammar.SetTablePrefix("prefix_")
	b1.Select().FromSub(func(builder *goeloquent.Builder) {
		builder.Select(goeloquent.Raw("max(last_seen_at) as last_seen_at")).From("user_sessions").Where("foo", "=", 1)
	}, "sessions").Where("bar", "<", 10)
	ShouldEqual(t, "select * from (select max(last_seen_at) as last_seen_at from `prefix_user_sessions` where `foo` = ?) as `prefix_sessions` where `bar` < ?", b1)
	//testFromRaw
	b2 := GetBuilder().Select().FromRaw("(select max(last_seen_at) as last_seen_at from `user_sessions`) as `sessions`)")
	ShouldEqual(t, "select * from (select max(last_seen_at) as last_seen_at from `user_sessions`) as `sessions`)", b2)

	b3 := GetBuilder().Select().FromRaw("(select max(last_seen_at) as last_seen_at from `user_sessions`) as `sessions`)").Where("last_seen_at", ">", 1520652582)
	ShouldEqual(t, "select * from (select max(last_seen_at) as last_seen_at from `user_sessions`) as `sessions`) where `last_seen_at` > ?", b3)
	ElementsShouldMatch(t, []interface{}{1520652582}, b3.GetBindings())

}

func TestEmptyWhereIns(t *testing.T) {
	//testEmptyWhereIns
	b := GetBuilder().Select().From("users").WhereIn("id", []interface{}{})
	ShouldEqual(t, "select * from `users` where 0 = 1", b)
	b1 := GetBuilder().Select().From("users").Where("id", "=", 1).OrWhereIn("id", []interface{}{})
	ShouldEqual(t, "select * from `users` where `id` = ? or 0 = 1", b1)

}

func TestEmptyWhereNotIns(t *testing.T) {
	//testEmptyWhereNotIns
	b := GetBuilder().Select().From("users").WhereNotIn("id", []interface{}{})
	ShouldEqual(t, "select * from `users` where 1 = 1", b)
	b1 := GetBuilder().Select().From("users").Where("id", "=", 1).OrWhereNotIn("id", []interface{}{})
	ShouldEqual(t, "select * from `users` where `id` = ? or 1 = 1", b1)
}

func TestWhereBetweenColumns(t *testing.T) {
	//testWhereBetweenColumns
	b := GetBuilder().Select().From("users").WhereBetweenColumns("id", []interface{}{"users.created_at", "users.updated_at"})
	ShouldEqual(t, "select * from `users` where `id` between `users`.`created_at` and `users`.`updated_at`", b)
	assert.Empty(t, b.GetBindings())

	b1 := GetBuilder().Select().From("users").WhereBetweenColumns("id", []interface{}{"created_at", "updated_at"})
	ShouldEqual(t, "select * from `users` where `id` between `created_at` and `updated_at`", b1)
	assert.Empty(t, b1.GetBindings())

	b2 := GetBuilder().Select().From("users").WhereBetweenColumns("id", []interface{}{goeloquent.Raw("1"), goeloquent.Raw("2")})
	ShouldEqual(t, "select * from `users` where `id` between 1 and 2", b2)
	assert.Empty(t, b2.GetBindings())
}

func TestWhereColumnArray(t *testing.T) {
	//testArrayWhereColumn
	conditions := [][]interface{}{
		{"first_name", "last_name"},
		{"updated_at", ">", "created_at"},
	}
	b := GetBuilder().Select().From("users").WhereColumn(conditions)
	ShouldEqual(t, "select * from `users` where (`first_name` = `last_name` and `updated_at` > `created_at`)", b)
	assert.Empty(t, b.GetBindings())
}
func TestOrderByInvalidDirectionParam(t *testing.T) {
	//testOrderByInvalidDirectionParam
	assert.PanicsWithErrorf(t, "wrong order direction: asec", func() {
		GetBuilder().Select().From("users").OrderBy("age", "asec")
	}, "wrong order should panic with msg:[wrong order direction: asec]")

}
func TestNestedWhereBindings(t *testing.T) {
	//testNestedWhereBindings
	b := GetBuilder().Where("email", "=", "foo").Where(func(builder *goeloquent.Builder) {
		builder.SelectRaw("?", []interface{}{goeloquent.Expression("ignore")}).Where("name", "=", "bar")
	})
	ElementsShouldMatch(t, []interface{}{"foo", "bar"}, b.GetBindings())
}
func TestClone(t *testing.T) {
	//testClone
	b := GetBuilder().Select().From("users")
	cb := goeloquent.Clone(b)
	cb.Where("email", "foo")
	ShouldEqual(t, "select * from `users`", b)
	ShouldEqual(t, "select * from `users` where `email` = ?", cb)
	assert.NotEqual(t, b, cb)
}

func TestCloneWithout(t *testing.T) {
	//testCloneWithout
	b := GetBuilder().Select().From("users").Where("email", "foo").OrderBy("email")
	cb := goeloquent.CloneWithout(b, goeloquent.TYPE_ORDER)
	ShouldEqual(t, "select * from `users` where `email` = ? order by `email` asc", b)
	ShouldEqual(t, "select * from `users` where `email` = ?", cb)

}
func TestCloneWithoutBindings(t *testing.T) {
	//testCloneWithoutBindings
	b := GetBuilder().Select().From("users").Where("email", "foo").OrderBy("email")
	cb := b.Clone().CloneWithout(goeloquent.TYPE_WHERE).CloneWithoutBindings(goeloquent.TYPE_WHERE)
	ShouldEqual(t, "select * from `users` where `email` = ? order by `email` asc", b)
	ElementsShouldMatch(t, []interface{}{"foo"}, b.GetBindings())

	ShouldEqual(t, "select * from `users` order by `email` asc", cb)
	ElementsShouldMatch(t, []interface{}{}, cb.GetBindings())
}
func TestCrossJoins(t *testing.T) {
	//testCrossJoins
	b := GetBuilder().Select().From("sizes").CrossJoin("colors")
	ShouldEqual(t, "select * from `sizes` cross join `colors`", b)
	b1 := GetBuilder().Select("*").From("tableB").Join("tableA", "tableA.column1", "=", "tableB.column2", goeloquent.JOIN_TYPE_CROSS)
	ShouldEqual(t, "select * from `tableB` cross join `tableA` on `tableA`.`column1` = `tableB`.`column2`", b1)
	b2 := GetBuilder().Select().From("tableB").CrossJoin("tableA", "tableA.column1", "=", "tableB.column2")
	ShouldEqual(t, "select * from `tableB` cross join `tableA` on `tableA`.`column1` = `tableB`.`column2`", b2)
}
func TestCrossJoinSubs(t *testing.T) {
	//testCrossJoinSubs
	b := GetBuilder().SelectRaw("(sale / overall.sales) * 100 AS percent_of_total").From("sales").CrossJoinSub(GetBuilder().SelectRaw("SUM(sale) AS sales").From("sales"), "overall")
	ShouldEqual(t, "select (sale / overall.sales) * 100 AS percent_of_total from `sales` cross join (select SUM(sale) AS sales from `sales`) as `overall`", b)
}
func TestInsertMethod(t *testing.T) {
	//testInsertMethod
	create := `
CREATE TABLE "users" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "name" varchar(255) NOT NULL,
  "age" tinyint(10) unsigned NOT NULL DEFAULT '0',
  "created_at" datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  "deleted_at" datetime DEFAULT NULL,
  PRIMARY KEY ("id")
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4;
`
	Connect()
	DB().Raw("default").Exec(strings.ReplaceAll(create, `"`, "`"))
	b := DB().Table("users")
	now := time.Now()
	insert, err := b.Insert(map[string]interface{}{
		"name":       "go-eloquent",
		"age":        18,
		"created_at": now,
	})
	assert.Nil(t, err)
	c, err := insert.RowsAffected()
	assert.Nil(t, err)
	assert.Equal(t, c, int64(1))
	ElementsShouldMatch(t, []interface{}{"go-eloquent", 18, now}, b.GetBindings())
	assert.Equal(t, "insert into `users` (`name`, `age`, `created_at`) values (?, ?, ?)", b.PreparedSql)
	DB().Raw("default").Exec("DROP TABLE IF EXISTS users")

	//testMySqlInsertOrIgnoreMethod
	DB().Raw("default").Exec(strings.ReplaceAll(create, `"`, "`"))
	b = DB().Table("users")
	insert, err = b.InsertOrIgnore(map[string]interface{}{
		"name":       "go-eloquent",
		"age":        18,
		"created_at": now,
	})
	assert.Nil(t, err)
	c, err = insert.RowsAffected()
	assert.Nil(t, err)
	assert.Equal(t, c, int64(1))
	ElementsShouldMatch(t, []interface{}{"go-eloquent", 18, now}, b.GetBindings())
	assert.Equal(t, "insert ignore into `users` (`name`, `age`, `created_at`) values (?, ?, ?)", b.PreparedSql)
	DB().Raw("default").Exec("DROP TABLE IF EXISTS users")

}
func TestAggregate(t *testing.T) {
	//testAggregateFunctions
}
