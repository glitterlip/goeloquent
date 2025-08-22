package goeloquent

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

func PrepareScanValues(values []interface{}, columnTypes []*sql.ColumnType, columns []string) {

	if len(columnTypes) > 0 {
		for i, columnType := range columnTypes {
			if columnType.ScanType() != nil {
				values[i] = reflect.New(columnType.ScanType()).Interface()
			} else {
				values[i] = new(interface{})
			}
		}
	} else {
		for i := range columns {
			values[i] = new(interface{})
		}
	}

}
func Scan(rows *sql.Rows, dest interface{}, mapping map[string]string) (count int64, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case string:
				err = errors.New("failed to scan rows: " + r.(string))
			case error:
				err = errors.New("failed to scan rows: " + r.(error).Error())
			}
		}
	}()
	if rows == nil {
		return 0, nil
	}
	columns, _ := rows.Columns()
	columnTypes, _ := rows.ColumnTypes()
	if len(mapping) > 0 {
		for i, column := range columns {
			if c, ok := mapping[column]; ok {
				columns[i] = c
			}
		}
	}

	switch destT := dest.(type) {
	case map[string]interface{}, *map[string]interface{}:
		if rows.Next() {
			count++
			values := make([]interface{}, len(columns))
			PrepareScanValues(values, columnTypes, columns)
			if err = rows.Scan(values...); err != nil {
				return 0, err
			}
			mapValue, ok := dest.(map[string]interface{})
			if ok {
				err = ScanMap(mapValue, values, columns)
				if err != nil {
					return 0, err
				}
			} else {
				if mapPtr, ok := dest.(*map[string]interface{}); ok {
					if *mapPtr == nil {
						*mapPtr = map[string]interface{}{}
					}
					err = ScanMap(*mapPtr, values, columns)
					if err != nil {
						return 0, err
					}
				}
			}
		}
	case *[]map[string]interface{}:
		for rows.Next() {
			count++
			values := make([]interface{}, len(columns))
			PrepareScanValues(values, columnTypes, columns)
			if err = rows.Scan(values...); err != nil {
				return 0, err
			}
			mapValue := map[string]interface{}{}
			err = ScanMap(mapValue, values, columns)
			if err != nil {
				return 0, err
			}
			*destT = append(*destT, mapValue)
		}

	case *int, *int8, *int16, *int32, *int64,
		*uint, *uint8, *uint16, *uint32, *uint64, *uintptr,
		*float32, *float64,
		*bool, *string, *time.Time,
		*sql.NullInt32, *sql.NullInt64, *sql.NullFloat64,
		*sql.NullBool, *sql.NullString, *sql.NullTime, []byte, *[]byte:
		for rows.Next() {
			count++
			err = rows.Scan(dest)
			if err != nil {
				return 0, err
			}
		}
	default:
		destPtr := reflect.ValueOf(dest)
		if destPtr.Kind() != reflect.Ptr {
			return 0, ErrorNotPtr
		}
		reflectDestValue := reflect.Indirect(destPtr)
		reflectDestType := reflectDestValue.Type()

		switch reflectDestValue.Kind() {
		case reflect.Slice, reflect.Array:
			reflectEleType := reflectDestType.Elem()
			isArray := reflectDestValue.Kind() == reflect.Array
			eleIsPtr := reflectEleType.Kind() == reflect.Ptr
			if eleIsPtr {
				reflectEleType = reflectEleType.Elem()
			}
			for rows.Next() {
				var elem reflect.Value
				count++
				switch reflectEleType.Kind() {
				case reflect.Map, reflect.Struct:
					if reflectEleType.Kind() == reflect.Map {
						values := make([]interface{}, len(columns))
						PrepareScanValues(values, columnTypes, columns)
						if err = rows.Scan(values...); err != nil {
							return count, err
						}
						mapValue := map[string]interface{}{}
						err = ScanMap(mapValue, values, columns)
						if err != nil {
							return 0, err
						}
						if eleIsPtr {
							elem = reflect.New(reflectEleType)
							elem.Elem().Set(reflect.ValueOf(mapValue))
						} else {
							elem = reflect.ValueOf(mapValue)
						}
					} else {
						elem = reflect.New(reflectEleType)
						ScanStruct(rows, elem, columns, columnTypes)
						if !eleIsPtr {
							elem = elem.Elem()
						}
					}
				default:
					elem = reflect.New(reflectEleType)
					err = rows.Scan(elem.Interface())
					if err != nil {
						return 0, err
					}
					if !eleIsPtr {
						elem = elem.Elem()
					}
				}
				if isArray {
					if reflectDestType.Len() >= int(count) {
						reflectDestValue.Index(int(count) - 1).Set(elem)
					}
				} else {
					reflectDestValue.Set(reflect.Append(reflectDestValue, elem))
				}
			}
		case reflect.Struct, reflect.Interface:
			if rows.Next() {
				var elem reflect.Value
				count++
				elem = reflect.New(reflectDestType)
				err = ScanStruct(rows, elem, columns, columnTypes)
				reflectDestValue.Set(elem.Elem())
				if err != nil {
					return 0, err
				}
			}
		default:
			err = rows.Scan(dest)
			if err != nil {
				return 0, err
			}
		}
	}

	return count, err

}

func ScanMap(mapValue map[string]interface{}, values []interface{}, columns []string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case string:
				err = errors.New("failed to scan map: " + r.(string))
			case error:
				err = r.(error)
			default:
				err = errors.New("failed to scan map")
			}
		}
	}()
	for i, column := range columns {
		v := reflect.Indirect(reflect.ValueOf(values[i]))
		if v.IsValid() {
			mapValue[column] = v.Interface()
			if valuer, ok := mapValue[column].(driver.Valuer); ok {
				mapValue[column], _ = valuer.Value()
			} else if bs, ok := mapValue[column].(sql.RawBytes); ok {
				mapValue[column] = string(bs)
			}
		} else {
			mapValue[column] = nil
		}
	}

	return nil
}

func ScanStruct(rows *sql.Rows, destPtr reflect.Value, columns []string, columnTypes []*sql.ColumnType) (err error) {
	scanArgs := make([]interface{}, len(columns))
	needCast := false
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case string:
				err = fmt.Errorf("failed to scan struct: %w", errors.New(r.(string)))
			case error:
				err = fmt.Errorf("failed to scan struct: %w", r.(error))
			}
		}
	}()
	config, err := GetParsedModel(destPtr.Interface())
	if err != nil {
		return err
	}
	var pivots, aggreagates map[string]interface{}
	for i, column := range columns {
		field, ok := config.FieldsByColumnName[column]
		if ok {
			if field.NeedCast {
				needCast = true
				scanArgs[i] = new(string)
			} else {
				scanArgs[i] = destPtr.Elem().FieldByName(field.Name).Addr().Interface()
			}
		} else if strings.Contains(column, OrmAggregateAlias) {
			if aggreagates == nil {
				aggreagates = make(map[string]interface{})
			}
			aggregateColumn := strings.TrimPrefix(column, OrmAggregateAlias)
			aggregateDest := reflect.New(columnTypes[i].ScanType())
			scanArgs[i] = aggregateDest.Interface()
			aggreagates[aggregateColumn] = i

		} else {
			if pivots == nil {
				pivots = make(map[string]interface{})
			}
			pivotColumn := strings.TrimPrefix(column, PivotAlias)
			pivotDest := reflect.New(columnTypes[i].ScanType())
			scanArgs[i] = pivotDest.Interface()
			pivots[pivotColumn] = i
		}

	}
	if err = rows.Scan(scanArgs...); err != nil {
		return err
	}
	if needCast {
		for i, column := range columns {
			field, ok := config.FieldsByColumnName[column]
			if ok {
				if field.NeedCast {
					ptr := reflect.New(field.FieldType).Interface()
					json.Unmarshal([]byte(*scanArgs[i].(*string)), ptr)
					destPtr.Elem().FieldByName(field.Name).Set(reflect.ValueOf(ptr).Elem())
				}
			}
		}
	}
	if config.IsEloquent && destPtr.Elem().Field(config.EloquentModelFieldIndex).IsNil() {
		origin := map[string]interface{}{}
		for i, c := range columns {
			origin[c] = scanArgs[i]
		}
		base := EloquentModel{
			Pivot:              map[string]interface{}{},
			WithAggregates:     map[string]float64{},
			Context:            nil,
			Booted:             false,
			Exists:             true,
			WasRecentlyCreated: false,
			Origin:             origin,
			Changes:            nil,
			ModelPointer:       destPtr,
			Related:            reflect.Value{},
			MutedEvents:        nil,
			Scopes:             nil,
		}
		destPtr.Elem().Field(config.EloquentModelFieldIndex).Set(reflect.ValueOf(&base))
	}
	if len(pivots) > 0 {
		for pivotColumn, i := range pivots {
			pivots[pivotColumn] = reflect.ValueOf(scanArgs[i.(int)]).Elem().Interface()
		}

		destPtr.Elem().FieldByIndex([]int{config.EloquentModelFieldIndex, EloquentModelPivotFieldIndex}).Set(reflect.ValueOf(pivots))
	}
	if len(aggreagates) > 0 {
		for aggregateColumn, i := range aggreagates {
			aggreagates[aggregateColumn] = reflect.ValueOf(scanArgs[i.(int)]).Elem().Interface()
			if _, ok := config.EagerRelationAggregates[aggregateColumn]; ok {
				destPtr.Elem().FieldByName(aggregateColumn).Set(reflect.ValueOf(aggreagates[aggregateColumn]))
			}
		}
		destPtr.Elem().FieldByIndex([]int{config.EloquentModelFieldIndex, EloquentModelAggregateFieldIndex}).Set(reflect.ValueOf(aggreagates))
	}

	return nil
}
