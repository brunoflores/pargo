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

// BatchUpdateProspectErrors is the result from a batch operation when errors were encountered.
//
// NOTE from http://developer.pardot.com: If any errors are found during the batch process,
// an error array will be returned for only the prospects with issues. The error array
// will be key/value pairs where the key is the index of the prospect submitted in the request.
// All other prospects will be processed as expected.
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

func (q BatchUpdateProspect) method() string {
	return http.MethodPost
}

func (q BatchUpdateProspect) path() string {
	return "prospect/" + version + "/do/batchUpdate"
}

func (q BatchUpdateProspect) query() (map[string]string, error) {
	query := make(map[string]string)
	b, err := json.Marshal(q.Prospects)
	if err != nil {
		return nil, err
	}
	query["prospects"] = string(b)
	return query, nil
}

func (q BatchUpdateProspect) read(res []byte) error {
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
		return result
	}
	return nil
}
