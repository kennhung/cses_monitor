FROM golang:1.14-alpine

ENV GO111MODULE=on

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o main .

WORKDIR /dist
RUN cp /build/main .

CMD ["/dist/main"]