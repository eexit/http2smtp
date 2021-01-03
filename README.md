# HTTP to SMTP

[![docker image](https://img.shields.io/docker/v/eexit/http2smtp?label=docker-image&sort=date)](https://hub.docker.com/repository/docker/eexit/http2smtp) [![ci](https://github.com/eexit/http2smtp/workflows/build/badge.svg)](https://github.com/eexit/http2smtp/actions) [![codecov](https://codecov.io/gh/eexit/http2smtp/branch/master/graph/badge.svg?token=XH18EYLDLZ)](https://codecov.io/gh/eexit/http2smtp)

This small app allows to connect any HTTP-based vendor mailer to a SMTP server. Developped because of the lack of capability to test email sending thru APIs.

### Supported vendors

- [SparkPost RFC 822 transmission](https://developers.sparkpost.com/api/transmissions/#transmissions-post-send-rfc822-content)

## Usage

### Docker image

1. Checkout this repo or only copy the `.env.dist` and `docker-compose.yml` files
1. Rename `.env.dist` into `.env`
2. Update the values accordingly

```bash
# Pull the images
docker-compose pull
# Up the stack
docker-compose up http2smtp
Creating http2smtp_smtp_1 ... done
Creating http2smtp_http2smtp_1 ... done
Attaching to http2smtp_http2smtp_1
http2smtp_1  | {"level":"info","version":"v0.1.0+dev","time":"2021-01-03T22:32:08Z","message":"app is starting"}
http2smtp_1  | {"level":"info","version":"v0.1.0+dev","smtp":{"addr":"smtp:1025","id":"go:net/smtp"},"time":"2021-01-03T22:32:08Z","message":"dialing to smtp server"}
http2smtp_1  | {"level":"info","version":"v0.1.0+dev","time":"2021-01-03T22:32:08Z","message":"listening on http:8080"}
```

## Vendor endpoints

### SparkPost

#### Inline transmission

_Not supported yet._

#### RFC 822 transmission

SparkPost documentation: https://developers.sparkpost.com/api/transmissions/#transmissions-post-send-rfc822-content

    POST /sparkpost/api/v1/transmissions

Basic validation is enforced, only the recipients list and the RFC 822 content are used.

