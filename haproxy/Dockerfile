FROM golang:1.20-alpine3.18 AS minica-builder
COPY minica.go /go
COPY genCert.sh /go
WORKDIR /go
RUN apk add --no-cache bash
RUN /bin/bash /go/genCert.sh


FROM haproxy:2.8-alpine3.18
MAINTAINER KhulnaSoft Ltd
LABEL khulnasoft.role=system

USER root

ARG is_dev_build

ENV ENABLE_AUTH=true \
    IS_DEV_BUILD=$is_dev_build \
    UI_SERVICE_NAME=khulnasoft-ui \
    UI_SERVICE_PORT=8081 \
    API_SERVICE_HOST=khulnasoft-server \
    API_SERVICE_PORT=8080 \
    KHULNASOFT_FILE_SERVER_HOST=khulnasoft-file-server \
    KHULNASOFT_FILE_SERVER_PORT=9000 \
    CUSTOMER_UNIQUE_ID="" \
    FORCE_HTTPS_REDIRECT="true"

COPY --from=minica-builder /go/minica.pem /usr/local/etc/haproxy/khulnasoft.crt
COPY --from=minica-builder /go/minica-key.pem /usr/local/etc/haproxy/khulnasoft.key
COPY router-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

RUN apk update --no-cache \
    && chmod +x /usr/local/bin/docker-entrypoint.sh \
    && apk add --no-cache bash lua5.3 lua5.3-socket curl \
    && rm -rf /var/cache/apk/* \
    && mkdir -p /var/log/haproxy \
    && touch /var/log/haproxy/haproxy.log

COPY haproxy.cfg /usr/local/etc/haproxy/haproxy.cfg
