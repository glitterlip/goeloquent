package goeloquent

import (
	"fmt"
	"reflect"
	"strings"
)

type Driver string

const (
	DriverMysql               Driver = "mysql"
	GlobalScopeWithoutTrashed        = "WithoutTrashed"
	ColumnDeletedAt                  = "DELETED_AT"
	ColumnCreatedAt                  = "CREATED_AT"
	ColumnUpdatedAt                  = "UPDATED_AT"
	ColumnPrimaryKey                 = "primaryKey"
)

type DBConfig struct {
	Driver          Driver
	Name            string
	ReadHost        []string
	WriteHost       []string
	Host            string
	Port            string
	Database        string
	Username        string
	Password        string
	Charset         string
	Prefix          string
	ConnMaxLifetime int
	ConnMaxIdleTime int
	MaxIdleConns    int
	MaxOpenConns    int
	ParseTime       bool
	// 	mysql
	Collation       string
	UnixSocket      string
	MultiStatements bool
	Dsn             string
	DsnExtraString  string
	Strict          bool
	Mode            string
	IsolationLevel  string
	// pgsql
	Sslmode   string
	TLS       string
	EnableLog bool
	//perf
	//interpolateParams TODO
}

type ModelConfig Model

/*
Model prased model config
*/
type Model struct {
	Name                       string //pkg + model name
	ModelType                  reflect.Type
	Table                      string                                //model db table name
	TableResolver              func(builder *EloquentBuilder) string //model table name
	FieldsByDbName             map[string]*Field                     //map of fields of database table columns, database column name=>Model.Field
	FieldsByStructName         map[string]*Field                     //map of struct fields, struct field name=>Model.Field
	Relations                  map[string]reflect.Value              //model field name => model getRelation method()
	ConnectionName             string                                //connection name
	ConnectionResolver         func(builder *EloquentBuilder) string //connection resolve func
	PrimaryKey                 *Field
	PrimaryKeyAutoIncrementing bool //auto incrementing primary key
	IsEloquent                 bool //eloquent model or plain struct
	PivotFieldIndex            int  //pivot column index
	EloquentModelFieldIndex    int  //*EloquentModel field index in model struct
	DefaultAttributes          map[string]interface{}
	CallBacks                  map[string]reflect.Value //model callbacks
	UpdatedAt                  string                   //database update timestamp column name
	CreatedAt                  string                   //database create timestamp column name
	DeletedAt                  string                   //database delete timestamp column name
	SoftDelete                 bool                     //has soft delete
	GlobalScopes               map[string]ScopeFunc     //registered global scopes
	Guards                     map[string]struct{}      //guarded model fields when use Save(map[string]interface{})/Fill(map[string]interface{})
	Fillables                  map[string]struct{}      //fillable model fields when use Save(map[string]interface{})/Fill(map[string]interface{})
	EagerRelations             map[string]RelationFunc  //eager loaded relations
	EagerRelationCounts        map[string]RelationFunc  //eager loaded relation counts
}
type Field struct {
	Name              string       //reflect.StructField.Name,struct field name
	ColumnName        string       //db column name
	Index             int          //Type.FieldByIndex, struct reflect.type fields index
	FieldType         reflect.Type //reflect.StructField.Type
	IndirectFieldType reflect.Type
	Tag               reflect.StructTag
	TagSettings       map[string]string
}

/*
Parse Parse model to ModelConfig and cache it
*/
func Parse(modelType reflect.Type) (model *Model, err error) {
	modelValue := reflect.New(modelType)
	model = &Model{
		Name:                modelType.Name(),
		ModelType:           modelType,
		FieldsByDbName:      make(map[string]*Field),
		FieldsByStructName:  make(map[string]*Field),
		Relations:           make(map[string]reflect.Value),
		PrimaryKey:          nil,
		CallBacks:           make(map[string]reflect.Value),
		PivotFieldIndex:     0,
		DefaultAttributes:   make(map[string]interface{}),
		EagerRelations:      make(map[string]RelationFunc),
		EagerRelationCounts: make(map[string]RelationFunc),
		GlobalScopes:        make(map[string]ScopeFunc),
	}
	if t, ok := modelValue.Interface().(TableName); ok {
		model.Table = t.TableName()
	} else if t, ok := modelValue.Interface().(DynamicTableName); ok {
		model.TableResolver = t.ResolveTableName
	} else {
		model.Table = ToSnakeCase(modelType.Name())
	}
	if c, ok := modelValue.Interface().(ConnectionName); ok {
		model.ConnectionName = c.ConnectionName()
	} else if c, ok := modelValue.Interface().(DynamicConnectionName); ok {
		model.ConnectionResolver = c.ResolveConnectionName
	} else {
		model.ConnectionName = "default"
	}

	for i := 0; i < modelType.NumField(); i++ {
		if _, ok := modelType.Field(i).Tag.Lookup(EloquentTagName); ok || modelType.Field(i).Name == EloquentName {
			model.ParseField(modelType.Field(i))
		}
	}

	funcs := map[string]string{
		EventSaving:     EventSaving,
		EventSaved:      EventSaved,
		EventCreating:   EventCreating,
		EventCreated:    EventCreated,
		EventUpdating:   EventUpdating,
		EventUpdated:    EventUpdated,
		EventDeleteing:  EventDeleteing,
		EventDeleted:    EventDeleted,
		EventRetrieving: EventRetrieving,
		EventRetrieved:  EventRetrieved,
		EventBooting:    EventBooting,
		EventBoot:       EventBoot,
		EventBooted:     EventBooted,
	}
	ptrReciver := reflect.PtrTo(modelType)
	for i := 0; i < ptrReciver.NumMethod(); i++ {
		if name, ok := funcs[ptrReciver.Method(i).Name]; ok {
			model.CallBacks[name] = modelValue.Method(i)
		}

		if ptrReciver.Method(i).Name == EloquentAddGlobalScopes {
			res := modelValue.MethodByName(EloquentAddGlobalScopes).Call([]reflect.Value{})
			model.GlobalScopes = res[0].Interface().(map[string]ScopeFunc)
		}
		if ptrReciver.Method(i).Name == EloquentGetFillable {
			res := modelValue.MethodByName(EloquentGetFillable).Call([]reflect.Value{})
			model.Fillables = res[0].Interface().(map[string]struct{})
		}
		if ptrReciver.Method(i).Name == EloquentGetGuarded {
			if model.Fillables != nil && len(model.Fillables) > 0 {
				panic("can not use guarded with fillable")
			}
			res := modelValue.MethodByName(EloquentGetGuarded).Call([]reflect.Value{})
			model.Guards = res[0].Interface().(map[string]struct{})
		}
		if ptrReciver.Method(i).Name == EloquentGetWithRelations {
			res := modelValue.MethodByName(EloquentGetWithRelations).Call([]reflect.Value{})
			model.EagerRelations = res[0].Interface().(map[string]RelationFunc)
		}
		if ptrReciver.Method(i).Name == EloquentGetWithRelationCounts {
			res := modelValue.MethodByName(EloquentGetWithRelationCounts).Call([]reflect.Value{})
			model.EagerRelationCounts = res[0].Interface().(map[string]RelationFunc)
		}
		if ptrReciver.Method(i).Name == EloquentGetDefaultAttributes {
			res := modelValue.MethodByName(EloquentGetDefaultAttributes).Call([]reflect.Value{})
			model.DefaultAttributes = res[0].Interface().(map[string]interface{})
		}

	}
	if model.PrimaryKey == nil {
		model.PrimaryKey = model.FieldsByDbName["id"]
	}

	return model, err
}

/*
ParseField parse model field to Model.Field
*/
func (m *Model) ParseField(field reflect.StructField) *Field {
	modelField := &Field{
		Name:              field.Name,
		ColumnName:        field.Name,
		FieldType:         field.Type,
		IndirectFieldType: field.Type,
		Tag:               field.Tag,
		TagSettings:       make(map[string]string),
		Index:             field.Index[0],
	}
	tag, ok := field.Tag.Lookup(EloquentTagName)
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
				m.FieldsByStructName[modelField.Name] = modelField
			case ColumnPrimaryKey:
				m.PrimaryKey = modelField
			case ColumnCreatedAt:
				m.CreatedAt = modelField.ColumnName
			case ColumnUpdatedAt:
				m.UpdatedAt = modelField.ColumnName
			case ColumnDeletedAt:
				m.DeletedAt = modelField.ColumnName
				m.SoftDelete = true
				m.GlobalScopes[GlobalScopeWithoutTrashed] = func(builder *EloquentBuilder) *EloquentBuilder {
					return builder.WhereNull(m.DeletedAt)
				}

			case string(RelationHasMany),
				string(RelationHasOne),
				string(RelationBelongsToMany),
				string(RelationBelongsTo),
				string(RelationHasManyThrough),
				string(RelationHasOneThrough),
				string(RelationMorphByMany),
				string(RelationMorphOne),
				string(RelationMorphMany),
				string(RelationMorphTo),
				string(RelationMorphToMany):
				methodName := value
				v := reflect.New(m.ModelType)
				m.Relations[field.Name] = v.MethodByName(methodName)
				if !m.Relations[field.Name].IsValid() {
					panic(fmt.Sprintf("relation method %s on model %s not found", methodName, m.Name))
				}

			default:
				panic(fmt.Sprintf("unknown tag %s", key))
			}
		}

		m.FieldsByStructName[modelField.Name] = modelField
	}

	if modelField.Name == EloquentName {
		m.IsEloquent = true
		m.EloquentModelFieldIndex = modelField.Index
		m.PivotFieldIndex = EloquentModelPivotFieldIndex
	}

	return modelField
}
func GetRegisteredModel(name string) reflect.Value {
	v, ok := RegisteredModelsMap.Load(name)
	if !ok {
		panic(fmt.Sprintf("no registered model found for %s", name))
	}
	return v.(reflect.Value)
}
func GetParsed(name string) (interface{}, bool) {
	return ParsedModelsMap.Load(name)
}

/*
GetParsedModel get parsed ModelConfig from cache or parse it
 1. if model is a string, get model by name from cache
 2. if model is a reflect.Type, get model by name from cache
*/
func GetParsedModel(model interface{}) *Model {
	var target reflect.Type
	switch t := model.(type) {
	case *Model:
		return t
	case Model:
		return &t
	case reflect.Type:
		target = t
	case string:
		target = GetRegisteredModel(t).Type()
	default:
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

	name := target.PkgPath() + "." + target.Name()
	i, ok := GetParsed(name)
	if ok {
		return i.(*Model)
	}
	parsed, _ := Parse(target)
	ParsedModelsMap.Store(name, parsed)
	return parsed

}
