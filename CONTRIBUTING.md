# Contributing to Poteto

I appreciate your consideration to contribute to this project! This document is a guide to help make your contribution easier and more effective. Thank you :)

## Getting Started

### Prerequisites

- Docker or go1.21.8

## Build

- You can use devcontainer
- Test
```bash
./scripts/run_test.sh
```

## How to Contribute

### Reporting Issues

If you find a bug or have a feature request, please open an issue on GitHub.

1. Check [the Issue Tracker](https://github.com/poteto-go/poteto/issues) for existing issues.
2. When requesting a new issue or feature, please use [the templates](https://github.com/poteto-go/poteto/issues/new?assignees=&labels=&projects=&template=-bug-----feature--issue-title.md&title=) and provide as much detail as possible.

### Development

1. Check [the Issue Tracker](https://github.com/poteto-go/poteto/issues), make sure if there is anything relevant to the problem you are trying to solve. Or create New Issue.
1. Fork repo
1. Create a new branch.

   ```bash
   git checkout -b your-account/feature
   ```

1. Make changes to the code and run tests to make sure everything is working properly.
1. Write a clear commit message.
1. Run linter

   ```bash
   golangci-lint run -c rules/.golangci.yaml
   ```

1. Pull request

### Thank you :)