CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
TAG=${TAG:-latest}
COMMIT=`git rev-parse --short HEAD`
GO15VENDOREXPERIMENT=1

all: build media

clean:
	@rm -rf shipyard

build: clean
	@go build -a -tags "netgo static_build" -installsuffix netgo -ldflags "-w -X github.com/shipyard/shipyard/version.GitCommit=$(COMMIT)" .

remote-build:
	@docker build -t shipyard-build -f Dockerfile.build .
	@rm -f ./controller/controller
	@cd controller && docker run --rm -w /go/src/github.com/shipyard/shipyard --entrypoint /bin/bash shipyard-build -c "make build 1>&2 && cd controller && tar -czf - controller" | tar zxf -

media:
	@cd controller/static && bower -s install --allow-root -p || `echo "Please inspect that Bower is installed and available." && exit -1`

image: media build
	@echo Building Shipyard image $(TAG)
	@cd controller && docker build -t shipyard/shipyard:$TAG .

dist-clean:
	@rm -rf dist

dist: dist-clean all
	@mkdir -p dist/controller && cp shipyard dist/ && cp -R controller/static dist/controller

release-clean:
	@rm -rf release

release: release-clean dist
	@mkdir -p release && cd dist && tar -czf ../release/shipyard-ilm.tar.gz * && cd ../ && echo "Release available under release/ directory"

test: clean
	@go test -v `go list ./... | grep -v /vendor | grep -v /test-assets`