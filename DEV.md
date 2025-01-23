# Dev Docs

## Building

### Pre-requisites

- Golang 1.22
- [Go Imports](#go-imports)

### Build Commands

```shell
make all
```

## Go Imports

One of the Makefile steps includes running an automatic imports fixing step. It requires `goimports`:

```shell
go install golang.org/x/tools/cmd/goimports@latest
```
