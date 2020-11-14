// Package pargo provides a Go client for the Pardot REST API.
package pargo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

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
	hostSalesforce = "apnic.my.salesforce.com"
	base           = "pi.pardot.com"
	version        = "version/4"
)

// Pargo is the state of a client.
type Pargo struct {
	client *http.Client // HTTP Client we delegate calls to.
	user   UserAccount  // Stored so the token can be refreshed as needed.

	apiKey string // Initially empty, refreshed by login.

	// apiKeyMu protects the api key.
	// It is quickly released after just some memory reads/writes.
	apiKeyMu sync.Mutex

	businessUnitId string // Introduced after SSO migration to Salesforce.
}

// UserAccount is the set of required credentials.
type UserAccount struct {
	ClientId     string // Client id used to login.
	ClientSecret string // Client secret used to login.
	Email        string // Email used as username to login.
	Pass         string // Password to login.
}

// NewPargo returns a pointer to a newly instantiated client.
func NewPargo(u UserAccount, businessUnitId string, confs ...func(*Pargo)) *Pargo {
	client := Pargo{
		client:         &http.Client{}, // Default client.
		user:           u,
		businessUnitId: businessUnitId,
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

func (p *Pargo) Call(req *http.Request) ([]byte, error) {
	if err := p.maybeAuth(); err != nil {
		return nil, err
	}

	req.Header = p.addAuthHeaders(req.Header)

	res, err := p.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "issuing request")
	}
	defer res.Body.Close()
	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading response bytes")
	}
	switch c := res.StatusCode; c {
	case 200, 201, 204:
	default:
		// For status codes not in the case above.
		return nil, errors.New(
			fmt.Sprintf(
				"got status code %d for %s",
				c, string(resBytes),
			),
		)
	}
	resBytes, err = p.parseRes(resBytes, req)
	if err != nil {
		return nil, err
	}
	return resBytes, nil
}

func (p *Pargo) parseRes(resBytes []byte, req *http.Request) ([]byte, error) {
	resBody := struct {
		Err  *string `json:"err,omitempty"`
		Attr *struct {
			ErrCode int `json:"err_code"`
		} `json:"@attributes,omitempty"`
	}{}
	err := json.Unmarshal(resBytes, &resBody)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling response")
	}
	if resBody.Err != nil {
		switch resBody.Attr.ErrCode {
		case 1:
			// API key expired so refresh key and try again with
			// the same body.
			p.apiKey = ""
			p.maybeAuth()
			return p.Call(req)
		case 15:
			return nil, ErrLoginFailed{*resBody.Err}
		case 71:
			return nil, ErrInvalidJSON{*resBody.Err}
		}
	}
	return resBytes, nil
}

func (p *Pargo) NewRequest(
	e Endpoint,
	header http.Header,
) (*http.Request, error) {

	header.Add("Content-Type", "application/x-www-form-urlencoded")
	req := http.Request{
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

	return &req, nil
}

func (p *Pargo) maybeAuth() error {
	// Synchronisation is needed so threads do not try to read the api
	// key while another thread is trying to write a new one.
	p.apiKeyMu.Lock()
	defer p.apiKeyMu.Unlock()

	if p.apiKey != "" {
		// Bails if we already have an api key.
		// Try and use the one we've got.
		return nil
	}

	headers := make(http.Header)
	headers.Add("Content-Type", "application/x-www-form-urlencoded")
	req := http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Scheme: "https",
			Host:   hostSalesforce,
			Path:   "/services/oauth2/token",
		},
		Header: headers,
	}

	q := req.URL.Query()
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()

	req.Body = ioutil.NopCloser(strings.NewReader(
		fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=%s&username=%s&password=%s",
			p.user.ClientId, p.user.ClientSecret,
			"password", p.user.Email, p.user.Pass)))
	res, err := p.client.Do(&req)
	if err != nil {
		return errors.Wrap(err, "issuing login request")
	}

	// At this point we have a body. Ensure it is closed before we return.
	defer res.Body.Close()

	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "reading login response bytes")
	}

	switch c := res.StatusCode; c {
	case 200, 201, 204:
	default:
		// Status codes not in the case above.
		return errors.New(
			fmt.Sprintf(
				"got status code %d with body %s",
				c, string(resBytes),
			),
		)
	}

	loginParsed := struct {
		Key string `json:"access_token"`
	}{}
	// Discard error and assume that the JSON from Pardot is valid.
	_ = json.Unmarshal(resBytes, &loginParsed)

	// Finally, store the key.
	// This is the schema of the response body:
	// {
	//    "access_token": "",
	//    "instance_url": "",
	//    "id": "",
	//    "token_type": "",
	//    "issued_at": "",
	//    "signature": ""
	// }

	p.apiKey = loginParsed.Key

	return nil
}

func (p *Pargo) addAuthHeaders(headers http.Header) http.Header {
	// Synchronisation is needed so threads do not try to read the api
	// key while another thread is trying to write a new one.
	p.apiKeyMu.Lock()
	defer p.apiKeyMu.Unlock()

	headers.Set("Authorization",
		fmt.Sprintf(
			"Bearer %s",
			p.apiKey,
		),
	)
	headers.Set("Pardot-Business-Unit-Id",
		fmt.Sprintf("%s", p.businessUnitId))
	return headers
}
