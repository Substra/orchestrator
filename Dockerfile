# Build stage
FROM golang:1.13.8-alpine AS build
RUN apk add --no-cache git

ENV GO111MODULE=on
ENV GOPATH=/go
ENV SRC_DIR=${GOPATH}/src/github.com/substrafoundation/substra-orchestrator
ENV CHAINCODE_DIR=${SRC_DIR}/chaincode

COPY . ${SRC_DIR}
WORKDIR ${SRC_DIR}

RUN go mod download
RUN go mod verify
RUN CGO_ENABLED=0 go build -o /bin/chaincode -v ${CHAINCODE_DIR}


# Expose the binary
FROM alpine:3.11 as prod
COPY --from=build /bin/chaincode /app/chaincode
USER 1000
WORKDIR /app

CMD ./chaincode
