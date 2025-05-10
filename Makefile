
### definition
APP=chatbot-go
MAIN_FILE=cmd/main.go
RELEASE=$(shell git describe --tags)
RELEASE:=$(if ${RELEASE},${RELEASE},"v0.0.1")
CONTAINER_IMAGE?=imilair.com/${APP}


### build
build:
	GEN_TARGET=no sh build.sh
	@echo "build successfully"

.PHONY: build

clean:
	rm -rf $(APP) target

swag:
	swag init -g $(MAIN_FILE) --parseDependency --outputTypes json


container:
	docker build -t $(CONTAINER_IMAGE):$(RELEASE) -f Dockerfile .
container_tag: container
	echo "container_tag"
	docker tag $(CONTAINER_IMAGE):${RELEASE}  $(CONTAINER_IMAGE):${RELEASE}
	docker tag $(CONTAINER_IMAGE):${RELEASE}  $(CONTAINER_IMAGE):latest
container_rc: container_tag
	echo "container_rc"
	docker push $(CONTAINER_IMAGE):${RELEASE}
	docker push $(CONTAINER_IMAGE):latest
push: container_rc
	echo "finished"
