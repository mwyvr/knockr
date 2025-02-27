BINARY=knockr

# -s -w strips debug info as well
build:
	go build -v -ldflags="-s -w -X 'main.version=`git describe --tags --abbrev=0`'" 

nocgo:
	CGO_ENABLED=0 go build -v -ldflags="-s -w -X 'main.version=`git describe --tags --abbrev=0`'" 

clean:
	rm ${BINARY}

lint:
	golangci-lint run

