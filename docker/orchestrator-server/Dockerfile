# Build stage
FROM golang:1.17-alpine3.15 AS build
RUN apk add --no-cache git=2.34.1-r0 make=4.3-r0 protoc=3.18.1-r1 protobuf-dev=3.18.1-r1

ARG VERSION=dev
ENV GO111MODULE=on
ENV SRC_DIR=/usr/src/orchestrator

RUN wget -q https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.3.6/grpc_health_probe-linux-amd64 -O /bin/grpc_health_probe && chmod +x /bin/grpc_health_probe

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
COPY lib ./lib
COPY utils ./utils

RUN make ./bin/orchestrator VERSION=$VERSION && mv ./bin/orchestrator /bin/orchestrator


# Expose the binary
FROM alpine:3.13 as prod

COPY --from=build /bin/orchestrator /app/orchestrator
COPY --from=build /bin/grpc_health_probe /bin/grpc_health_probe
USER 1000
WORKDIR /app

CMD ["/app/orchestrator"]
