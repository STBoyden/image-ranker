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
    go run github.com/air-verse/air@latest

build-docker:
    docker build . -t image-ranker:latest

run-docker: build-docker
    docker run --rm -p 3000:3000 --name image-ranker-app image-ranker:latest