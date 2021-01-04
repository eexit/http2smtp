FROM golang:1-alpine AS builder
RUN apk --update add ca-certificates
RUN update-ca-certificates
ARG version
LABEL version=${version}
WORKDIR /go/src/github.com/eexit/http2smtp
COPY . .
# Inject the build version: https://blog.alexellis.io/inject-build-time-vars-golang/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X github.com/eexit/http2smtp/internal/api.Version=${version}" \
    -o /http2smtp \
    ./cmd/http2smtp

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /http2smtp /
EXPOSE 8080
CMD ["/http2smtp"]
