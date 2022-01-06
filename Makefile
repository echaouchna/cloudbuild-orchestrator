IMAGE_NAME  ?= cork
DOCKER_USER ?= cork
VERSION=$(shell git describe --tags --always)

build:
	go build -ldflags '-X main.corkVersion=${VERSION}' -o cork

dist:
	@gox \
		-ldflags='-X main.corkVersion=${VERSION}' \
		--osarch "!darwin/386" \
		-output="bin/cork-{{.OS}}-{{.Arch}}"

build-docker:
	docker build --build-arg user=${DOCKER_USER} -t ${IMAGE_NAME} .

run-docker:
	docker run --rm -it -v ~/.config/gcloud:/home/${DOCKER_USER}/.config/gcloud ${IMAGE_NAME} $(DOCKER_ARGS)

clean:
	@rm -f cork
	@rm -rf bin