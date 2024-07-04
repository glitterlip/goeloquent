package goeloquent

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type TableName interface {
	TableName() string
}
type ConnectionName interface {
	ConnectionName() string
}
type DynamicTableName interface {
	ResolveTableName(builder *EloquentBuilder) string
}
type DynamicConnectionName interface {
	ResolveConnectionName(builder *EloquentBuilder) string
}

const (
	EloquentName                  = "EloquentModel"
	EloquentTagName               = "goelo"
	EloquentAddGlobalScopes       = "EloquentAddGlobalScopes"
	EloquentGetDefaultAttributes  = "EloquentGetDefaultAttributes"
	EloquentGetConnection         = "EloquentGetConnection"
	EloquentGetTable              = "EloquentGetTable"
	EloquentGetFillable           = "EloquentGetFillable"
	EloquentGetGuarded            = "EloquentGetGuarded"
	EloquentGetWithRelations      = "EloquentGetWithRelations"
	EloquentGetWithRelationCounts = "EloquentGetWithRelationCounts"
)

/*
BatchSync set models' EloquentModel field
*/
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
				newModel := reflect.ValueOf(NewEloquentModel(model.Addr().Interface(), exist))
				if !model.Field(parsed.EloquentModelFieldIndex).IsNil() && !model.Field(parsed.EloquentModelFieldIndex).IsZero() {
					//if pivot is not nil ,copy old pivot to new eloquentmodel
					newModel.Elem().Field(parsed.PivotFieldIndex).Set(model.Field(parsed.EloquentModelFieldIndex))
				}
				model.Field(parsed.FieldsByStructName[EloquentName].Index).Set(newModel)

			}
		}
	} else if realModels.Type().Kind() == reflect.Struct {
		parsed := GetParsedModel(realModels.Type())
		if parsed.IsEloquent {
			model := realModels
			newModel := reflect.ValueOf(NewEloquentModel(model.Addr().Interface(), exist))
			if !model.Field(parsed.EloquentModelFieldIndex).IsNil() && !model.Field(parsed.PivotFieldIndex).IsZero() {
				newModel.Elem().Field(parsed.PivotFieldIndex).Set(model.Field(parsed.PivotFieldIndex))
			}
			realModels.Field(parsed.FieldsByStructName[EloquentName].Index).Set(reflect.ValueOf(NewEloquentModel(realModels.Addr().Interface(), exist)))
		}
	}
	return
}

const EloquentModelPivotFieldIndex = 5

// EloquentModel is the base model for all models
type EloquentModel struct {
	IsBooted      bool                   `json:"-"` //model is booted
	Origin        map[string]interface{} `json:"-"` //store original attributes that get from the database or default attributes map[model.FiledName]value
	Changes       map[string]interface{} `json:"-"` //store changed attribute after save to database map[model.FiledName]value
	ModelPointer  reflect.Value          `json:"-"` //model pointer points to the model which hold this e.g. reflect.ValueOf(&User{})
	Pivot         map[string]interface{} `json:"-"` //pivot relation table attributes map[Relation.Field]value //TODO: field type mapping
	Exists        bool                   `json:"-"` //indicate whether the model is get from database or newly created and not store to the database yet
	Related       reflect.Value          `json:"-"` //when call save/create on relationship ,this holds the related key
	Muted         string                 `json:"-"` //muted events, comma separated
	OnlyColumns   map[string]interface{} `json:"-"` //only update/save these columns
	ExceptColumns map[string]interface{} `json:"-"` //exclude update/save there columns
	Tx            *Transaction           `json:"-"` //use same transaction
	Context       context.Context        `json:"-"`
}

/*
Init set modelPointer after initialized
*/
func Init(modelPointer interface{}) {
	parsed := GetParsedModel(modelPointer)
	isBooted := reflect.Indirect(reflect.ValueOf(modelPointer)).Field(parsed.EloquentModelFieldIndex).Elem().Field(0).Interface().(bool)
	if !isBooted {
		InitModel(modelPointer)
	}
}
func InitModelInTx(modelPointer interface{}, tx *Transaction, exists ...bool) *EloquentModel {
	e := InitModel(modelPointer, exists...)
	e.Tx = tx
	return e
}

func InitModel(modelPointer interface{}, exists ...bool) *EloquentModel {
	m := reflect.Indirect(reflect.ValueOf(modelPointer))
	parsed := GetParsedModel(modelPointer)
	e := NewEloquentModel(modelPointer, exists...)
	m.Field(parsed.EloquentModelFieldIndex).Set(reflect.ValueOf(e))
	return e
}
func NewEloquentModel(modelPointer interface{}, exists ...bool) *EloquentModel {
	m := EloquentModel{
		Origin: make(map[string]interface{}),
	}
	m.ModelPointer = reflect.ValueOf(modelPointer)
	m.Origin = make(map[string]interface{}, 4)
	m.Changes = make(map[string]interface{}, 4)
	m.Pivot = make(map[string]interface{})
	if len(exists) == 0 {
		exists = []bool{false}
	}
	m.Exists = exists[0]
	m.IsBooted = true
	m.SyncOrigin()
	return &m
}

/*
Fill the model with a map of attributes.
*/
func Fill(target interface{}, values ...map[string]interface{}) error {
	if eloquent, ok := target.(*EloquentModel); ok {
		for _, value := range values {
			eloquent.Fill(value)
		}
		return nil
	}
	parsed := GetParsedModel(target)
	if !parsed.IsEloquent {
		return errors.New(fmt.Sprintf("target: %s is not eloquent model", parsed.Name))
	}
	v := reflect.Indirect(reflect.ValueOf(target))
	if eloquent, ok := v.Field(parsed.EloquentModelFieldIndex).Interface().(*EloquentModel); !ok {
		return errors.New(fmt.Sprintf("target: %s is not eloquent model", parsed.Name))
	} else {
		for _, value := range values {
			eloquent.Fill(value)
		}
		return nil
	}

}
func (m *EloquentModel) WithContent(ctx context.Context) {
	m.Context = ctx
}
func (m *EloquentModel) BootIfNotBooted() {
	if !m.IsBooted {
		//m.FireModelEvent(EventBooting,nil)
		m.Booting()
		m.Boot()
		m.Booted()
		//m.FireModelEvent(EventBooted,nil)
	}
}
func (m *EloquentModel) Booting() {

}
func (m *EloquentModel) Boot() {

}
func (m *EloquentModel) Booted() {

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
	return keyValue.Interface() != m.Origin[key]
}

/*
SyncOrigin Sync the original attributes with the current.
*/
func (m *EloquentModel) SyncOrigin() {
	parsed := GetParsedModel(reflect.Indirect(m.ModelPointer).Type())
	model := reflect.Indirect(m.ModelPointer)
	for _, field := range parsed.FieldsByDbName {
		m.Origin[field.StructField.Name] = model.Field(field.StructField.Index[0]).Interface()
	}
}

/*
Save save the model to the database
*/
func (m *EloquentModel) Save() (res Result, err error) {
	parsed := GetParsedModel(reflect.Indirect(m.ModelPointer).Type())
	builder := DB.Model(parsed.ModelType)
	if eventErr := m.FireModelEvent(EventSaving, builder); eventErr != nil {
		return Result{Error: eventErr}, eventErr
	}
	if !m.Exists {
		if eventErr := m.FireModelEvent(EventCreating, builder); eventErr != nil {
			return Result{Error: eventErr}, eventErr
		}
		//TODO:SetDefaults
		res, err = builder.Insert(m.GetAttributesForCreate())
		if err != nil {
			return
		}
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
		if eventErr := m.FireModelEvent(EventCreated, builder); eventErr != nil {
			return Result{Error: eventErr}, eventErr
		}
	} else {
		if eventErr := m.FireModelEvent(EventUpdating, builder); eventErr != nil {
			return Result{Error: eventErr}, eventErr
		}
		m.Changes = m.GetDirty()
		res, err = builder.Where(parsed.PrimaryKey.ColumnName, reflect.Indirect(m.ModelPointer).Field(parsed.PrimaryKey.Index).Interface()).Update(m.GetAttributesForUpdate())
		if eventErr := m.FireModelEvent(EventUpdated, builder); eventErr != nil {
			return Result{Error: eventErr}, eventErr
		}
	}
	//FIXME: call save again in saved callback will cause primarykey dirty
	if eventErr := m.FireModelEvent(EventSaved, builder); eventErr != nil {
		return Result{Error: eventErr}, eventErr
	}
	m.SyncOrigin()
	return

}

/*
Create Save a new model and return the instance.
*/
func (m *EloquentModel) Create() (Result, error) {
	return m.Save()
}

/*
Delete Delete the model from the database.
*/
func (m *EloquentModel) Delete() (res Result, err error) {
	parsed := GetParsedModel(reflect.Indirect(m.ModelPointer).Type())
	b := DB.Model(parsed.ModelType)
	b.Where(parsed.PrimaryKey.ColumnName, m.ModelPointer.Elem().Field(parsed.PrimaryKey.Index).Interface())
	if eventErr := m.FireModelEvent(EventDeleteing, b); eventErr != nil {
		return Result{Error: eventErr}, eventErr
	}
	res, err = b.Delete()
	if eventErr := m.FireModelEvent(EventDeleted, b); eventErr != nil {
		return Result{Error: eventErr}, eventErr
	}
	return
}
func (m *EloquentModel) Mute(events ...string) *EloquentModel {
	for i := 0; i < len(events); i++ {
		if events[i] == EventALL {
			m.Muted = "Creating,Created,Updating,Updated,Saving,Saved,Deleting,Deleted"
			break
		}
		m.Muted = m.Muted + "," + events[i]
	}
	return m
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
		// TODO should insert all fields include zero value or just no-zero value?withzero/withoutzero?
		if !model.Field(keyIndex).IsZero() && keyIndex != modelType.PrimaryKey.Index {
			attrs[key] = model.Field(keyIndex).Interface()
		}
	}
	if modelType.UpdatedAt != "" {
		//if user set it manually,we won't change it
		//if !m.IsDirty(modelType.FieldsByDbName[modelType.UpdatedAt].Name) {
		switch modelType.FieldsByDbName[modelType.UpdatedAt].FieldType.Name() {
		case "NullTime":
			attrs[modelType.UpdatedAt] = sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			}
			m.Fill(map[string]interface{}{
				modelType.UpdatedAt: attrs[modelType.UpdatedAt],
			})
		case "Time":
			attrs[modelType.UpdatedAt] = time.Now()
			m.Fill(map[string]interface{}{
				modelType.UpdatedAt: time.Now(),
			})
		}
		//}
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
		//TODO: should update all fields or just dirty value?

		//TODO: scanner/valuer pointer
		if !model.Field(keyIndex).IsZero() {
			attrs[key] = model.Field(keyIndex).Interface()
		}
	}
	if modelType.CreatedAt != "" {
		//if user set it manually,we won't change it
		//if _, ok := attrs[modelType.CreatedAt]; !ok {
		switch modelType.FieldsByDbName[modelType.CreatedAt].FieldType.Name() {
		case "NullTime":
			attrs[modelType.CreatedAt] = sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			}
			m.Fill(map[string]interface{}{
				modelType.CreatedAt: attrs[modelType.CreatedAt],
			})
		case "Time":
			attrs[modelType.CreatedAt] = time.Now()
			m.Fill(map[string]interface{}{
				modelType.CreatedAt: time.Now(),
			})
		}
		//}
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

func (m *EloquentModel) FireModelEvent(eventName string, b *EloquentBuilder) error {
	reverted := m.ModelPointer.Interface()
	switch eventName {
	case EventSaving:
		if model, ok := reverted.(ISaving); ok && !strings.Contains(m.Muted, EventSaving) {
			return model.EloquentSaving()
		}
	case EventSaved:
		if model, ok := reverted.(ISaved); ok && !strings.Contains(m.Muted, EventSaved) {
			return model.EloquentSaved()
		}
	case EventCreating:
		if model, ok := reverted.(ICreating); ok && !strings.Contains(m.Muted, EventCreating) {
			return model.EloquentCreating()
		}
	case EventCreated:
		if model, ok := reverted.(ICreated); ok && !strings.Contains(m.Muted, EventCreated) {
			return model.EloquentCreated()
		}
	case EventUpdating:
		if model, ok := reverted.(IUpdating); ok && !strings.Contains(m.Muted, EventUpdating) {
			return model.EloquentUpdating()
		}
	case EventUpdated:
		if model, ok := reverted.(IUpdated); ok && !strings.Contains(m.Muted, EventUpdated) {
			return model.EloquentUpdated()
		}
	case EventDeleteing:
		if model, ok := reverted.(IDeleting); ok && !strings.Contains(m.Muted, EventDeleteing) {
			return model.EloquentDeleting()
		}
	case EventDeleted:
		if model, ok := reverted.(IDeleted); ok && !strings.Contains(m.Muted, EventDeleted) {
			return model.EloquentDeleted()
		}
	}
	return nil

}

/*
Fill fill the model with a map of attributes.
*/
func (m *EloquentModel) Fill(attrs map[string]interface{}) *EloquentModel {
	if !m.IsBooted {
		panic("model not inited yet,call Init first")
	}
	model := reflect.Indirect(m.ModelPointer)
	modelType := GetParsedModel(model.Type())
	for k, v := range attrs {
		//TODO check guard or fillable
		if f, ok := modelType.FieldsByDbName[k]; ok {
			model.Field(f.Index).Set(reflect.ValueOf(v))
		} else if f, ok := modelType.FieldsByStructName[k]; ok {
			model.Field(f.Index).Set(reflect.ValueOf(v))
		}
	}
	return m
}
func (m *EloquentModel) QualifyColumn(column string) string {
	if strings.Contains(column, ".") {
		return column
	}
	return GetParsedModel(m.ModelPointer.Type()).Table + "." + column
}

/*
NewInstance Create a new instance of the given model.
*/
func (m *EloquentModel) NewInstance(attributes map[string]interface{}, exist bool) interface{} {
	modelPointer := reflect.New(m.ModelPointer.Elem().Type())
	eloquentModel := NewEloquentModel(modelPointer.Interface(), exist)
	eloquentModel.Fill(attributes)
	return modelPointer.Interface()
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

	var b *EloquentBuilder
	b = NewEloquentBuilder(m.ModelPointer)

	b.With(relations...)
	b.Dest = m.ModelPointer.Interface()

	b.EagerLoadRelations(b.Dest)
}
