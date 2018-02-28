FROM golang:1
RUN curl https://glide.sh/get | sh
WORKDIR /go/src/git.opendaylight.org/gerrit/p/coe.git/odlCNIPlugin/odlovs-cni
COPY . .
RUN glide install -v
RUN go-wrapper install
CMD ["go-wrapper", "run"] # ["app"]
