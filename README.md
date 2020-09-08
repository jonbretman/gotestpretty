# gotestpretty

Prettier output for `go test`. Formatting is heavily inspired by (read "stolen from") [Jest](https://jestjs.io/en/).

### Install

```
go get -u github.com/jonbretman/gotestpretty
```

### Usage

`gotestpretty` takes the output of `go test` with the `-json` flag. For example:

```
$ go test ./... -v -json | gotestpretty
```

#### Turns this...

![Before](https://user-images.githubusercontent.com/1671025/59179078-d5b06180-8b58-11e9-96bd-2cb6bc4a1553.png)

#### ...into this

![After](https://user-images.githubusercontent.com/1671025/59179081-d5b06180-8b58-11e9-87bc-67f7fad22ac0.png)

### Features

- Colorized output to make it clear which tests passed and which failed
- Summary line at the end indicating how many passes, failures, and skipped tests there were
- Summary of failed tests at the end with a small code snippet pointing at the line the error occurred
- Subtests are shown clearly under the parent test
