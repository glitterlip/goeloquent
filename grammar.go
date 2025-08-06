package goeloquent

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Expression string

func Raw(expr string) Expression {
	return Expression(expr)
}

type MysqlGrammar struct {
	Error       error
	TablePrefix string
}

func NewMysqlGrammar() *MysqlGrammar {
	return &MysqlGrammar{}
}
func (g *MysqlGrammar) SetTablePrefix(prefix string) {
	g.TablePrefix = prefix
}
func (g *MysqlGrammar) AddError(err error) {
	if err != nil {
		if g.Error == nil {
			g.Error = err
		} else {
			g.Error = fmt.Errorf("%v; %w", g.Error, err)
		}
	}
}
func (g *MysqlGrammar) GetError() error {
	return g.Error
}
func (g *MysqlGrammar) Wrap(value interface{}) string {
	switch value.(type) {
	case Expression:
		return string(value.(Expression))
	case string:
		if strings.Contains(value.(string), " as ") || strings.Contains(value.(string), " AS ") {
			return g.WrapAliasedValue(value.(string))
		} else if IsJsonSelector(value.(string)) {
			return g.WrapJsonSelector(value.(string))
		} else {
			return g.WrapSegments(strings.Split(value.(string), "."))
		}
	}
	g.AddError(errors.New(fmt.Sprintf("wrap value type %T is not string or expression", value)))

	return ""
}

func (g *MysqlGrammar) WrapAliasedValue(value string) string {

	var segments []string
	if strings.Contains(value, " AS ") {
		segments = strings.SplitN(value, " AS ", 2)
	} else {
		segments = strings.SplitN(value, " as ", 2)
	}
	return g.Wrap(segments[0]) + " as " + g.WrapValue(segments[1])
}

func (g *MysqlGrammar) WrapAliasedTable(table string, prefix ...string) string {

	var tablePrefix string
	if len(prefix) > 0 && len(prefix[0]) > 0 {
		tablePrefix = prefix[0]
	}
	var segments []string
	if strings.Contains(table, " AS ") {
		segments = strings.SplitN(table, " AS ", 2)
	} else {
		segments = strings.SplitN(table, " as ", 2)
	}
	return g.WrapTable(segments[0], tablePrefix) + " as " + g.WrapValue(tablePrefix+segments[1])

}

func (g *MysqlGrammar) WrapValue(value string) string {
	if value == "*" {
		return value
	} else {
		return "`" + strings.ReplaceAll(value, "`", "``") + "`"
	}
}

func (g *MysqlGrammar) WrapJsonSelector(value string) string {

	field, path := g.WrapJsonFieldAndPath(value)
	return "json_unquote(json_extract(" + field + path + "))"
}
func (g *MysqlGrammar) WrapJsonFieldAndPath(value interface{}) (string, string) {
	var str string
	if expression, ok := value.(Expression); ok {
		str = string(expression)
	} else {
		str = value.(string)
	}
	parts := strings.SplitN(str, "->", 2)

	if len(parts) > 1 {
		return g.Wrap(parts[0]), ", " + g.WrapJsonPath(parts[1], "->")
	} else {
		return g.Wrap(parts[0]), ""
	}
}
func (g *MysqlGrammar) WrapJsonPath(value, delimiter string) string {
	re := regexp.MustCompile(`([\\\\]+)?\\'`)
	value = re.ReplaceAllString(value, "''")

	segments := strings.Split(value, delimiter)
	for i, segment := range segments {
		segments[i] = g.WrapJsonPathSegment(segment)
	}

	jsonPath := strings.Join(segments, ".")

	if strings.HasPrefix(jsonPath, "[") {
		return fmt.Sprintf("'$%s'", jsonPath)
	}
	return fmt.Sprintf("'$.%s'", jsonPath)
}
func (g *MysqlGrammar) WrapJsonPathSegment(segment string) string {
	re := regexp.MustCompile(`(\[[^\]]+\])+$`)
	matches := re.FindStringSubmatch(segment)

	if len(matches) > 0 {
		key := strings.TrimSuffix(segment, matches[0])
		if key != "" {
			return fmt.Sprintf(`"%s"%s`, key, matches[0])
		}
		return matches[0]
	}

	return fmt.Sprintf(`"%s"`, segment)
}
func (g *MysqlGrammar) WrapJsonBooleanSelector(value interface{}) string {
	field, path := g.WrapJsonFieldAndPath(value)
	return "json_extract(" + field + path + ")"
}
func (g *MysqlGrammar) WrapJsonBooleanValue(value string) string {
	return value
}
func (g *MysqlGrammar) Columnize(value []interface{}) string {
	var columns []string
	for _, v := range value {
		switch v.(type) {
		case string, Expression:
			columns = append(columns, g.Wrap(v))
		default:
			g.AddError(errors.New(fmt.Sprintf("columnize value type %T is not string or expression", v)))
		}
	}
	return strings.Join(columns, ", ")
}

func (g *MysqlGrammar) Parameter(value interface{}) string {
	if expression, ok := value.(Expression); ok {
		return string(expression)
	}
	return "?"
}

func (g *MysqlGrammar) Parameterize(value []interface{}) string {
	var parameters []string
	for _, v := range value {
		if expression, ok := v.(Expression); ok {
			parameters = append(parameters, string(expression))
		} else {
			parameters = append(parameters, g.Parameter(v))
		}
	}
	return strings.Join(parameters, ", ")
}

func (g *MysqlGrammar) QuoteString(value interface{}) string {

	switch value.(type) {
	case string:
		return "'" + value.(string) + "'"
	case []string:
		var quoted []string
		for _, v := range value.([]string) {
			quoted = append(quoted, g.QuoteString(v))
		}
		return strings.Join(quoted, ", ")
	default:
		g.AddError(errors.New(fmt.Sprintf("quote string value type %T is not string or array of string", value)))
	}
	return ""

}
func (g *MysqlGrammar) GetDateFormat() string {
	return "2006-01-02 15:04:05"
}

func (g *MysqlGrammar) WrapTable(tableName interface{}, prefix ...string) string {
	var tablePrefix string
	if len(prefix) > 0 && len(prefix[0]) > 0 {
		tablePrefix = prefix[0]
	} else {
		tablePrefix = g.TablePrefix
	}
	switch t := tableName.(type) {
	case Expression:
		return string(t)
	case string:
		if strings.Contains(t, " as ") || strings.Contains(t, " AS ") {
			return g.WrapAliasedTable(t, tablePrefix)
		}
		if strings.Contains(t, ".") {

			segments := strings.SplitN(t, ".", 2)

			return g.WrapValue(segments[0]) + "." + g.WrapValue(tablePrefix+segments[1])
		}

	}
	return g.WrapValue(tablePrefix + tableName.(string))
}
func (g *MysqlGrammar) WrapSegments(segments []string) string {
	if len(segments) == 0 {
		return ""
	}
	if len(segments) == 1 {
		return g.WrapValue(segments[0])
	}
	var wrappedSegments []string
	for i, segment := range segments {
		if i == 0 {
			wrappedSegments = append(wrappedSegments, g.WrapTable(segment))
		} else {
			wrappedSegments = append(wrappedSegments, g.WrapValue(segment))
		}
	}
	return strings.Join(wrappedSegments, ".")
}

func (g *MysqlGrammar) CompileSelect(query *QueryBuilder) string {

	if len(query.Havings) > 0 && len(query.Aggregates.AggregateName) > 0 {
		return g.CompileUnionAggregate(query)
	}
	if query.Grouplimit.Value > 0 {
		if len(query.Columns) == 0 {
			query.Columns = []interface{}{"*"}
		}
		return g.CompileGroupLimit(query)
	}
	columns := query.Columns
	if len(query.Columns) == 0 {
		query.Select("*")
	}
	parts := g.CompileComponents(query)
	query.Columns = columns
	return Concatenate(parts)
}
func Concatenate(parts map[Component]string) string {
	var sb strings.Builder
	for _, component := range SelectComponents {
		if part, ok := parts[component]; ok && len(part) > 0 {
			sb.WriteString(strings.TrimSuffix(part, " ") + " ")
		}
	}

	return strings.Trim(sb.String(), " ")
}
func (g *MysqlGrammar) CompileDelete(query *QueryBuilder) string {

	table := g.WrapTable(query.FromTable)
	where := g.CompileWheres(query)
	if len(query.Joins) > 0 {
		return g.CompileDeleteWithJoins(query, table, where)
	} else {
		return g.CompileDeleteWithoutJoins(query, table, where)
	}
}
func (g *MysqlGrammar) CompileDeleteWithJoins(query *QueryBuilder, table, where string) string {
	ts := strings.Split(table, " as ")
	alias := ts[len(ts)-1]
	joins := g.CompileJoins(query)
	return fmt.Sprintf("delete %s from %s %s %s", alias, table, joins, where)
}
func (g *MysqlGrammar) CompileDeleteWithoutJoins(query *QueryBuilder, table, where string) string {
	sqlStr := "delete from " + table + " " + where
	if len(query.Orders) > 0 {
		sqlStr += " " + g.CompileOrders(query)
	}
	if query.LimitNum > 0 {
		sqlStr += " " + g.CompileLimit(query)
	}
	return strings.TrimSuffix(sqlStr, " ")
}
func (g *MysqlGrammar) CompileExists(query *QueryBuilder) string {
	return "select exists(" + g.CompileSelect(query) + ") as " + g.Wrap("exists")
}
func (g *MysqlGrammar) CompileInsert(query *QueryBuilder, values []map[string]interface{}) (string, []interface{}) {
	var res []interface{}
	first := values[0]
	var keys []string
	for key, _ := range first {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var cols []interface{}
	for _, key := range keys {
		cols = append(cols, key)
	}
	columns := g.Columnize(cols)
	var sqls []string
	for _, value := range values {
		sql := "("
		for _, key := range keys {
			sql += g.Parameter(value[key])
			if _, ok := value[key].(Expression); !ok {
				res = append(res, value[key])
			}
			sql += ", "
		}
		sql = strings.TrimSuffix(sql, ", ")
		sql += ")"
		sqls = append(sqls, sql)
	}

	return fmt.Sprintf("insert into %s (%s) values %s", g.WrapTable(query.FromTable), columns, strings.Join(sqls, ", ")), res
}
func (g *MysqlGrammar) CompileUpsert(query *QueryBuilder, values []map[string]interface{}, uniqueBy []string, update []string) (string, []interface{}) {
	sql, bindings := g.CompileInsert(query, values)
	sql += " on duplicate key update "
	var parts []string
	for _, value := range update {
		parts = append(parts, g.Wrap(value)+" = values("+g.Wrap(value)+")")
	}
	return sql + strings.Join(parts, ", "), bindings

}
func (g *MysqlGrammar) CompileInsertGetId(query *QueryBuilder, values map[string]interface{}) (string, []interface{}) {
	return g.CompileInsert(query, []map[string]interface{}{values})
}
func (g *MysqlGrammar) CompileInsertOrIgnore(query *QueryBuilder, values []map[string]interface{}) (string, []interface{}) {
	str, bindings := g.CompileInsert(query, values)
	return strings.Replace(str, "insert", "insert ignore", 1), bindings
}

func (g *MysqlGrammar) CompileUpdate(query *QueryBuilder, values map[string]interface{}) (string, []interface{}) {

	var sql string
	table := g.WrapTable(query.FromTable)
	columns, bindings := g.CompileUpdateColumns(query, values)
	where := g.CompileWheres(query)
	if len(query.Joins) > 0 {
		sql = g.CompileUpdateWithJoins(query, table, columns, where)
	} else {
		sql = g.CompileUpdateWithoutJoins(query, table, columns, where)
	}
	return sql, bindings
}
func (g *MysqlGrammar) CompileUpdateColumns(query *QueryBuilder, values map[string]interface{}) (string, []interface{}) {
	var parts []string
	var bindings []interface{}
	var keys []string
	for key, _ := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		switch value := values[key].(type) {
		case *QueryBuilder:
			parts = append(parts, g.Wrap(key)+" = ("+g.CompileSelect(value)+")")
			bindings = append(bindings, value.RawBindings...)
			continue
		case QueryBuilder:
			parts = append(parts, g.Wrap(key)+" = ("+g.CompileSelect(&value)+")")
			bindings = append(bindings, value.RawBindings...)
			continue
		case func(*QueryBuilder):
			subQuery := NewQueryBuilder()
			value(subQuery)
			parts = append(parts, g.Wrap(key)+" = ("+g.CompileSelect(subQuery)+")")
			bindings = append(bindings, subQuery.RawBindings...)
			continue
		case func(*QueryBuilder) *QueryBuilder:
			subQuery := NewQueryBuilder()
			subQuery = value(subQuery)
			parts = append(parts, g.Wrap(key)+" = ("+g.CompileSelect(subQuery)+")")
			bindings = append(bindings, subQuery.GetBindings()...)
			continue
		}

		if IsJsonSelector(key) {
			parts = append(parts, g.CompileJsonUpdateColumn(key, values[key]))
			if _, ok := values[key].(bool); !ok {
				if _, ok := values[key].(Expression); !ok {
					bindings = append(bindings, values[key])
				}
			}
		} else {
			parts = append(parts, g.Wrap(key)+" = "+g.Parameter(values[key]))
			if _, ok := values[key].(Expression); !ok {
				bindings = append(bindings, values[key])
			}
		}

	}

	return strings.Join(parts, ", "), bindings
}
func (g *MysqlGrammar) CompileUpdateWithoutJoins(query *QueryBuilder, table, columns, where string) string {
	sql := fmt.Sprintf("update %s set %s %s", table, columns, where)
	if len(query.Orders) > 0 {
		sql += " " + g.CompileOrders(query)
	}
	if query.LimitNum > 0 {
		sql += " " + g.CompileLimit(query)
	}
	return sql
}
func (g *MysqlGrammar) CompileUpdateWithJoins(query *QueryBuilder, table, columns, where string) string {
	joins := g.CompileJoins(query)
	return strings.Trim(fmt.Sprintf("update %s %s set %s %s", table, joins, columns, where), " ")
}
func (g *MysqlGrammar) CompileJsonUpdateColumn(key string, value interface{}) string {
	switch value.(type) {
	case bool:
		value = strconv.FormatBool(value.(bool))
	case []interface{}:
		value = "cast(? as json)"
	default:
		value = g.Parameter(value)
	}

	field, path := g.WrapJsonFieldAndPath(key)
	return fmt.Sprintf("%s = json_set(%s%s, %s)", field, field, path, value)
}
func (g *MysqlGrammar) CompileInsertUsing(query *QueryBuilder, columns []interface{}, sql string) string {
	if len(columns) == 0 || (len(columns) == 1 && columns[0] == "*") {
		return "insert into " + g.WrapTable(query.FromTable) + " " + sql
	}
	return "insert into " + g.WrapTable(query.FromTable) + " (" + g.Columnize(columns) + ") " + sql
}
func (g *MysqlGrammar) CompileInsertOrIgnoreUsing(query *QueryBuilder, columns []interface{}, sql string) string {

	return strings.Replace(g.CompileInsertUsing(query, columns, sql), "insert", "insert ignore", 1)
}
func (g *MysqlGrammar) CompileJoins(query *QueryBuilder) string {
	var parts []string
	for _, join := range query.Joins {
		table := g.WrapTable(join.Table)
		var nestedJoins string
		if len(join.QueryBuilder.Joins) > 0 {
			nestedJoins = " " + g.CompileJoins(join.QueryBuilder)
			table = "(" + g.WrapTable(join.Table) + nestedJoins + ")"
		}
		if join.Lateral {
			parts = append(parts, g.CompileJoinLateral(join, table))
		} else {
			parts = append(parts, strings.TrimSuffix(fmt.Sprintf("%s join %s %s", join.Type, table, g.CompileWheres(join.QueryBuilder)), " "))

		}
	}
	return strings.Join(parts, " ")
}
func (g *MysqlGrammar) CompileJoinLateral(query *JoinBuilder, expression string) string {
	return strings.TrimPrefix(fmt.Sprintf("%s join lateral %s on true", query.Type, expression), " ")
}
func (g *MysqlGrammar) CompileUnionAggregate(query *QueryBuilder) string {

	sql := g.CompileAggregate(query)
	query.Aggregates = Aggregate{
		AggregateName:    "",
		AggregateColumns: []interface{}{},
	}

	return fmt.Sprintf("%s from (%s) as %s", sql, g.CompileSelect(query), g.Wrap("temp_table"))
}

func (g *MysqlGrammar) CompileGroupLimit(query *QueryBuilder) string {

	var bindings []interface{}
	for _, i := range query.GetRawBindings()[COMPONENT_SELECT] {
		bindings = append(bindings, i)
	}
	for _, i := range query.GetRawBindings()[COMPONENT_ORDER] {
		bindings = append(bindings, i)
	}
	query.SetBindings(bindings, COMPONENT_SELECT)
	query.SetBindings([]interface{}{}, COMPONENT_ORDER)
	limit := query.Grouplimit.Value
	offset := query.OffsetNum
	if offset > 0 {
		limit = limit + offset
		query.OffsetNum = 0
	}

	components := g.CompileComponents(query)
	orders, ok := components[COMPONENT_ORDER]
	if !ok {
		orders = ""
	}

	components[COMPONENT_COLUMN] = components[COMPONENT_COLUMN] + g.CompileRowNumber(query.Grouplimit.Column, orders)

	delete(components, COMPONENT_ORDER)
	table := g.Wrap(Eloquent + "_table")
	row := g.Wrap(Eloquent + "_row")
	sql := Concatenate(components)
	sql = fmt.Sprintf("select * from (%s) as %s where %s <= %d", sql, table, row, limit)
	if offset > 0 {
		sql = " and " + row + " > " + strconv.Itoa(offset)
	}

	return sql + " order by " + row
}

func (g *MysqlGrammar) CompileRowNumber(partition, orders string) string {

	over := "partition by " + g.Wrap(partition) + " " + orders
	return fmt.Sprintf(", row_number() over (%s) as %s", over, g.Wrap(Eloquent+"_row"))
}
func (g *MysqlGrammar) CompileComponents(query *QueryBuilder) map[Component]string {
	parts := make(map[Component]string)
	for key, component := range query.Components {
		switch key {
		case COMPONENT_AGGREGRATE:
			parts[key] = g.CompileAggregate(query)
		case COMPONENT_COLUMN:
			parts[key] = g.CompileColumns(query)
		case COMPONENT_FROM:
			parts[key] = g.CompileFrom(query)
		case COMPONENT_INDEX_HINT:
			parts[key] = g.CompileIndexHint(query)
		case COMPONENT_JOIN:
			parts[key] = g.CompileJoins(query)
		case COMPONENT_WHERE:
			parts[key] = g.CompileWheres(query)
		case COMPONENT_GROUP_BY:
			parts[key] = g.CompileGroups(query)
		case COMPONENT_HAVING:
			parts[key] = g.CompileHavings(query.Havings)
		case COMPONENT_ORDER:
			parts[key] = g.CompileOrders(query)
		case COMPONENT_LIMIT:
			parts[key] = g.CompileLimit(query)
		case COMPONENT_OFFSET:
			parts[key] = g.CompileOffset(query)
		case COMPONENT_LOCK:
			parts[key] = g.CompileLock(query)
		default:
			g.AddError(errors.New(fmt.Sprintf("unsupported component %s", component)))
			return nil
		}
	}
	return parts
}

func (g *MysqlGrammar) CompileAggregate(query *QueryBuilder) string {

	if query.Aggregates.AggregateName == "" {
		return ""
	}
	column := g.Columnize(query.Aggregates.AggregateColumns)
	if cs, ok := query.IsDistinct.([]interface{}); ok {
		column = "distinct " + g.Columnize(cs)
	} else if b, ok := query.IsDistinct.(bool); ok && b && column != "*" {
		column = "distinct " + column
	}

	return fmt.Sprintf("select %s(%s) as aggregate", query.Aggregates.AggregateName, column)
}

func (g *MysqlGrammar) CompileColumns(query *QueryBuilder) string {
	if query.Aggregates.AggregateName != "" {
		return ""
	}
	if query.IsDistinct != nil {
		return "select distinct " + g.Columnize(query.Columns)
	} else {
		return "select " + g.Columnize(query.Columns)
	}
}

func (g *MysqlGrammar) CompileFrom(query *QueryBuilder) string {
	return "from " + g.WrapTable(query.FromTable)
}

func (g *MysqlGrammar) CompileIndexHint(query *QueryBuilder) string {

	switch query.IndexHint.Type {
	case "hint":
		return fmt.Sprintf("use index (%s)", query.IndexHint.Index)
	case "force":
		return fmt.Sprintf("force index (%s)", query.IndexHint.Index)
	default:
		return fmt.Sprintf("ignore index (%s)", query.IndexHint.Index)
	}
}

func (g *MysqlGrammar) CompileGroups(query *QueryBuilder) string {
	return "group by " + g.Columnize(query.Groups)
}
func (g *MysqlGrammar) CompileHavings(havings []Having) string {
	var sqls []string
	for _, having := range havings {
		sqls = append(sqls, having.HavingBoolean+" "+g.CompileHaving(having))
	}
	return "having " + removeLeadingBoolean(strings.Join(sqls, " "))
}
func (g *MysqlGrammar) CompileHaving(having Having) string {
	switch having.Type {

	case HavingTypeRaw:
		return g.CompileHavingRaw(having)
	case HavingTypeBetween:
		return g.CompileHavingBetween(having)
	case HavingTypeBitwise:
		return g.CompileHavingBitwise(having)
	case HavingTypeNull:
		return g.CompileHavingNull(having)
	case HavingTypeNotNull:
		return g.CompileHavingNotNull(having)
	case HavingTypeNested:
		return g.CompileHavingNested(having)
	case HavingTypeExpression:
		return g.CompileHavingExpression(having)
	default:
		return g.CompileHavingBasic(having)
	}
}

func (g *MysqlGrammar) CompileOrders(query *QueryBuilder) string {

	if len(query.Orders) > 0 {
		return "order by " + strings.Join(g.CompileOrdersToArray(query), ", ")
	}
	return ""
}
func (g *MysqlGrammar) CompileOrdersToArray(query *QueryBuilder) []string {
	var orders []string
	for _, order := range query.Orders {
		if order.RawSql != "" {
			orders = append(orders, order.RawSql+" "+order.Direction)
		} else {
			orders = append(orders, g.Wrap(order.Column)+" "+order.Direction)
		}
	}
	return orders
}
func (g *MysqlGrammar) CompileRandom(seed ...int) string {
	if len(seed) > 0 {
		return "RAND(" + strconv.Itoa(seed[0]) + ")"
	}
	return "RAND()"
}
func (g *MysqlGrammar) CompileLimit(query *QueryBuilder) string {
	return "limit " + strconv.Itoa(query.LimitNum)
}

func (g *MysqlGrammar) CompileOffset(query *QueryBuilder) string {
	return "offset " + strconv.Itoa(query.OffsetNum)
}

func (g *MysqlGrammar) CompileTruncate(query *QueryBuilder) string {

	return "truncate table " + g.WrapTable(query.FromTable)
}
func (g *MysqlGrammar) CompileLock(query *QueryBuilder) string {

	return query.Locks
}
func removeLeadingBoolean(str string) string {
	re := regexp.MustCompile(`(?)(^and\s|^or\s)`)
	return re.ReplaceAllString(str, "")
}
func (g *MysqlGrammar) CompileWheres(query *QueryBuilder) string {
	if len(query.Wheres) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, where := range query.Wheres {
		sb.WriteString(where.Boolean + " ")
		switch where.Type {
		case WhereTypeBasic:
			sb.WriteString(g.CompileWhereBasic(where))
		case WhereTypeExpression:
			sb.WriteString(g.CompileWhereExpression(where))
		case WhereTypeBitwise:
			sb.WriteString(g.CompileWhereBitwise(where))
		case WhereTypeJsonBoolean:
			sb.WriteString(g.CompileWhereJsonBoolean(where))
		case WhereTypeColumn:
			sb.WriteString(g.CompileWhereColumn(where))
		case WhereTypeRaw:
			sb.WriteString(g.CompileWhereRaw(where))
		case WhereTypeLike:
			sb.WriteString(g.CompileWhereLike(where))
		case WhereTypeNotIn:
			sb.WriteString(g.CompileWhereNotIn(where))
		case WhereTypeIn:
			sb.WriteString(g.CompileWhereIn(where))
		case WhereTypeNotInRaw:
			sb.WriteString(g.CompileWhereNotInRaw(where))
		case WhereTypeInRaw:
			sb.WriteString(g.CompileWhereInRaw(where))
		case WhereTypeNotNull:
			sb.WriteString(g.CompileWhereNotNull(where))
		case WhereTypeNull:
			sb.WriteString(g.CompileWhereNull(where))
		case WhereTypeBetween:
			sb.WriteString(g.CompileWhereBetween(where))
		case WhereTypeBetweenColumn:
			sb.WriteString(g.CompileWhereBetweenColumn(where))
		case WhereTypeNested:
			sb.WriteString(g.CompileWhereNested(where))
		case WhereTypeSub:
			sb.WriteString(g.CompileWhereSub(where))
		case WhereTypeNotExists:
			sb.WriteString(g.CompileWhereNotExists(where))
		case WhereTypeExists:
			sb.WriteString(g.CompileWhereExists(where))
		case WhereTypeRowValues:
			sb.WriteString(g.CompileWhereRowValues(where))
		case WhereTypeJsonContains:
			sb.WriteString(g.CompileWhereJsonContains(where))
		case WhereTypeJsonOverlaps:
			sb.WriteString(g.CompileWhereJsonOverlaps(where))
		case WhereTypeJsonContainsKey:
			sb.WriteString(g.CompileWhereJsonContainsKey(where))
		case WhereTypeJsonLength:
			sb.WriteString(g.CompileWhereJsonLength(where))
		case WhereTypeFulltext:
			sb.WriteString(g.CompileWhereFulltext(where))
		case WhereTypeDate:
			sb.WriteString(g.CompileWhereDate(where))
		case WhereTypeTime:
			sb.WriteString(g.CompileWhereTime(where))
		case WhereTypeDay:
			sb.WriteString(g.CompileWhereDay(where))
		case WhereTypeMonth:
			sb.WriteString(g.CompileWhereMonth(where))
		case WhereTypeYear:
			sb.WriteString(g.CompileWhereYear(where))
		}
		sb.WriteString(" ")

	}

	str := strings.TrimSuffix(sb.String(), " ")
	str = removeLeadingBoolean(str)

	if query.IsJoin {
		return "on " + str
	}

	return "where " + str
}

func (g *MysqlGrammar) CompileWhereExpression(where Where) string {
	return where.RawSql.(string)
}

func (g *MysqlGrammar) CompileWhereBitwise(where Where) string {

	return g.CompileWhereBasic(where)
}

func (g *MysqlGrammar) CompileWhereJsonBoolean(where Where) string {

	column := g.WrapJsonBooleanSelector(where.Column)
	value := g.WrapJsonBooleanValue(g.Parameter(where.Value))

	return column + " " + where.Operator + " " + value
}

func (g *MysqlGrammar) CompileWhereColumn(where Where) string {

	return g.Wrap(where.First) + " " + where.Operator + " " + g.Wrap(where.Second)
}

func (g *MysqlGrammar) CompileWhereRaw(where Where) string {

	if expression, ok := where.RawSql.(Expression); ok {
		return string(expression)
	}
	return where.RawSql.(string)
}

func (g *MysqlGrammar) CompileWhereLike(where Where) string {
	operator := ""
	if where.Not {
		operator = "not "
	}
	if where.CaseSensitive {
		operator = operator + "like binary"
	} else {
		operator = operator + "like"
	}
	where.Operator = operator
	return g.CompileWhereBasic(where)
}

func (g *MysqlGrammar) CompileWhereNotIn(where Where) string {

	if len(where.Values) == 0 {
		return "1 = 1"
	}
	return g.Wrap(where.Column) + " not in (" + g.Parameterize(where.Values) + ")"
}

func (g *MysqlGrammar) CompileWhereIn(where Where) string {

	if len(where.Values) == 0 {
		return "0 = 1"
	}
	return g.Wrap(where.Column) + " in (" + g.Parameterize(where.Values) + ")"
}

func (g *MysqlGrammar) CompileWhereNotInRaw(where Where) string {

	if len(where.Values) == 0 {
		return "1 = 1"
	}
	return g.Wrap(where.Column) + " not in (" + g.Parameterize(where.Values) + ")"
}

func (g *MysqlGrammar) CompileWhereInRaw(where Where) string {

	if len(where.Values) == 0 {
		return "0 = 1"
	}
	return g.Wrap(where.Column) + " in (" + g.Parameterize(where.Values) + ")"
}

func (g *MysqlGrammar) CompileWhereNotNull(where Where) string {
	var columnStr string
	if str, ok := where.Column.(string); ok {
		columnStr = str
	} else if expression, ok := where.Column.(Expression); ok {
		columnStr = string(expression)
	}
	if IsJsonSelector(columnStr) {
		field, path := g.WrapJsonFieldAndPath(where.Column)
		return fmt.Sprintf("(json_extract(%s%s) is not null AND json_type(json_extract(%s%s)) != 'NULL')", field, path, field, path)
	}
	return g.Wrap(where.Column) + " is not null"
}

func (g *MysqlGrammar) CompileWhereNull(where Where) string {
	var columnStr string
	if str, ok := where.Column.(string); ok {
		columnStr = str
	} else if expression, ok := where.Column.(Expression); ok {
		columnStr = string(expression)
	}
	if IsJsonSelector(columnStr) {
		field, path := g.WrapJsonFieldAndPath(where.Column)
		return fmt.Sprintf("(json_extract(%s%s) is null OR json_type(json_extract(%s%s)) = 'NULL')", field, path, field, path)
	}
	return g.Wrap(where.Column) + " is null"
}

func (g *MysqlGrammar) CompileWhereBetween(where Where) string {

	between := "between"
	if where.Not {
		between = "not between"
	}
	return g.Wrap(where.Column) + " " + between + " " + g.Parameter(where.Values[0]) + " and " + g.Parameter(where.Values[1])

}

func (g *MysqlGrammar) CompileWhereBetweenColumn(where Where) string {

	between := "between"
	if where.Not {
		between = "not between"
	}
	return g.Wrap(where.Column) + " " + between + " " + g.Wrap(where.Values[0]) + " and " + g.Wrap(where.Values[1])
}

func (g *MysqlGrammar) CompileWhereNested(where Where) string {
	sql := g.CompileWheres(where.Query)

	if where.Query.IsJoin {
		return "(" + sql[3:] + ")"
	} else {
		return "(" + sql[6:] + ")"
	}

}

func (g *MysqlGrammar) CompileWhereSub(where Where) string {

	sql := g.CompileSelect(where.Query)
	return g.Wrap(where.Column) + " " + where.Operator + " (" + sql + ")"
}

func (g *MysqlGrammar) CompileWhereNotExists(where Where) string {

	return "not exists (" + g.CompileSelect(where.Query) + ")"
}

func (g *MysqlGrammar) CompileWhereExists(where Where) string {

	return "exists (" + g.CompileSelect(where.Query) + ")"
}

func (g *MysqlGrammar) CompileWhereRowValues(where Where) string {

	columns := g.Columnize(where.Columns)
	values := g.Parameterize(where.Values)
	return fmt.Sprintf("(%s) %s (%s)", columns, where.Operator, values)
}

func (g *MysqlGrammar) CompileWhereJsonContains(where Where) string {

	not := ""
	if where.Not {
		not = "not "
	}

	field, path := g.WrapJsonFieldAndPath(where.Column)
	return not + "json_contains(" + field + ", " + where.Value.(string) + path + ")"
}

func (g *MysqlGrammar) CompileWhereJsonOverlaps(where Where) string {

	not := ""
	if where.Not {
		not = "not "
	}

	field, path := g.WrapJsonFieldAndPath(where.Column)

	return not + "json_overlaps(" + field + ", " + g.Parameter(where.Value) + path + ")"
}

func (g *MysqlGrammar) CompileWhereJsonContainsKey(where Where) string {

	not := ""
	if where.Not {
		not = "not "
	}

	field, path := g.WrapJsonFieldAndPath(where.Column)
	return fmt.Sprintf("%sifnull(json_contains_path(%s, \\'one\\'%s), 0)", not, field, path)
}

func (g *MysqlGrammar) CompileWhereJsonLength(where Where) string {

	field, path := g.WrapJsonFieldAndPath(where.Column)
	return "json_length(" + field + path + ") " + where.Operator + " " + g.Parameter(where.Value)
}

func (g *MysqlGrammar) CompileWhereFulltext(where Where) string {

	columns := g.Columnize(where.Columns)
	value := g.Parameter(where.Value)
	mode := "natural language mode"
	if where.Mode == "boolean" {
		mode = "boolean mode"
	}
	expansion := ""
	if where.Expanded {
		expansion = " with query expansion"
	}
	return fmt.Sprintf("match (%s) against (%s in %s%s)", columns, value, mode, expansion)
}

func (g *MysqlGrammar) CompileWhereDate(where Where) string {

	return string(where.Type) + "(" + g.Wrap(where.Column) + ") " + where.Operator + " " + g.Parameter(where.Value)
}

func (g *MysqlGrammar) CompileWhereTime(where Where) string {

	return string(where.Type) + "(" + g.Wrap(where.Column) + ") " + where.Operator + " " + g.Parameter(where.Value)
}

func (g *MysqlGrammar) CompileWhereDay(where Where) string {

	return string(where.Type) + "(" + g.Wrap(where.Column) + ") " + where.Operator + " " + g.Parameter(where.Value)
}

func (g *MysqlGrammar) CompileWhereMonth(where Where) string {

	return string(where.Type) + "(" + g.Wrap(where.Column) + ") " + where.Operator + " " + g.Parameter(where.Value)
}

func (g *MysqlGrammar) CompileWhereYear(where Where) string {

	return string(where.Type) + "(" + g.Wrap(where.Column) + ") " + where.Operator + " " + g.Parameter(where.Value)
}

func (g *MysqlGrammar) CompileWhereBasic(where Where) string {
	operator := strings.ReplaceAll(where.Operator, "?", "??")

	return g.Wrap(where.Column) + " " + operator + " " + g.Parameter(where.Value)
}

func (g *MysqlGrammar) CompileHavingRaw(having Having) string {

	return having.RawSql
}
func (g *MysqlGrammar) CompileHavingBasic(having Having) string {

	return g.Wrap(having.HavingColumn) + " " + having.HavingOperator + " " + g.Parameter(having.HavingValue[0])
}

func (g *MysqlGrammar) CompileHavingBetween(having Having) string {

	between := "between"
	if having.Not {
		between = "not between"
	}
	return g.Wrap(having.HavingColumn) + " " + between + " " + g.Parameter(having.HavingValue[0]) + " and " + g.Parameter(having.HavingValue[1])
}

func (g *MysqlGrammar) CompileHavingBitwise(having Having) string {

	return fmt.Sprintf("(%s %s %s) != 0", g.Wrap(having.HavingColumn), having.HavingOperator, g.Parameter(having.HavingValue[0]))
}

func (g *MysqlGrammar) CompileHavingExpression(having Having) string {
	return having.HavingColumn
}

func (g *MysqlGrammar) CompileHavingNull(having Having) string {

	return g.Wrap(having.HavingColumn) + " is null"
}

func (g *MysqlGrammar) CompileHavingNotNull(having Having) string {

	return g.Wrap(having.HavingColumn) + " is not null"
}

func (g *MysqlGrammar) CompileHavingNested(having Having) string {
	return "(" + g.CompileHavings(having.Query.Havings)[7:] + ")"
}

func PrepareBindsForDelete(bindings map[Component][]interface{}) []interface{} {
	var res []interface{}
	for com, binds := range bindings {
		if com != COMPONENT_SELECT {
			for _, bind := range binds {
				if _, ok := bind.(Expression); !ok {
					res = append(res, bind)
				}
			}
		}
	}
	return res
}

func (g *MysqlGrammar) GetOperators() map[string]struct{} {
	return map[string]struct{}{
		"sounds like": {},
	}
}
