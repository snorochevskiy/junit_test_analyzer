package main

import (
	"database/sql"
	"reflect"
)

func ScanStruct(rows *sql.Rows, obj interface{}) error {
	cols, _ := rows.Columns()

	fieldsMap := make(map[string]interface{})
	collectFields(fieldsMap, reflect.ValueOf(obj), cols)

	fieldPtrs := make([]interface{}, len(cols), len(cols))
	for i := 0; i < len(cols); i++ {
		fieldPtrs[i] = fieldsMap[cols[i]]
	}

	return rows.Scan(fieldPtrs...)
}

func collectFields(fieldsMap map[string]interface{}, dest reflect.Value, cols []string) {

	if dest.Kind() == reflect.Ptr {
		dest = dest.Elem()
	}

	for i := 0; i < dest.Type().NumField(); i++ {
		if dest.Field(i).Kind() == reflect.Struct {
			collectFields(fieldsMap, dest.Field(i), cols)
		}

		fieldName := dest.Type().Field(i).Name
		if isValueInList(fieldName, cols) {
			fieldsMap[fieldName] = dest.Field(i).Addr().Interface()
			continue
		}

		fieldColumnTag := dest.Type().Field(i).Tag.Get("column")
		if isValueInList(fieldColumnTag, cols) {
			fieldsMap[fieldColumnTag] = dest.Field(i).Addr().Interface()
			continue
		}
	}
}

func isValueInList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
