package query_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/yansal/query"
)

func ExampleValidate() {
	v := url.Values{
		"goal":   []string{"not an int"},
		"q":      []string{"two", "strings"},
		"status": []string{"foo"},
		"url":    []string{"http://google.com"},
	}
	_, err := query.Validate(v,
		query.WithIntParam("goal"),
		query.WithStringParam("q"),
		query.WithStringsParam("status", []string{"pending", "processing", "success", "failure"}),
		query.WithParam("url", func(values []string) (interface{}, error) {
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
	// Output: 3
}

func TestWithDefault(t *testing.T) {
	defaultgoal := 42
	q, err := query.Validate(url.Values{},
		query.WithIntParam("goal", query.WithDefault(defaultgoal)),
	)
	if err != nil {
		t.Fatal(err)
	}
	goal, ok := q.Params["goal"]
	if !ok {
		t.Error("expected to have goal")
	}
	if goal != defaultgoal {
		t.Errorf("expected goal to be %d, got %d", defaultgoal, goal)
	}
}

func TestHasNoDefault(t *testing.T) {
	q, err := query.Validate(url.Values{},
		query.WithIntParam("goal"),
	)
	if err != nil {
		t.Fatal(err)
	}
	goal, ok := q.Params["goal"]
	if ok {
		t.Errorf("expected to not have goal, got %v", goal)
	}
}

func TestUnknownKey(t *testing.T) {
	_, err := query.Validate(url.Values{
		"unknown": []string{"key"},
	})
	qerr, ok := err.(query.Errors)
	if !ok {
		t.Errorf("expected to have query.Errors, got %v (%T)", err, err)
	}
	if len(qerr) != 1 {
		t.Errorf("expected to have 1 err, got %d", len(qerr))
	}
	err = qerr[0]
	uerr, ok := err.(query.UnknownKeyError)
	if !ok {
		t.Errorf("expected to have query.UnknownKeyError, got %v (%T)", err, err)
	}
	if len(uerr) != 1 {
		t.Errorf("expected to have 1 unknown key, got %d", len(uerr))
	}
	if uerr[0] != "unknown" {
		t.Errorf(`expected unknown key to be "unknown", got %q`, uerr[0])
	}
}
