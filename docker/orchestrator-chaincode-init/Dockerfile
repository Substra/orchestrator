FROM alpine:3.15

RUN apk add --no-cache curl=7.80.0-r0 bash=5.1.8-r0

COPY ./scripts/create-indexes.sh /usr/bin/create-indexes

CMD ["/usr/bin/create-indexes"]
