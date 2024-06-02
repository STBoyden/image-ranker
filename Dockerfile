FROM golang:alpine AS builder
RUN apk add just
WORKDIR /src
COPY . .
RUN just build

FROM alpine:latest
COPY --from=builder /src/out/app /bin/app
EXPOSE 3000
CMD ["/bin/app"]