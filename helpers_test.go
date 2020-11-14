package pargo_test

import (
	"net/http"

	"github.com/brunoflores/pargo"
)

func newTestClient(c *http.Client) *pargo.Pargo {
	return pargo.NewPargo(
		pargo.UserAccount{
			Email:        "a@b.com",
			Pass:         "pass",
			ClientId:     "clientid",
			ClientSecret: "clientsecret",
		},
		"somebusinessunitid",
		pargo.WithCustomClient(c),
	)
}

func newTestHTTPClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn),
	}
}

type mockEndpoint struct {
	MethodFunc func() string
	PathFunc   func() string
	ReadFunc   func([]byte) error
}

func (e mockEndpoint) Method() string {
	if e.MethodFunc == nil {
		return http.MethodGet
	}
	return e.MethodFunc()
}

func (e mockEndpoint) Path() string {
	return e.PathFunc()
}

func (e mockEndpoint) Read(r []byte) error {
	if e.ReadFunc == nil {
		return nil
	}
	return e.ReadFunc(r)
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}
