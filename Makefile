build:
	GOOS=linux GOARCH=amd64 go build -o ./lab/fit .

test:
	go test ./...
