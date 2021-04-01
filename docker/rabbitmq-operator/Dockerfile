FROM debian:stable-slim

RUN apt-get update && apt-get install python3 curl -y && apt-get clean

# Install rabbitmqadmin
RUN curl https://raw.githubusercontent.com/rabbitmq/rabbitmq-server/v3.8.14/deps/rabbitmq_management/bin/rabbitmqadmin > /usr/bin/rabbitmqadmin && chmod +x /usr/bin/rabbitmqadmin
