# 📩 HTTP to SMTP
[![ci](https://github.com/eexit/http2smtp/workflows/build/badge.svg)](https://github.com/eexit/http2smtp/actions) [![codecov](https://codecov.io/gh/eexit/http2smtp/branch/master/graph/badge.svg?token=XH18EYLDLZ)](https://codecov.io/gh/eexit/http2smtp) [![version](https://img.shields.io/github/v/tag/eexit/http2smtp?label=version&logo=github&sort=semver)](https://github.com/eexit/http2smtp/releases) [![license](https://img.shields.io/github/license/eexit/http2smtp)](https://github.com/eexit/http2smtp/blob/master/LICENSE)

An API that forwards HTTP-backed vendor mailer calls to SMTP.

Plug a [MailHog](https://github.com/mailhog/MailHog) or [MailCatcher](https://mailcatcher.me/) to API email sending vendors such as [SparkPost](https://www.sparkpost.com/), [Mailgun](https://www.mailgun.com/) or [SendGrid](https://sendgrid.com/) for testing purposes.

#### Features

- :white_check_mark: 100% code coverage
- :whale: Light Docker image available
- :zap: AWS Lambda Function 3-commands deployment

## Usage

See [examples](examples).

:zap: ProTip: for tracing purposes, this app kinda supports [W3C Trace Context recommendation](https://www.w3.org/TR/trace-context/). Configure the env var `TRACEPARENT_HEADER` and inject any trace into this header value. All log entries will be contextualized with the given value.

### Docker image [![docker pull](https://img.shields.io/docker/pulls/eexit/http2smtp)](https://hub.docker.com/repository/docker/eexit/http2smtp) [![size](https://img.shields.io/docker/image-size/eexit/http2smtp?sort=semver)](https://hub.docker.com/repository/docker/eexit/http2smtp)

1. Checkout this repo or only copy the `.env.dist` and `docker-compose.yml` files
1. Rename `.env.dist` into `.env`
1. Optional: update the values accordingly
1. Pull images and run `docker-compose up http2smtp`

### AWS Lambda Function ![aws-lambda-ready](https://img.shields.io/badge/aws-lambda--ready-orange?logo=amazon-aws&style=flat)

:zap: This project is also shipped for an AWS Lambda Function-ready. Check out the [README](cmd/http2smtp-lambda).

## Vendors

### [SparkPost](https://developers.sparkpost.com/api/)

    POST /sparkpost/api/v1/transmissions

SparkPost supports either [inline](https://developers.sparkpost.com/api/transmissions/#transmissions-post-send-inline-content) or [RFC 822 transmissions](https://developers.sparkpost.com/api/transmissions/#transmissions-post-send-rfc822-content). For now, only the latter one is supported.

Basic validation is enforced, only the recipients list email and the RFC 822 content are used and mandatory.

### [Mailgun](https://documentation.mailgun.com/en/latest)

    POST /mailgun/api/v3/messages.mime



## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributors

[![contributors](https://contrib.rocks/image?repo=eexit/http2smtp)](https://github.com/eexit/http2smtp/graphs/contributors)
