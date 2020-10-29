package pargo

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type DeleteProspect struct {
	ProspectID int
}

func (p *Pargo) DeleteProspect(args DeleteProspect) error {
	headers := make(http.Header)
	req, err := p.NewRequest(args, headers)
	if err != nil {
		return errors.Wrap(err, "building request")
	}
	_, err = p.Call(req)
	if err != nil {
		return errors.Wrap(err, "requesting")
	}
	return nil
}

func (DeleteProspect) Method() string {
	return http.MethodPost
}

func (q DeleteProspect) Path() string {
	return fmt.Sprintf("prospect/%s/do/delete/id/%d", version, q.ProspectID)
}

func (DeleteProspect) Query() (map[string]string, error) {
	return nil, nil
}
