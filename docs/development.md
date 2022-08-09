# Development tips

## Enabling SSL passthrough with nginx-ingress on minikube

In order to enable [SSL-passthrough](https://kubernetes.github.io/ingress-nginx/user-guide/tls/#ssl-passthrough), you need to patch nginx controller.

This can be done with this snippet:

```sh
cat <<\EOF | kubectl patch deployment ingress-nginx-controller -n ingress-nginx --patch "$(cat -)"
---
spec:
  template:
    spec:
      containers:
      - name: controller
        args:
          - /nginx-ingress-controller
          - --configmap=$(POD_NAMESPACE)/nginx-load-balancer-conf
          - --report-node-internal-ip-address
          - --tcp-services-configmap=$(POD_NAMESPACE)/tcp-services
          - --udp-services-configmap=$(POD_NAMESPACE)/udp-services
          - --validating-webhook=:8443
          - --validating-webhook-certificate=/usr/local/certificates/cert
          - --validating-webhook-key=/usr/local/certificates/key
          - --enable-ssl-passthrough

EOF
```

## Running the backend on arm64 architecture (apple M1)

Bitnami does not yet provide a rabbitmq docker image for the arm64 processor. We are using the original rabbitmq image from dockerhub directly. Compatible image should be released in the future. (see [github issue](https://github.com/bitnami/charts/issues/7305))
The following patches are necessary as the bitnami charts used to install the rabbitmq image are not fully compatible.

1. Deploy with `skaffold run -p arm64`

2. After deploying run the patch
`./examples/tools/patch-rabbitmq-statefulset-arm64.sh`

## Go Language Server

If you're running into issues with things like "Go to definition" or "Find references", it could be because of [this gopls bug](https://github.com/golang/go/issues/29202). Try adding build flags to your editor config. Example for VSCode:

```json
"gopls": {
    "build.buildFlags": ["-tags=e2e"],
}
```
