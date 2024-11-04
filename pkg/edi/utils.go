package edi

import (
	"reflect"
	"slices"
	"strings"

	"github.com/azarc-io/go-edi/internal/model"
)

type orderedEdiItem struct {
	Name     string
	Property *model.Property
}

func getOrderedEdi(input model.Properties) []orderedEdiItem {
	var items []orderedEdiItem
	for k, s := range input {
		items = append(items, orderedEdiItem{
			Name:     k,
			Property: s,
		})
	}
	slices.SortFunc(items, func(e1 orderedEdiItem, e2 orderedEdiItem) int {
		if e1.Property.XEdi.Order > e2.Property.XEdi.Order {
			return 1
		} else {
			return -1
		}
	})
	return items
}

func isMapStringAny(input any) bool {
	val := reflect.ValueOf(input)
	return val.Kind() == reflect.Map && val.Type().Key().Kind() == reflect.String && val.Type().Elem().Kind() == reflect.Interface
}
func isStruct(input any) bool {
	if reflect.ValueOf(input).Kind() == reflect.Ptr {
		return isStruct(reflect.ValueOf(input).Elem())
	}
	return reflect.ValueOf(input).Kind() == reflect.Struct
}
func isSlice(input any) bool {
	if reflect.ValueOf(input).Kind() == reflect.Ptr {
		return isSlice(reflect.ValueOf(input).Elem())
	}
	return reflect.ValueOf(input).Kind() == reflect.Slice
}

func toMapRecursive(input any) (any, bool) {
	v := reflect.ValueOf(input)

	// Unwrap pointers until we reach a non-pointer type
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		result := make(map[string]any)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := t.Field(i)

			// Only export public fields
			if field.CanInterface() {
				// Get the JSON tag if it exists, otherwise use the field name
				jsonTag := strings.Split(fieldType.Tag.Get("json"), ",")[0]
				fieldName := fieldType.Name
				if jsonTag != "" && jsonTag != "-" {
					fieldName = jsonTag
				}
				val, hasValue := toMapRecursive(field.Interface())
				if hasValue {
					result[fieldName] = val
				}
			}
		}
		return result, len(result) > 0

	case reflect.Slice, reflect.Array:
		sliceResult := []any{}
		for i := 0; i < v.Len(); i++ {
			val, hasValue := toMapRecursive(v.Index(i).Interface())
			if hasValue {
				sliceResult = append(sliceResult, val)
			}
		}
		return sliceResult, len(sliceResult) > 0

	case reflect.Map:
		mapResult := make(map[string]any)
		for _, key := range v.MapKeys() {
			// Ensure the map key is a string, otherwise skip or handle as needed
			if key.Kind() == reflect.String {
				strKey := key.Interface().(string)
				val, hasValue := toMapRecursive(v.MapIndex(key).Interface())
				if hasValue {
					mapResult[strKey] = val
				}
			}
		}
		return mapResult, len(mapResult) > 0

	case reflect.String:
		return input, len(v.String()) > 0
	default:
		// Return the primitive value as is
		return input, true
	}
}
