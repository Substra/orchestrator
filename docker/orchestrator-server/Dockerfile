# Build stage
FROM golang:1.16-alpine AS build
RUN apk add --no-cache git make protoc protobuf-dev

ARG version=dev
ENV GO111MODULE=on
ENV SRC_DIR=/usr/src/orchestrator

RUN wget https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.3.6/grpc_health_probe-linux-amd64 -O /bin/grpc_health_probe
RUN chmod +x /bin/grpc_health_probe

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

RUN make ./bin/orchestrator VERSION=$version
RUN mv ./bin/orchestrator /bin/orchestrator


# Expose the binary
FROM alpine:3.13 as prod

COPY --from=build /bin/orchestrator /app/orchestrator
COPY --from=build /bin/grpc_health_probe /bin/grpc_health_probe
USER 1000
WORKDIR /app

CMD ./orchestrator
