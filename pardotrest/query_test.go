package pardotrest

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

// To make sure that the REST client respects all our json annotations,
// here we assert that when the server responds with `p` marshaled,
// the response is exactly like `s` when marshaled again.
func TestQueryProspects(t *testing.T) {
	type prospect struct {
		ID            int    `json:"id"`
		Email         string `json:"email"`
		PocInProgress string `json:"poc_in_progress,omitempty"`
	}

	tests := []struct {
		p prospect
		s string
	}{
		{
			prospect{ID: 10, PocInProgress: "APNIC"},
			`{
 "id": 10,
 "email": "",
 "poc_in_progress": "APNIC"
}`,
		},
		{
			prospect{ID: 20, Email: "a@b.com"},
			`{
 "id": 20,
 "email": "a@b.com"
}`,
		},
	}

	testClient := newTestClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `login/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{"api_key":"anyapikey"}`)),
				Header:     make(http.Header)}
		case strings.Contains(u, `/query`):
			strs := []string{}
			for _, test := range tests {
				b, err := json.Marshal(test.p)
				if err != nil {
					t.Fatal(err)
				}
				strs = append(strs, string(b))
			}
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(
					`{"result":{"prospect":[` + strings.Join(strs, ",") + `]}}`)),
				Header: make(http.Header)}
		default:
			t.Fatal("endpoint not called")
			return nil
		}
	})

	res := []prospect{}
	pardot := NewPardotREST().WithCustomClient(testClient)
	err := pardot.Call(QueryProspects{
		Offset:      0,
		Limit:       1,
		Fields:      []string{"id"},
		PlaceHolder: &res,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != len(tests) {
		t.Fatalf("expected %d prospects, got %d", len(tests), len(res))
	}
	for i, r := range res {
		b, err := json.MarshalIndent(r, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != tests[i].s {
			t.Errorf("expected %s, got %s", tests[i].s, string(b))
		}
	}
}
