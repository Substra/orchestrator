#!/bin/bash

#######################
# Set parameters here #
#######################
ORGANIZATION="MyOrg2MSP"
NAMESPACE="org-2"
DOMAIN="node-2.com"


###########
# CA Cert #
###########
# By default, we use the CA cert key/pair provided in this folder (ca.cert/ca.key).
# Uncomment these lines to generate a new CA cert/key pair instead.
#
# openssl genrsa -out ca.key 2048
# openssl req -new -x509 -days 365 -key ca.key -subj "/C=FR/ST=Loire-Atlantique/L=Nantes/O=Orchestrator Root CA/CN=Orchestrator Root CA" -out ca.crt


###############
# Target cert #
###############
openssl req -newkey rsa:2048 -nodes -keyout tls.key -subj "/C=CN/ST=GD/L=SZ/O=${ORGANIZATION}/CN=orchestrator.${DOMAIN}" -out cert.csr
openssl x509 -req \
    -days 365 -in cert.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt \
    -extfile <(printf "subjectAltName=DNS:orchestrator.${DOMAIN},DNS:owkin-orchestrator-${NAMESPACE}.${NAMESPACE}.svc.cluster.local,DNS:owkin-orchestrator-${NAMESPACE}-rabbitmq.${NAMESPACE}.svc.cluster.local")
rm cert.csr ca.srl
