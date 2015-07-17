haproxy-librato-linux-amd64: main.go
	docker run --rm -v "$(CURDIR)":/usr/src/myapp -w /usr/src/myapp golang:1.3 /bin/sh -c 'go get -d -v && go build -v'
	mv myapp haproxy-librato-linux-amd64
