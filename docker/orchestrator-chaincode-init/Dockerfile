FROM alpine:3.14

RUN apk add --no-cache curl=7.79.1-r0 bash=5.1.4-r0

COPY ./scripts/create-indexes.sh /usr/bin/create-indexes

CMD ["/usr/bin/create-indexes"]
