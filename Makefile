SHELL := /bin/bash

test:
	go test -v --cover -coverprofile=coverage.txt -covermode=atomic --race --count=1 ./...
