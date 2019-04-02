package pardotrest

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
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

	testClient := newTestClient(func(req *http.Request) *http.Response {
		defer func() { currentIndex++ }()
		u := req.URL.Path
		switch {
		case strings.Contains(u, `login/`):
			if requests[currentIndex].path != `login/` {
				t.Fatalf("request #%d was not expected to be a login", currentIndex)
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
				jsonStr = `{"@attributes":{"err_code": 1}}`
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

	pardot := NewPardotREST().
		WithCustomClient(testClient).
		WithPardotUserAccount("a@b.com", "secretpass", "clientkey")

	for range []int{0, 1, 2} {
		err := pardot.Call(NopRequest{})
		if err != nil {
			t.Fatalf("no errors expected, got %s", err)
		}
	}
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn),
	}
}

type NopRequest struct{}

func (NopRequest) method() string {
	return "GET"
}

func (NopRequest) path() string {
	return "/query"
}

func (NopRequest) query() (map[string][]byte, error) {
	return nil, nil
}

func (NopRequest) body() (io.ReadCloser, error) {
	return nil, nil
}

func (NopRequest) read(r []byte) error {
	return nil
}
