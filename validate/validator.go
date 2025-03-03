package validate

type Validator interface {
	Validate(value any) error
}
