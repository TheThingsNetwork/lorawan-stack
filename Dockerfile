FROM alpine AS builder

COPY ttn-lw-stack /bin/ttn-lw-stack
RUN chmod 755 /bin/ttn-lw-stack

COPY ttn-lw-cli /bin/ttn-lw-cli
RUN chmod 755 /bin/ttn-lw-cli

COPY data/lorawan-frequency-plans /srv/ttn-lorawan/lorawan-frequency-plans
RUN rm -rf /srv/ttn-lorawan/lorawan-frequency-plans/.git

COPY data/lorawan-webhook-templates /srv/ttn-lorawan/lorawan-webhook-templates
RUN rm -rf /srv/ttn-lorawan/lorawan-webhook-templates/.git

FROM alpine:3.19

RUN addgroup -g 886 thethings && adduser -u 886 -S -G thethings thethings

RUN apk --update --no-cache add ca-certificates curl

COPY --from=builder /bin/ttn-lw-stack /bin/ttn-lw-stack
RUN ln -s /bin/ttn-lw-stack /bin/stack

COPY --from=builder /bin/ttn-lw-cli /bin/ttn-lw-cli
RUN ln -s /bin/ttn-lw-cli /bin/cli

COPY public /srv/ttn-lorawan/public

COPY --from=builder /srv/ttn-lorawan/lorawan-frequency-plans /srv/ttn-lorawan/lorawan-frequency-plans
COPY --from=builder /srv/ttn-lorawan/lorawan-webhook-templates /srv/ttn-lorawan/lorawan-webhook-templates
COPY data/lorawan-devices-index /srv/ttn-lorawan/lorawan-devices-index
RUN chmod 755 -R /srv/ttn-lorawan/lorawan-devices-index

EXPOSE 1700/udp 1881 8881 1882 8882 1883 8883 1884 8884 1885 8885 1887 8887

RUN mkdir /srv/ttn-lorawan/public/blob

VOLUME ["/srv/ttn-lorawan/public/blob"]

ENV TTN_LW_HEALTHCHECK_URL=http://localhost:1885/healthz

HEALTHCHECK --interval=1m --timeout=5s CMD curl -f $TTN_LW_HEALTHCHECK_URL || exit 1

USER thethings:thethings
