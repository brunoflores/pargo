package pargo_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"gitlab.xyz.apnic.net/go-pkg/pargo"
)

func TestDeleteProspect(t *testing.T) {
	var got string
	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `login/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header)}
		case strings.Contains(u, `/delete`):
			got = u
			return &http.Response{
				StatusCode: 204,
				Body:       ioutil.NopCloser(bytes.NewBufferString("")),
				Header:     make(http.Header)}
		default:
			t.Fatal("unknown endpoint called")
			return nil
		}
	})
	want := "/api/prospect/version/4/do/delete/id/46"
	client := newTestClient(testClient)
	client.DeleteProspect(pargo.DeleteProspect{
		ProspectID: 46,
	})
	if got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
}
