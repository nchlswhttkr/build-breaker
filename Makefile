.PHONY: build clean

build:
	GOOS=linux go build -o bin/hello src/hello.go
	mkdir -p handlers
	zip handlers/hello.zip bin/hello

clean:
	rm -rf bin handlers