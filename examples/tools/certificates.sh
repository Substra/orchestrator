#!/bin/bash

# By default, we use the CA cert key/pair provided in this folder (ca.cert/ca.key).
# Call this function to generate a new CA cert/key pair instead.
function generate_cacert {
    openssl genrsa -out ca.key 2048
    openssl req -new -x509 -days 365 -sha256 -key ca.key -extensions v3_ca -config openssl-with-ca.cnf -subj "/C=FR/ST=Loire-Atlantique/L=Nantes/O=Orchestrator Root CA/CN=Orchestrator Root CA" -out ca.crt
}

function generate_new_k8s_cacert {
    kubectl -n "org-1" create configmap orchestrator-tls-cacert --from-file=ca.crt --dry-run=client --output=yaml > "cm-tls-org-1-cacert.yaml"
    kubectl -n "org-2" create configmap orchestrator-tls-cacert --from-file=ca.crt --dry-run=client --output=yaml > "cm-tls-org-2-cacert.yaml"
    kubectl -n "org-3" create configmap orchestrator-tls-cacert --from-file=ca.crt --dry-run=client --output=yaml > "cm-tls-org-3-cacert.yaml"
    kubectl -n "cert-manager" create secret tls orchestrator-tls-ca --key="ca.key" --cert="ca.crt" --dry-run=client --output=yaml > "secret-cacert-certmanager.yaml"
    echo "Don't forget to update the \"orchestrator-tls-cacert\" ConfigMaps in substra-backend"
}

generate_cacert
generate_new_k8s_cacert
