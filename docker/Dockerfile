ARG DOCKER_BUILDER_VERSION
ARG DOCKER_RUNTIME_VERSION

# ----------------------------------------------------- builder ----------------
FROM golang:${DOCKER_BUILDER_VERSION} as builder

# download dependencies
WORKDIR /tmp
COPY go.mod go.sum bisquitt/
RUN cd bisquitt/ && go mod download

COPY . /opt/bisquitt
WORKDIR /opt/bisquitt

ARG WITH_RACE_DETECTION
ENV WITH_RACE_DETECTION=${WITH_RACE_DETECTION}
RUN make build

# ----------------------------------------------------- runtime ----------------
FROM debian:${DOCKER_RUNTIME_VERSION}

COPY --from=builder \
    /opt/bisquitt/cmd/bisquitt/bisquitt \
    /opt/bisquitt/cmd/bisquitt-pub/bisquitt-pub \
    /opt/bisquitt/cmd/bisquitt-sub/bisquitt-sub \
    /usr/local/bin/

RUN groupadd --system --gid 983 bisquitt && \
    useradd --system --uid 983 --no-create-home --home-dir /var/empty \
        --shell /sbin/nologin --gid bisquitt bisquitt

EXPOSE 1883:1883/udp

CMD ["/usr/local/bin/bisquitt", "--host", "0.0.0.0"]
