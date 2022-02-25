# Build stage
FROM golang:1.17-alpine3.15 AS build
COPY ./docker/orchestrator-chaincode/dependencies.json /tmp/dependencies.json

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache jq \
	&& jq -r '.[] | "\(.name)=\(.version)"' /tmp/dependencies.json | xargs apk add --no-cache \
	&& rm /tmp/dependencies.json

ARG VERSION=dev
ENV GO111MODULE=on
ENV SRC_DIR=/usr/src/chaincode

# Install protobuf codegen dependencies
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

WORKDIR ${SRC_DIR}

# Cache dependencies
COPY ./go.mod ./go.sum ${SRC_DIR}/
RUN go mod download

COPY . ${SRC_DIR}

RUN make ./bin/chaincode VERSION=$VERSION && mv ./bin/chaincode /bin/chaincode


# Expose the binary
FROM alpine:3.15 as prod

COPY --from=build /bin/chaincode /app/chaincode
USER 1000
WORKDIR /app

CMD ["/app/chaincode"]
