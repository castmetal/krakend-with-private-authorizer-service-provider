FROM golang:1.17.9-alpine
USER root

ARG ERROR_FLAG
ARG PRIVATE_FLAG
ARG TOKEN_HEADER
ARG AUTHORIZER_SERVICE_URL

ENV ERROR_FLAG ${ERROR_FLAG}
ENV PRIVATE_FLAG ${PRIVATE_FLAG}
ENV TOKEN_HEADER ${TOKEN_HEADER}
ENV AUTHORIZER_SERVICE_URL ${AUTHORIZER_SERVICE_URL}
ENV LOGIN_SERVICE_URL ${LOGIN_SERVICE_URL}


RUN wget -O krakend.tgz https://repo.krakend.io/bin/krakend_2.0.4_amd64.tar.gz
RUN tar -C / -xzf krakend.tgz

RUN apk update \
    && apk add --upgrade apk-tools \
    && apk upgrade --available

RUN apk add make gcc musl-dev

ENV GLIBC_REPO=https://github.com/sgerrand/alpine-pkg-glibc
ENV GLIBC_VERSION=2.31-r0

RUN set -ex && \
    apk --update add libstdc++ curl ca-certificates && \
    for pkg in glibc-${GLIBC_VERSION} glibc-bin-${GLIBC_VERSION}; \
    do curl -sSL ${GLIBC_REPO}/releases/download/${GLIBC_VERSION}/${pkg}.apk -o /tmp/${pkg}.apk; done && \
    apk add --allow-untrusted /tmp/*.apk && \
    rm -v /tmp/*.apk && \
    /usr/glibc-compat/sbin/ldconfig /lib /usr/glibc-compat/lib

COPY ./plugins /etc/krakend/plugins
COPY krakend.json /etc/krakend/

RUN cd /etc/krakend/plugins/krakend-private-auth-server-response &&  go get krakend-private-auth-server-response
RUN cd /etc/krakend/plugins/krakend-private-auth-server-response && GO111MODULE=on CGO_ENABLED=1 GOOS=linux go build -mod=mod -buildmode=plugin -o krakend-private-auth-server-response.so krakend-private-auth-server-response.go
RUN mv /etc/krakend/plugins/krakend-private-auth-server-response/krakend-private-auth-server-response.so /etc/krakend/plugins/

EXPOSE 8001

CMD CGO_ENABLED=1 \
    FC_ENABLE=1 \
    ERROR_FLAG=${ERROR_FLAG} \
    PRIVATE_FLAG=${PRIVATE_FLAG} \
    TOKEN_HEADER=${TOKEN_HEADER} \
    AUTHORIZER_SERVICE_URL=${AUTHORIZER_SERVICE_URL} \
    LOGIN_SERVICE_URL=${LOGIN_SERVICE_URL} \
    KRAKEND_PORT=8001 \
    krakend run -d -c /etc/krakend/krakend.json -p 8001
