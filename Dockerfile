FROM alpine:3.7
ARG release_dir=release
RUN apk --update --no-cache add ca-certificates
ADD $release_dir/ttn-*-linux-amd64 /usr/local/bin/
RUN cd /usr/local/bin && \
  for bin in `ls ttn-*-linux-amd64`; \
  do ln -sf $bin /usr/local/bin/${bin%-linux-amd64}; \
  done
RUN chmod 755 /usr/local/bin/ttn-*
