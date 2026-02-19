# Этап 1: Подготовка окружения и инструментов
FROM alpine:3.23.3 AS builder

# Установка зависимостей
RUN apk add --no-cache \
    git \
    cmake \
    make \
    curl \
    tar \
    xz \
    bash \
    python3

# Установка Zig
ARG TARGETARCH
ENV ZIG_VERSION=0.15.2
# Маппинг имен: Docker "arm64" -> Zig "aarch64", Docker "amd64" -> Zig "x86_64"
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

# Клонирование llama.cpp
WORKDIR /src
RUN git clone https://github.com/ggerganov/llama.cpp.git . && \
    git checkout b7952

# Скрипты-обертки
COPY <<EOF /usr/local/bin/zig-cc
#!/bin/sh
zig cc -target \$ZIG_TARGET "\$@"
EOF

COPY <<EOF /usr/local/bin/zig-cxx
#!/bin/sh
zig c++ -target \$ZIG_TARGET "\$@"
EOF

COPY <<EOF /usr/local/bin/zig-ar
#!/bin/sh
zig ar "\$@"
EOF

COPY <<EOF /usr/local/bin/zig-ranlib
#!/bin/sh
zig ranlib "\$@"
EOF

RUN chmod +x /usr/local/bin/zig-cc /usr/local/bin/zig-cxx /usr/local/bin/zig-ar /usr/local/bin/zig-ranlib

# --- СБОРКА ПОД LINUX (x86_64) ---
FROM builder AS build-linux
ENV ZIG_TARGET=x86_64-linux-musl
WORKDIR /build/linux
RUN cmake /src \
    -DCMAKE_C_COMPILER="/usr/local/bin/zig-cc" \
    -DCMAKE_CXX_COMPILER="/usr/local/bin/zig-cxx" \
    -DCMAKE_AR="/usr/local/bin/zig-ar" \
    -DCMAKE_RANLIB="/usr/local/bin/zig-ranlib" \
    -DCMAKE_C_FLAGS="-O3 -g0 -march=haswell" \
    -DCMAKE_CXX_FLAGS="-O3 -g0 -march=haswell" \
    -DCMAKE_SYSTEM_NAME=Linux \
    -DCMAKE_SYSTEM_PROCESSOR=x86_64 \
    -DBUILD_SHARED_LIBS=OFF \
    -DLLAMA_STATIC=ON \
    -DLLAMA_NATIVE=OFF \
    -DLLAMA_BUILD_SERVER=OFF \
    -DLLAMA_BUILD_TESTS=OFF \
    -DLLAMA_BUILD_EXAMPLES=OFF
RUN make -j$(nproc) llama ggml

# --- СБОРКА ПОД LINUX (ARM64) ---
FROM builder AS build-linux-arm64
ENV ZIG_TARGET=aarch64-linux-musl
WORKDIR /build/linux-arm64
RUN cmake /src \
    -DCMAKE_C_COMPILER="/usr/local/bin/zig-cc" \
    -DCMAKE_CXX_COMPILER="/usr/local/bin/zig-cxx" \
    -DCMAKE_AR="/usr/local/bin/zig-ar" \
    -DCMAKE_RANLIB="/usr/local/bin/zig-ranlib" \
    -DCMAKE_C_FLAGS="-O3 -g0" \
    -DCMAKE_CXX_FLAGS="-O3 -g0" \
    -DCMAKE_SYSTEM_NAME=Linux \
    -DCMAKE_SYSTEM_PROCESSOR=aarch64 \
    -DBUILD_SHARED_LIBS=OFF \
    -DLLAMA_STATIC=ON \
    -DLLAMA_NATIVE=OFF \
    -DLLAMA_BUILD_SERVER=OFF \
    -DLLAMA_BUILD_TESTS=OFF \
    -DLLAMA_BUILD_EXAMPLES=OFF
RUN make -j$(nproc) llama ggml

# --- СБОРКА ПОД MACOS (Apple Silicon / ARM64) ---
FROM builder AS build-macos
ENV ZIG_TARGET=aarch64-macos-none
WORKDIR /build/macos
# Создаем фейковый install_name_tool (он нужен CMake для тестов, но нам не пригодится для .a при сборке через Zig)
RUN touch /usr/local/bin/install_name_tool && chmod +x /usr/local/bin/install_name_tool
RUN cmake /src \
    -DCMAKE_C_COMPILER="/usr/local/bin/zig-cc" \
    -DCMAKE_CXX_COMPILER="/usr/local/bin/zig-cxx" \
    -DCMAKE_AR="/usr/local/bin/zig-ar" \
    -DCMAKE_RANLIB="/usr/local/bin/zig-ranlib" \
    -DCMAKE_INSTALL_NAME_TOOL="/usr/local/bin/install_name_tool" \
    -DCMAKE_C_FLAGS="-O3 -g0" \
    -DCMAKE_CXX_FLAGS="-O3 -g0" \
    -DCMAKE_SYSTEM_NAME=Darwin \
    -DCMAKE_SYSTEM_PROCESSOR=aarch64 \
    -DBUILD_SHARED_LIBS=OFF \
    -DLLAMA_STATIC=ON \
    -DLLAMA_NATIVE=OFF \
    -DLLAMA_METAL=OFF \
    -DLLAMA_BUILD_SERVER=OFF \
    -DLLAMA_BUILD_TESTS=OFF \
    -DLLAMA_BUILD_EXAMPLES=OFF \
    -DGGML_BLAS=OFF \
    -DGGML_NATIVE=OFF \
    -DGGML_METAL=OFF
RUN make -j$(nproc) llama ggml

# --- СБОР ХЕДЕРОВ И АРТЕФАКТОВ ---
FROM alpine:3.23.3 AS artifacts
WORKDIR /output

# Хедеры
COPY --from=builder /src/include/llama.h /output/include/
COPY --from=builder /src/include/llama-cpp.h /output/include/
COPY --from=builder /src/ggml/include/ggml*.h /output/include/
#COPY --from=builder /src/include /output/include/

# Linux Libs
COPY --from=build-linux /build/linux/src/libllama.a /output/lib/linux_amd64/libllama.a
COPY --from=build-linux /build/linux/ggml/src/libggml.a /output/lib/linux_amd64/libggml.a
COPY --from=build-linux /build/linux/ggml/src/libggml-base.a /output/lib/linux_amd64/libggml-base.a
COPY --from=build-linux /build/linux/ggml/src/libggml-cpu.a /output/lib/linux_amd64/libggml-cpu.a
#COPY --from=build-linux /build/linux/ /output/linux

# Linux ARM64 Libs
COPY --from=build-linux-arm64 /build/linux-arm64/src/libllama.a /output/lib/linux_arm64/libllama.a
COPY --from=build-linux-arm64 /build/linux-arm64/ggml/src/libggml.a /output/lib/linux_arm64/libggml.a
COPY --from=build-linux-arm64 /build/linux-arm64/ggml/src/libggml-base.a /output/lib/linux_arm64/libggml-base.a
COPY --from=build-linux-arm64 /build/linux-arm64/ggml/src/libggml-cpu.a /output/lib/linux_arm64/libggml-cpu.a

# MacOS Libs
COPY --from=build-macos /build/macos/src/libllama.a /output/lib/darwin_arm64/libllama.a
COPY --from=build-macos /build/macos/ggml/src/libggml.a /output/lib/darwin_arm64/libggml.a
COPY --from=build-macos /build/macos/ggml/src/libggml-base.a /output/lib/darwin_arm64/libggml-base.a
COPY --from=build-macos /build/macos/ggml/src/libggml-cpu.a /output/lib/darwin_arm64/libggml-cpu.a
#COPY --from=build-macos /build/macos/ /output/macos

CMD ["/bin/sh", "-c", "echo 'Artifacts are ready inside /output'"]