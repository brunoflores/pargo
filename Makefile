SHELL := /bin/bash

test:
	go test --cover --race --count=1 ./...
