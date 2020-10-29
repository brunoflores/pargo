package pargo

import (
	"encoding/json"
	"net/http"
	"strings"
)

// BatchUpdateProspect is an endpoint to update a batch of prospects.
type BatchUpdateProspect struct {
	Prospects interface{}
}

// BatchUpdateProspects executes the endpoint with arguments.
func (p *Pargo) BatchUpdateProspects(args BatchUpdateProspect) error {
	headers := make(http.Header)
	req, err := p.NewRequest(args, headers)
	if err != nil {
		return err
	}
	body, err := p.Call(req)
	if err != nil {
		return err
	}
	err = readBatchUpdateProspect(body)
	if err != nil {
		return err
	}
	return nil
}

// BatchUpdateProspectErrors is the result from a batch operation when
// errors were encountered.
//
// NOTE from http://developer.pardot.com: If any errors are found during the
// batch process, an error array will be returned for only the prospects
// with issues. The error array will be key/value pairs where the key is
// the index of the prospect submitted in the request. All other prospects
// will be processed as expected.
type BatchUpdateProspectErrors struct {
	Errors map[int]string
}

func (b BatchUpdateProspectErrors) Error() string {
	var concat []string
	for _, v := range b.Errors {
		concat = append(concat, v)
	}
	return strings.Join(concat, ", ")
}

func (q BatchUpdateProspect) Method() string {
	return http.MethodPost
}

func (q BatchUpdateProspect) Path() string {
	return "prospect/" + version + "/do/batchUpdate"
}

func (q BatchUpdateProspect) Query() (map[string]string, error) {
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

func readBatchUpdateProspect(res []byte) error {
	body := struct {
		Errors *map[int]string `json:"errors,string,omitempty"`
	}{}
	// Discard error and assume the JSON from Pardot is unmarshable.
	_ = json.Unmarshal(res, &body)
	if body.Errors != nil {
		result := BatchUpdateProspectErrors{make(map[int]string)}
		for k, v := range *body.Errors {
			result.Errors[k] = v
		}
		if len(result.Errors) > 0 {
			return result
		}
	}
	return nil
}
