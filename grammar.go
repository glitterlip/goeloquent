package goeloquent

import (
	"errors"
	"fmt"
	"strings"
)

type Expression string

func Raw(expr string) Expression {
	return Expression(expr)
}

type MysqlGrammar struct {
	Error error
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
			return g.WrapAliasedValue(value)
		} else if isJsonSelector(value.(string)) {
			return g.WrapJsonSelector(value.(string))
		} else {
			return g.WrapSegments(strings.Split(value.(string), "."))
		}
	}
	g.AddError(errors.New(fmt.Sprintf("wrap value type %T is not string or expression", value)))

	return ""
}

func (g *MysqlGrammar) WrapAliasedValue(value interface{}) string {
	//TODO implement me
	panic("implement me")
}

func (g *MysqlGrammar) WrapAliasedTable(table string, prefix ...string) string {
	//TODO implement me
	panic("implement me")
}

func (g *MysqlGrammar) WrapValue(value string) string {
	//TODO implement me
	panic("implement me")
}

func (g *MysqlGrammar) WrapJsonSelector(value string) string {
	//TODO implement me
	panic("implement me")
}

func (g *MysqlGrammar) Columnize(value []interface{}) string {
	//TODO implement me
	panic("implement me")
}

func (g *MysqlGrammar) Parameter(value interface{}) string {
	//TODO implement me
	panic("implement me")
}

func (g *MysqlGrammar) Parameterize(value []interface{}) string {
	//TODO implement me
	panic("implement me")
}

func (g *MysqlGrammar) QuoteString(value interface{}) string {
	//TODO implement me
	panic("implement me")
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
		if i == 0 && len(segments) > 1 {
			wrappedSegments = append(wrappedSegments, g.WrapTable(segment))
		} else {
			wrappedSegments = append(wrappedSegments, g.WrapValue(segment))
		}
	}
	return strings.Join(wrappedSegments, ".")
}
func NewMysqlGrammar() *MysqlGrammar {
	return &MysqlGrammar{}
}
