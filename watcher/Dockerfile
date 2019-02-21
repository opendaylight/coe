FROM golang:1-alpine as build
RUN apk --no-cache add curl gcc git git musl-dev
WORKDIR /go/src/git.opendaylight.org/gerrit/p/coe.git/watcher
ADD . .
ENV GO111MODULE=on
RUN go install -ldflags "-linkmode external -extldflags -static"

FROM scratch as runtime
COPY --from=build /go/bin/watcher /go/bin/watcher
ENTRYPOINT ["/go/bin/watcher"]
