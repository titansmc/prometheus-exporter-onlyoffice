VERSION   := $(shell cat VERSION)
GO    		:= GO111MODULE=on go
PROMU 		:= $(shell $(GO) env GOPATH)/bin/promu
BIN       := prometheus_onlyoffice_exporter
CONTAINER := prometheus_onlyoffice_exporter
GOOS      ?= linux
GOARCH    ?= amd64

GOFLAGS   := -ldflags "-X main.Version=$(VERSION)" -a -installsuffix cgo
TAR       := $(BIN)-$(VERSION)-$(GOOS)-$(GOARCH).tar.gz

PREFIX    ?= $(shell pwd)

default: $(BIN)

$(BIN):
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(PROMU) build --prefix $(PREFIX)

release: $(TAR)
	curl -XPOST --data-binary @$< $(DST)/$<

build-docker: $(BIN)
	docker build -t $(CONTAINER) .

$(TAR): $(BIN)
	tar czf $@ $<