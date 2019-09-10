// Package pargo provides a Go client for the Pardot REST API.
package pargo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"sync"

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
// See http://developer.pardot.com/kb/error-codes-messages.
type ErrInvalidJSON struct {
	msg string
}

func (e ErrInvalidJSON) Error() string {
	return e.msg
}

const (
	base    = "pi.pardot.com"
	version = "version/4"
)

// Pargo is the state of a client.
type Pargo struct {
	client *http.Client // HTTP Client we delegate calls to.
	user   UserAccount  // Stored so the token can be refreshed as needed.

	apiKey   string // Initially empty, refreshed by login.
	apiKeyMu sync.Mutex
}

// UserAccount is the set of required credentials.
type UserAccount struct {
	UserKey string // Client key used for login.
	Email   string // Email used as username for login.
	Pass    string // Password for login.
}

// NewPargo returns a pointer to a newly instantiated client.
func NewPargo(u UserAccount, confs ...func(*Pargo)) *Pargo {
	client := Pargo{
		client: &http.Client{}, // Default client.
		user:   u,
	}
	for _, conf := range confs {
		conf(&client)
	}
	return &client
}

// WithCustomClient sets a custom http.Client.
func WithCustomClient(c *http.Client) func(*Pargo) {
	return func(client *Pargo) {
		client.client = c
	}
}

// Endpoint is the behaviour required for an endpoint.
type Endpoint interface {
	Method() string
	Path() string
	Read([]byte) error
}

// EndpointBody is an endpoint with a body.
type EndpointBody interface {
	Endpoint
	Body() (io.ReadCloser, error)
}

// EndpointQuery is an endpoint with query strings.
type endpointQuery interface {
	Endpoint
	Query() (map[string]string, error)
}

func (p *Pargo) Call(e Endpoint) error {
	header := make(http.Header)
	_, isLogin := e.(*Login)
	if isLogin == false {
		if err := p.maybeAuth(); err != nil {
			return err
		}
		header.Add("Authorization",
			fmt.Sprintf(
				"Pardot api_key=%s, user_key=%s",
				p.apiKey, p.user.UserKey,
			),
		)
	}
	req, err := p.NewRequest(e, header)
	if err != nil {
		return errors.Wrap(err, "building request")
	}
	res, err := p.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "issuing request")
	}
	defer res.Body.Close()
	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "reading response bytes")
	}
	switch c := res.StatusCode; c {
	case 200, 201, 204:
	default:
		// For status codes not in the case above.
		return errors.New(
			fmt.Sprintf(
				"got status code %d for %s",
				c, string(resBytes),
			),
		)
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
		case 1:
			// API key expired so refresh key and try again.
			p.apiKey = ""
			return p.Call(e)
		case 15:
			return ErrLoginFailed{*resBody.Err}
		case 71:
			return ErrInvalidJSON{*resBody.Err}
		}
	}
	return e.Read(resBytes)
}

func (p *Pargo) NewRequest(
	e Endpoint,
	header http.Header,
) (*http.Request, error) {

	header.Add("Content-Type", "application/x-www-form-urlencoded")
	req := &http.Request{
		Method: e.Method(),
		URL: &url.URL{
			Scheme: "https",
			Host:   base,
			Path:   "/api/" + e.Path(),
		},
		Header: header,
	}

	if _, ok := e.(EndpointBody); ok {
		body, err := e.(EndpointBody).Body()
		if err != nil {
			return nil, err
		}
		req.Body = body
	}

	q := req.URL.Query()
	q.Add("format", "json")
	if _, ok := e.(endpointQuery); ok {
		query, err := e.(endpointQuery).Query()
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
	p.apiKeyMu.Lock()
	defer p.apiKeyMu.Unlock()
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
