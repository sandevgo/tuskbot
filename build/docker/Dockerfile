FROM node:20-bookworm-slim

RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-venv \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt-get/lists/*

ARG TARGETARCH

COPY ./bin/tusk-linux-${TARGETARCH} /usr/local/bin/tusk

ENV TUSK_RUNTIME_PATH=/root/.tuskbot
VOLUME ["/root/.tuskbot"]

ENTRYPOINT ["tusk"]
CMD ["start"]