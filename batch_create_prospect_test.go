package pargo_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"gitlab.xyz.apnic.net/go-pkg/pargo"
)

func TestBatchCreateProspects(t *testing.T) {
	type prospect struct {
		Email string `json:"email"`
	}
	prospects := []prospect{
		prospect{"a@a.com"},
		prospect{"b@b.com"},
	}
	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `oauth2/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header),
			}
		case strings.Contains(u, `/batchCreate`):
			expected := `{"prospects":[{"email":"a@a.com"},{"email":"b@b.com"}]}`
			if got := req.FormValue("prospects"); got != expected {
				t.Fatalf("expected: %s, got: %s", expected, got)
			}
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header),
			}
		default:
			t.Fatal("no endpoint called")
			return nil
		}
	})
	pardot := newTestClient(testClient)
	err := pardot.BatchCreateProspects(pargo.BatchCreateProspect{
		Prospects: &prospects,
	})
	if err != nil {
		t.Fatalf("expected no errors, got %s", err)
	}
}

func TestBatchCreateProspectsReturnsErrors(t *testing.T) {
	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `oauth2/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{"api_key":"anyapikey"}`)),
				Header:     make(http.Header)}
		case strings.Contains(u, `/batchCreate`):
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
			t.Fatal("no endpoint called")
			return nil
		}
	})
	prospects := []struct{}{}
	pardot := newTestClient(testClient)
	err := pardot.BatchCreateProspects(pargo.BatchCreateProspect{
		Prospects: &prospects,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	switch err.(type) {
	case pargo.BatchCreateProspectErrors:
	default:
		t.Fatalf("expected error of type BatchCreateProspectErrors")
	}
	if err.Error() == "" {
		t.Fatal("expected a non-empty error message")
	}
	var expected string
	for index, msg := range err.(pargo.BatchCreateProspectErrors).Errors {
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
