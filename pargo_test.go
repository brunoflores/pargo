package pargo_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/brunoflores/pargo"
)

// The requirement is that an API Key should be requested if none is found,
// and then it should be reused until any authenticated request returns
// err code 1 from Pardot.
// When err code 1 occurs, a new API Key should be requested and then used.
func TestReuseAPIKeyUntilExpired(t *testing.T) {
	requests := []struct {
		path, apiKey  string
		returnExpired bool
	}{
		{
			"oauth2/",
			"apikey#0",
			false,
		},
		{
			"/query",
			"apikey#0",
			false,
		},
		{
			"/query",
			"apikey#0",
			false,
		},
		{
			"/query",
			"apikey#0",
			true,
		},
		{
			"oauth2/",
			"apikey#1",
			false,
		},
		{
			"/query",
			"apikey#1",
			false,
		},
	}

	// State kept over executions of requests.
	// It is used to validate the current state by comparison with
	// the expected.
	currentIndex := 0

	const (
		keyExpired = `{
"err":"Invalid API key or user key",
"@attributes":{"err_code": 1}
}`
	)

	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {

		// Index is incremented on exit.
		defer func() { currentIndex++ }()

		u := req.URL.Path
		switch {
		case strings.Contains(u, `oauth2/`):
			if requests[currentIndex].path != `oauth2/` {
				t.Fatalf("request #%d expected to be oauth2", currentIndex)
			}
			if got := req.PostFormValue("username"); got != "a@b.com" {
				t.Fatalf("expected credential: username=%s, got: %s", "a@b.com", got)
			}
			if got := req.PostFormValue("password"); got != "pass" {
				t.Fatalf("expected credential: password=%s, got: %s", "pass", got)
			}
			if got := req.PostFormValue("client_secret"); got != "clientsecret" {
				t.Fatalf("expected credential: client_secret=%s, got: %s", "clientsecret", got)
			}
			if got := req.PostFormValue("client_id"); got != "clientid" {
				t.Fatalf("expected credential: client_id=%s, got: %s", "clientid", got)
			}
			if got := req.PostFormValue("grant_type"); got != "password" {
				t.Fatalf("expected credential: grant_type=%s, got: %s", "password", got)
			}
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(
					fmt.Sprintf(`{"access_token":"%s"}`, requests[currentIndex].apiKey))),
				Header: make(http.Header)}
		case strings.Contains(u, `/query`):
			if got := requests[currentIndex].path; got != `/query` {
				t.Fatalf("request #%d was expected to be a query; got %q", currentIndex, got)
			}
			expected := fmt.Sprintf(
				"Bearer %s", requests[currentIndex].apiKey)
			if req.Header["Authorization"] == nil {
				t.Fatal("Authorization header missing")
			}
			if auth := req.Header["Authorization"][0]; auth != expected {
				t.Fatalf(`expected Authorization header %q, got %q, index %d`,
					expected, auth, currentIndex)
			}
			var jsonStr = `{"result":{}}`
			if requests[currentIndex].returnExpired {
				jsonStr = keyExpired
			}
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(jsonStr)),
				Header:     make(http.Header)}
		default:
			return nil
		}
	})

	client := newTestClient(testClient)

	for range []int{0, 1, 2} {
		req, err := client.NewRequest(
			mockEndpoint{PathFunc: func() string { return "/query" }},
			make(http.Header))
		if err != nil {
			t.Fatalf("no errors expected, got %s", err)
		}
		_, err = client.Call(req)
		if err != nil {
			t.Fatalf("no errors expected, got %s", err)
		}
	}
}

func TestResultsInErr15(t *testing.T) {
	expected := "Login failed"
	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(
				bytes.NewBufferString(
					`{"err":"` + expected + `","@attributes":{"err_code":15}}`)),
			Header: make(http.Header)}
	})

	client := newTestClient(testClient)
	req, _ := client.NewRequest(
		mockEndpoint{PathFunc: func() string { return "/query" }},
		make(http.Header))
	_, err := client.Call(req)

	if err == nil {
		t.Fatal("expected error")
	}
	switch err.(type) {
	case pargo.ErrLoginFailed:
	default:
		t.Fatal("expected type: ErrLoginFailed")
	}
	if err.Error() != expected {
		t.Fatalf("expected: %s, got: %s", expected, err.Error())
	}
}

func TestResultsInErr71(t *testing.T) {
	expected := "Input needs to be valid JSON or XML"
	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `oauth2/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{"api_key":"anyapikey"}`)),
				Header:     make(http.Header)}
		case strings.Contains(u, `/query`):
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(
					bytes.NewBufferString(`{"err":"` + expected + `","@attributes":{"err_code":71}}`)),
				Header: make(http.Header)}
		default:
			t.Fatal("endpoint not called")
			return nil
		}
	})

	client := newTestClient(testClient)
	req, _ := client.NewRequest(
		mockEndpoint{PathFunc: func() string { return "/query" }},
		make(http.Header))
	_, err := client.Call(req)

	if err == nil {
		t.Fatal("expected error")
	}
	switch err.(type) {
	case pargo.ErrInvalidJSON:
	default:
		t.Fatal("expected error of type ErrInvalidJSON")
	}
	if err.Error() != expected {
		t.Fatalf("expected: %s, got: %s", expected, err.Error())
	}
}

func TestFormatAllJSON(t *testing.T) {
	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {
		if got := req.FormValue("format"); got != "json" {
			t.Fatalf("expected query string format=%q, got: %q", "json", got)
		}
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
			Header:     make(http.Header)}
	})
	client := newTestClient(testClient)
	req, _ := client.NewRequest(
		mockEndpoint{
			MethodFunc: func() string { return "" },
			PathFunc:   func() string { return "" },
			ReadFunc:   func([]byte) error { return nil },
		}, make(http.Header))
	_, _ = client.Call(req)
}
