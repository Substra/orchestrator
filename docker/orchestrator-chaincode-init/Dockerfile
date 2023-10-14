FROM debian:bookworm-slim

SHELL ["/bin/bash", "-eo", "pipefail", "-c"]

# Install necessary utilities
RUN apt-get update && apt-get install -y --no-install-recommends curl && rm -rf /var/lib/apt/lists/*

COPY ./scripts/create-indexes.sh /usr/bin/create-indexes

CMD ["/usr/bin/create-indexes"]
