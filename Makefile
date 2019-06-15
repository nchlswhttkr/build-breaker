.PHONY: build clean

build: clean
	GOOS=linux go build -o bin/hello src/hello/main.go
	mkdir -p handlers/${BB_VERSION}
	zip handlers/${BB_VERSION}/hello.zip bin/hello

clean:
	rm -rf bin handlers
	