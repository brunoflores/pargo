package pardotrest

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// ErrLoginFailed is the error code 15 in Pardot.
// See http://developer.pardot.com/kb/error-codes-messages.
type ErrLoginFailed struct {
	msg string
}

func (e ErrLoginFailed) Error() string {
	return e.msg
}

const (
	base    = "https://pi.pardot.com/api"
	version = "version/4"
)

// Endpoint is the behaviour required for an endpoint in the REST API.
type Endpoint interface {
	method() string
	path() string
	body() (io.ReadCloser, error)
	query() (map[string][]byte, error)
	read([]byte) error
}

// PardotREST is a client of the Pardot REST API.
type PardotREST struct {
	client  *http.Client
	apiKey  string
	userKey string
	email   string
	pass    string
}

// NewPardotREST returns a pointer to the REST client.
func NewPardotREST() *PardotREST {
	return &PardotREST{client: &http.Client{}}
}

// WithCustomClient sets a custom http.Client.
func (p *PardotREST) WithCustomClient(c *http.Client) *PardotREST {
	p.client = c
	return p
}

// WithPardotUserAccount configures a given user account.
func (p *PardotREST) WithPardotUserAccount(email, pass, key string) *PardotREST {
	p.email = email
	p.pass = pass
	p.userKey = key
	return p
}

// Call makes a call to the REST API and returns an error.
func (p *PardotREST) Call(e Endpoint) error {
	if err := p.maybeAuth(); err != nil {
		return err
	}
	header := make(http.Header)
	header.Add("Authorization", fmt.Sprintf("Pardot api_key=%s, user_key=%s", p.apiKey, p.userKey))
	body, err := e.body()
	if err != nil {
		return err
	}
	query, err := e.query()
	if err != nil {
		return err
	}
	req := p.newRequest(e.method(), e.path(), body, query, header)
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
		case 1:
			p.apiKey = ""
			return p.Call(e)
		default:
			return errors.New(*resBody.Err)
		}
	}
	e.read(resBytes)
	return nil
}

func (p *PardotREST) newRequest(method, path string, body io.ReadCloser, query map[string][]byte, header http.Header) *http.Request {
	u := &url.URL{
		Host: base,
		Path: path,
	}
	q := u.Query()
	for k, v := range query {
		q.Add(k, string(v))
	}
	u.RawQuery = q.Encode()
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	req := &http.Request{
		Method:     method,
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     header,
		Host:       u.Host,
		Body:       body,
	}
	return req
}

func (p *PardotREST) maybeAuth() error {
	if p.apiKey != "" {
		return nil // bails if we already have an api key.
	}
	const loginPath = "login/" + version
	body := ioutil.NopCloser(strings.NewReader(fmt.Sprintf("email=%s&password=%s&user_key=%s", p.email, p.pass, p.userKey)))
	header := make(http.Header)
	req := p.newRequest("POST", loginPath, body, make(map[string][]byte), header)
	res, err := p.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "issuing get request")
	}
	defer res.Body.Close()
	if c := res.StatusCode; c != 200 {
		return errors.New(fmt.Sprintf("status code %d", c))
	}
	resB, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "reading response bytes")
	}
	loginRes := struct {
		Key  string  `json:"api_key,omitempty"`
		Err  *string `json:"err,omitempty"`
		Attr *struct {
			ErrCode int `json:"err_code"`
		} `json:"@attributes,omitempty"`
	}{}
	err = json.Unmarshal(resB, &loginRes)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unmarshaling auth result: %s", string(resB)))
	}
	if loginRes.Err != nil {
		switch loginRes.Attr.ErrCode {
		case 15:
			return ErrLoginFailed{*loginRes.Err}
		default:
			return errors.New(*loginRes.Err)
		}
	}
	p.apiKey = loginRes.Key
	return nil
}
