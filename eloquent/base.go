package eloquent

import (
	"fmt"
	"github.com/glitterlip/goeloquent"
	"reflect"
	"strings"
)

type ModelConfig Model

// Model prased model config
type Model struct {
	Name                       string //pkg + model name
	ModelType                  reflect.Type
	Table                      string                        //model db table name
	DbFields                   []string                      //model db table columns
	FieldsByDbName             map[string]*Field             //map of fields of database table columns, database column name=>Model.Field
	Fields                     []*Field                      //all struct fields
	FieldsByStructName         map[string]*Field             //map of struct fields, struct field name=>Model.Field
	Relations                  map[string]reflect.Value      //model field name => model getRelation method()
	ConnectionName             string                        //connection name
	ConnectionResolver         func(builder *Builder) string //connection resolve func
	PrimaryKey                 *Field
	PrimaryKeyAutoIncrementing bool //auto incrementing primary key
	IsEloquent                 bool //eloquent model or plain struct
	PivotFieldIndex            int  //pivot column index
	PivotsFieldIndex           int  //pivots column index
	EloquentModelFieldIndex    int  //*EloquentModel field index in model struct
	DefaultAttributes          map[string]interface{}
	DispatchesEvents           map[string]reflect.Value
	UpdatedAt                  string               //database update timestamp column name
	CreatedAt                  string               //database create timestamp column name
	DeletedAt                  string               //database delete timestamp column name
	SoftDelete                 bool                 //has soft delete
	GlobalScopes               map[string]ScopeFunc //registered global scopes
	WithRelations              []string             //eager load relations
	WithRelationCounts         []string             //eager load relation counts
	Guards                     []string             //guarded model fields
	Fillables                  []string             //fillable model fields

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

func Parse(modelType reflect.Type) (model *Model, err error) {
	//cache
	modelValue := reflect.New(modelType)
	var tableName string
	var connectionName string
	if t, ok := modelValue.Interface().(TableName); ok {
		tableName = t.TableName()
	} else {
		tableName = goeloquent.ToSnakeCase(modelType.Name())
	}
	if c, ok := modelValue.Interface().(ConnectionName); ok {
		connectionName = c.ConnectionName()
	} else {
		connectionName = "default"
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
		ConnectionName:     connectionName,
		PivotFieldIndex:    0,
		DefaultAttributes:  make(map[string]interface{}),
		//RelationTypes:      make(map[string]reflect.Type),
	}
	//TODO: check if modeltype is map panic
	for i := 0; i < modelType.NumField(); i++ {
		model.ParseField(modelType.Field(i))
	}
	if len(model.DbFields) == 0 {
		for i := 0; i < len(model.Fields); i++ {
			if model.Fields[i].Name == "Table" && strings.Contains(model.Fields[i].Tag.Get(TagName), "TableName:") {
				model.Table = strings.Replace(model.Fields[i].Tag.Get(TagName), "TableName:", "", 1)
			} else {
				tag := model.Fields[i].Tag.Get(TagName)
				if strings.Contains(tag, "primaryKey") {
					model.PrimaryKey = model.Fields[i]
				}
				var name string
				if t := model.Fields[i].Tag.Get("column"); len(t) > 0 {
					name = t
				} else {
					name = goeloquent.ToSnakeCase(model.Fields[i].Name)
				}
				model.Fields[i].ColumnName = name
				model.FieldsByDbName[name] = model.Fields[i]
				model.DbFields = append(model.DbFields, name)
			}

		}
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
		if ptrReciver.Method(i).Name == "EloquentGetDefaultAttributes" {
			//TODO set model DefaultAttributes
		}
		if ptrReciver.Method(i).Name == "EloquentAddEloquentGlobalScopes" {
			res := modelValue.MethodByName("AddGlobalScopes").Call([]reflect.Value{})
			model.GlobalScopes = res[0].Interface().(map[string]ScopeFunc)
		}
	}
	if model.PrimaryKey == nil {
		model.PrimaryKey = model.FieldsByDbName["id"]
	}

	return model, err
}

// TODO only process field with goelo tag
func (m *Model) ParseField(field reflect.StructField) *Field {
	modelField := &Field{
		Name:              field.Name,
		ColumnName:        field.Name,
		FieldType:         field.Type,
		IndirectFieldType: field.Type,
		StructField:       field,
		Tag:               field.Tag,
		TagSettings:       make(map[string]string),
		//Model:             m,
		Index: field.Index[0],
		//ReflectValueOf:    nil,
		//ValueOf:           nil,
	}
	tag, ok := field.Tag.Lookup(TagName)
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

				//todo case defaultvalue,case updateable,insertable
			}
		}
		//} else {
		//maybe only a few fileds has goelo tag
		//user need add a goelo tag for every struct filed that has a db column otherwise no cant decide which field to be created/updated
		//modelField.ColumnName = ToSnakeCase(field.Name)
		//m.FieldsByDbName[modelField.ColumnName] = modelField
		//m.DbFields = append(m.DbFields, modelField.ColumnName)
	}
	//if  { ignored tag
	m.Fields = append(m.Fields, modelField)
	m.FieldsByStructName[modelField.Name] = modelField
	if modelField.Name == EloquentName {
		m.IsEloquent = true
		m.EloquentModelFieldIndex = modelField.Index
		//embeded := reflect.TypeOf(EloquentModel{})
		//for i := 0; i < embeded.NumField(); i++ {
		//	if embeded.Field(i).Name == "Pivot" {
		//		m.PivotFieldIndex = append(m.PivotFieldIndex, i)
		//
		//	}
		//}
	}
	//}

	return modelField
}

// reflect.Type of model
// string pkgpath+name of model
// pointer of model
// pointer of slice of model
func GetParsedModel(model interface{}) *Model {
	var target reflect.Type

	if t, ok := model.(reflect.Type); ok {
		target = t
	} else if s, ok := model.(string); ok {
		target = goeloquent.GetRegisteredModel(s).Type()
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
	i, ok := goeloquent.GetParsed(name)
	if ok {
		return i.(*Model)
	}
	parsed, _ := Parse(target)
	goeloquent.Eloquent.ParsedModelsMap.Store(name, parsed)
	return parsed

}
