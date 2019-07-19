FROM alpine:3.8

RUN addgroup -g 886 thethings && adduser -u 886 -S -G thethings thethings

RUN apk --update --no-cache add ca-certificates curl

COPY ttn-lw-stack /bin/ttn-lw-stack
RUN ln -s /bin/ttn-lw-stack /bin/stack
RUN chmod 755 /bin/ttn-lw-stack

COPY ttn-lw-cli /bin/ttn-lw-cli
RUN ln -s /bin/ttn-lw-cli /bin/cli
RUN chmod 755 /bin/ttn-lw-cli

COPY public /srv/ttn-lorawan/public

EXPOSE 1700/udp 1882 8882 1883 8883 1884 8884 1885 8885 1887 8887

RUN mkdir /srv/ttn-lorawan/public/blob

VOLUME ["/srv/ttn-lorawan/public/blob"]

ENV TTN_LW_BLOB_LOCAL_DIRECTORY=/srv/ttn-lorawan/public/blob \
    TTN_LW_IS_DATABASE_URI=postgres://root@cockroach:26257/ttn_lorawan?sslmode=disable \
    TTN_LW_REDIS_ADDRESS=redis:6379 \
    TTN_LW_HEALTHCHECK_URL=http://localhost:1885/healthz/live

HEALTHCHECK --interval=1m --timeout=5s CMD curl -f $TTN_LW_HEALTHCHECK_URL || exit 1

USER thethings:thethings
