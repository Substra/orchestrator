FROM alpine:3.18

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

COPY ./docker/orchestrator-chaincode-init/dependencies.json /tmp/dependencies.json

# hadolint ignore=DL3018
RUN apk update && apk add --no-cache jq \
	&& jq -r '.[] | "\(.name)=\(.version)"' /tmp/dependencies.json | xargs apk add --no-cache \
	&& rm /tmp/dependencies.json

COPY ./scripts/create-indexes.sh /usr/bin/create-indexes

CMD ["/usr/bin/create-indexes"]
