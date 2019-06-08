# gotestpretty

Prettier output for `go test`.

### Install

```
go get -u github.com/jonbretman/gotestpretty
```

### Usage

`gotestpretty` takes the output of `go test` with the `-json` flag. For example:

```
$ go test ./... -v | gotestpretty
```

**Before**
<img width="648" src="https://user-images.githubusercontent.com/1671025/59151635-5de61800-8a2e-11e9-9941-6b0db09eafdb.png">

**After**
<img width="665" src="https://user-images.githubusercontent.com/1671025/59151636-5de61800-8a2e-11e9-90d9-eb6e57383017.png">
