package pardot

import (
	"errors"
	"fmt"

	"gopkg.in/resty.v1"
)

const (
	base                   = "https://pi.pardot.com/api"
	version                = "version/4"
	prospectQueryEnd       = base + "/prospect/" + version + "/do/query"
	prospectBatchUpdateEnd = base + "/prospect/" + version + "/do/batchUpdate"
	loginEnd               = base + "/login/" + version
)

// Pardot is a connection to Pardot.
type Pardot struct {
	userKey string
	apiKey  string
}

// NewPardot returns a pointer to an uninitialised connection.
func NewPardot(key string) *Pardot {
	return &Pardot{userKey: key}
}

// Auth tries to connect given credentials and might return an error.
func (p *Pardot) Auth(user, pass string) error {
	type AuthResponse struct {
		APIKey string `json:"api_key,omitempty"`
		Error  string `json:"err,omitempty"`
	}
	resp, err := resty.R().
		SetHeader("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW").
		SetBody(fmt.Sprintf("------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"email\"\r\n\r\n%s\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"password\"\r\n\r\n%s\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"user_key\"\r\n\r\n%s\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"format\"\r\n\r\njson\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW--",
			user, pass, p.userKey)).
		SetResult(&AuthResponse{}).
		Post(loginEnd)
	if err != nil {
		return err
	}
	auth := resp.Result().(*AuthResponse)
	if auth.Error != "" {
		return errors.New(auth.Error)
	}
	p.apiKey = auth.APIKey
	return nil
}
