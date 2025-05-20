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
	case string, Expression:
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
