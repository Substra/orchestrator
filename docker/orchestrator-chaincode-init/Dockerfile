FROM alpine:3.14

RUN apk add --no-cache curl bash

COPY ./scripts/create-indexes.sh /usr/bin/create-indexes

CMD /usr/bin/create-indexes
