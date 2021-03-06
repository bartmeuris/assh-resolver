# ########################################################## #
# Makefile for Golang Project
# Includes cross-compiling, installation, cleanup
# ########################################################## #

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd dep
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

BINARY=assh-resolver
VERSION=0.0.2
BUILD=`git rev-parse HEAD`

PLATFORMS := linux/amd64 windows/amd64/.exe windows/386/.exe darwin/amd64 darwin/386
temp = $(subst /, ,$@)
tos   = $(word 1, $(temp))
tarch = $(word 2, $(temp))
ext = $(word 3, $(temp))
DEBUG = false

# Setup linker flags option for build that interoperate with variable names in src code
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD} -X main.Debug=${DEBUG}"

default: build

all: clean build_all

deps:
	dep ensure

# Simple build builds debug
build: DEBUG := true
build:
	go build ${LDFLAGS} -o ${BINARY}

build_all: $(PLATFORMS)

test: build
	go vet $$(go list ./... | grep -v /vendor/)
	go test -v -race ./...

$(PLATFORMS): deps
	CGO_ENABLED=0 GOOS=$(tos) GOARCH=$(tarch) go build $(LDFLAGS) -o '$(BINARY)_$(tos)-$(tarch)$(ext)'

install:
	go install ${LDFLAGS}

# Remove only what we've created
clean:
	-rm ${BINARY} ${BINARY}.exe ${BINARY}_*

.PHONY: check clean install build_all all
