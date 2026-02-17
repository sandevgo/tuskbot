FROM node:20-bookworm-slim

RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-venv \
    git \
    && rm -rf /var/lib/apt-get/lists/*

COPY ./bin/tusk-linux-amd64 /usr/local/bin/tusk

ENV TUSK_RUNTIME_PATH=/data
VOLUME ["/data"]

ENTRYPOINT ["tusk"]
CMD ["start"]