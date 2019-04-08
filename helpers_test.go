package pargo

import (
	"net/http"
)

func NewTestClient(c *http.Client) *Pargo {
	return NewPargo(
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

type MockEndpoint struct {
	MethodFunc func() string
	PathFunc   func() string
	ReadFunc   func([]byte) error
}

func (p *Pargo) MockEndpoint(args MockEndpoint) error {
	return p.call(args)
}

func (e MockEndpoint) method() string {
	if e.MethodFunc == nil {
		return http.MethodGet
	}
	return e.MethodFunc()
}

func (e MockEndpoint) path() string {
	return e.PathFunc()
}

func (e MockEndpoint) read(r []byte) error {
	if e.ReadFunc == nil {
		return nil
	}
	return e.ReadFunc(r)
}
