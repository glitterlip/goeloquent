package goeloquent

import (
	"context"
	"database/sql"
	"database/sql/driver"
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
				model.Field(parsed.EloquentModelFieldIndex).Set(newModel)
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
			model.Field(parsed.EloquentModelFieldIndex).Set(newModel)
		}
	}
	return
}

const EloquentModelPivotFieldIndex = 5
const EloquentModelContextFieldIndex = 11

// EloquentModel is the base model for all models
type EloquentModel struct {
	IsBooted           bool                   `json:"-"` //model is booted
	Origin             map[string]interface{} `json:"-"` //store original attributes that get from the database or default attributes map[model.FiledName]value
	Changes            map[string]interface{} `json:"-"` //store changed attribute after save to database map[model.FiledName]value
	ModelPointer       reflect.Value          `json:"-"` //model pointer points to the model which hold this e.g. reflect.ValueOf(&User{})
	Pivot              map[string]interface{} `json:"-"` //pivot relation table attributes map[Relation.Field]value //TODO: field type mapping
	Exists             bool                   `json:"-"` //indicate whether the model is get from database or newly created and not store to the database yet
	Related            reflect.Value          `json:"-"` //when call save/create on relationship ,this holds the related key
	Muted              string                 `json:"-"` //muted events, comma separated
	OnlyColumns        map[string]interface{} `json:"-"` //only update/save these columns
	ExceptColumns      map[string]interface{} `json:"-"` //exclude update/save there columns
	Tx                 *Transaction           `json:"-"` //use same transaction
	Context            context.Context        `json:"-"`
	WasRecentlyCreated bool                   `json:"-"`
	//RelationLoaded     map[string]interface{} //todo consider add a map of loaded relation Result to make debug much easier
}

/*
Init set modelPointer after initialized
*/
func Init(modelPointer interface{}) {
	config := GetParsedModel(modelPointer)
	model := reflect.Indirect(reflect.ValueOf(modelPointer))
	elo := model.Field(config.EloquentModelFieldIndex).Elem()
	isBooted := elo.IsValid() && elo.Field(0).Interface().(bool)
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
	if len(parsed.DefaultAttributes) > 0 {
		e.Fill(parsed.DefaultAttributes, true)
	}
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
func (m *EloquentModel) WithContext(ctx context.Context) *EloquentModel {
	m.Context = ctx
	return m
}

//	func (m *EloquentModel) IsEager() bool {
//		return !m.Exists
//	}
func (m *EloquentModel) IsDirty(structFieldName string) bool {
	current := reflect.Indirect(m.ModelPointer)
	parsed := GetParsedModel(reflect.Indirect(m.ModelPointer).Type())
	fieldIndex := parsed.FieldsByStructName[structFieldName].Index
	keyValue := current.Field(fieldIndex)
	if !keyValue.IsValid() || keyValue.IsZero() {
		return m.Origin[structFieldName] == nil
	}
	return keyValue.Interface() != m.Origin[structFieldName]
}

/*
SyncOrigin Sync the original attributes with the current.

u := &User{}
u.Name = "asd"
u.Email = "asd@zz.cc"
u.Only("Name").Save(&u)
u.IsDirty() = false
*/
func (m *EloquentModel) SyncOrigin() {
	parsed := GetParsedModel(reflect.Indirect(m.ModelPointer).Type())
	model := reflect.Indirect(m.ModelPointer)
	for _, field := range parsed.FieldsByDbName {
		m.Origin[field.Name] = model.Field(field.Index).Interface()
	}
	//reset after create/update
	m.OnlyColumns = nil
	m.ExceptColumns = nil
}

/*
Save save the model to the database
*/
func (m *EloquentModel) Save(ps ...interface{}) (res Result, err error) {
	if len(ps) > 0 && reflect.ValueOf(m).IsNil() {
		parsed := GetParsedModel(reflect.Indirect(reflect.ValueOf(ps[0])).Type())
		e := NewEloquentModel(ps[0])
		reflect.ValueOf(ps[0]).Elem().Field(parsed.EloquentModelFieldIndex).Set(reflect.ValueOf(e))
		return e.Save()
	}

	if reflect.ValueOf(m).IsNil() {
		panic("call Init(&model) first,or set modelPointer by call Save(&model)")
	}
	var saved map[string]interface{}
	parsed := GetParsedModel(reflect.Indirect(m.ModelPointer).Type())
	builder := DB.Model(parsed).WithContext(m.Context)
	if eventErr := m.FireModelEvent(EventSaving, builder); eventErr != nil {
		return Result{Error: eventErr}, eventErr
	}
	if !m.Exists {
		if eventErr := m.FireModelEvent(EventCreating, builder); eventErr != nil {
			return Result{Error: eventErr}, eventErr
		}
		//TODO:SetDefaults
		saved = m.GetAttributesForCreate()
		res, err = builder.Insert(saved)
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
		m.Changes = m.GetDirty()
		m.WasRecentlyCreated = true
		if eventErr := m.FireModelEvent(EventCreated, builder); eventErr != nil {
			return Result{Error: eventErr}, eventErr
		}
	} else {
		if eventErr := m.FireModelEvent(EventUpdating, builder); eventErr != nil {
			return Result{Error: eventErr}, eventErr
		}
		m.Changes = m.GetDirty()
		saved = m.GetAttributesForUpdate()
		res, err = builder.Where(parsed.PrimaryKey.ColumnName, reflect.Indirect(m.ModelPointer).Field(parsed.PrimaryKey.Index).Interface()).Update(saved)
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
	b := DB.Model(parsed.ModelType).WithContext(m.Context)
	b.Where(parsed.PrimaryKey.ColumnName, m.ModelPointer.Elem().Field(parsed.PrimaryKey.Index).Interface())
	if eventErr := m.FireModelEvent(EventDeleteing, b); eventErr != nil {
		return Result{Error: eventErr}, eventErr
	}
	if parsed.SoftDelete {
		b.Update(map[string]interface{}{
			parsed.DeletedAt: sql.NullTime{Time: time.Now(), Valid: true},
		})
	} else {
		res, err = b.Delete()
	}
	if eventErr := m.FireModelEvent(EventDeleted, b); eventErr != nil {
		return Result{Error: eventErr}, eventErr
	}
	return
}
func (m *EloquentModel) Mute(events ...string) *EloquentModel {
	for i := 0; i < len(events); i++ {
		if events[i] == EventALL {
			m.Muted = strings.Join([]string{EventSaving, EventSaved, EventCreating, EventCreated, EventUpdating, EventUpdated, EventDeleteing, EventDeleted, EventRetrieved, EventRetrieving}, ",")
			break
		}
		m.Muted = m.Muted + "," + events[i]
	}
	return m
}

/*
GetAttributesForUpdate Get the attributes that should be updated.
*/
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
	for columnName, field := range modelType.FieldsByDbName {
		keyIndex := field.Index
		if hasOnlyColumns {
			if _, ok := m.OnlyColumns[columnName]; !ok {
				continue
			}
		}
		if hasExceptColumns {
			if _, ok := m.ExceptColumns[columnName]; ok {
				continue
			}
		}

		// TODO should insert all fields include zero value or just no-zero value?withzero/withoutzero?
		if !model.Field(keyIndex).IsZero() {
			v := model.Field(keyIndex).Interface()
			if value, ok := v.(driver.Valuer); ok {
				v, _ = value.Value()
			}
			attrs[columnName] = v
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
	for columnName, field := range modelType.FieldsByDbName {
		keyIndex := field.Index

		if hasOnlyColumns {
			if _, ok := m.OnlyColumns[columnName]; !ok {
				continue
			}
		}

		if hasExceptColumns {
			if _, ok := m.ExceptColumns[columnName]; ok {
				continue
			}
		}
		//TODO: should update all fields or just dirty value?

		if !model.Field(keyIndex).IsZero() {
			v := model.Field(keyIndex).Interface()

			if value, ok := v.(driver.Valuer); ok {
				v, _ = value.Value()
			}

			attrs[columnName] = v
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

/*
GetAttributes Get the model's attributes. database column name=>value
*/
func (m *EloquentModel) GetAttributes() (attrs map[string]interface{}) {

	model := reflect.Indirect(m.ModelPointer)
	attrs = make(map[string]interface{})
	modelType := GetParsedModel(model.Type())
	for columnName, field := range modelType.FieldsByDbName {
		keyIndex := field.Index
		attrs[columnName] = model.Field(keyIndex).Interface()
	}
	return
}
func (m *EloquentModel) GetModel() interface{} {
	return m.ModelPointer.Interface()
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
	case EventRetrieved:
		if model, ok := reverted.(IRetrieved); ok && !strings.Contains(m.Muted, EventRetrieved) {
			return model.EloquentRetrieved()
		}
	case EventRetrieving:
		if model, ok := reverted.(IRetrieving); ok && !strings.Contains(m.Muted, EventRetrieving) {
			return model.EloquentRetrieving()
		}
	}
	return nil

}

/*
Fill fill the model with a map of attributes.
 1. user.Fill(map[string]interface{}) key can be struct field name / db column name
    fill user with map[string]interface{}
 2. user.Fill(map[string]interface{},true)
    fill user with map[string]interface{},force fill ignore guard or fillable
 3. user.Fill(map[string]interface{},false,&user)
    fill user with map[string]interface{},honor fillable or guard and init with user
*/
func (m *EloquentModel) Fill(attrs map[string]interface{}, ps ...interface{}) *EloquentModel {
	force := false
	if len(ps) > 1 {
		parsed := GetParsedModel(reflect.Indirect(reflect.ValueOf(ps[1])).Type())
		e := NewEloquentModel(ps[1])
		reflect.ValueOf(ps[1]).Elem().Field(parsed.EloquentModelFieldIndex).Set(reflect.ValueOf(e))
		return e.Fill(attrs, ps[0])
	}
	if len(ps) > 0 {
		force = ps[0].(bool)
	}
	if reflect.ValueOf(m).IsNil() || !m.IsBooted {
		panic("model not inited yet,call Init first")
	}
	model := reflect.Indirect(m.ModelPointer)
	config := GetParsedModel(model.Type())
	for k, v := range attrs {
		if _, ok := config.Fillables[k]; !ok && len(config.Fillables) > 0 && !force {
			continue
		}
		if _, ok := config.Guards[k]; ok && len(config.Guards) > 0 && !force {
			continue
		}
		if f, ok := config.FieldsByDbName[k]; ok {
			model.Field(f.Index).Set(reflect.ValueOf(v))
		} else if f, ok := config.FieldsByStructName[k]; ok {
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
