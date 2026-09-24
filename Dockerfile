
FROM golang:alpine3.24 AS build-go

WORKDIR /usr/src/faucet.florin.so

COPY go.mod go.mod
COPY go.sum go.sum

RUN go get ./...

COPY main.go main.go
COPY internal internal
COPY graph graph

RUN go build -o faucet.florin.so

FROM alpine:3.24 AS runner

COPY --from=build-go /usr/src/faucet.florin.so/faucet.florin.so /bin/faucet.florin.so

ENTRYPOINT ["/bin/faucet.florin.so"]
