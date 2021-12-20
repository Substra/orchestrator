FROM debian:bullseye-slim

RUN apt-get update && apt-get install --no-install-recommends python3=3.9.2-3 curl=7.74.0-1.3+deb11u1 ca-certificates=20210119 -y \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install rabbitmqadmin
RUN curl https://raw.githubusercontent.com/rabbitmq/rabbitmq-server/v3.8.14/deps/rabbitmq_management/bin/rabbitmqadmin > /usr/bin/rabbitmqadmin && chmod +x /usr/bin/rabbitmqadmin
