package pargo_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	pardotrest "gitlab.xyz.apnic.net/go-pkg/pargo"
)

// The requirement is that an API Key should be requested if none is found,
// and then it should be reused until any authenticated request returns err code 1 from Pardot.
// When err code 1 occurs, a new API Key should be requested and then used, transparently for the client.
func TestReuseAPIKeyUntilExpired(t *testing.T) {
	requests := []struct {
		path, apiKey  string
		returnExpired bool
	}{
		{
			"login/",
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
			"login/",
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
	// It is used to validate the current state by comparison with the expected.
	currentIndex := 0

	testClient := pardotrest.NewTestHTTPClient(func(req *http.Request) *http.Response {
		defer func() { currentIndex++ }()
		u := req.URL.Path
		switch {
		case strings.Contains(u, `login/`):
			if requests[currentIndex].path != `login/` {
				t.Fatalf("request #%d was not expected to be a login", currentIndex)
			}
			if got := req.PostFormValue("email"); got != "a@b.com" {
				t.Fatalf("expected credential: email=%s, got: %s", "a@b.com", got)
			}
			if got := req.PostFormValue("password"); got != "pass" {
				t.Fatalf("expected credential: password=%s, got: %s", "pass", got)
			}
			if got := req.PostFormValue("user_key"); got != "clientkey" {
				t.Fatalf("expected credential: user_key=%s, got: %s", "clientkey", got)
			}
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(
					fmt.Sprintf(`{"api_key":"%s"}`, requests[currentIndex].apiKey))),
				Header: make(http.Header)}
		case strings.Contains(u, `/query`):
			if requests[currentIndex].path != `/query` {
				t.Fatalf("request #%d was not expected to be a query", currentIndex)
			}
			expected := fmt.Sprintf(
				"Pardot api_key=%s, user_key=%s", requests[currentIndex].apiKey, "clientkey")
			if auth := req.Header["Authorization"][0]; auth != expected {
				t.Fatalf(`expected Authorization header %s, got %s`, expected, auth)
			}
			var jsonStr string
			if requests[currentIndex].returnExpired {
				jsonStr = `{"err":"Invalid API key or user key","@attributes":{"err_code": 1}}`
			} else {
				jsonStr = `{"result":{}}`
			}
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(jsonStr)),
				Header:     make(http.Header)}
		default:
			return nil
		}
	})

	pardot := pardotrest.NewTestClient(testClient)

	for range []int{0, 1, 2} {
		err := pardot.Call(pardotrest.NopEndpoint{P: "/query"})
		if err != nil {
			t.Fatalf("no errors expected, got %s", err)
		}
	}
}

func TestResultsInErr15(t *testing.T) {
	expected := "Login failed"
	testClient := pardotrest.NewTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `login/`):
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(
					bytes.NewBufferString(`{"err":"` + expected + `","@attributes":{"err_code":15}}`)),
				Header: make(http.Header)}
		default:
			t.Fatal("endpoint not called")
			return nil
		}
	})

	pardot := pardotrest.NewTestClient(testClient)
	err := pardot.Call(pardotrest.NopEndpoint{})

	if err == nil {
		t.Fatal("expected error")
	}
	switch err.(type) {
	case pardotrest.ErrLoginFailed:
	default:
		t.Fatal("expected type: ErrLoginFailed")
	}
	if err.Error() != expected {
		t.Fatalf("expected: %s, got: %s", expected, err.Error())
	}
}

func TestResultsInErr71(t *testing.T) {
	expected := "Input needs to be valid JSON or XML"
	testClient := pardotrest.NewTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `login/`):
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

	pardot := pardotrest.NewTestClient(testClient)
	err := pardot.Call(pardotrest.NopEndpoint{P: "/query"})

	if err == nil {
		t.Fatal("expected error")
	}
	switch err.(type) {
	case pardotrest.ErrInvalidJSON:
	default:
		t.Fatal("expected error of type ErrInvalidJSON")
	}
	if err.Error() != expected {
		t.Fatalf("expected: %s, got: %s", expected, err.Error())
	}
}

func TestFormatAllJSON(t *testing.T) {
	testClient := pardotrest.NewTestHTTPClient(func(req *http.Request) *http.Response {
		if got := req.FormValue("format"); got != "json" {
			t.Fatalf("expected query string format=%s, got: %s", "json", got)
		}
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
			Header:     make(http.Header)}
	})
	pardot := pardotrest.NewTestClient(testClient)
	_ = pardot.Call(pardotrest.NopEndpoint{})
}
