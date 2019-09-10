package pargo

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
)

// ListMemberships is an endpoint to query prospects subscribed to a list id.
type ListMemberships struct {
	ListID        int
	Offset, Limit int
	Placeholder   *[]ListMembership
}

// ListMembership is an instance of a list membership.
// A link between a prospect and a list.
type ListMembership struct {
	ListID     int `json:"list_id"`
	ProspectID int `json:"prospect_id"`
}

// ListMemberships executes the endpoint with arguments.
func (p *Pargo) ListMemberships(args ListMemberships) error {
	return p.Call(args)
}

func (ListMemberships) Method() string {
	return http.MethodGet
}

func (ListMemberships) Path() string {
	return "listMembership/" + version + "/do/query"
}

func (q ListMemberships) Query() (map[string]string, error) {
	query := make(map[string]string)
	query["offset"] = strconv.Itoa(q.Offset)
	query["limit"] = strconv.Itoa(q.Limit)
	query["list_id"] = strconv.Itoa(q.ListID)
	return query, nil
}

func (q ListMemberships) Read(res []byte) error {
	body := struct {
		Result struct {
			Total int             `json:"total_results"`
			List  json.RawMessage `json:"list_membership"`
		} `json:"result"`
	}{}
	// Discard error and assume that the JSON from Pardot is valid.
	_ = json.Unmarshal(res, &body)

	// Got an empty page.
	if body.Result.List == nil {
		return nil
	}

	switch i := body.Result.Total; i {
	case 1:
		var p ListMembership
		err := json.Unmarshal(body.Result.List, &p)
		if err != nil {
			return errors.Wrap(err, "unmarshaling single membership")
		}
		*q.Placeholder = append(*q.Placeholder, p)
	default:
		err := json.Unmarshal(body.Result.List, q.Placeholder)
		if err != nil {
			return errors.Wrap(err, "unmarshaling memberships")
		}
	}
	return nil
}
