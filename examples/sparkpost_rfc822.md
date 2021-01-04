# SparkPost RFC 822

API documentation: https://developers.sparkpost.com/api/transmissions/#transmissions-post-send-rfc822-content

![sparkpost_rfc822.gif](sparkpost_rfc822.gif)

Launch the API:

```bash
docker-compose up http2smtp
Creating http2smtp_smtp_1 ... done
Creating http2smtp_http2smtp_1 ... done
Attaching to http2smtp_http2smtp_1
http2smtp_1  | {"level":"info","version":"v0.1.0+dev","time":"2021-01-03T22:32:08Z","message":"app is starting"}
http2smtp_1  | {"level":"info","version":"v0.1.0+dev","smtp":{"addr":"smtp:1025","id":"go:net/smtp"},"time":"2021-01-03T22:32:08Z","message":"dialing to smtp server"}
http2smtp_1  | {"level":"info","version":"v0.1.0+dev","time":"2021-01-03T22:32:08Z","message":"listening on http:8080"}
```

Send the example request:

```bash
http POST :8080/sparkpost/api/v1/transmissions traceparent:$(openssl rand -hex 16) < sparkpost_rfc822.json
```

Logs:

```bash
http2smtp_1  | {"level":"info","version":"v0.1.0+dev","smtp":{"addr":"smtp:1025","id":"go:net/smtp"},"trace_id":"304dfb8a7fbcfbdb1db373da9e39354a","time":"2021-01-04T00:24:27Z","message":"sending message"}
http2smtp_1  | {"level":"debug","version":"v0.1.0+dev","smtp":{"addr":"smtp:1025","id":"go:net/smtp"},"trace_id":"304dfb8a7fbcfbdb1db373da9e39354a","tos":["bob@example.com"],"time":"2021-01-04T00:24:27Z","message":"executing transaction"}
http2smtp_1  | {"level":"debug","version":"v0.1.0+dev","smtp":{"addr":"smtp:1025","id":"go:net/smtp"},"trace_id":"304dfb8a7fbcfbdb1db373da9e39354a","from":"Test <test@example.com>","time":"2021-01-04T00:24:27Z","message":"sending MAIL FROM cmd"}
http2smtp_1  | {"level":"debug","version":"v0.1.0+dev","smtp":{"addr":"smtp:1025","id":"go:net/smtp"},"trace_id":"304dfb8a7fbcfbdb1db373da9e39354a","to":"bob@example.com","time":"2021-01-04T00:24:27Z","message":"sending RCPT cmd"}
http2smtp_1  | {"level":"debug","version":"v0.1.0+dev","smtp":{"addr":"smtp:1025","id":"go:net/smtp"},"trace_id":"304dfb8a7fbcfbdb1db373da9e39354a","time":"2021-01-04T00:24:27Z","message":"sending DATA cmd"}
http2smtp_1  | {"level":"debug","version":"v0.1.0+dev","smtp":{"addr":"smtp:1025","id":"go:net/smtp"},"trace_id":"304dfb8a7fbcfbdb1db373da9e39354a","data":"From: Test <test@example.com>\nTo: Bob <bob@example.com>\nSubject: Hello world!\n\nHello world!","time":"2021-01-04T00:24:27Z","message":"writing data"}
http2smtp_1  | {"level":"debug","version":"v0.1.0+dev","smtp":{"addr":"smtp:1025","id":"go:net/smtp"},"trace_id":"304dfb8a7fbcfbdb1db373da9e39354a","tos":["bob@example.com"],"time":"2021-01-04T00:24:27Z","message":"transaction executed"}
http2smtp_1  | {"level":"info","version":"v0.1.0+dev","smtp":{"addr":"smtp:1025","id":"go:net/smtp"},"trace_id":"304dfb8a7fbcfbdb1db373da9e39354a","accepted":1,"time":"2021-01-04T00:24:27Z","message":"message sent"}
http2smtp_1  | {"level":"info","version":"v0.1.0+dev","trace_id":"304dfb8a7fbcfbdb1db373da9e39354a","verb":"POST","ip":"172.24.0.1","user_agent":"HTTPie/2.3.0","url":"/sparkpost/api/v1/transmissions","code":201,"size":97,"duration":3.273861,"time":"2021-01-04T00:24:27Z","message":"served request"}
```
