VERSION = 0.2.0

DOCKER_REGISTRY ?= docker.io
DOCKER_BUILDER_VERSION ?= 1.17.8-bullseye
DOCKER_RUNTIME_VERSION ?= bullseye-20220228-slim
DOCKER_IMAGE = energomonitor/bisquitt

# Enable data race detection in the compiled binaries.
ifeq ($(WITH_RACE_DETECTION), 1)
EXTRA_BUILD_ARGS += -race
endif

.PHONY: build
build:
	cd cmd/bisquitt && go build -ldflags="-X 'main.Version=$(VERSION)'" $(EXTRA_BUILD_ARGS)
	cd cmd/bisquitt-pub && go build -ldflags="-X 'main.Version=$(VERSION)'" $(EXTRA_BUILD_ARGS)
	cd cmd/bisquitt-sub && go build -ldflags="-X 'main.Version=$(VERSION)'" $(EXTRA_BUILD_ARGS)

.PHONY: update
update:
	go get -u ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	goimports -w ./

.PHONY: test
test:
	go test ./...

.PHONY: docker/test
docker/test:
	docker-compose -f docker-compose.test.yaml build \
					--build-arg "WITH_RACE_DETECTION=1" \
					--build-arg "DOCKER_BUILDER_VERSION=$(DOCKER_BUILDER_VERSION)" \
					--build-arg "DOCKER_RUNTIME_VERSION=$(DOCKER_RUNTIME_VERSION)"
	docker-compose -f docker-compose.test.yaml up --remove-orphans --abort-on-container-exit --exit-code-from bisquitt-test

.PHONY: docker/build
docker/build:
	docker build --build-arg "DOCKER_BUILDER_VERSION=$(DOCKER_BUILDER_VERSION)" \
					--build-arg "DOCKER_RUNTIME_VERSION=$(DOCKER_RUNTIME_VERSION)" \
					-f docker/Dockerfile -t "$(DOCKER_IMAGE):latest" -t "$(DOCKER_IMAGE):$(VERSION)" .

.PHONY: docker/push
docker/push:
	docker tag "$(DOCKER_IMAGE):latest" "$(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest"
	docker tag "$(DOCKER_IMAGE):$(VERSION)" "$(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(VERSION)"
	docker push "$(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest"
	docker push "$(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(VERSION)"

.PHONY: docker/run
docker/run:
	docker-compose build --build-arg "DOCKER_BUILDER_VERSION=$(DOCKER_BUILDER_VERSION)" \
					--build-arg "DOCKER_RUNTIME_VERSION=$(DOCKER_RUNTIME_VERSION)"
	docker-compose up --remove-orphans --abort-on-container-exit
