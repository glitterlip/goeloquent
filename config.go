package goeloquent

import (
	"fmt"
	"reflect"
	"strings"
)

type Driver string

const (
	DriverMysql Driver = "mysql"
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
	DbFields                   []string                              //model db table columns
	FieldsByDbName             map[string]*Field                     //map of fields of database table columns, database column name=>Model.Field
	Fields                     []*Field                              //all struct fields
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
	Guards                     map[string]struct{}      //guarded model fields
	Fillables                  map[string]struct{}      //fillable model fields
	EagerRelations             map[string]RelationFunc  //eager loaded relations
	EagerRelationCounts        map[string]RelationFunc  //eager loaded relation counts
}
type Field struct {
	Name              string       //reflect.StructField.Name,struct field name
	ColumnName        string       //db column name
	Index             int          //Type.FieldByIndex, struct reflect.type fields index
	FieldType         reflect.Type //reflect.StructField.Type
	IndirectFieldType reflect.Type
	StructField       reflect.StructField
	Tag               reflect.StructTag
	TagSettings       map[string]string
}

/*
Parse Parse model to ModelConfig and cache it
*/
func Parse(modelType reflect.Type) (model *Model, err error) {
	modelValue := reflect.New(modelType)
	var tableName string
	var connectionName string
	if t, ok := modelValue.Interface().(TableName); ok {
		tableName = t.TableName()
	} else {
		tableName = ToSnakeCase(modelType.Name())
	}
	if c, ok := modelValue.Interface().(ConnectionName); ok {
		connectionName = c.ConnectionName()
	} else {
		connectionName = "default"
	}
	model = &Model{
		Name:                modelType.Name(),
		ModelType:           modelType,
		Table:               tableName,
		FieldsByDbName:      make(map[string]*Field),
		FieldsByStructName:  make(map[string]*Field),
		Relations:           make(map[string]reflect.Value),
		PrimaryKey:          nil,
		CallBacks:           make(map[string]reflect.Value),
		ConnectionName:      connectionName,
		PivotFieldIndex:     0,
		DefaultAttributes:   make(map[string]interface{}),
		EagerRelations:      make(map[string]RelationFunc),
		EagerRelationCounts: make(map[string]RelationFunc),
		GlobalScopes:        make(map[string]ScopeFunc),
	}

	for i := 0; i < modelType.NumField(); i++ {
		model.ParseField(modelType.Field(i))
	}
	if len(model.DbFields) == 0 {
		for i := 0; i < len(model.Fields); i++ {
			if model.Fields[i].Name == "Table" && strings.Contains(model.Fields[i].Tag.Get(EloquentTagName), "TableName:") {
				model.Table = strings.Replace(model.Fields[i].Tag.Get(EloquentTagName), "TableName:", "", 1)
			} else {
				tag := model.Fields[i].Tag.Get(EloquentTagName)
				if strings.Contains(tag, "primaryKey") {
					model.PrimaryKey = model.Fields[i]
				}
				var name string
				if t := model.Fields[i].Tag.Get("column"); len(t) > 0 {
					name = t
				} else {
					name = ToSnakeCase(model.Fields[i].Name)
				}
				model.Fields[i].ColumnName = name
				model.FieldsByDbName[name] = model.Fields[i]
				model.DbFields = append(model.DbFields, name)
			}

		}
	}
	funcs := map[string]string{
		EventSaving:           EventSaving,
		EventSaved:            EventSaved,
		EventCreating:         EventCreating,
		EventCreated:          EventCreated,
		EventUpdating:         EventUpdating,
		EventUpdated:          EventUpdated,
		EventDeleteing:        EventDeleteing,
		EventDeleted:          EventDeleted,
		EventRetrieving:       EventRetrieving,
		EventRetrieved:        EventRetrieved,
		EventBooting:          EventBooting,
		EventBoot:             EventBoot,
		EventBooted:           EventBooted,
		EloquentGetConnection: EloquentGetConnection,
		EloquentGetTable:      EloquentGetTable,
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
		StructField:       field,
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

			}
		}

		m.Fields = append(m.Fields, modelField)
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

	if t, ok := model.(reflect.Type); ok {
		target = t
	} else if s, ok := model.(string); ok {
		target = GetRegisteredModel(s).Type()
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
	name := target.PkgPath() + "." + target.Name()
	i, ok := GetParsed(name)
	if ok {
		return i.(*Model)
	}
	parsed, _ := Parse(target)
	ParsedModelsMap.Store(name, parsed)
	return parsed

}
