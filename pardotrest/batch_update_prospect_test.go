package pardotrest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
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

	testClient := newTestClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `login/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{"api_key":"anyapikey"}`)),
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

	pardot := NewPardotREST().WithCustomClient(testClient)
	err := pardot.Call(BatchUpdateProspect{
		Prospects: &prospects,
	})
	if err != nil {
		t.Fatal(err)
	}
}
