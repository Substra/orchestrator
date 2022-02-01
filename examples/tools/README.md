# Tools
These tools are presented for local development and documentation purposes only.

They are maintained on a "best effort" policy.

## Generating a new CA certificate

If at some point you need to generate a new cacert you can run `certificates.sh`.

Running the script will generate kubernetes secrets containing crypto materials used by cert-manager to generate certificates for the orchestrator and eventually the backend server.

You will need to move the generated ConfigMaps and Secrets to `/examples/k8s/`.

## Retrieving a client certificate for debug purposes

After the deployment of the orchestrator in your cluster you can use `./download_client_cert.sh` to generate a valid client certificate to interact with your orchestrator.
