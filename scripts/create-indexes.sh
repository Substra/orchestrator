#!/usr/bin/env bash

set -e -o pipefail

COUCHDB_BASEURL="http://$COUCHDB_USER:$COUCHDB_PASSWORD@$COUCHDB_INSTANCE/"

IFS="," read -r -a CHANS <<< $CHANNELS

for CHAN in "${CHANS[@]}"; do
    DB="${CHAN}_$CHAINCODE_NAME"

    http_code=$(curl -I -s -o /dev/null $COUCHDB_BASEURL/$DB -w '%{http_code}')
    while [ $http_code != "200" ]
    do
        echo "$DB does not exist yet, waiting..."
        http_code=$(curl -I -s -o /dev/null $COUCHDB_BASEURL/$DB -w '%{http_code}')
        sleep 10s
    done

    echo "create doc_type index on $DB"
    curl -S -s -o /dev/null -i -X POST -H "Content-Type: application/json" -d \
         "{\"index\":{\"fields\":[\"doc_type\"]},
         \"name\":\"ix_doc_type\",
         \"ddoc\":\"genericAssetDoc\",
         \"type\":\"json\"}" \
             "$COUCHDB_BASEURL/$DB/_index"

    echo "create event_ts index on $DB"
    curl -S -s -o /dev/null -i -X POST -H "Content-Type: application/json" -d \
         "{\"index\":{\"fields\":[\"doc_type\",\"asset.timestamp\",\"asset.id\"]},
         \"name\":\"ix_event_ts\",
         \"ddoc\":\"eventDoc\",
         \"type\":\"json\"}" \
             "$COUCHDB_BASEURL/$DB/_index"
done
