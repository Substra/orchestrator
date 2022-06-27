#! /usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

kubectl apply -f "${SCRIPT_DIR}/../k8s/client-cert.yaml"

# Sleep a few seconds to make sure the cert is created/populated
sleep 3

kubectl get secret -n org-1 -o json orchestrator-client-cert | jq -r '.data."tls.crt"' | base64 -d > "${SCRIPT_DIR}/client-org-1.crt"
kubectl get secret -n org-1 -o json orchestrator-client-cert | jq -r '.data."tls.key"' | base64 -d > "${SCRIPT_DIR}/client-org-1.key"
