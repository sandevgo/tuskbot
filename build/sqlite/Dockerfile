# Stage 1: Environment and Tools
FROM alpine:3.23.3 AS builder

# Install dependencies
RUN apk add --no-cache \
    curl \
    tar \
    xz \
    bash \
    unzip

# Install Zig (Same setup as llamacpp)
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

# Wrapper scripts for Zig
COPY <<EOF /usr/local/bin/zig-cc
#!/bin/sh
zig cc -target \$ZIG_TARGET "\$@"
EOF

COPY <<EOF /usr/local/bin/zig-ar
#!/bin/sh
zig ar "\$@"
EOF

RUN chmod +x /usr/local/bin/zig-cc /usr/local/bin/zig-ar

# --- DOWNLOAD SOURCES ---
WORKDIR /src

# 1. Download SQLite Amalgamation (needed for sqlite3ext.h header)
ENV SQLITE_YEAR=2026
ENV SQLITE_VERSION=3510200
RUN curl -L "https://www.sqlite.org/${SQLITE_YEAR}/sqlite-amalgamation-${SQLITE_VERSION}.zip" -o sqlite.zip && \
    unzip sqlite.zip && \
    mv sqlite-amalgamation-${SQLITE_VERSION} sqlite

# 2. Download sqlite-vec Amalgamation
ENV VEC_VERSION=0.1.6
RUN curl -L "https://github.com/asg017/sqlite-vec/releases/download/v${VEC_VERSION}/sqlite-vec-${VEC_VERSION}-amalgamation.tar.gz" -o vec.tar.gz && \
    tar -xf vec.tar.gz

# --- BUILD FOR LINUX (x86_64) ---
FROM builder AS build-linux
ENV ZIG_TARGET=x86_64-linux-musl
WORKDIR /build/linux

# Compile sqlite-vec.c -> libsqlite_vec.a
# -O3: Optimize
# -fPIC: Position Independent Code (important for CGO)
RUN /usr/local/bin/zig-cc \
    -O3 -fPIC \
    -D_GNU_SOURCE \
    -I/src/sqlite \
    -include sys/types.h \
    -c /src/sqlite-vec.c \
    -o sqlite-vec.o

RUN /usr/local/bin/zig-ar rcs libsqlite_vec.a sqlite-vec.o

# --- BUILD FOR LINUX (ARM64) ---
FROM builder AS build-linux-arm64
ENV ZIG_TARGET=aarch64-linux-musl
WORKDIR /build/linux-arm64

RUN /usr/local/bin/zig-cc \
    -O3 -fPIC \
    -D_GNU_SOURCE \
    -I/src/sqlite \
    -include sys/types.h \
    -c /src/sqlite-vec.c \
    -o sqlite-vec.o

RUN /usr/local/bin/zig-ar rcs libsqlite_vec.a sqlite-vec.o

# --- BUILD FOR MACOS (Apple Silicon / ARM64) ---
FROM builder AS build-macos
ENV ZIG_TARGET=aarch64-macos-none
WORKDIR /build/macos

RUN /usr/local/bin/zig-cc \
    -O3 -fPIC \
    -D_GNU_SOURCE \
    -I/src/sqlite \
    -include sys/types.h \
    -c /src/sqlite-vec.c \
    -o sqlite-vec.o

RUN /usr/local/bin/zig-ar rcs libsqlite_vec.a sqlite-vec.o

# --- GATHER ARTIFACTS ---
FROM alpine:3.23.3 AS artifacts
WORKDIR /output

# Headers
COPY --from=builder /src/sqlite/sqlite3.h /output/include/
COPY --from=builder /src/sqlite/sqlite3ext.h /output/include/
COPY --from=builder /src/sqlite-vec.h /output/include/

# Linux Libs
COPY --from=build-linux /build/linux/libsqlite_vec.a /output/lib/linux_amd64/libsqlite_vec.a
COPY --from=build-linux-arm64 /build/linux-arm64/libsqlite_vec.a /output/lib/linux_arm64/libsqlite_vec.a

# MacOS Libs
COPY --from=build-macos /build/macos/libsqlite_vec.a /output/lib/darwin_arm64/libsqlite_vec.a

CMD ["/bin/sh", "-c", "echo 'Artifacts are ready inside /output'"]
