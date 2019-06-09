test: bin/gotestpretty
	go test ./... -json | bin/gotestpretty

bin/gotestpretty: main.go
	go build -o bin/gotestpretty main.go