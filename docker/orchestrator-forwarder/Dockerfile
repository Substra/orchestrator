# Build stage
FROM golang:1.16-alpine AS build
RUN apk add --no-cache git make protoc protobuf-dev

ENV GO111MODULE=on
ENV SRC_DIR=/usr/src/orchestrator

# Install protobuf codegen dependencies
RUN go get google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc \
    github.com/go-bindata/go-bindata/...

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

RUN make ./bin/forwarder
RUN mv ./bin/forwarder /bin/forwarder


# Expose the binary
FROM alpine:3.13 as prod

COPY --from=build /bin/forwarder /app/forwarder
USER 1000
WORKDIR /app

CMD ./forwarder
