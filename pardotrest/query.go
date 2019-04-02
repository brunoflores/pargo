package pardotrest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// QueryProspects queries a page of prospects from Pardot.
type QueryProspects struct {
	Offset, Limit int
	Fields        []string
	PlaceHolder   interface{}
}

func (q QueryProspects) method() string {
	return http.MethodGet
}

func (q QueryProspects) path() string {
	return "prospect/" + version + "/do/query"
}

func (q QueryProspects) query() (map[string][]byte, error) {
	return nil, nil
}

func (q QueryProspects) body() (io.ReadCloser, error) {
	return nil, nil
}

func (q QueryProspects) read(res []byte) error {
	body := struct {
		Result struct {
			Prospect json.RawMessage `json:"prospect"`
		} `json:"result,omitempty"`
	}{}
	err := json.Unmarshal(res, &body)
	if err != nil {
		return errors.Wrap(err, "unmarshaling bytes")
	}
	err = json.Unmarshal(body.Result.Prospect, q.PlaceHolder)
	if err != nil {
		return errors.Wrap(err, "unmarshaling prospects")
	}
	return nil
}
