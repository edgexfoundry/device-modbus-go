GO=GO111MODULE=on go
HOOKS=cmd/configure/configure cmd/install/install

tidy:
	go mod tidy

build: tidy $(HOOKS)

cmd/configure/configure:
	$(GO) build -o $@ ./cmd/configure

cmd/install/install:
	$(GO) build -o $@ ./cmd/install

clean:
	rm -f $(HOOKS)

