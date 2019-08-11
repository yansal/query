package query_test

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
		query.IntParam("goal"),
		query.StringParam("q"),
		query.StringsParam("status", []string{"pending", "processing", "success", "failure"}),
		query.CustomParam("url", func(values []string) (interface{}, error) {
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
	if uerr != "unknown" {
		t.Errorf(`expected unknown key to be "unknown", got %q`, uerr)
	}
}

func Test(t *testing.T) {
	var (
		i int64 = 42
		s       = "foo"
	)
	q, err := query.Validate(url.Values{
		"i": []string{strconv.FormatInt(i, 10)},
		"s": []string{s},
	},
		query.IntParam("i"),
		query.StringParam("s"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(q) != 2 {
		t.Errorf("expected 2 params, got %d", len(q))
	}
	ii := q["i"].(int64)
	if ii != i {
		t.Errorf("expected i to be %d, got %d", i, ii)
	}
	ss := q["s"].(string)
	if ss != s {
		t.Errorf("expected s to be %s, got %s", s, ss)
	}
}
func TestNoParam(t *testing.T) {
	q, err := query.Validate(url.Values{},
		query.StringParam("s"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(q) != 0 {
		t.Errorf("expected 0 params, got %d", len(q))
	}
}
