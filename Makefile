build:
	GOOS=linux GOARCH=amd64 go build -o ./lab/fit .

test:
	go test ./...

fmt:
	go fmt ./...
	cd codemods && go run ./shadow/main.go -w ../fit
	cd codemods && go run ./compacterr/main.go -w ../fit
	go fmt ./...

check: fmt test

start-mom:
	docker compose down && docker compose up
