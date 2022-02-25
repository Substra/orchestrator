FROM debian:bullseye-slim

COPY ./docker/orchestrator-rabbitmq-operator/dependencies.json /tmp/dependencies.json

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends jq \
          && jq -r '.[] | "\(.name)=\(.version)"' /tmp/dependencies.json | xargs apt-get install -y --no-install-recommends \
	  && apt-get clean \
	  && rm -rf /var/lib/apt/lists/* \
	  && rm /tmp/dependencies.json

# Install rabbitmqadmin
RUN curl https://raw.githubusercontent.com/rabbitmq/rabbitmq-server/v3.8.14/deps/rabbitmq_management/bin/rabbitmqadmin > /usr/bin/rabbitmqadmin && chmod +x /usr/bin/rabbitmqadmin
