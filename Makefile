VERSION=$(shell git describe --tags --always)
IMAGE_NAME  ?= echaouchna/cork
DOCKER_USER ?= cork

build:
	go build -ldflags '-X main.corkVersion=${VERSION}' -o cork

dist:
	@gox \
		-ldflags='-X cork/cmd.corkVersion=${VERSION}' \
		--osarch "!darwin/386" \
		-output="bin/cork-{{.OS}}-{{.Arch}}"

build-docker:
	docker build --build-arg user=${DOCKER_USER} -t ${IMAGE_NAME}:${VERSION} .

publish-docker: build-docker
	docker tag ${IMAGE_NAME}:${VERSION} ${IMAGE_NAME}:latest
	docker push ${IMAGE_NAME}:${VERSION}
	docker push ${IMAGE_NAME}:latest

run-docker:
	docker run --rm -it -v ~/.config/gcloud:/home/${DOCKER_USER}/.config/gcloud ${IMAGE_NAME} $(DOCKER_ARGS)

clean:
	@rm -f cork
	@rm -rf bin