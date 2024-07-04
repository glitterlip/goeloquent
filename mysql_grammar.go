package goeloquent

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type MysqlGrammar struct {
	Prefix  string
	Builder *Builder
}

func (m *MysqlGrammar) SetTablePrefix(prefix string) {
	m.Prefix = prefix
}

func (m *MysqlGrammar) GetTablePrefix() string {
	return m.Prefix
}

func (m *MysqlGrammar) SetBuilder(builder *Builder) {
	m.Builder = builder
}

func (m *MysqlGrammar) GetBuilder() *Builder {
	return m.Builder
}

func (m *MysqlGrammar) CompileInsert(values []map[string]interface{}) string {
	b := m.GetBuilder()
	b.PreSql.WriteString("insert into ")
	b.PreSql.WriteString(m.CompileComponentTable())
	b.PreSql.WriteString(" (")
	first := values[0]
	length := len(values)
	var columns []string
	var columnizeVars []interface{}
	for key := range first {
		if (b.OnlyColumns == nil && b.ExceptColumns == nil) || b.FileterColumn(key) {
			columns = append(columns, key)
			columnizeVars = append(columnizeVars, key)
		}
	}
	columnLength := len(columns)
	b.PreSql.WriteString(m.columnize(columnizeVars))
	b.PreSql.WriteString(") values ")

	for k, v := range values {
		b.PreSql.WriteString("(")
		for i, key := range columns {
			b.PreSql.WriteString(m.parameter(v[key]))
			b.AddBinding([]interface{}{v[key]}, TYPE_INSERT)
			if i != columnLength-1 {
				b.PreSql.WriteString(", ")
			} else {
				b.PreSql.WriteString(")")
			}
		}
		if k != length-1 {
			b.PreSql.WriteString(", ")
		}
	}
	b.PreparedSql = b.PreSql.String()
	b.PreSql.Reset()
	return b.PreparedSql
}

func (m *MysqlGrammar) CompileInsertOrIgnore(values []map[string]interface{}) string {
	m.GetBuilder().PreparedSql = strings.Replace(m.CompileInsert(values), "insert", "insert ignore", 1)
	return m.GetBuilder().PreparedSql
}
func (m *MysqlGrammar) CompileDelete() string {
	b := m.GetBuilder()
	b.PreSql.WriteString("delete from ")
	b.PreSql.WriteString(m.CompileComponentTable())
	b.PreSql.WriteString(m.CompileComponentWheres())
	if len(b.Orders) > 0 {
		b.PreSql.WriteString(m.CompileComponentOrders())
	}
	if b.LimitNum > 0 {
		b.PreSql.WriteString(m.CompileComponentLimitNum())
	}
	m.GetBuilder().PreparedSql = m.GetBuilder().PreSql.String()
	b.PreSql.Reset()
	return m.GetBuilder().PreparedSql
}

func (m *MysqlGrammar) CompileUpdate(value map[string]interface{}) string {
	b := m.GetBuilder()
	b.PreSql.WriteString("update ")
	b.PreSql.WriteString(m.CompileComponentTable())
	b.PreSql.WriteString(" set ")
	count := 0
	length := len(value)
	for k, v := range value {
		count++
		if (b.OnlyColumns == nil && b.ExceptColumns == nil) || b.FileterColumn(k) {
			b.PreSql.WriteString(m.Wrap(k))
			b.PreSql.WriteString(" = ")
			b.AddBinding([]interface{}{v}, TYPE_UPDATE)
			if e, ok := v.(Expression); ok {
				b.PreSql.WriteString(string(e))
			} else {
				b.PreSql.WriteString(m.parameter(v))
			}
		}
		if count != length {
			b.PreSql.WriteString(" , ")
		}
	}
	b.PreSql.WriteString(m.CompileComponentWheres())
	m.GetBuilder().PreparedSql = b.PreSql.String()
	b.PreSql.Reset()
	return m.GetBuilder().PreparedSql
}

func (m *MysqlGrammar) CompileSelect() string {
	b := m.GetBuilder()
	if len(b.Aggregates) > 0 && len(b.Havings) > 0 {
		return m.CompileUnionAggregate()
	}
	b.PreparedSql = ""
	b.PreSql = strings.Builder{}
	if _, ok := b.Components[TYPE_COLUMN]; !ok || len(b.Columns) == 0 {
		b.Components[TYPE_COLUMN] = struct{}{}
		b.Columns = append(b.Columns, "*")
	}
	for _, componentName := range SelectComponents {
		if _, ok := b.Components[componentName]; ok {
			b.PreSql.WriteString(m.compileComponent(componentName))
		}
	}
	b.PreparedSql = b.PreSql.String()
	b.PreSql.Reset()
	return b.PreparedSql
}
func (m *MysqlGrammar) CompileUnionAggregate() string {
	b := m.GetBuilder()
	sql := m.CompileComponentAggregate(b.Aggregates...)
	b.Aggregates = []Aggregate{}
	b.PreparedSql = fmt.Sprintf("%s from (%s) as %s", sql, m.CompileSelect(), m.Wrap("temp_table"))
	return b.PreparedSql
}
func (m *MysqlGrammar) CompileExists() string {
	sql := m.CompileSelect()
	m.GetBuilder().PreparedSql = fmt.Sprintf("select exists(%s) as %s", sql, m.Wrap("exists"))
	return m.GetBuilder().PreparedSql
}

func (m *MysqlGrammar) compileComponent(componentName string) string {
	switch componentName {
	case TYPE_AGGREGRATE:
		return m.CompileComponentAggregate()
	case TYPE_COLUMN:
		return m.CompileComponentColumns()
	case TYPE_FROM:
		return m.CompileComponentFromTable()
	case TYPE_JOIN:
		return m.CompileComponentJoins()
	case TYPE_WHERE:
		return m.CompileComponentWheres()
	case TYPE_GROUP_BY:
		return m.CompileComponentGroups()
	case TYPE_HAVING:
		return m.CompileComponentHavings()
	case TYPE_ORDER:
		return m.CompileComponentOrders()
	case TYPE_LIMIT:
		return m.CompileComponentLimitNum()
	case TYPE_OFFSET:
		return m.CompileComponentOffsetNum()
	case "unions":
	case TYPE_LOCK:
		return m.CompileLock()
	}
	return ""
}
func (m *MysqlGrammar) CompileComponentAggregate(aggregates ...Aggregate) string {
	builder := strings.Builder{}
	var aggregate Aggregate
	if len(aggregates) == 0 {
		b := m.GetBuilder()
		if len(b.Aggregates) == 0 {
			return ""
		}
		aggregate = b.Aggregates[0]
	} else {
		aggregate = aggregates[0]
	}
	builder.WriteString("select ")
	builder.WriteString(aggregate.AggregateName)
	builder.WriteString("(")
	if m.GetBuilder().IsDistinct && aggregate.AggregateColumn != "*" {
		builder.WriteString("distinct ")
		builder.WriteString(m.Wrap(aggregate.AggregateColumn))
	} else {
		builder.WriteString(m.Wrap(aggregate.AggregateColumn))
	}
	builder.WriteString(") as aggregate")
	return builder.String()
}

// Convert []string column names into a delimited string.
// Compile the "select *" portion of the
func (m *MysqlGrammar) CompileComponentColumns() string {
	if len(m.GetBuilder().Aggregates) == 0 {
		builder := strings.Builder{}
		if m.GetBuilder().IsDistinct {
			builder.WriteString("select distinct ")
		} else {
			builder.WriteString("select ")
		}
		builder.WriteString(m.columnize(m.GetBuilder().Columns))
		return builder.String()
	}

	return ""

}

func (m *MysqlGrammar) CompileComponentFromTable() string {
	builder := strings.Builder{}
	builder.WriteString(" from ")
	builder.WriteString(m.WrapTable(m.GetBuilder().FromTable))
	return builder.String()
}
func (m *MysqlGrammar) CompileComponentTable() string {
	return m.WrapTable(m.GetBuilder().FromTable)
}
func (m *MysqlGrammar) CompileComponentJoins() string {
	builder := strings.Builder{}
	for _, join := range m.GetBuilder().Joins {
		var tableAndNestedJoins string
		if len(join.Joins) > 0 {
			//nested join
			tableAndNestedJoins = fmt.Sprintf("(%s%s)", m.WrapTable(join.Table), join.Grammar.CompileComponentJoins())
		} else {
			tableAndNestedJoins = m.WrapTable(join.Table)
		}
		onStr := join.Grammar.CompileComponentWheres()
		s := ""
		if len(onStr) > 0 {
			s = fmt.Sprintf(" %s join %s %s", join.JoinType, tableAndNestedJoins, strings.TrimSpace(onStr))
		} else {
			s = fmt.Sprintf(" %s join %s", join.JoinType, tableAndNestedJoins)
		}
		builder.WriteString(s)
	}
	return builder.String()
}

func (m *MysqlGrammar) CompileComponentWheres() string {
	if len(m.GetBuilder().Wheres) == 0 {
		return ""
	}
	builder := strings.Builder{}

	for _, w := range m.GetBuilder().Wheres {
		//TODO: filter columns for select
		switch w.Type {
		case CONDITION_TYPE_NESTED:
			builder.WriteString(w.Boolean + " ")
			builder.WriteString("(")
			cloneBuilder := w.Value.(*Builder)
			s := cloneBuilder.Grammar.CompileComponentWheres()
			var offset int
			if cloneBuilder.IsJoin {
				offset = 4 // remove leading " on "
			} else {
				offset = 7 // remove leading " where "
			}
			builder.WriteString(s[offset:])
			builder.WriteString(")")
		case CONDITION_TYPE_SUB:
			builder.WriteString(w.Boolean + " ")
			builder.WriteString(m.Wrap(w.Column))
			builder.WriteString(" " + w.Operator + " ")
			builder.WriteString("(")
			cb := NewQueryBuilder()

			if temp, ok := w.Value.(*Builder); ok {
				sql := temp.Grammar.CompileSelect()
				builder.WriteString(sql)
			} else if clousure, ok := w.Value.(func(builder *Builder)); ok {
				clousure(cb)
				sql := cb.Grammar.CompileSelect()
				builder.WriteString(sql)
			} else if clousure, ok := w.Value.(func(builder *Builder) *Builder); ok {
				sql := clousure(cb).Grammar.CompileSelect()
				builder.WriteString(sql)
			} else {
				panic("sub value must be a function")
			}
			builder.WriteString(")")
		default:
			builder.WriteString(m.CompileWhere(w))
		}

		builder.WriteString(" ")
	}

	str := builder.String()

	str = strings.TrimRight(str, " ")

	str = removeLeadingBoolean(str)

	if m.GetBuilder().IsJoin {
		str = " on " + str
	} else {
		str = " where " + str
	}
	return str
}
func removeLeadingBoolean(str string) string {
	re := regexp.MustCompile(`(?)(^and\s|^or\s)`)
	return re.ReplaceAllString(str, "")
}
func (m *MysqlGrammar) CompileWhere(w Where) (sql string) {
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString(w.Boolean + " ")
	switch w.Type {
	case CONDITION_TYPE_BASIC:
		sqlBuilder.WriteString(m.Wrap(w.Column))
		sqlBuilder.WriteString(" " + w.Operator + " ")
		sqlBuilder.WriteString(m.parameter(w.Value))
	case CONDITION_TYPE_BETWEEN:
		sqlBuilder.WriteString(m.Wrap(w.Column))
		if w.Not {
			sqlBuilder.WriteString(" not between ")
		} else {
			sqlBuilder.WriteString(" between ")
		}
		sqlBuilder.WriteString(m.parameter(w.Values[0]))
		sqlBuilder.WriteString(" and ")
		sqlBuilder.WriteString(m.parameter(w.Values[1]))
	case CONDITION_TYPE_BETWEEN_COLUMN:
		sqlBuilder.WriteString(m.Wrap(w.Column))
		if w.Not {
			sqlBuilder.WriteString(" not between ")
		} else {
			sqlBuilder.WriteString(" between ")
		}
		sqlBuilder.WriteString(m.Wrap(w.Values[0]))
		sqlBuilder.WriteString(" and ")
		sqlBuilder.WriteString(m.Wrap(w.Values[1]))
	case CONDITION_TYPE_IN:
		if len(w.Values) == 0 {
			if w.Not {
				sqlBuilder.WriteString("1 = 1")
			} else {
				sqlBuilder.WriteString("0 = 1")
			}
		} else {
			sqlBuilder.WriteString(m.Wrap(w.Column))
			if w.Not {
				sqlBuilder.WriteString(" not in (")
			} else {
				sqlBuilder.WriteString(" in (")
			}
			sqlBuilder.WriteString(m.parameter(w.Values...))
			sqlBuilder.WriteString(")")
		}

	case CONDITION_TYPE_DATE, CONDITION_TYPE_TIME, CONDITION_TYPE_DAY, CONDITION_TYPE_MONTH, CONDITION_TYPE_YEAR:
		sqlBuilder.WriteString(w.Type)
		sqlBuilder.WriteString("(")
		sqlBuilder.WriteString(m.Wrap(w.Column))
		sqlBuilder.WriteString(") ")
		sqlBuilder.WriteString(w.Operator)
		sqlBuilder.WriteString(" ")
		sqlBuilder.WriteString(m.parameter(w.Value))
	case CONDITION_TYPE_NULL:
		sqlBuilder.WriteString(m.Wrap(w.Column))
		sqlBuilder.WriteString(" is ")
		if w.Not {
			sqlBuilder.WriteString("not ")
		}
		sqlBuilder.WriteString("null")
	case CONDITION_TYPE_COLUMN:
		sqlBuilder.WriteString(m.Wrap(w.FirstColumn))
		sqlBuilder.WriteString(" ")
		sqlBuilder.WriteString(w.Operator)
		sqlBuilder.WriteString(" ")
		sqlBuilder.WriteString(m.Wrap(w.SecondColumn))
	case CONDITION_TYPE_RAW:
		sqlBuilder.WriteString(string(w.RawSql.(Expression)))
	case CONDITION_TYPE_NESTED:
		sqlBuilder.WriteString("(")
		sqlBuilder.WriteString(w.Value.(*Builder).Grammar.CompileComponentWheres())
		sqlBuilder.WriteString(") ")
	case CONDITION_TYPE_EXIST:
		if w.Not {
			sqlBuilder.WriteString("not ")
		}
		sqlBuilder.WriteString("exists ")
		sqlBuilder.WriteString(fmt.Sprintf("(%s)", w.Query.ToSql()))
	case CONDITION_TYPE_ROW_VALUES:
		var columns []interface{}
		for _, column := range w.Columns {
			columns = append(columns, column)
		}
		sqlBuilder.WriteString(fmt.Sprintf("(%s) %s (%s)", m.columnize(columns), w.Operator, m.parameter(w.Values...)))
	case CONDITION_TYPE_JSON_CONTAINS:

	default:
		panic("where type not Found")
	}
	return sqlBuilder.String()

}
func (m *MysqlGrammar) parameter(values ...interface{}) string {
	var ps []string
	for _, value := range values {
		if expre, ok := value.(Expression); ok {
			ps = append(ps, string(expre))
		} else {
			ps = append(ps, "?")
		}
	}
	return strings.Join(ps, ",")
}
func (m *MysqlGrammar) CompileComponentGroups() string {
	builder := strings.Builder{}
	builder.WriteString(" group by ")
	builder.WriteString(m.columnize(m.GetBuilder().Groups))
	return builder.String()
}

func (m *MysqlGrammar) CompileComponentHavings(query ...Builder) string {
	builder := strings.Builder{}
	builder.WriteString(" having ")
	hs := m.GetBuilder().Havings
	if len(query) > 0 {
		hs = query[0].Havings
	}
	for i, having := range hs {
		if i != 0 {
			builder.WriteString(" " + having.HavingBoolean + " ")
		}
		switch having.HavingType {
		case CONDITION_TYPE_BASIC:
			builder.WriteString(m.Wrap(having.HavingColumn))
			builder.WriteString(" ")
			builder.WriteString(having.HavingOperator)
			builder.WriteString(" ")
			builder.WriteString(m.parameter(having.HavingValue))
		case CONDITION_TYPE_RAW:
			builder.WriteString(string(having.RawSql.(Expression)))
		case CONDITION_TYPE_BETWEEN:
			vs := having.HavingValue.([]interface{})
			builder.WriteString(m.Wrap(having.HavingColumn))
			builder.WriteString(" ")
			builder.WriteString(CONDITION_TYPE_BETWEEN)
			builder.WriteString(" ")
			builder.WriteString(m.parameter(vs[0]))
			builder.WriteString(" and ")
			builder.WriteString(m.parameter(vs[1]))
		case CONDITION_TYPE_NESTED:
			builder.WriteString("(")
			compiled := m.CompileComponentHavings(having.HavingQuery)[8:]
			builder.WriteString(compiled)
			builder.WriteString(")")
		case CONDITION_TYPE_NULL:
			builder.WriteString(m.Wrap(having.HavingColumn))
			builder.WriteString(" is ")
			if having.Not {
				builder.WriteString("not ")
			}
			builder.WriteString("null")
		}

	}
	return builder.String()
}

func (m *MysqlGrammar) CompileComponentOrders() string {
	builder := strings.Builder{}
	builder.WriteString(" order by ")
	for i, order := range m.GetBuilder().Orders {
		if i != 0 {
			builder.WriteString(", ")
		}
		if order.OrderType == CONDITION_TYPE_RAW {
			builder.WriteString(string(order.RawSql.(Expression)))
			continue
		}
		builder.WriteString(m.Wrap(order.Column))
		builder.WriteString(" ")
		builder.WriteString(order.Direction)
	}
	return builder.String()
}

func (m *MysqlGrammar) CompileComponentLimitNum() string {
	if m.GetBuilder().LimitNum >= 0 {
		builder := strings.Builder{}
		builder.WriteString(" limit ")
		builder.WriteString(fmt.Sprintf("%v", m.GetBuilder().LimitNum))
		return builder.String()
	}
	return ""

}

func (m *MysqlGrammar) CompileComponentOffsetNum() string {
	if m.GetBuilder().OffsetNum >= 0 {
		builder := strings.Builder{}
		builder.WriteString(" offset ")
		builder.WriteString(fmt.Sprintf("%v", m.GetBuilder().OffsetNum))
		return builder.String()
	}
	return ""
}
func (m *MysqlGrammar) CompileLock() string {
	switch m.GetBuilder().LockMode.(type) {
	case string:
		return " " + m.GetBuilder().LockMode.(string)
	case bool:
		boolean := m.GetBuilder().LockMode.(bool)
		if boolean {
			return " for update"
		} else {
			return " lock in share mode"
		}
	case nil:
		return " for update"
	}
	return ""
}
func (m *MysqlGrammar) columnize(columns []interface{}) string {
	builder := strings.Builder{}
	var t []string
	for _, value := range columns {
		if s, ok := value.(string); ok {
			t = append(t, m.Wrap(s))
		} else if e, ok := value.(Expression); ok {
			t = append(t, string(e))
		}
	}
	builder.WriteString(strings.Join(t, ", "))
	return builder.String()
}

/*
Wrap a value in keyword identifiers.
*/
func (m *MysqlGrammar) Wrap(value interface{}, prefixAlias ...bool) string {
	prefix := false
	if expr, ok := value.(Expression); ok {
		return string(expr)
	}
	str := value.(string)
	if strings.Contains(str, " as ") || strings.Contains(str, " AS ") {
		if len(prefixAlias) > 0 && prefixAlias[0] {
			prefix = true
		}
		return m.WrapAliasedValue(str, prefix)
	}
	return m.WrapSegments(strings.Split(str, "."))
}

func (m *MysqlGrammar) WrapAliasedValue(value string, prefixAlias ...bool) string {
	var result strings.Builder
	separator := " as "
	segments := strings.SplitN(value, separator, 2)
	if len(segments) == 1 {
		separator = " AS "
		segments = strings.SplitN(value, separator, 2)
	}
	if len(prefixAlias) > 0 && prefixAlias[0] {
		segments[1] = m.GetTablePrefix() + segments[1]
	}
	result.WriteString(m.Wrap(segments[0]))
	result.WriteString(" as ")
	result.WriteString(m.WrapValue(segments[1]))
	return result.String()
}

/*
WrapSegments Wrap the given value segments.

user.name => "prefix_user"."name"
*/
func (m *MysqlGrammar) WrapSegments(values []string) string {
	var segments []string
	paramLength := len(values)
	for i, value := range values {
		if paramLength > 1 && i == 0 {
			segments = append(segments, m.WrapTable(value))
		} else {
			segments = append(segments, m.WrapValue(value))
		}
	}
	return strings.Join(segments, ".")
}

/*
WrapTable wrap a table in keyword identifiers.
*/
func (m *MysqlGrammar) WrapTable(tableName interface{}) string {
	if str, ok := tableName.(string); ok {
		return m.Wrap(m.GetTablePrefix()+str, true)
	} else if expr, ok := tableName.(Expression); ok {
		return string(expr)
	} else {
		panic(errors.New("tablename type mismatch"))
	}
}

/*
WrapValue Wrap a value in keyword identifiers.
*/
func (m *MysqlGrammar) WrapValue(value string) string {
	if value != "*" {
		return fmt.Sprintf("`%s`", strings.ReplaceAll(value, "`", "``"))
	}
	return value
}

/*
CompileRandom Compile the random statement into SQL.
*/
func (m *MysqlGrammar) CompileRandom(seed ...int) string {
	if len(seed) > 0 {
		return fmt.Sprintf("RAND(%d)", seed[0])
	}
	return "RAND()"
}
func (m *MysqlGrammar) CompileUpsert(values []map[string]interface{}, uniqueColumns []string, updateColumns interface{}) string {
	sql := m.CompileInsert(values)
	sql = sql + " on duplicate key update "

	var columns []string
	switch t := updateColumns.(type) {
	case nil:
		for s, _ := range values[0] {
			columns = append(columns, fmt.Sprintf("%s = values(%s)", m.Wrap(s), m.Wrap(s)))
		}
	case []string:
		for _, column := range updateColumns.([]string) {
			columns = append(columns, fmt.Sprintf("%s = values(%s)", m.Wrap(column), m.Wrap(column)))
		}
	case map[string]interface{}:
		for c, column := range updateColumns.(map[string]interface{}) {
			columns = append(columns, fmt.Sprintf("%s = %s, ", m.Wrap(c), m.parameter(column)))
		}
	default:
		panic(fmt.Sprintf("wrong type:%v", t))
	}

	return sql + strings.Join(columns, ", ")
}
