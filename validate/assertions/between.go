package assertions

import (
	"fmt"
	"reflect"

	"golang.org/x/exp/constraints"
)

func parseBetween(typ reflect.Type, args string) (Assertion, error) {
	switch typ.Kind() {
	case reflect.Int:
		return parseBetweenInt[int](args)
	case reflect.Int8:
		return parseBetweenInt[int8](args)
	case reflect.Int16:
		return parseBetweenInt[int16](args)
	case reflect.Int32:
		return parseBetweenInt[int32](args)
	case reflect.Int64:
		return parseBetweenInt[int64](args)
	case reflect.Uint:
		return parseBetweenInt[uint](args)
	case reflect.Uint8:
		return parseBetweenInt[uint8](args)
	case reflect.Uint16:
		return parseBetweenInt[uint16](args)
	case reflect.Uint32:
		return parseBetweenInt[uint32](args)
	case reflect.Uint64:
		return parseBetweenInt[uint64](args)
	case reflect.Float32:
		return parseBetweenFloat[float32](args)
	case reflect.Float64:
		return parseBetweenFloat[float64](args)
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return parseBetweenLen(args)
	}

	return nil, fmt.Errorf("unsupported type %s", typ)
}

// betweenInt asserts that an integer value is within a given range.
type betweenInt[T constraints.Integer] struct {
	lower T
	upper T
}

func parseBetweenInt[T constraints.Integer](args string) (Assertion, error) {
	var lower, upper T
	if _, err := fmt.Sscanf(args, "%d:%d", &lower, &upper); err != nil {
		return nil, fmt.Errorf("parse args (%s): %w", args, err)
	}

	return &betweenInt[T]{
		lower: lower,
		upper: upper,
	}, nil
}

func (b *betweenInt[T]) Validate(value any) error {
	v, ok := value.(T)
	if !ok {
		var t T
		return fmt.Errorf("bad type, expected %T, go %T", t, value)
	}

	if v < b.lower || v > b.upper {
		return errBetweenInt[T](b.lower, b.upper, v)
	}

	return nil
}

func errBetweenInt[T constraints.Integer](min, max, actual T) error {
	return &result{
		code:     "between",
		message:  fmt.Sprintf("the value must be between %d and %d (is %d)", min, max, actual),
		template: "the value must be between {min} and {max} (is {actual})",
		values: map[string]any{
			"min":    min,
			"max":    max,
			"actual": actual,
		},
	}
}

// betweenFloat asserts that a float value is within a given range.
type betweenFloat[T constraints.Float] struct {
	lower T
	upper T
}

func parseBetweenFloat[T constraints.Float](args string) (Assertion, error) {
	var lower, upper T
	if _, err := fmt.Sscanf(args, "%f:%f", &lower, &upper); err != nil {
		return nil, fmt.Errorf("parse args (%s): %w", args, err)
	}

	return &betweenFloat[T]{
		lower: lower,
		upper: upper,
	}, nil
}

func (b *betweenFloat[T]) Validate(value any) error {
	v, ok := value.(T)
	if !ok {
		var t T
		return fmt.Errorf("bad type, expected %T, go %T", t, value)
	}

	if v < b.lower || v > b.upper {
		return errBetweenFloat[T](b.lower, b.upper, v)
	}

	return nil
}

func errBetweenFloat[T constraints.Float](min, max, actual T) error {
	return &result{
		code:     "between",
		message:  fmt.Sprintf("the value must be between %f and %f (is %f)", min, max, actual),
		template: "the value must be between {min} and {max} (is {actual})",
		values: map[string]any{
			"min":    min,
			"max":    max,
			"actual": actual,
		},
	}
}

// betweenLen asserts that a string, slice, map, or array has a length within a
// given range.
type betweenLen struct {
	lower uint64
	upper uint64
}

func parseBetweenLen(args string) (Assertion, error) {
	var lower, upper uint64
	if _, err := fmt.Sscanf(args, "%d:%d", &lower, &upper); err != nil {
		return nil, fmt.Errorf("parse args (%s): %w", args, err)
	}
	return &betweenLen{
		lower: lower,
		upper: upper,
	}, nil
}

func (b *betweenLen) Validate(value any) error {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.String && v.Kind() != reflect.Slice && v.Kind() != reflect.Map && v.Kind() != reflect.Array {
		return fmt.Errorf("bad type, expected string, slice, map, or array, got %T", value)
	}

	l := uint64(v.Len())
	if l < b.lower || l > b.upper {
		return errBetweenLen(b.lower, b.upper, l)
	}

	return nil
}

func errBetweenLen(min, max, actual uint64) error {
	return &result{
		code:     "between",
		message:  fmt.Sprintf("length must be between %d and %d (is %d)", min, max, actual),
		template: "length must be between {min} and {max} (is {actual})",
		values: map[string]any{
			"min":    min,
			"max":    max,
			"actual": actual,
		},
	}
}
