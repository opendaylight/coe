FROM golang:1
RUN curl https://glide.sh/get | sh
WORKDIR /go/src/git.opendaylight.org/gerrit/p/coe.git/watcher
COPY . .
RUN glide install
RUN go-wrapper install
CMD ["go-wrapper", "run"] # ["app"]
