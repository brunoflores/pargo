package pargo_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"gitlab.xyz.apnic.net/go-pkg/pargo"
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

	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `oauth2/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header)}
		case strings.Contains(u, `/query`):
			if got := req.FormValue("offset"); got != "100" {
				t.Fatalf("expected query string offset=%s, got: %s", "100", got)
			}
			if got := req.FormValue("limit"); got != "200" {
				t.Fatalf("expected query string limit=%s, got: %s", "200", got)
			}
			if got := req.FormValue("fields"); got != "id,email" {
				t.Fatalf("expected query string fields=%s, got: %s", "id,email", got)
			}
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
	client := newTestClient(testClient)
	err := client.QueryProspects(pargo.QueryProspects{
		Offset:      100,
		Limit:       200,
		Fields:      []string{"id", "email"},
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

func TestQueryReadsEmptyPage(t *testing.T) {
	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `oauth2/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header)}
		case strings.Contains(u, `/query`):
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(
					`{"result":{"total_results": 2, "prospect":[]}}`)),
				Header: make(http.Header)}
		default:
			t.Fatal("no endpoint called")
			return nil
		}
	})

	client := newTestClient(testClient)
	var prospects []struct{}
	err := client.QueryProspects(pargo.QueryProspects{
		Offset:      100,
		Limit:       200,
		Fields:      []string{"id"},
		PlaceHolder: &prospects,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(prospects) != 0 {
		t.Fatalf("expected 0 prospects, got %d", len(prospects))
	}
}
