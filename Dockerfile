FROM alpine:3.8
ARG release_dir=release
RUN apk --update --no-cache add ca-certificates
ADD $release_dir/ttn-*-linux-amd64 /bin/
ADD public /srv/ttn-lorawan/public
RUN cd /bin && \
  for binary in `ls ttn-*-linux-amd64`; \
  do ln -sf $binary /bin/${binary%-linux-amd64}; \
  done
RUN chmod 755 /bin/ttn-*
