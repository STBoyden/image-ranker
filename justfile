set export
CGO_ENABLED := "1"

[private]
prebuild:
    mkdir -p out

generate:
    go generate

build: generate prebuild
    go build -o out/app main.go

run: build
    ./out/app

format:
    go run mvdan.cc/gofumpt@latest -w -extra **/*.go

dev:
    go run github.com/cosmtrek/air@latest