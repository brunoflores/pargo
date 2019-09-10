package pargo

import (
	"encoding/json"
	"sync"
)

type QueryAllProspects struct {
	Fields []string
	Page   func(json.RawMessage)
}

// QueryAllProspects will return a slice of all *prospects* in Pardot.
// The order in which prospects are returned is not guaranteed.
func (p *Pargo) QueryAllProspects(query QueryAllProspects) {

	// The Pardot REST API allows at most 5 parallel requests.
	// Here we are making 4 to be on the safe side.
	// If we exceed 5 at any point, the usual response is a time-out.
	const workers = 4

	var done = make(chan struct{}, workers)
	var jobs = make(chan int)
	var shut = make(chan struct{}, 1)
	var wg sync.WaitGroup

	queryWorker := func(jobs <-chan int, done chan<- struct{}) {
		defer func() { done <- struct{}{} }()
		for n := range jobs {
			err := p.QueryProspects(QueryProspects{
				Offset:    n * 200,
				Limit:     200,
				Fields:    query.Fields,
				Marshaler: query.Page,
			})
			if err != nil {
				switch err.(type) {
				case QueryProspectsEOF:
					return
				default:
					return
				}
			}
		}
	}

	for i := 0; i < workers; i++ {
		go queryWorker(jobs, done)
	}

	go func() {
		page := 0
		for {
			select {
			case <-shut:
				close(jobs)
				close(done)
			default:
				jobs <- page
				page++
			}
		}
	}()

	wg.Add(workers)
	go func() {
		for range done {
			wg.Done()
		}
	}()
	wg.Wait()
	shut <- struct{}{}
}
