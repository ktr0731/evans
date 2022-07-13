# syntax=docker/dockerfile:1.2

# Image page: <https://hub.docker.com/_/golang>
FROM golang:1.18-alpine as builder

WORKDIR /src

COPY . /src

# arguments to pass on each go tool link invocation
ENV LDFLAGS="-s -w"

RUN set -x \
    && CGO_ENABLED=0 go build -trimpath -ldflags "$LDFLAGS" -o ./evans . \
    && ./evans --version

WORKDIR /tmp/rootfs

# prepare the rootfs for scratch
RUN set -x \
    && mkdir -p ./bin ./etc/ssl ./tmp ./mount ./.config/evans ./.cache \
    && mv /src/evans ./bin/evans \
    && echo 'evans:x:10001:10001::/tmp:/sbin/nologin' > ./etc/passwd \
    && echo 'evans:x:10001:' > ./etc/group \
    && cp -R /etc/ssl/certs ./etc/ssl/certs \
    && chown -R 10001:10001 ./.config ./.cache \
    && chmod -R 777 ./tmp ./mount ./.config ./.cache

# use empty filesystem
FROM scratch as runtime

LABEL \
    # Docs: <https://github.com/opencontainers/image-spec/blob/master/annotations.md>
    org.opencontainers.image.title="evans" \
    org.opencontainers.image.description="more expressive universal gRPC client" \
    org.opencontainers.image.url="https://github.com/ktr0731/evans" \
    org.opencontainers.image.source="https://github.com/ktr0731/evans" \
    org.opencontainers.image.vendor="evans" \
    org.opencontainers.image.licenses="MIT"

# use an unprivileged user
USER 10001:10001

# import from builder
COPY --from=builder /tmp/rootfs /

WORKDIR "/mount"

ENTRYPOINT ["/bin/evans"]
