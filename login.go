package pargo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// Login is used to request an api key given credentials.
type Login struct {
	userKey, email, pass, apiKey string
}

func (*Login) Method() string {
	return http.MethodPost
}

func (*Login) Path() string {
	return "login/" + version
}

func (l *Login) Read(res []byte) error {
	loginRes := struct {
		Key string `json:"api_key"`
	}{}
	// Discard error and assume that the JSON from Pardot is valid.
	_ = json.Unmarshal(res, &loginRes)
	l.apiKey = loginRes.Key
	return nil
}

func (l *Login) Body() (io.ReadCloser, error) {
	return ioutil.NopCloser(
			strings.NewReader(
				fmt.Sprintf("email=%s&password=%s&user_key=%s",
					l.email, l.pass, l.userKey))),
		nil
}
