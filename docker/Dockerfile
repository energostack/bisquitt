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

RUN addgroup --system --gid 1883 bisquitt 2>/dev/null && \
    adduser --system --uid 1883 --disabled-password --no-create-home \
        --home /var/empty --shell /sbin/nologin --ingroup bisquitt bisquitt 2>/dev/null

EXPOSE 1883:1883/udp

CMD ["/usr/local/bin/bisquitt", "--host", "0.0.0.0"]
