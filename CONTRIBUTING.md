# How to contribute

Thank you for reading this, this project is open and my only wish is to make it better.

Before you get started, be aware that I'm not a Go expert and still in the learning curve. I believe some pieces of this software could be done in a more idiomatic and elegant way, and my desire is to learn from more experienced Go engineers.

## Testing

This project is heavily tested: coverage near 100% and I would like to keep it that way. If some code is not testable, this probably means the design is not right.

How to test the coverage?

```bash
go test ./... -count=1 -coverprofile=coverage.txt -covermode=atomic
```

Then, visualize the coverage report:

```bash
go tool cover -html=coverage.txt
```

## Linting

This projects uses the tool [golangci-lint](https://golangci-lint.run/). You can [install it locally](https://golangci-lint.run/usage/install/#local-installation) and run it before you commit your code.

You may adapt the configuration to improve the standard but avoid lowering it.
