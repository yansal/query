package query

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// A Query is a validated query string.
type Query map[string]interface{}

// A Param is a param.
type Param struct {
	key     string
	handler ParamHandler
}

// Validate validates v and returns a new Query.
func Validate(v url.Values, params ...Param) (Query, error) {
	var q Query
	handlers := make(map[string]ParamHandler)
	for _, param := range params {
		handlers[param.key] = param.handler
	}
	var errs Errors
	for key, values := range v {
		handler, ok := handlers[key]
		if !ok {
			errs = append(errs, UnknownKeyError(key))
			continue
		}
		value, err := handler(values)
		if perr, ok := err.(ParamError); ok {
			errs = append(errs, perr)
			continue
		} else if err != nil {
			return nil, err
		}

		if q == nil {
			q = make(map[string]interface{})
		}
		q[key] = value
	}
	if errs != nil {
		return nil, errs
	}
	return q, nil
}

// Errors is a list of errors.
type Errors []error

func (e Errors) Error() string {
	var errs []string
	for _, err := range e {
		errs = append(errs, err.Error())
	}
	return strings.Join(errs, "\n")
}

// UnknownKeyError is an unknown key error.
type UnknownKeyError string

func (u UnknownKeyError) Error() string {
	return fmt.Sprintf("%s: unknown", string(u))
}

// ParamError is an error with a param.
type ParamError struct {
	Key     string
	Message string
}

func (f ParamError) Error() string {
	return fmt.Sprintf("%s: %s", f.Key, f.Message)
}

// StringParam is a string param.
func StringParam(key string) Param {
	return Param{
		key: key,
		handler: func(values []string) (interface{}, error) {
			if len(values) != 1 {
				return nil, ParamError{Key: key, Message: fmt.Sprintf("expected one value, got %d", len(values))}
			}
			return values[0], nil
		},
	}
}

// IntParam is an int param.
func IntParam(key string) Param {
	return Param{
		key: key,
		handler: func(values []string) (interface{}, error) {
			if len(values) != 1 {
				return nil, ParamError{Key: key, Message: fmt.Sprintf("expected one value, got %d", len(values))}
			}
			i, err := strconv.ParseInt(values[0], 0, 0)
			if err != nil {
				return nil, ParamError{Key: key, Message: err.Error()}
			}
			return i, nil
		},
	}
}

// StringsParam is a strings param.
func StringsParam(key string, choices []string) Param {
	return Param{
		key: key,
		handler: func(values []string) (interface{}, error) {
			var out []string
			for _, v := range values {
				var ok bool
				for i := range choices {
					if v == choices[i] {
						ok = true
						out = append(out, v)
						break
					}
				}
				if !ok {
					return nil, ParamError{Key: key, Message: fmt.Sprintf("%q is not a valid choice, expected one of %v", v, choices)}
				}
			}
			return out, nil
		},
	}
}

// CustomParam is a custom param.
func CustomParam(key string, handler ParamHandler) Param {
	return Param{
		key:     key,
		handler: handler,
	}
}

// A ParamHandler parses a list of values.
type ParamHandler func(values []string) (interface{}, error)
