build:
	GOOS=linux GOARCH=amd64 go build -o ./lab/fit .

test:
	go test ./...

fmt:
	go fmt ./...

start-mom:
	docker compose down && docker compose up
