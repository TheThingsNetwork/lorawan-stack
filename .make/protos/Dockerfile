FROM golang:1.10 AS go_builder

ENV GOGO_PROTOBUF_VERSION=1.1.1

RUN git clone https://github.com/gogo/protobuf.git ${GOPATH}/src/github.com/gogo/protobuf && \
    cd ${GOPATH}/src/github.com/gogo/protobuf && \
    git checkout v${GOGO_PROTOBUF_VERSION}

ADD ./.make/protos/protoc-gen-gogottn.go ${GOPATH}/src/protoc-gen-gogottn.go

RUN go build -o /go/bin/protoc-gen-gogottn ${GOPATH}/src/protoc-gen-gogottn.go

FROM htdvisser/protocontainer:20180821

COPY --from=go_builder /go/bin/protoc-gen-* /usr/local/bin/
