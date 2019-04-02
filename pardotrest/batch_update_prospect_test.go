package pardotrest_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"gitlab.xyz.apnic.net/go-pkg/pardot/pardotrest"
)

func TestBatchUpdateProspects(t *testing.T) {
	type prospect struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}

	prospects := []prospect{
		prospect{10, "a@a.com"},
		prospect{20, "b@b.com"},
	}

	testClient := pardotrest.NewTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `login/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header)}
		case strings.Contains(u, `/batchUpdate`):
			expected := `[{"id":10,"email":"a@a.com"},{"id":20,"email":"b@b.com"}]`
			if got := req.FormValue("prospects"); got != expected {
				t.Fatalf("expected: %s, got: %s", expected, got)
			}
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header)}
		default:
			t.Fatal("endpoint not called")
			return nil
		}
	})

	pardot := pardotrest.NewTestClient(testClient)
	err := pardot.Call(pardotrest.BatchUpdateProspect{
		Prospects: &prospects,
	})
	if err != nil {
		t.Fatalf("expected no errors, got %s", err)
	}
}

func TestBatchUpdateProspectsReturnsErrors(t *testing.T) {
	testClient := pardotrest.NewTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `login/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{"api_key":"anyapikey"}`)),
				Header:     make(http.Header)}
		case strings.Contains(u, `/batchUpdate`):
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(`
				{"errors":
					{
						"0":"Invalid prospect email address",
						"1":"Invalid prospect"
					}
				}`)),
				Header: make(http.Header)}
		default:
			t.Fatal("endpoint not called")
			return nil
		}
	})

	prospects := []struct{}{}
	pardot := pardotrest.NewTestClient(testClient)
	err := pardot.Call(pardotrest.BatchUpdateProspect{
		Prospects: &prospects,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	switch err.(type) {
	case pardotrest.BatchUpdateProspectErrors:
	default:
		t.Fatal("expected error of type BatchUpdateProspectErrors")
	}
	if err.Error() == "" {
		t.Fatal("expected a non-empty error message")
	}
	var expected string
	for index, msg := range err.(pardotrest.BatchUpdateProspectErrors).Errors {
		switch index {
		case 0:
			expected = "Invalid prospect email address"
		case 1:
			expected = "Invalid prospect"
		}
		if msg != expected {
			t.Fatalf("expected: %s, got: %s", expected, msg)
		}
	}
}
