package query_test

import (
	"fmt"
	"net/http"

	"github.com/yansal/query"
)

func Example_Validate() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q, err := query.Validate(r.URL.Query(),
			query.WithIntParam("goal"),
			query.WithStringParam("q"),
			query.WithLangParam("lang"),
			query.WithCountryParam("country"),
			query.WithStringsParam("choices", []string{"pending", "processing", "success", "failure"}),
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
		if qerr, ok := err.(query.Errors); ok {
			http.Error(w, qerr.Error(), http.StatusBadRequest)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, q.Params)
	})
	http.ListenAndServe(":8080", nil)
}
