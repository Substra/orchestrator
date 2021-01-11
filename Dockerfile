# Build stage
FROM golang:1.13.8-alpine AS build
RUN apk add --no-cache git make protoc

ENV GO111MODULE=on
ENV SRC_DIR=/usr/src/chaincode

# Install protobuf codegen dependencies
RUN go get google.golang.org/protobuf/cmd/protoc-gen-go \
         google.golang.org/grpc/cmd/protoc-gen-go-grpc

COPY . ${SRC_DIR}
WORKDIR ${SRC_DIR}

RUN make chaincode
RUN mv ./bin/chaincode /bin/chaincode


# Expose the binary
FROM alpine:3.12 as prod

COPY --from=build /bin/chaincode /app/chaincode
USER 1000
WORKDIR /app

CMD ./chaincode
