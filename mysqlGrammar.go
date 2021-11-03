package goeloquent

import (
	"fmt"
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
	m.compileComponentTable()
	b.PreSql.WriteString(" ( ")
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
	m.columnize(columnizeVars)
	b.PreSql.WriteString(" ) values ")

	for k, v := range values {
		b.PreSql.WriteString(" ( ")
		for i, key := range columns {
			b.PreSql.WriteString(m.parameter(v[key]))
			if i != columnLength-1 {
				b.PreSql.WriteString(" , ")
			} else {
				b.PreSql.WriteString(" ) ")
			}
		}
		if k != length-1 {
			b.PreSql.WriteString(" , ")
		}
	}
	b.PreparedSql = b.PreSql.String()
	return b.PreparedSql
}

func (m *MysqlGrammar) CompileDelete() string {
	b := m.GetBuilder()
	b.PreSql.WriteString(" delete from ")
	m.compileComponentTable()
	m.compileComponentWheres()
	m.GetBuilder().PreparedSql = m.GetBuilder().PreSql.String()

	return m.GetBuilder().PreparedSql
}

func (m *MysqlGrammar) CompileUpdate(value map[string]interface{}) string {
	b := m.GetBuilder()
	b.PreSql.WriteString("update ")
	m.compileComponentTable()
	b.PreSql.WriteString(" set ")
	count := 0
	length := len(value)
	for k, v := range value {
		count++
		if (b.OnlyColumns == nil && b.ExceptColumns == nil) || b.FileterColumn(k) {
			b.PreSql.WriteString(m.Wrap(k))
			b.PreSql.WriteString(" = ")
			if e, ok := v.(Expression); ok {
				b.PreSql.WriteString(e.Value)
			} else {
				b.PreSql.WriteString(m.parameter(v))
			}
		}
		if count != length {
			b.PreSql.WriteString(" , ")
		}
	}
	m.compileComponentWheres()
	m.GetBuilder().PreparedSql = m.GetBuilder().PreSql.String()
	return m.GetBuilder().PreparedSql
}

func (m *MysqlGrammar) CompileSelect() string {
	b := m.GetBuilder()
	for _, componentName := range SelectComponents {
		if _, ok := b.Components[componentName]; ok {
			m.compileComponent(componentName)
		}
	}
	b.PreparedSql = b.PreSql.String()
	return b.PreSql.String()
}

func (m *MysqlGrammar) CompileExists() string {
	panic("implement me")
}

func (m *MysqlGrammar) compileComponent(componentName string) {
	switch componentName {
	case "aggregate":
		m.compileComponentAggregate()
	case "columns":
		m.compileComponentColumns()
	case "from":
		m.compileComponentFromTable()
	case "joins":
		m.compileComponentJoins()
	case "wheres":
		m.compileComponentWheres()
	case "groups":
		m.compileComponentGroups()
	case "havings":
		m.compileComponentHavings()
	case "orders":
		m.compileComponentOrders()
	case "limit":
		m.compileComponentLimitNum()
	case "offset":
		m.compileComponentOffsetNum()
	case "unions":
	case "lock":
		m.compileLock()
	}
}
func (m *MysqlGrammar) compileComponentAggregate() {
	aggregate := m.GetBuilder().Aggregates[0]
	m.GetBuilder().PreSql.WriteString("select ")
	m.GetBuilder().PreSql.WriteString(aggregate.AggregateName)
	m.GetBuilder().PreSql.WriteString("(")
	if m.GetBuilder().IsDistinct && aggregate.AggregateColumn != "*" {
		m.GetBuilder().PreSql.WriteString("distinct ")
		m.GetBuilder().PreSql.WriteString(m.Wrap(aggregate.AggregateColumn))
	} else {
		m.GetBuilder().PreSql.WriteString(m.Wrap(aggregate.AggregateColumn))
	}
	m.GetBuilder().PreSql.WriteString(") as aggregate")
}

// Convert []string column names into a delimited string.
// Compile the "select *" portion of the query.
func (m *MysqlGrammar) compileComponentColumns() {
	if m.GetBuilder().IsDistinct {
		m.GetBuilder().PreSql.WriteString("select distinct ")
	} else {
		m.GetBuilder().PreSql.WriteString("select ")
	}
	m.columnize(m.GetBuilder().Columns)
}

func (m *MysqlGrammar) compileComponentFromTable() {
	m.GetBuilder().PreSql.WriteString(" from ")
	m.GetBuilder().PreSql.WriteString(m.WrapTable(m.GetBuilder().FromTable))
}
func (m *MysqlGrammar) compileComponentTable() {
	m.GetBuilder().PreSql.WriteString(m.WrapTable(m.GetBuilder().FromTable))
}
func (m *MysqlGrammar) compileComponentJoins() {
	for _, join := range m.GetBuilder().Joins {
		//left join test_userinfo on user.id = test_userinfo.user_id
		s := fmt.Sprintf(" %s join %s %s %s %s %s", join.JoinType, m.GetTablePrefix()+join.JoinTable, join.JoinOperator, m.Wrap(join.FirstColumn), join.ColumnOperator, m.Wrap(join.SecondColumn))
		m.GetBuilder().PreSql.WriteString(s)
	}

}

//	db.Query("select * from `groups` where `id` in (?,?,?,?)", []interface{}{7,"8","9","10"}...)
func (m *MysqlGrammar) compileComponentWheres() {
	if len(m.GetBuilder().Wheres) == 0 {
		return
	}
	m.GetBuilder().PreSql.WriteString(" where ")
	for i, w := range m.GetBuilder().Wheres {
		if i != 0 {
			m.GetBuilder().PreSql.WriteString(" " + w.Boolean + " ")
		}
		if w.Type == CONDITION_TYPE_NESTED {
			m.GetBuilder().PreSql.WriteString("(")
			cloneBuilder := w.Value.(*Builder)
			for j := 0; j < len(cloneBuilder.Wheres); j++ {
				nestedWhere := cloneBuilder.Wheres[j]
				if j != 0 {
					m.GetBuilder().PreSql.WriteString(" " + nestedWhere.Boolean + " ")
				}
				//when compile nested where,we need bind the generated sql and params to the current builder
				g := cloneBuilder.Grammar.(*MysqlGrammar)
				//this will bind params to current builder when call m.parameter()
				g.SetBuilder(m.GetBuilder())
				nestedSql := g.compileWhere(nestedWhere)
				m.GetBuilder().PreSql.WriteString(nestedSql)
			}
			m.GetBuilder().PreSql.WriteString(")")
		} else if w.Type == CONDITION_TYPE_SUB {
			m.GetBuilder().PreSql.WriteString(m.Wrap(w.Column))
			m.GetBuilder().PreSql.WriteString(" " + w.Operator + " ")
			m.GetBuilder().PreSql.WriteString("(")
			cb := CloneBuilder(m.GetBuilder())

			if clousure, ok := w.Value.(func(builder *Builder)); ok {
				clousure(cb)
				sql := cb.Grammar.CompileSelect()
				m.GetBuilder().PreSql.WriteString(sql)
				m.GetBuilder().Bindings = append(m.GetBuilder().Bindings, cb.Bindings...)
			}
			m.GetBuilder().PreSql.WriteString(")")

		} else {
			m.GetBuilder().PreSql.WriteString(m.compileWhere(w))
		}

	}
}
func (m *MysqlGrammar) compileWhere(w Where) (sql string) {
	var sqlBuilder strings.Builder
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
	case CONDITION_TYPE_IN:
		sqlBuilder.WriteString(m.Wrap(w.Column))
		if w.Not {
			sqlBuilder.WriteString(" not in (")
		} else {
			sqlBuilder.WriteString(" in (")
		}
		sqlBuilder.WriteString(m.parameter(w.Values...))
		sqlBuilder.WriteString(")")
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
			sqlBuilder.WriteString(" not ")
		}
		sqlBuilder.WriteString("null ")
	case CONDITION_TYPE_COLUMN:
		sqlBuilder.WriteString(m.Wrap(w.FirstColumn))
		sqlBuilder.WriteString(" ")
		sqlBuilder.WriteString(w.Operator)
		sqlBuilder.WriteString(" ")
		sqlBuilder.WriteString(m.Wrap(w.SecondColumn))
	case CONDITION_TYPE_RAW:
		sqlBuilder.WriteString(w.RawSql)
	default:
		panic("where type not Found")
	}
	return sqlBuilder.String()

}
func (m *MysqlGrammar) parameter(values ...interface{}) string {
	var ps []string
	for _, v := range values {
		m.GetBuilder().Bindings = append(m.GetBuilder().Bindings, v)
		ps = append(ps, "?")
	}
	return strings.Join(ps, ",")
}
func (m *MysqlGrammar) compileComponentGroups() {
	m.GetBuilder().PreSql.WriteString(" group by ")
	m.columnize(m.GetBuilder().Groups)
}

func (m *MysqlGrammar) compileComponentHavings() {
	m.GetBuilder().PreSql.WriteString(" having ")
	for i, having := range m.GetBuilder().Havings {
		if i != 0 {
			m.GetBuilder().PreSql.WriteString(" " + having.HavingBoolean + " ")
		}
		if having.HavingType == CONDITION_TYPE_RAW {
			m.GetBuilder().PreSql.WriteString(having.RawSql)
		} else {
			m.GetBuilder().PreSql.WriteString(m.Wrap(having.HavingColumn))
			m.GetBuilder().PreSql.WriteString(" ")
			m.GetBuilder().PreSql.WriteString(having.HavingOperator)
			m.GetBuilder().PreSql.WriteString(" ")
			m.GetBuilder().PreSql.WriteString(m.parameter(having.HavingValue))
		}
	}
}

func (m *MysqlGrammar) compileComponentOrders() {
	m.GetBuilder().PreSql.WriteString(" order by ")
	for _, order := range m.GetBuilder().Orders {
		m.GetBuilder().PreSql.WriteString(m.Wrap(order.Column))
		m.GetBuilder().PreSql.WriteString(" ")
		m.GetBuilder().PreSql.WriteString(order.Direction)
	}
}

func (m *MysqlGrammar) compileComponentLimitNum() {
	m.GetBuilder().PreSql.WriteString(" limit ")
	m.GetBuilder().PreSql.WriteString(fmt.Sprintf("%v", m.GetBuilder().LimitNum))
}

func (m *MysqlGrammar) compileComponentOffsetNum() {
	m.GetBuilder().PreSql.WriteString(" offset ")
	m.GetBuilder().PreSql.WriteString(fmt.Sprintf("%v", m.GetBuilder().OffsetNum))
}
func (m *MysqlGrammar) compileLock() {
	if len(m.GetBuilder().Lock) > 0 {
		m.GetBuilder().PreSql.WriteString(m.GetBuilder().Lock)
	}
}
func (m *MysqlGrammar) columnize(columns []interface{}) {
	var t []string
	for _, value := range columns {
		if s, ok := value.(string); ok {
			t = append(t, m.Wrap(s))
		} else if e, ok := value.(Expression); ok {
			t = append(t, e.Value)
		}
	}
	m.GetBuilder().PreSql.WriteString(strings.Join(t, ","))
}
func (m *MysqlGrammar) Wrap(value string) string {
	if strings.Contains(value, " as ") {
		return m.WrapAliasedValue(value)
	}
	return m.WrapSegments(strings.Split(value, "."))
}

func (m *MysqlGrammar) WrapAliasedValue(value string) string {
	var result strings.Builder
	separator := " as "
	segments := strings.SplitN(value, separator, 2)
	result.WriteString(m.Wrap(segments[0]))
	result.WriteString(" as ")
	result.WriteString(m.WrapValue(segments[1]))
	return result.String()
}

//user.name => "prefix_user"."name"
func (m *MysqlGrammar) WrapSegments(values []string) string {
	var segments []string
	paramLength := len(values)
	if paramLength > 1 {
		segments = append(segments, m.WrapTable(values[0]))
	}
	segments = append(segments, m.WrapValue(values[paramLength-1]))
	return strings.Join(segments, ".")
}
func (m *MysqlGrammar) WrapTable(tableName string) string {
	return m.Wrap(m.GetTablePrefix() + tableName)
}

// table => `table`
// t1"t2 => `t1``t2`
func (m *MysqlGrammar) WrapValue(value string) string {
	if value != "*" {
		return fmt.Sprintf("`%s`", strings.ReplaceAll(value, "`", "``"))
	}
	return value
}
