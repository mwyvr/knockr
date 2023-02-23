BINARY=knockr

# build a linux CGO-free binary for upload to github release page
# -s -w strips debug info as well
build:
	GOOS=linux CGO=0 go build -v -ldflags="-s -w -X 'main.version=`git describe --tags --abbrev=0`'" 

clean:
	rm ${BINARY}

lint:
	golangci-lint run

