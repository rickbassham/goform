FROM golang:1.13-stretch

RUN go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

RUN mkdir /app
COPY . /app

WORKDIR /app

RUN make go-test
RUN make go-lint
