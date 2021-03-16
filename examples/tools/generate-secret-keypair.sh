#!/bin/bash

# Params
NAMESPACE="org-1"
SERVER_CLIENT="server"  # replace with "client" for client certs

# Vars
SECRET_NAME="orchestrator-tls-${SERVER_CLIENT}-pair"
FILENAME="secret-tls-${NAMESPACE}-${SERVER_CLIENT}-pair.yaml"

kubectl -n ${NAMESPACE} create secret generic ${SECRET_NAME} \
    --from-file=ca.crt `# rabbitmq chart needs this` \
    --from-file=tls.crt \
    --from-file=tls.key \
    --dry-run=client --output=yaml > ${FILENAME}
