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
		case reflect.Struct:
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
