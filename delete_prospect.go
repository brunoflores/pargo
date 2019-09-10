package pargo

import (
	"fmt"
	"net/http"
)

type DeleteProspect struct {
	ProspectID int
}

func (p *Pargo) DeleteProspect(args DeleteProspect) error {
	return p.Call(args)
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

func (q DeleteProspect) Read(res []byte) error {
	return nil
}
