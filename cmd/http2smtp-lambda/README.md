# :zap: http2smtp for AWS Lambda

Deploy in 2 commands!

## Usage

```
$ make
+-------------------------------------------------------------------+
| Make Usage                                                        |
+-------------------------------------------------------------------+
|- config          -> Prints the resolved stack config
|- deploy          -> Deploys the stack
|- download        -> Gets the latest version of the binary
|- install-sls     -> Installs serverless (via NPM)
|- remove          -> Removes the stack
```

## Deploying

- Install [serverless](https://www.serverless.com/) [yourself](https://www.serverless.com/framework/docs/getting-started/) or run `make install-sls`
- Edit the `serverless.yml` file with the [wished configuration](https://www.serverless.com/framework/docs/providers/aws/guide/serverless.yml/)
- Run `make download && make deploy`

### Deploying a specific version

If you wish to download a specific version of http2smtp, you can prefix the download URL as env var prior calling `make download`:

```bash
$ URL=https://api.github.com/repos/eexit/http2smtp/releases/tags/v1.2.3 make download
```

Then deploy using `make deploy`.
