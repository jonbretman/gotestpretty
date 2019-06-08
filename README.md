# gotestpretty

### Install

```
go get -u github.com/jonbretman/gotestpretty
```

### Usage

`gotestpretty` takes the output of `go test` with the `-json` flag. For example:

```
$ go test ./... -v | gotestpretty
```
