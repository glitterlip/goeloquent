package goeloquent

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
)

func ScanAll(rows *sql.Rows, dest interface{}, mapping map[string]interface{}) (result Result) {
	defer func() {
		if r := recover(); r != nil {
			result.Error = errors.New("scan error:" + r.(string))
		}
	}()
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr {
		result.Error = errors.New("result should be a pointer")
		return
	}
	realDest := reflect.Indirect(v)
	if realDest.Kind() == reflect.Slice {
		sliceType := realDest.Type()
		sliceItem := sliceType.Elem()
		if sliceItem.Kind() == reflect.Map {
			return scanMapSlice(rows, dest, mapping)
		} else if sliceItem.Kind() == reflect.Struct {
			return scanStructSlice(rows, dest, mapping)
		} else if sliceItem.Kind() == reflect.Ptr {
			if sliceItem.Elem().Kind() == reflect.Struct {
				return scanStructSlice(rows, dest, mapping)
			}
		} else {
			return scanValues(rows, dest)
		}
	} else if _, ok := dest.(*reflect.Value); ok {
		return scanRelations(rows, dest, mapping)
	} else if realDest.Kind() == reflect.Struct {
		return scanStruct(rows, dest, mapping)
	} else if realDest.Kind() == reflect.Map {
		return scanMap(rows, dest, mapping)
	} else {
		for rows.Next() {
			result.Count++
			err := rows.Scan(dest)
			if err != nil {
				panic(err.Error())
			}
		}
	}
	return
}

func scanMapSlice(rows *sql.Rows, dest interface{}, mapping map[string]interface{}) (result Result) {
	columns, _ := rows.Columns()
	realDest := reflect.Indirect(reflect.ValueOf(dest))
	for rows.Next() {
		scanArgs := make([]interface{}, len(columns))
		element := make(map[string]interface{})
		result.Count++
		for i, column := range columns {
			if _, ok := mapping[PivotAlias+column]; ok {
				if reflect.ValueOf(mapping[PivotAlias+column]).Kind() != reflect.Ptr {
					scanArgs[i] = reflect.New(reflect.TypeOf(mapping[PivotAlias+column])).Interface()
				} else {
					scanArgs[i] = mapping[PivotAlias+column]
				}
			} else {
				scanArgs[i] = new(interface{})
			}
		}
		err := rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}
		for i, column := range columns {
			element[column] = reflect.ValueOf(scanArgs[i]).Elem().Interface()
		}
		realDest.Set(reflect.Append(realDest, reflect.ValueOf(element)))
	}
	return
}

func scanStructSlice(rows *sql.Rows, dest interface{}, mapping map[string]interface{}) (result Result) {
	realDest := reflect.Indirect(reflect.ValueOf(dest))
	columns, _ := rows.Columns()
	slice := realDest.Type()
	sliceItem := slice.Elem()
	itemIsPtr := sliceItem.Kind() == reflect.Ptr
	model := GetParsedModel(dest)
	scanArgs := make([]interface{}, len(columns))

	var needProcessPivot bool
	var needProcessAggregate bool
	var pivotColumnMap = make(map[string]int, 2)
	var aggregateColumnMap = make(map[string]int, 2)
	for rows.Next() {
		result.Count++
		var v, vp reflect.Value
		if itemIsPtr {
			vp = reflect.New(sliceItem.Elem())
			v = reflect.Indirect(vp)
		} else {
			vp = reflect.New(sliceItem)
			v = reflect.Indirect(vp)
		}
		for i, column := range columns {
			if f, ok := model.FieldsByDbName[column]; ok {
				scanArgs[i] = v.Field(f.Index).Addr().Interface()
			} else if strings.Contains(column, PivotAlias) {
				//process  withpivot column
				needProcessPivot = true
				pivotColumnMap[column] = i
				//check if user defined a datetype mapping
				if t, ok := mapping[column]; ok {
					scanArgs[i] = reflect.New(reflect.TypeOf(t)).Interface()
				} else {
					scanArgs[i] = new(interface{})
				}
			} else if strings.Contains(column, OrmPivotAlias) {
				//process orm pivot keys as string
				needProcessPivot = true
				pivotColumnMap[column] = i
				var ts string
				scanArgs[i] = &ts
			} else if strings.Contains(column, OrmAggregateAlias) {
				//process orm pivot keys as string
				needProcessPivot = true
				pivotColumnMap[column] = i
				var ts float64
				scanArgs[i] = &ts
			} else {
				scanArgs[i] = new(interface{})
			}
		}
		err := rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}
		if needProcessPivot || needProcessAggregate {
			t := make(map[string]interface{}, 2)
			t1 := make(map[string]float64, 2)
			for columnName, index := range pivotColumnMap {
				if strings.Contains(columnName, OrmPivotAlias) {
					t[columnName] = *scanArgs[index].(*string)
				}
				if strings.Contains(columnName, PivotAlias) {
					t[strings.Replace(columnName, PivotAlias, "", 1)] = reflect.Indirect(reflect.ValueOf(scanArgs[index])).Interface()
				}
			}
			for columnName, index := range aggregateColumnMap {
				t1[columnName] = *scanArgs[index].(*float64)
			}
			eloquentPtr := reflect.New(reflect.TypeOf(EloquentModel{}))
			eloquentModel := reflect.Indirect(eloquentPtr)
			eloquentModel.Field(EloquentModelPivotFieldIndex).Set(reflect.ValueOf(t))
			eloquentModel.Field(EloquentModelAggregateFieldIndex).Set(reflect.ValueOf(t1))
			v.Field(model.EloquentModelFieldIndex).Set(eloquentPtr)
		}
		if itemIsPtr {
			realDest.Set(reflect.Append(realDest, vp))
		} else {
			realDest.Set(reflect.Append(realDest, v))
		}
	}
	return
}

func scanRelations(rows *sql.Rows, dest interface{}, mapping map[string]interface{}) (result Result) {

	columns, _ := rows.Columns()
	destValue := dest.(*reflect.Value)
	//relation reflect results
	slice := destValue.Type()
	sliceItem := slice.Elem()
	//itemIsPtr := base.Kind() == reflect.Ptr
	model := GetParsedModel(sliceItem)
	scanArgs := make([]interface{}, len(columns))
	vp := reflect.New(sliceItem)
	v := reflect.Indirect(vp)
	var needProcessPivot bool
	var needProcessAggregate bool
	var pivotColumnMap = make(map[string]int, 2)
	var aggregateColumnMap = make(map[string]int, 2)
	for rows.Next() {
		result.Count++
		for i, column := range columns {
			if f, ok := model.FieldsByDbName[column]; ok {
				scanArgs[i] = v.Field(f.Index).Addr().Interface()
			} else if strings.Contains(column, PivotAlias) {
				//process user's withpivot column
				needProcessPivot = true
				pivotColumnMap[column] = i
				//check if user defined a datetype mapping
				if t, ok := mapping[column]; ok {
					scanArgs[i] = reflect.New(reflect.TypeOf(t)).Interface()
				} else {
					scanArgs[i] = new(interface{})
				}
			} else if strings.Contains(column, OrmPivotAlias) {
				//process orm pivot keys as string
				needProcessPivot = true
				pivotColumnMap[column] = i
				var ts string
				scanArgs[i] = &ts
			} else if strings.Contains(column, OrmAggregateAlias) {
				//process orm pivot keys as string
				needProcessAggregate = true
				aggregateColumnMap[column] = i
				var ts float64
				scanArgs[i] = &ts
			} else {
				scanArgs[i] = new(interface{})
			}
		}
		err := rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}
		if needProcessPivot || needProcessAggregate {
			t := make(map[string]interface{}, 2)
			t1 := make(map[string]float64, 2)
			for columnName, index := range pivotColumnMap {
				if strings.Contains(columnName, OrmPivotAlias) {
					t[columnName] = *scanArgs[index].(*string)
					//t[strings.Replace(columnName, OrmPivotAlias, "", 1)] = *scanArgs[index].(*string)
				}
				if strings.Contains(columnName, PivotAlias) {
					t[strings.Replace(columnName, PivotAlias, "", 1)] = reflect.Indirect(reflect.ValueOf(scanArgs[index])).Interface()
				}
			}
			for columnName, index := range aggregateColumnMap {
				t1[columnName] = *scanArgs[index].(*float64)
			}

			eloquentPtr := reflect.New(reflect.TypeOf(EloquentModel{}))
			eloquentModel := reflect.Indirect(eloquentPtr)
			eloquentModel.Field(EloquentModelPivotFieldIndex).Set(reflect.ValueOf(t))
			eloquentModel.Field(EloquentModelAggregateFieldIndex).Set(reflect.ValueOf(t1))
			v.Field(model.EloquentModelFieldIndex).Set(eloquentPtr)
		}
		*destValue = reflect.Append(*destValue, v)
	}
	return
}

func scanStruct(rows *sql.Rows, dest interface{}, mapping map[string]interface{}) (result Result) {
	realDest := reflect.Indirect(reflect.ValueOf(dest))
	model := GetParsedModel(dest)
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	vp := reflect.New(realDest.Type())
	v := reflect.Indirect(vp)
	var needProcessPivot bool
	var needProcessAggregate bool
	var pivotColumnMap = make(map[string]int, 2)
	var aggregateColumnMap = make(map[string]int, 2)
	for rows.Next() {
		result.Count++
		for i, column := range columns {
			if f, ok := model.FieldsByDbName[column]; ok {
				if t, ok := mapping[column]; ok {
					scanArgs[i] = reflect.New(reflect.TypeOf(t)).Interface()
				} else {
					scanArgs[i] = v.Field(f.Index).Addr().Interface()
				}
			} else if strings.Contains(column, PivotAlias) {
				//process user's withpivot column
				needProcessPivot = true
				pivotColumnMap[column] = i
				//check if user defined a datetype mapping
				if t, ok := mapping[column]; ok {
					scanArgs[i] = reflect.New(reflect.TypeOf(t)).Interface()
				} else {
					scanArgs[i] = new(interface{})
				}
			} else if strings.Contains(column, OrmPivotAlias) {
				//process orm pivot keys as string
				needProcessPivot = true
				pivotColumnMap[column] = i
				var ts string
				scanArgs[i] = &ts
			} else if strings.Contains(column, OrmAggregateAlias) {
				//process orm pivot keys as string
				needProcessAggregate = true
				aggregateColumnMap[column] = i
				var ts float64
				scanArgs[i] = &ts
			} else {
				scanArgs[i] = new(interface{})
			}
		}
		err := rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}
		if needProcessPivot || needProcessAggregate {
			t := make(map[string]interface{}, 2)
			t1 := make(map[string]float64, 2)
			for columnName, index := range pivotColumnMap {
				if strings.Contains(columnName, OrmPivotAlias) {
					t[columnName] = *scanArgs[index].(*string)
				}
				if strings.Contains(columnName, PivotAlias) {
					t[strings.Replace(columnName, PivotAlias, "", 1)] = reflect.Indirect(reflect.ValueOf(scanArgs[index])).Interface()
				}
			}
			for columnName, index := range aggregateColumnMap {
				t1[strings.Replace(columnName, OrmAggregateAlias, "", 1)] = *scanArgs[index].(*float64)
			}

			eloquentPtr := reflect.New(reflect.TypeOf(EloquentModel{}))
			eloquentModel := reflect.Indirect(eloquentPtr)
			eloquentModel.Field(EloquentModelPivotFieldIndex).Set(reflect.ValueOf(t))
			eloquentModel.Field(EloquentModelAggregateFieldIndex).Set(reflect.ValueOf(t1))

			v.Field(model.EloquentModelFieldIndex).Set(eloquentPtr)
		}
		if realDest.Kind() == reflect.Ptr {
			realDest.Set(vp)
		} else {
			realDest.Set(v)
		}
	}
	return
}
func scanMap(rows *sql.Rows, dest interface{}, mapping map[string]interface{}) (result Result) {
	columns, _ := rows.Columns()
	realDest := reflect.Indirect(reflect.ValueOf(dest))
	scanArgs := make([]interface{}, len(columns))
	for rows.Next() {
		result.Count++
		for i, column := range columns {
			if _, ok := mapping[PivotAlias+column]; ok {
				if reflect.ValueOf(mapping[PivotAlias+column]).Kind() != reflect.Ptr {
					scanArgs[i] = reflect.New(reflect.TypeOf(mapping[PivotAlias+column])).Interface()
				} else {
					scanArgs[i] = mapping[PivotAlias+column]
				}
			} else {
				scanArgs[i] = new(interface{})
			}
		}
		err := rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}
		for i, column := range columns {
			//If elem is the zero Value, SetMapIndex deletes the key from the map.
			realDest.SetMapIndex(reflect.ValueOf(column), reflect.ValueOf(reflect.ValueOf(scanArgs[i]).Elem().Interface()))
		}
	}
	return
}
func scanValues(rows *sql.Rows, dest interface{}) (result Result) {
	columns, _ := rows.Columns()
	realDest := reflect.Indirect(reflect.ValueOf(dest))
	scanArgs := make([]interface{}, len(columns))
	for rows.Next() {
		result.Count++
		scanArgs[0] = reflect.New(realDest.Type().Elem()).Interface()
		err := rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}
		realDest.Set(reflect.Append(realDest, reflect.ValueOf(reflect.ValueOf(scanArgs[0]).Elem().Interface())))
	}
	return
}
