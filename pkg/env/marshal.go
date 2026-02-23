package env

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// MarshalEnv reflects over the struct and creates .env content from tags
func MarshalEnv(c any) (string, error) {
	var lines []string
	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("env")

		// Skip fields without env tag or unexported fields
		if tag == "" || !field.IsExported() {
			continue
		}

		// Parse tag: "KEY,required,notEmpty" or "KEY" or "KEY" envDefault:"value"
		parts := strings.Split(tag, ",")
		key := parts[0]

		// Skip if no key specified
		if key == "" {
			continue
		}

		val := v.Field(i)

		// Check if value is empty/zero
		if isZeroValue(val) {
			continue
		}

		// Format value based on type
		strVal := formatValue(val)
		lines = append(lines, fmt.Sprintf("%s=%s", key, strVal))
	}

	result := strings.Join(lines, "\n")
	if result != "" && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	return result, nil
}

// isZeroValue checks if a reflect.Value is the zero value for its type
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}

// formatValue converts a reflect.Value to its string representation
func formatValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
