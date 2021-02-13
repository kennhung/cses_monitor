FROM golang:1.14-alpine as builder

ENV GO111MODULE=on

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o main .

# production environment
FROM ubuntu:latest

WORKDIR /dist
COPY --from=builder /build/main .

CMD ["/dist/main"]