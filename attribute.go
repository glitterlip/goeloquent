package goeloquent

import (
	"database/sql"
	"reflect"
)

/*
GetOrigin Get the original attribute values.
*/
func (m *EloquentModel) GetOrigin() map[string]interface{} {
	return m.Origin
}

/*
GetChanges Get the attributes that were changed.
*/
func (m *EloquentModel) GetChanges() map[string]interface{} {
	return m.Changes
}

/*
GetDirty Get the attributes that have been changed since the last sync.
*/
func (m *EloquentModel) GetDirty() map[string]interface{} {
	parsed := GetParsedModel(reflect.Indirect(m.ModelPointer).Type())
	dirty := make(map[string]interface{})
	current := reflect.Indirect(m.ModelPointer)
	for key, _ := range m.Origin {
		fieldIndex := parsed.FieldsByStructName[key].Index
		currentField := current.Field(fieldIndex)
		var equal bool
		switch m.Origin[key].(type) {
		case sql.NullString:
			c := currentField.Interface().(sql.NullString)
			o := m.Origin[key].(sql.NullString)
			equal = c.Valid == o.Valid && c.String == o.String
		case sql.NullBool:
			c := currentField.Interface().(sql.NullBool)
			o := m.Origin[key].(sql.NullBool)
			equal = c.Valid == o.Valid && c.Bool == o.Bool
		case sql.NullFloat64:
			c := currentField.Interface().(sql.NullFloat64)
			o := m.Origin[key].(sql.NullFloat64)
			equal = c.Valid == o.Valid && c.Float64 == o.Float64
		case sql.NullInt32:
			c := currentField.Interface().(sql.NullInt32)
			o := m.Origin[key].(sql.NullInt32)
			equal = c.Valid == o.Valid && c.Int32 == o.Int32
		case sql.NullInt64:
			c := currentField.Interface().(sql.NullInt64)
			o := m.Origin[key].(sql.NullInt64)
			equal = c.Valid == o.Valid && c.Int64 == o.Int64
		case sql.NullTime:
			c := currentField.Interface().(sql.NullTime)
			o := m.Origin[key].(sql.NullTime)
			equal = c.Valid == o.Valid && c.Time == c.Time
		default:
			equal = currentField.Interface() == m.Origin[key]
		}
		if !equal {
			dirty[key] = currentField.Interface()
		}
	}
	return dirty
}
