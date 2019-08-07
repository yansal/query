package query

import (
	"errors"
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
	var q Query
	for _, o := range options {
		o(&q)
	}

	var errs Errors
	for key, handler := range q.handlers {
		value, err := handler(v[key])
		if err == errHasNoDefault {
			continue
		}
		delete(v, key)
		if perr, ok := err.(ParamError); ok {
			errs = append(errs, perr)
			continue
		} else if err != nil {
			return nil, err
		}
		if q.Params == nil {
			q.Params = make(map[string]interface{})
		}
		q.Params[key] = value
	}

	var unknown []string
	for k := range v {
		unknown = append(unknown, k)
	}
	if unknown != nil {
		errs = append(errs, UnknownKeyError(unknown))
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
type UnknownKeyError []string

func (u UnknownKeyError) Error() string {
	return fmt.Sprintf("%s: unknown", strings.Join(u, ","))
}

var errHasNoDefault = errors.New("has no default")

// ParamError is an error with a param.
type ParamError struct {
	Key     string
	Message string
}

func (f ParamError) Error() string {
	return fmt.Sprintf("%s: %s", f.Key, f.Message)
}

// WithStringParam is a string param option.
func WithStringParam(key string, options ...ParamOption) Option {
	var opts ParamOptions
	for _, o := range options {
		o(&opts)
	}
	return func(q *Query) {
		if q.handlers == nil {
			q.handlers = make(map[string]ParamHandler)
		}
		q.handlers[key] = func(values []string) (interface{}, error) {
			if values == nil {
				if opts.defaultvalue == nil {
					return nil, errHasNoDefault
				}
				return opts.defaultvalue, nil
			}
			if len(values) != 1 {
				return nil, ParamError{Key: key, Message: fmt.Sprintf("expected one value, got %d", len(values))}
			}
			return values[0], nil
		}
	}
}

// WithIntParam is an int param option.
func WithIntParam(key string, options ...ParamOption) Option {
	var opts ParamOptions
	for _, o := range options {
		o(&opts)
	}

	return func(q *Query) {
		if q.handlers == nil {
			q.handlers = make(map[string]ParamHandler)
		}
		q.handlers[key] = func(values []string) (interface{}, error) {
			if values == nil {
				if opts.defaultvalue == nil {
					return nil, errHasNoDefault
				}
				return opts.defaultvalue, nil
			}
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
func WithStringsParam(key string, choices []string, options ...ParamOption) Option {
	var opts ParamOptions
	for _, o := range options {
		o(&opts)
	}

	return func(q *Query) {
		if q.handlers == nil {
			q.handlers = make(map[string]ParamHandler)
		}
		q.handlers[key] = func(values []string) (interface{}, error) {
			if values == nil {
				if opts.defaultvalue == nil {
					return nil, errHasNoDefault
				}
				return opts.defaultvalue, nil
			}
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
		}
	}
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

// A ParamOption is a functional option for params.
type ParamOption func(*ParamOptions)

// ParamOptions are param options.
type ParamOptions struct {
	defaultvalue interface{}
}

// WithDefault sets a param default value.
func WithDefault(value interface{}) ParamOption {
	return func(o *ParamOptions) {
		o.defaultvalue = value
	}
}
