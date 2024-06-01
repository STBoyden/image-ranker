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
    go -exec go run mvdan.cc/gofumpt@latest -w -extra **/*.go