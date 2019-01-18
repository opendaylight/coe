FROM golang:1 as builder
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
# TODO Isn't there already an 'official' Go base image with dep already available?!
WORKDIR /go/src/git.opendaylight.org/gerrit/p/coe.git/odlCNIPlugin/odlovs-cni
COPY . .
RUN dep ensure -vendor-only
RUN go install
CMD ["go-wrapper", "run"] # ["app"]

FROM alpine:latest as runtime
ADD https://storage.googleapis.com/kubernetes-release/release/v1.12.0/bin/linux/amd64/kubectl /usr/local/bin/kubectl
RUN set -x && \
    apk add --no-cache curl ca-certificates && \
    chmod +x /usr/local/bin/kubectl
RUN mkdir -p /root/.kube/
RUN mkdir -p /opt/cni/bin/
RUN mkdir -p /etc/cni/net.d/
RUN mkdir -p /root/odlcni/
WORKDIR /root/odlcni/
COPY ./container/odlcni.sh /root/odlcni/
COPY --from=builder /go/bin/odlovs-cni /root/odlcni/
RUN chmod 775 /root/odlcni/*
ENTRYPOINT ["./odlcni.sh"]
