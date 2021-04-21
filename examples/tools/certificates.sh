#!/bin/bash

#######################
# Set nodes here      #
#######################
nodes='[
    {
        "organization":"MyOrg1MSP",
        "namespace":"org-1",
        "domain": "node-1.com"
    },
    {
        "organization":"MyOrg2MSP",
        "namespace":"org-2",
        "domain": "node-2.com"
    }
]'

# By default, we use the CA cert key/pair provided in this folder (ca.cert/ca.key).
# Call this function to generate a new CA cert/key pair instead.
function generate_cacert {
    openssl genrsa -out ca.key 2048
    openssl req -new -x509 -days 365 -key ca.key -subj "/C=FR/ST=Loire-Atlantique/L=Nantes/O=Orchestrator Root CA/CN=Orchestrator Root CA" -out ca.crt
}

function generate_new_k8s_cacert {
    local namespace
    namespace=$(echo "${1}" | jq -r .namespace)
    kubectl -n "${namespace}" create secret generic orchestrator-tls-cacert --from-file=ca.crt --dry-run=client --output=yaml > "secret-tls-${namespace}-cacert.yaml"
    echo "Don't forget to update the \"orchestrator-tls-cacert\" ConfigMaps in substra-backend"
}

# Generates a new cert for the target corresponding to the provided arguments
function generate_target_cert {
    local organization="$1"
    local namespace="$2"
    local domain="$3"
    openssl req -newkey rsa:2048 -nodes -keyout tls.key -subj "/C=FR/ST=Loire-Atlantique/L=Nantes/O=${organization}/CN=orchestrator.${domain}" -out cert.csr
    openssl x509 -req \
    -days 365 -in cert.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt \
    -extfile <(printf "subjectAltName=DNS:orchestrator.${domain},DNS:owkin-orchestrator-${namespace}-server.${namespace}.svc.cluster.local,DNS:owkin-orchestrator-${namespace}-rabbitmq.${namespace}.svc.cluster.local")
    rm cert.csr ca.srl
}

function generate_new_k8s_cert {
    local server_client="$1"
    local namespace="$2"
    local secret_name="orchestrator-tls-${server_client}-pair"
    local filename

    if [ "${server_client}" = "server" ]
    then
        filename="secret-tls-${namespace}-${server_client}-pair.yaml"
    else
        filename="secret-orchestrator-tls-${namespace}-${server_client}-pair.yaml"
    fi

    kubectl -n "${namespace}" create secret generic "${secret_name}" \
        --from-file=ca.crt `# rabbitmq chart needs this` \
        --from-file=tls.crt \
        --from-file=tls.key \
        --dry-run=client --output=yaml > "${filename}"
}

function process {
    local organization
    local namespace
    local domain
    organization=$(echo "${1}" | jq -r .organization)
    namespace=$(echo "${1}" | jq -r .namespace)
    domain=$(echo "${1}" | jq -r .domain)

    generate_target_cert "${organization}" "${namespace}" "${domain}"
    generate_new_k8s_cert "server" "${namespace}"
    generate_new_k8s_cert "client" "${namespace}"

    rm tls.crt tls.key
}

########
# Main #
########

# Uncomment to generate cacerts
# generate_cacert

for row in $(echo "${nodes}" | jq -c '.[]'); do
    process "${row}"
    # Uncomment to generate cacerts
    # generate_new_k8s_cacert "${row}"
done
