package pargo_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"

	"gitlab.xyz.apnic.net/go-pkg/pargo"
)

func TestQueryAllWithError(t *testing.T) {
	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `oauth2/`):
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(
					bytes.NewBufferString(`{}`)),
				Header: make(http.Header)}
		case strings.Contains(u, `/query`):
			return &http.Response{
				StatusCode: 503,
				Body: ioutil.NopCloser(
					bytes.NewBufferString("{}")),
				Header: make(http.Header)}
		default:
			t.Fatalf("unknown endpoint called %q", u)
			return nil
		}
	})
	client := newTestClient(testClient)
	err := client.QueryAllProspects(pargo.QueryAllProspects{
		Fields: []string{"id"},
		Page:   func(json.RawMessage) {},
	})
	if err != nil {
		// OK
		return
	}
	t.Fatal("want error; got nil")
}

func TestQueryAllProspects(t *testing.T) {
	const twoProspects = `{
  "result":{
    "total_results": 0,
    "prospect":[
      {
        "id": 10
      }
    ]
  }
}`
	const noProspects = `{
  "result":{
    "total_results": 0
  }
}`
	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `oauth2/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header)}
		case strings.Contains(u, `/query`):
			req.ParseForm()
			offset := req.FormValue("offset")
			var bodyStr string
			switch offset {
			case "0":
				bodyStr = twoProspects
			default:
				bodyStr = noProspects
			}
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(
					bytes.NewBufferString(bodyStr)),
				Header: make(http.Header)}
		default:
			t.Fatalf("unknown endpoint called %q", u)
			return nil
		}
	})
	client := newTestClient(testClient)
	type p struct {
		ID int `json:"id"`
	}
	var prospects []p
	var mu sync.Mutex
	readPage := func(data json.RawMessage) {
		mu.Lock()
		defer mu.Unlock()
		var tmp []p
		err := json.Unmarshal(data, &tmp)
		if err != nil {
			t.Fatal(err)
		}
		prospects = append(prospects, tmp...)
	}
	err := client.QueryAllProspects(pargo.QueryAllProspects{
		Fields: []string{"id"},
		Page:   readPage,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := len(prospects); got != 1 {
		t.Fatalf("len(prospects) = %d; want %d", got, 1)
	}
}
