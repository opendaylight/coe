FROM golang:1 as build
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/git.opendaylight.org/gerrit/p/coe.git/odlCNIPlugin/odlovs-cni
COPY . .
RUN dep ensure -vendor-only
RUN go install
CMD ["go-wrapper", "run"] # ["app"]

FROM centos as runtime
COPY --from=build /go/bin/odlovs-cni /go/bin/odlovs-cni
