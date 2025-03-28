<a href="https://kurrent.io">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="KurrentLogo-White.png">
    <source media="(prefers-color-scheme: light)" srcset="KurrentLogo-Black.png">
    <img alt="Kurrent" src="KurrentLogo-Plum.png" height="50%" width="50%">
  </picture>
</a>

# KurrentDB Go Client

[![PkgGoDev](https://pkg.go.dev/badge/github.com/kurrent-io/KurrentDB-Client-Go)](https://pkg.go.dev/github.com/kurrent-io/KurrentDB-Client-Go)
[![CI](https://github.com/kurrent-io/KurrentDB-Client-Go/actions/workflows/ci.yml/badge.svg)](https://github.com/kurrent-io/KurrentDB-Client-Go/actions/workflows/ci.yml)

KurrentDB is the event-native database, where business events are immutably stored and streamed. Designed for
event-sourced, event-driven, and microservices architectures.

"KurrentDB Go Client" is the client for talking to [KurrentDB](https://kurrent.io/).

The fastest way to add this client to a project is to run `go get github.com/kurrent-io/KurrentDB-Client-Go@latest` with
go, See [INSTALL.md](/INSTALL.md) for detailed installation instructions and troubleshooting.

### Documentation

* [API Reference](https://pkg.go.dev/github.com/kurrent.io/KurrentDB-Client-Go?tab=doc)
* [Samples](https://github.com/kurrent-io/KurrentDB-Client-Go/tree/main/samples)

## Communities

[Join our global community](https://www.kurrent.io/community) of developers.

- [Discuss](https://discuss.kurrent.io/)
- [Discord (Kurrent)](https://discord.gg/Phn9pmCw3t)
- [Discord (ddd-cqrs-es)](https://discord.com/invite/sEZGSHNNbH)

## Contributing

Development is done on the `main` branch.
We attempt to do our best to ensure that the history remains clean and to do so, we generally ask contributors to squash
their commits into a set or single logical commit.

- [Create an issue](https://github.com/kurrent-io/KurrentDB-Client-Go/issues)
- [Documentation](https://docs.kurrent.io/)
- [Contributing guide](https://github.com/kurrent-io/KurrentDB-Client-Go/blob/main/CONTRIBUTING.md)

## Building the client

The client is built using the [Go](https://golang.org/) programming language. To build the client, you need to have Go
installed on your machine. You can download it from the official Go website.
Once you have Go installed, you can build the client by running the following command in the root directory of the
project:

```bash
make build
```

The build scripts: `build.sh` and `build.ps1` are also available for Linux and Windows respectively to simplify the
build process.

### Running the tests

Testing requires [Docker](https://www.docker.com/) and [Docker Compose](https://www.docker.com/) to be installed.

Start all required KurrentDB services using the provided `docker-compose` configuration:

```bash
make start-kurrentdb
```

To stop the services, you can run:

```bash
make stop-kurrentdb
```

You can launch the tests as follows:

```
make test
```

Alternatively, you can run the tests using the `go test` command:

```bash
go test ./...
```

By default the tests use `docker.eventstore.com/eventstore-ce/eventstoredb-ce:latest`. If you want to run the tests with
a specific Docker image, you can set the following environment variables:

For example, to use `docker.kurrent.io/kurrent-staging/kurrentdb:ci`, you would set:

| Variable Name                | Value                             |
|------------------------------|-----------------------------------|
| `EVENTSTORE_DOCKER_REGISTRY` | docker.kurrent.io/kurrent-staging |
| `EVENTSTORE_DOCKER_IMAGE`    | kurrentdb                         |
| `EVENTSTORE_DOCKER_TAG`      | ci                                |

These variables combine to form the complete image reference: `docker.kurrent.io/kurrent-staging/kurrentdb:ci`

## More resources

- [Release notes](https://kurrent.io/blog/release-notes)
- [Beginners Guide to Event Sourcing](https://kurrent.io/event-sourcing)
- [Articles](https://kurrent.io/blog)
- [Webinars](https://kurrent.io/webinars)
- [Contact us](https://kurrent.io/contact)