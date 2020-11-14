SHELL := /bin/bash

test:
	go test -v --cover --race --count=1 ./...
