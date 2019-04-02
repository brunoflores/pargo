package pargo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// ErrLoginFailed is the error code 15 in Pardot.
// It implements `error`.
// See http://developer.pardot.com/kb/error-codes-messages.
type ErrLoginFailed struct {
	msg string
}

func (e ErrLoginFailed) Error() string {
	return e.msg
}

// ErrInvalidJSON is the error code 71 in Pardot.
// It implements `error`.
type ErrInvalidJSON struct {
	msg string
}

func (e ErrInvalidJSON) Error() string {
	return e.msg
}

const (
	base    = "https://pi.pardot.com/api"
	version = "version/4"
)

// Endpoint is the behaviour required for an endpoint.
type Endpoint interface {
	method() string
	path() string
	read([]byte) error
}

// EndpointBody is an endpoint with a body.
type EndpointBody interface {
	Endpoint
	body() (io.ReadCloser, error)
}

// EndpointQuery is an endpoint with query strings.
type EndpointQuery interface {
	Endpoint
	query() (map[string]string, error)
}

// Pargo is a client of the Pardot REST API.
type Pargo struct {
	client *http.Client // HTTP Client we delegate calls to.
	apiKey string       // Initially empty, refreshed by login.
	user   UserAccount  // Credentials.
}

// UserAccount is the set of required credentials.
type UserAccount struct {
	UserKey string // Client key used for login.
	Email   string // Email used as username for login.
	Pass    string // Password for login.
}

// NewPargo returns a pointer to an instance of ParGo.
func NewPargo(u UserAccount) *Pargo {
	return &Pargo{
		client: &http.Client{},
		user:   u,
	}
}

// WithCustomClient sets a custom http.Client.
// Otherwise, a default client is used.
func (p *Pargo) WithCustomClient(c *http.Client) *Pargo {
	p.client = c
	return p
}

// Call makes a call to the REST API and returns an error.
func (p *Pargo) Call(e Endpoint) error {
	header := make(http.Header)
	_, isLogin := e.(*Login)
	if isLogin == false {
		if err := p.maybeAuth(); err != nil {
			return err
		}
		header.Add("Authorization", fmt.Sprintf("Pardot api_key=%s, user_key=%s", p.apiKey, p.user.UserKey))
	}
	req, err := p.newRequest(e, header)
	if err != nil {
		return errors.Wrap(err, "building request")
	}
	res, err := p.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "issuing request")
	}
	defer res.Body.Close()
	if c := res.StatusCode; c != 200 {
		return errors.New(fmt.Sprintf("status code %d", c))
	}
	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "reading response bytes")
	}
	resBody := struct {
		Err  *string `json:"err,omitempty"`
		Attr *struct {
			ErrCode int `json:"err_code"`
		} `json:"@attributes,omitempty"`
	}{}
	err = json.Unmarshal(resBytes, &resBody)
	if err != nil {
		return errors.Wrap(err, "unmarshaling response")
	}
	if resBody.Err != nil {
		switch resBody.Attr.ErrCode {
		case 1: // API key expired so refresh key and try again.
			p.apiKey = ""
			return p.Call(e)
		case 15:
			return ErrLoginFailed{*resBody.Err}
		case 71:
			return ErrInvalidJSON{*resBody.Err}
		}
	}
	return e.read(resBytes)
}

func (p *Pargo) newRequest(e Endpoint, header http.Header) (*http.Request, error) {
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	req := &http.Request{
		Method: e.method(),
		URL: &url.URL{
			Host: base,
			Path: e.path(),
		},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     header,
		Host:       base,
	}

	if _, ok := e.(EndpointBody); ok {
		body, err := e.(EndpointBody).body()
		if err != nil {
			return nil, err
		}
		req.Body = body
	}

	q := req.URL.Query()
	q.Add("format", "json")
	if _, ok := e.(EndpointQuery); ok {
		query, err := e.(EndpointQuery).query()
		if err != nil {
			return nil, err
		}
		for k, v := range query {
			q.Add(k, string(v))
		}
	}
	req.URL.RawQuery = q.Encode()

	return req, nil
}

func (p *Pargo) maybeAuth() error {
	if p.apiKey != "" {
		// Bails if we already have an api key.
		// Try and use the one we've got.
		return nil
	}
	req := Login{
		userKey: p.user.UserKey,
		email:   p.user.Email,
		pass:    p.user.Pass,
	}
	err := p.Call(&req)
	if err != nil {
		return err
	}
	p.apiKey = req.apiKey
	return nil
}