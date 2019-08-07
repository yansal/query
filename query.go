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

// WithStringFilter is a string filter option.
func WithStringFilter(key string) Option {
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

// WithIntFilter is an int filter option.
func WithIntFilter(key string) Option {
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

// WithChoicesFilter is a choices filter option.
func WithChoicesFilter(key string, choices []string) Option {
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

// WithMapFilter is a map filter option.
func WithMapFilter(key string, m map[string]int64) Option {
	var choices []string
	for key := range m {
		choices = append(choices, key)
	}
	return func(q *Query) {
		if q.handlers == nil {
			q.handlers = make(map[string]ParamHandler)
		}
		q.handlers[key] = func(values []string) (interface{}, error) {
			commaseparated := strings.Join(values, ",")
			var out []int64
			for _, v := range strings.Split(commaseparated, ",") {
				var ok bool
				for mapkey, mapvalue := range m {
					if v == mapkey {
						out = append(out, mapvalue)
						ok = true
						break
					}
				}
				if !ok {
					return nil, ParamError{Key: key, Message: fmt.Sprintf("value %q is not a valid choice, expected one of %v", v, choices)}
				}
			}
			return out, nil
		}
	}
}

// WithLangFilter is a lang filter option.
func WithLangFilter(key string) Option {
	return WithChoicesFilter(key, []string{"en", "fr", "es", "it", "de"})
}

// WithCountryFilter is a country filter option.
func WithCountryFilter(key string) Option {
	return WithChoicesFilter(key, []string{"UK", "US", "FR", "BE", "CA", "ES", "IT", "DE"})
}

// WithFilter is a custom filter option.
func WithFilter(key string, handler ParamHandler) Option {
	return func(q *Query) {
		if q.handlers == nil {
			q.handlers = make(map[string]ParamHandler)
		}
		q.handlers[key] = handler
	}
}

// A ParamHandler parses a list of values.
type ParamHandler func(values []string) (interface{}, error)
