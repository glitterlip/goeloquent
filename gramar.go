package goeloquent

import "github.com/glitterlip/goeloquent/query"

type IGrammar interface {
	SetTablePrefix(prefix string)

	GetTablePrefix() string

	SetBuilder(builder *query.Builder)

	GetBuilder() *query.Builder

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

	CompileRandom(seed string) string
	//Wrap(value string, b *query.Builder) string

	CompileUpsert([]map[string]interface{}, []string, interface{}) string
}
