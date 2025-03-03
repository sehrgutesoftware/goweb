package assertions

import (
	"fmt"
	"reflect"
)

// ErrUnknownValidator is returned when the tag of a struct field contains
// an unknown validator name.
var ErrUnknownValidator = fmt.Errorf("unknown validator")

type Assertion interface {
	Validate(value any) error
}

type parser func(typ reflect.Type, args string) (Assertion, error)

var parsers = map[string]parser{
	"required": parseRequired,
	"between":  parseBetween,
}

// Resolve returns an Assertion for the given name, type, and arguments.
func Resolve(name string, typ reflect.Type, args string) (Assertion, error) {
	parser, ok := parsers[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownValidator, name)
	}
	return parser(typ, args)
}
