.EXPORT_ALL_VARIABLES:
NAMESPACE ?= camabeh
APP ?= etcd-extractor
VERSION ?= latest
GIT_HASH ?= $(shell git describe --tags --dirty --always)

.PHONY: clean prepare build docker-build docker-push docker-clean

prepare:
	go mod download

build:
	mkdir -p bin \
		&& GOOS=linux GOARCH=amd64 go build -o ./bin/${APP}.amd64 \
			-ldflags "-X github.com/camabeh/etcd-extractor/pkg/version.value=$(GIT_HASH)" ./cli

clean:
	rm -rf ./bin

docker-build: build
	# Buildkit needs to be enabled to avoid copying the whole
	# context (contents of folder where Dockerfile is) to DOCKER_HOST
	# if not building on the same machine.
	# .dockerignore works only for COPY/ADD commands
	DOCKER_BUILDKIT=1 docker image build \
		-t ${NAMESPACE}/${APP}:${VERSION} .

docker-push: docker-build
	docker push ${NAMESPACE}/${APP}:${VERSION}

docker-clean:
	docker rmi -f ${NAMESPACE}/${APP}:${VERSION}
