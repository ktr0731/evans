# Contributing to Evans

Evans has been built with the help of many contributors.

Thank you so much for contributing to Evans!

## Filing an issue

If you find any strange behavior, bugs, or crashes in Evans, please follow the steps below.

1. Read [README.md](https://github.com/ktr0731/evans/blob/master/README.md) and check whether this is the intended behavior.
1. Search for existing issues to see if it has been discussed in the past.

If there are any documents or discussions, please submit an issue from [here](https://github.com/ktr0731/evans/issues).  
If you can't follow the steps or are unsure, please feel free to file an issue.

## Submitting changes

If you would like to merge a change, such as a bug fix or documentation update, into the mainstream, please submit a Pull Request on GitHub.

If you want to add a new feature, it is strongly recommended that you submit an issue for discussion before submitting a Pull Request. By discussing first, you can decide if the feature is needed and if you have a better approach before you start implementing it.

## What should I know before I get started?
You will need some tools for development Evans.

- Go

Evans is written in Go, so you will need a Go environment. It is recommended to use the latest stable Go.

- GNU make

GNU make is used as a task runner. It is used to install, build and run tests for the tools that the project depends on.

We also need Go tools for testing and linting the source code. These can be set up by running the following commands:

``` bash
$ make tools
```

To run the test, run the following command:

``` bash
$ make test
```

E2E tests and some tests use [golden file testing](https://speakerdeck.com/mitchellh/advanced-testing-with-go?slide=19).  
If a golden file should be updated, please run `go test` with `-update` flag. For example:

``` bash 
$ go test -update ./e2e
```

Don't forget to update your credits if you add new dependent modules:

``` bash
$ make credits
```

## Where do I start?
If you're feeling overwhelmed by contributions, it's easiest and safest to start by updating your documentation.
For example, we can correct ambiguous parts of existing documents, correct typos, and proofread English grammar.

It is also a good idea to find issues that have not yet been untouched.
