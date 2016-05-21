build:
	docker run --rm -v "$(PWD)":/go/src/github.com/moznion/resque_exporter -w /go/src/github.com/moznion/resque_exporter golang:1.6 bash author/build.sh $(VERSION)

