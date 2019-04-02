package pardotrest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// QueryProspects is an endpoint to query a page of prospects.
type QueryProspects struct {
	Offset, Limit int
	Fields        []string
	PlaceHolder   interface{}
}

func (QueryProspects) method() string {
	return http.MethodGet
}

func (QueryProspects) path() string {
	return "prospect/" + version + "/do/query"
}

func (q QueryProspects) query() (map[string]string, error) {
	query := make(map[string]string)
	query["offset"] = strconv.Itoa(q.Offset)
	query["limit"] = strconv.Itoa(q.Limit)
	query["fields"] = strings.Join(q.Fields, ",")
	return query, nil
}

func (q QueryProspects) read(res []byte) error {
	body := struct {
		Result struct {
			Prospect json.RawMessage `json:"prospect"`
		} `json:"result"`
	}{}
	// Discard error and assume that the JSON from Pardot is valid.
	_ = json.Unmarshal(res, &body)
	err := json.Unmarshal(body.Result.Prospect, q.PlaceHolder)
	if err != nil {
		return errors.Wrap(err, "unmarshaling prospects")
	}
	return nil
}
