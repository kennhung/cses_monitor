FROM golang:1.14-alpine as builder

ENV GO111MODULE=on

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o main .

# production environment
FROM alpine

WORKDIR /app
COPY --from=builder /build/main /app/main

CMD ["/app/main"]
