package assertions

import (
	"reflect"
)

// required asserts that a value is present.
//
// For primitive types, the zero value counts as present. For slices, maps, and
// strings, the length must be greater than zero. For pointers, the value must
// be non-nil.
type required struct{}

func parseRequired(typ reflect.Type, args string) (Assertion, error) {
	return required{}, nil
}

func (r required) Validate(value any) error {
	vt := reflect.TypeOf(value)

	switch vt.Kind() {
	case reflect.Ptr:
		if reflect.ValueOf(value).IsNil() {
			return errRequired()
		}
	case reflect.Slice, reflect.Map, reflect.String:
		if reflect.ValueOf(value).Len() == 0 {
			return errRequired()
		}
	}

	return nil
}

func errRequired() error {
	return &result{
		code:     "required",
		message:  "the value is required",
		template: "the value is required",
		values:   nil,
	}
}
