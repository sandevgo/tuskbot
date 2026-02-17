# Use Alpine for musl compatibility (matching Zig-built libs)
FROM golang:alpine AS builder

RUN apk add --no-cache \
    make \
    curl \
    git \
    tar \
    xz

# Install Zig
ARG TARGETARCH
ENV ZIG_VERSION=0.15.2
RUN case "${TARGETARCH}" in \
        "amd64") ZIG_ARCH="x86_64" ;; \
        "arm64") ZIG_ARCH="aarch64" ;; \
        *) echo "Unsupported architecture: ${TARGETARCH}"; exit 1 ;; \
    esac && \
    echo "Downloading Zig for Linux ${ZIG_ARCH}..." && \
    curl -L "https://ziglang.org/download/${ZIG_VERSION}/zig-${ZIG_ARCH}-linux-${ZIG_VERSION}.tar.xz" -o zig.tar.xz && \
    tar -xf zig.tar.xz && \
    mv zig-${ZIG_ARCH}-linux-${ZIG_VERSION} /opt/zig && \
    ln -s /opt/zig/zig /usr/bin/zig && \
    rm zig.tar.xz

# MacOS SDK for cross-compilation
ENV MACOS_SDK=/opt/MacOSX12.3.sdk
RUN curl -L "https://github.com/joseluisq/macosx-sdks/releases/download/12.3/MacOSX12.3.sdk.tar.xz" -o macosx-sdk.tar.xz && \
    tar -xf macosx-sdk.tar.xz -C /opt && \
    rm macosx-sdk.tar.xz

WORKDIR /app

# Cache Go modules
COPY . .
RUN make deps

# Build the binary using Makefile target
ARG BUILD_TARGET
RUN make _build_${BUILD_TARGET}

# Export stage
FROM scratch AS export
COPY --from=builder /app/bin/tusk /
