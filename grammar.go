package goeloquent

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Expression string

func Raw(expr string) Expression {
	return Expression(expr)
}

type MysqlGrammar struct {
	Error error
}

func NewMysqlGrammar() *MysqlGrammar {
	return &MysqlGrammar{}
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
func (g *MysqlGrammar) Wrap(value interface{}) string {
	switch value.(type) {
	case Expression:
		return string(value.(Expression))
	case string:
		if strings.Contains(value.(string), " as ") || strings.Contains(value.(string), " AS ") {
			return g.WrapAliasedValue(value.(string))
		} else if isJsonSelector(value.(string)) {
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
	if len(prefix) > 0 {
		tablePrefix = prefix[0]
	}
	var segments []string
	if strings.Contains(table, " AS ") {
		segments = strings.SplitN(table, " AS ", 2)
	} else {
		segments = strings.SplitN(table, " as ", 2)
	}
	return g.WrapTable(segments[0], prefix...) + " as " + g.WrapValue(tablePrefix+segments[1])

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
func (g *MysqlGrammar) WrapJsonFieldAndPath(value string) (string, string) {
	parts := strings.SplitN(value, "->", 2)

	if len(parts) > 1 {
		return g.Wrap(parts[0]), g.WrapJsonPath(parts[1], "->")
	} else {
		return g.Wrap(parts[0]), ""
	}
}
func (g *MysqlGrammar) WrapJsonPath(value, delimiter string) string {
	re := regexp.MustCompile(`([\\]+)?'`)
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

	switch value.(type) {
	case string:
		return "?"
	case Expression:
		return value.(string)
	default:
		g.AddError(errors.New(fmt.Sprintf("parameter value type %T is not string or expression", value)))
	}
	return ""
}

func (g *MysqlGrammar) Parameterize(value []interface{}) string {
	var parameters []string
	for _, v := range value {
		switch v.(type) {
		case string, Expression:
			parameters = append(parameters, g.Parameter(v))
		default:
			g.AddError(errors.New(fmt.Sprintf("parameterize value type %T is not string or expression", v)))
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
	if len(prefix) > 0 {
		tablePrefix = prefix[0]
	} else {
		tablePrefix = ""
	}
	switch t := tableName.(type) {
	case Expression:
		return string(t)
	case string:
		if strings.Contains(t, " as ") || strings.Contains(t, " AS ") {
			return g.WrapAliasedTable(t)
		}
		if strings.Contains(t, ".") {

			segments := strings.SplitN(t, ".", 2)
			ts := segments[0] + "." + tablePrefix + segments[1]

			return g.WrapSegments(strings.Split(ts, "."))
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

	if len(query.Havings) > 0 && len(query.Aggregate.AggregateName) > 0 {
		return g.CompileUnionAggregate(query)
	}
	if query.GroupLimit > 0 {
		if len(query.Columns) == 0 {
			query.Columns = []interface{}{"*"}
		}
		return g.CompileGroupLimit(query)
	}
	columns := query.Columns
	if len(query.Columns) == 0 {
		columns = []interface{}{"*"}
	}
	parts := g.CompileComponents(query)
	var sb strings.Builder
	for i, part := range parts {
		if len(part) > 0 {
			sb.Write([]byte(strings.TrimSuffix(part, " ")))
		}
		if i != len(parts)-1 {
			sb.WriteByte(' ')
		}
	}
	query.Columns = columns
	return sb.String()
}

func (g *MysqlGrammar) CompileUnionAggregate(query *QueryBuilder) string {

}

func (g *MysqlGrammar) CompileGroupLimit(query *QueryBuilder) string {

}

func (g *MysqlGrammar) CompileComponents(query *QueryBuilder) []string {
	var parts []string
	for key, component := range query.Components {
		switch key {
		case COMPONENT_AGGREGRATE:
			parts = append(parts, g.CompileAggregate(query))
		case COMPONENT_COLUMN:
			parts = append(parts, g.CompileColumn(query))
		case COMPONENT_FROM:
			parts = append(parts, g.CompileFrom(query))
		case COMPONENT_INDEX_HINT:
			parts = append(parts, g.CompileIndexHint(query))
		case COMPONENT_JOIN:
			parts = append(parts, g.CompileJoin(query, query.Joins))
		case COMPONENT_WHERE:
			parts = append(parts, g.CompileWhere(query))
		case COMPONENT_GROUP_BY:
			parts = append(parts, g.CompileGroup(query))
		case COMPONENT_HAVING:
			parts = append(parts, g.CompileHaving(query))
		case COMPONENT_ORDER:
			parts = append(parts, g.CompileOrder(query))
		case COMPONENT_LIMIT:
			parts = append(parts, g.CompileLimit(query))
		case COMPONENT_OFFSET:
			parts = append(parts, g.CompileOffset(query))
		case COMPONENT_LOCK:
			parts = append(parts, g.CompileLock(query))
		default:
			g.AddError(errors.New(fmt.Sprintf("unsupported component %s", component)))
			return nil
		}
	}
	return parts
}

func (g *MysqlGrammar) CompileAggregate(query *QueryBuilder) string {

	column := g.Columnize(query.Aggregate.AggregateColumns)
	if cs, ok := query.Distinct.([]interface{}); ok {
		column = "distinct " + g.Columnize(cs)
	} else if b, ok := query.Distinct.(bool); ok && b && column != "*" {
		column = "distinct " + column
	}

	return fmt.Sprintf("select %s(%s) as aggregate", query.Aggregate.AggregateName, column)
}

func (g *MysqlGrammar) CompileColumn(query *QueryBuilder) string {
	if query.Aggregate.AggregateName != "" {
		return ""
	}
	if query.Distinct != nil {
		return "select distinct " + g.Columnize(query.Columns)
	} else {
		return "select " + g.Columnize(query.Columns)
	}
}

func (g *MysqlGrammar) CompileFrom(query *QueryBuilder) string {
	return "from " + g.WrapTable(query.From)
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

func (g *MysqlGrammar) CompileJoin(query *QueryBuilder, joins []*JoinBuilder) string {

	var parts []string
	for _, join := range joins {
		table := g.WrapTable(join.Table)
		var nested string
		tableAndNested := table
		if len(join.Joins) > 0 {
			nested = g.CompileJoin(query, join.Joins)
			tableAndNested = fmt.Sprintf("(%s%s)", table, nested)
		}

		parts = append(parts, fmt.Sprintf("%s join %s %s", join.Type, tableAndNested, g.CompileWhere(join.QueryBuilder)))

	}
	return strings.TrimSuffix(strings.Join(parts, " "), " ")
}

func (g *MysqlGrammar) CompileGroup(query *QueryBuilder) string {
	return ""
}

func (g *MysqlGrammar) CompileHaving(query *QueryBuilder) string {
	return ""
}

func (g *MysqlGrammar) CompileOrder(query *QueryBuilder) string {
	return ""
}

func (g *MysqlGrammar) CompileLimit(query *QueryBuilder) string {
	return ""
}

func (g *MysqlGrammar) CompileOffset(query *QueryBuilder) string {
	return ""
}

func (g *MysqlGrammar) CompileLock(query *QueryBuilder) string {
	return ""
}
