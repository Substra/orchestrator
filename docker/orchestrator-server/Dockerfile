# Build stage
FROM golang:1.19-alpine3.16 AS build

COPY ./docker/orchestrator-server/dependencies.json /tmp/dependencies.json

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache jq \
        && jq -r '.[] | "\(.name)=\(.version)"' /tmp/dependencies.json | xargs apk add --no-cache \
        && rm /tmp/dependencies.json

ARG VERSION=dev
ENV GO111MODULE=on
ENV SRC_DIR=/usr/src/orchestrator

RUN wget -q https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.11/grpc_health_probe-linux-amd64 -O /bin/grpc_health_probe && chmod +x /bin/grpc_health_probe

# Install migration tool
RUN wget -qO- https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.tar.gz | tar xzf - -C /bin/

# Install protobuf codegen dependencies
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

WORKDIR ${SRC_DIR}

# Download dependencies (cache them in a layer)
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY Makefile ./
COPY server ./server
COPY chaincode ./chaincode
COPY lib ./lib
COPY utils ./utils
COPY forwarder ./forwarder

RUN make ./bin/orchestrator VERSION=$VERSION && mv ./bin/orchestrator /bin/orchestrator


# Expose the binary
FROM alpine:3.16 as prod

COPY --from=build /bin/orchestrator /app/orchestrator
COPY --from=build /bin/grpc_health_probe /bin/grpc_health_probe
COPY --from=build /bin/migrate /bin/migrate

COPY ./server/standalone/migration /migrations

USER 1000
WORKDIR /app

CMD ["/app/orchestrator"]
