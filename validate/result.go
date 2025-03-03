package validate

import "net/http"

// result is an error type that is returned when validation fails.
//
// It implements the [error], [goweb.ErrorCoder], [goweb.APIError] interfaces.
type result struct {
	fields map[string][]error
}

// Error returns the error message.
func (r *result) Error() string {
	return "entity validation failed"
}

// ErrorCoder returns the error code.
func (r *result) ErrorCode() string {
	return "invalid_entity"
}

// StatusCode returns the HTTP status code.
func (r *result) StatusCode() int {
	return http.StatusUnprocessableEntity
}

// ErrorDetail returns the optional data associated with the error.
func (r *result) ErrorDetail() any {
	result := make(map[string][]struct {
		Code     string         `json:"code"`
		Message  string         `json:"message"`
		Template string         `json:"template"`
		Values   map[string]any `json:"values"`
	})
	for k, v := range r.fields {
		for _, e := range v {
			var fe struct {
				Code     string         `json:"code"`
				Message  string         `json:"message"`
				Template string         `json:"template"`
				Values   map[string]any `json:"values"`
			}
			if f, ok := e.(interface {
				Code() string
				Message() string
				Template() string
				Values() map[string]any
			}); ok {
				fe.Code = f.Code()
				fe.Message = f.Message()
				fe.Template = f.Template()
				fe.Values = f.Values()
			} else {
				fe.Code = "unknown"
				fe.Message = e.Error()
			}
			result[k] = append(result[k], fe)
		}
	}
	return result
}
