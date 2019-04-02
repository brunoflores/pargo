package pardotrest

import (
	"net/http"
)

func NewTestClient(c *http.Client) *PardotREST {
	return NewPardotREST(
		UserAccount{
			Email:   "a@b.com",
			Pass:    "pass",
			UserKey: "clientkey",
		},
	).WithCustomClient(c)
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestHTTPClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn),
	}
}

type NopEndpoint struct {
	M, P string
}

func (e NopEndpoint) method() string {
	return e.M
}

func (e NopEndpoint) path() string {
	return e.P
}

func (NopEndpoint) read(r []byte) error {
	return nil
}
