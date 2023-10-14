# Build stage
FROM golang:1.19-bookworm AS build

# Install necessary utilities
RUN apt-get update && apt-get install -y git make protoc protobuf-dev --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

SHELL ["/bin/bash", "-eo", "pipefail", "-c"]

ARG VERSION=dev
ENV GO111MODULE=on
ENV SRC_DIR=/usr/src/chaincode

# Install protobuf codegen dependencies
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

WORKDIR ${SRC_DIR}

# Cache dependencies
COPY ./go.mod ./go.sum ${SRC_DIR}/
RUN go mod download

COPY . ${SRC_DIR}

RUN make ./bin/chaincode VERSION=$VERSION && mv ./bin/chaincode /bin/chaincode


# Expose the binary
FROM debian:bookworm-slim as prod

COPY --from=build /bin/chaincode /app/chaincode
USER 1000
WORKDIR /app

CMD ["/app/chaincode"]
