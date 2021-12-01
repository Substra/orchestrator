# Build stage
FROM golang:1.16-alpine3.15 AS build
RUN apk add --no-cache git=2.34.1-r0 make=4.3-r0 protoc=3.18.1-r1 protobuf-dev=3.18.1-r1

ARG VERSION=dev
ENV GO111MODULE=on
ENV SRC_DIR=/usr/src/chaincode

# Install protobuf codegen dependencies
RUN go get google.golang.org/protobuf/cmd/protoc-gen-go \
         google.golang.org/grpc/cmd/protoc-gen-go-grpc

WORKDIR ${SRC_DIR}

# Cache dependencies
COPY ./go.mod ./go.sum ${SRC_DIR}/
RUN go mod download

COPY . ${SRC_DIR}

RUN make ./bin/chaincode VERSION=$VERSION && mv ./bin/chaincode /bin/chaincode


# Expose the binary
FROM alpine:3.13 as prod

COPY --from=build /bin/chaincode /app/chaincode
USER 1000
WORKDIR /app

CMD ["/app/chaincode"]
