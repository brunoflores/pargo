package pargo

import (
	"encoding/json"
	"sync"
)

type QueryAllProspects struct {
	Fields    []string
	Page      func(json.RawMessage)
	Heartbeat func(offset, limit int)
}

// QueryAllProspects will return a slice of all *prospects* in Pardot.
// The order in which prospects are returned is not guaranteed.
func (p *Pargo) QueryAllProspects(query QueryAllProspects) error {

	// The Pardot REST API allows at most 5 parallel requests.
	// Here we are making 4 to be on the safe side.
	const workers = 4

	var done = make(chan struct{}, workers)
	var quit = make(chan error, workers)
	var err error
	var jobs = make(chan int)
	var shut = make(chan struct{}, workers)
	var wg sync.WaitGroup

	maybeLog := func(fn func(int, int), offset, limit int) {
		if fn == nil {
			return
		}
		fn(offset, limit)
	}

	queryWorker := func(
		jobs <-chan int,
		done chan<- struct{},
		quit chan<- error) {

		defer func() { done <- struct{}{} }()
		for n := range jobs {
			var (
				limit  = 200
				offset = n * limit
			)
			maybeLog(query.Heartbeat, offset, limit)
			err := p.QueryProspects(QueryProspects{
				Offset:    offset,
				Limit:     limit,
				Fields:    query.Fields,
				Marshaler: query.Page,
			})
			if err != nil {
				switch err.(type) {
				case QueryProspectsEOF:
					return
				default:
					quit <- err
					return
				}
			}
		}
	}

	for i := 0; i < workers; i++ {
		go queryWorker(jobs, done, quit)
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

	go func() {
		for {
			select {
			case e := <-quit:
				err = e
				shut <- struct{}{}
			case <-shut:
				return
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

	return err
}
