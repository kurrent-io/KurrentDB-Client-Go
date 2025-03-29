# Installing KurrentDB-Client-Go

## Using the latest version

If you want to use the latest version of this library run:

```sh
go get github.com/kurrent-io/KurrentDB-Client-Go@latest
```

This will record a dependency on `github.com/kurrent-io/KurrentDB-Client-Go` in your go module. You can now import and
use the APIs in your project. The next time you `go build`, `go test`, or `go run` your project,
`github.com/kurrent-io/KurrentDB-Client-Go` and its dependencies will be downloaded (if needed), and detailed dependency
version info will be added to your `go.mod` file (or you can also run `go mod tidy` to do this directly).

## Using a specific version

If you want to use a particular version of the `github.com/kurrent-io/KurrentDB-Client-Go` library,
you can indicate which version of `KurrentDB-Client-Go` your project requires:

```sh
go get github.com/kurrent-io/KurrentDB-Client-Go@v0.1.0
```

You can now import and use the `github.com/kurrent-io/KurrentDB-Client-Go` APIs in your project.
The next time you `go build`, `go test`, or `go run` your project,
`github.com/kurrent-io/KurrentDB-Client-Go` and its dependencies will be downloaded (if needed),
and detailed dependency version info will be added to your `go.mod` file
(or you can also run `go mod tidy` to do this directly).

## Troubleshooting

### Go modules disabled

If you get a message like `cannot use path@version syntax in GOPATH mode`,
you likely do not have go modules enabled. This should be on by default in all
supported versions of Go.

```sh
export GO111MODULE=on
```

Ensure your project has a `go.mod` file defined at the root of your project.
If you do not already have one, `go mod init` will create one for you:

```sh
go mod init
```