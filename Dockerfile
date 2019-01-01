FROM alpine:latest

COPY dist/goxr-linux-amd64 /usr/bin/goxr
COPY dist/goxr-server-linux-amd64 /usr/bin/goxr-server
COPY .docker/install-generic.sh /tmp/install-generic.sh
COPY .docker/build.env /tmp/build.env

RUN chmod +x /usr/bin/goxr \
    && chmod +x /usr/bin/goxr-server \
    && apk add --no-cache curl wget make tar zip gzip \
    && sh /tmp/install-generic.sh \
    && rm -rf \
        /usr/share/man \
        /tmp/* \
        /var/cache/apk
