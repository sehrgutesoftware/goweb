package goweb

// ErrorsMap is a map of error codes to errors. It can be used to as a lookup
// for error codes.
//
// It fulfills the [goweb.serror.CodeResolver] interface.
type ErrorMap map[string]*codeError

// Resolve returns the error for the given code.
func (m ErrorMap) Resolve(code string) (error, bool) {
	err, ok := m[code]
	return err, ok
}
