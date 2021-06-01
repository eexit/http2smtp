FROM eexit/curl-healthchecker:v1.0.0 AS curl

FROM golang:1.16 AS builder
RUN apt-get update -y \
    && apt-get install -y upx \
    && update-ca-certificates
ARG version
ARG goos=linux
ARG goarch=amd64
LABEL version=${version}
WORKDIR /go/src/github.com/eexit/http2smtp
COPY . .
# Inject the build version: https://blog.alexellis.io/inject-build-time-vars-golang/
RUN CGO_ENABLED=0 GOOS=${goos} GOARCH=${goarch} go build \
    -ldflags "-X github.com/eexit/http2smtp/internal/api.Version=${version}" \
    -o /http2smtp \
    ./cmd/http2smtp
RUN upx /http2smtp

FROM scratch
COPY --from=curl /curl /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /http2smtp /
EXPOSE 80
HEALTHCHECK --interval=5s --timeout=1s --retries=3 \
    CMD ["/curl", "-fIA", "cURL healthcheck", "http://127.0.0.1/healthcheck"]
CMD ["/http2smtp"]
