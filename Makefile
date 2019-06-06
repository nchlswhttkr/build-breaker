.PHONY: build clean

build:
	go build -o bin/hello src/hello.go

clean:
	rm -rf bin