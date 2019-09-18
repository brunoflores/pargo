package pargo

import (
	"encoding/json"
	"net/http"
	"strings"
)

// BatchCreateProspect is an endpoint to create a batch of prospects.
type BatchCreateProspect struct {
	Prospects interface{}
}

// BatchCreateProspects executes the endpoint with arguments.
func (p *Pargo) BatchCreateProspects(args BatchCreateProspect) error {
	return p.Call(args)
}

// BatchCreateProspectErrors is the result from a batch operation when errors
// were encountered.
//
// NOTE from http://developer.pardot.com: If any errors are found during
// the batch process, an error array will be returned for only the prospects
// with issues. The error array will be key/value pairs where the key is the
// index of the prospect submitted in the request. All other prospects will
// be processed as expected.
type BatchCreateProspectErrors struct {
	Errors map[int]string
}

func (b BatchCreateProspectErrors) Error() string {
	var concat []string
	for _, v := range b.Errors {
		concat = append(concat, v)
	}
	return strings.Join(concat, ", ")
}

func (q BatchCreateProspect) Method() string {
	return http.MethodPost
}

func (q BatchCreateProspect) Path() string {
	return "prospect/" + version + "/do/batchCreate"
}

func (q BatchCreateProspect) Query() (map[string]string, error) {
	query := make(map[string]string)
	type wrap struct {
		Prospects interface{} `json:"prospects"`
	}
	w := wrap{q.Prospects}
	b, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}
	query["prospects"] = string(b)
	return query, nil
}

func (q BatchCreateProspect) Read(res []byte) error {
	body := struct {
		Errors *map[int]string `json:"errors,string,omitempty"`
	}{}
	// Discard error and assume the JSON from Pardot is valid.
	_ = json.Unmarshal(res, &body)
	if body.Errors != nil {
		result := BatchCreateProspectErrors{make(map[int]string)}
		for k, v := range *body.Errors {
			result.Errors[k] = v
		}
		if len(result.Errors) > 0 {
			return result
		}
	}
	return nil
}
