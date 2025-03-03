package assertions

type result struct {
	code     string
	message  string
	template string
	values   map[string]any
}

func (ar *result) Error() string {
	return ar.message
}

func (ar *result) Code() string {
	return ar.code
}

func (ar *result) Message() string {
	return ar.message
}

func (ar *result) Template() string {
	return ar.template
}

func (ar *result) Values() map[string]any {
	return ar.values
}
