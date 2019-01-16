FROM alpine:latest
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
COPY ./bin/odlovs-cni /root/odlcni/
RUN chmod 775 /root/odlcni/*
ENTRYPOINT ["./odlcni.sh"]
