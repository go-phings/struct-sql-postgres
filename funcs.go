package structsqlpostgres

import (
	"reflect"
)

// GetStructName returns struct name of a struct instance
func GetStructName(u interface{}) string {
	v := reflect.ValueOf(u)
	i := reflect.Indirect(v)
	s := i.Type()
	return s.Name()
}

// GetStructFieldNames returns list of names of fields of a struct instance, which are supported in generating SQL query.
func GetStructFieldNames(u interface{}) []string {
	v := reflect.ValueOf(u)
	i := reflect.Indirect(v)
	s := i.Type()

	names := []string{}

	for j := 0; j < s.NumField(); j++ {
		f := s.Field(j)
		k := f.Type.Kind()

		if !IsFieldKindSupported(k) {
			continue
		}

		names = append(names, f.Name)
	}

	return names
}

// GetStructNamesFromConstructors returns a list of struct names from a list of constructors (functions that return struct instances)
func GetStructNamesFromConstructors(objFuncs ...func() interface{}) []string {
	names := []string{}
	for _, objFunc := range objFuncs {
		o := objFunc()
		v := reflect.ValueOf(o)
		i := reflect.Indirect(v)
		s := i.Type()
		names = append(names, s.Name())
	}
	return names
}

// IsFieldKindSupported checks if a field kind is supported by this module
func IsFieldKindSupported(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.String, reflect.Bool:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// IsStructField checks if a specific string is a name of a field
func IsStructField(u interface{}, field string) bool {
	v := reflect.ValueOf(u)
	i := reflect.Indirect(v)
	s := i.Type()

	for j := 0; j < s.NumField(); j++ {
		f := s.Field(j)
		k := f.Type.Kind()

		if !IsFieldKindSupported(k) {
			continue
		}

		if f.Name == field {
			return true
		}
	}

	return false
}
