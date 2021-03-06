package pargo_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/brunoflores/pargo"
)

func TestReadsAList(t *testing.T) {
	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `oauth2/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header)}
		case strings.Contains(u, `listMembership/`):
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(
					`{"result":{"total_results": 2,"list_membership":[{"list_id": 24323,"prospect_id": 7666184},{"list_id": 24323,"prospect_id": 8058232}]}}`)),
				Header: make(http.Header)}
		default:
			t.Fatal("no endpoint called")
			return nil
		}
	})

	pardot := newTestClient(testClient)
	var memberships []pargo.ListMembership
	err := pardot.ListMemberships(pargo.ListMemberships{
		Offset:      100,
		Limit:       200,
		ListID:      24323,
		Placeholder: &memberships,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(memberships) != 2 {
		t.Fatalf("expected %d memberships, got %d", 2, len(memberships))
	}
}

func TestReadsASingle(t *testing.T) {
	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `oauth2/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header)}
		case strings.Contains(u, `listMembership/`):
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(
					`{"result":{"total_results": 1,"list_membership":{"list_id": 24323,"prospect_id": 7666184}}}`)),
				Header: make(http.Header)}
		default:
			t.Fatal("no endpoint called")
			return nil
		}
	})

	pardot := newTestClient(testClient)
	var memberships []pargo.ListMembership
	err := pardot.ListMemberships(pargo.ListMemberships{
		Offset:      100,
		Limit:       200,
		ListID:      24323,
		Placeholder: &memberships,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(memberships) != 1 {
		t.Fatalf("expected 1 membership, got %d", len(memberships))
	}
}

func TestReadsEmptyPage(t *testing.T) {
	testClient := newTestHTTPClient(func(req *http.Request) *http.Response {
		u := req.URL.Path
		switch {
		case strings.Contains(u, `oauth2/`):
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header)}
		case strings.Contains(u, `listMembership/`):
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewBufferString(
					`{"result":{"total_results": 2}}`)),
				Header: make(http.Header)}
		default:
			t.Fatal("no endpoint called")
			return nil
		}
	})

	pardot := newTestClient(testClient)
	var memberships []pargo.ListMembership
	err := pardot.ListMemberships(pargo.ListMemberships{
		Offset:      100,
		Limit:       200,
		ListID:      24323,
		Placeholder: &memberships,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(memberships) != 0 {
		t.Fatalf("expected 0 memberships, got %d", len(memberships))
	}
}
