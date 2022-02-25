# Build stage
FROM golang:1.17-alpine3.15 AS build

COPY ./docker/orchestrator-forwarder/dependencies.json /tmp/dependencies.json

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache jq \
        && jq -r '.[] | "\(.name)=\(.version)"' /tmp/dependencies.json | xargs apk add --no-cache \
        && rm /tmp/dependencies.json

ENV GO111MODULE=on
ENV SRC_DIR=/usr/src/orchestrator

# Install protobuf codegen dependencies
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

WORKDIR ${SRC_DIR}

# Download dependencies (cache them in a layer)
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY Makefile ./
COPY server ./server
COPY chaincode ./chaincode
COPY forwarder ./forwarder
COPY lib ./lib
COPY utils ./utils

RUN make ./bin/forwarder && mv ./bin/forwarder /bin/forwarder


# Expose the binary
FROM alpine:3.15 as prod

COPY --from=build /bin/forwarder /app/forwarder
USER 1000
WORKDIR /app

CMD ["/app/forwarder"]
