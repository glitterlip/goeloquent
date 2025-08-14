package goeloquent

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

const (
	FieldTagPrimaryKey        = "primaryKey"
	FieldTagColumnName        = "column"
	FieldTagCreateTimestamp   = "CREATED_AT"
	FieldTagUpdateTimestamp   = "UPDATED_AT"
	FieldTagDeleteTimestamp   = "DELETED_AT"
	FieldTagAutoIncrement     = "autoIncrement"
	FieldTagCast              = "cast" //todo
	WithAggregate             = "withAggregate"
	GlobalScopeWithoutTrashed = "withoutTrashed"
)

type Field struct {
	Name        string       //reflect.StructField.Name,struct field name
	ColumnName  string       //db column name
	Index       int          //Type.FieldByIndex, struct reflect.type fields index
	FieldType   reflect.Type //reflect.StructField.Type
	Tag         reflect.StructTag
	TagSettings map[string]string
	IsPtr       bool
	Kind        reflect.Kind
	NeedCast    bool
}

func ParseField(m *ModelConfig, field reflect.StructField) (*Field, error) {
	modelField := &Field{
		Name:        field.Name,
		ColumnName:  field.Name,
		FieldType:   field.Type,
		Tag:         field.Tag,
		TagSettings: make(map[string]string),
		Index:       field.Index[0],
		IsPtr:       field.Type.Kind() == reflect.Ptr,
		Kind:        field.Type.Kind(),
		NeedCast:    false,
	}
	//switch modelField.FieldType.Kind() {
	//case reflect.Slice, reflect.Array, reflect.Map:
	//	modelField.NeedCast = true
	//case reflect.Struct:
	ifaceType := reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	if modelField.FieldType.Implements(ifaceType) {
		modelField.NeedCast = true
	}
	//}
	tag, ok := field.Tag.Lookup(EloquentTagName)
	if ok {
		for _, pair := range strings.Split(tag, ";") {
			if pair == "" {
				continue
			}
			as := strings.SplitN(pair, ":", 2)
			key := as[0]
			var value string
			if len(as) > 1 {
				value = as[1]
			}
			switch key {
			case FieldTagColumnName:
				modelField.ColumnName = value
				m.FieldsByColumnName[modelField.ColumnName] = modelField
				m.FieldsByStructName[modelField.Name] = modelField
			case FieldTagPrimaryKey:
				m.PrimaryKeyField = modelField
			case FieldTagAutoIncrement:
				if value == "false" {
					m.PrimaryKeyAutoIncrementing = false
				}
			case FieldTagCreateTimestamp:
				m.CreatedAt = &ModelTimestamp{
					IsTimestamp: field.Type != reflect.TypeOf(time.Time{}),
					ColumnName:  modelField.ColumnName,
					Enabled:     true,
					Name:        field.Name,
				}
			case FieldTagUpdateTimestamp:
				m.UpdatedAt = &ModelTimestamp{
					IsTimestamp: field.Type != reflect.TypeOf(time.Time{}),
					ColumnName:  modelField.ColumnName,
					Enabled:     true,
					Name:        field.Name,
				}
			case FieldTagDeleteTimestamp:
				m.DeletedAt = &ModelTimestamp{
					IsTimestamp: field.Type != reflect.TypeOf(time.Time{}),
					ColumnName:  modelField.ColumnName,
					Enabled:     true,
					Name:        field.Name,
				}
				m.GlobalScopes[GlobalScopeWithoutTrashed] = func(st *Statement) {
					st.WhereNull(st.QualifyColumn(m.DeletedAt))
				}
			case FieldTagCast:
				modelField.NeedCast = true
				//todo: support cast

			case string(RelationHasMany),
				string(RelationHasOne),
				string(RelationBelongsToMany),
				string(RelationBelongsTo),
				string(RelationHasManyThrough),
				string(RelationHasOneThrough),
				string(RelationMorphedByMany),
				string(RelationMorphOne),
				string(RelationMorphMany),
				string(RelationMorphTo),
				string(RelationMorphToMany):
				methodName := value
				v := reflect.New(m.ModelType)
				m.Relations[field.Name] = v.MethodByName(methodName)
				if !m.Relations[field.Name].IsValid() {
					return nil, fmt.Errorf("Model: %s %s relation %s method not found", m.ModelType.Name(), key, methodName)
				}
			case WithAggregate:
				m.FieldsByStructName[modelField.Name] = modelField
				if len(as) > 1 {
					m.Aggregates[value] = modelField.Name
				}
			default:
				return nil, fmt.Errorf("unknown tag %s in model:%s field:%s", key, m.ModelType.Name(), field.Name)
			}
		}

		m.FieldsByStructName[modelField.Name] = modelField
	}

	if modelField.Name == EloquentModelName {
		m.IsEloquent = true
		m.EloquentModelFieldIndex = modelField.Index
	}
	if !m.IsEloquent && !ok {
		m.FieldsByColumnName[ToSnakeCase(modelField.ColumnName)] = modelField
		m.FieldsByStructName[modelField.Name] = modelField
		m.PrimaryKeyAutoIncrementing = false
	}

	return modelField, nil
}
