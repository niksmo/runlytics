package field

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNotStruct       = errors.New("given obj is not a struct")
	ErrNoField         = errors.New("field is not present in the struct")
	ErrUnexportedField = errors.New("specified field is not an exported or public field")
)

func Value(obj any, field string) (any, error) {
	sV, err := getReflectValue(obj)
	if err != nil {
		return nil, fmt.Errorf("%w, obj: %#v", err, obj)
	}

	fV := sV.FieldByName(field)

	if !fV.IsValid() {
		return nil, ErrNoField
	}

	if !fV.CanInterface() {
		return nil, ErrUnexportedField
	}

	return fV.Interface(), nil
}

// Return a reflect value of struct
func getReflectValue(obj any) (reflect.Value, error) {
	var v reflect.Value
	switch v = reflect.ValueOf(obj); v.Kind() {
	case reflect.Struct:
		return v, nil
	case reflect.Pointer:
		if v = v.Elem(); v.Kind() == reflect.Struct {
			return v, nil
		}
	}

	return v, ErrNotStruct
}
