.PHONY: build clean

build: clean
	GOOS=linux go build -o bin/hello src/hello/main.go
	GOOS=linux go build -o bin/tip src/tip/main.go
	mkdir -p handlers/${BB_VERSION}
	zip handlers/${BB_VERSION}/hello.zip bin/hello
	zip handlers/${BB_VERSION}/tip.zip bin/tip

clean:
	rm -rf bin handlers
	