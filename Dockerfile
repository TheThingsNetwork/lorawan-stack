FROM alpine:3.8

RUN addgroup -g 886 thethings && adduser -u 886 -S -G thethings thethings

ARG release_dir=release
RUN apk --update --no-cache add ca-certificates
ADD $release_dir/ttn-*-linux-amd64 /bin/
ADD public /srv/ttn-lorawan/public
RUN cd /bin && \
  for binary in `ls ttn-*-linux-amd64`; \
  do ln -sf $binary /bin/${binary%-linux-amd64}; \
  done
RUN chmod 755 /bin/ttn-*

EXPOSE 1700/udp 1882 8882 1883 8883 1884 8884 1885 8885

VOLUME ["/srv/ttn-lorawan/device-repository", "/srv/ttn-lorawan/blob", "/srv/ttn-lorawan/frequency-plans"]

ENV TTN_LW_AS_DEVICE_REPOSITORY_DIRECTORY=/srv/ttn-lorawan/device-repository \
    TTN_LW_BLOB_LOCAL_DIRECTORY=/srv/ttn-lorawan/blob \
    TTN_LW_FREQUENCY_PLANS_DIRECTORY=/srv/ttn-lorawan/frequency-plans \
    TTN_LW_IS_DATABASE_URI=postgres://root@cockroach:26257/ttn_lorawan?sslmode=disable \
    TTN_LW_REDIS_ADDRESS=redis:6379

USER thethings:thethings
