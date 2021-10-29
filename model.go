package goeloquent

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	EloquentName = "EloquentModel"
)

const (
	EventSaving    = "Saving"
	EventSaved     = "Saved"
	EventCreating  = "Creating"
	EventCreated   = "Created"
	EventUpdating  = "Updating"
	EventUpdated   = "Updated"
	EventDeleteing = "Deleting"
	EventDeleted   = "Deleted"
)

type TableName interface {
	TableName() string
}
type ConnectionName interface {
	ConnectionName() string
}
type Model struct {
	Name               string
	ModelType          reflect.Type
	Table              string
	DbFields           []string
	FieldsByDbName     map[string]*Field //map of fields of database table columns
	Fields             []*Field          //all struct fields
	FieldsByStructName map[string]*Field
	Relations          map[string]reflect.Value
	RelationFieldNames map[string]string
	RelationNames      []string
	ConnectionName     string
	Exists             bool //exists in database
	PrimaryKey         *Field
	IsEloquent         bool
	PivotFieldIndex    []int
	//DefaultAttributes  map[string]interface{}
	UpdatedAt  string
	CreatedAt  string
	DeletedAt  string
	SoftDelete bool
	//EagerLoad     []
	//GlobalScopes map[string]interface{}
	DispatchesEvents map[string]reflect.Value
	DB               *DB
}
type Field struct {
	Name       string
	ColumnName string
	Index      int
	//BindNames  []string
	//DataType               DataType
	//GORMDataType           DataType
	//PrimaryKey             bool
	//AutoIncrement          bool
	//AutoIncrementIncrement int64
	//Creatable              bool
	//Updatable              bool
	//Readable               bool
	//HasDefaultValue        bool
	//AutoCreateTime         TimeType
	//AutoUpdateTime         TimeType
	//DefaultValue           string
	//DefaultValueInterface  interface{}
	//DefaultValueInterface  interface{}

	FieldType         reflect.Type
	IndirectFieldType reflect.Type
	StructField       reflect.StructField
	Tag               reflect.StructTag
	TagSettings       map[string]string
	Model             *Model
	//EmbeddedSchema         *Schema
	//OwnerSchema            *Schema
	//ReflectValueOf func(reflect.Value) reflect.Value
	//ValueOf        func(reflect.Value) (value interface{}, zero bool)
}
type HasTimestamps interface {
	//Update the model's update timestamp
	Touch() bool
	//
	SetCreatedAt() Builder
	SetUpdatedAt() Builder
	GetCreatedAtColumn() Field
	GetUpdatedAtColumn() Field
}

func Parse(modelType reflect.Type) (model *Model, err error) {
	//cache
	modelValue := reflect.New(modelType)
	var tableName string
	if t, ok := modelValue.Interface().(TableName); ok {
		tableName = t.TableName()
	} else {
		tableName = ToSnakeCase(modelType.Name())
	}
	model = &Model{
		Name:               modelType.Name(),
		ModelType:          modelType,
		Table:              tableName,
		FieldsByDbName:     make(map[string]*Field),
		FieldsByStructName: make(map[string]*Field),
		Relations:          make(map[string]reflect.Value),
		PrimaryKey:         nil,
		DispatchesEvents:   make(map[string]reflect.Value),
		//DefaultAttributes:  make(map[string]interface{}),
		//RelationTypes:      make(map[string]reflect.Type),
	}
	for i := 0; i < modelType.NumField(); i++ {
		model.ParseField(modelType.Field(i))
	}
	funcs := map[string]string{
		EventSaving:    EventSaving,
		EventSaved:     EventSaved,
		EventCreating:  EventCreating,
		EventCreated:   EventCreated,
		EventUpdating:  EventUpdating,
		EventUpdated:   EventUpdated,
		EventDeleteing: EventDeleteing,
		EventDeleted:   EventDeleted,
	}
	ptrReciver := reflect.PtrTo(modelType)
	for i := 0; i < ptrReciver.NumMethod(); i++ {
		if name, ok := funcs[ptrReciver.Method(i).Name]; ok {
			model.DispatchesEvents[name] = modelValue.Method(i)
		}
	}
	if model.PrimaryKey == nil {
		model.PrimaryKey = model.FieldsByDbName["id"]
	}

	//callbacks
	//'retrieved', 'creating', 'created', 'updating', 'updated',
	//                'saving', 'saved', 'restoring', 'restored', 'replicating',
	//                'deleting', 'deleted', 'forceDeleted',
	return model, err
}
func (m *Model) ParseField(field reflect.StructField) *Field {
	modelField := &Field{
		Name:              field.Name,
		ColumnName:        field.Name,
		FieldType:         field.Type,
		IndirectFieldType: field.Type,
		StructField:       field,
		Tag:               field.Tag,
		TagSettings:       make(map[string]string),
		Model:             m,
		Index:             field.Index[0],
		//ReflectValueOf:    nil,
		//ValueOf:           nil,
	}
	tag, ok := field.Tag.Lookup("goelo")
	if ok {
		for _, pair := range strings.Split(tag, ";") {
			as := strings.SplitN(pair, ":", 2)
			key := as[0]
			var value string
			if len(as) > 1 {
				value = as[1]
			}
			switch key {
			case "column":
				modelField.ColumnName = value
				m.FieldsByDbName[modelField.ColumnName] = modelField
				m.DbFields = append(m.DbFields, modelField.ColumnName)
			case "primaryKey":
				m.PrimaryKey = modelField
			case "CREATED_AT":
				m.CreatedAt = modelField.ColumnName
			case "UPDATED_AT":
				m.UpdatedAt = modelField.ColumnName
			case "DELETED_AT":
				m.DeletedAt = modelField.ColumnName
				m.SoftDelete = true
			case RelationHasMany,
				RelationHasOne,
				RelationBelongsToMany,
				RelationBelongsTo,
				RelationHasManyThrough,
				RelationHasOneThrough,
				RelationMorphByMany,
				RelationMorphOne,
				RelationMorphMany,
				RelationMorphTo,
				RelationMorphToMany:
				methodName := value
				v := reflect.New(m.ModelType)
				m.Relations[field.Name] = v.MethodByName(methodName)
				if !m.Relations[field.Name].IsValid() {
					panic(fmt.Sprintf("relation method %s on model %s not found", methodName, m.Name))
				}
				m.RelationNames = append(m.RelationNames, field.Name)

				//todo case defaultvalue,case updateable,insertable
			}
		}
	}
	//if  { ignored tag
	m.Fields = append(m.Fields, modelField)
	m.FieldsByStructName[modelField.Name] = modelField
	if modelField.Name == EloquentName {
		m.IsEloquent = true
		m.PivotFieldIndex = append(m.PivotFieldIndex, modelField.Index)
		embeded := reflect.New(m.ModelType).Elem().Field(modelField.Index).Type()
		for i := 0; i < embeded.NumField(); i++ {
			if embeded.Field(i).Name == "Pivot" {
				m.PivotFieldIndex = append(m.PivotFieldIndex, i)

			}
		}
	}
	//}

	return modelField
}

func (m *Model) GetConnection() *IConnection {
	return m.DB.Connection("default")
}
func GetParsedModel(model interface{}) *Model {
	var target reflect.Type

	if t, ok := model.(reflect.Type); ok {
		target = t
	} else if s, ok := model.(string); ok {
		target = Eloquent.Models[s].Type()
	} else {
		value := reflect.ValueOf(model)
		if value.Kind() != reflect.Ptr {
			panic("must be a pointer")
		}
		modelValue := reflect.Indirect(value)

		switch modelValue.Kind() {
		case reflect.Array, reflect.Slice:
			target = modelValue.Type().Elem()
			if target.Kind() == reflect.Ptr {
				target = target.Elem()
			}
		default:
			target = modelValue.Type()
		}
	}

	parsed, _ := Parse(target)
	return parsed

}
func BatchSync(models interface{}, exists ...bool) {
	var exist bool
	if len(exists) == 0 {
		exist = false
	} else {
		exist = exists[0]
	}
	var realModels reflect.Value
	if v, ok := models.(*reflect.Value); ok {
		realModels = reflect.Indirect(*v)
	} else {
		realModels = reflect.Indirect(reflect.ValueOf(models))
	}

	if realModels.Type().Kind() == reflect.Slice {
		item := realModels.Type().Elem()
		var parsed *Model
		if item.Kind() == reflect.Ptr {
			parsed = GetParsedModel(item.Elem())
		} else {
			parsed = GetParsedModel(item)
		}
		if parsed.IsEloquent {
			for i := 0; i < realModels.Len(); i++ {
				model := realModels.Index(i)
				if model.Kind() == reflect.Ptr {
					model = model.Elem()
				}
				model.Field(parsed.FieldsByStructName[EloquentName].Index).Set(reflect.ValueOf(NewEloquentModel(model.Addr().Interface(), exist)))
			}
		}
	} else if realModels.Type().Kind() == reflect.Struct {
		item := realModels
		p := GetParsedModel(item.Type())
		if p.IsEloquent {
			item.Field(p.FieldsByStructName[EloquentName].Index).Set(reflect.ValueOf(NewEloquentModel(item.Addr().Interface(), exist)))
		}
	}
	return
}

type EloquentModel struct {
	Origin        map[string]interface{}    `json:"-"` //store original attribute that get from database or default
	Changes       map[string]interface{}    `json:"-"` //store changes attribute after save to database
	ModelPointer  reflect.Value             `json:"-"` //model pointer points to the model hold this
	Pivot         map[string]sql.NullString `json:"-"` //pivot relation table attribute
	Exists        bool                      `json:"-"` //indicate whether the model is get from database or newly created and not store to db yet
	Related       reflect.Value             `json:"-"` //when call save/create on relationship ,this holds the related key
	Muted         []string                  `json:"-"` //mute events
	OnlyColumns   map[string]interface{}    `json:"-"` //only update/save these columns
	ExceptColumns map[string]interface{}    `json:"-"` //exclude update/save there columns
}

func NewModel(modelPointer interface{}) *EloquentModel {
	e := NewEloquentModel(modelPointer, false)
	m := reflect.Indirect(reflect.ValueOf(modelPointer))
	parsed := GetParsedModel(m.Type())
	m.Field(parsed.PivotFieldIndex[0]).Set(reflect.ValueOf(e))
	return &e
}
func InitModel(modelPointer interface{}, exists ...bool) *EloquentModel {
	m := reflect.Indirect(reflect.ValueOf(modelPointer))
	parsed := GetParsedModel(m.Type())
	if len(exists) == 0 {
		exists = []bool{m.Field(parsed.PrimaryKey.Index).IsZero()}
	}
	e := NewEloquentModel(modelPointer, exists[0])
	m.Field(parsed.PivotFieldIndex[0]).Set(reflect.ValueOf(e))
	return &e
}
func NewEloquentModel(modelPointer interface{}, exists ...bool) EloquentModel {
	m := EloquentModel{
		Origin: make(map[string]interface{}),
	}
	m.ModelPointer = reflect.ValueOf(modelPointer)
	m.Origin = make(map[string]interface{}, 4)
	m.Changes = make(map[string]interface{}, 4)
	m.Pivot = make(map[string]sql.NullString, 0)
	if len(exists) == 0 {
		exists = []bool{false}
	}
	m.Exists = exists[0]
	var tmap reflect.Value
	parsed := GetParsedModel(modelPointer)
	if len(parsed.PivotFieldIndex) > 1 {
		model := reflect.Indirect(m.ModelPointer)
		tmap = model.Field(parsed.PivotFieldIndex[0]).Field(parsed.PivotFieldIndex[1])
	}
	if !tmap.IsNil() {
		m.Pivot = tmap.Interface().(map[string]sql.NullString)
	}
	m.SyncOrigin()
	return m
}
func (m *EloquentModel) IsEager() bool {
	return !m.Exists
}
func (m *EloquentModel) IsDirty(key string) bool {
	current := reflect.Indirect(m.ModelPointer)
	parsed := GetParsedModel(reflect.Indirect(m.ModelPointer).Type())
	fieldIndex := parsed.FieldsByStructName[key].Index
	keyValue := current.Field(fieldIndex)
	if !keyValue.IsValid() || keyValue.IsZero() {
		return m.Origin[key] == nil
	}
	return keyValue.Interface() == m.Origin[key]
}
func (m *EloquentModel) SyncOrigin() {
	parsed := GetParsedModel(reflect.Indirect(m.ModelPointer).Type())
	model := reflect.Indirect(m.ModelPointer)
	for _, field := range parsed.FieldsByDbName {
		m.Origin[field.StructField.Name] = model.Field(field.StructField.Index[0]).Interface()
	}
}
func (m *EloquentModel) GetOrigin() map[string]interface{} {
	return m.Origin
}

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
func (m *EloquentModel) Save() (res sql.Result, err error) {
	parsed := GetParsedModel(reflect.Indirect(m.ModelPointer).Type())
	if !m.FireModelEvent(EventSaving) {
		return nil, errors.New("abort by EventSaving func")
	}
	if !m.Exists {
		if !m.FireModelEvent(EventCreating) {
			return nil, errors.New("abort by EventCreating func")
		}
		//SetDefaults
		res, err = Eloquent.Model(parsed.ModelType).Insert(m.GetAttributesForCreate())
		if err != nil {
			return
		}
		//todo:set created_at timestamps
		id, err1 := res.LastInsertId()
		if err1 == nil {
			if parsed.PrimaryKey != nil {
				reflect.Indirect(m.ModelPointer).Field(parsed.PrimaryKey.Index).Set(reflect.ValueOf(id))
			}
		} else {
			err = err1
			return
		}
		m.Exists = true
		//m.WasRecentlyCreated = true for createOrNew createOrUpdate?
		m.FireModelEvent(EventCreated)

	} else {
		if !m.FireModelEvent(EventUpdating) {
			return nil, errors.New("abort by EventUpdating func")
		}
		//todo:set updated_at timestamps
		m.Changes = m.GetDirty()
		res, err = Eloquent.Model(parsed.ModelType).Where(parsed.PrimaryKey.ColumnName, m.ModelPointer.Elem().Field(parsed.PrimaryKey.Index).Interface()).Update(m.GetAttributesForUpdate())
		m.FireModelEvent(EventUpdated)
	}
	m.FireModelEvent(EventSaved)
	m.SyncOrigin()
	return

}

func (m *EloquentModel) Create() (sql.Result, error) {
	return m.Save()
}
func (m *EloquentModel) GetAttributesForUpdate() (attrs map[string]interface{}) {
	model := reflect.Indirect(m.ModelPointer)
	attrs = make(map[string]interface{})
	modelType := GetParsedModel(model.Type())
	var hasOnlyColumns, hasExceptColumns bool
	if m.OnlyColumns != nil && len(m.OnlyColumns) > 0 {
		hasOnlyColumns = true
	}
	if m.ExceptColumns != nil && len(m.ExceptColumns) > 0 {
		hasExceptColumns = true
	}
	for _, key := range modelType.DbFields {
		keyIndex := modelType.FieldsByDbName[key].Index
		if hasExceptColumns {
			if _, ok := m.ExceptColumns[key]; ok {
				continue
			}
			attrs[key] = model.Field(keyIndex).Interface()
			continue
		}
		if hasOnlyColumns {
			if _, ok := m.OnlyColumns[key]; !ok {
				continue
			}
			attrs[key] = model.Field(keyIndex).Interface()
			continue
		}
		if !model.Field(keyIndex).IsZero() && keyIndex != modelType.PrimaryKey.Index {
			attrs[key] = model.Field(keyIndex).Interface()
		}
	}
	if modelType.UpdatedAt != "" {
		attrs[modelType.UpdatedAt] = time.Now()
	}
	return
}
func (m *EloquentModel) GetAttributesForCreate() (attrs map[string]interface{}) {
	model := reflect.Indirect(m.ModelPointer)
	attrs = make(map[string]interface{})
	modelType := GetParsedModel(model.Type())
	var hasOnlyColumns, hasExceptColumns bool
	if m.OnlyColumns != nil && len(m.OnlyColumns) > 0 {
		hasOnlyColumns = true
	}
	if m.ExceptColumns != nil && len(m.ExceptColumns) > 0 {
		hasExceptColumns = true
	}
	for _, key := range modelType.DbFields {
		keyIndex := modelType.FieldsByDbName[key].Index
		if hasExceptColumns {
			if _, ok := m.ExceptColumns[key]; ok {
				continue
			}
			attrs[key] = model.Field(keyIndex).Interface()
			continue
		}
		if hasOnlyColumns {
			if _, ok := m.OnlyColumns[key]; !ok {
				continue
			}
			attrs[key] = model.Field(keyIndex).Interface()
			continue
		}
		if !model.Field(keyIndex).IsZero() {
			attrs[key] = model.Field(keyIndex).Interface()
		}
	}
	if modelType.CreatedAt != "" {
		attrs[modelType.CreatedAt] = time.Now()
	}
	return

}
func (m *EloquentModel) GetAttributes() (attrs map[string]interface{}) {

	model := reflect.Indirect(m.ModelPointer)
	attrs = make(map[string]interface{})
	modelType := GetParsedModel(model.Type())
	for _, key := range modelType.DbFields {
		keyIndex := modelType.FieldsByDbName[key].Index
		attrs[key] = model.Field(keyIndex).Interface()
	}
	return
}

func (m *EloquentModel) FireModelEvent(name string) bool {
	parsed := GetParsedModel(reflect.Indirect(m.ModelPointer).Type())
	method, ok := parsed.DispatchesEvents[name]
	if ok {
		return method.Call([]reflect.Value{m.ModelPointer})[0].Interface().(bool)
	} else {
		return true
	}
}

//scope
//increment decrement
//fresh
//paginate
//touch
func (m *EloquentModel) Fill(attrs map[string]interface{}) *EloquentModel {
	model := reflect.Indirect(m.ModelPointer)
	modelType := GetParsedModel(model.Type())
	for k, v := range attrs {
		if f, ok := modelType.FieldsByDbName[k]; ok {
			model.Field(f.Index).Set(reflect.ValueOf(v))
		} else if f, ok := modelType.FieldsByStructName[k]; ok {
			model.Field(f.Index).Set(reflect.ValueOf(v))
		}
	}
	return m
}
func (m *EloquentModel) Only(columns ...string) *EloquentModel {
	m.OnlyColumns = make(map[string]interface{}, len(columns))
	for i := 0; i < len(columns); i++ {
		m.OnlyColumns[columns[i]] = nil
	}
	return m
}
func (m *EloquentModel) Except(columns ...string) *EloquentModel {
	m.ExceptColumns = make(map[string]interface{}, len(columns))
	for i := 0; i < len(columns); i++ {
		m.ExceptColumns[columns[i]] = nil
	}
	return m
}
func (m *EloquentModel) Load(relations ...interface{}) {
	var connection IConnection
	if c, ok := m.ModelPointer.Elem().Interface().(ConnectionName); ok {
		connection = *Eloquent.Connection(c.ConnectionName())
	} else {
		connection = *Eloquent.Connection("default")
	}
	b := NewBuilder(connection)
	b.SetModel(m.ModelPointer.Interface())
	b.With(relations...)
	b.Dest = m.ModelPointer.Interface()
	rb := RelationBuilder{
		Builder: b,
	}
	rb.EagerLoadRelations(b.Dest)
}
