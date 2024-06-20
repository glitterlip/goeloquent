package eloquent

import (
	"fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/glitterlip/goeloquent/query"
	"reflect"
	"strings"
)

type ScopeFunc = func(builder *Builder) *Builder

type Builder struct {
	*query.Builder
	BaseModel     *Model
	EagerLoad     map[string]func(builder *Builder) *Builder
	Scopes        []ScopeFunc
	RemovedScopes []string
	Pivots        []string
	PivotWheres   []query.Where

	BeforeQueryCallBacks []func(*Builder) *Builder
	AfterQueryCallBacks  []func(*Builder) *Builder
}

func NewBuilder(base *query.Builder) *Builder {
	return &Builder{
		Builder: base,
	}
}
func NewEloquentBuilder(model interface{}, conn ...string) *Builder {
	if len(conn) > 0 {
		return NewBuilder(goeloquent.NewBuilder(goeloquent.Eloquent.Connection(conn[0]))).SetModel(model)
	}
	return NewBuilder(goeloquent.NewBuilder(goeloquent.Eloquent.Connection(goeloquent.DefaultConnectionName))).SetModel(model)
}
func (b *Builder) With(relations ...interface{}) *Builder {
	var name, relationName string
	var res = make(map[string]func(*Builder) *Builder)
	for _, relation := range relations {
		switch relation.(type) {
		case string:
			name = relation.(string)
			if pos := strings.Index(name, ":"); pos != -1 {
				relationName = name[0:pos]
			} else {
				relationName = name
			}
			res = b.addNestedWiths(relationName, res)
			res[relationName] = func(builder *Builder) *Builder {
				if pos := strings.Index(name, ":"); pos != -1 {
					builder.Select(strings.Split(name[pos+1:], ","))
				}
				return builder
			}
		case []string:
			for _, r := range relation.([]string) {
				name = r
				if pos := strings.Index(name, ":"); pos != -1 {
					relationName = name[0:pos]
				} else {
					relationName = name
				}
				res = b.addNestedWiths(relationName, res)
				res[relationName] = func(builder *Builder) *Builder {
					if pos := strings.Index(name, ":"); pos != -1 {
						builder.Select(strings.Split(name[pos+1:], ","))
					}
					return builder
				}
			}
		case map[string]func(builder *Builder) *Builder:
			for relationName, fn := range relation.(map[string]func(builder *Builder) *Builder) {
				name = relationName
				res = b.addNestedWiths(name, res)
				res[relationName] = fn
			}
		}
	}

	for s, f := range res {
		b.EagerLoad[s] = f
	}
	return b
}
func (b *Builder) addNestedWiths(name string, results map[string]func(builder *Builder) *Builder) map[string]func(builder *Builder) *Builder {
	var progress []string
	for _, segment := range strings.Split(name, ".") {
		progress = append(progress, segment)
		var ts string
		for j := 0; j < len(progress); j++ {
			ts = strings.Join(progress, ".")
		}
		if _, ok := results[ts]; !ok {
			results[ts] = func(builder *Builder) *Builder {
				return builder
			}
		}
	}
	return results
}

// TODO withGlobalScope
// TODO withoutGlobalScope
// TODO withoutGlobalScopes
// TODO removedScopes
// TODO whereKey
// TODO whereKeyNot
// TODO where
func (b *Builder) Get(dest interface{}, columns ...interface{}) (result goeloquent.Result, err error) {
	if b.FromTable == nil {
		b.BaseModel = GetParsedModel(dest)
		b.From(b.BaseModel.Table)
		b.Connection = goeloquent.Eloquent.Connection(b.BaseModel.ConnectionName)
		b.Grammar.SetTablePrefix(goeloquent.Eloquent.Configs[b.BaseModel.ConnectionName].Prefix)

	}
	b.ApplyGlobalScores()

	result, err = b.Builder.Get(dest, columns)

	d := reflect.TypeOf(b.Dest).Elem()
	if d.Kind() == reflect.Slice {
		d = d.Elem()
	}
	if result != nil && b.BaseModel != nil && b.BaseModel.IsEloquent && d.Kind() == reflect.Struct {
		c, _ := result.RowsAffected()
		BatchSync(b.Dest, c > 0)
	}

	if len(b.EagerLoad) == 0 {
		if len(b.Pivots) > 0 {
			WithPivots(b, b.Joins[0].JoinTable.(string), b.Pivots)
		}
		if len(b.PivotWheres) > 0 {
			WithPivots(b, b.Joins[0].JoinTable.(string), b.PivotWheres)
		}
	}
	if err == nil && len(b.EagerLoad) > 0 && result.Count > 0 {

		b.EagerLoadRelations(dest)
	}
	return
}
func (b *Builder) Find() {

}
func (b *Builder) First() {

}

/*
FirstOrNew get the first record matching the attributes or instantiate it.
*/
func (b *Builder) FirstOrNew(target interface{}, conditions ...map[string]interface{}) (found bool, err error) {
	res, err := b.SetModel(target).Where(conditions[0]).First(target)
	c, err := res.RowsAffected()
	found = c == 1
	if !found {
		err = Fill(target, conditions...)
		if err != nil {
			return
		}
		return false, nil
	}
	return true, nil
}

/*
FirstOrCreate get the first record matching the attributes or create it.
*/
func (b *Builder) FirstOrCreate(target interface{}, conditions ...map[string]interface{}) (found bool, err error) {
	res, err := b.SetModel(target).Where(conditions[0]).First(target)
	c, err := res.RowsAffected()
	found = c == 1
	if !found {
		err = Fill(target, conditions...)
		if err != nil {
			return
		}
		_, err = goeloquent.Eloquent.Save(target)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func (b *Builder) EagerLoadRelations(models interface{}) {
	//dest := r.Builder.Dest
	//value := reflect.ValueOf(dest)
	//get pointer value
	//realDest := reflect.Indirect(value)

	var model *Model
	//with reflect.Value of makeslice
	if t, ok := models.(*reflect.Value); ok {
		sliceEleType := t.Type().Elem()
		model = GetParsedModel(sliceEleType.PkgPath() + "." + sliceEleType.Name())
	} else {
		model = GetParsedModel(models)
	}

	//models = realDest.Interface()
	for relationName, fn := range b.EagerLoad {
		if !strings.Contains(relationName, ".") {
			b.EagerLoadRelation(models, model, relationName, fn)
		}
	}
}

func (b *Builder) EagerLoadRelation(models interface{}, model *Model, relationName string, constraints func(*RelationBuilder) *RelationBuilder) {
	if pos := strings.Index(relationName, ":"); pos != -1 {
		relationName = relationName[0:pos]
	}
	if relationMethod, ok := model.Relations[relationName]; ok {
		var params []reflect.Value
		builderI := relationMethod.Call(params)[0].Interface()
		//FIXME: builder config lost cause log missing
		if builder, ok := builderI.(*RelationBuilder); ok {
			//load nested relations
			wanted := relationName + "."
			nestedRelation := make(map[string]func(*RelationBuilder) *RelationBuilder)
			for name, f := range r.Builder.EagerLoad {
				if strings.Index(name, wanted) == 0 {
					nestedRelation[strings.Replace(name, wanted, "", -1)] = f
				}
			}
			if len(nestedRelation) > 0 {
				builder.With(nestedRelation)
			}
			//make sure this line runs first so we clear previous wheres and then add dynamic constraints
			builder.Relation.AddEagerConstraints(models)
			builder.LoadPivotColumns(r.Builder.Pivots...)
			builder.LoadPivotWheres(r.Builder.PivotWheres...)
			builder.DataMapping = r.Builder.DataMapping
			//dynamic constraints
			builder = constraints(builder)
			//copy conn/tx to new builder
			builder.Connection = r.Connection
			builder.Tx = r.Tx
			relationResults := builder.GetEager(builder.Relation)
			r.Match(models, relationResults, builder.Relation, relationName)
		}
	} else {
		panic(fmt.Sprintf(" relation : %s for model: %s didn't return a relationbuilder ", relationName, model.Name))
	}
}
func (b *Builder) GetRelation(name string) RelationI {
	if relationMethod, ok := b.BaseModel.Relations[name]; ok {
		rs := relationMethod.Call([]reflect.Value{})
		return rs[0].Interface().(RelationI)
	}
	panic(fmt.Sprintf("relation %s not found in model:%s", name, b.BaseModel.Name))

}
func (b *Builder) Model(model interface{}) *Builder {
	return b.SetModel(model)
}

/*
SetModel set the model for the eloquent builder

	Model should be a struct or a pointer to a struct with an embedded *eloquent.EloquentModel/*go-eloquent.EloquentModel interface


	e.g. SetModel(&User{})
*/
func (b *Builder) SetModel(model interface{}) *Builder {
	if model != nil {
		b.BaseModel = GetParsedModel(model)
		b.From(b.BaseModel.Table)
	}
	if b.Connection == nil {
		b.Connection = goeloquent.Eloquent.Connection(b.BaseModel.ConnectionName)
	}
	return b

}
func (b *Builder) WherePivot(params ...interface{}) *Builder {

	column := params[0].(string)
	paramsLength := len(params)
	var operator string
	var value interface{}
	var boolean = query.BOOLEAN_AND
	switch paramsLength {
	case 2:
		operator = "="
		value = params[1]
	case 3:
		operator = params[1].(string)
		value = params[2]
	case 4:
		operator = params[1].(string)
		value = params[2]
		boolean = params[3].(string)
	}

	b.PivotWheres = append(b.PivotWheres, Where{
		Type:     CONDITION_TYPE_BASIC,
		Column:   column,
		Operator: operator,
		Value:    value,
		Boolean:  boolean,
	})
	return b
}
func (b *Builder) WithPivot(columns ...string) *Builder {
	b.Pivots = append(b.Pivots, columns...)
	return b
}
func (b *Builder) ApplyGlobalScores() {
	if b.BaseModel != nil && len(b.BaseModel.GlobalScopes) > 0 {
		for name, scopeFunc := range b.BaseModel.GlobalScopes {
			if _, removed := b.RemovedScopes[name]; !removed {
				b.callScope(scopeFunc)
			}
		}
	}
}
func (b *Builder) callScope(scope ScopeFunc) *Builder {
	originalWhereCount := len(b.Wheres)
	scope(b)
	if len(b.Wheres) > originalWhereCount {
		b.addNewWheresWithInGroup(originalWhereCount)
	}
	return b
}
func (b *Builder) ApplyScopes(scopes ...func(builder *Builder) *Builder) *Builder {
	for _, scope := range scopes {
		scope(b)
	}
	return b
}
func (b *Builder) WithOutGlobalScopes(names ...string) *Builder {
	if len(names) == 0 {
		//remove all
		for name, _ := range b.BaseModel.GlobalScopes {
			b.RemovedScopes[name] = struct{}{}
		}
	} else {
		for _, name := range names {
			b.RemovedScopes[name] = struct{}{}
		}
	}

	return b
}

/*
UpdateOrCreate Create or update a record matching the attributes, and fill it with values.
*/
func (b *Builder) UpdateOrCreate(target interface{}, conditions, values map[string]interface{}) (updated bool, err error) {
	rows, err := b.SetModel(target).Where(conditions).First(target)
	if err != nil {
		return
	}
	c, _ := rows.RowsAffected()
	if c == 1 {
		err = eloquent.Fill(target, values)
	} else {
		err = eloquent.Fill(target, conditions, values)
	}
	if err != nil {
		return
	}
	_, err = goeloquent.Eloquent.Save(target)
	if err != nil {
		return
	}
	return c == 1, nil

}
