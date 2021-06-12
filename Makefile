.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -o ./bin/main

.PHONY: tidy
tidy:
	go mod tidy -v
	
.PHONY: pr-prep
pr-prep:
	go test ./cmd

.PHONY: update-dependencies
update-dependencies:
	go get -u
