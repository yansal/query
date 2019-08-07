package query

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// A Query is a validated query string.
type Query struct {
	Params   map[string]interface{}
	handlers map[string]ParamHandler
}

// Validate validates v and returns a new Query.
func Validate(v url.Values, options ...Option) (*Query, error) {
	q := Query{}
	for _, o := range options {
		o(&q)
	}

	var errs Errors
	for key, values := range v {
		handler, ok := q.handlers[key]
		if !ok {
			errs = append(errs, UnknownKeyError(key))
			continue
		}

		value, err := handler(values)
		if ferr, ok := err.(ParamError); ok {
			errs = append(errs, ferr)
			continue
		} else if err != nil {
			return nil, err
		}

		if q.Params == nil {
			q.Params = make(map[string]interface{})
		}
		q.Params[key] = value
	}
	if errs != nil {
		return nil, errs
	}
	return &q, nil
}

// An Option is a functional option for query parsing.
type Option func(q *Query)

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
	return fmt.Sprintf("%s: unknown key", string(u))
}

// ParamError is an error with a param.
type ParamError struct {
	Key     string
	Message string
}

func (f ParamError) Error() string {
	return fmt.Sprintf("%s: %s", f.Key, f.Message)
}

// WithStringParam is a string param option.
func WithStringParam(key string) Option {
	return func(q *Query) {
		if q.handlers == nil {
			q.handlers = make(map[string]ParamHandler)
		}
		q.handlers[key] = func(values []string) (interface{}, error) {
			if len(values) != 1 {
				return nil, ParamError{Key: key, Message: fmt.Sprintf("expected one value, got %d", len(values))}
			}
			return values[0], nil
		}
	}
}

// WithIntParam is an int param option.
func WithIntParam(key string) Option {
	return func(q *Query) {
		if q.handlers == nil {
			q.handlers = make(map[string]ParamHandler)
		}
		q.handlers[key] = func(values []string) (interface{}, error) {
			if len(values) != 1 {
				return nil, ParamError{Key: key, Message: fmt.Sprintf("expected one value, got %d", len(values))}
			}
			i, err := strconv.ParseInt(values[0], 0, 0)
			if err != nil {
				return nil, ParamError{Key: key, Message: err.Error()}
			}
			return i, nil
		}
	}
}

// WithStringsParam is a strings param option.
func WithStringsParam(key string, choices []string) Option {
	return func(q *Query) {
		if q.handlers == nil {
			q.handlers = make(map[string]ParamHandler)
		}
		q.handlers[key] = func(values []string) (interface{}, error) {
			commaseparated := strings.Join(values, ",")
			var out []string
			for _, v := range strings.Split(commaseparated, ",") {
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
		}
	}
}

// WithLangParam is a lang param option.
func WithLangParam(key string) Option {
	return WithStringsParam(key, []string{"en", "fr", "es", "it", "de"})
}

// WithCountryParam is a country param option.
func WithCountryParam(key string) Option {
	return WithStringsParam(key, []string{"UK", "US", "FR", "BE", "CA", "ES", "IT", "DE"})
}

// WithParam is a custom param option.
func WithParam(key string, handler ParamHandler) Option {
	return func(q *Query) {
		if q.handlers == nil {
			q.handlers = make(map[string]ParamHandler)
		}
		q.handlers[key] = handler
	}
}

// A ParamHandler parses a list of values.
type ParamHandler func(values []string) (interface{}, error)
