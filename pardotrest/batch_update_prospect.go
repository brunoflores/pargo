package pardotrest

import (
	"encoding/json"
	"io"
	"net/http"
)

// BatchUpdateProspect updates a batch of prospects or returns a fatal error.
// If `fatal` is non-nil, the complete batch failled.
// If `fatal` is nil, `errors` might contain errors where the key is the index of the prospect
// that caused it.
//
// NOTE from http://developer.pardot.com: If any errors are found during the batch process,
// an error array will be returned for only the prospects with issues. The error array
// will be key/value pairs where the key is the index of the prospect submitted in the request.
// All other prospects will be processed as expected.
type BatchUpdateProspect struct {
	Prospects interface{}
}

func (q BatchUpdateProspect) method() string {
	return http.MethodPost
}

func (q BatchUpdateProspect) path() string {
	return "prospect/" + version + "/do/batchUpdate"
}

func (q BatchUpdateProspect) query() (map[string][]byte, error) {
	query := make(map[string][]byte)
	b, err := json.Marshal(q.Prospects)
	if err != nil {
		return nil, err
	}
	query["prospects"] = b
	return query, nil
}

func (q BatchUpdateProspect) body() (io.ReadCloser, error) {
	return nil, nil
}

func (q BatchUpdateProspect) read(res []byte) error {
	return nil
}
