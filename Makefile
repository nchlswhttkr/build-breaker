.PHONY: build clean

build:
	GOOS=linux go build -o bin/hello src/hello/main.go
	GOOS=linux go build -o bin/tip src/tip/main.go
	mkdir -p handlers/
	zip handlers/hello.zip bin/hello
	zip handlers/tip.zip bin/tip

clean:
	rm -rf bin handlers
	