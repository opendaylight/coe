FROM golang:1-alpine as build
RUN apk --no-cache add curl gcc git git musl-dev
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/git.opendaylight.org/gerrit/p/coe.git/watcher
ADD . .
RUN dep ensure -vendor-only
RUN go install -ldflags "-linkmode external -extldflags -static"

FROM scratch as runtime
COPY --from=build /go/bin/watcher /go/bin/watcher
ENTRYPOINT ["/go/bin/watcher"]
CMD ["odl"]
