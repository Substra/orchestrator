#!/bin/bash

#
# Create a kubernetes secrets for ca.crt in namespaces org-1 and org-2
#
kubectl -n org-1 create secret generic orchestrator-tls-cacert --from-file=ca.crt --dry-run=client --output=yaml > secret-tls-org-1-cacert.yaml
kubectl -n org-2 create secret generic orchestrator-tls-cacert --from-file=ca.crt --dry-run=client --output=yaml > secret-tls-org-2-cacert.yaml

#
# The CA Cert in substra-backend should match
#
echo "Don't forget to update the \"orchestrator-tls-cacert\" ConfigMaps in substra-backend"
