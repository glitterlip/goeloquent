package goeloquent

type IGrammar interface {
	SetTablePrefix(prefix string)

	GetTablePrefix() string

	SetBuilder(builder *Builder)

	GetBuilder() *Builder

	CompileInsert([]map[string]interface{}) string

	CompileInsertOrIgnore([]map[string]interface{}) string

	CompileDelete() string

	CompileUpdate(map[string]interface{}) string

	CompileSelect() string

	CompileExists() string

	Wrap(interface{}, ...bool) string

	WrapTable(interface{}) string

	CompileComponentWheres() string

	CompileComponentJoins() string
	//Wrap(value string, b *query.Builder) string
}
