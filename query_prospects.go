package pargo

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type QueryProspectsEOF struct{}

func (QueryProspectsEOF) Error() string {
	return "Empty page."
}

// QueryProspects is an endpoint to query a page of prospects.
type QueryProspects struct {
	// Required fields.
	Offset, Limit int
	Fields        []string

	// Optional fields.
	PlaceHolder interface{}
	Marshaler   func(json.RawMessage)
}

// QueryProspects executes the endpoint with arguments.
func (p *Pargo) QueryProspects(args QueryProspects) error {
	headers := make(http.Header)
	req, err := p.NewRequest(args, headers)
	if err != nil {
		return errors.Wrap(err, "building request")
	}
	body, err := p.Call(req)
	if err != nil {
		return errors.Wrap(err, "requesting")
	}
	err = args.readQueryProspects(body)
	if err != nil {
		return err
	}
	return nil
}

func (QueryProspects) Method() string {
	return http.MethodGet
}

func (QueryProspects) Path() string {
	return "prospect/" + version + "/do/query"
}

func (q QueryProspects) Query() (map[string]string, error) {
	return map[string]string{
		"offset": strconv.Itoa(q.Offset),
		"limit":  strconv.Itoa(q.Limit),
		"fields": strings.Join(q.Fields, ","),
	}, nil
}

func (q QueryProspects) readQueryProspects(res []byte) error {
	body := struct {
		Result struct {
			Prospect json.RawMessage `json:"prospect"`
		} `json:"result"`
	}{}

	// Discard error and assume that the JSON from Pardot is valid.
	err := json.Unmarshal(res, &body)
	if err != nil {
		return errors.Wrap(err, "got invalid JSON from Pardot")
	}

	// Got an empty page.
	// Pardot does not tell how many records were returned, the only
	// indication of an empty page is that the array of prospects in the
	// key `prospect` is not present.
	if body.Result.Prospect == nil {
		return QueryProspectsEOF{}
	}

	if q.Marshaler != nil {
		q.Marshaler(body.Result.Prospect)
		return nil
	}

	err = json.Unmarshal(body.Result.Prospect, q.PlaceHolder)
	if err != nil {
		return errors.Wrap(err, "unmarshaling prospects")
	}
	return nil
}
