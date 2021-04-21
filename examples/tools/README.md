# Tools
These tools are presented for local development and documentation purposes only.

They are maintained on a "best effort" policy.

## Certificates

If at some point you need to generate new example certificates for the orchestrator, you can use the script `certificates.sh`.

Running the script will generate kubernetes secrets containing server certificates for the orchestrator and client certificates for any service that would need to communicate with the orchestrator.

You will need to move the server certificates to `/examples/secrets/`.

### Generating a new CA cert

If you need a new CA certificate you will need to uncomment two lines in `certficiates.sh` before running the script:
```sh
# generate_cacert
```
```sh
# generate_new_k8s_cacert "${row}"
```
