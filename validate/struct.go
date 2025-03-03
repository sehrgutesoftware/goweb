package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sehrgutesoftware/goweb/validate/assertions"
)

var (
	// ErrTypeMismatch is returned when the value passed to Validate is not of
	// the same type as the struct validator was created for.
	ErrTypeMismatch = fmt.Errorf("type mismatch")
)

type structValidator struct {
	typ    reflect.Type
	fields map[string]fieldSpec
}

// Struct creates a new validator for the type of the given struct.
//
// You can pass a pointer to the zero value of the struct type you want to
// validate. The struct must have fields with the `validate` tag. The tag
// value is a comma-separated list of assertions. Each assertion can have
// arguments separated by a colon.
func Struct(s any) (Validator, error) {
	st := reflect.TypeOf(s)

	if st.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %T", s)
	}

	sv := structValidator{
		typ:    st,
		fields: make(map[string]fieldSpec),
	}

	for i := range st.NumField() {
		var assertions []Validator
		var err error

		field := st.Field(i)
		tag, ok := field.Tag.Lookup("validate")
		if ok && tag != "" {
			assertions, err = assertionsFromTag(tag, field.Type)
			if err != nil {
				return nil, err
			}
		}

		if field.Type.Kind() == reflect.Struct {
			nestedValidator, err := Struct(reflect.New(field.Type).Elem().Interface())
			if err != nil {
				return nil, err
			}
			assertions = append(assertions, nestedValidator)
		}

		if len(assertions) > 0 {
			sv.fields[field.Name] = fieldSpec{
				alias:      fieldAlias(field),
				assertions: assertions,
			}
		}
	}

	return &sv, nil
}

func (v *structValidator) Validate(value any) error {
	vt := reflect.TypeOf(value)
	if vt != v.typ {
		return fmt.Errorf("%w: expected %s, got %s", ErrTypeMismatch, v.typ, vt)
	}

	valErr := result{
		fields: make(map[string][]error),
	}

	val := reflect.ValueOf(value)
	for name, spec := range v.fields {
		field := val.FieldByName(name)

		alias := name
		if spec.alias != "" {
			alias = spec.alias
		}

		for _, validator := range spec.assertions {
			r := validator.Validate(field.Interface())

			// If the result is a validationError itself, we flatten the fields
			// by prepending the current alias to the field names of the nested
			// validationError.
			if ve, ok := r.(*result); ok {
				for k, v := range ve.fields {
					valErr.fields[strings.Join([]string{alias, k}, ".")] = append(valErr.fields[k], v...)
				}
			} else if r != nil {
				valErr.fields[alias] = append(valErr.fields[alias], r)
			}
		}
	}

	if len(valErr.fields) > 0 {
		return &valErr
	}

	return nil
}

type fieldSpec struct {
	alias      string
	assertions []Validator
}

func assertionsFromTag(tag string, typ reflect.Type) ([]Validator, error) {
	var validators []Validator
	for spec := range strings.SplitSeq(tag, ",") {
		name, arg, _ := strings.Cut(spec, ":")

		validator, err := assertions.Resolve(name, typ, arg)
		if err != nil {
			return nil, err
		}

		validators = append(validators, validator)
	}
	return validators, nil
}

func fieldAlias(field reflect.StructField) string {
	jsonTag, ok := field.Tag.Lookup("json")
	if !ok {
		return ""
	}
	alias, _, _ := strings.Cut(jsonTag, ",")
	return alias
}
