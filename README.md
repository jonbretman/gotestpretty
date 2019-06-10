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

#### Before
![Before](https://user-images.githubusercontent.com/1671025/59179078-d5b06180-8b58-11e9-96bd-2cb6bc4a1553.png)

#### After
![After](https://user-images.githubusercontent.com/1671025/59179081-d5b06180-8b58-11e9-87bc-67f7fad22ac0.png)
