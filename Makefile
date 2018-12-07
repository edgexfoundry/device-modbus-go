.PHONY: build test clean prepare update docker

GO=CGO_ENABLED=0 go

MICROSERVICES=cmd/device-modbus

.PHONY: $(MICROSERVICES)

DOCKERS=docker_device_modbus_go
.PHONY: $(DOCKERS)

VERSION=$(shell cat ./VERSION)
GIT_SHA=$(shell git rev-parse HEAD)
GOFLAGS=-ldflags "-X github.com/edgexfoundry/device-modbus-go.Version=$(VERSION)"

build: $(MICROSERVICES)
	go build ./...

cmd/device-modbus:
	$(GO) build $(GOFLAGS) -o $@ ./cmd

test:
	go test ./... -cover

clean:
	rm -f $(MICROSERVICES)

prepare:
	glide install

update:
	glide update

docker: $(DOCKERS)

docker_device_modbus_go:
	docker build \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-device-modbus-go:$(GIT_SHA) \
		-t edgexfoundry/docker-device-modbus-go:$(VERSION)-dev \
		.