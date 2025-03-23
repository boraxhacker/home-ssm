.PHONY: clean run build

BINARY_NAME=home-ssm

build: clean 
	GOOS=linux GOARCH=amd64 go build -o ${BINARY_NAME}

run:
	go run main.go

clean:
	go clean
	rm -rf ${BINARY_NAME}
