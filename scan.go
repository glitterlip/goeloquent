package goeloquent

import (
	"database/sql"
	"errors"
	_ "fmt"
	"reflect"
	"strings"
)

type ScanResult struct {
	Count int
}

func (ScanResult) LastInsertId() (int64, error) {
	return 0, errors.New("no insert in select")
}

func (v ScanResult) RowsAffected() (int64, error) {
	return int64(v.Count), nil
}
func ScanAll(rows *sql.Rows, dest interface{}) (result ScanResult) {
	//*model *[]model *[]*model
	value := reflect.ValueOf(dest)
	//get pointer value could be a normal slice or a reflect.value created by reflect.MakeSlice
	realDest := reflect.Indirect(value)
	//&[]User &[]*User
	if realDest.Kind() == reflect.Slice {
		slice := value.Type().Elem()
		base := slice.Elem()
		itemIsPtr := base.Kind() == reflect.Ptr
		columns, _ := rows.Columns()
		model := GetParsedModel(dest)
		scanArgs := make([]interface{}, len(columns))
		var v, vp reflect.Value
		if itemIsPtr {
			vp = reflect.New(base.Elem())
			v = reflect.Indirect(vp)
		} else {
			vp = reflect.New(base)
			v = reflect.Indirect(vp)
		}
		var needProcessPivot bool
		var pivotColumnMap = make(map[string]int, 2)
		for rows.Next() {
			for i, column := range columns {
				if f, ok := model.FieldsByDbName[column]; ok {
					scanArgs[i] = v.Field(f.Index).Addr().Interface()
				} else {
					if strings.Contains(column, "goelo_pivot_") {
						if !needProcessPivot {
							needProcessPivot = true
						}
						pivotColumnMap[column] = i
						scanArgs[i] = new(sql.NullString)
					} else {
						scanArgs[i] = new(interface{})
					}
				}
			}
			err := rows.Scan(scanArgs...)
			if err != nil {
				panic(err.Error())
			}
			if needProcessPivot {
				t := make(map[string]sql.NullString, 2)
				for columnName, index := range pivotColumnMap {
					t[columnName] = *(scanArgs[index].(*sql.NullString))
				}
				v.Field(model.PivotFieldIndex[0]).Field(model.PivotFieldIndex[1]).Set(reflect.ValueOf(t))
			}
			if itemIsPtr {
				realDest.Set(reflect.Append(realDest, vp))
			} else {
				realDest.Set(reflect.Append(realDest, v))
			}
		}
	} else if destValue, ok := dest.(*reflect.Value); ok {
		//relation reflect results
		slice := destValue.Type()
		base := slice.Elem()
		//itemIsPtr := base.Kind() == reflect.Ptr
		columns, _ := rows.Columns()
		model := GetParsedModel(base)
		scanArgs := make([]interface{}, len(columns))
		vp := reflect.New(base)
		v := reflect.Indirect(vp)
		var needProcessPivot bool
		var pivotColumnMap = make(map[string]int, 2)
		for rows.Next() {
			result.Count++
			for i, column := range columns {
				if f, ok := model.FieldsByDbName[column]; ok {
					scanArgs[i] = v.Field(f.Index).Addr().Interface()
				} else if strings.Contains(column, "goelo_pivot_") {
					if !needProcessPivot {
						needProcessPivot = true
					}
					pivotColumnMap[column] = i
					scanArgs[i] = new(sql.NullString)
				} else {
					scanArgs[i] = new(interface{})
				}
			}
			err := rows.Scan(scanArgs...)
			if err != nil {
				panic(err.Error())
			}
			if needProcessPivot {
				t := make(map[string]sql.NullString, 2)
				for columnName, index := range pivotColumnMap {
					t[columnName] = *(scanArgs[index].(*sql.NullString))
				}
				v.Field(model.PivotFieldIndex[0]).Field(model.PivotFieldIndex[1]).Set(reflect.ValueOf(t))
			}
			*destValue = reflect.Append(*destValue, v)
		}
	} else if realDest.Kind() == reflect.Struct {
		columns, _ := rows.Columns()
		model := GetParsedModel(dest)
		scanArgs := make([]interface{}, len(columns))
		vp := reflect.New(realDest.Type())
		v := reflect.Indirect(vp)

		for rows.Next() {
			result.Count++
			for i, column := range columns {
				if f, ok := model.FieldsByDbName[column]; ok {
					scanArgs[i] = v.Field(f.Index).Addr().Interface()
				} else {
					scanArgs[i] = new(interface{})
				}
			}
			err := rows.Scan(scanArgs...)
			if err != nil {
				panic(err.Error())
			}
			if realDest.Kind() == reflect.Ptr {
				realDest.Set(vp)
			} else {
				realDest.Set(v)
			}
		}
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
