package tests

import (
	"errors"
	"fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func GetBuilder() *goeloquent.Builder {
	return goeloquent.NewQueryBuilder(DB.Connection("default"))
}

func TestBasicSelect(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users")
	assert.Equal(t, "select * from `users`", b.ToSql())
	b1 := GetBuilder()
	b1.Select("*").From("users")
	assert.Equal(t, "select * from `users`", b1.ToSql())
	b2 := GetBuilder()
	b2.Select("id", "name").From("users")
	assert.Equal(t, "select `id`, `name` from `users`", b2.ToSql())
	b3 := GetBuilder()
	b3.Select([]string{"id", "name"}).From("users")
	assert.Equal(t, "select `id`, `name` from `users`", b3.ToSql())
}

func TestBasicSelectWithGetColumns(t *testing.T) {
	b := GetBuilder()
	var dest []map[string]interface{}
	b.From("users")
	assert.Equal(t, "select * from `users`", b.ToSql())

	b1 := GetBuilder()
	b1.From("users").Get(&dest, "id", "name")
	assert.Equal(t, "select `id`, `name` from `users`", b1.PreparedSql)

	b2 := GetBuilder()
	b2.From("users").Get(&dest, []string{"id", "name"})
	assert.Equal(t, "select `id`, `name` from `users`", b2.PreparedSql)

	b3 := GetBuilder()
	b3.From("users").Get(&dest, "id")
	assert.Equal(t, "select `id` from `users`", b3.PreparedSql)
}

// TODO:testBasicSelectUseWritePdo
func TestAliasWrappingAsWholeConstant(t *testing.T) {
	b := GetBuilder()
	b.Select("x.y as foo.bar").From("baz")
	assert.Equal(t, "select `x`.`y` as `foo.bar` from `baz`", b.ToSql())
}
func TestAliasWrappingWithSpacesInDatabaseName(t *testing.T) {
	b1 := GetBuilder()
	b1.Select("w x.y.z as foo.bar").From("baz")
	assert.Equal(t, "select `w x`.`y`.`z` as `foo.bar` from `baz`", b1.ToSql())
}
func TestAddingSelects(t *testing.T) {
	b := GetBuilder()
	b.Select("foo").AddSelect("bar").AddSelect("baz", "boom").AddSelect("bar").From("users")
	//TODO consider add a columns map to eliminate duplicate columns
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
	cb := func(builder *goeloquent.Builder) {
		builder.Where("id", "=", 1)
	}
	b.Select("*").From("users").When(true, cb).Where("email", "foo")
	assert.Equal(t, "select * from `users` where `id` = ? and `email` = ?", b.ToSql())

	b1 := GetBuilder()
	b1.Select("*").From("users").When(false, cb).Where("email", "foo")
	assert.Equal(t, "select * from `users` where `email` = ?", b1.ToSql())
}
func TestWhenCallbackWithDefault(t *testing.T) {
	cb := func(builder *goeloquent.Builder) {
		builder.Where("id", "=", 1)
	}
	b2 := GetBuilder()
	b3 := GetBuilder()
	defaultCb := func(builder *goeloquent.Builder) {
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
	cb := func(builder *goeloquent.Builder) *goeloquent.Builder {
		return builder.Where("id", "=", 1)
	}

	b.Select("*").From("users").Tap(cb).Where("email", "foo")
	assert.Equal(t, "select * from `users` where `id` = ? and `email` = ?", b.ToSql())

}
func TestBasicWheres(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("id", 1)
	assert.Equal(t, "select * from `users` where `id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("id", "=", 1)
	assert.Equal(t, "select * from `users` where `id` = ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())

}
func TestBasicWhereNot(t *testing.T) {
	b1 := GetBuilder()
	b1.Select().From("users").WhereNot("name", "foot").WhereNot("name", "<>", "bar")
	assert.Equal(t, "select * from `users` where not `name` = ? and not `name` <> ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"foot", "bar"}, b1.GetBindings())
}
func TestWheresWithArrayValue(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `id` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("id", "=", []interface{}{1, 2, 3}, goeloquent.BOOLEAN_OR)
	assert.Equal(t, "select * from `users` where `id` = ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b1.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").Where("id", "!=", []interface{}{1, 2, 3}, goeloquent.BOOLEAN_OR)
	assert.Equal(t, "select * from `users` where `id` != ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b2.GetBindings())

	b3 := GetBuilder()
	b3.Select().From("users").Where("id", "<>", []interface{}{1, 2, 3}, goeloquent.BOOLEAN_OR)
	assert.Equal(t, "select * from `users` where `id` <> ?", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b3.GetBindings())
}

// testBasicTableWrappingProtectsQuotationMarks
func TestMySqlWrappingProtectsQuotationMarks(t *testing.T) {
	b3 := GetBuilder()
	b3.Select("*").From("some`table")
	assert.Equal(t, "select * from `some``table`", b3.ToSql())
}
func TestDateBasedWheresAcceptsTwoArguments(t *testing.T) {
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

	//TestWhereTimeOperatorOptionalMySql
	b4 := GetBuilder()
	b4.Select().From("users").WhereTime("created_at", "22:00").WhereMonth("created_at", 7).WhereDate("created_at", "=", "21", goeloquent.BOOLEAN_OR)
	assert.Equal(t, "select * from `users` where time(`created_at`) = ? and month(`created_at`) = ? or date(`created_at`) = ?", b4.ToSql())
	assert.Equal(t, []interface{}{"22:00", 7, "21"}, b4.GetBindings())

}

func TestDateBasedOrWheresAcceptsTwoArguments(t *testing.T) {

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

func TestWhereNested(t *testing.T) {

	m := []map[string]interface{}{}
	b := GetBuilder()
	b.From("users").WhereNested([][]interface{}{
		{"name", "foo"},
		{"email", "=", "bar", goeloquent.BOOLEAN_OR},
	}).WhereNested([][]interface{}{
		{"age", 18},
		{"admin", 1},
		{"id", "in", []interface{}{1, 2, 3}},
	}, goeloquent.BOOLEAN_OR).WhereNested(func(builder *goeloquent.Builder) *goeloquent.Builder {
		return builder.WhereNull("deleted_at")
	}).Pretend().Get(&m)
	assert.Equal(t, "select * from `users` where (`name` = ? or `email` = ?) or (`age` = ? and `admin` = ? and `id` in (?,?,?)) and (`deleted_at` is null)", b.ToSql())
	assert.Equal(t, []interface{}{"foo", "bar", 18, 1, 1, 2, 3}, b.GetBindings())
}
func TestDateBasedWheresExpressionIsNotBound(t *testing.T) {
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

func TestWhereBetweens(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereBetween("id", []interface{}{1, 2})
	assert.Equal(t, "select * from `users` where `id` between ? and ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2}, b.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, 2}, b.GetRawBindings()["where"])

	b1 := GetBuilder()
	b1.Select().From("users").WhereNotBetween("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `id` not between ? and ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2}, b1.GetBindings())
	assert.ElementsMatch(t, []interface{}{1, 2}, b1.GetRawBindings()["where"])

	b2 := GetBuilder()
	b2.Select().From("users").WhereBetween("id", []interface{}{goeloquent.Raw("1"), goeloquent.Raw("2")}, goeloquent.BOOLEAN_AND, true)
	assert.Equal(t, "select * from `users` where `id` not between 1 and 2", b2.ToSql())
	assert.Empty(t, b2.GetBindings())
	assert.Empty(t, b2.GetRawBindings()["where"])

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
func TestOrWhereBetween(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("id", 1).OrWhereBetween("age", []interface{}{18, 30})
	assert.Equal(t, "select * from `users` where `id` = ? or `age` between ? and ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 18, 30}, b.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").Where("id", 1).OrWhereBetween("age", []interface{}{goeloquent.Raw("18"), goeloquent.Raw("30")}, goeloquent.BOOLEAN_AND, true)
	assert.Equal(t, "select * from `users` where `id` = ? or `age` not between 18 and 30", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1}, b2.GetBindings())

	b3 := GetBuilder()
	b3.Select().From("users").Where("id", 1).OrWhereBetween("age", []interface{}{10, goeloquent.Raw("30")}).Where("admin", 1)
	assert.Equal(t, "select * from `users` where `id` = ? or `age` between ? and 30 and `admin` = ?", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 10, 1}, b3.GetBindings())

}
func TestOrWhereNotBetween(t *testing.T) {
	b1 := GetBuilder()
	b1.Select().From("users").Where("id", 1).OrWhereNotBetween("age", []interface{}{18, 30})
	assert.Equal(t, "select * from `users` where `id` = ? or `age` not between ? and ?", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 18, 30}, b1.GetBindings())
	b3 := GetBuilder()
	b3.Select().From("users").Where("id", 1).OrWhereNotBetween("age", []interface{}{10, goeloquent.Raw("30")}).Where("admin", 1)
	assert.Equal(t, "select * from `users` where `id` = ? or `age` not between ? and 30 and `admin` = ?", b3.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 10, 1}, b3.GetBindings())
}
func TestWhereBetweenColumns(t *testing.T) {
	b := GetBuilder().Select().From("users").WhereBetweenColumns("id", []interface{}{"users.created_at", "users.updated_at"})
	ShouldEqual(t, "select * from `users` where `id` between `users`.`created_at` and `users`.`updated_at`", b)
	assert.Empty(t, b.GetBindings())

	b1 := GetBuilder().Select().From("users").WhereNotBetweenColumns("id", []interface{}{"created_at", "updated_at"})
	ShouldEqual(t, "select * from `users` where `id` between `created_at` and `updated_at`", b1)
	assert.Empty(t, b1.GetBindings())

	b2 := GetBuilder().Select().From("users").WhereBetweenColumns("id", []interface{}{goeloquent.Raw("1"), goeloquent.Raw("2")})
	ShouldEqual(t, "select * from `users` where `id` between 1 and 2", b2)
	assert.Empty(t, b2.GetBindings())
}
func TestOrWhereBetweenColumns(t *testing.T) {
	b := GetBuilder().Select().From("users").Where("id", 1).OrWhereBetweenColumns("age", []interface{}{"users.created_at", "users.updated_at"})
	ShouldEqual(t, "select * from `users` where `id` = ? or `age` between `users`.`created_at` and `users`.`updated_at`", b)
	ElementsShouldMatch(t, []interface{}{1}, b.GetBindings())

	b1 := GetBuilder().Select().From("users").Where("id", 1).OrWhereBetweenColumns("age", []interface{}{goeloquent.Raw("1"), goeloquent.Raw("2")})
	ShouldEqual(t, "select * from `users` where `id` = ? or `age` between 1 and 2", b1)
	ElementsShouldMatch(t, []interface{}{1}, b1.GetBindings())
}
func TestOrWhereNotBetweenColumns(t *testing.T) {
	b := GetBuilder().Select().From("users").Where("id", 1).OrWhereNotBetweenColumns("age", []interface{}{"users.created_at", "users.updated_at"})
	ShouldEqual(t, "select * from `users` where `id` = ? or `age` not between `users`.`created_at` and `users`.`updated_at`", b)
	ElementsShouldMatch(t, []interface{}{1}, b.GetBindings())

	b1 := GetBuilder().Select().From("users").Where("id", 1).OrWhereNotBetweenColumns("age", []interface{}{goeloquent.Raw("1"), goeloquent.Raw("2")})
	ShouldEqual(t, "select * from `users` where `id` = ? or `age` not between 1 and 2", b1)
	ElementsShouldMatch(t, []interface{}{1}, b1.GetBindings())

}
func TestBasicOrWheres(t *testing.T) {
	builder := GetBuilder()
	builder.Select().From("users").Where("id", 1).OrWhere("email", "foo")
	assert.Equal(t, "select * from `users` where `id` = ? or `email` = ?", builder.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, builder.GetBindings())

}
func TestBasicOrWhereNot(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").Where("id", 1).OrWhereNot("email", "foo")
	assert.Equal(t, "select * from `users` where `id` = ? or not `email` = ?", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b.GetBindings())
}
func TestRawWheres(t *testing.T) {
	b2 := GetBuilder()
	b2.Select().From("users").WhereRaw("id = ? or email = ?", []interface{}{1, "foo"})
	assert.Equal(t, "select * from `users` where id = ? or email = ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "foo"}, b2.GetBindings())
}
func TestRawOrWheres(t *testing.T) {
	b2 := GetBuilder()
	b2.Select().From("users").Where("id", 1).OrWhereRaw("email = ?", []interface{}{"test@gmail.com"})
	assert.Equal(t, "select * from `users` where `id` = ? or email = ?", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{1, "test@gmail.com"}, b2.GetBindings())
}

func TestBasicWhereIns(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereIn("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `id` in (?,?,?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2, 3}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("name", "Joe").OrWhereIn("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `name` = ? or `id` in (?,?,?)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"Joe", 1, 2, 3}, b1.GetBindings())

}
func TestBasicWhereNotIns(t *testing.T) {

	b := GetBuilder()
	b.Select().From("users").WhereNotIn("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `id` not in (?,?,?)", b.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 2, 3}, b.GetBindings())

	b1 := GetBuilder()
	b1.Select().From("users").Where("name", "Joe").OrWhereNotIn("id", []interface{}{1, 2, 3})
	assert.Equal(t, "select * from `users` where `name` = ? or `id` not in (?,?,?)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{"Joe", 1, 2, 3}, b1.GetBindings())
}
func TestRawWhereIns(t *testing.T) {
	b1 := GetBuilder()
	b1.Select().From("users").WhereIn("id", []interface{}{1, goeloquent.Raw("test"), 3})
	assert.Equal(t, "select * from `users` where `id` in (?,test,?)", b1.ToSql())
	assert.ElementsMatch(t, []interface{}{1, 3}, b1.GetBindings())
	b4 := GetBuilder()
	b4.Select().From("users").Where("name", "Jim").OrWhereNotIn("id", []interface{}{1, goeloquent.Raw("test"), 3})
	assert.Equal(t, "select * from `users` where `name` = ? or `id` not in (?,test,?)", b4.ToSql())
	assert.ElementsMatch(t, []interface{}{"Jim", 1, 3}, b4.GetBindings())

}
func TestEmptyWhereIns(t *testing.T) {
	b := GetBuilder().Select().From("users").WhereIn("id", []interface{}{})
	ShouldEqual(t, "select * from `users` where 0 = 1", b)
	b1 := GetBuilder().Select().From("users").Where("id", "=", 1).OrWhereIn("id", []interface{}{})
	ShouldEqual(t, "select * from `users` where `id` = ? or 0 = 1", b1)
}
func TestEmptyWhereNotIns(t *testing.T) {
	b := GetBuilder().Select().From("users").WhereNotIn("id", []interface{}{})
	ShouldEqual(t, "select * from `users` where 1 = 1", b)
	b1 := GetBuilder().Select().From("users").Where("id", "=", 1).OrWhereNotIn("id", []interface{}{})
	ShouldEqual(t, "select * from `users` where `id` = ? or 1 = 1", b1)
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
	conditions := [][]interface{}{
		{"first_name", "last_name"},
		{"updated_at", ">", "created_at"},
	}
	b := GetBuilder().Select().From("users").WhereColumn(conditions)
	ShouldEqual(t, "select * from `users` where (`first_name` = `last_name` and `updated_at` > `created_at`)", b)
	assert.Empty(t, b.GetBindings())
}
func TestHavingAggregate(t *testing.T) {
	expected := "select count(*) as aggregate from (select (select `count(*)` from `videos` where `posts`.`id` = `videos`.`post_id`) as `videos_count` from `posts` having `videos_count` > ?) as `temp_table`"
	b := GetBuilder()
	var c int
	subFunc := func(builder *goeloquent.Builder) {
		builder.From("videos").Select("count(*)").WhereColumn("posts.id", "=", "videos.post_id")
	}
	b.From("posts").SelectSub(subFunc, "videos_count").Having("videos_count", ">", 1).Count(&c)
	assert.Equal(t, expected, b.PreparedSql)
	assert.ElementsMatch(t, []interface{}{1}, b.GetBindings())
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

func TestBasicWhereNulls(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").WhereNull("id")
	assert.Equal(t, "select * from `users` where `id` is null", b.ToSql())
	assert.Empty(t, b.GetBindings())

	b2 := GetBuilder()
	b2.Select().From("users").Where("id", 0).OrWhereNull("id")
	assert.Equal(t, "select * from `users` where `id` = ? or `id` is null", b2.ToSql())
	assert.ElementsMatch(t, []interface{}{0}, b2.GetBindings())
}
func TestArrayWhereNulls(t *testing.T) {
	b3 := GetBuilder()
	b3.Select().From("users").WhereNull([]interface{}{"id", "deleted_at"})
	assert.Equal(t, "select * from `users` where `id` is null and `deleted_at` is null", b3.ToSql())

	b4 := GetBuilder()
	b4.Select().From("users").Where("id", "<", 0).OrWhereNull([]interface{}{"id", "email"})
	assert.Equal(t, "select * from `users` where `id` < ? or `id` is null or `email` is null", b4.ToSql())
	assert.ElementsMatch(t, []interface{}{0}, b4.GetBindings())
}
func TestJsonWhereNull(t *testing.T)                   {}
func TestJsonWhereNotNullMysql(t *testing.T)           {}
func TestJsonWhereNullExpressionMysql(t *testing.T)    {}
func TestJsonWhereNotNullExpressionMysql(t *testing.T) {}
func TestBasicWhereNotNulls(t *testing.T) {
	b5 := GetBuilder()
	b5.Select().From("users").WhereNotNull("id")
	assert.Equal(t, "select * from `users` where `id` is not null", b5.ToSql())
	assert.Empty(t, b5.GetBindings())

	b6 := GetBuilder()
	b6.Select().From("users").Where("id", 0).OrWhereNotNull("id")
	assert.Equal(t, "select * from `users` where `id` = ? or `id` is not null", b6.ToSql())
	assert.ElementsMatch(t, []interface{}{0}, b6.GetBindings())
}
func TestArrayWhereNotNulls(t *testing.T) {

	b7 := GetBuilder()
	b7.Select().From("users").WhereNotNull([]interface{}{"id", "deleted_at"})
	assert.Equal(t, "select * from `users` where `id` is not null and `deleted_at` is not null", b7.ToSql())

	b8 := GetBuilder()
	b8.Select().From("users").Where("id", "<", 0).OrWhereNotNull([]interface{}{"id", "email"})
	assert.Equal(t, "select * from `users` where `id` < ? or `id` is not null or `email` is not null", b8.ToSql())
	assert.ElementsMatch(t, []interface{}{0}, b8.GetBindings())
}

func TestGroupBys(t *testing.T) {
	b := GetBuilder()
	b.Select().From("users").GroupBy("email")
	assert.Equal(t, "select * from `users` group by `email`", b.ToSql())

	b1 := GetBuilder()
	b1.Select().From("users").GroupBy("email", "id")
	ShouldEqual(t, "select * from `users` group by `email`, `id`", b1)

	ShouldEqual(t, "select * from `users` group by DATE(created_at)",
		GetBuilder().Select().From("users").GroupBy(goeloquent.Raw("DATE(created_at)")))

	groupByRaw := GetBuilder().Select().From("users").GroupByRaw("DATE(created_at), ? DESC", []interface{}{"foo"})
	ShouldEqual(t, "select * from `users` group by DATE(created_at), ? DESC", groupByRaw)
	ElementsShouldMatch(t, []interface{}{"foo"}, groupByRaw.GetBindings())

	ElementsShouldMatch(t, []interface{}{"whereRawBinding", "groupByRawBinding", "havingRawBinding"},
		GetBuilder().HavingRaw("?", []interface{}{"havingRawBinding"}).GroupByRaw("?", []interface{}{"groupByRawBinding"}).WhereRaw("?", []interface{}{"whereRawBinding"}).GetBindings())
}

func TestOrderBys(t *testing.T) {
	ShouldEqual(t, "select * from `users` order by `email` asc, `age` desc",
		GetBuilder().From("users").Select().OrderBy("email").OrderBy("age", "desc"))

	orderByRaw := GetBuilder().Select().From("users").OrderBy("email").OrderByRaw("`age` ? desc", []interface{}{"foo"})
	ShouldEqual(t, "select * from `users` order by `email` asc, `age` ? desc", orderByRaw)
	ElementsShouldMatch(t, []interface{}{"foo"}, orderByRaw.GetBindings())

	orderByRaw = GetBuilder().Select().From("users").OrderByDesc("email")
	ShouldEqual(t, "select * from `users` order by `email` desc", orderByRaw)

}
func TestInRandomOrderMySql(t *testing.T) {
	ShouldEqual(t, "select * from `users` order by RAND()", GetBuilder().Select().From("users").InRandomOrder())
	ShouldEqual(t, "select * from `users` order by RAND(123)", GetBuilder().Select().From("users").InRandomOrder(123))
}
func TestReorder(t *testing.T) {
	orderBy := GetBuilder().From("users").Select().OrderBy("email")
	ShouldEqual(t, "select * from `users` order by `email` asc", orderBy)
	orderBy = GetBuilder().From("users").Select().OrderBy("email").ReOrder()
	ShouldEqual(t, "select * from `users`", orderBy)

	orderBy = GetBuilder().From("users").Select().OrderBy("email")
	ShouldEqual(t, "select * from `users` order by `email` asc", orderBy)
	orderBy = GetBuilder().From("users").Select().OrderBy("email").ReOrder("name", "desc")
	ShouldEqual(t, "select * from `users` order by `name` desc", orderBy)

	orderBy = GetBuilder().Select().From("users").OrderByRaw("?", []interface{}{true})
	ElementsShouldMatch(t, []interface{}{true}, orderBy.GetBindings())

	orderBy.ReOrder()
	ElementsShouldMatch(t, []interface{}{}, orderBy.GetBindings())
}
func TestOrderBySubQueries(t *testing.T) {
	orderBySql := "select * from `users` order by (select `created_at` from `logins` where `user_id` = `users`.`id` limit 1)"
	ShouldEqual(t, orderBySql+" asc",
		GetBuilder().From("users").Select().OrderBy(func(builder *goeloquent.Builder) {
			builder.From("logins").Select("created_at").WhereColumn("user_id", "=", "users.id").Limit(1)
		}))

	ShouldEqual(t, orderBySql+" desc",
		GetBuilder().From("users").Select().OrderBy(func(builder *goeloquent.Builder) {
			builder.From("logins").Select("created_at").WhereColumn("user_id", "=", "users.id").Limit(1)
		}, "desc"))

}
func TestHavings(t *testing.T) {
	ShouldEqual(t, "select * from `users` having `email` > ?", GetBuilder().From("users").Select().Having("email", ">", 1))
	having := GetBuilder().From("users").Select().OrHaving("email", "=", "a@gmail.com").OrHaving("email", "=", "b@gmail.com")

	ShouldEqual(t, "select * from `users` having `email` = ? or `email` = ?", having)

	having = GetBuilder().From("users").Select().GroupBy("email").Having("email", ">", 1)
	ShouldEqual(t, "select * from `users` group by `email` having `email` > ?", having)
	ElementsShouldMatch(t, []interface{}{1}, having.GetBindings())

	ShouldEqual(t, "select `email` as `foo_email` from `users` having `foo_email` > ?", GetBuilder().From("users").Select("email as foo_email").Having("foo_email", ">", 1))

	ShouldEqual(t, "select `category`, count(*) as `total` from `item` where `department` = ? group by `category` having `total` > 3",
		GetBuilder().Select("category", goeloquent.Raw("count(*) as `total`")).From("item").Where("department", "popular").GroupBy("category").Having("total", ">", goeloquent.Raw("3")))

}
func TestNestedHavings(t *testing.T) {
	b := GetBuilder().Select().From("users").Having(func(builder *goeloquent.Builder) *goeloquent.Builder {
		return builder.Having("email", "foo").OrHaving("email", "bar")
	}).OrHaving("email", "baz")
	ShouldEqual(t, "select * from `users` having (`email` = ? or `email` = ?) or `email` = ?", b)
	ElementsShouldMatch(t, []interface{}{"foo", "bar", "baz"}, b.GetBindings())
}
func TestHavingBetweens(t *testing.T) {
	b := GetBuilder().Select().From("users").HavingBetween("id", []interface{}{1, 3, 5})
	ShouldEqual(t, "select * from `users` having `id` between ? and ?", b)
	ElementsShouldMatch(t, []interface{}{1, 3}, b.GetBindings())
}
func TestHavingNull(t *testing.T) {
	b := GetBuilder().Select().From("users").HavingNull("id")
	ShouldEqual(t, "select * from `users` having `id` is null", b)
	assert.Empty(t, b.GetBindings())

	b1 := GetBuilder().Select().From("users").HavingNotNull("id")
	ShouldEqual(t, "select * from `users` having `id` is not null", b1)
	assert.Empty(t, b1.GetBindings())

	b2 := GetBuilder().Select().From("users").WhereNull("deleted_at").Where("verified", 1).GroupBy("role").HavingNull("total")
	ShouldEqual(t, "select * from `users` where `deleted_at` is null and `verified` = ? group by `role` having `total` is null", b2)
	ElementsShouldMatch(t, []interface{}{1}, b2.GetBindings())
}
func TestHavingNotNull(t *testing.T) {
	b3 := GetBuilder().Select().From("users").HavingNotNull("email")
	ShouldEqual(t, "select * from `users` having `email` is not null", b3)

	b4 := GetBuilder().Select().From("users").HavingNotNull("email").HavingNotNull("name")
	ShouldEqual(t, "select * from `users` having `email` is not null and `name` is not null", b4)

	b5 := GetBuilder().Select().From("users").OrHavingNotNull("email").OrHavingNotNull("name")
	ShouldEqual(t, "select * from `users` having `email` is not null or `name` is not null", b5)

	b6 := GetBuilder().Select().From("users").GroupBy("email").HavingNotNull("email")
	ShouldEqual(t, "select * from `users` group by `email` having `email` is not null", b6)

	b7 := GetBuilder().Select().From("users").Where("id", 1).GroupBy("email").HavingNotNull("email")
	ShouldEqual(t, "select * from `users` where `id` = ? group by `email` having `email` is not null", b7)

}
func TestHavingShortcut(t *testing.T) {
	ShouldEqual(t, "select * from `users` having `email` = ? or `email` = ?", GetBuilder().Select().From("users").Having("email", 1).OrHaving("email", 2))
}
func TestHavingFollowedBySelectGet(t *testing.T) {
	s := GetBuilder().Select("category", goeloquent.Raw("count(*) as total")).From("users").Where("id", ">", 1).GroupBy("role").Having("total", ">", 1).OrHaving("email", "<", 2)
	ShouldEqual(t, "select `category`, count(*) as total from `users` where `id` > ? group by `role` having `total` > ? or `email` < ?", s)

}
func TestRawHavings(t *testing.T) {
	ShouldEqual(t, "select * from `users` having user_foo < user_bar", GetBuilder().Select().From("users").HavingRaw("user_foo < user_bar"))
	ShouldEqual(t, "select * from `users` having `baz` = ? or user_foo < user_bar", GetBuilder().Select().From("users").Having("baz", "=", 1).OrHavingRaw("user_foo < user_bar"))
	ShouldEqual(t, "select * from `users` having `last_login_date` between ? and ? or user_foo < user_bar", GetBuilder().Select().From("users").HavingBetween("last_login_date", []interface{}{"2018-11-11", "2022-02-02"}).OrHavingRaw("user_foo < user_bar"))
}

func TestLimiAndOffsets(t *testing.T) {
	ShouldEqual(t, "select * from `users` limit 10 offset 5", GetBuilder().Select().From("users").Offset(5).Limit(10))

	ShouldEqual(t, "select * from `users` limit 0", GetBuilder().Select().From("users").Limit(0))

}

func TestForPage(t *testing.T) {
	ShouldEqual(t, "select * from `users` limit 15 offset 15", GetBuilder().Select().From("users").ForPage(2, 15))
	ShouldEqual(t, "select * from `users` limit 15 offset 0", GetBuilder().Select().From("users").ForPage(0, 15))
	ShouldEqual(t, "select * from `users` limit 15 offset 0", GetBuilder().Select().From("users").ForPage(-2, 15))
	ShouldEqual(t, "select * from `users` limit 0 offset 0", GetBuilder().Select().From("users").ForPage(2, 0))
	ShouldEqual(t, "select * from `users` limit 0 offset 0", GetBuilder().Select().From("users").ForPage(0, 0))
	ShouldEqual(t, "select * from `users` limit 0 offset 0", GetBuilder().Select().From("users").ForPage(-2, 0))

	//TODO: testGetCountForPaginationWithColumnAliases
}

//	func TestGetCountForPaginationWithBindings(t *testing.T) {
//		b := GetBuilder()
//		b.From("users").Select(func(builder *goeloquent.Builder) *goeloquent.Builder {
//			return builder.Select("body").From("posts").Where("id", 4)
//		}, "post")
//		c, e := b.GetCountForPagination()
//		assert.Equal(t, c, int64(0))
//		assert.Nil(t, e)
//	}
func TestWhereShortcut(t *testing.T) {
	b3 := GetBuilder().Select().From("users").Where("id", 1).OrWhere("name", "Jack")
	ShouldEqual(t, "select * from `users` where `id` = ? or `name` = ?", b3)
	ElementsShouldMatch(t, []interface{}{1, "Jack"}, b3.GetBindings())
}

func TestWhereWithArrayConditions(t *testing.T) {
	b4 := GetBuilder().Select().From("users").Where([][]interface{}{
		{"admin", "=", 1, goeloquent.BOOLEAN_AND},
		{"id", "<", 10, goeloquent.BOOLEAN_OR},
		{"source", "=", "301"},
		{"deleted", 0},
		{"role", "in", []interface{}{"admin", "manager", "owner"}},
		{"age", "between", []interface{}{18, 60, 100}},
		{map[string]interface{}{"name": "Bob", "location": "NY"}, goeloquent.BOOLEAN_OR},
		{func(builder *goeloquent.Builder) *goeloquent.Builder {
			return builder.WhereYear("created_at", "<", 2010).WhereColumn("first_name", "last_name").OrWhereNull("created_at")
		}},
		{goeloquent.Raw("year(birthday) < 1998")},
		{"suspend", goeloquent.Raw("'nodoublequotes'")},
		{goeloquent.Where{
			Type:     goeloquent.CONDITION_TYPE_BASIC,
			Column:   "alias",
			Operator: "=",
			Value:    "boss",
			Boolean:  goeloquent.BOOLEAN_OR,
		}},
	})
	assert.Contains(t, b4.ToSql(),
		"select * from `users` where `admin` = ? or `id` < ? and `source` = ? and `deleted` = ? and `role` in (?,?,?) and `age` between ? and ? or ",
		"(`name` = ? and `location` = ?)",
		"(year(`created_at`) < ? and `first_name` = `last_name` or `created_at` is null)",
		"and year(birthday) < 1998", "and `suspend` = 'nodoublequotes'")
	ElementsShouldMatch(t, []interface{}{1, 10, "301", 0, "admin", "manager", "owner", 18, 60, "Bob", "NY", 2010}, b4.GetBindings())

}
func TestNestedWheres(t *testing.T) {

	b5 := GetBuilder().Select().From("users").Where("email", "foo").OrWhere(func(builder *goeloquent.Builder) {
		builder.Where("name", "bar").Where("age", "=", 25)
	})
	ShouldEqual(t, "select * from `users` where `email` = ? or (`name` = ? and `age` = ?)", b5)
	ElementsShouldMatch(t, []interface{}{"foo", "bar", 25}, b5.GetBindings())
}

func TestWhereNot(t *testing.T) {
	b6 := GetBuilder().Select().From("users").Where("email", "foo").OrWhereNot("name", "bar").OrWhereNot(func(builder *goeloquent.Builder) {

		builder.WhereNot("name", "baz").OrWhereNot("age", "=", 25)
	})
	ShouldEqual(t, "select * from `users` where `email` = ? or not `name` = ? or (not `name` = ? or not `age` = ?)", b6)
	ElementsShouldMatch(t, []interface{}{"foo", "bar", "baz", 25}, b6.GetBindings())
}
func TestFullSubSelects(t *testing.T) {
	b := GetBuilder().Select().From("users").Where("id", 1).OrWhere("id", "=", func(builder *goeloquent.Builder) *goeloquent.Builder {
		return builder.Select(goeloquent.Raw("max(id)")).From("users").Where("email", "like", "gmail.com")
	}).OrWhere("id", func(builder *goeloquent.Builder) *goeloquent.Builder {
		return builder.Select("id").From("users").Where("age", ">", 25)
	})

	ShouldEqual(t, "select * from `users` where `id` = ? or `id` = (select max(id) from `users` where `email` like ?) or `id` = (select `id` from `users` where `age` > ?)", b)
	ElementsShouldMatch(t, []interface{}{1, "gmail.com", 25}, b.GetBindings())
}
func TestWhereExists(t *testing.T) {
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

func TestBasicJoins(t *testing.T) {
	b := GetBuilder().Select().From("users").Join("contacts", "users.id", "=", "contacts.id")
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id`", b)

	b1 := GetBuilder().Select().From("users").Join("contacts", "users.id", "=", "contacts.id").LeftJoin("photos", "users.id", "=", "photos.id")
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` left join `photos` on `users`.`id` = `photos`.`id`", b1)

	b2 := GetBuilder().Select().From("users").LeftJoinWhere("photos", "users.id", "=", "bar").JoinWhere("photos", "users.id", "=", "foo")
	ShouldEqual(t, "select * from `users` left join `photos` on `users`.`id` = ? inner join `photos` on `users`.`id` = ?", b2)
	ElementsShouldMatch(t, []interface{}{"bar", "foo"}, b2.GetBindings())

}
func TestCrossJoins(t *testing.T) {
	b := GetBuilder().Select().From("sizes").CrossJoin("colors")
	ShouldEqual(t, "select * from `sizes` cross join `colors`", b)
	b1 := GetBuilder().Select("*").From("tableB").Join("tableA", "tableA.column1", "=", "tableB.column2", goeloquent.JOIN_TYPE_CROSS)
	ShouldEqual(t, "select * from `tableB` cross join `tableA` on `tableA`.`column1` = `tableB`.`column2`", b1)
	b2 := GetBuilder().Select().From("tableB").CrossJoin("tableA", "tableA.column1", "=", "tableB.column2")
	ShouldEqual(t, "select * from `tableB` cross join `tableA` on `tableA`.`column1` = `tableB`.`column2`", b2)
}
func TestCrossJoinSubs(t *testing.T) {
	b := GetBuilder().SelectRaw("(sale / overall.sales) * 100 AS percent_of_total").From("sales").CrossJoinSub(GetBuilder().SelectRaw("SUM(sale) AS sales").From("sales"), "overall")
	ShouldEqual(t, "select (sale / overall.sales) * 100 AS percent_of_total from `sales` cross join (select SUM(sale) AS sales from `sales`) as `overall`", b)
}
func TestComplexJoin(t *testing.T) {
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
}
func TestJoinWhereNull(t *testing.T) {
	b5 := GetBuilder().Select().From("users").Join("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").WhereNull("contacts.deleted_at")
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` and `contacts`.`deleted_at` is null", b5)
	b6 := GetBuilder().Select().From("users").Join("contacts", func(clasuse *goeloquent.Builder) {
		clasuse.On("users.id", "=", "contacts.id").OrWhereNull("contacts.deleted_at")
	})
	ShouldEqual(t, "select * from `users` inner join `contacts` on `users`.`id` = `contacts`.`id` or `contacts`.`deleted_at` is null", b6)
}
func TestJoinWhereNotNull(t *testing.T) {

}
func TestJoinWhereIn(t *testing.T) {
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
}
func TestJoinWhereInSubquery(t *testing.T) {
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
}

//	func TestJoinScan(t *testing.T) {
//		//Test join struct
//		c, d := CreateRelationTables()
//		type UserWithAddress struct {
//			Id       int64  `goelo:"column:id"`
//			Age      int    `goelo:"column:age"`
//			Name     string `goelo:"column:name"`
//			Country  string `goelo:"column:country"`
//			Province string `goelo:"column:province"`
//			City     string `goelo:"column:city"`
//			Address  string `goelo:"column:detail"`
//		}
//		RunWithDB(c, d, func() {
//			var u UserWithAddress
//			user, err := DB.Insert("insert into user (name,age) values ('Alex',18)", nil)
//			assert.Nil(t, err)
//			uid, _ := user.LastInsertId()
//			add, err := DB.Insert("insert into address (user_id,country,state,city,detail) values (?,?,'IL','Chicago','2609 S Halsted St')", []interface{}{uid, 1})
//			assert.Nil(t, err)
//			addId, err := add.LastInsertId()
//			assert.Greater(t, addId, int64(0))
//
//			b1 := DB.Query()
//			_, err = b1.From("user").Select("user.id as id", "user.age as age", "user.name as name").
//				AddSelect("address.country as country", "address.state as state", "address.city as city", "address.detail as detail").
//				Join("address", "user.id", "address.user_id").First(&u)
//			assert.Nil(t, err)
//			assert.ObjectsAreEqualValues(u, UserWithAddress{
//				Id:       1,
//				Age:      18,
//				Name:     "Alex",
//				Country:  "1",
//				Province: "IL",
//				City:     "Chicage",
//				Address:  "2609 S Halsted St",
//			})
//			joinMap := make(map[string]interface{})
//			_, err = DB.Query().From("user").Select("user.*").
//				AddSelect("address.*").
//				Join("address", "user.id", "address.user_id").First(&joinMap)
//			assert.Nil(t, err)
//			assert.Equal(t, u.City, string((joinMap["city"]).([]byte)))
//			assert.Equal(t, u.Name, string((joinMap["name"]).([]byte)))
//			assert.Equal(t, u.Address, string((joinMap["detail"]).([]byte)))
//
//			//Test where query
//			joinMap = make(map[string]interface{})
//			_, err = DB.Query().From("user").Select("user.*").Where("user.id", uid).
//				AddSelect("address.*").
//				Join("address", "user.id", "address.user_id").First(&joinMap)
//			assert.Nil(t, err)
//			assert.Equal(t, u.City, string((joinMap["city"]).([]byte)))
//			assert.Equal(t, u.Name, string((joinMap["name"]).([]byte)))
//			assert.Equal(t, u.Address, string((joinMap["detail"]).([]byte)))
//		})
//	}
func TestJoinWhereNotIn(t *testing.T) {

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
func TestJoinsWithNestedConditions(t *testing.T) {
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
}
func TestJoinsWithAdvancedConditions(t *testing.T) {
	b17 := GetBuilder().Select().From("users").LeftJoin("contacts", func(builder *goeloquent.Builder) {
		builder.On("users.id", "contacts.id").Where(func(builder2 *goeloquent.Builder) {
			builder2.Where("role", "admin").OrWhereNull("contacts.disabled").OrWhereRaw("year(contacts.created_at) = 2016")
		})
	})
	ShouldEqual(t, "select * from `users` left join `contacts` on `users`.`id` = `contacts`.`id` and (`role` = ? or `contacts`.`disabled` is null or year(contacts.created_at) = 2016)", b17)
	ElementsShouldMatch(t, []interface{}{"admin"}, b17.GetBindings())
}
func TestJoinsWithSubqueryCondition(t *testing.T) {
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
func TestJoinsWithAdvancedSubqueryCondition(t *testing.T) {
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

}
func TestJoinsWithNestedJoins(t *testing.T) {
	b1 := GetBuilder().Select("users.id", "contacts.id", "contact_types.id").From("users").LeftJoin("contacts", func(builder2 *goeloquent.Builder) {
		builder2.On("users.id", "contacts.id").Join("contact_types", "contacts.contact_type_id", "=", "contact_types.id")
	})
	ShouldEqual(t, "select `users`.`id`, `contacts`.`id`, `contact_types`.`id` from `users` left join (`contacts` inner join `contact_types` on `contacts`.`contact_type_id` = `contact_types`.`id`) on `users`.`id` = `contacts`.`id`", b1)
}
func TestJoinsWithMultipleNestedJoins(t *testing.T) {
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
}
func TestJoinsWithNestedJoinWithAdvancedSubqueryCondition(t *testing.T) {
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
}
func TestJoinWithNestedOnCondition(t *testing.T) {

}
func TestJoinSub(t *testing.T) {
	b4 := GetBuilder().Select().From("users").JoinSub(goeloquent.Raw("select * from `contacts`"), "sub", "users.id", "=", "sub.id")
	ShouldEqual(t, "select * from `users` inner join (select * from `contacts`) as `sub` on `users`.`id` = `sub`.`id`", b4)

	b5 := GetBuilder().Select().From("users").JoinSub(GetBuilder().Select().From("contacts"), "sub", "users.id", "=", "sub.id")
	ShouldEqual(t, "select * from `users` inner join (select * from `contacts`) as `sub` on `users`.`id` = `sub`.`id`", b5)
	c1 := GetBuilder().From("contacts").Select().Where("name", "foo")
	c2 := GetBuilder().From("contacts").Select().Where("name", "bar")
	b6 := GetBuilder().Select().From("users").JoinSub(c1, "sub1", "users.id", "=", 1, "inner", true).JoinSub(c2, "sub2", "users.id", "=", "sub2.user_id")
	expected := "select * from `users` inner join (select * from `contacts` where `name` = ?) as `sub1` on `users`.`id` = ? inner join (select * from `contacts` where `name` = ?) as `sub2` on `users`.`id` = `sub2`.`user_id`"
	ShouldEqual(t, expected, b6)
	ElementsShouldMatch(t, []interface{}{"foo", 1, "bar"}, b6.GetBindings())
}
func TestJoinSubWithPrefix(t *testing.T) {
	b7 := GetBuilder()
	b7.Grammar.SetTablePrefix("prefix_")
	b7.Select().From("users").JoinSub("select * from `contacts`", "sub", "users.id", "=", "sub.id")
	ShouldEqual(t, "select * from `prefix_users` inner join (select * from `contacts`) as `prefix_sub` on `prefix_users`.`id` = `prefix_sub`.`id`", b7)
}
func TestLeftJoinSub(t *testing.T) {
	b8 := GetBuilder().Select().From("users").LeftJoinSub(GetBuilder().Select().From("contacts"), "sub", "users.id", "=", "sub.id")
	ShouldEqual(t, "select * from `users` left join (select * from `contacts`) as `sub` on `users`.`id` = `sub`.`id`", b8)
}
func TestRightJoinSub(t *testing.T) {
	b9 := GetBuilder().Select().From("users").RightJoinSub(GetBuilder().Select().From("contacts"), "sub", "users.id", "=", "sub.id")
	ShouldEqual(t, "select * from `users` right join (select * from `contacts`) as `sub` on `users`.`id` = `sub`.`id`", b9)
	//

}
func TestJoinLateral(t *testing.T)           {}
func TestJoinLateralWithPrefix(t *testing.T) {}
func TestLeftJoinLateral(t *testing.T)       {}
func TestRawExpressionsInSelect(t *testing.T) {
	b10 := GetBuilder().Select(goeloquent.Raw("substr(foo,6)")).From("users")
	ShouldEqual(t, "select substr(foo,6) from `users`", b10)
}
func TestFindReturnsFirstResultByID(t *testing.T) {

	var user = make(map[string]interface{})
	b2 := DB.Query()
	rows, err := b2.From("user_models").Find(&user, 1)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), rows.Count)
	assert.Equal(t, user["id"], int64(1))
	assert.Equal(t, "select * from `user_models` where `id` = ? limit 1", b2.PreparedSql)
	assert.Equal(t, []interface{}{1}, b2.GetBindings())
}
func TestFirstMethodReturnsFirstResult(t *testing.T) {
	var user2 = make(map[string]interface{})
	b3 := DB.Query()
	b3.From("users").Where("id", "=", 1).First(&user2)
	assert.Equal(t, "select * from `users` where `id` = ? limit 1", b3.PreparedSql)
	assert.Equal(t, []interface{}{1}, b3.GetBindings())
}
func TestValueMethodReturnsSingleColumn(t *testing.T) {
	var name string
	b1 := DB.Query()
	b1.From("users").Where("id", 1).Value(&name, "name")
	assert.Equal(t, "select `name` from `users` where `id` = ? limit 1", b1.PreparedSql)
}
func TestRawValueMethodReturnsSingleColumn(t *testing.T) {

}
func TestAggregateFunctions(t *testing.T) {
	b := DB.Query()
	c := 0
	b.From("users").Count(&c)
	assert.Equal(t, "select count(*) as aggregate from `users`", b.ToSql())

	b1 := DB.Query()
	b1.From("users").Select().Exists()
	assert.Equal(t, "select exists(select * from `users`) as `exists`", b1.PreparedSql)

	b2 := DB.Query()
	b2.From("users").Select().Exists()
	assert.Equal(t, "select exists(select * from `users`) as `exists`", b1.PreparedSql)

	var m = 0
	b3 := DB.Query()
	b3.From("users").Max(&m, "id")
	assert.Equal(t, "select max(`id`) as aggregate from `users`", b3.ToSql())

	var m1 = 0
	b4 := DB.Query()
	b4.From("users").Min(&m1, "id")
	assert.Equal(t, "select min(`id`) as aggregate from `users`", b4.ToSql())

	b5 := DB.Query()
	b5.From("users").Sum(&m1, "age")
	assert.Equal(t, "select sum(`age`) as aggregate from `users`", b5.ToSql())
}
func TestAggregateWithSubSelect(t *testing.T) {

}
func TestSubqueriesBindings(t *testing.T) {
	f := GetBuilder()
	f.Select().From("users").Where("email", "=", func(builder *goeloquent.Builder) {
		builder.Select(goeloquent.Raw("max(id)")).From("users").Where("email", "=", "bar").
			OrderByRaw("email like ?", []interface{}{"%.com"}).GroupBy("id").Having("id", "=", 4)
	}).OrWhere("id", "=", "foo").GroupBy("id").Having("id", "=", 5)
	ElementsShouldMatch(t, []interface{}{"bar", 4, "%.com", "foo", 5}, f.GetBindings())
}
func TestInsertMethod(t *testing.T) {
	//TestInsertMethod
	createUsers, dropUsers := UserTableSql()
	now := time.Now()
	u1 := map[string]interface{}{
		"name":       "go-eloquent",
		"age":        18,
		"created_at": now,
	}
	RunWithDB(createUsers, dropUsers, func() {
		b := DB.Table("users")
		insert, err := b.Insert(u1)
		assert.Nil(t, err)
		c, err := insert.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, c, int64(1))
		ElementsShouldMatch(t, []interface{}{"go-eloquent", 18, now}, b.GetBindings())
		assert.True(t, strings.Contains(b.PreparedSql, "insert into `users` ("))

	})
	//TestMySqlInsertOrIgnoreMethod
	RunWithDB(createUsers, dropUsers, func() {
		b := DB.Table("users")
		insert, err := b.InsertOrIgnore(u1)
		assert.Nil(t, err)
		c, err := insert.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, c, int64(1))
		ElementsShouldMatch(t, []interface{}{"go-eloquent", 18, now}, b.GetBindings())
		assert.True(t, strings.Contains(b.PreparedSql, "insert ignore into `users` ("))
	})

	//

	//

}
func TestInsertUsingMethod(t *testing.T) {

}
func TestInsertOrIgnoreMethod(t *testing.T)       {}
func TestInsertOrIgnoreUsingMethod(t *testing.T)  {}
func TestInsertGetIdMethod(t *testing.T)          {}
func TestInsertGetIdWithEmptyValues(t *testing.T) {}
func TestInsertMethodRespectsRawBindings(t *testing.T) {
	b4 := DB.Query().Pretend()
	b4.From("users").Insert(map[string]interface{}{
		"email": goeloquent.Raw("CURRENT TIMESTAMP"),
	})
	assert.Equal(t, "insert into `users` (`email`) values (CURRENT TIMESTAMP)", b4.PreparedSql)
	assert.Nil(t, b4.GetBindings())
	createUsers, dropUsers := UserTableSql()
	now := time.Now()
	RunWithDB(createUsers, dropUsers, func() {
		b := DB.Table("users")
		insert, err := b.Insert(map[string]interface{}{
			"name":       "go-eloquent",
			"age":        18,
			"created_at": now,
			"deleted_at": goeloquent.Raw("CURRENT_TIMESTAMP"),
		})
		assert.Nil(t, err)
		c, err := insert.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, c, int64(1))
		assert.True(t, strings.Contains(b.PreparedSql, "insert into `users` ("))
		ElementsShouldMatch(t, []interface{}{"go-eloquent", 18, now}, b.GetBindings())
	})
}
func TestMultipleInsertsWithExpressionValues(t *testing.T) {
	createUsers, dropUsers := UserTableSql()
	now := time.Now()
	RunWithDB(createUsers, dropUsers, func() {
		b := DB.Table("users")
		insert, err := b.Insert(map[string]interface{}{
			"name":       goeloquent.Raw("CONCAT (UPPER('foo'),LOWER('BAR'))"),
			"age":        18,
			"created_at": now,
		})
		assert.Nil(t, err)
		c, err := insert.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, c, int64(1))
		ElementsShouldMatch(t, []interface{}{18, now}, b.GetBindings())
		assert.Contains(t, b.PreparedSql, "insert into `users` (", "CONCAT (UPPER('foo'),LOWER('BAR'))")
	})
}
func TestUpdateMethod(t *testing.T) {
	//TestUpdateMethod
	createUsers, dropUsers := UserTableSql()
	now := time.Now()
	u1 := map[string]interface{}{
		"name":       "go-eloquent",
		"age":        8,
		"created_at": now,
	}
	RunWithDB(createUsers, dropUsers, func() {
		b := DB.Table("users")
		insert, err := b.Insert(u1)
		assert.Nil(t, err)
		id, err := insert.LastInsertId()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), id)
		b1 := DB.Table("users")
		result, err := b1.Where("id", id).Update(map[string]interface{}{
			"name":       "newname",
			"age":        18,
			"updated_at": now.Add(time.Hour * 1),
		})
		assert.Nil(t, err)
		ElementsShouldMatch(t, []interface{}{id, "newname", 18, now.Add(time.Hour * 1)}, b1.GetBindings())
		updated, err := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), updated)
		res := map[string]interface{}{}
		_, err = DB.Table("users").Find(&res, id)
		assert.Equal(t, res["name"].([]byte), []byte("newname"))
		assert.Equal(t, res["age"].(int64), int64(18))
		assert.InDelta(t, res["updated_at"].(time.Time).Unix(), now.Add(time.Hour*1).Unix(), 1)

	})
}
func TestUpsertMethod(t *testing.T) {
	//b := GetBuilder()
	//b.From("users").Upsert([]map[string]interface{}{
	//	{
	//		"email": "foo",
	//		"name":  "bar",
	//	},
	//	{
	//		"name":  "bar2",
	//		"email": "foo2",
	//	},
	//},
	//	[]string{"email"}, nil,
	//)
	//assert.Equal(t, "insert into `users` (`email`, `name`) values (?, ?), (?, ?) on duplicate key update `email` = values(`email`), `name` = values(`name`)", b.PreparedSql)
	//assert.Equal(t, []interface{}{"foo", "bar", "foo2", "bar2"}, b.GetBindings())

}
func TestUpsertMethodWithUpdateColumns(t *testing.T) {
	//b := GetBuilder()
	//b.From("users").Upsert([]map[string]interface{}{
	//	{
	//		"email": "foo",
	//		"name":  "bar",
	//	},
	//	{
	//		"name":  "bar2",
	//		"email": "foo2",
	//	},
	//},
	//	[]string{"email"}, []string{"name"},
	//)
	//assert.Equal(t, "insert into `users` (`email`, `name`) values (?, ?), (?, ?) on duplicate key update `name` = values(`name`)", b.PreparedSql)
	//assert.Equal(t, []interface{}{"foo", "bar", "foo2", "bar2"}, b.GetBindings())
}
func TestUpdateMethodWithJoinsOnMySql(t *testing.T)      {}
func TestUpdateMethodRespectsRaw(t *testing.T)           {}
func TestUpdateMethodWorksWithQueryAsValue(t *testing.T) {}
func TestUpdateOrInsertMethod(t *testing.T)              {}
func TestDeleteMethod(t *testing.T) {
	//TestDeleteMethod
	createUsers, dropUsers := UserTableSql()
	now := time.Now()
	u1 := map[string]interface{}{
		"name":       "go-eloquent",
		"age":        8,
		"created_at": now,
	}
	RunWithDB(createUsers, dropUsers, func() {
		insert, err := DB.Table("users").Insert(u1)
		id, err := insert.LastInsertId()
		assert.Nil(t, err)
		b := DB.Table("users")
		result, err := b.Where("id", id).Delete()
		deleted, _ := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), deleted)
		assert.Equal(t, "delete from `users` where `id` = ?", b.PreparedSql)
		assert.Equal(t, []interface{}{id}, b.GetBindings())
		result, err = DB.Table("users").Insert([]map[string]interface{}{
			{
				"name":       "child",
				"age":        8,
				"created_at": now,
			},
			{
				"name":       "teen",
				"age":        18,
				"created_at": now,
			},
		})
		assert.Nil(t, err)
		created, _ := result.RowsAffected()
		assert.Equal(t, int64(2), created)
		var c int
		_, err = DB.Table("users").Select().Count(&c)
		assert.Nil(t, err)
		assert.Equal(t, 2, c)
		b1 := DB.Table("users")
		result, err = b1.Where("age", ">", 10).Delete()
		deleted, err = result.RowsAffected()
		assert.Equal(t, "delete from `users` where `age` > ?", b1.PreparedSql)
		assert.Equal(t, []interface{}{10}, b1.GetBindings())
		assert.Equal(t, int64(1), deleted)

		b2 := DB.Table("users")
		result, err = b2.Where("age", 8).OrderBy("id").Limit(1).Delete()
		assert.Nil(t, err)
		deleted, err = result.RowsAffected()
		assert.Equal(t, int64(1), deleted)
		assert.Equal(t, "delete from `users` where `age` = ? order by `id` asc limit 1", b2.PreparedSql)

	})
}
func TestDeleteWithJoinMethod(t *testing.T)                                   {}
func TestTruncateMethod(t *testing.T)                                         {}
func TestMySqlWrapping(t *testing.T)                                          {}
func TestMySqlUpdateWrappingJson(t *testing.T)                                {}
func TestMySqlUpdateWrappingNestedJson(t *testing.T)                          {}
func TestMySqlUpdateWrappingJsonArray(t *testing.T)                           {}
func TestMySqlUpdateWrappingJsonPathArrayIndex(t *testing.T)                  {}
func TestMySqlUpdateWithJsonPreparesBindingsCorrectly(t *testing.T)           {}
func TestMySqlWrappingJsonWithString(t *testing.T)                            {}
func TestMySqlWrappingJsonWithInteger(t *testing.T)                           {}
func TestMySqlWrappingJsonWithDouble(t *testing.T)                            {}
func TestMySqlWrappingJsonWithBoolean(t *testing.T)                           {}
func TestMySqlWrappingJsonWithBooleanAndIntegerThatLooksLikeOne(t *testing.T) {}
func TestJsonPathEscaping(t *testing.T)                                       {}
func TestMySqlWrappingJson(t *testing.T)                                      {}
func TestBitwiseOperators(t *testing.T)                                       {}
func TestMergeWheresCanMergeWheresAndBindings(t *testing.T)                   {}
func TestProvidingNullWithOperatorsBuildsCorrectly(t *testing.T)              {}
func TestMysqlLock(t *testing.T) {
	//TestMySqlLock
	b := DB.Query()
	b.Select().From("foo").Where("bar", "baz").Lock()
	ElementsShouldMatch(t, []interface{}{"baz"}, b.GetBindings())
	assert.Equal(t, "select * from `foo` where `bar` = ? for update", b.ToSql())

	b1 := DB.Query()
	b1.Select().From("foo").Where("bar", "baz").Lock(false)
	ElementsShouldMatch(t, []interface{}{"baz"}, b1.GetBindings())
	assert.Equal(t, "select * from `foo` where `bar` = ? lock in share mode", b1.ToSql())

	b2 := DB.Query()
	b2.Select().From("foo").Where("bar", "baz").Lock("lock in share mode")
	ElementsShouldMatch(t, []interface{}{"baz"}, b2.GetBindings())
	assert.Equal(t, "select * from `foo` where `bar` = ? lock in share mode", b2.ToSql())

}
func TestBindingOrder(t *testing.T)                                    {}
func TestAddBindingWithArrayMergesBindingsInCorrectOrder(t *testing.T) {}
func TestSubSelect(t *testing.T) {
	b := DB.Table("one").Select("foo", "bar").Where("key", "val")
	b.SelectSub(func(builder *goeloquent.Builder) {
		builder.From("two").Select("baz").Where("subkey", "subval")
	}, "sub")
	assert.Equal(t, "select `foo`, `bar`, (select `baz` from `two` where `subkey` = ?) as `sub` from `one` where `key` = ?", b.ToSql())
	ElementsShouldMatch(t, []interface{}{"subval", "val"}, b.GetBindings())

	b1 := DB.Table("one").Select("foo", "bar").Where("key", "val")
	b2 := DB.Query().From("two").Select("baz").Where("subkey", "subval")
	b1.SelectSub(b2, "sub")
	assert.Equal(t, "select `foo`, `bar`, (select `baz` from `two` where `subkey` = ?) as `sub` from `one` where `key` = ?", b1.ToSql())
	ElementsShouldMatch(t, []interface{}{"subval", "val"}, b1.GetBindings())

	b = GetBuilder()
	tempB := GetBuilder()
	tempB.Select("phone").From("user_info").Limit(1)
	b.Select(map[string]interface{}{
		"address": func(builder *goeloquent.Builder) {
			builder.Select("post_code").From("addresses").Limit(1)
		},
		"phone":   tempB,
		"balance": goeloquent.Raw("(select balance from accounts limit 1) as balance"),
		"time":    "updated_at",
	}).From("users")
	assert.Contains(t, b.ToSql(), "(select balance from accounts limit 1) as balance", `updated_at`, "(select `post_code` from `addresses` limit 1) as `address`", "(select `phone` from `user_info` limit 1) as `phone`")
	//assert.Equal(t, "select (select balance from accounts limit 1) as balance, `updated_at`, (select `post_code` from `addresses` limit 1) as `address`, ( select `phone` from `user_info` limit 1 ) as `phone` from `users`", b.ToSql())
}
func TestSubSelectResetBindings(t *testing.T)                   {}
func TestUppercaseLeadingBooleansAreRemoved(t *testing.T)       {}
func TestLowercaseLeadingBooleansAreRemoved(t *testing.T)       {}
func TestCaseInsensitiveLeadingBooleansAreRemoved(t *testing.T) {}

func TestChunkWithLastChunkComplete(t *testing.T) {
	createUsers, dropUsers := UserTableSql()
	RunWithDB(createUsers, dropUsers, func() {
		var ts []map[string]interface{}
		now := time.Now()
		for i := 0; i < 50; i++ {
			ts = append(ts, map[string]interface{}{
				"name":       fmt.Sprintf("user-%d", i),
				"age":        i,
				"created_at": now,
			})
		}
		result, err := DB.Table("users").Insert(&ts)
		assert.Nil(t, err)
		c, _ := result.RowsAffected()
		assert.Equal(t, int64(len(ts)), c)
		var total int
		totalP := &total
		err = DB.Table("users").OrderBy("id").Chunk(&[]UserScan{}, 10, func(dest interface{}) error {
			us := dest.(*[]UserScan)
			for _, user := range *us {
				assert.Equal(t, user.Name, fmt.Sprintf("user-%d", user.Id-1))
				*totalP++
			}
			return nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 50, *totalP)
	})

}
func TestChunkWithLastChunkPartial(t *testing.T)                  {}
func TestChunkCanBeStoppedByReturningFalse(t *testing.T)          {}
func TestChunkWithCountZero(t *testing.T)                         {}
func TestChunkByIdOnArrays(t *testing.T)                          {}
func TestChunkPaginatesUsingIdWithLastChunkComplete(t *testing.T) {}
func TestChunkPaginatesUsingIdWithLastChunkPartial(t *testing.T)  {}
func TestChunkPaginatesUsingIdWithCountZero(t *testing.T)         {}
func TestChunkPaginatesUsingIdWithAlias(t *testing.T)             {}
func TestChunkPaginatesUsingIdDesc(t *testing.T)                  {}
func TestPaginate(t *testing.T) {
	//TestPaginate
	c, d := UserTableSql()

	RunWithDB(c, d, func() {
		var usersSlice []map[string]interface{}
		var users, users1 []map[string]interface{}
		for i := 0; i < 50; i++ {
			user := map[string]interface{}{
				"name": fmt.Sprintf("bot_%d", i),
				"age":  100 - i,
			}
			usersSlice = append(usersSlice, user)
		}
		c, err := DB.Table("users").Insert(usersSlice)
		assert.Nil(t, err)
		count, err := c.RowsAffected()
		assert.Equal(t, int64(len(usersSlice)), count)
		paginator, err := DB.Table("users").Where("id", ">", 10).Where("id", "<", 28).Paginate(&users, 10, 1)
		assert.Nil(t, err)
		assert.Equal(t, int64(17), paginator.Total)
		assert.Equal(t, int64(1), paginator.CurrentPage)
		assert.Equal(t, 10, len(users))
		for _, u := range users {
			assert.True(t, u["id"].(int64) > 10 && u["id"].(int64) < 28)
		}

		paginator1, err := DB.Table("users").Where("id", ">", 13).Paginate(&users1, 5, 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(37), paginator1.Total)
		assert.Equal(t, int64(2), paginator1.CurrentPage)
		assert.Equal(t, 5, len(users1))
		for _, u := range users1 {
			assert.True(t, u["id"].(int64) > 13)
		}
		//Test paginate slice map
		var uSlice []map[string]interface{}
		p, err := DB.Table("users").Where("id", ">", 13).Paginate(&uSlice, 5, 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(37), p.Total)
		assert.Equal(t, int64(2), p.CurrentPage)
		assert.Equal(t, 5, len(users1))

		//Test paginate select columns
		var uSlice1 []map[string]interface{}
		p1, err := DB.Table("users").
			Where("id", ">", 13).
			Paginate(&uSlice1, 5, 2, []string{"name", "id"})
		assert.Nil(t, err)
		assert.Equal(t, int64(37), p1.Total)
		assert.Equal(t, int64(2), p1.CurrentPage)
		assert.Equal(t, 5, len(uSlice1))

	})

}
func TestPaginateWhenNoResults(t *testing.T)       {}
func TestPaginateWithSpecificColumns(t *testing.T) {}
func TestPaginateWithTotalOverride(t *testing.T)   {}
func TestWhereExpression(t *testing.T)             {}
func TestWhereRowValues(t *testing.T) {
	//TestWhereRowValues
	b := DB.Query()
	_, err := b.Select().From("orders").WhereRowValues([]string{"last_update", "order_number"}, "<", []interface{}{1, 2}).Pretend().Get(&[]map[string]interface{}{})
	assert.Equal(t, "select * from `orders` where (`last_update`, `order_number`) < (?,?)", b.ToSql())
	assert.Nil(t, err)

	b1 := DB.Query()
	_, err = b1.Select().From("orders").WhereRowValues([]string{"last_update", "order_number"}, "<", []interface{}{1, goeloquent.Raw("2")}).Pretend().Get(&[]map[string]interface{}{})
	assert.Equal(t, "select * from `orders` where (`last_update`, `order_number`) < (?,2)", b1.ToSql())
	assert.Nil(t, err)

}
func TestWhereJsonContainsMySql(t *testing.T)         {}
func TestWhereJsonOverlapsMySql(t *testing.T)         {}
func TestWhereJsonDoesntContainMySql(t *testing.T)    {}
func TestWhereJsonDoesntOverlapMySql(t *testing.T)    {}
func TestWhereJsonContainsKeyMySql(t *testing.T)      {}
func TestWhereJsonDoesntContainKeyMySql(t *testing.T) {}
func TestWhereJsonLengthMySql(t *testing.T)           {}

func TestFrom(t *testing.T) {
	b := GetBuilder().Select().FromSub(func(builder *goeloquent.Builder) {
		builder.Select(goeloquent.Raw("max(last_seen_at) as last_seen_at")).From("user_sessions").Where("foo", "=", 1)
	}, "sessions").Where("bar", "<", 10)
	ShouldEqual(t, "select * from (select max(last_seen_at) as last_seen_at from `user_sessions` where `foo` = ?) as `sessions` where `bar` < ?", b)
	ElementsShouldMatch(t, []interface{}{1, 10}, b.GetBindings())
	//

}
func TestFromSub(t *testing.T)                {}
func TestFromSubWithoutBindings(t *testing.T) {}

func TestFromSubWithPrefix(t *testing.T) {
	b1 := GetBuilder()
	b1.Grammar.SetTablePrefix("prefix_")
	b1.Select().FromSub(func(builder *goeloquent.Builder) {
		builder.Select(goeloquent.Raw("max(last_seen_at) as last_seen_at")).From("user_sessions").Where("foo", "=", 1)
	}, "sessions").Where("bar", "<", 10)
	ShouldEqual(t, "select * from (select max(last_seen_at) as last_seen_at from `prefix_user_sessions` where `foo` = ?) as `prefix_sessions` where `bar` < ?", b1)
}
func TestFromRaw(t *testing.T) {
	b2 := GetBuilder().Select().FromRaw("(select max(last_seen_at) as last_seen_at from `user_sessions`) as `sessions`)")
	ShouldEqual(t, "select * from (select max(last_seen_at) as last_seen_at from `user_sessions`) as `sessions`)", b2)

	b3 := GetBuilder().Select().FromRaw("(select max(last_seen_at) as last_seen_at from `user_sessions`) as `sessions`)").Where("last_seen_at", ">", 1520652582)
	ShouldEqual(t, "select * from (select max(last_seen_at) as last_seen_at from `user_sessions`) as `sessions`) where `last_seen_at` > ?", b3)
	ElementsShouldMatch(t, []interface{}{1520652582}, b3.GetBindings())
}
func TestFromRawWithWhereOnTheMainQuery(t *testing.T) {

}
func TestUseIndexMySql(t *testing.T)    {}
func TestForceIndexMySql(t *testing.T)  {}
func TestIgnoreIndexMySql(t *testing.T) {}
func TestOrderByInvalidDirectionParam(t *testing.T) {
	//TestOrderByInvalidDirectionParam
	assert.PanicsWithErrorf(t, "wrong order direction: asec", func() {
		GetBuilder().Select().From("users").OrderBy("age", "asec")
	}, "wrong order should panic with msg:[wrong order direction: asec]")

}
func TestNestedWhereBindings(t *testing.T) {
	//TestNestedWhereBindings
	b := GetBuilder().Where("email", "=", "foo").Where(func(builder *goeloquent.Builder) {
		builder.SelectRaw("?", []interface{}{goeloquent.Expression("ignore")}).Where("name", "=", "bar")
	})
	ElementsShouldMatch(t, []interface{}{"foo", "bar"}, b.GetBindings())
}
func TestClone(t *testing.T) {
	//TestClone
	b := GetBuilder().Select().From("users")
	cb := goeloquent.Clone(b)
	cb.Where("email", "foo")
	ShouldEqual(t, "select * from `users`", b)
	ShouldEqual(t, "select * from `users` where `email` = ?", cb)
	assert.NotEqual(t, b, cb)
}

func TestCloneWithout(t *testing.T) {
	//TestCloneWithout
	b := GetBuilder().Select().From("users").Where("email", "foo").OrderBy("email")
	cb := goeloquent.CloneWithout(b, goeloquent.TYPE_ORDER)
	ShouldEqual(t, "select * from `users` where `email` = ? order by `email` asc", b)
	ShouldEqual(t, "select * from `users` where `email` = ?", cb)

}
func TestCloneWithoutBindings(t *testing.T) {
	b := GetBuilder().Select().From("users").Where("email", "foo").OrderBy("email")
	cb := b.Clone().CloneWithout(goeloquent.TYPE_WHERE).CloneWithoutBindings(goeloquent.TYPE_WHERE)
	ShouldEqual(t, "select * from `users` where `email` = ? order by `email` asc", b)
	ElementsShouldMatch(t, []interface{}{"foo"}, b.GetBindings())

	ShouldEqual(t, "select * from `users` order by `email` asc", cb)
	ElementsShouldMatch(t, []interface{}{}, cb.GetBindings())
}

func TestPluckMethod(t *testing.T) {
	createUsers, dropUsers := UserTableSql()
	now := time.Now()
	u1 := map[string]interface{}{
		"name":       "go-eloquent",
		"age":        18,
		"created_at": now,
	}
	RunWithDB(createUsers, dropUsers, func() {
		var names, ns []string
		u2 := map[string]interface{}{
			"name":       "senc",
			"age":        12,
			"created_at": now,
		}
		us := []map[string]interface{}{
			u1, u2,
		}
		DB.Table("users").Insert(us)
		b := DB.Table("users")
		b.Pluck(&names, "name")
		assert.Equal(t, "select `name` from `users`", b.ToSql())
		assert.Equal(t, []string{"go-eloquent", "senc"}, names)
		b1 := DB.Table("users").Where("age", ">=", 18)
		b1.Pluck(&ns, "name")
		assert.Equal(t, "select `name` from `users` where `age` >= ?", b1.ToSql())
		assert.Equal(t, []string{"go-eloquent"}, ns)
	})
}
func TestStruct(t *testing.T) {
	//Test simple struct
	type Temp struct {
		Id     int64  `goelo:"column:tid;primaryKey"`
		Name   string `goelo:"column:name"`
		Status int8   `goelo:"column:status"`
	}
	type TempB struct {
		Id   int64  `goelo:"column:id;primaryKey"`
		Name string `goelo:"column:name"`
	}

	var tagt, tag Temp
	var tagtt TempB
	tag.Name = "test"
	tagt.Name = "test"
	tagt.Status = 1
	tagtt.Name = "test"

	//Test ignore zero value for struct
	b := DB.Table("tag")
	b.Insert(tag)
	ElementsShouldMatch(t, []interface{}{"test"}, b.GetBindings())
	assert.Equal(t, "insert into `tag` (`name`) values (?)", b.PreparedSql)

	b1 := DB.Table("tag")
	b1.Only("name").Insert(tagt)
	ElementsShouldMatch(t, []interface{}{"test"}, b1.GetBindings())
	assert.Equal(t, "insert into `tag` (`name`) values (?)", b1.PreparedSql)

	b2 := DB.Model()
	b2.Except("id").Insert(&tagtt)
	ElementsShouldMatch(t, []interface{}{"test"}, b2.GetBindings())
	assert.Equal(t, "insert into `temp_b` (`name`) values (?)", b2.PreparedSql)
}

func TestChunk(t *testing.T) {
	//TestChunkWithLastChunkComplete
	createUsers, dropUsers := UserTableSql()
	//TestChunkCanBeStoppedByReturningError
	RunWithDB(createUsers, dropUsers, func() {
		var ts []map[string]interface{}
		now := time.Now()
		for i := 0; i < 50; i++ {
			ts = append(ts, map[string]interface{}{
				"name":       fmt.Sprintf("user-%d", i),
				"age":        i,
				"created_at": now,
			})
		}
		result, err := DB.Table("users").Insert(&ts)
		assert.Nil(t, err)
		c, _ := result.RowsAffected()
		assert.Equal(t, int64(len(ts)), c)
		err = DB.Model(&UserScan{}).OrderBy("id").Chunk(&[]*UserScan{}, 10, func(dest interface{}) error {
			us := dest.(*[]*UserScan)
			for _, scan := range *us {
				scan.Age = 100
				scan.Status = 1
				_, e := DB.Save(scan)
				assert.Nil(t, e)
				//assert.Equal()
				if scan.Id == 5 {
					return errors.New("stopped")
				}
			}
			return nil
		})
		assert.Equal(t, "stopped", err.Error())
		var count int
		DB.Table("users").Where("status", 1).Count(&count)
		assert.Equal(t, 5, count)

	})
}
func TestChunkById(t *testing.T) {
	//TestChunkWithLastChunkComplete
	createUsers, dropUsers := UserTableSql()
	RunWithDB(createUsers, dropUsers, func() {
		var ts []map[string]interface{}
		now := time.Now()
		for i := 0; i < 50; i++ {
			ts = append(ts, map[string]interface{}{
				"name":       fmt.Sprintf("user-%d", i),
				"age":        i,
				"created_at": now,
			})
		}
		result, err := DB.Table("users").Insert(&ts)
		assert.Nil(t, err)
		c, _ := result.RowsAffected()
		assert.Equal(t, int64(len(ts)), c)
		var total int
		err = DB.Table("users").ChunkById(&[]UserScan{}, 10, func(dest interface{}) error {
			us := dest.(*[]UserScan)
			for _, user := range *us {
				assert.Equal(t, user.Name, fmt.Sprintf("user-%d", user.Age))
				total++
			}
			return nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 50, total)

		total = 0
		err = DB.Table("users").OrderBy("id").ChunkById(&[]map[string]interface{}{}, 10, func(dest interface{}) error {
			us := dest.(*[]map[string]interface{})
			for _, user := range *us {
				assert.Equal(t, string(user["name"].([]uint8)), fmt.Sprintf("user-%d", user["age"]))
				total++
			}
			return nil
		}, "id")
		assert.Nil(t, err)
		assert.Equal(t, 50, total)
	})
	//TestChunkCanBeStoppedByReturningError
	RunWithDB(createUsers, dropUsers, func() {
		var ts []map[string]interface{}
		now := time.Now()
		for i := 0; i < 50; i++ {
			ts = append(ts, map[string]interface{}{
				"name":       fmt.Sprintf("user-%d", i),
				"age":        i,
				"created_at": now,
			})
		}
		result, err := DB.Table("users").Insert(&ts)
		assert.Nil(t, err)
		c, _ := result.RowsAffected()
		assert.Equal(t, int64(len(ts)), c)
		err = DB.Model(&UserScan{}).OrderBy("id").Chunk(&[]UserScan{}, 10, func(dest interface{}) error {
			us := dest.(*[]UserScan)
			for i := 0; i < len(*us); i++ {
				(*us)[i].Age = 1
				(*us)[i].Status = 2
				DB.Save(&((*us)[i]))
				if (*us)[i].Id == 5 {
					return errors.New("stopped")
				}
			}
			return nil
		})
		assert.Equal(t, "stopped", err.Error())
		var count int
		DB.Table("users").Where("status", 2).Count(&count)
		assert.Equal(t, 5, count)

	})
}
func TestBase(t *testing.T) {
	createUsers, dropUsers := UserTableSql()
	RunWithDB(createUsers, dropUsers, func() {
		var ts []map[string]interface{}
		now := time.Now()
		for i := 0; i < 5; i++ {
			ts = append(ts, map[string]interface{}{
				"name":       fmt.Sprintf("user-%d", i),
				"age":        i,
				"created_at": now,
			})
		}
		result, err := DB.Table("users").Insert(&ts)
		assert.Nil(t, err)
		c, _ := result.RowsAffected()
		assert.Equal(t, int64(len(ts)), c)

		//TestImplode
		b5 := DB.Query()
		implode, err := b5.From("users").Where("id", "<", 3).Implode("name")
		assert.Nil(t, err)
		assert.Equal(t, "user-0user-1", implode)

		b6 := DB.Query()
		implode2, err := b6.From("users").Where("id", "<", 3).Implode("name", ",")
		assert.Nil(t, err)
		assert.Equal(t, "user-0,user-1", implode2)

		//Testfindmap
		var m = make(map[string]interface{})
		var m1 = make(map[string]interface{})
		b7 := DB.Query()
		a := int64(0)
		a1 := int64(0)
		b := ""
		b7.From("users").Mapping(map[string]interface{}{
			"id":   &a,
			"age":  &a1,
			"name": &b,
		}).Find(&m, 3)
		assert.Equal(t, "select * from `users` where `id` = ? limit 1", b7.ToSql())
		assert.Equal(t, int64(2), m["age"])
		assert.Equal(t, int64(3), m["id"])
		assert.Equal(t, "user-2", m["name"])
		b10 := DB.Query()
		b10.From("users").Mapping(map[string]interface{}{
			"id":   int64(0),
			"age":  int64(0),
			"name": "",
		}).Find(&m1, 4)
		assert.Equal(t, "select * from `users` where `id` = ? limit 1", b10.ToSql())
		assert.Equal(t, int64(3), m1["age"])
		assert.Equal(t, int64(4), m1["id"])
		assert.Equal(t, "user-3", m1["name"])
		//Testgetmap
		var ms []map[string]interface{}
		b8 := DB.Query()
		_, err = b8.From("users").Where("id", "<", 10).Limit(2).Mapping(map[string]interface{}{
			"id":   int64(0),
			"name": "",
		}).Get(&ms, "id", "name")
		assert.Nil(t, err)
		assert.Equal(t, "select `id`, `name` from `users` where `id` < ? limit 2", b8.ToSql())
		assert.Equal(t, 2, len(ms))
		for _, mt := range ms {
			assert.Equal(t, fmt.Sprintf("user-%d", mt["id"].(int64)-1), mt["name"])
			assert.IsType(t, a, mt["id"])
		}

		//Test findmany
		var users []map[string]interface{}
		b9 := DB.Query()
		rows, err := b9.From("users").Find(&users, 1, 2, 3, 4)
		c = rows.Count
		assert.Nil(t, err)
		assert.Equal(t, rows.Count, c)
		assert.Equal(t, users[0]["id"], int64(1))
		assert.Equal(t, "select * from `users` where `id` in (?,?,?,?)", b9.PreparedSql)

	})

}

func TestUpdateOrInsert(t *testing.T) {
	createUsers, dropUsers := UserTableSql()
	RunWithDB(createUsers, dropUsers, func() {
		var ts []map[string]interface{}
		now := time.Now()
		for i := 0; i < 5; i++ {
			ts = append(ts, map[string]interface{}{
				"name":       fmt.Sprintf("user-%d", i),
				"age":        i,
				"created_at": now,
			})
		}
		result, err := DB.Table("users").Insert(&ts)
		assert.Nil(t, err)
		c, _ := result.RowsAffected()
		assert.Equal(t, int64(len(ts)), c)

		//Test updated
		q := DB.Query()
		updated, err := q.Table("users").UpdateOrInsert(map[string]interface{}{
			"age":  2,
			"name": "user-2",
		}, map[string]interface{}{
			"age": 18,
		})
		assert.Nil(t, err)
		assert.True(t, updated)
		assert.Contains(t, q.PreparedSql, "update `users` set `age` = ? where (")
		assert.Equal(t, []interface{}{18, 2, "user-2"}, q.GetBindings())
		exist, err := DB.Query().Table("users").Where(map[string]interface{}{
			"age":  18,
			"name": "user-2",
		}).Exists()
		assert.True(t, exist)

		//Test inserted
		q1 := DB.Query()
		updated, err = q1.Table("users").UpdateOrInsert(map[string]interface{}{
			"age": 7,
		}, map[string]interface{}{
			"age":  18,
			"name": "user-7",
		})
		assert.Nil(t, err)
		assert.False(t, updated)
		assert.Equal(t, "insert into `users` (`age`, `name`) values (?, ?)", q1.PreparedSql)
		assert.Equal(t, []interface{}{18, "user-7"}, q1.GetBindings())
		exist, err = DB.Query().Table("users").Where(map[string]interface{}{
			"age":  18,
			"name": "user-7",
		}).Exists()
		assert.True(t, exist)
		assert.Nil(t, err)
	})
}
func TestUpdateOrCreate(t *testing.T) {
	//createUsers, dropUsers := UserTableSql()
	//RunWithDB(createUsers, dropUsers, func() {
	//	//Test found
	//	user := UserScan{
	//		Name:   "user-x",
	//		Status: 1,
	//	}
	//	_, err := DB.Save(&user)
	//	assert.Nil(t, err)
	//	q := DB.Model(&user)
	//	updated, err := q.UpdateOrCreate(&user, map[string]interface{}{
	//		"age":    200,
	//		"status": 1,
	//	}, map[string]interface{}{
	//		"name":   "impossible",
	//		"status": 2,
	//	})
	//	assert.Nil(t, err)
	//	assert.True(t, updated)
	//	assert.Equal(t, "select * from `users` where (`age` = ? and `status` = ?) limit 1", q.PreparedSql)
	//	assert.Equal(t, []interface{}{uint8(200), 1}, q.GetBindings())
	//	exist, err := DB.Query().Table("users").Where(map[string]interface{}{
	//		"age":    uint8(200),
	//		"name":   "impossible",
	//		"status": 2,
	//	}).Exists()
	//	assert.Nil(t, err)
	//	assert.True(t, exist)
	//
	//	//Test created
	//	var user2 User
	//	q1 := DB.Model(&user2)
	//	updated, err = q1.UpdateOrCreate(&user2, map[string]interface{}{
	//		"age": -1,
	//	}, map[string]interface{}{
	//		"age": uint8(28),
	//		"name": sql.NullString{
	//			String: "user-28",
	//			Valid:  true,
	//		},
	//	})
	//	assert.Nil(t, err)
	//	assert.False(t, updated)
	//	assert.Equal(t, "select * from `users` where (`age` = ?) limit 1", q1.PreparedSql)
	//	assert.Equal(t, []interface{}{-1}, q1.GetBindings())
	//	exist, err = DB.Query().Table("users").Where(map[string]interface{}{
	//		"age":  uint8(28),
	//		"name": "user-28",
	//	}).Exists()
	//	assert.True(t, exist)
	//	assert.Nil(t, err)
	//})
}

//	func TestFirstOrNew(t *testing.T) {
//		createUsers, dropUsers := UserTableSql()
//		RunWithDB(createUsers, dropUsers, func() {
//			const StatusPending = 1
//			//Test found
//			user := User{
//				Name: "john@gmail.com",
//				Age:      200,
//				Status:   1,
//			}
//			DB.Save(&user)
//			var tu User
//			q := DB.Query()
//			found, err := q.FirstOrNew(&tu, map[string]interface{}{
//				"name": "john@gmail.com",
//			}, map[string]interface{}{
//				"name": sql.NullString{
//					String: "john@gmail.com", Valid: true,
//				},
//				"age":    18,
//				"status": StatusPending,
//			})
//			assert.Nil(t, err)
//			assert.True(t, found)
//			assert.ObjectsAreEqualValues(user, tu)
//			assert.Equal(t, "select * from `users` where (`name` = ?) limit 1", q.PreparedSql)
//
//			//Test new
//			var t2 User
//			q2 := DB.Query()
//			found, err = q2.FirstOrNew(&t2, map[string]interface{}{
//				"name": sql.NullString{
//					String: "john2@gmail.com", Valid: true,
//				},
//			}, map[string]interface{}{
//				"name": sql.NullString{
//					String: "john2@gmail.com", Valid: true,
//				},
//				"age":    18,
//				"status": StatusPending,
//			})
//			assert.Nil(t, err)
//			assert.False(t, found)
//			assert.Equal(t, "select * from `users` where (`name` = ?) limit 1", q.PreparedSql)
//			assert.ObjectsAreEqualValues(t2, User{
//				Name: "john2@gmail.com",
//				Age:      18,
//				Status:   StatusPending,
//			})
//		})
//	}
//
//	func TestFirstOrCreate(t *testing.T) {
//		createUsers, dropUsers := UserTableSql()
//		RunWithDB(createUsers, dropUsers, func() {
//			const StatusPending = 1
//			//Test found
//			user := User{
//				Name: sql.NullString{
//					String: "john@gmail.com",
//					Valid:  true,
//				},
//				Age:    200,
//				Status: 1,
//			}
//			DB.Save(&user)
//			var tu User
//			q := DB.Query()
//			found, err := q.FirstOrCreate(&tu, map[string]interface{}{
//				"name": "john@gmail.com",
//			}, map[string]interface{}{
//				"name": sql.NullString{
//					String: "john@gmail.com", Valid: true,
//				},
//				"age":    18,
//				"status": StatusPending,
//			})
//			assert.Nil(t, err)
//			assert.True(t, found)
//			assert.ObjectsAreEqualValues(user, tu)
//			assert.Equal(t, "select * from `users` where (`name` = ?) limit 1", q.PreparedSql)
//
//			//Test new
//			var t2 User
//			q2 := DB.Query()
//			found, err = q2.FirstOrCreate(&t2, map[string]interface{}{
//				"name": sql.NullString{
//					String: "john2@gmail.com", Valid: true,
//				},
//			}, map[string]interface{}{
//				"name": sql.NullString{
//					String: "john2@gmail.com", Valid: true,
//				},
//				"age":    18,
//				"status": StatusPending,
//			})
//			assert.Nil(t, err)
//			assert.False(t, found)
//			assert.Greater(t, t2.ID, int64(0))
//			assert.Equal(t, "select * from `users` where (`name` = ?) limit 1", q.PreparedSql)
//
//		})
//	}
func TestCloneBuilder(t *testing.T) {

	b := GetBuilder()
	b.Model(&User{}).Select("email").AddSelect(goeloquent.Raw("DATE(created_at)")).
		Where("id", ">", 1).OrWhereIn("role", []interface{}{"ADMIN", "MANAGER", "OWNER"}).WherePivot("model_has_roles.status", 1).
		Only("id", "name").Limit(5).Offset(7).OrderByDesc("level")

	newB := b.Clone()
	assert.Equal(t, b.TableAlias, newB.TableAlias)
	assert.Equal(t, b.FromTable, newB.FromTable)
	assert.Equal(t, b.LimitNum, newB.LimitNum)
	assert.Equal(t, b.OffsetNum, newB.OffsetNum)
	for i, _ := range b.Columns {
		assert.Equal(t, b.Columns[i], newB.Columns[i])
	}
	for k, bindings := range b.Bindings {
		for i, binding := range bindings {
			assert.Equal(t, binding, newB.Bindings[k][i])
		}
	}
	for i, _ := range b.OnlyColumns {
		assert.Equal(t, b.OnlyColumns[i], newB.OnlyColumns[i])
	}
	for i, _ := range b.Wheres {
		assert.EqualValues(t, b.Wheres[i], newB.Wheres[i])
	}

	for i, _ := range b.Orders {
		assert.EqualValues(t, b.Orders[i], newB.Orders[i])
	}
}
func TestContext(t *testing.T) {
	//b := GetBuilder()
	//b.WithContext()
	//assert.Equal(t, "test", b.GetContext())
}
func TestBeforeCallbacks(t *testing.T) {
	//TODO:TestBeforeCallbacks
}
func TestAfterCallbacks(t *testing.T) {
	//TODO:TestAfterCallbacks
}

//func TestWhereStruct(t *testing.T) {
//	b := GetBuilder()
//
//	type OrderQueryParams struct {
//		VendorId      int64  `form:"pid" validate:"required|int" `
//		ChannelType   int    `form:"type" validate:"required|int" `
//		VendorOrderId string `form:"out_trade_no" validate:"required" `
//		OrderDesc     string `form:"name" validate:"required" column:"name"`
//		ClientIp      string `form:"clientip" validate:"required|ip"`
//		OrderStatus   *int   `form:"order_status" column:"status"`
//	}
//	pending := 0
//	test := OrderQueryParams{
//		VendorId:      0,
//		ChannelType:   1,
//		VendorOrderId: "2208091533zcjgk",
//		OrderDesc:     "macbook",
//		ClientIp:      "192.168.0.1",
//		OrderStatus:   &pending,
//	}
//	var res = make(map[string]interface{})
//	b.Table("orders").Where(&test).First(&res)
//	assert.Equal(t, "select * from `orders` where (`vendor_order_id` = ? and `name` = ? and `client_ip` = ? and `status` = ? and `channel_type` = ?) limit 1", b.ToSql())
//	assert.Equal(t, []interface{}{"2208091533zcjgk", "macbook", "192.168.0.1", &pending, 1}, b.GetBindings())
//
//}

//pending
//TODO: testTapCallback
//TODO: testAggregateResetFollowedByGet
//TODO: testAggregateResetFollowedBySelectGet
//TODO: testAggregateResetFollowedByGetWithColumns
//TODO: testInsertUsingInvalidSubquery
//TODO: testInsertGetIdMethodRemovesExpressions
//TODO: testUpdateOrInsertMethodWorksWithEmptyUpdateValues
//TODO: testTruncateMethod
//TODO: testPreserveAddsClosureToArray
//TODO: testApplyPreserveCleansArray
//TODO: testPreservedAreAppliedByToSql
//TODO: testPreservedAreAppliedByInsert
//TODO: testPreservedAreAppliedByInsertGetId
//TODO: testPreservedAreAppliedByInsertUsing
//TODO: testPreservedAreAppliedByUpsert
//TODO: testPreservedAreAppliedByUpdate
//TODO: testPreservedAreAppliedByDelete
//TODO: testPreservedAreAppliedByTruncate
//TODO: testPreservedAreAppliedByExists
//TODO: testMergeWheresCanMergeWheresAndBindings
//TODO: testProvidingNullWithOperatorsBuildsCorrectly
//TODO: testDynamicWhere
//TODO: testDynamicWhereIsNotGreedy
//TODO: testCallTriggersDynamicWhere
//TODO: testSelectWithLockUsesWritePdo
//TODO: testBindingOrder
//TODO: testAddBindingWithArrayMergesBindings
//TODO: testAddBindingWithArrayMergesBindingsInCorrectOrder
//TODO: testMergeBuilders
//TODO: testMergeBuildersBindingOrder
//TODO: testChunkPaginatesUsingIdWithLastChunkComplete
//TODO: testChunkPaginatesUsingIdWithLastChunkPartial
//TODO: testChunkPaginatesUsingIdWithCountZero
//TODO: testChunkPaginatesUsingIdWithAlias
//TODO: testPaginateWithDefaultArguments
//TODO: testPaginateWhenNoResults
//TODO: testCursorPaginate
//TODO: testCursorPaginateMultipleOrderColumns
//TODO: testCursorPaginateWithDefaultArguments
//TODO: testCursorPaginateWhenNoResults
//TODO: testCursorPaginateWithSpecificColumns
//TODO: testCursorPaginateWithMixedOrders
//TODO: testWhereRowValuesArityMismatch
//TODO: testGetCountForPaginationWithUnion
//TODO: testBitwiseOperators
//TODO: testBuilderThrowsExpectedExceptionWithUndefinedMethod
