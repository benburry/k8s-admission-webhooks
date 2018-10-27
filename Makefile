NAME = k8s-admission-webhooks
REPO ?= benburry/$(NAME)
GOPKG = github.com/benburry/$(NAME)
DOCKER ?= docker
GOVERSION = 1.9.2
DOCKERENV = -e CGO_ENABLED=0 -e GOOS="linux" -e GOARCH="amd64"

SHA = $(shell git show-ref --hash=10 --head | head -n1)
TIMESTAMP := $(shell date +%Y%m%d-%H%M%S-%Z)
VERSION = $(SHA)-$(TIMESTAMP)

default: container

test:
	-$(DOCKER) rm -f builder
	$(DOCKER) run --name builder --rm -v ${CURDIR}:/go/src/${GOPKG} -w /go/src/${GOPKG} ${DOCKERENV} golang:$(GOVERSION) /bin/sh -c 'go test -v $$(go list ./... | grep -v /vendor/)'

$(NAME): test
	-$(DOCKER) rm -f builder
	$(DOCKER) run --name builder --rm -v ${CURDIR}:/go/src/${GOPKG} -w /go/src/${GOPKG} ${DOCKERENV} golang:$(GOVERSION) go build -a -o $(NAME) .

clean:
	-rm -f $(NAME)

clean:
	rm $(NAME)

container: $(NAME)
	$(DOCKER) build -t $(NAME):$(VERSION) .

deploy: container
	$(DOCKER) tag $(NAME):$(VERSION) $(REPO):$(VERSION)
	$(DOCKER) push $(REPO):$(VERSION)

deploy-images: deploy

.PHONY: container deploy-images deploy default clean builder test

