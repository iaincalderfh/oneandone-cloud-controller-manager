FROM golang:1.10-alpine as build-env
COPY main.go /go/src/fasthosts.com/claas/oneandone-cloud-controller-manager/
COPY vendor /go/src/fasthosts.com/claas/oneandone-cloud-controller-manager/vendor
COPY pkg /go/src/fasthosts.com/claas/oneandone-cloud-controller-manager/pkg
WORKDIR /go/src/fasthosts.com/claas/oneandone-cloud-controller-manager
RUN go build -o oneandone-cloud-controller-manager

FROM alpine:3.7
RUN apk add --no-cache ca-certificates
COPY --from=build-env /go/src/fasthosts.com/claas/oneandone-cloud-controller-manager/oneandone-cloud-controller-manager /bin/
CMD ["/bin/oneandone-cloud-controller-manager", "--cloud-provider=oneandone"]