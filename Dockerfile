FROM golang:1-alpine AS builder
RUN apk --update add ca-certificates
RUN update-ca-certificates
ARG version
LABEL version=${version}
WORKDIR /go/src/github.com/eexit/httpsmtp
COPY . .
# Inject the build version: https://blog.alexellis.io/inject-build-time-vars-golang/
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X github.com/eexit/httpsmtp/internal/server.Version=${version}" \
    -o /httpsmtp \
    ./cmd/httpsmtp

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE 8080
COPY --from=builder /httpsmtp /
CMD ["/httpsmtp"]
