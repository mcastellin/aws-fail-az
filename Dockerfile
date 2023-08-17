############################################################################
# This is an example Dockerfile to use as a reference for tool installation
# until we have a setup to distribute binaries
############################################################################
FROM golang:1.21 as builder

WORKDIR /go/src/

RUN : \
    && git clone https://github.com/mcastellin/aws-fail-az.git \
    && cd aws-fail-az/ \
    && git checkout feat/fail-ecs \
    && go build \
    && :

FROM debian

COPY --from=builder /go/src/aws-fail-az/aws-fail-az /usr/local/sbin/

ENTRYPOINT ["/usr/local/sbin/aws-fail-az"]

CMD ["help"]
