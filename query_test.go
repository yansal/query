package query_test

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/yansal/query"
)

func ExampleValidate() {
	v := url.Values{
		"unknown": []string{"key"},
		"goal":    []string{"not an int"},
		"q":       []string{"two", "strings"},
		"status":  []string{"foo"},
		"url":     []string{"http://google.com"},
	}
	_, err := query.Validate(v,
		query.WithIntParam("goal"),
		query.WithStringParam("q"),
		query.WithStringsParam("status", []string{"pending", "processing", "success", "failure"}),
		query.WithCustomParam("url", func(values []string) (interface{}, error) {
			if len(values) != 1 {
				return nil, query.ParamError{Key: "url", Message: fmt.Sprintf("expected one value, got %d", len(values))}
			}
			v := values[0]
			resp, err := http.Get(v)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			return resp.Status, nil
		}),
	)
	qerr := err.(query.Errors)
	fmt.Println(len(qerr))
	// Output: 4
}
